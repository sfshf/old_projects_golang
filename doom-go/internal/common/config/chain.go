package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

type ChainConfig struct {
	Chains []*Chain
}

type Chain struct {
	Name    string
	Address string
}

var (
	DefaultChainConfig *ChainConfig
)

func LoadDefaultChainConfig() error {
	configPath := os.Getenv("CRYPTOCURRENCY_CHAIN_CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/application/cryptocurrency/chains.toml"
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var config ChainConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return err
	}
	DefaultChainConfig = &config
	return nil
}

var (
	ErrNilDefaultPriceConfig = errors.New("nil default price config")
	ErrNilDefaultChainConfig = errors.New("default chain config is nil")
	ErrInvalidChainName      = errors.New("invalid chain name")
)

func ChainByName(name string) (*Chain, error) {
	for _, chain := range DefaultChainConfig.Chains {
		if name == chain.Name {
			return chain, nil
		}
	}
	return nil, ErrInvalidChainName
}
