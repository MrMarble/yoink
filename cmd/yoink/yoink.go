package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
	"github.com/dustin/go-humanize"
	"github.com/mrmarble/yoink"
)

var (
	// Populated by goreleaser during build
	version = "master"
	commit  = "?"
	date    = ""
)

type VersionFlag string

func (v VersionFlag) Decode(_ *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                       { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong) error {
	fmt.Printf("yoink has version %s built from %s on %s\n", version, commit, date)
	app.Exit(0)

	return nil
}

type cli struct {
	ProwlarrURL    string `help:"Prowlarr URL" env:"PROWLARR_API_URL"`
	ProwlarrAPIKey string `help:"Prowlarr API Key." env:"PROWLARR_API_KEY"`

	QBitTorrentURL  string `help:"qBitTorrent URL" name:"qbittorrent-url" env:"QBITTORRENT_URL"`
	QbitTorrentUser string `help:"qBitTorrent user to authenticante with" name:"qbittorrent-user" env:"QBITTORRENT_USER"`
	QbitTorrentPass string `help:"qBitTorrent password to authenticante with" name:"qbittorrent-pass" env:"QBITTORRENT_PASS"`

	Config *config `name:"config" help:"configuration file." type:"yamlfile" short:"c"`

	Run      RunCmd      `cmd:"" help:"Run yoink." default:"1" hidden:""`
	Indexers IndexersCmd `cmd:"" help:"List indexers."`
	Version  VersionFlag `name:"version" help:"print version information and quit"`
}

type config struct {
	QbitTorrent struct {
		Host string
		User string
		Pass string
	}
	Prowlarr struct {
		Host   string
		APIKey string `yaml:"api_key"`
	}
	DownloadDir       string `yaml:"download_dir"`
	TotalFreelechSize string `yaml:"total_freelech_size"`
	Indexers          []struct {
		ID         int
		MaxSeeders int    `yaml:"max_seeders"`
		MaxSize    string `yaml:"max_size"`
	} `yaml:"indexers"`
	Category string
}

type Context struct {
	config *yoink.Config
}

func main() {
	var cli cli
	ctx := kong.Parse(&cli,
		kong.Name("yoink"),
		kong.Description("Yoink! Command line tool for finding and downloading freeleech torrents."),
		kong.UsageOnError(),
		kong.NamedMapper("yamlfile", kongyaml.YAMLFileMapper),
	)

	if ctx.Validate() == nil {
		fmt.Printf("yoink %s\n\n", version)
	}

	cfg, err := unifyConfig(&cli)
	ctx.FatalIfErrorf(err)

	ctx.FatalIfErrorf(ctx.Run(&Context{config: cfg}))
}

func unifyConfig(cli *cli) (*yoink.Config, error) {
	config := &yoink.Config{
		DownloadDir: cli.Config.DownloadDir,
		Category:    cli.Config.Category,
	}

	// Override config with CLI flags
	if cli.ProwlarrURL != "" {
		config.Prowlarr.Host = cli.ProwlarrURL
	}
	if cli.ProwlarrAPIKey != "" {
		config.Prowlarr.APIKey = cli.ProwlarrAPIKey
	}

	if cli.QBitTorrentURL != "" {
		config.QbitTorrent.Host = cli.QBitTorrentURL
	}
	if cli.QbitTorrentUser != "" {
		config.QbitTorrent.User = cli.QbitTorrentUser
	}
	if cli.QbitTorrentPass != "" {
		config.QbitTorrent.Pass = cli.QbitTorrentPass
	}

	// Validate config
	if config.Prowlarr.Host == "" {
		return nil, fmt.Errorf("Prowlarr URL must be specified")
	}
	if config.Prowlarr.APIKey == "" {
		return nil, fmt.Errorf("Prowlarr API Key must be specified")
	}
	if config.QbitTorrent.Host == "" {
		return nil, fmt.Errorf("qBitTorrent URL must be specified")
	}

	if cli.Config.TotalFreelechSize != "" {
		size, err := humanize.ParseBytes(cli.Config.TotalFreelechSize)
		if err != nil {
			return nil, fmt.Errorf("failed to parse total freelech size: %w", err)
		}
		config.TotalFreelechSize = size
	}

	for _, indexer := range cli.Config.Indexers {
		size, err := humanize.ParseBytes(indexer.MaxSize)
		if err != nil {
			return nil, fmt.Errorf("failed to parse max size for indexer %d: %w", indexer.ID, err)
		}
		config.Indexers = append(config.Indexers, struct {
			ID         int
			MaxSeeders int
			SeedTime   int
			MaxSize    uint
		}{
			ID:         indexer.ID,
			MaxSeeders: indexer.MaxSeeders,
			MaxSize:    uint(size),
		})
	}

	return config, nil
}
