package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"os/exec"
	"strings"
	"testing"
)

func TestDfsVariantsPreserveSimpleMode(t *testing.T) {
	cases := []struct {
		name string
		copy func(*Gia_Man_t) *Gia_Man_t
	}{
		{"dfs", Gia_ManDupDfs}, {"dfs2", Gia_ManDupDfs2},
		{"one", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupDfsOnePo(g, 0) }},
		{"skip", Gia_ManDupDfsSkip},
		{"node", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupDfsNode(g, Gia_ObjFanin0(Gia_ManCo(g, 0))) }},
		{"literals", func(g *Gia_Man_t) *Gia_Man_t {
			return Gia_ManDupDfsLitArray(g, permutation(Gia_ObjFanin0(Gia_ManCo(g, 0)).id<<1))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Errorf("valid simple-mode copy panicked: %v", err)
				}
			}()
			src := Gia_ManStart(1)
			src.fGiaSimple = 1
			a := Gia_ManAppendCi(src)
			Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, a^1))
			out := tc.copy(src)
			if out.fGiaSimple != 1 {
				t.Fatal("lost simple construction mode")
			}
			for _, input := range []bool{false, true} {
				checkQuantificationOutputs(t, out, []bool{input}, []bool{false})
			}
		})
	}
}

func TestDfs2PreservesConstraintClassification(t *testing.T) {
	src := Gia_ManStart(1)
	a, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, a^1)
	Gia_ManAppendCo(src, r)
	Gia_ManSetRegNum(src, 1)
	src.nConstrs = 1
	if out := Gia_ManDupDfs2(src); out.nConstrs != 1 {
		t.Fatal("full DFS2 copy reclassified a constraint as a property")
	}
}

func dfsReviewFullRecords(t *testing.T, typed, layout bool, constraints int32) string {
	t.Helper()
	var records strings.Builder
	for mode, copy := range []func(*Gia_Man_t) *Gia_Man_t{Gia_ManDupDfs, Gia_ManDupDfs2} {
		for signs := int32(0); signs < 32; signs++ {
			for _, regs := range []int32{0, 2, 4} {
				src, _ := levelReviewGraph(signs, 0, typed)
				Gia_ManAppendCi(src)
				Gia_ManSetRegNum(src, regs)
				src.nConstrs = constraints
				if layout {
					src = Gia_ManDupOrderDfsReverse(src)
				}
				out := copy(src)
				name, spec := out.Identity()
				wantNodes := src.nObjs
				if typed && !layout {
					wantNodes--
				}
				if out.nObjs != wantNodes || Gia_ManCiNum(out) != 5 || Gia_ManCoNum(out) != Gia_ManCoNum(src) || out.nRegs != regs || out.nConstrs != constraints || name != "levels" || spec != "levels.aig" {
					t.Fatal("DFS copy changed interface, live logic or identity")
				}
				if mode == 0 && Gia_ManIsNormalized(out) == 0 {
					t.Fatal("CIs-first DFS result not normalized")
				}
				for i := int32(0); i < 5; i++ {
					if Gia_ManCi(src, i).Value != uint32(Gia_ManCi(out, i).id<<1) {
						t.Fatal("DFS input order changed")
					}
				}
				for i := int32(0); i < Gia_ManCoNum(src); i++ {
					if Gia_ManCo(src, i).Value != uint32(Gia_ManCo(out, i).id<<1) {
						t.Fatal("DFS output order changed")
					}
				}
				fmt.Fprintf(&records, "%d %d %d %d %d %d %d", mode, signs, regs, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
				for _, obj := range src.pObjs {
					fmt.Fprintf(&records, " %d", obj.Value)
				}
				for mask := 0; mask < 32; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
					want, _ := selectionReviewTruth(signs, typed, inputs)
					checkQuantificationOutputs(t, out, inputs, want)
					packed := 0
					for i, value := range want {
						if value {
							packed |= 1 << uint(i)
						}
					}
					fmt.Fprintf(&records, " %d", packed)
				}
				records.WriteByte('\n')
			}
		}
	}
	return records.String()
}

func dfsReviewRepresentativeGraph(signs, regs int32) (*Gia_Man_t, int32) {
	src := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCi(src)
	Gia_ManAppendCi(src)
	x := Gia_ManAppendAnd(src, a^(signs&1), b^((signs>>1)&1))
	y := Gia_ManAppendAnd(src, x^1, 1)
	z := Gia_ManAppendAnd(src, y, 1)
	u := Gia_ManAppendAnd(src, (a^(signs&1))^1, b^((signs>>1)&1))
	zero := Gia_ManAppendAnd(src, x, u)
	one := Gia_ManAppendAnd(src, zero^1, 1)
	for _, lit := range []int32{x ^ ((signs >> 2) & 1), y ^ ((signs >> 3) & 1), z ^ ((signs >> 4) & 1), zero, one} {
		Gia_ManAppendCo(src, lit)
	}
	Gia_ManSetRegNum(src, regs)
	Gia_ManSetPhase(src)
	src.pReprsOld = make([]int32, src.nObjs)
	for i := range src.pReprsOld {
		src.pReprsOld[i] = -1
	}
	src.pReprsOld[y>>1] = x >> 1
	src.pReprsOld[z>>1] = y >> 1
	src.pReprsOld[zero>>1] = 0
	src.pReprsOld[one>>1] = 0
	return src, u >> 1
}

func dfsReviewRepresentativeRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	for mode := 0; mode < 2; mode++ {
		for signs := int32(0); signs < 32; signs++ {
			for _, regs := range []int32{0, 2, 4} {
				src, unused := dfsReviewRepresentativeGraph(signs, regs)
				var out *Gia_Man_t
				if mode == 0 {
					out = Gia_ManDupDfs2(src)
				} else {
					out = giaDupMappedStart(src)
					for i := int32(0); i < Gia_ManCiNum(src); i++ {
						Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
					}
					Gia_ManHashAlloc(out)
					for i := int32(0); i < Gia_ManCoNum(src); i++ {
						Gia_ManDupDfs2_rec(out, src, Gia_ManCo(src, i))
					}
					Gia_ManHashStop(out)
					Gia_ManSetRegNum(out, regs)
				}
				count := out.nObjs
				first := Gia_ManCo(src, 0)
				if lit := Gia_ManDupDfs2_rec(out, src, first); lit != int32(first.Value) || out.nObjs != count {
					t.Fatal("memoized DFS allocated or changed its result")
				}

				if Gia_ManCiNum(out) != 4 || Gia_ManCoNum(out) != 5 || out.nRegs != regs || Gia_ManAndNum(out) != 1 || src.pObjs[unused].Value != ^uint32(0) || Gia_ManHashStarted(out) != 0 {
					t.Fatal("representative substitution retained redundant logic or lost interface")
				}
				fmt.Fprintf(&records, "%d %d %d %d", mode, signs, regs, out.nObjs)
				for _, obj := range src.pObjs {
					fmt.Fprintf(&records, " %d", obj.Value)
				}
				for mask := 0; mask < 16; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
					x := (inputs[0] != (signs&1 != 0)) && (inputs[1] != (signs&2 != 0))
					want := []bool{x != (signs&4 != 0), (!x) != (signs&8 != 0), (!x) != (signs&16 != 0), false, true}
					checkQuantificationOutputs(t, src, inputs, want)
					checkQuantificationOutputs(t, out, inputs, want)
					packed := 0
					for i, value := range want {
						if value {
							packed |= 1 << uint(i)
						}
					}
					fmt.Fprintf(&records, " %d", packed)
				}
				records.WriteByte('\n')
			}
		}
	}
	return records.String()
}

func TestDfsFullCopiesAndRepresentativeSubstitution(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, layout := range []bool{false, true} {
			for _, constraints := range []int32{0, 1} {
				t.Run(fmt.Sprintf("typed%v_layout%v_constraints%d", typed, layout, constraints), func(t *testing.T) { dfsReviewFullRecords(t, typed, layout, constraints) })
			}
		}
	}
	dfsReviewRepresentativeRecords(t)
}

func TestDfsFullAndRepresentativeCopiesMatchSource(t *testing.T) {
	exe := compileGiaSourceOracle(t, "dfs_representatives.c")
	for _, mode := range []string{"full", "representatives"} {
		t.Run(mode, func(t *testing.T) {
			native, err := exec.Command(exe, mode).CombinedOutput()
			if err != nil {
				t.Fatalf("C DFS: %v: %s", err, native)
			}
			var records string
			if mode == "full" {
				records = dfsReviewFullRecords(t, false, false, 0)
			} else {
				records = dfsReviewRepresentativeRecords(t)
			}
			got, want := strings.Split(strings.TrimSpace(records), "\n"), strings.Split(strings.TrimSpace(string(native)), "\n")
			if len(got) != len(want) {
				t.Fatalf("Go cases %d, C cases %d", len(got), len(want))
			}
			for i := range got {
				if got[i] != want[i] {
					t.Fatalf("case %d:\nGo %s\nC %s", i, got[i], want[i])
				}
			}
		})
	}
	witness, err := exec.Command(exe, "constraints").CombinedOutput()
	if err != nil || string(witness) != "1 1 0\n" {
		t.Fatalf("inherited DFS2 constraint witness: %v: %s", err, witness)
	}
}

func TestDfsCounterexampleOwnership(t *testing.T) {
	for _, regs := range []int32{0, 2, 4} {
		src, _ := levelReviewGraph(0, 0, false)
		Gia_ManAppendCi(src)
		Gia_ManSetRegNum(src, regs)
		src.pCexSeq = util.Abc_CexAlloc(regs, 5-regs, 37)
		src.pCexSeq.IFrame = 36
		for i := regs; i < src.pCexSeq.NBits; i++ {
			if (i-regs)%5 < 2 {
				util.Abc_InfoSetBit(src.pCexSeq.PData, i)
			}
		}
		out := Gia_ManDupDfs(src)
		if out.pCexSeq == nil || out.pCexSeq == src.pCexSeq || out.pCexSeq.NRegs != regs || out.pCexSeq.NPis != 5-regs || out.pCexSeq.NBits != src.pCexSeq.NBits || out.pCexSeq.IFrame != 36 || out.pCexSeq.IPo != 0 {
			t.Fatal("counterexample copy changed dimensions or aliased source")
		}
		for i := int32(0); i < src.pCexSeq.NBits; i++ {
			if util.Abc_InfoHasBit(out.pCexSeq.PData, i) != util.Abc_InfoHasBit(src.pCexSeq.PData, i) {
				t.Fatal("counterexample copy changed valid zero-init trace bits")
			}
		}
		out.pCexSeq.PData[0] ^= 1
		if out.pCexSeq.PData[0] == src.pCexSeq.PData[0] {
			t.Fatal("counterexample data aliases source")
		}
		if copy := Gia_ManDupDfs2(src); copy.pCexSeq != nil {
			t.Fatal("DFS2 copied counterexample contrary to C's metadata policy")
		}
	}
	src := Gia_ManStart(1)
	src.pCexSeq = util.AbcCexSentinel
	if out := Gia_ManDupDfs(src); out.pCexSeq != util.AbcCexSentinel {
		t.Fatal("counterexample sentinel changed")
	}
}

func TestDfsFullCopyBoundariesAndInvalidRepresentatives(t *testing.T) {
	for mode, copy := range []func(*Gia_Man_t) *Gia_Man_t{Gia_ManDupDfs, Gia_ManDupDfs2} {
		for _, kind := range []string{"empty", "constants", "unused-input", "register-only"} {
			t.Run(fmt.Sprintf("copy%d_%s", mode, kind), func(t *testing.T) {
				src := Gia_ManStart(1)
				var inputs, want []bool
				switch kind {
				case "constants":
					Gia_ManAppendCo(src, 0)
					Gia_ManAppendCo(src, 1)
					want = []bool{false, true}
				case "unused-input":
					Gia_ManAppendCi(src)
					inputs = []bool{true}
				case "register-only":
					r := Gia_ManAppendCi(src)
					Gia_ManAppendCo(src, r^1)
					Gia_ManSetRegNum(src, 1)
					inputs = []bool{true}
					want = []bool{false}
				}
				out := copy(src)
				if out.nObjs != src.nObjs || Gia_ManCiNum(out) != Gia_ManCiNum(src) || Gia_ManCoNum(out) != Gia_ManCoNum(src) || out.nRegs != src.nRegs {
					t.Fatal("boundary copy changed graph/interface")
				}
				checkQuantificationOutputs(t, out, inputs, want)
			})
		}
	}
	for _, kind := range []string{"short", "self-cycle", "out-of-range"} {
		t.Run(kind, func(t *testing.T) {
			src := Gia_ManStart(1)
			a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			x := Gia_ManAppendAnd(src, a, b)
			Gia_ManAppendCo(src, x)
			src.pReprsOld = make([]int32, src.nObjs)
			for i := range src.pReprsOld {
				src.pReprsOld[i] = -1
			}
			switch kind {
			case "short":
				src.pReprsOld = src.pReprsOld[:1]
			case "self-cycle":
				src.pReprsOld[x>>1] = x >> 1
			case "out-of-range":
				src.pReprsOld[x>>1] = src.nObjs
			}
			defer func() {
				if recover() == nil {
					t.Fatal("invalid representative metadata accepted")
				}
			}()
			Gia_ManDupDfs2(src)
		})
	}
}
