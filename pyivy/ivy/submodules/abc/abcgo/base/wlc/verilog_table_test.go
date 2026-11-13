package wlc

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestWriteTableOneMatchesC(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	source, err := os.ReadFile(filepath.Join(filepath.Dir(file), "../../../src/base/wlc/wlcWriteVer.c"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	start := strings.Index(text, "void Wlc_WriteTableOne(")
	if start < 0 {
		t.Fatal("C table writer missing")
	}
	brace := strings.IndexByte(text[start:], '{') + start
	depth, end := 0, brace
	for ; end < len(text); end++ {
		if text[end] == '{' {
			depth++
		}
		if text[end] == '}' {
			depth--
			if depth == 0 {
				end++
				break
			}
		}
	}
	preamble := "#include <stdio.h>\n#include <stdlib.h>\n#include <assert.h>\ntypedef unsigned long long word;\nstatic word Abc_Tt6Mask(int n) {return (~(word)0)>>(64-n);}\n"
	main := "\nint main(int argc,char **argv) {word t[4]={0x0123456789abcdefULL,0xfedcba9876543210ULL,0xaaaaaaaa55555555ULL,0xf123456789abcdefULL};Wlc_WriteTableOne(stdout,2,atoi(argv[1]),t,7);return 0;}\n"
	dir := t.TempDir()
	cfile, binary := filepath.Join(dir, "table.c"), filepath.Join(dir, "table")
	if err = os.WriteFile(cfile, []byte(preamble+text[start:end]+main), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(cc, cfile, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile C: %v: %s", err, output)
	}
	words := []uint64{0x0123456789abcdef, 0xfedcba9876543210, 0xaaaaaaaa55555555, 0xf123456789abcdef}
	for _, width := range []int32{1, 8, 16, 32, 64} {
		want, err := exec.Command(binary, strconv.Itoa(int(width))).Output()
		if err != nil {
			t.Fatal(err)
		}
		var got bytes.Buffer
		Wlc_WriteTableOne(&got, 2, width, words, 7)
		if !bytes.Equal(got.Bytes(), want) {
			t.Fatalf("width %d differs\nGo: %s\nC: %s", width, got.Bytes(), want)
		}
	}
}
