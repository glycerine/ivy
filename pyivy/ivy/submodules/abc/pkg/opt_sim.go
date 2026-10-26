package abc

// Port of src/opt/sim/sim.h, simMan.c, simSeq.c, simSupp.c,
// simSwitch.c, simSym.c, simSymSat.c, simSymSim.c, simSymStr.c, and simUtils.c.

import (
	"fmt"
	"math"
	"math/bits"
	"slices"
	"sort"
	"strings"
	"time"
)

type SimOptions struct {
	MaxObjects, MaxWords, MaxFrames, MaxPairs int
	Seed                                      uint64
	Stop                                      func() bool
	SAT                                       SatLimits
}

func normalizeSimOptions(o SimOptions) (SimOptions, error) {
	if o.MaxObjects < 0 || o.MaxWords < 0 || o.MaxFrames < 0 || o.MaxPairs < 0 {
		return o, ErrDimension
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxWords == 0 {
		o.MaxWords = 1 << 20
	}
	if o.MaxFrames == 0 {
		o.MaxFrames = 1 << 20
	}
	if o.MaxPairs == 0 {
		o.MaxPairs = 1 << 24
	}
	if o.Seed == 0 {
		o.Seed = 0xabc
	}
	return o, nil
}

func simStopped(o SimOptions) bool { return o.Stop != nil && o.Stop() }

type simRandom struct{ state uint64 }

func (r *simRandom) word() uint32 {
	r.state ^= r.state << 13
	r.state ^= r.state >> 7
	r.state ^= r.state << 17
	return uint32(r.state ^ r.state>>32)
}

// SimInfo is owned packed simulation storage indexed by stable AIG object ID.
type SimInfo struct {
	rows   [][]uint32
	words  int
	closed bool
}

func NewSimInfo(entries, words int, clean bool, o SimOptions) (*SimInfo, error) {
	o, err := normalizeSimOptions(o)
	if err != nil || entries < 0 || words < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if entries > o.MaxObjects || words > o.MaxWords || entries != 0 && words > (1<<28)/entries {
		return nil, ErrLimit
	}
	if simStopped(o) {
		return nil, ErrCancelled
	}
	rows := make([][]uint32, entries)
	for i := range rows {
		rows[i] = make([]uint32, words)
	}
	_ = clean // Go allocation is always clean; retained for source correspondence.
	return &SimInfo{rows: rows, words: words}, nil
}

func (s *SimInfo) check() error {
	if s == nil {
		return ErrDimension
	}
	if s.closed || s.words < 0 {
		return ErrInvariant
	}
	for _, row := range s.rows {
		if len(row) != s.words {
			return ErrInvariant
		}
	}
	return nil
}
func (s *SimInfo) Entries() int {
	if s == nil || s.closed {
		return 0
	}
	return len(s.rows)
}
func (s *SimInfo) Words() int {
	if s == nil || s.closed {
		return 0
	}
	return s.words
}
func (s *SimInfo) Row(id int) ([]uint32, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	if id < 0 || id >= len(s.rows) {
		return nil, ErrDimension
	}
	return slices.Clone(s.rows[id]), nil
}
func (s *SimInfo) Write(id int, words []uint32) error {
	if err := s.check(); err != nil {
		return err
	}
	if id < 0 || id >= len(s.rows) || len(words) != s.words {
		return ErrDimension
	}
	copy(s.rows[id], words)
	return nil
}
func (s *SimInfo) Close() error {
	if err := s.check(); err != nil {
		return err
	}
	for i := range s.rows {
		s.rows[i] = nil
	}
	s.rows = nil
	s.words = 0
	s.closed = true
	return nil
}

func SimInfoAdd(first, second []uint32) error {
	if len(first) != len(second) {
		return ErrDimension
	}
	for i := range first {
		first[i] |= second[i]
	}
	return nil
}
func SimInfoDetectDiffs(first, second []uint32) ([]int, error) {
	if len(first) != len(second) {
		return nil, ErrDimension
	}
	result := make([]int, 0)
	for w := range first {
		for x := first[w] ^ second[w]; x != 0; x &= x - 1 {
			result = append(result, 32*w+bits.TrailingZeros32(x))
		}
	}
	return result, nil
}
func SimInfoDetectNews(first, second []uint32) ([]int, error) {
	if len(first) != len(second) {
		return nil, ErrDimension
	}
	result := make([]int, 0)
	for w := range first {
		for x := second[w] &^ first[w]; x != 0; x &= x - 1 {
			result = append(result, 32*w+bits.TrailingZeros32(x))
		}
	}
	return result, nil
}
func SimCountOnes(words []uint32) int {
	n := 0
	for _, word := range words {
		n += bits.OnesCount32(word)
	}
	return n
}
func SimCountOnesArray(info *SimInfo) ([]int, error) {
	if err := info.check(); err != nil {
		return nil, err
	}
	r := make([]int, len(info.rows))
	for i := range r {
		r[i] = SimCountOnes(info.rows[i])
	}
	return r, nil
}
func SimSetComplement(words []uint32) {
	for i := range words {
		words[i] = ^words[i]
	}
}
func SimSetConstant(words []uint32, one bool) {
	var v uint32
	if one {
		v = math.MaxUint32
	}
	for i := range words {
		words[i] = v
	}
}
func SimSetRandom(words []uint32, seed uint64) uint64 {
	if seed == 0 {
		seed = 0xabc
	}
	random := simRandom{seed}
	for i := range words {
		words[i] = random.word()
	}
	return random.state
}
func SimInfoEqual(a, b []uint32) (bool, error) {
	if len(a) != len(b) {
		return false, ErrDimension
	}
	return slices.Equal(a, b), nil
}
func SimInfoImplication(a, b []uint32) (bool, error) {
	if len(a) != len(b) {
		return false, ErrDimension
	}
	for i := range a {
		if a[i]&^b[i] != 0 {
			return false, nil
		}
	}
	return true, nil
}
func SimInfoClause(a, b []uint32) (bool, error) {
	if len(a) != len(b) {
		return false, ErrDimension
	}
	for i := range a {
		if ^a[i]&^b[i] != 0 {
			return false, nil
		}
	}
	return true, nil
}

func simRead(row []uint32, complement bool, offset, words int) []uint32 {
	r := row[offset : offset+words]
	if !complement {
		return r
	}
	c := make([]uint32, words)
	for i := range c {
		c[i] = ^r[i]
	}
	return c
}

func simPackedFrame(g *AigManager, info *SimInfo, offset, words int) error {
	if g == nil || info == nil || offset < 0 || words < 0 || offset+words > info.words || len(info.rows) != len(g.objects) {
		return ErrDimension
	}
	for id, object := range g.objects {
		switch object.Type {
		case AigConstant1:
			for w := 0; w < words; w++ {
				info.rows[id][offset+w] = math.MaxUint32
			}
		case AigAND:
			a := simRead(info.rows[object.Fanin0.ID()], object.Fanin0.Complemented(), offset, words)
			b := simRead(info.rows[object.Fanin1.ID()], object.Fanin1.Complemented(), offset, words)
			for w := 0; w < words; w++ {
				info.rows[id][offset+w] = a[w] & b[w]
			}
		case AigBuffer, AigCO:
			a := simRead(info.rows[object.Fanin0.ID()], object.Fanin0.Complemented(), offset, words)
			copy(info.rows[id][offset:offset+words], a)
		}
	}
	return nil
}

func SimulatePackedFrame(graph *AigManager, info *SimInfo, offset, words int) error {
	if graph == nil {
		return ErrDimension
	}
	if err := info.check(); err != nil {
		return err
	}
	if err := graph.Check(); err != nil {
		return err
	}
	return simPackedFrame(graph, info, offset, words)
}

type SimStats struct {
	Inputs, Outputs, SimulationWords                                          int
	FunctionalSupport, StructuralSupport                                      int
	SatRuns, SatRunsSAT, SatRunsUNSAT                                         int
	PairsTotal, PairsSymmetric, PairsStructurallySymmetric, PairsNonSymmetric int
	SimulationTime, SATTime, TotalTime                                        time.Duration
}

func (s SimStats) Text() string {
	return fmt.Sprintf("Total func supps   = %8d.\nTotal struct supps = %8d.\nSat runs SAT       = %8d.\nSat runs UNSAT     = %8d.\nTotal symm         = %8d.\nStructural symm    = %8d.\nTotal non-sym      = %8d.\nTotal var pairs    = %8d.\n", s.FunctionalSupport, s.StructuralSupport, s.SatRunsSAT, s.SatRunsUNSAT, s.PairsSymmetric, s.PairsStructurallySymmetric, s.PairsNonSymmetric, s.PairsTotal)
}

type SimPattern struct {
	Input, Output int
	Data          []uint32
}
type SimManager struct {
	graph                  *AigManager
	options                SimOptions
	lightweight            bool
	sim0, sim1             *SimInfo
	structural, functional [][]uint32
	stats                  SimStats
	closed                 bool
}

func NewSimManager(graph *AigManager, lightweight bool, options SimOptions) (*SimManager, error) {
	options, err := normalizeSimOptions(options)
	if err != nil {
		return nil, err
	}
	if graph == nil {
		return nil, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	if graph.ObjectCount() > options.MaxObjects {
		return nil, ErrLimit
	}
	if simStopped(options) {
		return nil, ErrCancelled
	}
	copyGraph := graph.Clone()
	sim0, err := NewSimInfo(copyGraph.ObjectCount(), 64, false, options)
	if err != nil {
		return nil, err
	}
	m := &SimManager{graph: copyGraph, options: options, lightweight: lightweight, sim0: sim0}
	m.stats.Inputs, m.stats.Outputs, m.stats.SimulationWords = copyGraph.InputCount(), copyGraph.OutputCount(), 64
	if !lightweight {
		m.sim1, err = NewSimInfo(copyGraph.ObjectCount(), 64, false, options)
		if err != nil {
			sim0.Close()
			return nil, err
		}
		m.structural, err = SimStructuralSupport(copyGraph, options)
		if err != nil {
			m.Close()
			return nil, err
		}
		m.stats.StructuralSupport = simSupportSize(m.structural)
	}
	return m, nil
}
func (m *SimManager) check() error {
	if m == nil {
		return ErrDimension
	}
	if m.closed {
		return ErrInvariant
	}
	return nil
}
func (m *SimManager) Stats() (SimStats, error) {
	if err := m.check(); err != nil {
		return SimStats{}, err
	}
	return m.stats, nil
}
func (m *SimManager) StatsText() (string, error) {
	s, e := m.Stats()
	if e != nil {
		return "", e
	}
	return s.Text(), nil
}
func (m *SimManager) AllocPattern() (*SimPattern, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	if m.lightweight {
		return nil, ErrUnsupported
	}
	return &SimPattern{Output: -1, Data: make([]uint32, (m.graph.InputCount()+31)/32)}, nil
}
func (m *SimManager) FreePattern(pattern *SimPattern) error {
	if err := m.check(); err != nil {
		return err
	}
	if pattern == nil {
		return ErrDimension
	}
	clear(pattern.Data)
	pattern.Data = nil
	return nil
}
func (m *SimManager) Close() error {
	if err := m.check(); err != nil {
		return err
	}
	if m.sim0 != nil {
		m.sim0.Close()
	}
	if m.sim1 != nil {
		m.sim1.Close()
	}
	m.graph = nil
	m.structural = nil
	m.functional = nil
	m.closed = true
	return nil
}

func simSupportSize(support [][]uint32) int {
	n := 0
	for _, row := range support {
		n += SimCountOnes(row)
	}
	return n
}

func SimStructuralSupport(graph *AigManager, options SimOptions) ([][]uint32, error) {
	options, err := normalizeSimOptions(options)
	if err != nil {
		return nil, err
	}
	if graph == nil {
		return nil, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	words := (graph.InputCount() + 31) / 32
	if words > options.MaxWords || graph.ObjectCount() > options.MaxObjects {
		return nil, ErrLimit
	}
	all := make([][]uint32, graph.ObjectCount())
	for i := range all {
		all[i] = make([]uint32, words)
	}
	for i, id := range graph.cis {
		all[id][i>>5] |= 1 << uint(i&31)
	}
	for id, obj := range graph.objects {
		if simStopped(options) {
			return nil, ErrCancelled
		}
		switch obj.Type {
		case AigAND:
			for w := 0; w < words; w++ {
				all[id][w] = all[obj.Fanin0.ID()][w] | all[obj.Fanin1.ID()][w]
			}
		case AigBuffer, AigCO:
			copy(all[id], all[obj.Fanin0.ID()])
		}
	}
	result := make([][]uint32, len(graph.cos))
	for i, id := range graph.cos {
		result[i] = slices.Clone(all[id])
	}
	return result, nil
}

func simAddAIGCNF(solver *SatSolver, graph *AigManager, offset int, stop func() bool) error {
	literal := func(signal AigLit) SatLit {
		result := SatLit(int(signal.ID())+offset) << 1
		if signal.Complemented() {
			result |= 1
		}
		return result
	}
	add := func(clause ...SatLit) error {
		ok, err := solver.AddClause(clause)
		if err != nil {
			return err
		}
		if !ok {
			return ErrInvariant
		}
		return nil
	}
	if err := add(SatLit(offset) << 1); err != nil {
		return err
	}
	for id, object := range graph.objects {
		if stop != nil && stop() {
			return ErrCancelled
		}
		output := SatLit(id+offset) << 1
		switch object.Type {
		case AigAND:
			first, second := literal(object.Fanin0), literal(object.Fanin1)
			if err := add(output, first.Not(), second.Not()); err != nil {
				return err
			}
			if err := add(output.Not(), first); err != nil {
				return err
			}
			if err := add(output.Not(), second); err != nil {
				return err
			}
		case AigBuffer, AigCO:
			first := literal(object.Fanin0)
			if err := add(output, first.Not()); err != nil {
				return err
			}
			if err := add(output.Not(), first); err != nil {
				return err
			}
		}
	}
	return nil
}
func simSATPair(graph *AigManager, output, first, second int, symmetry bool, options SimOptions) (bool, []bool, error) {
	if graph == nil {
		return false, nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return false, nil, err
	}
	if output < 0 || output >= graph.OutputCount() || first < 0 || first >= graph.InputCount() || second < 0 || second >= graph.InputCount() || symmetry && first == second {
		return false, nil, ErrDimension
	}
	n := graph.ObjectCount()
	solver, err := NewSatSolverWithLimit(2 * n)
	if err != nil {
		return false, nil, err
	}
	if err = solver.SetVariableCount(2 * n); err != nil {
		return false, nil, err
	}
	if err = simAddAIGCNF(solver, graph, 0, options.Stop); err != nil {
		return false, nil, err
	}
	if err = simAddAIGCNF(solver, graph, n, options.Stop); err != nil {
		return false, nil, err
	}
	inputs := make([]SatLit, len(graph.cis))
	for i, id := range graph.cis {
		inputs[i] = SatLit(id) << 1
	}
	outputs := make([]SatLit, len(graph.cos))
	for i, id := range graph.cos {
		outputs[i] = SatLit(id) << 1
	}
	inconsistent := false
	add := func(clause ...SatLit) error {
		if inconsistent {
			return nil
		}
		ok, e := solver.AddClause(clause)
		if e != nil {
			return e
		}
		if !ok {
			inconsistent = true
		}
		return nil
	}
	for i, l := range inputs {
		r := SatLit(int(l.Var())+n) << 1
		if i == first || symmetry && i == second {
			continue
		}
		if err = add(l.Not(), r); err != nil {
			return false, nil, err
		}
		if err = add(l, r.Not()); err != nil {
			return false, nil, err
		}
	}
	l1 := inputs[first]
	r1 := SatLit(int(l1.Var())+n) << 1
	if symmetry {
		l2 := inputs[second]
		r2 := SatLit(int(l2.Var())+n) << 1
		for _, c := range [][]SatLit{{l1.Not()}, {l2}, {r1}, {r2.Not()}} {
			if err = add(c...); err != nil {
				return false, nil, err
			}
		}
	} else {
		if err = add(l1, r1); err != nil {
			return false, nil, err
		}
		if err = add(l1.Not(), r1.Not()); err != nil {
			return false, nil, err
		}
	}
	o1 := outputs[output]
	o2 := SatLit(int(o1.Var())+n) << 1
	if err = add(o1, o2); err != nil {
		return false, nil, err
	}
	if err = add(o1.Not(), o2.Not()); err != nil {
		return false, nil, err
	}
	if inconsistent {
		return false, nil, nil
	}
	limits := options.SAT
	if limits.Stop == nil {
		limits.Stop = options.Stop
	}
	result, err := solver.Solve(nil, limits)
	if err != nil {
		return false, nil, err
	}
	if result.Status == SatUnknown {
		if result.Reason == SatCancelled {
			return false, nil, ErrCancelled
		}
		return false, nil, ErrLimit
	}
	if result.Status == SatUnsatisfiable {
		return false, nil, nil
	}
	model := make([]bool, len(inputs))
	for i, l := range inputs {
		model[i] = result.Model[l.Var()]
	}
	return true, model, nil
}

type SimSupportResult struct {
	Supports [][]uint32
	Stats    SimStats
}

func SimFunctionalSupport(graph *AigManager, options SimOptions) (SimSupportResult, error) {
	start := time.Now()
	options, err := normalizeSimOptions(options)
	if err != nil {
		return SimSupportResult{}, err
	}
	structural, err := SimStructuralSupport(graph, options)
	if err != nil {
		return SimSupportResult{}, err
	}
	result := make([][]uint32, len(structural))
	for i := range result {
		result[i] = make([]uint32, len(structural[i]))
	}
	stats := SimStats{Inputs: graph.InputCount(), Outputs: graph.OutputCount(), StructuralSupport: simSupportSize(structural)}
	pairs := 0
	for out, row := range structural {
		for input := 0; input < graph.InputCount(); input++ {
			if row[input>>5]&(1<<uint(input&31)) == 0 {
				continue
			}
			pairs++
			if pairs > options.MaxPairs {
				return SimSupportResult{}, ErrLimit
			}
			if simStopped(options) {
				return SimSupportResult{}, ErrCancelled
			}
			sat, _, e := simSATPair(graph, out, input, input, false, options)
			stats.SatRuns++
			if e != nil {
				return SimSupportResult{}, e
			}
			if sat {
				result[out][input>>5] |= 1 << uint(input&31)
				stats.SatRunsSAT++
			} else {
				stats.SatRunsUNSAT++
			}
		}
	}
	stats.FunctionalSupport = simSupportSize(result)
	stats.TotalTime = time.Since(start)
	stats.SATTime = stats.TotalTime
	return SimSupportResult{Supports: result, Stats: stats}, nil
}

type SimInitial uint8

const (
	SimInitZero SimInitial = iota
	SimInitOne
	SimInitUnknown
)

type SimSequentialTrace struct {
	Frames, Words int
	info          *SimInfo
}

func (t *SimSequentialTrace) ObjectWords(id AigID, frame int) ([]uint32, error) {
	if t == nil || t.info == nil || frame < 0 || frame >= t.Frames {
		return nil, ErrDimension
	}
	if uint64(id) >= uint64(t.info.Entries()) {
		return nil, ErrDimension
	}
	return slices.Clone(t.info.rows[id][frame*t.Words : (frame+1)*t.Words]), nil
}
func (t *SimSequentialTrace) Close() error {
	if t == nil || t.info == nil {
		return ErrDimension
	}
	e := t.info.Close()
	t.info = nil
	return e
}

func simSequential(graph *AigManager, frames, words int, model []bool, initial []SimInitial, random bool, options SimOptions) (*SimSequentialTrace, error) {
	options, err := normalizeSimOptions(options)
	if err != nil {
		return nil, err
	}
	if graph == nil || frames < 0 || words <= 0 || len(initial) != graph.RegisterCount() {
		return nil, ErrDimension
	}
	for _, value := range initial {
		if value > SimInitUnknown {
			return nil, ErrDimension
		}
	}
	if frames > options.MaxFrames || words > options.MaxWords || frames != 0 && words > options.MaxWords/frames {
		return nil, ErrLimit
	}
	if !random && len(model) != frames*graph.PrimaryInputCount() {
		return nil, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	info, err := NewSimInfo(graph.ObjectCount(), frames*words, false, options)
	if err != nil {
		return nil, err
	}
	if frames == 0 {
		return &SimSequentialTrace{Frames: frames, Words: words, info: info}, nil
	}
	rng := simRandom{options.Seed}
	pis := graph.cis[:graph.PrimaryInputCount()]
	for i, id := range pis {
		for f := 0; f < frames; f++ {
			for w := 0; w < words; w++ {
				if random {
					info.rows[id][f*words+w] = rng.word()
				} else if model[f*len(pis)+i] {
					info.rows[id][f*words+w] = math.MaxUint32
				}
			}
		}
	}
	regs := graph.cis[graph.PrimaryInputCount():]
	for i, id := range regs {
		for w := 0; w < words; w++ {
			switch initial[i] {
			case SimInitOne:
				info.rows[id][w] = math.MaxUint32
			case SimInitUnknown:
				info.rows[id][w] = rng.word()
			}
		}
	}
	for f := 0; f < frames; f++ {
		if simStopped(options) {
			info.Close()
			return nil, ErrCancelled
		}
		if err = simPackedFrame(graph, info, f*words, words); err != nil {
			info.Close()
			return nil, err
		}
		if f+1 < frames {
			for i, id := range regs {
				co := graph.cos[graph.PrimaryOutputCount()+i]
				copy(info.rows[id][(f+1)*words:(f+2)*words], info.rows[co][f*words:(f+1)*words])
			}
		}
	}
	return &SimSequentialTrace{Frames: frames, Words: words, info: info}, nil
}
func SimulateSequentialModel(graph *AigManager, frames int, model []bool, initial []SimInitial, options SimOptions) (*SimSequentialTrace, error) {
	return simSequential(graph, frames, 1, model, initial, false, options)
}
func SimulateSequentialRandom(graph *AigManager, frames, words int, initial []SimInitial, options SimOptions) (*SimSequentialTrace, error) {
	return simSequential(graph, frames, words, nil, initial, true, options)
}

func SimSwitching(words []uint32) float32 {
	if len(words) == 0 {
		return 0
	}
	ones := float64(SimCountOnes(words))
	total := float64(32 * len(words))
	return float32(2 * ones / total * (total - ones) / total)
}
func SimNetworkSwitching(graph *AigManager, patterns int, options SimOptions) ([]float32, error) {
	options, err := normalizeSimOptions(options)
	if err != nil {
		return nil, err
	}
	if graph == nil || patterns < 0 {
		return nil, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	words := (patterns + 31) / 32
	if words > options.MaxWords {
		return nil, ErrLimit
	}
	info, err := NewSimInfo(graph.ObjectCount(), words, false, options)
	if err != nil {
		return nil, err
	}
	defer info.Close()
	rng := simRandom{options.Seed}
	for _, id := range graph.cis {
		for w := range info.rows[id] {
			info.rows[id][w] = rng.word()
		}
	}
	if err = simPackedFrame(graph, info, 0, words); err != nil {
		return nil, err
	}
	r := make([]float32, graph.ObjectCount())
	for id, obj := range graph.objects {
		if obj.Type == AigCI || obj.Type == AigAND || obj.Type == AigBuffer {
			r[id] = SimSwitching(info.rows[id])
		}
	}
	return r, nil
}

type SimBitMatrix struct {
	size int
	rows [][]uint32
}

func NewSimBitMatrix(size int) (*SimBitMatrix, error) {
	if size < 0 {
		return nil, ErrDimension
	}
	w := (size + 31) / 32
	r := make([][]uint32, size)
	for i := range r {
		r[i] = make([]uint32, w)
	}
	return &SimBitMatrix{size: size, rows: r}, nil
}
func (m *SimBitMatrix) Has(a, b int) (bool, error) {
	if m == nil || a < 0 || b < 0 || a >= m.size || b >= m.size {
		return false, ErrDimension
	}
	return m.rows[a][b>>5]&(1<<uint(b&31)) != 0, nil
}
func (m *SimBitMatrix) set(a, b int) {
	m.rows[a][b>>5] |= 1 << uint(b&31)
	m.rows[b][a>>5] |= 1 << uint(a&31)
}
func (m *SimBitMatrix) Rows() [][]uint32 {
	r := make([][]uint32, len(m.rows))
	for i := range r {
		r[i] = slices.Clone(m.rows[i])
	}
	return r
}

type SimSymmetryResult struct {
	Symmetric, NonSymmetric []*SimBitMatrix
	Supports                [][]uint32
	Stats                   SimStats
}

func simStructurallyEqualAfterSwap(graph *AigManager, output, first, second int) bool {
	ciIndex := make([]int, len(graph.objects))
	for i := range ciIndex {
		ciIndex[i] = -1
	}
	for i, id := range graph.cis {
		ciIndex[id] = i
	}
	canonical := func(swapped bool) string {
		memo := make(map[AigLit]string)
		var signal func(AigLit) string
		var conjunction func(AigID) string
		conjunction = func(id AigID) string {
			factors := make([]string, 0, 4)
			var collect func(AigLit)
			collect = func(lit AigLit) {
				object := graph.objects[lit.ID()]
				if !lit.Complemented() && object.Type == AigAND {
					collect(object.Fanin0)
					collect(object.Fanin1)
					return
				}
				factors = append(factors, signal(lit))
			}
			object := graph.objects[id]
			collect(object.Fanin0)
			collect(object.Fanin1)
			sort.Strings(factors)
			return "(" + strings.Join(factors, "&") + ")"
		}
		signal = func(lit AigLit) string {
			if cached, ok := memo[lit]; ok {
				return cached
			}
			object := graph.objects[lit.ID()]
			var text string
			switch object.Type {
			case AigConstant1:
				if lit.Complemented() {
					text = "0"
				} else {
					text = "1"
				}
			case AigCI:
				index := ciIndex[lit.ID()]
				if swapped {
					if index == first {
						index = second
					} else if index == second {
						index = first
					}
				}
				text = fmt.Sprintf("x%d", index)
				if lit.Complemented() {
					text = "!" + text
				}
			case AigBuffer, AigCO:
				text = signal(object.Fanin0.NotIf(lit.Complemented()))
			case AigAND:
				text = conjunction(lit.ID())
				if lit.Complemented() {
					text = "!" + text
				}
			default:
				text = "?"
			}
			memo[lit] = text
			return text
		}
		return signal(graph.objects[graph.cos[output]].Fanin0)
	}
	return canonical(false) == canonical(true)
}

func SimStructuralSymmetries(graph *AigManager, supports [][]uint32, options SimOptions) ([]*SimBitMatrix, int, error) {
	options, err := normalizeSimOptions(options)
	if err != nil {
		return nil, 0, err
	}
	if graph == nil || len(supports) != graph.OutputCount() {
		return nil, 0, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return nil, 0, err
	}
	if graph.ObjectCount() > options.MaxObjects {
		return nil, 0, ErrLimit
	}
	result := make([]*SimBitMatrix, graph.OutputCount())
	count, pairs := 0, 0
	for output, support := range supports {
		if len(support) != (graph.InputCount()+31)/32 {
			return nil, 0, ErrDimension
		}
		result[output], _ = NewSimBitMatrix(graph.InputCount())
		for first := 0; first < graph.InputCount(); first++ {
			if support[first>>5]&(1<<uint(first&31)) == 0 {
				continue
			}
			for second := first + 1; second < graph.InputCount(); second++ {
				if support[second>>5]&(1<<uint(second&31)) == 0 {
					continue
				}
				pairs++
				if pairs > options.MaxPairs {
					return nil, 0, ErrLimit
				}
				if simStopped(options) {
					return nil, 0, ErrCancelled
				}
				if simStructurallyEqualAfterSwap(graph, output, first, second) {
					result[output].set(first, second)
					count++
				}
			}
		}
	}
	return result, count, nil
}

func SimComputeTwoVariableSymmetries(graph *AigManager, options SimOptions) (SimSymmetryResult, error) {
	start := time.Now()
	options, err := normalizeSimOptions(options)
	if err != nil {
		return SimSymmetryResult{}, err
	}
	support, err := SimFunctionalSupport(graph, options)
	if err != nil {
		return SimSymmetryResult{}, err
	}
	r := SimSymmetryResult{Supports: support.Supports, Stats: support.Stats}
	r.Symmetric = make([]*SimBitMatrix, graph.OutputCount())
	r.NonSymmetric = make([]*SimBitMatrix, graph.OutputCount())
	structural, structuralCount, err := SimStructuralSymmetries(graph, r.Supports, options)
	if err != nil {
		return SimSymmetryResult{}, err
	}
	r.Stats.PairsSymmetric = structuralCount
	r.Stats.PairsStructurallySymmetric = structuralCount
	pairs := 0
	for out, row := range r.Supports {
		r.Symmetric[out] = structural[out]
		r.NonSymmetric[out], _ = NewSimBitMatrix(graph.InputCount())
		for a := 0; a < graph.InputCount(); a++ {
			if row[a>>5]&(1<<uint(a&31)) == 0 {
				continue
			}
			for b := a + 1; b < graph.InputCount(); b++ {
				if row[b>>5]&(1<<uint(b&31)) == 0 {
					continue
				}
				pairs++
				if pairs > options.MaxPairs {
					return SimSymmetryResult{}, ErrLimit
				}
				known, _ := r.Symmetric[out].Has(a, b)
				if known {
					continue
				}
				sat, _, e := simSATPair(graph, out, a, b, true, options)
				r.Stats.SatRuns++
				if e != nil {
					return SimSymmetryResult{}, e
				}
				if sat {
					r.NonSymmetric[out].set(a, b)
					r.Stats.PairsNonSymmetric++
					r.Stats.SatRunsSAT++
				} else {
					r.Symmetric[out].set(a, b)
					r.Stats.PairsSymmetric++
					r.Stats.SatRunsUNSAT++
				}
			}
		}
	}
	r.Stats.PairsTotal = pairs
	r.Stats.TotalTime = time.Since(start)
	r.Stats.SATTime = r.Stats.TotalTime
	return r, nil
}

// SimSymmetryCounterexample returns a SAT witness for a non-symmetric pair.
func SimSymmetryCounterexample(graph *AigManager, output, first, second int, options SimOptions) ([]bool, bool, error) {
	options, e := normalizeSimOptions(options)
	if e != nil {
		return nil, false, e
	}
	sat, model, e := simSATPair(graph, output, first, second, true, options)
	return model, sat, e
}

func SimCountAllPairs(support [][]uint32, inputs int) ([]int, int, error) {
	if inputs < 0 {
		return nil, 0, ErrDimension
	}
	counts := make([]int, len(support))
	total := 0
	for i, row := range support {
		if len(row) != (inputs+31)/32 {
			return nil, 0, ErrDimension
		}
		n := SimCountOnes(row)
		counts[i] = n * (n - 1) / 2
		total += counts[i]
	}
	return counts, total, nil
}
func SimMatricesDisjoint(first, second []*SimBitMatrix) (bool, error) {
	if len(first) != len(second) {
		return false, ErrDimension
	}
	for i := range first {
		if first[i] == nil || second[i] == nil || first[i].size != second[i].size {
			return false, ErrDimension
		}
		for r := range first[i].rows {
			for w := range first[i].rows[r] {
				if first[i].rows[r][w]&second[i].rows[r][w] != 0 {
					return false, nil
				}
			}
		}
	}
	return true, nil
}

func SimSupportString(support [][]uint32, inputs int) (string, error) {
	if inputs < 0 {
		return "", ErrDimension
	}
	var b strings.Builder
	for out, row := range support {
		if len(row) != (inputs+31)/32 {
			return "", ErrDimension
		}
		fmt.Fprintf(&b, "%d:", out)
		for i := 0; i < inputs; i++ {
			if row[i>>5]&(1<<uint(i&31)) != 0 {
				fmt.Fprintf(&b, " %d", i)
			}
		}
		b.WriteByte('\n')
	}
	return b.String(), nil
}
