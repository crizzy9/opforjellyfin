package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/web/handlers"
)

//go:embed templates static
var content embed.FS

var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseFS(content, "templates/*.html")
	if err != nil {
		logger.Log(true, "Failed to parse templates: %v", err)
	}
}

func StartServer(port int) error {
	mux := http.NewServeMux()

	staticFS := http.FS(content)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(staticFS)))

	mux.HandleFunc("/", handlers.HandleIndex(templates))
	mux.HandleFunc("/episodes", handlers.HandleEpisodes(templates))
	mux.HandleFunc("/activity", handlers.HandleActivity(templates))
	mux.HandleFunc("/settings", handlers.HandleSettings(templates))
	mux.HandleFunc("/system", handlers.HandleSystem(templates))

	mux.HandleFunc("/api/episodes/list", handlers.APIListEpisodes)
	mux.HandleFunc("/api/episodes/search", handlers.APISearchEpisodes)
	mux.HandleFunc("/api/episodes/download", handlers.APIDownloadEpisode)
	mux.HandleFunc("/api/settings/update", handlers.APIUpdateSettings)
	mux.HandleFunc("/api/settings/test-client", handlers.APITestClient)
	mux.HandleFunc("/api/system/sync", handlers.APISync)
	mux.HandleFunc("/api/activity/status", handlers.APIActivityStatus)

	addr := fmt.Sprintf(":%d", port)
	logger.Log(true, "🌐 Starting web server on http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}
