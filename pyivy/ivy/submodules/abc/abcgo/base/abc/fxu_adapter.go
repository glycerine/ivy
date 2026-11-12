package abc

import (
	"fmt"
	"strings"
)

// FXUCovers uses stable object IDs, leaving non-SOP rows empty. Complemented
// input edges are folded into the exported cover without mutating the graph.
func (n *Network) FXUCovers() ([]string, [][]int32, error) {
	if n.Type != uint8(ABC_NTK_LOGIC) || n.Function != uint8(ABC_FUNC_SOP) {
		return nil, nil, fmt.Errorf("FXU requires SOP logic")
	}
	if err := n.Check(); err != nil {
		return nil, nil, err
	}
	sops := make([]string, len(n.objects))
	fanins := make([][]int32, len(n.objects))
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		text, ok := o.Data.(string)
		if !ok || !checkSOPSyntax(text, len(o.fanins)) {
			return nil, nil, fmt.Errorf("object %d has invalid SOP functionality", o.ID)
		}
		if text == "" {
			text = strings.Repeat("-", len(o.fanins)) + " 0\n"
		}
		rows := fanSOPRows(text, len(o.fanins))
		for i, row := range rows {
			pattern := []byte(row.pattern)
			odd := false
			for pin, edge := range o.fanins {
				if !edge.Complement {
					continue
				}
				odd = !odd
				if pattern[pin] == '0' {
					pattern[pin] = '1'
				} else if pattern[pin] == '1' {
					pattern[pin] = '0'
				}
			}
			if odd && (row.output == 'x' || row.output == 'n') {
				if row.output == 'x' {
					rows[i].output = 'n'
				} else {
					rows[i].output = 'x'
				}
			}
			rows[i].pattern = string(pattern)
		}
		sops[o.ID] = fanSOPText(rows)
		for _, edge := range o.fanins {
			fanins[o.ID] = append(fanins[o.ID], int32(edge.Object.ID))
		}
	}
	return sops, fanins, nil
}

// ReplaceFXUCovers validates the complete candidate cone before editing any
// object. Appended rows may refer forward to other appended definitions.
func (n *Network) ReplaceFXUCovers(sops []string, fanins [][]int32) error {
	if n.Type != uint8(ABC_NTK_LOGIC) || n.Function != uint8(ABC_FUNC_SOP) {
		return fmt.Errorf("FXU requires SOP logic")
	}
	if len(sops) != len(fanins) || len(sops) < len(n.objects) {
		return fmt.Errorf("invalid FXU row count")
	}
	if err := n.Check(); err != nil {
		return err
	}
	oldCount := len(n.objects)
	defined := func(row int) bool {
		return row >= 0 && row < len(sops) && (row < oldCount && n.Object(row) != nil || row >= oldCount && sops[row] != "")
	}
	sourceAllowed := func(row int) bool {
		if !defined(row) {
			return false
		}
		if row >= oldCount {
			return true
		}
		o := n.Object(row)
		return o.Kind == ObjectConst1 || o.IsCI() || o.Kind == ObjectNode
	}
	for row, text := range sops {
		if text == "" {
			continue
		}
		if row < oldCount {
			o := n.Object(row)
			if o == nil || o.Kind != ObjectNode || o.Barrier {
				return fmt.Errorf("FXU row %d is not a writable SOP node", row)
			}
		}
		if !checkSOPSyntax(text, len(fanins[row])) {
			return fmt.Errorf("invalid SOP in FXU row %d", row)
		}
		for _, id := range fanins[row] {
			if !sourceAllowed(int(id)) {
				return fmt.Errorf("invalid fanin %d in FXU row %d", id, row)
			}
		}
	}
	state := make([]byte, len(sops))
	var visit func(int) error
	visit = func(row int) error {
		if state[row] == 2 {
			return nil
		}
		if state[row] == 1 {
			return fmt.Errorf("FXU introduces a combinational cycle at row %d", row)
		}
		if row < oldCount {
			object := n.Object(row)
			if object == nil || object.Kind != ObjectNode {
				return nil
			}
		}
		state[row] = 1
		var inputs []int32
		if sops[row] != "" {
			inputs = fanins[row]
		} else {
			for _, edge := range n.Object(row).fanins {
				inputs = append(inputs, int32(edge.Object.ID))
			}
		}
		for _, id := range inputs {
			if err := visit(int(id)); err != nil {
				return err
			}
		}
		state[row] = 2
		return nil
	}
	for row := range sops {
		if defined(row) {
			if err := visit(row); err != nil {
				return err
			}
		}
	}
	objects := make([]*Object, len(sops))
	copy(objects, n.objects)
	for row := oldCount; row < len(sops); row++ {
		if sops[row] != "" {
			objects[row] = n.NewObject(ObjectNode)
		}
	}
	for row, text := range sops {
		if text == "" {
			continue
		}
		object := objects[row]
		Abc_ObjRemoveFanins(object)
		object.Data = text
		for _, id := range fanins[row] {
			if err := n.AddFanin(object, objects[id].Signal()); err != nil {
				panic(err)
			}
		}
	}
	return nil
}
