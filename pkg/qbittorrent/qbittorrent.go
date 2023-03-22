package qbittorrent

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	url    string
	client *http.Client
}

type Torrent struct {
	Name     string `json:"name"`
	Size     uint64 `json:"size"`
	Category string `json:"category"`
}

func NewClient(url string) *Client {
	return &Client{
		url:    url,
		client: &http.Client{},
	}
}

func (c *Client) GetTorrents() ([]Torrent, error) {
	resp, err := c.client.Get(c.url + "/api/v2/torrents/info")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var torrents []Torrent
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, err
	}

	return torrents, nil
}

func (c *Client) AddTorrentFromUrl(urls string, options map[string]string) error {
	data := url.Values{}
	data.Set("urls", urls)
	for k, v := range options {
		data.Set(k, v)
	}
	resp, err := c.client.Post(c.url+"/api/v2/torrents/add", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	return nil
}

func (c *Client) AddTorrentFromBuffer(file *bytes.Buffer, fileName string, options map[string]string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("torrents", fileName)
	io.Copy(part, file)

	for k, v := range options {
		err := writer.WriteField(k, v)
		if err != nil {
			return err
		}
	}
	writer.Close()

	resp, err := c.client.Post(c.url+"/api/v2/torrents/add", writer.FormDataContentType(), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	return nil
}
