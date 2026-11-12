package glucose2

/************************************************************************************[ParseUtils.h]
Copyright (c) 2003-2006, Niklas Een, Niklas Sorensson
Copyright (c) 2007-2010, Niklas Sorensson

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// StreamBuffer reads plain or gzip-compressed DIMACS without C zlib handles.
// Position is the absolute byte offset in decompressed input.
type StreamBuffer struct {
	reader       *bufio.Reader
	offset       int64
	line, column int
	err          error
	closer       io.Closer
}

func NewStreamBuffer(input io.Reader) (*StreamBuffer, error) {
	b := bufio.NewReaderSize(input, 1<<20)
	magic, _ := b.Peek(2)
	s := &StreamBuffer{reader: b, line: 1, column: 1}
	if len(magic) == 2 && magic[0] == 0x1f && magic[1] == 0x8b {
		z, err := gzip.NewReader(b)
		if err != nil {
			return nil, err
		}
		s.reader = bufio.NewReaderSize(z, 1<<20)
		s.closer = z
	}
	return s, nil
}
func (s *StreamBuffer) Peek() int {
	p, err := s.reader.Peek(1)
	if err != nil {
		if err != io.EOF {
			s.err = err
		}
		return -1
	}
	return int(p[0])
}
func (s *StreamBuffer) Advance() {
	b, err := s.reader.ReadByte()
	if err != nil {
		if err != io.EOF {
			s.err = err
		}
		return
	}
	s.offset++
	if b == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}
}
func (s *StreamBuffer) Position() int64        { return s.offset }
func (s *StreamBuffer) LineColumn() (int, int) { return s.line, s.column }
func (s *StreamBuffer) Err() error             { return s.err }
func (s *StreamBuffer) Close() error {
	if s.closer != nil {
		return s.closer.Close()
	}
	return nil
}
func IsEOF(s *StreamBuffer) bool { return s.Peek() < 0 }
func SkipWhitespace(s *StreamBuffer) {
	for c := s.Peek(); c == 32 || c >= 9 && c <= 13; c = s.Peek() {
		s.Advance()
	}
}
func SkipLine(s *StreamBuffer) {
	for c := s.Peek(); c >= 0; c = s.Peek() {
		s.Advance()
		if c == '\n' {
			break
		}
	}
}
func ParseInt(s *StreamBuffer) (int, error) {
	SkipWhitespace(s)
	var token strings.Builder
	c := s.Peek()
	if c == '-' || c == '+' {
		token.WriteByte(byte(c))
		s.Advance()
	}
	if s.Peek() < '0' || s.Peek() > '9' {
		return 0, fmt.Errorf("line %d column %d: expected integer", s.line, s.column)
	}
	for c = s.Peek(); c >= '0' && c <= '9'; c = s.Peek() {
		token.WriteByte(byte(c))
		s.Advance()
	}
	value, err := strconv.ParseInt(token.String(), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("line %d column %d: %w", s.line, s.column, err)
	}
	return int(value), nil
}
func ParseDouble(s *StreamBuffer) (float64, error) {
	SkipWhitespace(s)
	var token strings.Builder
	for c := s.Peek(); c >= 0 && c != ' ' && !(c >= 9 && c <= 13); c = s.Peek() {
		token.WriteByte(byte(c))
		s.Advance()
	}
	v, err := strconv.ParseFloat(token.String(), 64)
	if err != nil {
		return 0, fmt.Errorf("line %d column %d: %w", s.line, s.column, err)
	}
	return v, nil
}

// Match advances only if the whole string matches. EagerMatch retains the
// original eager iterator behavior, consuming the matching prefix on failure.
func Match(s *StreamBuffer, text string) bool {
	p, err := s.reader.Peek(len(text))
	if err != nil || string(p) != text {
		return false
	}
	for i := 0; i < len(text); i++ {
		s.Advance()
	}
	return true
}
func EagerMatch(s *StreamBuffer, text string) bool {
	for i := 0; i < len(text); i++ {
		if s.Peek() != int(text[i]) {
			return false
		}
		s.Advance()
	}
	return true
}
