package wlc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"os"
	"testing"
)

func TestBlastDividerExhaustive(t *testing.T) {
	for n := int32(1); n <= 4; n++ {
		for quotient := int32(0); quotient < 2; quotient++ {
			g := gia.Gia_ManStart(1000)
			gia.Gia_ManHashAlloc(g)
			a, b := make([]int32, n), make([]int32, n)
			for i := range a {
				a[i] = gia.Gia_ManAppendCi(g)
			}
			for i := range b {
				b[i] = gia.Gia_ManAppendCi(g)
			}
			out := vec.Vec_IntAlloc(0)
			Wlc_BlastDivider(g, a, n, b, n, quotient, out)
			for _, v := range vec.Vec_IntArray(out) {
				gia.Gia_ManAppendCo(g, v)
			}
			for x := int32(0); x < 1<<n; x++ {
				for y := int32(0); y < 1<<n; y++ {
					inputs := make([]bool, 2*n)
					for i := int32(0); i < n; i++ {
						inputs[i] = x&(1<<i) != 0
						inputs[n+i] = y&(1<<i) != 0
					}
					result, err := g.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					want := x
					if y != 0 {
						if quotient != 0 {
							want = x / y
						} else {
							want = x % y
						}
					} else if quotient != 0 {
						want = (1 << n) - 1
					}
					for i := int32(0); i < n; i++ {
						if result[i] != (want&(1<<i) != 0) {
							t.Fatalf("width=%d quo=%d x=%d y=%d result=%v want=%d", n, quotient, x, y, result, want)
						}
					}
				}
			}
		}
	}
}

func TestBlastSignedDividerExhaustive(t *testing.T) {
	for n := int32(1); n <= 4; n++ {
		for quotient := int32(0); quotient < 2; quotient++ {
			g := gia.Gia_ManStart(3000)
			gia.Gia_ManHashAlloc(g)
			a, b := make([]int32, n), make([]int32, n)
			for i := range a {
				a[i] = gia.Gia_ManAppendCi(g)
			}
			for i := range b {
				b[i] = gia.Gia_ManAppendCi(g)
			}
			out := vec.Vec_IntAlloc(0)
			Wlc_BlastDividerSigned(g, a, n, b, n, quotient, out)
			for _, v := range vec.Vec_IntArray(out) {
				gia.Gia_ManAppendCo(g, v)
			}
			for x := int32(0); x < 1<<n; x++ {
				for y := int32(1); y < 1<<n; y++ {
					inputs := make([]bool, 2*n)
					for i := int32(0); i < n; i++ {
						inputs[i] = x&(1<<i) != 0
						inputs[n+i] = y&(1<<i) != 0
					}
					sx, sy := x, y
					if x&(1<<(n-1)) != 0 {
						sx -= 1 << n
					}
					if y&(1<<(n-1)) != 0 {
						sy -= 1 << n
					}
					result, err := g.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					want := sx % sy
					if quotient != 0 {
						want = sx / sy
					}
					for i := int32(0); i < n; i++ {
						if result[i] != (want&(1<<i) != 0) {
							t.Fatalf("width=%d quo=%d x=%d y=%d result=%v want=%d", n, quotient, x, y, result, want)
						}
					}
				}
			}
		}
	}
}

// Source fixtures include zero divisors and signed overflow cases.
func TestBlastSignedDividerMatchesC(t *testing.T) {
	file, err := os.Open("test_vectors/divide_signed_small.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var n, q, x, y, want int32
	for {
		_, err = fmt.Fscan(file, &n, &q, &x, &y, &want)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		g := gia.Gia_ManStart(1000)
		gia.Gia_ManHashAlloc(g)
		a, b := make([]int32, n), make([]int32, n)
		for i := range a {
			a[i] = (x >> i) & 1
			b[i] = (y >> i) & 1
		}
		out := vec.Vec_IntAlloc(0)
		Wlc_BlastDividerSigned(g, a, n, b, n, q, out)
		var got int32
		for i, lit := range vec.Vec_IntArray(out) {
			if lit > 1 {
				t.Fatal("expected constant")
			}
			got |= lit << i
		}
		if got != want {
			t.Fatalf("n=%d quotient=%d x=%d y=%d got=%d C=%d", n, q, x, y, got, want)
		}
		gia.Gia_ManStop(g)
	}
}

func TestBlastZeroCondition(t *testing.T) {
	g := gia.Gia_ManStart(32)
	gia.Gia_ManHashAlloc(g)
	a, b, result := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	res := vec.Vec_IntAllocArrayCopy([]int32{result}, 1)
	Wlc_BlastZeroCondition(g, []int32{a, b}, 2, res)
	gia.Gia_ManAppendCo(g, vec.Vec_IntEntry(res, 0))
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		out, err := g.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		if out[0] != ((in[0] || in[1]) && in[2]) {
			t.Fatalf("inputs %v result %v", in, out)
		}
	}
}
