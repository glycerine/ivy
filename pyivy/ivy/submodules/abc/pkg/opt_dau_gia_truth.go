package abc

// Truth-table dispatch and mapped-network derivation from opt/dau/dauGia.c.

import "fmt"

type DauGiaStatistics struct {
	Calls, NonDSD, PrimeBlocks int
}

func FormatDauGiaStatistics(statistics DauGiaStatistics) string {
	return fmt.Sprintf("Calls = %d. NonDSD = %d. Non1Step = %d.\n", statistics.Calls, statistics.NonDSD, statistics.PrimeBlocks)
}

type DauGiaTruthResult struct {
	DauGiaResult
	LargestPrime int
}

// WithDauTruth decomposes a truth table and lowers the resulting formula onto
// an owned graph clone. Inputs may include unused trailing leaves, matching the
// mapped-cut caller in the source. No package-global statistics are retained.
func (m *GiaManager) WithDauTruth(truth *TruthTable, inputs []GiaLit, options DauGiaOptions) (DauGiaTruthResult, error) {
	if m == nil || truth == nil || truth.variables != len(inputs) || len(inputs) > DauMaxVariables {
		return DauGiaTruthResult{}, ErrDimension
	}
	if err := truth.Check(); err != nil {
		return DauGiaTruthResult{}, err
	}
	decomposition, err := DauDSDDecompose(truth, DauDSDDecomposeOptions{
		MaxBytes: options.MaxBytes, MaxNodes: options.MaxNodes, MaxDepth: options.MaxDepth,
		MaxWords: options.MaxWords, Stop: options.Stop,
	})
	if err != nil {
		return DauGiaTruthResult{}, err
	}
	lowered, err := m.WithDauDSD(decomposition.Formula, inputs, options)
	if err != nil {
		return DauGiaTruthResult{}, err
	}
	return DauGiaTruthResult{DauGiaResult: lowered, LargestPrime: decomposition.LargestPrime}, nil
}

type DauGiaDeriveOptions struct {
	DauGiaOptions
	MaxMappingEntries int
	MaxCutObjects     int
}

type DauGiaDeriveResult struct {
	Graph      *GiaManager
	Statistics DauGiaStatistics
}

// DeriveDauGia rebuilds a LUT-mapped graph by evaluating each ordered LUT
// boundary, decomposing its truth, and lowering it into one structurally hashed
// target. Unmapped ANDs are omitted; CIs, mapped roots, buffers, CO order,
// registers, constraints and metadata are preserved as in Dsm_ManDeriveGia.
func (m *GiaManager) DeriveDauGia(mapping *GiaLUTMapping, options DauGiaDeriveOptions) (DauGiaDeriveResult, error) {
	fail := func(err error) (DauGiaDeriveResult, error) { return DauGiaDeriveResult{}, err }
	if m == nil || mapping == nil || options.MaxMappingEntries < 0 || options.MaxCutObjects < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects < 0 || options.MaxBytes < 0 || options.MaxNodes < 0 || options.MaxDepth < 0 || options.MaxWords < 0 {
		return fail(ErrDimension)
	}
	if err := mapping.checkGraph(m); err != nil {
		return fail(err)
	}
	mappingLimit := options.MaxMappingEntries
	if mappingLimit == 0 {
		mappingLimit = 1 << 20
	}
	if len(m.objects) > mappingLimit {
		return fail(ErrLimit)
	}
	if _, err := mapping.Analyze(m, GiaLUTAnalysisOptions{MaxObjects: mappingLimit, MaxVisits: mappingLimit, Stop: options.Stop}); err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	target := NewGiaManager()
	if options.MaxObjects != 0 {
		target.objectLimit = min(options.MaxObjects, int(MaxGiaID)+1)
	}
	target.Name, target.Spec = m.Name, m.Spec
	if err := target.HashStart(); err != nil {
		return fail(err)
	}
	correspondence := make([]GiaLit, len(m.objects))
	for index := range correspondence {
		correspondence[index] = InvalidGiaLit
	}
	correspondence[0] = GiaConst0
	statistics := DauGiaStatistics{}
	lowerOptions := options.DauGiaOptions
	lowerOptions.Balance = true // Dsm_ManDeriveGia always allocates level storage.
	for id, object := range m.objects {
		if stopped() {
			return fail(ErrCancelled)
		}
		if id == 0 {
			continue
		}
		objectID := GiaID(id)
		if object.IsCI() {
			literal, err := target.AppendCI()
			if err != nil {
				return fail(err)
			}
			correspondence[id] = literal
			continue
		}
		if object.IsBuffer() {
			fanin, err := object.Fanin0(objectID)
			if err != nil || !correspondence[fanin.ID()].Valid() {
				return fail(ErrInvariant)
			}
			literal, err := target.AppendBuffer(correspondence[fanin.ID()].NotIf(fanin.Complemented()))
			if err != nil {
				return fail(err)
			}
			correspondence[id] = literal
			continue
		}
		record, mapped := mapping.Record(objectID)
		if !mapped {
			continue
		}
		leaves := make([]GiaLit, len(record.Fanins))
		for index, fanin := range record.Fanins {
			if !correspondence[fanin].Valid() {
				return fail(ErrInvariant)
			}
			leaves[index] = correspondence[fanin]
		}
		cut, err := m.CutTruth(GiaLit(objectID)<<1, record.Fanins, GiaCutTruthOptions{
			MaxObjects: options.MaxCutObjects, MaxMappingEntries: mappingLimit,
			MaxWords: options.MaxWords, Stop: options.Stop,
		})
		if err != nil {
			return fail(err)
		}
		lowered, err := target.WithDauTruth(cut.Table, leaves, lowerOptions)
		if err != nil {
			return fail(err)
		}
		target = lowered.Graph
		correspondence[id] = lowered.Root
		statistics.Calls++
		if lowered.LargestPrime != 0 {
			statistics.NonDSD++
		}
		statistics.PrimeBlocks += lowered.PrimeCount
	}
	for _, id := range m.cos {
		object := m.objects[id]
		fanin, err := object.Fanin0(id)
		if err != nil || !correspondence[fanin.ID()].Valid() {
			return fail(ErrInvariant)
		}
		if _, err = target.AppendCO(correspondence[fanin.ID()].NotIf(fanin.Complemented())); err != nil {
			return fail(err)
		}
	}
	target.HashStop()
	if err := target.SetRegisterCount(m.registers); err != nil {
		return fail(err)
	}
	if err := target.SetConstraintCount(m.constraints); err != nil {
		return fail(err)
	}
	cleaned, _, err := target.Cleanup()
	if err != nil {
		return fail(err)
	}
	return DauGiaDeriveResult{Graph: cleaned, Statistics: statistics}, nil
}
