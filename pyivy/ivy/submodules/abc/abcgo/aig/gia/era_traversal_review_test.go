package gia

import (
	"fmt"
	"slices"
	"testing"
)

func eraReviewTypedMachine(phase int32) *Gia_Man_t {
	g := Gia_ManStart(128)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	r := make([]int32, 65)
	for i := range r {
		r[i] = Gia_ManAppendCi(g)
	}
	g.EnableMuxes()
	x := Gia_ManAppendXorReal(g, a^(phase&1), r[32])
	m := Gia_ManAppendMuxReal(g, b^((phase>>1)&1), r[64], x)
	Gia_ManAppendCo(g, Gia_ManAppendAnd(g, r[0], r[32])^((phase>>2)&1))
	Gia_ManAppendCo(g, r[64])
	for i, ro := range r {
		switch i {
		case 0:
			ro = x
		case 32:
			ro = m
		case 64:
			ro = r[0]
		}
		Gia_ManAppendCo(g, ro)
	}
	Gia_ManSetRegNum(g, 65)
	return g
}

// A plain scalar BFS owns its own Boolean states, queue, predecessor tree and
// visited map. It does not use reachability hash, simulation or trace helpers.
func TestExplicitReachabilityAgainstScalarBFS(t *testing.T) {
	for phase := int32(0); phase < 8; phase++ {
		for miter := int32(0); miter < 2; miter++ {
			t.Run(fmt.Sprintf("phase%d/miter%d", phase, miter), func(t *testing.T) {
				g := eraReviewTypedMachine(phase)
				defer Gia_ManStop(g)
				p := Gia_ManEraCreate(g)
				defer Gia_ManEraFree(p)
				initial := Gia_ManEraCreateState(p)
				p.pBins[Gia_ManEraStateHash(initial.pData, p.nWordsDat, p.nBins)] = 1
				key := func(state []bool) string {
					bits := make([]byte, len(state))
					for i, v := range state {
						if v {
							bits[i] = 1
						}
					}
					return string(bits)
				}
				states := [][]bool{make([]bool, 65)}
				previous, conditions := []int{0}, []int32{0}
				seen := map[string]int{key(states[0]): 0}
				var rows []int32
				failed := false
				for at := 0; at < len(states); at++ {
					if at > 16 {
						t.Fatal("scalar machine exceeded its three changing state bits")
					}
					s := p.vStates[at+1]
					for i, v := range states[at] {
						if got := s.pData[i>>5]&(1<<uint(i&31)) != 0; got != v {
							t.Fatalf("state%d register%d got%v want%v", at, i, got, v)
						}
					}
					Gia_ManInsertState(p, s)
					Gia_ManPerformOneIter(p)
					var truth [4][]bool
					for input := 0; input < 4; input++ {
						in := append([]bool{input&1 != 0, input&2 != 0}, states[at]...)
						out, err := g.Evaluate(in)
						if err != nil {
							t.Fatal(err)
						}
						truth[input] = out
					}
					bad := -1
					if miter != 0 {
						for output := 0; output < 2 && bad < 0; output++ {
							for input := 0; input < 4; input++ {
								if truth[input][output] {
									bad = input
									break
								}
							}
						}
					}
					gotBad := Gia_ManAnalyzeResult(p, s, miter, 1)
					if bad >= 0 {
						failed = true
						if gotBad != 1 {
							t.Fatal("missed scalar violation")
						}
						trace := []int32{int32(bad)}
						for cursor := at; ; cursor = previous[cursor] {
							trace = append(trace, conditions[cursor])
							if cursor == 0 {
								break
							}
						}
						slices.Reverse(trace)
						if got := Vec_IntArray(p.vBugTrace); !slices.Equal(got, trace) {
							t.Fatalf("trace%v want%v", got, trace)
						}
						break
					}
					if gotBad != 0 {
						t.Fatal("false scalar violation")
					}
					for input := 0; input < 4; input++ {
						next := truth[input][2:]
						code := key(next)
						index, ok := seen[code]
						if !ok {
							index = len(states)
							seen[code] = index
							states = append(states, append([]bool(nil), next...))
							previous = append(previous, at)
							conditions = append(conditions, int32(input))
						}
						output := int32(0)
						if truth[input][0] {
							output |= 1
						}
						if truth[input][1] {
							output |= 2
						}
						rows = append(rows, int32(input), int32(at+1), int32(index+1), output)
					}
				}
				if got := Vec_IntArray(p.vStgDump); !slices.Equal(got, rows) {
					t.Fatalf("transition rows%v want%v", got, rows)
				}
				count := len(p.vStates) - 1
				if p.pStateNew != nil {
					count--
				}
				if count != len(states) {
					t.Fatalf("discovered%d states, scalar%d", count, len(states))
				}
				want := int32(1)
				if failed {
					want = 0
				}
				var got int32
				hashReviewStdout(t, func() { got = Gia_ManCollectReachable(g, 100, miter, 0, 0) })
				if got != want {
					t.Fatalf("driver result%d want%d", got, want)
				}
			})
		}
	}
}

func TestExplicitReachabilityFirstOutputPrecedesFirstInput(t *testing.T) {
	g := Gia_ManStart(16)
	defer Gia_ManStop(g)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	ro := Gia_ManAppendCi(g)
	// PO0 first fails at input 3; PO1 fails earlier at input 0. The original
	// algorithm prioritizes output index, then the earliest input for that PO.
	Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, b))
	Gia_ManAppendCo(g, a^1)
	Gia_ManAppendCo(g, ro)
	Gia_ManSetRegNum(g, 1)
	p := Gia_ManEraCreate(g)
	defer Gia_ManEraFree(p)
	initial := Gia_ManEraCreateState(p)
	Gia_ManInsertState(p, initial)
	Gia_ManPerformOneIter(p)
	if Gia_ManAnalyzeResult(p, initial, 1, 1) != 1 {
		t.Fatal("missed initial violation")
	}
	assertIDs(t, p.vBugTrace, 0, 3)
	if Vec_IntSize(p.vStgDump) != 0 || len(p.vStates) != 2 {
		t.Fatal("violating iteration collected successors")
	}
}
