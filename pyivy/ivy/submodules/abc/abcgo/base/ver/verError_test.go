package ver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestParseErrorReportsLineAndReleasesDesign(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "input.v")
	if err := os.WriteFile(source, []byte("a\nb"), 0600); err != nil {
		t.Fatal(err)
	}
	reader := Ver_StreamAlloc(source)
	defer Ver_StreamFree(reader)
	Ver_StreamPopChar(reader)
	Ver_StreamPopChar(reader)
	out, err := os.CreateTemp(dir, "errors")
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	p := &Ver_Man_t{pFileName: "input.v", pReader: reader, Output: out, pDesign: abc.Abc_DesCreate("example")}
	copy(p.sError[:], "bad token")
	Ver_ParsePrintErrorMessage(p)
	if p.fError != 1 || p.pDesign != nil {
		t.Fatal("error state or cleanup missing")
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "input.v (line 2): bad token\n" {
		t.Fatalf("error output %q", got)
	}
}
