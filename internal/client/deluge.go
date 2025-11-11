package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"opforjellyfin/internal/shared"
)

type DelugeClient struct {
	config    shared.TorrentClientConfig
	client    *http.Client
	sessionID string
}

type delugeRequest struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
	ID     int    `json:"id"`
}

type delugeResponse struct {
	Result any `json:"result"`
	Error  any `json:"error"`
	ID     int `json:"id"`
}

func NewDelugeClient(cfg shared.TorrentClientConfig) (*DelugeClient, error) {
	client := &DelugeClient{
		config: cfg,
		client: &http.Client{},
	}

	if err := client.login(); err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return client, nil
}

func (d *DelugeClient) login() error {
	req := delugeRequest{
		Method: "auth.login",
		Params: []any{d.config.Password},
		ID:     1,
	}

	var resp delugeResponse
	if err := d.makeRequest(req, &resp); err != nil {
		return err
	}

	if result, ok := resp.Result.(bool); !ok || !result {
		return fmt.Errorf("login failed")
	}

	return nil
}

func (d *DelugeClient) makeRequest(req delugeRequest, resp *delugeResponse) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", d.config.URL+"/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := d.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	return json.NewDecoder(httpResp.Body).Decode(resp)
}

func (d *DelugeClient) TestConnection() error {
	req := delugeRequest{
		Method: "daemon.info",
		Params: []any{},
		ID:     1,
	}

	var resp delugeResponse
	return d.makeRequest(req, &resp)
}

func (d *DelugeClient) AddTorrent(ctx context.Context, torrentURL string, savePath string) (string, error) {
	req := delugeRequest{
		Method: "core.add_torrent_url",
		Params: []any{torrentURL, map[string]any{"download_location": savePath}},
		ID:     1,
	}

	var resp delugeResponse
	if err := d.makeRequest(req, &resp); err != nil {
		return "", err
	}

	if hash, ok := resp.Result.(string); ok {
		return hash, nil
	}

	return "", fmt.Errorf("failed to add torrent")
}

func (d *DelugeClient) GetTorrentStatus(torrentID string) (*TorrentStatus, error) {
	req := delugeRequest{
		Method: "core.get_torrent_status",
		Params: []any{torrentID, []string{"name", "state", "progress", "total_done", "total_size", "download_payload_rate", "upload_payload_rate", "save_path"}},
		ID:     1,
	}

	var resp delugeResponse
	if err := d.makeRequest(req, &resp); err != nil {
		return nil, err
	}

	statusMap, ok := resp.Result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid status response")
	}

	progress, _ := statusMap["progress"].(float64)

	return &TorrentStatus{
		ID:            torrentID,
		Name:          statusMap["name"].(string),
		State:         statusMap["state"].(string),
		Progress:      progress,
		Downloaded:    int64(statusMap["total_done"].(float64)),
		TotalSize:     int64(statusMap["total_size"].(float64)),
		DownloadSpeed: int64(statusMap["download_payload_rate"].(float64)),
		UploadSpeed:   int64(statusMap["upload_payload_rate"].(float64)),
		SavePath:      statusMap["save_path"].(string),
		IsComplete:    progress >= 100,
	}, nil
}

func (d *DelugeClient) RemoveTorrent(torrentID string, deleteFiles bool) error {
	req := delugeRequest{
		Method: "core.remove_torrent",
		Params: []any{torrentID, deleteFiles},
		ID:     1,
	}

	var resp delugeResponse
	return d.makeRequest(req, &resp)
}

func (d *DelugeClient) PauseTorrent(torrentID string) error {
	req := delugeRequest{
		Method: "core.pause_torrent",
		Params: []any{[]string{torrentID}},
		ID:     1,
	}

	var resp delugeResponse
	return d.makeRequest(req, &resp)
}

func (d *DelugeClient) ResumeTorrent(torrentID string) error {
	req := delugeRequest{
		Method: "core.resume_torrent",
		Params: []any{[]string{torrentID}},
		ID:     1,
	}

	var resp delugeResponse
	return d.makeRequest(req, &resp)
}

func (d *DelugeClient) GetClientInfo() (*ClientInfo, error) {
	req := delugeRequest{
		Method: "daemon.info",
		Params: []any{},
		ID:     1,
	}

	var resp delugeResponse
	if err := d.makeRequest(req, &resp); err != nil {
		return nil, err
	}

	return &ClientInfo{
		Version: "Deluge",
	}, nil
}
