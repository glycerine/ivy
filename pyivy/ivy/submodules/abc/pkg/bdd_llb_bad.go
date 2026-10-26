package abc

// Bad-state construction and PI abstraction from src/bdd/llb/llb2Bad.c.

type LLBBadOptions struct {
	MaxObjects, MaxVisits, MaxOperations, MaxBytes int
	Stop                                           func() bool
}

func (options LLBBadOptions) traversal() AigTraversalOptions {
	return AigTraversalOptions{MaxObjects: options.MaxObjects, MaxVisits: options.MaxVisits, Stop: options.Stop}
}

func (options LLBBadOptions) bdd() ExtraBDDOptions {
	return ExtraBDDOptions{MaxOperations: options.MaxOperations, MaxBytes: options.MaxBytes, Stop: options.Stop}
}

func llbMappedBDD(mapping []ExtraBDDLit, ready []bool, literal AigLit) (ExtraBDDLit, error) {
	if !literal.Valid() || uint64(literal.ID()) >= uint64(len(mapping)) || !ready[literal.ID()] {
		return 0, ErrInvalidLiteral
	}
	mapped := mapping[literal.ID()]
	if literal.Complemented() {
		mapped = mapped.Not()
	}
	return mapped, nil
}

// LLBComputeBad constructs the disjunction of all primary outputs. BDD
// variables follow the native LLB order: latch outputs first, then primary
// inputs, even though AigManager stores primary inputs first.
func LLBComputeBad(graph *AigManager, manager *ExtraBDDManager, options LLBBadOptions) (ExtraBDDLit, error) {
	if manager == nil {
		return 0, ErrDimension
	}
	if options.MaxObjects < 0 || options.MaxVisits < 0 || options.MaxOperations < 0 || options.MaxBytes < 0 {
		return 0, ErrDimension
	}
	if err := manager.check(); err != nil {
		return 0, err
	}
	if graph == nil || manager.variables != graph.InputCount() {
		return 0, ErrDimension
	}
	roots := graph.PrimaryOutputs()
	order, err := graph.DFSNodes(roots, options.traversal())
	if err != nil {
		return 0, err
	}
	return manager.runBooleanOperation(nil, options.bdd(), func(work *extraBDDWork) (ExtraBDDLit, error) {
		mapping := make([]ExtraBDDLit, len(graph.objects))
		ready := make([]bool, len(graph.objects))
		mapping[0], ready[0] = ExtraBDDTrue, true
		primaryInputs := graph.PrimaryInputCount()
		for index, id := range graph.cis {
			variable := index - primaryInputs
			if index < primaryInputs {
				variable = graph.RegisterCount() + index
			}
			literal, variableErr := manager.Variable(variable)
			if variableErr != nil {
				return 0, variableErr
			}
			mapping[id], ready[id] = literal, true
		}
		for _, id := range order {
			object := graph.objects[id]
			switch object.Type {
			case AigCI:
				continue
			case AigBuffer:
				mapped, mapErr := llbMappedBDD(mapping, ready, object.Fanin0)
				if mapErr != nil {
					return 0, mapErr
				}
				mapping[id], ready[id] = mapped, true
			case AigAND:
				first, mapErr := llbMappedBDD(mapping, ready, object.Fanin0)
				if mapErr != nil {
					return 0, mapErr
				}
				second, mapErr := llbMappedBDD(mapping, ready, object.Fanin1)
				if mapErr != nil {
					return 0, mapErr
				}
				mapped, mapErr := manager.andRecursive(first, second, work)
				if mapErr != nil {
					return 0, mapErr
				}
				mapping[id], ready[id] = mapped, true
			default:
				return 0, ErrInvariant
			}
		}
		result := ExtraBDDFalse
		for _, outputID := range roots {
			driver := graph.objects[outputID].Fanin0
			mapped, mapErr := llbMappedBDD(mapping, ready, driver)
			if mapErr != nil {
				return 0, mapErr
			}
			conjunction, mapErr := manager.andRecursive(result.Not(), mapped.Not(), work)
			if mapErr != nil {
				return 0, mapErr
			}
			result = conjunction.Not()
		}
		return result, nil
	})
}

// LLBQuantifyPIs existentially abstracts every primary-input variable from a
// BDD produced with the LLB variable order.
func LLBQuantifyPIs(graph *AigManager, manager *ExtraBDDManager, root ExtraBDDLit, options LLBBadOptions) (ExtraBDDLit, error) {
	if manager == nil {
		return 0, ErrDimension
	}
	if options.MaxObjects < 0 || options.MaxVisits < 0 || options.MaxOperations < 0 || options.MaxBytes < 0 {
		return 0, ErrDimension
	}
	if err := manager.checkLiteral(root); err != nil {
		return 0, err
	}
	if graph == nil || manager.variables != graph.InputCount() {
		return 0, ErrDimension
	}
	if _, err := checkAigTraversal(graph, options.traversal()); err != nil {
		return 0, err
	}
	variables := make([]int, graph.PrimaryInputCount())
	for index := range variables {
		variables[index] = graph.RegisterCount() + index
	}
	return manager.ExistAbstract(root, variables, options.bdd())
}
