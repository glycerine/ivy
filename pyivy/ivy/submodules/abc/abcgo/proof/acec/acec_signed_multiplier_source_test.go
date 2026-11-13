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

// A 4x4 signed product is built by sign-extending operands to eight bits and
// retaining the low eight product bits. This mirrors the compiled C fixture.
func sourceSignedMultiplier4(swap bool, flip int) *gia.Gia_Man_t {
	p := gia.Gia_ManStart(1000)
	var a, b [8]int32
	for i := 0; i < 4; i++ {
		a[i] = gia.Gia_ManAppendCi(p)
	}
	for i := 0; i < 4; i++ {
		b[i] = gia.Gia_ManAppendCi(p)
	}
	for i := 4; i < 8; i++ {
		a[i] = a[3]
		b[i] = b[3]
	}
	var columns [8][]int32
	for i := 0; i < 8; i++ {
		for j := 0; i+j < 8; j++ {
			x, y := a[i], b[j]
			if swap {
				x, y = b[j], a[i]
			}
			columns[i+j] = append(columns[i+j], gia.Gia_ManAppendAnd(p, x, y))
		}
	}
	for k := 0; k < 8; k++ {
		for len(columns[k]) >= 3 {
			in := columns[k][:3]
			columns[k] = columns[k][3:]
			var out [2]int32
			Acec_InsertFadd(p, in, out[:])
			columns[k] = append(columns[k], out[0])
			if k < 7 {
				columns[k+1] = append(columns[k+1], out[1])
			}
		}
		if len(columns[k]) == 2 {
			x, y := columns[k][0], columns[k][1]
			columns[k] = []int32{gia.Gia_ManAppendXor(p, x, y)}
			if k < 7 {
				columns[k+1] = append(columns[k+1], gia.Gia_ManAppendAnd(p, x, y))
			}
		}
	}
	for k := 0; k < 8; k++ {
		out := int32(0)
		if len(columns[k]) != 0 {
			out = columns[k][0]
		}
		if k == flip {
			out ^= 1
		}
		gia.Gia_ManAppendCo(p, out)
	}
	return p
}

func TestSourceSignedMultiplierTruth(t *testing.T) {
	for _, swap := range []bool{false, true} {
		p := sourceSignedMultiplier4(swap, -1)
		for a := 0; a < 16; a++ {
			for b := 0; b < 16; b++ {
				inputs := make([]bool, 8)
				for i := 0; i < 4; i++ {
					inputs[i] = a&(1<<i) != 0
					inputs[i+4] = b&(1<<i) != 0
				}
				av, bv := a, b
				if av&8 != 0 {
					av -= 16
				}
				if bv&8 != 0 {
					bv -= 16
				}
				want := uint8(av * bv)
				got, err := p.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				for i, v := range got {
					if v != (want&(1<<i) != 0) {
						t.Fatalf("swap=%v %d*%d bit %d got %v", swap, av, bv, i, got)
					}
				}
			}
		}
	}
}

func TestSourceSignedMultiplierAcec(t *testing.T) {
	for _, tc := range []struct {
		flip int
		want int32
	}{{-1, 1}, {0, 0}, {4, 0}, {6, 1}} {
		t.Run(fmt.Sprintf("flip_%d", tc.flip), func(t *testing.T) {
			a, b := sourceSignedMultiplier4(false, -1), sourceSignedMultiplier4(true, tc.flip)
			var pars Acec_ParCec_t
			Acec_ManCecSetDefaultParams(&pars)
			got := Acec_Solve(a, b, &pars)
			if got != tc.want {
				t.Fatalf("status %d; compiled C %d", got, tc.want)
			}
			check := func(inputs []bool) {
				t.Helper()
				x, e1 := a.Evaluate(inputs)
				y, e2 := b.Evaluate(inputs)
				if e1 != nil || e2 != nil {
					t.Fatalf("evaluate %v %v", e1, e2)
				}
				for i := range x {
					if x[i] != y[i] {
						return
					}
				}
				t.Fatalf("invalid witness %v", inputs)
			}
			if got == 0 {
				w := a.CombinationalCounterexample()
				if w == nil {
					t.Fatal("missing Go witness")
				}
				in := make([]bool, 8)
				for i := range in {
					in[i] = util.Abc_InfoHasBit(w.PData, int32(i))
				}
				check(in)
			}
			if binary := os.Getenv("ACEC_SIGNED_MULTIPLIER_ORACLE"); binary != "" {
				data, err := exec.Command(binary, strconv.Itoa(tc.flip)).CombinedOutput()
				if err != nil {
					t.Fatalf("C oracle %v: %s", err, data)
				}
				if !strings.Contains(string(data), fmt.Sprintf("SOURCE_STATUS=%d\n", tc.want)) {
					t.Fatalf("C status: %s", data)
				}
				if got == 0 {
					_, line, ok := strings.Cut(string(data), "SOURCE_WITNESS=")
					if !ok {
						t.Fatalf("missing C witness: %s", data)
					}
					line, _, _ = strings.Cut(line, "\n")
					if len(line) != 8 {
						t.Fatalf("invalid C witness %q", line)
					}
					in := make([]bool, 8)
					for i, c := range line {
						if c != '0' && c != '1' {
							t.Fatalf("invalid C witness %q", line)
						}
						in[i] = c == '1'
					}
					check(in)
				}
			}
		})
	}
}
