package downloader

import (
	"context"
	"fmt"
	"opforjellyfin/internal/client"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/matcher"
	"opforjellyfin/internal/metadata"
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

	hash, err := torrentClient.AddTorrent(ctx, torrentURL, "")
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
		Imported:     false,
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

func ImportCompletedDownload(td *shared.TorrentDownload, status *client.TorrentStatus, cfg shared.Config) error {
	if td.Imported {
		logger.Log(false, "Skipping import - already imported: %s", td.Title)
		return nil
	}

	if !status.IsComplete {
		return fmt.Errorf("torrent not complete")
	}

	td.SavePath = status.SavePath
	td.PlacementProgress = "🔗 Importing files..."
	shared.SaveTorrentDownload(td)

	logger.Log(true, "🔄 Starting import for: %s from %s", td.Title, status.SavePath)

	index := metadata.LoadMetadataCache()
	if index == nil {
		logger.Log(true, "❌ Failed to import %s: metadata cache not loaded", td.Title)
		return fmt.Errorf("metadata cache not loaded")
	}

	// Process the files and check if any were placed
	matcher.ProcessTorrentFiles(status.SavePath, cfg.TargetDir, td, index)

	// Only mark as imported and placed if files were actually placed
	if len(td.PlacementFull) > 0 {
		td.Imported = true
		td.Placed = true
		td.Done = true
		logger.Log(true, "✅ Successfully imported %d file(s) for: %s", len(td.PlacementFull), td.Title)
	} else {
		logger.Log(true, "⚠️  Warning: No files were placed for: %s", td.Title)
		td.PlacementProgress = "⚠️ No files placed"
	}

	shared.SaveTorrentDownload(td)
	return nil
}
