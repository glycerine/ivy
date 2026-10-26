package abc

// Fault-pattern persistence and reporting from src/sat/bmc/bmcFault.c.

import (
	"os"
	"slices"
)

type FaultPatternOptions struct {
	MaxBits  int
	MaxBytes int
	Stop     func() bool
}

func faultPatternLimits(options FaultPatternOptions) (FaultPatternOptions, error) {
	if options.MaxBits < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxBits == 0 {
		options.MaxBits = 16 << 20
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 32 << 20
	}
	return options, nil
}

func MarshalFaultPatterns(patterns []bool, width int, options FaultPatternOptions) ([]byte, error) {
	options, err := faultPatternLimits(options)
	if err != nil || width <= 0 || len(patterns)%width != 0 {
		return nil, ErrDimension
	}
	if len(patterns) > options.MaxBits || len(patterns)+len(patterns)/width > options.MaxBytes {
		return nil, ErrLimit
	}
	result := make([]byte, 0, len(patterns)+len(patterns)/width)
	for i, bit := range patterns {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if bit {
			result = append(result, '1')
		} else {
			result = append(result, '0')
		}
		if (i+1)%width == 0 {
			result = append(result, '\n')
		}
	}
	return result, nil
}

func ParseFaultPatterns(data []byte, options FaultPatternOptions) ([]bool, error) {
	options, err := faultPatternLimits(options)
	if err != nil {
		return nil, err
	}
	if len(data) > options.MaxBytes {
		return nil, ErrLimit
	}
	result := make([]bool, 0, min(len(data), options.MaxBits))
	for _, character := range data {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		switch character {
		case '0':
			result = append(result, false)
		case '1':
			result = append(result, true)
		case ' ', '\t', '\r', '\n':
			continue
		default:
			return nil, ErrDimension
		}
		if len(result) > options.MaxBits {
			return nil, ErrLimit
		}
	}
	return result, nil
}

func WriteFaultPatterns(path string, patterns []bool, width int, options FaultPatternOptions) error {
	if path == "" {
		return ErrDimension
	}
	data, err := MarshalFaultPatterns(patterns, width, options)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o666)
}

func ReadFaultPatterns(path string, options FaultPatternOptions) ([]bool, error) {
	if path == "" {
		return nil, ErrDimension
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseFaultPatterns(data, options)
}

// MarshalFaultDelayPatterns converts rows laid out as [RO0, PI0, PI1] into
// two conventional [PI, state] lines, deriving state1 from the source graph.
func MarshalFaultDelayPatterns(graph *GiaManager, patterns []bool, options FaultPatternOptions) ([]byte, error) {
	options, err := faultPatternLimits(options)
	if err != nil || graph == nil || graph.RegisterCount() == 0 {
		return nil, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	width := 2*graph.PrimaryInputCount() + graph.RegisterCount()
	if width <= 0 || len(patterns)%width != 0 || len(patterns) > options.MaxBits {
		return nil, ErrDimension
	}
	lineWidth := graph.PrimaryInputCount() + graph.RegisterCount()
	rows := len(patterns) / width
	if 2*rows*(lineWidth+1) > options.MaxBytes {
		return nil, ErrLimit
	}
	result := make([]byte, 0, 2*rows*(lineWidth+1))
	appendBits := func(bits []bool) {
		for _, bit := range bits {
			if bit {
				result = append(result, '1')
			} else {
				result = append(result, '0')
			}
		}
	}
	for row := 0; row < rows; row++ {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		entry := patterns[row*width : (row+1)*width]
		registers := entry[:graph.RegisterCount()]
		first := entry[graph.RegisterCount() : graph.RegisterCount()+graph.PrimaryInputCount()]
		second := entry[graph.RegisterCount()+graph.PrimaryInputCount():]
		appendBits(first)
		appendBits(registers)
		result = append(result, '\n')
		_, next, err := graph.Step(first, registers)
		if err != nil {
			return nil, err
		}
		appendBits(second)
		appendBits(next)
		result = append(result, '\n')
	}
	return result, nil
}

type GiaFaultRunStatistics struct {
	Name                                     string
	PrimaryInputs, PrimaryOutputs, Registers int
	Gates, Variables, Clauses                int
	Conflicts                                uint64
	Iterations                               int
}

func FaultRunStatistics(graph *GiaManager, solver *SatSolver, iterations int) (GiaFaultRunStatistics, error) {
	if graph == nil || solver == nil || iterations < 0 {
		return GiaFaultRunStatistics{}, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return GiaFaultRunStatistics{}, err
	}
	return GiaFaultRunStatistics{
		Name: graph.Name, PrimaryInputs: graph.PrimaryInputCount(), PrimaryOutputs: graph.PrimaryOutputCount(),
		Registers: graph.RegisterCount(), Gates: graph.AndCount(), Variables: solver.VariableCount(),
		Clauses: solver.OriginalClauseCount(), Conflicts: solver.ConflictCount(), Iterations: iterations,
	}, nil
}

// DuplicateGiaWithPrimaryInputs inserts new PIs before the RO suffix.
func DuplicateGiaWithPrimaryInputs(source *GiaManager, count int, options GiaCofactorOptions) (*GiaManager, []GiaID, error) {
	if source == nil || count < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if err := source.Check(); err != nil {
		return nil, nil, err
	}
	if source.ObjectCount() > options.MaxMappingEntries || source.ObjectCount()+count > options.MaxObjects {
		return nil, nil, ErrLimit
	}
	graph, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, nil, err
	}
	graph.Name, graph.Spec = source.Name, source.Spec
	mapping := make([]GiaLit, source.ObjectCount())
	for i := 0; i < source.PrimaryInputCount(); i++ {
		mapping[source.cis[i]], err = graph.AppendCI()
		if err != nil {
			return nil, nil, err
		}
	}
	extra := make([]GiaID, count)
	for i := range extra {
		if options.Stop != nil && options.Stop() {
			return nil, nil, ErrCancelled
		}
		literal, err := graph.AppendCI()
		if err != nil {
			return nil, nil, err
		}
		extra[i] = literal.ID()
	}
	for i := source.PrimaryInputCount(); i < source.InputCount(); i++ {
		mapping[source.cis[i]], err = graph.AppendCI()
		if err != nil {
			return nil, nil, err
		}
	}
	for i, object := range source.objects {
		if object.IsAnd() {
			mapping[i], err = appendGiaCopiedGate(source, graph, GiaID(i), mapping)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	for _, id := range source.cos {
		driver, _ := source.objects[id].Fanin0(id)
		copied, err := mappedGiaSignal(mapping, driver)
		if err != nil {
			return nil, nil, err
		}
		if _, err = graph.AppendCO(copied); err != nil {
			return nil, nil, err
		}
	}
	if err = graph.SetRegisterCount(source.RegisterCount()); err != nil {
		return nil, nil, err
	}
	if err = graph.SetConstraintCount(source.ConstraintCount()); err != nil {
		return nil, nil, err
	}
	if err = graph.Check(); err != nil {
		return nil, nil, err
	}
	return graph, slices.Clone(extra), nil
}
