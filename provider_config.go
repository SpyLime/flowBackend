package main

// ProviderConfig represents the configuration for an OAuth provider
type ProviderConfig struct {
	Enabled      bool   `yaml:"enabled"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

// AppleProviderConfig represents the configuration for Apple Sign In
type AppleProviderConfig struct {
	Enabled      bool   `yaml:"enabled"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	TeamID       string `yaml:"team_id"`
	KeyID        string `yaml:"key_id"`
}

// ProvidersConfig represents the configuration for all SSO providers
type ProvidersConfig struct {
	Google    ProviderConfig      `yaml:"google"`
	Microsoft ProviderConfig      `yaml:"microsoft"`
	Facebook  ProviderConfig      `yaml:"facebook"`
	Apple     AppleProviderConfig `yaml:"apple"`
	Discord   ProviderConfig      `yaml:"discord"`
	Twitter   ProviderConfig      `yaml:"twitter"`
}
