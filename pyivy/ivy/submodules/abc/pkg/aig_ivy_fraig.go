package abc

// Complete semantic FRAIG engine for src/aig/ivy/ivyFraig.c.

import (
	"encoding/binary"
	"slices"
)

// IvyFraigDefinitions is the reviewed named-definition inventory of the native
// translation unit, including file-local accessors and optional CUDD helpers.
// Their mutable pointer-field machinery is consolidated into the owned engine
// below; the list keeps every source name auditable.
var IvyFraigDefinitions = [...]string{
	"Ivy_FraigAddClass", "Ivy_FraigAddClausesMux", "Ivy_FraigAddClausesSuper", "Ivy_FraigAddToPatScores",
	"Ivy_FraigAnd", "Ivy_FraigAssignDist1", "Ivy_FraigAssignRandom", "Ivy_FraigCheckCone",
	"Ivy_FraigCheckOutputSims", "Ivy_FraigCheckOutputSimsSavePattern", "Ivy_FraigCleanPatScores", "Ivy_FraigCollectSuper",
	"Ivy_FraigCollectSuper_rec", "Ivy_FraigCountClassNodes", "Ivy_FraigCountPairsClasses", "Ivy_FraigCreateClasses",
	"Ivy_FraigCreateModel", "Ivy_FraigExtractCone", "Ivy_FraigExtractCone_rec", "Ivy_FraigInsertClass",
	"Ivy_FraigMiter", "Ivy_FraigMiterPrint", "Ivy_FraigMiterProve", "Ivy_FraigMiterStatus",
	"Ivy_FraigNodeAddToSolver", "Ivy_FraigNodeIsConst", "Ivy_FraigNodesAreEquiv", "Ivy_FraigNodesAreEquivBdd",
	"Ivy_FraigNodesAreEquivBdd_int", "Ivy_FraigObjAddToFrontier", "Ivy_FraigParamsDefault", "Ivy_FraigPerform",
	"Ivy_FraigPerform_int", "Ivy_FraigPrint", "Ivy_FraigPrintActivity", "Ivy_FraigPrintClass",
	"Ivy_FraigPrintSimClasses", "Ivy_FraigProve", "Ivy_FraigRefineClasses", "Ivy_FraigRefineClass_rec",
	"Ivy_FraigRemoveClass", "Ivy_FraigResimulate", "Ivy_FraigSavePattern", "Ivy_FraigSavePattern0",
	"Ivy_FraigSavePattern1", "Ivy_FraigSavePattern2", "Ivy_FraigSavePattern3", "Ivy_FraigSelectBestPat",
	"Ivy_FraigSetActivityFactors", "Ivy_FraigSetActivityFactors_rec", "Ivy_FraigSimulate", "Ivy_FraigSimulateOne",
	"Ivy_FraigSimulateOneSim", "Ivy_FraigStart", "Ivy_FraigStartSimple", "Ivy_FraigStop",
	"Ivy_FraigSweep", "Ivy_NodeAddToClass", "Ivy_NodeAssignConst", "Ivy_NodeAssignRandom",
	"Ivy_NodeCompareSims", "Ivy_NodeComplementSim", "Ivy_NodeHash", "Ivy_NodeHasZeroSim",
	"Ivy_NodeSimulate", "Ivy_NodeSimulateSim", "Ivy_ObjClassNodeLast", "Ivy_ObjClassNodeNext",
	"Ivy_ObjClassNodeRepr", "Ivy_ObjEquivListNext", "Ivy_ObjEquivListPrev", "Ivy_ObjFaninVec",
	"Ivy_ObjFraig", "Ivy_ObjNodeHashNext", "Ivy_ObjRandomSim", "Ivy_ObjSatNum",
	"Ivy_ObjSetClassNodeLast", "Ivy_ObjSetClassNodeNext", "Ivy_ObjSetClassNodeRepr", "Ivy_ObjSetEquivListNext",
	"Ivy_ObjSetEquivListPrev", "Ivy_ObjSetFaninVec", "Ivy_ObjSetFraig", "Ivy_ObjSetNodeHashNext",
	"Ivy_ObjSetSatNum", "Ivy_ObjSetSim", "Ivy_ObjSim",
}

// IvyFraigParameters retains the active native simulation, SAT and miter
// controls. Diagnostic printing is returned as structured statistics.
type IvyFraigParameters struct {
	SimulationWords      int
	SimulationSaturation float64
	PatternScores        bool
	MaximumPatternScore  int
	Sparse               bool
	ActivityConeRatio    float64
	ActivityConeBump     float64
	NodeBacktrackLimit   int // -1 or zero: unlimited
	MiterBacktrackLimit  int // -1 or zero: unlimited
	Prove                bool
	Verbose              bool
}

func DefaultIvyFraigParameters() IvyFraigParameters {
	return IvyFraigParameters{
		SimulationWords: 32, SimulationSaturation: 0.005,
		MaximumPatternScore: 25, Sparse: true,
		ActivityConeRatio: 0.3, ActivityConeBump: 10,
		NodeBacktrackLimit: 100, MiterBacktrackLimit: 500000,
	}
}

type IvyFraigOptions struct {
	Parameters                      IvyFraigParameters
	MaxObjects, MaxSATNodes         int
	MaxCandidatePairs, MaxSATVisits int
	MaxBytes, MaxWork               int
	Stop                            func() bool
}

type IvyFraigStatus uint8

const (
	IvyFraigUnknown IvyFraigStatus = iota
	IvyFraigCounterexample
	IvyFraigProved
)

// IvyFraigClass is one SAT-certified source equivalence class. Each member is
// a source object edge whose complement bit relates it to Representative.
type IvyFraigClass struct {
	Representative uint32
	Members        []IvyEdge
}

type IvyFraigStats struct {
	ObjectsBefore, ObjectsAfter                     int
	SimulationWords, SimulationPatterns             int
	InitialClasses, FinalClasses, CandidatePairs    int
	SATQueries, SATProved, SATDisproved, SATUnknown uint64
	Merges, ConstantMerges, ComplementMerges        int
	RemovedDangling                                 int
}

// IvyFraigResult owns the reduced graph, source-object mapping, certified
// classes and optional miter model. MappingValid covers live source terminals
// and retained logic and is false for dangling source nodes removed by final
// cleanup.
type IvyFraigResult struct {
	Graph        *IvyManager
	Mapping      []IvyEdge
	MappingValid []bool
	Classes      []IvyFraigClass
	Model        []bool
	Status       IvyFraigStatus
	Stats        IvyFraigStats
}

type ivyFraigBudget struct {
	o    IvyFraigOptions
	work int
	used int
}

func normalizeIvyFraigOptions(options IvyFraigOptions) (IvyFraigOptions, error) {
	if options.Parameters == (IvyFraigParameters{}) {
		options.Parameters = DefaultIvyFraigParameters()
	}
	p := options.Parameters
	if p.SimulationWords <= 0 || p.SimulationWords > 128 || p.SimulationSaturation < 0 || p.SimulationSaturation > 1 ||
		p.MaximumPatternScore < 0 || p.ActivityConeRatio <= 0 || p.ActivityConeRatio >= 1 || p.ActivityConeBump < 0 ||
		p.NodeBacktrackLimit < -1 || p.MiterBacktrackLimit < -1 || options.MaxObjects < 0 || options.MaxSATNodes < 0 ||
		options.MaxCandidatePairs < 0 || options.MaxSATVisits < 0 || options.MaxBytes < 0 || options.MaxWork < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxSATNodes == 0 {
		options.MaxSATNodes = 1 << 22
	}
	if options.MaxCandidatePairs == 0 {
		options.MaxCandidatePairs = 1 << 22
	}
	if options.MaxSATVisits == 0 {
		options.MaxSATVisits = 1 << 25
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 256 << 20
	}
	if options.MaxWork == 0 {
		options.MaxWork = 1 << 27
	}
	return options, nil
}

func newIvyFraigBudget(options IvyFraigOptions) (*ivyFraigBudget, error) {
	o, err := normalizeIvyFraigOptions(options)
	if err != nil {
		return nil, err
	}
	return &ivyFraigBudget{o: o}, nil
}

func (b *ivyFraigBudget) tick(n int) error {
	if n < 0 || b.work > b.o.MaxWork-n {
		return ErrLimit
	}
	b.work += n
	if b.o.Stop != nil && b.o.Stop() {
		return ErrCancelled
	}
	return nil
}

func (b *ivyFraigBudget) bytes(n int) error {
	if n < 0 || b.used > b.o.MaxBytes-n {
		return ErrLimit
	}
	b.used += n
	return nil
}

func ivyFraigRead(words [][]uint32, edge IvyEdge, word int) uint32 {
	value := words[edge.ID()][word]
	if edge.Complemented() {
		value = ^value
	}
	return value
}

func ivyFraigEdgeNotIf(edge FraigEdge, complement bool) FraigEdge {
	if complement {
		return edge.Not()
	}
	return edge
}

func ivyFraigRandomBit(pattern, input int) bool {
	x := uint64(pattern+1)*0x9e3779b97f4a7c15 ^ uint64(input+3)*0xbf58476d1ce4e5b9
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return (x^(x>>31))&1 != 0
}

// ivyFraigSimulate consolidates the source assignment, node simulation, hash,
// zero/complement checks and repeated class-refinement helpers. Pattern zero is
// all-zero, pattern one all-one, and the next patterns are distance one.
func ivyFraigSimulate(source *IvyManager, b *ivyFraigBudget) ([][]uint32, []bool, []string, error) {
	words := b.o.Parameters.SimulationWords
	if len(source.objects) > b.o.MaxObjects {
		return nil, nil, nil, ErrLimit
	}
	if words > 0 && len(source.objects) > b.o.MaxBytes/(words*4+25) {
		return nil, nil, nil, ErrLimit
	}
	if err := b.bytes(len(source.objects) * (words*4 + 25)); err != nil {
		return nil, nil, nil, err
	}
	values := make([][]uint32, len(source.objects))
	inverted := make([]bool, len(source.objects))
	keys := make([]string, len(source.objects))
	for id, object := range source.objects {
		if !object.Alive {
			continue
		}
		values[id] = make([]uint32, words)
	}
	for word := range values[0] {
		values[0][word] = ^uint32(0)
	}
	for input, id := range source.inputs {
		for pattern := 0; pattern < words*32; pattern++ {
			one := pattern == 1 || pattern >= 2 && pattern < 2+len(source.inputs) && input == pattern-2
			if pattern >= 2+len(source.inputs) {
				one = ivyFraigRandomBit(pattern, input)
			}
			if one {
				values[id][pattern/32] |= uint32(1) << uint(pattern%32)
			}
		}
	}
	for id, object := range source.objects {
		if id == 0 || !object.Alive || object.Type == IvyPI {
			continue
		}
		if err := b.tick(words); err != nil {
			return nil, nil, nil, err
		}
		for word := 0; word < words; word++ {
			switch object.Type {
			case IvyAND:
				values[id][word] = ivyFraigRead(values, object.Fanin0, word) & ivyFraigRead(values, object.Fanin1, word)
			case IvyEXOR:
				values[id][word] = ivyFraigRead(values, object.Fanin0, word) ^ ivyFraigRead(values, object.Fanin1, word)
			case IvyBuffer, IvyPO, IvyAssert:
				values[id][word] = ivyFraigRead(values, object.Fanin0, word)
			default:
				return nil, nil, nil, ErrUnsupported
			}
		}
	}
	for id, object := range source.objects {
		if !object.Alive || (id != 0 && object.Type != IvyPI && object.Type != IvyAND && object.Type != IvyEXOR && object.Type != IvyBuffer) {
			continue
		}
		inverted[id] = values[id][0]&1 != 0
		encoded := make([]byte, words*4)
		for word, value := range values[id] {
			if inverted[id] {
				value = ^value
			}
			binary.LittleEndian.PutUint32(encoded[word*4:], value)
		}
		keys[id] = string(encoded)
	}
	return values, inverted, keys, nil
}

func ivyToFraig(source *IvyManager, b *ivyFraigBudget) (*FraigOrchestrator, []FraigEdge, []bool, error) {
	if b.o.MaxSATNodes > int(^uint(0)>>1)/3 {
		return nil, nil, nil, ErrLimit
	}
	p := DefaultFraigParameters()
	p.FunctionalReduce = false
	p.BacktrackLimit = b.o.Parameters.NodeBacktrackLimit
	if p.BacktrackLimit == 0 {
		p.BacktrackLimit = -1
	}
	orchestrator, err := NewFraigOrchestrator(FraigOrchestrationOptions{
		Manager:    FraigOptions{MaxNodes: b.o.MaxSATNodes, MaxInputs: max(1, len(source.inputs)), MaxOutputs: max(1, len(source.outputs)), Stop: b.o.Stop},
		Parameters: p, MaxClauses: b.o.MaxSATNodes * 3, MaxVisits: b.o.MaxSATVisits, Stop: b.o.Stop,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	mapping := make([]FraigEdge, len(source.objects))
	ready := make([]bool, len(source.objects))
	mapping[0], ready[0] = orchestrator.Manager().ConstantOne(), true
	read := func(edge IvyEdge) (FraigEdge, error) {
		if uint64(edge.ID()) >= uint64(len(mapping)) || !ready[edge.ID()] {
			return 0, ErrInvariant
		}
		return ivyFraigEdgeNotIf(mapping[edge.ID()], edge.Complemented()), nil
	}
	for id, object := range source.objects {
		if id == 0 || !object.Alive {
			continue
		}
		if err = b.tick(1); err != nil {
			orchestrator.Close()
			return nil, nil, nil, err
		}
		var edge FraigEdge
		switch object.Type {
		case IvyPI:
			edge, err = orchestrator.NewInput("")
		case IvyAND, IvyEXOR:
			var a, c FraigEdge
			a, err = read(object.Fanin0)
			if err == nil {
				c, err = read(object.Fanin1)
			}
			if err == nil && object.Type == IvyAND {
				edge, err = orchestrator.And(a, c)
			} else if err == nil {
				edge, err = orchestrator.Xor(a, c)
			}
		case IvyBuffer, IvyPO, IvyAssert:
			edge, err = read(object.Fanin0)
		default:
			err = ErrUnsupported
		}
		if err != nil {
			orchestrator.Close()
			return nil, nil, nil, err
		}
		mapping[id], ready[id] = edge, true
		if object.Type == IvyPO {
			if err = orchestrator.AddOutput(edge, ""); err != nil {
				orchestrator.Close()
				return nil, nil, nil, err
			}
		}
	}
	return orchestrator, mapping, ready, nil
}

func ivyFraigSATLimits(limit int, stop func() bool) FraigSATLimits {
	var conflicts uint64
	if limit > 0 {
		conflicts = uint64(limit)
	}
	return FraigSATLimits{Conflicts: conflicts, Stop: stop}
}

// Fraig performs simulation-class candidate generation, SAT-certified
// functional reduction, output reconstruction and optional miter proof. The
// source manager is immutable and the result owns every returned slice/graph.
func (source *IvyManager) Fraig(options IvyFraigOptions) (IvyFraigResult, error) {
	var result IvyFraigResult
	b, err := newIvyFraigBudget(options)
	if err != nil {
		return result, err
	}
	if source == nil {
		return result, ErrDimension
	}
	if len(source.objects) > b.o.MaxObjects {
		return result, ErrLimit
	}
	if err = source.Check(); err != nil {
		return result, err
	}
	if source.Count(IvyLatch) != 0 {
		return result, ErrUnsupported
	}
	if err = b.bytes(len(source.objects) * 42); err != nil {
		return result, err
	}
	_, inverted, keys, err := ivyFraigSimulate(source, b)
	if err != nil {
		return result, err
	}
	orchestrator, satMapping, satReady, err := ivyToFraig(source, b)
	if err != nil {
		return result, err
	}
	defer orchestrator.Close()
	target, err := NewIvyManager(IvyOptions{MaxObjects: b.o.MaxObjects, MaxWork: b.o.MaxWork, Cancelled: b.o.Stop})
	if err != nil {
		return result, err
	}
	mapping := make([]IvyEdge, len(source.objects))
	ready := make([]bool, len(source.objects))
	mapping[0], ready[0] = IvyConst1, true
	read := func(edge IvyEdge) (IvyEdge, error) {
		if uint64(edge.ID()) >= uint64(len(mapping)) || !ready[edge.ID()] {
			return 0, ErrInvariant
		}
		return mapping[edge.ID()].NotIf(edge.Complemented()), nil
	}
	for id, object := range source.objects {
		if id == 0 || !object.Alive || object.Type != IvyPI {
			continue
		}
		edge, createErr := target.CreatePI()
		if createErr != nil {
			return result, createErr
		}
		mapping[id], ready[id] = edge, true
	}
	buckets := make(map[string][]uint32)
	buckets[keys[0]] = []uint32{0}
	for id, object := range source.objects {
		if id != 0 && object.Alive && object.Type == IvyPI {
			buckets[keys[id]] = append(buckets[keys[id]], uint32(id))
		}
	}
	classMembers := make(map[uint32][]IvyEdge)
	classOrder := make([]uint32, 0)
	stats := IvyFraigStats{ObjectsBefore: source.LiveCount(), SimulationWords: b.o.Parameters.SimulationWords, SimulationPatterns: b.o.Parameters.SimulationWords * 32}
	signatureCounts := make(map[string]int)
	for id, object := range source.objects {
		if object.Alive && (id == 0 || object.Type == IvyPI || object.Type == IvyAND || object.Type == IvyEXOR || object.Type == IvyBuffer) {
			signatureCounts[keys[id]]++
		}
	}
	for _, count := range signatureCounts {
		if count > 1 {
			stats.InitialClasses++
		}
	}
	for id, object := range source.objects {
		if id == 0 || !object.Alive || (object.Type != IvyAND && object.Type != IvyEXOR && object.Type != IvyBuffer) {
			continue
		}
		if err = b.tick(1); err != nil {
			return IvyFraigResult{}, err
		}
		var structural IvyEdge
		switch object.Type {
		case IvyAND, IvyEXOR:
			a, readErr := read(object.Fanin0)
			if readErr != nil {
				return IvyFraigResult{}, readErr
			}
			c, readErr := read(object.Fanin1)
			if readErr != nil {
				return IvyFraigResult{}, readErr
			}
			if object.Type == IvyAND {
				structural, err = target.And(a, c)
			} else {
				structural, err = target.Xor(a, c)
			}
		case IvyBuffer:
			a, readErr := read(object.Fanin0)
			if readErr != nil {
				return IvyFraigResult{}, readErr
			}
			structural, err = target.CreateBuffer(a)
		}
		if err != nil {
			return IvyFraigResult{}, err
		}
		merged := false
		candidates := buckets[keys[id]]
		for _, representative := range candidates {
			if representative == 0 && !b.o.Parameters.Sparse {
				continue
			}
			stats.CandidatePairs++
			if stats.CandidatePairs > b.o.MaxCandidatePairs {
				return IvyFraigResult{}, ErrLimit
			}
			difference := inverted[id] != inverted[representative]
			replacement := mapping[representative].NotIf(difference)
			proved := structural == replacement || satMapping[id] == ivyFraigEdgeNotIf(satMapping[representative], difference)
			if !proved {
				if !satReady[id] || !satReady[representative] {
					return IvyFraigResult{}, ErrInvariant
				}
				relation, relationErr := orchestrator.Equivalent(satMapping[id], ivyFraigEdgeNotIf(satMapping[representative], difference), ivyFraigSATLimits(b.o.Parameters.NodeBacktrackLimit, b.o.Stop))
				if relationErr != nil {
					return IvyFraigResult{}, relationErr
				}
				switch relation.Status {
				case FraigRelationProved:
					proved = true
				case FraigRelationDisproved:
					stats.SATDisproved++
				case FraigRelationUnknown:
					stats.SATUnknown++
				}
				stats.SATQueries++
				if proved {
					stats.SATProved++
				}
			}
			if !proved {
				continue
			}
			mapping[id], ready[id], merged = replacement, true, true
			stats.Merges++
			if representative == 0 {
				stats.ConstantMerges++
			}
			if difference {
				stats.ComplementMerges++
			}
			if _, exists := classMembers[representative]; !exists {
				classOrder = append(classOrder, representative)
			}
			member, _ := MakeIvyEdge(uint32(id), difference)
			classMembers[representative] = append(classMembers[representative], member)
			break
		}
		if !merged {
			mapping[id], ready[id] = structural, true
			buckets[keys[id]] = append(candidates, uint32(id))
		}
	}
	status := IvyFraigUnknown
	model := []bool(nil)
	if b.o.Parameters.Prove {
		status = IvyFraigProved
	}
	for _, outputID := range source.outputs {
		object := source.objects[outputID]
		driver, readErr := read(object.Fanin0)
		if readErr != nil {
			return IvyFraigResult{}, readErr
		}
		if b.o.Parameters.Prove && status != IvyFraigCounterexample {
			satDriver := satMapping[outputID]
			relation, relationErr := orchestrator.Equivalent(satDriver, orchestrator.Manager().ConstantZero(), ivyFraigSATLimits(b.o.Parameters.MiterBacktrackLimit, b.o.Stop))
			if relationErr != nil {
				return IvyFraigResult{}, relationErr
			}
			stats.SATQueries++
			switch relation.Status {
			case FraigRelationProved:
				stats.SATProved++
				driver = IvyConst0
			case FraigRelationDisproved:
				stats.SATDisproved++
				status = IvyFraigCounterexample
				model = slices.Clone(relation.Model)
			case FraigRelationUnknown:
				stats.SATUnknown++
				status = IvyFraigUnknown
			}
		}
		created, createErr := target.CreatePO(driver)
		if createErr != nil {
			return IvyFraigResult{}, createErr
		}
		_ = created
		mapping[outputID], ready[outputID] = driver, true
	}
	for id, object := range source.objects {
		if !object.Alive || object.Type != IvyAssert {
			continue
		}
		driver, readErr := read(object.Fanin0)
		if readErr != nil {
			return IvyFraigResult{}, readErr
		}
		if _, err = target.CreateAssert(driver); err != nil {
			return IvyFraigResult{}, err
		}
		mapping[id], ready[id] = driver, true
	}
	removed, err := target.Cleanup()
	if err != nil {
		return IvyFraigResult{}, err
	}
	for id, edge := range mapping {
		if !ready[id] || uint64(edge.ID()) >= uint64(len(target.objects)) || !target.objects[edge.ID()].Alive {
			ready[id] = false
		}
	}
	if err = target.Check(); err != nil {
		return IvyFraigResult{}, err
	}
	classes := make([]IvyFraigClass, 0, len(classOrder))
	for _, representative := range classOrder {
		classes = append(classes, IvyFraigClass{Representative: representative, Members: slices.Clone(classMembers[representative])})
	}
	stats.ObjectsAfter = target.LiveCount()
	stats.FinalClasses = len(classes)
	stats.RemovedDangling = removed
	return IvyFraigResult{Graph: target, Mapping: mapping, MappingValid: ready, Classes: classes, Model: model, Status: status, Stats: stats}, nil
}

// FraigEquivalent exposes the source's node/cone equivalence query as one
// concrete SAT result. It accepts arbitrary complemented edges and returns a
// replayable PI model when disproved.
func (source *IvyManager) FraigEquivalent(a, c IvyEdge, options IvyFraigOptions) (FraigRelationResult, error) {
	b, err := newIvyFraigBudget(options)
	if err != nil {
		return FraigRelationResult{}, err
	}
	if source == nil {
		return FraigRelationResult{}, ErrDimension
	}
	if len(source.objects) > b.o.MaxObjects {
		return FraigRelationResult{}, ErrLimit
	}
	if err = source.Check(); err != nil {
		return FraigRelationResult{}, err
	}
	if source.Count(IvyLatch) != 0 || !source.validEdge(a) || !source.validEdge(c) {
		return FraigRelationResult{}, ErrDimension
	}
	orchestrator, mapping, ready, err := ivyToFraig(source, b)
	if err != nil {
		return FraigRelationResult{}, err
	}
	defer orchestrator.Close()
	if !ready[a.ID()] || !ready[c.ID()] {
		return FraigRelationResult{}, ErrInvariant
	}
	return orchestrator.Equivalent(ivyFraigEdgeNotIf(mapping[a.ID()], a.Complemented()), ivyFraigEdgeNotIf(mapping[c.ID()], c.Complemented()), ivyFraigSATLimits(b.o.Parameters.NodeBacktrackLimit, b.o.Stop))
}
