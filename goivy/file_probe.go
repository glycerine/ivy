package goivy

import (
	"github.com/glycerine/ivy/goivy/fileops"
)

func ivyReadableFileExists(filename string) bool {
	info, err := fileops.Stat(filename)
	return err == nil && !info.IsDir
}
