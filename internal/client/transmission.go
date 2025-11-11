package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"opforjellyfin/internal/shared"
)

type TransmissionClient struct {
	config    shared.TorrentClientConfig
	client    *http.Client
	sessionID string
}

type transmissionRequest struct {
	Method    string `json:"method"`
	Arguments any    `json:"arguments,omitempty"`
}

type transmissionResponse struct {
	Result    string `json:"result"`
	Arguments any    `json:"arguments,omitempty"`
}

func NewTransmissionClient(cfg shared.TorrentClientConfig) (*TransmissionClient, error) {
	client := &TransmissionClient{
		config: cfg,
		client: &http.Client{},
	}

	if err := client.TestConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}

func (t *TransmissionClient) makeRequest(req transmissionRequest, resp *transmissionResponse) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", t.config.URL+"/transmission/rpc", bytes.NewReader(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if t.sessionID != "" {
		httpReq.Header.Set("X-Transmission-Session-Id", t.sessionID)
	}

	if t.config.Username != "" {
		httpReq.SetBasicAuth(t.config.Username, t.config.Password)
	}

	httpResp, err := t.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == 409 {
		t.sessionID = httpResp.Header.Get("X-Transmission-Session-Id")
		return t.makeRequest(req, resp)
	}

	return json.NewDecoder(httpResp.Body).Decode(resp)
}

func (t *TransmissionClient) TestConnection() error {
	req := transmissionRequest{
		Method: "session-get",
	}

	var resp transmissionResponse
	return t.makeRequest(req, &resp)
}

func (t *TransmissionClient) AddTorrent(ctx context.Context, torrentURL string, savePath string) (string, error) {
	req := transmissionRequest{
		Method: "torrent-add",
		Arguments: map[string]any{
			"filename":     torrentURL,
			"download-dir": savePath,
		},
	}

	var resp transmissionResponse
	if err := t.makeRequest(req, &resp); err != nil {
		return "", err
	}

	if resp.Result != "success" {
		return "", fmt.Errorf("failed to add torrent: %s", resp.Result)
	}

	argsMap, ok := resp.Arguments.(map[string]any)
	if !ok {
		return "", fmt.Errorf("invalid response")
	}

	torrentAdded, ok := argsMap["torrent-added"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("torrent not added")
	}

	hashString, _ := torrentAdded["hashString"].(string)
	return hashString, nil
}

func (t *TransmissionClient) GetTorrentStatus(torrentID string) (*TorrentStatus, error) {
	req := transmissionRequest{
		Method: "torrent-get",
		Arguments: map[string]any{
			"fields": []string{"id", "name", "status", "percentDone", "downloadedEver", "totalSize", "rateDownload", "rateUpload", "downloadDir", "error", "errorString"},
			"ids":    []string{torrentID},
		},
	}

	var resp transmissionResponse
	if err := t.makeRequest(req, &resp); err != nil {
		return nil, err
	}

	argsMap := resp.Arguments.(map[string]any)
	torrents := argsMap["torrents"].([]any)

	if len(torrents) == 0 {
		return nil, fmt.Errorf("torrent not found")
	}

	torrent := torrents[0].(map[string]any)
	percentDone := torrent["percentDone"].(float64)

	status := &TorrentStatus{
		ID:            fmt.Sprintf("%v", torrent["id"]),
		Name:          torrent["name"].(string),
		Progress:      percentDone * 100,
		Downloaded:    int64(torrent["downloadedEver"].(float64)),
		TotalSize:     int64(torrent["totalSize"].(float64)),
		DownloadSpeed: int64(torrent["rateDownload"].(float64)),
		UploadSpeed:   int64(torrent["rateUpload"].(float64)),
		SavePath:      torrent["downloadDir"].(string),
		IsComplete:    percentDone >= 1.0,
	}

	if errNum, ok := torrent["error"].(float64); ok && errNum != 0 {
		status.Error = torrent["errorString"].(string)
	}

	return status, nil
}

func (t *TransmissionClient) RemoveTorrent(torrentID string, deleteFiles bool) error {
	req := transmissionRequest{
		Method: "torrent-remove",
		Arguments: map[string]any{
			"ids":               []string{torrentID},
			"delete-local-data": deleteFiles,
		},
	}

	var resp transmissionResponse
	return t.makeRequest(req, &resp)
}

func (t *TransmissionClient) PauseTorrent(torrentID string) error {
	req := transmissionRequest{
		Method: "torrent-stop",
		Arguments: map[string]any{
			"ids": []string{torrentID},
		},
	}

	var resp transmissionResponse
	return t.makeRequest(req, &resp)
}

func (t *TransmissionClient) ResumeTorrent(torrentID string) error {
	req := transmissionRequest{
		Method: "torrent-start",
		Arguments: map[string]any{
			"ids": []string{torrentID},
		},
	}

	var resp transmissionResponse
	return t.makeRequest(req, &resp)
}

func (t *TransmissionClient) GetClientInfo() (*ClientInfo, error) {
	req := transmissionRequest{
		Method: "session-get",
	}

	var resp transmissionResponse
	if err := t.makeRequest(req, &resp); err != nil {
		return nil, err
	}

	argsMap := resp.Arguments.(map[string]any)
	version, _ := argsMap["version"].(string)

	return &ClientInfo{
		Version: version,
	}, nil
}
