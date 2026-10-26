package abc

// Sequential dual-rail and property-blocking transforms from src/aig/saig/saigDual.c.

type SAIGDualRailOptions struct {
	DualPrimaryInputs int
	DCRegisters       []bool
	DualRegisters     bool
	MiterRegisters    bool
	ComplementOutput  bool
	CheckZero         bool
	CheckOne          bool
	MaxObjects        int
	MaxWork           int
	Stop              func() bool
}

type SAIGBlockPropertyOptions struct {
	MaxObjects int
	MaxWork    int
	Stop       func() bool
}

type saigDualBudget struct {
	maxObjects int
	maxWork    int
	work       int
	stop       func() bool
}

func newSAIGDualBudget(maxObjects, maxWork int, stop func() bool) (*saigDualBudget, error) {
	if maxObjects < 0 || maxWork < 0 {
		return nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if maxWork == 0 {
		maxWork = 8 << 20
	}
	return &saigDualBudget{maxObjects: maxObjects, maxWork: maxWork, stop: stop}, nil
}

func (budget *saigDualBudget) step(graph *AigManager) error {
	if budget.stop != nil && budget.stop() {
		return ErrCancelled
	}
	if budget.work == budget.maxWork {
		return ErrLimit
	}
	budget.work++
	if graph != nil && graph.ObjectCount() > budget.maxObjects {
		return ErrLimit
	}
	return nil
}

func saigDualAnd(graph *AigManager, budget *saigDualBudget, left, right AigLit) (AigLit, error) {
	if err := budget.step(graph); err != nil {
		return InvalidAigLit, err
	}
	result, err := graph.And(left, right)
	if err != nil {
		return InvalidAigLit, err
	}
	if graph.ObjectCount() > budget.maxObjects {
		return InvalidAigLit, ErrLimit
	}
	return result, nil
}

func saigDualOr(graph *AigManager, budget *saigDualBudget, left, right AigLit) (AigLit, error) {
	if err := budget.step(graph); err != nil {
		return InvalidAigLit, err
	}
	result, err := graph.Or(left, right)
	if err != nil {
		return InvalidAigLit, err
	}
	if graph.ObjectCount() > budget.maxObjects {
		return InvalidAigLit, ErrLimit
	}
	return result, nil
}

func saigDualCI(graph *AigManager, budget *saigDualBudget) (AigLit, error) {
	if err := budget.step(graph); err != nil {
		return InvalidAigLit, err
	}
	result, err := graph.CreateCI()
	if err != nil {
		return InvalidAigLit, err
	}
	if graph.ObjectCount() > budget.maxObjects {
		return InvalidAigLit, ErrLimit
	}
	return result, nil
}

func saigDualCO(graph *AigManager, budget *saigDualBudget, driver AigLit) error {
	if err := budget.step(graph); err != nil {
		return err
	}
	if _, err := graph.CreateCO(driver); err != nil {
		return err
	}
	if graph.ObjectCount() > budget.maxObjects {
		return ErrLimit
	}
	return nil
}

func saigDualFanin(graph *AigManager, budget *saigDualBudget, copies [][2]AigLit, fanin AigLit) (AigLit, AigLit, error) {
	if !fanin.Valid() || uint64(fanin.ID()) >= uint64(len(copies)) || !copies[fanin.ID()][0].Valid() || !copies[fanin.ID()][1].Valid() {
		return InvalidAigLit, InvalidAigLit, ErrInvariant
	}
	zero, one := copies[fanin.ID()][0], copies[fanin.ID()][1]
	if !fanin.Complemented() {
		return zero, one, nil
	}
	care, err := saigDualOr(graph, budget, zero, one)
	if err != nil {
		return InvalidAigLit, InvalidAigLit, err
	}
	newZero, err := saigDualAnd(graph, budget, one, care)
	if err != nil {
		return InvalidAigLit, InvalidAigLit, err
	}
	newOne, err := saigDualAnd(graph, budget, zero, care)
	if err != nil {
		return InvalidAigLit, InvalidAigLit, err
	}
	return newZero, newOne, nil
}

// SAIGDuplicateDual transforms a sequential AIG into the source's ternary
// dual-rail miter. Each rail pair encodes definitely-zero, definitely-one or
// unknown. If both checks are selected, CheckZero has the source's priority.
func SAIGDuplicateDual(source *AigManager, options SAIGDualRailOptions) (*AigManager, error) {
	budget, err := newSAIGDualBudget(options.MaxObjects, options.MaxWork, options.Stop)
	if err != nil {
		return nil, err
	}
	if source == nil || options.DualPrimaryInputs < 0 {
		return nil, ErrDimension
	}
	if err = source.Check(); err != nil {
		return nil, err
	}
	primaryInputs := source.PrimaryInputCount()
	registers := source.RegisterCount()
	if source.PrimaryOutputCount() == 0 || options.DualPrimaryInputs > primaryInputs || options.DCRegisters != nil && len(options.DCRegisters) != registers {
		return nil, ErrDimension
	}
	if source.ObjectCount() > budget.maxObjects {
		return nil, ErrLimit
	}
	result := NewAigManager()
	result.Name = source.Name
	copies := make([][2]AigLit, source.ObjectCount())
	for index := range copies {
		copies[index] = [2]AigLit{InvalidAigLit, InvalidAigLit}
	}
	copies[0] = [2]AigLit{AigConst0, AigConst1}
	for index, id := range source.cis {
		if err = budget.step(result); err != nil {
			return nil, err
		}
		var zeroInput, oneInput AigLit
		switch {
		case index < options.DualPrimaryInputs:
			zeroInput, err = saigDualCI(result, budget)
			if err == nil {
				oneInput, err = saigDualCI(result, budget)
			}
		case index < primaryInputs:
			oneInput, err = saigDualCI(result, budget)
			zeroInput = oneInput.Not()
		default:
			zeroInput, err = saigDualCI(result, budget)
			if err == nil {
				oneInput, err = saigDualCI(result, budget)
			}
			dual := options.DualRegisters
			if options.DCRegisters != nil {
				dual = options.DCRegisters[index-primaryInputs]
			}
			zeroInput = zeroInput.NotIf(!dual)
		}
		if err != nil {
			return nil, err
		}
		zero, andErr := saigDualAnd(result, budget, zeroInput, oneInput.Not())
		if andErr != nil {
			return nil, andErr
		}
		one, andErr := saigDualAnd(result, budget, oneInput, zeroInput.Not())
		if andErr != nil {
			return nil, andErr
		}
		copies[id] = [2]AigLit{zero, one}
	}
	for id, object := range source.objects {
		if err = budget.step(result); err != nil {
			return nil, err
		}
		switch object.Type {
		case AigNone, AigConstant1, AigCI, AigCO:
			continue
		case AigBuffer:
			zero, one, faninErr := saigDualFanin(result, budget, copies, object.Fanin0)
			if faninErr != nil {
				return nil, faninErr
			}
			copies[id] = [2]AigLit{zero, one}
		case AigAND:
			zero0, one0, faninErr := saigDualFanin(result, budget, copies, object.Fanin0)
			if faninErr != nil {
				return nil, faninErr
			}
			zero1, one1, faninErr := saigDualFanin(result, budget, copies, object.Fanin1)
			if faninErr != nil {
				return nil, faninErr
			}
			zero, gateErr := saigDualOr(result, budget, zero0, zero1)
			if gateErr != nil {
				return nil, gateErr
			}
			one, gateErr := saigDualAnd(result, budget, one0, one1)
			if gateErr != nil {
				return nil, gateErr
			}
			copies[id] = [2]AigLit{zero, one}
		default:
			return nil, ErrUnsupported
		}
	}
	miter := AigConst0
	start, end := 0, source.PrimaryOutputCount()
	if options.MiterRegisters {
		start, end = source.PrimaryOutputCount(), source.OutputCount()
	}
	for _, outputID := range source.cos[start:end] {
		zero, one, faninErr := saigDualFanin(result, budget, copies, source.objects[outputID].Fanin0)
		if faninErr != nil {
			return nil, faninErr
		}
		var violation AigLit
		if options.CheckZero {
			violation, err = saigDualAnd(result, budget, zero, one.Not())
		} else if options.CheckOne {
			violation, err = saigDualAnd(result, budget, zero.Not(), one)
		} else {
			violation, err = saigDualAnd(result, budget, zero.Not(), one.Not())
		}
		if err == nil {
			miter, err = saigDualOr(result, budget, miter, violation)
		}
		if err != nil {
			return nil, err
		}
	}
	if err = saigDualCO(result, budget, miter.NotIf(options.ComplementOutput)); err != nil {
		return nil, err
	}
	for index, outputID := range source.cos[source.PrimaryOutputCount():] {
		zero, one, faninErr := saigDualFanin(result, budget, copies, source.objects[outputID].Fanin0)
		if faninErr != nil {
			return nil, faninErr
		}
		dual := options.DualRegisters
		if options.DCRegisters != nil {
			dual = options.DCRegisters[index]
		}
		zero = zero.NotIf(!dual)
		if err = saigDualCO(result, budget, zero); err != nil {
			return nil, err
		}
		if err = saigDualCO(result, budget, one); err != nil {
			return nil, err
		}
	}
	if err = result.SetRegisterCount(2 * registers); err != nil {
		return nil, err
	}
	result.Cleanup()
	if err = result.Check(); err != nil {
		return nil, err
	}
	return result, nil
}

func saigCopyDualSignal(mapping []AigLit, literal AigLit) (AigLit, error) {
	if !literal.Valid() || uint64(literal.ID()) >= uint64(len(mapping)) || !mapping[literal.ID()].Valid() {
		return InvalidAigLit, ErrInvariant
	}
	return mapping[literal.ID()].NotIf(literal.Complemented()), nil
}

// SAIGBlockPropertyCycles returns an owned form of Saig_ManBlockPo. It retains
// the source's transient PO boundary while adding the delay registers, which
// also gates the prefix of old latch-input outputs selected by the C routine.
func SAIGBlockPropertyCycles(source *AigManager, cycles int, options SAIGBlockPropertyOptions) (*AigManager, error) {
	budget, err := newSAIGDualBudget(options.MaxObjects, options.MaxWork, options.Stop)
	if err != nil {
		return nil, err
	}
	if source == nil || cycles <= 0 {
		return nil, ErrDimension
	}
	if err = source.Check(); err != nil {
		return nil, err
	}
	if source.ObjectCount() > budget.maxObjects || cycles > budget.maxObjects-source.ObjectCount() {
		return nil, ErrLimit
	}
	result := NewAigManager()
	result.Name, result.Spec = source.Name, source.Spec
	mapping := make([]AigLit, source.ObjectCount())
	for index := range mapping {
		mapping[index] = InvalidAigLit
	}
	mapping[0] = AigConst1
	for _, id := range source.cis {
		mapping[id], err = saigDualCI(result, budget)
		if err != nil {
			return nil, err
		}
	}
	delays := make([]AigLit, cycles)
	for index := range delays {
		delays[index], err = saigDualCI(result, budget)
		if err != nil {
			return nil, err
		}
	}
	condition := AigConst1
	for _, delay := range delays {
		condition, err = saigDualAnd(result, budget, condition, delay)
		if err != nil {
			return nil, err
		}
	}
	for id, object := range source.objects {
		if err = budget.step(result); err != nil {
			return nil, err
		}
		switch object.Type {
		case AigNone, AigConstant1, AigCI, AigCO:
			continue
		case AigBuffer:
			mapping[id], err = saigCopyDualSignal(mapping, object.Fanin0)
		case AigAND:
			var left, right AigLit
			left, err = saigCopyDualSignal(mapping, object.Fanin0)
			if err == nil {
				right, err = saigCopyDualSignal(mapping, object.Fanin1)
			}
			if err == nil {
				mapping[id], err = saigDualAnd(result, budget, left, right)
			}
		default:
			err = ErrUnsupported
		}
		if err != nil {
			return nil, err
		}
	}
	drivers := make([]AigLit, 0, source.OutputCount()+cycles)
	for _, outputID := range source.cos {
		driver, copyErr := saigCopyDualSignal(mapping, source.objects[outputID].Fanin0)
		if copyErr != nil {
			return nil, copyErr
		}
		drivers = append(drivers, driver)
	}
	drivers = append(drivers, AigConst1)
	if cycles > 1 {
		drivers = append(drivers, delays[:cycles-1]...)
	}
	transientPrimaryOutputs := source.PrimaryOutputCount() + cycles
	for index := 0; index < transientPrimaryOutputs; index++ {
		drivers[index], err = saigDualAnd(result, budget, drivers[index], condition)
		if err != nil {
			return nil, err
		}
	}
	for _, driver := range drivers {
		if err = saigDualCO(result, budget, driver); err != nil {
			return nil, err
		}
	}
	if err = result.SetRegisterCount(source.RegisterCount() + cycles); err != nil {
		return nil, err
	}
	if err = result.SetConstraintCount(source.ConstraintCount()); err != nil {
		return nil, err
	}
	result.Cleanup()
	if err = result.Check(); err != nil {
		return nil, err
	}
	return result, nil
}
