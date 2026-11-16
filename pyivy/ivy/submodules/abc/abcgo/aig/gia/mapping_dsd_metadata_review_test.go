package gia_test

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	if_ "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/opt/dau"
	"os/exec"
	"strings"
	"testing"
)

func mappingDSDReviewBit(value bool) int {
	if value {
		return 1
	}
	return 0
}
func TestMappingDSDNativeManagerMetadata(t *testing.T) {
	executable := gia.CompileSourceOracleForTest(t, "if_dsd_metadata_review.c")
	source, err := exec.Command(executable).CombinedOutput()
	if err != nil {
		t.Fatalf("original DSD metadata: %v %s", err, source)
	}
	var expected, got strings.Builder
	for _, line := range strings.Split(string(source), "\n") {
		if strings.HasPrefix(line, "meta ") {
			expected.WriteString(line + "\n")
		}
	}
	p := if_.If_DsdManAlloc(4, 0)
	defer if_.If_DsdManFree(p, 0)
	if_.If_DsdManCompute(p, dau.Dau_DsdToTruth("(ab)", 2), 2, make([]byte, 2), "")
	for i := 0; i < 3; i++ {
		if_.If_DsdManCompute(p, dau.Dau_DsdToTruth("((ab)[cd])", 4), 4, make([]byte, 4), "")
	}
	q := if_.If_DsdManFilter(p, 2)
	fmt.Fprintf(&got, "meta filter %d %016x\n", if_.If_DsdManObjNum(q), if_.If_DsdManComputeTruth(q, 2*(if_.If_DsdManObjNum(q)-1), nil)[0])
	if_.If_DsdManFree(q, 0)
	dst := if_.If_DsdManAlloc(4, 0)
	if_.If_DsdManCompute(dst, dau.Dau_DsdToTruth("[ab]", 2), 2, make([]byte, 2), "")
	if_.If_DsdManMerge(dst, p)
	fmt.Fprintf(&got, "meta merge %d %016x\n", if_.If_DsdManObjNum(dst), if_.If_DsdManComputeTruth(dst, 2*(if_.If_DsdManObjNum(dst)-1), nil)[0])
	if_.If_DsdManFree(dst, 0)
	r := if_.If_DsdManAlloc(2, 0)
	defer if_.If_DsdManFree(r, 0)
	if_.Id_DsdManTuneStr1(r, "{ab}", 0, 0)
	config := if_.If_DsdManGetFuncConfig(r, 0)
	if_.If_DsdManInvertMarks(r, 0)
	after := if_.If_DsdManGetFuncConfig(r, 0)
	same := len(config) > 0 && len(after) > 0 && &config[0] == &after[0]
	fmt.Fprintf(&got, "meta invert %d %d %d", mappingDSDReviewBit(if_.If_DsdManGetCellStr(r) != ""), mappingDSDReviewBit(after != nil), mappingDSDReviewBit(same))
	for id := int32(0); id < if_.If_DsdManObjNum(r); id++ {
		fmt.Fprintf(&got, " %d", if_.If_DsdManReadMark(r, id<<1))
	}
	got.WriteByte('\n')
	if_.If_DsdManCleanMarks(r, 0)
	fmt.Fprintf(&got, "meta clean %d %d", mappingDSDReviewBit(if_.If_DsdManGetCellStr(r) != ""), mappingDSDReviewBit(if_.If_DsdManGetFuncConfig(r, 0) != nil))
	for id := int32(0); id < if_.If_DsdManObjNum(r); id++ {
		fmt.Fprintf(&got, " %d", if_.If_DsdManReadMark(r, id<<1))
	}
	got.WriteByte('\n')
	if got.String() != expected.String() {
		t.Fatalf("Go metadata:\n%soriginal C:\n%s", got.String(), expected.String())
	}
}
