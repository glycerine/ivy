package gia_test

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/internal/testoracle"
	if_ "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/opt/dau"
	"strings"
	"testing"
)

func TestMappingDSDNativeCanonicalCellTuning(t *testing.T) {
	executable := gia.CompileSourceOracleForTest(t, "if_dsd_tuning_review.c")
	source, err := testoracle.Command(t, executable).CombinedOutput()
	if err != nil {
		t.Fatalf("original tuning: %v %s", err, source)
	}
	var expected strings.Builder
	for _, line := range strings.Split(string(source), "\n") {
		if strings.HasPrefix(line, "contract ") || strings.HasPrefix(line, "node ") || strings.HasPrefix(line, "two-lut ") {
			expected.WriteString(line + "\n")
		}
	}
	if expected.Len() == 0 {
		t.Fatalf("missing source tuning contract: %s", source)
	}
	p := if_.If_DsdManAlloc(4, 0)
	defer if_.If_DsdManFree(p, 0)
	if_.If_DsdManCompute(p, dau.Dau_DsdToTruth("((ab)[cd])", 4), 4, make([]byte, 4), "")
	if_.Id_DsdManTuneStr1(p, "{abcd}", 0, 0)
	var got strings.Builder
	fmt.Fprintf(&got, "contract %d %d %d\n", if_.If_DsdManObjNum(p), if_.If_DsdManLutSize(p), len(if_.If_DsdManGetFuncConfig(p, 0)))
	for id := int32(0); id < if_.If_DsdManObjNum(p); id++ {
		fmt.Fprintf(&got, "node %d %d", id, if_.If_DsdManReadMark(p, id<<1))
		for _, value := range if_.If_DsdManGetFuncConfig(p, id<<1) {
			fmt.Fprintf(&got, " %016x", value)
		}
		got.WriteByte('\n')
	}
	q := if_.If_DsdManAlloc(11, 0)
	defer if_.If_DsdManFree(q, 0)
	if_.If_DsdManCompute(q, dau.Dau_DsdToTruth("[(abcdefghij)k]", 11), 11, make([]byte, 11), "")
	if_.If_DsdManTune(q, 6, 0, 0, 0, 0)
	fmt.Fprintf(&got, "two-lut %d", if_.If_DsdManObjNum(q))
	for id := int32(0); id < if_.If_DsdManObjNum(q); id++ {
		fmt.Fprintf(&got, " %d", if_.If_DsdManReadMark(q, id<<1))
	}
	got.WriteByte('\n')
	if got.String() != expected.String() {
		t.Fatalf("Go tuning:\n%soriginal C:\n%s", got.String(), expected.String())
	}
}
