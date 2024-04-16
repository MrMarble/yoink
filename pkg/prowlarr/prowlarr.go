/*
Prowlarr API Client.
Prowlarr docs: https://prowlarr.com/docs/api
*/
package prowlarr

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	url    string
	apiKey string
	client *http.Client
}

// Indexer represents a Prowlarr indexer (Tracker).
type Indexer struct {
	ID          int      `json:"id"`
	Protocol    string   `json:"protocol"`
	Name        string   `json:"name"`
	Enable      bool     `json:"enable"`
	IndexerUrls []string `json:"indexerUrls"`
}

// SearchResult represents a search result from Prowlarr.
type SearchResult struct {
	Size         uint     `json:"size"`
	IndexerID    int      `json:"indexerId"`
	Title        string   `json:"title"`
	FileName     string   `json:"fileName"`
	DownloadURL  string   `json:"downloadUrl"`
	IndexerFlags []string `json:"indexerFlags"`
	Seeders      int      `json:"seeders"`
	Leechers     int      `json:"leechers"`
	Protocol     string   `json:"protocol"`
}

// SearchConfig represents the search configuration.
type SearchConfig struct {
	Indexers  []int // Indexer IDs
	Limit     int   // Limit the number of results. Currently not working in Prowlarr
	Offset    int   // Offset the number of results. Currently not working in Prowlarr
	FreeLeech bool  // Only return freeleech torrents
}

func NewClient(url, apiKey string) *Client {
	return &Client{
		url:    url,
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (c *Client) GetAPIVersion() (string, error) {
	req, err := c.authenticateRequest("/api")
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	var status struct {
		Version string `json:"current"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return "", err
	}

	return status.Version, nil
}

func (c *Client) authenticateRequest(path string) (*http.Request, error) {
	req, err := http.NewRequest("GET", c.url+path, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Api-Key", c.apiKey)

	return req, nil
}

func (c *Client) GetIndexers() ([]Indexer, error) {
	req, err := c.authenticateRequest("/api/v1/indexer")
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var indexers []Indexer
	if err := json.NewDecoder(resp.Body).Decode(&indexers); err != nil {
		return nil, err
	}

	return indexers, nil
}

func (c *Client) generateQueryString(config *SearchConfig) string {
	query := ""
	if len(config.Indexers) > 0 {
		indexersID := ""
		for _, v := range config.Indexers {
			indexersID += fmt.Sprintf("&indexerIds=%d", v)
		}
		query += indexersID
	}

	if config.Limit > 0 {
		query += fmt.Sprintf("&limit=%d", config.Limit)
	}

	if config.Offset > 0 {
		query += fmt.Sprintf("&offset=%d", config.Offset)
	}

	return query
}

// Search performs a search on Prowlarr.
func (c *Client) Search(config *SearchConfig) ([]SearchResult, error) {
	url := "/api/v1/search?type=search" + c.generateQueryString(config)

	req, err := c.authenticateRequest(url)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var results []SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}
	if config.FreeLeech {
		freeleechResults := make([]SearchResult, 0)
		for _, result := range results {
			if result.IsFreeleech() {
				freeleechResults = append(freeleechResults, result)
			}
		}

		return freeleechResults, nil
	}

	return results, nil
}

// IsFreeleech returns true if the search result is freeleech.
func (s *SearchResult) IsFreeleech() bool {
	for _, v := range s.IndexerFlags {
		if v == "freeleech" {
			return true
		}
	}

	return false
}
