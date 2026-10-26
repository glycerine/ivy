package abc

// Cone transfer and CI composition from aig/aig/aigDfs.c.

// AigRewriteOptions bounds both the input manager and the committed result.
// Zero values use the traversal defaults. Failed operations leave every
// caller-owned manager unchanged.
type AigRewriteOptions struct {
	MaxObjects int // zero: one million source and result objects
	MaxVisits  int // zero: eight million traversal/build events
	Stop       func() bool
}

func checkAigRewrite(g *AigManager, o AigRewriteOptions) (AigTraversalOptions, error) {
	return checkAigTraversal(g, AigTraversalOptions{
		MaxObjects: o.MaxObjects,
		MaxVisits:  o.MaxVisits,
		Stop:       o.Stop,
	})
}

func aigRewriteOrder(g *AigManager, root AigLit, o AigTraversalOptions) ([]AigID, error) {
	if !g.validSignal(root) {
		return nil, ErrInvalidLiteral
	}
	visit := aigTraversalVisit(o)
	order, err := aigDFSOrder(g, []AigID{root.ID()}, make([]bool, len(g.objects)), false, visit, nil)
	if err != nil {
		return nil, err
	}
	for _, id := range order {
		switch g.objects[id].Type {
		case AigConstant1, AigCI, AigAND:
		default:
			return nil, ErrDimension
		}
	}
	return order, nil
}

func aigRewriteStep(o AigTraversalOptions, visits *int) error {
	if o.Stop != nil && o.Stop() {
		return ErrCancelled
	}
	if *visits == o.MaxVisits {
		return ErrLimit
	}
	*visits++
	return nil
}

func checkAigRewriteSize(g *AigManager, maximum int) error {
	if len(g.objects) > maximum {
		return ErrLimit
	}
	return nil
}

// TransferAigCone implements Aig_Transfer. The first inputCount source CIs
// map by index to destination CIs; IthVar-compatible missing destination CIs
// are created. The destination is mutated only after complete success.
func TransferAigCone(source, destination *AigManager, root AigLit, inputCount int, options AigRewriteOptions) (AigLit, error) {
	if source == nil || destination == nil || inputCount < 0 {
		return InvalidAigLit, ErrDimension
	}
	checked, err := checkAigRewrite(source, options)
	if err != nil {
		return InvalidAigLit, err
	}
	if err = destination.Check(); err != nil {
		return InvalidAigLit, err
	}
	if !source.validSignal(root) {
		return InvalidAigLit, ErrInvalidLiteral
	}
	if source == destination {
		return root, nil
	}
	if root.ID() == 0 {
		return AigConst1.NotIf(root.Complemented()), nil
	}
	order, err := aigRewriteOrder(source, root, checked)
	if err != nil {
		return InvalidAigLit, err
	}
	ciIndex := make([]int, len(source.objects))
	for i := range ciIndex {
		ciIndex[i] = -1
	}
	for i, id := range source.cis {
		ciIndex[id] = i
	}
	for _, id := range order {
		if source.objects[id].Type == AigCI && ciIndex[id] >= inputCount {
			return InvalidAigLit, ErrDimension
		}
	}

	work := destination.Clone()
	if err = checkAigRewriteSize(work, checked.MaxObjects); err != nil {
		return InvalidAigLit, err
	}
	mapping := make([]AigLit, len(source.objects))
	for i := range mapping {
		mapping[i] = InvalidAigLit
	}
	mapping[0] = AigConst1
	visits := 0
	for i, id := range source.cis {
		if i == inputCount {
			break
		}
		if err = aigRewriteStep(checked, &visits); err != nil {
			return InvalidAigLit, err
		}
		mapping[id], err = work.IthVar(i)
		if err != nil {
			return InvalidAigLit, err
		}
		if err = checkAigRewriteSize(work, checked.MaxObjects); err != nil {
			return InvalidAigLit, err
		}
	}
	for _, id := range order {
		if err = aigRewriteStep(checked, &visits); err != nil {
			return InvalidAigLit, err
		}
		obj := source.objects[id]
		switch obj.Type {
		case AigConstant1, AigCI:
			if mapping[id] == InvalidAigLit {
				return InvalidAigLit, ErrDimension
			}
		case AigAND:
			a, b := mapping[obj.Fanin0.ID()], mapping[obj.Fanin1.ID()]
			if a == InvalidAigLit || b == InvalidAigLit {
				return InvalidAigLit, ErrInvariant
			}
			mapping[id], err = work.And(a.NotIf(obj.Fanin0.Complemented()), b.NotIf(obj.Fanin1.Complemented()))
			if err != nil {
				return InvalidAigLit, err
			}
			if err = checkAigRewriteSize(work, checked.MaxObjects); err != nil {
				return InvalidAigLit, err
			}
		}
	}
	result := mapping[root.ID()].NotIf(root.Complemented())
	if !work.validSignal(result) {
		return InvalidAigLit, ErrInvariant
	}
	if err = work.Check(); err != nil {
		return InvalidAigLit, err
	}
	*destination = *work
	return result, nil
}

// ComposeAigInput implements Aig_Compose by replacing the indexed CI in root
// with function. It may append structurally hashed nodes to manager and commits
// them atomically. The replacement may itself depend on the replaced CI.
func (manager *AigManager) ComposeAigInput(root, function AigLit, inputIndex int, options AigRewriteOptions) (AigLit, error) {
	if manager == nil || inputIndex < 0 {
		return InvalidAigLit, ErrDimension
	}
	checked, err := checkAigRewrite(manager, options)
	if err != nil {
		return InvalidAigLit, err
	}
	if inputIndex >= len(manager.cis) {
		return InvalidAigLit, ErrDimension
	}
	if !manager.validSignal(root) || !manager.validSignal(function) {
		return InvalidAigLit, ErrInvalidLiteral
	}
	order, err := aigRewriteOrder(manager, root, checked)
	if err != nil {
		return InvalidAigLit, err
	}
	work := manager.Clone()
	mapping := make([]AigLit, len(manager.objects))
	for i := range mapping {
		mapping[i] = InvalidAigLit
	}
	mapping[0] = AigConst1
	target := manager.cis[inputIndex]
	visits := 0
	for _, id := range order {
		if err = aigRewriteStep(checked, &visits); err != nil {
			return InvalidAigLit, err
		}
		obj := manager.objects[id]
		switch obj.Type {
		case AigConstant1:
		case AigCI:
			if id == target {
				mapping[id] = function
			} else {
				mapping[id] = AigLit(id) << 1
			}
		case AigAND:
			a, b := mapping[obj.Fanin0.ID()], mapping[obj.Fanin1.ID()]
			if a == InvalidAigLit || b == InvalidAigLit {
				return InvalidAigLit, ErrInvariant
			}
			mapping[id], err = work.And(a.NotIf(obj.Fanin0.Complemented()), b.NotIf(obj.Fanin1.Complemented()))
			if err != nil {
				return InvalidAigLit, err
			}
			if err = checkAigRewriteSize(work, checked.MaxObjects); err != nil {
				return InvalidAigLit, err
			}
		}
	}
	result := mapping[root.ID()].NotIf(root.Complemented())
	if !work.validSignal(result) {
		return InvalidAigLit, ErrInvariant
	}
	if err = work.Check(); err != nil {
		return InvalidAigLit, err
	}
	*manager = *work
	return result, nil
}
