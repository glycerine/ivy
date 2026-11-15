package gia

import (
	"fmt"
	"testing"
)

// Expand without hashing or simplification so ternary precision follows the
// original source gate expansions, rather than a different optimized network.
func frontierReviewExpanded(g *Gia_Man_t) *Gia_Man_t {
	out := Gia_ManStart(4 * g.nObjs)
	copies := make([]int32, g.nObjs)
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		copies[Gia_ManCi(g, i).id] = Gia_ManAppendCi(out)
	}
	read := func(lit int32) int32 { return copies[lit>>1] ^ (lit & 1) }
	for id := int32(1); id < g.nObjs; id++ {
		a, b, c, kind := g.Gate(id)
		switch kind {
		case 2:
			copies[id] = Gia_ManAppendAnd(out, read(a), read(b))
		case 3:
			copies[id] = Gia_ManAppendXor(out, read(a), read(b))
		case 4:
			copies[id] = Gia_ManAppendMux(out, read(a), read(b), read(c))
		case 5:
			copies[id] = Gia_ManAppendBuf(out, read(a))
		}
	}
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		o := Gia_ManCo(g, i)
		Gia_ManAppendCo(out, read(Gia_ObjFaninLit0(o, o.id)))
	}
	Gia_ManSetRegNum(out, g.nRegs)
	return out
}

// Enumerate each fanin's possible Boolean values. This reference neither calls
// the production ternary algebra nor uses frontier slots or reference scratch.
func frontierReviewTernaryValues(g *Gia_Man_t, inputs []int32) []int32 {
	values := make([]int32, g.nObjs)
	values[0] = 1 // set containing false
	for i, v := range inputs {
		values[Gia_ManCi(g, int32(i)).id] = v
	}
	read := func(lit int32) int32 {
		v := values[lit>>1]
		if lit&1 != 0 {
			v = (v&1)<<1 | (v&2)>>1
		}
		return v
	}
	for id := int32(1); id < g.nObjs; id++ {
		a, b, _, kind := g.Gate(id)
		switch kind {
		case 2:
			x, y := read(a), read(b)
			for av := int32(0); av < 2; av++ {
				for bv := int32(0); bv < 2; bv++ {
					if x&(1<<av) != 0 && y&(1<<bv) != 0 {
						values[id] |= 1 << (av & bv)
					}
				}
			}
		case 5, 6:
			values[id] = read(a)
		case 1:
		default:
			panic("reference requires expanded ordinary graph")
		}
	}
	return values
}

func TestFrontierTernaryUnknownsMatchSourceExpansion(t *testing.T) {
	for seed := uint32(0); seed < 8; seed++ {
		for _, allNodes := range []bool{false, true} {
			g, _ := falseReviewTypedGraph(seed)
			if allNodes {
				for _, obj := range g.pObjs[1:] {
					if Gia_ObjIsAnd(obj) != 0 {
						Gia_ManAppendCo(g, obj.id<<1)
					}
				}
			}
			expanded := frontierReviewExpanded(g)
			p := Gia_ManTerCreate(g)
			for pattern := 0; pattern < 81; pattern++ {
				inputs, code := make([]int32, 4), pattern
				for ci := range inputs {
					inputs[ci], code = int32(code%3)+1, code/3
					Gia_ManTerSimInfoSet(p.pDataSimCis, int32(ci), inputs[ci])
				}
				want := frontierReviewTernaryValues(expanded, inputs)
				Gia_ManTerSimulateRound(p)
				for co := int32(0); co < Gia_ManCoNum(g); co++ {
					if got, expected := Gia_ManTerSimInfoGet(p.pDataSimCos, co), want[Gia_ManCo(expanded, co).id]; got != expected {
						t.Fatalf("seed%d all%t pattern%d output%d got%d want%d", seed, allNodes, pattern, co, got, expected)
					}
				}
			}
			Gia_ManTerDelete(p)
			Gia_ManStop(expanded)
			Gia_ManStop(g)
		}
	}
}

func TestFrontierTypedMultiwordAndSwitchingCounts(t *testing.T) {
	for seed := uint32(0); seed < 16; seed++ {
		for _, words := range []int32{1, 2, 5} {
			for _, transitions := range []int32{0, 1} {
				t.Run(fmt.Sprintf("%d/%d/%d", seed, words, transitions), func(t *testing.T) {
					g, _ := falseReviewTypedGraph(seed)
					defer Gia_ManStop(g)
					simPars := Gia_ParSim_t{NWords: words}
					sim := Gia_ManSimCreate(g, &simPars)
					defer Gia_ManSimDelete(sim)
					swiPars := Gia_ParSwi_t{nWords: words, fProbTrans: transitions}
					swi := Gia_ManSwiCreate(g, &swiPars)
					defer Gia_ManSwiDelete(swi)
					rng := seed
					for ci := int32(0); ci < Gia_ManCiNum(g); ci++ {
						for w := int32(0); w < words; w++ {
							rng = rng*1664525 + 1013904223
							Gia_SimDataCi(sim, ci)[w], Gia_SwiDataCi(swi, ci)[w] = rng, rng
						}
					}
					want := make([][]uint32, g.nObjs)
					for id := range want {
						want[id] = make([]uint32, words)
					}
					for w := int32(0); w < words; w++ {
						for bit := uint(0); bit < 32; bit++ {
							input := 0
							for ci := int32(0); ci < Gia_ManCiNum(g); ci++ {
								input |= int(Gia_SimDataCi(sim, ci)[w]>>bit&1) << ci
							}
							for id, v := range cofReviewValues(g, input) {
								if v {
									want[id][w] |= 1 << bit
								}
							}
						}
					}
					for round := int32(1); round <= 2; round++ {
						Gia_ManSimulateRound(sim)
						Gia_ManSwiSimulateRound(swi, 1)
						for co := int32(0); co < Gia_ManCoNum(g); co++ {
							for w, expected := range want[Gia_ManCo(g, co).id] {
								if Gia_SimDataCo(sim, co)[w] != expected || Gia_SwiDataCo(swi, co)[w] != expected {
									t.Fatalf("round%d output%d word%d differs", round, co, w)
								}
							}
						}
						for id, obj := range g.pObjs {
							if Gia_ObjIsCo(obj) != 0 {
								continue
							}
							var expected int32
							for _, word := range want[id] {
								for bit := uint(0); bit < 32; bit++ {
									if transitions == 0 {
										expected += int32(word >> bit & 1)
									} else if bit < 16 && (word>>bit&1 != word>>(bit+16)&1) {
										expected += 2
									}
								}
							}
							if swi.pData1[id] != round*expected {
								t.Fatalf("round%d node%d count%d want%d", round, id, swi.pData1[id], round*expected)
							}
						}
					}
				})
			}
		}
	}
}

func BenchmarkFrontierWordSimulation(b *testing.B) {
	for _, words := range []int32{1, 8, 64} {
		b.Run(fmt.Sprint(words), func(b *testing.B) {
			g := frontierReviewGraph(7)
			defer Gia_ManStop(g)
			pars := Gia_ParSim_t{NWords: words}
			p := Gia_ManSimCreate(g, &pars)
			defer Gia_ManSimDelete(p)
			for i := range p.pDataSimCis {
				p.pDataSimCis[i] = uint32(i)*1664525 + 1013904223
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Gia_ManSimulateRound(p)
			}
		})
	}
}
