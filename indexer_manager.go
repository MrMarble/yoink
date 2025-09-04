package yoink

// IndexerManager provides a common interface for both Prowlarr and Jackett clients
type IndexerManager interface {
	// GetAPIVersion returns the API version string
	GetAPIVersion() (string, error)

	// Search performs a search for torrents and returns results
	Search(config SearchConfig) ([]SearchResult, error)

	// GetIndexers returns the list of available indexers
	GetIndexers() ([]IndexerInfo, error)
}

// SearchConfig represents common search configuration for both indexer types
type SearchConfig struct {
	Indexers  []interface{} // Can be []int for Prowlarr or []string for Jackett
	FreeLeech bool          // Only return freeleech torrents
	Query     string        // Search query (optional)
	Limit     int           // Limit the number of results
}

// SearchResult represents a common search result interface
type SearchResult interface {
	GetTitle() string
	GetSize() uint64
	GetSeeders() int
	GetLeechers() int
	GetDownloadURL() string
	GetIndexerID() interface{} // Can be int for Prowlarr or string for Jackett
	IsFreeleech() bool
}

// IndexerInfo represents common indexer information
type IndexerInfo interface {
	GetID() interface{} // Can be int for Prowlarr or string for Jackett
	GetName() string
	IsEnabled() bool
}
