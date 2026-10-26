package abc

// Property-output prime decomposition from src/aig/saig/saigOutDec.c.

import "slices"

type SAIGPropertyPrimeOptions struct {
	MaxObjects, MaxVisits, MaxCandidates  int
	MaxPairs, MaxQueries, MaxCubes        int
	MaxVariables, MaxClauses, MaxLiterals int
	MaxWork                               int
	SAT                                   SatLimits
	Stop                                  func() bool
}

type SAIGPrimeCube struct {
	Literals []AigLit
}

type SAIGPropertyPrimes struct {
	Cubes                   []SAIGPrimeCube
	Cone                    []AigID
	Single, Pair            int
	Queries, Unknown, Pairs int
}

type saigPrimeBudget struct {
	options SAIGPropertyPrimeOptions
	work    int
}

func normalizeSAIGPropertyPrimeOptions(options SAIGPropertyPrimeOptions) (SAIGPropertyPrimeOptions, error) {
	if options.MaxObjects < 0 || options.MaxVisits < 0 || options.MaxCandidates < 0 || options.MaxPairs < 0 || options.MaxQueries < 0 || options.MaxCubes < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 || options.MaxWork < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 8 << 20
	}
	if options.MaxCandidates == 0 {
		options.MaxCandidates = 1 << 20
	}
	if options.MaxPairs == 0 {
		options.MaxPairs = 4 << 20
	}
	if options.MaxQueries == 0 {
		options.MaxQueries = 16 << 20
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 4 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 16 << 20
	}
	if options.MaxWork == 0 {
		options.MaxWork = 32 << 20
	}
	return options, nil
}

func (budget *saigPrimeBudget) step(count int) error {
	if count < 0 || budget.work > budget.options.MaxWork-count {
		return ErrLimit
	}
	budget.work += count
	if budget.options.Stop != nil && budget.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func saigPrimeSATLimits(options SAIGPropertyPrimeOptions) SatLimits {
	limits := options.SAT
	satStop := limits.Stop
	limits.Stop = func() bool {
		return options.Stop != nil && options.Stop() || satStop != nil && satStop()
	}
	return limits
}

func saigPrimeLiteral(id AigID, complement bool) AigLit {
	literal := AigLit(id) << 1
	return literal.NotIf(complement)
}

func saigPrimeAssumption(relation *AigCNF, graph *AigManager, id AigID, extraComplement bool) (SatLit, error) {
	literal, err := relation.ObjectLiteral(id)
	if err != nil {
		return InvalidSatLit, err
	}
	if graph.objects[id].Phase != extraComplement {
		literal = literal.Not()
	}
	return literal, nil
}

// SAIGFindPropertyPrimes preserves native DFS order, phase-relative literals,
// single-prime marking and the source's three tested pair polarities. Only
// UNSAT assumptions produce a prime; bounded unknown answers are reported and
// skipped.
func SAIGFindPropertyPrimes(graph *AigManager, literals int, options SAIGPropertyPrimeOptions) (SAIGPropertyPrimes, error) {
	options, err := normalizeSAIGPropertyPrimeOptions(options)
	if err != nil {
		return SAIGPropertyPrimes{}, err
	}
	if graph == nil || (literals != 1 && literals != 2) {
		return SAIGPropertyPrimes{}, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return SAIGPropertyPrimes{}, err
	}
	if graph.PrimaryOutputCount() < 1 || graph.ObjectCount() > options.MaxObjects {
		if graph.ObjectCount() > options.MaxObjects {
			return SAIGPropertyPrimes{}, ErrLimit
		}
		return SAIGPropertyPrimes{}, ErrDimension
	}
	budget := &saigPrimeBudget{options: options}
	propertyID := graph.cos[0]
	root := graph.objects[propertyID].Fanin0.ID()
	cone, err := graph.DFSNodes([]AigID{root}, AigTraversalOptions{MaxObjects: options.MaxObjects, MaxVisits: options.MaxVisits, Stop: options.Stop})
	if err != nil {
		return SAIGPropertyPrimes{}, err
	}
	if len(cone) > options.MaxCandidates {
		return SAIGPropertyPrimes{}, ErrLimit
	}
	relation, err := EncodeAigCNF(graph)
	if err != nil {
		return SAIGPropertyPrimes{}, err
	}
	if relation.VariableCount() > options.MaxVariables || len(relation.clauses) > options.MaxClauses {
		return SAIGPropertyPrimes{}, ErrLimit
	}
	literalCount := 0
	for _, clause := range relation.clauses {
		if literalCount > options.MaxLiterals-len(clause) {
			return SAIGPropertyPrimes{}, ErrLimit
		}
		literalCount += len(clause)
	}
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return SAIGPropertyPrimes{}, err
	}
	if ok, loadErr := relation.loadIntoWithStop(solver, 0, options.Stop); loadErr != nil || !ok {
		if loadErr != nil {
			return SAIGPropertyPrimes{}, loadErr
		}
		return SAIGPropertyPrimes{}, ErrInvariant
	}
	outputs := relation.OutputLiterals()
	if len(outputs) == 0 {
		return SAIGPropertyPrimes{}, ErrInvariant
	}
	propertyFalse := outputs[0].Not()
	result := SAIGPropertyPrimes{Cone: slices.Clone(cone)}
	marked := make([]bool, len(cone))
	solve := func(assumptions []SatLit) (SatStatus, error) {
		if result.Queries == options.MaxQueries {
			return SatUnknown, ErrLimit
		}
		if err := budget.step(len(assumptions) + 1); err != nil {
			return SatUnknown, err
		}
		result.Queries++
		answer, solveErr := solver.Solve(assumptions, saigPrimeSATLimits(options))
		if solveErr != nil {
			return SatUnknown, solveErr
		}
		if answer.Status == SatUnknown {
			result.Unknown++
		}
		return answer.Status, nil
	}
	appendCube := func(cube SAIGPrimeCube) error {
		if len(result.Cubes) == options.MaxCubes {
			return ErrLimit
		}
		cube.Literals = slices.Clone(cube.Literals)
		result.Cubes = append(result.Cubes, cube)
		return nil
	}
	for index, id := range cone {
		if id == root {
			continue
		}
		assumption, assumptionErr := saigPrimeAssumption(relation, graph, id, false)
		if assumptionErr != nil {
			return SAIGPropertyPrimes{}, assumptionErr
		}
		status, solveErr := solve([]SatLit{propertyFalse, assumption})
		if solveErr != nil {
			return SAIGPropertyPrimes{}, solveErr
		}
		if status == SatUnsatisfiable {
			if err = appendCube(SAIGPrimeCube{Literals: []AigLit{saigPrimeLiteral(id, graph.objects[id].Phase)}}); err != nil {
				return SAIGPropertyPrimes{}, err
			}
			marked[index] = true
			result.Single++
		}
	}
	if literals == 2 {
		for first, firstID := range cone {
			if firstID == root {
				continue
			}
			for second := first + 1; second < len(cone); second++ {
				secondID := cone[second]
				if secondID == root || marked[first] || marked[second] {
					continue
				}
				if result.Pairs == options.MaxPairs {
					return SAIGPropertyPrimes{}, ErrLimit
				}
				result.Pairs++
				for mode := 0; mode < 3; mode++ {
					firstComplement := mode&1 != 0
					secondComplement := mode&2 != 0
					firstAssumption, firstErr := saigPrimeAssumption(relation, graph, firstID, firstComplement)
					if firstErr != nil {
						return SAIGPropertyPrimes{}, firstErr
					}
					secondAssumption, secondErr := saigPrimeAssumption(relation, graph, secondID, secondComplement)
					if secondErr != nil {
						return SAIGPropertyPrimes{}, secondErr
					}
					status, solveErr := solve([]SatLit{propertyFalse, firstAssumption, secondAssumption})
					if solveErr != nil {
						return SAIGPropertyPrimes{}, solveErr
					}
					if status != SatUnsatisfiable {
						continue
					}
					cube := SAIGPrimeCube{Literals: []AigLit{
						saigPrimeLiteral(firstID, graph.objects[firstID].Phase != firstComplement),
						saigPrimeLiteral(secondID, graph.objects[secondID].Phase != secondComplement),
					}}
					if err = appendCube(cube); err != nil {
						return SAIGPropertyPrimes{}, err
					}
					result.Pair++
					break
				}
			}
		}
	}
	return result, nil
}

func saigCopySignal(mapping []AigLit, literal AigLit) (AigLit, error) {
	if !literal.Valid() || uint64(literal.ID()) >= uint64(len(mapping)) || !mapping[literal.ID()].Valid() {
		return InvalidAigLit, ErrInvariant
	}
	return mapping[literal.ID()].NotIf(literal.Complemented()), nil
}

// SAIGDecomposePropertyOutput duplicates the source graph, retains all original
// primary outputs, inserts each prime cube as another primary output, then
// appends the original latch inputs and restores the register partition.
func SAIGDecomposePropertyOutput(graph *AigManager, literals int, options SAIGPropertyPrimeOptions) (*AigManager, SAIGPropertyPrimes, error) {
	options, err := normalizeSAIGPropertyPrimeOptions(options)
	if err != nil {
		return nil, SAIGPropertyPrimes{}, err
	}
	primes, err := SAIGFindPropertyPrimes(graph, literals, options)
	if err != nil {
		return nil, SAIGPropertyPrimes{}, err
	}
	result := NewAigManager()
	result.Name, result.Spec = graph.Name, graph.Spec
	mapping := make([]AigLit, graph.ObjectCount())
	for index := range mapping {
		mapping[index] = InvalidAigLit
	}
	mapping[0] = AigConst1
	check := func() error {
		if result.ObjectCount() > options.MaxObjects {
			return ErrLimit
		}
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		return nil
	}
	for _, id := range graph.cis {
		mapping[id], err = result.CreateCI()
		if err != nil {
			return nil, SAIGPropertyPrimes{}, err
		}
		if err = check(); err != nil {
			return nil, SAIGPropertyPrimes{}, err
		}
	}
	for id, object := range graph.objects {
		switch object.Type {
		case AigNone, AigConstant1, AigCI, AigCO:
			continue
		case AigBuffer:
			mapping[id], err = saigCopySignal(mapping, object.Fanin0)
		case AigAND:
			var first, second AigLit
			first, err = saigCopySignal(mapping, object.Fanin0)
			if err == nil {
				second, err = saigCopySignal(mapping, object.Fanin1)
			}
			if err == nil {
				mapping[id], err = result.And(first, second)
			}
		default:
			err = ErrUnsupported
		}
		if err != nil {
			return nil, SAIGPropertyPrimes{}, err
		}
		if err = check(); err != nil {
			return nil, SAIGPropertyPrimes{}, err
		}
	}
	primaryOutputs := graph.PrimaryOutputCount()
	appendOutput := func(driver AigLit) error {
		if _, outputErr := result.CreateCO(driver); outputErr != nil {
			return outputErr
		}
		return check()
	}
	for _, outputID := range graph.cos[:primaryOutputs] {
		driver, driverErr := saigCopySignal(mapping, graph.objects[outputID].Fanin0)
		if driverErr != nil {
			return nil, SAIGPropertyPrimes{}, driverErr
		}
		if err = appendOutput(driver); err != nil {
			return nil, SAIGPropertyPrimes{}, err
		}
	}
	for _, cube := range primes.Cubes {
		driver := AigConst1
		for _, literal := range cube.Literals {
			mapped, mapErr := saigCopySignal(mapping, literal)
			if mapErr != nil {
				return nil, SAIGPropertyPrimes{}, mapErr
			}
			driver, err = result.And(driver, mapped)
			if err != nil {
				return nil, SAIGPropertyPrimes{}, err
			}
		}
		if err = appendOutput(driver); err != nil {
			return nil, SAIGPropertyPrimes{}, err
		}
	}
	for _, outputID := range graph.cos[primaryOutputs:] {
		driver, driverErr := saigCopySignal(mapping, graph.objects[outputID].Fanin0)
		if driverErr != nil {
			return nil, SAIGPropertyPrimes{}, driverErr
		}
		if err = appendOutput(driver); err != nil {
			return nil, SAIGPropertyPrimes{}, err
		}
	}
	if err = result.SetRegisterCount(graph.RegisterCount()); err != nil {
		return nil, SAIGPropertyPrimes{}, err
	}
	if err = result.SetConstraintCount(graph.ConstraintCount()); err != nil {
		return nil, SAIGPropertyPrimes{}, err
	}
	result.Cleanup()
	if err = result.Check(); err != nil {
		return nil, SAIGPropertyPrimes{}, err
	}
	return result, primes, nil
}
