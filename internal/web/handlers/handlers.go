package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"opforjellyfin/internal/downloader"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/scraper"
	"opforjellyfin/internal/shared"
	"strconv"
	"strings"
)

func HandleIndex(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/episodes", http.StatusSeeOther)
	}
}

func HandleEpisodes(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"Page": "episodes",
		}
		if err := templates.ExecuteTemplate(w, "episodes.html", data); err != nil {
			logger.Log(true, "Template error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func HandleActivity(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"Page": "activity",
		}
		if err := templates.ExecuteTemplate(w, "activity.html", data); err != nil {
			logger.Log(true, "Template error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func HandleSettings(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := shared.LoadConfig()
		data := map[string]any{
			"Page":   "settings",
			"Config": cfg,
		}
		if err := templates.ExecuteTemplate(w, "settings.html", data); err != nil {
			logger.Log(true, "Template error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func HandleSystem(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"Page": "system",
		}
		if err := templates.ExecuteTemplate(w, "system.html", data); err != nil {
			logger.Log(true, "Template error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func APIListEpisodes(w http.ResponseWriter, r *http.Request) {
	cfg := shared.LoadConfig()
	if cfg.Source.BaseURL == "" {
		http.Error(w, "Please run sync first", http.StatusBadRequest)
		return
	}

	torrents, err := scraper.FetchTorrents(cfg)
	if err != nil {
		logger.Log(true, "Error fetching torrents: %v", err)
		http.Error(w, "Failed to fetch torrents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(torrents)
}

func APISearchEpisodes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	titleFilter := r.URL.Query().Get("title")
	rangeFilter := r.URL.Query().Get("range")

	cfg := shared.LoadConfig()
	if cfg.Source.BaseURL == "" {
		http.Error(w, "Please run sync first", http.StatusBadRequest)
		return
	}

	torrents, err := scraper.FetchTorrents(cfg)
	if err != nil {
		http.Error(w, "Failed to fetch torrents", http.StatusInternalServerError)
		return
	}

	var filtered []shared.TorrentEntry
	for _, t := range torrents {
		if query != "" && !strings.Contains(strings.ToLower(t.TorrentName), strings.ToLower(query)) {
			continue
		}
		if titleFilter != "" && !strings.Contains(strings.ToLower(t.TorrentName), strings.ToLower(titleFilter)) {
			continue
		}
		if rangeFilter != "" && !strings.Contains(t.ChapterRange, rangeFilter) {
			continue
		}
		filtered = append(filtered, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}

func APIDownloadEpisode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	downloadKeyStr := r.FormValue("downloadKey")
	downloadKey, err := strconv.Atoi(downloadKeyStr)
	if err != nil {
		http.Error(w, "Invalid download key", http.StatusBadRequest)
		return
	}

	cfg := shared.LoadConfig()
	if cfg.TargetDir == "" {
		http.Error(w, "Target directory not set", http.StatusBadRequest)
		return
	}

	torrents, err := scraper.FetchTorrents(cfg)
	if err != nil {
		http.Error(w, "Failed to fetch torrents", http.StatusInternalServerError)
		return
	}

	var match *shared.TorrentEntry
	for _, t := range torrents {
		if t.DownloadKey == downloadKey {
			if match == nil || t.Seeders > match.Seeders {
				tmp := t
				match = &tmp
			}
		}
	}

	if match == nil {
		http.Error(w, "Torrent not found", http.StatusNotFound)
		return
	}

	torrentURL := fmt.Sprintf("%s/download/%d.torrent", cfg.Source.BaseURL, match.TorrentID)

	if err := downloader.QueueDownload(match, torrentURL, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to queue download: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": fmt.Sprintf("Added %s to download queue", match.TorrentName),
		"torrent": match,
	})
}

func APIUpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	cfg := shared.LoadConfig()

	if targetDir := r.FormValue("targetDir"); targetDir != "" {
		cfg.TargetDir = targetDir
	}

	if clientType := r.FormValue("clientType"); clientType != "" {
		cfg.TorrentClient.Type = clientType
	}

	if clientURL := r.FormValue("clientUrl"); clientURL != "" {
		cfg.TorrentClient.URL = clientURL
	}

	if clientUsername := r.FormValue("clientUsername"); clientUsername != "" {
		cfg.TorrentClient.Username = clientUsername
	}

	if clientPassword := r.FormValue("clientPassword"); clientPassword != "" {
		cfg.TorrentClient.Password = clientPassword
	}

	shared.SaveConfig(cfg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Settings updated successfully",
	})
}

func APITestClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := shared.LoadConfig()

	if cfg.TorrentClient.Type == "" || cfg.TorrentClient.Type == "internal" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"message": "Using internal torrent client",
		})
		return
	}

	testClient, err := downloader.TestConnection(cfg.TorrentClient)
	if err != nil {
		http.Error(w, fmt.Sprintf("Connection failed: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": fmt.Sprintf("Successfully connected to %s", testClient),
	})
}

func APISync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := shared.LoadConfig()
	if cfg.TargetDir == "" {
		http.Error(w, "Target directory not set", http.StatusBadRequest)
		return
	}

	if err := metadata.SyncMetadata(cfg.TargetDir, cfg); err != nil {
		logger.Log(true, "Sync failed: %v", err)
		http.Error(w, fmt.Sprintf("Sync failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Metadata synced successfully",
	})
}

func APIActivityStatus(w http.ResponseWriter, r *http.Request) {
	downloads := shared.GetActiveDownloads()
	cfg := shared.LoadConfig()

	for _, dl := range downloads {
		if dl.UseExternal {
			status, err := downloader.GetDownloadStatus(dl, cfg)
			if err == nil && status != nil {
				dl.Progress = status.Downloaded
				dl.TotalSize = status.TotalSize
				dl.Done = status.IsComplete
				if status.IsComplete && !dl.Placed {
					dl.PlacementProgress = "✅ Complete - Ready to organize"
				} else if !dl.Done {
					dl.PlacementProgress = fmt.Sprintf("⏳ Downloading... %.1f%%", status.Progress)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"downloads": downloads,
		"count":     len(downloads),
	})
}
