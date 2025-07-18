package types

// DaoConfig represents the structure of individual DAO config files
type DaoConfig struct {
	Name                  string `yaml:"name"`
	Code                  string `yaml:"code"`
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
	} `yaml:"indexer"`
	Contracts struct {
		Governor      string `yaml:"governor"`
		GovernorToken struct {
			Address  string `yaml:"address"`
			Standard string `yaml:"standard"`
		} `yaml:"governorToken"`
		TimeLock string `yaml:"timeLock"`
	} `yaml:"contracts"`
}
