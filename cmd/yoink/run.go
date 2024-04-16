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

	// 2. Get indexers from Prowlarr
	fmt.Print("Checking prowlarr connection...")
	_, err = prowlarr.NewClient(ctx.config.Prowlarr.Host, ctx.config.Prowlarr.APIKey).GetIndexers()
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

	fmt.Printf("Found %d freeleech torrents:\n", len(prTorrents))

	// Filter torrents
	downloadSize := uint64(0)
	for i, prTorrent := range prTorrents {

		// Filter out torrents that are too big
		if uint64(prTorrent.Size) > spaceLeft {
			prTorrents[i] = prTorrents[len(prTorrents)-1]
			prTorrents = prTorrents[:len(prTorrents)-1]
			continue
		}

		isDownloading := false

		// Filter out torrents that are already downloading
		for _, qbTorrent := range qbTorrents {
			if qbTorrent.Name == prTorrent.Title &&
				qbTorrent.Size == uint64(prTorrent.Size) {
				isDownloading = true
				break
			}
		}

		if isDownloading {
			prTorrents[i] = prTorrents[len(prTorrents)-1]
			prTorrents = prTorrents[:len(prTorrents)-1]
			continue
		}

		downloadSize += uint64(prTorrent.Size)
	}
	fmt.Printf("Uploading %d torrents (%s) to qBittorrent...\n", len(prTorrents), humanize.Bytes(downloadSize))
	for i := range prTorrents {
		torrent := prTorrents[i]
		fmt.Printf("  [%s] [%d/%d] %s\n", humanize.Bytes(uint64(torrent.Size)), torrent.Seeders, torrent.Leechers, cutString(torrent.Title, 50))
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
