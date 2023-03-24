package main

import (
	"fmt"

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
	for i := range torrents {
		torrent := torrents[i]
		fmt.Printf("Downloading %s\n", torrent.Title)
		data, err := yoink.DownloadTorrent(&torrent, ctx.config)
		if err != nil {
			return err
		}
		if data == nil {
			fmt.Printf("Skipping %s because it's already seeding\n", torrent.Title)
			continue
		}
		fmt.Printf("Uploading %s\n", torrent.Title)
		err = qClient.AddTorrentFromBuffer(data, torrent.FileName, map[string]string{"category": ctx.config.Category})
		if err != nil {
			return err
		}
	}
	fmt.Println("Done!")
	return nil
}
