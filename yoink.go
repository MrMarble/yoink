// Package yoink provides utilities to manage freeleech
// downloads automatically
package yoink

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/dustin/go-humanize"
	"github.com/mrmarble/yoink/pkg/qbittorrent"
)

// GetTorrents searches for freeleech torrents using the configured indexer and filters them based on the indexer configuration
func GetTorrents(cfg *Config, indexers []Indexer) ([]SearchResult, error) {
	manager, err := GetIndexerManager(cfg)
	if err != nil {
		return nil, err
	}

	// Convert indexers to the appropriate ID format
	var indexerIDs []interface{}
	for _, indexer := range indexers {
		if cfg.IndexerType == "prowlarr" {
			indexerIDs = append(indexerIDs, indexer.GetProwlarrID())
		} else {
			indexerIDs = append(indexerIDs, indexer.GetJackettID())
		}
	}

	// Search for torrents
	searchConfig := SearchConfig{
		Indexers:  indexerIDs,
		FreeLeech: true,
	}

	results, err := manager.Search(searchConfig)
	if err != nil {
		return nil, err
	}

	// Filter results based on indexer configuration
	var filteredResults []SearchResult
	for _, result := range results {
		if isStale(result) {
			continue
		}

		// Check if result matches any indexer criteria
		validForAnyIndexer := false
		for _, indexer := range indexers {
			if isValidForIndexer(result, indexer, cfg.IndexerType) {
				validForAnyIndexer = true
				break
			}
		}

		if validForAnyIndexer {
			filteredResults = append(filteredResults, result)
		}
	}

	return filteredResults, nil
}

// isValidForIndexer checks if a result matches an indexer's criteria
func isValidForIndexer(result SearchResult, indexer Indexer, indexerType string) bool {
	// For Prowlarr, we can match by exact indexer ID
	if indexerType == "prowlarr" {
		if result.GetIndexerID() != indexer.GetProwlarrID() {
			return false
		}
	}
	// For Jackett, indexer matching is less reliable, so we apply filters to all results

	maxSize, _ := humanize.ParseBytes(indexer.MaxSize)
	validSeeders := indexer.MaxSeeders == 0 || result.GetSeeders() <= indexer.MaxSeeders
	validSize := maxSize == 0 || result.GetSize() <= maxSize
	validLeechers := indexer.MinLeechers == 0 || result.GetLeechers() >= indexer.MinLeechers

	return validSeeders && validSize && validLeechers
}

// isStale checks if the torrent is stale (no seeders)
func isStale(result SearchResult) bool {
	return result.GetSeeders() == 0
}

// DownloadTorrent downloads the torrents to qBittorrent
// Filters out any torrents that are already downloading
//
// 1. Connect to qBittorrent and get the list of torrents
//
// 3. Download to memory and check if the torrent is already downloading
func DownloadTorrent(result SearchResult) (*bytes.Buffer, error) {
	buf, err := downloadFile(result.GetDownloadURL())
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func downloadFile(url string) (*bytes.Buffer, error) {
	// Get the data
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	return buf, err
}

// GetDownloadingTorrents retrieves the list of torrents currently downloading in qBittorrent
func GetDownloadingTorrents(config *Config, qClient *qbittorrent.Client) ([]qbittorrent.Torrent, error) {
	qTorrents, err := qClient.GetTorrents()
	if err != nil {
		return nil, err
	}

	var downloadingTorrents []qbittorrent.Torrent
	for _, torrent := range qTorrents {
		if torrent.Category == config.Category {
			downloadingTorrents = append(downloadingTorrents, torrent)
		}
	}

	return downloadingTorrents, nil
}
