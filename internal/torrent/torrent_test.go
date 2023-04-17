package torrent

import (
	"io"
	"os"
	"testing"
)

func TestParseTorrentFile(t *testing.T) {
	file, err := os.Open("testdata/test.torrent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	torrent, err := ParseTorrentFile(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	want := Torrent{
		Announce: "http://yoink.tracker",
		Name:     "yoink",
		Hash:     "2cebb70fef00a76d7f8adcee2b334fe1d2b0332d",
	}

	if *torrent != want {
		t.Errorf("expected %+v, got %+v", want, torrent)
	}
}
