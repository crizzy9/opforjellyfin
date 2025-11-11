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
	writer.WriteField("savepath", savePath)
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

	return torrentURL, nil
}

func (q *QBittorrentClient) GetTorrentStatus(torrentID string) (*TorrentStatus, error) {
	resp, err := q.client.Get(q.config.URL + "/api/v2/torrents/info")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var torrents []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, err
	}

	for _, t := range torrents {
		hash, _ := t["hash"].(string)
		if hash == torrentID || strings.Contains(t["name"].(string), torrentID) {
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
	}

	return nil, fmt.Errorf("torrent not found")
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
