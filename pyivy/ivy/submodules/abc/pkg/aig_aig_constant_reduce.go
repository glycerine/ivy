package abc

// Constant-register reduction from src/aig/aig/aigTsim.c.

type AigConstantReduceOptions struct {
	Multivalued bool
	MaxPasses   int // zero: initial register count plus one
	Ternary     AigTernaryStateOptions
	Remap       AigRewriteOptions
	Cleanup     AigSequentialCleanupOptions
}

type AigConstantReduction struct {
	Graph            *AigManager
	Passes           int
	RemovedRegisters int
	Registers        []int
}

// ReduceTernaryConstants implements Aig_ManConstReduce's ordinary ternary
// branch. It returns an owned graph and never consumes or mutates its receiver.
func (graph *AigManager) ReduceTernaryConstants(options AigConstantReduceOptions) (AigConstantReduction, error) {
	result := AigConstantReduction{}
	if graph == nil || options.MaxPasses < 0 {
		return result, ErrDimension
	}
	if options.Multivalued || graph.equivs != nil {
		return result, ErrUnsupported
	}
	if err := graph.Check(); err != nil {
		return result, err
	}
	maximumPasses := options.MaxPasses
	if maximumPasses == 0 {
		maximumPasses = graph.registers + 1
	}
	current := graph.Clone()
	initialRegisters := graph.registers
	projection := make([]int, initialRegisters)
	for index := range projection {
		projection[index] = index
	}
	for current.registers > 0 {
		analysis, err := current.SequentialTernaryConstants(options.Ternary)
		if err != nil {
			return AigConstantReduction{}, err
		}
		if analysis.Mapping == nil {
			break
		}
		if result.Passes == maximumPasses {
			return AigConstantReduction{}, ErrLimit
		}
		remapped, err := current.RemapInputs(analysis.Mapping, options.Remap)
		if err != nil {
			return AigConstantReduction{}, err
		}
		cleaned, err := remapped.SequentialCleanup(options.Cleanup)
		if err != nil {
			return AigConstantReduction{}, err
		}
		if cleaned.Graph.registers >= current.registers {
			return AigConstantReduction{}, ErrInvariant
		}
		nextProjection := make([]int, len(cleaned.Registers))
		for index, retained := range cleaned.Registers {
			if retained < 0 || retained >= len(projection) {
				return AigConstantReduction{}, ErrInvariant
			}
			nextProjection[index] = projection[retained]
		}
		projection = nextProjection
		current = cleaned.Graph
		result.Passes++
	}
	result.Graph = current
	result.RemovedRegisters = initialRegisters - current.registers
	result.Registers = projection
	return result, nil
}
