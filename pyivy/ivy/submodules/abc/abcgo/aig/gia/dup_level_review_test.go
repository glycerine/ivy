package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func TestLevelizedPreservesConstraintClassification(t *testing.T) {
	src := Gia_ManStart(1)
	a := Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, a^1)
	src.nConstrs = 1
	out := Gia_ManDupLevelized(src)
	if out.nConstrs != 1 {
		t.Fatalf("levelized copy reclassified %d constraint as an ordinary property", src.nConstrs)
	}
}

func TestDfsCopyPreservesConstraintClassification(t *testing.T) {
	src := Gia_ManStart(1)
	a := Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, a^1)
	src.nConstrs = 1
	out := Gia_ManDupOrderDfs(src)
	if out.nConstrs != 1 {
		t.Fatalf("DFS copy reclassified %d constraint as an ordinary property", src.nConstrs)
	}
}

func TestDfsCopyPreservesSimpleAnds(t *testing.T) {
	src := Gia_ManStart(1)
	src.fGiaSimple = 1
	a := Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, a^1))
	out := Gia_ManDupOrderDfs(src)
	if out.fGiaSimple != 1 || out.nObjs != src.nObjs {
		t.Fatal("DFS copy changed simple-mode AND storage")
	}
	for _, input := range []bool{false, true} {
		checkQuantificationOutputs(t, out, []bool{input}, []bool{false})
	}
}

var levelReviewSink *Gia_Man_t

func BenchmarkLevelizedLayerOrdering(b *testing.B) {
	src := Gia_ManStart(1)
	ci := make([]int32, 32)
	for i := range ci {
		ci[i] = Gia_ManAppendCi(src)
	}
	for block := 0; block < 256; block++ {
		root := ci[block%len(ci)]
		for depth := 0; depth < 16; depth++ {
			root = Gia_ManAppendAnd(src, root, ci[(block+depth+1)%len(ci)])
		}
		Gia_ManAppendCo(src, root)
	}
	Gia_ManLevelNum(src)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		levelReviewSink = Gia_ManDupLevelized(src)
	}
}

func levelReviewGraph(signs, registers int32, typed bool) (*Gia_Man_t, []int32) {
	g := Gia_ManStart(1)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	var order []int32
	if !typed {
		x := Gia_ManAppendAnd(g, a^(signs&1), b)
		y := Gia_ManAppendAnd(g, x, c)
		z := Gia_ManAppendAnd(g, y, d)
		Gia_ManAppendCo(g, z^((signs>>1)&1))
		u := Gia_ManAppendAnd(g, a, c^((signs>>2)&1))
		v := Gia_ManAppendAnd(g, b^((signs>>3)&1), d)
		w := Gia_ManAppendAnd(g, u, v)
		for _, lit := range []int32{w, u ^ ((signs >> 4) & 1), d, y, 1} {
			Gia_ManAppendCo(g, lit)
		}
		order = []int32{x, u, v, y, w, z}
	} else {
		x := Gia_ManAppendAnd(g, a, b)
		y := Gia_ManAppendAnd(g, x, c)
		Gia_ManAppendCo(g, y)
		b0 := Gia_ManAppendBuf(g, d^(signs&1))
		b1 := Gia_ManAppendBuf(g, b0^((signs>>1)&1))
		z := Gia_ManAppendXorReal(g, b1, a^((signs>>2)&1))
		b2 := Gia_ManAppendBuf(g, x^((signs>>3)&1))
		g.EnableMuxes()
		mux := Gia_ManAppendMuxReal(g, c^((signs>>4)&1), z, b2)
		root := Gia_ManAppendAnd(g, mux, y)
		for _, lit := range []int32{root, b1 ^ 1, z, b2} {
			Gia_ManAppendCo(g, lit)
		}
		dangling := Gia_ManAppendAnd(g, b, d)
		order = []int32{b0, b1, x, b2, dangling, y, z, mux, root}
	}
	Gia_ManSetRegNum(g, registers)
	g.SetIdentity("levels", "levels.aig")
	return g, order
}

func levelReviewRecords(t *testing.T, typed, dfs bool, constraints int32) string {
	t.Helper()
	var records strings.Builder
	for signs := int32(0); signs < 32; signs++ {
		for _, registers := range []int32{0, 2, 4} {
			src, order := levelReviewGraph(signs, registers, typed)
			src.nConstrs = constraints
			if dfs {
				src = Gia_ManDupOrderDfs(src)
			}
			out := Gia_ManDupLevelized(src)
			name, spec := out.Identity()
			if out.nObjs != src.nObjs || Gia_ManCiNum(out) != 4 || Gia_ManCoNum(out) != Gia_ManCoNum(src) || out.nRegs != registers || out.nConstrs != constraints || name != "levels" || spec != "levels.aig" {
				t.Fatalf("typed=%v dfs=%v signs=%d registers=%d: changed graph/interface/identity", typed, dfs, signs, registers)
			}
			if !dfs {
				for rank, lit := range order {
					if got, want := src.pObjs[lit>>1].Value, uint32((5+rank)<<1); got != want {
						t.Fatalf("gate %d: mapped literal %d, want %d", lit>>1, got, want)
					}
				}
			}
			// Check a bijective, stable level ordering even after DFS has interleaved CIs.
			original := make([]int32, Gia_ManAndNum(src))
			for _, obj := range src.pObjs {
				if Gia_ObjIsAnd(obj) == 0 {
					continue
				}
				mapped := int32(obj.Value >> 1)
				if mapped < 5 || int(mapped-5) >= len(original) || original[mapped-5] != 0 {
					t.Fatal("non-bijective internal node mapping")
				}
				original[mapped-5] = obj.id
				a, b, c, kind := src.Gate(obj.id)
				oa, ob, oc, okind := out.Gate(mapped)
				mapLit := func(lit int32) int32 { return int32(src.pObjs[lit>>1].Value) ^ (lit & 1) }
				// AND/XOR data pins may swap into canonical order during reconstruction.
				if kind != okind {
					t.Fatal("copy changed gate kind")
				}
				if kind == 2 || kind == 3 {
					if !((oa == mapLit(a) && ob == mapLit(b)) || (oa == mapLit(b) && ob == mapLit(a))) {
						t.Fatal("changed AND/XOR pins")
					}
				} else if kind == 5 && oa != mapLit(a) {
					t.Fatal("changed buffer pin")
				} else if kind == 4 {
					if !((oa == mapLit(a) && ob == mapLit(b) && oc == mapLit(c)) || (oa == mapLit(a)^1 && ob == mapLit(c) && oc == mapLit(b))) {
						t.Fatal("changed MUX pins")
					}
				}
			}
			for i, id := range original {
				if id == 0 {
					t.Fatal("missing mapped node")
				}
				if i > 0 {
					prev := original[i-1]
					l, pl := Gia_ObjLevel(src, src.pObjs[id]), Gia_ObjLevel(src, src.pObjs[prev])
					if l < pl || (l == pl && id < prev) {
						t.Fatal("copy is not stable within levels")
					}
				}
			}
			fmt.Fprintf(&records, "%d %d %d %d %d %d", signs, registers, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
			for _, obj := range src.pObjs {
				if Gia_ObjIsAnd(obj) != 0 {
					fmt.Fprintf(&records, " %d", obj.Value)
				}
			}
			for mask := 0; mask < 16; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				a, b, c, d := inputs[0], inputs[1], inputs[2], inputs[3]
				var want []bool
				if !typed {
					x := (a != (signs&1 != 0)) && b
					y := x && c
					u := a && (c != (signs&4 != 0))
					v := (b != (signs&8 != 0)) && d
					want = []bool{(y && d) != (signs&2 != 0), u && v, u != (signs&16 != 0), d, y, true}
				} else {
					x := a && b
					y := x && c
					b1 := (d != (signs&1 != 0)) != (signs&2 != 0)
					z := b1 != (a != (signs&4 != 0))
					b2 := x != (signs&8 != 0)
					mux := b2
					if c != (signs&16 != 0) {
						mux = z
					}
					want = []bool{y, mux && y, !b1, z, b2}
				}
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
	return records.String()
}

func TestLevelizedStableOrderingAndSemantics(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, dfs := range []bool{false, true} {
			for _, constraints := range []int32{0, 1} {
				t.Run(fmt.Sprintf("typed%v_dfs%v_constraints%d", typed, dfs, constraints), func(t *testing.T) { levelReviewRecords(t, typed, dfs, constraints) })
			}
		}
	}
}

func TestLevelizedMatchesOriginalSource(t *testing.T) {
	exe := compileGiaSourceOracle(t, "levelized.c")
	want, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C levelized: %v: %s", err, want)
	}
	got := levelReviewRecords(t, false, false, 0)
	if got != string(want) {
		t.Fatalf("levelized output differs from original C:\nGo %s\nC %s", got, want)
	}
	// These are inherited C omissions; the Go regressions above require repairs.
	witness, err := testoracle.Command(t, exe, "--constraints").CombinedOutput()
	if err != nil || string(witness) != "1 0 0\n" {
		t.Fatalf("constraint classification witness: %v: %s", err, witness)
	}
}

func TestLevelizedEmptyAndZeroLevelBuffers(t *testing.T) {
	for _, mode := range []string{"empty", "wires", "buffers", "simple"} {
		t.Run(mode, func(t *testing.T) {
			src := Gia_ManStart(1)
			if mode != "empty" {
				a := Gia_ManAppendCi(src)
				switch mode {
				case "buffers":
					a = Gia_ManAppendBuf(src, Gia_ManAppendBuf(src, 1))
				case "simple":
					src.fGiaSimple = 1
					a = Gia_ManAppendAnd(src, a, a^1)
				}
				Gia_ManAppendCo(src, a)
				Gia_ManAppendCo(src, 0)
				Gia_ManAppendCo(src, 1)
			}
			out := Gia_ManDupLevelized(src)
			if out.nObjs != src.nObjs || out.fGiaSimple != src.fGiaSimple {
				t.Fatal("changed graph storage/mode")
			}
			rows := 1
			if mode != "empty" {
				rows = 2
			}
			for mask := 0; mask < rows; mask++ {
				var inputs []bool
				if mode != "empty" {
					inputs = []bool{mask != 0}
				}
				want, err := src.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				checkQuantificationOutputs(t, out, inputs, want)
			}
		})
	}
}
