package cnf

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Aig_ManScanMapping_rec(manager *Cnf_Man_t, object *Aig_Obj_t, mapped *Vec_Ptr_t) int32 {
	return scanMapping(manager, edgeValue(object), mapped, false, make(map[int]bool))
}
func Aig_ManScanMapping(manager *Cnf_Man_t, collect int32) *Vec_Ptr_t {
	result := vec.Vec_PtrAllocExact(0)
	if manager == nil || manager.pManAig == nil {
		return result
	}
	if len(manager.selectedCuts) == 0 {
		Cnf_DeriveMapping(manager)
	}
	visited := make(map[int]bool)
	for _, output := range manager.pManAig.Outputs() {
		node := manager.pManAig.Node(output)
		scanMapping(manager, node.Fanins[0], result, false, visited)
	}
	if collect == 0 {
		vec.Vec_PtrClear(result)
	}
	return result
}
func Cnf_ManScanMapping_rec(manager *Cnf_Man_t, object *Aig_Obj_t, mapped *Vec_Ptr_t, preorder int32) int32 {
	return scanMapping(manager, edgeValue(object), mapped, preorder != 0, make(map[int]bool))
}
func Cnf_ManScanMapping(manager *Cnf_Man_t, collect, preorder int32) *Vec_Ptr_t {
	result := vec.Vec_PtrAllocExact(0)
	if manager == nil || manager.pManAig == nil {
		return result
	}
	if len(manager.selectedCuts) == 0 {
		Cnf_DeriveMapping(manager)
	}
	visited := make(map[int]bool)
	for _, output := range manager.pManAig.Outputs() {
		node := manager.pManAig.Node(output)
		scanMapping(manager, node.Fanins[0], result, preorder != 0, visited)
	}
	if collect == 0 {
		vec.Vec_PtrClear(result)
	}
	return result
}
func scanMapping(manager *Cnf_Man_t, edge aig.Edge, mapped *Vec_Ptr_t, preorder bool, visited map[int]bool) int32 {
	if manager == nil || manager.pManAig == nil || edge == aig.InvalidEdge {
		return 0
	}
	id := edge.NodeID()
	if id <= 0 || id >= len(manager.selectedCuts) || visited[id] {
		return 0
	}
	node := manager.pManAig.Node(edge)
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer {
		return 0
	}
	visited[id] = true
	cut := manager.selectedCuts[id]
	cost := int32(1)
	if preorder {
		vec.Vec_PtrPushUnique(mapped, edge)
	}
	if cut != nil {
		for _, leaf := range cut.pFanins {
			cost += scanMapping(manager, aig.Edge(uint32(leaf)<<1), mapped, preorder, visited)
		}
	} else {
		cost += scanMapping(manager, node.Fanins[0], mapped, preorder, visited)
		if node.Kind == aig.And || node.Kind == aig.ExclusiveOr {
			cost += scanMapping(manager, node.Fanins[1], mapped, preorder, visited)
		}
	}
	if !preorder {
		vec.Vec_PtrPushUnique(mapped, edge)
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

var _ = fmt.Sprintf
