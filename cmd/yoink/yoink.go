package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/ilyakaznacheev/cleanenv"
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
	Config string `name:"config" help:"configuration file." type:"existingfile" short:"c" required:""`
	DryRun bool   `help:"Dry run. Don't upload torrents to qBittorrent."`

	Run         RunCmd         `cmd:"" help:"Run yoink." default:"1" hidden:""`
	Indexers    IndexersCmd    `cmd:"" help:"List indexers."`
	PrintConfig PrintConfigCmd `cmd:"" help:"Print the configuration."`
	Version     VersionFlag    `name:"version" help:"print version information and quit"`
}

type Context struct {
	config *yoink.Config
	dryRun bool
}

func main() {
	var cli cli //nolint:govet
	ctx := kong.Parse(&cli,
		kong.Name("yoink"),
		kong.Description("Yoink! Command line tool for finding and downloading freeleech torrents."),
		kong.UsageOnError(),
		kong.NamedMapper("yamlfile", kongyaml.YAMLFileMapper),
	)

	if ctx.Validate() == nil {
		printBanner(cli.DryRun)
	}

	cfg, err := parseConfig(&cli)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	ctx.FatalIfErrorf(ctx.Run(&Context{config: cfg, dryRun: cli.DryRun}))
}

func parseConfig(cli *cli) (*yoink.Config, error) {
	config := &yoink.Config{}
	err := cleanenv.ReadConfig(cli.Config, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
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

	if config.TotalFreeleechSize != "" {
		_, err := humanize.ParseBytes(config.TotalFreeleechSize)
		if err != nil {
			return nil, fmt.Errorf("failed to parse total freelech size: %w", err)
		}
	}

	for _, indexer := range config.Indexers {
		_, err := humanize.ParseBytes(indexer.MaxSize)
		if err != nil {
			return nil, fmt.Errorf("failed to parse max size for indexer %d: %w", indexer.ID, err)
		}
	}

	return config, nil
}

func printBanner(dryRun bool) {
	const banner = `
██╗   ██╗ ██████╗ ██╗███╗   ██╗██╗  ██╗
╚██╗ ██╔╝██╔═══██╗██║████╗  ██║██║ ██╔╝
 ╚████╔╝ ██║   ██║██║██╔██╗ ██║█████╔╝ 
  ╚██╔╝  ██║   ██║██║██║╚██╗██║██╔═██╗ 
   ██║   ╚██████╔╝██║██║ ╚████║██║  ██╗
   ╚═╝    ╚═════╝ ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝`
	w := lipgloss.Width(banner)

	fmt.Println(lipgloss.JoinVertical(lipgloss.Top, banner, lipgloss.PlaceHorizontal(w, lipgloss.Center, fmt.Sprintf("%s - MrMarble", version))))
	fmt.Println()
	if dryRun {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).
			Bold(true).
			Align(lipgloss.Center).
			Width(w).
			Border(lipgloss.NormalBorder(), true).
			Render("Running in dry-run mode.\nNo torrents will be downloaded."))
		fmt.Println()
	}
}
