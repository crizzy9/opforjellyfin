package downloader

import (
	"context"
	"fmt"
	"opforjellyfin/internal/client"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"time"
)

func QueueDownload(entry *shared.TorrentEntry, torrentURL string, cfg shared.Config) error {
	if client.IsInternalClient(cfg.TorrentClient) {
		return queueInternalDownload(entry)
	}

	return queueExternalDownload(entry, torrentURL, cfg)
}

func queueInternalDownload(entry *shared.TorrentEntry) error {
	td := &shared.TorrentDownload{
		Title:        entry.TorrentName,
		TorrentID:    entry.TorrentID,
		FullTitle:    entry.Title,
		Started:      time.Now(),
		ChapterRange: entry.ChapterRange,
		UseExternal:  false,
	}

	shared.SaveTorrentDownload(td)
	logger.Log(false, "Queued internal download: %s", entry.TorrentName)

	return nil
}

func queueExternalDownload(entry *shared.TorrentEntry, torrentURL string, cfg shared.Config) error {
	torrentClient, err := client.NewClient(cfg.TorrentClient)
	if err != nil {
		return fmt.Errorf("failed to create torrent client: %w", err)
	}

	ctx := context.Background()
	savePath := cfg.TargetDir

	hash, err := torrentClient.AddTorrent(ctx, torrentURL, savePath)
	if err != nil {
		return fmt.Errorf("failed to add torrent to client: %w", err)
	}

	td := &shared.TorrentDownload{
		Title:        entry.TorrentName,
		TorrentID:    entry.TorrentID,
		FullTitle:    entry.Title,
		Started:      time.Now(),
		ChapterRange: entry.ChapterRange,
		ExternalHash: hash,
		UseExternal:  true,
	}

	shared.SaveTorrentDownload(td)
	logger.Log(false, "Queued external download: %s (hash: %s)", entry.TorrentName, hash)

	return nil
}

func GetDownloadStatus(td *shared.TorrentDownload, cfg shared.Config) (*client.TorrentStatus, error) {
	if !td.UseExternal {
		return nil, nil
	}

	torrentClient, err := client.NewClient(cfg.TorrentClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create torrent client: %w", err)
	}

	return torrentClient.GetTorrentStatus(td.ExternalHash)
}

func TestConnection(cfg shared.TorrentClientConfig) (string, error) {
	torrentClient, err := client.NewClient(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}

	if err := torrentClient.TestConnection(); err != nil {
		return "", fmt.Errorf("connection test failed: %w", err)
	}

	info, err := torrentClient.GetClientInfo()
	if err != nil {
		return cfg.Type, nil
	}

	return fmt.Sprintf("%s v%s", cfg.Type, info.Version), nil
}
