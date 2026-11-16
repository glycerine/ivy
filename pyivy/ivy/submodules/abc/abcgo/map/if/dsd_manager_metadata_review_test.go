package if_

import (
	"github.com/glycerine/abcgo/opt/dau"
	"slices"
	"testing"
)

func dsdMetadataReviewManager(t *testing.T) (*If_DsdMan_t, int32) {
	t.Helper()
	p := If_DsdManAlloc(4, 0)
	t.Cleanup(func() { If_DsdManFree(p, 0) })
	If_DsdManCompute(p, dau.Dau_DsdToTruth("(ab)", 2), 2, make([]byte, 2), "")
	var root int32
	for i := 0; i < 3; i++ {
		root = If_DsdManCompute(p, dau.Dau_DsdToTruth("((ab)[cd])", 4), 4, make([]byte, 4), "")
	}
	return p, root
}
func TestDsdFilterUsesOccurrencesAndRemapsDependencies(t *testing.T) {
	p, root := dsdMetadataReviewManager(t)
	expected := append([]word(nil), If_DsdManComputeTruth(p, root, nil)...)
	q := If_DsdManFilter(p, 2)
	defer If_DsdManFree(q, 0)
	if got := If_DsdManObjNum(q); got != 4 {
		t.Fatalf("filter kept%d nodes, want bases + required XOR + frequent root", got)
	}
	last := If_DsdManObjNum(q) - 1
	if !slices.Equal(If_DsdManComputeTruth(q, last<<1, nil), expected) {
		t.Fatal("filter did not remap canonical fanins")
	}
	if If_DsdVecObj(&q.vObjs, last).Count != 3 {
		t.Fatal("filter lost occurrence count")
	}
}
func TestDsdMergeInternsAndRemapsCanonicalNodes(t *testing.T) {
	src, root := dsdMetadataReviewManager(t)
	expected := append([]word(nil), If_DsdManComputeTruth(src, root, nil)...)
	dst := If_DsdManAlloc(4, 0)
	defer If_DsdManFree(dst, 0)
	If_DsdManCompute(dst, dau.Dau_DsdToTruth("[ab]", 2), 2, make([]byte, 2), "")
	If_DsdManMerge(dst, src)
	if got := If_DsdManObjNum(dst); got != 5 {
		t.Fatalf("merge has%d nodes, want shared bases/XOR and two added AND structures", got)
	}
	if !slices.Equal(If_DsdManComputeTruth(dst, 8, nil), expected) {
		t.Fatal("merge did not remap canonical fanins")
	}
	if If_DsdVecObj(&dst.vObjs, 4).Count != 0 {
		t.Fatal("source merge does not import occurrence counts")
	}
}
func TestDsdMarkResetsClearSourceConfigurationState(t *testing.T) {
	for _, invert := range []bool{false, true} {
		t.Run(map[bool]string{false: "clean", true: "invert"}[invert], func(t *testing.T) {
			p := If_DsdManAlloc(2, 0)
			defer If_DsdManFree(p, 0)
			Id_DsdManTuneStr1(p, "{ab}", 0, 0)
			config := p.vConfigs
			if invert {
				If_DsdManInvertMarks(p, 0)
			} else {
				If_DsdManCleanMarks(p, 0)
			}
			if p.pCellStr != "" {
				t.Errorf("mark reset retained cell description%q", p.pCellStr)
			}
			if invert && p.vConfigs != config {
				t.Error("invert must retain source configurations")
			}
			if !invert && p.vConfigs != nil {
				t.Error("clean must release source configurations")
			}
			for id := int32(0); id < If_DsdManObjNum(p); id++ {
				expected := int32(0)
				if invert {
					expected = 1
				}
				if If_DsdVecObj(&p.vObjs, id).fMark != expected {
					t.Errorf("node%d mark was not reset", id)
				}
			}
		})
	}
}

func TestDsdCheckDecUsesStoredMark(t *testing.T) {
	p := If_DsdManAlloc(5, 3)
	defer If_DsdManFree(p, 0)
	for _, text := range []string{"(abcde)", "[ab]"} {
		n := dau.Dau_DsdFindVarNum(text)
		lit := If_DsdManCompute(p, dau.Dau_DsdToTruth(text, n), n, make([]byte, n), "")
		// AND5 fits two LUT3s despite having support greater than one LUT.
		// Marks can later change during cell tuning, independently of support.
		for _, mark := range []int32{0, 1, 0} {
			If_DsdVecObj(&p.vObjs, lit>>1).fMark = mark
			for _, phase := range []int32{0, 1} {
				if got := If_DsdManCheckDec(p, lit^phase); got != mark {
					t.Errorf("%s phase%d stored mark%d, CheckDec%d", text, phase, mark, got)
				}
			}
		}
	}
}
