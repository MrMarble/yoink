// Package yoink provides utilities to manage freeleech
// downloads automatically
package yoink

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/dustin/go-humanize"
	"github.com/mrmarble/yoink/pkg/jackett"
	"github.com/mrmarble/yoink/pkg/prowlarr"
	"github.com/mrmarble/yoink/pkg/qbittorrent"
)

// GetTorrents searches for freeleech torrents using the configured indexer and filters them based on the indexer configuration
func GetTorrents(cfg *Config, indexers []Indexer) ([]CommonSearchResult, error) {
	if cfg.IndexerType == "jackett" {
		return getTorrentsFromJackett(cfg, indexers)
	}
	return getTorrentsFromProwlarr(cfg, indexers)
}

// getTorrentsFromProwlarr searches for freeleech torrents in Prowlarr
func getTorrentsFromProwlarr(cfg *Config, indexers []Indexer) ([]CommonSearchResult, error) {
	pClient := prowlarr.NewClient(cfg.Prowlarr.Host, cfg.Prowlarr.APIKey)

	indexerIDs := make([]int, len(indexers))
	for i, indexer := range indexers {
		indexerIDs[i] = indexer.GetProwlarrID()
	}
	var filteredResults []CommonSearchResult

	// TODO: Add support for multiple pages once Prowlarr supports it (currently broken)
	results, err := pClient.Search(&prowlarr.SearchConfig{
		Indexers:  indexerIDs,
		FreeLeech: true,
	})
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		for _, indexer := range indexers {
			if result.IndexerID == indexer.GetProwlarrID() {
				commonResult := NewProwlarrResult(result)
				if isStale(&commonResult) {
					continue
				}

				maxSize, _ := humanize.ParseBytes(indexer.MaxSize)
				validSeeders := indexer.MaxSeeders == 0 || commonResult.GetSeeders() <= indexer.MaxSeeders
				validSize := maxSize == 0 || commonResult.GetSize() <= maxSize
				validLeechers := indexer.MinLeechers == 0 || commonResult.GetLeechers() >= indexer.MinLeechers

				if validSeeders && validSize && validLeechers {
					filteredResults = append(filteredResults, commonResult)
				}
			}
		}
	}

	return filteredResults, nil
}

// getTorrentsFromJackett searches for torrents in Jackett and applies freeleech filtering
func getTorrentsFromJackett(cfg *Config, indexers []Indexer) ([]CommonSearchResult, error) {
	jClient := jackett.NewClient(cfg.Jackett.Host, cfg.Jackett.APIKey)

	indexerIDs := make([]string, len(indexers))
	for i, indexer := range indexers {
		indexerIDs[i] = indexer.GetJackettID()
	}
	var filteredResults []CommonSearchResult

	// Search with freeleech filtering enabled (heuristic-based)
	results, err := jClient.Search(&jackett.SearchConfig{
		Indexers:  indexerIDs,
		FreeLeech: true, // This applies heuristic filtering in the jackett client
	})
	if err != nil {
		return nil, err
	}

	// For Jackett, we don't have reliable indexer ID matching from the response
	// So we apply filtering to all returned results
	for _, result := range results {
		commonResult := NewJackettResult(result)
		if isStale(&commonResult) {
			continue
		}

		// Apply size and seeder/leecher filters based on any of the configured indexers
		// This is a simplified approach since Jackett doesn't provide reliable indexer mapping
		validForAnyIndexer := false
		for _, indexer := range indexers {
			maxSize, _ := humanize.ParseBytes(indexer.MaxSize)
			validSeeders := indexer.MaxSeeders == 0 || commonResult.GetSeeders() <= indexer.MaxSeeders
			validSize := maxSize == 0 || commonResult.GetSize() <= maxSize
			validLeechers := indexer.MinLeechers == 0 || commonResult.GetLeechers() >= indexer.MinLeechers

			if validSeeders && validSize && validLeechers {
				validForAnyIndexer = true
				break
			}
		}

		if validForAnyIndexer {
			filteredResults = append(filteredResults, commonResult)
		}
	}

	return filteredResults, nil
}

// isStale checks if the torrent is stale (no seeders)
func isStale(torrent *CommonSearchResult) bool {
	return torrent.GetSeeders() == 0
}

// DownloadTorrent downloads the torrents to qBittorrent
// Filters out any torrents that are already downloading
//
// 1. Connect to qBittorrent and get the list of torrents
//
// 3. Download to memory and check if the torrent is already downloading
func DownloadTorrent(result *CommonSearchResult) (*bytes.Buffer, error) {
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
