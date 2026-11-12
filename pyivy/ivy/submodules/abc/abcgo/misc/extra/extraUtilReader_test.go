package extra

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileReaderTokenGroups(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input")
	input := "# header\na b\n\nc/d // comment\ne\tlast# eof comment"
	if err := os.WriteFile(path, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	p := Extra_FileReaderAlloc(path, "#/", "\n", " \t\r")
	if p == nil {
		t.Fatal("open")
	}
	defer Extra_FileReaderFree(p)
	want := [][]string{{"a", "b"}, {"c/d"}, {"e", "last"}}
	for _, row := range want {
		got := Extra_FileReaderGetTokens(p)
		if strings.Join(got, "|") != strings.Join(row, "|") {
			t.Fatal(got, row)
		}
	}
	if Extra_FileReaderGetTokens(p) != nil || Extra_FileReaderErr(p) != nil {
		t.Fatal("EOF")
	}
	if Extra_FileReaderGetCurPosition(p) != int64(len(input)) || Extra_FileReaderGetFileSize(p) != int64(len(input)) {
		t.Fatal("position")
	}
}
func TestFileReaderLongTokensAndLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input")
	long := strings.Repeat("x", 100000)
	if err := os.WriteFile(path, []byte("\n"+long+" z\nlast"), 0600); err != nil {
		t.Fatal(err)
	}
	p := Extra_FileReaderAlloc(path, "#", "\n", " ")
	defer Extra_FileReaderFree(p)
	row := Extra_FileReaderGetTokens(p)
	if len(row) != 2 || row[0] != long || Extra_FileReaderGetLineNumber(p, 0) != 2 {
		t.Fatal("long token")
	}
	if got := Extra_FileReaderGetTokens(p); len(got) != 1 || got[0] != "last" || Extra_FileReaderGetLineNumber(p, 0) != 3 {
		t.Fatal(got)
	}
	if row[0] != long {
		t.Fatal("retained token mutated")
	}
}

func TestFileReaderCharacterPrecedenceAndFree(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input")
	if err := os.WriteFile(path, []byte("a#b/c//tail\nz"), 0600); err != nil {
		t.Fatal(err)
	}
	// Clean overrides stop, which overrides comment, matching C initialization.
	p := Extra_FileReaderAlloc(path, "#/", "#\n", "#")
	got := Extra_FileReaderGetTokens(p)
	if strings.Join(got, "|") != "a|b/c" {
		t.Fatal(got)
	}
	got = Extra_FileReaderGetTokens(p)
	if len(got) != 1 || got[0] != "z" {
		t.Fatal(got)
	}
	Extra_FileReaderFree(p)
	Extra_FileReaderFree(p)
	if Extra_FileReaderGetTokens(p) != nil {
		t.Fatal("reader used after free")
	}
}
