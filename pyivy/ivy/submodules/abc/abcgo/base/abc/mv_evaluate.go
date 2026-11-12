package abc

import (
	"fmt"
	"sort"
)

// EvaluateMV evaluates a multi-valued combinational cone. The result contains
// every allowed output value, making nondeterministic covers explicit.
func (n *Network) EvaluateMV(root *Object, inputs map[*Object]int) ([]int, error) {
	order, err := n.Topological([]*Object{root})
	if err != nil {
		return nil, err
	}
	values := map[*Object][]int{}
	for _, o := range order {
		if o.Kind == ObjectConst1 {
			values[o] = []int{1}
			continue
		}
		if o.IsCI() {
			value, ok := inputs[o]
			if !ok {
				return nil, fmt.Errorf("missing MV input %d", o.ID)
			}
			domain := 2
			if len(o.fanouts) > 0 && o.fanouts[0].Kind == ObjectNet {
				domain = int(Abc_ObjMvVarNum(o.fanouts[0]))
			}
			if value < 0 || value >= domain {
				return nil, fmt.Errorf("input %d outside domain", o.ID)
			}
			values[o] = []int{value}
			continue
		}
		choices := make([][]int, len(o.fanins))
		for i, e := range o.fanins {
			choices[i] = append([]int(nil), values[e.Object]...)
			if e.Complement {
				for j, v := range choices[i] {
					if v != 0 && v != 1 {
						return nil, fmt.Errorf("complement on nonbinary edge")
					}
					choices[i][j] = 1 - v
				}
			}
		}
		if o.Kind == ObjectNet || o.IsCO() || isBarrierBuffer(o) {
			if len(choices) != 1 {
				return nil, fmt.Errorf("identity object has wrong arity")
			}
			values[o] = choices[0]
			continue
		}
		if o.Kind != ObjectNode {
			return nil, fmt.Errorf("cannot evaluate box %d", o.ID)
		}
		var table *MVTable
		switch data := o.Data.(type) {
		case *MVTable:
			table = data
		case string:
			if data != aigANDMarker {
				table, err = nodeMVTable(o)
				if err != nil {
					return nil, err
				}
			}
		}
		result := map[int]bool{}
		assignment := make([]int, len(choices))
		combinations := 0
		var walk func(int) error
		walk = func(index int) error {
			if index < len(choices) {
				for _, v := range choices[index] {
					assignment[index] = v
					if err := walk(index + 1); err != nil {
						return err
					}
				}
				return nil
			}
			combinations++
			if combinations > 1<<20 {
				return fmt.Errorf("MV evaluation combination limit")
			}
			if table != nil {
				outputs, err := table.Evaluate(assignment)
				if err != nil {
					return err
				}
				for _, v := range outputs {
					result[v] = true
				}
				return nil
			}
			bits := make([]bool, len(assignment))
			for i, v := range assignment {
				if v != 0 && v != 1 {
					return fmt.Errorf("nonbinary input to Boolean node")
				}
				bits[i] = v != 0
			}
			value := false
			if fn, ok := o.Data.(func([]bool) bool); ok {
				value = fn(bits)
			} else if o.Data == aigANDMarker && len(bits) == 2 {
				value = bits[0] && bits[1]
			} else {
				return fmt.Errorf("object has no MV/Boolean function")
			}
			if value {
				result[1] = true
			} else {
				result[0] = true
			}
			return nil
		}
		if err := walk(0); err != nil {
			return nil, err
		}
		var outputs []int
		for v := range result {
			outputs = append(outputs, v)
		}
		sort.Ints(outputs)
		values[o] = outputs
	}
	return append([]int(nil), values[root]...), nil
}
