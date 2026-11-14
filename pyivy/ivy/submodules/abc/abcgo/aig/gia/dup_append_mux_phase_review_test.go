package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func appendMuxPhaseReviewGraph(signs int32) *Gia_Man_t {
	src := Gia_ManStart(1)
	a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	src.EnableMuxes()
	m := Gia_ManAppendMuxReal(src, c^((signs>>2)&1), b^(signs&1), a^((signs>>1)&1))
	Gia_ManAppendCo(src, m^((signs>>3)&1))
	Gia_ManAppendCo(src, Gia_ManAppendAnd(src, m, a^1))
	return src
}

func appendMuxPhaseReviewRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	for signs := int32(0); signs < 16; signs++ {
		src := appendMuxPhaseReviewGraph(signs)
		dst := Gia_ManStart(1)
		d := Gia_ManAppendCi(dst)
		Gia_ManAppendCo(dst, d^1)
		oldObjects := dst.nObjs
		Gia_ManDupAppend(dst, src)
		if dst.nObjs != oldObjects+src.nObjs-1 || dst.nMuxes != 1 || Gia_ManCiNum(dst) != 4 || Gia_ManCoNum(dst) != 3 {
			t.Fatal("append changed source node kinds or terminal counts")
		}
		fmt.Fprintf(&records, "%d %d %d", signs, Gia_ManCiNum(dst), Gia_ManCoNum(dst))
		for mask := 0; mask < 16; mask++ {
			d, a, b, c := mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0
			mux := a != (signs&2 != 0)
			if c != (signs&4 != 0) {
				mux = b != (signs&1 != 0)
			}
			checkQuantificationOutputs(t, dst, []bool{d, a, b, c}, []bool{!d, mux != (signs&8 != 0), mux && !a})
			values, err := dst.Evaluate([]bool{d, a, b, c})
			if err != nil {
				t.Fatal(err)
			}
			packed := 0
			for i, value := range values {
				if value {
					packed |= 1 << uint(i)
				}
			}
			fmt.Fprintf(&records, " %d", packed)
		}
		records.WriteByte('\n')
	}
	return records.String()
}

func TestUnhashedAppendPreservesMuxDataPhases(t *testing.T) {
	appendMuxPhaseReviewRecords(t)
}

func TestDfsMuxPhasesWithReorderedLeaves(t *testing.T) {
	for signs := int32(0); signs < 16; signs++ {
		for _, hash := range []bool{false, true} {
			src := appendMuxPhaseReviewGraph(signs)
			dst := Gia_ManStart(1)
			dst.EnableMuxes()
			Gia_ManFillValue(src)
			src.pObjs[0].Value = 0
			for i := int32(2); i >= 0; i-- {
				Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(dst))
			}
			if hash {
				Gia_ManHashStart(dst)
			}
			for i := int32(0); i < 2; i++ {
				obj := Gia_ManCo(src, i)
				lit := Gia_ManDupOrderDfs_rec(dst, src, obj)
				objects := dst.nObjs
				if got := Gia_ManDupOrderDfs_rec(dst, src, obj); got != lit || dst.nObjs != objects {
					t.Fatal("DFS reused a phase incorrectly or duplicated a copied cone")
				}
			}
			if dst.nObjs != src.nObjs || dst.nMuxes != 1 {
				t.Fatal("DFS changed source node kinds or counts")
			}
			for mask := 0; mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				want, err := src.Evaluate([]bool{inputs[2], inputs[1], inputs[0]})
				if err != nil {
					t.Fatal(err)
				}
				checkQuantificationOutputs(t, dst, inputs, want)
			}
		}
	}
}

func TestAppendMuxPhasesMatchOriginalExpandedGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "append_mux_phases.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C append MUX phases: %v: %s", err, want)
	}
	gotRows := strings.Split(strings.TrimSpace(appendMuxPhaseReviewRecords(t)), "\n")
	wantRows := strings.Split(strings.TrimSpace(string(want)), "\n")
	if len(gotRows) != len(wantRows) {
		t.Fatalf("Go cases %d, C cases %d", len(gotRows), len(wantRows))
	}
	for i := range gotRows {
		if gotRows[i] != wantRows[i] {
			t.Fatalf("case %d:\nGo %s\nC  %s", i, gotRows[i], wantRows[i])
		}
	}
}
