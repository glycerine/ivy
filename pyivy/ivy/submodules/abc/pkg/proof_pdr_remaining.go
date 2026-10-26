package abc

// Standalone Go semantic replacements for portable behavior in
// src/proof/pdr/pdrClass.c, pdrIncr.c, pdrTsim.c, pdrTsim2.c and pdrTsim3.c.
// Solver/GIA construction in pdrCnf.c and command-network entry points in
// pdrIncr.c deliberately remain unsupported. The three ternary simulators'
// priority heuristics are replaced by exhaustive truth-table generalization.

import "slices"

// PdrRemainingUnit records the disposition of each build unit which was not
// represented by an earlier Go source file. It is immutable caller-owned data.
type PdrRemainingUnit struct {
	Source, Disposition string
	Portable            bool
}

// PdrRemainingUnits is the auditable module.make inventory for this lane.
func PdrRemainingUnits() []PdrRemainingUnit {
	return []PdrRemainingUnit{
		{"pdr.h", "portable data contracts use concrete Go types; C manager and command APIs are excluded", true},
		{"pdrClass.c", "complete caller-supplied observations produce and validate a deterministic equivalence partition; AIG rehash iteration is excluded", true},
		{"pdrCnf.c", "native CNF/GIA variable mapping and solver construction are unsupported here", false},
		{"pdrIncr.c", "owned transactional clause archive; SAT loop, abstraction, bridge, network and command entry points are excluded", true},
		{"pdrTsim.c", "exhaustively checked truth-table predecessor generalization replaces the AIG ternary heuristic", true},
		{"pdrTsim2.c", "exhaustively checked truth-table predecessor generalization replaces the priority/GIA ternary heuristic", true},
		{"pdrTsim3.c", "exhaustively checked truth-table predecessor generalization replaces the abstraction/GIA ternary heuristic", true},
		{"module.make", "build list surveyed; previously ported units remain in their concrete Go files", true},
	}
}

type PdrRemainingLimits struct {
	MaxNodes, MaxStates, MaxTransitions, MaxFrames, MaxClauses, MaxCubes int
	MaxBytes, MaxWork, MaxDepth                                          uint64
	Stop                                                                 func() bool
}

type PdrRemainingStats struct {
	Nodes, States, Transitions, Frames, Clauses, Cubes int
	Bytes, Work, Depth                                 uint64
}

func (l PdrRemainingLimits) check(s PdrRemainingStats) error {
	if l.Stop != nil && l.Stop() {
		return ErrCancelled
	}
	if l.MaxNodes < 0 || l.MaxStates < 0 || l.MaxTransitions < 0 || l.MaxFrames < 0 || l.MaxClauses < 0 || l.MaxCubes < 0 {
		return ErrDimension
	}
	if (l.MaxNodes > 0 && s.Nodes > l.MaxNodes) || (l.MaxStates > 0 && s.States > l.MaxStates) || (l.MaxTransitions > 0 && s.Transitions > l.MaxTransitions) ||
		(l.MaxFrames > 0 && s.Frames > l.MaxFrames) || (l.MaxClauses > 0 && s.Clauses > l.MaxClauses) ||
		(l.MaxCubes > 0 && s.Cubes > l.MaxCubes) || (l.MaxBytes > 0 && s.Bytes > l.MaxBytes) ||
		(l.MaxWork > 0 && s.Work > l.MaxWork) || (l.MaxDepth > 0 && s.Depth > l.MaxDepth) {
		return ErrLimit
	}
	return nil
}

// PdrEquivalencePartition is a deterministic partition. Representative is
// the least member and Classes are ordered by representative.
type PdrEquivalencePartition struct {
	representative []int
	classes        [][]int
}

// NewPdrEquivalencePartition partitions columns with identical complete
// Boolean observations. This is the portable, data-set based part of
// Pdr_ManEquivClasses; proving completeness of observations remains a caller
// responsibility instead of silently invoking a native SAT/GIA oracle.
func NewPdrEquivalencePartition(observations [][]bool, columns int, limits PdrRemainingLimits) (*PdrEquivalencePartition, PdrRemainingStats, error) {
	stats := PdrRemainingStats{Nodes: columns, States: len(observations)}
	if columns < 0 {
		return nil, stats, ErrDimension
	}
	rows, cols := uint64(len(observations)), uint64(columns)
	if cols != 0 && (cols > ^uint64(0)/40 || rows > ^uint64(0)/cols-40) {
		return nil, stats, ErrLimit
	}
	// Account conservatively for observations, representatives, class members,
	// and one slice header per possible singleton class before allocating.
	stats.Bytes = cols * (rows + 40)
	if err := limits.check(stats); err != nil {
		return nil, stats, err
	}
	for _, row := range observations {
		if len(row) != columns {
			return nil, stats, ErrDimension
		}
	}
	p := &PdrEquivalencePartition{representative: make([]int, columns)}
	for column := 0; column < columns; column++ {
		if limits.Stop != nil && limits.Stop() {
			return nil, stats, ErrCancelled
		}
		rep := column
		for candidate := 0; candidate < column; candidate++ {
			equal := true
			for _, row := range observations {
				stats.Work++
				if err := limits.check(stats); err != nil {
					return nil, stats, err
				}
				if row[column] != row[candidate] {
					equal = false
					break
				}
			}
			if equal {
				rep = candidate
				break
			}
		}
		p.representative[column] = rep
	}
	for rep := 0; rep < columns; rep++ {
		if p.representative[rep] != rep {
			continue
		}
		class := make([]int, 0, 1)
		for member, representative := range p.representative {
			if representative == rep {
				class = append(class, member)
			}
		}
		p.classes = append(p.classes, class)
	}
	return p, stats, nil
}

func (p *PdrEquivalencePartition) Representatives() []int {
	if p == nil {
		return nil
	}
	return slices.Clone(p.representative)
}
func (p *PdrEquivalencePartition) Classes() [][]int {
	if p == nil {
		return nil
	}
	r := make([][]int, len(p.classes))
	for i := range p.classes {
		r[i] = slices.Clone(p.classes[i])
	}
	return r
}
func (p *PdrEquivalencePartition) Check() error {
	if p == nil {
		return ErrDimension
	}
	seen := make([]bool, len(p.representative))
	for _, class := range p.classes {
		if len(class) == 0 {
			return ErrInvariant
		}
		rep := class[0]
		for i, member := range class {
			if member < 0 || member >= len(seen) || seen[member] || (i > 0 && class[i-1] >= member) || p.representative[member] != rep {
				return ErrInvariant
			}
			seen[member] = true
		}
	}
	for _, yes := range seen {
		if !yes {
			return ErrInvariant
		}
	}
	return nil
}

// CheckObservations validates that this partition is exactly the equality
// partition induced by complete caller-supplied Boolean observations.
func (p *PdrEquivalencePartition) CheckObservations(observations [][]bool) error {
	if err := p.Check(); err != nil {
		return err
	}
	for _, row := range observations {
		if len(row) != len(p.representative) {
			return ErrDimension
		}
	}
	for left := range p.representative {
		for right := 0; right < left; right++ {
			equal := true
			for _, row := range observations {
				if row[left] != row[right] {
					equal = false
					break
				}
			}
			if equal != (p.representative[left] == p.representative[right]) {
				return ErrInvariant
			}
		}
	}
	return nil
}

// PdrClauseArchive owns frame clauses used by save/restore/rebuild operations
// in incremental PDR. Mutation is copy-on-success, so cancellation or a limit
// never leaves a partially updated archive.
type PdrClauseArchive struct {
	frames [][]*PdrCube
	stats  PdrRemainingStats
}

func NewPdrClauseArchive(frames int, limits PdrRemainingLimits) (*PdrClauseArchive, error) {
	stats := PdrRemainingStats{Frames: frames}
	if frames < 0 {
		return nil, ErrDimension
	}
	if uint64(frames) > ^uint64(0)/24 {
		return nil, ErrLimit
	}
	stats.Bytes = uint64(frames) * 24
	if err := limits.check(stats); err != nil {
		return nil, err
	}
	return &PdrClauseArchive{frames: make([][]*PdrCube, frames), stats: stats}, nil
}
func (a *PdrClauseArchive) Stats() PdrRemainingStats {
	if a == nil {
		return PdrRemainingStats{}
	}
	return a.stats
}
func (a *PdrClauseArchive) Frames() [][]*PdrCube {
	if a == nil {
		return nil
	}
	r := make([][]*PdrCube, len(a.frames))
	for f := range a.frames {
		for _, c := range a.frames[f] {
			r[f] = append(r[f], c.Clone())
		}
	}
	return r
}

// Save returns an owned archive snapshot. DropLast mirrors the incremental C
// helper: there is no useful snapshot when only frame zero would remain.
func (a *PdrClauseArchive) Save(dropLast bool, limits PdrRemainingLimits) ([][]*PdrCube, error) {
	if a == nil {
		return nil, ErrDimension
	}
	if err := limits.check(a.stats); err != nil {
		return nil, err
	}
	end := len(a.frames)
	if end == 1 || (dropLast && end == 2) {
		return nil, nil
	}
	if dropLast && end > 0 {
		end--
	}
	stats := a.stats
	for frame := 0; frame < end; frame++ {
		for range a.frames[frame] {
			stats.Work++
			if err := limits.check(stats); err != nil {
				return nil, err
			}
		}
	}
	return pdrCloneFrames(a.frames[:end]), nil
}
func pdrArchiveStats(frames [][]*PdrCube, prior PdrRemainingStats) PdrRemainingStats {
	s := prior
	s.Frames = len(frames)
	s.Clauses, s.Cubes = 0, 0
	s.Bytes = uint64(len(frames)) * 24
	for _, frame := range frames {
		for _, c := range frame {
			s.Clauses++
			s.Cubes++
			s.Bytes += uint64(64 + 4*c.LiteralCount())
		}
	}
	return s
}
func pdrCloneFrames(frames [][]*PdrCube) [][]*PdrCube {
	r := make([][]*PdrCube, len(frames))
	for f := range frames {
		for _, c := range frames[f] {
			r[f] = append(r[f], c.Clone())
		}
	}
	return r
}

// Add inserts a blocked cube, discarding contained clauses and rejecting a
// redundant superset, as Pdr_SetContains-based frame maintenance does.
func (a *PdrClauseArchive) Add(frame int, cube *PdrCube, limits PdrRemainingLimits) error {
	if a == nil || cube == nil || frame < 0 || frame >= len(a.frames) {
		return ErrDimension
	}
	if err := cube.Check(); err != nil {
		return err
	}
	if len(cube.inputs) != 0 {
		return ErrDimension
	}
	if err := limits.check(a.stats); err != nil {
		return err
	}
	stats := a.stats
	for _, old := range a.frames[frame] {
		stats.Work++
		if err := limits.check(stats); err != nil {
			return err
		}
		if cube.Contains(old) {
			return nil
		}
	}
	for _, old := range a.frames[frame] {
		stats.Work++
		if err := limits.check(stats); err != nil {
			return err
		}
		if old.Contains(cube) {
			stats.Clauses--
			stats.Cubes--
			stats.Bytes -= uint64(64 + 4*old.LiteralCount())
		}
	}
	stats.Clauses++
	stats.Cubes++
	stats.Bytes += uint64(64 + 4*cube.LiteralCount())
	stats.Work++
	if err := limits.check(stats); err != nil {
		return err
	}
	next := pdrCloneFrames(a.frames)
	kept := next[frame][:0]
	for _, old := range next[frame] {
		if !old.Contains(cube) {
			kept = append(kept, old)
		}
	}
	next[frame] = append(kept, cube.Clone())
	slices.SortFunc(next[frame], func(x, y *PdrCube) int { return x.Compare(y) })
	a.frames, a.stats = next, stats
	return nil
}

// Push copies a clause into every later frame through destination. The entire
// operation commits or rolls back as one transaction.
func (a *PdrClauseArchive) Push(from, destination int, cube *PdrCube, limits PdrRemainingLimits) error {
	if a == nil || cube == nil || from < 0 || destination < from || destination >= len(a.frames) {
		return ErrDimension
	}
	if err := cube.Check(); err != nil {
		return err
	}
	if len(cube.inputs) != 0 {
		return ErrDimension
	}
	if err := limits.check(a.stats); err != nil {
		return err
	}
	copyArchive := &PdrClauseArchive{frames: pdrCloneFrames(a.frames), stats: a.stats}
	for f := from + 1; f <= destination; f++ {
		if err := copyArchive.Add(f, cube, limits); err != nil {
			return err
		}
	}
	a.frames, a.stats = copyArchive.frames, copyArchive.stats
	return nil
}
func (a *PdrClauseArchive) Restore(snapshot [][]*PdrCube, limits PdrRemainingLimits) error {
	if a == nil {
		return ErrDimension
	}
	stats := a.stats
	for _, frame := range snapshot {
		for i, c := range frame {
			stats.Work++
			if err := limits.check(stats); err != nil {
				return err
			}
			if err := c.Check(); err != nil {
				return err
			}
			if len(c.inputs) != 0 {
				return ErrDimension
			}
			if i > 0 && frame[i-1].Compare(c) >= 0 {
				return ErrInvariant
			}
			for j := 0; j < i; j++ {
				stats.Work++
				if err := limits.check(stats); err != nil {
					return err
				}
				if frame[j].Contains(c) || c.Contains(frame[j]) {
					return ErrInvariant
				}
			}
		}
	}
	stats = pdrArchiveStats(snapshot, stats)
	if err := limits.check(stats); err != nil {
		return err
	}
	next := pdrCloneFrames(snapshot)
	a.frames, a.stats = next, stats
	return nil
}
func (a *PdrClauseArchive) Check() error {
	if a == nil {
		return ErrDimension
	}
	for _, frame := range a.frames {
		for i, c := range frame {
			if err := c.Check(); err != nil {
				return err
			}
			if len(c.inputs) != 0 {
				return ErrInvariant
			}
			if i > 0 && frame[i-1].Compare(c) >= 0 {
				return ErrInvariant
			}
			for j := 0; j < i; j++ {
				if frame[j].Contains(c) || c.Contains(frame[j]) {
					return ErrInvariant
				}
			}
		}
	}
	return nil
}

// PdrExplicitTransitionSystem is a complete deterministic truth-table model.
// It makes generalization and certificate replay independent of the C code.
type PdrExplicitTransitionSystem struct {
	StateBits, InputBits int
	Initial              []uint64
	Next                 []uint64
	Bad                  []bool
}

func (m *PdrExplicitTransitionSystem) dimensions() (states, inputs int, err error) {
	if m == nil || m.StateBits < 0 || m.InputBits < 0 || m.StateBits > 20 || m.InputBits > 20 || m.StateBits+m.InputBits > 24 {
		return 0, 0, ErrDimension
	}
	states, inputs = 1<<m.StateBits, 1<<m.InputBits
	if len(m.Next) != states*inputs || len(m.Bad) != states*inputs {
		return 0, 0, ErrDimension
	}
	for _, s := range m.Initial {
		if s >= uint64(states) {
			return 0, 0, ErrDimension
		}
	}
	for _, s := range m.Next {
		if s >= uint64(states) {
			return 0, 0, ErrDimension
		}
	}
	return states, inputs, nil
}
func pdrUintCube(value uint64, bits int, input bool) (*PdrCube, error) {
	lits := make([]PdrLit, bits)
	for bit := 0; bit < bits; bit++ {
		l, e := MakePdrLit(uint32(bit), value&(1<<bit) == 0)
		if e != nil {
			return nil, e
		}
		lits[bit] = l
	}
	if input {
		return NewPdrCube(nil, lits)
	}
	return NewPdrCube(lits, nil)
}

// GeneralizePredecessor greedily removes state literals then input literals
// while every completion still reaches target. Exhaustive enumeration makes
// every removal decision exact; it does not claim the C heuristics' same cube.
func (m *PdrExplicitTransitionSystem) GeneralizePredecessor(state, input uint64, target *PdrCube, limits PdrRemainingLimits) (*PdrCube, PdrRemainingStats, error) {
	states, inputs, err := m.dimensions()
	if err != nil {
		return nil, PdrRemainingStats{}, err
	}
	stats := PdrRemainingStats{Nodes: states, States: states, Transitions: states * inputs, Clauses: 1, Cubes: 1, Depth: 1}
	if target == nil {
		return nil, stats, ErrDimension
	}
	if state >= uint64(states) || input >= uint64(inputs) {
		return nil, stats, ErrDimension
	}
	if err = target.Check(); err != nil {
		return nil, stats, err
	}
	if len(target.inputs) != 0 {
		return nil, stats, ErrDimension
	}
	for _, l := range target.state {
		if int(l.Index()) >= m.StateBits {
			return nil, stats, ErrDimension
		}
	}
	sc, _ := pdrUintCube(state, m.StateBits, false)
	ic, _ := pdrUintCube(input, m.InputBits, true)
	current, _ := NewPdrCube(sc.state, ic.inputs)
	stats.Bytes = uint64(64 + 4*(target.LiteralCount()+current.LiteralCount()))
	if err := limits.check(stats); err != nil {
		return nil, stats, err
	}
	forces := func(candidate *PdrCube) (bool, error) {
		for s := 0; s < states; s++ {
			sm, _ := candidate.Matches(pdrBits(uint64(s), m.StateBits))
			if !sm {
				continue
			}
			for in := 0; in < inputs; in++ {
				if !pdrInputMatches(candidate.inputs, uint64(in)) {
					continue
				}
				stats.Work++
				if err := limits.check(stats); err != nil {
					return false, err
				}
				ok, e := target.Matches(pdrBits(m.Next[s*inputs+in], m.StateBits))
				if e != nil || !ok {
					return false, e
				}
			}
		}
		return true, nil
	}
	ok, err := forces(current)
	if err != nil {
		return nil, stats, err
	}
	if !ok {
		return nil, stats, ErrInvariant
	}
	for i := 0; i < len(current.state); {
		candidate, _ := current.Without(i)
		stats.Cubes++
		stats.Bytes += uint64(32 + 4*candidate.LiteralCount())
		if ok, e := forces(candidate); e != nil {
			return nil, stats, e
		} else if ok {
			current = candidate
			continue
		}
		i++
	}
	for i := 0; i < len(current.inputs); {
		inputsCopy := append(slices.Clone(current.inputs[:i]), current.inputs[i+1:]...)
		candidate, _ := NewPdrCube(current.state, inputsCopy)
		stats.Cubes++
		stats.Bytes += uint64(32 + 4*candidate.LiteralCount())
		if ok, e := forces(candidate); e != nil {
			return nil, stats, e
		} else if ok {
			current = candidate
			continue
		}
		i++
	}
	if err := limits.check(stats); err != nil {
		return nil, stats, err
	}
	return current.Clone(), stats, nil
}
func pdrBits(value uint64, count int) []bool {
	r := make([]bool, count)
	for i := range r {
		r[i] = value&(1<<i) != 0
	}
	return r
}
func pdrInputMatches(lits []PdrLit, input uint64) bool {
	for _, l := range lits {
		if (input&(1<<l.Index()) != 0) == l.Negative() {
			return false
		}
	}
	return true
}

type PdrExplicitInvariantPhase uint8

const (
	PdrExplicitComplete PdrExplicitInvariantPhase = iota
	PdrExplicitInitiation
	PdrExplicitInduction
	PdrExplicitSafety
)

type PdrExplicitVerification struct {
	Valid bool
	Phase PdrExplicitInvariantPhase
	Stats PdrRemainingStats
}

// VerifyInvariant exhaustively checks initiation, induction and safety in that
// order. A success is therefore an actual inductive certificate.
func (m *PdrExplicitTransitionSystem) VerifyInvariant(invariant *PdrInvariant, limits PdrRemainingLimits) (PdrExplicitVerification, error) {
	states, inputs, err := m.dimensions()
	result := PdrExplicitVerification{Stats: PdrRemainingStats{Nodes: states, States: states, Transitions: states * inputs}}
	if err != nil || invariant == nil {
		if err == nil {
			err = ErrDimension
		}
		return result, err
	}
	result.Stats.Clauses, result.Stats.Cubes = len(invariant.clauses), len(invariant.clauses)
	for _, cube := range invariant.clauses {
		if err := cube.Check(); err != nil {
			return result, err
		}
		if len(cube.inputs) != 0 {
			return result, ErrDimension
		}
		for _, literal := range cube.state {
			if int(literal.Index()) >= m.StateBits {
				return result, ErrDimension
			}
		}
		result.Stats.Bytes += uint64(32 + 4*cube.LiteralCount())
	}
	if err := limits.check(result.Stats); err != nil {
		return result, err
	}
	holds := func(s uint64) (bool, error) {
		result.Stats.Work++
		if e := limits.check(result.Stats); e != nil {
			return false, e
		}
		return invariant.Holds(pdrBits(s, m.StateBits))
	}
	for _, s := range m.Initial {
		ok, e := holds(s)
		if e != nil {
			return result, e
		}
		if !ok {
			result.Phase = PdrExplicitInitiation
			return result, nil
		}
	}
	for s := 0; s < states; s++ {
		ok, e := holds(uint64(s))
		if e != nil {
			return result, e
		}
		if !ok {
			continue
		}
		for in := 0; in < inputs; in++ {
			nextOK, e := holds(m.Next[s*inputs+in])
			if e != nil {
				return result, e
			}
			if !nextOK {
				result.Phase = PdrExplicitInduction
				return result, nil
			}
		}
	}
	for s := 0; s < states; s++ {
		ok, e := holds(uint64(s))
		if e != nil {
			return result, e
		}
		if !ok {
			continue
		}
		for in := 0; in < inputs; in++ {
			result.Stats.Work++
			if e := limits.check(result.Stats); e != nil {
				return result, e
			}
			if m.Bad[s*inputs+in] {
				result.Phase = PdrExplicitSafety
				return result, nil
			}
		}
	}
	result.Valid, result.Phase = true, PdrExplicitComplete
	return result, nil
}

// VerifyWitness replays an owned state/input trace and requires the last step
// to assert Bad. It rejects tampering rather than trusting cached states.
func (m *PdrExplicitTransitionSystem) VerifyWitness(statesTrace, inputsTrace []uint64, limits PdrRemainingLimits) (bool, PdrRemainingStats, error) {
	states, inputs, err := m.dimensions()
	stats := PdrRemainingStats{Nodes: states, States: states, Transitions: states * inputs, Depth: uint64(len(inputsTrace))}
	if err != nil {
		return false, stats, err
	}
	if len(inputsTrace) == 0 || len(statesTrace) != len(inputsTrace)+1 {
		return false, stats, ErrDimension
	}
	stats.Bytes = uint64(8 * (len(statesTrace) + len(inputsTrace)))
	if err := limits.check(stats); err != nil {
		return false, stats, err
	}
	for _, state := range statesTrace {
		if state >= uint64(states) {
			return false, stats, ErrDimension
		}
	}
	for _, input := range inputsTrace {
		if input >= uint64(inputs) {
			return false, stats, ErrDimension
		}
	}
	if !slices.Contains(m.Initial, statesTrace[0]) {
		return false, stats, nil
	}
	for frame, in := range inputsTrace {
		index := int(statesTrace[frame])*inputs + int(in)
		stats.Work++
		if err := limits.check(stats); err != nil {
			return false, stats, err
		}
		if m.Next[index] != statesTrace[frame+1] {
			return false, stats, nil
		}
		if frame+1 == len(inputsTrace) && !m.Bad[index] {
			return false, stats, nil
		}
	}
	return true, stats, nil
}
