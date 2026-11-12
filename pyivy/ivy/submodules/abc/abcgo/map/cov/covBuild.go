package cov

import (
	"fmt"
	"math/bits"
	"sort"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/cov/minimize"
)

type nodeCover struct {
	support  []*abc.Object
	sop      []minimize.Cube
	sopPhase byte
	esop     []minimizeCube
	useESOP  bool
}

// deriveAIGToSOP computes bounded covers from the outputs toward the inputs.
// If a cone exceeds either limit, its already-computed fanin covers become a
// cut boundary and the next pass retries above that cut. This mirrors the
// iterative boundary growth in covCore.c while evaluating exact local truth
// tables on the Go-owned ABC graph.
func deriveAIGToSOP(source *abc.Network, nFaninMax, nCubesMax int32, useESOP, useSOP int32) (*abc.Network, error) {
	if source == nil || source.Type != uint8(abc.ABC_NTK_STRASH) || source.Function != uint8(abc.ABC_FUNC_AIG) {
		return nil, fmt.Errorf("cov: expected a strashed AIG network")
	}
	if nFaninMax < 0 || nCubesMax < 0 || useESOP == 0 && useSOP == 0 {
		return nil, fmt.Errorf("cov: invalid support, cube, or cover-mode limit")
	}
	if _, err := source.Topological(source.Objects()); err != nil {
		return nil, fmt.Errorf("cov: input network traversal: %w", err)
	}
	boundary := make(map[*abc.Object]bool)
	boundary[source.One().Object] = true
	for _, input := range source.ObjectsOfKind(abc.ObjectPI) {
		boundary[input] = true
	}
	allCovers := make(map[*abc.Object]*nodeCover)
	outputs := source.ObjectsOfKind(abc.ObjectPO)
	for pass := 0; pass <= len(source.Objects()); pass++ {
		current := make(map[*abc.Object]*nodeCover)
		state := make(map[*abc.Object]uint8)
		candidates := make(map[*abc.Object]bool)
		var compute func(*abc.Object) (*nodeCover, bool)
		compute = func(object *abc.Object) (*nodeCover, bool) {
			if boundary[object] {
				return &nodeCover{support: []*abc.Object{object}}, true
			}
			if cover := current[object]; cover != nil {
				return cover, true
			}
			if state[object] == 1 {
				return nil, false
			}
			state[object] = 1
			if object.Kind != abc.ObjectNode {
				state[object] = 2
				return nil, false
			}
			fanins := object.Fanins()
			if len(fanins) != 2 {
				state[object] = 2
				return nil, false
			}
			var support []*abc.Object
			children := make([]*abc.Object, 0, len(fanins))
			for _, edge := range fanins {
				children = append(children, edge.Object)
				childCover, ok := compute(edge.Object)
				if !ok || childCover == nil {
					state[object] = 2
					markBoundaryCandidates(children, boundary, current, allCovers, candidates)
					return nil, false
				}
				support = append(support, childCover.support...)
			}
			support = uniqueObjects(support)
			if int32(len(support)) > nFaninMax {
				state[object] = 2
				markBoundaryCandidates(children, boundary, current, allCovers, candidates)
				return nil, false
			}
			truth, err := truthAtCut(object, support)
			if err != nil {
				state[object] = 2
				markBoundaryCandidates(children, boundary, current, allCovers, candidates)
				return nil, false
			}
			cover, err := chooseNodeCover(int32(len(support)), truth, useESOP != 0, useSOP != 0, nCubesMax)
			if err != nil {
				state[object] = 2
				markBoundaryCandidates(children, boundary, current, allCovers, candidates)
				return nil, false
			}
			cover.support = support
			current[object] = cover
			allCovers[object] = cover
			state[object] = 2
			return cover, true
		}

		incomplete := false
		for _, output := range outputs {
			fanins := output.Fanins()
			if len(fanins) != 1 {
				return nil, fmt.Errorf("cov: output %d has %d fanins, want 1", output.ID, len(fanins))
			}
			root := fanins[0].Object
			if boundary[root] {
				continue
			}
			if _, ok := compute(root); !ok {
				incomplete = true
			}
		}
		if !incomplete {
			break
		}
		promoted := make([]*abc.Object, 0, len(candidates))
		for object := range candidates {
			promoted = append(promoted, object)
		}
		sort.Slice(promoted, func(i, j int) bool { return promoted[i].ID < promoted[j].ID })
		if len(promoted) == 0 {
			return nil, fmt.Errorf("cov: no admissible cut boundary satisfies fanin=%d cubes=%d", nFaninMax, nCubesMax)
		}
		for _, object := range promoted {
			boundary[object] = true
		}
		if pass == len(source.Objects()) {
			return nil, fmt.Errorf("cov: cut boundary growth did not converge")
		}
	}

	target := abc.NewNetwork(source.Name)
	target.Type = uint8(abc.ABC_NTK_LOGIC)
	target.Function = uint8(abc.ABC_FUNC_SOP)
	target.Spec = source.Spec
	mapped := make(map[*abc.Object]*abc.Object, len(source.Objects()))
	nameOf := func(old *abc.Object, prefix string) string {
		if name := old.Name(); name != "" {
			return name
		}
		return fmt.Sprintf("%s%d", prefix, old.ID)
	}
	for _, old := range source.ObjectsOfKind(abc.ObjectPI) {
		input := target.NewObject(abc.ObjectPI)
		input.SetName(nameOf(old, "pi_"))
		mapped[old] = input
	}
	mapped[source.One().Object] = target.One().Object
	var build func(*abc.Object) (*abc.Object, error)
	building := make(map[*abc.Object]bool)
	build = func(old *abc.Object) (*abc.Object, error) {
		if object := mapped[old]; object != nil {
			return object, nil
		}
		if building[old] {
			return nil, fmt.Errorf("cov: cycle while deriving object %d", old.ID)
		}
		building[old] = true
		defer delete(building, old)
		cover := allCovers[old]
		if cover == nil {
			return nil, fmt.Errorf("cov: object %d has no computed cut cover", old.ID)
		}
		faninObjects := make([]*abc.Object, len(cover.support))
		for i, sourceFanin := range cover.support {
			fanin, err := build(sourceFanin)
			if err != nil {
				return nil, err
			}
			faninObjects[i] = fanin
		}
		var result *abc.Object
		if cover.useESOP {
			terms := make([]*abc.Object, 0, len(cover.esop))
			for _, term := range cover.esop {
				var termNode *abc.Object
				if term.mask == 0 {
					termNode = target.One().Object
				} else {
					indices := make([]int, 0, bits.OnesCount64(term.mask))
					var row strings.Builder
					for variable := range cover.support {
						bit := uint64(1) << variable
						if term.mask&bit == 0 {
							continue
						}
						indices = append(indices, variable)
						if term.bits&bit != 0 {
							row.WriteByte('1')
						} else {
							row.WriteByte('0')
						}
					}
					row.WriteString(" 1\n")
					termNode = target.NewObject(abc.ObjectNode)
					termNode.Data = row.String()
					for _, variable := range indices {
						if err := target.AddFanin(termNode, faninObjects[variable].Signal()); err != nil {
							return nil, fmt.Errorf("cov: connect ESOP term: %w", err)
						}
					}
				}
				terms = append(terms, termNode)
			}
			switch len(terms) {
			case 0:
				result = target.NewObject(abc.ObjectNode)
				result.Data = func([]bool) bool { return false }
			case 1:
				result = terms[0]
			default:
				result = target.NewObject(abc.ObjectNode)
				result.Data = fmt.Sprintf("%s x\n", strings.Repeat("-", len(terms)))
				for _, term := range terms {
					if err := target.AddFanin(result, term.Signal()); err != nil {
						return nil, fmt.Errorf("cov: connect ESOP XOR: %w", err)
					}
				}
			}
			if result != target.One().Object {
				result.SetName(nameOf(old, "node_"))
			}
		} else {
			result = target.NewObject(abc.ObjectNode)
			result.SetName(nameOf(old, "node_"))
			switch {
			case cover.sopPhase == '0' && len(cover.sop) == 0:
				result = target.One().Object
			case cover.sopPhase == '1' && len(cover.sop) == 0:
				result.Data = func([]bool) bool { return false }
			default:
				result.Data = formatSOP(cover.sop, len(faninObjects), cover.sopPhase)
			}
			for _, fanin := range faninObjects {
				if result == target.One().Object || len(cover.sop) == 0 {
					break
				}
				if err := target.AddFanin(result, fanin.Signal()); err != nil {
					return nil, fmt.Errorf("cov: connect SOP fanin: %w", err)
				}
			}
		}
		mapped[old] = result
		return result, nil
	}
	for _, output := range outputs {
		edge := output.Fanins()[0]
		fanin, err := build(edge.Object)
		if err != nil {
			return nil, err
		}
		po := target.NewObject(abc.ObjectPO)
		po.SetName(nameOf(output, "po_"))
		if err := target.AddFanin(po, abc.Edge{Object: fanin, Complement: edge.Complement}); err != nil {
			return nil, fmt.Errorf("cov: connect output: %w", err)
		}
	}
	if err := target.Check(); err != nil {
		return nil, fmt.Errorf("cov: constructed network invariant: %w", err)
	}
	return target, nil
}

func markBoundaryCandidates(children []*abc.Object, boundary map[*abc.Object]bool, current, all map[*abc.Object]*nodeCover, candidates map[*abc.Object]bool) {
	for _, child := range children {
		if child.Kind == abc.ObjectNode && !boundary[child] && (current[child] != nil || all[child] != nil) {
			candidates[child] = true
		}
	}
}

func uniqueObjects(objects []*abc.Object) []*abc.Object {
	sort.Slice(objects, func(i, j int) bool { return objects[i].ID < objects[j].ID })
	result := objects[:0]
	for _, object := range objects {
		if len(result) == 0 || result[len(result)-1] != object {
			result = append(result, object)
		}
	}
	return result
}

func truthAtCut(root *abc.Object, support []*abc.Object) ([]bool, error) {
	if len(support) > 16 {
		return nil, fmt.Errorf("cov: support size %d exceeds exact truth-table limit", len(support))
	}
	limit := 1 << len(support)
	truth := make([]bool, limit)
	values := make(map[*abc.Object]bool, len(support))
	var evaluate func(*abc.Object, map[*abc.Object]bool, map[*abc.Object]bool) (bool, error)
	evaluate = func(object *abc.Object, overrides, memo map[*abc.Object]bool) (bool, error) {
		if value, ok := overrides[object]; ok {
			return value, nil
		}
		if value, ok := memo[object]; ok {
			return value, nil
		}
		switch object.Kind {
		case abc.ObjectConst1:
			return true, nil
		case abc.ObjectPI:
			return false, fmt.Errorf("cov: primary input %d is outside cut support", object.ID)
		case abc.ObjectNode:
			fanins := object.Fanins()
			if len(fanins) != 2 {
				return false, fmt.Errorf("cov: AIG node %d has %d fanins", object.ID, len(fanins))
			}
			args := [2]bool{}
			for i, edge := range fanins {
				value, err := evaluate(edge.Object, overrides, memo)
				if err != nil {
					return false, err
				}
				args[i] = value != edge.Complement
			}
			value := args[0] && args[1]
			memo[object] = value
			return value, nil
		default:
			return false, fmt.Errorf("cov: unsupported object kind %d in AIG cone", object.Kind)
		}
	}
	for assignment := 0; assignment < limit; assignment++ {
		for variable, object := range support {
			values[object] = assignment&(1<<variable) != 0
		}
		value, err := evaluate(root, values, make(map[*abc.Object]bool))
		if err != nil {
			return nil, err
		}
		truth[assignment] = value
	}
	return truth, nil
}

func chooseNodeCover(variables int32, truth []bool, useESOP, useSOP bool, maxCubes int32) (*nodeCover, error) {
	if !useESOP && !useSOP {
		return nil, fmt.Errorf("cov: no cover mode enabled")
	}
	on := make([]uint64, 0)
	off := make([]uint64, 0)
	for minterm, value := range truth {
		if value {
			on = append(on, uint64(minterm))
		} else {
			off = append(off, uint64(minterm))
		}
	}
	result := &nodeCover{}
	var sopPos, sopNeg []minimize.Cube
	if useSOP {
		sopPos, _ = minimize.Minimize(int(variables), on, nil)
		sopNeg, _ = minimize.Minimize(int(variables), off, nil)
	}
	if useESOP {
		if variables <= 4 {
			result.esop = minimizeESOPExact(int(variables), truth)
		} else {
			result.esop = algebraicNormalForm(int(variables), truth)
		}
	}
	validPos := useSOP && sopPos != nil && len(sopPos) <= int(maxCubes)
	validNeg := useSOP && sopNeg != nil && len(sopNeg) <= int(maxCubes)
	validESOP := useESOP && len(result.esop) <= int(maxCubes)
	if !validPos && !validNeg && !validESOP {
		return nil, fmt.Errorf("cov: no bounded cover for support %d", variables)
	}
	// Match covBuild.c's cube-count selection: positive SOP wins ties, then
	// negative SOP, while ESOP must be strictly smaller than both.
	posCount, negCount, esopCount := int(^uint(0)>>1), int(^uint(0)>>1), int(^uint(0)>>1)
	if validPos {
		posCount = len(sopPos)
	}
	if validNeg {
		negCount = len(sopNeg)
	}
	if validESOP {
		esopCount = len(result.esop)
	}
	if validPos && posCount <= negCount && posCount <= esopCount {
		result.sop, result.sopPhase = sopPos, '1'
	} else if validNeg && negCount <= posCount && negCount <= esopCount {
		result.sop, result.sopPhase = sopNeg, '0'
	} else if validESOP {
		result.useESOP = true
	} else {
		return nil, fmt.Errorf("cov: no bounded cover for support %d", variables)
	}
	return result, nil
}

func formatSOP(cubes []minimize.Cube, variables int, phase byte) string {
	if len(cubes) == 0 {
		pattern := strings.Repeat("-", variables)
		if variables == 0 {
			pattern = ""
		}
		return pattern + " 0\n"
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
		out.WriteByte(' ')
		out.WriteByte(phase)
		out.WriteByte('\n')
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
	result, err := deriveAIGToSOP(source, p.nFaninMax, p.nCubesMax, p.fUseEsop, p.fUseSop)
	if err != nil {
		return nil
	}
	return result
}
