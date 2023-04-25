package yoink

//go:generate go run config_sample.go

// Config holds the configuration for yoink
type Config struct {
	// Max space to use for downloads. If 0, no limit is applied
	TotalFreeleechSize string `yaml:"total_freeleech_size" env:"TOTAL_FREELEECH_SIZE" env-default:"200GB" env-description:"Max space to use for downloads. If 0, no limit is applied" comment:"Max space to use for downloads. If 0, no limit is applied"`
	// Category to use for downloads.
	Category string `yaml:"category" env:"CATEGORY" env-default:"FreeLeech" env-description:"Category to use for downloads." comment:"Category to use for downloads."`
	// Whether to pause torrents after adding them to qBittorrent
	Paused bool `yaml:"paused" env:"PAUSED" env-default:"true" env-description:"Whether to pause torrents after adding them to qBittorrent" comment:"Whether to pause torrents after adding them to qBittorrentf"`
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

	// List of indexers to use. Filters out any indexers not in this list
	Indexers []Indexer `yaml:"indexers" comment:"List of indexers to use. Filters out any indexers not in this list"`
}

type Indexer struct {
	// ID of the indexer in Prowlarr
	ID int `yaml:"id" env:"INDEXER_ID" env-description:"ID of the indexer in Prowlarr" comment:"ID of the indexer in Prowlarr"`
	// Maximum number of seeders to allow. 0 = no limit
	MaxSeeders int `yaml:"max_seeders" env:"INDEXER_MAX_SEEDERS" env-default:"0" env-description:"Maximum number of seeders to allow. 0 = no limit" comment:"Maximum number of seeders to allow. 0 = no limit"`
	// Maximum file size to allow. 0 = no limit
	MaxSize string `yaml:"max_size" env:"INDEXER_MAX_SIZE" env-default:"0" env-description:"Maximum file size to allow. 0 = no limit" comment:"Maximum file size to allow. 0 = no limit"`
}
