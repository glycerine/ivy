package startingpoint

import (
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCommentEncoding(t *testing.T) {
	source := "// café \x92 �\n/* \xa8\xff */\nchar *s = \"\x92\"; '\xa8'; code\xff;\n"
	want := "// café _ �\n/* __ */\nchar *s = \"\x92\"; '\xa8'; code\xff;\n"
	got, err := SanitizeCommentUTF8(source)
	if err != nil {
		t.Fatal(err)
	}
	if got != want || len(got) != len(source) {
		t.Fatalf("got %q", got)
	}
}

func TestABCCommentEncoding(t *testing.T) {
	for _, path := range []string{"../src/aig/gia/giaEmbed.c", "../src/opt/dau/dauCanon.c", "../src/proof/int/int.h", "../src/proof/int2/int2.h"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if utf8.Valid(source) {
			t.Fatalf("fixture no longer has invalid encoding: %s", path)
		}
		r, err := Analyze(path, string(source))
		if err != nil {
			t.Fatal(err)
		}
		if !utf8.ValidString(r.Draft) {
			t.Fatalf("invalid output encoding: %s", path)
		}
	}
}

func TestCommentEncodingInMacroAndFunction(t *testing.T) {
	source := "#define M(x) /* \x92 */ x\nstatic inline int f() { /* \xa8 */ return 1; }"
	r, err := Analyze("p/file.h", source)
	if err != nil {
		t.Fatal(err)
	}
	if !utf8.ValidString(r.Draft) || !utf8.ValidString(r.Functions[0].Draft) {
		t.Fatal("invalid draft encoding")
	}
	if !strings.Contains(r.Functions[0].Original, "\xa8") {
		t.Fatal("original changed")
	}
}
