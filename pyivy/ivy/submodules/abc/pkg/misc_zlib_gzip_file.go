package abc

// Concrete gzip file lifecycle from src/misc/zlib/gzlib.c, gzread.c,
// gzwrite.c and gzclose.c.

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

type GZIPReaderOptions struct {
	BufferSize int
	MaxOutput  int64 // zero: 1 GiB
	Stop       func() bool
}

type GZIPReader struct {
	path       string
	file       *os.File
	start      int64
	seekable   bool
	buffer     *bufio.Reader
	stream     *gzip.Reader
	bufferSize int
	maxOutput  int64
	stop       func() bool
	position   int64
	direct     bool
	eof        bool
	closed     bool
	err        error
	pushback   []byte
	pending    bool
	pendingTo  int64
}

func OpenGZIPReader(path string, options GZIPReaderOptions) (*GZIPReader, error) {
	if path == "" {
		return nil, ErrDimension
	}
	options, err := checkGZIPReaderOptions(options)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader, err := openGZIPReaderFile(file, path, options)
	if err != nil {
		file.Close()
		return nil, err
	}
	return reader, nil
}

// OpenGZIPReaderFile implements the descriptor-backed behavior of gzdopen.
// On success, Close owns and closes file. Reading begins at its current offset.
func OpenGZIPReaderFile(file *os.File, options GZIPReaderOptions) (*GZIPReader, error) {
	if file == nil {
		return nil, ErrDimension
	}
	options, err := checkGZIPReaderOptions(options)
	if err != nil {
		return nil, err
	}
	return openGZIPReaderFile(file, file.Name(), options)
}

func checkGZIPReaderOptions(options GZIPReaderOptions) (GZIPReaderOptions, error) {
	if options.BufferSize < 0 || options.MaxOutput < 0 {
		return options, ErrDimension
	}
	if options.BufferSize == 0 {
		options.BufferSize = 32 << 10
	}
	if options.MaxOutput == 0 {
		options.MaxOutput = 1 << 30
	}
	if options.Stop != nil && options.Stop() {
		return options, ErrCancelled
	}
	return options, nil
}

func openGZIPReaderFile(file *os.File, path string, options GZIPReaderOptions) (*GZIPReader, error) {
	start, err := file.Seek(0, io.SeekCurrent)
	seekable := err == nil
	if !seekable {
		start = 0
	}
	reader := &GZIPReader{path: path, file: file, start: start, seekable: seekable, bufferSize: options.BufferSize, maxOutput: options.MaxOutput, stop: options.Stop}
	if err = reader.initializeStream(); err != nil {
		return nil, err
	}
	return reader, nil
}

func (reader *GZIPReader) initializeStream() error {
	if reader.stream != nil {
		reader.stream.Close()
		reader.stream = nil
	}
	reader.buffer = bufio.NewReaderSize(reader.file, reader.bufferSize)
	header, err := reader.buffer.Peek(2)
	if err != nil && err != io.EOF && err != bufio.ErrBufferFull {
		return err
	}
	reader.direct = len(header) < 2 || header[0] != 0x1f || header[1] != 0x8b
	if !reader.direct {
		reader.stream, err = gzip.NewReader(reader.buffer)
		if err != nil {
			return fmt.Errorf("%w: gzip header: %v", ErrInvariant, err)
		}
		reader.stream.Multistream(false)
	}
	reader.position, reader.eof, reader.pushback, reader.pending, reader.err = 0, false, reader.pushback[:0], false, nil
	return nil
}

func (reader *GZIPReader) resetStream() error {
	if !reader.seekable {
		return ErrDimension
	}
	if _, err := reader.file.Seek(reader.start, io.SeekStart); err != nil {
		return err
	}
	return reader.initializeStream()
}

func (reader *GZIPReader) sourceRead(buffer []byte) (int, error) {
	written := 0
	for written < len(buffer) {
		if reader.direct {
			count, err := reader.buffer.Read(buffer[written:])
			written += count
			if err != nil {
				return written, err
			}
			if count == 0 {
				return written, nil
			}
			continue
		}
		count, err := reader.stream.Read(buffer[written:])
		written += count
		if err != io.EOF {
			if err != nil || count == 0 {
				return written, err
			}
			continue
		}
		advanced, err := reader.advanceReadStream()
		if err != nil {
			return written, err
		}
		if !advanced {
			return written, io.EOF
		}
	}
	return written, nil
}

// advanceReadStream follows native gzread LOOK behavior after a verified gzip
// trailer: another gzip header starts a new member, while any other remaining
// bytes switch the reader to transparent direct copying.
func (reader *GZIPReader) advanceReadStream() (bool, error) {
	header, err := reader.buffer.Peek(2)
	if err != nil && err != io.EOF && err != bufio.ErrBufferFull {
		return false, err
	}
	if len(header) == 0 {
		return false, nil
	}
	if len(header) >= 2 && header[0] == 0x1f && header[1] == 0x8b {
		if err = reader.stream.Reset(reader.buffer); err != nil {
			return false, fmt.Errorf("%w: gzip member header: %v", ErrInvariant, err)
		}
		reader.stream.Multistream(false)
		return true, nil
	}
	if err = reader.stream.Close(); err != nil {
		return false, err
	}
	reader.stream = nil
	reader.direct = true
	return true, nil
}

func (reader *GZIPReader) Read(buffer []byte) (int, error) {
	if reader == nil || reader.closed {
		return 0, ErrDimension
	}
	if reader.err != nil {
		return 0, reader.err
	}
	if reader.stop != nil && reader.stop() {
		reader.err = ErrCancelled
		return 0, reader.err
	}
	if len(buffer) == 0 {
		return 0, nil
	}
	if err := reader.applyPendingSeek(); err != nil {
		return 0, err
	}
	written := 0
	for written < len(buffer) && len(reader.pushback) != 0 {
		last := len(reader.pushback) - 1
		buffer[written] = reader.pushback[last]
		reader.pushback = reader.pushback[:last]
		reader.position++
		written++
	}
	if written == len(buffer) {
		return written, nil
	}
	remaining := reader.maxOutput - reader.position
	if remaining <= 0 {
		reader.err = ErrLimit
		return written, reader.err
	}
	want := len(buffer) - written
	if int64(want) > remaining {
		want = int(remaining)
	}
	count, err := reader.sourceRead(buffer[written : written+want])
	reader.position += int64(count)
	written += count
	if err == io.EOF {
		reader.eof = true
		if written != 0 {
			return written, nil
		}
		return 0, io.EOF
	}
	if err != nil {
		reader.err = fmt.Errorf("%w: gzip read: %v", ErrInvariant, err)
		return written, reader.err
	}
	return written, nil
}

func (reader *GZIPReader) applyPendingSeek() error {
	if !reader.pending {
		return nil
	}
	target := reader.pendingTo
	reader.pending = false
	scratch := make([]byte, 32<<10)
	for reader.position < target {
		if reader.stop != nil && reader.stop() {
			reader.err = ErrCancelled
			return reader.err
		}
		want := min(int64(len(scratch)), target-reader.position)
		count, err := reader.sourceRead(scratch[:want])
		reader.position += int64(count)
		if err == io.EOF {
			reader.eof = true
			return nil
		}
		if err != nil {
			reader.err = fmt.Errorf("%w: gzip seek: %v", ErrInvariant, err)
			return reader.err
		}
		if count == 0 {
			reader.err = ErrInvariant
			return reader.err
		}
	}
	return nil
}

func (reader *GZIPReader) ReadByte() (byte, error) {
	var one [1]byte
	count, err := reader.Read(one[:])
	if count == 1 {
		return one[0], nil
	}
	return 0, err
}

// PushByte implements gzungetc with a bounded LIFO. As in the native code,
// successful pushes decrement the logical position and pushed bytes need not
// equal bytes previously read.
func (reader *GZIPReader) PushByte(value byte) error {
	if reader == nil || reader.closed || reader.err != nil {
		return ErrDimension
	}
	if err := reader.applyPendingSeek(); err != nil {
		return err
	}
	if len(reader.pushback) >= reader.bufferSize*2 {
		reader.err = ErrLimit
		return reader.err
	}
	reader.pushback = append(reader.pushback, value)
	reader.position--
	return nil
}

// ReadLine implements gzgets: at most maximum-1 bytes, including a newline.
func (reader *GZIPReader) ReadLine(maximum int) ([]byte, error) {
	if maximum < 1 {
		return nil, ErrDimension
	}
	result := make([]byte, 0, maximum-1)
	for len(result) < maximum-1 {
		value, err := reader.ReadByte()
		if err == io.EOF {
			if len(result) != 0 {
				return result, nil
			}
			return nil, io.EOF
		}
		if err != nil {
			return nil, err
		}
		result = append(result, value)
		if value == '\n' {
			break
		}
	}
	return result, nil
}

func (reader *GZIPReader) Rewind() error {
	if reader == nil || reader.closed {
		return ErrDimension
	}
	return reader.resetStream()
}

func (reader *GZIPReader) Seek(offset int64, whence int) (int64, error) {
	if reader == nil || reader.closed || (whence != io.SeekStart && whence != io.SeekCurrent) {
		return -1, ErrDimension
	}
	target := offset
	if whence == io.SeekCurrent {
		target += reader.Tell()
	}
	if target < 0 || target > reader.maxOutput {
		return -1, ErrDimension
	}
	current := reader.Tell()
	if target < current {
		if err := reader.resetStream(); err != nil {
			return -1, err
		}
	}
	reader.pushback = reader.pushback[:0]
	reader.pending, reader.pendingTo, reader.eof = target != reader.position, target, false
	return target, nil
}

func (reader *GZIPReader) Tell() int64 {
	if reader == nil {
		return -1
	}
	if reader.pending {
		return reader.pendingTo
	}
	return reader.position
}
func (reader *GZIPReader) Direct() bool { return reader != nil && reader.direct }
func (reader *GZIPReader) EOF() bool {
	return reader != nil && reader.eof && len(reader.pushback) == 0 && !reader.pending
}

// Offset reports the effective compressed/raw descriptor offset. It excludes
// bytes already buffered but not consumed, matching gzoffset64.
func (reader *GZIPReader) Offset() (int64, error) {
	if reader == nil || reader.closed || !reader.seekable {
		return -1, ErrDimension
	}
	offset, err := reader.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return -1, err
	}
	return offset - int64(reader.buffer.Buffered()), nil
}
func (reader *GZIPReader) LastError() error {
	if reader == nil {
		return ErrDimension
	}
	return reader.err
}
func (reader *GZIPReader) ClearError() {
	if reader != nil && !reader.closed {
		reader.err, reader.eof = nil, false
	}
}
func (reader *GZIPReader) Close() error {
	if reader == nil || reader.closed {
		return ErrDimension
	}
	reader.closed = true
	var first error
	if reader.stream != nil {
		first = reader.stream.Close()
	}
	if err := reader.file.Close(); first == nil {
		first = err
	}
	return first
}

type GZIPWriterOptions struct {
	Level      int
	BufferSize int
	MaxInput   int64 // zero: 1 GiB
	Append     bool
	Stop       func() bool
}

type GZIPWriter struct {
	file     *os.File
	buffer   *bufio.Writer
	stream   *gzip.Writer
	level    int
	maxInput int64
	stop     func() bool
	position int64
	closed   bool
	err      error
}

func OpenGZIPWriter(path string, options GZIPWriterOptions) (*GZIPWriter, error) {
	if path == "" {
		return nil, ErrDimension
	}
	options, err := checkGZIPWriterOptions(options)
	if err != nil {
		return nil, err
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if options.Append {
		flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}
	file, err := os.OpenFile(path, flags, 0666)
	if err != nil {
		return nil, err
	}
	return openGZIPWriterFile(file, options), nil
}

// OpenGZIPWriterFile implements descriptor-backed gzdopen behavior. On
// success, Close owns and closes file and writes at its current descriptor
// position (or according to its O_APPEND flag).
func OpenGZIPWriterFile(file *os.File, options GZIPWriterOptions) (*GZIPWriter, error) {
	if file == nil {
		return nil, ErrDimension
	}
	options, err := checkGZIPWriterOptions(options)
	if err != nil {
		return nil, err
	}
	return openGZIPWriterFile(file, options), nil
}

func checkGZIPWriterOptions(options GZIPWriterOptions) (GZIPWriterOptions, error) {
	if options.BufferSize < 0 || options.MaxInput < 0 || options.Level < gzip.DefaultCompression || options.Level > gzip.BestCompression {
		return options, ErrDimension
	}
	if options.BufferSize == 0 {
		options.BufferSize = 32 << 10
	}
	if options.MaxInput == 0 {
		options.MaxInput = 1 << 30
	}
	if options.Stop != nil && options.Stop() {
		return options, ErrCancelled
	}
	return options, nil
}

func openGZIPWriterFile(file *os.File, options GZIPWriterOptions) *GZIPWriter {
	return &GZIPWriter{file: file, buffer: bufio.NewWriterSize(file, options.BufferSize), level: options.Level, maxInput: options.MaxInput, stop: options.Stop}
}

func (writer *GZIPWriter) ensureStream() error {
	if writer.stream != nil {
		return nil
	}
	stream, err := gzip.NewWriterLevel(writer.buffer, writer.level)
	if err != nil {
		writer.err = fmt.Errorf("%w: gzip level: %v", ErrDimension, err)
		return writer.err
	}
	writer.stream = stream
	return nil
}

func (writer *GZIPWriter) Write(data []byte) (int, error) {
	if writer == nil || writer.closed {
		return 0, ErrDimension
	}
	if writer.err != nil {
		return 0, writer.err
	}
	if writer.stop != nil && writer.stop() {
		writer.err = ErrCancelled
		return 0, writer.err
	}
	if int64(len(data)) > writer.maxInput-writer.position {
		writer.err = ErrLimit
		return 0, writer.err
	}
	if len(data) == 0 {
		return 0, nil
	}
	if err := writer.ensureStream(); err != nil {
		return 0, err
	}
	count, err := writer.stream.Write(data)
	writer.position += int64(count)
	if err != nil {
		writer.err = err
	}
	return count, err
}

func (writer *GZIPWriter) WriteByte(value byte) error {
	_, err := writer.Write([]byte{value})
	return err
}
func (writer *GZIPWriter) WriteString(value string) (int, error) {
	return writer.Write([]byte(value))
}
func (writer *GZIPWriter) Tell() int64 {
	if writer == nil {
		return -1
	}
	return writer.position
}

// Offset reports the descriptor offset, excluding compressed bytes still held
// by the Go buffer, matching the native descriptor-based definition.
func (writer *GZIPWriter) Offset() (int64, error) {
	if writer == nil || writer.closed {
		return -1, ErrDimension
	}
	return writer.file.Seek(0, io.SeekCurrent)
}

func (writer *GZIPWriter) Seek(offset int64, whence int) (int64, error) {
	if writer == nil || writer.closed || (whence != io.SeekStart && whence != io.SeekCurrent) {
		return -1, ErrDimension
	}
	target := offset
	if whence == io.SeekCurrent {
		target += writer.position
	}
	if target < writer.position || target > writer.maxInput {
		return -1, ErrDimension
	}
	zeroes := make([]byte, 32<<10)
	for writer.position < target {
		count := min(int64(len(zeroes)), target-writer.position)
		if _, err := writer.Write(zeroes[:count]); err != nil {
			return -1, err
		}
	}
	return writer.position, nil
}

func (writer *GZIPWriter) SetLevel(level int) error {
	if writer == nil || writer.closed || writer.stream != nil || level < gzip.DefaultCompression || level > gzip.BestCompression {
		return ErrDimension
	}
	writer.level = level
	return nil
}

func (writer *GZIPWriter) Flush() error {
	if writer == nil || writer.closed {
		return ErrDimension
	}
	if writer.err != nil {
		return writer.err
	}
	if err := writer.ensureStream(); err != nil {
		return err
	}
	if err := writer.stream.Flush(); err != nil {
		writer.err = err
		return err
	}
	if err := writer.buffer.Flush(); err != nil {
		writer.err = err
		return err
	}
	return nil
}

func (writer *GZIPWriter) LastError() error {
	if writer == nil {
		return ErrDimension
	}
	return writer.err
}

func (writer *GZIPWriter) Close() error {
	if writer == nil || writer.closed {
		return ErrDimension
	}
	writer.closed = true
	first := writer.err
	if first == nil {
		first = writer.ensureStream()
	}
	if writer.stream != nil {
		if err := writer.stream.Close(); first == nil {
			first = err
		}
	}
	if err := writer.buffer.Flush(); first == nil {
		first = err
	}
	if err := writer.file.Close(); first == nil {
		first = err
	}
	return first
}
