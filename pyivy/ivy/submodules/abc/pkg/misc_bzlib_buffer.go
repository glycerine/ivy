package abc

// Buffer decompression and metadata from src/misc/bzlib/bzlib.c.

import (
	"bytes"
	"compress/bzip2"
	"fmt"
	"io"
)

const BZip2Version = "1.0.5, 10-Dec-2007"

// DecompressBZip2 returns an owned decompressed buffer. A zero maximumOutput
// selects a 1 GiB default. Failures and cancellation expose no partial bytes.
func DecompressBZip2(source []byte, maximumOutput int64, stop func() bool) ([]byte, error) {
	if maximumOutput < 0 {
		return nil, ErrDimension
	}
	if maximumOutput == 0 {
		maximumOutput = 1 << 30
	}
	if stop != nil && stop() {
		return nil, ErrCancelled
	}
	reader := bzip2.NewReader(bytes.NewReader(source))
	result := make([]byte, 0, min(maximumOutput, 32<<10))
	buffer := make([]byte, 32<<10)
	for {
		if stop != nil && stop() {
			return nil, ErrCancelled
		}
		count, err := reader.Read(buffer)
		if int64(count) > maximumOutput-int64(len(result)) {
			return nil, ErrLimit
		}
		result = append(result, buffer[:count]...)
		if err == io.EOF {
			return result, nil
		}
		if err != nil {
			return nil, fmt.Errorf("%w: bzip2 decode: %v", ErrFormat, err)
		}
		if count == 0 {
			return nil, ErrInvariant
		}
	}
}
