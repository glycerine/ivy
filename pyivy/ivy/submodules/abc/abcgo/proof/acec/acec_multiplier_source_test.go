package acec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// sourceMultiplier3 constructs a 3x3 unsigned product using partial-product
// columns, full adders, and half adders without relying on WLC.
func sourceMultiplier3(swapped bool, invertOutput int) *gia.Gia_Man_t {
	return sourceMultiplier(3, swapped, invertOutput)
}

func sourceMultiplier(width int, swapped bool, invertOutput int) *gia.Gia_Man_t {
	p := gia.Gia_ManStart(800)
	a, b := make([]int32, width), make([]int32, width)
	for i := range a {
		a[i] = gia.Gia_ManAppendCi(p)
	}
	for i := range b {
		b[i] = gia.Gia_ManAppendCi(p)
	}
	columns := make([][]int32, 2*width)
	for i := 0; i < width; i++ {
		for j := 0; j < width; j++ {
			x, y := a[i], b[j]
			if swapped {
				x, y = y, x
			}
			columns[i+j] = append(columns[i+j], gia.Gia_ManAppendAnd(p, x, y))
		}
	}
	for k := 0; k < 2*width-1; k++ {
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
	for k := 0; k < 2*width; k++ {
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
			if binary := os.Getenv("ACEC_MULTIPLIER_ORACLE"); binary != "" {
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

func TestSourceMultiplier4Truth(t *testing.T) { testSourceMultiplierTruth(t, 4) }
func TestSourceMultiplier5Truth(t *testing.T) { testSourceMultiplierTruth(t, 5) }

func testSourceMultiplierTruth(t *testing.T, width int) {
	t.Helper()
	for _, swap := range []bool{false, true} {
		p := sourceMultiplier(width, swap, -1)
		for a := 0; a < 1<<width; a++ {
			for b := 0; b < 1<<width; b++ {
				in := make([]bool, 2*width)
				for i := 0; i < width; i++ {
					in[i] = a&(1<<i) != 0
					in[width+i] = b&(1<<i) != 0
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

// A mismatch in the second-to-last output exposes the source's discarded-output behavior;
// a mismatch in output 0 instead supplies a replayable witness.
func TestSourceMultiplier4Acec(t *testing.T) { testSourceMultiplierAcec(t, 4) }
func TestSourceMultiplier5Acec(t *testing.T) { testSourceMultiplierAcec(t, 5) }

func testSourceMultiplierAcec(t *testing.T, width int) {
	t.Helper()
	for _, tc := range []struct {
		flip   int
		status int32
	}{{-1, 1}, {0, 0}, {width, 0}, {2*width - 2, 1}} {
		t.Run(fmt.Sprintf("flip_%d", tc.flip), func(t *testing.T) {
			a, b := sourceMultiplier(width, false, -1), sourceMultiplier(width, true, tc.flip)
			var pars Acec_ParCec_t
			Acec_ManCecSetDefaultParams(&pars)
			if got := Acec_Solve(a, b, &pars); got != tc.status {
				t.Fatalf("Go status=%d source=%d", got, tc.status)
			}
			checkWitness := func(in []bool) {
				t.Helper()
				x, e1 := a.Evaluate(in)
				y, e2 := b.Evaluate(in)
				if e1 != nil || e2 != nil {
					t.Fatalf("witness evaluation: %v / %v", e1, e2)
				}
				different := false
				for i := range x {
					different = different || x[i] != y[i]
				}
				if !different {
					t.Fatalf("invalid witness %v", in)
				}
			}
			if tc.status == 0 {
				cex := a.CombinationalCounterexample()
				if cex == nil {
					t.Fatal("missing Go witness")
				}
				in := make([]bool, 2*width)
				for i := range in {
					in[i] = util.Abc_InfoHasBit(cex.PData, int32(i))
				}
				checkWitness(in)
			}
			if tc.flip == 2*width-2 {
				x, _ := a.Evaluate(make([]bool, 2*width))
				y, _ := b.Evaluate(make([]bool, 2*width))
				if x[tc.flip] == y[tc.flip] {
					t.Fatal("discarded output should differ")
				}
			}
			if binary := os.Getenv("ACEC_MULTIPLIER_ORACLE"); binary != "" {
				output, err := exec.Command(binary, strconv.Itoa(tc.flip), strconv.Itoa(width)).CombinedOutput()
				if err != nil {
					t.Fatalf("C oracle: %v: %s", err, output)
				}
				if !strings.Contains(string(output), fmt.Sprintf("SOURCE_STATUS=%d\n", tc.status)) {
					t.Fatalf("C status mismatch: %s", output)
				}
				if tc.status == 0 {
					_, line, ok := strings.Cut(string(output), "SOURCE_WITNESS=")
					if !ok {
						t.Fatalf("missing C witness: %s", output)
					}
					line, _, _ = strings.Cut(line, "\n")
					if len(line) != 2*width {
						t.Fatalf("malformed C witness %q", line)
					}
					in := make([]bool, 2*width)
					for i, c := range line {
						if c != '0' && c != '1' {
							t.Fatalf("malformed C witness %q", line)
						}
						in[i] = c == '1'
					}
					checkWitness(in)
				}
			}
		})
	}
}
