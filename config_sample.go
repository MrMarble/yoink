// The following directive is necessary to make the package coherent:

//go:build ignore
// +build ignore

package main

import (
	"os"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/mrmarble/yoink"
	marshal "github.com/rockholla/go-lib/marshal"
)

func main() {
	cfg := yoink.Config{
		TotalFreeleechSize: "200GB",
		Category:           "FreeLeech",
		Paused:             true,
		QbitTorrent: struct {
			Host string "yaml:\"host\" env:\"QBIT_HOST\" env-description:\"Connection details for qBittorrent\""
			User string "yaml:\"username\" env:\"QBIT_USER\" env-description:\"Connection details for qBittorrent\""
			Pass string "yaml:\"password\" env:\"QBIT_PASS\" env-description:\"Connection details for qBittorrent\""
		}{
			Host: "http://localhost:8080",
			User: "admin",
			Pass: "adminadmin",
		},
		Prowlarr: struct {
			Host   string "yaml:\"host\" env:\"PROWLARR_HOST\" env-description:\"Connection details for Prowlarr\""
			APIKey string "yaml:\"api_key\" env:\"PROWLARR_API_KEY\" env-description:\"Connection details for Prowlarr\""
		}{
			Host:   "http://localhost:8081",
			APIKey: "1234567890",
		},
		Indexers: []yoink.Indexer{
			{
				ID:         1,
				MaxSeeders: 20,
				MaxSize:    "50GB",
			}, {
				ID:         3,
				MaxSeeders: 10,
				MaxSize:    "50GB",
			},
		},
	}

	// Save config to config.sample.yaml
	out, err := marshal.YAMLWithComments(cfg, 0)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("config.sample.yaml", []byte(out), os.ModePerm)
	if err != nil {
		panic(err)
	}

	// Modify sample config in readme file
	readme, err := os.ReadFile("README.md")
	if err != nil {
		panic(err)
	}

	readmeStr := string(readme)
	configStart := strings.Index(readmeStr, "<!-- CONFIG_FILE -->")
	configEnd := strings.Index(readmeStr, "<!-- END_CONFIG_FILE -->")

	readmeStr = readmeStr[:configStart+21] + "\n```yaml\n" + out + "```\n" + readmeStr[configEnd:]
	envStart := strings.Index(readmeStr, "<!-- ENV_VARS -->")
	envEnd := strings.Index(readmeStr, "<!-- END_ENV_VARS -->")
	env, err := cleanenv.GetDescription(&cfg, nil)
	if err != nil {
		panic(err)
	}
	readmeStr = readmeStr[:envStart+18] + "\n```\n" + env + "\n```\n" + readmeStr[envEnd:]
	err = os.WriteFile("README.md", []byte(readmeStr), os.ModePerm)
	if err != nil {
		panic(err)
	}
}
