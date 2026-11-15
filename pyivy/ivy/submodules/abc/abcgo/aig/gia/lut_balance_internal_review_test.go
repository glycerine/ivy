package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// Check the actual C helpers' intermediate decisions, not merely equivalent
// output truth. The native fixture includes giaBalLut.c to expose static APIs.
func TestLutBalanceInternalNativeOracle(t *testing.T) {
	exe := compileGiaSourceOracle(t, "lut_balance_internal.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v\n%s", err, native)
	}
	var got strings.Builder
	lutInternalGraphSnapshots(&got)
	lutCollectorBoundarySnapshots(t, &got)
	lutSimplificationSnapshots(t, &got)
	lutMergeSnapshots(&got)
	lutMuxMergeSnapshots(&got)

	if got.String() != string(native) {
		wantLines, gotLines := strings.Split(string(native), "\n"), strings.Split(got.String(), "\n")
		for i := 0; i < len(wantLines) && i < len(gotLines); i++ {
			if wantLines[i] != gotLines[i] {
				t.Fatalf("line %d\nC:  %s\nGo: %s", i+1, wantLines[i], gotLines[i])
			}
		}
		t.Fatalf("line counts C %d Go %d", len(wantLines), len(gotLines))
	}
	t.Log("560 configurations, 26,880 gate/cut and pair decisions, plus all collected supergates, boundary cases, simplifications and signature-collision merges match C")
}

func lutInternalGraphSnapshots(got *strings.Builder) {
	for size := int32(2); size <= 6; size++ {
		for cuts := int32(2); cuts <= 8; cuts++ {
			for seed := uint32(1); seed <= 16; seed++ {
				g := Gia_ManStart(1)
				pool := []int32{0, 1}
				for i := 0; i < 6; i++ {
					pool = append(pool, Gia_ManAppendCi(g))
				}
				g.EnableMuxes()
				p := Bal_ManAlloc(g, g, size, cuts, 0)
				Gia_ManHashStart(g)
				state := seed
				random := func() uint32 { state = state*1664525 + 1013904223; return state }
				for step := 0; step < 48; step++ {
					a := pool[int(random())%len(pool)] ^ int32(random()&1)
					b := pool[int(random())%len(pool)] ^ int32(random()&1)
					c := pool[int(random())%len(pool)] ^ int32(random()&1)
					kinds := 3
					if size == 2 {
						kinds = 2
					}
					var lit int32
					switch step % kinds {
					case 0:
						lit = Gia_ManHashAnd(g, a, b)
					case 1:
						lit = Gia_ManHashXorReal(g, a, b)
					case 2:
						lit = Gia_ManHashMuxReal(g, c, a, b)
					}
					cost := Bal_ManSetGateLevel(p, nil, lit)
					pool = append(pool, lit)
					fmt.Fprintf(got, "gate %d %d %d %d %d %d %d", size, cuts, seed, step, lit, cost, g.nObjs)
					for _, cut := range p.Cuts[lit>>1] {
						fmt.Fprintf(got, " cut %d %d %d", cut.Delay, cut.Func, cut.Sign)
						for _, leaf := range cut.Leaves {
							fmt.Fprintf(got, " %d", leaf)
						}
					}
					got.WriteByte('\n')
					lits := append([]int32(nil), pool[len(pool)-6:]...)
					balLutSortCosts(p, lits)
					i, k := Bal_ManFindBestPair(p, lits, 0)
					fmt.Fprintf(got, "pair %d %d", i, k)
					for _, lit := range lits {
						fmt.Fprintf(got, " %d", lit)
					}
					got.WriteByte('\n')
				}
				Gia_ManHashStop(g)
				Gia_ManCreateRefs(g)
				for id := int32(1); id < g.nObjs; id++ {
					_, _, _, kind := g.Gate(id)
					if kind != 2 && kind != 3 {
						continue
					}
					fmt.Fprintf(got, "super %d", id)
					for _, lit := range giaLutCollect(g, id) {
						fmt.Fprintf(got, " %d", lit)
					}
					got.WriteByte('\n')
				}
				Bal_ManFree(p)
				Gia_ManStop(g)
			}
		}
	}
}

func lutCollectorBoundarySnapshots(t *testing.T, got *strings.Builder) {
	t.Helper()
	for mode := 0; mode < 2; mode++ {
		for extra := 0; extra < 4; extra++ {
			g := Gia_ManStart(1)
			g.EnableMuxes()
			inputs := make([]int32, 64)
			for i := range inputs {
				inputs[i] = Gia_ManAppendCi(g)
			}
			appendGate := Gia_ManAppendAnd
			if mode != 0 {
				appendGate = Gia_ManAppendXorReal
			}
			base := appendGate(g, inputs[0], inputs[1])
			root := base
			for _, lit := range inputs[2:] {
				root = appendGate(g, root, lit)
			}
			Gia_ManAppendCo(g, root)
			for i := 0; i < extra; i++ {
				Gia_ManAppendCo(g, base)
			}
			Gia_ManCreateRefs(g)
			for j, lit := range []int32{root, base} {
				lits := giaLutCollect(g, lit>>1)
				fmt.Fprintf(got, "boundary %d %d %d", mode, extra, j)
				for _, leaf := range lits {
					fmt.Fprintf(got, " %d", leaf)
				}
				got.WriteByte('\n')
				// Independently evaluate the cut expression against the source node.
				state := uint64(77)
				for row := 0; row < 128; row++ {
					bits := make([]bool, 64)
					for i := range bits {
						state = state*6364136223846793005 + 1442695040888963407
						bits[i] = state>>63 != 0
					}
					values, _, err := g.evaluateObjects(bits)
					if err != nil {
						t.Fatal(err)
					}
					value := mode == 0
					for _, leaf := range lits {
						x := values[leaf>>1]
						if leaf&1 != 0 {
							x = !x
						}
						if mode == 0 {
							value = value && x
						} else {
							value = value != x
						}
					}
					if value != values[lit>>1] {
						t.Fatalf("collector mode%d extra%d root%d row%d", mode, extra, j, row)
					}
				}
			}
			Gia_ManStop(g)
		}
	}

}

func lutSimplificationSnapshots(t *testing.T, got *strings.Builder) {
	t.Helper()
	// Exhaust all multisets of up to eight literals of three variables,
	// including constants and both polarities. Reverse input tests sorting.
	var simplify func([]int32, int32)
	simplify = func(values []int32, minimum int32) {
		for mode := 0; mode < 2; mode++ {
			fmt.Fprintf(got, "simplify %d", mode)
			for _, lit := range values {
				fmt.Fprintf(got, " %d", lit)
			}
			got.WriteString(" =>")
			input := append([]int32(nil), values...)
			for i, j := 0, len(input)-1; i < j; i, j = i+1, j-1 {
				input[i], input[j] = input[j], input[i]
			}
			out := giaLutSimplify(input, mode == 1)
			for _, lit := range out {
				fmt.Fprintf(got, " %d", lit)
			}
			got.WriteByte('\n')
			for mask := 0; mask < 8; mask++ {
				eval := func(lits []int32) bool {
					result := mode == 0
					for _, lit := range lits {
						value := lit>>1 != 0 && mask&(1<<((lit>>1)-1)) != 0
						if lit&1 != 0 {
							value = !value
						}
						if mode == 0 {
							result = result && value
						} else {
							result = result != value
						}
					}
					return result
				}
				if eval(values) != eval(out) {
					t.Fatalf("mode%d input%v simplified%v mask%d", mode, values, out, mask)
				}
			}
		}
		if len(values) == 8 {
			return
		}
		for lit := minimum; lit < 8; lit++ {
			simplify(append(values, lit), lit)
		}
	}
	simplify(nil, 0)
}

func lutMergeSnapshots(got *strings.Builder) {
	leaves := []int32{1, 2, 65, 66, 129, 130}
	for mask0 := 1; mask0 < 64; mask0++ {
		for mask1 := 1; mask1 < 64; mask1++ {
			for size := int32(2); size <= 6; size++ {
				a, b := Bal_Cut_t{Delay: 2}, Bal_Cut_t{Delay: 3}
				for i, leaf := range leaves {
					if mask0>>i&1 != 0 {
						a.Leaves = append(a.Leaves, leaf)
					}
					if mask1>>i&1 != 0 {
						b.Leaves = append(b.Leaves, leaf)
					}
				}
				if len(a.Leaves) > int(size) || len(b.Leaves) > int(size) {
					continue
				}
				a.Sign, b.Sign = Bal_CutGetSign(a.Leaves), Bal_CutGetSign(b.Leaves)
				var out Bal_Cut_t
				ok := Bal_CutMergeOrder(&a, &b, &out, size)
				fmt.Fprintf(got, "merge %d %d %d %d", mask0, mask1, size, giaBool(ok))
				if ok {
					fmt.Fprintf(got, " %d %d %d", out.Delay, out.Func, out.Sign)
					for _, leaf := range out.Leaves {
						fmt.Fprintf(got, " %d", leaf)
					}
				}
				got.WriteByte('\n')
			}
		}
	}
}

func lutMuxMergeSnapshots(got *strings.Builder) {
	leaves := []int32{1, 2, 65, 66}
	cut := func(mask int, delay int32) Bal_Cut_t {
		c := Bal_Cut_t{Delay: delay}
		for i, leaf := range leaves {
			if mask>>i&1 != 0 {
				c.Leaves = append(c.Leaves, leaf)
			}
		}
		c.Sign = Bal_CutGetSign(c.Leaves)
		return c
	}
	for m0 := 1; m0 < 16; m0++ {
		for m1 := 1; m1 < 16; m1++ {
			for m2 := 1; m2 < 16; m2++ {
				for size := int32(2); size <= 6; size++ {
					a, b, c := cut(m0, 2), cut(m1, 3), cut(m2, 4)
					if len(a.Leaves) > int(size) || len(b.Leaves) > int(size) || len(c.Leaves) > int(size) {
						continue
					}
					var out Bal_Cut_t
					ok := Bal_CutMergeOrderMux(&a, &b, &c, &out, size)
					fmt.Fprintf(got, "muxmerge %d %d %d %d %d", m0, m1, m2, size, giaBool(ok))
					if ok {
						fmt.Fprintf(got, " %d %d %d", out.Delay, out.Func, out.Sign)
						for _, leaf := range out.Leaves {
							fmt.Fprintf(got, " %d", leaf)
						}
					}
					got.WriteByte('\n')
				}
			}
		}
	}
}
