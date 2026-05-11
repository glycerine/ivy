package goivy

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/fileops"
)

func ivyReadableFileExists(filename string) bool {
	info, err := fileops.Stat(filename)
	if err != nil {
		fmt.Printf("fileops.Stat(filename='%v') gave err='%v'\n", filename, err)
	}
	return err == nil && !info.IsDir
}
