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
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	arcsCache      []ArcStatus
	arcsCacheMutex sync.RWMutex
	arcsCacheTime  time.Time
	arcsCacheTTL   = 5 * time.Minute

	arcDetailsCache      map[string]map[string]any
	arcDetailsCacheMutex sync.RWMutex
	arcDetailsCacheTime  map[string]time.Time
)

func init() {
	arcDetailsCache = make(map[string]map[string]any)
	arcDetailsCacheTime = make(map[string]time.Time)
}

func InvalidateArcsCache() {
	arcsCacheMutex.Lock()
	arcDetailsCacheMutex.Lock()
	defer arcsCacheMutex.Unlock()
	defer arcDetailsCacheMutex.Unlock()

	arcsCache = nil
	arcsCacheTime = time.Time{}
	arcDetailsCache = make(map[string]map[string]any)
	arcDetailsCacheTime = make(map[string]time.Time)
}

func HandleIndex(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/arcs", http.StatusSeeOther)
	}
}

func HandleArcs(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"Page": "arcs",
		}
		if err := templates.ExecuteTemplate(w, "base", data); err != nil {
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
		if err := templates.ExecuteTemplate(w, "base", data); err != nil {
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
		if err := templates.ExecuteTemplate(w, "base", data); err != nil {
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
		if err := templates.ExecuteTemplate(w, "base", data); err != nil {
			logger.Log(true, "Template error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

type ArcStatus struct {
	Name         string `json:"name"`
	SeasonKey    string `json:"seasonKey"`
	SeasonNumber int    `json:"seasonNumber"`
	ChapterRange string `json:"chapterRange"`
	HasMetadata  bool   `json:"hasMetadata"`
	VideoStatus  int    `json:"videoStatus"`
	EpisodeCount int    `json:"episodeCount"`
	DownloadKey  int    `json:"downloadKey"`
}

type EpisodeStatus struct {
	Title        string `json:"title"`
	ChapterRange string `json:"chapterRange"`
	HasVideo     bool   `json:"hasVideo"`
	DownloadKey  int    `json:"downloadKey"`
}

func APIListArcs(w http.ResponseWriter, r *http.Request) {
	forceRefresh := r.URL.Query().Get("refresh") == "true"

	arcsCacheMutex.RLock()
	if !forceRefresh && arcsCache != nil && time.Since(arcsCacheTime) < arcsCacheTTL {
		cachedArcs := arcsCache
		arcsCacheMutex.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cachedArcs)
		return
	}
	arcsCacheMutex.RUnlock()

	cfg := shared.LoadConfig()
	if cfg.TargetDir == "" {
		http.Error(w, "Please set target directory first", http.StatusBadRequest)
		return
	}

	index := metadata.LoadMetadataCache()
	if index == nil || len(index.Seasons) == 0 {
		http.Error(w, "Please run sync first", http.StatusBadRequest)
		return
	}

	torrents, err := scraper.FetchTorrents(cfg)
	if err != nil {
		logger.Log(true, "Error fetching torrents: %v", err)
		http.Error(w, "Failed to fetch torrents", http.StatusInternalServerError)
		return
	}

	downloadKeyMap := make(map[string]int)
	for _, t := range torrents {
		if t.DownloadKey > 0 {
			downloadKeyMap[t.ChapterRange] = t.DownloadKey
		}
	}

	var arcs []ArcStatus
	for seasonKey, season := range index.Seasons {
		if season.Range == "" {
			continue
		}

		seasonNum := season.SeasonNumber
		if seasonNum == 0 && seasonKey != "Specials" {
			var snum int
			fmt.Sscanf(seasonKey, "Season %d", &snum)
			seasonNum = snum
		}

		arc := ArcStatus{
			Name:         season.Name,
			SeasonKey:    seasonKey,
			SeasonNumber: seasonNum,
			ChapterRange: season.Range,
			HasMetadata:  true,
			VideoStatus:  metadata.HaveVideoStatus(season.Range),
			EpisodeCount: len(season.EpisodeRange),
			DownloadKey:  downloadKeyMap[season.Range],
		}
		arcs = append(arcs, arc)
	}

	sort.Slice(arcs, func(i, j int) bool {
		return arcs[i].SeasonNumber < arcs[j].SeasonNumber
	})

	arcsCacheMutex.Lock()
	arcsCache = arcs
	arcsCacheTime = time.Now()
	arcsCacheMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(arcs)
}

func APIGetArcDetails(w http.ResponseWriter, r *http.Request) {
	seasonKey := r.URL.Query().Get("seasonKey")
	if seasonKey == "" {
		http.Error(w, "seasonKey parameter required", http.StatusBadRequest)
		return
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"

	arcDetailsCacheMutex.RLock()
	if !forceRefresh {
		if cachedResponse, exists := arcDetailsCache[seasonKey]; exists {
			if cacheTime, timeExists := arcDetailsCacheTime[seasonKey]; timeExists {
				if time.Since(cacheTime) < arcsCacheTTL {
					arcDetailsCacheMutex.RUnlock()
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(cachedResponse)
					return
				}
			}
		}
	}
	arcDetailsCacheMutex.RUnlock()

	cfg := shared.LoadConfig()
	if cfg.TargetDir == "" {
		http.Error(w, "Please set target directory first", http.StatusBadRequest)
		return
	}

	index := metadata.LoadMetadataCache()
	season, exists := index.Seasons[seasonKey]
	if !exists {
		http.Error(w, "Arc not found", http.StatusNotFound)
		return
	}

	torrents, err := scraper.FetchTorrents(cfg)
	if err != nil {
		logger.Log(true, "Error fetching torrents: %v", err)
		http.Error(w, "Failed to fetch torrents", http.StatusInternalServerError)
		return
	}

	downloadKeyMap := make(map[string]int)
	for _, t := range torrents {
		if t.DownloadKey > 0 {
			downloadKeyMap[t.ChapterRange] = t.DownloadKey
		}
	}

	var episodes []EpisodeStatus
	seasonDir := filepath.Join(cfg.TargetDir, seasonKey)

	for epRange, epData := range season.EpisodeRange {
		videoPathMP4 := filepath.Join(seasonDir, epData.Title+".mp4")
		videoPathMKV := filepath.Join(seasonDir, epData.Title+".mkv")
		hasVideo := shared.FileExists(videoPathMP4) || shared.FileExists(videoPathMKV)

		ep := EpisodeStatus{
			Title:        epData.Title,
			ChapterRange: epRange,
			HasVideo:     hasVideo,
			DownloadKey:  downloadKeyMap[epRange],
		}
		episodes = append(episodes, ep)
	}

	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].Title < episodes[j].Title
	})

	seasonNum := season.SeasonNumber
	if seasonNum == 0 && seasonKey != "Specials" {
		var snum int
		fmt.Sscanf(seasonKey, "Season %d", &snum)
		seasonNum = snum
	}

	response := map[string]any{
		"arc": ArcStatus{
			Name:         season.Name,
			SeasonKey:    seasonKey,
			SeasonNumber: seasonNum,
			ChapterRange: season.Range,
			HasMetadata:  true,
			VideoStatus:  metadata.HaveVideoStatus(season.Range),
			EpisodeCount: len(season.EpisodeRange),
			DownloadKey:  downloadKeyMap[season.Range],
		},
		"episodes": episodes,
	}

	arcDetailsCacheMutex.Lock()
	arcDetailsCache[seasonKey] = response
	arcDetailsCacheTime[seasonKey] = time.Now()
	arcDetailsCacheMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func APISearchArcs(w http.ResponseWriter, r *http.Request) {
	rangeFilter := r.URL.Query().Get("range")
	if rangeFilter == "" {
		http.Error(w, "range parameter required", http.StatusBadRequest)
		return
	}

	cfg := shared.LoadConfig()
	if cfg.Source.BaseURL == "" {
		http.Error(w, "Please run sync first", http.StatusBadRequest)
		return
	}

	torrents, err := scraper.FetchTorrents(cfg)
	if err != nil {
		logger.Log(true, "Failed to fetch torrents for search: %v", err)
		http.Error(w, "Failed to fetch torrents", http.StatusInternalServerError)
		return
	}

	logger.Log(false, "Searching for torrents matching range: %s", rangeFilter)

	var filtered []shared.TorrentEntry
	for _, t := range torrents {
		// Match exact chapter range or if the search range is contained within the torrent's range
		if matchesChapterRange(t.ChapterRange, rangeFilter) {
			logger.Log(false, "Found match: %s (range: %s)", t.TorrentName, t.ChapterRange)
			filtered = append(filtered, t)
		}
	}

	logger.Log(true, "Search for '%s' returned %d results", rangeFilter, len(filtered))

	// Sort by quality (highest first), then by seeders
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Quality != filtered[j].Quality {
			qi, _ := strconv.Atoi(strings.TrimSuffix(filtered[i].Quality, "p"))
			qj, _ := strconv.Atoi(strings.TrimSuffix(filtered[j].Quality, "p"))
			return qi > qj
		}
		return filtered[i].Seeders > filtered[j].Seeders
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}

// matchesChapterRange checks if a torrent's chapter range matches the search filter
// It handles exact matches, range overlaps, and single episode matches
func matchesChapterRange(torrentRange, searchRange string) bool {
	if torrentRange == "" || searchRange == "" {
		return false
	}

	// Normalize dashes
	torrentRange = shared.NormalizeDash(torrentRange)
	searchRange = shared.NormalizeDash(searchRange)

	// Exact match
	if torrentRange == searchRange {
		return true
	}

	// Parse ranges
	tStart, tEnd := shared.ParseRange(torrentRange)
	sStart, sEnd := shared.ParseRange(searchRange)

	// If parsing failed, try simple contains
	if tStart == -1 || sStart == -1 {
		return strings.Contains(torrentRange, searchRange)
	}

	// Check if ranges overlap or match
	// Torrent contains the search range
	if tStart <= sStart && tEnd >= sEnd {
		return true
	}

	// Search range contains the torrent (for searching full seasons)
	if sStart <= tStart && sEnd >= tEnd {
		return true
	}

	return false
}

func APISearchAndDownloadAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Method not allowed",
		})
		return
	}

	rangeFilter := r.FormValue("range")
	seasonKey := r.FormValue("seasonKey")

	if rangeFilter == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Range parameter required",
		})
		return
	}

	cfg := shared.LoadConfig()
	if cfg.TargetDir == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Target directory not set",
		})
		return
	}

	torrents, err := scraper.FetchTorrents(cfg)
	if err != nil {
		logger.Log(true, "Error fetching torrents: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Failed to fetch torrents",
		})
		return
	}

	// First try to find a full season torrent
	var fullSeasonTorrent *shared.TorrentEntry
	for i := range torrents {
		t := &torrents[i]
		if matchesChapterRange(t.ChapterRange, rangeFilter) && t.ChapterRange == rangeFilter {
			if fullSeasonTorrent == nil || t.Seeders > fullSeasonTorrent.Seeders {
				fullSeasonTorrent = t
			}
		}
	}

	queuedCount := 0

	if fullSeasonTorrent != nil {
		// Found full season, download it
		logger.Log(true, "Found full season torrent for %s: %s", rangeFilter, fullSeasonTorrent.TorrentName)
		torrentURL := fmt.Sprintf("%s/download/%d.torrent", cfg.Source.BaseURL, fullSeasonTorrent.TorrentID)

		if err := downloader.QueueDownload(fullSeasonTorrent, torrentURL, cfg); err != nil {
			logger.Log(true, "Failed to queue full season: %v", err)
		} else {
			queuedCount++
		}
	} else {
		// No full season found, try to download individual episodes
		logger.Log(true, "No full season found for %s, searching for individual episodes", rangeFilter)

		// Get episode list from metadata
		index := metadata.LoadMetadataCache()
		if index == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"message": "Metadata cache not loaded",
			})
			return
		}

		// Find the season
		var season *shared.SeasonIndex
		for key, s := range index.Seasons {
			if key == seasonKey || s.Range == rangeFilter {
				season = &s
				break
			}
		}

		if season == nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"message": "Season not found in metadata",
			})
			return
		}

		// Build map of available torrents by chapter range
		torrentMap := make(map[string]*shared.TorrentEntry)
		for i := range torrents {
			t := &torrents[i]
			if existing, ok := torrentMap[t.ChapterRange]; !ok || t.Seeders > existing.Seeders {
				torrentMap[t.ChapterRange] = t
			}
		}

		// Queue each episode
		for epRange := range season.EpisodeRange {
			if torrent, ok := torrentMap[epRange]; ok {
				torrentURL := fmt.Sprintf("%s/download/%d.torrent", cfg.Source.BaseURL, torrent.TorrentID)
				if err := downloader.QueueDownload(torrent, torrentURL, cfg); err != nil {
					logger.Log(true, "Failed to queue episode %s: %v", epRange, err)
				} else {
					queuedCount++
					logger.Log(false, "Queued episode: %s", torrent.TorrentName)
				}
			} else {
				logger.Log(true, "No torrent found for episode range: %s", epRange)
			}
		}
	}

	InvalidateArcsCache()

	if queuedCount == 0 {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "No torrents found to download",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": fmt.Sprintf("Queued %d torrent(s) for download", queuedCount),
		"count":   queuedCount,
	})
}

func APIDownloadArc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Method not allowed",
		})
		return
	}

	downloadKeyStr := r.FormValue("downloadKey")
	downloadKey, err := strconv.Atoi(downloadKeyStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Invalid download key",
		})
		return
	}

	cfg := shared.LoadConfig()
	if cfg.TargetDir == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Target directory not set",
		})
		return
	}

	torrents, err := scraper.FetchTorrents(cfg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Failed to fetch torrents",
		})
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
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Torrent not found",
		})
		return
	}

	torrentURL := fmt.Sprintf("%s/download/%d.torrent", cfg.Source.BaseURL, match.TorrentID)

	if err := downloader.QueueDownload(match, torrentURL, cfg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": fmt.Sprintf("Failed to queue download: %v", err),
		})
		return
	}

	InvalidateArcsCache()

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

	InvalidateArcsCache()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Metadata synced successfully",
	})
}

func APIActivityStatus(w http.ResponseWriter, r *http.Request) {
	downloads := shared.GetActiveDownloads()
	cfg := shared.LoadConfig()

	hasPlacedFiles := false
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
		if dl.Placed {
			hasPlacedFiles = true
		}
	}

	if hasPlacedFiles {
		InvalidateArcsCache()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"downloads": downloads,
		"count":     len(downloads),
	})
}

func APIBrowseDirectories(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read directory: %v", err), http.StatusBadRequest)
		return
	}

	var dirs []map[string]string

	parent := filepath.Dir(path)
	if parent != path {
		dirs = append(dirs, map[string]string{
			"name": "..",
			"path": parent,
		})
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(path, entry.Name())
			dirs = append(dirs, map[string]string{
				"name": entry.Name(),
				"path": fullPath,
			})
		}
	}

	sort.Slice(dirs, func(i, j int) bool {
		if dirs[i]["name"] == ".." {
			return true
		}
		if dirs[j]["name"] == ".." {
			return false
		}
		return dirs[i]["name"] < dirs[j]["name"]
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"currentPath": path,
		"directories": dirs,
	})
}
