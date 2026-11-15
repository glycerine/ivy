package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func remapReviewLiteralRecords(t *testing.T) string {
	t.Helper()
	options := []uint32{0, 1, 2, 3, 8, 9, ^uint32(0)}
	var records strings.Builder
	for a, av := range options {
		for b, bv := range options {
			for c, cv := range options {
				src := Gia_ManStart(1)
				for i := 0; i < 3; i++ {
					Gia_ManAppendCi(src)
				}
				mappings := []uint32{0, av, bv, cv}
				for id, value := range mappings {
					src.pObjs[id].Value = value
				}
				original := []int32{-1, -7, -1 << 31, 0, 1, 2, 3, 4, 5, 6, 7}
				lits := permutation(original...)
				Gia_ManDupRemapLiterals(lits, src)
				for id, value := range mappings {
					if src.pObjs[id].Value != value {
						t.Fatal("literal remapping changed source Values")
					}
				}
				fmt.Fprintf(&records, "%d %d %d", a, b, c)
				for i, lit := range Vec_IntArray(lits) {
					want := original[i]
					if want >= 0 {
						mapped := mappings[want/2]
						if mapped == ^uint32(0) {
							want = -1
						} else {
							want = int32(mapped) ^ (want % 2)
						}
					}
					if lit != want {
						t.Fatalf("literal %d: got %d, want %d", original[i], lit, want)
					}
					fmt.Fprintf(&records, " %d", lit)
				}
				records.WriteByte('\n')
			}
		}
	}
	return records.String()
}

func remapReviewPermutations() [][]int32 {
	var all [][]int32
	var visit func([]int32)
	visit = func(prefix []int32) {
		if len(prefix) == 4 {
			all = append(all, slices.Clone(prefix))
			return
		}
		for i := int32(0); i < 4; i++ {
			if !slices.Contains(prefix, i) {
				visit(append(slices.Clone(prefix), i))
			}
		}
	}
	visit(nil)
	return all
}

func remapReviewTerminalRecords(t *testing.T, dfs bool) string {
	t.Helper()
	var records strings.Builder
	for pi, inputs := range remapReviewPermutations() {
		for po, outputs := range remapReviewPermutations() {
			for _, regs := range []int32{0, 2, 4} {
				src := Gia_ManStart(1)
				a, b, c, d := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
				x := Gia_ManAppendAnd(src, a, b)
				for _, lit := range []int32{a ^ 1, x, c, d} {
					Gia_ManAppendCo(src, lit)
				}
				Gia_ManSetRegNum(src, regs)
				if dfs {
					src = Gia_ManDupOrderDfsReverse(src)
				}
				dst := Gia_ManStart(1)
				physical := make([]int32, 4)
				for i := range physical {
					physical[i] = Gia_ManAppendCi(dst)
				}
				src.pObjs[0].Value = 0
				for i, index := range inputs {
					Gia_ManCi(src, int32(i)).Value = uint32(physical[index])
				}
				giaDupMappedNodes(dst, src)
				for position := int32(0); position < 4; position++ {
					index := slices.Index(outputs, position)
					obj := Gia_ManCo(src, int32(index))
					obj.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj)))
				}
				Gia_ManSetRegNum(dst, regs)
				dst.SetIdentity("destination", "original.aig")
				dst.nConstrs = 1
				if regs == 4 {
					dst.nConstrs = 0
				}
				before := make([]uint32, src.nObjs)
				for id, obj := range src.pObjs {
					before[id] = obj.Value
				}
				Gia_ManDupRemapCis(dst, src)
				Gia_ManDupRemapCos(dst, src)
				name, spec := dst.Identity()
				if dst.nRegs != regs || dst.nConstrs != giaBool(regs != 4) || name != "destination" || spec != "original.aig" {
					t.Fatal("terminal remapping changed unrelated metadata")
				}
				for id, obj := range src.pObjs {
					if obj.Value != before[id] {
						t.Fatal("terminal remapping changed source Values")
					}
				}
				fmt.Fprintf(&records, "%d %d %d", pi, po, regs)
				for i, index := range inputs {
					obj := Gia_ManCi(dst, int32(i))
					if obj.id != physical[index]>>1 || Gia_ObjCioId(obj) != int32(i) {
						t.Fatal("CI permutation or ordinal incorrect")
					}
					fmt.Fprintf(&records, " %d:%d", obj.id, Gia_ObjCioId(obj))
				}
				for i, index := range outputs {
					obj := Gia_ManCo(dst, int32(i))
					if obj.id != 6+index || Gia_ObjCioId(obj) != int32(i) {
						t.Fatal("CO permutation or ordinal incorrect")
					}
					fmt.Fprintf(&records, " %d:%d", obj.id, Gia_ObjCioId(obj))
				}
				for mask := 0; mask < 16; mask++ {
					in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
					want := []bool{!in[0], in[0] && in[1], in[2], in[3]}
					checkQuantificationOutputs(t, dst, in, want)
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

func remapReviewEquivRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	for rotation := int32(0); rotation < 8; rotation++ {
		for reverse := 0; reverse < 2; reverse++ {
			for phase := 0; phase < 4; phase++ {
				for drop := 0; drop < 2; drop++ {
					src, dst := Gia_ManStart(1), Gia_ManStart(1)
					for i := 0; i < 8; i++ {
						Gia_ManAppendCi(src)
						Gia_ManAppendCi(dst)
					}
					src.StartRepresentatives()
					Gia_ObjSetRepr(src, 3, 1)
					Gia_ObjSetRepr(src, 5, 1)
					Gia_ObjSetRepr(src, 4, 2)
					Gia_ObjSetRepr(src, 6, 0)
					Gia_ObjSetRepr(src, 7, 0)
					src.PNexts = Gia_ManDeriveNexts(src)
					src.pObjs[0].Value = 0
					mapped := make([]int32, 9)
					for id := int32(1); id <= 8; id++ {
						index := (id - 1 + rotation) % 8
						if reverse != 0 {
							index = 7 - index
						}
						mapped[id] = index + 1
						polarity := int32(0)
						switch phase {
						case 1:
							polarity = 1
						case 2:
							polarity = id & 1
						case 3:
							polarity = (id >> 1) & 1
						}
						src.pObjs[id].Value = uint32((mapped[id] << 1) | polarity)
					}
					if drop != 0 {
						src.pObjs[8].Value = ^uint32(0)
					} // Unclassified and never visited by C's class walk.
					Gia_ManDupRemapEquiv(dst, src)
					expected := make([]int32, 9)
					next := make([]int32, 9)
					for i := range expected {
						expected[i] = GIA_VOID
					}
					for _, class := range [][]int32{{1, 3, 5}, {2, 4}} {
						ids := make([]int32, len(class))
						for i, id := range class {
							ids[i] = mapped[id]
						}
						slices.Sort(ids)
						for i := 1; i < len(ids); i++ {
							expected[ids[i]] = ids[0]
							next[ids[i-1]] = ids[i]
						}
					}
					expected[mapped[6]], expected[mapped[7]] = 0, 0
					fmt.Fprintf(&records, "%d %d %d %d", rotation, reverse, phase, drop)
					for id := int32(0); id < dst.nObjs; id++ {
						if Gia_ObjRepr(dst, id) != expected[id] || Gia_ObjNext(dst, id) != next[id] {
							t.Fatal("changed representative class, constant membership or next chain")
						}
						fmt.Fprintf(&records, " %d:%d", Gia_ObjRepr(dst, id), Gia_ObjNext(dst, id))
					}
					records.WriteByte('\n')
					Gia_ObjUnsetRepr(dst, mapped[3])
					if Gia_ObjRepr(src, 3) != 1 {
						t.Fatal("representatives alias source")
					}
				}
			}
		}
	}
	return records.String()
}

func TestDuplicationRemapPhasesTerminalsAndClasses(t *testing.T) {
	remapReviewLiteralRecords(t)
	remapReviewTerminalRecords(t, false)
	remapReviewTerminalRecords(t, true)
	remapReviewEquivRecords(t)
}

func TestDuplicationRemapsMatchOriginalSource(t *testing.T) {
	exe := compileGiaSourceOracle(t, "duplication_remapping.c")
	for _, mode := range []string{"literals", "terminals", "equiv"} {
		t.Run(mode, func(t *testing.T) {
			native, err := exec.Command(exe, mode).CombinedOutput()
			if err != nil {
				t.Fatalf("C remapping: %v: %s", err, native)
			}
			var records string
			switch mode {
			case "literals":
				records = remapReviewLiteralRecords(t)
			case "terminals":
				records = remapReviewTerminalRecords(t, false)
			case "equiv":
				records = remapReviewEquivRecords(t)
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
}

func TestRemapBoundaryAndRejectedMappings(t *testing.T) {
	src, dst := Gia_ManStart(1), Gia_ManStart(1)
	empty := permutation()
	Gia_ManDupRemapLiterals(empty, src)
	Gia_ManDupRemapCis(dst, src)
	Gia_ManDupRemapCos(dst, src)
	if Vec_IntSize(empty) != 0 || dst.nObjs != 1 {
		t.Fatal("empty remapping changed graph/vector")
	}
	for _, literal := range []int32{2, 3, 1<<31 - 1} {
		t.Run(fmt.Sprintf("invalidLiteral%d", literal), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("out of range literal accepted")
				}
			}()
			Gia_ManDupRemapLiterals(permutation(literal), Gia_ManStart(1))
		})
	}
	for _, outputs := range []bool{false, true} {
		for _, mode := range []string{"count", "unset", "complemented", "range", "nonterminal"} {
			t.Run(fmt.Sprintf("outputs%v_%s", outputs, mode), func(t *testing.T) {
				src, dst := Gia_ManStart(1), Gia_ManStart(1)
				var source, target *Gia_Obj_t
				if outputs {
					source = Gia_ManObj(src, Gia_ManAppendCo(src, 1)>>1)
					if mode != "count" {
						target = Gia_ManObj(dst, Gia_ManAppendCo(dst, 0)>>1)
					}
				} else {
					source = Gia_ManObj(src, Gia_ManAppendCi(src)>>1)
					if mode != "count" {
						target = Gia_ManObj(dst, Gia_ManAppendCi(dst)>>1)
					}
				}
				if target != nil {
					source.Value = uint32(target.id << 1)
				}
				switch mode {
				case "unset":
					source.Value = ^uint32(0)
				case "complemented":
					source.Value |= 1
				case "range":
					source.Value = uint32(dst.nObjs << 1)
				case "nonterminal":
					source.Value = 0
				}
				defer func() {
					if recover() == nil {
						t.Fatal("invalid terminal mapping accepted")
					}
				}()
				if outputs {
					Gia_ManDupRemapCos(dst, src)
				} else {
					Gia_ManDupRemapCis(dst, src)
				}
			})
		}
	}
	// Absence of source representatives is a no-op even with destination classes.
	src, dst = Gia_ManStart(1), Gia_ManStart(1)
	Gia_ManAppendCi(dst)
	Gia_ManAppendCi(dst)
	dst.StartRepresentatives()
	Gia_ObjSetRepr(dst, 2, 1)
	dst.PNexts = Gia_ManDeriveNexts(dst)
	Gia_ManDupRemapEquiv(dst, src)
	if Gia_ObjRepr(dst, 2) != 1 || Gia_ObjNext(dst, 1) != 2 {
		t.Fatal("absent source metadata cleared destination")
	}
	for _, mode := range []string{"reprs", "nexts", "bad map", "bad source"} {
		t.Run(mode, func(t *testing.T) {
			src, dst := Gia_ManStart(1), Gia_ManStart(1)
			Gia_ManAppendCi(src)
			Gia_ManAppendCi(dst)
			src.StartRepresentatives()
			src.pObjs[0].Value = 0
			src.pObjs[1].Value = 2
			switch mode {
			case "reprs":
				dst.StartRepresentatives()
			case "nexts":
				dst.PNexts = make([]int32, dst.nObjs)
			case "bad map":
				src.pObjs[1].Value = uint32(dst.nObjs << 1)
			case "bad source":
				src.PReprs = src.PReprs[:1]
			}
			defer func() {
				if recover() == nil {
					t.Fatal("invalid representative remap accepted")
				}
			}()
			Gia_ManDupRemapEquiv(dst, src)
		})
	}
}

func TestRemapRepresentativesAfterDroppingOrCollapsing(t *testing.T) {
	// TransferRepresentativesFrom explicitly supports these cases in Go. C's
	// class loop assumes every class still contains at least two mapped nodes;
	// do not invoke its assertion/invalid indexing for the dropped-node cases.
	cases := []struct {
		name                string
		mapping, repr, next []int32
	}{
		{"dropped head", []int32{0, -1, 1, 2}, []int32{GIA_VOID, GIA_VOID, 1}, []int32{0, 2, 0}},
		{"singleton", []int32{0, 1, -1, -1}, []int32{GIA_VOID, GIA_VOID, GIA_VOID}, []int32{0, 0, 0}},
		{"collapsed members", []int32{0, 1, 1, 2}, []int32{GIA_VOID, GIA_VOID, 1}, []int32{0, 2, 0}},
		{"collapsed class", []int32{0, 1, 1, 1}, []int32{GIA_VOID, GIA_VOID, GIA_VOID}, []int32{0, 0, 0}},
		{"constant class", []int32{0, 0, 0, 2}, []int32{GIA_VOID, GIA_VOID, 0}, []int32{0, 0, 0}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src, dst := Gia_ManStart(1), Gia_ManStart(1)
			for i := 0; i < 3; i++ {
				Gia_ManAppendCi(src)
			}
			Gia_ManAppendCi(dst)
			Gia_ManAppendCi(dst)
			src.StartRepresentatives()
			Gia_ObjSetRepr(src, 2, 1)
			Gia_ObjSetRepr(src, 3, 1)
			src.PNexts = Gia_ManDeriveNexts(src)
			for id, mapped := range tc.mapping {
				if mapped < 0 {
					src.pObjs[id].Value = ^uint32(0)
				} else {
					src.pObjs[id].Value = uint32(mapped << 1)
				}
			}
			Gia_ManDupRemapEquiv(dst, src)
			for id := int32(0); id < dst.nObjs; id++ {
				if Gia_ObjRepr(dst, id) != tc.repr[id] || Gia_ObjNext(dst, id) != tc.next[id] {
					t.Fatal("wrong surviving class or constant association")
				}
			}
		})
	}
}
