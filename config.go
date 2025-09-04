package yoink

import "fmt"

//go:generate go run config_sample.go

// Config holds the configuration for yoink
type Config struct {
	// Max space to use for downloads. If 0, no limit is applied
	TotalFreeleechSize string `yaml:"total_freeleech_size" env:"TOTAL_FREELEECH_SIZE" env-default:"200GB" env-description:"Max space to use for downloads. If 0, no limit is applied" comment:"Max space to use for downloads. If 0, no limit is applied"`
	// Category to use for downloads.
	Category string `yaml:"category" env:"CATEGORY" env-default:"FreeLeech" env-description:"Category to use for downloads." comment:"Category to use for downloads."`
	// Whether to pause torrents after adding them to qBittorrent
	Paused bool `yaml:"paused" env:"PAUSED" env-default:"true" env-description:"Whether to pause torrents after adding them to qBittorrent" comment:"Whether to pause torrents after adding them to qBittorrentf"`
	// Indexer type: prowlarr or jackett
	IndexerType string `yaml:"indexer_type" env:"INDEXER_TYPE" env-default:"prowlarr" env-description:"Type of indexer to use: prowlarr or jackett" comment:"Type of indexer to use: prowlarr or jackett"`
	// Connection details for qBittorrent
	QbitTorrent struct {
		Host string `yaml:"host" env:"QBIT_HOST" env-description:"Connection details for qBittorrent"`
		User string `yaml:"username" env:"QBIT_USER" env-description:"Connection details for qBittorrent"`
		Pass string `yaml:"password" env:"QBIT_PASS" env-description:"Connection details for qBittorrent"`
	} `yaml:"qbittorrent" comment:"Connection details for qBittorrent"`
	// Connection details for Prowlarr
	Prowlarr struct {
		Host   string `yaml:"host" env:"PROWLARR_HOST" env-description:"Connection details for Prowlarr"`
		APIKey string `yaml:"api_key" env:"PROWLARR_API_KEY" env-description:"Connection details for Prowlarr"`
	} `yaml:"prowlarr" comment:"Connection details for Prowlarr"`
	// Connection details for Jackett
	Jackett struct {
		Host   string `yaml:"host" env:"JACKETT_HOST" env-description:"Connection details for Jackett"`
		APIKey string `yaml:"api_key" env:"JACKETT_API_KEY" env-description:"Connection details for Jackett"`
	} `yaml:"jackett" comment:"Connection details for Jackett"`

	// List of indexers to use. Filters out any indexers not in this list
	Indexers []Indexer `yaml:"indexers" comment:"List of indexers to use. Filters out any indexers not in this list"`
}

type Indexer struct {
	// ID of the indexer (integer for Prowlarr, string for Jackett)
	ID   interface{} `yaml:"id" env:"INDEXER_ID" env-description:"ID of the indexer (integer for Prowlarr, string for Jackett)" comment:"ID of the indexer (integer for Prowlarr, string for Jackett)"`
	// Maximum number of seeders to allow. 0 = no limit
	MaxSeeders int `yaml:"max_seeders" env:"INDEXER_MAX_SEEDERS" env-default:"0" env-description:"Maximum number of seeders to allow. 0 = no limit" comment:"Maximum number of seeders to allow. 0 = no limit"`
	// Maximum file size to allow. 0 = no limit
	MaxSize string `yaml:"max_size" env:"INDEXER_MAX_SIZE" env-default:"0" env-description:"Maximum file size to allow. 0 = no limit" comment:"Maximum file size to allow. 0 = no limit"`
	// Minimum number of leechers to allow. 0 = no limit
	MinLeechers int `yaml:"min_leechers" env:"INDEXER_MIN_LEECHERS" env-default:"0" env-description:"Minimum number of leechers to allow. 0 = no limit" comment:"Minimum number of leechers to allow. 0 = no limit"`
}

// GetProwlarrID returns the indexer ID as an integer for Prowlarr compatibility
func (i *Indexer) GetProwlarrID() int {
	switch v := i.ID.(type) {
	case int:
		return v
	case float64: // JSON numbers are parsed as float64
		return int(v)
	default:
		return 0
	}
}

// GetJackettID returns the indexer ID as a string for Jackett compatibility
func (i *Indexer) GetJackettID() string {
	switch v := i.ID.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		return ""
	}
}
