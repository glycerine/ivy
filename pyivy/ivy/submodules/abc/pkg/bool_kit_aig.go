package abc

// KIT decomposition-graph to AIG translation from src/bool/kit/kitAig.c.

type KitAigOptions struct {
	MaxObjects int
	MaxWork    int
	MaxBytes   int
	Stop       func() bool
}

func normalizeKitAigOptions(options KitAigOptions) (KitAigOptions, error) {
	if options.MaxObjects < 0 || options.MaxWork < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxWork == 0 {
		options.MaxWork = 1 << 20
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 128 << 20
	}
	return options, nil
}

type kitAigWork struct {
	options KitAigOptions
	used    int
}

func (work *kitAigWork) take() error {
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	if work.used >= work.options.MaxWork {
		return ErrLimit
	}
	work.used++
	return nil
}

// KitGraphToAig returns an owned clone of source with graph strashed into it.
// Returning a new manager makes every error atomic while preserving the native
// fanin and complemented-edge semantics.
func KitGraphToAig(source *AigManager, fanins []AigLit, graph *KitGraph, options KitAigOptions) (*AigManager, AigLit, error) {
	if source == nil || graph == nil {
		return nil, InvalidAigLit, ErrDimension
	}
	options, err := normalizeKitAigOptions(options)
	if err != nil {
		return nil, InvalidAigLit, err
	}
	if err = source.Check(); err != nil {
		return nil, InvalidAigLit, err
	}
	graph.mu.RLock()
	defer graph.mu.RUnlock()
	if len(fanins) != graph.leaves || (!graph.constant && !graph.validEdgeLocked(graph.root)) {
		return nil, InvalidAigLit, ErrDimension
	}
	if len(graph.nodes) < graph.leaves || len(graph.nodes) > graph.maxNodes {
		return nil, InvalidAigLit, ErrInvariant
	}
	internal := len(graph.nodes) - graph.leaves
	if source.ObjectCount() > options.MaxObjects-internal {
		return nil, InvalidAigLit, ErrLimit
	}
	// Conservative retained clone, hash and mapping accounting.
	if source.ObjectCount() > options.MaxBytes/96 || len(graph.nodes) > (options.MaxBytes-96*source.ObjectCount())/16 {
		return nil, InvalidAigLit, ErrLimit
	}
	work := kitAigWork{options: options}
	if err = work.take(); err != nil {
		return nil, InvalidAigLit, err
	}
	for _, fanin := range fanins {
		if err = work.take(); err != nil {
			return nil, InvalidAigLit, err
		}
		if !source.validSignal(fanin) {
			return nil, InvalidAigLit, ErrInvalidLiteral
		}
	}
	target := source.Clone()
	if graph.constant {
		result := AigConst1
		if graph.root.Complement {
			result = result.Not()
		}
		return target, result, nil
	}
	mapping := make([]AigLit, len(graph.nodes))
	copy(mapping, fanins)
	for index := graph.leaves; index < len(graph.nodes); index++ {
		if err = work.take(); err != nil {
			return nil, InvalidAigLit, err
		}
		node := graph.nodes[index]
		if node.first.Node < 0 || node.first.Node >= index || node.second.Node < 0 || node.second.Node >= index {
			return nil, InvalidAigLit, ErrInvariant
		}
		first := mapping[node.first.Node].NotIf(node.first.Complement)
		second := mapping[node.second.Node].NotIf(node.second.Complement)
		mapping[index], err = target.And(first, second)
		if err != nil {
			return nil, InvalidAigLit, err
		}
	}
	result := mapping[graph.root.Node].NotIf(graph.root.Complement)
	return target, result, nil
}
