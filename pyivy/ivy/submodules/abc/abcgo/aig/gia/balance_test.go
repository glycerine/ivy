package gia

import (
	"math/rand"
	"testing"
)

func TestBalanceAndAreaBalanceTruth(t *testing.T) {
	rng := rand.New(rand.NewSource(922))
	for run := 0; run < 30; run++ {
		src := Gia_ManStart(80)
		literals := []int32{0, 1}
		for i := 0; i < 5; i++ {
			literals = append(literals, Gia_ManAppendCi(src))
		}
		Gia_ManHashStart(src)
		for i := 0; i < 35; i++ {
			a, b := literals[rng.Intn(len(literals))]^int32(rng.Intn(2)), literals[rng.Intn(len(literals))]^int32(rng.Intn(2))
			var lit int32
			switch rng.Intn(3) {
			case 0:
				lit = Gia_ManHashAnd(src, a, b)
			case 1:
				lit = Gia_ManHashXor(src, a, b)
			default:
				lit = Gia_ManHashMux(src, literals[rng.Intn(len(literals))], a, b)
			}
			literals = append(literals, lit)
		}
		for i := 0; i < 4; i++ {
			Gia_ManAppendCo(src, literals[len(literals)-1-i])
		}
		Gia_ManHashStop(src)
		for _, simple := range []int32{0, 1} {
			for _, area := range []bool{false, true} {
				var dst *Gia_Man_t
				if area {
					dst = Gia_ManAreaBalance(src, simple, 20, 0, 0)
				} else {
					dst = Gia_ManBalance(src, simple, 1, 0)
				}
				if Gia_ManCiNum(dst) != 5 || Gia_ManCoNum(dst) != 4 {
					t.Fatal("changed interface")
				}
				for mask := 0; mask < 32; mask++ {
					inputs := make([]bool, 5)
					for i := range inputs {
						inputs[i] = mask&(1<<uint(i)) != 0
					}
					want, e := src.Evaluate(inputs)
					if e != nil {
						t.Fatal(e)
					}
					got, e := dst.Evaluate(inputs)
					if e != nil {
						t.Fatal(e)
					}
					for i := range want {
						if got[i] != want[i] {
							t.Fatal("balance changed function", run, simple, area, mask, i)
						}
					}
				}
			}
		}
	}
}

func TestBalanceReducesSerialAndDepth(t *testing.T) {
	src := Gia_ManStart(40)
	root := Gia_ManAppendCi(src)
	for i := 0; i < 15; i++ {
		root = Gia_ManAppendAnd(src, root, Gia_ManAppendCi(src))
	}
	Gia_ManAppendCo(src, root)
	if before := Gia_ManLevelNum(src); before != 15 {
		t.Fatal(before)
	}
	dst := Gia_ManBalance(src, 1, 1, 0)
	if after := Gia_ManLevelNum(dst); after != 4 {
		t.Fatal("unbalanced result", after)
	}
}
