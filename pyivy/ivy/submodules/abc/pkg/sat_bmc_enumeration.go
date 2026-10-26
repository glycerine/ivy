package abc

// Assignment justification and enumeration from src/sat/bmc/bmcEnum.c.

import "slices"

// DeriveGiaAssignmentJustification implements Gia_ManDeriveOne. Each returned
// output is true on the supplied CI assignment and implies that the matching
// source output retains the value observed on that assignment.
func (manager *GiaManager) DeriveAssignmentJustification(values []bool, options GiaCofactorOptions) (*GiaManager, error) {
	if manager == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if err := manager.Check(); err != nil {
		return nil, err
	}
	if len(values) != len(manager.cis) {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 4 << 20
	}
	if len(manager.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	for id, object := range manager.objects {
		if object.IsAnd() && (object.IsBuffer() || object.IsXor()) {
			return nil, ErrUnsupported
		}
		if _, mux := manager.MuxControl(GiaID(id)); mux {
			return nil, ErrUnsupported
		}
	}
	phases := make([]bool, len(manager.objects))
	for index, id := range manager.cis {
		phases[id] = values[index]
	}
	readPhase := func(literal GiaLit) bool {
		return phases[literal.ID()] != literal.Complemented()
	}
	for id, object := range manager.objects {
		if !object.IsAnd() {
			continue
		}
		if stopped() {
			return nil, ErrCancelled
		}
		left, _ := object.Fanin0(GiaID(id))
		right, _ := object.Fanin1(GiaID(id))
		phases[id] = readPhase(left) && readPhase(right)
	}
	marked := make([]bool, len(manager.objects))
	for _, id := range manager.cos {
		driver, err := manager.objects[id].Fanin0(id)
		if err != nil {
			return nil, err
		}
		marked[driver.ID()] = true
	}
	for id := len(manager.objects) - 1; id > 0; id-- {
		object := manager.objects[id]
		if !marked[id] || !object.IsAnd() {
			continue
		}
		left, _ := object.Fanin0(GiaID(id))
		right, _ := object.Fanin1(GiaID(id))
		leftPhase, rightPhase := readPhase(left), readPhase(right)
		switch {
		case leftPhase == rightPhase:
			marked[left.ID()], marked[right.ID()] = true, true
		case leftPhase:
			marked[right.ID()] = true
		case rightPhase:
			marked[left.ID()] = true
		}
	}
	// The source assigns literal zero to the constant object's temporary Value,
	// which cannot represent the true equality predicate "constant is zero".
	// Reject a justification that actually depends on this broken case.
	if marked[0] {
		return nil, ErrUnsupported
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name, target.Spec = manager.Name, manager.Spec
	mapping := make([]GiaLit, len(manager.objects))
	for index := range mapping {
		mapping[index] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for index, id := range manager.cis {
		literal, appendErr := target.AppendCI()
		if appendErr != nil {
			return nil, appendErr
		}
		mapping[id] = literal.NotIf(!values[index])
	}
	for id, object := range manager.objects {
		if !marked[id] || !object.IsAnd() {
			continue
		}
		if stopped() {
			return nil, ErrCancelled
		}
		left, _ := object.Fanin0(GiaID(id))
		right, _ := object.Fanin1(GiaID(id))
		leftPhase, rightPhase := readPhase(left), readPhase(right)
		if leftPhase == rightPhase {
			if phases[id] {
				mapping[id], err = target.AppendAndSimplified(mapping[left.ID()], mapping[right.ID()])
			} else {
				mapping[id], err = target.AppendOrSimplified(mapping[left.ID()], mapping[right.ID()])
			}
		} else if leftPhase {
			mapping[id] = mapping[right.ID()]
		} else {
			mapping[id] = mapping[left.ID()]
		}
		if err != nil {
			return nil, err
		}
	}
	for _, id := range manager.cos {
		driver, faninErr := manager.objects[id].Fanin0(id)
		if faninErr != nil || mapping[driver.ID()] == InvalidGiaLit {
			return nil, ErrInvariant
		}
		if _, err = target.AppendCO(mapping[driver.ID()]); err != nil {
			return nil, err
		}
	}
	if err = target.SetRegisterCount(manager.registers); err != nil {
		return nil, err
	}
	return target, nil
}

// DeriveZeroAssignmentJustification implements Gia_ManDeriveOneTest2.
func (manager *GiaManager) DeriveZeroAssignmentJustification(options GiaCofactorOptions) (*GiaManager, error) {
	if manager == nil {
		return nil, ErrDimension
	}
	return manager.DeriveAssignmentJustification(make([]bool, manager.InputCount()), options)
}

type GiaJustificationEnumerationOptions struct {
	Output                                int
	Desired                               bool
	MaxAssignments, MaxTotalObjects       int
	MaxObjects, MaxMappingEntries         int
	MaxVariables, MaxClauses, MaxLiterals int
	SAT                                   SatLimits
	Stop                                  func() bool
}

type GiaJustificationRecord struct {
	Inputs        []bool
	Observed      []bool
	Justification *GiaManager
}

type GiaJustificationEnumeration struct {
	Records                      []GiaJustificationRecord
	Status                       SatStatus
	Reason                       SatStopReason
	Stats                        SatStats
	Complete                     bool
	Iterations                   int
	Variables, Clauses, Literals int
	JustificationObjects         int
}

func normalizeGiaJustificationEnumerationOptions(options GiaJustificationEnumerationOptions) (GiaJustificationEnumerationOptions, error) {
	if options.Output < 0 || options.MaxAssignments < 0 || options.MaxTotalObjects < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
		return options, ErrDimension
	}
	if options.MaxAssignments == 0 {
		options.MaxAssignments = 10000
	}
	if options.MaxTotalObjects == 0 {
		options.MaxTotalObjects = 16 << 20
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 4 << 20
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
	return options, nil
}

func giaAddJustificationCNF(loader *GiaLazyCNF, source, justification *GiaManager, output int) (SatLit, error) {
	if loader == nil || source == nil || justification == nil || output < 0 || output >= justification.OutputCount() || justification.InputCount() != source.InputCount() {
		return 0, ErrDimension
	}
	mapping := make([]SatLit, justification.ObjectCount())
	constant, err := loader.satLiteral(GiaConst0)
	if err != nil {
		return 0, err
	}
	mapping[0] = constant
	for index, id := range justification.cis {
		sourceLiteral, literalErr := MakeGiaLit(source.cis[index], false)
		if literalErr != nil {
			return 0, literalErr
		}
		mapping[id], err = loader.satLiteral(sourceLiteral)
		if err != nil {
			return 0, err
		}
	}
	read := func(literal GiaLit) SatLit {
		mapped := mapping[literal.ID()]
		if literal.Complemented() {
			mapped = mapped.Not()
		}
		return mapped
	}
	for id, object := range justification.objects {
		if !object.IsAnd() {
			continue
		}
		if loader.solver.VariableCount() == loader.options.MaxVariables {
			return 0, ErrLimit
		}
		variable := loader.solver.VariableCount()
		if err = loader.solver.SetVariableCount(variable + 1); err != nil {
			return 0, err
		}
		outputLiteral := giaMeshLiteral(variable, false)
		mapping[id] = outputLiteral
		left, _ := object.Fanin0(GiaID(id))
		right, _ := object.Fanin1(GiaID(id))
		leftLiteral, rightLiteral := read(left), read(right)
		for _, clause := range [][]SatLit{
			{outputLiteral.Not(), leftLiteral},
			{outputLiteral.Not(), rightLiteral},
			{outputLiteral, leftLiteral.Not(), rightLiteral.Not()},
		} {
			if err = loader.addClause(clause); err != nil {
				return 0, err
			}
		}
	}
	coID := justification.cos[output]
	driver, err := justification.objects[coID].Fanin0(coID)
	if err != nil {
		return 0, err
	}
	return read(driver), nil
}

// EnumerateAssignmentJustifications implements Gia_ManDeriveOneTest without
// stdout. It constrains the selected output, derives a justification for each
// SAT assignment, adds exact Tseitin definitions for that formula, and carries
// its negated root as a later assumption so an entire covered region is
// excluded rather than only one model.
func (manager *GiaManager) EnumerateAssignmentJustifications(options GiaJustificationEnumerationOptions) (GiaJustificationEnumeration, error) {
	fail := func(err error) (GiaJustificationEnumeration, error) { return GiaJustificationEnumeration{}, err }
	if manager == nil {
		return fail(ErrDimension)
	}
	options, err := normalizeGiaJustificationEnumerationOptions(options)
	if err != nil {
		return fail(err)
	}
	if err = manager.Check(); err != nil {
		return fail(err)
	}
	if options.Output >= manager.OutputCount() {
		return fail(ErrDimension)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	loader, err := NewGiaLazyCNF(manager, GiaLazyCNFOptions{
		MaxVariables: options.MaxVariables, MaxClauses: options.MaxClauses,
		MaxLiterals: options.MaxLiterals, Stop: options.Stop,
	})
	if err != nil {
		return fail(err)
	}
	defer loader.Close()
	coID := manager.cos[options.Output]
	driver, err := manager.objects[coID].Fanin0(coID)
	if err != nil {
		return fail(err)
	}
	if err = loader.LoadCone(driver); err != nil {
		return fail(err)
	}
	outputLiteral, err := loader.satLiteral(driver)
	if err != nil {
		return fail(err)
	}
	if !options.Desired {
		outputLiteral = outputLiteral.Not()
	}
	inputVariables := make([]SatVar, len(manager.cis))
	for index, id := range manager.cis {
		literal, literalErr := MakeGiaLit(id, false)
		if literalErr != nil {
			return fail(literalErr)
		}
		encoded, literalErr := loader.satLiteral(literal)
		if literalErr != nil {
			return fail(literalErr)
		}
		inputVariables[index] = encoded.Var()
	}
	assumptions := []SatLit{outputLiteral}
	limits := options.SAT
	callerStop := limits.Stop
	limits.Stop = func() bool { return stopped() || callerStop != nil && callerStop() }
	result := GiaJustificationEnumeration{Records: []GiaJustificationRecord{}, Status: SatUnknown}
	for {
		if stopped() {
			return fail(ErrCancelled)
		}
		result.Iterations++
		answer, solveErr := loader.solver.Solve(assumptions, limits)
		if solveErr != nil {
			return fail(solveErr)
		}
		result.Status, result.Reason, result.Stats = answer.Status, answer.Reason, answer.Stats
		if answer.Status == SatUnsatisfiable {
			result.Complete = true
			break
		}
		if answer.Status != SatSatisfiable {
			break
		}
		if len(result.Records) == options.MaxAssignments {
			result.Variables = loader.solver.VariableCount()
			result.Clauses, result.Literals = len(loader.clauses), loader.literalCount
			return result, ErrLimit
		}
		inputs := make([]bool, len(inputVariables))
		for index, variable := range inputVariables {
			if int(variable) >= len(answer.Model) {
				return fail(ErrInvariant)
			}
			inputs[index] = answer.Model[variable]
		}
		observed, evalErr := manager.Evaluate(inputs)
		if evalErr != nil || observed[options.Output] != options.Desired {
			return fail(ErrInvariant)
		}
		justification, deriveErr := manager.DeriveAssignmentJustification(inputs, GiaCofactorOptions{
			MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries, Stop: options.Stop,
		})
		if deriveErr != nil {
			return fail(deriveErr)
		}
		if justification.ObjectCount() > options.MaxTotalObjects-result.JustificationObjects {
			return fail(ErrLimit)
		}
		root, addErr := giaAddJustificationCNF(loader, manager, justification, options.Output)
		if addErr != nil {
			return fail(addErr)
		}
		assumptions = append(assumptions, root.Not())
		result.JustificationObjects += justification.ObjectCount()
		result.Records = append(result.Records, GiaJustificationRecord{
			Inputs: slices.Clone(inputs), Observed: slices.Clone(observed), Justification: justification,
		})
	}
	result.Variables = loader.solver.VariableCount()
	result.Clauses, result.Literals = len(loader.clauses), loader.literalCount
	return result, nil
}
