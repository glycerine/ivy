package wlc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"os"
	"testing"
)

func TestBlastMultiplierTruth(t *testing.T) {
	for n := int32(1); n <= 3; n++ {
		for signed := int32(0); signed <= 1; signed++ {
			g := gia.Gia_ManStart(1000)
			gia.Gia_ManHashAlloc(g)
			a, b := make([]int32, n), make([]int32, n)
			for i := range a {
				a[i] = gia.Gia_ManAppendCi(g)
			}
			for i := range b {
				b[i] = gia.Gia_ManAppendCi(g)
			}
			temp, res, small, neg := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
			Wlc_BlastMultiplier(g, a, b, n, n, temp, res, signed)
			Wlc_BlastMultiplier2(g, a, b, n, temp, small)
			Wlc_BlastMinus(g, a, n, neg)
			for _, v := range vec.Vec_IntArray(res) {
				gia.Gia_ManAppendCo(g, v)
			}
			for _, v := range vec.Vec_IntArray(small) {
				gia.Gia_ManAppendCo(g, v)
			}
			for _, v := range vec.Vec_IntArray(neg) {
				gia.Gia_ManAppendCo(g, v)
			}
			bound := int32(1) << n
			for x := int32(0); x < bound; x++ {
				for y := int32(0); y < bound; y++ {
					inputs := make([]bool, 2*n)
					for i := int32(0); i < n; i++ {
						inputs[i] = x&(1<<i) != 0
						inputs[n+i] = y&(1<<i) != 0
					}
					got, err := g.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					sx, sy := x, y
					if signed != 0 {
						if x&(bound>>1) != 0 {
							sx -= bound
						}
						if y&(bound>>1) != 0 {
							sy -= bound
						}
					}
					for i := int32(0); i < 2*n; i++ {
						if got[i] != ((sx*sy)&(1<<i) != 0) {
							t.Fatalf("width=%d signed=%d x=%d y=%d product bit=%d", n, signed, x, y, i)
						}
					}
					for i := int32(0); i < n; i++ {
						if got[2*n+i] != ((x*y)&(1<<i) != 0) || got[3*n+i] != ((-x)&(1<<i) != 0) {
							t.Fatalf("width=%d x=%d y=%d low bit=%d", n, x, y, i)
						}
					}
				}
			}
		}
	}
}

// The original C signed multiplier can differ from mathematical signed
// multiplication for unequal widths, so these vectors preserve its output.
func TestBlastMultiplierMatchesCForUnequalWidths(t *testing.T) {
	data, err := os.Open("test_vectors/multiply_small.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
	var na, nb, sign, x, y, want int32
	for {
		_, err = fmt.Fscan(data, &na, &nb, &sign, &x, &y, &want)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		g := gia.Gia_ManStart(100)
		gia.Gia_ManHashAlloc(g)
		a, b := make([]int32, na), make([]int32, nb)
		for i := range a {
			a[i] = (x >> i) & 1
		}
		for i := range b {
			b[i] = (y >> i) & 1
		}
		tmp, out := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
		Wlc_BlastMultiplier(g, a, b, na, nb, tmp, out, sign)
		var got int32
		for i, bit := range vec.Vec_IntArray(out) {
			if bit > 1 {
				t.Fatal("expected constant output")
			}
			got |= bit << i
		}
		if got != want {
			t.Fatalf("widths=%d/%d signed=%d x=%d y=%d: got %d want C %d", na, nb, sign, x, y, got, want)
		}
		gia.Gia_ManStop(g)
	}
}
