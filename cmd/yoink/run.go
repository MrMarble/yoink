package main

import (
	"fmt"
	"strconv"

	"github.com/dustin/go-humanize"
	"github.com/mrmarble/yoink"
	"github.com/mrmarble/yoink/pkg/qbittorrent"
)

type RunCmd struct{}

func (r *RunCmd) Run(ctx *Context) error {
	torrents, err := yoink.GetTorrents(ctx.config)
	if err != nil {
		return err
	}
	fmt.Printf("Found %d torrents from %d indexers\n", len(torrents), len(ctx.config.Indexers))
	if ctx.config.TotalFreelechSize > 0 {
		fmt.Printf("Filtering by total freeleech size: %s\n", humanize.Bytes(ctx.config.TotalFreelechSize))
		torrents, err = yoink.FilterTorrentBySize(torrents, ctx.config)
		if err != nil {
			return err
		}
		fmt.Printf("%d torrents after filtering\n", len(torrents))
	}

	qClient := qbittorrent.NewClient(ctx.config.QbitTorrent.Host)
	fmt.Println("Uploading torrents to qBittorrent...")
	for i := range torrents {
		torrent := torrents[i]
		data, err := yoink.DownloadTorrent(&torrent, ctx.config)
		if err != nil {
			return err
		}
		if data == nil {
			fmt.Printf("Skipping %s because it's already seeding\n", torrent.Title)
			continue
		}
		fmt.Printf("[%s] [%d/%d] %s\n", humanize.Bytes(uint64(torrent.Size)), torrent.Seeders, torrent.Leechers, torrent.Title)
		if !ctx.dryRun {
			err = qClient.AddTorrentFromBuffer(data, torrent.FileName, map[string]string{"category": ctx.config.Category, "paused": strconv.FormatBool(ctx.config.Paused)})
		}
		if err != nil {
			return err
		}
	}
	fmt.Println("Done!")
	return nil
}
