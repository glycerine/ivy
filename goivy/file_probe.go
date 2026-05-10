package goivy

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/fileops"
)

func ivyReadableFileExists(filename string) bool {
	info, err := fileops.Stat(filename)
	fmt.Printf("fileops.Stat(filename='%v') gave err='%v' and info='%v'\n", filename, err, info)
	return err == nil && !info.IsDir
}
