package goivy

import (
	"os"
	"path/filepath"
	"sort"
)

// ListIsolates returns the sorted list of isolates registered by the normal
// Ivy module load path.
func ListIsolates(pathToDotIvy string) (isoList []string, err error) {
	mod := New()
	err = SourceFile(pathToDotIvy, mod, mod.Sig, map[string]interface{}{"create_isolate": false})
	if err != nil {
		return nil, err
	}
	for name := range mod.Isolates {
		isoList = append(isoList, name)
	}
	sort.Strings(isoList)
	return isoList, nil
}

func ListAllIvyPathsRecursively(startingDir string) (ivySpecs []string, err error) {
	err = filepath.Walk(startingDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info == nil || info.IsDir() {
			// Returning nil for a directory tells filepath.Walk to recurse into it.
			return nil
		}
		if filepath.Ext(path) == ".ivy" {
			ok, err := IvyVersionSupported(path)
			if err == nil && ok {
				ivySpecs = append(ivySpecs, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(ivySpecs)
	return ivySpecs, nil
}

// a quick read of the first line of the file.
// If the read fails return false + the filesystem error.
// Otherwise parse the version and return true for Ivy versions > 1.5,
// and false for all versions <= 1.5
// For example, a file "#lang ivy1.5" will return above15 == false,
// but #lang ivy1.6 will return above15 true.
func IvyVersionSupported(path string) (above15 bool, err error) {
	panic("implement me")
}
