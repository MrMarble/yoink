package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/mrmarble/yoink"
	"github.com/mrmarble/yoink/pkg/jackett"
	"github.com/mrmarble/yoink/pkg/prowlarr"
	"github.com/mrmarble/yoink/pkg/qbittorrent"
)

type RunCmd struct{}

func (r *RunCmd) Run(ctx *Context) error {
	fmt.Println("Total download size:", ctx.config.TotalFreeleechSize)
	parsedTotalSize, _ := humanize.ParseBytes(ctx.config.TotalFreeleechSize)

	// 1. Fetch already downloading torrents from qBittorrent
	qClient := qbittorrent.NewClient(ctx.config.QbitTorrent.Host)
	fmt.Print("Checking qbt connection...")
	err := qClient.Login(ctx.config.QbitTorrent.User, ctx.config.QbitTorrent.Pass)
	if err != nil {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render(" FAIL"))
		return err
	}
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#bfff00")).Render(" OK"))
	qbTorrents, err := yoink.GetDownloadingTorrents(ctx.config, qClient)
	if err != nil {
		return err
	}
	fmt.Printf("Found %d torrents in qBittorrent:\n", len(qbTorrents))
	usedSpace := uint64(0)
	for _, t := range qbTorrents {
		fmt.Printf("  [%s] %s\n", humanize.Bytes(t.Size), cutString(t.Name, 50))
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

	// 2. Check indexer connection
	if ctx.config.IndexerType == "prowlarr" {
		fmt.Print("Checking prowlarr connection...")
		_, err = prowlarr.NewClient(ctx.config.Prowlarr.Host, ctx.config.Prowlarr.APIKey).GetIndexers()
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render(" FAIL"))
			return err
		}
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#bfff00")).Render(" OK"))
	} else if ctx.config.IndexerType == "jackett" {
		fmt.Print("Checking jackett connection...")
		_, err = jackett.NewClient(ctx.config.Jackett.Host, ctx.config.Jackett.APIKey).GetIndexers()
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render(" FAIL"))
			return err
		}
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#bfff00")).Render(" OK"))
	}

	// 3. Fetch new freeleech torrents from indexer
	fmt.Printf("Searching for freeleech torrents using %s...\n", ctx.config.IndexerType)
	torrents, err := yoink.GetTorrents(ctx.config, ctx.config.Indexers)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d freeleech torrents:\n", len(torrents))

	// Filter torrents
	downloadSize := uint64(0)
	filteredResults := make([]yoink.CommonSearchResult, 0)
	for _, torrent := range torrents {

		// Filter out torrents that are too big
		if torrent.GetSize() > spaceLeft {
			continue
		}

		// Filter out torrents that are already downloading
		if isTorrentDownloading(torrent, qbTorrents) {
			continue
		}

		// Filter out torrents that would exceed the total space limit
		if downloadSize+torrent.GetSize() > spaceLeft {
			continue
		}

		downloadSize += torrent.GetSize()
		filteredResults = append(filteredResults, torrent)
	}
	fmt.Printf("Uploading %d torrents (%s) to qBittorrent...\n", len(filteredResults), humanize.Bytes(downloadSize))
	for i := range filteredResults {
		torrent := filteredResults[i]
		fmt.Printf("  [%s] [%d/%d] %s\n", humanize.Bytes(torrent.GetSize()), torrent.GetSeeders(), torrent.GetLeechers(), cutString(torrent.GetTitle(), 50))
		if !ctx.dryRun {
			data, err := yoink.DownloadTorrent(&torrent)
			if err != nil {
				return err
			}
			
			// Extract filename for qBittorrent
			filename := torrent.GetTitle()
			if torrent.GetProwlarrResult() != nil {
				filename = torrent.GetProwlarrResult().FileName
			}
			
			err = qClient.AddTorrentFromBuffer(data, filename, map[string]string{"category": ctx.config.Category, "paused": strconv.FormatBool(ctx.config.Paused)})
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func isTorrentDownloading(torrent yoink.CommonSearchResult, torrents []qbittorrent.Torrent) bool {
	for _, t := range torrents {
		if t.Name == torrent.GetTitle() && t.Size == torrent.GetSize() {
			return true
		}
	}
	return false
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
