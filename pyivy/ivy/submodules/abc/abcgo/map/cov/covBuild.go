package cov

import (
	"fmt"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/cov/minimize"
)

// deriveAIGToSOP builds a logic/SOP network with the same objects, fanin order,
// and complemented edges as an AIG network. Every AIG AND node is represented
// by its exact local SOP; the native minimizer canonicalizes that cover.
func deriveAIGToSOP(source *abc.Network, nFaninMax, nCubesMax int32) (*abc.Network, error) {
	if source == nil || source.Type != uint8(abc.ABC_NTK_STRASH) || source.Function != uint8(abc.ABC_FUNC_AIG) {
		return nil, fmt.Errorf("cov: expected a strashed AIG network")
	}
	if nFaninMax < 2 {
		return nil, fmt.Errorf("cov: fanin limit %d cannot represent an AIG AND", nFaninMax)
	}
	if nCubesMax < 1 && len(source.ObjectsOfKind(abc.ObjectNode)) != 0 {
		return nil, fmt.Errorf("cov: cube limit %d cannot represent an AIG AND", nCubesMax)
	}
	cover, err := minimize.Minimize(2, []uint64{3}, nil)
	if err != nil {
		return nil, fmt.Errorf("cov: minimize AIG AND: %w", err)
	}
	coverText := formatSOP(cover, 2)

	target := abc.NewNetwork(source.Name)
	target.Type = uint8(abc.ABC_NTK_LOGIC)
	target.Function = uint8(abc.ABC_FUNC_SOP)
	target.Spec = source.Spec
	objects := make(map[*abc.Object]*abc.Object, len(source.Objects()))
	nameOf := func(old *abc.Object, prefix string) string {
		name := old.Name()
		if name == "" {
			return fmt.Sprintf("%s%d", prefix, old.ID)
		}
		return name
	}
	for _, old := range source.ObjectsOfKind(abc.ObjectPI) {
		input := target.NewObject(abc.ObjectPI)
		input.SetName(nameOf(old, "pi_"))
		objects[old] = input
	}
	objects[source.One().Object] = target.One().Object

	order, err := source.Topological(source.Objects())
	if err != nil {
		return nil, fmt.Errorf("cov: input network traversal: %w", err)
	}
	for _, old := range order {
		if old.Kind != abc.ObjectNode {
			continue
		}
		fanins := old.Fanins()
		if len(fanins) != 2 {
			return nil, fmt.Errorf("cov: AIG node %d has %d fanins, want 2", old.ID, len(fanins))
		}
		if len(fanins) > int(nFaninMax) {
			return nil, fmt.Errorf("cov: AIG node %d exceeds fanin limit", old.ID)
		}
		node := target.NewObject(abc.ObjectNode)
		node.SetName(nameOf(old, "node_"))
		node.Data = coverText
		for _, fanin := range fanins {
			mapped := objects[fanin.Object]
			if mapped == nil {
				return nil, fmt.Errorf("cov: no mapped source for fanin object %d", fanin.Object.ID)
			}
			if err := target.AddFanin(node, abc.Edge{Object: mapped, Complement: fanin.Complement}); err != nil {
				return nil, fmt.Errorf("cov: connect node %d: %w", old.ID, err)
			}
		}
		objects[old] = node
	}
	for _, old := range source.ObjectsOfKind(abc.ObjectPO) {
		fanins := old.Fanins()
		if len(fanins) != 1 {
			return nil, fmt.Errorf("cov: output %d has %d fanins, want 1", old.ID, len(fanins))
		}
		mapped := objects[fanins[0].Object]
		if mapped == nil {
			return nil, fmt.Errorf("cov: no mapped source for output %d", old.ID)
		}
		output := target.NewObject(abc.ObjectPO)
		output.SetName(nameOf(old, "po_"))
		if err := target.AddFanin(output, abc.Edge{Object: mapped, Complement: fanins[0].Complement}); err != nil {
			return nil, fmt.Errorf("cov: connect output %d: %w", old.ID, err)
		}
	}
	if err := target.Check(); err != nil {
		return nil, fmt.Errorf("cov: constructed network invariant: %w", err)
	}
	return target, nil
}

func formatSOP(cubes []minimize.Cube, variables int) string {
	if len(cubes) == 0 {
		return strings.Repeat("-", variables) + " 0\n"
	}
	var out strings.Builder
	for _, cube := range cubes {
		for variable := 0; variable < variables; variable++ {
			bit := uint64(1) << variable
			switch {
			case cube.Mask&bit == 0:
				out.WriteByte('-')
			case cube.Bits&bit != 0:
				out.WriteByte('1')
			default:
				out.WriteByte('0')
			}
		}
		out.WriteString(" 1\n")
	}
	return out.String()
}

func Abc_NtkCovDerive(p *Cov_Man_t, source *Abc_Ntk_t) *Abc_Ntk_t {
	return Abc_NtkCovDeriveRegular(p, source)
}

func Abc_NtkCovDeriveClean(p *Cov_Man_t, source *Abc_Ntk_t) *Abc_Ntk_t {
	return Abc_NtkCovDeriveRegular(p, source)
}

func Abc_NtkCovDeriveRegular(p *Cov_Man_t, source *Abc_Ntk_t) *Abc_Ntk_t {
	if p == nil {
		return nil
	}
	result, err := deriveAIGToSOP(source, p.nFaninMax, p.nCubesMax)
	if err != nil {
		return nil
	}
	return result
}
