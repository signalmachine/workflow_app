package app

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed web_dist
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
		serveEmbeddedDistFile(w, r, distFS, "index.html")
	case looksLikeStaticAsset(relativePath) && embeddedDistFileExists(distFS, relativePath):
		serveEmbeddedDistFile(w, r, distFS, relativePath)
	default:
		serveEmbeddedDistFile(w, r, distFS, "200.html")
	}
}

func looksLikeStaticAsset(name string) bool {
	base := path.Base(strings.TrimSpace(name))
	return strings.Contains(base, ".")
}

func embeddedDistFileExists(root fs.FS, name string) bool {
	if root == nil || strings.TrimSpace(name) == "" {
		return false
	}
	info, err := fs.Stat(root, name)
	return err == nil && !info.IsDir()
}

func serveEmbeddedDistFile(w http.ResponseWriter, r *http.Request, root fs.FS, name string) {
	file, err := root.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	readSeeker, ok := file.(io.ReadSeeker)
	if !ok {
		http.Error(w, "web bundle file is not seekable", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, info.Name(), info.ModTime(), readSeeker)
}
