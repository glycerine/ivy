package abc

// CI substitution from src/aig/aig/aigScl.c:Aig_ManRemap.

// RemapInputs rebuilds graph after replacing each CI with a signed source CI
// or constant literal. The returned graph is independently owned; all original
// CI slots and the register partition remain until a later sequential cleanup.
func (graph *AigManager) RemapInputs(replacements []AigLit, options AigRewriteOptions) (*AigManager, error) {
	if graph == nil || len(replacements) != len(graph.cis) {
		return nil, ErrDimension
	}
	checked, err := checkAigRewrite(graph, options)
	if err != nil {
		return nil, err
	}
	if graph.equivs != nil {
		return nil, ErrUnsupported
	}
	ciIndex := make([]int, len(graph.objects))
	for index := range ciIndex {
		ciIndex[index] = -1
	}
	for index, id := range graph.cis {
		ciIndex[id] = index
	}
	for _, replacement := range replacements {
		if !graph.validSignal(replacement) {
			return nil, ErrInvalidLiteral
		}
		if replacement.ID() != 0 && graph.objects[replacement.ID()].Type != AigCI {
			return nil, ErrDimension
		}
	}
	target := NewAigManager()
	target.Name, target.Spec = graph.Name, graph.Spec
	objectMap := make([]AigLit, len(graph.objects))
	for index := range objectMap {
		objectMap[index] = InvalidAigLit
	}
	objectMap[0] = AigConst1
	createdInputs := make([]AigLit, len(graph.cis))
	visits := 0
	for index := range graph.cis {
		if err = aigRewriteStep(checked, &visits); err != nil {
			return nil, err
		}
		createdInputs[index], err = target.CreateCI()
		if err != nil {
			return nil, err
		}
		if err = checkAigRewriteSize(target, checked.MaxObjects); err != nil {
			return nil, err
		}
	}
	for index, id := range graph.cis {
		replacement := replacements[index]
		base := AigConst1
		if replacement.ID() != 0 {
			ordinal := ciIndex[replacement.ID()]
			if ordinal < 0 {
				return nil, ErrInvariant
			}
			base = createdInputs[ordinal]
		}
		objectMap[id] = base.NotIf(replacement.Complemented())
	}
	for id, object := range graph.objects {
		if err = aigRewriteStep(checked, &visits); err != nil {
			return nil, err
		}
		switch object.Type {
		case AigNone, AigConstant1, AigCI, AigCO:
			continue
		case AigBuffer:
			fanin := objectMap[object.Fanin0.ID()]
			if fanin == InvalidAigLit {
				return nil, ErrInvariant
			}
			objectMap[id] = fanin.NotIf(object.Fanin0.Complemented())
		case AigAND:
			left, right := objectMap[object.Fanin0.ID()], objectMap[object.Fanin1.ID()]
			if left == InvalidAigLit || right == InvalidAigLit {
				return nil, ErrInvariant
			}
			objectMap[id], err = target.And(left.NotIf(object.Fanin0.Complemented()), right.NotIf(object.Fanin1.Complemented()))
			if err != nil {
				return nil, err
			}
		default:
			return nil, ErrUnsupported
		}
		if err = checkAigRewriteSize(target, checked.MaxObjects); err != nil {
			return nil, err
		}
	}
	for _, id := range graph.cos {
		if err = aigRewriteStep(checked, &visits); err != nil {
			return nil, err
		}
		driver := graph.objects[id].Fanin0
		mapped := objectMap[driver.ID()]
		if mapped == InvalidAigLit {
			return nil, ErrInvariant
		}
		if _, err = target.CreateCO(mapped.NotIf(driver.Complemented())); err != nil {
			return nil, err
		}
		if err = checkAigRewriteSize(target, checked.MaxObjects); err != nil {
			return nil, err
		}
	}
	if err = target.SetRegisterCount(graph.registers); err != nil {
		return nil, err
	}
	if err = target.SetConstraintCount(graph.constraints); err != nil {
		return nil, err
	}
	if err = target.Check(); err != nil {
		return nil, err
	}
	return target, nil
}
