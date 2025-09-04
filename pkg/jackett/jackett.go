/*
Package jackett provides a client for the Jackett API.
Jackett docs: https://github.com/Jackett/Jackett/wiki/Jackett-API
*/
package jackett

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Client struct {
	url    string
	apiKey string
	client *http.Client
}

// Indexer represents a Jackett indexer (Tracker).
type Indexer struct {
	ID          string `xml:"id,attr"`
	Type        string `xml:"type,attr"`
	Configured  string `xml:"configured,attr"`
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Language    string `xml:"language"`
}

// SearchResult represents a search result from Jackett.
type SearchResult struct {
	Title                string  `xml:"title"`
	Link                 string  `xml:"link"`
	Description          string  `xml:"description"`
	GUID                 string  `xml:"guid"`
	Category             string  `xml:"category"`
	Size                 uint64  `xml:"size"`
	Seeders              int     `xml:"seeders"`
	Peers                int     `xml:"peers"`
	DownloadVolumeFactor float64 // Torznab field: 0 = freeleech, 1 = normal download
	HasVolumeFactor      bool    // Track if DownloadVolumeFactor was explicitly set
	DownloadURL          string
	IndexerID            string
}

// SearchConfig represents the search configuration for Jackett.
type SearchConfig struct {
	Indexers   []string // Indexer IDs to search (empty means all)
	Query      string   // Search query
	Categories []int    // Category IDs to search
	Limit      int      // Limit the number of results
	FreeLeech  bool     // Attempt to filter for freeleech (heuristic-based)
}

// TorznabResponse represents the XML response from Jackett API
type TorznabResponse struct {
	XMLName xml.Name       `xml:"rss"`
	Channel TorznabChannel `xml:"channel"`
}

type TorznabChannel struct {
	Title       string        `xml:"title"`
	Description string        `xml:"description"`
	Items       []TorznabItem `xml:"item"`
}

type TorznabItem struct {
	Title       string        `xml:"title"`
	Link        string        `xml:"link"`
	GUID        string        `xml:"guid"`
	Category    string        `xml:"category"`
	Description string        `xml:"description"`
	Attributes  []TorznabAttr `xml:"http://torznab.com/schemas/2015/feed attr"`
}

type TorznabAttr struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

func NewClient(url, apiKey string) *Client {
	return &Client{
		url:    strings.TrimSuffix(url, "/"),
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (c *Client) GetAPIVersion() (string, error) {
	// Jackett doesn't have a version endpoint like Prowlarr
	// Return a placeholder version for compatibility
	return "jackett-api", nil
}

func (c *Client) authenticateRequest(path string) (*http.Request, error) {
	fullURL := c.url + path
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// Add API key to query parameters
	q := req.URL.Query()
	q.Add("apikey", c.apiKey)
	req.URL.RawQuery = q.Encode()

	return req, nil
}

func (c *Client) GetIndexers() ([]Indexer, error) {
	req, err := c.authenticateRequest("/api/v2.0/indexers")
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	var indexers []Indexer
	// Jackett returns indexer list in a different format, we'll need to parse it
	// For now, return empty slice - this can be implemented based on actual Jackett response format
	return indexers, nil
}

func (c *Client) Search(config *SearchConfig) ([]SearchResult, error) {
	// Determine the search endpoint
	var endpoint string
	if len(config.Indexers) == 1 {
		endpoint = fmt.Sprintf("/api/v2.0/indexers/%s/results", config.Indexers[0])
	} else {
		endpoint = "/api/v2.0/indexers/all/results"
	}

	req, err := c.authenticateRequest(endpoint)
	if err != nil {
		return nil, err
	}

	// Add search parameters
	q := req.URL.Query()
	if config.Query != "" {
		q.Add("Query", config.Query)
	}
	if len(config.Categories) > 0 {
		for _, cat := range config.Categories {
			q.Add("Category", strconv.Itoa(cat))
		}
	}
	if config.Limit > 0 {
		q.Add("limit", strconv.Itoa(config.Limit))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	var torznab TorznabResponse
	if err := xml.NewDecoder(resp.Body).Decode(&torznab); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, item := range torznab.Channel.Items {
		result := SearchResult{
			Title:       item.Title,
			Link:        item.Link,
			GUID:        item.GUID,
			Category:    item.Category,
			Description: item.Description,
			DownloadURL: item.Link,              // Jackett usually provides direct download links
			IndexerID:   extractIndexerID(item), // Extract indexer ID from the item
		}

		// Parse attributes for additional data
		for _, attr := range item.Attributes {
			switch attr.Name {
			case "size":
				if size, err := strconv.ParseUint(attr.Value, 10, 64); err == nil {
					result.Size = size
				}
			case "seeders":
				if seeders, err := strconv.Atoi(attr.Value); err == nil {
					result.Seeders = seeders
				}
			case "peers":
				if peers, err := strconv.Atoi(attr.Value); err == nil {
					result.Peers = peers
				}
			case "downloadvolumefactor":
				if factor, err := strconv.ParseFloat(attr.Value, 64); err == nil {
					result.DownloadVolumeFactor = factor
					result.HasVolumeFactor = true
				}
			}
		}

		results = append(results, result)
	}

	// Apply freeleech filtering if requested
	if config.FreeLeech {
		freelechResults := make([]SearchResult, 0)
		for _, result := range results {
			if result.IsFreeleech() {
				freelechResults = append(freelechResults, result)
			}
		}
		return freelechResults, nil
	}

	return results, nil
}

// IsFreeleech determines if a torrent is freeleech
// First checks DownloadVolumeFactor (Torznab standard) if available, then falls back to heuristics
func (s *SearchResult) IsFreeleech() bool {
	// Primary method: Check DownloadVolumeFactor if explicitly set
	// 0 means freeleech, 1 means normal download
	if s.HasVolumeFactor && s.DownloadVolumeFactor == 0 {
		return true
	}

	// If DownloadVolumeFactor is explicitly set to non-zero, it's not freeleech
	if s.HasVolumeFactor && s.DownloadVolumeFactor > 0 {
		return false
	}

	// Fallback to heuristics if DownloadVolumeFactor is not available
	// Check title for common freeleech indicators
	title := strings.ToLower(s.Title)
	freelechIndicators := []string{
		"freeleech", "free leech", "fl", "free", "0%",
		"[fl]", "(fl)", "[free]", "(free)",
	}

	for _, indicator := range freelechIndicators {
		if strings.Contains(title, indicator) {
			return true
		}
	}

	// Check description for freeleech indicators
	desc := strings.ToLower(s.Description)
	for _, indicator := range freelechIndicators {
		if strings.Contains(desc, indicator) {
			return true
		}
	}

	return false
}

// extractIndexerID attempts to extract indexer ID from torznab item
// This is a best-effort approach since Jackett's response format may vary
func extractIndexerID(item TorznabItem) string {
	// Try to extract from guid or other attributes
	// For now, return empty string as indexer matching will be handled differently for Jackett
	return ""
}
