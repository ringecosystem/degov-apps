package internal

import (
	"context"
	"fmt"

	"github.com/machinebox/graphql"
)

// DataMetrics represents the data metrics structure from GraphQL response
type DataMetrics struct {
	ProposalsCount          int    `json:"proposalsCount"`
	MemberCount             int    `json:"memberCount"`
	PowerSum                string `json:"powerSum"`
	VotesCount              int    `json:"votesCount"`
	VotesWeightAbstainSum   string `json:"votesWeightAbstainSum"`
	VotesWeightAgainstSum   string `json:"votesWeightAgainstSum"`
	VotesWeightForSum       string `json:"votesWeightForSum"`
	VotesWithParamsCount    int    `json:"votesWithParamsCount"`
	VotesWithoutParamsCount int    `json:"votesWithoutParamsCount"`
	ID                      string `json:"id"`
}

// DataMetricsResponse represents the GraphQL response structure
type DataMetricsResponse struct {
	DataMetrics []DataMetrics `json:"dataMetrics"`
}

// Proposal represents the proposal structure from GraphQL response
type Proposal struct {
	ID             string `json:"id"`
	BlockNumber    string `json:"blockNumber"`
	BlockTimestamp string `json:"blockTimestamp"`
	ProposalID     string `json:"proposalId"`
}

// ProposalsResponse represents the GraphQL response structure for proposals
type ProposalsResponse struct {
	Proposals []Proposal `json:"proposals"`
}

type VoteCast struct {
	ProposalID      string `json:"proposalId"`
	Reason          string `json:"reason"`
	Support         int    `json:"support"`
	Voter           string `json:"voter"`
	Weight          string `json:"weight"`
	TransactionHash string `json:"transactionHash"`
	ID              string `json:"id"`
	BlockNumber     string `json:"blockNumber"`
	BlockTimestamp  string `json:"blockTimestamp"`
}

type VoteCastsResponse struct {
	VoteCasts []VoteCast `json:"voteCasts"`
}

// DegovIndexer handles GraphQL queries to fetch governance data
type DegovIndexer struct {
	client   *graphql.Client
	endpoint string
}

// NewDegovIndexer creates a new DegovIndexer instance with the given endpoint
func NewDegovIndexer(endpoint string) *DegovIndexer {
	client := graphql.NewClient(endpoint)
	return &DegovIndexer{
		client:   client,
		endpoint: endpoint,
	}
}

// GetEndpoint returns the current GraphQL endpoint
func (d *DegovIndexer) GetEndpoint() string {
	return d.endpoint
}

// QueryDataMetrics executes the QueryDataMetrics GraphQL query and returns a single DataMetrics object
func (d *DegovIndexer) QueryGlobalDataMetrics(ctx context.Context) (*DataMetrics, error) {
	query := `
		query QueryDataMetrics {
			dataMetrics(where: {id_eq: "global"}) {
				proposalsCount
				memberCount
				powerSum
				votesCount
				votesWeightAbstainSum
				votesWeightAgainstSum
				votesWeightForSum
				votesWithParamsCount
				votesWithoutParamsCount
				id
			}
		}
	`

	req := graphql.NewRequest(query)

	var response DataMetricsResponse
	if err := d.client.Run(ctx, req, &response); err != nil {
		return nil, fmt.Errorf("failed to execute QueryDataMetrics: %w", err)
	}

	// Return the first item if available, otherwise return nil
	if len(response.DataMetrics) > 0 {
		return &response.DataMetrics[0], nil
	}

	return nil, fmt.Errorf("no data metrics found for global id")
}

// QueryProposalsOffset executes the QueryProposalsOffset GraphQL query and returns proposals list
func (d *DegovIndexer) QueryProposalsOffset(ctx context.Context, offset int) ([]Proposal, error) {
	query := `
		query QueryProposalsOffset($limit: Int!, $offset: Int!) {
			proposals(orderBy: blockNumber_ASC_NULLS_FIRST, limit: $limit, offset: $offset) {
				id
				blockNumber
				blockTimestamp
				proposalId
			}
		}
	`

	req := graphql.NewRequest(query)
	req.Var("limit", 30)
	req.Var("offset", offset)

	var response ProposalsResponse
	if err := d.client.Run(ctx, req, &response); err != nil {
		return nil, fmt.Errorf("failed to execute QueryProposalsOffset: %w", err)
	}

	return response.Proposals, nil
}

func (d *DegovIndexer) QueryVotesOffset(ctx context.Context, offset int, proposalId string) ([]VoteCast, error) {
	query := `
		query QueryVotesOffset($limit: Int!, $offset: Int!, $proposalId: String!) {
			voteCasts(orderBy: blockNumber_ASC_NULLS_FIRST, limit: $limit, offset: $offset, where: {proposalId_eq: $proposalId}) {
				proposalId
				reason
				support
				voter
				weight
				transactionHash
				id
				blockNumber
				blockTimestamp
			}
		}
	`
	req := graphql.NewRequest(query)
	req.Var("limit", 30)
	req.Var("offset", offset)
	req.Var("proposalId", proposalId)

	var response VoteCastsResponse
	if err := d.client.Run(ctx, req, &response); err != nil {
		return nil, fmt.Errorf("failed to execute QueryVotesOffset: %w", err)
	}

	return response.VoteCasts, nil
}
