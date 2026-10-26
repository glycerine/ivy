package abc

// Concrete command dispatch from src/misc/espresso/main.c.

type EspressoDriverExecutionOptions struct {
	MaxCubes, MaxOperations, MaxRows, MaxDepth int
	MaxIterations, MaxBytes                    int
	Stop                                       func() bool
}

type EspressoDriverExecution struct {
	PLA               *EspressoPLA
	Text              string
	Decision, Decided bool
	Verification      EspressoCoverVerification
	Consistency       EspressoConsistency
	EquivalentOutputs []EspressoOutputEquivalence
	FSM               EspressoFSMDisassembly
	PositivePhase     []bool
	Pairing           EspressoPairing
}

func espressoDriverWithOn(source *EspressoPLA, on *EspressoSetFamily) (*EspressoPLA, error) {
	return espressoDriverWithFamilies(source, on, source.DontCare, source.Off)
}

func espressoDriverWithFamilies(source *EspressoPLA, on, dontCare, off *EspressoSetFamily) (*EspressoPLA, error) {
	result, err := NewEspressoPLA(source.Layout, on, dontCare, off, source.Labels)
	if err != nil {
		return nil, err
	}
	result.Filename, result.Type = source.Filename, source.Type
	return result, nil
}

func espressoDriverSharpOptions(options EspressoDriverExecutionOptions) EspressoSharpOptions {
	return EspressoSharpOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop}
}

func espressoDriverContainOptions(options EspressoDriverExecutionOptions) EspressoContainOptions {
	return EspressoContainOptions{MaxSets: options.MaxCubes, MaxChecks: options.MaxOperations, Stop: options.Stop}
}

func espressoDriverVerifyOptions(options EspressoDriverExecutionOptions) EspressoVerifyOptions {
	return EspressoVerifyOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop}
}

func espressoDriverManipulationOptions(options EspressoDriverExecutionOptions) EspressoCoverManipulationOptions {
	return EspressoCoverManipulationOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop}
}

func espressoDriverMinimizeOptions(runtime EspressoRuntimeOptions, options EspressoDriverExecutionOptions) (EspressoMinimizeOptions, error) {
	result, err := runtime.MinimizeOptions()
	if err != nil {
		return result, err
	}
	result.MaxCubes, result.MaxOperations = options.MaxCubes, options.MaxOperations
	result.MaxRows, result.MaxDepth = options.MaxRows, options.MaxDepth
	result.MaxIterations, result.Stop = options.MaxIterations, options.Stop
	return result, nil
}

func espressoDriverExactOptions(options EspressoDriverExecutionOptions) EspressoExactOptions {
	return EspressoExactOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxSearchNodes: options.MaxRows, MaxDepth: options.MaxDepth, Stop: options.Stop}
}

func espressoDriverPhaseOptions(config EspressoDriverConfig, options EspressoDriverExecutionOptions) (EspressoPhaseOptions, error) {
	heuristic, err := espressoDriverMinimizeOptions(config.Runtime, options)
	if err != nil {
		return EspressoPhaseOptions{}, err
	}
	return EspressoPhaseOptions{MaxPhases: options.MaxRows, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop, Heuristic: heuristic, Exact: espressoDriverExactOptions(options)}, nil
}

func espressoDriverPairOptions(config EspressoDriverConfig, options EspressoDriverExecutionOptions) (EspressoPairOptions, error) {
	heuristic, err := espressoDriverMinimizeOptions(config.Runtime, options)
	if err != nil {
		return EspressoPairOptions{}, err
	}
	phase, err := espressoDriverPhaseOptions(config, options)
	if err != nil {
		return EspressoPairOptions{}, err
	}
	return EspressoPairOptions{MaxPairings: options.MaxIterations, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, MaxDepth: options.MaxDepth, MaxSearchNodes: options.MaxRows, Stop: options.Stop, Heuristic: heuristic, Exact: espressoDriverExactOptions(options), Phase: phase}, nil
}

func espressoDriverValidateInputs(config EspressoDriverConfig, first, second *EspressoPLA) error {
	known, ok := LookupEspressoCommand(config.Command.Name)
	if !ok || known != config.Command || first == nil {
		return ErrDimension
	}
	if err := first.check(); err != nil {
		return err
	}
	if config.Command.InputPLAs == 2 {
		if second == nil {
			return ErrDimension
		}
		if err := second.check(); err != nil {
			return err
		}
		if first.Layout.bitCount != second.Layout.bitCount || first.Layout.binaryVariables != second.Layout.binaryVariables || len(first.Layout.partSizes) != len(second.Layout.partSizes) {
			return ErrDimension
		}
		for index, size := range first.Layout.partSizes {
			if second.Layout.partSizes[index] != size {
				return ErrDimension
			}
		}
	}
	return nil
}

// ExecuteEspressoCommand runs one already-parsed command without file or
// process side effects. Returned PLAs and diagnostic structures are owned.
func ExecuteEspressoCommand(config EspressoDriverConfig, first, second *EspressoPLA, options EspressoDriverExecutionOptions) (EspressoDriverExecution, error) {
	if options.MaxCubes < 0 || options.MaxOperations < 0 || options.MaxRows < 0 || options.MaxDepth < 0 || options.MaxIterations < 0 || options.MaxBytes < 0 {
		return EspressoDriverExecution{}, ErrDimension
	}
	if options.Stop != nil && options.Stop() {
		return EspressoDriverExecution{}, ErrCancelled
	}
	if err := espressoDriverValidateInputs(config, first, second); err != nil {
		return EspressoDriverExecution{}, err
	}
	layout := first.Layout
	sharp := espressoDriverSharpOptions(options)
	contain := espressoDriverContainOptions(options)
	verify := espressoDriverVerifyOptions(options)
	manipulate := espressoDriverManipulationOptions(options)
	result := EspressoDriverExecution{}
	var on *EspressoSetFamily
	var err error

	switch config.Command.Kind {
	case EspressoCommandEcho:
		result.PLA, err = first.Clone()
	case EspressoCommandMinimize:
		minimize, projectionErr := espressoDriverMinimizeOptions(config.Runtime, options)
		if projectionErr != nil {
			return result, projectionErr
		}
		minimized, minimizeErr := layout.MinimizeEspressoCover(first.On, first.DontCare, first.Off, minimize)
		if minimizeErr != nil {
			return result, minimizeErr
		}
		result.PLA, err = espressoDriverWithOn(first, minimized.Cover)
	case EspressoCommandExact, EspressoCommandQuineMcCluskey:
		exact := espressoDriverExactOptions(options)
		minimized, exactErr := layout.MinimizeExact(first.On, first.DontCare, first.Off, config.Command.Kind == EspressoCommandExact, exact)
		if exactErr != nil {
			return result, exactErr
		}
		result.PLA, err = espressoDriverWithOn(first, minimized.Cover)
	case EspressoCommandSimplify:
		on, err = layout.SimplifyCover(first.On, sharp)
	case EspressoCommandSingleOutput, EspressoCommandSingleOutputBoth:
		strategy := EspressoOutputHeuristic
		if config.Strategy == 1 {
			strategy = EspressoOutputExact
		}
		heuristic, optionErr := espressoDriverMinimizeOptions(config.Runtime, options)
		if optionErr != nil {
			return result, optionErr
		}
		minimized, minimizeErr := first.MinimizeOutputs(EspressoOutputMinimizeOptions{Strategy: strategy, ChoosePolarity: config.Command.Kind == EspressoCommandSingleOutputBoth, Manipulation: manipulate, Heuristic: heuristic, Exact: espressoDriverExactOptions(options)})
		if minimizeErr != nil {
			return result, minimizeErr
		}
		result.PLA, result.PositivePhase = minimized.PLA, minimized.PositivePhase
	case EspressoCommandOutputPhase, EspressoCommandAllOutputPhases:
		phase, optionErr := espressoDriverPhaseOptions(config, options)
		if optionErr != nil {
			return result, optionErr
		}
		if config.Command.Kind == EspressoCommandOutputPhase {
			phase.Strategy = EspressoPhaseStrategy((config.Strategy >> 2) & 1)
			phase.Repeated = config.Strategy&2 != 0
			phase.Heuristic.SkipMakeSparse = config.Strategy&1 != 0
			assigned, assignErr := first.AssignOutputPhase(phase)
			if assignErr != nil {
				return result, assignErr
			}
			result.PLA, result.PositivePhase = assigned.PLA, assigned.PositivePhase
		} else {
			if config.Strategy == 1 {
				phase.Strategy = EspressoPhaseExact
			}
			start, end := config.First, config.Last
			output := layout.OutputVariable()
			if output < 0 {
				return result, ErrDimension
			}
			if start < 0 || start >= layout.partSizes[output] {
				start = 0
			}
			if end < 0 || end >= layout.partSizes[output] {
				end = layout.partSizes[output] - 1
			}
			assigned, assignErr := first.EnumerateOutputPhases(start, end, phase)
			if assignErr != nil {
				return result, assignErr
			}
			result.PLA, result.PositivePhase = assigned.PLA, assigned.PositivePhase
		}
	case EspressoCommandPair, EspressoCommandAllPairs:
		pair, optionErr := espressoDriverPairOptions(config, options)
		if optionErr != nil {
			return result, optionErr
		}
		if config.Command.Kind == EspressoCommandPair {
			strategy := config.Strategy
			if strategy < 0 || strategy > int(EspressoPairExact) {
				strategy = 0
			}
			paired, pairErr := first.FindOptimalPairing(EspressoPairCostStrategy(strategy), pair)
			if pairErr != nil {
				return result, pairErr
			}
			result.PLA, result.Pairing, result.PositivePhase = paired.PLA, paired.Pairing, paired.PositivePhase
		} else {
			strategy := config.Strategy
			if strategy < 0 || strategy > int(EspressoPairOptimizePhase) {
				strategy = 0
			}
			paired, pairErr := first.OptimizePairing(EspressoPairOptimizationStrategy(strategy), pair)
			if pairErr != nil {
				return result, pairErr
			}
			result.PLA, result.Pairing, result.PositivePhase = paired.PLA, paired.Pairing, paired.PositivePhase
		}
	case EspressoCommandContain:
		on, err = first.On.Contain(contain)
	case EspressoCommandIntersect:
		on, err = layout.IntersectCovers(first.On, second.On, sharp)
	case EspressoCommandUnion:
		on, err = ContainedUnionEspressoSetFamilies(first.On, second.On, contain)
	case EspressoCommandDisjoint:
		on, err = layout.MakeDisjoint(first.On, sharp)
	case EspressoCommandDisjointSharp:
		on, err = layout.DisjointSharpCovers(first.On, second.On, sharp)
	case EspressoCommandSharp:
		on, err = layout.SharpCovers(first.On, second.On, sharp)
	case EspressoCommandLexicalSort:
		on, err = first.On.SortLexical(manipulate)
	case EspressoCommandMinterms:
		start, end := config.First, config.Last
		if start < 0 || start >= layout.VariableCount() {
			start = 0
		}
		if end < 0 || end >= layout.VariableCount() {
			end = layout.VariableCount() - 1
		}
		on, err = layout.UnravelRange(first.On, start, end, manipulate)
	case EspressoCommandDistanceOneMerge, EspressoCommandDistanceOneMergeInputs:
		on, _ = first.On.Clone()
		start, end := config.First, config.Last
		if config.Command.Kind == EspressoCommandDistanceOneMergeInputs {
			start, end = 0, layout.binaryVariables-1
		} else {
			if start < 0 || start >= layout.VariableCount() {
				start = 0
			}
			if end < 0 || end >= layout.VariableCount() {
				end = layout.VariableCount() - 1
			}
		}
		for variable := start; variable <= end && err == nil; variable++ {
			on, err = on.DistanceOneMerge(layout, variable, contain)
		}
	case EspressoCommandPrimes:
		list, listErr := NewEspressoCubeList(layout, first.On, first.DontCare)
		if listErr != nil {
			return result, listErr
		}
		on, err = layout.PrimeConsensus(list, EspressoPrimeOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxDepth: options.MaxDepth, Stop: options.Stop})
	case EspressoCommandEssentials:
		marked, _ := first.On.Clone()
		for _, cube := range marked.sets {
			cube.Flags |= EspressoSetRelEssential
			cube.Flags &^= EspressoSetNonessential
		}
		extracted, extractErr := layout.ExtractEssentials(marked, first.DontCare, EspressoEssentialOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
		if extractErr != nil {
			return result, extractErr
		}
		result.PLA, err = espressoDriverWithFamilies(first, extracted.Remaining, extracted.DontCare, first.Off)
	case EspressoCommandExpand:
		on, err = layout.ExpandCover(first.On, first.Off, EspressoExpandOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxDepth: options.MaxDepth, Stop: options.Stop})
	case EspressoCommandIrredundant:
		irredundant, irredundantErr := layout.Irredundant(first.On, first.DontCare, EspressoIrredundantOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, Stop: options.Stop})
		if irredundantErr != nil {
			return result, irredundantErr
		}
		on = irredundant.Cover
	case EspressoCommandReduce:
		on, err = layout.ReduceCover(first.On, first.DontCare, EspressoReduceOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
	case EspressoCommandGasp, EspressoCommandSuperGasp:
		gasp := EspressoGaspOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, MaxDepth: options.MaxDepth, Stop: options.Stop}
		if config.Command.Kind == EspressoCommandGasp {
			on, err = layout.LastGasp(first.On, first.DontCare, first.Off, gasp)
		} else {
			on, err = layout.SuperGasp(first.On, first.DontCare, first.Off, gasp)
		}
	case EspressoCommandMakeSparse:
		on, err = layout.MakeSparse(first.On, first.DontCare, first.Off, EspressoMakeSparseOptions{ForceIrredundant: config.Runtime.ForceIrredundant, MaxIterations: options.MaxIterations, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, MaxDepth: options.MaxDepth, Stop: options.Stop})
	case EspressoCommandTest:
		blocked, joinErr := JoinEspressoSetFamilies(first.DontCare, first.Off)
		if joinErr != nil {
			return result, joinErr
		}
		testOn, testErr := layout.ComplementCover(blocked, sharp)
		if testErr != nil {
			return result, testErr
		}
		testOn, testErr = layout.ExpandCover(testOn, blocked, EspressoExpandOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxDepth: options.MaxDepth, Stop: options.Stop})
		if testErr != nil {
			return result, testErr
		}
		empty, _ := NewEspressoSetFamily(0, layout.bitCount)
		testIrred, testErr := layout.Irredundant(testOn, empty, EspressoIrredundantOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, Stop: options.Stop})
		if testErr != nil {
			return result, testErr
		}
		dcBlocked, joinErr := JoinEspressoSetFamilies(testIrred.Cover, first.Off)
		if joinErr != nil {
			return result, joinErr
		}
		testDC, testErr := layout.ExpandCover(first.DontCare, dcBlocked, EspressoExpandOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxDepth: options.MaxDepth, Stop: options.Stop})
		if testErr != nil {
			return result, testErr
		}
		testDCIrred, testErr := layout.Irredundant(testDC, empty, EspressoIrredundantOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, Stop: options.Stop})
		if testErr != nil {
			return result, testErr
		}
		result.PLA, err = espressoDriverWithFamilies(first, testIrred.Cover, testDCIrred.Cover, first.Off)
	case EspressoCommandSeparate:
		blocked, joinErr := JoinEspressoSetFamilies(first.DontCare, first.Off)
		if joinErr != nil {
			return result, joinErr
		}
		on, err = layout.ComplementCover(blocked, sharp)
	case EspressoCommandXOR:
		left, leftErr := layout.IntersectCovers(first.On, second.Off, sharp)
		if leftErr != nil {
			return result, leftErr
		}
		right, rightErr := layout.IntersectCovers(second.On, first.Off, sharp)
		if rightErr != nil {
			return result, rightErr
		}
		on, err = ContainedUnionEspressoSetFamilies(left, right, contain)
	case EspressoCommandCheck:
		result.Consistency, err = first.CheckConsistency(verify)
		result.Decision, result.Decided = result.Consistency.Consistent, true
		if err == nil {
			result.PLA, err = first.Clone()
		}
	case EspressoCommandVerify:
		result.Verification, err = layout.VerifyCovers(second.On, first.On, first.DontCare, verify)
		result.Decision, result.Decided = result.Verification.Equivalent, true
	case EspressoCommandPLAVerify:
		result.Verification, err = first.Verify(second, verify)
		result.Decision, result.Decided = result.Verification.Equivalent, true
	case EspressoCommandStats:
		result.Text, err = (EspressoPLAParseResult{PLA: first}).Summary(EspressoCoverCostOptions{MaxCubes: options.MaxCubes, MaxChecks: options.MaxOperations, Stop: options.Stop})
	case EspressoCommandMap:
		result.Text, err = layout.FormatKarnaughMap(first.On, EspressoMapOptions{MaxCubes: options.MaxCubes, MaxMinterms: options.MaxRows, MaxVisits: options.MaxOperations, Stop: options.Stop})
	case EspressoCommandMapDontCare:
		result.PLA, err = first.MapDontCareVariable(EspressoSymbolicOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxDepth: options.MaxDepth, Stop: options.Stop})
	case EspressoCommandEquivalentOutputs:
		result.EquivalentOutputs, err = first.FindEquivalentOutputs(verify)
	case EspressoCommandFSM:
		result.FSM, err = first.DisassembleFSM(EspressoFSMDisassembleOptions{Verbose: config.Runtime.Summary, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, MaxDepth: options.MaxDepth, MaxIterations: options.MaxIterations, MaxBytes: options.MaxBytes, Stop: options.Stop})
		result.PLA = result.FSM.PLA
	case EspressoCommandTautology:
		result.Decision, err = layout.CoverIsTautology(first.On, EspressoIrredundantOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, Stop: options.Stop})
		result.Decided = err == nil
		if err == nil {
			result.PLA, err = first.Clone()
		}
	default:
		return EspressoDriverExecution{}, ErrUnsupported
	}
	if err != nil {
		return EspressoDriverExecution{}, err
	}
	if on != nil {
		result.PLA, err = espressoDriverWithOn(first, on)
	}
	return result, err
}
