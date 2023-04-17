package torrent

import (
	"crypto/sha1" //nolint:gosec // we don't need a strong hash here
	"encoding/hex"
	"fmt"

	"github.com/IncSW/go-bencode"
	"github.com/mitchellh/mapstructure"
)

// Torrent represents a torrent file
type Torrent struct {
	Announce string
	Name     string
	Hash     string
}

type encodedTorrent struct {
	Announce []uint8 `mapstructure:"announce"`
	Info     struct {
		Name []uint8 `mapstructure:"name"`
	} `mapstructure:"info"`
}

func ParseTorrentFile(file []byte) (*Torrent, error) {
	var t encodedTorrent
	data, err := bencode.Unmarshal(file)
	if err != nil {
		return nil, err
	}
	err = mapstructure.Decode(data, &t)
	if err != nil {
		return nil, err
	}

	// compute the hash of the info value
	hash, err := computeHash(file)
	if err != nil {
		return nil, err
	}
	torrent := Torrent{
		Announce: string(t.Announce),
		Name:     string(t.Info.Name),
		Hash:     hex.EncodeToString(hash),
	}
	return &torrent, nil
}

func computeHash(data []byte) ([]byte, error) {
	needle := []byte{0x3A, 0x69, 0x6E, 0x66, 0x6F}

	// sliding window of 5 bytes
	for i := 0; i < len(data)-6; i += 6 {
		if data[i] == needle[0] && data[i+1] == needle[1] && data[i+2] == needle[2] && data[i+3] == needle[3] && data[i+4] == needle[4] {
			h := sha1.New() //nolint:gosec // we don't need a strong hash here
			_, err := h.Write(data[i+5 : len(data)-1])
			if err != nil {
				return nil, err
			}
			return h.Sum(nil), nil
		}
	}

	return nil, fmt.Errorf("could not find info key")
}
