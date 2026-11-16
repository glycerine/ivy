package gia_test

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	if_ "github.com/glycerine/abcgo/map/if"
	"os/exec"
	"strings"
	"testing"
)

func TestMappingDSDOriginalCanonicalContracts(t *testing.T) {
	executable := gia.CompileSourceOracleForTest(t, "if_dsd_canonical_review.c")
	want, err := exec.Command(executable).CombinedOutput()
	if err != nil {
		t.Fatalf("original DSD: %v %s", err, want)
	}
	p := if_.If_DsdManAlloc(3, 2)
	defer if_.If_DsdManFree(p, 0)
	var got strings.Builder
	fmt.Fprintf(&got, "initial %d %s\n", if_.If_DsdManObjNum(p), if_.If_DsdManFileName(p))
	truth := []uint64{0x8080808080808080}
	perm := make([]byte, 3)
	lit := if_.If_DsdManCompute(p, truth, 3, perm, "22")
	fmt.Fprintf(&got, "computed %d %d %d permutation", lit, if_.If_DsdManObjNum(p), if_.If_DsdManSuppSize(p, lit))
	for _, pin := range perm {
		fmt.Fprintf(&got, " %d", pin)
	}
	got.WriteByte('\n')
	fmt.Fprintf(&got, "partition %x\n", if_.If_DsdManCheckXY(p, lit, 2, 1, 0, 1, 0))
	lit = if_.If_DsdManCompute(p, truth, 3, perm, "22")
	fmt.Fprintf(&got, "reused %d %d\n", lit, if_.If_DsdManObjNum(p))
	if got.String() != string(want) {
		t.Fatalf("Go DSD:\n%soriginal C:\n%s", got.String(), want)
	}
}
