package qbittorrent

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/torrents/info":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"name":"test","size":123,"category":"test","hash":"test","tracker":"test"}]`))
		default:
			t.Errorf("unexpected request: %s", r.URL.Path)
		}
	}))
}

func TestClient_GetTorrents(t *testing.T) {
	server := mockServer(t)
	defer server.Close()

	client := NewClient(server.URL)
	torrents, err := client.GetTorrents()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(torrents) != 1 {
		t.Errorf("expected 1 torrent, got %d", len(torrents))
	}

	want := Torrent{
		Name:     "test",
		Size:     123,
		Category: "test",
		Hash:     "test",
		Tracker:  "test",
	}

	if torrents[0] != want {
		t.Errorf("expected %+v, got %+v", want, torrents[0])
	}
}
