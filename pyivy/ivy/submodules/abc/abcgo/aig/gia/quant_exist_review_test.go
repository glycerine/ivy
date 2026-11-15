package gia

import (
	"fmt"
	"testing"
)

func quantReviewMux() (*Gia_Man_t, int32) {
	g := Gia_ManStart(128)
	Gia_ManQuantSetSuppStart(g)
	var inputs [3]int32
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(g)
		Gia_ManQuantSetSuppZero(g)
		Gia_ManQuantSetSuppCi(g, Gia_ManObj(g, inputs[i]>>1))
	}
	Gia_ManHashAlloc(g)
	return g, Gia_ManHashMux(g, inputs[0], inputs[1], inputs[2])
}

func TestQuantSupportRowAppendKeepsNeighbor(t *testing.T) {
	g, _ := quantReviewMux()
	defer Gia_ManStop(g)
	first, second := Gia_ManCi(g, 0).id, Gia_ManCi(g, 1).id
	before := append([]word(nil), Gia_ManQuantInfoId(g, second)...)
	row := Gia_ManQuantInfoId(g, first)
	grown := append(row, word(12345))
	neighbor := Gia_ManQuantInfoId(g, second)
	for i := range before {
		if neighbor[i] != before[i] {
			t.Fatalf("append changed neighboring support word%d from%d to%d", i, before[i], neighbor[i])
		}
	}
	if grown[len(row)] != 12345 {
		t.Fatal("append did not retain value")
	}
	// The original row is still a writable view of precisely its own words.
	row[0] = 9
	if Gia_ManQuantInfoId(g, first)[0] != 9 {
		t.Fatal("row became an unrelated copy")
	}
}

func TestQuantExistEliminatesAllMuxInputs(t *testing.T) {
	for _, mode := range []int{0, 1} {
		for phase := int32(0); phase < 2; phase++ {
			t.Run(fmt.Sprintf("mode%d/phase%d", mode, phase), func(t *testing.T) {
				defer func() {
					if err := recover(); err != nil {
						t.Fatalf("quantifying all MUX inputs panicked: %v", err)
					}
				}()
				g, root := quantReviewMux()
				defer Gia_ManStop(g)
				keepNone := func(any, int32) int32 { return 0 }
				var lit int32
				if mode == 0 {
					lit = Gia_ManQuantExist(g, root^phase, keepNone, nil)
				} else {
					lit = Gia_ManQuantExist2(g, root^phase, keepNone, nil)
				}
				// Both a MUX and its complement are satisfiable, so eliminating every
				// input must leave true, even when true appears before the last variable.
				if lit != 1 {
					t.Fatalf("got literal%d, want constant true", lit)
				}
			})
		}
	}
}

func quantReviewGraph(seed uint32) (*Gia_Man_t, int32) {
	g := Gia_ManStart(256)
	Gia_ManQuantSetSuppStart(g)
	pool := []int32{0, 1}
	rng := seed
	random := func() uint32 { rng = rng*1664525 + 1013904223; return rng >> 8 }
	edge := func() int32 { return pool[int(random())%len(pool)] ^ int32(random()&1) }
	for i := 0; i < 4; i++ {
		lit := Gia_ManAppendCi(g)
		pool = append(pool, lit)
		Gia_ManQuantSetSuppZero(g)
		Gia_ManQuantSetSuppCi(g, Gia_ManObj(g, lit>>1))
	}
	Gia_ManHashAlloc(g)
	for i := 0; i < 12; i++ {
		a, b, c := edge(), edge(), edge()
		var lit int32
		switch random() % 3 {
		case 0:
			lit = Gia_ManHashAnd(g, a, b)
		case 1:
			lit = Gia_ManHashXor(g, a, b)
		case 2:
			lit = Gia_ManHashMux(g, a, b, c)
		}
		pool = append(pool, lit)
	}
	return g, pool[len(pool)-1]
}

func quantReviewExpected(src []bool, keep, bits int) bool {
	for assignment, value := range src {
		if assignment&keep == bits&keep && value {
			return true
		}
	}
	return false
}
func TestQuantExistGeneratedTruthMatrix(t *testing.T) {
	queries := 0
	for seed := uint32(0); seed < 64; seed++ {
		for _, mode := range []int{0, 1} {
			for phase := int32(0); phase < 2; phase++ {
				g, root := quantReviewGraph(seed)
				root ^= phase
				truth := make([]bool, 16)
				for bits := range truth {
					values := cofReviewValues(g, bits)
					truth[bits] = values[root>>1] != (root&1 != 0)
				}
				// Reuse the graph across keep masks to expose stale traversal/copy/support
				// scratch. Each query must retain the original root's Boolean function.
				for _, keepMask := range []int{0, 15, 1, 14, 2, 13, 3, 12, 4, 11, 5, 10, 6, 9, 7, 8, 0, 15} {
					keep := func(data any, ci int32) int32 {
						if ci < 0 || ci >= 4 {
							t.Fatalf("callback received non-CI index%d", ci)
						}
						return giaBool(data.(int)&(1<<ci) != 0)
					}
					var result int32
					func() {
						defer func() {
							if err := recover(); err != nil {
								t.Fatalf("seed%d mode%d phase%d keep%x panic: %v", seed, mode, phase, keepMask, err)
							}
						}()
						if mode == 0 {
							result = Gia_ManQuantExist(g, root, keep, keepMask)
						} else {
							result = Gia_ManQuantExist2(g, root, keep, keepMask)
						}
					}()
					if result < 0 || result>>1 >= g.nObjs {
						t.Fatalf("invalid result%d", result)
					}
					for bits := range truth {
						values := cofReviewValues(g, bits)
						got := values[result>>1] != (result&1 != 0)
						want := quantReviewExpected(truth, keepMask, bits)
						if got != want {
							t.Fatalf("seed%d mode%d phase%d keep%x input%x literal%d got%t want%t", seed, mode, phase, keepMask, bits, result, got, want)
						}
					}
					queries++
				}
				Gia_ManStop(g)
			}
		}
	}
	t.Logf("%d repeated quantification queries match scalar existential truth", queries)
}

func TestQuantExist2RetainsUnaffectedCiCofactor(t *testing.T) {
	g := Gia_ManStart(32)
	defer Gia_ManStop(g)
	Gia_ManQuantSetSuppStart(g)
	var inputs [3]int32
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(g)
		Gia_ManQuantSetSuppZero(g)
		Gia_ManQuantSetSuppCi(g, Gia_ManObj(g, inputs[i]>>1))
	}
	Gia_ManHashAlloc(g)
	root := Gia_ManHashAnd(g, inputs[2], Gia_ManHashOr(g, inputs[0], inputs[1]))
	keepC := func(_ any, ci int32) int32 { return giaBool(ci == 2) }
	result := Gia_ManQuantExist2(g, root, keepC, nil)
	// exists a,b: c AND (a OR b) == c. After eliminating b, one cofactor
	// is just c; eliminating a must leave that unaffected input intact.
	if result != inputs[2] {
		t.Fatalf("got literal%d, want retained c literal%d", result, inputs[2])
	}
}

func TestQuantExist2TypedConeTruth(t *testing.T) {
	for _, kind := range []string{"xor", "mux", "buffer"} {
		for phase := int32(0); phase < 2; phase++ {
			t.Run(fmt.Sprintf("%s/phase%d", kind, phase), func(t *testing.T) {
				defer func() {
					if err := recover(); err != nil {
						t.Fatalf("typed quantification panicked: %v", err)
					}
				}()
				for mask := 0; mask < 8; mask++ {
					g := Gia_ManStart(32)
					g.EnableMuxes()
					a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
					var root int32
					switch kind {
					case "xor":
						root = Gia_ManAppendXorReal(g, a, b)
					case "mux":
						root = Gia_ManAppendMuxReal(g, c, a, b)
					case "buffer":
						root = Gia_ManAppendBuf(g, a^1)
					}
					root ^= phase
					truth := make([]bool, 8)
					for bits := range truth {
						values := cofReviewValues(g, bits)
						truth[bits] = values[root>>1] != (root&1 != 0)
					}
					Gia_ManHashAlloc(g)
					keep := func(data any, ci int32) int32 { return int32(data.(int)>>ci) & 1 }
					result := Gia_ManQuantExist2(g, root, keep, mask)
					for bits := range truth {
						values := cofReviewValues(g, bits)
						got := values[result>>1] != (result&1 != 0)
						want := quantReviewExpected(truth, mask, bits)
						if got != want {
							t.Fatalf("mask%x input%d got%t want%t literal%d", mask, bits, got, want, result)
						}
					}
					Gia_ManStop(g)
				}
			})
		}
	}
}

func TestQuantRawConstantFanins(t *testing.T) {
	for _, shape := range []string{"zero", "one", "repeated", "opposite"} {
		for mode := 0; mode < 2; mode++ {
			for phase := int32(0); phase < 2; phase++ {
				t.Run(fmt.Sprintf("%s/mode%d/phase%d", shape, mode, phase), func(t *testing.T) {
					defer func() {
						if err := recover(); err != nil {
							t.Fatalf("raw-constant quantification panicked: %v", err)
						}
					}()
					for mask := 0; mask < 4; mask++ {
						g := Gia_ManStart(32)
						g.fGiaSimple = 1
						Gia_ManQuantSetSuppStart(g)
						a := Gia_ManAppendCi(g)
						Gia_ManQuantSetSuppZero(g)
						Gia_ManQuantSetSuppCi(g, Gia_ManObj(g, a>>1))
						b := Gia_ManAppendCi(g)
						Gia_ManQuantSetSuppZero(g)
						Gia_ManQuantSetSuppCi(g, Gia_ManObj(g, b>>1))
						var root int32
						switch shape {
						case "zero":
							root = Gia_ManAppendAnd(g, 0, a)
						case "one":
							root = Gia_ManAppendAnd(g, 1, a)
						case "repeated":
							root = Gia_ManAppendAnd(g, a, a)
						case "opposite":
							root = Gia_ManAppendAnd(g, a, a^1)
						}
						root ^= phase
						truth := make([]bool, 4)
						for bits := range truth {
							values := cofReviewValues(g, bits)
							truth[bits] = values[root>>1] != (root&1 != 0)
						}
						Gia_ManHashAlloc(g)
						keep := func(_ any, ci int32) int32 { return int32(mask>>ci) & 1 }
						var result int32
						if mode == 0 {
							result = Gia_ManQuantExist(g, root, keep, nil)
						} else {
							result = Gia_ManQuantExist2(g, root, keep, nil)
						}
						for bits := range truth {
							values := cofReviewValues(g, bits)
							got := values[result>>1] != (result&1 != 0)
							want := quantReviewExpected(truth, mask, bits)
							if got != want {
								t.Fatalf("keep%x input%d literal%d got%t want%t", mask, bits, result, got, want)
							}
						}
						Gia_ManStop(g)
					}
				})
			}
		}
	}
}

func TestQuantConeCopyMappingAndOwnership(t *testing.T) {
	for _, kind := range []string{"xor", "mux", "buffer"} {
		for phase := int32(0); phase < 2; phase++ {
			src := esopReviewTypedGraph(kind, 5)
			saved := make([]uint32, src.nObjs)
			for i, obj := range src.pObjs {
				obj.Value = uint32(123 + i*17)
				saved[i] = obj.Value
			}
			root := Gia_ObjFaninLit0(Gia_ManCo(src, 0), Gia_ManCo(src, 0).id) ^ phase
			keep := func(_ any, ci int32) int32 { return giaBool(ci == 1) }
			var cis *Vec_Int_t
			var lit int32
			out := Gia_ManQuantDupConeSupp(src, root, keep, nil, &cis, &lit)
			if Vec_IntSize(cis) != int(Gia_ManCiNum(out)) {
				t.Fatal("mapping does not cover copied CIs")
			}
			for i, obj := range src.pObjs {
				if obj.Value != saved[i] {
					t.Fatal("cone copy modified unrelated Value scratch")
				}
			}
			for _, copy := range Vec_IntArray(&src.vCopies) {
				if copy != -1 {
					t.Fatal("cone copy left occupied slots")
				}
			}
			quantVars := int32(0)
			for _, ci := range Vec_IntArray(cis) {
				if ci != 1 {
					quantVars++
				}
			}
			if out.iSuppPi != quantVars {
				t.Fatalf("tracked%d variables, want%d", out.iSuppPi, quantVars)
			}
			Gia_ManAppendCo(out, lit)
			for bits := 0; bits < 8; bits++ {
				oldInputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
				newInputs := make([]bool, Vec_IntSize(cis))
				for i, ci := range Vec_IntArray(cis) {
					newInputs[i] = oldInputs[ci]
				}
				before, e1 := src.Evaluate(oldInputs)
				after, e2 := out.Evaluate(newInputs)
				if e1 != nil || e2 != nil || after[0] != (before[0] != (phase != 0)) {
					t.Fatalf("kind%s phase%d input%d old%v new%v errors%v/%v", kind, phase, bits, before, after, e1, e2)
				}
			}
			// Source disposal cannot invalidate the copied manager or its CI map.
			Gia_ManStop(src)
			if _, err := out.Evaluate(make([]bool, Vec_IntSize(cis))); err != nil {
				t.Fatal(err)
			}
			Gia_ManStop(out)
		}
	}
}
