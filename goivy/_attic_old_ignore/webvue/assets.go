package webvue

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const DefaultRunwebDir = ".runweb"

//go:embed static
var embeddedStatic embed.FS

// StaticAssets returns the built browser assets rooted at webvue/static.
func StaticAssets() fs.FS {
	static, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		panic(err)
	}
	return static
}

// MaterializeRunwebDir copies embedded browser assets into dst so the HTTP
// server can serve normal files from disk.
func MaterializeRunwebDir(dst string) error {
	if strings.TrimSpace(dst) == "" {
		dst = DefaultRunwebDir
	}
	return copyFS(StaticAssets(), dst)
}

func copyFS(src fs.FS, dst string) error {
	return fs.WalkDir(src, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		target := filepath.Join(dst, filepath.FromSlash(path))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(src, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
