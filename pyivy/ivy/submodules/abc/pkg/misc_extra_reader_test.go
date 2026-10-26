package abc

// Go-only parser model, file, lifecycle, and fuzz tests for extraUtilReader.c.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestExtraByteReaderTokensLinesPositionAndOwnership(t *testing.T) {
	data := []byte("a b,#comment\nc,d;\n// hello\n/e f;")
	reader, err := NewExtraByteReader("owned.logic", data, "#/", ";", " \t\r\n,", ExtraFileReaderLimits{})
	if err != nil {
		t.Fatal(err)
	}
	data[0] = 'z'
	name, _ := reader.FileName()
	size, _ := reader.FileSize()
	if name != "owned.logic" || size != len(data) {
		t.Fatal(name, size)
	}
	first, ok, err := reader.Next()
	if err != nil || !ok || !slices.Equal(first.Tokens, []string{"a", "b", "c", "d"}) || !slices.Equal(first.Lines, []int{1, 1, 2, 2}) {
		t.Fatal(first, ok, err)
	}
	position, _ := reader.Position()
	if position != 17 {
		t.Fatal(position)
	}
	first.Tokens[0] = "changed"
	first.Lines[0] = 99
	second, ok, err := reader.Next()
	if err != nil || !ok || !slices.Equal(second.Tokens, []string{"/e", "f"}) || !slices.Equal(second.Lines, []int{4, 4}) {
		t.Fatal(second, ok, err)
	}
	line, err := reader.LineNumber(0)
	if err != nil || line != 4 {
		t.Fatal(line, err)
	}
	if _, err = reader.LineNumber(2); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	last, ok, err := reader.Next()
	if err != nil || ok || len(last.Tokens) != 0 {
		t.Fatal(last, ok, err)
	}
	if !reader.Close() || reader.Close() {
		t.Fatal("close lifecycle")
	}
	if _, _, err = reader.Next(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestExtraByteReaderEmptyGroupsPrecedenceAndSlash(t *testing.T) {
	reader, _ := NewExtraByteReader("", []byte(";;x;;"), "", ";", " ", ExtraFileReaderLimits{})
	batch, ok, err := reader.Next()
	if err != nil || !ok || !slices.Equal(batch.Tokens, []string{"x"}) {
		t.Fatal(batch, ok, err)
	}
	if _, ok, err = reader.Next(); err != nil || ok {
		t.Fatal(ok, err)
	}
	precedence, _ := NewExtraByteReader("", []byte("a#b;"), "#", "#;", "#", ExtraFileReaderLimits{})
	batch, ok, err = precedence.Next()
	if err != nil || !ok || !slices.Equal(batch.Tokens, []string{"a", "b"}) {
		t.Fatal(batch, ok, err)
	}
	slash, _ := NewExtraByteReader("", []byte("/a //gone\n/b;"), "/", ";", " \n", ExtraFileReaderLimits{})
	batch, ok, err = slash.Next()
	if err != nil || !ok || !slices.Equal(batch.Tokens, []string{"/a", "/b"}) || !slices.Equal(batch.Lines, []int{1, 2}) {
		t.Fatal(batch, ok, err)
	}
}

func TestExtraFileReaderFileAndErrors(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "tokens.txt")
	if err := os.WriteFile(path, []byte("one,two;"), 0o600); err != nil {
		t.Fatal(err)
	}
	reader, err := NewExtraFileReader(path, "#", ";", ",", ExtraFileReaderLimits{})
	if err != nil {
		t.Fatal(err)
	}
	batch, ok, err := reader.Next()
	if err != nil || !ok || !slices.Equal(batch.Tokens, []string{"one", "two"}) {
		t.Fatal(batch, ok, err)
	}
	if _, err = NewExtraFileReader(filepath.Join(directory, "missing"), "", "", "", ExtraFileReaderLimits{}); err == nil {
		t.Fatal("missing file accepted")
	}
	if _, err = NewExtraFileReader(path, "", "", "", ExtraFileReaderLimits{MaxBytes: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = NewExtraByteReader("", []byte{'a', 0, 'b'}, "", "", "", ExtraFileReaderLimits{}); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
	comment, _ := NewExtraByteReader("", []byte("a #unterminated"), "#", ";", " ", ExtraFileReaderLimits{})
	position, _ := comment.Position()
	if _, _, err = comment.Next(); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
	after, _ := comment.Position()
	if after != position {
		t.Fatal("failed parse changed position", position, after)
	}
	if _, err = NewExtraByteReader("", nil, "", "", "", ExtraFileReaderLimits{MaxTokens: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestExtraByteReaderLimitsCancellationAndAtomicity(t *testing.T) {
	for _, row := range []struct {
		limits ExtraFileReaderLimits
		data   string
	}{{ExtraFileReaderLimits{MaxTokenBytes: 2}, "long;"}, {ExtraFileReaderLimits{MaxTokens: 1}, "a b;"}, {ExtraFileReaderLimits{MaxOperations: 2}, "abc;"}} {
		reader, err := NewExtraByteReader("", []byte(row.data), "", ";", " ", row.limits)
		if err != nil {
			t.Fatal(err)
		}
		before, _ := reader.Position()
		if _, _, err = reader.Next(); !errors.Is(err, ErrLimit) {
			t.Fatal(row, err)
		}
		after, _ := reader.Position()
		if before != after {
			t.Fatal("limit changed position", before, after)
		}
	}
	if reader, err := NewExtraByteReader("", []byte("a;"), "", ";", "", ExtraFileReaderLimits{Stop: func() bool { return true }}); reader != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(reader, err)
	}
	calls := 0
	reader, _ := NewExtraByteReader("", []byte("abcdef;"), "", ";", "", ExtraFileReaderLimits{Stop: func() bool { calls++; return calls > 1 }})
	if _, _, err := reader.Next(); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	position, _ := reader.Position()
	if position != 0 {
		t.Fatal(position)
	}
}

func extraReaderModel(data []byte) []ExtraTokenBatch {
	result := []ExtraTokenBatch{}
	batch := ExtraTokenBatch{}
	line, tokenLine := 1, 0
	var token []byte
	finish := func() {
		if token != nil {
			batch.Tokens = append(batch.Tokens, string(token))
			batch.Lines = append(batch.Lines, tokenLine)
			token = nil
		}
	}
	flush := func() {
		finish()
		if len(batch.Tokens) > 0 {
			result = append(result, batch)
		}
		batch = ExtraTokenBatch{}
	}
	for position := 0; position < len(data); {
		character := data[position]
		if character == '#' || character == '/' && position+1 < len(data) && data[position+1] == '/' {
			finish()
			for position < len(data) && data[position] != '\n' {
				position++
			}
			continue
		}
		if character == ';' {
			flush()
			position++
			continue
		}
		if character == ' ' || character == ',' || character == '\n' {
			finish()
			if character == '\n' {
				line++
			}
			position++
			continue
		}
		if token == nil {
			token = []byte{}
			tokenLine = line
		}
		token = append(token, character)
		position++
	}
	flush()
	return result
}

func FuzzExtraFileReader(f *testing.F) {
	for _, seed := range [][]byte{[]byte("a,b;c"), []byte("//x\na;"), []byte("/a #x\nb;")} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		alphabet := []byte{'a', 'b', ' ', ',', ';', '#', '\n', '/'}
		data := make([]byte, min(len(raw), 256)+1)
		for i := 0; i < len(data)-1; i++ {
			data[i] = alphabet[int(raw[i])%len(alphabet)]
		}
		data[len(data)-1] = '\n'
		reader, err := NewExtraByteReader("fuzz", data, "#/", ";", " ,\n", ExtraFileReaderLimits{MaxBytes: 1024, MaxTokens: 1024, MaxTokenBytes: 1024, MaxOperations: 4096})
		if err != nil {
			t.Fatal(err)
		}
		got := []ExtraTokenBatch{}
		for {
			batch, ok, nextErr := reader.Next()
			if nextErr != nil {
				t.Fatal(nextErr, string(data))
			}
			if !ok {
				break
			}
			got = append(got, batch)
		}
		want := extraReaderModel(data)
		if len(got) != len(want) {
			t.Fatal("batch count", string(data), got, want)
		}
		for i := range want {
			if !slices.Equal(got[i].Tokens, want[i].Tokens) || !slices.Equal(got[i].Lines, want[i].Lines) {
				t.Fatal(i, string(data), got[i], want[i])
			}
		}
	})
}
