package qbittorrent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// Client is a qBittorrent client.
type Client struct {
	url    string
	client *http.Client
}

// Torrent is a torrent.
type Torrent struct {
	Name     string `json:"name"`
	Size     uint64 `json:"size"`
	Category string `json:"category"`
	Hash     string `json:"hash"`
	Tracker  string `json:"tracker"`
}

func NewClient(url string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		url: url,
		client: &http.Client{
			Jar: jar,
		},
	}
}

func (c *Client) Login(username, password string) error {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	// set headers
	req, err := http.NewRequest("POST", c.url+"/api/v2/auth/login", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Referer", c.url)
	req.Header.Set("Origin", c.url)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	// check if login was successful
	if resp.Header.Get("Set-Cookie") == "" {
		return fmt.Errorf("login failed")
	}

	return nil
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

func (c *Client) AddTorrentFromURL(urls string, options map[string]string) error {
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
	_, err := io.Copy(part, file)
	if err != nil {
		return err
	}

	for k, v := range options {
		err = writer.WriteField(k, v)
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
