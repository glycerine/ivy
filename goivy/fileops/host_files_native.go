//go:build !tinygo || !js || !wasm

package fileops

import "os"

type FileStat struct {
	Size  int64
	IsDir bool
}

func ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func WriteFile(name string, data []byte) error {
	return os.WriteFile(name, data, 0o666)
}

func Stat(name string) (FileStat, error) {
	info, err := os.Stat(name)
	if err != nil {
		return FileStat{}, err
	}
	return FileStat{
		Size:  info.Size(),
		IsDir: info.IsDir(),
	}, nil
}
