package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ringecosystem/degov-apps/database"
	dbmodels "github.com/ringecosystem/degov-apps/database/models"
	gqlmodels "github.com/ringecosystem/degov-apps/graph/models"
	"github.com/ringecosystem/degov-apps/internal/utils"
	"gorm.io/gorm"
)

// Defines the structure for the response from the Etherscan ABI API.
type etherscanAbiResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

// Defines the structure for the result from the Etherscan Source Code API.
type etherscanSourceCodeResult struct {
	Proxy          string `json:"Proxy"`
	Implementation string `json:"Implementation"`
}

// Defines the structure for the response from the Etherscan Source Code API.
type etherscanSourceCodeResponse struct {
	Status  string                      `json:"status"`
	Message string                      `json:"message"`
	Result  []etherscanSourceCodeResult `json:"result"`
}

// Defines the structure for the response from the Swiss-knife API.
type swissKnifeResponse struct {
	Status string `json:"status"`
	Data   struct {
		IsProxy               bool   `json:"isProxy"`
		ImplementationAddress string `json:"implementationAddress"`
		Abi                   string `json:"abi"`
	} `json:"data"`
}

type EvmChainService struct {
	db         *gorm.DB
	httpClient *http.Client
}

func NewEvmChainService() *EvmChainService {
	return &EvmChainService{
		db: database.GetDB(),
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // Set a 15-second timeout.
		},
	}
}

// GetAbi is the main entry point for fetching contract ABI details.
// It returns a slice of results: one for an implementation, or two (proxy + implementation) for a proxy.
func (s *EvmChainService) GetAbi(input gqlmodels.EvmAbiInput) ([]*gqlmodels.EvmAbiOutput, error) {
	initialAddress := strings.ToLower(input.Contract)
	chainId := int(input.Chain)

	// Use a map to detect circular dependencies in proxy contracts.
	visited := make(map[string]bool)

	// 1. Get information for the initial contract address.
	initialContractInfo, err := s.getContractInfo(chainId, initialAddress)
	if err != nil {
		return nil, fmt.Errorf("could not resolve initial contract %s on chain %d: %w", initialAddress, chainId, err)
	}

	// 2. Check if the initial contract is a proxy or an implementation.
	if initialContractInfo.Type == dbmodels.ContractsAbiTypeImplementation {
		// It's a simple implementation contract.
		output := &gqlmodels.EvmAbiOutput{
			Address: initialContractInfo.Address,
			Abi:     initialContractInfo.Abi,
			Type:    gqlmodels.AbiTypeImplementation,
		}
		return []*gqlmodels.EvmAbiOutput{output}, nil
	}

	if initialContractInfo.Type == dbmodels.ContractsAbiTypeProxy {
		// It's a proxy contract. We need to find the final implementation.
		visited[initialAddress] = true // Mark initial address as visited.

		finalImplementation, err := s.findFinalImplementation(chainId, initialContractInfo.Implementation, visited)
		if err != nil {
			return nil, fmt.Errorf("could not find final implementation for proxy %s: %w", initialAddress, err)
		}

		proxyOutput := &gqlmodels.EvmAbiOutput{
			Address: initialAddress,
			Abi:     "", // Proxy itself has no ABI.
			Type:    gqlmodels.AbiTypeProxy,
		}
		implementationOutput := &gqlmodels.EvmAbiOutput{
			Address: finalImplementation.Address,
			Abi:     finalImplementation.Abi,
			Type:    gqlmodels.AbiTypeImplementation,
		}

		return []*gqlmodels.EvmAbiOutput{proxyOutput, implementationOutput}, nil
	}

	return nil, fmt.Errorf("resolved contract %s has an unknown type: %s", initialAddress, initialContractInfo.Type)
}

// findFinalImplementation recursively traverses the proxy chain to find the ultimate implementation contract.
func (s *EvmChainService) findFinalImplementation(chainId int, address string, visited map[string]bool) (*dbmodels.ContractsAbi, error) {
	// Prevent circular proxy references.
	if visited[address] {
		return nil, fmt.Errorf("circular proxy dependency detected for address %s", address)
	}
	visited[address] = true

	// Get info for the current address in the chain.
	contractInfo, err := s.getContractInfo(chainId, address)
	if err != nil {
		return nil, err
	}

	// Check the type to decide whether to continue recursion.
	if contractInfo.Type == dbmodels.ContractsAbiTypeProxy && contractInfo.Implementation != "" {
		slog.Info("Following proxy chain...", "from", address, "to", contractInfo.Implementation)
		return s.findFinalImplementation(chainId, contractInfo.Implementation, visited)
	}

	if contractInfo.Type == dbmodels.ContractsAbiTypeImplementation {
		// Found the end of the chain.
		return contractInfo, nil
	}

	return nil, fmt.Errorf("contract %s is not a valid implementation or proxy", address)
}

// getContractInfo fetches information for a single contract address from various sources.
// It checks DB -> Explorer -> Swiss-knife and caches the result. It is not recursive.
func (s *EvmChainService) getContractInfo(chainId int, address string) (*dbmodels.ContractsAbi, error) {
	// 1. Try to read from the database.
	dbContract, err := s.getAbiFromDB(chainId, address)
	if err != nil {
		slog.Error("Error querying database", "chainId", chainId, "address", address, "error", err)
		return nil, err
	}
	if dbContract != nil {
		return dbContract, nil
	}

	// 2. If not in DB, try fetching from an Explorer.
	slog.Info("Contract not in DB, trying Explorer...", "chainId", chainId, "address", address)
	explorerContract, err := s.getAbiFromExplorer(chainId, address)
	if err != nil {
		slog.Warn("Failed to get info from explorer", "chainId", chainId, "address", address, "error", err)
	}
	if explorerContract != nil {
		if err := s.saveAbiToDB(explorerContract); err != nil {
			slog.Error("Failed to save explorer data to DB", "error", err)
		}
		return explorerContract, nil
	}

	// 3. If not in Explorer, try fetching from Swiss-knife.
	slog.Info("Contract not in Explorer, trying Swiss-knife...", "chainId", chainId, "address", address)
	swissKnifeContract, err := s.getAbiFromSwissKnife(chainId, address)
	if err != nil {
		slog.Warn("Failed to get info from Swiss-knife", "chainId", chainId, "address", address, "error", err)
	}
	if swissKnifeContract != nil {
		if err := s.saveAbiToDB(swissKnifeContract); err != nil {
			slog.Error("Failed to save Swiss-knife data to DB", "error", err)
		}
		return swissKnifeContract, nil
	}

	// 4. Not found in any data source.
	return nil, fmt.Errorf("contract info not found for address %s on chain %d", address, chainId)
}

// getAbiFromDB retrieves contract information from the database.
func (s *EvmChainService) getAbiFromDB(chainId int, address string) (*dbmodels.ContractsAbi, error) {
	var contract dbmodels.ContractsAbi
	err := s.db.Where("chain_id = ? AND address = ?", chainId, address).First(&contract).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Record not found is not considered an error.
		}
		return nil, err // Other database errors.
	}
	return &contract, nil
}

// getAbiFromExplorer fetches the ABI from a blockchain explorer (like Etherscan).
func (s *EvmChainService) getAbiFromExplorer(chainId int, address string) (*dbmodels.ContractsAbi, error) {
	apiUrl := os.Getenv(fmt.Sprintf("ETHERSCAN_API_URL_%d", chainId))
	apiKey := os.Getenv(fmt.Sprintf("ETHERSCAN_API_KEY_%d", chainId))

	if apiUrl == "" {
		return nil, fmt.Errorf("ETHERSCAN_API_URL_%d is not set in environment variables", chainId)
	}

	// 1. First, call getsourcecode to check if it's a proxy contract.
	sourceCodeUrl := fmt.Sprintf("%s?module=contract&action=getsourcecode&address=%s&apikey=%s", apiUrl, address, apiKey)
	resp, err := s.httpClient.Get(sourceCodeUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to call explorer source code API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var sourceCodeResp etherscanSourceCodeResponse
	if err := json.Unmarshal(body, &sourceCodeResp); err != nil {
		return nil, fmt.Errorf("failed to parse explorer source code response: %w", err)
	}

	if sourceCodeResp.Status == "1" && len(sourceCodeResp.Result) > 0 && sourceCodeResp.Result[0].Proxy == "1" && sourceCodeResp.Result[0].Implementation != "" {
		slog.Info("Explorer identified contract as a proxy", "address", address, "implementation", sourceCodeResp.Result[0].Implementation)
		return &dbmodels.ContractsAbi{
			ChainId:        chainId,
			Address:        address,
			Type:           dbmodels.ContractsAbiTypeProxy,
			Implementation: strings.ToLower(sourceCodeResp.Result[0].Implementation),
		}, nil
	}

	// 2. If it's not a proxy, call getabi to get the ABI.
	abiUrl := fmt.Sprintf("%s?module=contract&action=getabi&address=%s&apikey=%s", apiUrl, address, apiKey)
	resp, err = s.httpClient.Get(abiUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to call explorer ABI API: %w", err)
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	var abiResp etherscanAbiResponse
	if err := json.Unmarshal(body, &abiResp); err != nil {
		return nil, fmt.Errorf("failed to parse explorer ABI response: %w", err)
	}

	if abiResp.Status == "1" && abiResp.Result != "Contract source code not verified" && abiResp.Result != "" {
		slog.Info("Explorer found ABI for contract", "address", address)
		return &dbmodels.ContractsAbi{
			ChainId: chainId,
			Address: address,
			Type:    dbmodels.ContractsAbiTypeImplementation,
			Abi:     abiResp.Result,
		}, nil
	}

	return nil, nil // Not found in Explorer.
}

// getAbiFromSwissKnife fetches the ABI from the Swiss-knife.xyz API.
func (s *EvmChainService) getAbiFromSwissKnife(chainId int, address string) (*dbmodels.ContractsAbi, error) {
	url := fmt.Sprintf("https://swiss-knife.xyz/api/evm/contract/%s?network=%d", address, chainId)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "degov-app/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Swiss-knife API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Swiss-knife API returned non-200 status: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var swissResp swissKnifeResponse
	if err := json.Unmarshal(body, &swissResp); err != nil {
		return nil, fmt.Errorf("failed to parse Swiss-knife response: %w", err)
	}

	if swissResp.Status == "success" {
		if swissResp.Data.IsProxy && swissResp.Data.ImplementationAddress != "" {
			slog.Info("Swiss-knife identified contract as a proxy", "address", address, "implementation", swissResp.Data.ImplementationAddress)
			return &dbmodels.ContractsAbi{
				ChainId:        chainId,
				Address:        address,
				Type:           dbmodels.ContractsAbiTypeProxy,
				Implementation: strings.ToLower(swissResp.Data.ImplementationAddress),
			}, nil
		}
		if swissResp.Data.Abi != "" && swissResp.Data.Abi != "[]" {
			slog.Info("Swiss-knife found ABI for contract", "address", address)
			return &dbmodels.ContractsAbi{
				ChainId: chainId,
				Address: address,
				Type:    dbmodels.ContractsAbiTypeImplementation,
				Abi:     swissResp.Data.Abi,
			}, nil
		}
	}

	return nil, nil // Not found in Swiss-knife.
}

// saveAbiToDB saves or updates contract information without using ON CONFLICT.
func (s *EvmChainService) saveAbiToDB(contract *dbmodels.ContractsAbi) error {
	var existing dbmodels.ContractsAbi
	err := s.db.Where("chain_id = ? AND address = ?", contract.ChainId, contract.Address).First(&existing).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err // A genuine database error.
	}

	// The record already exists, so update it.
	if err == nil {
		existing.Type = contract.Type
		existing.Implementation = contract.Implementation
		// Only overwrite the ABI if the new one is not empty.
		if contract.Abi != "" {
			existing.Abi = contract.Abi
		}
		now := time.Now()
		existing.UTime = &now
		return s.db.Save(&existing).Error
	}

	// The record does not exist, so create it.
	contract.ID = utils.NextIDString()
	contract.CTime = time.Now()
	now := time.Now()
	contract.UTime = &now
	return s.db.Create(contract).Error
}
