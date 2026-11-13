package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"os"
	"strings"
	"testing"
)

func TestVerifyClassesReportsOnlyTruthMismatches(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	and := gia.Gia_ManAppendAnd(p, a, b)
	or := gia.Gia_ManAppendOr(p, a, b)
	lits, classes := vec.Vec_WecStart(1), vec.Vec_WecStart(1)
	vec.Vec_WecPush(lits, 0, and)
	vec.Vec_WecPush(lits, 0, or)
	vec.Vec_WecPush(classes, 0, 4)
	vec.Vec_WecPush(classes, 0, 4)
	old := os.Stdout
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = write
	defer func() { os.Stdout = old }()
	Acec_VerifyClasses(p, lits, classes)
	write.Close()
	out, err := io.ReadAll(read)
	read.Close()
	if err != nil || !strings.Contains(string(out), "Total errors in equivalence classes = 1") {
		t.Fatalf("output %q error %v", out, err)
	}
}
