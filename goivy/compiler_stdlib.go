package goivy

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/glycerine/ivy/goivy/fileops"
)

// StandardLibrary is an in-memory snapshot of Ivy's versioned standard
// include tree. It caches source text, not AST nodes, because compilation
// phases may annotate or otherwise mutate parsed nodes.
type StandardLibrary struct {
	BaseDir  string
	versions []string
	files    map[string]map[string]string
}

var (
	standardLibraryMu     sync.Mutex
	standardLibraryByBase = make(map[string]*StandardLibrary)
)

// PreloadStandardLibrary loads Ivy's standard include tree into cfg. The web
// UI calls this at backend startup so browser uploads do not lazily discover or
// reread the standard library.
func PreloadStandardLibrary(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("nil Ivy config")
	}
	if cfg.IuCfg == nil {
		cfg.IuCfg = NewIvyUtilsConfig()
	}
	baseDir := strings.TrimSpace(cfg.IuCfg.IncludeBaseDir)
	if baseDir == "" {
		baseDir = getIncludeBaseDir()
	}
	lib, err := loadStandardLibrary(baseDir)
	if err != nil {
		return err
	}
	cfg.StandardLibrary = lib
	cfg.IuCfg.IncludeBaseDir = lib.BaseDir
	return nil
}

func loadStandardLibrary(baseDir string) (*StandardLibrary, error) {
	baseDir = filepath.Clean(baseDir)
	standardLibraryMu.Lock()
	if lib := standardLibraryByBase[baseDir]; lib != nil {
		standardLibraryMu.Unlock()
		return lib, nil
	}
	standardLibraryMu.Unlock()

	lib, err := readStandardLibrary(baseDir)
	if err != nil {
		return nil, err
	}

	standardLibraryMu.Lock()
	if existing := standardLibraryByBase[baseDir]; existing != nil {
		standardLibraryMu.Unlock()
		return existing, nil
	}
	standardLibraryByBase[baseDir] = lib
	standardLibraryMu.Unlock()
	return lib, nil
}

func readStandardLibrary(baseDir string) (*StandardLibrary, error) {
	entries, err := fileops.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("read Ivy standard library %s: %w", baseDir, err)
	}

	lib := &StandardLibrary{
		BaseDir: filepath.Clean(baseDir),
		files:   make(map[string]map[string]string),
	}
	for _, entry := range entries {
		if !entry.IsDir || !incDirPat.MatchString(entry.Name) {
			continue
		}
		version := entry.Name
		dir := filepath.Join(lib.BaseDir, version)
		files, err := fileops.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("read Ivy standard library %s: %w", dir, err)
		}
		for _, file := range files {
			if file.IsDir || !strings.HasSuffix(file.Name, ".ivy") {
				continue
			}
			data, err := fileops.ReadFile(filepath.Join(dir, file.Name))
			if err != nil {
				return nil, fmt.Errorf("read Ivy standard library %s: %w", filepath.Join(dir, file.Name), err)
			}
			if lib.files[version] == nil {
				lib.files[version] = make(map[string]string)
				lib.versions = append(lib.versions, version)
			}
			lib.files[version][file.Name] = string(data)
		}
	}
	sort.Strings(lib.versions)
	if len(lib.versions) == 0 {
		return nil, fmt.Errorf("Ivy standard library %s has no versioned include directories", baseDir)
	}
	return lib, nil
}

func (lib *StandardLibrary) includeSource(languageVersion, moduleName string) (filename, source string, ok bool) {
	if lib == nil {
		return "", "", false
	}
	fileName := moduleName
	if !strings.HasSuffix(fileName, ".ivy") {
		fileName += ".ivy"
	}
	bestVersion := ""
	for _, version := range lib.versions {
		if !VersionLE(languageVersion, version) {
			continue
		}
		if bestVersion == "" || VersionLE(version, bestVersion) {
			bestVersion = version
		}
	}
	if bestVersion == "" {
		return "", "", false
	}
	source, ok = lib.files[bestVersion][fileName]
	if !ok {
		return "", "", false
	}
	return filepath.Join(lib.BaseDir, bestVersion, fileName), source, true
}
