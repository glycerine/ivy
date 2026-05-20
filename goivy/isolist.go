package goivy

import "sort"

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
