package yoink

import (
	"github.com/mrmarble/yoink/pkg/prowlarr"
)

// ProwlarrManager wraps the Prowlarr client to implement IndexerManager interface
type ProwlarrManager struct {
	client *prowlarr.Client
}

// NewProwlarrManager creates a new ProwlarrManager
func NewProwlarrManager(host, apiKey string) *ProwlarrManager {
	return &ProwlarrManager{
		client: prowlarr.NewClient(host, apiKey),
	}
}

// GetAPIVersion returns the Prowlarr API version
func (p *ProwlarrManager) GetAPIVersion() (string, error) {
	return p.client.GetAPIVersion()
}

// Search performs a search using Prowlarr
func (p *ProwlarrManager) Search(config SearchConfig) ([]SearchResult, error) {
	// Convert generic indexer IDs to int slice for Prowlarr
	var indexerIDs []int
	for _, id := range config.Indexers {
		if intID, ok := id.(int); ok {
			indexerIDs = append(indexerIDs, intID)
		}
	}

	prowlarrConfig := &prowlarr.SearchConfig{
		Indexers:  indexerIDs,
		FreeLeech: config.FreeLeech,
		Limit:     config.Limit,
	}

	results, err := p.client.Search(prowlarrConfig)
	if err != nil {
		return nil, err
	}

	// Convert prowlarr results to SearchResult interface
	searchResults := make([]SearchResult, len(results))
	for i, result := range results {
		searchResults[i] = &ProwlarrSearchResult{result: result}
	}

	return searchResults, nil
}

// GetIndexers returns available Prowlarr indexers
func (p *ProwlarrManager) GetIndexers() ([]IndexerInfo, error) {
	indexers, err := p.client.GetIndexers()
	if err != nil {
		return nil, err
	}

	// Convert prowlarr indexers to IndexerInfo interface, filtering for torrent protocol
	var indexerInfos []IndexerInfo
	for _, indexer := range indexers {
		if indexer.Protocol == "torrent" {
			indexerInfos = append(indexerInfos, &ProwlarrIndexerInfo{indexer: indexer})
		}
	}

	return indexerInfos, nil
}

// ProwlarrSearchResult wraps prowlarr.SearchResult to implement SearchResult interface
type ProwlarrSearchResult struct {
	result prowlarr.SearchResult
}

func (p *ProwlarrSearchResult) GetTitle() string          { return p.result.Title }
func (p *ProwlarrSearchResult) GetSize() uint64           { return uint64(p.result.Size) }
func (p *ProwlarrSearchResult) GetSeeders() int           { return p.result.Seeders }
func (p *ProwlarrSearchResult) GetLeechers() int          { return p.result.Leechers }
func (p *ProwlarrSearchResult) GetDownloadURL() string    { return p.result.DownloadURL }
func (p *ProwlarrSearchResult) GetIndexerID() interface{} { return p.result.IndexerID }
func (p *ProwlarrSearchResult) IsFreeleech() bool         { return p.result.IsFreeleech() }

// ProwlarrIndexerInfo wraps prowlarr.Indexer to implement IndexerInfo interface
type ProwlarrIndexerInfo struct {
	indexer prowlarr.Indexer
}

func (p *ProwlarrIndexerInfo) GetID() interface{} { return p.indexer.ID }
func (p *ProwlarrIndexerInfo) GetName() string    { return p.indexer.Name }
func (p *ProwlarrIndexerInfo) IsEnabled() bool    { return p.indexer.Enable }
