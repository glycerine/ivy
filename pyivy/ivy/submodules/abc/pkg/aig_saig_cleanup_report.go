package abc

// Sequential cleanup diagnostics from src/aig/saig/saigScl.c.

type AigSequentialDiagnosticOptions struct {
	MaxObjects int
	Stop       func() bool
}

type AigUselessRegisterReport struct {
	Registers                     int
	RegistersDrivingPrimaryOutput int
	RegistersDrivenByPrimaryInput int
}

// UselessRegisterReport ports Saig_ManReportUselessRegisters without CIO-ID
// scratch or stdout. A register output counts only when it is the regular
// driver of a PO and has one total graph reference. A PI similarly counts only
// when it is the regular driver of an RI and has one total graph reference.
func (m *AigManager) UselessRegisterReport(options AigSequentialDiagnosticOptions) (AigUselessRegisterReport, error) {
	var result AigUselessRegisterReport
	if m == nil || options.MaxObjects < 0 {
		return result, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if err := m.Check(); err != nil {
		return result, err
	}
	if len(m.objects) > options.MaxObjects {
		return result, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return result, ErrCancelled
	}
	result.Registers = m.registers
	registerOutput := make([]bool, len(m.objects))
	for _, id := range m.cis[m.PrimaryInputCount():] {
		registerOutput[id] = true
	}
	primaryInput := make([]bool, len(m.objects))
	for _, id := range m.cis[:m.PrimaryInputCount()] {
		primaryInput[id] = true
	}
	for _, id := range m.cos[:m.PrimaryOutputCount()] {
		if stopped() {
			return AigUselessRegisterReport{}, ErrCancelled
		}
		driver := m.objects[id].Fanin0.ID()
		if registerOutput[driver] && m.objects[driver].RefCount == 1 {
			result.RegistersDrivingPrimaryOutput++
		}
	}
	for _, id := range m.cos[m.PrimaryOutputCount():] {
		if stopped() {
			return AigUselessRegisterReport{}, ErrCancelled
		}
		driver := m.objects[id].Fanin0.ID()
		if primaryInput[driver] && m.objects[driver].RefCount == 1 {
			result.RegistersDrivenByPrimaryInput++
		}
	}
	return result, nil
}

// ComplementedRegisterPairs ports Saig_ManReportComplements. Despite its C
// name, the observable result counts every RI after the first whose regular
// driver object has already appeared; edge phase is deliberately ignored.
func (m *AigManager) ComplementedRegisterPairs(options AigSequentialDiagnosticOptions) (int, error) {
	if m == nil || options.MaxObjects < 0 {
		return 0, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if err := m.Check(); err != nil {
		return 0, err
	}
	if m.registers == 0 {
		return 0, ErrDimension
	}
	if len(m.objects) > options.MaxObjects {
		return 0, ErrLimit
	}
	seen := make([]bool, len(m.objects))
	duplicates := 0
	for _, id := range m.cos[m.PrimaryOutputCount():] {
		if options.Stop != nil && options.Stop() {
			return 0, ErrCancelled
		}
		driver := m.objects[id].Fanin0.ID()
		if seen[driver] {
			duplicates++
		} else {
			seen[driver] = true
		}
	}
	return duplicates, nil
}
