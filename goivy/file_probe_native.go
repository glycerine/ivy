//go:build !wasm

package goivy

import "os"

func ivyReadableFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
