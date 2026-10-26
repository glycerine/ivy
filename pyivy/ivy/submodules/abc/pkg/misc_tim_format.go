package abc

// Deterministic timing-manager diagnostics from misc/tim/timMan.c.

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

type TimBoxKindStats struct {
	DelayTable, Count, Inputs, Outputs int
}

type TimManagerStats struct {
	CIs, COs, PIs, POs                   int
	Boxes, BlackBoxes                    int
	BlackInputs, BlackOutputs, And2Delay int
	Kinds                                []TimBoxKindStats
}

// Stats returns owned, deterministically ordered hierarchy statistics.
func (m *TimManager) Stats(and2Delay int) (TimManagerStats, error) {
	if m == nil {
		return TimManagerStats{}, ErrDimension
	}
	pis, pos := m.PrimaryCounts()
	blackBoxes, blackInputs, blackOutputs := m.BlackBoxCounts()
	result := TimManagerStats{
		CIs: len(m.cis), COs: len(m.cos), PIs: pis, POs: pos,
		Boxes: len(m.boxes), BlackBoxes: blackBoxes, BlackInputs: blackInputs,
		BlackOutputs: blackOutputs, And2Delay: and2Delay,
	}
	for _, box := range m.boxes {
		index := -1
		for i := range result.Kinds {
			if result.Kinds[i].DelayTable == box.DelayTable {
				index = i
				break
			}
		}
		if index < 0 {
			result.Kinds = append(result.Kinds, TimBoxKindStats{DelayTable: box.DelayTable})
			index = len(result.Kinds) - 1
		}
		result.Kinds[index].Count++
		result.Kinds[index].Inputs = box.Inputs
		result.Kinds[index].Outputs = box.Outputs
	}
	slices.SortFunc(result.Kinds, func(a, b TimBoxKindStats) int {
		if a.DelayTable < b.DelayTable {
			return -1
		}
		if a.DelayTable > b.DelayTable {
			return 1
		}
		return 0
	})
	return result, nil
}

// Format returns a deterministic owned diagnostic. Float32 values are emitted
// as exact hexadecimal bits so NaN payloads and signed zero survive logging.
func (m *TimManager) Format() (string, error) {
	if m == nil {
		return "", ErrDimension
	}
	stats, err := m.Stats(0)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	fmt.Fprintf(&out, "TIMING MANAGER PI=%d CI=%d PO=%d CO=%d Box=%d\n", stats.PIs, stats.CIs, stats.POs, stats.COs, stats.Boxes)
	for _, terminal := range m.cis {
		fmt.Fprintf(&out, "CI %d box=%d pos=%d trav=%d arrival=%08x required=%08x\n", terminal.ID, terminal.Box, terminal.Position, terminal.TravID, math.Float32bits(terminal.Arrival), math.Float32bits(terminal.Required))
	}
	for _, terminal := range m.cos {
		fmt.Fprintf(&out, "CO %d box=%d pos=%d trav=%d arrival=%08x required=%08x\n", terminal.ID, terminal.Box, terminal.Position, terminal.TravID, math.Float32bits(terminal.Arrival), math.Float32bits(terminal.Required))
	}
	for _, box := range m.boxes {
		fmt.Fprintf(&out, "BOX %d in=%d+%d out=%d+%d table=%d copy=%d black=%t trav=%d\n", box.ID, box.FirstInput, box.Inputs, box.FirstOutput, box.Outputs, box.DelayTable, box.Copy, box.Black, box.TravID)
	}
	for tableID, table := range m.tables {
		if table.Inputs == 0 {
			fmt.Fprintf(&out, "TABLE %d empty\n", tableID)
			continue
		}
		fmt.Fprintf(&out, "TABLE %d %dx%d", tableID, table.Inputs, table.Outputs)
		for _, delay := range table.Delays {
			fmt.Fprintf(&out, " %08x", math.Float32bits(delay))
		}
		out.WriteByte('\n')
	}
	return out.String(), nil
}

// FormatBoxCopies is the deterministic replacement for Tim_ManPrintBoxCopy.
func (m *TimManager) FormatBoxCopies() (string, error) {
	if m == nil {
		return "", ErrDimension
	}
	var out strings.Builder
	for i, box := range m.boxes {
		if i != 0 {
			out.WriteByte(' ')
		}
		fmt.Fprint(&out, box.Copy)
	}
	out.WriteByte('\n')
	return out.String(), nil
}
