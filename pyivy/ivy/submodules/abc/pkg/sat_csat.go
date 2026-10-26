package abc

// CSAT compatibility layer from src/sat/csat/csat_apis.h and csat_apis.c.

import (
	"os"
	"slices"
	"strconv"
	"strings"
)

type CSATGateType uint8

const (
	CSATConst CSATGateType = iota
	CSATBPI
	CSATBPPI
	CSATBAND
	CSATBNAND
	CSATBOR
	CSATBNOR
	CSATBXOR
	CSATBXNOR
	CSATBINV
	CSATBBUF
	CSATBMUX
	CSATBDFF
	CSATBSDFF
	CSATBTRIH
	CSATBTRIL
	CSATBBUS
	CSATBPPO
	CSATBPO
	CSATBCNF
	CSATBDC
)

type CSATStatus uint8

const (
	CSATUndetermined CSATStatus = iota
	CSATUnsatisfiable
	CSATSatisfiable
	CSATTimeOut
	CSATFrameOut
	CSATNoTarget
	CSATAborted
	CSATSequentialSatisfiable
)

type CSATCaller uint8

const (
	CSATCallerBLS CSATCaller = iota
	CSATCallerSatori
	CSATCallerNone
)

type CSATSolveOption uint8

const (
	CSATBaseline CSATSolveOption = iota
	CSATImplicitLearning
	CSATExplicitLearning
)

type CSATTargetResult struct {
	Status                         CSATStatus
	Decisions, Implications        int
	ConflictGates, ConflictSignals int
	Seconds                        float64
	Names                          []string
	Values                         []int
}

func (r CSATTargetResult) Clone() CSATTargetResult {
	r.Names = slices.Clone(r.Names)
	r.Values = slices.Clone(r.Values)
	return r
}

type CSATLimits struct {
	MaxGates, MaxFanins, MaxTargets, MaxNameBytes int
	Stop                                          func() bool
}

type csatGate struct {
	type_    CSATGateType
	name     string
	fanins   []int
	variable SatVar
	dcAttr   int
}

type csatTarget struct {
	gates  []int
	values []bool
}

type CSATManager struct {
	limits                     CSATLimits
	gates                      []csatGate
	byName                     map[string]int
	inputs, outputs, internals []int
	fanouts                    []int
	target                     csatTarget
	result                     *CSATTargetResult
	clauses                    [][]SatLit
	ready, finalized, closed   bool
	coreOnly                   bool
	option                     CSATSolveOption
	caller                     CSATCaller
	dumpPath                   string
	solveBacktracks            uint64
	totalBacktrackLimit        uint64
	totalInspectLimit          uint64
	totalBacktracks            uint64
	totalInspects              uint64
	faninCount, nameBytes      int
}

func NewCSATManager(limits CSATLimits) (*CSATManager, error) {
	if limits.MaxGates < 0 || limits.MaxFanins < 0 || limits.MaxTargets < 0 || limits.MaxNameBytes < 0 {
		return nil, ErrDimension
	}
	if limits.MaxGates == 0 {
		limits.MaxGates = 1 << 20
	}
	if limits.MaxFanins == 0 {
		limits.MaxFanins = 1 << 24
	}
	if limits.MaxTargets == 0 {
		limits.MaxTargets = 1 << 20
	}
	if limits.MaxNameBytes == 0 {
		limits.MaxNameBytes = 1 << 28
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &CSATManager{limits: limits, byName: make(map[string]int), option: CSATImplicitLearning, caller: CSATCallerNone}, nil
}

func (m *CSATManager) check() error {
	if m == nil || m.closed || m.byName == nil {
		return ErrInvariant
	}
	return nil
}

func (m *CSATManager) stopped() bool { return m.limits.Stop != nil && m.limits.Stop() }

func (m *CSATManager) Close() error {
	if m == nil || m.closed {
		return nil
	}
	m.closed = true
	m.gates = nil
	m.byName = nil
	m.inputs = nil
	m.outputs = nil
	m.internals = nil
	m.fanouts = nil
	m.target = csatTarget{}
	m.result = nil
	m.clauses = nil
	return nil
}

func (m *CSATManager) SetSolveOption(option CSATSolveOption) error {
	if err := m.check(); err != nil {
		return err
	}
	if option > CSATExplicitLearning {
		return ErrDimension
	}
	m.option = option
	return nil
}

func (m *CSATManager) UseOnlyCoreSatSolver() error {
	if err := m.check(); err != nil {
		return err
	}
	m.coreOnly = true
	return nil
}

func (m *CSATManager) SetCaller(caller CSATCaller) error {
	if err := m.check(); err != nil {
		return err
	}
	if caller > CSATCallerNone {
		return ErrDimension
	}
	m.caller = caller
	return nil
}

func csatNameValid(name string) bool {
	return name != "" && !strings.ContainsAny(name, "\x00\r\n")
}

func csatSupportedGate(type_ CSATGateType) bool {
	return type_ <= CSATBBUF || type_ == CSATBPPO || type_ == CSATBPO
}

func csatGateArity(type_ CSATGateType, count int) bool {
	switch type_ {
	case CSATBPI, CSATBPPI, CSATConst:
		return count == 0
	case CSATBAND, CSATBNAND, CSATBOR, CSATBNOR:
		return count >= 1
	case CSATBXOR, CSATBXNOR:
		return count >= 1 && count <= 2
	case CSATBINV, CSATBBUF, CSATBPPO, CSATBPO:
		return count == 1
	default:
		return false
	}
}

func (m *CSATManager) AddGate(type_ CSATGateType, name string, fanins []string, dcAttr int) error {
	if err := m.check(); err != nil {
		return err
	}
	if !csatSupportedGate(type_) {
		return ErrUnsupported
	}
	if m.stopped() {
		return ErrCancelled
	}
	if !csatNameValid(name) || !csatGateArity(type_, len(fanins)) {
		return ErrDimension
	}
	if _, exists := m.byName[name]; exists {
		return ErrInvariant
	}
	if len(m.gates) == m.limits.MaxGates || len(fanins) > m.limits.MaxFanins-m.faninCount {
		return ErrLimit
	}
	if len(name) > m.limits.MaxNameBytes-m.nameBytes {
		return ErrLimit
	}
	ids := make([]int, len(fanins))
	for i, fanin := range fanins {
		if m.stopped() {
			return ErrCancelled
		}
		id, found := m.byName[fanin]
		if !found {
			return ErrInvalidLiteral
		}
		ids[i] = id
	}
	id := len(m.gates)
	gate := csatGate{type_: type_, name: name, fanins: ids, variable: SatVar(id), dcAttr: dcAttr}
	m.gates = append(m.gates, gate)
	m.fanouts = append(m.fanouts, 0)
	m.byName[name] = id
	m.faninCount += len(ids)
	m.nameBytes += len(name)
	for _, fanin := range ids {
		m.fanouts[fanin]++
	}
	switch type_ {
	case CSATBPI, CSATBPPI:
		m.inputs = append(m.inputs, id)
	case CSATBPPO, CSATBPO:
		m.outputs = append(m.outputs, id)
	default:
		m.internals = append(m.internals, id)
	}
	m.ready = false
	m.finalized = false
	m.result = nil
	return nil
}

func (m *CSATManager) GateCount() int {
	if m == nil || m.closed {
		return 0
	}
	return len(m.gates)
}

func (m *CSATManager) CheckIntegrity() (bool, error) {
	if err := m.check(); err != nil {
		return false, err
	}
	for index, id := range m.internals {
		if m.stopped() {
			return false, ErrCancelled
		}
		if index != 0 && m.fanouts[id] == 0 {
			return false, nil
		}
		for _, fanin := range m.gates[id].fanins {
			if fanin < 0 || fanin >= id {
				return false, ErrInvariant
			}
		}
	}
	return true, nil
}

func (m *CSATManager) Finalize() error {
	if err := m.check(); err != nil {
		return err
	}
	if m.stopped() {
		return ErrCancelled
	}
	m.finalized = true
	return nil
}

// The source accepts these three settings but leaves them unimplemented.
func (m *CSATManager) SetTimeLimit(runtime int) error           { return m.noopNonnegative(runtime) }
func (m *CSATManager) SetLearnLimit(limit int) error            { return m.noopNonnegative(limit) }
func (m *CSATManager) SetLearnBacktrackLimit(limit int) error   { return m.noopNonnegative(limit) }
func (m *CSATManager) SetSolveImplicationLimit(limit int) error { return m.noopNonnegative(limit) }

func (m *CSATManager) noopNonnegative(value int) error {
	if err := m.check(); err != nil {
		return err
	}
	if value < 0 {
		return ErrDimension
	}
	return nil
}

func (m *CSATManager) SetSolveBacktrackLimit(limit int) error {
	if err := m.noopNonnegative(limit); err != nil {
		return err
	}
	m.solveBacktracks = uint64(limit)
	return nil
}

func (m *CSATManager) SetTotalBacktrackLimit(limit uint64) error {
	if err := m.check(); err != nil {
		return err
	}
	m.totalBacktrackLimit = limit
	return nil
}

func (m *CSATManager) SetTotalInspectLimit(limit uint64) error {
	if err := m.check(); err != nil {
		return err
	}
	m.totalInspectLimit = limit
	return nil
}

func (m *CSATManager) TotalBacktracksMade() uint64 { return m.totalBacktracks }
func (m *CSATManager) TotalInspectsMade() uint64   { return m.totalInspects }

func (m *CSATManager) EnableDump(path string) error {
	if err := m.check(); err != nil {
		return err
	}
	if path != "" && strings.ContainsRune(path, 0) {
		return ErrDimension
	}
	m.dumpPath = path
	return nil
}

func (m *CSATManager) AddTarget(names []string, values []int) (int, error) {
	if err := m.check(); err != nil {
		return -1, err
	}
	if len(names) == 0 || len(names) != len(values) {
		return -1, ErrDimension
	}
	if len(names) > m.limits.MaxTargets {
		return -1, ErrLimit
	}
	target := csatTarget{gates: make([]int, len(names)), values: make([]bool, len(names))}
	for i, name := range names {
		if m.stopped() {
			return -1, ErrCancelled
		}
		id, found := m.byName[name]
		if !found {
			return -1, ErrInvalidLiteral
		}
		if values[i] != 0 && values[i] != 1 {
			return -1, ErrDimension
		}
		target.gates[i], target.values[i] = id, values[i] == 1
	}
	m.target = target
	m.ready = false
	m.result = nil
	return 1, nil
}

func csatLit(variable SatVar, negative bool) SatLit {
	literal, _ := MakeSatLit(variable, negative)
	return literal
}

func csatEquivalence(output, input SatVar, inverted bool) [][]SatLit {
	return [][]SatLit{
		{csatLit(output, true), csatLit(input, inverted)},
		{csatLit(output, false), csatLit(input, !inverted)},
	}
}

func csatAndClauses(output SatVar, inputs []SatVar, inverted bool) [][]SatLit {
	clauses := make([][]SatLit, 0, len(inputs)+1)
	for _, input := range inputs {
		clauses = append(clauses, []SatLit{csatLit(output, !inverted), csatLit(input, false)})
	}
	last := []SatLit{csatLit(output, inverted)}
	for _, input := range inputs {
		last = append(last, csatLit(input, true))
	}
	return append(clauses, last)
}

func csatOrClauses(output SatVar, inputs []SatVar, inverted bool) [][]SatLit {
	clauses := make([][]SatLit, 0, len(inputs)+1)
	for _, input := range inputs {
		clauses = append(clauses, []SatLit{csatLit(output, inverted), csatLit(input, true)})
	}
	last := []SatLit{csatLit(output, !inverted)}
	for _, input := range inputs {
		last = append(last, csatLit(input, false))
	}
	return append(clauses, last)
}

func (m *CSATManager) deriveClauses() ([][]SatLit, error) {
	clauses := make([][]SatLit, 0, 3*len(m.gates))
	for _, gate := range m.gates {
		if m.stopped() {
			return nil, ErrCancelled
		}
		inputs := make([]SatVar, len(gate.fanins))
		for i, id := range gate.fanins {
			inputs[i] = m.gates[id].variable
		}
		switch gate.type_ {
		case CSATBPI, CSATBPPI:
		case CSATConst:
			clauses = append(clauses, []SatLit{csatLit(gate.variable, false)})
		case CSATBAND:
			clauses = append(clauses, csatAndClauses(gate.variable, inputs, false)...)
		case CSATBNAND:
			clauses = append(clauses, csatAndClauses(gate.variable, inputs, true)...)
		case CSATBOR:
			clauses = append(clauses, csatOrClauses(gate.variable, inputs, false)...)
		case CSATBNOR:
			clauses = append(clauses, csatOrClauses(gate.variable, inputs, true)...)
		case CSATBXOR, CSATBXNOR:
			if len(inputs) == 1 {
				clauses = append(clauses, csatEquivalence(gate.variable, inputs[0], gate.type_ == CSATBXNOR)...)
			} else {
				xor, err := SatSolver3XorClauses(int(gate.variable), int(inputs[0]), int(inputs[1]), gate.type_ == CSATBXNOR)
				if err != nil {
					return nil, err
				}
				clauses = append(clauses, xor...)
			}
		case CSATBINV:
			clauses = append(clauses, csatEquivalence(gate.variable, inputs[0], true)...)
		case CSATBBUF, CSATBPPO, CSATBPO:
			clauses = append(clauses, csatEquivalence(gate.variable, inputs[0], false)...)
		default:
			return nil, ErrUnsupported
		}
	}
	return clauses, nil
}

func (m *CSATManager) SolveInit() error {
	if err := m.check(); err != nil {
		return err
	}
	if len(m.target.gates) == 0 {
		return ErrDimension
	}
	clauses, err := m.deriveClauses()
	if err != nil {
		return err
	}
	m.clauses = clauses
	m.ready = true
	return nil
}

func (m *CSATManager) AnalyzeTargets() error {
	if err := m.check(); err != nil {
		return err
	}
	return nil
}

func nonzeroMinimum(first, second uint64) uint64 {
	if first == 0 {
		return second
	}
	if second == 0 || first < second {
		return first
	}
	return second
}

func remainingCSATLimit(limit, used uint64) (uint64, bool) {
	if limit == 0 {
		return 0, true
	}
	if used >= limit {
		return 0, false
	}
	return limit - used, true
}

func (m *CSATManager) Solve() (CSATStatus, error) {
	if err := m.check(); err != nil {
		return CSATAborted, err
	}
	if !m.ready {
		return CSATUndetermined, ErrInvariant
	}
	if m.stopped() {
		return CSATAborted, ErrCancelled
	}
	solver, err := NewSatSolverWithLimit(len(m.gates))
	if err != nil {
		return CSATAborted, err
	}
	if err = solver.SetVariableCount(len(m.gates)); err != nil {
		return CSATAborted, err
	}
	for _, clause := range m.clauses {
		if m.stopped() {
			return CSATAborted, ErrCancelled
		}
		if _, err = solver.AddClause(clause); err != nil {
			return CSATAborted, err
		}
	}
	assumptions := make([]SatLit, len(m.target.gates))
	for i, id := range m.target.gates {
		assumptions[i] = csatLit(m.gates[id].variable, !m.target.values[i])
	}
	remainingBacktracks, backtracksAvailable := remainingCSATLimit(m.totalBacktrackLimit, m.totalBacktracks)
	remainingInspects, inspectsAvailable := remainingCSATLimit(m.totalInspectLimit, m.totalInspects)
	if !backtracksAvailable || !inspectsAvailable {
		result := CSATTargetResult{Status: CSATUndetermined}
		m.result = &result
		m.ready = false
		m.clauses = nil
		return result.Status, nil
	}
	conflicts := nonzeroMinimum(m.solveBacktracks, remainingBacktracks)
	propagations := remainingInspects
	answer, err := solver.Solve(assumptions, SatLimits{Conflicts: conflicts, Propagations: propagations, Stop: m.limits.Stop})
	if err != nil {
		return CSATAborted, err
	}
	if answer.Reason == SatCancelled {
		return CSATAborted, ErrCancelled
	}
	m.totalBacktracks += answer.Stats.Conflicts
	m.totalInspects += answer.Stats.Inspections
	result := CSATTargetResult{Decisions: int(answer.Stats.Decisions), Implications: int(answer.Stats.Propagations)}
	switch answer.Status {
	case SatUnsatisfiable:
		result.Status = CSATUnsatisfiable
	case SatSatisfiable:
		result.Status = CSATSatisfiable
		result.Names = make([]string, len(m.inputs))
		result.Values = make([]int, len(m.inputs))
		for i, id := range m.inputs {
			result.Names[i] = m.gates[id].name
			if answer.Model[m.gates[id].variable] {
				result.Values[i] = 1
			}
		}
	default:
		result.Status = CSATUndetermined
	}
	m.result = &result
	m.ready = false
	m.clauses = nil
	return result.Status, nil
}

func (m *CSATManager) TargetResult(targetID int) (CSATTargetResult, error) {
	if err := m.check(); err != nil {
		return CSATTargetResult{}, err
	}
	if targetID != 1 || m.result == nil {
		return CSATTargetResult{}, ErrDimension
	}
	return m.result.Clone(), nil
}

func (m *CSATManager) Bench() ([]byte, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	var builder strings.Builder
	for _, id := range m.inputs {
		builder.WriteString("INPUT(")
		builder.WriteString(m.gates[id].name)
		builder.WriteString(")\n")
	}
	for _, id := range m.outputs {
		builder.WriteString("OUTPUT(")
		builder.WriteString(m.gates[id].name)
		builder.WriteString(")\n")
	}
	labels := [...]string{"CONST", "PI", "PPI", "AND", "NAND", "OR", "NOR", "XOR", "XNOR", "NOT", "BUFF"}
	for _, gate := range m.gates {
		if m.stopped() {
			return nil, ErrCancelled
		}
		switch gate.type_ {
		case CSATBPI, CSATBPPI:
			continue
		case CSATBPPO, CSATBPO:
			builder.WriteString(gate.name)
			builder.WriteString(" = BUFF(")
			builder.WriteString(m.gates[gate.fanins[0]].name)
			builder.WriteString(")\n")
			continue
		}
		builder.WriteString(gate.name)
		builder.WriteString(" = ")
		if gate.type_ == CSATConst {
			builder.WriteString("1\n")
			continue
		}
		builder.WriteString(labels[gate.type_])
		builder.WriteByte('(')
		for i, fanin := range gate.fanins {
			if i != 0 {
				builder.WriteByte(',')
			}
			builder.WriteString(m.gates[fanin].name)
		}
		builder.WriteString(")\n")
	}
	return []byte(builder.String()), nil
}

func (m *CSATManager) DumpBench() error {
	if err := m.check(); err != nil {
		return err
	}
	path := m.dumpPath
	if path == "" {
		path = "abc_test.bench"
	}
	data, err := m.Bench()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (m *CSATManager) FormatResult(targetID int) (string, error) {
	result, err := m.TargetResult(targetID)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString("status=")
	builder.WriteString(strconv.Itoa(int(result.Status)))
	for i, name := range result.Names {
		builder.WriteByte(' ')
		builder.WriteString(name)
		builder.WriteByte('=')
		builder.WriteString(strconv.Itoa(result.Values[i]))
	}
	return builder.String(), nil
}
