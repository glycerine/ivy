package abc

// Whole-network unit scheduling from aig/gia/giaSatLut.c:Gia_ManLutSat.

import "encoding/binary"

type GiaUnitRemapPassOptions struct {
	MaxPackedEntries      int // packed wrapper output storage; zero: one million
	WindowLimit           int // zero: 64; supported range 1..128
	Reverse               bool
	MaxImprovements       int // zero: unlimited, as in C
	MaxPivots, MaxWindows int // zero: one million pivots, 65536 distinct windows
	Search                GiaUnitWindowSearchOptions
}
type GiaUnitRemapAttempt struct {
	Pivot  GiaID
	Ands   int
	Reason string // improved, unchanged, repeated, small, large
}
type GiaUnitRemapPassResult struct {
	DiscoveryBuilds, ReferenceRefreshes, DiscoveryQueries int
	SolverBuilds, WindowResets                            int // compiled cardinality prefixes and successful window rebindings
	Network                                               *GiaLUTNetwork
	Functions                                             []GiaLUTFunction
	Timing                                                GiaLUTUnitTiming
	Attempts                                              []GiaUnitRemapAttempt
	Improvements                                          int
}

// RemapUnitPass visits currently mapped IDs in ascending order, so replacements
// can remove future pivots or introduce later ones. Window hashes persist across
// improvements, matching C's schedule. Native windows below ten ANDs are skipped.
// The result owns a complete replacement; errors expose no partial pass. Search
// budgets apply per window/phase, bounded by MaxPivots and MaxWindows.
// Discovery scratch/static fanouts and SAT prefixes are reused across windows.
// Box-aware entry points share this schedule with owned timing-state updates.
func (n *GiaLUTNetwork) RemapUnitPass(m *GiaManager, delayStart int, o GiaUnitRemapPassOptions) (GiaUnitRemapPassResult, error) {
	r, e := n.remapPass(m, delayStart, o, nil, nil)
	return r.GiaUnitRemapPassResult, e
}

type giaScheduledRemapResult struct {
	GiaUnitRemapPassResult
	edges      *GiaEdgeAssignment
	edgeTiming GiaEdgeTiming
	manager    *TimManager
	order      []GiaID
}
type giaEdgeRemapContext struct {
	maximum  int
	prior    *GiaEdgeAssignment
	baseline GiaEdgeTiming
}

func (n *GiaLUTNetwork) remapPass(m *GiaManager, delayStart int, o GiaUnitRemapPassOptions, edge *giaEdgeRemapContext, boxes *TimManager) (giaScheduledRemapResult, error) {
	fail := func(e error) (giaScheduledRemapResult, error) { return giaScheduledRemapResult{}, e }
	if o.MaxPackedEntries < 0 || o.WindowLimit < 0 || o.WindowLimit > 128 || o.MaxImprovements < 0 || o.MaxPivots < 0 || o.MaxWindows < 0 {
		return fail(ErrDimension)
	}
	if o.Search.LutSize < 0 || o.Search.LutSize > 6 {
		return fail(ErrDimension)
	}
	if edge != nil && o.Search.LutSize > 4 {
		return fail(ErrUnsupported)
	}
	if o.Search.MaxIterations < 0 || o.Search.Cuts.MaxCuts < 0 || o.Search.Cuts.MaxVisits < 0 || o.Search.CNF.MaxVariables < 0 || o.Search.CNF.MaxClauses < 0 || o.Search.CNF.MaxLiterals < 0 || o.Search.Analysis.MaxObjects < 0 || o.Search.Analysis.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if o.Search.MaxIterations > 10000 {
		return fail(ErrLimit)
	}
	if o.WindowLimit == 0 {
		o.WindowLimit = 64
	}
	if o.MaxPivots == 0 {
		o.MaxPivots = 1 << 20
	}
	if o.MaxWindows == 0 {
		o.MaxWindows = 1 << 16
	}
	analysis := o.Search.Analysis
	prior := analysis.Stop
	analysis.Stop = func() bool { return o.Search.Stop != nil && o.Search.Stop() || prior != nil && prior() }
	if _, e := n.ValidateFunctions(m, analysis); e != nil {
		return fail(e)
	}
	current, e := n.WithWindow(nil, nil, analysis)
	if e != nil {
		return fail(e)
	}
	r := giaScheduledRemapResult{}
	if boxes != nil {
		r.manager = boxes.Clone()
	}
	if edge != nil {
		r.edges = &GiaEdgeAssignment{rows: append([][2]GiaID(nil), edge.prior.rows...), skipped: edge.prior.skipped}
		r.edgeTiming = edge.baseline
	}
	log := 0
	for 1<<uint(log) < o.WindowLimit {
		log++
	}
	session := &giaUnitWindowSolverSession{log: log}
	discovery := &giaWindowDiscoverySession{}
	seen := make(map[string]struct{}) // lookup only; iteration order comes from the ID scan
	for id := 1; id < len(current.fanins); id++ {
		if analysis.Stop() {
			return fail(ErrCancelled)
		}
		if len(current.fanins[id]) == 0 {
			continue
		}
		if len(r.Attempts) == o.MaxPivots {
			return fail(ErrLimit)
		}
		attempt := GiaUnitRemapAttempt{Pivot: GiaID(id)}
		window, e := current.discoverWindow(m, GiaID(id), o.WindowLimit, o.Reverse, analysis, discovery)
		if e != nil {
			return fail(e)
		}
		discovery.queries++
		attempt.Ands = len(window.Ands)
		if !window.Found {
			attempt.Reason = "small"
			r.Attempts = append(r.Attempts, attempt)
			continue
		}
		bytes := make([]byte, 4*len(window.Ands))
		for i, v := range window.Ands {
			binary.LittleEndian.PutUint32(bytes[4*i:], uint32(v))
		}
		key := string(bytes)
		if _, ok := seen[key]; ok {
			attempt.Reason = "repeated"
			r.Attempts = append(r.Attempts, attempt)
			continue
		}
		if len(seen) == o.MaxWindows {
			return fail(ErrLimit)
		}
		seen[key] = struct{}{}
		if len(window.Leaves) > 128 || len(window.Ands) > o.WindowLimit {
			attempt.Reason = "large"
			r.Attempts = append(r.Attempts, attempt)
			continue
		}
		if len(window.Ands) < 10 {
			attempt.Reason = "small"
			r.Attempts = append(r.Attempts, attempt)
			continue
		}
		attempt.Reason = "unchanged"
		if edge != nil && o.Search.AreaOnly {
			raw, e := current.searchWindow(m, window.Leaves, window.Ands, window.Roots, 0, o.Search, session, &giaWindowSearchTiming{maximum: edge.maximum, boxes: r.manager})
			if e != nil {
				return fail(e)
			}
			if raw.Improved {
				proposal, e := current.WithWindow(window.Ands, raw.BestWindow, analysis)
				if e != nil {
					return fail(e)
				}
				if _, e = proposal.ValidateFunctions(m, analysis); e != nil {
					return fail(e)
				}
				current = proposal
				r.Improvements++
				attempt.Reason = "improved"
			}
		} else if edge == nil && boxes == nil {
			proposal, e := current.remapUnitWindow(m, window.Leaves, window.Ands, window.Roots, delayStart, o.Search, session)
			if e != nil {
				return fail(e)
			}
			if proposal.Replacement.Network != nil {
				current = proposal.Replacement.Network
				r.Improvements++
				attempt.Reason = "improved"
			}
		} else if boxes == nil {
			proposal, e := current.remapEdgeWindow(m, window.Leaves, window.Ands, window.Roots, edge.maximum, o.Search, session)
			if e != nil {
				return fail(e)
			}
			if proposal.Replacement.Network != nil {
				current = proposal.Replacement.Network
				r.edges = proposal.Replacement.Edges
				r.edgeTiming = proposal.Replacement.Timing
				r.Improvements++
				attempt.Reason = "improved"
			}
		} else if edge == nil {
			proposal, e := current.remapUnitWindowWithBoxes(m, r.manager, window.Leaves, window.Ands, window.Roots, delayStart, o.Search, session)
			if e != nil {
				return fail(e)
			}
			if proposal.Replacement.Network != nil {
				current = proposal.Replacement.Network
				r.manager = proposal.Replacement.Timing.Manager
				r.Improvements++
				attempt.Reason = "improved"
			}
		} else {
			proposal, e := current.remapEdgeWindowWithBoxes(m, r.manager, window.Leaves, window.Ands, window.Roots, edge.maximum, o.Search, session)
			if e != nil {
				return fail(e)
			}
			if proposal.Replacement.Network != nil {
				current = proposal.Replacement.Network
				r.manager = proposal.Replacement.Manager
				r.edges = proposal.Replacement.Edges
				r.edgeTiming = proposal.Replacement.Timing
				r.Improvements++
				attempt.Reason = "improved"
			}
		}
		r.Attempts = append(r.Attempts, attempt)
		if o.MaxImprovements > 0 && r.Improvements == o.MaxImprovements {
			break
		}
	}
	r.DiscoveryBuilds = discovery.builds
	r.ReferenceRefreshes = discovery.refreshes
	r.DiscoveryQueries = discovery.queries
	r.SolverBuilds = session.builds
	r.WindowResets = session.resets
	r.Network = current
	r.Functions, e = current.ValidateFunctions(m, analysis)
	if e != nil {
		return fail(e)
	}
	if edge == nil {
		if boxes == nil {
			r.Timing, e = current.UnitTiming(m, delayStart, analysis)
		} else {
			var timing GiaBoxLUTUnitTiming
			timing, e = current.UnitTimingWithBoxes(m, r.manager, delayStart, analysis)
			if e == nil {
				r.Timing, r.manager, r.order = timing.Timing, timing.Manager, timing.Order
			}
		}
		if e != nil {
			return fail(e)
		}
	} else {
		mapping, e := current.selectionMapping(analysis)
		if e != nil {
			return fail(e)
		}
		if boxes == nil {
			r.edgeTiming, e = r.edges.EvaluateDelay(m, mapping, analysis)
		} else {
			var timing GiaBoxEdgeTiming
			timing, e = r.edges.EvaluateDelayWithBoxes(m, mapping, r.manager, analysis)
			if e == nil {
				r.edgeTiming, r.manager, r.order = timing.Timing, timing.Manager, timing.Order
			}
		}
		if e != nil {
			return fail(e)
		}
		if !o.Search.AreaOnly && r.Improvements > 0 && r.edgeTiming.Maximum > edge.maximum {
			return fail(ErrInvariant)
		}
	}
	return r, nil
}
