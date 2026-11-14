package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func inputDropReviewGraph(keep, signs int32, last bool) *Gia_Man_t {
	g := Gia_ManStart(1)
	ci := make([]int32, 6)
	for i := range ci {
		ci[i] = Gia_ManAppendCi(g)
	}
	if last {
		ci = ci[6-keep:]
	} else {
		ci = ci[:keep]
	}
	g.EnableMuxes()
	m := signs & 1
	if keep != 0 {
		m = ci[0] ^ (signs & 1)
		if keep > 1 {
			m = Gia_ManAppendXorReal(g, m, ci[1])
		}
		m = Gia_ManAppendBuf(g, m^((signs>>1)&1))
		if keep > 2 {
			m = Gia_ManAppendMuxReal(g, ci[2]^((signs>>2)&1), m, ci[0]^1)
		}
		for i := int32(3); i < keep; i++ {
			m = Gia_ManAppendXorReal(g, m, ci[i])
		}
	}
	for _, lit := range []int32{m, m ^ 1, (signs >> 2) & 1} {
		Gia_ManAppendCo(g, lit)
	}
	g.SetIdentity("drop", "drop.aig")
	return g
}

func inputDualReviewRecords(t *testing.T, dfs bool) string {
	t.Helper()
	var records strings.Builder
	check := func(src, out *Gia_Man_t, mode int, signs, parameter int32) {
		t.Helper()
		wantCI, wantCO, wantReg := parameter, int32(3), int32(0)
		if mode == 2 {
			wantCI, wantCO, wantReg = 6, 12, src.nRegs
		}
		if Gia_ManCiNum(out) != wantCI || Gia_ManCoNum(out) != wantCO || out.nRegs != wantReg || out.nConstrs != 0 {
			t.Fatal("input/dual conversion changed interface")
		}
		wantObjects := src.nObjs - (6 - parameter)
		if mode == 2 {
			wantObjects = src.nObjs + Gia_ManCoNum(src)
		}
		if out.nObjs != wantObjects || out.nMuxes != src.nMuxes || out.nXors != src.nXors || out.nBufs != src.nBufs {
			t.Fatal("input/dual conversion changed internal gate kinds or counts")
		}
		name, spec := src.Identity()
		if gotName, gotSpec := out.Identity(); gotName != name || gotSpec != spec {
			t.Fatal("input/dual conversion changed identity")
		}
		fmt.Fprintf(&records, "%d %d %d %d %d %d", mode, signs, parameter, wantCI, wantCO, wantReg)
		for mask := 0; mask < 1<<uint(wantCI); mask++ {
			inputs := make([]bool, wantCI)
			for i := range inputs {
				inputs[i] = mask&(1<<uint(i)) != 0
			}
			original := make([]bool, 6)
			offset := 0
			if mode == 1 {
				offset = int(6 - parameter)
			}
			copy(original[offset:], inputs)
			values, err := src.Evaluate(original)
			if err != nil {
				t.Fatal(err)
			}
			want := values
			if mode == 2 {
				// Source doubles every raw CO, including RI terminals. The old
				// register count is retained; this is not two state machines.
				want = nil
				for _, value := range values {
					want = append(want, value, !value)
				}
			}
			checkQuantificationOutputs(t, out, inputs, want)
			got, err := out.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			packed := 0
			for i, value := range got {
				if value {
					packed |= 1 << uint(i)
				}
			}
			fmt.Fprintf(&records, " %d", packed)
		}
		records.WriteByte('\n')
	}
	for signs := int32(0); signs < 8; signs++ {
		for keep := int32(0); keep <= 6; keep++ {
			for mode := 0; mode < 2; mode++ {
				src := inputDropReviewGraph(keep, signs, mode == 1)
				if dfs {
					src = Gia_ManDupOrderDfs(src)
				}
				var out *Gia_Man_t
				if mode == 0 {
					out = Gia_ManDupRemovePis(src, 6-keep)
				} else {
					out = Gia_ManDupLastPis(src, keep)
				}
				check(src, out, mode, signs, keep)
			}
		}
	}
	for signs := int32(0); signs < 8; signs++ {
		for _, regs := range []int32{0, 3, 6} {
			src := blockReviewGraph(signs, regs)
			if dfs {
				src = Gia_ManDupOrderDfs(src)
			}
			check(src, Gia_ManDupOnsetOffset(src), 2, signs, regs)
		}
	}
	return records.String()
}

func TestInputRemovalAndDualTypedTruth(t *testing.T) {
	inputDualReviewRecords(t, false)
}

func TestInputRemovalAndDualDfsLayouts(t *testing.T) {
	inputDualReviewRecords(t, true)
}

func TestInputRemovalAndDualMatchOriginalExpandedGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "input_dual.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C input/dual conversion: %v: %s", err, want)
	}
	gotRows := strings.Split(strings.TrimSpace(inputDualReviewRecords(t, false)), "\n")
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

func TestInputRemovalKeepsRegisterPairs(t *testing.T) {
	// C's RemovePis call sites are combinational. The Go entry point also
	// supports dropping trailing PIs without deleting register terminals.
	src := Gia_ManStart(1)
	a := Gia_ManAppendCi(src)
	Gia_ManAppendCi(src)
	r0, r1 := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	src.EnableMuxes()
	x := Gia_ManAppendXorReal(src, a, r0)
	Gia_ManAppendCo(src, Gia_ManAppendMuxReal(src, r1, x, a))
	Gia_ManAppendCo(src, r1)
	Gia_ManAppendCo(src, a^1)
	Gia_ManSetRegNum(src, 2)
	out := Gia_ManDupRemovePis(src, 1)
	if Gia_ManPiNum(out) != 1 || Gia_ManPoNum(out) != 1 || out.nRegs != 2 {
		t.Fatal("dropping a PI changed register pairing")
	}
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		want, err := src.Evaluate([]bool{inputs[0], true, inputs[1], inputs[2]})
		if err != nil {
			t.Fatal(err)
		}
		checkQuantificationOutputs(t, out, inputs, want)
	}
}

func TestInputRemovalAndDualSimpleAnds(t *testing.T) {
	for mode := 0; mode < 3; mode++ {
		src := Gia_ManStart(1)
		src.fGiaSimple = 1
		if mode == 1 {
			Gia_ManAppendCi(src)
		}
		a := Gia_ManAppendCi(src)
		if mode != 1 {
			Gia_ManAppendCi(src)
		}
		Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, a^1))
		var out *Gia_Man_t
		switch mode {
		case 0:
			out = Gia_ManDupRemovePis(src, 1)
		case 1:
			out = Gia_ManDupLastPis(src, 1)
		case 2:
			out = Gia_ManDupOnsetOffset(src)
		}
		if out.fGiaSimple != 1 {
			t.Fatal("input/dual conversion lost simple mode")
		}
		for mask := 0; mask < 1<<uint(Gia_ManCiNum(out)); mask++ {
			inputs := make([]bool, Gia_ManCiNum(out))
			for i := range inputs {
				inputs[i] = mask&(1<<uint(i)) != 0
			}
			want := []bool{false}
			if mode == 2 {
				want = append(want, true)
			}
			checkQuantificationOutputs(t, out, inputs, want)
		}
	}
}

func TestInputRemovalRejectsInvalidOrUsedInputs(t *testing.T) {
	for mode := 0; mode < 2; mode++ {
		for _, count := range []int32{-1, 7} {
			t.Run(fmt.Sprintf("mode%d_count%d", mode, count), func(t *testing.T) {
				defer func() {
					if recover() == nil {
						t.Fatal("invalid input count accepted")
					}
				}()
				src := inputDropReviewGraph(6, 0, false)
				if mode == 0 {
					Gia_ManDupRemovePis(src, count)
				} else {
					Gia_ManDupLastPis(src, count)
				}
			})
		}
		for _, dangling := range []bool{false, true} {
			t.Run(fmt.Sprintf("mode%d_dangling%v", mode, dangling), func(t *testing.T) {
				defer func() {
					if recover() == nil {
						t.Fatal("removed an input still used by source logic")
					}
				}()
				src := Gia_ManStart(1)
				a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
				if dangling {
					Gia_ManAppendAnd(src, a, b)
				} else {
					Gia_ManAppendCo(src, a)
					Gia_ManAppendCo(src, b)
				}
				if mode == 0 {
					Gia_ManDupRemovePis(src, 1)
				} else {
					Gia_ManDupLastPis(src, 1)
				}
			})
		}
	}
	t.Run("last_pis_sequential", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("LastPis accepted a sequential graph")
			}
		}()
		Gia_ManDupLastPis(blockReviewGraph(0, 3), 1)
	})
}
