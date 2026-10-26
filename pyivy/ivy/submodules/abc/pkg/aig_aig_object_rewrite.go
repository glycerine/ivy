package abc

// Owned buffer propagation and object replacement from src/aig/aig/aigObj.c.

// AigObjectRewrite is an independently owned, topologically ordered rewrite.
// ObjectMap maps every old live object to its new literal-shaped object ID.
// Entries for old signal objects include their rewrite polarity; CO entries are
// uncomplemented IDs and tombstones remain InvalidAigLit.
type AigObjectRewrite struct {
	Graph           *AigManager
	ObjectMap       []AigLit
	BypassedBuffers int
}

type aigObjectRewriteFrame struct {
	id   AigID
	exit bool
}

// PropagateBuffers implements the completed effect of
// Aig_ManPropagateBuffers. It bypasses every buffer, rebuilds every live
// internal object in dependency order and preserves CI/CO and sequential
// partitions. Rebuilding is necessary because native propagation temporarily
// permits forward-pointing buffers, while AigManager always remains
// topologically ordered.
func (graph *AigManager) PropagateBuffers(options AigRewriteOptions) (AigObjectRewrite, error) {
	return graph.rewriteObjects(InvalidAigID, InvalidAigLit, false, options)
}

// ReplaceObject implements the completed semantic effect of Aig_ObjReplace
// followed by Aig_ManPropagateBuffers. Every use of old is redirected to
// replacement, all buffers are bypassed, and the returned graph is
// independently owned. The replacement must not be a buffer and must not
// transitively depend on old. ObjectMap makes the unavoidable ID remapping
// explicit instead of exposing native forward references.
func (graph *AigManager) ReplaceObject(old AigID, replacement AigLit, options AigRewriteOptions) (AigObjectRewrite, error) {
	return graph.rewriteObjects(old, replacement, true, options)
}

func (graph *AigManager) rewriteObjects(old AigID, replacement AigLit, replace bool, options AigRewriteOptions) (AigObjectRewrite, error) {
	checked, err := checkAigRewrite(graph, options)
	if err != nil {
		return AigObjectRewrite{}, err
	}
	if graph.equivs != nil {
		return AigObjectRewrite{}, ErrUnsupported
	}
	if replace {
		if uint64(old) >= uint64(len(graph.objects)) {
			return AigObjectRewrite{}, ErrInvalidLiteral
		}
		oldType := graph.objects[old].Type
		if oldType != AigAND && oldType != AigBuffer {
			return AigObjectRewrite{}, ErrUnsupported
		}
		if !graph.validSignal(replacement) {
			return AigObjectRewrite{}, ErrInvalidLiteral
		}
		if replacement.ID() == old {
			return AigObjectRewrite{}, ErrInvariant
		}
		replacementType := graph.objects[replacement.ID()].Type
		if replacementType == AigBuffer {
			return AigObjectRewrite{}, ErrUnsupported
		}
	}

	target := NewAigManager()
	target.Name, target.Spec = graph.Name, graph.Spec
	mapping := make([]AigLit, len(graph.objects))
	for index := range mapping {
		mapping[index] = InvalidAigLit
	}
	state := make([]uint8, len(graph.objects))
	mapping[0], state[0] = AigConst1, 2
	visits := 0
	for _, id := range graph.cis {
		if err = aigRewriteStep(checked, &visits); err != nil {
			return AigObjectRewrite{}, err
		}
		mapping[id], err = target.CreateCI()
		if err != nil {
			return AigObjectRewrite{}, err
		}
		state[id] = 2
		if err = checkAigRewriteSize(target, checked.MaxObjects); err != nil {
			return AigObjectRewrite{}, err
		}
	}
	for id, object := range graph.objects {
		if object.Type == AigNone {
			state[id] = 2
		}
	}

	ensure := func(root AigID) error {
		stack := []aigObjectRewriteFrame{{id: root}}
		for len(stack) != 0 {
			if err := aigRewriteStep(checked, &visits); err != nil {
				return err
			}
			frame := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if uint64(frame.id) >= uint64(len(graph.objects)) {
				return ErrInvalidLiteral
			}
			if frame.exit {
				if replace && frame.id == old {
					mapped := mapping[replacement.ID()]
					if mapped == InvalidAigLit {
						return ErrInvariant
					}
					mapping[frame.id] = mapped.NotIf(replacement.Complemented())
					state[frame.id] = 2
					continue
				}
				object := graph.objects[frame.id]
				switch object.Type {
				case AigBuffer:
					mapped := mapping[object.Fanin0.ID()]
					if mapped == InvalidAigLit {
						return ErrInvariant
					}
					mapping[frame.id] = mapped.NotIf(object.Fanin0.Complemented())
				case AigAND:
					left, right := mapping[object.Fanin0.ID()], mapping[object.Fanin1.ID()]
					if left == InvalidAigLit || right == InvalidAigLit {
						return ErrInvariant
					}
					var buildErr error
					mapping[frame.id], buildErr = target.And(
						left.NotIf(object.Fanin0.Complemented()),
						right.NotIf(object.Fanin1.Complemented()),
					)
					if buildErr != nil {
						return buildErr
					}
					if err := checkAigRewriteSize(target, checked.MaxObjects); err != nil {
						return err
					}
				default:
					return ErrInvariant
				}
				state[frame.id] = 2
				continue
			}
			if state[frame.id] == 2 {
				continue
			}
			if state[frame.id] == 1 {
				return ErrInvariant
			}
			state[frame.id] = 1
			stack = append(stack, aigObjectRewriteFrame{id: frame.id, exit: true})
			if replace && frame.id == old {
				stack = append(stack, aigObjectRewriteFrame{id: replacement.ID()})
				continue
			}
			object := graph.objects[frame.id]
			switch object.Type {
			case AigBuffer:
				stack = append(stack, aigObjectRewriteFrame{id: object.Fanin0.ID()})
			case AigAND:
				// Push fanin one first so fanin zero is visited first.
				stack = append(stack,
					aigObjectRewriteFrame{id: object.Fanin1.ID()},
					aigObjectRewriteFrame{id: object.Fanin0.ID()},
				)
			default:
				return ErrInvariant
			}
		}
		return nil
	}

	bypassed := 0
	for id, object := range graph.objects {
		switch object.Type {
		case AigBuffer:
			bypassed++
			fallthrough
		case AigAND:
			if err = ensure(AigID(id)); err != nil {
				return AigObjectRewrite{}, err
			}
		}
	}
	for _, id := range graph.cos {
		if err = aigRewriteStep(checked, &visits); err != nil {
			return AigObjectRewrite{}, err
		}
		driver := graph.objects[id].Fanin0
		mapped := mapping[driver.ID()]
		if mapped == InvalidAigLit {
			return AigObjectRewrite{}, ErrInvariant
		}
		newID, createErr := target.CreateCO(mapped.NotIf(driver.Complemented()))
		if createErr != nil {
			return AigObjectRewrite{}, createErr
		}
		mapping[id] = AigLit(newID) << 1
		if err = checkAigRewriteSize(target, checked.MaxObjects); err != nil {
			return AigObjectRewrite{}, err
		}
	}
	if err = target.SetRegisterCount(graph.registers); err != nil {
		return AigObjectRewrite{}, err
	}
	if err = target.SetConstraintCount(graph.constraints); err != nil {
		return AigObjectRewrite{}, err
	}
	if err = target.Check(); err != nil {
		return AigObjectRewrite{}, err
	}
	return AigObjectRewrite{Graph: target, ObjectMap: mapping, BypassedBuffers: bypassed}, nil
}
