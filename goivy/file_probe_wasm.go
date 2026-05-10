//go:build wasm

package goivy

import "os"

func ivyReadableFileExists(filename string) bool {
	// TinyGo's js/wasm os.Stat is a stub that returns ErrNotImplemented.
	// os.Open lowers to the file APIs that our Node-backed TinyGo WASI shim
	// supports, and ReadModule needs the file to be readable anyway.
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}
