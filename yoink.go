// Package yoink provides utilities to manage freeleech
// downloads automatically
package yoink

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/dustin/go-humanize"
	"github.com/mrmarble/yoink/pkg/prowlarr"
	"github.com/mrmarble/yoink/pkg/qbittorrent"
)

// GetTorrents searches for freeleech torrents in Prowlarr and filters them based on the indexer configuration
func GetTorrents(cfg *Config, indexers []Indexer) ([]prowlarr.SearchResult, error) {
	pClient := prowlarr.NewClient(cfg.Prowlarr.Host, cfg.Prowlarr.APIKey)

	indexerIDs := make([]int, len(indexers))
	for i, indexer := range indexers {
		indexerIDs[i] = indexer.ID
	}
	var filteredResults []prowlarr.SearchResult

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
			if result.IndexerID == indexer.ID {
				result := result
				if isStale(&result) {
					continue
				}

				maxSize, _ := humanize.ParseBytes(indexer.MaxSize)
				validSeeders := indexer.MaxSeeders == 0 || result.Seeders <= indexer.MaxSeeders
				validSize := maxSize == 0 || uint64(result.Size) <= maxSize
				validLeechers := indexer.MinLeechers == 0 || result.Leechers >= indexer.MinLeechers

				if validSeeders && validSize && validLeechers {
					filteredResults = append(filteredResults, result)
				}
			}
		}
	}

	return filteredResults, nil
}

// isStale checks if the torrent is stale (no seeders)
func isStale(torrent *prowlarr.SearchResult) bool {
	return torrent.Seeders == 0
}

// DownloadTorrent downloads the torrents to qBittorrent
// Filters out any torrents that are already downloading
//
// 1. Connect to qBittorrent and get the list of torrents
//
// 3. Download to memory and check if the torrent is already downloading
func DownloadTorrent(result *prowlarr.SearchResult) (*bytes.Buffer, error) {
	buf, err := downloadFile(result.DownloadURL)
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
