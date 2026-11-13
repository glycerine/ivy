package acec

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// sourceRectMultiplier constructs the exact 3x5 unsigned multiplier used by
// the C oracle, retaining all eight product bits.
func sourceRectMultiplier(swap bool, flip int) *gia.Gia_Man_t {
	p := gia.Gia_ManStart(800)
	var a [3]int32
	var b [5]int32
	for i := range a {
		a[i] = gia.Gia_ManAppendCi(p)
	}
	for i := range b {
		b[i] = gia.Gia_ManAppendCi(p)
	}
	var columns [8][]int32
	for i := range a {
		for j := range b {
			x, y := a[i], b[j]
			if swap {
				x, y = y, x
			}
			if flip == -2 && i == 0 && j == 0 {
				y ^= 1
			}
			columns[i+j] = append(columns[i+j], gia.Gia_ManAppendAnd(p, x, y))
		}
	}
	for k := 0; k < 7; k++ {
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
	for k := 0; k < 8; k++ {
		lit := int32(0)
		if len(columns[k]) != 0 {
			lit = columns[k][0]
		}
		if k == flip {
			lit ^= 1
		}
		gia.Gia_ManAppendCo(p, lit)
	}
	return p
}

func TestSourceRectMultiplierTruth(t *testing.T) {
	for _, swap := range []bool{false, true} {
		p := sourceRectMultiplier(swap, -1)
		for a := 0; a < 8; a++ {
			for b := 0; b < 32; b++ {
				in := make([]bool, 8)
				for i := 0; i < 3; i++ {
					in[i] = a&(1<<i) != 0
				}
				for i := 0; i < 5; i++ {
					in[i+3] = b&(1<<i) != 0
				}
				out, err := p.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				for i, v := range out {
					if v != ((a*b)&(1<<i) != 0) {
						t.Fatalf("swap=%v %d*%d bit %d: %v", swap, a, b, i, out)
					}
				}
			}
		}
	}
}

func TestSourceRectMultiplierAcec(t *testing.T) {
	for _, tc := range []struct {
		flip   int
		status int32
	}{{-2, 0}, {-1, 1}, {0, 0}, {4, 0}, {6, 1}} {
		t.Run(fmt.Sprintf("flip_%d", tc.flip), func(t *testing.T) {
			a, b := sourceRectMultiplier(false, -1), sourceRectMultiplier(true, tc.flip)
			var p Acec_ParCec_t
			Acec_ManCecSetDefaultParams(&p)
			got := Acec_Solve(a, b, &p)
			if got != tc.status {
				t.Fatalf("status %d, C %d", got, tc.status)
			}
			if tc.flip == -2 {
				x, _ := a.Evaluate(make([]bool, 8))
				y, _ := b.Evaluate(make([]bool, 8))
				for i := range x {
					if x[i] != y[i] {
						t.Fatal("partial-product mutation must agree on all-zero input")
					}
				}
			}
			replay := func(in []bool) {
				t.Helper()
				x, e1 := a.Evaluate(in)
				y, e2 := b.Evaluate(in)
				if e1 != nil || e2 != nil {
					t.Fatalf("evaluate %v %v", e1, e2)
				}
				for i := range x {
					if x[i] != y[i] {
						return
					}
				}
				t.Fatalf("invalid witness %v", in)
			}
			if got == 0 {
				cex := a.CombinationalCounterexample()
				if cex == nil {
					t.Fatal("missing Go witness")
				}
				in := make([]bool, 8)
				for i := range in {
					in[i] = util.Abc_InfoHasBit(cex.PData, int32(i))
				}
				replay(in)
			}
			if tc.flip == 6 {
				x, _ := a.Evaluate(make([]bool, 8))
				y, _ := b.Evaluate(make([]bool, 8))
				if x[6] == y[6] {
					t.Fatal("source false-equivalence case must differ")
				}
			}
			if binary := os.Getenv("ACEC_RECT_MULTIPLIER_ORACLE"); binary != "" {
				out, err := exec.Command(binary, strconv.Itoa(tc.flip)).CombinedOutput()
				if err != nil {
					t.Fatalf("C oracle %v: %s", err, out)
				}
				if !strings.Contains(string(out), fmt.Sprintf("SOURCE_STATUS=%d\n", tc.status)) {
					t.Fatalf("C result %s", out)
				}
				if got == 0 {
					_, line, ok := strings.Cut(string(out), "SOURCE_WITNESS=")
					if !ok {
						t.Fatalf("missing C witness: %s", out)
					}
					line, _, _ = strings.Cut(line, "\n")
					if len(line) != 8 {
						t.Fatalf("C witness %q", line)
					}
					in := make([]bool, 8)
					for i, c := range line {
						if c != '0' && c != '1' {
							t.Fatalf("C witness %q", line)
						}
						in[i] = c == '1'
					}
					replay(in)
				}
			}
		})
	}
}
