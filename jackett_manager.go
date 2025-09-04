package yoink

import (
	"github.com/mrmarble/yoink/pkg/jackett"
)

// JackettManager wraps the Jackett client to implement IndexerManager interface
type JackettManager struct {
	client *jackett.Client
}

// NewJackettManager creates a new JackettManager
func NewJackettManager(host, apiKey string) *JackettManager {
	return &JackettManager{
		client: jackett.NewClient(host, apiKey),
	}
}

// GetAPIVersion returns the Jackett API version
func (j *JackettManager) GetAPIVersion() (string, error) {
	return j.client.GetAPIVersion()
}

// Search performs a search using Jackett
func (j *JackettManager) Search(config SearchConfig) ([]SearchResult, error) {
	// Convert generic indexer IDs to string slice for Jackett
	var indexerIDs []string
	for _, id := range config.Indexers {
		if stringID, ok := id.(string); ok {
			indexerIDs = append(indexerIDs, stringID)
		}
	}

	jackettConfig := &jackett.SearchConfig{
		Indexers:  indexerIDs,
		FreeLeech: config.FreeLeech,
		Query:     config.Query,
		Limit:     config.Limit,
	}

	results, err := j.client.Search(jackettConfig)
	if err != nil {
		return nil, err
	}

	// Convert jackett results to SearchResult interface
	searchResults := make([]SearchResult, len(results))
	for i, result := range results {
		searchResults[i] = &JackettSearchResult{result: result}
	}

	return searchResults, nil
}

// GetIndexers returns available Jackett indexers
func (j *JackettManager) GetIndexers() ([]IndexerInfo, error) {
	indexers, err := j.client.GetIndexers()
	if err != nil {
		return nil, err
	}

	// Convert jackett indexers to IndexerInfo interface
	indexerInfos := make([]IndexerInfo, len(indexers))
	for i, indexer := range indexers {
		indexerInfos[i] = &JackettIndexerInfo{indexer: indexer}
	}

	return indexerInfos, nil
}

// JackettSearchResult wraps jackett.SearchResult to implement SearchResult interface
type JackettSearchResult struct {
	result jackett.SearchResult
}

func (j *JackettSearchResult) GetTitle() string          { return j.result.Title }
func (j *JackettSearchResult) GetSize() uint64           { return j.result.Size }
func (j *JackettSearchResult) GetSeeders() int           { return j.result.Seeders }
func (j *JackettSearchResult) GetLeechers() int          { return j.result.Peers }
func (j *JackettSearchResult) GetDownloadURL() string    { return j.result.DownloadURL }
func (j *JackettSearchResult) GetIndexerID() interface{} { return j.result.IndexerID }
func (j *JackettSearchResult) IsFreeleech() bool         { return j.result.IsFreeleech() }

// JackettIndexerInfo wraps jackett.Indexer to implement IndexerInfo interface
type JackettIndexerInfo struct {
	indexer jackett.Indexer
}

func (j *JackettIndexerInfo) GetID() interface{} { return j.indexer.ID }
func (j *JackettIndexerInfo) GetName() string    { return j.indexer.Title }
func (j *JackettIndexerInfo) IsEnabled() bool {
	// Jackett indexers don't have an explicit enabled field in the current struct
	// Assume all returned indexers are enabled
	return true
}
