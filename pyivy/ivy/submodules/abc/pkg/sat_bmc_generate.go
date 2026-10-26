package abc

// Packed random simulation and SAT-generation diagnostics from
// src/sat/bmc/bmcGen.c.

import "slices"

type GiaPackedSimulationOptions struct {
	Words, MaxWords, MaxEntries int
	Stop                        func() bool
}

// GiaPackedSimulation owns object-major simulation words. Constants remain
// zero, CIs use the reset Gia_ManRandomW stream, and gates/COs are propagated
// in object order.
type GiaPackedSimulation struct {
	words, objects int
	values         []uint64
}

func (simulation GiaPackedSimulation) Words() int   { return simulation.words }
func (simulation GiaPackedSimulation) Objects() int { return simulation.objects }
func (simulation GiaPackedSimulation) Values() []uint64 {
	return slices.Clone(simulation.values)
}

func (simulation GiaPackedSimulation) ObjectWords(id GiaID) ([]uint64, error) {
	if simulation.words < 1 || uint64(id) >= uint64(simulation.objects) || len(simulation.values) != simulation.words*simulation.objects {
		return nil, ErrInvalidLiteral
	}
	begin := int(id) * simulation.words
	return slices.Clone(simulation.values[begin : begin+simulation.words]), nil
}

func (simulation GiaPackedSimulation) Bit(id GiaID, sample int) (bool, error) {
	if sample < 0 || simulation.words < 1 || sample >= 64*simulation.words || uint64(id) >= uint64(simulation.objects) || len(simulation.values) != simulation.words*simulation.objects {
		return false, ErrDimension
	}
	word := simulation.values[int(id)*simulation.words+sample/64]
	return word>>uint(sample&63)&1 != 0, nil
}

func normalizeGiaPackedSimulationOptions(options GiaPackedSimulationOptions) (GiaPackedSimulationOptions, error) {
	if options.Words < 1 || options.MaxWords < 0 || options.MaxEntries < 0 {
		return options, ErrDimension
	}
	if options.MaxWords == 0 {
		options.MaxWords = 1 << 16
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 16 << 20
	}
	if options.Words > options.MaxWords {
		return options, ErrLimit
	}
	return options, nil
}

// SimulateRandomWords implements Gia_ManMoFindSimulate without mutating GIA
// simulation scratch. Native buffers, XORs and muxes are rejected because the
// source routine treats every Gia_ObjIsAnd object as an ordinary AND.
func (manager *GiaManager) SimulateRandomWords(options GiaPackedSimulationOptions) (GiaPackedSimulation, error) {
	fail := func(err error) (GiaPackedSimulation, error) { return GiaPackedSimulation{}, err }
	if manager == nil {
		return fail(ErrDimension)
	}
	if err := manager.Check(); err != nil {
		return fail(err)
	}
	options, err := normalizeGiaPackedSimulationOptions(options)
	if err != nil {
		return fail(err)
	}
	entries, err := checkedSize(0, len(manager.objects), options.Words, options.MaxEntries)
	if err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	for id, object := range manager.objects {
		if object.IsAnd() && (object.IsBuffer() || object.IsXor()) {
			return fail(ErrUnsupported)
		}
		if _, mux := manager.MuxControl(GiaID(id)); mux {
			return fail(ErrUnsupported)
		}
	}
	result := GiaPackedSimulation{words: options.Words, objects: len(manager.objects), values: make([]uint64, entries)}
	z, w := uint32(3716960521), uint32(2174103536)
	random := func() uint32 {
		z = 36969*(z&65535) + (z >> 16)
		w = 18000*(w&65535) + (w >> 16)
		return (z << 16) + w
	}
	random() // Gia_ManRandomW(1) leaves the generator one step after reset.
	randomWord := func() uint64 { return uint64(random())<<32 | uint64(random()) }
	for _, id := range manager.cis {
		for word := 0; word < options.Words; word++ {
			if stopped() {
				return fail(ErrCancelled)
			}
			result.values[int(id)*options.Words+word] = randomWord()
		}
	}
	read := func(literal GiaLit, word int) uint64 {
		value := result.values[int(literal.ID())*options.Words+word]
		if literal.Complemented() {
			value = ^value
		}
		return value
	}
	for id, object := range manager.objects {
		if !object.IsAnd() {
			continue
		}
		left, _ := object.Fanin0(GiaID(id))
		right, _ := object.Fanin1(GiaID(id))
		for word := 0; word < options.Words; word++ {
			if stopped() {
				return fail(ErrCancelled)
			}
			result.values[id*options.Words+word] = read(left, word) & read(right, word)
		}
	}
	for _, id := range manager.cos {
		object := manager.objects[id]
		driver, _ := object.Fanin0(id)
		for word := 0; word < options.Words; word++ {
			if stopped() {
				return fail(ErrCancelled)
			}
			result.values[int(id)*options.Words+word] = read(driver, word)
		}
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return result, nil
}

type GiaSATEnumerationOptions struct {
	Words, MaxWords, MaxEntries, MaxSamples int
	MaxVariables, MaxClauses, MaxLiterals   int
	SAT                                     SatLimits
	Stop                                    func() bool
}

type GiaSATEnumeration struct {
	OutputBits []bool
	Statuses   []SatStatus
	SATCount   int
	Stats      SatStats
}

// EnumerateSATDiagnostic implements the active behavior of Gia_ManTestSatEnum.
// The source computes simulation-derived polarity candidates but its polarity
// call is commented out; consequently it solves the same output-constrained
// formula for every simulation bit. Go returns those bits and statuses instead
// of printing them or reporting wall-clock time.
func (manager *GiaManager) EnumerateSATDiagnostic(output int, options GiaSATEnumerationOptions) (GiaSATEnumeration, error) {
	fail := func(err error) (GiaSATEnumeration, error) { return GiaSATEnumeration{}, err }
	if manager == nil || output < 0 || output >= manager.PrimaryOutputCount() || options.Words < 0 || options.MaxSamples < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
		return fail(ErrDimension)
	}
	if options.Words == 0 {
		options.Words = 1
	}
	if options.MaxSamples == 0 {
		options.MaxSamples = 1 << 20
	}
	if options.Words > options.MaxSamples/64 {
		return fail(ErrLimit)
	}
	stopped := func() bool {
		return options.Stop != nil && options.Stop()
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	simulation, err := manager.SimulateRandomWords(GiaPackedSimulationOptions{Words: options.Words, MaxWords: options.MaxWords, MaxEntries: options.MaxEntries, Stop: options.Stop})
	if err != nil {
		return fail(err)
	}
	loader, err := NewGiaLazyCNF(manager, GiaLazyCNFOptions{MaxVariables: options.MaxVariables, MaxClauses: options.MaxClauses, MaxLiterals: options.MaxLiterals, Stop: options.Stop})
	if err != nil {
		return fail(err)
	}
	defer loader.Close()
	coID := manager.cos[output]
	driver, err := manager.objects[coID].Fanin0(coID)
	if err != nil {
		return fail(err)
	}
	if err = loader.LoadCone(driver); err != nil {
		return fail(err)
	}
	assumption, err := loader.satLiteral(driver)
	if err != nil {
		return fail(err)
	}
	limits := options.SAT
	callerStop := limits.Stop
	limits.Stop = func() bool {
		return stopped() || callerStop != nil && callerStop()
	}
	samples := 64 * options.Words
	result := GiaSATEnumeration{OutputBits: make([]bool, samples), Statuses: make([]SatStatus, samples)}
	for sample := 0; sample < samples; sample++ {
		if stopped() {
			return fail(ErrCancelled)
		}
		result.OutputBits[sample], err = simulation.Bit(coID, sample)
		if err != nil {
			return fail(err)
		}
		answer, solveErr := loader.solver.Solve([]SatLit{assumption}, limits)
		if solveErr != nil {
			return fail(solveErr)
		}
		result.Statuses[sample] = answer.Status
		result.Stats = answer.Stats
		if answer.Status == SatSatisfiable {
			result.SATCount++
		}
	}
	return result, nil
}
