package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"github.com/ringecosystem/degov-apps/internal"
	"github.com/ringecosystem/degov-apps/internal/database"
	"github.com/ringecosystem/degov-apps/internal/utils"
	"github.com/ringecosystem/degov-apps/models"
)

type DaoSyncService struct {
	db *gorm.DB
}

func NewDaoSyncService() *DaoSyncService {
	return &DaoSyncService{
		db: database.GetDB(),
	}
}

// DaoRegistryConfig represents the structure of the config.yml file
type DaoRegistryConfig map[string][]struct {
	Name   string `yaml:"name"`
	Code   string `yaml:"code"`
	Config string `yaml:"config"`
}

// DaoConfig represents the structure of individual DAO config files
type DaoConfig struct {
	ChainID int    `yaml:"chain_id"`
	Name    string `yaml:"name"`
	// Add other fields from DAO config as needed
}

// StartDaoSyncScheduler starts the scheduled task to sync DAO configurations
func (s *DaoSyncService) StartDaoSyncScheduler(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Run sync immediately on startup
	if err := s.SyncDaos(); err != nil {
		slog.Error("Initial DAO sync failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("DAO sync scheduler stopped")
			return
		case <-ticker.C:
			if err := s.SyncDaos(); err != nil {
				slog.Error("DAO sync failed", "error", err)
			}
		}
	}
}

// SyncDaos fetches the latest DAO configuration and syncs it with the database
func (s *DaoSyncService) SyncDaos() error {
	slog.Info("Starting DAO synchronization")

	// Fetch the registry config
	registryConfig, err := s.fetchRegistryConfig()
	if err != nil {
		return fmt.Errorf("failed to fetch registry config: %w", err)
	}

	// Get chain ID mappings
	chainIDMap := s.getChainIDMap()

	// Track active DAO codes for marking inactive ones
	activeDaoCodes := make(map[string]bool)

	// Process each chain and its DAOs
	for chainName, daos := range registryConfig {
		chainID, exists := chainIDMap[chainName]
		if !exists {
			slog.Warn("Unknown chain name, skipping", "chain", chainName)
			continue
		}

		for _, daoInfo := range daos {
			activeDaoCodes[daoInfo.Code] = true

			// Fetch DAO config details
			daoConfig, err := s.fetchDaoConfig(daoInfo.Config)
			if err != nil {
				slog.Error("Failed to fetch DAO config", "dao", daoInfo.Code, "error", err)
				continue
			}

			// Upsert DAO in database
			dao := &models.Dao{
				ID:         internal.NextIDString(),
				ChainID:    chainID,
				ChainName:  chainName,
				Name:       daoInfo.Name,
				Code:       daoInfo.Code,
				Seq:        0,
				State:      "ACTIVE",
				ConfigLink: daoInfo.Config,
				TimeSyncd:  utils.TimePtrNow(),
			}

			// Override with config values if available
			if daoConfig.ChainID != 0 {
				dao.ChainID = daoConfig.ChainID
			}
			if daoConfig.Name != "" {
				dao.Name = daoConfig.Name
			}

			if err := s.upsertDao(dao); err != nil {
				slog.Error("Failed to upsert DAO", "dao", daoInfo.Code, "error", err)
				continue
			}

			slog.Debug("Successfully synced DAO", "dao", daoInfo.Code, "chain", chainName)
		}
	}

	// Mark DAOs as inactive if they're not in the current config
	if err := s.markInactiveDAOs(activeDaoCodes); err != nil {
		return fmt.Errorf("failed to mark inactive DAOs: %w", err)
	}

	slog.Info("DAO synchronization completed", "active_daos", len(activeDaoCodes))
	return nil
}

// fetchRegistryConfig fetches and parses the main registry configuration
func (s *DaoSyncService) fetchRegistryConfig() (DaoRegistryConfig, error) {
	const configURL = "https://raw.githubusercontent.com/ringecosystem/degov-registry/refs/heads/daos-config/config.yml"

	resp, err := http.Get(configURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var config DaoRegistryConfig
	decoder := yaml.NewDecoder(resp.Body)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return config, nil
}

// fetchDaoConfig fetches and parses individual DAO configuration
func (s *DaoSyncService) fetchDaoConfig(configURL string) (*DaoConfig, error) {
	resp, err := http.Get(configURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DAO config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var config DaoConfig
	decoder := yaml.NewDecoder(resp.Body)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse DAO config YAML: %w", err)
	}

	return &config, nil
}

// getChainIDMap returns a mapping of chain names to chain IDs
func (s *DaoSyncService) getChainIDMap() map[string]int {
	return map[string]int{
		"darwinia": 46,    // Darwinia chain ID
		"ethereum": 1,     // Ethereum mainnet
		"polygon":  137,   // Polygon
		"arbitrum": 42161, // Arbitrum One
		"optimism": 10,    // Optimism
		"bsc":      56,    // Binance Smart Chain
		// Add more chains as needed
	}
}

// upsertDao inserts or updates a DAO in the database
func (s *DaoSyncService) upsertDao(dao *models.Dao) error {
	var existingDao models.Dao
	result := s.db.Where("code = ?", dao.Code).First(&existingDao)

	if result.Error == gorm.ErrRecordNotFound {
		// Insert new DAO
		dao.CTime = time.Now()
		return s.db.Create(dao).Error
	} else if result.Error != nil {
		return result.Error
	}

	// Update existing DAO
	dao.ID = existingDao.ID
	dao.CTime = existingDao.CTime
	dao.UTime = utils.TimePtrNow()

	return s.db.Save(dao).Error
}

// markInactiveDAOs marks DAOs as inactive if they're not in the active list
func (s *DaoSyncService) markInactiveDAOs(activeCodes map[string]bool) error {
	var allDaos []models.Dao
	if err := s.db.Find(&allDaos).Error; err != nil {
		return err
	}

	for _, dao := range allDaos {
		if !activeCodes[dao.Code] && dao.State != "INACTIVE" {
			slog.Info("Marking DAO as inactive", "dao", dao.Code)
			dao.State = "INACTIVE"
			dao.UTime = utils.TimePtrNow()
			if err := s.db.Save(&dao).Error; err != nil {
				slog.Error("Failed to mark DAO as inactive", "dao", dao.Code, "error", err)
			}
		}
	}

	return nil
}
