package frontend

import (
	"compress/gzip"
	"crypto/sha256"
	"embed"
	"fmt"
	"io"
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

type fileMeta struct {
	etag string
	size int64
}

var (
	metaMap map[string]*fileMeta
	fileSys http.FileSystem
	once    sync.Once
)

func initFS() {
	subFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}

	fileSys = http.FS(subFS)

	metaMap = make(map[string]*fileMeta)

	err = fs.WalkDir(subFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		f, err := subFS.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return err
		}

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}

		metaMap["/"+p] = &fileMeta{
			etag: fmt.Sprintf(`"%x"`, h.Sum(nil)),
			size: info.Size(),
		}

		log.Printf("[fs]\tadded: /%s", p)

		return nil
	})
	if err != nil {
		panic(err)
	}
}

// FS returns the embedded filesystem rooted at dist/.
func FS() http.FileSystem {
	once.Do(initFS)
	return fileSys
}

// Handler returns an http.Handler serving the embedded SPA.
//
// Features:
//   - SPA fallback to index.html
//   - ETag support
//   - Cache-Control
//   - MIME type detection
//   - Conditional gzip compression
func Handler() http.Handler {
	once.Do(initFS)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := r.URL.Path

		if reqPath == "/" {
			reqPath = "/index.html"
		}

		meta, ok := metaMap[reqPath]
		if !ok {
			// SPA fallback.
			reqPath = "/index.html"
			meta = metaMap[reqPath]

			if meta == nil {
				http.NotFound(w, r)
				return
			}
		}

		// Caching.
		w.Header().Set("ETag", meta.etag)
		w.Header().Set("Cache-Control", "public, max-age=3600")

		// Conditional request.
		if r.Header.Get("If-None-Match") == meta.etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// MIME type.
		ext := path.Ext(reqPath)
		if contentType := mime.TypeByExtension(ext); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}

		// Open embedded file.
		f, err := FS().Open(strings.TrimPrefix(reqPath, "/"))
		if err != nil {
			log.Printf("frontend: open %q: %v", reqPath, err)
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		// Gzip text assets when supported.
		if acceptsGzip(r) && isTextAsset(reqPath) {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")

			gz := gzip.NewWriter(w)
			defer gz.Close()

			_, _ = io.Copy(gz, f)
			return
		}

		// Serve uncompressed content.
		w.Header().Set("Content-Length", fmt.Sprintf("%d", meta.size))

		_, _ = io.Copy(w, f)
	})
}

func acceptsGzip(r *http.Request) bool {
	return strings.Contains(
		strings.ToLower(r.Header.Get("Accept-Encoding")),
		"gzip",
	)
}

func isTextAsset(p string) bool {
	p = strings.ToLower(p)

	switch path.Ext(p) {
	case ".html", ".css", ".js", ".json", ".svg", ".txt":
		return true
	default:
		return false
	}
}
