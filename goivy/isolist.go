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
			ivySpecs = append(ivySpecs, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(ivySpecs)
	return ivySpecs, nil
}
