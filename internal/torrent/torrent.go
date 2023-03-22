package torrent

import (
	"github.com/IncSW/go-bencode"
)

// Torrent is a torrent
type Torrent struct {
	Announce string
	Name     string
	Length   int
}

func ParseTorrentFile(file []byte) (*Torrent, error) {
	var t Torrent
	data, err := bencode.Unmarshal(file)
	if err != nil {
		return nil, err
	}

	t.Announce = string(data.(map[string]interface{})["announce"].([]uint8))
	t.Name = string(data.(map[string]interface{})["info"].(map[string]interface{})["name"].([]uint8))
	// t.Length = int(data.(map[string]interface{})["info"].(map[string]interface{})["length"].(int64))
	return &t, nil
}
