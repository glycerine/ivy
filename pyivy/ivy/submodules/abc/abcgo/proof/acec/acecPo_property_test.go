package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"testing"
)

func TestPolynomialBuildersMixedDagAgainstSimulation(t *testing.T) {
	rng := rand.New(rand.NewSource(17))
	for graph := 0; graph < 30; graph++ {
		p := gia.Gia_ManStart(64)
		lits := make([]int32, 4)
		leaves := vec.Vec_IntAlloc(4)
		for i := range lits {
			lits[i] = gia.Gia_ManAppendCi(p)
			vec.Vec_IntPush(leaves, lits[i]>>1)
		}
		nodes := vec.Vec_IntAlloc(8)
		for i := 0; i < 8; i++ {
			a := rng.Intn(len(lits))
			b := rng.Intn(len(lits) - 1)
			if b >= a {
				b++
			}
			x, y := lits[a]^int32(rng.Intn(2)), lits[b]^int32(rng.Intn(2))
			var lit int32
			if rng.Intn(2) == 0 {
				lit = gia.Gia_ManAppendAnd(p, x, y)
			} else {
				lit = gia.Gia_ManAppendXorReal(p, x, y)
			}
			lits = append(lits, lit)
			vec.Vec_IntPush(nodes, lit>>1)
		}
		roots := vec.Vec_IntAlloc(3)
		for i := 0; i < 3; i++ {
			lit := lits[4+rng.Intn(8)] ^ int32(rng.Intn(2))
			vec.Vec_IntPush(roots, lit)
			gia.Gia_ManAppendCo(p, lit)
		}
		obj := Gia_PolynBuildNew(p, nil, roots, 0, leaves, nodes, 0, 0, 0)
		lit := Gia_PolynBuildNew2(p, roots, 0, leaves, nodes, 0, 0, 0)
		for mask := 0; mask < 16; mask++ {
			inputs := make([]bool, 4)
			objVals, litVals := map[int32]bool{}, map[int32]bool{}
			for i := range inputs {
				inputs[i] = mask&(1<<i) != 0
				objVals[lits[i]>>1] = inputs[i]
				litVals[lits[i]] = inputs[i]
			}
			outputs, err := p.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			var want int64
			for i, v := range outputs {
				if v {
					want += 1 << i
				}
			}
			if got := polynomialValue(obj, objVals); got != want {
				for _, id := range vec.Vec_IntArray(nodes) {
					a, b, c, k := p.Gate(id)
					t.Logf("node %d: %d %d %d kind %d", id, a, b, c, k)
				}
				t.Fatalf("graph %d input %d object got %d want %d roots %v nodes %v polynomial %v outputs %v", graph, mask, got, want, vec.Vec_IntArray(roots), vec.Vec_IntArray(nodes), poRows(obj), outputs)
			}
			if got := polynomialValue(lit, litVals); got != want {
				t.Fatalf("graph %d input %d literal got %d want %d", graph, mask, got, want)
			}
		}
	}
}
