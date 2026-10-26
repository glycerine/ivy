package abc

// Derived from ABC src/proof/fra. The source cluster is represented by the
// source inventory below; specialized native partition, cut, file, and timing
// flows remain explicitly unsupported rather than being silently approximated.

import (
	"errors"
	"fmt"
	"slices"
	"sort"
)

var ErrFRAClosed = errors.New("FRA manager is closed")

type FRASourceUnit struct {
	File        string
	Disposition string
}

// FRASourceInventory returns every build unit in src/proof/fra/module.make and
// the public header. The returned slice is owned by the caller.
func FRASourceInventory() []FRASourceUnit {
	return []FRASourceUnit{
		{"fra.h", "partial: concrete manager, classes, simulation, SAT, sequential witnesses, limits"},
		{"fra_.c", "validated: empty non-build placeholder with no definitions"},
		{"fraBmc.c", "partial: bounded explicit-state frames and replayable witnesses"},
		{"fraCec.c", "partial: owned SAT-verified node equivalence"},
		{"fraClass.c", "partial: deterministic phase-aware signature classes"},
		{"fraClau.c", "unsupported: native clause minimization workflow"},
		{"fraClaus.c", "unsupported: cut-based inductive-clause mining"},
		{"fraCnf.c", "partial: concrete FRAIG SAT encoding is reused"},
		{"fraCore.c", "partial: verified atomic functional sweep"},
		{"fraHot.c", "unsupported: one-hot EXDC construction"},
		{"fraImp.c", "unsupported: implication mining"},
		{"fraInd.c", "partial: bounded sequential equivalence, not unbounded induction"},
		{"fraIndVer.c", "unsupported: external invariant-clause verification"},
		{"fraLcr.c", "partial: sequential phase/latch correspondence checks"},
		{"fraMan.c", "partial: owned lifecycle, accounting, cancellation, rollback"},
		{"fraPart.c", "unsupported: partitioned native FRAIG experiments"},
		{"fraSat.c", "partial: SAT-verified equivalence with replayed models"},
		{"fraSec.c", "partial: bounded sequential comparison only"},
		{"fraSim.c", "partial: deterministic combinational/sequential simulation"},
		{"module.make", "validated: all 17 compiled units inventoried"},
	}
}

const (
	fraNodeBytes    = 40
	fraPatternBytes = 32
	fraClassBytes   = 48
)

type FRAOptions struct {
	MaxNodes, MaxClauses, MaxFrames, MaxPatterns, MaxClasses int
	MaxBytes, MaxWork, MaxDepth                              int
	Stop                                                     func() bool
}

func (o FRAOptions) normalized() (FRAOptions, error) {
	if o.MaxNodes < 0 || o.MaxClauses < 0 || o.MaxFrames < 0 || o.MaxPatterns < 0 || o.MaxClasses < 0 || o.MaxBytes < 0 || o.MaxWork < 0 || o.MaxDepth < 0 {
		return o, ErrDimension
	}
	if o.MaxNodes == 0 {
		o.MaxNodes = 1 << 20
	}
	if o.MaxClauses == 0 {
		o.MaxClauses = 1 << 22
	}
	if o.MaxFrames == 0 {
		o.MaxFrames = 64
	}
	if o.MaxPatterns == 0 {
		o.MaxPatterns = 1 << 16
	}
	if o.MaxClasses == 0 {
		o.MaxClasses = 1 << 20
	}
	if o.MaxBytes == 0 {
		o.MaxBytes = 1 << 28
	}
	if o.MaxWork == 0 {
		o.MaxWork = 1 << 27
	}
	if o.MaxDepth == 0 {
		o.MaxDepth = 64
	}
	return o, nil
}

type FRAStats struct {
	Nodes, Clauses, Frames, Patterns, Classes, Bytes, Work int
	SATQueries, VerifiedMerges                             uint64
}

type FRAPattern struct{ Inputs []bool }

type FRASignature struct {
	Node AigID
	Bits []uint64
}

type FRAClassMember struct {
	Node       AigID
	Complement bool
}

type FRAClass struct {
	Representative AigID
	Members        []FRAClassMember
}

type FRAWitness struct {
	Inputs []bool
	States [][]bool
	Frames [][]bool
}

func (w FRAWitness) Clone() FRAWitness {
	r := FRAWitness{Inputs: slices.Clone(w.Inputs), States: make([][]bool, len(w.States)), Frames: make([][]bool, len(w.Frames))}
	for i := range w.States {
		r.States[i] = slices.Clone(w.States[i])
	}
	for i := range w.Frames {
		r.Frames[i] = slices.Clone(w.Frames[i])
	}
	return r
}

type FRAEquivalence struct {
	Status  FraigRelationStatus
	Witness FRAWitness
	// Bounded is true only for explicit-state sequential checks. A Proved
	// bounded result proves exactly BoundedFrames frames; it is not an
	// unbounded induction result.
	Bounded       bool
	BoundedFrames int
}

type FRASweepResult struct {
	Graph   *AigManager
	Mapping []AigLit
	Classes []FRAClass
}

type FRAManager struct {
	graph   *AigManager
	opts    FRAOptions
	oracle  *FraigOrchestrator
	toFraig []FraigEdge
	stats   FRAStats
	closed  bool
}

func fraEdgeNotIf(edge FraigEdge, condition bool) FraigEdge {
	if condition {
		return edge.Not()
	}
	return edge
}

func NewFRAManager(source *AigManager, options FRAOptions) (*FRAManager, error) {
	o, err := options.normalized()
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, ErrInvariant
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	if err = source.Check(); err != nil {
		return nil, err
	}
	if source.ObjectCount() > o.MaxNodes || source.ObjectCount() > o.MaxBytes/fraNodeBytes {
		return nil, ErrLimit
	}
	clauses, depth := 1, 0
	for _, obj := range source.objects {
		if obj.Type == AigAND {
			clauses += 3
		}
		if int(obj.Level) > depth {
			depth = int(obj.Level)
		}
	}
	if clauses > o.MaxClauses || depth > o.MaxDepth {
		return nil, ErrLimit
	}
	m := &FRAManager{graph: source.Clone(), opts: o, stats: FRAStats{Nodes: source.ObjectCount(), Clauses: clauses, Bytes: source.ObjectCount() * fraNodeBytes}}
	if err = m.buildOracle(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *FRAManager) buildOracle() error {
	fo := FraigOrchestrationOptions{Manager: FraigOptions{MaxNodes: m.opts.MaxNodes, MaxInputs: m.opts.MaxNodes, MaxOutputs: m.opts.MaxNodes, Stop: m.opts.Stop}, MaxClauses: m.opts.MaxClauses, MaxVisits: m.opts.MaxWork, Stop: m.opts.Stop}
	o, err := NewFraigOrchestrator(fo)
	if err != nil {
		return err
	}
	mapLit := make([]FraigEdge, len(m.graph.objects))
	mapLit[0] = o.manager.ConstantOne()
	for _, id := range m.graph.cis {
		e, eerr := o.NewInput("")
		if eerr != nil {
			o.Close()
			return eerr
		}
		mapLit[id] = e
	}
	for id, obj := range m.graph.objects {
		if id == 0 || obj.Type == AigCI || obj.Type == AigCO || obj.Type == AigNone {
			continue
		}
		get := func(l AigLit) FraigEdge { return fraEdgeNotIf(mapLit[l.ID()], l.Complemented()) }
		switch obj.Type {
		case AigAND:
			mapLit[id], err = o.And(get(obj.Fanin0), get(obj.Fanin1))
		case AigBuffer:
			mapLit[id] = get(obj.Fanin0)
		default:
			err = ErrInvariant
		}
		if err != nil {
			o.Close()
			return err
		}
	}
	m.oracle, m.toFraig = o, mapLit
	return nil
}

func (m *FRAManager) check() error {
	if m == nil || m.closed || m.graph == nil || m.oracle == nil {
		return ErrFRAClosed
	}
	if m.opts.Stop != nil && m.opts.Stop() {
		return ErrCancelled
	}
	return nil
}

func (m *FRAManager) Close() {
	if m != nil && !m.closed {
		m.closed = true
		if m.oracle != nil {
			m.oracle.Close()
		}
		m.graph = nil
		m.toFraig = nil
	}
}
func (m *FRAManager) Stats() FRAStats {
	if m == nil {
		return FRAStats{}
	}
	return m.stats
}
func (m *FRAManager) Graph() *AigManager {
	if m == nil || m.closed {
		return nil
	}
	return m.graph.Clone()
}

func (m *FRAManager) spend(work, bytes int) error {
	if work < 0 || bytes < 0 || work > m.opts.MaxWork-m.stats.Work || bytes > m.opts.MaxBytes-m.stats.Bytes {
		return ErrLimit
	}
	m.stats.Work += work
	m.stats.Bytes += bytes
	if m.opts.Stop != nil && m.opts.Stop() {
		return ErrCancelled
	}
	return nil
}

func (m *FRAManager) values(inputs []bool) ([]bool, error) {
	if len(inputs) != len(m.graph.cis) {
		return nil, ErrDimension
	}
	if err := m.spend(len(m.graph.objects), 0); err != nil {
		return nil, err
	}
	v := make([]bool, len(m.graph.objects))
	v[0] = true
	for i, id := range m.graph.cis {
		v[id] = inputs[i]
	}
	read := func(l AigLit) bool { return v[l.ID()] != l.Complemented() }
	for id, obj := range m.graph.objects {
		switch obj.Type {
		case AigAND:
			v[id] = read(obj.Fanin0) && read(obj.Fanin1)
		case AigBuffer, AigCO:
			v[id] = read(obj.Fanin0)
		}
	}
	return v, nil
}

func (m *FRAManager) Signatures(patterns []FRAPattern) ([]FRASignature, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	if len(patterns) > m.opts.MaxPatterns {
		return nil, ErrLimit
	}
	if len(patterns) > m.opts.MaxPatterns-m.stats.Patterns {
		return nil, ErrLimit
	}
	for _, pattern := range patterns {
		if len(pattern.Inputs) != len(m.graph.cis) {
			return nil, ErrDimension
		}
	}
	words := (len(patterns) + 63) / 64
	remainingBytes := m.opts.MaxBytes - m.stats.Bytes
	if words != 0 && len(m.graph.objects) > remainingBytes/(words*8) {
		return nil, ErrLimit
	}
	bytes := len(m.graph.objects) * words * 8
	if len(patterns) > (remainingBytes-bytes)/fraPatternBytes {
		return nil, ErrLimit
	}
	bytes += len(patterns) * fraPatternBytes
	if err := m.spend(0, bytes); err != nil {
		return nil, err
	}
	result := make([]FRASignature, len(m.graph.objects))
	for id := range result {
		result[id] = FRASignature{AigID(id), make([]uint64, words)}
	}
	for p, pattern := range patterns {
		values, err := m.values(pattern.Inputs)
		if err != nil {
			return nil, err
		}
		for id, value := range values {
			if value {
				result[id].Bits[p>>6] |= uint64(1) << uint(p&63)
			}
		}
	}
	m.stats.Patterns += len(patterns)
	return result, nil
}

func complementSignature(bits []uint64, count int) []uint64 {
	r := make([]uint64, len(bits))
	for i := range bits {
		r[i] = ^bits[i]
	}
	if count&63 != 0 && len(r) != 0 {
		r[len(r)-1] &= uint64(1)<<uint(count&63) - 1
	}
	return r
}

func signatureKey(bits []uint64) string { return fmt.Sprint(bits) }

func (m *FRAManager) Classes(patterns []FRAPattern) ([]FRAClass, error) {
	sigs, err := m.Signatures(patterns)
	if err != nil {
		return nil, err
	}
	groups := make(map[string][]FRAClassMember)
	for id, sig := range sigs {
		if m.graph.objects[id].Type == AigCO || m.graph.objects[id].Type == AigNone {
			continue
		}
		direct, inv := signatureKey(sig.Bits), signatureKey(complementSignature(sig.Bits, len(patterns)))
		key, phase := direct, false
		if inv < direct {
			key, phase = inv, true
		}
		groups[key] = append(groups[key], FRAClassMember{AigID(id), phase})
	}
	keys := make([]string, 0, len(groups))
	for key, members := range groups {
		if len(members) > 1 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if len(keys) > m.opts.MaxClasses {
		return nil, ErrLimit
	}
	if len(keys) > m.opts.MaxClasses-m.stats.Classes {
		return nil, ErrLimit
	}
	classes := make([]FRAClass, 0, len(keys))
	for _, key := range keys {
		members := groups[key]
		sort.Slice(members, func(i, j int) bool { return members[i].Node < members[j].Node })
		base := members[0].Complement
		for i := range members {
			members[i].Complement = members[i].Complement != base
		}
		classes = append(classes, FRAClass{members[0].Node, members})
	}
	if err = m.spend(0, len(classes)*fraClassBytes); err != nil {
		return nil, err
	}
	m.stats.Classes += len(classes)
	return classes, nil
}

func (m *FRAManager) VerifyEquivalent(a, b AigLit) (FRAEquivalence, error) {
	if err := m.check(); err != nil {
		return FRAEquivalence{}, err
	}
	if !m.graph.validSignal(a) || !m.graph.validSignal(b) {
		return FRAEquivalence{}, ErrInvalidLiteral
	}
	m.stats.SATQueries++
	remaining := m.opts.MaxWork - m.stats.Work
	aEdge := fraEdgeNotIf(m.toFraig[a.ID()], a.Complemented())
	bEdge := fraEdgeNotIf(m.toFraig[b.ID()], b.Complemented())
	if remaining == 0 && aEdge != bEdge {
		return FRAEquivalence{}, ErrLimit
	}
	r, err := m.oracle.Equivalent(aEdge, bEdge, FraigSATLimits{Conflicts: uint64(remaining), Propagations: uint64(remaining), Decisions: uint64(remaining), Stop: m.opts.Stop})
	if err != nil {
		return FRAEquivalence{}, err
	}
	work64 := r.Stats.Conflicts
	if r.Stats.Propagations > uint64(remaining)-min(work64, uint64(remaining)) {
		return FRAEquivalence{}, ErrLimit
	}
	work64 += r.Stats.Propagations
	if work64 > uint64(remaining) || r.Stats.Decisions > uint64(remaining)-work64 {
		return FRAEquivalence{}, ErrLimit
	}
	work64 += r.Stats.Decisions
	if err = m.spend(int(work64), 0); err != nil {
		return FRAEquivalence{}, err
	}
	if r.Status == FraigRelationDisproved {
		values, verr := m.values(r.Model)
		if verr != nil {
			return FRAEquivalence{}, verr
		}
		if values[a.ID()] != a.Complemented() == (values[b.ID()] != b.Complemented()) {
			return FRAEquivalence{}, ErrInvariant
		}
		return FRAEquivalence{Status: r.Status, Witness: FRAWitness{Inputs: slices.Clone(r.Model)}}, nil
	}
	return FRAEquivalence{Status: r.Status}, nil
}

// VerifiedClasses refines simulation candidates with the concrete SAT oracle.
// Every returned member has been proved equivalent to its representative with
// the recorded complement phase; disproved and resource-unknown candidates
// are omitted.
func (m *FRAManager) VerifiedClasses(patterns []FRAPattern) ([]FRAClass, error) {
	candidates, err := m.Classes(patterns)
	if err != nil {
		return nil, err
	}
	verified := make([]FRAClass, 0, len(candidates))
	for _, candidate := range candidates {
		members := []FRAClassMember{candidate.Members[0]}
		for _, member := range candidate.Members[1:] {
			representative := AigLit(candidate.Representative) << 1
			query := (AigLit(member.Node) << 1).NotIf(member.Complement)
			proof, proveErr := m.VerifyEquivalent(representative, query)
			if proveErr != nil {
				return nil, proveErr
			}
			if proof.Status == FraigRelationProved {
				members = append(members, member)
			}
		}
		if len(members) > 1 {
			verified = append(verified, FRAClass{Representative: candidate.Representative, Members: members})
		}
	}
	return verified, nil
}

// VerifySequentialEquivalent checks nodes at every reachable frame from the
// all-zero latch state. It is a bounded check and never claims unbounded proof.
func (m *FRAManager) VerifySequentialEquivalent(a, b AigLit, frames int) (FRAEquivalence, error) {
	if err := m.check(); err != nil {
		return FRAEquivalence{}, err
	}
	if !m.graph.validSignal(a) || !m.graph.validSignal(b) {
		return FRAEquivalence{}, ErrInvalidLiteral
	}
	if frames < 0 {
		return FRAEquivalence{}, ErrDimension
	}
	if frames > m.opts.MaxFrames || frames > m.opts.MaxDepth {
		return FRAEquivalence{}, ErrLimit
	}
	pi, regs := m.graph.PrimaryInputCount(), m.graph.RegisterCount()
	if pi >= 63 || uint64(1)<<uint(pi) > uint64(m.opts.MaxPatterns) {
		return FRAEquivalence{}, ErrLimit
	}
	type path struct {
		state  []bool
		inputs [][]bool
		states [][]bool
	}
	front := []path{{state: make([]bool, regs), states: [][]bool{make([]bool, regs)}}}
	for frame := 0; frame < frames; frame++ {
		if m.stats.Frames == m.opts.MaxFrames {
			return FRAEquivalence{}, ErrLimit
		}
		m.stats.Frames++
		next := make([]path, 0)
		for _, p := range front {
			for mask := uint64(0); mask < uint64(1)<<uint(pi); mask++ {
				if m.stats.Patterns == m.opts.MaxPatterns {
					return FRAEquivalence{}, ErrLimit
				}
				m.stats.Patterns++
				pathBytes := fraPatternBytes + pi + regs + (frame+1)*(pi+regs)
				if pathBytes < 0 {
					return FRAEquivalence{}, ErrLimit
				}
				if err := m.spend(0, pathBytes); err != nil {
					return FRAEquivalence{}, err
				}
				in := make([]bool, pi)
				for i := range in {
					in[i] = mask>>uint(i)&1 != 0
				}
				all := append(slices.Clone(in), p.state...)
				values, err := m.values(all)
				if err != nil {
					return FRAEquivalence{}, err
				}
				av, bv := values[a.ID()] != a.Complemented(), values[b.ID()] != b.Complemented()
				if av != bv {
					return FRAEquivalence{Status: FraigRelationDisproved, Witness: FRAWitness{States: slices.Clone(p.states), Frames: append(slices.Clone(p.inputs), in)}, Bounded: true, BoundedFrames: frame + 1}, nil
				}
				_, ns, err := m.graph.Step(in, p.state)
				if err != nil {
					return FRAEquivalence{}, err
				}
				np := path{state: ns, inputs: append(slices.Clone(p.inputs), in), states: append(slices.Clone(p.states), slices.Clone(ns))}
				next = append(next, np)
			}
		}
		front = next
	}
	return FRAEquivalence{Status: FraigRelationProved, Bounded: true, BoundedFrames: frames}, nil
}

func (m *FRAManager) ReplaySequential(a, b AigLit, witness FRAWitness) error {
	if err := m.check(); err != nil {
		return err
	}
	if !m.graph.validSignal(a) || !m.graph.validSignal(b) {
		return ErrInvalidLiteral
	}
	if len(witness.Frames) == 0 || len(witness.Frames) != len(witness.States) {
		return ErrDimension
	}
	if len(witness.Frames) > m.opts.MaxFrames || len(witness.Frames) > m.opts.MaxDepth {
		return ErrLimit
	}
	state := make([]bool, m.graph.RegisterCount())
	for frame, in := range witness.Frames {
		if len(in) != m.graph.PrimaryInputCount() || len(witness.States[frame]) != len(state) || !slices.Equal(witness.States[frame], state) {
			return ErrDimension
		}
		values, err := m.values(append(slices.Clone(in), state...))
		if err != nil {
			return err
		}
		different := (values[a.ID()] != a.Complemented()) != (values[b.ID()] != b.Complemented())
		if different != (frame == len(witness.Frames)-1) {
			return ErrInvariant
		}
		_, state, err = m.graph.Step(in, state)
		if err != nil {
			return err
		}
	}
	return nil
}

// Sweep only commits a candidate merge after the SAT oracle proves it. The
// source manager is never mutated; any cancellation or limit leaves it intact.
func (m *FRAManager) Sweep(patterns []FRAPattern) (FRASweepResult, error) {
	if err := m.check(); err != nil {
		return FRASweepResult{}, err
	}
	classes, err := m.VerifiedClasses(patterns)
	if err != nil {
		return FRASweepResult{}, err
	}
	replace := make(map[AigID]AigLit)
	mergeCount := uint64(0)
	for _, class := range classes {
		for _, member := range class.Members[1:] {
			target := AigLit(class.Representative) << 1
			target = target.NotIf(member.Complement)
			replace[member.Node] = target
			mergeCount++
		}
	}
	target := NewAigManager()
	mapping := make([]AigLit, len(m.graph.objects))
	mapping[0] = AigConst1
	for _, id := range m.graph.cis {
		if err = m.spend(1, 0); err != nil {
			return FRASweepResult{}, err
		}
		mapping[id], err = target.CreateCI()
		if err != nil {
			return FRASweepResult{}, err
		}
	}
	// Every target CI is created to preserve the PI/register partition, but a
	// verified CI replacement must still be reflected in the public old-to-new
	// signal mapping and in subsequent fanin resolution.
	for _, id := range m.graph.cis {
		if replacement, ok := replace[id]; ok {
			mapping[id] = mapping[replacement.ID()].NotIf(replacement.Complemented())
		}
	}
	resolve := func(l AigLit) AigLit {
		seen := 0
		for {
			r, ok := replace[l.ID()]
			if !ok {
				break
			}
			l = r.NotIf(l.Complemented())
			seen++
			if seen > m.opts.MaxDepth {
				return InvalidAigLit
			}
		}
		return mapping[l.ID()].NotIf(l.Complemented())
	}
	for id, obj := range m.graph.objects {
		if id == 0 || obj.Type == AigCI || obj.Type == AigCO || obj.Type == AigNone {
			continue
		}
		if _, ok := replace[AigID(id)]; ok {
			mapping[id] = resolve(AigLit(id) << 1)
			continue
		}
		if err = m.spend(1, 0); err != nil {
			return FRASweepResult{}, err
		}
		switch obj.Type {
		case AigAND:
			mapping[id], err = target.And(resolve(obj.Fanin0), resolve(obj.Fanin1))
		case AigBuffer:
			mapping[id], err = target.CreateBuffer(resolve(obj.Fanin0))
		default:
			err = ErrInvariant
		}
		if err != nil {
			return FRASweepResult{}, err
		}
	}
	for _, id := range m.graph.cos {
		if err = m.spend(1, 0); err != nil {
			return FRASweepResult{}, err
		}
		driver := resolve(m.graph.objects[id].Fanin0)
		var newID AigID
		newID, err = target.CreateCO(driver)
		if err != nil {
			return FRASweepResult{}, err
		}
		mapping[id] = AigLit(newID) << 1
	}
	if err = target.SetRegisterCount(m.graph.registers); err != nil {
		return FRASweepResult{}, err
	}
	if err = target.SetConstraintCount(m.graph.constraints); err != nil {
		return FRASweepResult{}, err
	}
	if err = target.Check(); err != nil {
		return FRASweepResult{}, err
	}
	m.stats.VerifiedMerges += mergeCount
	return FRASweepResult{target, mapping, classes}, nil
}
