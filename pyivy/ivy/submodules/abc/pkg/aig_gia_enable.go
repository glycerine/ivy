package abc

// Structural signal detection from src/aig/gia/giaEnable.c:
// Gia_CollectSuper[_rec], Gia_ManPrintSignals and Gia_ManDetectSeqSignals.

import (
	"fmt"
	"strings"
)

type GiaSequentialSignals struct {
	SetCandidates, ResetCandidates, EnableCandidates []uint64
	RegistersWithSharedSetReset, RegistersWithEnable int
}

type GiaSignalFrequencyEntry struct {
	ID         GiaID
	References uint64
	Frequency  uint64
}

type GiaSignalFrequencyReport struct {
	Label                     string
	RepeatedSignals, TotalUse uint64
	Frequent                  []GiaSignalFrequencyEntry
}

// Format preserves Gia_ManPrintSignals' deterministic diagnostic layout.
func (r GiaSignalFrequencyReport) Format() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s (total = %d  driven = %d)\n", r.Label, r.RepeatedSignals, r.TotalUse)
	for i, entry := range r.Frequent {
		fmt.Fprintf(&b, "%3d :   Obj = %6d   Refs = %6d   Freq = %6d\n", i+1, entry.ID, entry.References, entry.Frequency)
	}
	return b.String()
}

// SignalFrequencyReport owns the information printed by Gia_ManPrintSignals.
// Frequencies above one contribute to the summary and frequencies above ten
// appear in increasing object-ID order. No manager scratch or stdout is used.
func (m *GiaManager) SignalFrequencyReport(frequencies []uint64, label string, maxObjects int) (GiaSignalFrequencyReport, error) {
	var result GiaSignalFrequencyReport
	if m == nil || maxObjects < 0 || len(frequencies) != len(m.objects) {
		return result, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if len(m.objects) > maxObjects {
		return result, ErrLimit
	}
	refs, err := m.ReferenceCounts(maxObjects)
	if err != nil {
		return result, err
	}
	result.Label = label
	for id, frequency := range frequencies {
		if frequency > 1 {
			result.RepeatedSignals++
			result.TotalUse += frequency
		}
		if frequency > 10 {
			result.Frequent = append(result.Frequent, GiaSignalFrequencyEntry{ID: GiaID(id), References: refs[id], Frequency: frequency})
		}
	}
	return result, nil
}

// CollectSupergate ports both Gia_CollectSuper and Gia_CollectSuper_rec.
// recursive=false returns unique immediate fanin IDs, ignoring their phase.
// recursive=true flattens regular ANDs in left-to-right DFS order and stops at
// complemented edges and CIs. Zero maxObjects selects one million.
func (m *GiaManager) CollectSupergate(root GiaLit, recursive bool, maxObjects int) ([]GiaID, error) {
	if m == nil || maxObjects < 0 || !m.validSignal(root) || root.Complemented() && !recursive {
		return nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if len(m.objects) > maxObjects {
		return nil, ErrLimit
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	seen := make([]bool, len(m.objects))
	result := make([]GiaID, 0)
	add := func(id GiaID) {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	if !recursive {
		obj := m.objects[root.ID()]
		if !obj.IsAnd() {
			add(root.ID())
			return result, nil
		}
		a, err := obj.Fanin0(root.ID())
		if err != nil {
			return nil, err
		}
		b, err := obj.Fanin1(root.ID())
		if err != nil {
			return nil, err
		}
		add(a.ID())
		add(b.ID())
		return result, nil
	}
	stack := []GiaLit{root}
	for len(stack) > 0 {
		last := len(stack) - 1
		lit := stack[last]
		stack = stack[:last]
		obj := m.objects[lit.ID()]
		if lit.Complemented() || obj.IsCI() {
			add(lit.ID())
			continue
		}
		if !obj.IsAnd() || obj.IsBuffer() {
			return nil, ErrDimension
		}
		a, err := obj.Fanin0(lit.ID())
		if err != nil {
			return nil, err
		}
		b, err := obj.Fanin1(lit.ID())
		if err != nil {
			return nil, err
		}
		stack = append(stack, b, a)
	}
	return result, nil
}

// superInputs preserves the active C collector: unique immediate fanin IDs,
// ignoring literal phase. C's recursive flattening call is commented out.
func (m *GiaManager) superInputs(id GiaID) ([]GiaID, error) {
	return m.CollectSupergate(GiaLit(id)<<1, false, len(m.objects))
}

// DetectSequentialSignals counts ABC's structural set/reset and enable
// candidates. These are syntactic diagnostics, not verified signal functions:
// feedback comparison ignores phase, and immediate fanin frequencies ignore
// phase within each gate. Counts are indexed by object ID and caller-owned.
// Zero maxObjects selects one million; at least one register is required.
func (m *GiaManager) DetectSequentialSignals(maxObjects int) (GiaSequentialSignals, error) {
	var result GiaSequentialSignals
	if maxObjects < 0 {
		return result, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if err := m.Check(); err != nil {
		return result, err
	}
	if m.registers == 0 {
		return result, ErrDimension
	}
	if len(m.objects) > maxObjects {
		return result, ErrLimit
	}
	result.SetCandidates = make([]uint64, len(m.objects))
	result.ResetCandidates = make([]uint64, len(m.objects))
	result.EnableCandidates = make([]uint64, len(m.objects))
	candidates := make([][]GiaID, m.registers)
	for reg, ri := range m.cos[m.PrimaryOutputCount():] {
		driver, err := m.objects[ri].Fanin0(ri)
		if err != nil {
			return GiaSequentialSignals{}, err
		}
		if !m.objects[driver.ID()].IsAnd() {
			continue
		}
		immediate, err := m.superInputs(driver.ID())
		if err != nil {
			return GiaSequentialSignals{}, err
		}
		candidates[reg] = immediate
		counts := result.ResetCandidates
		if driver.Complemented() {
			counts = result.SetCandidates
		}
		for _, id := range immediate {
			counts[id]++
		}
		control, yes, no, ok, err := m.RecognizeMux(driver.ID())
		if err != nil {
			return GiaSequentialSignals{}, err
		}
		ro := m.cis[m.PrimaryInputCount()+reg]
		if ok && (yes.ID() == ro || no.ID() == ro) {
			result.EnableCandidates[control.ID()]++
			result.RegistersWithEnable++
		}
	}
	for _, immediate := range candidates {
		for _, id := range immediate {
			if result.SetCandidates[id] > 1 || result.ResetCandidates[id] > 1 {
				result.RegistersWithSharedSetReset++
				break
			}
		}
	}
	return result, nil
}

// SequentialSignalsAboveFanout ports Gia_ManDetectSeqSignalsWithFanout. A
// candidate is included only when its register-use count is strictly greater
// than threshold. IDs are deduplicated across categories while preserving
// C's set, reset, enable category order and increasing IDs in each category.
// Constant zero is excluded. These remain syntactic candidate diagnostics.
func (m *GiaManager) SequentialSignalsAboveFanout(threshold, maxObjects int) ([]GiaID, error) {
	if threshold < 0 {
		return nil, ErrDimension
	}
	signals, err := m.DetectSequentialSignals(maxObjects)
	if err != nil {
		return nil, err
	}
	seen := make([]bool, len(m.objects))
	result := make([]GiaID, 0)
	for _, counts := range [][]uint64{signals.SetCandidates, signals.ResetCandidates, signals.EnableCandidates} {
		for id := 1; id < len(counts); id++ {
			if counts[id] > uint64(threshold) && !seen[id] {
				result = append(result, GiaID(id))
				seen[id] = true
			}
		}
	}
	return result, nil
}
