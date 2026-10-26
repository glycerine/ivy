package abc

// Derived from src/proof/dch/{dch.h,dchInt.h,dchAig.c,dchChoice.c,
// dchClass.c,dchCnf.c,dchCore.c,dchMan.c,dchSat.c,dchSim.c,
// dchSimSat.c,dchSweep.c,module.make}.

import (
	"encoding/binary"
	"hash/fnv"
	"slices"
)

// DchOptions controls deterministic simulation, CNF/SAT equivalence checking,
// sweeping and choice construction. Zero numeric values select bounded
// defaults. Stop is polled throughout and must not mutate the graph.
type DchOptions struct {
	SimulationWords      int
	ConflictLimit        uint64
	MaxObjects           int
	MaxSimulationWords   int
	MaxCandidatePairs    int
	MaxSATVariables      int
	SkipRedundantSupport bool
	Stop                 func() bool
}

// DefaultDchOptions returns the operational counterparts of DCH's native
// defaults: eight simulation words and a 1000-conflict per-comparison budget.
// The larger Go limits bound owned storage rather than relying on allocation
// failure. Passing the zero value to an operation selects the same defaults.
func DefaultDchOptions() DchOptions {
	return DchOptions{
		SimulationWords:    8,
		ConflictLimit:      1000,
		MaxObjects:         1 << 20,
		MaxSimulationWords: 1024,
		MaxCandidatePairs:  8 << 20,
		MaxSATVariables:    1 << 20,
	}
}

type DchMember struct {
	ID       AigID
	Inverted bool
}

// DchClass is ordered representative first and then by object ID. Inverted
// means that member's function is the complement of the representative.
type DchClass struct {
	Representative AigID
	Members        []DchMember
}

type DchStats struct {
	Candidates, CandidatePairs int
	SATCalls                   int
	Proved, Disproved, Unknown int
	ConstantMembers            int
	ChoiceClasses, Choices     int
}

// DchResult owns every returned slice and graph. Replacements is indexed by
// source object ID. Graph preserves the source structure and attaches proven
// same-polarity choice chains; Swept rebuilds outputs through representatives.
type DchResult struct {
	Graph, Swept *AigManager
	Classes      []DchClass
	Replacements []AigLit
	Stats        DchStats
}

func normalizeDchOptions(o DchOptions) (DchOptions, error) {
	if o.SimulationWords < 0 || o.MaxObjects < 0 || o.MaxSimulationWords < 0 || o.MaxCandidatePairs < 0 || o.MaxSATVariables < 0 {
		return o, ErrDimension
	}
	if o.SimulationWords == 0 {
		o.SimulationWords = 8
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxSimulationWords == 0 {
		o.MaxSimulationWords = 1024
	}
	if o.MaxCandidatePairs == 0 {
		o.MaxCandidatePairs = 8 << 20
	}
	if o.MaxSATVariables == 0 {
		o.MaxSATVariables = 1 << 20
	}
	if o.SimulationWords > o.MaxSimulationWords {
		return o, ErrLimit
	}
	if o.ConflictLimit == 0 {
		o.ConflictLimit = 1000
	}
	return o, nil
}

func dchStop(o DchOptions) error {
	if o.Stop != nil && o.Stop() {
		return ErrCancelled
	}
	return nil
}

func dchMix(state *uint64) uint64 {
	*state += 0x9e3779b97f4a7c15
	z := *state
	z = (z ^ z>>30) * 0xbf58476d1ce4e5b9
	z = (z ^ z>>27) * 0x94d049bb133111eb
	return z ^ z>>31
}

func dchSimulate(g *AigManager, words int, o DchOptions) ([][]uint64, error) {
	if len(g.objects) > o.MaxObjects {
		return nil, ErrLimit
	}
	if uint64(len(g.objects))*uint64(words) > uint64(^uint(0)>>1) {
		return nil, ErrLimit
	}
	sim := make([][]uint64, len(g.objects))
	for id := range sim {
		sim[id] = make([]uint64, words)
	}
	for word := range sim[0] {
		sim[0][word] = ^uint64(0)
	}
	state := uint64(1)
	for _, id := range g.cis {
		for word := range sim[id] {
			sim[id][word] = dchMix(&state)
		}
		if words > 0 {
			sim[id][0] &^= 1
		}
	}
	read := func(l AigLit, word int) uint64 {
		value := sim[l.ID()][word]
		if l.Complemented() {
			return ^value
		}
		return value
	}
	for id, object := range g.objects {
		if id&255 == 0 {
			if err := dchStop(o); err != nil {
				return nil, err
			}
		}
		switch object.Type {
		case AigAND:
			for word := 0; word < words; word++ {
				sim[id][word] = read(object.Fanin0, word) & read(object.Fanin1, word)
			}
		case AigBuffer, AigCO:
			for word := 0; word < words; word++ {
				sim[id][word] = read(object.Fanin0, word)
			}
		}
	}
	return sim, nil
}

func dchCanonicalSimulation(words []uint64) ([]uint64, bool) {
	if len(words) == 0 || words[0]>>63 == 0 {
		return words, false
	}
	result := make([]uint64, len(words))
	for i, word := range words {
		result[i] = ^word
	}
	return result, true
}

func dchSimulationKey(words []uint64) uint64 {
	h := fnv.New64a()
	var data [8]byte
	for _, word := range words {
		binary.LittleEndian.PutUint64(data[:], word)
		_, _ = h.Write(data[:])
	}
	return h.Sum64()
}

type dchCandidate struct {
	id       AigID
	inverted bool
	words    []uint64
}

func dchCandidates(g *AigManager, sim [][]uint64) []dchCandidate {
	result := make([]dchCandidate, 0, g.AndCount()+1)
	canonical, inverted := dchCanonicalSimulation(sim[0])
	result = append(result, dchCandidate{0, inverted, canonical})
	for id, object := range g.objects {
		if object.Type != AigAND {
			continue
		}
		canonical, inverted = dchCanonicalSimulation(sim[id])
		result = append(result, dchCandidate{AigID(id), inverted, canonical})
	}
	return result
}

func dchSolveMismatch(s *SatSolver, first, second SatLit, inverted bool, o DchOptions) (SatStatus, int, error) {
	patterns := [2][2]bool{{true, inverted}, {false, !inverted}}
	for call, pattern := range patterns {
		if err := dchStop(o); err != nil {
			return SatUnknown, call, err
		}
		firstAssumption, secondAssumption := first, second
		if !pattern[0] {
			firstAssumption = first.Not()
		}
		if !pattern[1] {
			secondAssumption = second.Not()
		}
		assumptions := []SatLit{firstAssumption, secondAssumption}
		result, err := s.Solve(assumptions, SatLimits{Conflicts: o.ConflictLimit, Stop: o.Stop})
		if err != nil {
			return SatUnknown, call + 1, err
		}
		if result.Status != SatUnsatisfiable {
			return result.Status, call + 1, nil
		}
	}
	return SatUnsatisfiable, 2, nil
}

type dchExactGroup struct {
	representative dchCandidate
	members        []DchMember
}

func dchFindClasses(g *AigManager, o DchOptions) ([]DchClass, []AigLit, DchStats, error) {
	var stats DchStats
	sim, err := dchSimulate(g, o.SimulationWords, o)
	if err != nil {
		return nil, nil, stats, err
	}
	candidates := dchCandidates(g, sim)
	stats.Candidates = len(candidates)
	cnf, err := EncodeAigCNF(g)
	if err != nil {
		return nil, nil, stats, err
	}
	if cnf.VariableCount() > o.MaxSATVariables {
		return nil, nil, stats, ErrLimit
	}
	solver, err := NewSatSolverWithLimit(o.MaxSATVariables)
	if err != nil {
		return nil, nil, stats, err
	}
	if ok, loadErr := cnf.LoadInto(solver, 0); loadErr != nil || !ok {
		if loadErr != nil {
			return nil, nil, stats, loadErr
		}
		return nil, nil, stats, ErrInvariant
	}
	buckets := make(map[uint64][]dchCandidate)
	for _, candidate := range candidates {
		key := dchSimulationKey(candidate.words)
		buckets[key] = append(buckets[key], candidate)
	}
	replacements := make([]AigLit, len(g.objects))
	for id := range replacements {
		replacements[id] = AigLit(id) << 1
	}
	var classes []DchClass
	// Candidate IDs, rather than map enumeration, determine all output order.
	processed := make([]bool, len(g.objects))
	for _, seed := range candidates {
		if processed[seed.id] {
			continue
		}
		bucket := buckets[dchSimulationKey(seed.words)]
		var exact []dchExactGroup
		for _, candidate := range bucket {
			if processed[candidate.id] || !slices.Equal(seed.words, candidate.words) {
				continue
			}
			processed[candidate.id] = true
			placed := false
			for groupIndex := range exact {
				if stats.CandidatePairs == o.MaxCandidatePairs {
					return nil, nil, stats, ErrLimit
				}
				stats.CandidatePairs++
				representative := exact[groupIndex].representative
				first, firstErr := cnf.ObjectLiteral(representative.id)
				second, secondErr := cnf.ObjectLiteral(candidate.id)
				if firstErr != nil || secondErr != nil {
					return nil, nil, stats, ErrInvariant
				}
				inverted := representative.inverted != candidate.inverted
				status, calls, solveErr := dchSolveMismatch(solver, first, second, inverted, o)
				stats.SATCalls += calls
				if solveErr != nil {
					return nil, nil, stats, solveErr
				}
				switch status {
				case SatUnsatisfiable:
					stats.Proved++
					exact[groupIndex].members = append(exact[groupIndex].members, DchMember{candidate.id, inverted})
					placed = true
				case SatSatisfiable:
					stats.Disproved++
				case SatUnknown:
					stats.Unknown++
				}
				if placed {
					break
				}
			}
			if !placed {
				exact = append(exact, dchExactGroup{candidate, []DchMember{{ID: candidate.id}}})
			}
		}
		for _, group := range exact {
			if len(group.members) < 2 {
				continue
			}
			class := DchClass{Representative: group.members[0].ID, Members: slices.Clone(group.members)}
			for _, member := range class.Members[1:] {
				replacements[member.ID] = (AigLit(class.Representative) << 1).NotIf(member.Inverted)
				if class.Representative == 0 {
					stats.ConstantMembers++
				}
			}
			classes = append(classes, class)
		}
	}
	return classes, replacements, stats, nil
}

func dchSupport(g *AigManager, id AigID, memo [][]AigID) []AigID {
	if memo[id] != nil {
		return memo[id]
	}
	object := g.objects[id]
	switch object.Type {
	case AigConstant1:
		memo[id] = []AigID{}
	case AigCI:
		memo[id] = []AigID{id}
	case AigBuffer, AigCO:
		memo[id] = slices.Clone(dchSupport(g, object.Fanin0.ID(), memo))
	case AigAND:
		first := dchSupport(g, object.Fanin0.ID(), memo)
		second := dchSupport(g, object.Fanin1.ID(), memo)
		memo[id] = make([]AigID, 0, len(first)+len(second))
		i, j := 0, 0
		for i < len(first) || j < len(second) {
			if j == len(second) || i < len(first) && first[i] < second[j] {
				memo[id] = append(memo[id], first[i])
				i++
			} else if i == len(first) || second[j] < first[i] {
				memo[id] = append(memo[id], second[j])
				j++
			} else {
				memo[id] = append(memo[id], first[i])
				i++
				j++
			}
		}
	}
	return memo[id]
}

func dchChoiceChains(g *AigManager, classes []DchClass, o DchOptions) ([][]AigID, int) {
	memo := make([][]AigID, len(g.objects))
	var chains [][]AigID
	choices := 0
	for _, class := range classes {
		if class.Representative == 0 || g.objects[class.Representative].Type != AigAND {
			continue
		}
		chain := []AigID{class.Representative}
		representativeSupport := dchSupport(g, class.Representative, memo)
		for _, member := range class.Members[1:] {
			if member.Inverted || g.objects[member.ID].Type != AigAND {
				continue
			}
			if o.SkipRedundantSupport && !slices.Equal(representativeSupport, dchSupport(g, member.ID, memo)) {
				continue
			}
			chain = append(chain, member.ID)
		}
		if len(chain) > 1 {
			chains = append(chains, chain)
			choices += len(chain) - 1
		}
	}
	return chains, choices
}

func dchChoiceReaches(g *AigManager, start AigID, marked []bool, next []AigID, o DchOptions) (bool, error) {
	seen := make([]bool, len(g.objects))
	stack := []AigID{start}
	for len(stack) > 0 {
		if err := dchStop(o); err != nil {
			return false, err
		}
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if marked[id] {
			return true, nil
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		obj := g.objects[id]
		if obj.Type != AigAND {
			continue
		}
		stack = append(stack, obj.Fanin0.ID(), obj.Fanin1.ID())
		if next[id] != InvalidAigID {
			stack = append(stack, next[id])
		}
	}
	return false, nil
}

// dchDeriveChoiceGraph follows Dch_DeriveChoiceAigNode: fanins are rebuilt
// through proved representatives, structurally collapsed alternatives are
// skipped, and only dangling, acyclic, same-polarity alternatives are linked.
func dchDeriveChoiceGraph(source *AigManager, replacements []AigLit, o DchOptions) (*AigManager, int, int, error) {
	g := NewAigManager()
	g.Name, g.Spec = source.Name, source.Spec
	mapping := make([]AigLit, len(source.objects))
	for i := range mapping {
		mapping[i] = InvalidAigLit
	}
	mapping[0] = AigConst1
	for _, id := range source.cis {
		literal, err := g.CreateCI()
		if err != nil {
			return nil, 0, 0, err
		}
		mapping[id] = literal
	}
	choiceNext := make([]AigID, len(source.objects))
	for i := range choiceNext {
		choiceNext[i] = InvalidAigID
	}
	choiceTail := make(map[AigID]AigID)
	choiceMembers := make(map[AigID]bool)
	readRepresentative := func(literal AigLit) AigLit {
		relation := replacements[literal.ID()]
		base := mapping[literal.ID()]
		if relation.Valid() && relation.ID() < literal.ID() {
			base = mapping[relation.ID()].NotIf(relation.Complemented())
		}
		return base.NotIf(literal.Complemented())
	}
	classes, choices := 0, 0
	for id, object := range source.objects {
		if id&255 == 0 {
			if err := dchStop(o); err != nil {
				return nil, 0, 0, err
			}
		}
		if object.Type != AigAND {
			continue
		}
		relation := replacements[id]
		if relation.Valid() && relation.ID() < AigID(id) && source.objects[relation.ID()].Type != AigAND {
			mapping[id] = mapping[relation.ID()].NotIf(relation.Complemented())
			continue
		}
		created, err := g.And(readRepresentative(object.Fanin0), readRepresentative(object.Fanin1))
		if err != nil {
			return nil, 0, 0, err
		}
		mapping[id] = created
		if !relation.Valid() || relation.ID() >= AigID(id) || relation.Complemented() {
			continue
		}
		representative := mapping[relation.ID()].ID()
		alternative := created.ID()
		if representative == alternative || g.objects[representative].Type != AigAND || g.objects[alternative].Type != AigAND || representative >= alternative || g.objects[alternative].RefCount != 0 || choiceMembers[alternative] {
			continue
		}
		head := representative
		if choiceMembers[head] {
			if _, isHead := choiceTail[head]; !isHead {
				continue
			}
		}
		if uint64(head) >= uint64(len(choiceNext)) || uint64(alternative) >= uint64(len(choiceNext)) {
			continue
		}
		marked := make([]bool, len(g.objects))
		for member := head; member != InvalidAigID; member = choiceNext[member] {
			marked[member] = true
		}
		reaches, err := dchChoiceReaches(g, alternative, marked, choiceNext, o)
		if err != nil {
			return nil, 0, 0, err
		}
		if reaches {
			continue
		}
		if o.SkipRedundantSupport {
			memo := make([][]AigID, len(g.objects))
			if !slices.Equal(dchSupport(g, head, memo), dchSupport(g, alternative, memo)) {
				continue
			}
		}
		tail, exists := choiceTail[head]
		if !exists {
			tail = head
			choiceMembers[head] = true
			classes++
		}
		choiceNext[tail] = alternative
		choiceTail[head] = alternative
		choiceMembers[alternative] = true
		choices++
		if g.ObjectCount() > o.MaxObjects {
			return nil, 0, 0, ErrLimit
		}
	}
	for _, co := range source.cos {
		driver := source.objects[co].Fanin0
		if _, err := g.CreateCO(readRepresentative(driver)); err != nil {
			return nil, 0, 0, err
		}
	}
	if err := g.SetRegisterCount(source.registers); err != nil {
		return nil, 0, 0, err
	}
	if err := g.SetConstraintCount(source.constraints); err != nil {
		return nil, 0, 0, err
	}
	chains := make([][]AigID, 0, classes)
	for head := range choiceTail {
		chain := []AigID{head}
		for member := choiceNext[head]; member != InvalidAigID; member = choiceNext[member] {
			chain = append(chain, member)
		}
		chains = append(chains, chain)
	}
	slices.SortFunc(chains, func(a, b []AigID) int {
		if a[0] < b[0] {
			return -1
		}
		if a[0] > b[0] {
			return 1
		}
		return 0
	})
	if len(chains) > 0 {
		if err := g.SetChoiceChains(chains, AigTraversalOptions{MaxObjects: o.MaxObjects, Stop: o.Stop}); err != nil {
			return nil, 0, 0, err
		}
	}
	if err := g.Check(); err != nil {
		return nil, 0, 0, err
	}
	return g, classes, choices, nil
}

func dchSweep(g *AigManager, replacements []AigLit, o DchOptions) (*AigManager, error) {
	result := NewAigManager()
	result.Name, result.Spec = g.Name, g.Spec
	mapping := make([]AigLit, len(g.objects))
	for id := range mapping {
		mapping[id] = InvalidAigLit
	}
	mapping[0] = AigConst1
	for _, id := range g.cis {
		literal, err := result.CreateCI()
		if err != nil {
			return nil, err
		}
		mapping[id] = literal
	}
	read := func(l AigLit) AigLit { return mapping[l.ID()].NotIf(l.Complemented()) }
	for id, object := range g.objects {
		if id&255 == 0 {
			if err := dchStop(o); err != nil {
				return nil, err
			}
		}
		switch object.Type {
		case AigAND:
			replacement := replacements[id]
			if replacement.ID() < AigID(id) {
				mapping[id] = mapping[replacement.ID()].NotIf(replacement.Complemented())
			} else {
				var err error
				mapping[id], err = result.And(read(object.Fanin0), read(object.Fanin1))
				if err != nil {
					return nil, err
				}
			}
		case AigBuffer:
			mapping[id] = read(object.Fanin0)
		case AigCO:
			if _, err := result.CreateCO(read(object.Fanin0)); err != nil {
				return nil, err
			}
		}
		if result.ObjectCount() > o.MaxObjects {
			return nil, ErrLimit
		}
	}
	if err := result.SetRegisterCount(g.registers); err != nil {
		return nil, err
	}
	if err := result.SetConstraintCount(g.constraints); err != nil {
		return nil, err
	}
	if err := result.Check(); err != nil {
		return nil, err
	}
	return result, nil
}

// ComputeDchEquivalences performs deterministic simulation-class refinement
// followed by exact SAT checking. The source is never mutated.
func ComputeDchEquivalences(source *AigManager, options DchOptions) (DchResult, error) {
	var result DchResult
	o, err := normalizeDchOptions(options)
	if err != nil {
		return result, err
	}
	if source == nil {
		return result, ErrDimension
	}
	if err = dchStop(o); err != nil {
		return result, err
	}
	if len(source.objects) > o.MaxObjects {
		return result, ErrLimit
	}
	if err = source.Check(); err != nil {
		return result, err
	}
	classes, replacements, stats, err := dchFindClasses(source, o)
	if err != nil {
		return result, err
	}
	swept, err := dchSweep(source, replacements, o)
	if err != nil {
		return result, err
	}
	result = DchResult{Swept: swept, Classes: classes, Replacements: replacements, Stats: stats}
	return result, nil
}

// ComputeDchChoices additionally attaches same-polarity AND alternatives to an
// independent source clone in deterministic representative/object order.
func ComputeDchChoices(source *AigManager, options DchOptions) (DchResult, error) {
	result, err := ComputeDchEquivalences(source, options)
	if err != nil {
		return DchResult{}, err
	}
	o, _ := normalizeDchOptions(options)
	graph, classes, choices, err := dchDeriveChoiceGraph(source, result.Replacements, o)
	if err != nil {
		return DchResult{}, err
	}
	result.Graph = graph
	result.Stats.ChoiceClasses = classes
	result.Stats.Choices = choices
	return result, nil
}

// DeriveDchTotalAig combines equal-terminal AIGs by sharing CIs and copying
// every output cone. It creates one CO per output position, driven by the first
// graph, matching the cumulative graph used before DCH choice discovery.
func DeriveDchTotalAig(graphs []*AigManager, options DchOptions) (*AigManager, error) {
	o, err := normalizeDchOptions(options)
	if err != nil {
		return nil, err
	}
	if len(graphs) == 0 || graphs[0] == nil {
		return nil, ErrDimension
	}
	inputs, outputs := graphs[0].InputCount(), graphs[0].OutputCount()
	for _, graph := range graphs {
		if graph == nil || graph.InputCount() != inputs || graph.OutputCount() != outputs || graph.RegisterCount() != graphs[0].RegisterCount() {
			return nil, ErrDimension
		}
		if err = graph.Check(); err != nil {
			return nil, err
		}
	}
	result := NewAigManager()
	result.Name, result.Spec = graphs[0].Name, graphs[0].Spec
	for i := 0; i < inputs; i++ {
		if _, err = result.CreateCI(); err != nil {
			return nil, err
		}
	}
	firstDrivers := make([]AigLit, outputs)
	for output := 0; output < outputs; output++ {
		for graphIndex, graph := range graphs {
			if err = dchStop(o); err != nil {
				return nil, err
			}
			driver := graph.objects[graph.cos[output]].Fanin0
			copied, copyErr := TransferAigCone(graph, result, driver, inputs, AigRewriteOptions{MaxObjects: o.MaxObjects, Stop: o.Stop})
			if copyErr != nil {
				return nil, copyErr
			}
			if graphIndex == 0 {
				firstDrivers[output] = copied
			}
		}
	}
	for _, driver := range firstDrivers {
		if _, err = result.CreateCO(driver); err != nil {
			return nil, err
		}
	}
	if err = result.SetRegisterCount(graphs[0].registers); err != nil {
		return nil, err
	}
	if err = result.SetConstraintCount(graphs[0].constraints); err != nil {
		return nil, err
	}
	return result, result.Check()
}
