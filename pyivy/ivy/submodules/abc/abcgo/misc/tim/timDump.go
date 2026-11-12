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
func Tim_ManLoad(data []byte, hierarchyOnly int) *Tim_Man_t {
	pos := 0
	getWord := func(order binary.ByteOrder) uint32 {
		if len(data)-pos < 4 {
			panic("tim: truncated timing stream")
		}
		v := order.Uint32(data[pos:])
		pos += 4
		return v
	}
	getInt := func() int { return int(int32(getWord(binary.BigEndian))) }
	getFloat := func() float32 { return math.Float32frombits(getWord(binary.NativeEndian)) }
	if getInt() != TIM_DUMP_VER_NUM {
		panic("tim: unsupported dump version")
	}
	cis, cos, pi, po := getInt(), getInt(), getInt(), getInt()
	// A valid timing file contains at least one value per terminal in full mode;
	// cap all counts by encoded bytes even in hierarchy-only mode.
	if cis < 0 || cos < 0 || pi < 0 || po < 0 || pi > cis || po > cos || cis > 1<<28 || cos > 1<<28 {
		panic("tim: invalid terminal counts")
	}
	p := Tim_ManStart(cis, cos)
	n := getInt()
	if n < 0 || n > (len(data)-pos)/16 {
		panic("tim: invalid box count")
	}
	ci, co := pi, 0
	for i := 0; i < n; i++ {
		ins, outs, table, copyID := getInt(), getInt(), getInt(), getInt()
		Tim_ManCreateBox(p, co, ins, ci, outs, table, 0)
		p.vBoxes[i].iCopy = copyID
		ci += outs
		co += ins
	}
	if ci != cis || co+po != cos {
		panic("tim: inconsistent terminal counts")
	}
	if hierarchyOnly != 0 {
		return p
	}
	tables := getInt()
	if tables < 0 || tables > (len(data)-pos)/12 {
		panic("tim: invalid table count")
	}
	p.vDelayTables = make([][]float32, tables)
	for i := 0; i < tables; i++ {
		id, x, y := getInt(), getInt(), getInt()
		if id != i || x < 0 || y < 0 || int64(x)*int64(y) > int64((len(data)-pos)/4) {
			panic("tim: invalid delay table dimensions")
		}
		table := make([]float32, 3+x*y)
		table[0] = float32(id)
		table[1] = float32(x)
		table[2] = float32(y)
		for k := 3; k < len(table); k++ {
			table[k] = getFloat()
		}
		p.vDelayTables[i] = table
	}
	for i := range p.pCis {
		if p.pCis[i].iObj2Box < 0 {
			p.pCis[i].timeArr = getFloat()
		}
	}
	for i := range p.pCos {
		if p.pCos[i].iObj2Box < 0 {
			p.pCos[i].timeReq = getFloat()
		}
	}
	if pos != len(data) {
		panic("tim: trailing timing stream data")
	}
	return p
}
