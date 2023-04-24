// Package yoink provides utilities to manage freeleech
// downloads automatically
package yoink

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/dustin/go-humanize" //nolint:typecheck
	"github.com/mrmarble/yoink/pkg/prowlarr"
	"github.com/mrmarble/yoink/pkg/qbittorrent"
)

// GetTorrents searches for freeleech torrents in Prowlarr and filters them based on the indexer configuration
func GetTorrents(cfg *Config, indexers []Indexer) ([]prowlarr.SearchResult, error) {
	pClient := prowlarr.NewClient(cfg.Prowlarr.Host, cfg.Prowlarr.APIKey)

	indexerIds := make([]int, len(indexers))
	for i, indexer := range indexers {
		indexerIds[i] = indexer.ID
	}
	var filteredResults []prowlarr.SearchResult

	// TODO: Add support for multiple pages once Prowlarr supports it (currently broken)
	results, err := pClient.Search(&prowlarr.SearchConfig{
		Indexers: indexerIds,
	})
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		for _, indexer := range indexers {
			if result.IndexerID == indexer.ID {
				maxSize, _ := humanize.ParseBytes(indexer.MaxSize)
				if (indexer.MaxSeeders == 0 || result.Seeders <= indexer.MaxSeeders) && uint64(result.Size) <= maxSize && result.Seeders > 0 && result.IsFreeleech() {
					filteredResults = append(filteredResults, result)
				}
			}
		}
	}

	return filteredResults, nil
}

// DownloadTorrents downloads the torrents to qBittorrent
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

func GetDownloadingTorrents(config *Config) ([]qbittorrent.Torrent, error) {
	qClient := qbittorrent.NewClient(config.QbitTorrent.Host) // TODO: Add user/pass

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
