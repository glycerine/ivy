package abc

// Memory-buffer compression from src/misc/zlib/compress_.c and uncompr.c.

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"math"
)

type ZlibBufferOptions struct {
	MaxOutput int // zero: 64 MiB
	Stop      func() bool
}

type zlibBoundedBuffer struct {
	data  []byte
	limit int
	stop  func() bool
}

func (b *zlibBoundedBuffer) Write(p []byte) (int, error) {
	if b.stop != nil && b.stop() {
		return 0, ErrCancelled
	}
	if len(p) > b.limit-len(b.data) {
		return 0, ErrLimit
	}
	b.data = append(b.data, p...)
	return len(p), nil
}

func checkZlibBufferOptions(options ZlibBufferOptions) (ZlibBufferOptions, error) {
	if options.MaxOutput < 0 {
		return options, ErrDimension
	}
	if options.MaxOutput == 0 {
		options.MaxOutput = 64 << 20
	}
	if options.Stop != nil && options.Stop() {
		return options, ErrCancelled
	}
	return options, nil
}

// ZlibCompressBound implements compressBound with checked uint64 arithmetic.
func ZlibCompressBound(sourceLength uint64) (uint64, error) {
	parts := [...]uint64{sourceLength >> 12, sourceLength >> 14, sourceLength >> 25, 13}
	bound := sourceLength
	for _, part := range parts {
		if math.MaxUint64-bound < part {
			return 0, ErrLimit
		}
		bound += part
	}
	return bound, nil
}

// ZlibCompress implements compress2. Valid native levels are -1 and 0 through
// 9. It returns owned bytes and never exposes a partial stream on failure.
func ZlibCompress(source []byte, level int, options ZlibBufferOptions) ([]byte, error) {
	if level < zlib.DefaultCompression || level > zlib.BestCompression {
		return nil, ErrDimension
	}
	options, err := checkZlibBufferOptions(options)
	if err != nil {
		return nil, err
	}
	buffer := zlibBoundedBuffer{limit: options.MaxOutput, stop: options.Stop}
	writer, err := zlib.NewWriterLevel(&buffer, level)
	if err != nil {
		return nil, fmt.Errorf("%w: zlib level: %v", ErrDimension, err)
	}
	for offset := 0; offset < len(source); {
		if options.Stop != nil && options.Stop() {
			writer.Close()
			return nil, ErrCancelled
		}
		end := min(len(source), offset+(32<<10))
		if _, err = writer.Write(source[offset:end]); err != nil {
			writer.Close()
			return nil, err
		}
		offset = end
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	return buffer.data, nil
}

// ZlibUncompress implements uncompress with an explicit decompressed-byte cap.
// Format failures wrap ErrInvariant; resource failures remain distinguishable.
func ZlibUncompress(source []byte, options ZlibBufferOptions) ([]byte, error) {
	options, err := checkZlibBufferOptions(options)
	if err != nil {
		return nil, err
	}
	reader, err := zlib.NewReader(bytes.NewReader(source))
	if err != nil {
		return nil, fmt.Errorf("%w: zlib stream: %v", ErrInvariant, err)
	}
	buffer := zlibBoundedBuffer{limit: options.MaxOutput, stop: options.Stop}
	scratch := make([]byte, 32<<10)
	for {
		if options.Stop != nil && options.Stop() {
			reader.Close()
			return nil, ErrCancelled
		}
		count, readErr := reader.Read(scratch)
		if count > 0 {
			if _, err = buffer.Write(scratch[:count]); err != nil {
				reader.Close()
				return nil, err
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			reader.Close()
			return nil, fmt.Errorf("%w: zlib stream: %v", ErrInvariant, readErr)
		}
	}
	if err = reader.Close(); err != nil {
		return nil, fmt.Errorf("%w: zlib close: %v", ErrInvariant, err)
	}
	return buffer.data, nil
}
