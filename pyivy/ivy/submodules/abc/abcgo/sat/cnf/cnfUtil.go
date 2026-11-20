package cnf

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Aig_ManScanMapping_rec(manager *Cnf_Man_t, object *Aig_Obj_t, mapped *Vec_Ptr_t) int32 {
	return scanMapping(manager, edgeValue(object), mapped, true, true)
}
func Aig_ManScanMapping(manager *Cnf_Man_t, collect int32) *Vec_Ptr_t {
	return scanOutputs(manager, collect != 0, true, true)
}
func Cnf_ManScanMapping_rec(manager *Cnf_Man_t, object *Aig_Obj_t, mapped *Vec_Ptr_t, preorder int32) int32 {
	return scanMapping(manager, edgeValue(object), mapped, preorder != 0, false)
}
func Cnf_ManScanMapping(manager *Cnf_Man_t, collect, preorder int32) *Vec_Ptr_t {
	return scanOutputs(manager, collect != 0, preorder != 0, false)
}

func scanOutputs(manager *Cnf_Man_t, collect, preorder, darCosts bool) *Vec_Ptr_t {
	if manager == nil || manager.pManAig == nil {
		if collect {
			return vec.Vec_PtrAllocExact(0)
		}
		return nil
	}
	if len(manager.selectedCuts) == 0 {
		Cnf_DeriveMapping(manager)
	}
	g := manager.pManAig
	for id := range aig.Aig_ManObjNumMax(g) {
		edge := aig.Aig_ManObj(g, id)
		if g.ValidEdge(edge) {
			aig.Aig_ObjClearRef(g, edge)
		}
	}
	var mapped *Vec_Ptr_t
	if collect {
		mapped = vec.Vec_PtrAllocExact(min(1000, g.NodeCount()))
	}
	manager.aArea = 0
	for i := range g.OutputCount() {
		edge := aig.Aig_ManCo(g, int32(i))
		if n := g.Node(edge); n.Kind == aig.Output {
			edge = n.Fanins[0]
		}
		manager.aArea += scanMapping(manager, edge, mapped, preorder, darCosts)
	}
	return mapped
}

// Each reference is an incoming edge in the selected mapping. Only the
// first visit to a gate contributes its cost and traverses its cut leaves.
func scanMapping(manager *Cnf_Man_t, edge aig.Edge, mapped *Vec_Ptr_t, preorder, darCosts bool) int32 {
	if manager == nil || manager.pManAig == nil || !manager.pManAig.ValidEdge(edge) {
		return 0
	}
	g := manager.pManAig
	edge = aig.Aig_Regular(edge)
	node := g.Node(edge)
	aig.Aig_ObjRef(g, edge)
	if node.Refs != 0 || node.Kind == aig.Input || node.Kind == aig.Constant {
		return 0
	}
	id := edge.NodeID()
	if id >= len(manager.selectedCuts) || node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer {
		return 0
	}
	if mapped != nil && preorder {
		vec.Vec_PtrPush(mapped, edge)
	}
	var cost int32
	switch cut := manager.selectedCuts[id]; {
	case node.MarkB && node.Kind != aig.Buffer:
		var leaves []aig.Edge
		aig.Aig_ObjCollectSuper(g, edge, &leaves)
		cost = int32(len(leaves) + 1)
		for _, leaf := range leaves {
			cost += scanMapping(manager, leaf, mapped, preorder, darCosts)
		}
	case cut != nil:
		cost = int32(cut.Cost)
		if darCosts && len(cut.truth) != 0 {
			truth := smallCutTruth(cut.truth[0], len(cut.pFanins))
			cost = int32(manager.pSopSizes[truth]) + int32(manager.pSopSizes[^truth])
		}
		for _, leaf := range cut.pFanins {
			cost += scanMapping(manager, aig.Edge(uint32(leaf)<<1), mapped, preorder, darCosts)
		}
	default:
		cost = 1 + scanMapping(manager, node.Fanins[0], mapped, preorder, darCosts)
		if node.Kind != aig.Buffer {
			cost += scanMapping(manager, node.Fanins[1], mapped, preorder, darCosts)
		}
	}
	if mapped != nil && !preorder {
		vec.Vec_PtrPush(mapped, edge)
	}
	return cost
}

func Cnf_DataCollectCiSatNums(data *Cnf_Dat_t, graph *Aig_Man_t) *Vec_Int_t {
	return Cnf_DataCollectPiSatNums(data, graph)
}
func Cnf_DataCollectCoSatNums(data *Cnf_Dat_t, graph *Aig_Man_t) *Vec_Int_t {
	result := vec.Vec_IntAllocExact(0)
	if data == nil || graph == nil {
		return result
	}
	for _, output := range graph.Outputs() {
		id := output.NodeID()
		variable := int32(-1)
		if id >= 0 && id < len(data.VarNums) {
			variable = data.VarNums[id]
		}
		vec.Vec_IntPush(result, variable)
	}
	return result
}

// Cnf_DataDeriveLitPolarities records the available polarity of each input
// variable in clauses grouped by their first (output) literal.
func Cnf_DataDeriveLitPolarities(data *Cnf_Dat_t) []byte {
	if data == nil {
		return nil
	}
	result := make([]byte, len(data.Clauses))
	if len(data.Obj2Clause) == 0 || len(data.Obj2Count) != len(data.Obj2Clause) {
		return result
	}
	for objectID, start := range data.Obj2Clause {
		if start < 0 || data.Obj2Count[objectID] <= 0 {
			continue
		}
		begin, end := int(start), int(start+data.Obj2Count[objectID])
		if begin >= len(data.Clauses) {
			continue
		}
		if end > len(data.Clauses) {
			end = len(data.Clauses)
		}
		positive, negative := make([]byte, data.Variables), make([]byte, data.Variables)
		for clauseID := begin; clauseID < end; clauseID++ {
			clause := data.Clauses[clauseID]
			if len(clause) == 0 {
				continue
			}
			polarity := negative
			if clause[0]&1 == 0 {
				polarity = positive
			}
			for _, lit := range clause[1:] {
				variable := int(lit >> 1)
				if variable >= len(polarity) {
					continue
				}
				mask := byte(2)
				if lit&1 != 0 {
					mask = 1
				}
				polarity[variable] |= mask
			}
		}
		for clauseID := begin; clauseID < end; clauseID++ {
			clause := data.Clauses[clauseID]
			if len(clause) == 0 {
				continue
			}
			polarity := positive
			if clause[0]&1 != 0 {
				polarity = negative
			}
			for slot, lit := range clause[1:] {
				if slot > 3 {
					break
				}
				variable := int(lit >> 1)
				if variable < len(polarity) {
					result[clauseID] |= polarity[variable] << (2 * slot)
				}
			}
		}
	}
	return result
}

func Cnf_DataReadFromFile(filename string) *Cnf_Dat_t {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 4<<20)
	variables, declaredClauses := -1, -1
	clauses := make([][]int32, 0)
	current := make([]int32, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "c") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[0] == "p" {
			if fields[1] != "cnf" || variables >= 0 {
				return nil
			}
			variables, err = strconv.Atoi(fields[2])
			if err != nil || variables < 0 {
				return nil
			}
			declaredClauses, err = strconv.Atoi(fields[3])
			if err != nil || declaredClauses < 0 {
				return nil
			}
			continue
		}
		if variables < 0 {
			return nil
		}
		for _, field := range fields {
			value, parseErr := strconv.ParseInt(field, 10, 32)
			if parseErr != nil {
				return nil
			}
			if value == 0 {
				clauses = append(clauses, current)
				current = nil
				continue
			}
			if value > int64(variables) || value < -int64(variables) {
				return nil
			}
			variable := int32(value)
			negative := variable < 0
			if negative {
				variable = -variable
			}
			lit := (variable - 1) << 1
			if negative {
				lit |= 1
			}
			current = append(current, lit)
		}
	}
	if scanner.Err() != nil || variables < 0 || len(current) != 0 || declaredClauses != len(clauses) {
		return nil
	}
	result := &Cnf_Dat_t_{Variables: variables, nVars: int32(variables), Clauses: clauses, pClauses: cloneClauses(clauses), VarNums: make([]int32, 0)}
	result.nClauses = int32(len(clauses))
	for _, clause := range clauses {
		result.nLiterals += int32(len(clause))
	}
	return result
}

// Cnf_DataSolveFromFile returns 1 for UNSAT, 0 for SAT, and -1 on error or limit.
func Cnf_DataSolveFromFile(filename string, conflictLimit, _ int32, _ int32, _ int32, _ int32, _ int32, model *[]int32, nPis int32) int32 {
	data := Cnf_DataReadFromFile(filename)
	if data == nil {
		return -1
	}
	solver := bsat.NewSolver()
	if conflictLimit > 0 {
		solver.SetLimits(int64(conflictLimit), 0)
	}
	for solver.NumVars() < int32(data.Variables) {
		solver.NewVariable()
	}
	for _, clause := range data.Clauses {
		lits := make([]bsat.Literal, len(clause))
		for i, lit := range clause {
			if lit&1 != 0 {
				lits[i] = bsat.Negative(lit >> 1)
			} else {
				lits[i] = bsat.Positive(lit >> 1)
			}
		}
		if !solver.AddClause(lits...) {
			return 1
		}
	}
	status := solver.Solve()
	if status == bsat.Unsatisfiable {
		return 1
	}
	if status != bsat.Satisfiable {
		return -1
	}
	if model != nil && nPis > 0 {
		count := int(nPis)
		if count > data.Variables {
			count = data.Variables
		}
		values := make([]int32, count)
		for i := 0; i < count; i++ {
			value, ok := solver.Value(int32(data.Variables - count + i))
			if ok && value {
				values[i] = 1
			} else {
				values[i] = 0
			}
		}
		*model = values
	}
	return 0
}
