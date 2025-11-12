package downloader

import (
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"time"
)

type Worker struct {
	cfg              shared.Config
	stopChan         chan struct{}
	pollInterval     time.Duration
	onImportCallback func()
}

func NewWorker(cfg shared.Config, onImport func()) *Worker {
	return &Worker{
		cfg:              cfg,
		stopChan:         make(chan struct{}),
		pollInterval:     30 * time.Second,
		onImportCallback: onImport,
	}
}

func (w *Worker) Start() {
	logger.Log(false, "Starting download worker (polling every %v)", w.pollInterval)
	go w.run()
}

func (w *Worker) Stop() {
	close(w.stopChan)
}

func (w *Worker) run() {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkAndImportDownloads()
		case <-w.stopChan:
			logger.Log(false, "Download worker stopped")
			return
		}
	}
}

func (w *Worker) checkAndImportDownloads() {
	downloads := shared.GetActiveDownloads()

	if len(downloads) > 0 {
		logger.Log(false, "Worker: Checking %d active downloads", len(downloads))
	}

	hasImports := false
	for _, td := range downloads {
		if !td.UseExternal {
			logger.Log(false, "Worker: Skipping internal download: %s", td.Title)
			continue
		}

		if td.Imported {
			logger.Log(false, "Worker: Already imported: %s", td.Title)
			continue
		}

		logger.Log(false, "Worker: Checking status for %s (hash: %s)", td.Title, td.ExternalHash)
		status, err := GetDownloadStatus(td, w.cfg)
		if err != nil {
			logger.Log(true, "Worker: Error checking download status for %s: %v", td.Title, err)
			continue
		}

		if status == nil {
			logger.Log(false, "Worker: No status returned for %s", td.Title)
			continue
		}

		logger.Log(false, "Worker: %s - Progress: %.1f%%, Complete: %v", td.Title, status.Progress, status.IsComplete)

		if !status.IsComplete {
			continue
		}

		logger.Log(true, "Worker: Download completed, importing: %s from %s", td.Title, status.SavePath)

		if err := ImportCompletedDownload(td, status, w.cfg); err != nil {
			logger.Log(true, "Worker: Failed to import download %s: %v", td.Title, err)
			td.PlacementProgress = "❌ Import failed"
			shared.SaveTorrentDownload(td)
		} else {
			logger.Log(true, "Worker: Successfully imported: %s", td.Title)
			hasImports = true
		}
	}

	if hasImports && w.onImportCallback != nil {
		w.onImportCallback()
	}
}
