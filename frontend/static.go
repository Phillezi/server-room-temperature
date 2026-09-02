package frontend

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"
	"sync"
)

//go:embed dist
var distFS embed.FS

// Config holds runtime parameters injected into the frontend SPA via SSR.
type Config struct {
	NatsWSURL    string `json:"nats_url,omitempty"`
	NatsUser     string `json:"nats_user,omitempty"`
	NatsPassword string `json:"nats_password,omitempty"`
	Subject      string `json:"subject,omitempty"`
}

type fileMeta struct {
	etag        string
	size        int64
	contentType string
	data        []byte
}

type spaHandler struct {
	indexMeta *fileMeta
	metaMap   map[string]*fileMeta
}

var (
	indexTmpl *template.Template
	rawFS     http.FileSystem
	rawMeta   map[string]*fileMeta
	initOnce  sync.Once
)

func initStaticFS() {
	subFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}

	rawFS = http.FS(subFS)
	rawMeta = make(map[string]*fileMeta)

	indexBytes, err := fs.ReadFile(subFS, "index.html")
	if err != nil {
		panic(err)
	}

	indexTmpl = template.Must(template.New("index.html").Parse(string(indexBytes)))

	err = fs.WalkDir(subFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := fs.ReadFile(subFS, p)
		if err != nil {
			return err
		}

		h := sha256.Sum256(data)
		ext := path.Ext(p)
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		rawMeta["/"+p] = &fileMeta{
			etag:        fmt.Sprintf(`"%x"`, h),
			size:        int64(len(data)),
			contentType: contentType,
			data:        data,
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
}

// FS returns the embedded filesystem rooted at dist/.
func FS() http.FileSystem {
	initOnce.Do(initStaticFS)
	return rawFS
}

// Handler returns an http.Handler serving the embedded SPA with optional SSR config injection.
func Handler(cfgs ...Config) http.Handler {
	var cfg Config
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}
	return NewHandler(cfg)
}

// NewHandler creates a new SPA handler with injected SSR configuration.
func NewHandler(cfg Config) http.Handler {
	initOnce.Do(initStaticFS)

	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		cfgJSON = []byte("{}")
	}

	var buf bytes.Buffer
	data := struct {
		ConfigJSON template.JS
	}{
		ConfigJSON: template.JS(cfgJSON),
	}

	if err := indexTmpl.Execute(&buf, data); err != nil {
		log.Printf("frontend: execute index template: %v", err)
		buf.Reset()
		buf.Write(rawMeta["/index.html"].data)
	}

	renderedHTML := buf.Bytes()
	h := sha256.Sum256(renderedHTML)

	indexMeta := &fileMeta{
		etag:        fmt.Sprintf(`"%x"`, h),
		size:        int64(len(renderedHTML)),
		contentType: "text/html; charset=utf-8",
		data:        renderedHTML,
	}

	return &spaHandler{
		indexMeta: indexMeta,
		metaMap:   rawMeta,
	}
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqPath := r.URL.Path
	if reqPath == "/" || reqPath == "/index.html" {
		h.serveFile(w, r, h.indexMeta)
		return
	}

	meta, ok := h.metaMap[reqPath]
	if !ok {
		// SPA fallback
		h.serveFile(w, r, h.indexMeta)
		return
	}

	h.serveFile(w, r, meta)
}

func (h *spaHandler) serveFile(w http.ResponseWriter, r *http.Request, meta *fileMeta) {
	// Caching
	w.Header().Set("ETag", meta.etag)
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// Conditional request
	if r.Header.Get("If-None-Match") == meta.etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", meta.contentType)

	// Gzip text assets when supported
	if acceptsGzip(r) && isTextAsset(meta.contentType) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		_, _ = gz.Write(meta.data)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", meta.size))
	_, _ = w.Write(meta.data)
}

func acceptsGzip(r *http.Request) bool {
	return strings.Contains(
		strings.ToLower(r.Header.Get("Accept-Encoding")),
		"gzip",
	)
}

func isTextAsset(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.HasPrefix(ct, "text/") ||
		strings.Contains(ct, "javascript") ||
		strings.Contains(ct, "json") ||
		strings.Contains(ct, "xml") ||
		strings.Contains(ct, "svg")
}
