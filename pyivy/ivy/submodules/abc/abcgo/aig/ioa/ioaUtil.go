package ioa

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Ioa_FileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
func Ioa_FileNameGeneric(path string) string { return strings.TrimSuffix(path, filepath.Ext(path)) }
func Ioa_FileNameGenericAppend(base, suffix string) string {
	if base == "" {
		return suffix
	}
	name := Ioa_FileNameGeneric(base) + suffix
	for i := len(name) - 1; i >= 0; i-- {
		c := name[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c == '.') {
			return name[i+1:]
		}
	}
	return name
}
func Ioa_TimeStamp() string { return time.Now().Format("Mon Jan _2 15:04:05 2006") }
