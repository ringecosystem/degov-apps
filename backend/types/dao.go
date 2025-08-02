package types

type RefreshDaoAndConfigInput struct {
	Code   string    `json:"code"`
	Tags   []string  `json:"tags"`
	Config DaoConfig `json:"config"`
	Raw    string    `json:"raw"`
}

// DaoConfig represents the structure of individual DAO config files
type DaoConfig struct {
	Name                  string `yaml:"name"`
	Logo                  string `yaml:"logo"`
	SiteURL               string `yaml:"siteUrl"`
	OffChainDiscussionURL string `yaml:"offChainDiscussionUrl"`
	Description           string `yaml:"description"`
	Chain                 struct {
		ID          int      `yaml:"id"`
		Name        string   `yaml:"name"`
		Logo        string   `yaml:"logo"`
		RPCs        []string `yaml:"rpcs"`
		Explorers   []string `yaml:"explorers"`
		NativeToken struct {
			Symbol   string `yaml:"symbol"`
			PriceID  string `yaml:"priceId"`
			Decimals int    `yaml:"decimals"`
			Logo     string `yaml:"logo"`
		} `yaml:"nativeToken"`
	} `yaml:"chain"`
	AIAgent struct {
		Endpoint string `yaml:"endpoint"`
	} `yaml:"aiAgent"`
	Links struct {
		Coingecko string `yaml:"coingecko"`
		Website   string `yaml:"website"`
		Twitter   string `yaml:"twitter"`
		GitHub    string `yaml:"github"`
	} `yaml:"links"`
	Wallet struct {
		WalletConnectProjectID string `yaml:"walletConnectProjectId"`
	} `yaml:"wallet"`
	Indexer struct {
		Endpoint   string `yaml:"endpoint"`
		StartBlock int    `yaml:"startBlock"`
		RPC        string `yaml:"rpc"`
		Gateway    string `yaml:"gateway"`
	} `yaml:"indexer"`
	Contracts struct {
		Governor      string `yaml:"governor"`
		GovernorToken struct {
			Address  string `yaml:"address"`
			Standard string `yaml:"standard"`
		} `yaml:"governorToken"`
		TimeLock string `yaml:"timeLock"`
	} `yaml:"contracts"`
	Safes []struct {
		Name    string `yaml:"name"`
		ChainID int    `yaml:"chainId"`
		Link    string `yaml:"link"`
	} `yaml:"safes"`
}
