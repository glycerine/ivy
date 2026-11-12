package bzlib

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	BZ_RUN int32 = iota
	BZ_FLUSH
	BZ_FINISH
)
const (
	BZ_OK               int32 = 0
	BZ_RUN_OK           int32 = 1
	BZ_FLUSH_OK         int32 = 2
	BZ_FINISH_OK        int32 = 3
	BZ_STREAM_END       int32 = 4
	BZ_SEQUENCE_ERROR   int32 = -1
	BZ_PARAM_ERROR      int32 = -2
	BZ_MEM_ERROR        int32 = -3
	BZ_DATA_ERROR       int32 = -4
	BZ_DATA_ERROR_MAGIC int32 = -5
	BZ_IO_ERROR         int32 = -6
	BZ_UNEXPECTED_EOF   int32 = -7
	BZ_OUTBUFF_FULL     int32 = -8
	BZ_CONFIG_ERROR     int32 = -9
	BZ_MAX_UNUSED             = 5000
)

var ErrOutputFull = errors.New("bzlib: output buffer full")
var ErrMagic = errors.New("bzlib: invalid bzip2 header")
var ErrSequence = errors.New("bzlib: operation is not valid in this stream state")
var ErrParameters = errors.New("bzlib: invalid parameters")

// NewReader uses the standard-library decoder, including concatenated streams.
func NewReader(r io.Reader) io.Reader { return bzip2.NewReader(r) }
func BZ2_bzlibVersion() string        { return "1.0.5 Go replacement" }
func errorCode(err error) int32 {
	if err == nil {
		return BZ_OK
	}
	if errors.Is(err, ErrOutputFull) {
		return BZ_OUTBUFF_FULL
	}
	if errors.Is(err, io.EOF) {
		return BZ_STREAM_END
	}
	if errors.Is(err, ErrMagic) {
		return BZ_DATA_ERROR_MAGIC
	}
	if errors.Is(err, ErrSequence) {
		return BZ_SEQUENCE_ERROR
	}
	if errors.Is(err, ErrParameters) {
		return BZ_PARAM_ERROR
	}
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return BZ_UNEXPECTED_EOF
	}
	var path *os.PathError
	if errors.As(err, &path) {
		return BZ_IO_ERROR
	}
	var structural bzip2.StructuralError
	if errors.As(err, &structural) {
		return BZ_DATA_ERROR
	}
	return BZ_IO_ERROR
}
func setCode(p *int32, code int32) {
	if p != nil {
		*p = code
	}
}
func validOptions(verbosity, small int32) bool {
	return verbosity >= 0 && verbosity <= 4 && (small == 0 || small == 1)
}

// File adapts ABC's BZFILE calls to Go readers and writers. ReadOpen/WriteOpen
// borrow their underlying file. Open and Bzdopen own the descriptor instead.
type File struct {
	buffered *bufio.Reader
	reader   io.Reader
	writer   *Writer
	owner    io.Closer
	closed   bool
	err      error
	lastCode int32
}
type BZFILE = File

func (f *File) record(err error) { f.err = err; f.lastCode = errorCode(err) }
func (f *File) Read(p []byte) (int, error) {
	if f == nil || f.closed || f.writer != nil {
		return 0, ErrSequence
	}
	if len(p) == 0 {
		return 0, nil
	}
	if f.lastCode == BZ_STREAM_END {
		return 0, io.EOF
	}
	if f.lastCode < 0 && f.err != nil {
		return 0, f.err
	}
	if f.reader == nil {
		head, err := f.buffered.Peek(4)
		// Report a magic mismatch as soon as it is provable. ABC uses that status
		// to fall back to plain AIGER/BLIF input, including very short text files.
		for i, want := range []byte{'B', 'Z', 'h'} {
			if i < len(head) && head[i] != want {
				f.record(ErrMagic)
				return 0, ErrMagic
			}
		}
		if len(head) == 4 && (head[3] < '1' || head[3] > '9') {
			f.record(ErrMagic)
			return 0, ErrMagic
		}
		if len(head) < 4 {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			f.record(err)
			return 0, err
		}

		f.reader = NewReader(f.buffered)
	}
	n, err := f.reader.Read(p)
	f.record(err)
	return n, err
}
func (f *File) Write(p []byte) (int, error) {
	if f == nil || f.closed || f.writer == nil {
		return 0, ErrSequence
	}
	n, err := f.writer.Write(p)
	f.record(err)
	return n, err
}
func (f *File) Flush() error {
	if f == nil || f.closed || f.writer == nil {
		return ErrSequence
	}
	err := f.writer.Flush()
	f.record(err)
	return err
}
func (f *File) Close() error {
	if f == nil || f.closed {
		return ErrSequence
	}
	f.closed = true
	var err error
	if f.writer != nil {
		err = f.writer.Close()
	}
	if f.owner != nil {
		closeErr := f.owner.Close()
		if err == nil {
			err = closeErr
		}
	}
	f.record(err)
	return err
}
func BZ2_bzReadOpen(code *int32, r io.Reader, verbosity, small int32, unused []byte, nUnused int32) *File {
	if r == nil || !validOptions(verbosity, small) || nUnused < 0 || nUnused > BZ_MAX_UNUSED || int(nUnused) > len(unused) {
		setCode(code, BZ_PARAM_ERROR)
		return nil
	}
	source := r
	if nUnused > 0 {
		source = io.MultiReader(bytes.NewReader(append([]byte(nil), unused[:nUnused]...)), r)
	}
	f := &File{buffered: bufio.NewReader(source)}
	setCode(code, BZ_OK)
	return f
}
func BZ2_bzRead(code *int32, f *File, buffer []byte, length int32) int32 {
	if f == nil || buffer == nil || length < 0 || int(length) > len(buffer) {
		setFileCode(code, f, BZ_PARAM_ERROR)
		return 0
	}
	if f.writer != nil || f.closed {
		setFileCode(code, f, BZ_SEQUENCE_ERROR)
		return 0
	}
	if length == 0 {
		setCode(code, BZ_OK)
		return 0
	}
	n, err := io.ReadFull(f, buffer[:length])
	if err == io.ErrUnexpectedEOF && f.lastCode == BZ_STREAM_END {
		err = io.EOF
	}
	if err != nil && err != io.EOF {
		setCode(code, errorCode(err))
		return 0
	}
	setCode(code, errorCode(err))
	return int32(n)
}
func BZ2_bzReadClose(code *int32, f *File) {
	if f == nil {
		setCode(code, BZ_OK)
		return
	}
	if f.writer != nil || f.closed {
		setFileCode(code, f, BZ_SEQUENCE_ERROR)
		return
	}
	f.closed = true
	setCode(code, BZ_OK)
}

// ReadGetUnused exposes buffered bytes after the standard decoder reaches EOF.
// Unlike the C decoder, this decoder processes all concatenated members first.
func BZ2_bzReadGetUnused(code *int32, f *File, unused *[]byte, nUnused *int32) {
	if f == nil || unused == nil || nUnused == nil {
		setFileCode(code, f, BZ_PARAM_ERROR)
		return
	}
	if f.closed || f.writer != nil || f.lastCode != BZ_STREAM_END {
		setFileCode(code, f, BZ_SEQUENCE_ERROR)
		return
	}
	n := f.buffered.Buffered()
	data, _ := f.buffered.Peek(n)
	*unused = append([]byte(nil), data...)
	*nUnused = int32(n)
	setCode(code, BZ_OK)
}
func BZ2_bzWriteOpen(code *int32, w io.Writer, blockSize, verbosity, workFactor int32) *File {
	if w == nil || blockSize < 1 || blockSize > 9 || verbosity < 0 || verbosity > 4 || workFactor < 0 || workFactor > 250 {
		setCode(code, BZ_PARAM_ERROR)
		return nil
	}
	writer, err := NewWriterLevel(w, int(blockSize))
	if err != nil {
		setCode(code, errorCode(err))
		return nil
	}
	setCode(code, BZ_OK)
	return &File{writer: writer}
}
func BZ2_bzWrite(code *int32, f *File, buffer []byte, length int32) {
	if f == nil || buffer == nil || length < 0 || int(length) > len(buffer) {
		setFileCode(code, f, BZ_PARAM_ERROR)
		return
	}
	if f.writer == nil || f.closed {
		setFileCode(code, f, BZ_SEQUENCE_ERROR)
		return
	}
	_, err := f.Write(buffer[:length])
	setCode(code, errorCode(err))
}
func BZ2_bzWriteClose(code *int32, f *File, abandon int32, nIn, nOut *uint32) {
	BZ2_bzWriteClose64(code, f, abandon, nIn, nil, nOut, nil)
}
func BZ2_bzWriteClose64(code *int32, f *File, abandon int32, inLo, inHi, outLo, outHi *uint32) {
	for _, p := range []*uint32{inLo, inHi, outLo, outHi} {
		if p != nil {
			*p = 0
		}
	}
	if f == nil {
		setCode(code, BZ_OK)
		return
	}
	if f.closed || f.writer == nil {
		setFileCode(code, f, BZ_SEQUENCE_ERROR)
		return
	}
	var err error
	if abandon != 0 {
		f.writer.abandon()
	} else {
		err = f.writer.Close()
	}
	f.closed = true
	f.record(err)
	values := []uint32{uint32(f.writer.TotalIn), uint32(f.writer.TotalIn >> 32), uint32(f.writer.TotalOut), uint32(f.writer.TotalOut >> 32)}
	for i, p := range []*uint32{inLo, inHi, outLo, outHi} {
		if p != nil {
			*p = values[i]
		}
	}
	setCode(code, errorCode(err))
}
func CompressBytes(source []byte, blockSize int) ([]byte, error) {
	var b bytes.Buffer
	w, err := NewWriterLevel(&b, blockSize)
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
func DecompressBytes(source []byte) ([]byte, error) {
	var code int32
	r := BZ2_bzReadOpen(&code, bytes.NewReader(source), 0, 0, nil, 0)
	return io.ReadAll(r)
}
func BZ2_bzBuffToBuffCompress(dest []byte, destLen *uint32, source []byte, sourceLen uint32, blockSize, verbosity, workFactor int32) int32 {
	if destLen == nil || uint64(*destLen) > uint64(len(dest)) || uint64(sourceLen) > uint64(len(source)) || blockSize < 1 || blockSize > 9 || verbosity < 0 || verbosity > 4 || workFactor < 0 || workFactor > 250 {
		return BZ_PARAM_ERROR
	}
	sink := &bufferWriter{data: dest[:*destLen]}
	writer, err := NewWriterLevel(sink, int(blockSize))
	if err != nil {
		return errorCode(err)
	}
	if _, err = writer.Write(source[:sourceLen]); err != nil {
		return errorCode(err)
	}
	if err = writer.Close(); err != nil {
		return errorCode(err)
	}
	*destLen = uint32(sink.written)
	return BZ_OK
}
func BZ2_bzBuffToBuffDecompress(dest []byte, destLen *uint32, source []byte, sourceLen uint32, small, verbosity int32) int32 {
	if destLen == nil || uint64(*destLen) > uint64(len(dest)) || uint64(sourceLen) > uint64(len(source)) || !validOptions(verbosity, small) {
		return BZ_PARAM_ERROR
	}
	var code int32
	f := BZ2_bzReadOpen(&code, bytes.NewReader(source[:sourceLen]), verbosity, small, nil, 0)
	n, err := io.ReadFull(f, dest[:*destLen])
	if err == io.ErrUnexpectedEOF && f.lastCode == BZ_STREAM_END {
		err = io.EOF
	}
	if err == io.EOF {
		*destLen = uint32(n)
		return BZ_OK
	}
	if err != nil {
		return errorCode(err)
	}
	var extra [1]byte
	count, err := f.Read(extra[:])
	if count > 0 {
		return BZ_OUTBUFF_FULL
	}
	if err == io.EOF {
		*destLen = uint32(n)
		return BZ_OK
	}
	return errorCode(err)
}

func parseMode(mode string) (writing bool, blockSize int32, small int32, err error) {
	blockSize = 9
	var access bool
	for _, c := range mode {
		switch {
		case c == 'r':
			if access {
				return false, 0, 0, ErrParameters
			}
			access = true
		case c == 'w':
			if access {
				return false, 0, 0, ErrParameters
			}
			access = true
			writing = true
		case c >= '0' && c <= '9':
			blockSize = int32(c - '0')
		case c == 's':
			small = 1
		case c == 'b':
		default:
			return false, 0, 0, ErrParameters
		}
	}
	if !access {
		return false, 0, 0, ErrParameters
	}
	if blockSize < 1 {
		blockSize = 1
	}
	return
}
func Open(path, mode string) (*File, error) {
	writing, size, small, err := parseMode(mode)
	if err != nil {
		return nil, err
	}
	var file *os.File
	own := path != ""
	if !own {
		if writing {
			file = os.Stdout
		} else {
			file = os.Stdin
		}
	} else if writing {
		file, err = os.Create(path)
	} else {
		file, err = os.Open(path)
	}
	if err != nil {
		return nil, err
	}
	var code int32
	var f *File
	if writing {
		f = BZ2_bzWriteOpen(&code, file, size, 0, 0)
	} else {
		f = BZ2_bzReadOpen(&code, file, 0, small, nil, 0)
	}
	if f == nil {
		if own {
			file.Close()
		}
		return nil, fmt.Errorf("bzlib: open status %d", code)
	}
	if own {
		f.owner = file
	}
	return f, nil
}
func BZ2_bzopen(path, mode string) *File { f, _ := Open(path, mode); return f }
func BZ2_bzdopen(fd int, mode string) *File {
	if fd < 0 {
		return nil
	}
	writing, size, small, err := parseMode(mode)
	if err != nil {
		return nil
	}
	file := os.NewFile(uintptr(fd), fmt.Sprintf("fd:%d", fd))
	if file == nil {
		return nil
	}
	var code int32
	var f *File
	if writing {
		f = BZ2_bzWriteOpen(&code, file, size, 0, 0)
	} else {
		f = BZ2_bzReadOpen(&code, file, 0, small, nil, 0)
	}
	if f == nil {
		return nil
	}
	f.owner = file
	return f
}
func BZ2_bzread(f *File, buffer []byte, length int32) int32 {
	var code int32
	n := BZ2_bzRead(&code, f, buffer, length)
	if code == BZ_OK || code == BZ_STREAM_END {
		return n
	}
	return -1
}
func BZ2_bzwrite(f *File, buffer []byte, length int32) int32 {
	var code int32
	BZ2_bzWrite(&code, f, buffer, length)
	if code != BZ_OK {
		return -1
	}
	return length
}
func BZ2_bzflush(f *File) int32 {
	if err := f.Flush(); err != nil {
		return -1
	}
	return 0
}
func BZ2_bzclose(f *File) {
	if f != nil {
		f.Close()
	}
}
func BZ2_bzerror(f *File, code *int32) string {
	if f == nil {
		setCode(code, BZ_PARAM_ERROR)
		return ErrParameters.Error()
	}
	status := f.lastCode
	if status > 0 {
		status = 0
	}
	setCode(code, status)
	if status == 0 {
		return "OK"
	}
	if f.err == nil || errorCode(f.err) != status {
		return statusText(status)
	}
	return f.err.Error()
}

func setFileCode(out *int32, f *File, code int32) {
	setCode(out, code)
	if f != nil {
		f.lastCode = code
	}
}
func statusText(code int32) string {
	switch code {
	case BZ_SEQUENCE_ERROR:
		return "SEQUENCE_ERROR"
	case BZ_PARAM_ERROR:
		return "PARAM_ERROR"
	case BZ_MEM_ERROR:
		return "MEM_ERROR"
	case BZ_DATA_ERROR:
		return "DATA_ERROR"
	case BZ_DATA_ERROR_MAGIC:
		return "DATA_ERROR_MAGIC"
	case BZ_IO_ERROR:
		return "IO_ERROR"
	case BZ_UNEXPECTED_EOF:
		return "UNEXPECTED_EOF"
	case BZ_OUTBUFF_FULL:
		return "OUTBUFF_FULL"
	case BZ_CONFIG_ERROR:
		return "CONFIG_ERROR"
	}
	return "OK"
}

type bufferWriter struct {
	data    []byte
	written int
}

func (w *bufferWriter) Write(p []byte) (int, error) {
	n := copy(w.data[w.written:], p)
	w.written += n
	if n != len(p) {
		return n, ErrOutputFull
	}
	return n, nil
}
