package abc

// Bounded proof-core construction from src/sat/bmc/bmcBCore.c.

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

type GiaBoundedCoreVariable struct {
	Object GiaID
	Frame  int
}

type GiaBoundedCoreCNF struct {
	Variables             []GiaBoundedCoreVariable
	Clauses               [][]SatLit
	PivotVariables        []SatVar
	InvalidPivotObjectIDs []int
	Literals              int
}

type GiaBoundedCoreOptions struct {
	Frame, Output                         int
	PivotObjectIDs                        []int
	MaxVariables, MaxClauses, MaxLiterals int
	SAT                                   SatLimits
	Proof                                 SatProofDAGLimits
	Stop                                  func() bool
}

type GiaBoundedCoreProofClause struct {
	Node     int
	Literals []SatLit
	Final    bool
}

type GiaBoundedCoreResult struct {
	Status                SatStatus
	Reason                SatStopReason
	Stats                 SatStats
	Variables             []GiaBoundedCoreVariable
	PivotVariables        []SatVar
	InvalidPivotObjectIDs []int
	RootClauseHandles     []int
	LearnedCore           []GiaBoundedCoreProofClause
}

func checkGiaBoundedCoreOptions(options GiaBoundedCoreOptions) (GiaBoundedCoreOptions, error) {
	if options.Frame < 0 || options.Output < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
		return options, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 4 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 12 << 20
	}
	if options.MaxVariables < 2 || options.MaxClauses < 1 || options.MaxLiterals < 1 {
		return options, ErrLimit
	}
	return options, nil
}

// ParseGiaBoundedCorePivots parses the whitespace-separated object IDs accepted
// by Bmc_ManBCoreReadPivots. maxPivots zero selects a one-million-entry bound.
func ParseGiaBoundedCorePivots(data []byte, maxPivots int) ([]int, error) {
	if maxPivots < 0 {
		return nil, ErrDimension
	}
	if maxPivots == 0 {
		maxPivots = 1 << 20
	}
	fields := bytes.Fields(data)
	if len(fields) > maxPivots {
		return nil, ErrLimit
	}
	result := make([]int, len(fields))
	for index, field := range fields {
		value, err := strconv.Atoi(string(field))
		if err != nil {
			return nil, fmt.Errorf("%w: pivot %q", ErrDimension, field)
		}
		result[index] = value
	}
	return result, nil
}

// ReadGiaBoundedCorePivots is the checked file counterpart of the source
// reader. maxBytes zero selects 16 MiB; maxPivots is passed to the parser.
func ReadGiaBoundedCorePivots(path string, maxBytes, maxPivots int) ([]int, error) {
	if maxBytes < 0 {
		return nil, ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = 16 << 20
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, int64(maxBytes)+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxBytes {
		return nil, ErrLimit
	}
	return ParseGiaBoundedCorePivots(data, maxPivots)
}

// CollectGiaBoundedCorePivots maps requested object IDs to every corresponding
// timeframe SAT variable, scanning the variable map in deterministic order.
func CollectGiaBoundedCorePivots(manager *GiaManager, objectIDs []int, variables []GiaBoundedCoreVariable) (pivots []SatVar, invalid []int, err error) {
	if manager == nil {
		return nil, nil, ErrDimension
	}
	marked := make([]bool, manager.ObjectCount())
	for _, id := range objectIDs {
		if id <= 0 || id >= len(marked) {
			invalid = append(invalid, id)
			continue
		}
		marked[id] = true
	}
	for variable, mapped := range variables {
		if mapped.Object == InvalidGiaID {
			continue
		}
		if uint64(mapped.Object) >= uint64(len(marked)) {
			return nil, nil, ErrInvariant
		}
		if marked[mapped.Object] {
			pivots = append(pivots, SatVar(variable))
		}
	}
	return pivots, invalid, nil
}

type giaBoundedCoreBuilder struct {
	manager *GiaManager
	options GiaBoundedCoreOptions
	cnf     GiaBoundedCoreCNF
}

func (b *giaBoundedCoreBuilder) stopped() bool {
	return b.options.Stop != nil && b.options.Stop()
}

func (b *giaBoundedCoreBuilder) variable(object GiaID, frame int) (int, error) {
	if len(b.cnf.Variables) == b.options.MaxVariables {
		return -1, ErrLimit
	}
	variable := len(b.cnf.Variables)
	b.cnf.Variables = append(b.cnf.Variables, GiaBoundedCoreVariable{Object: object, Frame: frame})
	return variable, nil
}

func (b *giaBoundedCoreBuilder) clause(literals ...SatLit) error {
	if len(literals) == 0 || len(b.cnf.Clauses) == b.options.MaxClauses || len(literals) > b.options.MaxLiterals-b.cnf.Literals {
		return ErrLimit
	}
	for _, literal := range literals {
		if !literal.Valid() || uint64(literal.Var()) >= uint64(len(b.cnf.Variables)) {
			return ErrInvariant
		}
	}
	b.cnf.Clauses = append(b.cnf.Clauses, slices.Clone(literals))
	b.cnf.Literals += len(literals)
	return nil
}

func boundedCoreLit(variable int, negative bool) SatLit {
	literal, _ := MakeSatLit(SatVar(variable), negative)
	return literal
}

func (b *giaBoundedCoreBuilder) constant(variable int, complement bool) error {
	return b.clause(boundedCoreLit(variable, complement))
}

func (b *giaBoundedCoreBuilder) buffer(a, other int, complement bool) error {
	if err := b.clause(boundedCoreLit(a, false), boundedCoreLit(other, !complement)); err != nil {
		return err
	}
	return b.clause(boundedCoreLit(a, true), boundedCoreLit(other, complement))
}

func (b *giaBoundedCoreBuilder) and(output, input0, input1 int, complement0, complement1 bool) error {
	if err := b.clause(boundedCoreLit(output, true), boundedCoreLit(input0, complement0)); err != nil {
		return err
	}
	if err := b.clause(boundedCoreLit(output, true), boundedCoreLit(input1, complement1)); err != nil {
		return err
	}
	return b.clause(boundedCoreLit(output, false), boundedCoreLit(input0, !complement0), boundedCoreLit(input1, !complement1))
}

func (b *giaBoundedCoreBuilder) constraint(variable, auxiliary int, complement bool) error {
	literal := boundedCoreLit(variable, complement)
	if err := b.clause(literal, boundedCoreLit(auxiliary, false)); err != nil {
		return err
	}
	return b.clause(literal, boundedCoreLit(auxiliary, true))
}

type giaBoundedCoreRoot struct {
	object   GiaID
	variable int
}

func (b *giaBoundedCoreBuilder) collectFrame(roots []giaBoundedCoreRoot, frame int, objectVariables []int) ([]giaBoundedCoreRoot, []GiaID, error) {
	visited := make([]bool, b.manager.ObjectCount())
	visited[0] = true
	constant, err := b.variable(0, frame)
	if err != nil {
		return nil, nil, err
	}
	objectVariables[0] = constant
	if err = b.constant(constant, true); err != nil {
		return nil, nil, err
	}
	nextRoots := make([]giaBoundedCoreRoot, 0)
	assigned := make([]GiaID, 0)
	for _, root := range roots {
		rootObject := b.manager.objects[root.object]
		driver, faninErr := rootObject.Fanin0(root.object)
		if faninErr != nil {
			return nil, nil, faninErr
		}
		stack := []GiaID{driver.ID()}
		for len(stack) > 0 {
			if b.stopped() {
				return nil, nil, ErrCancelled
			}
			last := len(stack) - 1
			id := stack[last]
			stack = stack[:last]
			if visited[id] {
				continue
			}
			visited[id] = true
			variable, variableErr := b.variable(id, frame)
			if variableErr != nil {
				return nil, nil, variableErr
			}
			objectVariables[id] = variable
			assigned = append(assigned, id)
			object := b.manager.objects[id]
			if object.IsCI() {
				if int(object.Diff1()) >= b.manager.PrimaryInputCount() {
					ri, pairErr := b.manager.RegisterInputForOutput(id)
					if pairErr != nil {
						return nil, nil, pairErr
					}
					nextRoots = append(nextRoots, giaBoundedCoreRoot{object: ri, variable: variable})
				}
				continue
			}
			if !object.IsAnd() || object.IsBuffer() || object.IsXor() {
				return nil, nil, ErrUnsupported
			}
			if _, mux := b.manager.MuxControl(id); mux {
				return nil, nil, ErrUnsupported
			}
			fanin0, firstErr := object.Fanin0(id)
			fanin1, secondErr := object.Fanin1(id)
			if firstErr != nil || secondErr != nil {
				return nil, nil, ErrInvariant
			}
			stack = append(stack, fanin1.ID(), fanin0.ID())
		}
	}
	return nextRoots, assigned, nil
}

// BuildGiaBoundedCoreCNF constructs the exact frame-local variable map and
// gate clauses used by Bmc_ManBCoreCollect without mutating GIA scratch.
func (manager *GiaManager) BuildGiaBoundedCoreCNF(options GiaBoundedCoreOptions) (GiaBoundedCoreCNF, error) {
	checked, err := checkGiaBoundedCoreOptions(options)
	if err != nil {
		return GiaBoundedCoreCNF{}, err
	}
	if manager == nil || checked.Output >= manager.PrimaryOutputCount() {
		return GiaBoundedCoreCNF{}, ErrDimension
	}
	if err = manager.Check(); err != nil {
		return GiaBoundedCoreCNF{}, err
	}
	builder := giaBoundedCoreBuilder{manager: manager, options: checked}
	builder.cnf.Variables = append(builder.cnf.Variables, GiaBoundedCoreVariable{Object: InvalidGiaID, Frame: -1})
	po := manager.cos[checked.Output]
	poVariable, err := builder.variable(po, checked.Frame)
	if err != nil {
		return GiaBoundedCoreCNF{}, err
	}
	roots := []giaBoundedCoreRoot{{object: po, variable: poVariable}}
	objectVariables := make([]int, manager.ObjectCount())
	for frame := checked.Frame; frame >= 0; frame-- {
		if builder.stopped() {
			return GiaBoundedCoreCNF{}, ErrCancelled
		}
		for index := range objectVariables {
			objectVariables[index] = -1
		}
		nextRoots, assigned, collectErr := builder.collectFrame(roots, frame, objectVariables)
		if collectErr != nil {
			return GiaBoundedCoreCNF{}, collectErr
		}
		if frame == checked.Frame {
			object := manager.objects[po]
			driver, _ := object.Fanin0(po)
			if err = builder.constraint(objectVariables[driver.ID()], poVariable, driver.Complemented()); err != nil {
				return GiaBoundedCoreCNF{}, err
			}
		} else {
			for _, root := range roots {
				object := manager.objects[root.object]
				driver, faninErr := object.Fanin0(root.object)
				if faninErr != nil {
					return GiaBoundedCoreCNF{}, faninErr
				}
				if err = builder.buffer(root.variable, objectVariables[driver.ID()], driver.Complemented()); err != nil {
					return GiaBoundedCoreCNF{}, err
				}
			}
		}
		for _, id := range assigned {
			object := manager.objects[id]
			if object.IsCI() {
				continue
			}
			fanin0, _ := object.Fanin0(id)
			fanin1, _ := object.Fanin1(id)
			if err = builder.and(objectVariables[id], objectVariables[fanin0.ID()], objectVariables[fanin1.ID()], fanin0.Complemented(), fanin1.Complemented()); err != nil {
				return GiaBoundedCoreCNF{}, err
			}
		}
		roots = nextRoots
	}
	for _, root := range roots {
		auxiliary := len(builder.cnf.Variables)
		registerOutput, pairErr := manager.RegisterOutputForInput(root.object)
		if pairErr != nil {
			return GiaBoundedCoreCNF{}, pairErr
		}
		if _, err = builder.variable(registerOutput, -1); err != nil {
			return GiaBoundedCoreCNF{}, err
		}
		if err = builder.constraint(root.variable, auxiliary, true); err != nil {
			return GiaBoundedCoreCNF{}, err
		}
	}
	builder.cnf.PivotVariables, builder.cnf.InvalidPivotObjectIDs, err = CollectGiaBoundedCorePivots(manager, checked.PivotObjectIDs, builder.cnf.Variables)
	if err != nil {
		return GiaBoundedCoreCNF{}, err
	}
	return builder.cnf, nil
}

// BoundedCore solves the collected BMC instance and, on UNSAT, returns both
// the independently rechecked root core and all reachable learned clauses.
func (manager *GiaManager) BoundedCore(options GiaBoundedCoreOptions) (GiaBoundedCoreResult, error) {
	cnf, err := manager.BuildGiaBoundedCoreCNF(options)
	if err != nil {
		return GiaBoundedCoreResult{}, err
	}
	checked, _ := checkGiaBoundedCoreOptions(options)
	result := GiaBoundedCoreResult{
		Variables:             slices.Clone(cnf.Variables),
		PivotVariables:        slices.Clone(cnf.PivotVariables),
		InvalidPivotObjectIDs: slices.Clone(cnf.InvalidPivotObjectIDs),
	}
	solver, err := NewSatSolver2(SatSolver2Options{MaxVariables: checked.MaxVariables, MaxClauses: checked.MaxClauses})
	if err != nil {
		return GiaBoundedCoreResult{}, err
	}
	defer solver.Stop()
	if err = solver.SetVariableCount(len(cnf.Variables)); err != nil {
		return GiaBoundedCoreResult{}, err
	}
	for index, clause := range cnf.Clauses {
		if _, err = solver.AddClause(clause, index+1); err != nil {
			return GiaBoundedCoreResult{}, err
		}
	}
	if err = solver.EnableResolutionProofTracking(checked.Proof); err != nil {
		return GiaBoundedCoreResult{}, err
	}
	answer, err := solver.Solve(nil, checked.SAT)
	if err != nil {
		return GiaBoundedCoreResult{}, err
	}
	result.Status, result.Reason, result.Stats = answer.Status, answer.Reason, answer.Stats
	if answer.Status != SatUnsatisfiable {
		return result, nil
	}
	result.RootClauseHandles, err = solver.ResolutionUnsatCore()
	if err != nil {
		return GiaBoundedCoreResult{}, err
	}
	learned, err := solver.ResolutionLearnedCore()
	if err != nil {
		return GiaBoundedCoreResult{}, err
	}
	result.LearnedCore = make([]GiaBoundedCoreProofClause, len(learned))
	for index, clause := range learned {
		result.LearnedCore[index] = GiaBoundedCoreProofClause{Node: clause.Node, Literals: slices.Clone(clause.Literals), Final: clause.Final}
	}
	return result, nil
}

// FormatGiaBoundedCoreProof renders learned clauses using object(frame)
// spellings, preserving the source report's stable human-readable contract.
func FormatGiaBoundedCoreProof(result GiaBoundedCoreResult) (string, error) {
	if result.Status != SatUnsatisfiable {
		return "", ErrInvariant
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "UNSAT contains %d learned clauses:\n", len(result.LearnedCore))
	for index, clause := range result.LearnedCore {
		fmt.Fprintf(&builder, "%6d : %6d : ", index, clause.Node)
		for _, literal := range clause.Literals {
			if uint64(literal.Var()) >= uint64(len(result.Variables)) {
				return "", ErrInvariant
			}
			mapped := result.Variables[literal.Var()]
			if literal.Negative() {
				builder.WriteByte('!')
			}
			fmt.Fprintf(&builder, "%d(%d) ", mapped.Object, mapped.Frame)
		}
		if len(clause.Literals) == 0 {
			builder.WriteString("Empty")
		}
		builder.WriteByte('\n')
	}
	return builder.String(), nil
}

func WriteGiaBoundedCoreProof(path string, result GiaBoundedCoreResult) error {
	text, err := FormatGiaBoundedCoreProof(result)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(text), 0o644)
}
