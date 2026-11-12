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
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompressBuffersAndStandardInteroperability(t *testing.T) {
	source := bytes.Repeat([]byte("abc\x00def\n"), 10000)
	for _, level := range []int{0, 1, 6, 9, -1, -2} {
		output, err := CompressBytes(source, level)
		if err != nil {
			t.Fatal(err)
		}
		r, err := stdzlib.NewReader(bytes.NewReader(output))
		if err != nil {
			t.Fatal(err)
		}
		got, err := io.ReadAll(r)
		r.Close()
		if err != nil || !bytes.Equal(got, source) {
			t.Fatal(level, err)
		}
		dest := make([]byte, CompressBound(uint64(len(source))))
		n := uint64(len(dest))
		if Compress2(dest, &n, source, level) != Z_OK {
			t.Fatal(level)
		}
		plain := make([]byte, len(source))
		size := uint64(len(plain))
		if Uncompress(plain, &size, dest[:n]) != Z_OK || !bytes.Equal(plain, source) {
			t.Fatal(level, size)
		}
		size = 10
		if Uncompress(plain, &size, dest[:n]) != Z_BUF_ERROR || size != 10 {
			t.Fatal("short output", size)
		}
	}
	n := uint64(0)
	if Compress(nil, &n, source) != Z_BUF_ERROR {
		t.Fatal("short compressed buffer")
	}
	if Compress2(nil, &n, nil, 99) != Z_STREAM_ERROR {
		t.Fatal("invalid level")
	}
	n = 5
	if Uncompress(make([]byte, 5), &n, []byte("bad")) != Z_DATA_ERROR {
		t.Fatal("invalid header")
	}
	empty, err := CompressBytes(nil, -1)
	if err != nil {
		t.Fatal(err)
	}
	n = 0
	if Uncompress(nil, &n, empty) != Z_OK {
		t.Fatal("empty roundtrip")
	}
	bad := append([]byte(nil), empty...)
	bad[len(bad)-1] ^= 1
	n = 0
	if Uncompress(nil, &n, bad) != Z_DATA_ERROR {
		t.Fatal("checksum accepted")
	}
}

func TestChecksumsAndCombines(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	data := make([]byte, 200000)
	rng.Read(data)
	for i := 0; i < 100; i++ {
		split := rng.Intn(len(data) + 1)
		a, b := data[:split], data[split:]
		if Adler32(1, data) != adler32.Checksum(data) {
			t.Fatal("adler")
		}
		if Adler32(Adler32(1, a), b) != adler32.Checksum(data) {
			t.Fatal("incremental adler")
		}
		if Adler32Combine(adler32.Checksum(a), adler32.Checksum(b), int64(len(b))) != adler32.Checksum(data) {
			t.Fatal("combined adler")
		}
		if CRC32(crc32.ChecksumIEEE(a), b) != crc32.ChecksumIEEE(data) {
			t.Fatal("incremental crc")
		}
		if CRC32Combine(crc32.ChecksumIEEE(a), crc32.ChecksumIEEE(b), int64(len(b))) != crc32.ChecksumIEEE(data) {
			t.Fatal("combined crc")
		}
	}
	if Adler32(99, nil) != 1 || CRC32(99, nil) != 0 {
		t.Fatal("nil initialization")
	}
}

func TestStreamingFormatsAndDictionary(t *testing.T) {
	dictionary := []byte("common dictionary words and content")
	for _, format := range []Format{Zlib, Gzip, Raw} {
		for _, dict := range [][]byte{nil, dictionary} {
			if format == Gzip && dict != nil {
				continue
			}
			var buffer bytes.Buffer
			writer, err := NewFormatWriter(&buffer, format, 6, dict)
			if err != nil {
				t.Fatal(err)
			}
			chunks := []string{"common dictionary ", "words and content ", strings.Repeat("a", 4096)}
			for _, s := range chunks {
				if _, err = io.WriteString(writer, s); err != nil {
					t.Fatal(err)
				}
				if err = writer.Flush(); err != nil {
					t.Fatal(err)
				}
			}
			if err = writer.Close(); err != nil {
				t.Fatal(err)
			}
			reader, err := NewFormatReader(bytes.NewReader(buffer.Bytes()), format, dict)
			if err != nil {
				t.Fatal(err)
			}
			got, err := io.ReadAll(reader)
			reader.Close()
			if err != nil || string(got) != strings.Join(chunks, "") {
				t.Fatal(format, err)
			}
			if dict == nil && format != Raw {
				auto, err := NewFormatReader(bytes.NewReader(buffer.Bytes()), Auto, nil)
				if err != nil {
					t.Fatal(err)
				}
				got, err = io.ReadAll(auto)
				auto.Close()
				if err != nil || string(got) != strings.Join(chunks, "") {
					t.Fatal(err)
				}
			}
		}
	}
	var encoded bytes.Buffer
	w, _ := NewWriterLevelDict(&encoded, 6, dictionary)
	w.Write(dictionary)
	w.Close()
	if _, err := NewReaderDict(bytes.NewReader(encoded.Bytes()), []byte("wrong")); !errors.Is(err, ErrDictionary) {
		t.Fatal(err)
	}
	// The returned reader supports the standard zlib reset contract.
	r, err := NewReaderDict(bytes.NewReader(encoded.Bytes()), dictionary)
	if err != nil {
		t.Fatal(err)
	}
	if err = r.(Resetter).Reset(bytes.NewReader(encoded.Bytes()), dictionary); err != nil {
		t.Fatal(err)
	}
	r.Close()
	if _, err = NewFormatWriter(io.Discard, Gzip, flate.DefaultCompression, dictionary); err == nil {
		t.Fatal("gzip dictionary accepted")
	}
}

func TestGzipFileConsumerAPI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "network.aig")
	f := Gzopen(path, "wb6")
	if f == nil {
		t.Fatal("open")
	}
	if Gzbuffer(f, 1024) != 0 {
		t.Fatal("buffer")
	}
	if Gzprintf(f, "p cnf %u %ld\n", uint32(17), int64(2)) != 11 {
		t.Fatal("printf")
	}
	if Gzputs(f, "1 -2 0\n\x00ignored") != 7 {
		t.Fatal("puts")
	}
	if Gzwrite(f, []byte{0, 1, 2}, 3) != 3 {
		t.Fatal("write")
	}
	if Gzclose(f) != Z_OK {
		t.Fatal("close")
	}
	compressed, err := os.ReadFile(path)
	if err != nil || !bytes.HasPrefix(compressed, []byte{31, 139}) {
		t.Fatal(err)
	}
	plainReader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatal(err)
	}
	plain, err := io.ReadAll(plainReader)
	plainReader.Close()
	if err != nil || string(plain) != "p cnf 17 2\n1 -2 0\n\x00\x01\x02" {
		t.Fatal(err, plain)
	}
	f = Gzopen(path, "rb")
	if Gzdirect(f) != 0 {
		t.Fatal("gzip detected as plain")
	}
	buffer := make([]byte, 100)
	got := Gzread(f, buffer, 100)
	if got != int32(len(plain)) || !bytes.Equal(buffer[:got], plain) || Gzeof(f) != 1 {
		t.Fatal(got, Gzeof(f))
	}
	if Gzrewind(f) != 0 || Gztell(f) != 0 {
		t.Fatal("rewind")
	}
	if Gzgetc(f) != 'p' || Gzungetc('X', f) != 'X' || Gzgetc(f) != 'X' {
		t.Fatal("pushback")
	}
	if Gzseek(f, 6, io.SeekStart) != 6 || Gzgetc(f) != '1' {
		t.Fatal("forward seek")
	}
	if Gzseek(f, -1, io.SeekCurrent) != 6 || Gzgetc(f) != '1' {
		t.Fatal("backward seek")
	}
	if Gzrewind(f) != 0 {
		t.Fatal("rewind")
	}
	line := make([]byte, 50)
	gotLine := Gzgets(f, line, 50)
	if string(gotLine) != "p cnf 17 2\n\x00" {
		t.Fatal(gotLine)
	}
	if Gzseek(f, 0, io.SeekEnd) != -1 {
		t.Fatal("seek-end accepted")
	}
	if GzcloseR(f) != Z_OK {
		t.Fatal("read close")
	}
}

func TestPlainInputAppendAndFinishedMembers(t *testing.T) {
	root := t.TempDir()
	plainPath := filepath.Join(root, "plain.gz")
	os.WriteFile(plainPath, []byte("plain input"), 0600)
	f := Gzopen(plainPath, "rb")
	if Gzdirect(f) != 1 {
		t.Fatal("plain detection")
	}
	got, err := io.ReadAll(f)
	if err != nil || string(got) != "plain input" {
		t.Fatal(err, string(got))
	}
	f.Close()
	path := filepath.Join(root, "members.gz")
	f = Gzopen(path, "wb")
	io.WriteString(f, "first")
	if Gzflush(f, Z_FINISH) != Z_OK {
		t.Fatal("finish")
	}
	io.WriteString(f, "second")
	f.Close()
	f = Gzopen(path, "ab")
	io.WriteString(f, "third")
	f.Close()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	file.WriteString("trailing garbage")
	file.Close()
	f = Gzopen(path, "rb")
	got, err = io.ReadAll(f)
	if err != nil || string(got) != "firstsecondthird" {
		t.Fatal(err, string(got))
	}
	f.Close()
	directPath := filepath.Join(root, "direct")
	f = Gzopen(directPath, "wbT")
	if Gzseek(f, 3, io.SeekStart) != 3 {
		t.Fatal("write seek")
	}
	Gzputc(f, 'x')
	f.Close()
	got, _ = os.ReadFile(directPath)
	if string(got) != "\x00\x00\x00x" {
		t.Fatal(got)
	}
}

func TestGzipCorruptionAndValidation(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "bad.gz")
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write([]byte("content"))
	w.Close()
	data := append([]byte(nil), b.Bytes()...)
	data[len(data)-1] ^= 1
	os.WriteFile(path, data, 0600)
	f := Gzopen(path, "rb")
	_, err := io.ReadAll(f)
	if err == nil {
		t.Fatal("checksum accepted")
	}
	var code int32
	if Gzerror(f, &code) == "" || code == Z_OK {
		t.Fatal(code)
	}
	Gzclearerr(f)
	if Gzerror(f, &code) != "" || code != Z_OK {
		t.Fatal(code)
	}
	f.Close()
	if _, err = Open(path, "r+"); err == nil {
		t.Fatal("invalid mode")
	}
	if Gzopen(filepath.Join(root, "missing"), "rb") != nil {
		t.Fatal("missing opened")
	}
	if Gzclose(nil) != Z_STREAM_ERROR || Gzseek(nil, 0, 0) != -1 || Gzread(nil, make([]byte, 1), 1) != -1 {
		t.Fatal("invalid handles")
	}
}

func TestBugReviewZeroLengthNilHandlesAndMemberFinish(t *testing.T) {
	if Gzread(nil, nil, 0) != -1 || Gzgets(nil, make([]byte, 1), 1) != nil {
		t.Fatal("invalid zero-length handles")
	}
	path := filepath.Join(t.TempDir(), "single.gz")
	f := Gzopen(path, "wb")
	Gzputs(f, "one")
	if Gzflush(f, Z_SYNC_FLUSH) != Z_OK || Gzflush(f, Z_FINISH) != Z_OK {
		t.Fatal("flush")
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if Gzclose(f) != Z_OK {
		t.Fatal("close")
	}
	after, _ := os.Stat(path)
	if before.Size() != after.Size() {
		t.Fatal("close added an unwanted empty member")
	}
	f = Gzopen(path, "rb")
	if Gzflush(f, Z_SYNC_FLUSH) != Z_STREAM_ERROR {
		t.Fatal("flushing read handle")
	}
	io.ReadAll(f)
	if Gzrewind(f) != 0 || Gzgetc(f) != 'o' {
		t.Fatal("rewind after eof")
	}
	f.Close()
	f = Gzopen(path, "wb")
	if Gzflush(f, Z_FULL_FLUSH) != Z_STREAM_ERROR {
		t.Fatal("unsupported dictionary-reset flush pretended supported")
	}
	f.Close()
}

func TestFileDescriptorOwnershipAndNonseekableRead(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	go func() { writer.Write([]byte("pipe data")); writer.Close() }()
	f, err := OpenFile(reader, "rb")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(f)
	if err != nil || string(data) != "pipe data" {
		t.Fatal(err, string(data))
	}
	if Gzrewind(f) != -1 {
		t.Fatal("pipe rewind should fail")
	}
	Gzclearerr(f)
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	var b [1]byte
	if _, err = reader.Read(b[:]); err == nil {
		t.Fatal("OpenFile did not close owned descriptor")
	}
}

func TestIncrementalPipeStreaming(t *testing.T) {
	for _, format := range []Format{Zlib, Gzip, Raw} {
		t.Run(fmt.Sprint(format), func(t *testing.T) {
			input, output := io.Pipe()
			ack := make(chan struct{})
			done := make(chan error, 1)
			go func() {
				w, err := NewFormatWriter(output, format, 6, nil)
				if err != nil {
					output.CloseWithError(err)
					done <- err
					return
				}
				_, err = io.WriteString(w, "first")
				if err == nil {
					err = w.Flush()
				}
				if err != nil {
					output.CloseWithError(err)
					done <- err
					return
				}
				<-ack
				_, err = io.WriteString(w, "second")
				if err == nil {
					err = w.Close()
				}
				output.CloseWithError(err)
				done <- err
			}()
			r, err := NewFormatReader(input, format, nil)
			if err != nil {
				close(ack)
				t.Fatal(err)
			}
			first := make([]byte, 5)
			if _, err = io.ReadFull(r, first); err != nil || string(first) != "first" {
				close(ack)
				t.Fatal(err, string(first))
			}
			close(ack)
			rest, err := io.ReadAll(r)
			r.Close()
			input.Close()
			if err != nil || string(rest) != "second" {
				t.Fatal(err, string(rest))
			}
			if err = <-done; err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestHuffmanStrategyIsNotCompressionLevel(t *testing.T) {
	if Z_HUFFMAN_ONLY != 2 || HuffmanOnly != flate.HuffmanOnly {
		t.Fatal("strategy/level conflated")
	}
	path := filepath.Join(t.TempDir(), "huffman.gz")
	f := Gzopen(path, "wbh")
	if f == nil {
		t.Fatal("huffman open")
	}
	Gzputs(f, "abcabcabc")
	if Gzsetparams(f, 6, Z_HUFFMAN_ONLY) != Z_OK {
		t.Fatal("huffman strategy")
	}
	Gzputs(f, "def")
	if Gzclose(f) != Z_OK {
		t.Fatal("close")
	}
	f = Gzopen(path, "rb")
	data, err := io.ReadAll(f)
	f.Close()
	if err != nil || string(data) != "abcabcabcdef" {
		t.Fatal(err, string(data))
	}
}

func TestPrintfUnsignedLengths(t *testing.T) {
	path := filepath.Join(t.TempDir(), "format.gz")
	f := Gzopen(path, "wb")
	Gzprintf(f, "%u %hhu %hu %llu %*u %%", int32(-1), int32(-1), int32(-1), int64(-1), 3, int32(7))
	if Gzclose(f) != Z_OK {
		t.Fatal("write close")
	}
	f = Gzopen(path, "rb")
	text, err := io.ReadAll(f)
	f.Close()
	if err != nil || string(text) != "4294967295 255 65535 18446744073709551615   7 %" {
		t.Fatal(err, string(text))
	}
}

func TestOwnedFileInitialOffsetAndRewind(t *testing.T) {
	path := filepath.Join(t.TempDir(), "embedded.gz")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.WriteString(file, "prefix"); err != nil {
		t.Fatal(err)
	}
	w := gzip.NewWriter(file)
	w.Write([]byte("payload"))
	w.Close()
	file.Close()
	file, err = os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	file.Seek(6, io.SeekStart)
	f, err := OpenFile(file, "rb")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(f)
	if err != nil || string(data) != "payload" {
		t.Fatal(err, string(data))
	}
	if Gzrewind(f) != 0 || Gzgetc(f) != 'p' || Gztell(f) != 1 {
		t.Fatal("rewind lost descriptor initial offset")
	}
	if Gzoffset(f) < 6 {
		t.Fatal("physical offset lost prefix")
	}
	f.Close()
}
