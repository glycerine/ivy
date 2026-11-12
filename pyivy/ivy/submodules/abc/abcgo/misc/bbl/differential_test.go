package bbl

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOriginalCInteroperability(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	root := "../../../src/misc/bbl"
	header, err := os.ReadFile(filepath.Join(root, "bblif.h"))
	if err != nil {
		t.Fatal(err)
	}
	implementation, err := os.ReadFile(filepath.Join(root, "bblif.c"))
	if err != nil {
		t.Fatal(err)
	}
	var code strings.Builder
	code.WriteString("#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n#include <assert.h>\n#define ABC_NAMESPACE_HEADER_START\n#define ABC_NAMESPACE_HEADER_END\n#define ABC_NAMESPACE_IMPL_START\n#define ABC_NAMESPACE_IMPL_END\n")
	code.Write(header)
	for _, line := range strings.Split(string(implementation), "\n") {
		if !strings.HasPrefix(line, "#include") {
			code.WriteString(line + "\n")
		}
	}
	code.WriteString(`int main(int argc,char **argv) {
 if(argc==1) { Bbl_ManSimpleDemo(); return 0; }
 Bbl_Man_t *p=Bbl_ManReadBinaryBlif(argv[1]); Bbl_Obj_t *o; int count=0;
 printf("%s\n",Bbl_ManName(p));
 Bbl_ManForEachObj(p,o) { count++; if(Bbl_ObjIsLut(o)) { unsigned *t=Bbl_ManSopToTruth(Bbl_ObjSop(p,o),Bbl_ObjFaninNumber(o)); printf("%08x\n",t[0]);free(t); } }
 printf("objects=%d\n",count); Bbl_ManStop(p); return 0;
}
`)
	dir := t.TempDir()
	source := filepath.Join(dir, "reference.c")
	binary := filepath.Join(dir, "reference")
	if err := os.WriteFile(source, []byte(code.String()), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, source, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("reference compile: %v: %s", err, out)
	}
	cmd := exec.Command(binary)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("C write: %v: %s", err, out)
	}
	p, err := Bbl_ManReadBinaryBlif(filepath.Join(dir, "hadder.bblif"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "hadder" || len(p.Objects) != 7 {
		t.Fatal(p)
	}
	if got := Bbl_ManSopToTruth(p.Objects[5].SOP, 3)[0]; got != 0x96969696 {
		t.Fatalf("C-written parity truth %x", got)
	}
	if got := Bbl_ManSopToTruth(p.Objects[6].SOP, 3)[0]; got != 0xe8e8e8e8 {
		t.Fatalf("C-written majority truth %x", got)
	}
	var wire bytes.Buffer
	if err := Bbl_ManSimpleDemo().WriteBinary(&wire); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "go.bblif")
	if err := os.WriteFile(path, wire.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(binary, path).CombinedOutput()
	if err != nil {
		t.Fatalf("C read Go: %v: %s", err, out)
	}
	if string(out) != "hadder\n96969696\ne8e8e8e8\nobjects=7\n" {
		t.Fatal(string(out))
	}
}
