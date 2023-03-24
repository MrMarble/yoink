// Package yoink provides utilities to manage freeleech
// downloads automatically
package yoink

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/mrmarble/yoink/internal/du"      //nolint:typecheck
	"github.com/mrmarble/yoink/internal/torrent" //nolint:typecheck
	"github.com/mrmarble/yoink/pkg/prowlarr"
	"github.com/mrmarble/yoink/pkg/qbittorrent"
)

// Config holds the configuration for yoink
type Config struct {
	// Connection details for qBittorrent
	QbitTorrent struct {
		Host string
		User string
		Pass string
	}

	// Connection details for Prowlarr
	Prowlarr struct {
		Host   string
		APIKey string
	}

	// Directory to download to. Used to get the available space on the drive. If empty, space is not checked
	DownloadDir string

	// Max space to use for downloads. If 0, no limit is applied
	TotalFreelechSize uint64

	// List of indexers to use. Filters out any indexers not in this list
	Indexers []struct {
		// ID of the indexer in Prowlarr
		ID int
		// Maximum number of seeders to allow. 0 = no limit
		MaxSeeders int
		// Minimum hours to seed for. 0 = no limit
		SeedTime int
		// Maximum file size to allow. 0 = no limit
		MaxSize uint
	}

	// Category to use for downloads. Used for automatic deletion of old downloads
	Category string
}

// GetTorrents searches for freeleech torrents in Prowlarr and filters them based on the indexer configuration
func GetTorrents(config *Config) ([]prowlarr.SearchResult, error) {
	pClient := prowlarr.NewClient(config.Prowlarr.Host, config.Prowlarr.APIKey)

	indexerIds := make([]int, len(config.Indexers))
	for i, indexer := range config.Indexers {
		indexerIds[i] = indexer.ID
	}

	results, err := pClient.Search(&prowlarr.SearchConfig{
		Indexers: indexerIds,
	})
	if err != nil {
		return nil, err
	}

	var filteredResults []prowlarr.SearchResult
	for _, result := range results {
		for _, indexer := range config.Indexers {
			if result.IndexerID == indexer.ID {
				if (indexer.MaxSeeders == 0 || result.Seeders <= indexer.MaxSeeders) && result.Size <= indexer.MaxSize && result.IsFreeleech() {
					filteredResults = append(filteredResults, result)
				}
			}
		}
	}

	return filteredResults, nil
}

// FilterTorrentBySize filters out torrents that are too large to download based on the available space on the drive.
//
// 1. Connect to qBittorrent and get the list of torrents
//
// 2. Get the total size of all the torrents
//
// 3. Get the available space on the drive
//
// 4. Filter out any torrents that would exceed the available space
func FilterTorrentBySize(torrents []prowlarr.SearchResult, config *Config) ([]prowlarr.SearchResult, error) {
	qClient := qbittorrent.NewClient(config.QbitTorrent.Host) // TODO: Add user/pass

	qTorrents, err := qClient.GetTorrents()
	if err != nil {
		return nil, err
	}

	var totalSize uint64
	for _, torrent := range qTorrents {
		if torrent.Category == config.Category {
			totalSize += torrent.Size
		}
	}

	// If the total size of the torrents is greater than the max freelech size, don't download anything
	if totalSize >= config.TotalFreelechSize {
		return nil, nil
	}

	// If the download directory is not set, don't check the available space
	if config.DownloadDir == "" {
		return torrents, nil
	}

	filteredTorrents := filterTorrentsByDiskSize(config, totalSize, torrents)

	return filteredTorrents, nil
}

func filterTorrentsByDiskSize(config *Config, totalSize uint64, torrents []prowlarr.SearchResult) []prowlarr.SearchResult {
	availableSpace := du.Available(config.DownloadDir) - totalSize

	// If the available space is less than the max freelech size, don't download anything
	if availableSpace < config.TotalFreelechSize {
		return nil
	}

	var filteredTorrents []prowlarr.SearchResult
	for _, torrent := range torrents {
		if uint64(torrent.Size) <= availableSpace {
			filteredTorrents = append(filteredTorrents, torrent)
		}
	}
	return filteredTorrents
}

// DownloadTorrents downloads the torrents to qBittorrent
// Filters out any torrents that are already downloading
//
// 1. Connect to qBittorrent and get the list of torrents
//
// 3. Download to memory and check if the torrent is already downloading
func DownloadTorrent(result *prowlarr.SearchResult, config *Config) (*bytes.Buffer, error) {
	qClient := qbittorrent.NewClient(config.QbitTorrent.Host) // TODO: Add user/pass

	qTorrents, err := qClient.GetTorrents()
	if err != nil {
		return nil, err
	}

	buf, err := downloadFile(result.DownloadURL)
	if err != nil {
		return nil, err
	}

	tFile, err := torrent.ParseTorrentFile(buf.Bytes())
	if err != nil {
		return nil, err
	}

	// Check if the torrent is already downloading
	for _, qTorrent := range qTorrents {
		if qTorrent.Name == tFile.Name {
			return nil, nil
		}
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
