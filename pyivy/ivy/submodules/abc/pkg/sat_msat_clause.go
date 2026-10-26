package abc

// Clause construction and behavior from src/sat/msat/msatClause.c.

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

type MsatClauseCreateStatus uint8

const (
	MsatClauseConflict MsatClauseCreateStatus = iota
	MsatClauseSatisfied
	MsatClauseUnit
	MsatClauseStored
)

type MsatClausePropagation struct {
	Literal     SatLit
	Unit        bool
	Satisfied   bool
	WatchMoved  bool
	WatchBucket SatLit
}

type MsatClauseContext struct {
	assignments   []int
	levels        []int
	reasons       []*MsatClause
	watched       []*MsatClauseVector
	activity      *MsatActivityState
	nextNumber    int
	decisionLevel int
	maximum       int
	closed        bool
}

func NewMsatClauseContext(variables, maximum int, activity *MsatActivityState) (*MsatClauseContext, error) {
	maximum, err := msatVectorMaximum(maximum)
	if err != nil || variables < 0 || activity == nil || activity.checkLive() != nil || len(activity.order.activities) != variables {
		return nil, ErrDimension
	}
	assignments := make([]int, variables)
	levels := make([]int, variables)
	for index := range assignments {
		assignments[index], levels[index] = -1, -1
	}
	watched := make([]*MsatClauseVector, 2*variables)
	for index := range watched {
		watched[index], err = NewMsatClauseVector(0, maximum)
		if err != nil {
			return nil, err
		}
	}
	return &MsatClauseContext{
		assignments: assignments,
		levels:      levels,
		reasons:     make([]*MsatClause, variables),
		watched:     watched,
		activity:    activity,
		maximum:     maximum,
	}, nil
}

func (c *MsatClauseContext) checkLive() error {
	if c == nil {
		return ErrDimension
	}
	if c.closed || c.maximum <= 0 || c.activity == nil || c.activity.checkLive() != nil ||
		len(c.assignments) != len(c.levels) || len(c.assignments) != len(c.reasons) || len(c.watched) != 2*len(c.assignments) {
		return ErrInvariant
	}
	return nil
}

func (c *MsatClauseContext) SetDecisionLevel(level int) error {
	if err := c.checkLive(); err != nil {
		return err
	}
	if level < 0 {
		return ErrDimension
	}
	c.decisionLevel = level
	return nil
}

func (c *MsatClauseContext) SetAssignment(variable SatVar, literal *SatLit, level int, reason *MsatClause) error {
	if err := c.checkLive(); err != nil {
		return err
	}
	if uint64(variable) >= uint64(len(c.assignments)) || level < -1 {
		return ErrDimension
	}
	if literal == nil {
		c.assignments[variable], c.levels[variable], c.reasons[variable] = -1, -1, nil
		return c.activity.order.Unassign(variable)
	}
	if !literal.Valid() || literal.Var() != variable || level < 0 || reason != nil && reason.closed {
		return ErrInvalidLiteral
	}
	c.assignments[variable], c.levels[variable], c.reasons[variable] = int(*literal), level, reason
	return c.activity.order.Assign(variable)
}

func (c *MsatClauseContext) enqueue(literal SatLit) MsatClauseCreateStatus {
	variable := literal.Var()
	if c.assignments[variable] == -1 {
		c.assignments[variable] = int(literal)
		c.levels[variable] = c.decisionLevel
		c.reasons[variable] = nil
		_ = c.activity.order.Assign(variable)
		return MsatClauseUnit
	}
	if c.assignments[variable] == int(literal) {
		return MsatClauseSatisfied
	}
	return MsatClauseConflict
}

func (c *MsatClauseContext) validateLiterals(literals []SatLit) error {
	if len(literals) > (1<<14)-1 {
		return ErrLimit
	}
	for _, literal := range literals {
		if !literal.Valid() || uint64(literal.Var()) >= uint64(len(c.assignments)) {
			return ErrInvalidLiteral
		}
	}
	return nil
}

// simplifyOriginal applies correct SAT duplicate/tautology semantics. The C
// source explicitly marks its opposite-polarity handling as buggy.
func (c *MsatClauseContext) simplifyOriginal(literals []SatLit) ([]SatLit, MsatClauseCreateStatus) {
	work := slices.Clone(literals)
	slices.Sort(work)
	result := work[:0]
	for index := 0; index < len(work); {
		literal := work[index]
		variable := literal.Var()
		havePositive, haveNegative := false, false
		for index < len(work) && work[index].Var() == variable {
			if work[index].Negative() {
				haveNegative = true
			} else {
				havePositive = true
			}
			index++
		}
		if havePositive && haveNegative {
			return nil, MsatClauseSatisfied
		}
		if haveNegative {
			literal, _ = MakeSatLit(variable, true)
		} else {
			literal, _ = MakeSatLit(variable, false)
		}
		assignment := c.assignments[variable]
		if assignment == int(literal) {
			return nil, MsatClauseSatisfied
		}
		if assignment == int(literal.Not()) {
			continue
		}
		result = append(result, literal)
	}
	return slices.Clone(result), MsatClauseStored
}

func (c *MsatClauseContext) canWatch(literal SatLit) bool {
	vector := c.watched[literal]
	return vector != nil && !vector.closed && vector.size < vector.maximum
}

func (c *MsatClauseContext) activityRollback() func() {
	order := c.activity.order
	activities := slices.Clone(order.activities)
	index := slices.Clone(order.index)
	heap := slices.Clone(order.heap)
	config := c.activity.config
	clauseActivities := make([]float32, c.activity.learned.size)
	for position, clause := range c.activity.learned.values[:c.activity.learned.size] {
		if clause != nil {
			clauseActivities[position] = clause.activity
		}
	}
	return func() {
		copy(order.activities, activities)
		order.index, order.heap = index, heap
		c.activity.config = config
		for position, clause := range c.activity.learned.values[:c.activity.learned.size] {
			if clause != nil {
				clause.activity = clauseActivities[position]
			}
		}
	}
}

// Create constructs and watches a clause but does not append it to a problem or
// learned database, matching Msat_ClauseCreate's caller-owned database step.
func (c *MsatClauseContext) Create(literals []SatLit, learned bool) (MsatClauseCreateStatus, *MsatClause, error) {
	if err := c.checkLive(); err != nil {
		return MsatClauseConflict, nil, err
	}
	if err := c.validateLiterals(literals); err != nil {
		return MsatClauseConflict, nil, err
	}
	work := slices.Clone(literals)
	if !learned {
		if c.decisionLevel != 0 {
			return MsatClauseConflict, nil, ErrInvariant
		}
		var status MsatClauseCreateStatus
		work, status = c.simplifyOriginal(work)
		if status == MsatClauseSatisfied {
			return status, nil, nil
		}
	}
	if len(work) == 0 {
		return MsatClauseConflict, nil, nil
	}
	if len(work) == 1 {
		return c.enqueue(work[0]), nil, nil
	}
	if c.nextNumber >= c.maximum {
		return MsatClauseConflict, nil, ErrLimit
	}
	if learned {
		best := 1
		if c.levels[work[best].Var()] < 0 {
			return MsatClauseConflict, nil, ErrInvariant
		}
		for index := 2; index < len(work); index++ {
			level := c.levels[work[index].Var()]
			if level < 0 {
				return MsatClauseConflict, nil, ErrInvariant
			}
			if c.levels[work[best].Var()] < level {
				best = index
			}
		}
		work[1], work[best] = work[best], work[1]
	}
	watch0, watch1 := work[0].Not(), work[1].Not()
	if !c.canWatch(watch0) || !c.canWatch(watch1) || watch0 == watch1 && c.watched[watch0].size+1 >= c.watched[watch0].maximum {
		return MsatClauseConflict, nil, ErrLimit
	}
	clause := &MsatClause{number: c.nextNumber, learned: learned, literals: slices.Clone(work)}
	rollback := func() {}
	if learned {
		rollback = c.activityRollback()
		if err := c.activity.ClauseBump(clause); err != nil {
			rollback()
			return MsatClauseConflict, nil, err
		}
		for _, literal := range clause.literals {
			if err := c.activity.VariableBump(literal); err != nil {
				rollback()
				return MsatClauseConflict, nil, err
			}
		}
	}
	if err := c.watched[watch0].Push(clause); err != nil {
		rollback()
		return MsatClauseConflict, nil, err
	}
	if err := c.watched[watch1].Push(clause); err != nil {
		_, _ = c.watched[watch0].Pop()
		rollback()
		return MsatClauseConflict, nil, err
	}
	c.nextNumber++
	return MsatClauseStored, clause, nil
}

func (c *MsatClauseContext) Assignments() []int {
	if c == nil || c.closed {
		return nil
	}
	return slices.Clone(c.assignments)
}

func (c *MsatClauseContext) Watch(literal SatLit) ([]*MsatClause, error) {
	if err := c.checkLive(); err != nil {
		return nil, err
	}
	if !literal.Valid() || int(literal) >= len(c.watched) {
		return nil, ErrInvalidLiteral
	}
	return c.watched[literal].Values(), nil
}

func (clause *MsatClause) checkLive() error {
	if clause == nil {
		return ErrDimension
	}
	if clause.closed || len(clause.literals) == 0 {
		return ErrInvariant
	}
	return nil
}

func (clause *MsatClause) Learned() bool { return clause != nil && !clause.closed && clause.learned }
func (clause *MsatClause) Len() int {
	if clause == nil || clause.closed {
		return 0
	}
	return len(clause.literals)
}
func (clause *MsatClause) Literals() []SatLit {
	if clause == nil || clause.closed {
		return nil
	}
	return slices.Clone(clause.literals)
}
func (clause *MsatClause) Marked() bool { return clause != nil && !clause.closed && clause.marked }
func (clause *MsatClause) SetMarked(marked bool) error {
	if err := clause.checkLive(); err != nil {
		return err
	}
	clause.marked = marked
	return nil
}
func (clause *MsatClause) Number() (int, error) {
	if err := clause.checkLive(); err != nil {
		return 0, err
	}
	return clause.number, nil
}
func (clause *MsatClause) SetNumber(number int) error {
	if err := clause.checkLive(); err != nil {
		return err
	}
	if number < 0 {
		return ErrDimension
	}
	clause.number = number
	return nil
}
func (clause *MsatClause) TypeA() bool { return clause != nil && !clause.closed && clause.typeA }
func (clause *MsatClause) SetTypeA(typeA bool) error {
	if err := clause.checkLive(); err != nil {
		return err
	}
	clause.typeA = typeA
	return nil
}
func (clause *MsatClause) Activity() (float32, error) {
	if err := clause.checkLive(); err != nil {
		return 0, err
	}
	return clause.activity, nil
}

func (clause *MsatClause) SetActivity(activity float32) error {
	if err := clause.checkLive(); err != nil {
		return err
	}
	if math.IsNaN(float64(activity)) || math.IsInf(float64(activity), 0) {
		return ErrDimension
	}
	clause.activity = activity
	return nil
}

func (clause *MsatClause) IsLocked(reasons []*MsatClause) (bool, error) {
	if err := clause.checkLive(); err != nil {
		return false, err
	}
	variable := clause.literals[0].Var()
	if uint64(variable) >= uint64(len(reasons)) {
		return false, ErrDimension
	}
	return reasons[variable] == clause, nil
}

func validateMsatAssignments(assignments []int, variables int) error {
	if len(assignments) < variables {
		return ErrDimension
	}
	for variable := 0; variable < variables; variable++ {
		assignment := assignments[variable]
		if assignment != -1 && (assignment < 0 || assignment>>1 != variable) {
			return ErrInvariant
		}
	}
	return nil
}

func (clause *MsatClause) Propagate(trigger SatLit, assignments []int) (MsatClausePropagation, error) {
	if err := clause.checkLive(); err != nil {
		return MsatClausePropagation{}, err
	}
	if len(clause.literals) < 2 || !trigger.Valid() {
		return MsatClausePropagation{}, ErrDimension
	}
	variables := 0
	for _, literal := range clause.literals {
		variables = max(variables, int(literal.Var())+1)
	}
	if err := validateMsatAssignments(assignments, variables); err != nil {
		return MsatClausePropagation{}, err
	}
	falseLiteral := trigger.Not()
	if clause.literals[0] == falseLiteral {
		clause.literals[0], clause.literals[1] = clause.literals[1], falseLiteral
	}
	if clause.literals[1] != falseLiteral {
		return MsatClausePropagation{}, ErrInvariant
	}
	first := clause.literals[0]
	if assignments[first.Var()] == int(first) {
		return MsatClausePropagation{Satisfied: true}, nil
	}
	for index := 2; index < len(clause.literals); index++ {
		literal := clause.literals[index]
		if assignments[literal.Var()] != int(literal.Not()) {
			clause.literals[1], clause.literals[index] = literal, falseLiteral
			return MsatClausePropagation{WatchMoved: true, WatchBucket: literal.Not()}, nil
		}
	}
	return MsatClausePropagation{Literal: first, Unit: true}, nil
}

func (clause *MsatClause) Simplify(assignments []int) (bool, error) {
	if err := clause.checkLive(); err != nil {
		return false, err
	}
	variables := 0
	for _, literal := range clause.literals {
		variables = max(variables, int(literal.Var())+1)
	}
	if err := validateMsatAssignments(assignments, variables); err != nil {
		return false, err
	}
	result := make([]SatLit, 0, len(clause.literals))
	for index, literal := range clause.literals {
		assignment := assignments[literal.Var()]
		if assignment == -1 {
			result = append(result, literal)
			continue
		}
		if assignment == int(literal) {
			return true, nil
		}
		if index < 2 {
			return false, ErrInvariant
		}
	}
	clause.literals = result
	return false, nil
}

func (clause *MsatClause) Reason(literal *SatLit, assignments []int, activity *MsatActivityState) ([]SatLit, error) {
	if err := clause.checkLive(); err != nil {
		return nil, err
	}
	start := 0
	if literal != nil {
		if *literal != clause.literals[0] {
			return nil, ErrInvalidLiteral
		}
		start = 1
	}
	variables := 0
	for _, item := range clause.literals {
		variables = max(variables, int(item.Var())+1)
	}
	if err := validateMsatAssignments(assignments, variables); err != nil {
		return nil, err
	}
	result := make([]SatLit, 0, len(clause.literals)-start)
	for _, item := range clause.literals[start:] {
		if assignments[item.Var()] != int(item.Not()) {
			return nil, ErrInvariant
		}
		result = append(result, item.Not())
	}
	if clause.learned {
		if activity == nil {
			return nil, ErrDimension
		}
		if err := activity.ClauseBump(clause); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func RemoveMsatClauseWatch(vector *MsatClauseVector, clause *MsatClause) error {
	if vector == nil || clause == nil {
		return ErrDimension
	}
	if err := vector.checkLive(); err != nil {
		return err
	}
	for index := 0; index < vector.size; index++ {
		if vector.values[index] == clause {
			copy(vector.values[index:], vector.values[index+1:vector.size])
			vector.size--
			vector.values[vector.size] = nil
			return nil
		}
	}
	return ErrInvariant
}

func (c *MsatClauseContext) Free(clause *MsatClause, removeWatched bool) error {
	if err := c.checkLive(); err != nil {
		return err
	}
	if err := clause.checkLive(); err != nil {
		return err
	}
	if removeWatched {
		if len(clause.literals) < 2 {
			return ErrInvariant
		}
		first, second := clause.literals[0].Not(), clause.literals[1].Not()
		count := func(bucket SatLit) int {
			found := 0
			for _, candidate := range c.watched[bucket].values[:c.watched[bucket].size] {
				if candidate == clause {
					found++
				}
			}
			return found
		}
		if first == second {
			if count(first) < 2 {
				return ErrInvariant
			}
		} else if count(first) < 1 || count(second) < 1 {
			return ErrInvariant
		}
		for _, bucket := range []SatLit{first, second} {
			_ = RemoveMsatClauseWatch(c.watched[bucket], clause)
		}
	}
	clear(clause.literals)
	clause.literals = nil
	clause.closed = true
	return nil
}

func FormatMsatClause(clause *MsatClause) string {
	if clause == nil {
		return "NULL pointer\n"
	}
	var builder strings.Builder
	if clause.learned {
		fmt.Fprintf(&builder, "Act = %.4f  ", clause.activity)
	}
	for _, literal := range clause.literals {
		if literal.Negative() {
			fmt.Fprintf(&builder, " -%d", literal.Var()+1)
		} else {
			fmt.Fprintf(&builder, " %d", literal.Var()+1)
		}
	}
	builder.WriteByte('\n')
	return builder.String()
}

func FormatMsatClauseSymbols(clause *MsatClause) string {
	if clause == nil {
		return "NULL pointer\n"
	}
	var builder strings.Builder
	for _, literal := range clause.literals {
		if literal.Negative() {
			fmt.Fprintf(&builder, " -%d", literal.Var()+1)
		} else {
			fmt.Fprintf(&builder, " %d", literal.Var()+1)
		}
	}
	builder.WriteByte('\n')
	return builder.String()
}

func (clause *MsatClause) DIMACS(increment bool) (string, error) {
	if err := clause.checkLive(); err != nil {
		return "", err
	}
	var builder strings.Builder
	for _, literal := range clause.literals {
		value := int(literal.Var())
		if increment {
			value++
		}
		if literal.Negative() {
			fmt.Fprintf(&builder, "-%d ", value)
		} else {
			fmt.Fprintf(&builder, "%d ", value)
		}
	}
	if increment {
		builder.WriteByte('0')
	}
	builder.WriteByte('\n')
	return builder.String(), nil
}

func (c *MsatClauseContext) Close() {
	if c == nil || c.closed {
		return
	}
	for _, vector := range c.watched {
		vector.Close()
	}
	clear(c.assignments)
	clear(c.levels)
	clear(c.reasons)
	c.assignments, c.levels, c.reasons, c.watched = nil, nil, nil, nil
	c.activity = nil
	c.closed = true
}
