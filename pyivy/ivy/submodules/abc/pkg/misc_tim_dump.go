package abc

// Binary timing-manager persistence from misc/tim/timDump.c.

import (
	"encoding/binary"
	"math"
)

const timDumpVersion = 1

// TimDumpOptions bounds owned output and untrusted input. Zero selects a
// conservative default. The wire format is the pinned native format: signed
// 32-bit integers are big-endian and float32 values are little-endian.
type TimDumpOptions struct {
	MaxBytes, MaxTerminals, MaxBoxes, MaxPins, MaxTables, MaxValues int
}

func (o TimDumpOptions) normalized() (TimDumpOptions, error) {
	if o.MaxBytes < 0 || o.MaxTerminals < 0 || o.MaxBoxes < 0 || o.MaxPins < 0 || o.MaxTables < 0 || o.MaxValues < 0 {
		return o, ErrDimension
	}
	if o.MaxBytes == 0 {
		o.MaxBytes = 64 << 20
	}
	if o.MaxTerminals == 0 {
		o.MaxTerminals = 1 << 20
	}
	if o.MaxBoxes == 0 {
		o.MaxBoxes = 1 << 20
	}
	if o.MaxPins == 0 {
		o.MaxPins = 1 << 20
	}
	if o.MaxTables == 0 {
		o.MaxTables = 1 << 20
	}
	if o.MaxValues == 0 {
		o.MaxValues = 1 << 20
	}
	return o, nil
}

func timDumpAdd(size, count, width, limit int) (int, error) {
	return checkedSize(size, count, width, limit)
}

// timCanonicalLayout verifies the topology that timDump.c can represent. The
// native stream omits terminal indices, so box inputs must be consecutive from
// CO zero and box outputs consecutive after the primary-input prefix.
func (m *TimManager) timCanonicalLayout() (nPis, nPos int, err error) {
	if m == nil {
		return 0, 0, ErrDimension
	}
	if len(m.boxes) == 0 {
		for _, p := range m.cis {
			if p.Box != -1 || p.Position != -1 {
				return 0, 0, ErrInvariant
			}
		}
		for _, p := range m.cos {
			if p.Box != -1 || p.Position != -1 {
				return 0, 0, ErrInvariant
			}
		}
		return len(m.cis), len(m.cos), nil
	}
	nPis = m.boxes[0].FirstOutput
	if nPis < 0 || nPis > len(m.cis) {
		return 0, 0, ErrInvariant
	}
	ci, co := nPis, 0
	for boxID, box := range m.boxes {
		if box.ID != boxID || box.Inputs < 1 || box.Outputs < 1 || box.FirstInput != co || box.FirstOutput != ci {
			return 0, 0, ErrInvariant
		}
		if box.FirstInput > len(m.cos) || box.Inputs > len(m.cos)-box.FirstInput || box.FirstOutput > len(m.cis) || box.Outputs > len(m.cis)-box.FirstOutput {
			return 0, 0, ErrInvariant
		}
		for i := 0; i < box.Inputs; i++ {
			p := m.cos[box.FirstInput+i]
			if p.Box != boxID || p.Position != i {
				return 0, 0, ErrInvariant
			}
		}
		for i := 0; i < box.Outputs; i++ {
			p := m.cis[box.FirstOutput+i]
			if p.Box != boxID || p.Position != i {
				return 0, 0, ErrInvariant
			}
		}
		co += box.Inputs
		ci += box.Outputs
	}
	if ci != len(m.cis) || co > len(m.cos) {
		return 0, 0, ErrInvariant
	}
	nPos = len(m.cos) - co
	for i := 0; i < nPis; i++ {
		if m.cis[i].Box != -1 || m.cis[i].Position != -1 {
			return 0, 0, ErrInvariant
		}
	}
	for i := co; i < len(m.cos); i++ {
		if m.cos[i].Box != -1 || m.cos[i].Position != -1 {
			return 0, 0, ErrInvariant
		}
	}
	return nPis, nPos, nil
}

// SaveBinary returns an owned native-compatible timing stream. Hierarchy-only
// streams stop after box metadata. Black-box flags are not representable in
// the version-1 native format and therefore reject instead of being lost.
func (m *TimManager) SaveBinary(hierarchyOnly bool, options TimDumpOptions) ([]byte, error) {
	options, err := options.normalized()
	if err != nil {
		return nil, err
	}
	nPis, nPos, err := m.timCanonicalLayout()
	if err != nil {
		return nil, err
	}
	if len(m.cis) > options.MaxTerminals || len(m.cos) > options.MaxTerminals || len(m.boxes) > options.MaxBoxes || m.boxPins > options.MaxPins {
		return nil, ErrLimit
	}
	size, err := timDumpAdd(0, 6, 4, options.MaxBytes)
	if err != nil {
		return nil, err
	}
	size, err = timDumpAdd(size, len(m.boxes), 16, options.MaxBytes)
	if err != nil {
		return nil, err
	}
	if !hierarchyOnly {
		if len(m.tables) > options.MaxTables || m.tableValues > options.MaxValues {
			return nil, ErrLimit
		}
		size, err = timDumpAdd(size, 1, 4, options.MaxBytes)
		if err != nil {
			return nil, err
		}
		for _, table := range m.tables {
			if table.Inputs < 1 || table.Outputs < 1 || len(table.Delays) != table.Inputs*table.Outputs {
				return nil, ErrInvariant
			}
			size, err = timDumpAdd(size, 3, 4, options.MaxBytes)
			if err == nil {
				size, err = timDumpAdd(size, len(table.Delays), 4, options.MaxBytes)
			}
			if err != nil {
				return nil, err
			}
		}
		size, err = timDumpAdd(size, nPis+nPos, 4, options.MaxBytes)
		if err != nil {
			return nil, err
		}
	}
	data := make([]byte, 0, size)
	putInt := func(value int) { data = binary.BigEndian.AppendUint32(data, uint32(int32(value))) }
	putFloat := func(value float32) { data = binary.LittleEndian.AppendUint32(data, math.Float32bits(value)) }
	putInt(timDumpVersion)
	putInt(len(m.cis))
	putInt(len(m.cos))
	putInt(nPis)
	putInt(nPos)
	putInt(len(m.boxes))
	for _, box := range m.boxes {
		if box.Black {
			return nil, ErrUnsupported
		}
		putInt(box.Inputs)
		putInt(box.Outputs)
		putInt(box.DelayTable)
		putInt(box.Copy)
	}
	if hierarchyOnly {
		return data, nil
	}
	putInt(len(m.tables))
	for tableID, table := range m.tables {
		putInt(tableID)
		putInt(table.Inputs)
		putInt(table.Outputs)
		for _, value := range table.Delays {
			putFloat(value)
		}
	}
	for i := 0; i < nPis; i++ {
		putFloat(m.cis[i].Arrival)
	}
	for i := len(m.cos) - nPos; i < len(m.cos); i++ {
		putFloat(m.cos[i].Required)
	}
	return data, nil
}

type timDumpReader struct {
	data []byte
	pos  int
}

func (r *timDumpReader) word() (uint32, error) {
	if r.pos > len(r.data)-4 {
		return 0, ErrDimension
	}
	value := r.data[r.pos : r.pos+4]
	r.pos += 4
	return binary.BigEndian.Uint32(value), nil
}

func (r *timDumpReader) integer() (int, error) {
	value, err := r.word()
	return int(int32(value)), err
}

func (r *timDumpReader) floating() (float32, error) {
	if r.pos > len(r.data)-4 {
		return 0, ErrDimension
	}
	value := math.Float32frombits(binary.LittleEndian.Uint32(r.data[r.pos : r.pos+4]))
	r.pos += 4
	return value, nil
}

// LoadTimManagerBinary parses one complete version-1 native stream. Unlike the
// assertion-based C loader, malformed, trailing and over-limit data is rejected
// without exposing a partial manager.
func LoadTimManagerBinary(data []byte, hierarchyOnly bool, options TimDumpOptions) (*TimManager, error) {
	options, err := options.normalized()
	if err != nil {
		return nil, err
	}
	if len(data) > options.MaxBytes {
		return nil, ErrLimit
	}
	r := timDumpReader{data: data}
	version, err := r.integer()
	if err != nil || version != timDumpVersion {
		return nil, ErrDimension
	}
	nCis, err := r.integer()
	if err != nil {
		return nil, err
	}
	nCos, err := r.integer()
	if err != nil {
		return nil, err
	}
	nPis, err := r.integer()
	if err != nil {
		return nil, err
	}
	nPos, err := r.integer()
	if err != nil {
		return nil, err
	}
	nBoxes, err := r.integer()
	if err != nil {
		return nil, err
	}
	if nCis < 0 || nCos < 0 || nPis < 0 || nPis > nCis || nPos < 0 || nPos > nCos || nBoxes < 0 {
		return nil, ErrDimension
	}
	if nCis > options.MaxTerminals || nCos > options.MaxTerminals || nBoxes > options.MaxBoxes {
		return nil, ErrLimit
	}
	m, err := NewTimManager(nCis, nCos, options.MaxTerminals)
	if err != nil {
		return nil, err
	}
	curCI, curCO := nPis, 0
	for boxID := 0; boxID < nBoxes; boxID++ {
		inputs, e1 := r.integer()
		outputs, e2 := r.integer()
		tableID, e3 := r.integer()
		copyID, e4 := r.integer()
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
			return nil, ErrDimension
		}
		if inputs < 1 || outputs < 1 || curCO > nCos-inputs || curCI > nCis-outputs {
			return nil, ErrDimension
		}
		_, err = m.AddBox(TimBox{FirstInput: curCO, Inputs: inputs, FirstOutput: curCI, Outputs: outputs, DelayTable: tableID}, TimBoxOptions{MaxBoxes: options.MaxBoxes, MaxPins: options.MaxPins})
		if err != nil {
			return nil, err
		}
		if err = m.SetBoxCopy(boxID, copyID); err != nil {
			return nil, err
		}
		curCO += inputs
		curCI += outputs
	}
	if curCI != nCis || curCO+nPos != nCos {
		return nil, ErrDimension
	}
	if hierarchyOnly {
		if r.pos != len(data) {
			return nil, ErrDimension
		}
		return m, nil
	}
	nTables, err := r.integer()
	if err != nil || nTables < 0 {
		return nil, ErrDimension
	}
	if nTables > options.MaxTables {
		return nil, ErrLimit
	}
	for tableID := 0; tableID < nTables; tableID++ {
		storedID, e1 := r.integer()
		inputs, e2 := r.integer()
		outputs, e3 := r.integer()
		if e1 != nil || e2 != nil || e3 != nil || storedID != tableID || inputs < 1 || outputs < 1 {
			return nil, ErrDimension
		}
		count, err := checkedSize(0, inputs, outputs, options.MaxValues)
		if err != nil {
			return nil, err
		}
		values := make([]float32, count)
		for i := range values {
			if values[i], err = r.floating(); err != nil {
				return nil, err
			}
		}
		if err = m.SetDelayTable(tableID, TimDelayTable{Inputs: inputs, Outputs: outputs, Delays: values}, TimDelayOptions{MaxTables: options.MaxTables, MaxValues: options.MaxValues}); err != nil {
			return nil, err
		}
	}
	for i := 0; i < nPis; i++ {
		value, err := r.floating()
		if err != nil {
			return nil, err
		}
		m.cis[i].Arrival = value
	}
	for i := nCos - nPos; i < nCos; i++ {
		value, err := r.floating()
		if err != nil {
			return nil, err
		}
		m.cos[i].Required = value
	}
	if r.pos != len(data) {
		return nil, ErrDimension
	}
	return m, nil
}
