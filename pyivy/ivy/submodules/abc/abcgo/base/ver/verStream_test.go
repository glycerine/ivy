package ver

import (
	"fmt"
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
	if got := Ver_StreamGetWord(p, " \n;"); got == nil || *got != "alpha" {
		t.Fatal(got)
	}
	Ver_StreamSkipChars(p, " \n")
	if Ver_StreamGetCurPosition(p) != 7 || Ver_StreamGetLineNumber(p) != 2 {
		t.Fatal("wrong location")
	}
	if got := Ver_StreamGetWord(p, ";"); got == nil || *got != "beta" {
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
	if got := Ver_StreamGetWord(p, ";"); got == nil || *got != "token" {
		t.Fatal(got)
	}
	if Ver_StreamGetCurPosition(p) != int32(VER_BUFFER_SIZE-VER_OFFSET_SIZE+6) {
		t.Fatal("wrong reload position")
	}
}

func TestStreamEmptyWordDiffersFromNull(t *testing.T) {
	name := filepath.Join(t.TempDir(), "empty.v")
	if err := os.WriteFile(name, nil, 0600); err != nil {
		t.Fatal(err)
	}
	p := Ver_StreamAlloc(name)
	defer Ver_StreamFree(p)
	if word := Ver_StreamGetWord(p, ";"); word == nil || *word != "" {
		t.Fatalf("first EOF word = %v", word)
	}
	if word := Ver_StreamGetWord(p, ";"); word != nil {
		t.Fatalf("stopped stream returned %q", *word)
	}
}

// verStream.c permits VER_WORD_SIZE-1 bytes, but rejects the next byte
// without advancing the stream cursor.
func TestStreamWordLengthBoundary(t *testing.T) {
	for _, n := range []int{VER_WORD_SIZE - 1, VER_WORD_SIZE} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			name := filepath.Join(t.TempDir(), "word.v")
			if err := os.WriteFile(name, []byte(strings.Repeat("a", n)+";"), 0600); err != nil {
				t.Fatal(err)
			}
			p := Ver_StreamAlloc(name)
			defer Ver_StreamFree(p)
			word := Ver_StreamGetWord(p, ";")
			if n == VER_WORD_SIZE {
				if word != nil || Ver_StreamGetCurPosition(p) != 0 || Ver_StreamIsOkey(p) != 1 {
					t.Fatal("overlong word changed source stream state")
				}
			} else if word == nil || len(*word) != n || Ver_StreamGetCurPosition(p) != int32(n) {
				t.Fatal("valid boundary word rejected")
			}
		})
	}
}
