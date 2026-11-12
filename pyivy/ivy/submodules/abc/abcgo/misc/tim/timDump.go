package tim

import (
	"encoding/binary"
	"math"
)

const TIM_DUMP_VER_NUM = 1

// Tim_ManSave preserves ABC's version-1 stream: integers are big endian,
// while float32 payloads use the host byte order (as in Vec_StrPutF).
func Tim_ManSave(p *Tim_Man_t, hierarchyOnly int) []byte {
	var data []byte
	putInt := func(n int) { data = binary.BigEndian.AppendUint32(data, uint32(int32(n))) }
	putFloat := func(n float32) { data = binary.NativeEndian.AppendUint32(data, math.Float32bits(n)) }
	for _, n := range []int{TIM_DUMP_VER_NUM, p.nCis, p.nCos, Tim_ManPiNum(p), Tim_ManPoNum(p), len(p.vBoxes)} {
		putInt(n)
	}
	for _, b := range p.vBoxes {
		for _, n := range []int{b.nInputs, b.nOutputs, b.iDelayTable, b.iCopy} {
			putInt(n)
		}
	}
	if hierarchyOnly != 0 {
		return data
	}
	putInt(len(p.vDelayTables))
	for i, t := range p.vDelayTables {
		if len(t) < 3 || int(t[0]) != i || len(t) != 3+int(t[1])*int(t[2]) {
			panic("tim: invalid table for serialization")
		}
		putInt(i)
		putInt(int(t[1]))
		putInt(int(t[2]))
		for _, n := range t[3:] {
			putFloat(n)
		}
	}
	for i, o := range p.pCis {
		if o.iObj2Box < 0 {
			putFloat(Tim_ManGetCiArrival(p, i))
		}
	}
	for i, o := range p.pCos {
		if o.iObj2Box < 0 {
			putFloat(Tim_ManGetCoRequired(p, i))
		}
	}
	return data
}

// Tim_ManLoad validates the encoded topology before allocating terminal arrays.
// Invalid counts must not force a large allocation before the inconsistency is
// discovered. Hierarchy-only input intentionally ends after the box records.
func Tim_ManLoad(data []byte, hierarchyOnly int) *Tim_Man_t {
	pos := 0
	getWord := func(order binary.ByteOrder) uint32 {
		if len(data)-pos < 4 {
			panic("tim: truncated timing stream")
		}
		value := order.Uint32(data[pos:])
		pos += 4
		return value
	}
	getInt := func() int { return int(int32(getWord(binary.BigEndian))) }
	getFloat := func() float32 { return math.Float32frombits(getWord(binary.NativeEndian)) }
	if getInt() != TIM_DUMP_VER_NUM {
		panic("tim: unsupported dump version")
	}
	cis, cos, pi, po := getInt(), getInt(), getInt(), getInt()
	if cis < 0 || cos < 0 || pi < 0 || po < 0 || pi > cis || po > cos || cis > 1<<28 || cos > 1<<28 {
		panic("tim: invalid terminal counts")
	}
	type boxRecord struct{ inputs, outputs, table, copyID int }
	count := getInt()
	if count < 0 || count > (len(data)-pos)/16 {
		panic("tim: invalid box count")
	}
	boxes := make([]boxRecord, count)
	ci, co := pi, 0
	for i := range boxes {
		box := boxRecord{getInt(), getInt(), getInt(), getInt()}
		if box.inputs < 0 || box.outputs < 0 || box.inputs > cos-co || box.outputs > cis-ci {
			panic("tim: invalid box terminal count")
		}
		boxes[i] = box
		ci += box.outputs
		co += box.inputs
	}
	if ci != cis || co+po != cos {
		panic("tim: inconsistent terminal counts")
	}
	var tables [][]float32
	var arrivals, required []float32
	if hierarchyOnly == 0 {
		tableCount := getInt()
		if tableCount < 0 || tableCount > (len(data)-pos)/12 {
			panic("tim: invalid table count")
		}
		if tableCount > 0 {
			tables = make([][]float32, tableCount)
		}
		for i := range tables {
			id, x, y := getInt(), getInt(), getInt()
			if id != i || x < 0 || y < 0 || int64(x)*int64(y) > int64((len(data)-pos)/4) {
				panic("tim: invalid delay table dimensions")
			}
			table := make([]float32, 3+x*y)
			table[0] = float32(id)
			table[1] = float32(x)
			table[2] = float32(y)
			for j := 3; j < len(table); j++ {
				table[j] = getFloat()
			}
			tables[i] = table
		}
		if pi > (len(data)-pos)/4 {
			panic("tim: truncated PI arrivals")
		}
		arrivals = make([]float32, pi)
		for i := range arrivals {
			arrivals[i] = getFloat()
		}
		if po > (len(data)-pos)/4 {
			panic("tim: truncated PO constraints")
		}
		required = make([]float32, po)
		for i := range required {
			required[i] = getFloat()
		}
		if pos != len(data) {
			panic("tim: trailing timing stream data")
		}
	}
	p := Tim_ManStart(cis, cos)
	ci, co = pi, 0
	for _, box := range boxes {
		Tim_ManCreateBox(p, co, box.inputs, ci, box.outputs, box.table, 0)
		p.vBoxes[len(p.vBoxes)-1].iCopy = box.copyID
		ci += box.outputs
		co += box.inputs
	}
	p.vDelayTables = tables
	for i, arrival := range arrivals {
		p.pCis[i].timeArr = arrival
	}
	for i, requirement := range required {
		p.pCos[cos-po+i].timeReq = requirement
	}
	return p
}
