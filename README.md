# Yoink - Free leech manager

![GitHub](https://img.shields.io/github/license/mrmarble/yoink)
[![Go Report Card](https://goreportcard.com/badge/github.com/mrmarble/yoink)](https://goreportcard.com/report/github.com/mrmarble/yoink)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mrmarble/yoink)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/mrmarble/yoink)

> An exclamation that transfers ownership of an object to the person who utters it, regardless of previous property rights.

`yoink` is an app designed to help you download torrents marked as `free leech` in order to mantain your ratio in private trackers.

`yoink` can search all your trackers using [prowlarr](https://github.com/Prowlarr/Prowlarr) or [jackett](https://github.com/Jackett/Jackett) as the indexer and automatically add them to your [qBitTorrent](https://github.com/qbittorrent/qBittorrent) client to start seeding

## Basic behavior

1. Get a list of torrents marked as **freeleech** from your indexer (Prowlarr or Jackett).
2. Connect to qBitTorrent and filter-out any already downloaded torrent.
3. Upload remaining torrents to qBitTorrent and start seeding.

## Indexer Support

### Prowlarr (Recommended)
Prowlarr has native support for freeleech filtering, making it the preferred indexer.

### Jackett
Jackett support is provided with the following limitations:
- **No native freeleech filtering**: Jackett doesn't provide explicit freeleech flags in the API
- **Heuristic-based detection**: The application uses keyword matching in titles and descriptions to identify freeleech torrents
- **Less reliable**: May miss some freeleech torrents or include false positives

Common freeleech indicators detected: `freeleech`, `free leech`, `fl`, `[fl]`, `(fl)`, `[free]`, `(free)`, `0%`, etc.

## Configuration

Some properties can be passed by environment variable.

### File

<!-- CONFIG_FILE -->
```yaml
total_freeleech_size: "200GB" # Max space to use for downloads. If 0, no limit is applied
category: "FreeLeech" # Category to use for downloads.
paused: true # Whether to pause torrents after adding them to qBittorrent
indexer_type: "prowlarr" # Type of indexer to use: prowlarr or jackett
qbittorrent: # Connection details for qBittorrent
  host: "http://localhost:8080"
  username: "admin"
  password: "adminadmin"
prowlarr: # Connection details for Prowlarr
  host: "http://localhost:8081"
  api_key: "1234567890"
jackett: # Connection details for Jackett  
  host: "http://localhost:9117"
  api_key: "1234567890"
indexers: # List of indexers to use. Filters out any indexers not in this list
- id: 1 # ID of the indexer (integer for Prowlarr, string for Jackett)
  max_seeders: 20 # Maximum number of seeders to allow. 0 = no limit
  max_size: "50GB" # Maximum file size to allow. 0 = no limit
  min_leechers: 0 # Minimum number of leechers to allow. 0 = no limit
- id: 3 # ID of the indexer (integer for Prowlarr, string for Jackett)
  max_seeders: 10 # Maximum number of seeders to allow. 0 = no limit
  max_size: "50GB" # Maximum file size to allow. 0 = no limit
  min_leechers: 0 # Minimum number of leechers to allow. 0 = no limit
```
<!-- END_CONFIG_FILE -->

#### Jackett Configuration Example

```yaml
total_freeleech_size: "200GB"
category: "FreeLeech" 
paused: true
indexer_type: "jackett" # Use jackett instead of prowlarr
qbittorrent:
  host: "http://localhost:8080"
  username: "admin"
  password: "adminadmin"
jackett:
  host: "http://localhost:9117"  # Default Jackett port
  api_key: "your-jackett-api-key"
indexers:
- id: "tracker1" # Use string IDs for Jackett indexers
  max_seeders: 20
  max_size: "50GB"
  min_leechers: 0
```

### Environment

Environment variables will override config file
<!-- ENV_VARS -->
```
Environment variables:
  INDEXER_TYPE string
    	Type of indexer to use: prowlarr or jackett (default "prowlarr")
  TOTAL_FREELEECH_SIZE string
    	Max space to use for downloads. If 0, no limit is applied (default "200GB")
  CATEGORY string
    	Category to use for downloads. (default "FreeLeech")
  PAUSED bool
    	Whether to pause torrents after adding them to qBittorrent (default "true")
  QBIT_HOST string
    	Connection details for qBittorrent
  QBIT_USER string
    	Connection details for qBittorrent
  QBIT_PASS string
    	Connection details for qBittorrent
  PROWLARR_HOST string
    	Connection details for Prowlarr
  PROWLARR_API_KEY string
    	Connection details for Prowlarr
  JACKETT_HOST string
    	Connection details for Jackett
  JACKETT_API_KEY string
    	Connection details for Jackett
```
<!-- END_ENV_VARS -->

## Usage

CLI parameters will override environment variables

```
$ yoink --help
Usage: yoink --config=STRING <command>

Yoink! Command line tool for finding and downloading freeleech torrents.

Flags:
  -h, --help             Show context-sensitive help.
  -c, --config=STRING    configuration file.
      --dry-run          Dry run. Don't upload torrents to qBittorrent.
      --version          print version information and quit

Commands:
  indexers --config=STRING
    List indexers.

  print-config --config=STRING
    Print the configuration.

Run "yoink <command> --help" for more information on a command.
```

Example:

```shell
# Using Prowlarr (recommended)
$ PROWLARR_API_KEY=XXXXXXXXXX QBIT_PASS=SecurePassword yoink --config ./config.yaml

# Using Jackett
$ INDEXER_TYPE=jackett JACKETT_API_KEY=XXXXXXXXXX QBIT_PASS=SecurePassword yoink --config ./config.yaml
```

Docker:

```shell
# Using Prowlarr
$ docker run -e "PROWLARR_API_KEY=XXXXXXXXXX" \
    -e "QBIT_PASS=SecurePassword" \
    -v ./config.yaml:/config.yaml:ro \
    ghcr.io/mrmarble/yoink:latest

# Using Jackett  
$ docker run -e "INDEXER_TYPE=jackett" \
    -e "JACKETT_API_KEY=XXXXXXXXXX" \
    -e "QBIT_PASS=SecurePassword" \
    -v ./config.yaml:/config.yaml:ro \
    ghcr.io/mrmarble/yoink:latest
```