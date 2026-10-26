package abc

// Tent/rank analysis and the all-output bounded-checking entry point from
// src/sat/bmc/bmcBmcS.c.

type GiaTentBoundary struct {
	Object                  GiaID
	Tent, Rank              int
	PrimaryInput, Candidate bool
}

type GiaTentAnalysis struct {
	Tents, Reachable, MaxRank  int
	TentByObject, RankByObject []int
	Candidates                 []GiaID
	Boundaries                 []GiaTentBoundary
	RankProfile                []int
}

func giaTentCandidateMarks(manager *GiaManager, stop func() bool) (int, []int, []bool, error) {
	firstTent := make([]int, manager.ObjectCount())
	for index := range firstTent {
		firstTent[index] = -1
	}
	firstTent[0] = 0
	candidates := make([]bool, manager.ObjectCount())
	roToRI := make([]GiaID, manager.ObjectCount())
	for index := range roToRI {
		roToRI[index] = InvalidGiaID
	}
	for register, ro := range manager.cis[manager.PrimaryInputCount():] {
		roToRI[ro] = manager.cos[manager.PrimaryOutputCount()+register]
	}
	roots := make([]GiaID, 0, manager.PrimaryOutputCount()+manager.RegisterCount())
	for _, po := range manager.cos[:manager.PrimaryOutputCount()] {
		driver, _ := manager.objects[po].Fanin0(po)
		roots = append(roots, driver.ID())
	}
	begin, tents := 0, 0
	for begin < len(roots) {
		end := len(roots)
		for _, root := range roots[begin:end] {
			stack := []GiaID{root}
			for len(stack) > 0 {
				if stop != nil && stop() {
					return 0, nil, nil, ErrCancelled
				}
				last := len(stack) - 1
				id := stack[last]
				stack = stack[:last]
				if firstTent[id] >= 0 {
					if firstTent[id] < tents {
						candidates[id] = true
					}
					continue
				}
				firstTent[id] = tents
				object := manager.objects[id]
				if object.IsAnd() {
					left, _ := object.Fanin0(id)
					right, _ := object.Fanin1(id)
					stack = append(stack, right.ID(), left.ID())
				} else if ri := roToRI[id]; ri != InvalidGiaID {
					driver, _ := manager.objects[ri].Fanin0(ri)
					roots = append(roots, driver.ID())
				}
			}
		}
		begin, tents = end, tents+1
	}
	candidates[0] = false
	return tents, firstTent, candidates, nil
}

// AnalyzeTents exposes the owned schedule used by NewGiaRankedUnroller. It
// replaces traversal marks and the source's parallel vectors with object-indexed
// tent/rank arrays and deterministic boundary/candidate slices.
func (manager *GiaManager) AnalyzeTents(options GiaRankedUnrollOptions) (GiaTentAnalysis, error) {
	var result GiaTentAnalysis
	if manager == nil || manager.PrimaryOutputCount() == 0 {
		return result, ErrDimension
	}
	if err := manager.Check(); err != nil {
		return result, err
	}
	for id, object := range manager.objects {
		if object.IsBuffer() || object.IsXor() {
			return result, ErrUnsupported
		}
		if _, mux := manager.MuxControl(GiaID(id)); mux {
			return result, ErrUnsupported
		}
	}
	unroller, err := NewGiaRankedUnroller(manager, options)
	if err != nil {
		return result, err
	}
	defer unroller.Stop()
	result.TentByObject = make([]int, manager.ObjectCount())
	result.RankByObject = make([]int, manager.ObjectCount())
	for index := range result.TentByObject {
		result.TentByObject[index], result.RankByObject[index] = -1, -1
	}
	result.TentByObject[0], result.RankByObject[0] = 0, 0
	stats := unroller.Statistics()
	tents, firstTent, candidates, err := giaTentCandidateMarks(manager, options.Stop)
	if err != nil {
		return GiaTentAnalysis{}, err
	}
	result.Tents, result.Reachable, result.MaxRank = tents, stats.ReachableObjects, stats.MaxRank
	result.RankProfile = unroller.RankProfile()
	for _, ids := range unroller.byTent {
		for _, id := range ids {
			node := unroller.nodes[id]
			rank := node.rankMax - 1
			result.TentByObject[id], result.RankByObject[id] = firstTent[id], rank
			candidate := candidates[id]
			if candidate {
				result.Candidates = append(result.Candidates, id)
			}
			if node.isPI || candidate {
				result.Boundaries = append(result.Boundaries, GiaTentBoundary{Object: id, Tent: firstTent[id], Rank: rank, PrimaryInput: node.isPI, Candidate: candidate})
			}
		}
	}
	return result, nil
}

type GiaSuperBMCOptions struct {
	Frames, FrameBatch, Workers                 int
	MaxObjects, MaxMappingEntries, MaxVariables int
	MaxClauses, MaxLiterals                     int
	GlobalConflicts                             uint64
	SAT                                         SatLimits
	OnFrameDone                                 func(frame, property int, failed bool)
}

// SuperBoundedCheck preserves bmcS target semantics over the concrete owned
// unroller/Tseitin/CDCL stack. FrameBatch and Workers are performance choices;
// one deterministic solver returns the same Boolean answer and witness without
// retaining a thread/solver portfolio. Every PO is a target, including outputs
// carrying AIGER constraint metadata.
func (manager *GiaManager) SuperBoundedCheck(options GiaSuperBMCOptions) (BmcResult, error) {
	var result BmcResult
	result.Frame, result.Property = -1, -1
	if manager == nil {
		return result, ErrDimension
	}
	if err := checkBmciPlainGraph(manager); err != nil {
		return result, err
	}
	if options.Frames < 1 || options.FrameBatch < 0 || options.Workers < 0 || options.Workers >= 100 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
		return result, ErrDimension
	}
	if options.FrameBatch == 0 {
		options.FrameBatch = 1
	}
	if options.Workers == 0 {
		options.Workers = 1
	}
	aig, _, err := manager.ToAig()
	if err != nil {
		return result, err
	}
	interval, err := aig.IntervalBoundedCheck(AigIntervalBMCOptions{
		Frames:     options.Frames,
		MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries,
		MaxVariables: options.MaxVariables, MaxClauses: options.MaxClauses, MaxLiterals: options.MaxLiterals,
		GlobalConflicts: options.GlobalConflicts, SAT: options.SAT,
		OnTargetDone: func(frame, property int, status SatStatus) {
			if options.OnFrameDone != nil && status != SatUnknown {
				options.OnFrameDone(frame, property, status == SatSatisfiable)
			}
		},
	})
	if err != nil {
		return result, err
	}
	result.Status, result.Reason, result.SatReason = interval.Status, interval.Reason, interval.SatReason
	result.CheckedFrames, result.Frame, result.Property = interval.CheckedFrames, interval.Frame, interval.Property
	result.Counterexample, result.Stats = interval.Counterexample, interval.Stats
	return result, nil
}
