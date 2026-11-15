package gia

import (
	"fmt"
	"testing"
)

func cubeReviewRandomMachine(seed uint32, real bool) *Gia_Man_t {
	g := Gia_ManStart(128)
	pool := []int32{0, 1}
	for i := 0; i < 6; i++ {
		pool = append(pool, Gia_ManAppendCi(g))
	}
	if real {
		g.EnableMuxes()
	}
	Gia_ManHashAlloc(g)
	random := func() uint32 { seed = seed*1664525 + 1013904223; return seed >> 8 }
	edge := func() int32 { return pool[int(random())%len(pool)] ^ int32(random()&1) }
	for i := 0; i < 18; i++ {
		a, b, c := edge(), edge(), edge()
		var out int32
		switch random() % 4 {
		case 0:
			out = Gia_ManHashAnd(g, a, b)
		case 1:
			if real {
				out = Gia_ManHashXorReal(g, a, b)
			} else {
				out = Gia_ManHashXor(g, a, b)
			}
		case 2:
			if real {
				out = Gia_ManHashMuxReal(g, a, b, c)
			} else {
				out = Gia_ManHashMux(g, a, b, c)
			}
		case 3:
			out = Gia_ManAppendBuf(g, a)
		}
		pool = append(pool, out)
	}
	next := [3]int32{edge(), edge(), edge()}
	outputs := [2]int32{edge(), edge()}
	// Include identities that need SAT rather than just structural hashing.
	a, b, c := int32(2), int32(4), int32(6)
	x := Gia_ManHashOr(g, Gia_ManHashAnd(g, a, b), Gia_ManHashAnd(g, a, c))
	y := Gia_ManHashAnd(g, a, Gia_ManHashOr(g, b, c))
	z := Gia_ManHashXor(g, x, y)
	if real {
		z = Gia_ManHashXorReal(g, x, y)
	}
	if seed&3 == 0 {
		outputs = [2]int32{z, 0}
		next[0] = z
	}
	if seed&7 == 0 {
		next[1] = z ^ 1
	}
	for _, out := range outputs {
		Gia_ManAppendCo(g, out)
	}
	for _, out := range next {
		Gia_ManAppendCo(g, out)
	}
	Gia_ManHashStop(g)
	Gia_ManSetRegNum(g, 3)
	return g
}

func cubeReviewScalarResult(t *testing.T, g *Gia_Man_t) int32 {
	t.Helper()
	queue := []int{0}
	seen := [8]bool{true}
	for at := 0; at < len(queue); at++ {
		state := queue[at]
		for input := 0; input < 8; input++ {
			values := []bool{input&1 != 0, input&2 != 0, input&4 != 0, state&1 != 0, state&2 != 0, state&4 != 0}
			out, err := g.Evaluate(values)
			if err != nil {
				t.Fatal(err)
			}
			if out[0] || out[1] {
				return 0
			}
			next := 0
			for i, v := range out[2:] {
				if v {
					next |= 1 << i
				}
			}
			if !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return 1
}

func TestCubeReachabilityRandomMachinesAgainstScalarBFS(t *testing.T) {
	for seed := uint32(0); seed < 64; seed++ {
		for _, real := range []bool{false, true} {
			t.Run(fmt.Sprintf("seed%d/real%t", seed, real), func(t *testing.T) {
				defer func() {
					if failure := recover(); failure != nil {
						t.Errorf("valid machine panicked: %v", failure)
					}
				}()
				g := cubeReviewRandomMachine(seed, real)
				defer Gia_ManStop(g)
				want := cubeReviewScalarResult(t, g)
				for miter := int32(0); miter < 2; miter++ {
					expected := int32(1)
					if miter != 0 {
						expected = want
					}
					var got int32
					hashReviewStdout(t, func() { got = Gia_ManArePerform(g, 1000, miter, 0) })
					if got != expected {
						t.Fatalf("miter%d result%d scalar%d", miter, got, expected)
					}
					if got == 0 && (g.pCexSeq == nil || Gia_ManVerifyCex(g, g.pCexSeq, 0) == 0) {
						t.Fatal("invalid failure trace")
					}
					if got == 1 && g.pCexSeq != nil {
						t.Fatal("safe run retained a trace")
					}
				}
			})
		}
	}
}
