package abc

// Integration tests for gzip file lifecycle from misc/zlib/gz*.c.

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

type gzipFileOracle struct {
	Buffer int   `json:"buffer"`
	Write  []int `json:"write"`
	Read   struct {
		Direct, Tell0, Tell1, Getc, Ungetc, Tell2, Pushed, Tell3 int
		Seek, EOFGetc, EOF, Cleared, Rewind, TellRewind          int
		SeekCurrent, Close                                       int
		Line1, Line2, Line3                                      string
	} `json:"read"`
	StateHex, AppendHex, AppendRead, RawRead string
	RawDirect                                int `json:"rawDirect"`
}

type gzipDescriptorOracle struct {
	Read    []int  `json:"read"`
	Write   []int  `json:"write"`
	Payload string `json:"payload"`
}

type gzipPushbackOracle struct {
	Setup    []int  `json:"setup"`
	Push     []int  `json:"push"`
	Read     []int  `json:"read"`
	Data     string `json:"data"`
	EOFAfter int    `json:"eofAfter"`
}

type gzipTrailingRawRecord struct {
	Count, Direct, Tell, EOF int
	Hex                      string
}

type gzipTrailingRawOracle struct {
	First, Second, Third gzipTrailingRawRecord
	End                  struct {
		Count, EOF, Close int
	}
}

func TestGZIPFileNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/zlib_gzip_file.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle gzipFileOracle
	if err = json.Unmarshal(data, &oracle); err != nil || oracle.Buffer != 0 || len(oracle.Write) != 7 {
		t.Fatal(err, oracle.Buffer, len(oracle.Write))
	}
	directory := t.TempDir()
	stateBytes, err := hex.DecodeString(oracle.StateHex)
	if err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(directory, "native-state.gz")
	if err = os.WriteFile(statePath, stateBytes, 0600); err != nil {
		t.Fatal(err)
	}
	reader, err := OpenGZIPReader(statePath, GZIPReaderOptions{})
	if err != nil || boolInt(reader.Direct()) != oracle.Read.Direct || reader.Tell() != int64(oracle.Read.Tell0) {
		t.Fatal(err, reader.Direct(), reader.Tell())
	}
	line, err := reader.ReadLine(32)
	if err != nil || hex.EncodeToString(line) != oracle.Read.Line1 || reader.Tell() != int64(oracle.Read.Tell1) {
		t.Fatal(err, string(line), reader.Tell())
	}
	value, err := reader.ReadByte()
	if err != nil || int(value) != oracle.Read.Getc {
		t.Fatal(value, err)
	}
	if err = reader.PushByte(byte(oracle.Read.Ungetc)); err != nil || reader.Tell() != int64(oracle.Read.Tell2) {
		t.Fatal(err, reader.Tell())
	}
	value, err = reader.ReadByte()
	if err != nil || int(value) != oracle.Read.Pushed || reader.Tell() != int64(oracle.Read.Tell3) {
		t.Fatal(value, reader.Tell(), err)
	}
	if position, err := reader.Seek(10, io.SeekStart); err != nil || position != int64(oracle.Read.Seek) {
		t.Fatal(position, err)
	}
	line, err = reader.ReadLine(5)
	if err != nil || hex.EncodeToString(line) != oracle.Read.Line2 {
		t.Fatal(string(line), err)
	}
	line, err = reader.ReadLine(5)
	if err != nil || hex.EncodeToString(line) != oracle.Read.Line3 {
		t.Fatal(string(line), err)
	}
	if _, err = reader.ReadByte(); err != io.EOF || boolInt(reader.EOF()) != oracle.Read.EOF {
		t.Fatal(err, reader.EOF())
	}
	reader.ClearError()
	if boolInt(reader.EOF()) != oracle.Read.Cleared {
		t.Fatal(reader.EOF())
	}
	if err = reader.Rewind(); err != nil || reader.Tell() != int64(oracle.Read.TellRewind) {
		t.Fatal(err, reader.Tell())
	}
	if position, err := reader.Seek(6, io.SeekCurrent); err != nil || position != int64(oracle.Read.SeekCurrent) {
		t.Fatal(position, err)
	}
	if err = reader.Close(); err != nil {
		t.Fatal(err)
	}
	appendBytes, err := hex.DecodeString(oracle.AppendHex)
	if err != nil {
		t.Fatal(err)
	}
	appendPath := filepath.Join(directory, "native-append.gz")
	if err = os.WriteFile(appendPath, appendBytes, 0600); err != nil {
		t.Fatal(err)
	}
	reader, err = OpenGZIPReader(appendPath, GZIPReaderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	all, err := io.ReadAll(reader)
	if err != nil || hex.EncodeToString(all) != oracle.AppendRead {
		t.Fatal(string(all), err)
	}
	reader.Close()
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func TestGZIPFileRoundTripState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.gz")
	writer, err := OpenGZIPWriter(path, GZIPWriterOptions{Level: 1, BufferSize: 17})
	if err != nil {
		t.Fatal(err)
	}
	if err = writer.SetLevel(9); err != nil {
		t.Fatal(err)
	}
	if count, err := writer.WriteString("first\n"); count != 6 || err != nil {
		t.Fatal(count, err)
	}
	if err = writer.WriteByte('X'); err != nil {
		t.Fatal(err)
	}
	if position, err := writer.Seek(10, io.SeekStart); position != 10 || err != nil {
		t.Fatal(position, err)
	}
	if _, err = writer.WriteString("last\n"); err != nil || writer.Tell() != 15 {
		t.Fatal(writer.Tell(), err)
	}
	if err = writer.Flush(); err != nil {
		t.Fatal(err)
	}
	if err = writer.SetLevel(1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenGZIPReader(path, GZIPReaderOptions{BufferSize: 19})
	if err != nil || reader.Direct() {
		t.Fatal(err, reader.Direct())
	}
	line, err := reader.ReadLine(32)
	if err != nil || string(line) != "first\n" || reader.Tell() != 6 {
		t.Fatal(string(line), reader.Tell(), err)
	}
	value, err := reader.ReadByte()
	if err != nil || value != 'X' {
		t.Fatal(value, err)
	}
	if err = reader.PushByte('Y'); err != nil {
		t.Fatal(err)
	}
	value, err = reader.ReadByte()
	if err != nil || value != 'Y' || reader.Tell() != 7 {
		t.Fatal(value, reader.Tell(), err)
	}
	if position, err := reader.Seek(10, io.SeekStart); position != 10 || err != nil {
		t.Fatal(position, err)
	}
	line, err = reader.ReadLine(5)
	if err != nil || string(line) != "last" {
		t.Fatal(string(line), err)
	}
	line, err = reader.ReadLine(5)
	if err != nil || string(line) != "\n" {
		t.Fatal(string(line), err)
	}
	if _, err = reader.ReadByte(); err != io.EOF || !reader.EOF() {
		t.Fatal(err, reader.EOF())
	}
	reader.ClearError()
	if reader.EOF() || reader.LastError() != nil {
		t.Fatal(reader.EOF(), reader.LastError())
	}
	if err = reader.Rewind(); err != nil || reader.Tell() != 0 {
		t.Fatal(reader.Tell(), err)
	}
	if position, err := reader.Seek(6, io.SeekCurrent); position != 6 || err != nil {
		t.Fatal(position, err)
	}
	if err = reader.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGZIPFileAppendTransparentAndErrors(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "append.gz")
	for index, text := range []string{"one", "two"} {
		writer, err := OpenGZIPWriter(path, GZIPWriterOptions{Level: 6, Append: index != 0})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = writer.WriteString(text); err != nil {
			t.Fatal(err)
		}
		if err = writer.Close(); err != nil {
			t.Fatal(err)
		}
	}
	reader, err := OpenGZIPReader(path, GZIPReaderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	all, err := io.ReadAll(reader)
	if err != nil || string(all) != "onetwo" {
		t.Fatal(string(all), err)
	}
	reader.Close()

	rawPath := filepath.Join(directory, "plain")
	if err = os.WriteFile(rawPath, []byte("plain\ntext"), 0600); err != nil {
		t.Fatal(err)
	}
	raw, err := OpenGZIPReader(rawPath, GZIPReaderOptions{})
	if err != nil || !raw.Direct() {
		t.Fatal(err, raw.Direct())
	}
	all, err = io.ReadAll(raw)
	if err != nil || string(all) != "plain\ntext" {
		t.Fatal(string(all), err)
	}
	raw.Close()

	limitedPath := filepath.Join(directory, "limited.gz")
	limited, err := OpenGZIPWriter(limitedPath, GZIPWriterOptions{Level: 6, MaxInput: 3})
	if err != nil {
		t.Fatal(err)
	}
	if count, err := limited.Write([]byte("four")); count != 0 || !errors.Is(err, ErrLimit) {
		t.Fatal(count, err)
	}
	if err = limited.Close(); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	for _, options := range []GZIPWriterOptions{{Level: -2}, {Level: 10}, {BufferSize: -1}, {MaxInput: -1}} {
		if value, err := OpenGZIPWriter(filepath.Join(directory, "bad"), options); value != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(value, err)
		}
	}
	cancelled, err := OpenGZIPWriter(filepath.Join(directory, "cancel"), GZIPWriterOptions{Level: 6, Stop: func() bool { return true }})
	if cancelled != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(cancelled, err)
	}
}

func TestGZIPReaderSourceOwnership(t *testing.T) {
	path := filepath.Join(t.TempDir(), "owned.gz")
	writer, _ := OpenGZIPWriter(path, GZIPWriterOptions{Level: 6})
	source := []byte("owned")
	writer.Write(source)
	source[0] = 'X'
	writer.Close()
	reader, err := OpenGZIPReader(path, GZIPReaderOptions{MaxOutput: 5})
	if err != nil {
		t.Fatal(err)
	}
	output := make([]byte, 5)
	if _, err = io.ReadFull(reader, output); err != nil || !slices.Equal(output, []byte("owned")) {
		t.Fatal(string(output), err)
	}
	reader.Close()
}

func TestGZIPDescriptorOffsetsAndDeferredSeek(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/zlib_gzip_descriptor.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle gzipDescriptorOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Read) != 12 || len(oracle.Write) != 6 {
		t.Fatal(err, len(oracle.Read), len(oracle.Write))
	}
	directory := t.TempDir()
	memberPath := filepath.Join(directory, "member.gz")
	writer, err := OpenGZIPWriter(memberPath, GZIPWriterOptions{Level: 6})
	if err != nil {
		t.Fatal(err)
	}
	writer.WriteString("descriptor")
	writer.Close()
	member, err := os.ReadFile(memberPath)
	if err != nil {
		t.Fatal(err)
	}
	containerPath := filepath.Join(directory, "container")
	container := append([]byte("pre"), member...)
	if err = os.WriteFile(containerPath, container, 0600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(containerPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = file.Seek(3, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	reader, err := OpenGZIPReaderFile(file, GZIPReaderOptions{})
	if err != nil || boolInt(reader.Direct()) != oracle.Read[0] {
		t.Fatal(err, reader.Direct())
	}
	if offset, err := reader.Offset(); err != nil || offset != int64(oracle.Read[1]) {
		t.Fatal(offset, err)
	}
	if position, err := reader.Seek(int64(oracle.Read[2]), io.SeekStart); err != nil || position != int64(oracle.Read[2]) || reader.Tell() != int64(oracle.Read[3]) {
		t.Fatal(position, reader.Tell(), err)
	}
	value, err := reader.ReadByte()
	if int(value) != 0 || err != io.EOF || oracle.Read[4] != -1 || boolInt(reader.EOF()) != oracle.Read[5] || reader.Tell() != int64(oracle.Read[6]) {
		t.Fatal(err, reader.EOF(), reader.Tell())
	}
	if err = reader.Rewind(); boolInt(err == nil) != boolInt(oracle.Read[7] == 0) || reader.Tell() != int64(oracle.Read[8]) {
		t.Fatal(err)
	}
	output, err := io.ReadAll(reader)
	if err != nil || string(output) != "descriptor" {
		t.Fatal(string(output), err)
	}
	if err = reader.Close(); boolInt(err == nil) != boolInt(oracle.Read[9] == 0) {
		t.Fatal(err)
	}
	_, err = file.Read(make([]byte, 1))
	if boolInt(err != nil) != oracle.Read[10] {
		t.Fatal("descriptor not closed")
	}

	writePath := filepath.Join(directory, "descriptor-write")
	writeFile, err := os.OpenFile(writePath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = writeFile.WriteString("pre"); err != nil {
		t.Fatal(err)
	}
	descriptorWriter, err := OpenGZIPWriterFile(writeFile, GZIPWriterOptions{Level: 6})
	if err != nil {
		t.Fatal(err)
	}
	if offset, err := descriptorWriter.Offset(); err != nil || offset != int64(oracle.Write[0]) {
		t.Fatal(offset, err)
	}
	if count, err := descriptorWriter.WriteString(oracle.Payload); err != nil || count != oracle.Write[1] {
		t.Fatal(count, err)
	}
	if err = descriptorWriter.Flush(); boolInt(err == nil) != boolInt(oracle.Write[2] == 0) {
		t.Fatal(err)
	}
	offset, err := descriptorWriter.Offset()
	descriptorOffset, seekErr := writeFile.Seek(0, io.SeekCurrent)
	if err != nil || seekErr != nil || offset != descriptorOffset || offset <= int64(oracle.Write[0]) || oracle.Write[3] <= oracle.Write[0] {
		t.Fatal(offset, descriptorOffset, err, seekErr)
	}
	if err = descriptorWriter.Close(); boolInt(err == nil) != boolInt(oracle.Write[4] == 0) {
		t.Fatal(err)
	}
	_, err = writeFile.WriteString("x")
	if boolInt(err != nil) != oracle.Write[5] {
		t.Fatal("descriptor not closed", err)
	}
	readFile, err := os.Open(writePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = readFile.Seek(3, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	reader, err = OpenGZIPReaderFile(readFile, GZIPReaderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	output, err = io.ReadAll(reader)
	if err != nil || len(output) != oracle.Read[11] || string(output) != oracle.Payload {
		t.Fatal(string(output), err)
	}
	reader.Close()
}

func TestGZIPExtendedPushbackNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/zlib_gzip_pushback.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle gzipPushbackOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Setup) != 3 || len(oracle.Push) != 11 || len(oracle.Read) != 9 {
		t.Fatal(err, len(oracle.Setup), len(oracle.Push), len(oracle.Read))
	}
	path := filepath.Join(t.TempDir(), "pushback.gz")
	writer, err := OpenGZIPWriter(path, GZIPWriterOptions{Level: 6})
	if err != nil {
		t.Fatal(err)
	}
	if count, err := writer.WriteString("abcdef"); err != nil || count != oracle.Setup[1] {
		t.Fatal(count, err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	reader, err := OpenGZIPReader(path, GZIPReaderOptions{BufferSize: 4})
	if err != nil {
		t.Fatal(err)
	}
	initial, err := io.ReadAll(reader)
	if err != nil || len(initial) != oracle.Setup[1] || reader.Tell() != int64(oracle.Setup[2]) {
		t.Fatal(string(initial), reader.Tell(), err)
	}
	for index, result := range oracle.Push[:8] {
		if err = reader.PushByte(byte(result)); err != nil {
			t.Fatal(index, err)
		}
	}
	if err = reader.PushByte('I'); !errors.Is(err, ErrLimit) || oracle.Push[8] != -1 || reader.Tell() != int64(oracle.Push[9]) {
		t.Fatal(err, reader.Tell())
	}
	buffer := make([]byte, 10)
	if count, err := reader.Read(buffer); count != 0 || !errors.Is(err, ErrLimit) || oracle.Push[10] != -1 {
		t.Fatal(count, err)
	}
	reader.ClearError()
	count, err := reader.Read(buffer[:8])
	if err != nil || count != oracle.Read[0] || reader.Tell() != int64(oracle.Read[1]) {
		t.Fatal(count, reader.Tell(), err)
	}
	if _, err = reader.ReadByte(); err != io.EOF || oracle.Read[2] != -1 || boolInt(reader.EOF()) != oracle.Read[3] {
		t.Fatal(err, reader.EOF())
	}
	if err = reader.PushByte(byte(oracle.Read[4])); err != nil {
		t.Fatal(err)
	}
	if err = reader.PushByte(byte(oracle.Read[5])); err != nil || boolInt(reader.EOF()) != oracle.Read[6] || reader.Tell() != int64(oracle.Read[7]) {
		t.Fatal(err, reader.EOF(), reader.Tell())
	}
	count, err = reader.Read(buffer[8:])
	if err != nil || count != oracle.Read[8] || boolInt(reader.EOF()) != oracle.EOFAfter || hex.EncodeToString(buffer) != oracle.Data {
		t.Fatal(count, err, reader.EOF(), hex.EncodeToString(buffer))
	}
	if err = reader.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGZIPTrailingRawNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/zlib_gzip_trailing_raw.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle gzipTrailingRawOracle
	if err = json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "trailing-raw.gz")
	for index, value := range []string{"one", "two"} {
		writer, err := OpenGZIPWriter(path, GZIPWriterOptions{Level: 6, Append: index != 0})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = writer.WriteString(value); err != nil {
			t.Fatal(err)
		}
		if err = writer.Close(); err != nil {
			t.Fatal(err)
		}
	}
	tail, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tail.WriteString("RAW"); err != nil {
		t.Fatal(err)
	}
	if err = tail.Close(); err != nil {
		t.Fatal(err)
	}
	reader, err := OpenGZIPReader(path, GZIPReaderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	check := func(record gzipTrailingRawRecord, size int) {
		t.Helper()
		buffer := make([]byte, size)
		count, readErr := reader.Read(buffer)
		if readErr != nil || count != record.Count || boolInt(reader.Direct()) != record.Direct || reader.Tell() != int64(record.Tell) || boolInt(reader.EOF()) != record.EOF || hex.EncodeToString(buffer[:count]) != record.Hex {
			t.Fatal(count, readErr, reader.Direct(), reader.Tell(), reader.EOF(), hex.EncodeToString(buffer[:count]))
		}
	}
	check(oracle.First, 3)
	check(oracle.Second, 2)
	check(oracle.Third, 16)
	buffer := make([]byte, 1)
	count, err := reader.Read(buffer)
	if count != oracle.End.Count || err != io.EOF || boolInt(reader.EOF()) != oracle.End.EOF {
		t.Fatal(count, err, reader.EOF())
	}
	if err = reader.Close(); boolInt(err == nil) != boolInt(oracle.End.Close == 0) {
		t.Fatal(err)
	}
}

func FuzzGZIPFileRoundTrip(f *testing.F) {
	f.Add([]byte("abcdef"), uint8(6), uint16(3), uint8(2))
	f.Add([]byte{}, uint8(0), uint16(0), uint8(0))
	f.Fuzz(func(t *testing.T, source []byte, rawLevel uint8, rawSplit uint16, rawGap uint8) {
		if len(source) > 2048 {
			source = source[:2048]
		}
		level := int(rawLevel % 10)
		split := 0
		if len(source) != 0 {
			split = int(rawSplit) % (len(source) + 1)
		}
		gap := int(rawGap % 17)
		path := filepath.Join(t.TempDir(), "fuzz.gz")
		writer, err := OpenGZIPWriter(path, GZIPWriterOptions{Level: level, MaxInput: int64(len(source) + gap + 1)})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = writer.Write(source[:split]); err != nil {
			t.Fatal(err)
		}
		if _, err = writer.Seek(int64(gap), io.SeekCurrent); err != nil {
			t.Fatal(err)
		}
		if _, err = writer.Write(source[split:]); err != nil {
			t.Fatal(err)
		}
		if err = writer.Close(); err != nil {
			t.Fatal(err)
		}
		reader, err := OpenGZIPReader(path, GZIPReaderOptions{MaxOutput: int64(len(source) + gap + 1)})
		if err != nil {
			t.Fatal(err)
		}
		output, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		reader.Close()
		want := make([]byte, 0, len(source)+gap)
		want = append(want, source[:split]...)
		want = append(want, make([]byte, gap)...)
		want = append(want, source[split:]...)
		if !slices.Equal(output, want) {
			t.Fatal(len(source), split, gap, len(output))
		}
	})
}

func FuzzGZIPPushback(f *testing.F) {
	f.Add([]byte("abcdef"), []byte("ABCDEFGH"), uint8(4))
	f.Add([]byte{}, []byte{0, 255}, uint8(1))
	f.Fuzz(func(t *testing.T, source, pushed []byte, rawSize uint8) {
		if len(source) > 256 {
			source = source[:256]
		}
		bufferSize := int(rawSize%64) + 1
		if len(pushed) > bufferSize*2 {
			pushed = pushed[:bufferSize*2]
		}
		path := filepath.Join(t.TempDir(), "pushback.gz")
		writer, err := OpenGZIPWriter(path, GZIPWriterOptions{Level: 6})
		if err != nil {
			t.Fatal(err)
		}
		if count, err := writer.Write(source); err != nil || count != len(source) {
			t.Fatal(count, err)
		}
		if err = writer.Close(); err != nil {
			t.Fatal(err)
		}
		reader, err := OpenGZIPReader(path, GZIPReaderOptions{BufferSize: bufferSize, MaxOutput: int64(len(source) + 1)})
		if err != nil {
			t.Fatal(err)
		}
		initial, err := io.ReadAll(reader)
		if err != nil || !slices.Equal(initial, source) {
			t.Fatal(len(initial), err)
		}
		for _, value := range pushed {
			if err = reader.PushByte(value); err != nil {
				t.Fatal(err)
			}
		}
		if reader.Tell() != int64(len(source)-len(pushed)) {
			t.Fatal(reader.Tell(), len(source), len(pushed))
		}
		output := make([]byte, len(pushed))
		if _, err = io.ReadFull(reader, output); err != nil {
			t.Fatal(err)
		}
		want := slices.Clone(pushed)
		slices.Reverse(want)
		if !slices.Equal(output, want) || reader.Tell() != int64(len(source)) || !reader.EOF() {
			t.Fatal(len(source), bufferSize, output, want, reader.Tell(), reader.EOF())
		}
		if err = reader.Close(); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzGZIPTrailingRaw(f *testing.F) {
	f.Add([]byte("one"), []byte("two"), []byte("RAW"))
	f.Add([]byte{}, []byte{}, []byte{0x1f})
	f.Fuzz(func(t *testing.T, first, second, raw []byte) {
		if len(first) > 128 {
			first = first[:128]
		}
		if len(second) > 128 {
			second = second[:128]
		}
		if len(raw) > 128 {
			raw = raw[:128]
		}
		raw = slices.Clone(raw)
		if len(raw) >= 2 && raw[0] == 0x1f && raw[1] == 0x8b {
			raw[1] ^= 1
		}
		path := filepath.Join(t.TempDir(), "trailing-raw.gz")
		for index, value := range [][]byte{first, second} {
			writer, err := OpenGZIPWriter(path, GZIPWriterOptions{Level: 6, Append: index != 0})
			if err != nil {
				t.Fatal(err)
			}
			if count, err := writer.Write(value); err != nil || count != len(value) {
				t.Fatal(count, err)
			}
			if err = writer.Close(); err != nil {
				t.Fatal(err)
			}
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
		if err != nil {
			t.Fatal(err)
		}
		if count, err := file.Write(raw); err != nil || count != len(raw) {
			t.Fatal(count, err)
		}
		if err = file.Close(); err != nil {
			t.Fatal(err)
		}
		reader, err := OpenGZIPReader(path, GZIPReaderOptions{MaxOutput: int64(len(first) + len(second) + len(raw) + 1)})
		if err != nil {
			t.Fatal(err)
		}
		output, err := io.ReadAll(reader)
		want := make([]byte, 0, len(first)+len(second)+len(raw))
		want = append(want, first...)
		want = append(want, second...)
		want = append(want, raw...)
		if err != nil || !slices.Equal(output, want) || reader.Direct() != (len(raw) != 0) || !reader.EOF() {
			t.Fatal(len(first), len(second), len(raw), len(output), err, reader.Direct(), reader.EOF())
		}
		if err = reader.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
