package sbd

import (
	"fmt"
	"slices"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

var sbdTruths6 = [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}

func sbdEnsureVariables(solver *bsat.Solver, count int32) {
	for solver.NumVars() < count {
		solver.NewVariable()
	}
}
func sbdModelValue(solver *bsat.Solver, variable int32) int32 {
	value, ok := solver.Value(variable)
	if !ok {
		panic("SAT model value unavailable")
	}
	if value {
		return 1
	}
	return 0
}
func sbdModelLiteral(solver *bsat.Solver, variable int32) bsat.Literal {
	return bsat.Literal(2*variable + 1 - sbdModelValue(solver, variable))
}
func sbdAnd(solver *bsat.Solver, out, a, b bsat.Literal) bool {
	return solver.AddClause(out^1, a) && solver.AddClause(out^1, b) && solver.AddClause(out, a^1, b^1)
}
func sbdXor(solver *bsat.Solver, out, a, b bsat.Literal) bool {
	return solver.AddClause(a, b, out^1) && solver.AddClause(a^1, b^1, out^1) && solver.AddClause(a, b^1, out) && solver.AddClause(a^1, b, out)
}
func Sbd_ManSatSolver(solver *bsat.Solver, network *Network, mirrors *vec.Vec_Int_t, pivot int32, window, objToVar, tfo, roots *vec.Vec_Int_t, qbf int32) *bsat.Solver {
	count, tail, rootCount := int32(vec.Vec_IntSize(window)), int32(vec.Vec_IntSize(tfo)), int32(vec.Vec_IntSize(roots))
	start := count - tail
	pivotVar := vec.Vec_IntEntry(objToVar, int(pivot))
	last := count + tail + rootCount
	if count < 1 || vec.Vec_IntEntry(window, 0) != 0 || start < 1 {
		panic("invalid SBD SAT window")
	}
	if solver == nil {
		solver = bsat.NewSolver()
	} else {
		solver.Reset()
	}
	sbdEnsureVariables(solver, last+SBD_FVAR_MAX)
	if !solver.AddClause(bsat.Negative(0)) {
		return nil
	}
	encode := func(id int32, second bool) bool {
		node := network.object(id)
		if node.Kind == aig.Input {
			return true
		}
		if mirrors != nil && vec.Vec_IntEntry(mirrors, int(id)) >= 0 {
			panic("window contains mirrored node")
		}
		variable := vec.Vec_IntEntry(objToVar, int(id))
		if second {
			variable += tail
		}
		input := func(edge aig.Edge) bsat.Literal {
			literal := sourceLiteral(edge)
			if mirrors != nil {
				if replacement := vec.Vec_IntEntry(mirrors, edge.NodeID()); replacement >= 0 {
					literal = replacement ^ (literal & 1)
				}
			}
			v := vec.Vec_IntEntry(objToVar, int(literal>>1))
			if second {
				if v >= start {
					v += tail
				}
				if qbf != 0 && v == pivotVar {
					v = last
				}
				if qbf == 0 && v == pivotVar {
					literal ^= 1
				}
			}
			return bsat.Literal(2*v + (literal & 1))
		}
		output := bsat.Positive(variable)
		switch node.Kind {
		case aig.And:
			return sbdAnd(solver, output, input(node.Fanins[0]), input(node.Fanins[1]))
		case aig.ExclusiveOr:
			return sbdXor(solver, output, input(node.Fanins[0]), input(node.Fanins[1]))
		case aig.Buffer:
			return solver.AddClause(output^1, input(node.Fanins[0])) && solver.AddClause(output, input(node.Fanins[0])^1)
		default:
			panic("window contains unsupported node")
		}
	}
	for _, id := range vec.Vec_IntArray(window)[1:] {
		if !encode(id, false) {
			return nil
		}
	}
	for _, id := range vec.Vec_IntArray(window)[start:] {
		if !encode(id, true) {
			return nil
		}
	}
	if rootCount > 0 {
		differences := make([]bsat.Literal, rootCount)
		for i, id := range vec.Vec_IntArray(roots) {
			variable := vec.Vec_IntEntry(objToVar, int(id))
			diff := count + tail + int32(i)
			differences[i] = bsat.Positive(diff)
			if !sbdXor(solver, differences[i], bsat.Positive(variable), bsat.Positive(variable+tail)) {
				return nil
			}
		}
		if !solver.AddClause(differences...) {
			return nil
		}
	} else if qbf != 0 {
		if !solver.AddClause(bsat.Positive(pivotVar), bsat.Positive(last)) || !solver.AddClause(bsat.Negative(pivotVar), bsat.Negative(last)) {
			return nil
		}
	}
	if !solver.PropagateUnits() {
		return nil
	}
	return solver
}
func Sbd_ManSolve(solver *bsat.Solver, pivot, free int32, divSet, divVars, divValues, temp *vec.Vec_Int_t) uint64 {
	if solver == nil || free >= solver.NumVars() || vec.Vec_IntSize(divSet) > 6 || vec.Vec_IntSize(divVars) != vec.Vec_IntSize(divValues) {
		panic("invalid SBD solve dimensions")
	}
	onset, enable := bsat.Positive(pivot), bsat.Positive(free)
	truth := uint64(0)
	for {
		status := solver.Solve(onset, enable)
		if status == bsat.Unknown {
			return SBD_SAT_UNDEC
		}
		if status == bsat.Unsatisfiable {
			return truth
		}
		for i, v := range vec.Vec_IntArray(divVars) {
			vec.Vec_IntWriteEntry(divValues, i, 2*sbdModelValue(solver, v))
		}
		assumptions := []bsat.Literal{onset ^ 1}
		for _, v := range vec.Vec_IntArray(divSet) {
			assumptions = append(assumptions, sbdModelLiteral(solver, v))
		}
		status = solver.Solve(assumptions...)
		if status == bsat.Unknown {
			return SBD_SAT_UNDEC
		}
		if status == bsat.Satisfiable {
			for i, v := range vec.Vec_IntArray(divVars) {
				vec.Vec_IntAddToEntry(divValues, i, sbdModelValue(solver, v)+12)
			}
			return SBD_SAT_SAT
		}
		cube := ^uint64(0)
		vec.Vec_IntClear(temp)
		vec.Vec_IntPush(temp, int32(enable^1))
		for _, literal := range solver.UnsatCore() {
			if literal == onset {
				continue
			}
			index := slices.Index(vec.Vec_IntArray(divSet), int32(literal>>1))
			if index < 0 {
				panic("SBD core references a non-divisor")
			}
			vec.Vec_IntPush(temp, int32(literal))
			mask := sbdTruths6[index]
			if literal&1 == 0 {
				mask = ^mask
			}
			cube &= mask
		}
		truth |= cube
		if !solver.AddClause(vec.Vec_IntArray(temp)...) {
			panic("SBD blocking cube made base inconsistent")
		}
	}
}
func Sbd_ManSolve2(solver *bsat.Solver, pivot, free int32, divVars, divValues, temp, sop *vec.Vec_Int_t) int32 {
	if free >= solver.NumVars() || vec.Vec_IntSize(divVars) != vec.Vec_IntSize(divValues) {
		panic("invalid SOP solve dimensions")
	}
	onset, enable := bsat.Positive(pivot), bsat.Positive(free)
	vec.Vec_IntClear(sop)
	for {
		status := solver.Solve(onset, enable)
		if status == bsat.Unknown {
			return 0
		}
		if status == bsat.Unsatisfiable {
			return 1
		}
		assumptions := []bsat.Literal{onset ^ 1}
		for _, v := range vec.Vec_IntArray(divVars) {
			assumptions = append(assumptions, sbdModelLiteral(solver, v))
		}
		status = solver.Solve(assumptions...)
		if status != bsat.Unsatisfiable {
			return 0
		}
		vec.Vec_IntClear(temp)
		vec.Vec_IntPush(temp, int32(enable^1))
		for _, literal := range solver.UnsatCore() {
			if literal == onset {
				continue
			}
			index := slices.Index(vec.Vec_IntArray(divVars), int32(literal>>1))
			if index < 0 {
				panic("SOP core references non-divisor")
			}
			vec.Vec_IntPush(temp, int32(literal))
			vec.Vec_IntPush(sop, int32(2*index)+1-int32(literal&1))
		}
		vec.Vec_IntPush(sop, -1)
		if !solver.AddClause(vec.Vec_IntArray(temp)...) {
			panic("SOP blocker made base inconsistent")
		}
	}
}
func Sbd_ManSolverSupp(sop *vec.Vec_Int_t, indices []int32) (uint64, int32) {
	var support uint64
	count := int32(0)
	for _, entry := range vec.Vec_IntArray(sop) {
		if entry == -1 {
			continue
		}
		v := entry >> 1
		if v < 0 || v >= 64 {
			panic("SOP support exceeds word")
		}
		if support>>v&1 != 0 {
			continue
		}
		indices[v] = count
		count++
		support |= uint64(1) << v
	}
	return support, count
}
func Sbd_ManSolverPrint(sop *vec.Vec_Int_t) {
	var indices [64]int32
	_, count := Sbd_ManSolverSupp(sop, indices[:])
	cube := make([]byte, count)
	for i := range cube {
		cube[i] = '-'
	}
	for _, entry := range vec.Vec_IntArray(sop) {
		if entry == -1 {
			fmt.Println(string(cube))
			for i := range cube {
				cube[i] = '-'
			}
		} else {
			cube[indices[entry>>1]] = '1' - byte(entry&1)
		}
	}
}
func Sbd_ManSolveSelect(network *Network, mirrors *vec.Vec_Int_t, pivot int32, divVars, divValues, window, objToVar, tfo, roots *vec.Vec_Int_t) {
	solver := Sbd_ManSatSolver(nil, network, mirrors, pivot, window, objToVar, tfo, roots, 0)
	if solver == nil {
		fmt.Println("UNSAT window.")
		return
	}
	sop, temp := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	status := Sbd_ManSolve2(solver, vec.Vec_IntEntry(objToVar, int(pivot)), int32(vec.Vec_IntSize(window)+vec.Vec_IntSize(tfo)+vec.Vec_IntSize(roots)), divVars, divValues, temp, sop)
	fmt.Printf("Pivot = %4d. Divs = %4d. Solved = %d\n", pivot, vec.Vec_IntSize(divVars), status)
	if status != 0 {
		Sbd_ManSolverPrint(sop)
	}
}
func sbdRandomPolarity(solver *bsat.Solver, random *aig.AigRandomSource) {
	for v := int32(0); v < solver.NumVars(); v++ {
		solver.SetPolarity(v, random.Uint32()&1 != 0)
	}
}
func Sbd_ManCollectConstants(solver *bsat.Solver, care [2]int32, pivot int32, sims [][]uint64, indices *vec.Vec_Int_t) int32 {
	if vec.Vec_IntSize(indices) != int(care[0]+care[1]) {
		panic("care pattern count mismatch")
	}
	random := aig.NewAigRandomSource()
	for i, index := range vec.Vec_IntArray(indices) {
		offset := int32(0)
		if int32(i) < care[0] {
			offset = 1
		}
		sbdRandomPolarity(solver, random)
		status := solver.Solve(bsat.Literal(2*pivot + offset))
		if status == bsat.Unknown {
			return -2
		}
		if status == bsat.Unsatisfiable {
			return offset
		}
		for v := int32(0); v <= pivot; v++ {
			mask := uint64(1) << uint(index&63)
			if sbdModelValue(solver, v) != 0 {
				sims[v][index>>6] |= mask
			} else {
				sims[v][index>>6] &^= mask
			}
		}
	}
	return -1
}
func Sbd_ManCollectConstantsNew(solver *bsat.Solver, divVars *vec.Vec_Int_t, count, pivot int32, onset, offset []uint64) int32 {
	if vec.Vec_IntSize(divVars) >= 64 || count < 0 {
		panic("invalid care sample dimensions")
	}
	random := aig.NewAigRandomSource()
	patterns := [2][]uint64{onset, offset}
	for polarity := int32(0); polarity < 2; polarity++ {
		for i := int32(0); i < count; i++ {
			sbdRandomPolarity(solver, random)
			status := solver.Solve(bsat.Literal(2*pivot + polarity))
			if status == bsat.Unknown {
				return -2
			}
			if status == bsat.Unsatisfiable {
				return polarity
			}
			value := uint64(1-polarity) << uint(vec.Vec_IntSize(divVars))
			for k, v := range vec.Vec_IntArray(divVars) {
				if sbdModelValue(solver, v) != 0 {
					value |= uint64(1) << uint(k)
				}
			}
			patterns[polarity][i] = value
		}
	}
	return -1
}
