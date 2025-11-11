package client

import (
	"context"
	"fmt"
	"opforjellyfin/internal/shared"
)

type TorrentClient interface {
	TestConnection() error
	AddTorrent(ctx context.Context, torrentURL string, savePath string) (string, error)
	GetTorrentStatus(torrentID string) (*TorrentStatus, error)
	RemoveTorrent(torrentID string, deleteFiles bool) error
	PauseTorrent(torrentID string) error
	ResumeTorrent(torrentID string) error
	GetClientInfo() (*ClientInfo, error)
}

type TorrentStatus struct {
	ID            string
	Name          string
	State         string
	Progress      float64
	Downloaded    int64
	TotalSize     int64
	DownloadSpeed int64
	UploadSpeed   int64
	Seeders       int
	Peers         int
	SavePath      string
	Error         string
	IsComplete    bool
}

type ClientInfo struct {
	Version    string
	FreeSpace  int64
	TotalSpace int64
}

func NewClient(cfg shared.TorrentClientConfig) (TorrentClient, error) {
	if cfg.Type == "" || cfg.Type == "internal" {
		return nil, nil
	}

	switch cfg.Type {
	case "qbittorrent":
		return NewQBittorrentClient(cfg)
	case "deluge":
		return NewDelugeClient(cfg)
	case "transmission":
		return NewTransmissionClient(cfg)
	default:
		return nil, fmt.Errorf("unknown client type: %s", cfg.Type)
	}
}

func IsInternalClient(cfg shared.TorrentClientConfig) bool {
	return cfg.Type == "" || cfg.Type == "internal"
}
