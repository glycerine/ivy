package abc

// Structural and functional cut decomposition checks from src/map/if/ifCheck.c.

type IfCutCheckOptions struct {
	MaxObjects        int
	MaxMappingEntries int
	MaxWords          int
	Stop              func() bool
}

type IfCutDecomposition struct {
	DirectLeaf         bool
	Structural         bool
	StructuralConflict bool
	Functional         bool
	FunctionalLeaf     AigID
	FunctionalCofactor int // -1 when no constant cofactor exists
}

type IfCutCheckStatistics struct {
	Calls, Structural, StructuralConflicts, Functional int
}

type IfCutChecker struct {
	statistics IfCutCheckStatistics
}

func (checker *IfCutChecker) Statistics() IfCutCheckStatistics {
	if checker == nil {
		return IfCutCheckStatistics{}
	}
	return checker.statistics
}

func checkIfCutOptions(options IfCutCheckOptions) (IfCutCheckOptions, error) {
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxWords < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if options.MaxWords == 0 {
		options.MaxWords = 1 << 20
	}
	return options, nil
}

func (graph *AigManager) checkIfCutBoundary(leaves []AigID, options IfCutCheckOptions) (IfCutCheckOptions, []bool, error) {
	options, err := checkIfCutOptions(options)
	if err != nil {
		return options, nil, err
	}
	if graph == nil || len(leaves) == 0 || len(leaves) > 7 {
		return options, nil, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return options, nil, err
	}
	if len(graph.objects) > options.MaxMappingEntries {
		return options, nil, ErrLimit
	}
	boundary := make([]bool, len(graph.objects))
	for _, leaf := range leaves {
		if leaf == 0 || uint64(leaf) >= uint64(len(graph.objects)) || boundary[leaf] {
			return options, nil, ErrDimension
		}
		typeOfLeaf := graph.objects[leaf].Type
		if typeOfLeaf != AigCI && typeOfLeaf != AigAND {
			return options, nil, ErrInvalidLiteral
		}
		boundary[leaf] = true
	}
	if options.Stop != nil && options.Stop() {
		return options, nil, ErrCancelled
	}
	return options, boundary, nil
}

func (graph *AigManager) checkIfCut(root AigID, leaves []AigID, options IfCutCheckOptions) (IfCutCheckOptions, []bool, error) {
	options, boundary, err := graph.checkIfCutBoundary(leaves, options)
	if err != nil {
		return options, nil, err
	}
	if uint64(root) >= uint64(len(graph.objects)) || graph.objects[root].Type != AigAND {
		return options, nil, ErrInvalidLiteral
	}
	return options, boundary, nil
}

// IfCutReach preserves If_ManCutReach_rec's cut-boundary traversal without
// borrowing object marks. Complemented edges do not affect reachability.
func (graph *AigManager) IfCutReach(path, leaf AigID, leaves []AigID, options IfCutCheckOptions) (bool, error) {
	options, boundary, err := graph.checkIfCutBoundary(leaves, options)
	if err != nil {
		return false, err
	}
	if uint64(path) >= uint64(len(graph.objects)) || (graph.objects[path].Type != AigCI && graph.objects[path].Type != AigAND) {
		return false, ErrInvalidLiteral
	}
	if uint64(leaf) >= uint64(len(boundary)) || !boundary[leaf] {
		return false, ErrDimension
	}
	seen := make([]bool, len(graph.objects))
	stack := []AigID{path}
	visited := 0
	for len(stack) != 0 {
		if options.Stop != nil && options.Stop() {
			return false, ErrCancelled
		}
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if id == leaf {
			return true, nil
		}
		if seen[id] || boundary[id] {
			continue
		}
		seen[id] = true
		object := graph.objects[id]
		if object.Type != AigAND {
			return false, ErrDimension
		}
		visited++
		if visited > options.MaxObjects {
			return false, ErrLimit
		}
		// Push fanin one first so fanin zero is visited first, as in C.
		stack = append(stack, object.Fanin1.ID(), object.Fanin0.ID())
	}
	return false, nil
}

// IfCutCofactorConstant preserves If_ManCutTruthCheck by fixing one boundary
// variable and determining whether the resulting function is constant.
func (graph *AigManager) IfCutCofactorConstant(root AigID, leaves []AigID, leafIndex int, value bool, options IfCutCheckOptions) (bool, error) {
	options, _, err := graph.checkIfCut(root, leaves, options)
	if err != nil {
		return false, err
	}
	if leafIndex < 0 || leafIndex >= len(leaves) {
		return false, ErrDimension
	}
	truth, err := graph.CutTruth(AigLit(root)<<1, leaves, AigCutTruthOptions{
		MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries,
		MaxWords: options.MaxWords, Stop: options.Stop,
	})
	if err != nil {
		return false, err
	}
	cofactor, err := truth.Table.Cofactor(leafIndex, value)
	if err != nil {
		return false, err
	}
	return cofactor.IsConstant(false) || cofactor.IsConstant(true), nil
}

// Analyze preserves If_ManCutCheck's leaf order and zero-before-one cofactor
// scan. Statistics replace its process-global counters and stdout report.
func (checker *IfCutChecker) Analyze(graph *AigManager, root AigID, leaves []AigID, options IfCutCheckOptions) (IfCutDecomposition, error) {
	result := IfCutDecomposition{FunctionalLeaf: InvalidAigID, FunctionalCofactor: -1}
	if checker == nil {
		return result, ErrDimension
	}
	options, _, err := graph.checkIfCut(root, leaves, options)
	if err != nil {
		return result, err
	}
	object := graph.objects[root]
	for _, leaf := range leaves {
		path := InvalidAigID
		switch leaf {
		case object.Fanin0.ID():
			path = object.Fanin1.ID()
		case object.Fanin1.ID():
			path = object.Fanin0.ID()
		}
		if path == InvalidAigID {
			continue
		}
		result.DirectLeaf = true
		reachable, reachErr := graph.IfCutReach(path, leaf, leaves, options)
		if reachErr != nil {
			return IfCutDecomposition{FunctionalLeaf: InvalidAigID, FunctionalCofactor: -1}, reachErr
		}
		result.Structural = !reachable
		result.StructuralConflict = reachable
		break
	}
	for index, leaf := range leaves {
		for cofactor := 0; cofactor < 2; cofactor++ {
			constant, cofactorErr := graph.IfCutCofactorConstant(root, leaves, index, cofactor != 0, options)
			if cofactorErr != nil {
				return IfCutDecomposition{FunctionalLeaf: InvalidAigID, FunctionalCofactor: -1}, cofactorErr
			}
			if constant {
				result.Functional = true
				result.FunctionalLeaf = leaf
				result.FunctionalCofactor = cofactor
				break
			}
		}
		if result.Functional {
			break
		}
	}
	checker.statistics.Calls++
	if result.Structural {
		checker.statistics.Structural++
	}
	if result.StructuralConflict {
		checker.statistics.StructuralConflicts++
	}
	if result.Functional {
		checker.statistics.Functional++
	}
	return result, nil
}
