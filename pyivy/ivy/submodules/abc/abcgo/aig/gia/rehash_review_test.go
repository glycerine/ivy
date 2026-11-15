package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func TestRehashPreservesOutputRoles(t *testing.T) {
	for additional := int32(0); additional < 2; additional++ {
		src := partitionReviewGraph(0, false, false)
		if err := src.SetAigerProperties(2, 2); err != nil {
			t.Fatal(err)
		}
		out := Gia_ManRehash(src, additional)
		if out.aigerBadOutputs != 2 || out.nConstrs != 2 {
			t.Errorf("additional%d: roles bad%d constraints%d, want2/2", additional, out.aigerBadOutputs, out.nConstrs)
		}
	}
}

func TestRehashPreservesInterleavedInterface(t *testing.T) {
	for additional := int32(0); additional < 2; additional++ {
		t.Run(fmt.Sprint(additional), func(t *testing.T) {
			src := partitionReviewGraph(5, false, true)
			out := Gia_ManRehash(src, additional)
			for mask := 0; mask < 32; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
				want, err := src.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				checkQuantificationOutputs(t, out, inputs, want)
			}
		})
	}
}

func TestRehashPreservesRealGateLogic(t *testing.T) {
	for _, kind := range []string{"xor", "mux", "buffer"} {
		for additional := int32(0); additional < 2; additional++ {
			t.Run(fmt.Sprintf("%s_%d", kind, additional), func(t *testing.T) {
				src := Gia_ManStart(1)
				a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
				var root int32
				switch kind {
				case "xor":
					root = Gia_ManAppendXorReal(src, a, b)
				case "mux":
					src.EnableMuxes()
					root = Gia_ManAppendMuxReal(src, c, a, b)
				case "buffer":
					root = Gia_ManAppendBuf(src, a^1)
				}
				Gia_ManAppendCo(src, root)
				out := Gia_ManRehash(src, additional)
				for mask := 0; mask < 8; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
					want, err := src.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					checkQuantificationOutputs(t, out, inputs, want)
				}
			})
		}
	}
}

func rehashReviewGraph(signs, family int, regs int32, typed, layout bool) *Gia_Man_t {
	src := Gia_ManStart(1)
	ci := make([]int32, 6)
	for i := range ci {
		ci[i] = Gia_ManAppendCi(src)
	}
	a, b, c, d, r0, r1 := ci[0], ci[1], ci[2], ci[3], ci[4], ci[5]
	phase := func(bit uint) int32 { return int32(signs>>bit) & 1 }
	var x, y, z int32
	if typed {
		x = Gia_ManAppendXorReal(src, a^phase(0), b^phase(1))
		src.EnableMuxes()
		y = Gia_ManAppendMuxReal(src, r0^phase(2), x, c^phase(2))
	} else {
		x = Gia_ManAppendAnd(src, a^phase(0), b^phase(1))
		y = Gia_ManAppendAnd(src, a^phase(0), c^phase(2))
	}
	z = Gia_ManAppendAnd(src, d^phase(3), r1^phase(4))
	if typed {
		z = Gia_ManAppendBuf(src, z^phase(5))
	}
	side := Gia_ManAppendAnd(src, a^phase(0), b^phase(1)^1)
	var root int32
	switch family {
	case 0:
		root = Gia_ManAppendAnd(src, x, a^phase(0))
	case 1:
		root = Gia_ManAppendAnd(src, x, a^phase(0)^1)
	case 2:
		root = Gia_ManAppendAnd(src, x, y)
	case 3:
		root = Gia_ManAppendAnd(src, x^1, y^1)
	case 4:
		root = Gia_ManAppendAnd(src, x^1, z^1)
	case 5:
		root = Gia_ManAppendAnd(src, x^1, side^1)
	}
	Gia_ManAppendAnd(src, b^phase(1), d^phase(3)) // Deliberately dangling; final cleanup must remove it.
	for _, lit := range []int32{root ^ phase(5), x, y ^ 1, z, 0, 1, root, r0 ^ 1, z ^ 1} {
		Gia_ManAppendCo(src, lit)
	}
	Gia_ManSetRegNum(src, regs)
	src.SetIdentity("rehash", "rehash.aig")
	if layout {
		src = Gia_ManDupOrderDfsReverse(src)
	}
	return src
}

func rehashReviewRecords(t *testing.T, typed, layout bool) string {
	t.Helper()
	var records strings.Builder
	for signs := 0; signs < 64; signs++ {
		for family := 0; family < 6; family++ {
			for _, regs := range []int32{0, 2, 4} {
				for additional := int32(0); additional < 2; additional++ {
					src := rehashReviewGraph(signs, family, regs, typed, layout)
					bad, constr := int32(signs%3), int32((signs/3)%3)
					if err := src.SetAigerProperties(bad, constr); err != nil {
						t.Fatal(err)
					}
					refs := make([]int32, src.nObjs)
					for i := range refs {
						refs[i] = int32(i + 123)
					}
					src.Refs = slices.Clone(refs)
					src.fAddStrash = additional ^ 1
					for id, obj := range src.pObjs {
						obj.Value = uint32(id + 999)
						obj.fMark0 = uint32(id & 1)
						obj.fMark1 = uint32((id >> 1) & 1)
					}
					out := Gia_ManRehash(src, additional)
					if out.nRegs != regs || Gia_ManCiNum(out) != 6 || Gia_ManCoNum(out) != 9 || out.nConstrs != constr || out.aigerBadOutputs != bad {
						t.Fatal("rehash changed complete logical interface/roles")
					}
					if out.nXors != 0 || out.nMuxes != 0 || out.nBufs != 0 || Gia_ManHashStarted(out) != 0 || out.fAddStrash != 0 || Gia_ManIsNormalized(out) == 0 {
						t.Fatal("rehash did not return a normalized ordinary AIG with inactive hashing")
					}
					if !slices.Equal(src.Refs, refs) || src.fAddStrash != additional^1 {
						t.Fatal("rehash changed source reference cache or hash mode")
					}
					for id, obj := range src.pObjs {
						if obj.Value == ^uint32(0) || obj.fMark0 != uint32(id&1) || obj.fMark1 != uint32((id>>1)&1) {
							t.Fatal("source mapping or independent marks lost")
						}
					}
					for i := int32(0); i < 6; i++ {
						if Gia_ManCi(src, i).Value != uint32((i+1)<<1) || Gia_ManCi(out, i).id != i+1 {
							t.Fatal("logical CI/RO order or stable CI copy map lost")
						}
					}
					if name, spec := out.Identity(); name != "rehash" || spec != "rehash.aig" {
						t.Fatal("rehash identity lost")
					}
					parsed := partitionReviewRoundTrip(t, out, bad, constr)
					emit := !typed && !layout
					if emit {
						fmt.Fprintf(&records, "%d %d %d %d result %d %d %d %d values", signs, family, regs, additional, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
						for _, obj := range src.pObjs {
							fmt.Fprintf(&records, " %d", obj.Value)
						}
						fmt.Fprint(&records, " truth")
					}
					for mask := 0; mask < 64; mask++ {
						inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0, mask&32 != 0}
						phase := func(bit uint) bool { return signs&(1<<bit) != 0 }
						a, b, c, d, r0, r1 := inputs[0], inputs[1], inputs[2], inputs[3], inputs[4], inputs[5]
						x := (a != phase(0)) && (b != phase(1))
						y := (a != phase(0)) && (c != phase(2))
						if typed {
							x = (a != phase(0)) != (b != phase(1))
							y = c != phase(2)
							if r0 != phase(2) {
								y = x
							}
						}
						z := (d != phase(3)) && (r1 != phase(4))
						if typed {
							z = z != phase(5)
						}
						side := (a != phase(0)) && !(b != phase(1))
						var root bool
						switch family {
						case 0:
							root = x && (a != phase(0))
						case 1:
							root = x && !(a != phase(0))
						case 2:
							root = x && y
						case 3:
							root = !x && !y
						case 4:
							root = !x && !z
						case 5:
							root = !x && !side
						}
						want := []bool{root != phase(5), x, !y, z, false, true, root, !r0, !z}
						checkQuantificationOutputs(t, out, inputs, want)
						checkQuantificationOutputs(t, parsed, inputs, want)
						if emit {
							packed := 0
							for i, v := range want {
								if v {
									packed |= 1 << uint(i)
								}
							}
							fmt.Fprintf(&records, " %d", packed)
						}
					}
					if emit {
						records.WriteByte('\n')
					}
					// Every retained internal node has a consumer; unused original PIs remain.
					uses := make([]bool, out.nObjs)
					for _, obj := range out.pObjs {
						if obj.id == 0 || Gia_ObjIsCi(obj) != 0 {
							continue
						}
						uses[Gia_ObjFaninId0(obj, obj.id)] = true
						if Gia_ObjIsAnd(obj) != 0 {
							uses[Gia_ObjFaninId1(obj, obj.id)] = true
						}
					}
					for _, obj := range out.pObjs {
						if Gia_ObjIsAnd(obj) != 0 && !uses[obj.id] {
							t.Fatal("rehash left dangling logic")
						}
					}
				}
			}
		}
	}
	return records.String()
}

func TestRehashTypedLayoutAndStrashMatrix(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, layout := range []bool{false, true} {
			rehashReviewRecords(t, typed, layout)
		}
	}
}

func TestRehashMatchesOriginalOrdinaryC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "rehash.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C rehash: %v: %s", err, native)
	}
	got, want := strings.Split(strings.TrimSpace(rehashReviewRecords(t, false, false)), "\n"), strings.Split(strings.TrimSpace(string(native)), "\n")
	if len(got) != len(want) {
		t.Fatalf("Go cases%d C cases%d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("case%d\nGo %s\nC %s", i, got[i], want[i])
		}
	}
}

func TestRehashPreservesActiveSourceHash(t *testing.T) {
	src := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	src.fAddStrash = 1
	Gia_ManHashStart(src)
	root := Gia_ManHashAnd(src, a, b)
	Gia_ManAppendCo(src, root)
	table, chains := slices.Clone(Vec_IntArray(&src.vHTable)), slices.Clone(Vec_IntArray(&src.vHash))
	count := src.nObjs
	out := Gia_ManRehash(src, 0)
	if Gia_ManHashStarted(src) == 0 || src.fAddStrash != 1 || !slices.Equal(Vec_IntArray(&src.vHTable), table) || !slices.Equal(Vec_IntArray(&src.vHash), chains) || Gia_ManHashAnd(src, a, b) != root || src.nObjs != count {
		t.Fatal("rehash disturbed the active source hash")
	}
	for mask := 0; mask < 4; mask++ {
		a, b := mask&1 != 0, mask&2 != 0
		checkQuantificationOutputs(t, out, []bool{a, b}, []bool{a && b})
	}
}

func TestRehashBoundaryGraphs(t *testing.T) {
	for _, kind := range []string{"empty", "constants", "state", "simple_zero", "simple_buffer"} {
		for additional := int32(0); additional < 2; additional++ {
			t.Run(fmt.Sprintf("%s_%d", kind, additional), func(t *testing.T) {
				src := Gia_ManStart(1)
				switch kind {
				case "constants":
					Gia_ManAppendCo(src, 0)
					Gia_ManAppendCo(src, 1)
					if err := src.SetAigerProperties(1, 1); err != nil {
						t.Fatal(err)
					}
				case "state":
					r0, r1 := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
					Gia_ManAppendCo(src, r1)
					Gia_ManAppendCo(src, r0^1)
					Gia_ManSetRegNum(src, 2)
				case "simple_zero", "simple_buffer":
					src.fGiaSimple = 1
					a := Gia_ManAppendCi(src)
					other := a
					if kind == "simple_zero" {
						other ^= 1
					}
					Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, other))
				}
				out := Gia_ManRehash(src, additional)
				if out.nRegs != src.nRegs || Gia_ManCiNum(out) != Gia_ManCiNum(src) || Gia_ManCoNum(out) != Gia_ManCoNum(src) || out.nBufs != 0 || out.fGiaSimple != 0 {
					t.Fatal("rehash boundary interface/simplification")
				}
				parsed := partitionReviewRoundTrip(t, out, src.aigerBadOutputs, src.nConstrs)
				for mask := 0; mask < 1<<uint(Gia_ManCiNum(src)); mask++ {
					inputs := make([]bool, Gia_ManCiNum(src))
					for i := range inputs {
						inputs[i] = mask&(1<<uint(i)) != 0
					}
					want, err := src.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					checkQuantificationOutputs(t, out, inputs, want)
					checkQuantificationOutputs(t, parsed, inputs, want)
				}
			})
		}
	}
}
