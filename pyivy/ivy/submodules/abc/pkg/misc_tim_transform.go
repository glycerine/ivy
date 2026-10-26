package abc

// Manager duplication and timing-array extraction from misc/tim/timMan.c.

import "slices"

// Duplicate returns an independent timing manager with a fresh traversal
// epoch. A nonzero unitDelay replaces every enabled box arc and resets primary
// arrivals/requirements on the duplicate. This corrects Tim_ManDup, which
// accidentally resets the source manager instead.
func (m *TimManager) Duplicate(unitDelay int) (*TimManager, error) {
	if m == nil || int64(unitDelay) < -2147483648 || int64(unitDelay) > 2147483647 {
		return nil, ErrDimension
	}
	copy := m.Clone()
	copy.epoch = 0
	copy.useTraversal = true
	for i := range copy.cis {
		copy.cis[i].TravID = 0
	}
	for i := range copy.cos {
		copy.cos[i].TravID = 0
	}
	for i := range copy.boxes {
		copy.boxes[i].TravID = 0
	}
	if unitDelay == 0 {
		return copy, nil
	}
	for i := range copy.cis {
		if copy.cis[i].Box < 0 {
			copy.cis[i].Arrival = 0
		}
	}
	for i := range copy.cos {
		if copy.cos[i].Box < 0 {
			copy.cos[i].Required = TimEternity
		}
	}
	value := float32(unitDelay)
	for tableID := range copy.tables {
		for i, delay := range copy.tables[tableID].Delays {
			if delay != TimNoArc {
				copy.tables[tableID].Delays[i] = value
			}
		}
	}
	return copy, nil
}

// PrimaryArrivalTimes returns nil when every primary arrival is zero. The
// owned result is in primary-input order rather than the oversized native CI
// allocation whose boxed tail is unused.
func (m *TimManager) PrimaryArrivalTimes() ([]float32, error) {
	if m == nil {
		return nil, ErrDimension
	}
	values := make([]float32, 0, len(m.cis))
	nontrivial := false
	for _, terminal := range m.cis {
		if terminal.Box >= 0 {
			continue
		}
		values = append(values, terminal.Arrival)
		nontrivial = nontrivial || terminal.Arrival != 0
	}
	if !nontrivial {
		return nil, nil
	}
	return slices.Clip(values), nil
}

// PrimaryRequiredTimes returns nil when every primary required time is
// TimEternity. It corrects Tim_ManGetReqTimes, which tests required values but
// accidentally copies arrival values into the returned array.
func (m *TimManager) PrimaryRequiredTimes() ([]float32, error) {
	if m == nil {
		return nil, ErrDimension
	}
	values := make([]float32, 0, len(m.cos))
	nontrivial := false
	for _, terminal := range m.cos {
		if terminal.Box >= 0 {
			continue
		}
		values = append(values, terminal.Required)
		nontrivial = nontrivial || terminal.Required != TimEternity
	}
	if !nontrivial {
		return nil, nil
	}
	return slices.Clip(values), nil
}

func (m *TimManager) rebuildBoxes(selected []int, nPis, nPos int) (*TimManager, error) {
	if m == nil || nPis < 0 || nPos < 0 {
		return nil, ErrDimension
	}
	oldPis, oldPos, err := m.timCanonicalLayout()
	if err != nil || nPis > oldPis || nPos > oldPos {
		return nil, ErrDimension
	}
	totalCIs, totalCOs, totalPins := nPis, nPos, 0
	for _, boxID := range selected {
		if boxID < 0 || boxID >= len(m.boxes) {
			return nil, ErrDimension
		}
		box := m.boxes[boxID]
		if totalCIs > int(^uint(0)>>1)-box.Outputs || totalCOs > int(^uint(0)>>1)-box.Inputs || totalPins > int(^uint(0)>>1)-box.Inputs-box.Outputs {
			return nil, ErrLimit
		}
		totalCIs += box.Outputs
		totalCOs += box.Inputs
		totalPins += box.Inputs + box.Outputs
	}
	if totalCIs > len(m.cis)-(oldPis-nPis) || totalCOs > len(m.cos)-(oldPos-nPos) {
		return nil, ErrDimension
	}
	result, err := NewTimManager(totalCIs, totalCOs, totalCIs+totalCOs)
	if err != nil {
		return nil, err
	}
	for i := 0; i < nPis; i++ {
		result.cis[i] = m.cis[i]
		result.cis[i].ID = i
		result.cis[i].TravID = 0
		result.cis[i].Box = -1
		result.cis[i].Position = -1
	}
	sourcePO := len(m.cos) - oldPos
	targetPO := totalCOs - nPos
	for i := 0; i < nPos; i++ {
		result.cos[targetPO+i] = m.cos[sourcePO+i]
		result.cos[targetPO+i].ID = targetPO + i
		result.cos[targetPO+i].TravID = 0
		result.cos[targetPO+i].Box = -1
		result.cos[targetPO+i].Position = -1
	}
	result.tables = slices.Clone(m.tables)
	for i := range result.tables {
		result.tables[i].Delays = slices.Clone(m.tables[i].Delays)
	}
	result.tableValues = m.tableValues
	firstCI, firstCO := nPis, 0
	for _, boxID := range selected {
		box := m.boxes[boxID]
		copyID := box.Copy
		if copyID == -1 {
			copyID = boxID
		}
		newID, err := result.AddBox(TimBox{FirstInput: firstCO, Inputs: box.Inputs, FirstOutput: firstCI, Outputs: box.Outputs, DelayTable: box.DelayTable, Black: box.Black}, TimBoxOptions{MaxBoxes: max(1, len(selected)), MaxPins: max(1, totalPins)})
		if err != nil {
			return nil, err
		}
		if err = result.SetBoxCopy(newID, copyID); err != nil {
			return nil, err
		}
		firstCI += box.Outputs
		firstCO += box.Inputs
	}
	return result, nil
}

// TrimBoxes retains boxes selected by a same-length mask and repacks their
// terminals between the original primary-input and primary-output regions.
// Primary constraints and all delay tables are owned by the result.
func (m *TimManager) TrimBoxes(preserve []bool) (*TimManager, error) {
	if m == nil || len(preserve) != len(m.boxes) {
		return nil, ErrDimension
	}
	selected := make([]int, 0, len(m.boxes))
	for boxID, keep := range preserve {
		if keep {
			selected = append(selected, boxID)
		}
	}
	if len(selected) == len(m.boxes) {
		return m.Duplicate(0)
	}
	nPis, nPos, err := m.timCanonicalLayout()
	if err != nil {
		return nil, err
	}
	return m.rebuildBoxes(selected, nPis, nPos)
}

// ReduceBoxes retains boxes in the supplied order while removing the same
// number of primary inputs and outputs. Repeated box IDs retain native behavior
// when the resulting dimensions remain valid.
func (m *TimManager) ReduceBoxes(boxes []int, terminalDifference int) (*TimManager, error) {
	if m == nil || terminalDifference < 0 || len(boxes) > len(m.boxes) {
		return nil, ErrDimension
	}
	nPis, nPos, err := m.timCanonicalLayout()
	if err != nil || terminalDifference > nPis || terminalDifference > nPos {
		return nil, ErrDimension
	}
	return m.rebuildBoxes(slices.Clone(boxes), nPis-terminalDifference, nPos-terminalDifference)
}

// AlignTimBoxes maps implementation boxes onto specification box IDs using
// their Copy fields. The result has one true entry per represented spec box.
func AlignTimBoxes(specification, implementation *TimManager) ([]bool, error) {
	if specification == nil || implementation == nil || len(specification.boxes) <= len(implementation.boxes) {
		return nil, ErrDimension
	}
	result := make([]bool, len(specification.boxes))
	for _, box := range implementation.boxes {
		if box.Copy < 0 || box.Copy >= len(result) || result[box.Copy] {
			return nil, ErrInvariant
		}
		result[box.Copy] = true
	}
	return result, nil
}
