package abc

// Stand-alone Go port of src/proof/ssc.

import (
	"encoding/binary"
	"slices"
)

// SSCParameters corresponds to Ssc_Pars_t. The Max fields are Go safety
// bounds; zero selects deterministic defaults.
type SSCParameters struct {
	SimulationWords, ConflictLimit, MaxSATVariables, CallsBeforeRecycle int
	AppendConstraints, Verify                                           bool
	MaxAssignments, MaxObjects                                          int
	Stop                                                                func() bool
}

func DefaultSSCParameters() SSCParameters {
	return SSCParameters{SimulationWords: 1, ConflictLimit: 1000, MaxSATVariables: 5000, CallsBeforeRecycle: 100, MaxAssignments: 1 << 20, MaxObjects: 1 << 20}
}

type SSCLiteral struct {
	ID           GiaID
	Complemented bool
}

type SSCClass struct {
	Representative GiaID
	Members        []SSCLiteral
}

type SSCWitness struct {
	PrimaryInputs []bool
	State         []bool
}

func (w SSCWitness) Inputs() []bool {
	r := make([]bool, 0, len(w.PrimaryInputs)+len(w.State))
	r = append(r, w.PrimaryInputs...)
	r = append(r, w.State...)
	return r
}

type SSCEquivalenceResult struct {
	Status      SatStatus
	Witness     *SSCWitness
	Assignments int
}

type SSCSweepResult struct {
	Manager       *GiaManager
	Classes       []SSCClass
	Pivot         *SSCWitness
	Unsatisfiable bool
	Assignments   int
}

func normalizeSSCParameters(p SSCParameters) (SSCParameters, error) {
	if p.SimulationWords < 0 || p.ConflictLimit < 0 || p.MaxSATVariables < 0 || p.CallsBeforeRecycle < 0 || p.MaxAssignments < 0 || p.MaxObjects < 0 {
		return p, ErrDimension
	}
	d := DefaultSSCParameters()
	if p.SimulationWords == 0 {
		p.SimulationWords = d.SimulationWords
	}
	if p.ConflictLimit == 0 {
		p.ConflictLimit = d.ConflictLimit
	}
	if p.MaxSATVariables == 0 {
		p.MaxSATVariables = d.MaxSATVariables
	}
	if p.CallsBeforeRecycle == 0 {
		p.CallsBeforeRecycle = d.CallsBeforeRecycle
	}
	if p.MaxAssignments == 0 {
		p.MaxAssignments = d.MaxAssignments
	}
	if p.MaxObjects == 0 {
		p.MaxObjects = d.MaxObjects
	}
	return p, nil
}

func sscCheckGraphs(aig, care *GiaManager, p SSCParameters) (SSCParameters, error) {
	var err error
	if p, err = normalizeSSCParameters(p); err != nil {
		return p, err
	}
	if aig == nil || care == nil {
		return p, ErrDimension
	}
	if err = aig.Check(); err != nil {
		return p, err
	}
	if err = care.Check(); err != nil {
		return p, err
	}
	if !aig.IsNormalized() || !care.IsNormalized() {
		return p, ErrInvariant
	}
	if aig.InputCount() != care.InputCount() || care.RegisterCount() != 0 {
		return p, ErrDimension
	}
	if aig.ObjectCount() > p.MaxObjects || care.ObjectCount() > p.MaxObjects {
		return p, ErrLimit
	}
	if p.Stop != nil && p.Stop() {
		return p, ErrCancelled
	}
	return p, nil
}

func sscAssignmentCount(inputs, maximum int) (int, error) {
	if inputs < 0 || inputs >= 63 {
		return 0, ErrLimit
	}
	total := uint64(1) << uint(inputs)
	if total > uint64(maximum) || total > uint64(^uint(0)>>1) {
		return 0, ErrLimit
	}
	return int(total), nil
}

func sscBits(value, count int) []bool {
	r := make([]bool, count)
	for i := range r {
		r[i] = value&(1<<uint(i)) != 0
	}
	return r
}

func sscWitness(aig *GiaManager, inputs []bool) *SSCWitness {
	n := aig.PrimaryInputCount()
	return &SSCWitness{PrimaryInputs: slices.Clone(inputs[:n]), State: slices.Clone(inputs[n:])}
}

func sscCareHolds(care *GiaManager, inputs []bool) (bool, error) {
	values, err := care.evaluateValues(inputs)
	if err != nil {
		return false, err
	}
	for _, id := range care.cos {
		if values[id] {
			return false, nil
		}
	}
	return true, nil
}

// SSCCheckEquivalence implements Ssc_ManCheckEquivalence's result convention:
// SAT carries a differing admissible assignment; UNSAT proves equivalence.
func SSCCheckEquivalence(aig, care *GiaManager, first, second GiaLit, p SSCParameters) (SSCEquivalenceResult, error) {
	p, err := sscCheckGraphs(aig, care, p)
	if err != nil {
		return SSCEquivalenceResult{}, err
	}
	if !aig.validSignal(first) || !aig.validSignal(second) {
		return SSCEquivalenceResult{}, ErrInvalidLiteral
	}
	total, err := sscAssignmentCount(aig.InputCount(), p.MaxAssignments)
	if err != nil {
		return SSCEquivalenceResult{}, err
	}
	for assignment := 0; assignment < total; assignment++ {
		if p.Stop != nil && p.Stop() {
			return SSCEquivalenceResult{}, ErrCancelled
		}
		inputs := sscBits(assignment, aig.InputCount())
		holds, e := sscCareHolds(care, inputs)
		if e != nil {
			return SSCEquivalenceResult{}, e
		}
		if !holds {
			continue
		}
		values, e := aig.evaluateValues(inputs)
		if e != nil {
			return SSCEquivalenceResult{}, e
		}
		read := func(l GiaLit) bool { return values[l.ID()] != l.Complemented() }
		if read(first) != read(second) {
			return SSCEquivalenceResult{Status: SatSatisfiable, Witness: sscWitness(aig, inputs), Assignments: assignment + 1}, nil
		}
	}
	return SSCEquivalenceResult{Status: SatUnsatisfiable, Assignments: total}, nil
}

type sscSignature struct {
	id       GiaID
	words    []uint64
	inverted bool
}

func sscKey(words []uint64) string {
	b := make([]byte, 8*len(words))
	for i, word := range words {
		binary.LittleEndian.PutUint64(b[8*i:], word)
	}
	return string(b)
}

func sscCanonical(words []uint64, bits int) (string, bool) {
	inv := make([]uint64, len(words))
	for i := range words {
		inv[i] = ^words[i]
	}
	if rem := bits & 63; rem != 0 && len(inv) > 0 {
		inv[len(inv)-1] &= uint64(1)<<uint(rem) - 1
	}
	a, b := sscKey(words), sscKey(inv)
	if b < a {
		return b, true
	}
	return a, false
}

func sscClasses(aig, care *GiaManager, p SSCParameters) ([]SSCClass, *SSCWitness, int, bool, error) {
	total, err := sscAssignmentCount(aig.InputCount(), p.MaxAssignments)
	if err != nil {
		return nil, nil, 0, false, err
	}
	candidates := []GiaID{0}
	for id := 1; id < aig.ObjectCount(); id++ {
		if aig.objects[id].IsAnd() {
			candidates = append(candidates, GiaID(id))
		}
	}
	sigs := make([][]uint64, len(candidates))
	var pivot *SSCWitness
	cares := 0
	for assignment := 0; assignment < total; assignment++ {
		if p.Stop != nil && p.Stop() {
			return nil, nil, 0, false, ErrCancelled
		}
		inputs := sscBits(assignment, aig.InputCount())
		holds, e := sscCareHolds(care, inputs)
		if e != nil {
			return nil, nil, 0, false, e
		}
		if !holds {
			continue
		}
		values, e := aig.evaluateValues(inputs)
		if e != nil {
			return nil, nil, 0, false, e
		}
		if pivot == nil {
			pivot = sscWitness(aig, inputs)
		}
		word := cares >> 6
		for i, id := range candidates {
			if word == len(sigs[i]) {
				sigs[i] = append(sigs[i], 0)
			}
			if values[id] {
				sigs[i][word] |= uint64(1) << uint(cares&63)
			}
		}
		cares++
	}
	if cares == 0 {
		return nil, nil, total, true, nil
	}
	type bucketMember struct {
		id  GiaID
		inv bool
	}
	buckets := make(map[string][]bucketMember)
	order := make([]string, 0)
	for i, id := range candidates {
		key, inv := sscCanonical(sigs[i], cares)
		if _, ok := buckets[key]; !ok {
			order = append(order, key)
		}
		buckets[key] = append(buckets[key], bucketMember{id, inv})
	}
	classes := make([]SSCClass, 0)
	for _, key := range order {
		bucket := buckets[key]
		if len(bucket) < 2 {
			continue
		}
		repInv := bucket[0].inv
		class := SSCClass{Representative: bucket[0].id, Members: make([]SSCLiteral, len(bucket))}
		for i, member := range bucket {
			class.Members[i] = SSCLiteral{ID: member.id, Complemented: member.inv != repInv}
		}
		classes = append(classes, class)
	}
	return classes, pivot, total, false, nil
}

// SSCClasses returns exact care-constrained functional classes of constant and
// internal nodes. Members are ordered by object ID.
func SSCClasses(aig, care *GiaManager, p SSCParameters) ([]SSCClass, *SSCWitness, error) {
	p, err := sscCheckGraphs(aig, care, p)
	if err != nil {
		return nil, nil, err
	}
	classes, pivot, _, _, err := sscClasses(aig, care, p)
	return classes, pivot, err
}

func sscZeroGraph(source *GiaManager, p SSCParameters) (*GiaManager, error) {
	g, err := NewGiaManagerWithLimit(p.MaxObjects)
	if err != nil {
		return nil, err
	}
	g.Name, g.Spec = source.Name, source.Spec
	for range source.cis {
		if _, err = g.AppendCI(); err != nil {
			return nil, err
		}
	}
	for range source.cos {
		if _, err = g.AppendCO(GiaConst0); err != nil {
			return nil, err
		}
	}
	if err = g.SetRegisterCount(source.registers); err != nil {
		return nil, err
	}
	if err = g.SetConstraintCount(source.constraints); err != nil {
		return nil, err
	}
	return g, nil
}

func sscRebuild(source *GiaManager, classes []SSCClass, p SSCParameters) (*GiaManager, error) {
	replace := make(map[GiaID]SSCLiteral)
	for _, class := range classes {
		for _, member := range class.Members[1:] {
			replace[member.ID] = SSCLiteral{ID: class.Representative, Complemented: member.Complemented}
		}
	}
	g, err := NewGiaManagerWithLimit(p.MaxObjects)
	if err != nil {
		return nil, err
	}
	g.Name, g.Spec = source.Name, source.Spec
	mapping := make([]GiaLit, source.ObjectCount())
	mapping[0] = GiaConst0
	for id := 1; id < source.ObjectCount(); id++ {
		if p.Stop != nil && p.Stop() {
			return nil, ErrCancelled
		}
		obj := source.objects[id]
		if obj.IsCI() {
			mapping[id], err = g.AppendCI()
		} else if obj.IsAnd() {
			if relation, ok := replace[GiaID(id)]; ok {
				mapping[id] = mapping[relation.ID].NotIf(relation.Complemented)
				continue
			}
			a, _ := obj.Fanin0(GiaID(id))
			b, _ := obj.Fanin1(GiaID(id))
			a = mapping[a.ID()].NotIf(a.Complemented())
			b = mapping[b.ID()].NotIf(b.Complemented())
			if obj.IsBuffer() {
				mapping[id], err = g.AppendBuffer(a)
			} else if obj.IsXor() {
				mapping[id], err = g.AppendXorReal(a, b)
			} else if control, yes := source.MuxControl(GiaID(id)); yes {
				mapping[id], err = g.AppendMuxReal(mapping[control.ID()].NotIf(control.Complemented()), b, a)
			} else {
				mapping[id], err = g.AppendAnd(a, b)
			}
		} else if obj.IsCO() {
			a, _ := obj.Fanin0(GiaID(id))
			mapping[id], err = g.AppendCO(mapping[a.ID()].NotIf(a.Complemented()))
		}
		if err != nil {
			return nil, err
		}
	}
	if err = g.SetRegisterCount(source.registers); err != nil {
		return nil, err
	}
	if err = g.SetConstraintCount(source.constraints); err != nil {
		return nil, err
	}
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}

// SSCVerify checks sequential combinational correspondence under the care set
// and returns the first replayable counterexample, if any.
func SSCVerify(original, reduced, care *GiaManager, p SSCParameters) (SSCEquivalenceResult, error) {
	p, err := sscCheckGraphs(original, care, p)
	if err != nil {
		return SSCEquivalenceResult{}, err
	}
	if reduced == nil || reduced.InputCount() != original.InputCount() || reduced.OutputCount() != original.OutputCount() || reduced.RegisterCount() != original.RegisterCount() {
		return SSCEquivalenceResult{}, ErrDimension
	}
	if err = reduced.Check(); err != nil {
		return SSCEquivalenceResult{}, err
	}
	if reduced.ObjectCount() > p.MaxObjects {
		return SSCEquivalenceResult{}, ErrLimit
	}
	total, err := sscAssignmentCount(original.InputCount(), p.MaxAssignments)
	if err != nil {
		return SSCEquivalenceResult{}, err
	}
	for assignment := 0; assignment < total; assignment++ {
		if p.Stop != nil && p.Stop() {
			return SSCEquivalenceResult{}, ErrCancelled
		}
		inputs := sscBits(assignment, original.InputCount())
		holds, e := sscCareHolds(care, inputs)
		if e != nil {
			return SSCEquivalenceResult{}, e
		}
		if !holds {
			continue
		}
		want, e := original.Evaluate(inputs)
		if e != nil {
			return SSCEquivalenceResult{}, e
		}
		got, e := reduced.Evaluate(inputs)
		if e != nil {
			return SSCEquivalenceResult{}, e
		}
		if !slices.Equal(want, got) {
			return SSCEquivalenceResult{Status: SatSatisfiable, Witness: sscWitness(original, inputs), Assignments: assignment + 1}, nil
		}
	}
	return SSCEquivalenceResult{Status: SatUnsatisfiable, Assignments: total}, nil
}

// SSCSweep performs exact constrained SAT sweeping. It never mutates either
// input; errors and cancellation return no partial manager.
func SSCSweep(aig, care *GiaManager, p SSCParameters) (SSCSweepResult, error) {
	p, err := sscCheckGraphs(aig, care, p)
	if err != nil {
		return SSCSweepResult{}, err
	}
	classes, pivot, assignments, unsat, err := sscClasses(aig, care, p)
	if err != nil {
		return SSCSweepResult{}, err
	}
	if unsat {
		g, e := sscZeroGraph(aig, p)
		if e != nil {
			return SSCSweepResult{}, e
		}
		return SSCSweepResult{Manager: g, Unsatisfiable: true, Assignments: assignments}, nil
	}
	g, err := sscRebuild(aig, classes, p)
	if err != nil {
		return SSCSweepResult{}, err
	}
	if p.Verify {
		verified, e := SSCVerify(aig, g, care, p)
		if e != nil {
			return SSCSweepResult{}, e
		}
		if verified.Status != SatUnsatisfiable {
			return SSCSweepResult{}, ErrInvariant
		}
	}
	return SSCSweepResult{Manager: g, Classes: classes, Pivot: pivot, Assignments: assignments}, nil
}

func sscSelectOutputs(source *GiaManager, outputs []int, registers int, p SSCParameters) (*GiaManager, error) {
	if source == nil {
		return nil, ErrDimension
	}
	if source.ObjectCount() > p.MaxObjects {
		return nil, ErrLimit
	}
	g, err := NewGiaManagerWithLimit(p.MaxObjects)
	if err != nil {
		return nil, err
	}
	mapping := make([]GiaLit, source.ObjectCount())
	for id := 1; id < source.ObjectCount(); id++ {
		if p.Stop != nil && p.Stop() {
			return nil, ErrCancelled
		}
		obj := source.objects[id]
		if obj.IsCI() {
			mapping[id], err = g.AppendCI()
		} else if obj.IsAnd() {
			a, _ := obj.Fanin0(GiaID(id))
			b, _ := obj.Fanin1(GiaID(id))
			a, b = mapping[a.ID()].NotIf(a.Complemented()), mapping[b.ID()].NotIf(b.Complemented())
			if obj.IsBuffer() {
				mapping[id], err = g.AppendBuffer(a)
			} else if obj.IsXor() {
				mapping[id], err = g.AppendXorReal(a, b)
			} else if control, yes := source.MuxControl(GiaID(id)); yes {
				mapping[id], err = g.AppendMuxReal(mapping[control.ID()].NotIf(control.Complemented()), b, a)
			} else {
				mapping[id], err = g.AppendAnd(a, b)
			}
		}
		if err != nil {
			return nil, err
		}
	}
	for _, output := range outputs {
		if output < 0 || output >= source.OutputCount() {
			return nil, ErrDimension
		}
		id := source.cos[output]
		driver, _ := source.objects[id].Fanin0(id)
		if _, err = g.AppendCO(mapping[driver.ID()].NotIf(driver.Complemented())); err != nil {
			return nil, err
		}
	}
	if err = g.SetRegisterCount(registers); err != nil {
		return nil, err
	}
	return g, g.Check()
}

// SSCSweepConstraints ports Ssc_PerformSweepingConstr. The final constrained
// primary outputs are violation outputs; register outputs remain in the subject.
func SSCSweepConstraints(source *GiaManager, p SSCParameters) (SSCSweepResult, error) {
	p, err := normalizeSSCParameters(p)
	if err != nil {
		return SSCSweepResult{}, err
	}
	if source == nil {
		return SSCSweepResult{}, ErrDimension
	}
	if err = source.Check(); err != nil {
		return SSCSweepResult{}, err
	}
	if source.ObjectCount() > p.MaxObjects {
		return SSCSweepResult{}, ErrLimit
	}
	nPO, nConstraints := source.PrimaryOutputCount(), source.ConstraintCount()
	if nConstraints < 0 || nConstraints > nPO {
		return SSCSweepResult{}, ErrInvariant
	}
	subjectOut := make([]int, 0, source.OutputCount()-nConstraints)
	for i := 0; i < nPO-nConstraints; i++ {
		subjectOut = append(subjectOut, i)
	}
	for i := nPO; i < source.OutputCount(); i++ {
		subjectOut = append(subjectOut, i)
	}
	careOut := make([]int, nConstraints)
	for i := range careOut {
		careOut[i] = nPO - nConstraints + i
	}
	subject, err := sscSelectOutputs(source, subjectOut, source.RegisterCount(), p)
	if err != nil {
		return SSCSweepResult{}, err
	}
	var care *GiaManager
	if nConstraints == 0 {
		care, err = sscSelectOutputs(source, nil, 0, p)
		if err == nil {
			_, err = care.AppendCO(GiaConst0)
		}
	} else {
		care, err = sscSelectOutputs(source, careOut, 0, p)
	}
	if err != nil {
		return SSCSweepResult{}, err
	}
	result, err := SSCSweep(subject, care, p)
	if err != nil || !p.AppendConstraints {
		return result, err
	}
	if result.Unsatisfiable {
		full := source.Clone()
		for i := 0; i < nPO-nConstraints; i++ {
			if err = full.PatchCODriver(i, GiaConst0); err != nil {
				return SSCSweepResult{}, err
			}
		}
		for i := nPO; i < source.OutputCount(); i++ {
			if err = full.PatchCODriver(i, GiaConst0); err != nil {
				return SSCSweepResult{}, err
			}
		}
		result.Manager = full
		return result, nil
	}
	// Gate IDs in the selected subject match the source because all internal
	// objects were copied. Rebuilding the full source therefore appends the
	// original constraints while sharing every proved replacement.
	full, err := sscRebuild(source, result.Classes, p)
	if err != nil {
		return SSCSweepResult{}, err
	}
	result.Manager = full
	return result, nil
}

// SSCDropContained ports Gia_ManDropContained. An output is retained exactly
// when it can be the sole asserted output for some input assignment.
func SSCDropContained(source *GiaManager, p SSCParameters) (*GiaManager, error) {
	p, err := normalizeSSCParameters(p)
	if err != nil {
		return nil, err
	}
	if source == nil || source.RegisterCount() != 0 {
		return nil, ErrDimension
	}
	if err = source.Check(); err != nil {
		return nil, err
	}
	if source.ObjectCount() > p.MaxObjects {
		return nil, ErrLimit
	}
	total, err := sscAssignmentCount(source.InputCount(), p.MaxAssignments)
	if err != nil {
		return nil, err
	}
	keep := make([]bool, source.OutputCount())
	for i := range keep {
		keep[i] = true
	}
	for candidate := range keep {
		essential := false
		for assignment := 0; assignment < total; assignment++ {
			if p.Stop != nil && p.Stop() {
				return nil, ErrCancelled
			}
			outputs, e := source.Evaluate(sscBits(assignment, source.InputCount()))
			if e != nil {
				return nil, e
			}
			if !outputs[candidate] {
				continue
			}
			valid := true
			for other, active := range keep {
				if other != candidate && active && outputs[other] {
					valid = false
					break
				}
			}
			if valid {
				essential = true
				break
			}
		}
		if !essential {
			keep[candidate] = false
		}
	}
	selected := make([]int, 0, len(keep))
	for i, yes := range keep {
		if yes {
			selected = append(selected, i)
		}
	}
	return sscSelectOutputs(source, selected, 0, p)
}

// SSCOptimizeRing ports Gia_ManOptimizeRing: contained outputs are removed,
// then each remaining output is swept while all other outputs are constraints.
func SSCOptimizeRing(source *GiaManager, p SSCParameters) (*GiaManager, error) {
	p, err := normalizeSSCParameters(p)
	if err != nil {
		return nil, err
	}
	current, err := SSCDropContained(source, p)
	if err != nil || current.OutputCount() <= 1 {
		return current, err
	}
	count := current.OutputCount()
	for pivot := 0; pivot < count; pivot++ {
		order := make([]int, count)
		for i := range order {
			order[i] = i
		}
		order[0], order[pivot] = order[pivot], order[0]
		reordered, e := sscSelectOutputs(current, order, 0, p)
		if e != nil {
			return nil, e
		}
		if e = reordered.SetConstraintCount(count - 1); e != nil {
			return nil, e
		}
		result, e := SSCSweepConstraints(reordered, SSCParameters{
			SimulationWords: p.SimulationWords, ConflictLimit: p.ConflictLimit,
			MaxSATVariables: p.MaxSATVariables, CallsBeforeRecycle: p.CallsBeforeRecycle,
			AppendConstraints: true, Verify: p.Verify, MaxAssignments: p.MaxAssignments,
			MaxObjects: p.MaxObjects, Stop: p.Stop,
		})
		if e != nil {
			return nil, e
		}
		result.Manager.constraints = 0
		current, e = sscSelectOutputs(result.Manager, order, 0, p)
		if e != nil {
			return nil, e
		}
	}
	return current, nil
}

// SSCGenerateOneHot ports the utility declaration in sscInt.h.
func SSCGenerateOneHot(variables int, p SSCParameters) (*GiaManager, error) {
	p, err := normalizeSSCParameters(p)
	if err != nil {
		return nil, err
	}
	return NewGiaOneHot(0, variables, GiaCofactorOptions{MaxObjects: p.MaxObjects, MaxMappingEntries: p.MaxObjects, Stop: p.Stop})
}
