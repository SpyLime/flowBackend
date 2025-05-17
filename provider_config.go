package main

// ProviderConfig represents the configuration for an OAuth provider
type ProviderConfig struct {
	Enabled      bool   `yaml:"enabled"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

// ProvidersConfig represents the configuration for all SSO providers
type ProvidersConfig struct {
	Google    ProviderConfig `yaml:"google"`
	Microsoft ProviderConfig `yaml:"microsoft"`
	Facebook  ProviderConfig `yaml:"facebook"`
	Discord   ProviderConfig `yaml:"discord"`
	Twitter   ProviderConfig `yaml:"twitter"`
}
