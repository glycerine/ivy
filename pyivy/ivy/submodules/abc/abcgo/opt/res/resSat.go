package res

import (
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Res_SatProveUnsat follows the source proof construction, including its call
// to Sat_SolverDoubleClauses. That checked-in ABC routine unconditionally
// asserts; this entry point consequently retains that source behavior.
func Res_SatProveUnsat(pAig *Abc_Ntk_t, vFanins *Vec_Ptr_t) any {
	count := Vec_PtrSize(vFanins)
	if count < 2 {
		panic("res: expected care and onset outputs")
	}
	roots := make([]*Abc_Obj_t, count)
	for i := range roots {
		roots[i] = Vec_PtrEntry(vFanins, int32(i))
	}
	if roots[0].Network() != pAig || abc.Abc_ObjIsPo(roots[0]) == 0 {
		panic("res: first fanin must be an output of the window AIG")
	}
	nodes := abc.Abc_NtkDfsNodes(pAig, roots, count)
	numbers := make(map[*Abc_Obj_t]int32)
	var n int32
	assign := func(obj *Abc_Obj_t) { numbers[obj] = n; n++ }
	assign(abc.Abc_AigConst1(pAig).Object)
	for i := int32(0); i < abc.Abc_NtkPiNum(pAig); i++ {
		assign(abc.Abc_NtkPi(pAig, i))
	}
	for _, obj := range nodes {
		assign(obj)
	}
	for _, obj := range roots {
		assign(obj)
	}
	solver := bsat.NewSolver()
	solver.EnableClauseStore()
	Res_SatAddConst1(solver, numbers[abc.Abc_AigConst1(pAig).Object], 0)
	for _, obj := range nodes {
		Res_SatAddAnd(solver, numbers[obj], numbers[abc.Abc_ObjFanin0(obj)], numbers[abc.Abc_ObjFanin1(obj)], abc.Abc_ObjFaninC0(obj), abc.Abc_ObjFaninC1(obj))
	}
	for _, obj := range roots {
		Res_SatAddEqual(solver, numbers[obj], numbers[abc.Abc_ObjFanin0(obj)], abc.Abc_ObjFaninC0(obj))
	}
	Res_SatAddConst1(solver, numbers[roots[0]], 0)
	Res_SatAddConst1(solver, numbers[roots[1]], 0)
	solver.MarkClauseStorePartitionA()
	bsat.Sat_SolverDoubleClauses(solver, numbers[roots[1]])
	for _, obj := range roots[2:] {
		Res_SatAddEqual(solver, numbers[obj], numbers[obj]+n, 0)
	}
	solver.MarkClauseStoreRoots()
	solver.SetLimits(10000, 0)
	if solver.Solve() == bsat.Unsatisfiable {
		return solver.ReleaseClauseStore()
	}
	return nil
}

// resSatSimulateConstr constructs the source's care/onset SAT instance and
// keeps the source object-to-variable numbering in a Go-owned map.
func resSatSimulateConstr(pAig *Abc_Ntk_t, onSet int32) (*bsat.Solver, map[*Abc_Obj_t]int32) {
	roots := []*Abc_Obj_t{abc.Abc_NtkPo(pAig, 0), abc.Abc_NtkPo(pAig, 1)}
	nodes := abc.Abc_NtkDfsNodes(pAig, roots, 2)
	numbers := make(map[*Abc_Obj_t]int32)
	var n int32
	assign := func(obj *Abc_Obj_t) { numbers[obj] = n; n++ }
	assign(abc.Abc_AigConst1(pAig).Object)
	for i := int32(0); i < abc.Abc_NtkPiNum(pAig); i++ {
		assign(abc.Abc_NtkPi(pAig, i))
	}
	for _, obj := range nodes {
		assign(obj)
	}
	for _, obj := range roots {
		assign(obj)
	}
	solver := bsat.NewSolver()
	Res_SatAddConst1(solver, numbers[abc.Abc_AigConst1(pAig).Object], 0)
	for _, obj := range nodes {
		Res_SatAddAnd(solver, numbers[obj], numbers[abc.Abc_ObjFanin0(obj)], numbers[abc.Abc_ObjFanin1(obj)], abc.Abc_ObjFaninC0(obj), abc.Abc_ObjFaninC1(obj))
	}
	for _, obj := range roots {
		Res_SatAddEqual(solver, numbers[obj], numbers[abc.Abc_ObjFanin0(obj)], abc.Abc_ObjFaninC0(obj))
	}
	Res_SatAddConst1(solver, numbers[roots[0]], 0)
	Res_SatAddConst1(solver, numbers[roots[1]], resInvertPhase(onSet))
	return solver, numbers
}

func Res_SatSimulateConstr(pAig *Abc_Ntk_t, onSet int32) *bsat.Solver {
	solver, _ := resSatSimulateConstr(pAig, onSet)
	return solver
}

func Res_SatSimulate(p *Res_Sim_t, limit, onSet int32) int32 {
	started := util.Abc_Clock()
	defer func() { p.timeSat += util.Abc_Clock() - started }()
	count, patterns := p.nPats0, p.vPats0
	if onSet != 0 {
		count, patterns = p.nPats1, p.vPats1
	}
	if count >= limit {
		panic("res: pattern limit already reached")
	}
	solver, numbers := resSatSimulateConstr(p.pAig, onSet)
	if !solver.Simplify() {
		if count == 0 {
			return 0
		}
		return -1
	}
	result := int32(1)
	k := count
	for ; k < limit; k++ {
		solver.SetLimits(10000, 0)
		switch solver.Solve() {
		case bsat.Unsatisfiable:
			if k == 0 {
				if onSet != 0 {
					p.fConst0 = 1
				} else {
					p.fConst1 = 1
				}
				result = 0
			}
			goto done
		case bsat.Satisfiable:
			block := make([]bsat.Literal, 0, p.nTruePis)
			for i := int32(0); i < p.nTruePis; i++ {
				variable := numbers[abc.Abc_NtkPi(p.pAig, i)]
				value, _ := solver.Value(variable)
				if value {
					resWords(Vec_PtrEntryAny(patterns, i))[k/32] |= 1 << uint(k%32)
				}
				literal := bsat.Positive(variable)
				if value {
					literal ^= 1
				}
				block = append(block, literal)
			}
			if !solver.AddClause(block...) {
				k++
				goto done
			}
		default:
			if k == 0 {
				result = 0
			}
			goto done
		}
	}
done:
	if onSet != 0 {
		p.nPats1 = k
	} else {
		p.nPats0 = k
	}
	return result
}
