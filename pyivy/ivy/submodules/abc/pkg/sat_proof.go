package abc

// Owned proof reconstruction from src/sat/proof/pr.c and pr.h.

import (
	"bytes"
	"os"
	"slices"
	"strconv"
	"strings"
)

type SatProofRecordingLimits struct {
	MaxVariables, MaxClauses, MaxLiterals, MaxSteps, MaxBytes int
	Stop                                                      func() bool
}

func normalizeSatProofRecordingLimits(l SatProofRecordingLimits) (SatProofRecordingLimits, error) {
	if l.MaxVariables < 0 || l.MaxClauses < 0 || l.MaxLiterals < 0 || l.MaxSteps < 0 || l.MaxBytes < 0 {
		return l, ErrDimension
	}
	if l.MaxVariables == 0 {
		l.MaxVariables = 1 << 24
	}
	if l.MaxClauses == 0 {
		l.MaxClauses = 1 << 24
	}
	if l.MaxLiterals == 0 {
		l.MaxLiterals = 1 << 27
	}
	if l.MaxSteps == 0 {
		l.MaxSteps = 1 << 27
	}
	if l.MaxBytes == 0 {
		l.MaxBytes = 1 << 30
	}
	return l, nil
}

type SatProofRecordingClause struct {
	Literals    []SatLit
	Root, PartA bool
	ProofID     int
	Interpolant uint32
}

type SatProofRecordingStep struct{ ID, Left, Right int }

type SatProofRecordingStats struct {
	Variables, Roots, Clauses, PartA, Learned, Literals, ResolutionSteps int
}

type SatProofRecordingResult struct {
	Clauses      []SatProofRecordingClause
	Steps        []SatProofRecordingStep
	DerivedEmpty bool
	FinalProofID int
	GlobalVars   []SatVar
	Interpolant  uint32
}

type SatProofRecorder struct {
	limits    SatProofRecordingLimits
	clauses   []SatProofRecordingClause
	variables int
	literals  int
	closed    bool
}

func NewSatProofRecorder(initialVariables int, limits SatProofRecordingLimits) (*SatProofRecorder, error) {
	limits, err := normalizeSatProofRecordingLimits(limits)
	if err != nil || initialVariables < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if initialVariables > limits.MaxVariables {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &SatProofRecorder{limits: limits, variables: initialVariables}, nil
}

func (r *SatProofRecorder) check() error {
	if r == nil || r.closed {
		return ErrInvariant
	}
	return nil
}

func cloneSatProofRecordingClauses(in []SatProofRecordingClause) []SatProofRecordingClause {
	out := make([]SatProofRecordingClause, len(in))
	for i := range in {
		out[i] = in[i]
		out[i].Literals = slices.Clone(in[i].Literals)
	}
	return out
}

func (r *SatProofRecorder) AddClause(literals []SatLit, root, partA bool) (int, error) {
	if err := r.check(); err != nil {
		return -1, err
	}
	if partA && !root {
		return -1, ErrDimension
	}
	if len(r.clauses) >= r.limits.MaxClauses || len(literals) > r.limits.MaxLiterals-r.literals {
		return -1, ErrLimit
	}
	if r.limits.Stop != nil && r.limits.Stop() {
		return -1, ErrCancelled
	}
	owned := slices.Clone(literals)
	for _, lit := range owned {
		if !lit.Valid() || uint64(lit.Var()) >= uint64(r.limits.MaxVariables) {
			return -1, ErrInvalidLiteral
		}
	}
	slices.Sort(owned)
	for i := 1; i < len(owned); i++ {
		if owned[i-1].Var() == owned[i].Var() {
			return -1, ErrInvalidLiteral
		}
	}
	maxVar := r.variables
	for _, lit := range owned {
		if int(lit.Var())+1 > maxVar {
			maxVar = int(lit.Var()) + 1
		}
	}
	if maxVar > r.limits.MaxVariables {
		return -1, ErrLimit
	}
	id := len(r.clauses)
	r.clauses = append(r.clauses, SatProofRecordingClause{Literals: owned, Root: root, PartA: partA})
	r.variables, r.literals = maxVar, r.literals+len(owned)
	return id, nil
}

func (r *SatProofRecorder) Clauses() ([]SatProofRecordingClause, error) {
	if err := r.check(); err != nil {
		return nil, err
	}
	return cloneSatProofRecordingClauses(r.clauses), nil
}

func (r *SatProofRecorder) Stats() (SatProofRecordingStats, error) {
	if err := r.check(); err != nil {
		return SatProofRecordingStats{}, err
	}
	s := SatProofRecordingStats{Variables: r.variables, Clauses: len(r.clauses), Literals: r.literals}
	for _, c := range r.clauses {
		if c.Root {
			s.Roots++
		} else {
			s.Learned++
		}
		if c.PartA {
			s.PartA++
		}
	}
	return s, nil
}

func satProofLitTrue(l SatLit, assignments []int8) bool {
	v := assignments[int(l.Var())]
	return v >= 0 && (v == 1) != l.Negative()
}

func satProofLitFalse(l SatLit, assignments []int8) bool {
	v := assignments[int(l.Var())]
	return v >= 0 && (v == 1) == l.Negative()
}

type satProofBCP struct {
	assigns []int8
	reasons []int
	trail   []SatLit
	watches [][2]int
	lists   [][]int
	clauses []SatProofRecordingClause
}

func (b *satProofBCP) enqueue(l SatLit, reason int) bool {
	v := int(l.Var())
	want := int8(1)
	if l.Negative() {
		want = 0
	}
	if b.assigns[v] >= 0 {
		return b.assigns[v] == want
	}
	b.assigns[v], b.reasons[v] = want, reason
	b.trail = append(b.trail, l)
	return true
}

func (b *satProofBCP) cancel(level int) {
	for i := len(b.trail) - 1; i >= level; i-- {
		v := int(b.trail[i].Var())
		b.assigns[v], b.reasons[v] = -1, -1
	}
	b.trail = b.trail[:level]
}

func newSatProofBCP(variables int, clauses []SatProofRecordingClause) *satProofBCP {
	b := &satProofBCP{
		assigns: make([]int8, variables), reasons: make([]int, variables),
		watches: make([][2]int, len(clauses)), lists: make([][]int, 2*variables), clauses: clauses,
	}
	for i := range b.assigns {
		b.assigns[i], b.reasons[i] = -1, -1
	}
	for i := range b.watches {
		b.watches[i] = [2]int{-1, -1}
	}
	return b
}

func (b *satProofBCP) watchClause(index int) error {
	if index < 0 || index >= len(b.clauses) || len(b.clauses[index].Literals) < 2 || b.watches[index][0] >= 0 {
		return ErrInvariant
	}
	b.watches[index] = [2]int{0, 1}
	for side := 0; side < 2; side++ {
		key := int(b.clauses[index].Literals[side].Not())
		b.lists[key] = append(b.lists[key], index)
	}
	return nil
}

// propagate preserves pr.c's two-watched-literal behavior. Watch lists are
// indexed by the assignment that falsifies the watched literal.
func (b *satProofBCP) propagate(start int, stop func() bool) (int, error) {
	if start < 0 || start > len(b.trail) {
		return -1, ErrDimension
	}
	for position := start; position < len(b.trail); position++ {
		if stop != nil && stop() {
			return -1, ErrCancelled
		}
		key := int(b.trail[position])
		old := b.lists[key]
		kept := old[:0]
		for offset, clauseIndex := range old {
			if stop != nil && stop() {
				b.lists[key] = append(kept, old[offset:]...)
				return -1, ErrCancelled
			}
			clause := b.clauses[clauseIndex].Literals
			watched := b.watches[clauseIndex]
			falseSide := -1
			for side := 0; side < 2; side++ {
				if clause[watched[side]].Not() == b.trail[position] {
					falseSide = side
					break
				}
			}
			if falseSide < 0 {
				return -1, ErrInvariant
			}
			otherSide := falseSide ^ 1
			otherPosition := watched[otherSide]
			other := clause[otherPosition]
			if satProofLitTrue(other, b.assigns) {
				kept = append(kept, clauseIndex)
				continue
			}
			replacement := -1
			for candidate, literal := range clause {
				if candidate == watched[0] || candidate == watched[1] || satProofLitFalse(literal, b.assigns) {
					continue
				}
				replacement = candidate
				break
			}
			if replacement >= 0 {
				b.watches[clauseIndex][falseSide] = replacement
				newKey := int(clause[replacement].Not())
				b.lists[newKey] = append(b.lists[newKey], clauseIndex)
				continue
			}
			kept = append(kept, clauseIndex)
			if !b.enqueue(other, clauseIndex) {
				b.lists[key] = append(kept, old[offset+1:]...)
				return clauseIndex, nil
			}
		}
		b.lists[key] = kept
	}
	return -1, nil
}

func satProofResolve(first, second []SatLit, variable int) ([]SatLit, error) {
	out := make([]SatLit, 0, len(first)+len(second)-2)
	foundA, foundB := false, false
	var pivotA, pivotB SatLit
	push := func(l SatLit) error {
		for _, x := range out {
			if x.Var() == l.Var() {
				if x != l {
					return ErrInvariant
				}
				return nil
			}
		}
		out = append(out, l)
		return nil
	}
	for _, l := range first {
		if int(l.Var()) == variable {
			foundA = true
			pivotA = l
			continue
		}
		if err := push(l); err != nil {
			return nil, err
		}
	}
	for _, l := range second {
		if int(l.Var()) == variable {
			foundB = true
			pivotB = l
			continue
		}
		if err := push(l); err != nil {
			return nil, err
		}
	}
	if !foundA || !foundB || pivotA != pivotB.Not() {
		return nil, ErrInvariant
	}
	slices.Sort(out)
	return out, nil
}

func equalSatProofRecordingClause(a, b []SatLit) bool { return slices.Equal(a, b) }

func satProofTrace(clauses []SatProofRecordingClause, conflict int, target int, b *satProofBCP, varTypes []int8, steps *[]SatProofRecordingStep, counter *int, maxSteps int, stop func() bool) (int, uint32, error) {
	res := slices.Clone(clauses[conflict].Literals)
	seen := make([]bool, len(b.assigns))
	for _, l := range res {
		seen[int(l.Var())] = true
	}
	prev := clauses[conflict].ProofID
	truth := clauses[conflict].Interpolant
	for i := len(b.trail) - 1; i >= 0; i-- {
		if stop != nil && stop() {
			return 0, 0, ErrCancelled
		}
		v := int(b.trail[i].Var())
		if !seen[v] {
			continue
		}
		seen[v] = false
		reason := b.reasons[v]
		if reason < 0 {
			continue
		}
		var err error
		res, err = satProofResolve(res, clauses[reason].Literals, v)
		if err != nil {
			return 0, 0, err
		}
		if len(*steps) >= maxSteps {
			return 0, 0, ErrLimit
		}
		*counter++
		id := *counter
		*steps = append(*steps, SatProofRecordingStep{ID: id, Left: prev, Right: clauses[reason].ProofID})
		prev = id
		if len(varTypes) > v {
			if varTypes[v] == 1 {
				truth |= clauses[reason].Interpolant
			} else {
				truth &= clauses[reason].Interpolant
			}
		}
		for _, l := range clauses[reason].Literals {
			if int(l.Var()) != v {
				seen[int(l.Var())] = true
			}
		}
	}
	if !equalSatProofRecordingClause(res, clauses[target].Literals) {
		return 0, 0, ErrInvariant
	}
	return prev, truth, nil
}

func satProofPrepareInterpolants(clauses []SatProofRecordingClause, variables, roots int) ([]int8, []SatVar, error) {
	if roots == 0 {
		return nil, nil, nil
	}
	hasPartA := false
	for i := 0; i < roots; i++ {
		hasPartA = hasPartA || clauses[i].PartA
	}
	if !hasPartA {
		return nil, nil, nil
	}
	types := make([]int8, variables)
	inB := make([]bool, variables)
	hasA := false
	for i := 0; i < roots; i++ {
		hasA = hasA || clauses[i].PartA
		for _, l := range clauses[i].Literals {
			if clauses[i].PartA {
				types[int(l.Var())] = 1
			} else {
				inB[int(l.Var())] = true
			}
		}
	}
	if !hasA {
		return nil, nil, nil
	}
	globals := make([]SatVar, 0, 5)
	for v := 0; v < variables; v++ {
		if types[v] == 1 && inB[v] {
			if len(globals) == 5 {
				return nil, nil, ErrLimit
			}
			types[v] = -int8(len(globals) + 1)
			globals = append(globals, SatVar(v))
		}
	}
	patterns := [5]uint32{0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00, 0xFFFF0000}
	for i := 0; i < roots; i++ {
		if !clauses[i].PartA {
			clauses[i].Interpolant = ^uint32(0)
			continue
		}
		var truth uint32
		for _, l := range clauses[i].Literals {
			t := types[int(l.Var())]
			if t < 0 {
				p := patterns[-t-1]
				if l.Negative() {
					p = ^p
				}
				truth |= p
			}
		}
		clauses[i].Interpolant = truth
	}
	return types, globals, nil
}

func (r *SatProofRecorder) Record() (*SatProofRecordingResult, error) {
	if err := r.check(); err != nil {
		return nil, err
	}
	if len(r.clauses) == 0 {
		return nil, ErrDimension
	}
	if r.limits.Stop != nil && r.limits.Stop() {
		return nil, ErrCancelled
	}
	clauses := cloneSatProofRecordingClauses(r.clauses)
	roots := 0
	learned := false
	partBRoot := false
	for i, c := range clauses {
		if c.Root {
			if learned {
				return nil, ErrInvariant
			}
			if c.PartA {
				if partBRoot {
					return nil, ErrInvariant
				}
			} else {
				partBRoot = true
			}
			roots++
		} else {
			learned = true
		}
		if len(c.Literals) == 0 && i != len(clauses)-1 {
			return nil, ErrInvariant
		}
	}
	if len(clauses[len(clauses)-1].Literals) != 0 || clauses[len(clauses)-1].Root {
		return nil, ErrInvariant
	}
	varTypes, globals, err := satProofPrepareInterpolants(clauses, r.variables, roots)
	if err != nil {
		return nil, err
	}
	b := newSatProofBCP(r.variables, clauses)
	for i := 0; i < roots; i++ {
		if len(clauses[i].Literals) > 1 {
			if err = b.watchClause(i); err != nil {
				return nil, err
			}
		} else if len(clauses[i].Literals) == 1 && !b.enqueue(clauses[i].Literals[0], i) {
			return nil, ErrInvariant
		}
	}
	if conflict, e := b.propagate(0, r.limits.Stop); e != nil {
		return nil, e
	} else if conflict >= 0 {
		return nil, ErrInvariant
	}
	rootLevel := len(b.trail)
	counter := 0
	for i := 0; i < roots; i++ {
		counter++
		clauses[i].ProofID = counter
	}
	steps := make([]SatProofRecordingStep, 0)
	finalID := 0
	var finalTruth uint32
	for ci := roots; ci < len(clauses)-1; ci++ {
		clause := clauses[ci]
		if len(clause.Literals) == 0 {
			return nil, ErrInvariant
		}
		for _, l := range clause.Literals {
			if !b.enqueue(l.Not(), -1) {
				return nil, ErrInvariant
			}
		}
		conflict, e := b.propagate(rootLevel, r.limits.Stop)
		if e != nil {
			return nil, e
		}
		if conflict < 0 {
			return nil, ErrInvariant
		}
		proof, truth, e := satProofTrace(clauses, conflict, ci, b, varTypes, &steps, &counter, r.limits.MaxSteps, r.limits.Stop)
		if e != nil {
			return nil, e
		}
		clauses[ci].ProofID, clauses[ci].Interpolant = proof, truth
		b.cancel(rootLevel)
		if len(clause.Literals) > 1 {
			if err = b.watchClause(ci); err != nil {
				return nil, err
			}
			continue
		}
		if !b.enqueue(clause.Literals[0], ci) {
			return nil, ErrInvariant
		}
		conflict, e = b.propagate(rootLevel, r.limits.Stop)
		if e != nil {
			return nil, e
		}
		if conflict >= 0 {
			empty := len(clauses) - 1
			proof, truth, e = satProofTrace(clauses, conflict, empty, b, varTypes, &steps, &counter, r.limits.MaxSteps, r.limits.Stop)
			if e != nil {
				return nil, e
			}
			clauses[empty].ProofID, clauses[empty].Interpolant = proof, truth
			finalID, finalTruth = proof, truth
			break
		}
		rootLevel = len(b.trail)
	}
	if finalID == 0 {
		return nil, ErrInvariant
	}
	result := &SatProofRecordingResult{Clauses: clauses, Steps: steps, DerivedEmpty: true, FinalProofID: finalID, GlobalVars: slices.Clone(globals), Interpolant: finalTruth}
	return result, nil
}

func (r *SatProofRecorder) Close() {
	if r != nil {
		r.closed = true
		r.clauses = nil
		r.variables = 0
		r.literals = 0
	}
}

func (result *SatProofRecordingResult) MarshalText(maxBytes int) ([]byte, error) {
	if result == nil || !result.DerivedEmpty {
		return nil, ErrInvariant
	}
	if maxBytes < 0 {
		return nil, ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = 1 << 30
	}
	var out strings.Builder
	for _, c := range result.Clauses {
		if !c.Root {
			continue
		}
		out.WriteString(strconv.Itoa(c.ProofID))
		for _, l := range c.Literals {
			out.WriteByte(' ')
			v := int(l.Var()) + 1
			if l.Negative() {
				v = -v
			}
			out.WriteString(strconv.Itoa(v))
		}
		out.WriteString(" 0 0\n")
		if out.Len() > maxBytes {
			return nil, ErrLimit
		}
	}
	for _, s := range result.Steps {
		out.WriteString(strconv.Itoa(s.ID))
		out.WriteString(" * ")
		out.WriteString(strconv.Itoa(s.Left))
		out.WriteByte(' ')
		out.WriteString(strconv.Itoa(s.Right))
		out.WriteString(" 0\n")
		if out.Len() > maxBytes {
			return nil, ErrLimit
		}
	}
	return []byte(out.String()), nil
}

func ParseSatProofRecording(data []byte, limits SatProofRecordingLimits) (*SatProofRecorder, error) {
	limits, err := normalizeSatProofRecordingLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	lines := bytes.Split(data, []byte{'\n'})
	var recorder *SatProofRecorder
	expected, roots, partA := -1, 0, 0
	counter := 0
	for _, raw := range lines {
		line := strings.TrimSpace(string(raw))
		if line == "" || strings.HasPrefix(line, "c") {
			continue
		}
		fields := strings.Fields(line)
		if fields[0] == "p" {
			if recorder != nil || len(fields) < 4 || len(fields) > 5 {
				return nil, ErrFormat
			}
			vars, e1 := strconv.Atoi(fields[1])
			n, e2 := strconv.Atoi(fields[2])
			rr, e3 := strconv.Atoi(fields[3])
			aa := 0
			var e4 error
			if len(fields) == 5 {
				aa, e4 = strconv.Atoi(fields[4])
			}
			if e1 != nil || e2 != nil || e3 != nil || e4 != nil || vars < 0 || n < 1 || rr < 0 || rr > n || aa < 0 || aa > rr {
				return nil, ErrFormat
			}
			recorder, err = NewSatProofRecorder(vars, limits)
			if err != nil {
				return nil, err
			}
			expected, roots, partA = n, rr, aa
			continue
		}
		if recorder == nil {
			return nil, ErrFormat
		}
		lits := make([]SatLit, 0, len(fields))
		terminated := false
		for _, field := range fields {
			v, e := strconv.Atoi(field)
			if e != nil {
				return nil, ErrFormat
			}
			if v == 0 {
				terminated = true
				break
			}
			if v == mathMinInt() {
				return nil, ErrFormat
			}
			variable := v
			if variable < 0 {
				variable = -variable
			}
			lit, e := MakeSatLit(SatVar(variable-1), v < 0)
			if e != nil {
				return nil, e
			}
			lits = append(lits, lit)
		}
		if !terminated {
			return nil, ErrFormat
		}
		if _, err = recorder.AddClause(lits, counter < roots, counter < partA); err != nil {
			return nil, err
		}
		counter++
	}
	if recorder == nil || counter != expected {
		return nil, ErrFormat
	}
	return recorder, nil
}

func ReadSatProofRecordingFile(path string, limits SatProofRecordingLimits) (*SatProofRecorder, error) {
	limits, err := normalizeSatProofRecordingLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > int64(limits.MaxBytes) {
		return nil, ErrLimit
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseSatProofRecording(data, limits)
}
