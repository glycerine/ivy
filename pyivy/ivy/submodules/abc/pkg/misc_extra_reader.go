package abc

// Owned token reader from src/misc/extra/extraUtilReader.c.

import (
	"fmt"
	"os"
	"slices"
)

type ExtraFileReaderLimits struct {
	MaxBytes, MaxTokens, MaxTokenBytes, MaxOperations int
	Stop                                              func() bool
}

type ExtraTokenBatch struct {
	Tokens []string
	Lines  []int
}

const (
	extraReaderComment uint8 = iota
	extraReaderNormal
	extraReaderStop
	extraReaderClean
)

type ExtraFileReader struct {
	name     string
	data     []byte
	classes  [256]uint8
	position int
	line     int
	last     ExtraTokenBatch
	limits   ExtraFileReaderLimits
	closed   bool
}

func normalizeExtraFileReaderLimits(limits ExtraFileReaderLimits) (ExtraFileReaderLimits, error) {
	if limits.MaxBytes < 0 || limits.MaxTokens < 0 || limits.MaxTokenBytes < 0 || limits.MaxOperations < 0 {
		return limits, ErrDimension
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 1 << 30
	}
	if limits.MaxTokens == 0 {
		limits.MaxTokens = 1 << 20
	}
	if limits.MaxTokenBytes == 0 {
		limits.MaxTokenBytes = 16 << 20
	}
	if limits.MaxOperations == 0 {
		limits.MaxOperations = 1 << 30
	}
	return limits, nil
}

func newExtraFileReader(name string, data []byte, comments, stops, clean string, limits ExtraFileReaderLimits) (*ExtraFileReader, error) {
	limits, err := normalizeExtraFileReaderLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	if slices.Contains(data, byte(0)) {
		return nil, fmt.Errorf("%w: NUL byte in token input", ErrFormat)
	}
	reader := &ExtraFileReader{name: name, data: slices.Clone(data), line: 1, limits: limits}
	for i := range reader.classes {
		reader.classes[i] = extraReaderNormal
	}
	for _, set := range []struct {
		text  string
		class uint8
	}{{comments, extraReaderComment}, {stops, extraReaderStop}, {clean, extraReaderClean}} {
		for i := 0; i < len(set.text); i++ {
			reader.classes[set.text[i]] = set.class
		}
	}
	return reader, nil
}

func NewExtraByteReader(name string, data []byte, comments, stops, clean string, limits ExtraFileReaderLimits) (*ExtraFileReader, error) {
	return newExtraFileReader(name, data, comments, stops, clean, limits)
}

func NewExtraFileReader(name, comments, stops, clean string, limits ExtraFileReaderLimits) (*ExtraFileReader, error) {
	limits, err := normalizeExtraFileReaderLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	info, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	if info.Size() < 0 || info.Size() > int64(limits.MaxBytes) {
		return nil, ErrLimit
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return newExtraFileReader(name, data, comments, stops, clean, limits)
}

func (reader *ExtraFileReader) check() error {
	if reader == nil {
		return ErrDimension
	}
	if reader.closed || reader.position < 0 || reader.position > len(reader.data) || reader.line < 1 {
		return ErrInvariant
	}
	return nil
}

func (reader *ExtraFileReader) FileName() (string, error) {
	if err := reader.check(); err != nil {
		return "", err
	}
	return reader.name, nil
}

func (reader *ExtraFileReader) FileSize() (int, error) {
	if err := reader.check(); err != nil {
		return 0, err
	}
	return len(reader.data), nil
}

func (reader *ExtraFileReader) Position() (int, error) {
	if err := reader.check(); err != nil {
		return 0, err
	}
	return reader.position, nil
}

func (reader *ExtraFileReader) LineNumber(token int) (int, error) {
	if err := reader.check(); err != nil {
		return 0, err
	}
	if token < 0 || token >= len(reader.last.Lines) {
		return 0, ErrDimension
	}
	return reader.last.Lines[token], nil
}

func cloneExtraTokenBatch(batch ExtraTokenBatch) ExtraTokenBatch {
	return ExtraTokenBatch{Tokens: slices.Clone(batch.Tokens), Lines: slices.Clone(batch.Lines)}
}

func (reader *ExtraFileReader) nextInternal(position, line int) (ExtraTokenBatch, int, int, bool, int, error) {
	batch := ExtraTokenBatch{}
	operations := 0
	tokenStart, tokenLine := -1, 0
	finishToken := func(end int) error {
		if tokenStart < 0 {
			return nil
		}
		if end-tokenStart > reader.limits.MaxTokenBytes || len(batch.Tokens) == reader.limits.MaxTokens {
			return ErrLimit
		}
		batch.Tokens = append(batch.Tokens, string(reader.data[tokenStart:end]))
		batch.Lines = append(batch.Lines, tokenLine)
		tokenStart = -1
		return nil
	}
	for position < len(reader.data) {
		if operations == reader.limits.MaxOperations {
			return ExtraTokenBatch{}, 0, 0, false, operations, ErrLimit
		}
		if operations&255 == 0 && reader.limits.Stop != nil && reader.limits.Stop() {
			return ExtraTokenBatch{}, 0, 0, false, operations, ErrCancelled
		}
		character := reader.data[position]
		operations++
		if character == '\n' {
			line++
		}
		switch reader.classes[character] {
		case extraReaderComment:
			if character == '/' && (position+1 >= len(reader.data) || reader.data[position+1] != '/') {
				if tokenStart < 0 {
					tokenStart, tokenLine = position, line
				}
				position++
				continue
			}
			if err := finishToken(position); err != nil {
				return ExtraTokenBatch{}, 0, 0, false, operations, err
			}
			for position < len(reader.data) && reader.data[position] != '\n' {
				if operations == reader.limits.MaxOperations {
					return ExtraTokenBatch{}, 0, 0, false, operations, ErrLimit
				}
				position++
				operations++
			}
			if position == len(reader.data) {
				return ExtraTokenBatch{}, 0, 0, false, operations, fmt.Errorf("%w: comment reaches end of input", ErrFormat)
			}
			continue
		case extraReaderNormal:
			if tokenStart < 0 {
				tokenStart, tokenLine = position, line
			}
			position++
		case extraReaderStop:
			if err := finishToken(position); err != nil {
				return ExtraTokenBatch{}, 0, 0, false, operations, err
			}
			position++
			return batch, position, line, true, operations, nil
		case extraReaderClean:
			if err := finishToken(position); err != nil {
				return ExtraTokenBatch{}, 0, 0, false, operations, err
			}
			position++
		default:
			return ExtraTokenBatch{}, 0, 0, false, operations, ErrInvariant
		}
	}
	if err := finishToken(position); err != nil {
		return ExtraTokenBatch{}, 0, 0, false, operations, err
	}
	return batch, position, line, false, operations, nil
}

func (reader *ExtraFileReader) Next() (ExtraTokenBatch, bool, error) {
	if err := reader.check(); err != nil {
		return ExtraTokenBatch{}, false, err
	}
	position, line, total := reader.position, reader.line, 0
	for {
		batch, nextPosition, nextLine, stopped, operations, err := reader.nextInternal(position, line)
		total += operations
		if total > reader.limits.MaxOperations {
			return ExtraTokenBatch{}, false, ErrLimit
		}
		if err != nil {
			return ExtraTokenBatch{}, false, err
		}
		position, line = nextPosition, nextLine
		if len(batch.Tokens) != 0 {
			reader.position, reader.line = position, line
			reader.last = cloneExtraTokenBatch(batch)
			return cloneExtraTokenBatch(batch), true, nil
		}
		if !stopped {
			reader.position, reader.line = position, line
			reader.last = ExtraTokenBatch{}
			return ExtraTokenBatch{}, false, nil
		}
	}
}

func (reader *ExtraFileReader) Close() bool {
	if reader == nil || reader.closed {
		return false
	}
	reader.data, reader.last = nil, ExtraTokenBatch{}
	reader.closed = true
	return true
}
