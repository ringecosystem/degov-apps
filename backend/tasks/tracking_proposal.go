package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	dbmodels "github.com/ringecosystem/degov-apps/database/models"
	gqlmodels "github.com/ringecosystem/degov-apps/graph/models"
	"github.com/ringecosystem/degov-apps/internal"
	"github.com/ringecosystem/degov-apps/services"
	"github.com/ringecosystem/degov-apps/types"
)

type TrackingProposalTask struct {
	daoService          *services.DaoService
	daoConfigService    *services.DaoConfigService
	proposalService     *services.ProposalService
	chipService         *services.DaoChipService
	notificationService *services.NotificationService
}

func NewTrackingProposalTask() *TrackingProposalTask {
	return &TrackingProposalTask{
		daoService:          services.NewDaoService(),
		daoConfigService:    services.NewDaoConfigService(),
		proposalService:     services.NewProposalService(),
		chipService:         services.NewDaoChipService(),
		notificationService: services.NewNotificationService(),
	}
}

// Name returns the task name
func (t *TrackingProposalTask) Name() string {
	return "tracking-proposal"
}

// Execute performs the DAO synchronization
func (t *TrackingProposalTask) Execute() error {
	return t.TrackingProposal()
}

// TrackingProposal tracks proposals for DAOs
func (t *TrackingProposalTask) TrackingProposal() error {
	// Get all DAOs from DaoService.ListDaos
	daos, err := t.daoService.ListDaos(types.BasicInput[*types.ListDaosInput]{})
	if err != nil {
		slog.Error("Failed to list DAOs", "error", err)
		return err
	}

	slog.Info("Found DAOs for proposal tracking", "count", len(daos))

	// Iterate through each DAO and get its config from DaoConfigService
	for _, dao := range daos {
		daoConfig, err := t.daoConfigService.StandardConfig(dao.Code)
		if err != nil {
			slog.Error("Failed to get DAO config", "dao_code", dao.Code, "error", err)
			continue
		}

		slog.Info("Processing DAO",
			"dao_code", dao.Code,
			"dao_name", daoConfig.Name,
			"indexer_endpoint", daoConfig.Indexer.Endpoint)

		if err := t.storeProposals(dao, daoConfig); err != nil {
			slog.Error("Failed to process proposal tracking", "dao_code", dao.Code, "error", err)
			continue
		}
		if err := t.updateProposalsStates(dao, daoConfig); err != nil {
			slog.Error("Failed to update proposal state", "dao_code", dao.Code, "error", err)
			continue
		}
	}
	if err := t.updateDaoChips(); err != nil {
		slog.Warn("Failed to update DAO chips", "error", err)
		return err
	}

	return nil
}

func (t *TrackingProposalTask) storeProposals(dao *gqlmodels.Dao, daoConfig *types.DaoConfig) error {
	indexer := internal.NewDegovIndexer(daoConfig.Indexer.Endpoint)

	offsetTrackingProposal := int(dao.OffsetTrackingProposal)

	lastOffsetTrackingProposal := offsetTrackingProposal

	slog.Info("Starting proposal tracking",
		"dao_code", dao.Code,
		"start_block", lastOffsetTrackingProposal)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		// Query proposals after the last tracked block (correct parameter order)
		proposals, err := indexer.QueryProposalsOffset(ctx, lastOffsetTrackingProposal)
		cancel()

		if err != nil {
			return fmt.Errorf("failed to query proposals: %w", err)
		}

		if len(proposals) == 0 {
			slog.Info("No more proposals found", "dao_code", dao.Code)
			break
		}

		slog.Info("Found proposals", "dao_code", dao.Code, "count", len(proposals))

		// Process each proposal
		for _, proposal := range proposals {
			// Parse block number and timestamp
			blockNumber, err := strconv.Atoi(proposal.BlockNumber)
			if err != nil {
				slog.Error("Failed to parse block number",
					"dao_code", dao.Code,
					"proposal_id", proposal.ProposalID,
					"block_number", proposal.BlockNumber,
					"error", err)
				continue
			}

			// Parse block timestamp
			var proposalCreatedAt *time.Time
			if proposal.BlockTimestamp != "" {
				if timestamp, err := strconv.ParseInt(proposal.BlockTimestamp, 10, 64); err == nil {
					// Convert milliseconds to seconds for time.Unix()
					createdAt := time.Unix(timestamp/1000, (timestamp%1000)*1000000)
					proposalCreatedAt = &createdAt
				}
			}

			// Create proposal link
			proposalLink := fmt.Sprintf("%s/proposal/%s", daoConfig.SiteURL, proposal.ProposalID)

			// Build proposal tracking input
			input := types.ProposalTrackingInput{
				DaoCode:           dao.Code,
				ChainId:           daoConfig.Chain.ID,
				ProposalLink:      proposalLink,
				ProposalID:        proposal.ProposalID,
				ProposalCreatedAt: proposalCreatedAt,
				ProposalAtBlock:   blockNumber,
			}

			// Store proposal tracking (handles existence check internally)
			created, err := t.proposalService.StoreProposalTracking(input)
			if err != nil {
				slog.Error("Failed to store proposal tracking",
					"dao_code", dao.Code,
					"proposal_id", proposal.ProposalID,
					"error", err)
				continue
			}

			if created {
				slog.Info("Inserted new proposal tracking",
					"dao_code", dao.Code,
					"proposal_id", proposal.ProposalID,
					"block_number", blockNumber,
					"proposal_link", proposalLink)
			} else {
				slog.Debug("Proposal already exists, skipping",
					"dao_code", dao.Code,
					"proposal_id", proposal.ProposalID)
			}

			// Update offset tracking proposal
			lastOffsetTrackingProposal += 1
		}

		if lastOffsetTrackingProposal != offsetTrackingProposal {
			if err := t.daoService.UpdateDaoOffsetTrackingProposal(dao.Code, lastOffsetTrackingProposal); err != nil {
				return fmt.Errorf("failed to update last tracking block: %w", err)
			}

			slog.Info("Updated last tracking offset",
				"dao_code", dao.Code,
				"old_offset", offsetTrackingProposal,
				"new_offset", lastOffsetTrackingProposal)
		}
	}

	return nil
}

func (t *TrackingProposalTask) updateProposalsStates(dao *gqlmodels.Dao, daoConfig *types.DaoConfig) error {
	proposals, err := t.proposalService.TrackingStateProposals(types.TrackingStateProposalsInput{
		DaoCode: dao.Code,
		States: []dbmodels.ProposalState{
			dbmodels.ProposalStateUnknown,
			dbmodels.ProposalStatePending,
			dbmodels.ProposalStateActive,
			dbmodels.ProposalStateSucceeded,
			dbmodels.ProposalStateQueued,
		},
	})
	if err != nil {
		slog.Error("Failed to get tracking state proposals",
			"dao_code", dao.Code,
			"error", err)
		return err
	}

	if len(proposals) == 0 {
		slog.Info("No proposals to update state for", "dao_code", dao.Code)
		return nil
	}

	governorAddress := daoConfig.Contracts.Governor
	if governorAddress == "" {
		slog.Warn("No governor contract address configured", "dao_code", dao.Code)
		return nil
	}

	// Get RPC URL
	rpcURL := internal.GetRPCURL(daoConfig.Chain.RPCs, daoConfig.Chain.ID)
	if rpcURL == "" {
		slog.Warn("No RPC URL available for chain", "dao_code", dao.Code, "chain_id", daoConfig.Chain.ID)
		return nil
	}

	// Create Governor contract client
	governorContract, err := internal.NewGovernorContract(rpcURL)
	if err != nil {
		slog.Error("Failed to create Governor contract client", "dao_code", dao.Code, "rpc_url", rpcURL, "error", err)
		return err
	}
	defer governorContract.Close()

	slog.Info("Updating proposal states",
		"dao_code", dao.Code,
		"count", len(proposals),
		"governor_address", governorAddress,
		"rpc_url", rpcURL)

	// Process each proposal individually
	for _, proposal := range proposals {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// Get proposal state from contract
		newState, err := governorContract.GetProposalState(ctx, governorAddress, proposal.ProposalId)
		cancel()

		if err != nil {
			// Update tracking info with error
			if updateErr := t.proposalService.UpdateProposalTrackingError(proposal.ProposalId, dao.Code, err.Error()); updateErr != nil {
				slog.Error("Failed to update proposal tracking error",
					"dao_code", dao.Code,
					"proposal_id", proposal.ProposalId,
					"original_error", err,
					"update_error", updateErr)
			} else {
				slog.Warn("Updated proposal tracking with error",
					"dao_code", dao.Code,
					"proposal_id", proposal.ProposalId,
					"error", err)
			}
			continue
		}

		// Check if state has changed
		if newState != proposal.State {
			// Update proposal state in database
			if err := t.proposalService.UpdateProposalState(proposal.ProposalId, dao.Code, newState); err != nil {
				// Update tracking info with error
				if updateErr := t.proposalService.UpdateProposalTrackingError(proposal.ProposalId, dao.Code, err.Error()); updateErr != nil {
					slog.Error("Failed to update proposal tracking error after state update failure",
						"dao_code", dao.Code,
						"proposal_id", proposal.ProposalId,
						"original_error", err,
						"update_error", updateErr)
				}
				continue
			}

			payload := "{\"old_state\": \"" + string(proposal.State) + "\", \"new_state\": \"" + string(newState) + "\"}"
			if proposal.State == dbmodels.ProposalStateUnknown {
				payload = "{\"new_state\": \"" + string(newState) + "\"}"
			}
			if err := t.notificationService.SaveEvent(dbmodels.NotificationEvent{
				ChainID:    proposal.ChainId,
				DaoCode:    proposal.DaoCode,
				Type:       dbmodels.NotificationTypeStateChanged,
				ProposalID: proposal.ProposalId,
				TimeEvent:  proposal.CTime,
				Payload:    payload,
			}); err != nil {
				slog.Error("Failed to save state change notification event",
					"dao_code", dao.Code,
					"proposal_id", proposal.ProposalId,
					"error", err,
				)
				continue
			}

			slog.Info("Updated proposal state",
				"dao_code", dao.Code,
				"proposal_id", proposal.ProposalId,
				"old_state", proposal.State,
				"new_state", newState)
		}
	}

	return nil
}

func (t *TrackingProposalTask) updateDaoChips() error {
	// Get proposal state counts for all active DAOs
	counts, err := t.proposalService.ProposalStateCount()
	if err != nil {
		slog.Error("Failed to get proposal state counts", "error", err)
		return err
	}

	slog.Info("Retrieved proposal state counts", "total_records", len(counts))

	// Store metrics state chips using DaoChipService
	storeInput := types.StoreDaoChipMetricsStateInput{
		MetricsStates: counts,
	}

	if err := t.chipService.StoreChipMetricsState(storeInput); err != nil {
		slog.Error("Failed to store chip metrics state", "error", err)
		return err
	}

	slog.Info("Successfully updated DAO chips", "metrics_count", len(counts))
	return nil
}
