package yoink

import "fmt"

// ManagerRegistry holds the available indexer managers
var ManagerRegistry = map[string]func(host, apiKey string) IndexerManager{
	"prowlarr": func(host, apiKey string) IndexerManager {
		return NewProwlarrManager(host, apiKey)
	},
	"jackett": func(host, apiKey string) IndexerManager {
		return NewJackettManager(host, apiKey)
	},
}

// GetIndexerManager creates the appropriate indexer manager based on configuration
func GetIndexerManager(cfg *Config) (IndexerManager, error) {
	createManager, exists := ManagerRegistry[cfg.IndexerType]
	if !exists {
		return nil, fmt.Errorf("unsupported indexer type: %s", cfg.IndexerType)
	}

	switch cfg.IndexerType {
	case "prowlarr":
		return createManager(cfg.Prowlarr.Host, cfg.Prowlarr.APIKey), nil
	case "jackett":
		return createManager(cfg.Jackett.Host, cfg.Jackett.APIKey), nil
	default:
		return nil, fmt.Errorf("unsupported indexer type: %s", cfg.IndexerType)
	}
}
