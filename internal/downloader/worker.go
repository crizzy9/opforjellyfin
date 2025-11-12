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

	hasImports := false
	for _, td := range downloads {
		if !td.UseExternal || td.Imported {
			continue
		}

		status, err := GetDownloadStatus(td, w.cfg)
		if err != nil {
			logger.Log(false, "Error checking download status for %s: %v", td.Title, err)
			continue
		}

		if status == nil || !status.IsComplete {
			continue
		}

		logger.Log(false, "Download completed, importing: %s", td.Title)

		if err := ImportCompletedDownload(td, status, w.cfg); err != nil {
			logger.Log(true, "Failed to import download %s: %v", td.Title, err)
			td.PlacementProgress = "❌ Import failed"
			shared.SaveTorrentDownload(td)
		} else {
			logger.Log(false, "Successfully imported: %s", td.Title)
			hasImports = true
		}
	}

	if hasImports && w.onImportCallback != nil {
		w.onImportCallback()
	}
}
