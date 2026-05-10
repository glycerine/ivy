//go:build tinygo && js && wasm

package fileops

import "fmt"

type FileStat struct {
	Size  int64
	IsDir bool
}

type fsScratch = uint32
type fsBytesHandle = uint32

func packedWord(data []byte, offset uint32) (uint32, uint32) {
	if offset >= uint32(len(data)) {
		return 0, 0
	}
	n := uint32(len(data)) - offset
	if n > 4 {
		n = 4
	}
	var word uint32
	for i := uint32(0); i < n; i++ {
		word |= uint32(data[offset+i]) << (8 * i)
	}
	return word, n
}

func scratchBytes(data []byte) fsScratch {
	h := goivyFSScratchBytesBegin(uint32(len(data)))
	for offset := uint32(0); offset < uint32(len(data)); {
		word, n := packedWord(data, offset)
		goivyFSScratchBytesWrite(h, offset, word, n)
		offset += n
	}
	return h
}

func readHostBytes(h fsBytesHandle) []byte {
	if h == 0 {
		return nil
	}
	n := goivyFSBytesLen(h)
	buf := make([]byte, n)
	for offset := uint32(0); offset < n; offset += 4 {
		word := goivyFSBytesWord(h, offset)
		for i := uint32(0); i < 4 && offset+i < n; i++ {
			buf[offset+i] = byte(word >> (8 * i))
		}
	}
	goivyFSBytesRelease(h)
	return buf
}

func pathScratch(name string) (fsScratch, uint32) {
	b := []byte(name)
	return scratchBytes(b), uint32(len(b))
}

func hostErr(op, name string, errno int32) error {
	return fmt.Errorf("%s %s: host errno %d", op, name, errno)
}

func ReadFile(name string) ([]byte, error) {
	path, pathLen := pathScratch(name)
	defer goivyFSScratchBytesRelease(path)
	h := goivyFSReadFile(path, pathLen)
	if h == 0 {
		return nil, hostErr("read", name, goivyFSLastErrno())
	}
	return readHostBytes(h), nil
}

func WriteFile(name string, data []byte) error {
	path, pathLen := pathScratch(name)
	defer goivyFSScratchBytesRelease(path)
	payload := scratchBytes(data)
	defer goivyFSScratchBytesRelease(payload)
	if errno := goivyFSWriteFile(path, pathLen, payload, uint32(len(data))); errno != 0 {
		return hostErr("write", name, errno)
	}
	return nil
}

func Stat(name string) (FileStat, error) {
	path, pathLen := pathScratch(name)
	defer goivyFSScratchBytesRelease(path)
	if errno := goivyFSStat(path, pathLen); errno != 0 {
		return FileStat{}, hostErr("stat", name, errno)
	}
	size := uint64(goivyFSLastSizeLo()) | (uint64(goivyFSLastSizeHi()) << 32)
	return FileStat{
		Size:  int64(size),
		IsDir: goivyFSLastIsDir() != 0,
	}, nil
}

//go:wasmimport goivy_fs __scratch_bytes_begin
func goivyFSScratchBytesBegin(n uint32) fsScratch

//go:wasmimport goivy_fs __scratch_bytes_write
func goivyFSScratchBytesWrite(h fsScratch, offset uint32, word uint32, n uint32)

//go:wasmimport goivy_fs __scratch_bytes_release
func goivyFSScratchBytesRelease(h fsScratch)

//go:wasmimport goivy_fs read_file
func goivyFSReadFile(path fsScratch, pathLen uint32) fsBytesHandle

//go:wasmimport goivy_fs write_file
func goivyFSWriteFile(path fsScratch, pathLen uint32, data fsScratch, dataLen uint32) int32

//go:wasmimport goivy_fs stat
func goivyFSStat(path fsScratch, pathLen uint32) int32

//go:wasmimport goivy_fs bytes_len
func goivyFSBytesLen(h fsBytesHandle) uint32

//go:wasmimport goivy_fs bytes_word
func goivyFSBytesWord(h fsBytesHandle, offset uint32) uint32

//go:wasmimport goivy_fs bytes_release
func goivyFSBytesRelease(h fsBytesHandle)

//go:wasmimport goivy_fs last_errno
func goivyFSLastErrno() int32

//go:wasmimport goivy_fs last_size_lo
func goivyFSLastSizeLo() uint32

//go:wasmimport goivy_fs last_size_hi
func goivyFSLastSizeHi() uint32

//go:wasmimport goivy_fs last_is_dir
func goivyFSLastIsDir() uint32
