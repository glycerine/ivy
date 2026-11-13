package cnf

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
)

// CNF is the common Go representation shared by AIG, BMC, and proof code.
// Literals use ABC's encoding: variable<<1 is positive and bit zero negates.
type CNF = Cnf_Dat_t_

func (cnf *CNF) Validate() error {
	if cnf == nil || cnf.Variables < 0 {
		return fmt.Errorf("cnf: invalid variable count")
	}
	for ci, clause := range cnf.Clauses {
		for _, lit := range clause {
			if lit < 0 || int(lit>>1) >= cnf.Variables {
				return fmt.Errorf("cnf: clause %d has invalid literal %d", ci, lit)
			}
		}
	}
	return nil
}

// EncodeAIG builds an exact Tseitin encoding for an AIG and asserts the first
// nOutputs outputs. A negative count asserts all outputs. VarNums is indexed by
// source object ID; the constant object is marked -1.
func EncodeAIG(graph *aig.Graph, nOutputs int) (*CNF, error) {
	if graph == nil {
		return nil, fmt.Errorf("cnf: nil AIG")
	}
	nodes := graph.Nodes()
	result := &CNF{pMan: graph, VarNums: make([]int32, len(nodes))}
	for i := range result.VarNums {
		result.VarNums[i] = -1
	}
	for id, node := range nodes {
		if id == 0 || node.Kind == aig.Constant {
			continue
		}
		result.VarNums[id] = int32(result.Variables)
		result.Variables++
	}
	result.nVars = int32(result.Variables)
	result.pVarNums = append([]int32(nil), result.VarNums...)

	type value struct {
		lit             int32
		constant, truth bool
	}
	get := func(edge aig.Edge) value {
		id := edge.NodeID()
		if id < 0 || id >= len(nodes) {
			return value{constant: true}
		}
		if id == 0 || nodes[id].Kind == aig.Constant {
			return value{constant: true, truth: !edge.Complemented()}
		}
		return value{lit: result.VarNums[id]<<1 | boolBit(edge.Complemented())}
	}
	appendClause := func(clause ...int32) {
		result.Clauses = append(result.Clauses, clause)
		result.nLiterals += int32(len(clause))
		result.nClauses++
	}
	force := func(v int32, truth bool) { appendClause(v<<1 | boolBit(!truth)) }
	appendEquiv := func(out int32, input value) {
		if input.constant {
			force(out, input.truth)
			return
		}
		appendClause(out<<1|1, input.lit)
		appendClause(out<<1, input.lit^1)
	}
	for id, node := range nodes {
		if id == 0 || result.VarNums[id] < 0 {
			continue
		}
		out := result.VarNums[id]
		a, b := get(node.Fanins[0]), get(node.Fanins[1])
		switch node.Kind {
		case aig.Input:
		case aig.Buffer, aig.Output:
			appendEquiv(out, a)
		case aig.And:
			if a.constant && b.constant {
				force(out, a.truth && b.truth)
				break
			}
			if a.constant {
				if !a.truth {
					force(out, false)
				} else {
					appendEquiv(out, b)
				}
				break
			}
			if b.constant {
				if !b.truth {
					force(out, false)
				} else {
					appendEquiv(out, a)
				}
				break
			}
			appendClause(out<<1|1, a.lit)
			appendClause(out<<1|1, b.lit)
			appendClause(out<<1, a.lit^1, b.lit^1)
		case aig.ExclusiveOr:
			if a.constant && b.constant {
				force(out, a.truth != b.truth)
				break
			}
			if a.constant {
				if a.truth {
					b.lit ^= 1
				}
				appendEquiv(out, b)
				break
			}
			if b.constant {
				if b.truth {
					a.lit ^= 1
				}
				appendEquiv(out, a)
				break
			}
			appendClause(a.lit^1, b.lit^1, out<<1|1)
			appendClause(a.lit, b.lit, out<<1|1)
			appendClause(a.lit, b.lit^1, out<<1)
			appendClause(a.lit^1, b.lit, out<<1)
		}
	}
	outputs := graph.Outputs()
	if nOutputs < 0 || nOutputs > len(outputs) {
		nOutputs = len(outputs)
	}
	for _, output := range outputs[:nOutputs] {
		node := graph.Node(output)
		driver := get(node.Fanins[0])
		if driver.constant {
			if !driver.truth {
				appendClause()
			}
			continue
		}
		appendClause(driver.lit)
	}
	result.pClauses = cloneClauses(result.Clauses)
	result.nClauses = int32(len(result.Clauses))
	result.nLiterals = 0
	for _, clause := range result.Clauses {
		result.nLiterals += int32(len(clause))
	}
	return result, nil
}

func boolBit(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
func cloneClauses(clauses [][]int32) [][]int32 {
	out := make([][]int32, len(clauses))
	for i := range clauses {
		out[i] = append([]int32(nil), clauses[i]...)
	}
	return out
}

// NewMappedCNF constructs CNF data from ABC-encoded clauses and mapping
// metadata. It copies its inputs so the caller can reuse their buffers.
func NewMappedCNF(variables int, clauses [][]int32, varNums, obj2Clause, obj2Count []int32) *CNF {
	result := &CNF{Variables: variables, nVars: int32(variables)}
	result.Clauses = cloneClauses(clauses)
	result.pClauses = cloneClauses(clauses)
	result.nClauses = int32(len(clauses))
	for _, clause := range clauses {
		result.nLiterals += int32(len(clause))
	}
	result.VarNums = append([]int32(nil), varNums...)
	result.pVarNums = append([]int32(nil), varNums...)
	result.Obj2Clause = append([]int32(nil), obj2Clause...)
	result.pObj2Clause = append([]int32(nil), obj2Clause...)
	result.Obj2Count = append([]int32(nil), obj2Count...)
	result.pObj2Count = append([]int32(nil), obj2Count...)
	return result
}
