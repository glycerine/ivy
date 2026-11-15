package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func balanceDepthReviewGraph(family, signs int) *Gia_Man_t {
	g := Gia_ManStart(128)
	ci := make([]int32, 6)
	for i := range ci {
		ci[i] = Gia_ManAppendCi(g)
	}
	g.EnableMuxes()
	Gia_ManHashStart(g)
	phase := func(bit uint) int32 { return int32(signs>>bit) & 1 }
	x := Gia_ManHashAnd(g, ci[0]^phase(0), ci[1]^phase(1))
	root := x
	switch family {
	case 0:
		for i := 2; i < 6; i++ {
			root = Gia_ManHashAnd(g, root, ci[i]^phase(uint(i%3)))
		}
	case 1:
		left, right := Gia_ManHashAnd(g, x, ci[2]), Gia_ManHashAnd(g, x, ci[3])
		root = Gia_ManHashAnd(g, left, right)
	case 2:
		x = Gia_ManHashXorReal(g, ci[0]^phase(0), ci[1]^phase(1))
		left := Gia_ManHashXorReal(g, x, ci[2]^phase(2))
		root = Gia_ManHashAnd(g, left, ci[4])
	case 3:
		x = Gia_ManHashMuxReal(g, ci[2]^phase(2), ci[0]^phase(0), ci[1]^phase(1))
		root = Gia_ManHashAnd(g, x, ci[3])
	case 4:
		first := Gia_ManAppendBuf(g, x^phase(2))
		left := Gia_ManHashAnd(g, first, ci[2])
		root = Gia_ManAppendBuf(g, left)
		Gia_ManAppendBuf(g, ci[3])
	case 5:
		other := Gia_ManHashAnd(g, ci[0]^phase(0), ci[1]^phase(1)^1)
		root = Gia_ManHashAnd(g, x^1, other^1)
	}
	for _, lit := range []int32{root ^ phase(3), x, ci[0], ci[1], root ^ 1, ci[4] ^ phase(2), ci[5] ^ 1} {
		Gia_ManAppendCo(g, lit)
	}
	Gia_ManSetRegNum(g, 2)
	Gia_ManHashStop(g)
	return g
}

func balanceDepthReviewArrivals(g *Gia_Man_t, mode int, delay int32) {
	g.And2Delay = delay
	if mode == 1 || mode == 3 {
		g.vCiArrs = permutation(0, 5, 2, 8, 1, 3)
	}
	if mode == 2 || mode == 3 {
		g.VInArrs = Vec_FltAlloc(6)
		for _, value := range []float32{4.25, 0.75, 3.5, 1.25, 8, 2.5} {
			Vec_FltPush(g.VInArrs, value)
		}
	}
}

func TestDepthBalanceNativeArrivalsOracle(t *testing.T) {
	exe := compileGiaSourceOracle(t, "balance_depth.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v\n%s", err, native)
	}
	var got strings.Builder
	cases := 0
	for family := 0; family < 6; family++ {
		for signs := 0; signs < 16; signs++ {
			for strict := int32(0); strict < 2; strict++ {
				for arrival := 0; arrival < 4; arrival++ {
					for _, delay := range []int32{0, 2, 3} {
						src := balanceDepthReviewGraph(family, signs)
						balanceDepthReviewArrivals(src, arrival, delay)
						out := Gia_ManBalanceInt(src, strict)
						fmt.Fprintf(&got, "case %d %d %d %d %d result %d %d %d %d values", family, signs, strict, arrival, delay, out.nObjs, out.nRegs, out.nXors, out.nMuxes)
						for _, obj := range src.pObjs {
							fmt.Fprintf(&got, " %d", obj.Value)
						}
						got.WriteString(" nodes")
						for _, obj := range out.pObjs {
							a, b, c, kind := out.Gate(obj.id)
							switch kind {
							case 0, 1:
								fmt.Fprintf(&got, " %d", kind)
							case 5, 6:
								fmt.Fprintf(&got, " %d:%d", kind, a)
							case 2, 3:
								fmt.Fprintf(&got, " %d:%d:%d", kind, a, b)
							case 4:
								fmt.Fprintf(&got, " 4:%d:%d:%d", a, b, c)
							default:
								t.Fatalf("invalid kind%d", kind)
							}
						}
						got.WriteString(" truth")
						for mask := 0; mask < 64; mask++ {
							inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0, mask&32 != 0}
							want, err := src.Evaluate(inputs)
							if err != nil {
								t.Fatal(err)
							}
							checkQuantificationOutputs(t, out, inputs, want)
							values, err := out.Evaluate(inputs)
							if err != nil {
								t.Fatal(err)
							}
							packed := 0
							for i, value := range values {
								if value {
									packed |= 1 << i
								}
							}
							fmt.Fprintf(&got, " %d", packed)
						}
						got.WriteByte('\n')
						cases++
					}
				}
			}
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(string(native), "\n"), strings.Split(got.String(), "\n")
		for i := 0; i < len(a) && i < len(b); i++ {
			if a[i] != b[i] {
				t.Fatalf("line%d\nC: %s\nGo: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("line counts C%d Go%d", len(a), len(b))
	}
	t.Logf("%d depth-balancing cases and %d truth rows match C, including strict sharing and integer/float arrival precedence", cases, cases*64)
}
