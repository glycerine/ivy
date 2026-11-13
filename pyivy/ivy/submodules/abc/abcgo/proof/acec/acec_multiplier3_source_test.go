package acec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// sourceMultiplier3 constructs a 3x3 unsigned product using partial-product
// columns, full adders, and half adders without relying on WLC.
func sourceMultiplier3(swapped bool, invertOutput int) *gia.Gia_Man_t {
	p := gia.Gia_ManStart(400)
	var a, b [3]int32
	for i := range a {
		a[i] = gia.Gia_ManAppendCi(p)
	}
	for i := range b {
		b[i] = gia.Gia_ManAppendCi(p)
	}
	var columns [6][]int32
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			x, y := a[i], b[j]
			if swapped {
				x, y = y, x
			}
			columns[i+j] = append(columns[i+j], gia.Gia_ManAppendAnd(p, x, y))
		}
	}
	for k := 0; k < 5; k++ {
		for len(columns[k]) >= 3 {
			in := columns[k][:3]
			columns[k] = columns[k][3:]
			var out [2]int32
			Acec_InsertFadd(p, in, out[:])
			columns[k] = append(columns[k], out[0])
			columns[k+1] = append(columns[k+1], out[1])
		}
		if len(columns[k]) == 2 {
			x, y := columns[k][0], columns[k][1]
			columns[k] = []int32{gia.Gia_ManAppendXor(p, x, y)}
			columns[k+1] = append(columns[k+1], gia.Gia_ManAppendAnd(p, x, y))
		}
	}
	for k := 0; k < 6; k++ {
		out := int32(0)
		if len(columns[k]) > 0 {
			out = columns[k][0]
		}
		if k == invertOutput {
			out ^= 1
		}
		gia.Gia_ManAppendCo(p, out)
	}
	return p
}
func TestSourceMultiplier3Truth(t *testing.T) {
	for _, swap := range []bool{false, true} {
		p := sourceMultiplier3(swap, -1)
		for a := 0; a < 8; a++ {
			for b := 0; b < 8; b++ {
				in := make([]bool, 6)
				for i := 0; i < 3; i++ {
					in[i] = a&(1<<i) != 0
					in[3+i] = b&(1<<i) != 0
				}
				out, err := p.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				for i, v := range out {
					if v != ((a*b)&(1<<i) != 0) {
						t.Fatalf("swap=%t a=%d b=%d out=%v", swap, a, b, out)
					}
				}
			}
		}
	}
}

// The source fixture records an inherited false equivalence when bit 4 is
// inverted: Acec_Solve discards its last two outputs after matching boxes.
func TestSourceMultiplier3Acec(t *testing.T) {
	for _, tc := range []struct {
		flip   int
		status int32
	}{{-1, 1}, {0, 0}, {4, 1}} {
		t.Run(fmt.Sprintf("flip_%d", tc.flip), func(t *testing.T) {
			a, b := sourceMultiplier3(false, -1), sourceMultiplier3(true, tc.flip)
			inputs := make([]bool, 6)
			x, e1 := a.Evaluate(inputs)
			y, e2 := b.Evaluate(inputs)
			if e1 != nil || e2 != nil {
				t.Fatalf("evaluate: %v / %v", e1, e2)
			}
			if tc.flip >= 0 && x[tc.flip] == y[tc.flip] {
				t.Fatal("inverted output is not different")
			}
			var p Acec_ParCec_t
			Acec_ManCecSetDefaultParams(&p)
			if got := Acec_Solve(a, b, &p); got != tc.status {
				t.Fatalf("Go status=%d, source status=%d", got, tc.status)
			}
			if binary := os.Getenv("ACEC_MULTIPLIER3_ORACLE"); binary != "" {
				output, err := exec.Command(binary, strconv.Itoa(tc.flip)).CombinedOutput()
				if err != nil {
					t.Fatalf("C oracle: %v: %s", err, output)
				}
				want := fmt.Sprintf("SOURCE_STATUS=%d\n", tc.status)
				if !strings.Contains(string(output), want) {
					t.Fatalf("C oracle missing %q: %s", want, output)
				}
			}
		})
	}
}
