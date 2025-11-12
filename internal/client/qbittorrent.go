package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"opforjellyfin/internal/shared"
	"strings"
	"time"
)

type QBittorrentClient struct {
	config shared.TorrentClientConfig
	client *http.Client
	cookie string
}

func NewQBittorrentClient(cfg shared.TorrentClientConfig) (*QBittorrentClient, error) {
	jar, _ := cookiejar.New(nil)
	client := &QBittorrentClient{
		config: cfg,
		client: &http.Client{
			Jar: jar,
		},
	}

	if err := client.login(); err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return client, nil
}

func (q *QBittorrentClient) login() error {
	data := url.Values{}
	data.Set("username", q.config.Username)
	data.Set("password", q.config.Password)

	resp, err := q.client.Post(
		q.config.URL+"/api/v2/auth/login",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (q *QBittorrentClient) TestConnection() error {
	resp, err := q.client.Get(q.config.URL + "/api/v2/app/version")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection test failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (q *QBittorrentClient) AddTorrent(ctx context.Context, torrentURL string, savePath string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("urls", torrentURL)
	if savePath != "" {
		writer.WriteField("savepath", savePath)
	}
	writer.WriteField("category", "OnePace")

	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", q.config.URL+"/api/v2/torrents/add", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := q.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to add torrent: %s", string(bodyBytes))
	}

	hash, err := q.getTorrentHashByURL(torrentURL)
	if err != nil {
		return "", fmt.Errorf("torrent added but failed to get hash: %w", err)
	}

	return hash, nil
}

func (q *QBittorrentClient) getTorrentHashByURL(torrentURL string) (string, error) {
	for i := 0; i < 10; i++ {
		resp, err := q.client.Get(q.config.URL + "/api/v2/torrents/info?category=OnePace")
		if err != nil {
			return "", err
		}

		var torrents []map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
			resp.Body.Close()
			return "", err
		}
		resp.Body.Close()

		for _, t := range torrents {
			hash, _ := t["hash"].(string)
			magnetURI, _ := t["magnet_uri"].(string)

			if strings.Contains(magnetURI, torrentURL) || strings.Contains(torrentURL, hash) {
				return hash, nil
			}
		}

		if i < 9 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return "", fmt.Errorf("could not find torrent hash after adding")
}

func (q *QBittorrentClient) GetTorrentStatus(torrentID string) (*TorrentStatus, error) {
	resp, err := q.client.Get(q.config.URL + "/api/v2/torrents/info?hashes=" + torrentID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var torrents []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, err
	}

	if len(torrents) == 0 {
		return nil, fmt.Errorf("torrent not found: %s", torrentID)
	}

	t := torrents[0]
	hash, _ := t["hash"].(string)
	name, _ := t["name"].(string)
	state, _ := t["state"].(string)
	progress, _ := t["progress"].(float64)
	downloaded, _ := t["downloaded"].(float64)
	totalSize, _ := t["size"].(float64)
	dlspeed, _ := t["dlspeed"].(float64)
	upspeed, _ := t["upspeed"].(float64)
	savePath, _ := t["save_path"].(string)

	return &TorrentStatus{
		ID:            hash,
		Name:          name,
		State:         state,
		Progress:      progress * 100,
		Downloaded:    int64(downloaded),
		TotalSize:     int64(totalSize),
		DownloadSpeed: int64(dlspeed),
		UploadSpeed:   int64(upspeed),
		SavePath:      savePath,
		IsComplete:    progress >= 1.0,
	}, nil
}

func (q *QBittorrentClient) RemoveTorrent(torrentID string, deleteFiles bool) error {
	data := url.Values{}
	data.Set("hashes", torrentID)
	data.Set("deleteFiles", fmt.Sprintf("%t", deleteFiles))

	resp, err := q.client.Post(
		q.config.URL+"/api/v2/torrents/delete",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (q *QBittorrentClient) PauseTorrent(torrentID string) error {
	data := url.Values{}
	data.Set("hashes", torrentID)

	resp, err := q.client.Post(
		q.config.URL+"/api/v2/torrents/pause",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (q *QBittorrentClient) ResumeTorrent(torrentID string) error {
	data := url.Values{}
	data.Set("hashes", torrentID)

	resp, err := q.client.Post(
		q.config.URL+"/api/v2/torrents/resume",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (q *QBittorrentClient) GetClientInfo() (*ClientInfo, error) {
	resp, err := q.client.Get(q.config.URL + "/api/v2/app/version")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	versionBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &ClientInfo{
		Version: strings.TrimSpace(string(versionBytes)),
	}, nil
}
