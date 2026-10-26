package abc

// Sparse ternary-guided unrolling and memory-efficient bounded checking from
// src/sat/bmc/bmcBmcAnd.c.

import (
	"errors"
	"fmt"
	"slices"
)

type GiaAndUnrollOptions struct {
	Frames, AdditionalFrames, MaxFrames int
	MaxObjects, MaxMappingEntries       int
	Stop                                func() bool
}

type GiaAndInput struct {
	Frame, Input int
	Literal      GiaLit
}

type GiaAndUnrolling struct {
	graph         *GiaManager
	frames        int
	firstUnknown  int
	values        []GiaTernary
	inputs        []GiaAndInput
	active        []bool
	sourceObjects int
}

func (u *GiaAndUnrolling) Graph() *GiaManager { return u.graph }
func (u *GiaAndUnrolling) Frames() int        { return u.frames }
func (u *GiaAndUnrolling) FirstUnknownFrame() int {
	return u.firstUnknown
}
func (u *GiaAndUnrolling) Inputs() []GiaAndInput { return slices.Clone(u.inputs) }
func (u *GiaAndUnrolling) FrameActive(frame int) (bool, error) {
	if frame < 0 || frame >= u.frames {
		return false, ErrDimension
	}
	return u.active[frame], nil
}
func (u *GiaAndUnrolling) FrameValues(frame int) ([]GiaTernary, error) {
	if frame < 0 || frame >= u.frames {
		return nil, ErrDimension
	}
	begin := frame * u.sourceObjects
	return slices.Clone(u.values[begin : begin+u.sourceObjects]), nil
}

func normalizeGiaAndUnrollOptions(options GiaAndUnrollOptions) (GiaAndUnrollOptions, error) {
	if options.Frames < 0 || options.AdditionalFrames < 0 || options.MaxFrames < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return options, ErrDimension
	}
	if options.Frames == 0 && options.AdditionalFrames == 0 {
		return options, ErrDimension
	}
	if options.Frames > 0 {
		if options.MaxFrames != 0 && options.Frames > options.MaxFrames {
			return options, ErrLimit
		}
		options.MaxFrames = options.Frames
	} else if options.MaxFrames == 0 {
		options.MaxFrames = 1 << 20
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 4 << 20
	}
	return options, nil
}

func checkGiaAndGraph(manager *GiaManager) error {
	if manager == nil || manager.PrimaryOutputCount() == 0 || manager.ConstraintCount() != 0 {
		return ErrDimension
	}
	if err := manager.Check(); err != nil {
		return err
	}
	for id, object := range manager.objects {
		if !object.IsAnd() {
			continue
		}
		if object.IsBuffer() || object.IsXor() {
			return ErrUnsupported
		}
		if _, mux := manager.MuxControl(GiaID(id)); mux {
			return ErrUnsupported
		}
	}
	return nil
}

func giaAndRead(values []GiaTernary, literal GiaLit) GiaTernary {
	value := values[literal.ID()]
	if literal.Complemented() {
		return giaTernaryNot(value)
	}
	return value
}

// AndTernaryUnroll performs the zero-state/unknown-input simulation, backward
// unknown-controlling cone selection and forward sparse construction used by
// bmcBmcAnd.c. A zero Frames value selects the source's adaptive mode: stop
// AdditionalFrames after the first unknown PO, with MaxFrames as a mandatory
// safety cap. Known-one POs are rejected because the source sparse builder
// silently turns them into zero; returning an error avoids an unsound check.
func (manager *GiaManager) AndTernaryUnroll(options GiaAndUnrollOptions) (*GiaAndUnrolling, error) {
	if err := checkGiaAndGraph(manager); err != nil {
		return nil, err
	}
	options, err := normalizeGiaAndUnrollOptions(options)
	if err != nil {
		return nil, err
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	objectCount := manager.ObjectCount()
	if objectCount > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	firstUnknown := -1
	values := make([]GiaTernary, 0)
	state := make([]GiaTernary, manager.RegisterCount())
	for index := range state {
		state[index] = GiaTernaryZero
	}
	for frame := 0; frame < options.MaxFrames; frame++ {
		if options.Frames == 0 && firstUnknown >= 0 && frame == firstUnknown+options.AdditionalFrames {
			break
		}
		if stopped() {
			return nil, ErrCancelled
		}
		if _, err := checkedSize(0, frame+1, objectCount, options.MaxMappingEntries); err != nil {
			return nil, err
		}
		row := make([]GiaTernary, objectCount)
		row[0] = GiaTernaryZero
		for _, id := range manager.cis[:manager.PrimaryInputCount()] {
			row[id] = GiaTernaryUnknown
		}
		for index, id := range manager.cis[manager.PrimaryInputCount():] {
			row[id] = state[index]
		}
		for index, object := range manager.objects {
			if stopped() {
				return nil, ErrCancelled
			}
			id := GiaID(index)
			switch {
			case object.IsAnd():
				first, _ := object.Fanin0(id)
				second, _ := object.Fanin1(id)
				row[index] = giaTernaryAnd(giaAndRead(row, first), giaAndRead(row, second))
			case object.IsCO():
				first, _ := object.Fanin0(id)
				row[index] = giaAndRead(row, first)
			}
		}
		for _, id := range manager.cos[:manager.PrimaryOutputCount()] {
			switch row[id] {
			case GiaTernaryUnknown:
				if firstUnknown < 0 {
					firstUnknown = frame
				}
			case GiaTernaryOne:
				return nil, fmt.Errorf("%w: ternary-guided unroll cannot represent a known-one PO", ErrUnsupported)
			}
		}
		for index, id := range manager.cos[manager.PrimaryOutputCount():] {
			state[index] = row[id]
		}
		values = append(values, row...)
	}
	frames := len(values) / objectCount
	if frames == 0 {
		return nil, ErrInvariant
	}
	needed := make([]bool, frames*objectCount)
	roots := make([][]GiaID, frames)
	leaves := []GiaID(nil)
	active := make([]bool, frames)
	for frame := frames - 1; frame >= 0; frame-- {
		if stopped() {
			return nil, ErrCancelled
		}
		row := values[frame*objectCount : (frame+1)*objectCount]
		for _, id := range manager.cos[:manager.PrimaryOutputCount()] {
			if row[id] == GiaTernaryUnknown {
				roots[frame] = append(roots[frame], id)
			}
		}
		roots[frame] = append(roots[frame], leaves...)
		active[frame] = true
		leaves = leaves[:0]
		stack := slices.Clone(roots[frame])
		for len(stack) != 0 {
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			object := manager.objects[id]
			if object.IsCO() {
				literal, _ := object.Fanin0(id)
				if giaAndRead(row, literal) != GiaTernaryUnknown {
					return nil, ErrInvariant
				}
				stack = append(stack, literal.ID())
				continue
			}
			mark := frame*objectCount + int(id)
			if needed[mark] {
				continue
			}
			needed[mark] = true
			switch {
			case object.IsAnd():
				for fanin := 0; fanin < 2; fanin++ {
					literal, _ := object.Fanin0(id)
					if fanin == 1 {
						literal, _ = object.Fanin1(id)
					}
					value := giaAndRead(row, literal)
					if value == GiaTernaryUnknown {
						stack = append(stack, literal.ID())
					} else if value != GiaTernaryOne {
						return nil, ErrInvariant
					}
				}
			case object.IsCI():
				terminal, _ := object.TerminalIndex()
				if terminal >= manager.PrimaryInputCount() {
					register := terminal - manager.PrimaryInputCount()
					leaves = append(leaves, manager.cos[manager.PrimaryOutputCount()+register])
				}
			default:
				if id != 0 {
					return nil, ErrInvariant
				}
			}
		}
		if len(leaves) == 0 {
			break
		}
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name, target.Spec = manager.Name, manager.Spec
	if err = target.HashStart(); err != nil {
		return nil, err
	}
	mapping := make([]GiaLit, objectCount)
	for index := range mapping {
		mapping[index] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	provenance := make([]GiaAndInput, 0)
	for frame := 0; frame < frames; frame++ {
		if stopped() {
			return nil, ErrCancelled
		}
		if !active[frame] {
			for range manager.cos[:manager.PrimaryOutputCount()] {
				if _, err = target.AppendCO(GiaConst0); err != nil {
					return nil, err
				}
			}
			continue
		}
		row := values[frame*objectCount : (frame+1)*objectCount]
		for index, id := range manager.cis[:manager.PrimaryInputCount()] {
			if !needed[frame*objectCount+int(id)] {
				continue
			}
			literal, appendErr := target.AppendCI()
			if appendErr != nil {
				return nil, appendErr
			}
			mapping[id] = literal
			provenance = append(provenance, GiaAndInput{Frame: frame, Input: index, Literal: literal})
		}
		copyFanin := func(literal GiaLit) (GiaLit, error) {
			if giaAndRead(row, literal) != GiaTernaryUnknown {
				if giaAndRead(row, literal) != GiaTernaryOne {
					return InvalidGiaLit, ErrInvariant
				}
				return GiaConst1, nil
			}
			mapped := mapping[literal.ID()]
			if !mapped.Valid() {
				return InvalidGiaLit, ErrInvariant
			}
			return mapped.NotIf(literal.Complemented()), nil
		}
		for index, object := range manager.objects {
			if !object.IsAnd() || !needed[frame*objectCount+index] {
				continue
			}
			id := GiaID(index)
			first, _ := object.Fanin0(id)
			second, _ := object.Fanin1(id)
			a, mapErr := copyFanin(first)
			if mapErr != nil {
				return nil, mapErr
			}
			b, mapErr := copyFanin(second)
			if mapErr != nil {
				return nil, mapErr
			}
			mapping[id], err = target.HashAnd(a, b)
			if err != nil {
				return nil, err
			}
		}
		rootSet := make([]bool, objectCount)
		for _, id := range roots[frame] {
			rootSet[id] = true
			object := manager.objects[id]
			driver, _ := object.Fanin0(id)
			mapping[id], err = copyFanin(driver)
			if err != nil {
				return nil, err
			}
		}
		for _, id := range manager.cos[:manager.PrimaryOutputCount()] {
			driver := GiaConst0
			if rootSet[id] {
				driver = mapping[id]
			}
			if _, err = target.AppendCO(driver); err != nil {
				return nil, err
			}
		}
		for _, id := range roots[frame] {
			object := manager.objects[id]
			terminal, _ := object.TerminalIndex()
			if terminal < manager.PrimaryOutputCount() {
				continue
			}
			register := terminal - manager.PrimaryOutputCount()
			mapping[manager.cis[manager.PrimaryInputCount()+register]] = mapping[id]
		}
	}
	target.HashStop()
	return &GiaAndUnrolling{graph: target, frames: frames, firstUnknown: firstUnknown, values: values, inputs: provenance, active: active, sourceObjects: objectCount}, nil
}

type GiaAndBMCOptions struct {
	Unroll       GiaAndUnrollOptions
	MaxVariables int
	SAT          SatLimits
}

// AndBoundedCheck uses the sparse unroll with the concrete Go CNF encoder and
// solver. SAT witnesses are translated back to the original frame/input space
// and replayed on the original transition system before return.
func (manager *GiaManager) AndBoundedCheck(options GiaAndBMCOptions) (result BmcResult, err error) {
	result.Frame, result.Property = -1, -1
	if options.MaxVariables < 0 {
		return result, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	unrolling, err := manager.AndTernaryUnroll(options.Unroll)
	if err != nil {
		if errors.Is(err, ErrCancelled) {
			result.Reason, result.SatReason = BmcCancelled, SatCancelled
			return result, nil
		}
		if errors.Is(err, ErrLimit) {
			result.Reason = BmcResourceLimit
			return result, nil
		}
		return result, err
	}
	aig, giaToAig, err := unrolling.graph.ToAig()
	if err != nil {
		return result, fmt.Errorf("sparse GIA conversion: %w", err)
	}
	relation, err := EncodeAigCNF(aig)
	if err != nil {
		return result, fmt.Errorf("sparse CNF encoding: %w", err)
	}
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return result, err
	}
	defer func() { result.Stats = solver.Statistics() }()
	ok, err := relation.loadIntoWithStop(solver, 0, options.SAT.Stop)
	if err != nil {
		if errors.Is(err, ErrCancelled) {
			result.Reason, result.SatReason = BmcCancelled, SatCancelled
			return result, nil
		}
		if errors.Is(err, ErrLimit) {
			result.Reason = BmcResourceLimit
			return result, nil
		}
		return result, fmt.Errorf("sparse CNF loading: %w", err)
	}
	if !ok {
		return result, fmt.Errorf("%w: sparse transition relation is inconsistent", ErrInvariant)
	}
	outputCount := manager.PrimaryOutputCount()
	for frame := 0; frame < unrolling.frames; frame++ {
		result.Frame = frame
		for property := 0; property < outputCount; property++ {
			answer, solveErr := solver.Solve([]SatLit{relation.outputs[frame*outputCount+property]}, options.SAT)
			if solveErr != nil {
				return result, fmt.Errorf("frame %d property %d solve: %w", frame, property, solveErr)
			}
			switch answer.Status {
			case SatUnsatisfiable:
				continue
			case SatUnknown:
				result.Status = BmcUnknown
				result.SatReason = answer.Reason
				result.Reason = BmcSatLimit
				if answer.Reason == SatCancelled {
					result.Reason = BmcCancelled
				}
				return result, nil
			case SatSatisfiable:
				counterexample, makeErr := NewCounterexample(manager.RegisterCount(), manager.PrimaryInputCount(), frame+1)
				if makeErr != nil {
					return result, fmt.Errorf("counterexample allocation: %w", makeErr)
				}
				counterexample.Property = property
				for index, input := range unrolling.inputs {
					if input.Frame > frame {
						continue
					}
					mapped := giaToAig[input.Literal.ID()].NotIf(input.Literal.Complemented())
					satLiteral, literalErr := relation.ObjectLiteral(mapped.ID())
					if literalErr != nil {
						return result, fmt.Errorf("input %d SAT literal: %w", index, literalErr)
					}
					satLiteral ^= SatLit(mapped & 1)
					if int(satLiteral.Var()) >= len(answer.Model) {
						return result, ErrInvariant
					}
					value := answer.Model[satLiteral.Var()] != satLiteral.Negative()
					if makeErr = counterexample.SetInput(input.Frame, input.Input, value); makeErr != nil {
						return result, fmt.Errorf("input %d provenance frame=%d input=%d: %w", index, input.Frame, input.Input, makeErr)
					}
					_ = index
				}
				sourceAig, _, convertErr := manager.ToAig()
				if convertErr != nil {
					return result, fmt.Errorf("source GIA conversion: %w", convertErr)
				}
				model, modelErr := NewAigerModel(sourceAig, outputCount, 0, 0, make([]LatchInit, manager.RegisterCount()))
				if modelErr != nil {
					return result, fmt.Errorf("source model: %w", modelErr)
				}
				if replayErr := model.ValidateCounterexample(counterexample); replayErr != nil {
					return result, fmt.Errorf("%w: sparse BMC witness failed replay: %v", ErrInvariant, replayErr)
				}
				result.Status = BmcCounterexample
				result.Reason = BmcWitnessFound
				result.Property = property
				result.Counterexample = counterexample
				return result, nil
			default:
				return result, ErrInvariant
			}
		}
		result.CheckedFrames = frame + 1
	}
	result.Status = BmcNoCounterexample
	result.Reason = BmcBoundReached
	return result, nil
}
