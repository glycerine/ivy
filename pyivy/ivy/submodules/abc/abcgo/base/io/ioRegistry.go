package io

import (
	"github.com/glycerine/abcgo/base/abc"
	"sync"
)

type FormatReader func(filename string, check bool) (*abc.Network, error)
type FormatWriter func(network *abc.Network, filename string) error
type formatHandler struct {
	read  FormatReader
	write FormatWriter
}

var formats = struct {
	sync.RWMutex
	handlers map[Io_FileType_t]formatHandler
}{handlers: make(map[Io_FileType_t]formatHandler)}

// RegisterFormat installs concrete format implementations. Nil leaves the
// corresponding operation unchanged, allowing read/write families to register separately.
func RegisterFormat(kind Io_FileType_t, read FormatReader, write FormatWriter) {
	formats.Lock()
	defer formats.Unlock()
	handler := formats.handlers[kind]
	if read != nil {
		handler.read = read
	}
	if write != nil {
		handler.write = write
	}
	formats.handlers[kind] = handler
}
