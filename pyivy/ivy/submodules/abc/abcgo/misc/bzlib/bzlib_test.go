package bzlib

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestABCFileAPIAndOwnership(t *testing.T) {
	path := filepath.Join(t.TempDir(), "network.aig.bz2")
	underlying, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	var code int32
	writer := BZ2_bzWriteOpen(&code, underlying, 9, 0, 0)
	if code != BZ_OK || writer == nil {
		t.Fatal(code)
	}
	payload := bytes.Repeat([]byte("aig 3 2 0 1 1\n\x00\x01\xff"), 10000)
	for start := 0; start < len(payload); {
		end := start + 333
		if end > len(payload) {
			end = len(payload)
		}
		BZ2_bzWrite(&code, writer, payload[start:end], int32(end-start))
		if code != BZ_OK {
			t.Fatal(code)
		}
		start = end
	}
	var inLo, inHi, outLo, outHi uint32
	BZ2_bzWriteClose64(&code, writer, 0, &inLo, &inHi, &outLo, &outHi)
	if code != BZ_OK || inLo != uint32(len(payload)) || inHi != 0 || outHi != 0 {
		t.Fatal(code, inLo, inHi, outHi)
	}
	info, err := underlying.Stat()
	if err != nil || uint32(info.Size()) != outLo {
		t.Fatal("borrowed writer closed file or wrong byte count", err, outLo)
	}
	underlying.Close()
	underlying, err = os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	reader := BZ2_bzReadOpen(&code, underlying, 0, 0, nil, 0)
	if code != BZ_OK {
		t.Fatal(code)
	}
	buffer := make([]byte, 1003)
	var decoded []byte
	for {
		n := BZ2_bzRead(&code, reader, buffer, int32(len(buffer)))
		decoded = append(decoded, buffer[:n]...)
		if code == BZ_STREAM_END {
			break
		}
		if code != BZ_OK {
			t.Fatal(code)
		}
	}
	if !bytes.Equal(decoded, payload) {
		t.Fatal("binary payload mismatch")
	}
	var unused []byte
	var count int32
	BZ2_bzReadGetUnused(&code, reader, &unused, &count)
	if code != BZ_OK || count != 0 {
		t.Fatal(code, count)
	}
	BZ2_bzReadClose(&code, reader)
	if code != BZ_OK {
		t.Fatal(code)
	}
	if _, err = underlying.Stat(); err != nil {
		t.Fatal("borrowed reader closed file", err)
	}
	underlying.Close()
	// A simple-open handle owns its descriptor and supports the legacy names.
	owned := BZ2_bzopen(path, "rb")
	if owned == nil {
		t.Fatal("simple open")
	}
	var all []byte
	for {
		n := BZ2_bzread(owned, buffer, int32(len(buffer)))
		if n < 0 {
			t.Fatal("simple read")
		}
		if n == 0 {
			break
		}
		all = append(all, buffer[:n]...)
	}
	BZ2_bzclose(owned)
	if !bytes.Equal(all, payload) {
		t.Fatal("simple read differs")
	}
}

func TestBufferAPIsAndErrors(t *testing.T) {
	source := bytes.Repeat([]byte("input\x00"), 1000)
	compressed := make([]byte, 10000)
	n := uint32(len(compressed))
	if code := BZ2_bzBuffToBuffCompress(compressed, &n, source, uint32(len(source)), 1, 0, 0); code != BZ_OK {
		t.Fatal(code)
	}
	dest := make([]byte, len(source))
	size := uint32(len(dest))
	if code := BZ2_bzBuffToBuffDecompress(dest, &size, compressed, n, 0, 0); code != BZ_OK || !bytes.Equal(dest, source) {
		t.Fatal(code, size)
	}
	size = 5
	if code := BZ2_bzBuffToBuffDecompress(dest, &size, compressed, n, 0, 0); code != BZ_OUTBUFF_FULL {
		t.Fatal("short output", code)
	}
	size = 0
	if code := BZ2_bzBuffToBuffCompress(nil, &size, source, uint32(len(source)), 9, 0, 0); code != BZ_OUTBUFF_FULL {
		t.Fatal("short compressed output", code)
	}
	size = uint32(len(dest))
	if code := BZ2_bzBuffToBuffDecompress(dest, &size, []byte("ordinary plain file"), 19, 0, 0); code != BZ_DATA_ERROR_MAGIC {
		t.Fatal("ABC plain-file fallback requires magic error", code)
	}
	size = uint32(len(dest))
	if code := BZ2_bzBuffToBuffDecompress(dest, &size, compressed, n-5, 0, 0); code != BZ_UNEXPECTED_EOF {
		t.Fatal("truncated stream", code)
	}
	corrupt := append([]byte(nil), compressed[:n]...)
	corrupt[10] ^= 1
	size = uint32(len(dest))
	if code := BZ2_bzBuffToBuffDecompress(dest, &size, corrupt, uint32(len(corrupt)), 0, 0); code != BZ_DATA_ERROR {
		t.Fatal("checksum", code)
	}
	for _, params := range [][3]int32{{0, 0, 0}, {10, 0, 0}, {1, 5, 0}, {1, 0, 251}} {
		size = uint32(len(compressed))
		if code := BZ2_bzBuffToBuffCompress(compressed, &size, source, uint32(len(source)), params[0], params[1], params[2]); code != BZ_PARAM_ERROR {
			t.Fatal(params, code)
		}
	}
}

func TestFlushConcatAndUnusedPrefix(t *testing.T) {
	var compressed bytes.Buffer
	w, _ := NewWriterLevel(&compressed, 1)
	w.Write([]byte("first"))
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("second"))
	w.Close()
	one := append([]byte(nil), compressed.Bytes()...)
	two, err := CompressBytes([]byte("third"), 9)
	if err != nil {
		t.Fatal(err)
	}
	joined := append(one, two...)
	got, err := DecompressBytes(joined)
	if err != nil || string(got) != "firstsecondthird" {
		t.Fatal(err, string(got))
	}
	var code int32
	r := BZ2_bzReadOpen(&code, bytes.NewReader(one[3:]), 0, 1, one[:3], 3)
	got, err = io.ReadAll(r)
	if err != nil || string(got) != "firstsecond" {
		t.Fatal(err, string(got))
	}
	empty, _ := CompressBytes(nil, 9)
	got, err = DecompressBytes(empty)
	if err != nil || len(got) != 0 {
		t.Fatal(err, got)
	}
}

type failingWriter struct{ left int }

func (w *failingWriter) Write(p []byte) (int, error) {
	if w.left <= 0 {
		return 0, io.ErrClosedPipe
	}
	if len(p) > w.left {
		p = p[:w.left]
	}
	n := len(p)
	w.left -= n
	return n, nil
}
func TestWriterFailuresAndAbandon(t *testing.T) {
	sink := &failingWriter{left: 8}
	w, err := NewWriter(sink)
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("some input"))
	if err = w.Close(); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	if _, err = w.Write([]byte("again")); err == nil {
		t.Fatal("write after close")
	}
	var compressed bytes.Buffer
	var code int32
	f := BZ2_bzWriteOpen(&code, &compressed, 9, 0, 0)
	BZ2_bzWrite(&code, f, []byte("discard"), 7)
	BZ2_bzWriteClose(&code, f, 1, nil, nil)
	if code != BZ_OK || compressed.Len() != 4 {
		t.Fatal(code, compressed.Len())
	}
	var result int32
	if BZ2_bzReadOpen(&result, nil, 0, 0, nil, 0) != nil || result != BZ_PARAM_ERROR {
		t.Fatal(result)
	}
	f = BZ2_bzReadOpen(&result, bytes.NewReader([]byte("bad header")), 0, 0, nil, 0)
	BZ2_bzWrite(&result, f, []byte{1}, 1)
	if result != BZ_SEQUENCE_ERROR {
		t.Fatal(result)
	}
}

func TestExternalBzip2Interoperability(t *testing.T) {
	binary, err := exec.LookPath("bzip2")
	if err != nil {
		t.Skip("optional independent bzip2 oracle unavailable")
	}
	source := bytes.Repeat([]byte("ABC bzip2 interoperability\n\x00\xff"), 20000)
	for _, level := range []int{1, 9} {
		compressed, err := CompressBytes(source, level)
		if err != nil {
			t.Fatal(err)
		}
		command := exec.Command(binary, "-dc")
		command.Stdin = bytes.NewReader(compressed)
		decoded, err := command.Output()
		if err != nil || !bytes.Equal(decoded, source) {
			t.Fatal("external decoder rejected Go stream", level, err)
		}
	}
	command := exec.Command(binary, "-c")
	command.Stdin = bytes.NewReader(source)
	compressed, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecompressBytes(compressed)
	if err != nil || !bytes.Equal(decoded, source) {
		t.Fatal(err)
	}
}

func TestUnusedPrefixCopiedAndBoundaryErrors(t *testing.T) {
	compressed, err := CompressBytes([]byte("prefix lifetime"), 1)
	if err != nil {
		t.Fatal(err)
	}
	prefix := append([]byte(nil), compressed[:4]...)
	var code int32
	r := BZ2_bzReadOpen(&code, bytes.NewReader(compressed[4:]), 0, 0, prefix, 4)
	clear(prefix)
	got, err := io.ReadAll(r)
	if err != nil || string(got) != "prefix lifetime" {
		t.Fatal("unused input was not copied", err, string(got))
	}
	BZ2_bzReadClose(&code, nil)
	if code != BZ_OK {
		t.Fatal("nil close")
	}
	if BZ2_bzRead(&code, nil, []byte{0}, 1) != 0 || code != BZ_PARAM_ERROR {
		t.Fatal("nil read", code)
	}
	if BZ2_bzReadOpen(&code, bytes.NewReader(nil), 0, 2, nil, 0) != nil || code != BZ_PARAM_ERROR {
		t.Fatal("small flag", code)
	}
	if _, err = NewWriterLevel(io.Discard, 0); err == nil {
		t.Fatal("invalid blocksize")
	}
	var compressedAgain bytes.Buffer
	w, _ := NewWriter(&compressedAgain)
	w.Write([]byte("x"))
	w.Flush()
	w.Write([]byte("y"))
	w.Flush()
	w.Write([]byte("z"))
	w.Close()
	got, err = DecompressBytes(compressedAgain.Bytes())
	if err != nil || string(got) != "xyz" {
		t.Fatal("unaligned flush blocks", err, string(got))
	}
}

func TestCompatibilityErrorQueryAndRecovery(t *testing.T) {
	compressed, err := CompressBytes([]byte("abc"), 1)
	if err != nil {
		t.Fatal(err)
	}
	var code int32
	f := BZ2_bzReadOpen(&code, bytes.NewReader(compressed), 0, 0, nil, 0)
	BZ2_bzRead(&code, f, make([]byte, 1), 2)
	if code != BZ_PARAM_ERROR {
		t.Fatal(code)
	}
	var queried int32
	if text := BZ2_bzerror(f, &queried); queried != BZ_PARAM_ERROR || text != "PARAM_ERROR" {
		t.Fatal(queried, text)
	}
	buf := make([]byte, 10)
	n := BZ2_bzRead(&code, f, buf, 10)
	if code != BZ_STREAM_END || string(buf[:n]) != "abc" {
		t.Fatal("parameter error poisoned valid read", code, n)
	}
	if text := BZ2_bzerror(f, &queried); queried != BZ_OK || text != "OK" {
		t.Fatal(queried, text)
	}
}

func TestShortPlainInputMagicForABCFallback(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  int32
	}{{"a", BZ_DATA_ERROR_MAGIC}, {"A\n", BZ_DATA_ERROR_MAGIC}, {"BX", BZ_DATA_ERROR_MAGIC}, {"BZx", BZ_DATA_ERROR_MAGIC}, {"BZh0", BZ_DATA_ERROR_MAGIC}, {"B", BZ_UNEXPECTED_EOF}, {"BZ", BZ_UNEXPECTED_EOF}, {"BZh", BZ_UNEXPECTED_EOF}, {"", BZ_UNEXPECTED_EOF}} {
		var code int32
		f := BZ2_bzReadOpen(&code, bytes.NewBufferString(tc.input), 0, 0, nil, 0)
		BZ2_bzRead(&code, f, make([]byte, 100), 100)
		if code != tc.want {
			t.Fatalf("%q got%d want%d", tc.input, code, tc.want)
		}
	}
}
