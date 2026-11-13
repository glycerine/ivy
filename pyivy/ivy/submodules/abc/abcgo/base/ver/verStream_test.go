package ver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStreamWordsPositionAndLines(t *testing.T) {
	name := filepath.Join(t.TempDir(), "input.v")
	if err := os.WriteFile(name, []byte("alpha\n beta;"), 0600); err != nil {
		t.Fatal(err)
	}
	p := Ver_StreamAlloc(name)
	if p == nil {
		t.Fatal("open failed")
	}
	defer Ver_StreamFree(p)
	if got := Ver_StreamGetWord(p, " \n;"); got != "alpha" {
		t.Fatal(got)
	}
	Ver_StreamSkipChars(p, " \n")
	if Ver_StreamGetCurPosition(p) != 7 || Ver_StreamGetLineNumber(p) != 2 {
		t.Fatal("wrong location")
	}
	if got := Ver_StreamGetWord(p, ";"); got != "beta" {
		t.Fatal(got)
	}
	if Ver_StreamPopChar(p) != ';' || Ver_StreamPopChar(p) != 0xff || Ver_StreamIsOkey(p) != 0 {
		t.Fatal("wrong end of file")
	}
}

func TestStreamReloadAcrossBufferBoundary(t *testing.T) {
	name := filepath.Join(t.TempDir(), "large.v")
	data := strings.Repeat(" ", VER_BUFFER_SIZE-VER_OFFSET_SIZE+1) + "token;" + strings.Repeat(" ", VER_OFFSET_SIZE)
	if err := os.WriteFile(name, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	p := Ver_StreamAlloc(name)
	if p == nil {
		t.Fatal("open failed")
	}
	defer Ver_StreamFree(p)
	Ver_StreamSkipChars(p, " ")
	if got := Ver_StreamGetWord(p, ";"); got != "token" {
		t.Fatal(got)
	}
	if Ver_StreamGetCurPosition(p) != int32(VER_BUFFER_SIZE-VER_OFFSET_SIZE+6) {
		t.Fatal("wrong reload position")
	}
}
