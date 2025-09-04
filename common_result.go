package yoink

import (
	"github.com/mrmarble/yoink/pkg/jackett"
	"github.com/mrmarble/yoink/pkg/prowlarr"
)

// CommonSearchResult is a wrapper that provides a common interface for both Prowlarr and Jackett search results
type CommonSearchResult struct {
	prowlarrResult *prowlarr.SearchResult
	jackettResult  *jackett.SearchResult
}

// NewProwlarrResult wraps a Prowlarr search result
func NewProwlarrResult(result prowlarr.SearchResult) CommonSearchResult {
	return CommonSearchResult{
		prowlarrResult: &result,
	}
}

// NewJackettResult wraps a Jackett search result  
func NewJackettResult(result jackett.SearchResult) CommonSearchResult {
	return CommonSearchResult{
		jackettResult: &result,
	}
}

// GetTitle returns the torrent title
func (r *CommonSearchResult) GetTitle() string {
	if r.prowlarrResult != nil {
		return r.prowlarrResult.Title
	}
	return r.jackettResult.Title
}

// GetSize returns the torrent size in bytes
func (r *CommonSearchResult) GetSize() uint64 {
	if r.prowlarrResult != nil {
		return uint64(r.prowlarrResult.Size)
	}
	return r.jackettResult.Size
}

// GetSeeders returns the number of seeders
func (r *CommonSearchResult) GetSeeders() int {
	if r.prowlarrResult != nil {
		return r.prowlarrResult.Seeders
	}
	return r.jackettResult.Seeders
}

// GetLeechers returns the number of leechers
func (r *CommonSearchResult) GetLeechers() int {
	if r.prowlarrResult != nil {
		return r.prowlarrResult.Leechers
	}
	return r.jackettResult.Peers
}

// GetDownloadURL returns the download URL
func (r *CommonSearchResult) GetDownloadURL() string {
	if r.prowlarrResult != nil {
		return r.prowlarrResult.DownloadURL
	}
	return r.jackettResult.DownloadURL
}

// GetIndexerID returns the indexer ID
func (r *CommonSearchResult) GetIndexerID() interface{} {
	if r.prowlarrResult != nil {
		return r.prowlarrResult.IndexerID
	}
	return r.jackettResult.IndexerID
}

// IsFreeleech returns whether the torrent is freeleech
func (r *CommonSearchResult) IsFreeleech() bool {
	if r.prowlarrResult != nil {
		return r.prowlarrResult.IsFreeleech()
	}
	return r.jackettResult.IsFreeleech()
}

// GetProwlarrResult returns the underlying Prowlarr result (for compatibility with existing code)
func (r *CommonSearchResult) GetProwlarrResult() *prowlarr.SearchResult {
	return r.prowlarrResult
}

// GetJackettResult returns the underlying Jackett result  
func (r *CommonSearchResult) GetJackettResult() *jackett.SearchResult {
	return r.jackettResult
}