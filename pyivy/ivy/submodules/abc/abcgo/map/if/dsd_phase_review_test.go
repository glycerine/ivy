package if_

import (
	"os"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestDsdBalanceRetainsMixedOutputPhases(t *testing.T) {
	// Mapping's DSD cache is keyed by the regular truth-table ID. An entry
	// computed from a complemented cut must not bake in that cut's phase.
	progress, err := os.CreateTemp(t.TempDir(), "mapping-progress")
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = progress
	defer func() { os.Stdout = original; progress.Close() }()
	for phase := int32(0); phase < 4; phase++ {
		src := gia.Gia_ManStart(32)
		a, b := gia.Gia_ManAppendCi(src), gia.Gia_ManAppendCi(src)
		gia.Gia_ManHashStart(src)
		x := gia.Gia_ManHashXor(src, a^(phase&1), b^(phase>>1&1))
		for _, lit := range []int32{x, x ^ 1, gia.Gia_ManHashAnd(src, a, b), gia.Gia_ManHashOr(src, a, b)} {
			gia.Gia_ManAppendCo(src, lit)
		}
		gia.Gia_ManHashStop(src)
		out := Gia_ManPerformDsdBalance(src, 6, 8, 0, 0)
		for mask := 0; mask < 4; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0}
			want, err := src.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			got, err := out.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, want) {
				t.Fatalf("phase%d assignment%d: got%v want%v", phase, mask, got, want)
			}
		}
	}
}

func TestDsdBalanceExhaustiveSmallFunctions(t *testing.T) {
	progress, err := os.CreateTemp(t.TempDir(), "mapping-progress")
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = progress
	defer func() { os.Stdout = original; progress.Close() }()
	cases, assignments := 0, 0
	for n := 2; n <= 4; n++ {
		var functions []uint64
		if n < 4 {
			for table := uint64(0); table < 1<<uint(1<<n); table++ {
				functions = append(functions, table)
			}
		} else {
			functions = []uint64{0, 0xffff, 0xaaaa, 0x5555, 0xcccc, 0x3333, 0xf0f0, 0x0f0f, 0xff00, 0x00ff, 0x6996, 0x9669, 0x8000, 0xfffe, 0xcaac, 0x9e79}
		}
		for _, table := range functions {
			src := gia.Gia_ManStart(128)
			cis := make([]int32, n)
			for i := range cis {
				cis[i] = gia.Gia_ManAppendCi(src)
			}
			gia.Gia_ManHashStart(src)
			var root int32
			for assignment := 0; assignment < 1<<n; assignment++ {
				if table>>assignment&1 == 0 {
					continue
				}
				term := cis[0] ^ int32(1-(assignment&1))
				for pin := 1; pin < n; pin++ {
					term = gia.Gia_ManHashAnd(src, term, cis[pin]^int32(1-(assignment>>pin&1)))
				}
				root = gia.Gia_ManHashOr(src, root, term)
			}
			for _, lit := range []int32{root, root ^ 1, gia.Gia_ManHashAnd(src, cis[0], cis[1]), gia.Gia_ManHashOr(src, cis[0], cis[1])} {
				gia.Gia_ManAppendCo(src, lit)
			}
			gia.Gia_ManHashStop(src)
			out := Gia_ManPerformDsdBalance(src, 6, 8, 0, 0)
			for assignment := 0; assignment < 1<<n; assignment++ {
				inputs := make([]bool, n)
				for i := range inputs {
					inputs[i] = assignment>>i&1 != 0
				}
				value := table>>assignment&1 != 0
				want := []bool{value, !value, inputs[0] && inputs[1], inputs[0] || inputs[1]}
				for _, g := range []*gia.Gia_Man_t{src, out} {
					got, err := g.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					if !slices.Equal(got, want) {
						t.Fatalf("vars%d table%#x assignment%d: got%v want%v", n, table, assignment, got, want)
					}
				}
				assignments++
			}
			cases++
		}
	}
	t.Logf("%d functions, %d assignments: all two/three-variable functions and 16 four-variable patterns", cases, assignments)
}
