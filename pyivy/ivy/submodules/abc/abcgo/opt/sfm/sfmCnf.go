package sfm

import (
	"fmt"
	"github.com/glycerine/abcgo/bool/kit"
)

// Sfm_TruthToCnf derives both ISOP covers and emits implications to the local
// output variable nVars. Clauses use ordinary var<<1|negative literals.
func Sfm_TruthToCnf(truth uint64, large []uint64, vars int) [][]int {
	if vars < 0 || vars > 16 {
		panic("sfm: invalid truth support")
	}
	words := []uint64{stretch(truth, vars)}
	if vars > 6 {
		if len(large) < truthWords(vars) {
			panic("sfm: short truth")
		}
		words = large[:truthWords(vars)]
	}
	all0, all1 := true, true
	for _, w := range words {
		all0 = all0 && w == 0
		all1 = all1 && w == ^uint64(0)
	}
	if all0 || all1 {
		lit := vars << 1
		if all0 {
			lit ^= 1
		}
		return [][]int{{lit}}
	}
	var clauses [][]int
	for c := 0; c < 2; c++ {
		t := make([]uint32, 2*len(words))
		for i, w := range words {
			if c != 0 {
				w = ^w
			}
			t[2*i], t[2*i+1] = uint32(w), uint32(w>>32)
		}
		var cover []uint32
		if kit.Kit_TruthIsop(t, int32(vars), &cover, 0) != 0 {
			panic("sfm: ISOP failed")
		}
		for _, cube := range cover {
			var clause []int
			for v := 0; v < vars; v++ {
				literal := (cube >> uint(2*v)) & 3
				if literal == 1 {
					clause = append(clause, v<<1)
				} else if literal == 2 {
					clause = append(clause, v<<1|1)
				}
			}
			clause = append(clause, vars<<1|c)
			clauses = append(clauses, clause)
		}
	}
	return clauses
}
func Sfm_TranslateCnf(cnf [][]int, faninMap []int, pivot int) [][]int {
	r := make([][]int, len(cnf))
	for i, c := range cnf {
		r[i] = make([]int, len(c))
		for j, l := range c {
			v := faninMap[l>>1]
			r[i][j] = v<<1 | (l & 1)
			if v == pivot {
				r[i][j] ^= 1
			}
		}
	}
	return r
}
func Sfm_CreateCnf(p *Sfm_Ntk_t) [][][]int {
	r := make([][][]int, len(p.Fanins))
	for id := p.Pis; id < len(p.Fanins)-p.Pos; id++ {
		if !p.Deleted[id] {
			r[id] = Sfm_TruthToCnf(p.Truths[id][0], p.Truths[id], len(p.Fanins[id]))
		}
	}
	return r
}
func Sfm_PrintCnf(cnf [][]int) {
	for _, c := range cnf {
		for _, l := range c {
			if l&1 != 0 {
				fmt.Print("-")
			}
			fmt.Printf("%d ", l>>1)
		}
		fmt.Println()
	}
}
