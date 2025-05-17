package main

import (
	"os"

	"github.com/go-pkgz/lgr"
	"gopkg.in/yaml.v3"
)

// ServerConfig represents the server configuration
type ServerConfig struct {
	AdminPassword string          `yaml:"adminpassword"`
	SecureCookies bool            `yaml:"securecookies"`
	EnableXSRF    bool            `yaml:"enablexsrf"`
	SecretKey     string          `yaml:"secretkey"`
	ServerPort    int             `yaml:"serverport"`
	SeedPassword  string          `yaml:"seedpassword"`
	EmailSMTP     string          `yaml:"emailsmtp"`
	PasswordSMTP  string          `yaml:"passwordsmtp"`
	Server        bool            `yaml:"server"` // server is true if the server is running on the server
	Production    bool            `yaml:"production"`
	Providers     ProvidersConfig `yaml:"providers"`
}

// LoadConfig loads the server configuration from the YAML file
func LoadConfig() ServerConfig {
	config := ServerConfig{
		SecretKey:     "secret",
		AdminPassword: "admin",
		ServerPort:    8080,
		SeedPassword:  "123qwe",
		EmailSMTP:     "qq@qq.com",
		PasswordSMTP:  "123qwe",
		Production:    false,
	}

	yamlFile, err := os.ReadFile("./flcfg.yml")
	if err != nil {
		lgr.Printf("ERROR cannot load config %v", err)
		return config
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		lgr.Printf("ERROR cannot decode config %v", err)
	}

	return config
}
