package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/mrmarble/yoink"
	"github.com/mrmarble/yoink/pkg/prowlarr"
	"github.com/mrmarble/yoink/pkg/qbittorrent"
)

type RunCmd struct{}

func (r *RunCmd) Run(ctx *Context) error {
	fmt.Println("Total download size:", ctx.config.TotalFreeleechSize)
	parsedTotalSize, _ := humanize.ParseBytes(ctx.config.TotalFreeleechSize)

	// 1. Fetch already downloading torrents from qBittorrent
	fmt.Print("Checking qbt connection...")
	qbTorrents, err := yoink.GetDownloadingTorrents(ctx.config)
	if err != nil {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render(" FAIL"))
		return err
	}
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#bfff00")).Render(" OK"))
	fmt.Printf("Found %d torrents in qBittorrent:\n", len(qbTorrents))
	usedSpace := uint64(0)
	for _, t := range qbTorrents {
		fmt.Printf("  [%s] %s\n", humanize.Bytes(t.Size), cutString(t.Name, 30))
		usedSpace += t.Size
	}

	if usedSpace >= parsedTotalSize {
		fmt.Println("Not enough space left to download new torrents. Exiting...")
		return nil
	}

	spaceLeft := parsedTotalSize - usedSpace

	if parsedTotalSize <= 0 {
		spaceLeft = 1024 ^ 6
	}

	fmt.Printf("Used space: %s, Left: %s\n", humanize.Bytes(usedSpace), humanize.Bytes(spaceLeft))

	// 2. Get indexers from Prowlarr
	fmt.Print("Checking prowlarr connection...")
	indexers, err := prowlarr.NewClient(ctx.config.Prowlarr.Host, ctx.config.Prowlarr.APIKey).GetIndexers()
	if err != nil {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render(" FAIL"))
		return err
	}
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#bfff00")).Render(" OK"))

	// 2. Fetch new freeleech torrents from Prowlarr
	fmt.Println("Searching for freeleech torrents...")
	prTorrents, err := yoink.GetTorrents(ctx.config, ctx.config.Indexers)
	if err != nil {
		return err
	}

	// Filter out already downloading torrents
	filteredTorrents := make([]prowlarr.SearchResult, 0)
	downloadSize := uint64(0)
	for _, prTorrent := range prTorrents {
		for _, qbTorrent := range qbTorrents {
			if uint64(prTorrent.Size) > spaceLeft {
				break
			}

			if qbTorrent.Name != prTorrent.Title && qbTorrent.Size != uint64(prTorrent.Size) {
				filteredTorrents = append(filteredTorrents, prTorrent)
				spaceLeft -= uint64(prTorrent.Size)
				downloadSize += uint64(prTorrent.Size)
				break
			}

			// check if is the same tracker
			sameTracker := false
			for _, indexer := range indexers {
				if indexer.ID == prTorrent.IndexerID {
					for _, url := range indexer.IndexerUrls {
						if strings.Contains(qbTorrent.Tracker, url) {
							sameTracker = true
							break
						}
					}
				}
			}

			if !sameTracker {
				filteredTorrents = append(filteredTorrents, prTorrent)
				spaceLeft -= uint64(prTorrent.Size)
				downloadSize += uint64(prTorrent.Size)
			}
		}
	}
	qClient := qbittorrent.NewClient(ctx.config.QbitTorrent.Host)
	fmt.Printf("Uploading %d torrents (%s) to qBittorrent...\n", len(filteredTorrents), humanize.Bytes(downloadSize))
	for i := range filteredTorrents {
		torrent := filteredTorrents[i]
		fmt.Printf("  [%s] [%d/%d] %s\n", humanize.Bytes(uint64(torrent.Size)), torrent.Seeders, torrent.Leechers, cutString(torrent.Title, 30))
		if !ctx.dryRun {
			data, err := yoink.DownloadTorrent(&torrent)
			if err != nil {
				return err
			}
			err = qClient.AddTorrentFromBuffer(data, torrent.FileName, map[string]string{"category": ctx.config.Category, "paused": strconv.FormatBool(ctx.config.Paused)})
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func cutString(s string, l int) string {
	if len(s) <= l {
		return s
	}
	words := strings.Fields(s)
	s = ""
	for _, w := range words {
		if len(s)+len(w) > l {
			break
		}
		s += w + " "
	}
	return strings.TrimSpace(s) + "..."
}
