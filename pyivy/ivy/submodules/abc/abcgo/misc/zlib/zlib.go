// Package zlib provides ABC's compression services using Go's standard library.
// This is a Go replacement, not the original zlib implementation.
package zlib

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	stdzlib "compress/zlib"
	"errors"
	"fmt"
	"hash/adler32"
	"hash/crc32"
	"io"
)

const (
	Z_NO_FLUSH int32 = iota
	Z_PARTIAL_FLUSH
	Z_SYNC_FLUSH
	Z_FULL_FLUSH
	Z_FINISH
	Z_BLOCK
	Z_TREES
)
const (
	Z_OK                  int32 = 0
	Z_STREAM_END          int32 = 1
	Z_NEED_DICT           int32 = 2
	Z_ERRNO               int32 = -1
	Z_STREAM_ERROR        int32 = -2
	Z_DATA_ERROR          int32 = -3
	Z_MEM_ERROR           int32 = -4
	Z_BUF_ERROR           int32 = -5
	Z_VERSION_ERROR       int32 = -6
	Z_NO_COMPRESSION            = flate.NoCompression
	Z_BEST_SPEED                = flate.BestSpeed
	Z_BEST_COMPRESSION          = flate.BestCompression
	Z_DEFAULT_COMPRESSION       = flate.DefaultCompression
	Z_HUFFMAN_ONLY              = 2
	Z_DEFLATED                  = 8
	Z_DEFAULT_STRATEGY          = 0
	Z_FILTERED                  = 1
	Z_RLE                       = 3
	Z_FIXED                     = 4
	ZLIB_VERSION                = "1.2.5"
)

// Compression levels are distinct from zlib's strategy constants.
const (
	HuffmanOnly        = flate.HuffmanOnly
	NoCompression      = flate.NoCompression
	BestSpeed          = flate.BestSpeed
	BestCompression    = flate.BestCompression
	DefaultCompression = flate.DefaultCompression
	Z_BINARY           = 0
	Z_TEXT             = 1
	Z_ASCII            = Z_TEXT
	Z_UNKNOWN          = 2
	MAX_WBITS          = 15
	MAX_MEM_LEVEL      = 9
	DEF_MEM_LEVEL      = 8
)

type Resetter = stdzlib.Resetter
type Writer = stdzlib.Writer
type GzipHeader = gzip.Header

var ErrChecksum = stdzlib.ErrChecksum
var ErrDictionary = stdzlib.ErrDictionary
var ErrHeader = stdzlib.ErrHeader

func NewReader(r io.Reader) (io.ReadCloser, error) { return stdzlib.NewReader(r) }
func NewReaderDict(r io.Reader, dict []byte) (io.ReadCloser, error) {
	return stdzlib.NewReaderDict(r, dict)
}
func NewWriter(w io.Writer) *Writer                          { return stdzlib.NewWriter(w) }
func NewWriterLevel(w io.Writer, level int) (*Writer, error) { return stdzlib.NewWriterLevel(w, level) }
func NewWriterLevelDict(w io.Writer, level int, dict []byte) (*Writer, error) {
	return stdzlib.NewWriterLevelDict(w, level, dict)
}
func ZlibVersion() string { return ZLIB_VERSION }
func ZError(code int32) string {
	switch code {
	case Z_OK:
		return ""
	case Z_STREAM_END:
		return "stream end"
	case Z_NEED_DICT:
		return "need dictionary"
	case Z_ERRNO:
		return "file error"
	case Z_STREAM_ERROR:
		return "stream error"
	case Z_DATA_ERROR:
		return "data error"
	case Z_MEM_ERROR:
		return "insufficient memory"
	case Z_BUF_ERROR:
		return "buffer error"
	case Z_VERSION_ERROR:
		return "incompatible version"
	}
	return "unknown error"
}

// CompressBytes produces a complete zlib stream at the requested level.
func CompressBytes(source []byte, level int) ([]byte, error) {
	var b bytes.Buffer
	w, err := NewWriterLevel(&b, level)
	if err != nil {
		return nil, err
	}
	if _, err = w.Write(source); err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
func UncompressBytes(source []byte) ([]byte, error) {
	r, err := NewReader(bytes.NewReader(source))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
func CompressBound(n uint64) uint64 { return n + (n >> 12) + (n >> 14) + (n >> 25) + 13 }
func Compress(dest []byte, destLen *uint64, source []byte) int32 {
	return Compress2(dest, destLen, source, Z_DEFAULT_COMPRESSION)
}
func Compress2(dest []byte, destLen *uint64, source []byte, level int) int32 {
	if destLen == nil || *destLen > uint64(len(dest)) {
		return Z_STREAM_ERROR
	}
	out, err := CompressBytes(source, level)
	if err != nil {
		return Z_STREAM_ERROR
	}
	n := copy(dest[:*destLen], out)
	*destLen = uint64(n)
	if n < len(out) {
		return Z_BUF_ERROR
	}
	return Z_OK
}
func Uncompress(dest []byte, destLen *uint64, source []byte) int32 {
	if destLen == nil || *destLen > uint64(len(dest)) {
		return Z_STREAM_ERROR
	}
	r, err := NewReader(bytes.NewReader(source))
	if err != nil {
		return Z_DATA_ERROR
	}
	defer r.Close()
	limit := *destLen
	n, err := io.ReadFull(r, dest[:limit])
	*destLen = uint64(n)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return Z_OK
	}
	if err != nil {
		return Z_DATA_ERROR
	}
	var extra [1]byte
	n, err = r.Read(extra[:])
	if n != 0 {
		return Z_BUF_ERROR
	}
	if err == io.EOF {
		return Z_OK
	}
	if err != nil {
		return Z_DATA_ERROR
	}
	return Z_BUF_ERROR
}
func Adler32(adler uint32, buf []byte) uint32 {
	if buf == nil {
		return 1
	}
	if adler == 1 {
		return adler32.Checksum(buf)
	}
	s1, s2 := adler&65535, adler>>16
	for len(buf) > 0 {
		n := len(buf)
		if n > 5552 {
			n = 5552
		}
		for _, b := range buf[:n] {
			s1 += uint32(b)
			s2 += s1
		}
		s1 %= 65521
		s2 %= 65521
		buf = buf[n:]
	}
	return s1 | s2<<16
}
func CRC32(crc uint32, buf []byte) uint32 {
	if buf == nil {
		return 0
	}
	return crc32.Update(crc, crc32.IEEETable, buf)
}
func Adler32Combine(a, b uint32, lenB int64) uint32 {
	if lenB < 0 {
		return ^uint32(0)
	}
	const base uint64 = 65521
	rem := uint64(lenB) % base
	s1 := uint64(a & 65535)
	s2 := rem * s1 % base
	s1 += (uint64(b) & 65535) + base - 1
	s2 += uint64(a>>16) + uint64(b>>16) + base - rem
	return uint32(s1%base) | uint32(s2%base)<<16
}
func gf2Times(matrix *[32]uint32, vector uint32) uint32 {
	var sum uint32
	for i := 0; vector != 0; i++ {
		if vector&1 != 0 {
			sum ^= matrix[i]
		}
		vector >>= 1
	}
	return sum
}
func gf2Square(square, matrix *[32]uint32) {
	for i := range square {
		square[i] = gf2Times(matrix, matrix[i])
	}
}
func CRC32Combine(a, b uint32, lenB int64) uint32 {
	if lenB <= 0 {
		return a
	}
	var odd, even [32]uint32
	odd[0] = 0xedb88320
	row := uint32(1)
	for i := 1; i < 32; i++ {
		odd[i] = row
		row <<= 1
	}
	gf2Square(&even, &odd)
	gf2Square(&odd, &even)
	for {
		gf2Square(&even, &odd)
		if lenB&1 != 0 {
			a = gf2Times(&even, a)
		}
		lenB >>= 1
		if lenB == 0 {
			break
		}
		gf2Square(&odd, &even)
		if lenB&1 != 0 {
			a = gf2Times(&odd, a)
		}
		lenB >>= 1
		if lenB == 0 {
			break
		}
	}
	return a ^ b
}
func GetCRCTable() [256]uint32 { return *crc32.IEEETable }

// Format selects a framing format; Auto detects gzip or zlib when reading.
type Format uint8

const (
	Zlib Format = iota
	Gzip
	Raw
	Auto
)

func NewFormatReader(r io.Reader, format Format, dict []byte) (io.ReadCloser, error) {
	switch format {
	case Zlib:
		return stdzlib.NewReaderDict(r, dict)
	case Gzip:
		return gzip.NewReader(r)
	case Raw:
		if dict != nil {
			return flate.NewReaderDict(r, dict), nil
		}
		return flate.NewReader(r), nil
	case Auto:
		b := bufioReader(r)
		head, err := b.Peek(2)
		if err != nil {
			return nil, err
		}
		if head[0] == 31 && head[1] == 139 {
			return gzip.NewReader(b)
		}
		return stdzlib.NewReaderDict(b, dict)
	}
	return nil, fmt.Errorf("zlib: invalid format %d", format)
}

type FlushWriteCloser interface {
	io.WriteCloser
	Flush() error
}

func NewFormatWriter(w io.Writer, format Format, level int, dict []byte) (FlushWriteCloser, error) {
	switch format {
	case Zlib:
		return stdzlib.NewWriterLevelDict(w, level, dict)
	case Gzip:
		if len(dict) != 0 {
			return nil, errors.New("zlib: gzip does not support preset dictionaries")
		}
		return gzip.NewWriterLevel(w, level)
	case Raw:
		if dict != nil {
			return flate.NewWriterDict(w, level, dict)
		}
		return flate.NewWriter(w, level)
	}
	return nil, fmt.Errorf("zlib: invalid output format %d", format)
}
