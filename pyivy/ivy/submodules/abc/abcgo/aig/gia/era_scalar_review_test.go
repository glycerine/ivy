package gia

import (
	"fmt"
	"testing"
)

// Scalar truth evaluation is independent of the reachability engine's packed
// PI truth tables, register broadcasting and bit extraction.
func TestExplicitReachabilityPackedSimulationAgainstScalar(t *testing.T) {
	for _, ni := range []int32{0, 1, 5, 6, 12} {
		for phase := int32(0); phase < 8; phase++ {
			t.Run(fmt.Sprintf("inputs%d/phase%d", ni, phase), func(t *testing.T) {
				g := Gia_ManStart(128)
				defer Gia_ManStop(g)
				a, b := int32(0), int32(1)
				for i := int32(0); i < ni; i++ {
					v := Gia_ManAppendCi(g)
					if i == 0 {
						a = v
					}
					b = v
				}
				var regs []int32
				for i := 0; i < 65; i++ {
					regs = append(regs, Gia_ManAppendCi(g))
				}
				g.EnableMuxes()
				x := Gia_ManAppendXorReal(g, a^(phase&1), regs[32]^((phase>>1)&1))
				m := Gia_ManAppendMuxReal(g, b^((phase>>2)&1), regs[64], x)
				z := Gia_ManAppendAnd(g, x^1, m)
				for _, v := range []int32{x, m, z, z ^ 1} {
					Gia_ManAppendCo(g, v)
				}
				for i, ro := range regs {
					v := ro
					switch i {
					case 0:
						v = m
					case 32:
						v = x
					case 64:
						v = z
					}
					Gia_ManAppendCo(g, v)
				}
				Gia_ManSetRegNum(g, 65)
				p := Gia_ManEraCreate(g)
				defer Gia_ManEraFree(p)
				state := Gia_ManEraCreateState(p)
				for _, words := range [][3]uint32{{0, 0, 0}, {^uint32(0), ^uint32(0), 1}, {0xaaaaaaaa, 0x55555555, 1}} {
					copy(state.pData, words[:])
					Gia_ManInsertState(p, state)
					Gia_ManPerformOneIter(p)
					in := make([]bool, ni+65)
					for i := 0; i < 65; i++ {
						in[int(ni)+i] = words[i>>5]&(1<<uint(i&31)) != 0
					}
					for condition := int32(0); condition < 1<<ni; condition++ {
						for i := int32(0); i < ni; i++ {
							in[i] = condition&(1<<i) != 0
						}
						want, err := g.Evaluate(in)
						if err != nil {
							t.Fatal(err)
						}
						for i, value := range want {
							data := Gia_ManEraData(p, Gia_ManCo(g, int32(i)).id)
							got := data[condition>>5]&(1<<uint(condition&31)) != 0
							if got != value {
								t.Fatalf("state%x input%d CO%d got%v want%v", words, condition, i, got, value)
							}
						}
					}
				}
			})
		}
	}
}

func TestExplicitReachabilityHashResizeAndScratch(t *testing.T) {
	g := eraReviewGraph(1, 65, 0)
	defer Gia_ManStop(g)
	p := Gia_ManEraCreate(g)
	defer Gia_ManEraFree(p)
	p.nBins, p.pBins = 3, make([]int32, 3)
	// Deliberately collide the low word while varying bits above each 32-bit
	// boundary. The pending scratch row must remain outside the hash table.
	for i := uint32(0); i < 40; i++ {
		s := Gia_ManEraCreateState(p)
		s.pData[0], s.pData[1], s.pData[2] = i%3, i, i%2
		var n int32
		place := Gia_ManEraHashFind(p, s, &n)
		if place == nil || n != -1 {
			t.Fatal("new state falsely matched")
		}
		*place = s.Num
	}
	p.pStateNew = Gia_ManEraCreateState(p)
	copy(p.pStateNew.pData, []uint32{999, 999, 1})
	for resize := 0; resize < 3; resize++ {
		previous := p.nBins
		Gia_ManEraHashResize(p)
		if p.nBins <= previous {
			t.Fatal("hash did not grow")
		}
		for i, s := range p.vStates[1:] {
			var n int32
			place := Gia_ManEraHashFind(p, s, &n)
			if s == p.pStateNew {
				if place == nil || n != -1 {
					t.Fatal("scratch state was inserted")
				}
			} else if place != nil || n != int32(i+1) {
				t.Fatalf("state%d lost after resize%d", i+1, resize)
			}
		}
	}
}
