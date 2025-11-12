package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
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

	staticSubFS, err := fs.Sub(content, "static")
	if err != nil {
		logger.Log(true, "Failed to create static sub filesystem: %v", err)
		return err
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSubFS))))
	logger.Log(true, "📁 Serving static files from embedded filesystem")

	mux.HandleFunc("/arcs", handlers.HandleArcs(templates))
	mux.HandleFunc("/activity", handlers.HandleActivity(templates))
	mux.HandleFunc("/settings", handlers.HandleSettings(templates))
	mux.HandleFunc("/system", handlers.HandleSystem(templates))

	mux.HandleFunc("/api/arcs/list", handlers.APIListArcs)
	mux.HandleFunc("/api/arcs/details", handlers.APIGetArcDetails)
	mux.HandleFunc("/api/arcs/search", handlers.APISearchArcs)
	mux.HandleFunc("/api/arcs/download", handlers.APIDownloadArc)
	mux.HandleFunc("/api/settings/update", handlers.APIUpdateSettings)
	mux.HandleFunc("/api/settings/test-client", handlers.APITestClient)
	mux.HandleFunc("/api/settings/browse", handlers.APIBrowseDirectories)
	mux.HandleFunc("/api/system/sync", handlers.APISync)
	mux.HandleFunc("/api/activity/status", handlers.APIActivityStatus)

	mux.HandleFunc("/", handlers.HandleIndex(templates))

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	logger.Log(true, "🌐 Starting web server on http://0.0.0.0:%d", port)
	return http.ListenAndServe(addr, mux)
}
