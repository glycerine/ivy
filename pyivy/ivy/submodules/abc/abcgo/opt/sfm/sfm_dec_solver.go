package sfm

import "github.com/glycerine/abcgo/sat/satoko"

// sfmDecWindow uses local object indices, with the pivot ending the TFI.
type sfmDecWindow struct {
	gates  []int
	fanins [][]int
	roots  []int
	pivot  int
}

func (w *sfmDecWindow) prepareSolver(gateCnfs [][][]int) (*satoko.Solver, bool) {
	n := len(w.gates)
	tfi := w.pivot + 1
	if w.pivot < 0 || tfi > n || len(w.fanins) != n {
		panic("sfm: invalid decomposition window")
	}
	satVars := 2*n - tfi
	solver := satoko.New()
	solver.SetVariables(satVars + len(w.roots))
	add := func(gate int, inputs []int, pivot int) bool {
		if gate < 0 || gate >= len(gateCnfs) {
			panic("sfm: invalid gate")
		}
		for _, clause := range Sfm_TranslateCnf(gateCnfs[gate], inputs, pivot) {
			if !solver.AddClause(clause) {
				return false
			}
		}
		return true
	}
	for id, gate := range w.gates {
		if gate < 0 {
			continue
		}
		inputs := append(append([]int(nil), w.fanins[id]...), id)
		if !add(gate, inputs, -1) {
			return nil, false
		}
	}
	for id := tfi; id < n; id++ {
		inputs := make([]int, 0, len(w.fanins[id])+1)
		for _, fan := range w.fanins[id] {
			if fan > w.pivot {
				fan += n - tfi
			}
			inputs = append(inputs, fan)
		}
		inputs = append(inputs, id+n-tfi)
		if !add(w.gates[id], inputs, w.pivot) {
			return nil, false
		}
	}
	if tfi < n {
		clause := make([]int, 0, len(w.roots))
		for _, root := range w.roots {
			a, b, z := root*2, (root+n-tfi)*2, satVars*2
			satVars++
			for _, c := range [][]int{{a, b, z ^ 1}, {a ^ 1, b ^ 1, z ^ 1}, {a, b ^ 1, z}, {a ^ 1, b, z}} {
				if !solver.AddClause(c) {
					return nil, false
				}
			}
			clause = append(clause, z)
		}
		if !solver.AddClause(clause) {
			return nil, false
		}
	} else if len(w.roots) != 1 {
		panic("sfm: expected single window root")
	}
	return solver, true
}
