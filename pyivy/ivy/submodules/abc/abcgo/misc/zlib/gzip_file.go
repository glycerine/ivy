package zlib

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func bufioReader(r io.Reader) *bufio.Reader {
	if b, ok := r.(*bufio.Reader); ok {
		return b
	}
	return bufio.NewReader(r)
}

// File implements the gzFile API. Each handle is used by one goroutine at a
// time, like zlib's gzFile; independent handles may be used concurrently.
type File struct {
	file                 *os.File
	mode                 byte
	level                int
	direct               bool
	bufferSize           int
	reader               io.Reader
	buffered             *bufio.Reader
	gzipReader           *gzip.Reader
	writer               *gzip.Writer
	start, position      int64
	pushed               []byte
	eof, closed, started bool
	finished             bool
	err                  error
}
type GzFile = *File

func parseMode(mode string) (byte, int, bool, error) {
	var access byte
	level := Z_DEFAULT_COMPRESSION
	direct := false
	for _, c := range mode {
		switch {
		case c == 'r' || c == 'w' || c == 'a':
			if access != 0 {
				return 0, 0, false, errors.New("zlib: invalid mode")
			}
			access = byte(c)
		case c >= '0' && c <= '9':
			level = int(c - '0')
		case c == 'T':
			direct = true
		case c == 'b' || c == 'e' || c == 'x':
		case c == 'h':
			level = HuffmanOnly
		case c == 'f' || c == 'R' || c == 'F':
			return 0, 0, false, fmt.Errorf("zlib: unsupported compression strategy %c", c)
		default:
			return 0, 0, false, fmt.Errorf("zlib: invalid mode %q", mode)
		}
	}
	if access == 0 {
		return 0, 0, false, errors.New("zlib: missing access mode")
	}
	return access, level, direct, nil
}
func Open(path, mode string) (*File, error) {
	access, level, direct, err := parseMode(mode)
	if err != nil {
		return nil, err
	}
	flags := os.O_RDONLY
	if access == 'w' {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	if access == 'a' {
		flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}
	if strings.Contains(mode, "x") {
		flags |= os.O_EXCL
	}
	f, err := os.OpenFile(path, flags, 0666)
	if err != nil {
		return nil, err
	}
	return wrapFile(f, access, level, direct), nil
}
func wrapFile(file *os.File, access byte, level int, direct bool) *File {
	start, _ := file.Seek(0, io.SeekCurrent)
	return &File{file: file, mode: access, level: level, direct: direct, start: start, bufferSize: 8192}
}

// OpenFile transfers ownership of file to the returned handle, as gzdopen does.
func OpenFile(file *os.File, mode string) (*File, error) {
	if file == nil {
		return nil, errors.New("zlib: nil file")
	}
	access, level, direct, err := parseMode(mode)
	if err != nil {
		return nil, err
	}
	return wrapFile(file, access, level, direct), nil
}
func Gzopen(path, mode string) *File { f, _ := Open(path, mode); return f }
func Gzdopen(fd int, mode string) *File {
	access, level, direct, err := parseMode(mode)
	if err != nil || fd < 0 {
		return nil
	}
	f := os.NewFile(uintptr(fd), fmt.Sprintf("fd:%d", fd))
	if f == nil {
		return nil
	}
	return wrapFile(f, access, level, direct)
}

func (f *File) initReader() error {
	if f.reader != nil {
		return nil
	}
	f.buffered = bufio.NewReaderSize(f.file, f.bufferSize)
	head, err := f.buffered.Peek(2)
	if err != nil && err != io.EOF {
		return err
	}
	f.direct = len(head) < 2 || head[0] != 31 || head[1] != 139
	if f.direct {
		f.reader = f.buffered
		return nil
	}
	r, err := gzip.NewReader(f.buffered)
	if err != nil {
		return err
	}
	r.Multistream(false)
	f.gzipReader = r
	f.reader = r
	return nil
}
func (f *File) nextMember() error {
	head, err := f.buffered.Peek(2)
	if err != nil && err != io.EOF {
		return err
	}
	if len(head) < 2 || head[0] != 31 || head[1] != 139 {
		return io.EOF
	}
	if err = f.gzipReader.Reset(f.buffered); err != nil {
		return err
	}
	f.gzipReader.Multistream(false)
	return nil
}
func (f *File) Read(p []byte) (int, error) {
	if f == nil || f.closed || f.mode != 'r' {
		return 0, os.ErrInvalid
	}
	if len(p) == 0 {
		return 0, nil
	}
	f.started = true
	if err := f.initReader(); err != nil {
		f.err = err
		return 0, err
	}
	n := 0
	for len(f.pushed) > 0 && n < len(p) {
		last := len(f.pushed) - 1
		p[n] = f.pushed[last]
		f.pushed = f.pushed[:last]
		n++
	}
	for n < len(p) {
		m, err := f.reader.Read(p[n:])
		n += m
		if err == io.EOF && f.gzipReader != nil {
			err = f.nextMember()
			if err == nil {
				continue
			}
		}
		if err != nil {
			f.position += int64(n)
			if err == io.EOF {
				f.eof = true
			} else {
				f.err = err
			}
			return n, err
		}
		if m == 0 {
			break
		}
	}
	f.position += int64(n)
	return n, nil
}
func (f *File) initWriter() error {
	if f.writer != nil || f.direct || f.finished {
		return nil
	}
	w, err := gzip.NewWriterLevel(f.file, f.level)
	if err != nil {
		return err
	}
	f.writer = w
	return nil
}
func (f *File) Write(p []byte) (int, error) {
	if f == nil || f.closed || f.mode == 'r' {
		return 0, os.ErrInvalid
	}
	f.started = true
	f.finished = false
	if err := f.initWriter(); err != nil {
		f.err = err
		return 0, err
	}
	var n int
	var err error
	if f.direct {
		n, err = f.file.Write(p)
	} else {
		n, err = f.writer.Write(p)
	}
	f.position += int64(n)
	if err != nil {
		f.err = err
	}
	return n, err
}
func (f *File) Flush() error {
	if f == nil || f.closed || f.mode == 'r' {
		return os.ErrInvalid
	}
	err := f.initWriter()
	if err == nil && f.writer != nil {
		err = f.writer.Flush()
	}
	if err != nil {
		f.err = err
	}
	return err
}
func (f *File) Close() error {
	if f == nil || f.closed {
		return os.ErrInvalid
	}
	f.closed = true
	var result error
	if f.mode != 'r' {
		if err := f.initWriter(); err != nil {
			result = err
		}
		if f.writer != nil {
			if err := f.writer.Close(); result == nil {
				result = err
			}
		}
	} else if f.gzipReader != nil {
		result = f.gzipReader.Close()
	}
	if err := f.file.Close(); result == nil {
		result = err
	}
	if result == nil {
		result = f.err
	}
	return result
}
func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f == nil || f.closed {
		return -1, os.ErrInvalid
	}
	target := offset
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		target += f.position
	default:
		return -1, errors.New("zlib: seek end is unsupported")
	}
	if target < 0 {
		return -1, errors.New("zlib: negative seek")
	}
	if f.mode != 'r' {
		if target < f.position {
			return -1, errors.New("zlib: compressed writes cannot seek backward")
		}
		zeros := make([]byte, 4096)
		for f.position < target {
			n := target - f.position
			if n > int64(len(zeros)) {
				n = int64(len(zeros))
			}
			if _, err := f.Write(zeros[:n]); err != nil {
				return -1, err
			}
		}
		return f.position, nil
	}
	if target < f.position {
		if _, err := f.file.Seek(f.start, io.SeekStart); err != nil {
			return -1, err
		}
		if f.gzipReader != nil {
			f.gzipReader.Close()
		}
		f.reader = nil
		f.gzipReader = nil
		f.buffered = nil
		f.pushed = nil
		f.position = 0
		f.eof = false
		f.err = nil
	}
	if target > f.position {
		_, err := io.CopyN(io.Discard, f, target-f.position)
		if err != nil {
			return -1, err
		}
	}
	f.eof = false
	return f.position, nil
}
func Gzread(f *File, buffer []byte, n uint32) int32 {
	if f == nil || n > 0x7fffffff || uint64(n) > uint64(len(buffer)) {
		return -1
	}
	count, err := io.ReadFull(f, buffer[:n])
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF && count == 0 {
		return -1
	}
	return int32(count)
}
func Gzwrite(f *File, buffer []byte, n uint32) int32 {
	if f == nil || n > 0x7fffffff || uint64(n) > uint64(len(buffer)) {
		return 0
	}
	count, _ := f.Write(buffer[:n])
	return int32(count)
}
func Gzprintf(f *File, format string, args ...any) int32 {
	format, args, formatErr := gzFormat(format, args)
	if formatErr != nil {
		if f != nil {
			f.err = formatErr
		}
		return 0
	}
	text := fmt.Sprintf(format, args...)
	n, err := f.Write([]byte(text))
	if err != nil {
		return 0
	}
	return int32(n)
}
func Gzputs(f *File, s string) int32 {
	if i := strings.IndexByte(s, 0); i >= 0 {
		s = s[:i]
	}
	n, err := f.Write([]byte(s))
	if err != nil {
		return -1
	}
	return int32(n)
}

// Gzgets includes the newline and writes the C terminating NUL. It returns nil
// when no bytes were read; a buffer of size one is valid and contains just NUL.
func Gzgets(f *File, buffer []byte, n int32) []byte {
	if f == nil || f.closed || f.mode != 'r' || n <= 0 || int(n) > len(buffer) {
		return nil
	}
	k := 0
	for k < int(n)-1 {
		c := Gzgetc(f)
		if c < 0 {
			break
		}
		buffer[k] = byte(c)
		k++
		if c == '\n' {
			break
		}
	}
	buffer[k] = 0
	if k == 0 && n > 1 {
		return nil
	}
	return buffer[:k+1]
}
func Gzgetc(f *File) int32 {
	var b [1]byte
	n, _ := f.Read(b[:])
	if n == 0 {
		return -1
	}
	return int32(b[0])
}
func Gzputc(f *File, c int32) int32 {
	if n, err := f.Write([]byte{byte(c)}); err != nil || n != 1 {
		return -1
	}
	return int32(byte(c))
}
func Gzungetc(c int32, f *File) int32 {
	if f == nil || f.closed || f.mode != 'r' || c < 0 {
		return -1
	}
	f.pushed = append(f.pushed, byte(c))
	f.position--
	f.eof = false
	return int32(byte(c))
}
func Gzflush(f *File, flush int32) int32 {
	if f == nil || f.closed || f.mode == 'r' || flush < 0 || flush > Z_FINISH || flush == Z_PARTIAL_FLUSH || flush == Z_FULL_FLUSH {
		return Z_STREAM_ERROR
	}
	if flush == Z_NO_FLUSH {
		return Z_OK
	}
	if flush == Z_FINISH {
		if f.mode == 'r' || f.closed {
			return Z_STREAM_ERROR
		}
		if err := f.initWriter(); err != nil {
			return Z_ERRNO
		}
		if f.writer != nil {
			err := f.writer.Close()
			f.writer = nil
			f.finished = true
			if err != nil {
				f.err = err
				return Z_ERRNO
			}
		}
		return Z_OK
	}
	if err := f.Flush(); err != nil {
		return Z_ERRNO
	}
	return Z_OK
}
func Gzseek(f *File, offset int64, whence int) int64 {
	n, err := f.Seek(offset, whence)
	if err != nil {
		return -1
	}
	return n
}
func Gzrewind(f *File) int32 {
	if f == nil || f.closed || f.mode != 'r' {
		return -1
	}
	if _, err := f.file.Seek(f.start, io.SeekStart); err != nil {
		f.err = err
		return -1
	}
	if f.gzipReader != nil {
		f.gzipReader.Close()
	}
	f.reader = nil
	f.gzipReader = nil
	f.buffered = nil
	f.pushed = nil
	f.position = 0
	f.eof = false
	f.err = nil
	return 0
}
func Gztell(f *File) int64 {
	if f == nil || f.closed {
		return -1
	}
	return f.position
}
func Gzoffset(f *File) int64 {
	if f == nil || f.closed {
		return -1
	}
	n, err := f.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return -1
	}
	if f.buffered != nil {
		n -= int64(f.buffered.Buffered())
	}
	return n
}
func Gzeof(f *File) int32 {
	if f != nil && f.eof {
		return 1
	}
	return 0
}
func Gzdirect(f *File) int32 {
	if f == nil || f.closed {
		return 0
	}
	if f.mode == 'r' {
		if err := f.initReader(); err != nil {
			f.err = err
			return 0
		}
	}
	if f.direct {
		return 1
	}
	return 0
}
func Gzclose(f *File) int32 {
	if f == nil {
		return Z_STREAM_ERROR
	}
	if err := f.Close(); err != nil {
		return Z_ERRNO
	}
	return Z_OK
}
func GzcloseR(f *File) int32 {
	if f == nil || f.mode != 'r' {
		return Z_STREAM_ERROR
	}
	return Gzclose(f)
}
func GzcloseW(f *File) int32 {
	if f == nil || f.mode == 'r' {
		return Z_STREAM_ERROR
	}
	return Gzclose(f)
}
func Gzerror(f *File, code *int32) string {
	if f == nil {
		if code != nil {
			*code = Z_STREAM_ERROR
		}
		return "invalid gzip handle"
	}
	if f.err == nil {
		if code != nil {
			*code = Z_OK
		}
		return ""
	}
	if code != nil {
		*code = Z_DATA_ERROR
		var path *os.PathError
		if errors.As(f.err, &path) {
			*code = Z_ERRNO
		}
	}
	return f.err.Error()
}
func Gzclearerr(f *File) {
	if f != nil {
		f.err = nil
		f.eof = false
	}
}
func Gzbuffer(f *File, size uint32) int32 {
	if f == nil || f.started || f.reader != nil || size == 0 {
		return -1
	}
	f.bufferSize = int(size)
	return 0
}
func Gzsetparams(f *File, level, strategy int) int32 {
	if f == nil || f.closed || f.mode == 'r' || level < -1 || level > 9 {
		return Z_STREAM_ERROR
	}
	if strategy != Z_DEFAULT_STRATEGY && strategy != Z_HUFFMAN_ONLY {
		return Z_STREAM_ERROR
	}
	if strategy == Z_HUFFMAN_ONLY {
		level = HuffmanOnly
	}
	if f.writer != nil {
		if err := f.writer.Close(); err != nil {
			f.err = err
			return Z_ERRNO
		}
		f.writer = nil
		f.finished = true
	}
	f.level = level
	return Z_OK
}

var gzVerb = regexp.MustCompile(`%([-+# 0]*)([0-9]*|\*)?(\.(?:[0-9]*|\*))?([hljztL]*)([diouxXfFeEgGaAcspn%])`)

func gzFormat(format string, args []any) (string, []any, error) {
	args = append([]any(nil), args...)
	argument := 0
	var err error
	format = gzVerb.ReplaceAllStringFunc(format, func(spec string) string {
		m := gzVerb.FindStringSubmatch(spec)
		verb := m[5]
		if verb == "n" || verb == "a" || verb == "A" || ((verb == "s" || verb == "c") && m[4] != "") {
			err = fmt.Errorf("zlib: unsupported C format %s", spec)
			return spec
		}
		if verb != "%" {
			if m[2] == "*" {
				argument++
			}
			if m[3] == ".*" {
				argument++
			}
			if (verb == "u" || verb == "o" || verb == "x" || verb == "X") && argument < len(args) {
				args[argument] = gzUnsigned(args[argument], m[4])
			}
			if verb == "s" && argument < len(args) {
				switch v := args[argument].(type) {
				case string:
					if end := strings.IndexByte(v, 0); end >= 0 {
						args[argument] = v[:end]
					}
				case []byte:
					if end := bytes.IndexByte(v, 0); end >= 0 {
						args[argument] = v[:end]
					}
				}
			}
			argument++
		}
		if verb == "i" || verb == "u" {
			verb = "d"
		}
		return "%" + m[1] + m[2] + m[3] + verb
	})
	return format, args, err
}

func gzUnsigned(value any, length string) any {
	var n uint64
	switch v := value.(type) {
	case int:
		n = uint64(v)
	case int8:
		n = uint64(v)
	case int16:
		n = uint64(v)
	case int32:
		n = uint64(v)
	case int64:
		n = uint64(v)
	case uint:
		n = uint64(v)
	case uint8:
		n = uint64(v)
	case uint16:
		n = uint64(v)
	case uint32:
		n = uint64(v)
	case uint64:
		n = v
	case uintptr:
		n = uint64(v)
	default:
		return value
	}
	switch length {
	case "hh":
		return uint8(n)
	case "h":
		return uint16(n)
	case "l", "ll", "j", "z", "t":
		return n
	default:
		return uint32(n)
	}
}
