package app

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:web_dist
var webDistFS embed.FS

func (h *AgentAPIHandler) handleSvelteApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}

	distFS, err := fs.Sub(webDistFS, "web_dist")
	if err != nil {
		http.Error(w, "web bundle unavailable", http.StatusServiceUnavailable)
		return
	}

	relativePath := strings.TrimPrefix(r.URL.Path, webAppPath)
	relativePath = strings.TrimPrefix(relativePath, "/")
	relativePath = path.Clean(relativePath)
	if relativePath == "." {
		relativePath = ""
	}

	switch {
	case relativePath == "":
		// Serve the SPA fallback shell at /app so asset URLs stay rooted under /app
		// even when the browser requests the base path without a trailing slash.
		serveEmbeddedDistFile(w, r, distFS, "200.html")
	case looksLikeStaticAsset(relativePath):
		if serveEmbeddedDistFileIfExists(w, r, distFS, relativePath) {
			return
		}
		serveEmbeddedDistFile(w, r, distFS, "200.html")
	default:
		serveEmbeddedDistFile(w, r, distFS, "200.html")
	}
}

func looksLikeStaticAsset(name string) bool {
	base := path.Base(strings.TrimSpace(name))
	return strings.Contains(base, ".")
}

func serveEmbeddedDistFileIfExists(w http.ResponseWriter, r *http.Request, root fs.FS, name string) bool {
	file, err := root.Open(name)
	if err != nil {
		return false
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil || info.IsDir() {
		return false
	}

	readSeeker, ok := file.(io.ReadSeeker)
	if !ok {
		http.Error(w, "web bundle file is not seekable", http.StatusInternalServerError)
		return true
	}

	http.ServeContent(w, r, info.Name(), info.ModTime(), readSeeker)
	return true
}

func serveEmbeddedDistFile(w http.ResponseWriter, r *http.Request, root fs.FS, name string) {
	if serveEmbeddedDistFileIfExists(w, r, root, name) {
		return
	}
	http.NotFound(w, r)
}
