package prowlarr

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func MockServer(path, response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// All requests require an API key
		if r.Header.Get("X-Api-Key") != "test" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Write([]byte(response))
	}))
}

func MockServerWithHandler(path string, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// All requests require an API key
		if r.Header.Get("X-Api-Key") != "test" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}))
}

func TestClient_generateQueryString(t *testing.T) {
	type args struct {
		config *SearchConfig
	}
	tests := []struct {
		name string
		c    *Client
		args args
		want string
	}{
		{"empty", &Client{}, args{&SearchConfig{}}, ""},
		{"indexers", &Client{}, args{&SearchConfig{Indexers: []int{1, 2, 3}}}, "&indexerIds=1&indexerIds=2&indexerIds=3"},
		{"limit", &Client{}, args{&SearchConfig{Limit: 10}}, "&limit=10"},
		{"offset", &Client{}, args{&SearchConfig{Offset: 10}}, "&offset=10"},
		{"all", &Client{}, args{&SearchConfig{Indexers: []int{1, 2, 3}, Limit: 10, Offset: 10}}, "&indexerIds=1&indexerIds=2&indexerIds=3&limit=10&offset=10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.generateQueryString(tt.args.config); got != tt.want {
				t.Errorf("Client.generateQueryString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetAPIVersion(t *testing.T) {
	server := MockServer("/api", `{"current":"v1"}`)
	defer server.Close()

	c := NewClient(server.URL, "test")
	got, err := c.GetAPIVersion()
	if err != nil {
		t.Errorf("Client.GetAPIVersion() error = %v", err)
		return
	}

	if got != "v1" {
		t.Errorf("Client.GetAPIVersion() = %v, want %v", got, "v1")
	}
}

func TestClient_GetIndexers(t *testing.T) {
	server := MockServer("/api/v1/indexer", `[ { "id": 1, "name": "Test", "protocol": "torrent", "enable": true } ]`)
	defer server.Close()

	c := NewClient(server.URL, "test")
	got, err := c.GetIndexers()
	if err != nil {
		t.Errorf("Client.GetIndexers() error = %v", err)
		return
	}

	if len(got) != 1 {
		t.Errorf("Client.GetIndexers() = %v, want %v", got, 1)
	}

	if got[0].ID != 1 {
		t.Errorf("Client.GetIndexers() = %v, want %v", got[0].ID, 1)
	}

	if got[0].Name != "Test" {
		t.Errorf("Client.GetIndexers() = %v, want %v", got[0].Name, "Test")
	}

	if got[0].Protocol != "torrent" {
		t.Errorf("Client.GetIndexers() = %v, want %v", got[0].Protocol, "torrent")
	}

	if got[0].Enable != true {
		t.Errorf("Client.GetIndexers() = %v, want %v", got[0].Enable, true)
	}
}

func TestClient_Search(t *testing.T) {
	server := MockServerWithHandler("/api/v1/search", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("offset") != "5" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Write([]byte(`[ { "title": "Test", "protocol": "torrent", "indexerId": 1, "downloadUrl": "http://test.com" } ]`))
	})
	defer server.Close()

	c := NewClient(server.URL, "test")
	got, err := c.Search(&SearchConfig{Offset: 5})
	if err != nil {
		t.Errorf("Client.Search() error = %v", err)
		return
	}

	if len(got) != 1 {
		t.Errorf("Client.Search() = %v, want %v", got, 1)
	}

	if got[0].Title != "Test" {
		t.Errorf("Client.Search() = %v, want %v", got[0].Title, "Test")
	}

	if got[0].Protocol != "torrent" {
		t.Errorf("Client.Search() = %v, want %v", got[0].Protocol, "torrent")
	}

	if got[0].IndexerID != 1 {
		t.Errorf("Client.Search() = %v, want %v", got[0].IndexerID, 1)
	}

	if got[0].DownloadURL != "http://test.com" {
		t.Errorf("Client.Search() = %v, want %v", got[0].DownloadURL, "http://test.com")
	}
}
