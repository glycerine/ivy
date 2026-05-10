package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/glycerine/ivy/goivy/fileops"
)

var readPath string
var writePath string
var dirPath string

func main() {
	if len(os.Args) == 3 {
		readPath = os.Args[1]
		writePath = os.Args[2]
	}
	if readPath == "" || writePath == "" {
		fmt.Fprintln(os.Stderr, "usage: tinyfsprobe read-path write-path")
		os.Exit(2)
	}

	readStat, err := fileops.Stat(readPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stat read path: %v\n", err)
		os.Exit(1)
	}
	if readStat.IsDir {
		fmt.Fprintf(os.Stderr, "read path is a directory: %s\n", readPath)
		os.Exit(1)
	}

	data, err := fileops.ReadFile(readPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read file: %v\n", err)
		os.Exit(1)
	}
	if int64(len(data)) != readStat.Size {
		fmt.Fprintf(os.Stderr, "read size mismatch: stat=%d read=%d\n", readStat.Size, len(data))
		os.Exit(1)
	}

	payload := append([]byte("tinyfsprobe:"), data...)
	if err := fileops.WriteFile(writePath, payload); err != nil {
		fmt.Fprintf(os.Stderr, "write file: %v\n", err)
		os.Exit(1)
	}
	writeStat, err := fileops.Stat(writePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stat write path: %v\n", err)
		os.Exit(1)
	}
	if writeStat.IsDir {
		fmt.Fprintf(os.Stderr, "write path is a directory: %s\n", writePath)
		os.Exit(1)
	}
	if int64(len(payload)) != writeStat.Size {
		fmt.Fprintf(os.Stderr, "write size mismatch: stat=%d payload=%d\n", writeStat.Size, len(payload))
		os.Exit(1)
	}
	written, err := fileops.ReadFile(writePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read written file: %v\n", err)
		os.Exit(1)
	}
	if !bytes.Equal(written, payload) {
		fmt.Fprintln(os.Stderr, "written payload mismatch")
		os.Exit(1)
	}
	if dirPath != "" {
		entries, err := fileops.ReadDir(dirPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "readdir: %v\n", err)
			os.Exit(1)
		}
		foundRead := false
		foundWrite := false
		for _, entry := range entries {
			if entry.Name == "input.ivy" && !entry.IsDir {
				foundRead = true
			}
			if entry.Name == "output.ivy" && !entry.IsDir {
				foundWrite = true
			}
		}
		if !foundRead || !foundWrite {
			fmt.Fprintf(os.Stderr, "readdir missing files: input=%v output=%v\n", foundRead, foundWrite)
			os.Exit(1)
		}
	}

	fmt.Printf("OK read=%d wrote=%d\n", len(data), len(payload))
}
