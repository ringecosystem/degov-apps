package tasks

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/ringecosystem/degov-apps/internal/config"
	"github.com/ringecosystem/degov-apps/services"
	"github.com/ringecosystem/degov-apps/types"
)

// GitHubTag represents a GitHub tag response
type GitHubTag struct {
	Name string `json:"name"`
}

type GithubConfigLink struct {
	BaseLink   string
	ConfigLink string
}

type DaoSyncTask struct {
	daoService *services.DaoService
}

// DaoRegistryConfig represents the structure of individual DAO configuration
type DaoRegistryConfig struct {
	Code   string   `yaml:"code"`
	Tags   []string `yaml:"tags,omitempty"` // Optional tags field
	Config string   `yaml:"config"`
}

type DaoRegistryConfigResult struct {
	RemoteLink GithubConfigLink
	Result     map[string][]DaoRegistryConfig
}

type DaoConfigResult struct {
	Raw    string           // Original YAML content as string
	Config *types.DaoConfig // Parsed YAML content
}

// NewDaoSyncTask creates a new DAO sync task
func NewDaoSyncTask() *DaoSyncTask {
	return &DaoSyncTask{
		daoService: services.NewDaoService(),
	}
}

// Name returns the task name
func (t *DaoSyncTask) Name() string {
	return "dao-sync"
}

// Execute performs the DAO synchronization
func (t *DaoSyncTask) Execute() error {
	return t.SyncDaos()
}

// SyncDaos fetches the latest DAO configuration and syncs it with the database
func (t *DaoSyncTask) SyncDaos() error {
	startTime := time.Now()
	slog.Info("Starting DAO synchronization", "timestamp", startTime.Format(time.RFC3339))

	// Fetch the registry config
	registryConfigResult, err := t.fetchRegistryConfig()
	if err != nil {
		return fmt.Errorf("failed to fetch registry config: %w", err)
	}

	slog.Info("Successfully fetched registry config", "chains", len(registryConfigResult.Result))

	// Track active DAO codes for marking inactive ones
	activeDaoCodes := make(map[string]bool)

	// Process each chain and its DAOs
	for chainName, daos := range registryConfigResult.Result {
		for _, daoInfo := range daos {
			if err := t.processSingleDao(registryConfigResult.RemoteLink, daoInfo, chainName, activeDaoCodes); err != nil {
				slog.Error("Failed to process DAO", "dao", daoInfo.Code, "chain", chainName, "error", err)
				continue
			}
		}
	}

	// Mark DAOs as inactive if they're not in the current config
	if err := t.daoService.MarkInactiveDAOs(activeDaoCodes); err != nil {
		return fmt.Errorf("failed to mark inactive DAOs: %w", err)
	}

	duration := time.Since(startTime)
	slog.Info("DAO synchronization completed",
		"active_daos", len(activeDaoCodes),
		"duration", duration.String(),
		"timestamp", time.Now().Format(time.RFC3339))
	return nil
}

// processSingleDao processes a single DAO configuration
func (t *DaoSyncTask) processSingleDao(remoteLink GithubConfigLink, daoInfo DaoRegistryConfig, chainName string, activeDaoCodes map[string]bool) error {
	configURL := daoInfo.Config
	// Convert relative URL to absolute if needed
	if !strings.HasPrefix(configURL, "http://") && !strings.HasPrefix(configURL, "https://") {
		configURL = fmt.Sprintf("%s/%s", remoteLink.BaseLink, configURL)
	}

	// Fetch DAO config details
	daoConfig, err := t.fetchDaoConfig(configURL, daoInfo.Code)
	if err != nil {
		return fmt.Errorf("failed to fetch DAO config: %w", err)
	}

	// Skip if essential fields are missing
	if daoInfo.Code == "" || daoConfig.Config.Name == "" {
		slog.Warn("DAO config missing essential fields", "config_url", daoInfo.Config)
		return nil
	}

	activeDaoCodes[daoInfo.Code] = true

	// Create DAO model using service
	dao := t.daoService.CreateDaoFromConfig(daoInfo.Code, daoConfig.Config, daoInfo.Tags, configURL)

	if err := t.daoService.UpsertDao(dao); err != nil {
		return fmt.Errorf("failed to upsert DAO: %w", err)
	}

	// Save DAO config to DgvDaoConfig table
	if err := t.daoService.UpsertDaoConfig(daoInfo.Code, daoConfig.Raw); err != nil {
		return fmt.Errorf("failed to upsert DAO config: %w", err)
	}

	slog.Debug("Successfully synced DAO", "dao", daoInfo.Code, "chain", chainName)
	return nil
}

// fetchRegistryConfig fetches and parses the main registry configuration
func (t *DaoSyncTask) fetchRegistryConfig() (DaoRegistryConfigResult, error) {
	configURLs := t.buildConfigURLs()

	for i, configURL := range configURLs {
		slog.Debug("Attempting to fetch registry config", "url", configURL, "attempt", i+1)

		var config map[string][]DaoRegistryConfig
		_, err := t.fetchAndParseYAML(configURL.ConfigLink, &config)
		if err != nil {
			if i == len(configURLs)-1 {
				return DaoRegistryConfigResult{}, fmt.Errorf("failed to fetch config from all URLs: %w", err)
			}
			slog.Warn("Failed to fetch config, trying next URL", "url", configURL, "error", err)
			continue
		}

		slog.Debug("Successfully fetched registry config", "url", configURL, "chains_count", len(config))
		return DaoRegistryConfigResult{
			RemoteLink: configURL,
			Result:     config,
		}, nil
	}

	return DaoRegistryConfigResult{}, fmt.Errorf("failed to fetch config from any URL")
}

// fetchAndParseYAML fetches content from URL and parses it as YAML
func (t *DaoSyncTask) fetchAndParseYAML(url string, target interface{}) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			slog.Error("Failed to fetch content", "url", url, "status_code", resp.StatusCode, "status", resp.Status, "body_read_error", readErr)
		} else {
			slog.Error("Failed to fetch content", "url", url, "status_code", resp.StatusCode, "status", resp.Status, "response_body", string(body))
		}
		return "", fmt.Errorf("unexpected status %d (%s) from %s", resp.StatusCode, resp.Status, url)
	}

	// Read the raw content first
	rawContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	// Parse the YAML content
	if err := yaml.Unmarshal(rawContent, target); err != nil {
		return "", fmt.Errorf("failed to parse YAML from %s: %w", url, err)
	}

	return string(rawContent), nil
}

// buildConfigURLs constructs the list of config URLs to try based on configuration
func (t *DaoSyncTask) buildConfigURLs() []GithubConfigLink {
	mode := config.GetString("REGISTRY_CONFIG_MODE")
	refs := config.GetString("REGISTRY_CONFIG_REFS")

	switch {
	case mode != "" && refs != "":
		configURL := t.buildConfigURL(mode, refs)
		slog.Debug("Using explicit config mode", "mode", mode, "refs", refs, "url", configURL)
		return []GithubConfigLink{configURL}

	case refs != "":
		urls := []GithubConfigLink{
			t.buildConfigURL("tag", refs),
			t.buildConfigURL("branch", refs),
		}
		slog.Debug("Using explicit refs with fallback", "refs", refs, "urls", urls)
		return urls

	default:
		return t.buildDefaultConfigURLs()
	}
}

// buildDefaultConfigURLs builds URLs when no explicit config is provided
func (t *DaoSyncTask) buildDefaultConfigURLs() []GithubConfigLink {
	latestTag, err := t.getLatestTag()
	if err != nil || latestTag == "" {
		if err != nil {
			slog.Warn("Failed to get latest tag, will use main branch", "error", err)
		} else {
			slog.Info("No tags found, using main branch")
		}
		return []GithubConfigLink{t.buildConfigURL("branch", "main")}
	}

	slog.Info("Using latest tag", "tag", latestTag)
	return []GithubConfigLink{t.buildConfigURL("tag", latestTag)}
}

func (t *DaoSyncTask) baseRawGithubLink(mode, refs string) string {
	baseURL := "https://raw.githubusercontent.com/ringecosystem/degov-registry"
	if mode == "tag" {
		return fmt.Sprintf("%s/tags/%s", baseURL, refs)
	}
	return fmt.Sprintf("%s/heads/%s", baseURL, refs)
}

// buildConfigURL constructs the config URL based on mode and refs
func (t *DaoSyncTask) buildConfigURL(mode, refs string) GithubConfigLink {
	baseURL := t.baseRawGithubLink(mode, refs)
	return GithubConfigLink{
		BaseLink:   baseURL,
		ConfigLink: fmt.Sprintf("%s/config.yml", baseURL),
	}
}

// getLatestTag fetches the latest tag from GitHub API
func (t *DaoSyncTask) getLatestTag() (string, error) {
	apiURL := "https://api.github.com/repos/ringecosystem/degov-registry/tags"

	slog.Debug("Fetching tags from GitHub API", "url", apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch tags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var tags []GitHubTag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	if len(tags) == 0 {
		return "", nil
	}

	// Return the latest (first) tag
	return tags[0].Name, nil
}

// fetchDaoConfig fetches and parses individual DAO configuration
func (t *DaoSyncTask) fetchDaoConfig(configURL string, daoCode string) (DaoConfigResult, error) {
	slog.Debug("Fetching DAO config", "url", configURL)

	var config types.DaoConfig
	rawContent, err := t.fetchAndParseYAML(configURL, &config)
	if err != nil {
		return DaoConfigResult{}, err
	}

	slog.Debug("Successfully fetched DAO config", "url", configURL, "dao_code", daoCode)
	return DaoConfigResult{
		Raw:    rawContent,
		Config: &config,
	}, nil
}
