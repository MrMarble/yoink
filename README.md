# Yoink - Free leech manager

![GitHub](https://img.shields.io/github/license/mrmarble/yoink)
[![Go Report Card](https://goreportcard.com/badge/github.com/mrmarble/yoink)](https://goreportcard.com/report/github.com/mrmarble/yoink)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mrmarble/yoink)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/mrmarble/yoink)

`yoink` is an app designed to help you download torrents marked as `free leech` in order to mantain your ratio in private trackers.

`yoink` can search all your trackers using [prowlarr](https://github.com/Prowlarr/Prowlarr) as the indexer and automatically add them to you [qBitTorrent](https://github.com/qbittorrent/qBittorrent) client to start seeding

## Basic behavior

1. Get a list of torrents marked as **freeleech** from prowlarr.
2. Connect to qBitTorrent and filter-out any already downloaded torrent.
3. Upload remaining torrents to qBitTorrent and start seeding.

## Configuration

Some properties can be passed by environment variable or cli argument.

### File

```yaml
prowlarr: # prowlarr connection
  host: "https://localhost:9696"
  api_key: "xxxxxxxxxxxxxxxxxxxxxxxxxx"

qbittorrent: # qbittorrent connection
  host: "https://localhost:8080"
  # not needed if access is open
  user: "admin"
  password: "admin"

# used to calculate available space.
download_dir: "/media/downloads/free_leech"
total_freeleech_size: 200Gb
# qBitTorrent category to set to torrents
category: "FreeLeech"

# Indexer configuration. If not present will be ignored
indexers:
  - id: 2
  - id: 11 # prowlarr indexer ID
    max_size: 40Gb # max file size to download
    max_seeders: 10 # if seeders are greater than this torrent will be ignored
```

### Environment

Environemnt variables will override config file

- PROWLARR_API_URL
- PROWLARR_API_KEY
- QBITTORRENT_URL
- QBITTORRENT_USER
- QBITTORRENT_PASS

## Usage

CLI parameters will override enviroment variables

```shell
$ yoink --help
Usage: yoink <command>

Yoink! Command line tool for finding and downloading freeleech torrents.

Flags:
  -h, --help                       Show context-sensitive help.
      --prowlarr-url=STRING        Prowlarr URL ($PROWLARR_API_URL)
      --prowlarr-api-key=STRING    Prowlarr API Key ($PROWLARR_API_KEY).
      --qbittorrent-url=STRING     qBitTorrent URL ($QBITTORRENT_URL)
      --qbittorrent-user=STRING    qBitTorrent user to authenticante with ($QBITTORRENT_USER)
      --qbittorrent-pass=STRING    qBitTorrent password to authenticante with ($QBITTORRENT_PASS)
  -c, --config=CONFIG              configuration file.
      --version                    print version information and quit

Commands:
  indexers
    List indexers.

Run "yoink <command> --help" for more information on a command.
```

Example:

```shell
# don't save sensible info in config file
$ yoink --prowlarr-api-key XXXXXXXX --qbittorrent-pass SecretPassword --config ./config.yaml
```

Docker:

```shell
$ docker run -e "PROWLARR_API_KEY=XXXXXXXXXX" \
    -e "QBITTORRENT_PASS=SecurePassword" \
    -v ./config.yaml:/config.yaml:ro \
    gcr.io/mrmarble/yoink:latest
```
