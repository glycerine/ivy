package mf

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
)

// CNF uses ABC literals (variable<<1|negative), including the source's unused
// variable zero in compact mode. Mapping is the source flat index/payload table.
type CNF struct {
	Variables             int
	Clauses               [][]int32
	VarNums               []int32
	Mapping               []int32
	Obj2Clause, Obj2Count []int32
}

func cubeLiteral(cube int32, variable int) int32 { return cube >> uint(2*variable) & 3 }
func encodedLiteral(variable int32, negative bool) int32 {
	lit := variable << 1
	if negative {
		lit |= 1
	}
	return lit
}
func filledIDs(n int) []int32 {
	ids := make([]int32, n)
	for i := range ids {
		ids[i] = -1
	}
	return ids
}

// DeriveCNF follows MF's CO-first, reverse-mapped-node, constant, CI variable
// ordering and emits each LUT clause's output literal before its input literals.
func DeriveCNF(m *Mapping, objectIDs, addOr, withMapping bool) (*CNF, error) {
	if m == nil || m.Network == nil {
		return nil, fmt.Errorf("mf: nil mapping")
	}
	if m.Options.LUTSize < 1 || m.Options.LUTSize > 8 {
		return nil, fmt.Errorf("mf: CNF LUT size must be 1..8")
	}
	n := m.Network.ObjectCount()
	if n < 1 {
		return nil, fmt.Errorf("mf: empty network")
	}
	inputs, outputs := m.Network.InputLiterals(), m.Network.OutputLiterals()
	var cos []int32
	for id := 0; id < n; id++ {
		_, _, _, kind := m.Network.Gate(int32(id))
		if kind == 6 {
			cos = append(cos, int32(id))
		}
	}
	synthetic := len(cos) == 0
	if synthetic {
		for i := range outputs {
			cos = append(cos, int32(n+i))
		}
	} else if len(cos) != len(outputs) {
		return nil, fmt.Errorf("mf: output count mismatch")
	}
	idCount := n
	if synthetic {
		idCount += len(outputs)
	}
	ids := filledIDs(idCount)
	next := int32(1)
	assign := func(id int32) {
		if objectIDs {
			ids[id] = id
		} else {
			ids[id] = next
		}
		next++
	}
	for _, id := range cos {
		assign(id)
	}
	var mapped []int32
	for id := n - 1; id >= 1; id-- {
		_, _, _, kind := m.Network.Gate(int32(id))
		if gateNode(kind) && id < len(m.Refs) && m.Refs[id] > 0 {
			if id >= len(m.Cuts) || len(m.Cuts[id]) == 0 {
				return nil, fmt.Errorf("mf: missing cut for object%d", id)
			}
			mapped = append(mapped, int32(id))
			assign(int32(id))
		}
	}
	assign(0)
	for _, lit := range inputs {
		if lit < 0 || lit&1 != 0 || int(lit>>1) >= n {
			return nil, fmt.Errorf("mf: invalid input literal%d", lit)
		}
		assign(lit >> 1)
	}
	result := &CNF{Variables: int(next), VarNums: ids}
	if objectIDs {
		for _, id := range ids {
			result.Variables = max(result.Variables, int(id)+1)
		}
	}
	if withMapping {
		result.Mapping = make([]int32, result.Variables)
	}
	add := func(lits ...int32) { result.Clauses = append(result.Clauses, lits) }
	if addOr {
		clause := make([]int32, len(cos))
		for i, id := range cos {
			clause[i] = encodedLiteral(ids[id], false)
		}
		add(clause...)
	}
	writeMap := func(id int32, leaves []int32, truth []uint64) {
		result.Mapping[ids[id]] = int32(len(result.Mapping))
		result.Mapping = append(result.Mapping, int32(len(leaves)))
		for _, leaf := range leaves {
			result.Mapping = append(result.Mapping, ids[leaf])
		}
		words := 1
		if len(leaves) >= 6 {
			words = 1 << (len(leaves) - 5)
		}
		for i := 0; i < words; i++ {
			result.Mapping = append(result.Mapping, int32(uint32(truth[i/2]>>uint(32*(i%2)))))
		}
	}
	for i, id := range cos {
		driver := outputs[i]
		if !synthetic {
			driver, _, _, _ = m.Network.Gate(id)
		}
		if driver < 0 || int(driver>>1) >= len(ids) || ids[driver>>1] < 0 {
			return nil, fmt.Errorf("mf: unmapped output driver%d", driver)
		}
		v, d := ids[id], ids[driver>>1]
		add(encodedLiteral(v, false), encodedLiteral(d, driver&1 == 0))
		add(encodedLiteral(v, true), encodedLiteral(d, driver&1 != 0))
		if withMapping {
			truth := uint64(0xAAAAAAAAAAAAAAAA)
			if driver&1 != 0 {
				truth = ^truth
			}
			writeMap(id, []int32{driver >> 1}, []uint64{truth})
		}
	}
	templates := map[int32][]int32{}
	for _, id := range mapped {
		cut := m.Cuts[id][0]
		if len(cut.Leaves) > m.Options.LUTSize {
			return nil, fmt.Errorf("mf: cut exceeds LUT size")
		}
		function, phase := cut.Function>>1, cut.Function&1 != 0
		if function < 0 || int(function) >= len(m.Truths) {
			return nil, fmt.Errorf("mf: invalid truth function%d", function)
		}
		if function == 0 {
			add(encodedLiteral(ids[id], !phase))
			if withMapping {
				return nil, fmt.Errorf("mf: source mapping does not support constant cuts")
			}
			continue
		}
		if len(m.Truths[function]) < 1<<max(0, len(cut.Leaves)-6) {
			return nil, fmt.Errorf("mf: short truth table")
		}
		cubes, ok := templates[function]
		if !ok {
			buffer := make([]int32, 512)
			var count int32
			if m.Options.LUTSize <= 6 {
				count = util.Abc_Tt6Cnf(m.Truths[function][0], int32(len(cut.Leaves)), buffer)
			} else {
				table := make([]uint64, 4)
				copy(table, m.Truths[function])
				count = util.Abc_Tt8Cnf(table, int32(len(cut.Leaves)), buffer)
			}
			cubes = append([]int32(nil), buffer[:count]...)
			templates[function] = cubes
			if int(function) < len(m.ClauseSizes) && m.ClauseSizes[function] != len(cubes) {
				return nil, fmt.Errorf("mf: clause count mismatch for function%d", function)
			}
		}
		for _, leaf := range cut.Leaves {
			if leaf < 0 || int(leaf) >= len(ids) || ids[leaf] < 0 {
				return nil, fmt.Errorf("mf: unmapped cut leaf%d", leaf)
			}
		}
		for _, cube := range cubes {
			output := cubeLiteral(cube, len(cut.Leaves))
			if output == 0 {
				return nil, fmt.Errorf("mf: clause template has no output")
			}
			clause := []int32{encodedLiteral(ids[id], (output == 2) != phase)}
			for k, leaf := range cut.Leaves {
				if literal := cubeLiteral(cube, k); literal != 0 {
					clause = append(clause, encodedLiteral(ids[leaf], literal == 2))
				}
			}
			add(clause...)
		}
		if withMapping {
			truth := append([]uint64(nil), m.Truths[function]...)
			if phase {
				for i := range truth {
					truth[i] = ^truth[i]
				}
			}
			writeMap(id, cut.Leaves, truth)
		}
	}
	add(encodedLiteral(ids[0], true))
	if objectIDs {
		result.Obj2Clause = filledIDs(len(ids))
		result.Obj2Count = filledIDs(len(ids))
		for i, clause := range result.Clauses {
			if len(clause) == 0 {
				continue
			}
			id := clause[0] >> 1
			if result.Obj2Clause[id] < 0 {
				result.Obj2Clause[id] = int32(i)
				result.Obj2Count[id] = 1
			} else {
				result.Obj2Count[id]++
			}
		}
	}
	return result, nil
}
