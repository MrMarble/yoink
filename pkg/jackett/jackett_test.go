package jackett

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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
}

func MockServerWithHandler(path string, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		handler(w, r)
	}))
}

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:9117", "test-api-key")
	if client.url != "http://localhost:9117" {
		t.Errorf("Expected URL to be http://localhost:9117, got %s", client.url)
	}
	if client.apiKey != "test-api-key" {
		t.Errorf("Expected API key to be test-api-key, got %s", client.apiKey)
	}
}

func TestClient_GetAPIVersion(t *testing.T) {
	client := NewClient("http://localhost:9117", "test")
	version, err := client.GetAPIVersion()
	if err != nil {
		t.Errorf("Client.GetAPIVersion() error = %v", err)
		return
	}
	if version != "jackett-api" {
		t.Errorf("Client.GetAPIVersion() = %v, want %v", version, "jackett-api")
	}
}

func TestClient_Search(t *testing.T) {
	mockResponse := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:torznab="http://torznab.com/schemas/2015/feed">
  <channel>
    <title>Jackett</title>
    <description>Jackett RSS Feed</description>
    <item>
      <title>Test Torrent [FL]</title>
      <link>http://test.com/download</link>
      <guid>123456</guid>
      <category>Movies</category>
      <description>A test torrent</description>
      <torznab:attr name="size" value="1073741824"/>
      <torznab:attr name="seeders" value="10"/>
      <torznab:attr name="peers" value="5"/>
    </item>
  </channel>
</rss>`

	server := MockServer("/api/v2.0/indexers/all/results", mockResponse)
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	results, err := client.Search(&SearchConfig{
		Query:     "test",
		FreeLeech: false,
	})

	if err != nil {
		t.Errorf("Client.Search() error = %v", err)
		return
	}

	if len(results) != 1 {
		t.Errorf("Client.Search() returned %d results, expected 1", len(results))
		return
	}

	result := results[0]
	if result.Title != "Test Torrent [FL]" {
		t.Errorf("Client.Search() result title = %v, want %v", result.Title, "Test Torrent [FL]")
	}

	if result.Size != 1073741824 {
		t.Errorf("Client.Search() result size = %v, want %v", result.Size, uint64(1073741824))
	}

	if result.Seeders != 10 {
		t.Errorf("Client.Search() result seeders = %v, want %v", result.Seeders, 10)
	}

	if result.Peers != 5 {
		t.Errorf("Client.Search() result peers = %v, want %v", result.Peers, 5)
	}

	if result.DownloadURL != "http://test.com/download" {
		t.Errorf("Client.Search() result download URL = %v, want %v", result.DownloadURL, "http://test.com/download")
	}
}

func TestClient_SearchWithFreeleech(t *testing.T) {
	mockResponse := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:torznab="http://torznab.com/schemas/2015/feed">
  <channel>
    <title>Jackett</title>
    <description>Jackett RSS Feed</description>
    <item>
      <title>Freeleech Torrent [FL]</title>
      <link>http://test.com/download1</link>
      <guid>123456</guid>
      <category>Movies</category>
      <description>A freeleech torrent</description>
      <torznab:attr name="size" value="1073741824"/>
      <torznab:attr name="seeders" value="10"/>
      <torznab:attr name="peers" value="5"/>
    </item>
    <item>
      <title>Regular Torrent</title>
      <link>http://test.com/download2</link>
      <guid>789012</guid>
      <category>Movies</category>
      <description>A regular torrent</description>
      <torznab:attr name="size" value="2147483648"/>
      <torznab:attr name="seeders" value="15"/>
      <torznab:attr name="peers" value="8"/>
    </item>
  </channel>
</rss>`

	server := MockServer("/api/v2.0/indexers/all/results", mockResponse)
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	results, err := client.Search(&SearchConfig{
		Query:     "test",
		FreeLeech: true,
	})

	if err != nil {
		t.Errorf("Client.Search() error = %v", err)
		return
	}

	// Should only return the freeleech torrent
	if len(results) != 1 {
		t.Errorf("Client.Search() with freeleech filter returned %d results, expected 1", len(results))
		return
	}

	result := results[0]
	if result.Title != "Freeleech Torrent [FL]" {
		t.Errorf("Client.Search() freeleech result title = %v, want %v", result.Title, "Freeleech Torrent [FL]")
	}
}

func TestSearchResult_IsFreeleech(t *testing.T) {
	tests := []struct {
		name     string
		result   SearchResult
		expected bool
	}{
		{
			name: "Title with [FL] indicator",
			result: SearchResult{
				Title:       "Test Torrent [FL]",
				Description: "A test torrent",
			},
			expected: true,
		},
		{
			name: "Title with freeleech indicator",
			result: SearchResult{
				Title:       "Test Torrent freeleech",
				Description: "A test torrent",
			},
			expected: true,
		},
		{
			name: "Description with free indicator",
			result: SearchResult{
				Title:       "Test Torrent",
				Description: "This is a free torrent",
			},
			expected: true,
		},
		{
			name: "No freeleech indicators",
			result: SearchResult{
				Title:       "Regular Torrent",
				Description: "A regular torrent",
			},
			expected: false,
		},
		{
			name: "Case insensitive matching",
			result: SearchResult{
				Title:       "Test Torrent [fl]",
				Description: "A test torrent",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsFreeleech(); got != tt.expected {
				t.Errorf("SearchResult.IsFreeleech() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClient_GetIndexers(t *testing.T) {
	// For now, test that it returns empty slice without error
	// This can be expanded once the actual indexer listing is implemented
	server := MockServer("/api/v2.0/indexers", `[]`)
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	indexers, err := client.GetIndexers()

	if err != nil {
		t.Errorf("Client.GetIndexers() error = %v", err)
		return
	}

	if len(indexers) != 0 {
		t.Errorf("Client.GetIndexers() returned %d indexers, expected 0 (placeholder implementation)", len(indexers))
	}
}
