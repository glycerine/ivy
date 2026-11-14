package aig

import "fmt"

type Rtm_Init_t uint8

const (
	RTM_VAL_NONE Rtm_Init_t = iota
	RTM_VAL_ZERO
	RTM_VAL_ONE
	RTM_VAL_VOID
)

type Rtm_Edg_t struct {
	nLats int32
	LData uint32
	big   []uint32
}
type Rtm_Obj_t struct {
	// Num is the fanin capacity during construction, then retiming depth.
	// Temp holds the fanout capacity, matching the source construction checks.
	pCopy              Edge
	Kind               NodeKind
	Type, fMark, fAuto bool
	Num, Temp, Id      int32
	fanins, fanouts    []*Rtm_Obj_t
	edges, outEdges    []*Rtm_Edg_t
	complements        []bool
}
type Rtm_Man_t struct {
	vObjs, vPis, vPos, latchObjs []*Rtm_Obj_t
	extra                        [][]uint32
	nExtraCur                    int32
	source                       *Graph
	sourceMap                    map[int]*Rtm_Obj_t
}

func Rtm_ObjFanin(object *Rtm_Obj_t, i int32) *Rtm_Obj_t      { return object.fanins[i] }
func Rtm_ObjFanout(object *Rtm_Obj_t, i int32) *Rtm_Obj_t     { return object.fanouts[i] }
func Rtm_ObjEdge(object *Rtm_Obj_t, i int32) *Rtm_Edg_t       { return object.edges[i] }
func Rtm_ObjFanoutEdge(object *Rtm_Obj_t, i int32) *Rtm_Edg_t { return object.outEdges[i] }
func Rtm_InitNot(value Rtm_Init_t) Rtm_Init_t {
	switch value {
	case RTM_VAL_ZERO:
		return RTM_VAL_ONE
	case RTM_VAL_ONE:
		return RTM_VAL_ZERO
	case RTM_VAL_VOID:
		return RTM_VAL_VOID
	}
	panic("aig: missing latch value")
}
func Rtm_InitNotCond(value Rtm_Init_t, condition int32) Rtm_Init_t {
	if condition != 0 {
		return Rtm_InitNot(value)
	}
	return value
}
func Rtm_InitAnd(a, b Rtm_Init_t) Rtm_Init_t {
	if a == RTM_VAL_ZERO || b == RTM_VAL_ZERO {
		return RTM_VAL_ZERO
	}
	if a == RTM_VAL_ONE && b == RTM_VAL_ONE {
		return RTM_VAL_ONE
	}
	if a == RTM_VAL_VOID || b == RTM_VAL_VOID {
		return RTM_VAL_VOID
	}
	panic("aig: missing latch value")
}
func Rtm_InitWordsNum(latches int32) int32 {
	if latches < 0 {
		panic("aig: negative latch count")
	}
	return (latches + 15) / 16
}
func Rtm_InitGetTwo(words []uint32, i int32) int32 {
	if i < 0 || int(i/16) >= len(words) {
		panic("aig: latch storage index outside range")
	}
	return int32(words[i/16] >> uint((i&15)*2) & 3)
}
func Rtm_InitSetTwo(words []uint32, i, value int32) {
	if i < 0 || int(i/16) >= len(words) || value < 0 || value > 3 {
		panic("aig: invalid latch storage write")
	}
	shift := uint((i & 15) * 2)
	words[i/16] = words[i/16]&^(3<<shift) | uint32(value)<<shift
}
func Rtm_InitXorTwo(words []uint32, i, value int32) {
	if i < 0 || int(i/16) >= len(words) || value < 0 || value > 3 {
		panic("aig: invalid latch storage XOR")
	}
	words[i/16] ^= uint32(value) << uint((i&15)*2)
}
func rtmLatchIndex(edge *Rtm_Edg_t, i int32) {
	if i < 0 || i >= edge.nLats {
		panic("aig: latch index outside edge")
	}
}
func rtmLatchValue(value Rtm_Init_t) {
	if value < RTM_VAL_ZERO || value > RTM_VAL_VOID {
		panic("aig: invalid latch initialization")
	}
}
func Rtm_ObjGetOne1(edge *Rtm_Edg_t, i int32) Rtm_Init_t {
	rtmLatchIndex(edge, i)
	if i >= 16 {
		panic("aig: compact latch overflow")
	}
	return Rtm_Init_t(edge.LData >> uint(i*2) & 3)
}
func Rtm_ObjGetFirst1(edge *Rtm_Edg_t) Rtm_Init_t { return Rtm_ObjGetOne1(edge, 0) }
func Rtm_ObjGetLast1(edge *Rtm_Edg_t) Rtm_Init_t  { return Rtm_ObjGetOne1(edge, edge.nLats-1) }
func Rtm_ObjRemFirst1(edge *Rtm_Edg_t) Rtm_Init_t {
	value := Rtm_ObjGetFirst1(edge)
	edge.LData >>= 2
	edge.nLats--
	return value
}
func Rtm_ObjRemLast1(edge *Rtm_Edg_t) Rtm_Init_t {
	value := Rtm_ObjGetLast1(edge)
	edge.nLats--
	edge.LData &^= 3 << uint(edge.nLats*2)
	return value
}
func Rtm_ObjAddFirst1(edge *Rtm_Edg_t, value Rtm_Init_t) {
	rtmLatchValue(value)
	if edge.nLats >= 10 {
		panic("aig: compact latch storage full")
	}
	edge.LData = edge.LData<<2 | uint32(value)
	edge.nLats++
}
func Rtm_ObjAddLast1(edge *Rtm_Edg_t, value Rtm_Init_t) {
	rtmLatchValue(value)
	if edge.nLats >= 10 {
		panic("aig: compact latch storage full")
	}
	edge.LData |= uint32(value) << uint(edge.nLats*2)
	edge.nLats++
}
func rtmAllocateWords(p *Rtm_Man_t, count int32) []uint32 {
	words := make([]uint32, count)
	if p != nil {
		p.extra = append(p.extra, words)
		p.nExtraCur += count
	}
	return words
}
func Rtm_ObjTransferToBig(p *Rtm_Man_t, edge *Rtm_Edg_t) {
	if edge.big != nil || edge.nLats > 10 {
		panic("aig: invalid compact-to-overflow conversion")
	}
	count := Rtm_InitWordsNum(edge.nLats + 1)
	if count < 1 {
		count = 1
	}
	edge.big = rtmAllocateWords(p, count)
	edge.big[0] = edge.LData
	edge.LData = 0
}
func Rtm_ObjTransferToSmall(p *Rtm_Man_t, edge *Rtm_Edg_t) {
	if edge.big == nil || edge.nLats > 10 {
		panic("aig: invalid overflow-to-compact conversion")
	}
	edge.LData = edge.big[0]
	if edge.nLats == 0 {
		edge.LData = 0
	} else {
		edge.LData &= (1 << uint(edge.nLats*2)) - 1
	}
	edge.big = nil
}
func Rtm_ObjTransferToBigger(p *Rtm_Man_t, edge *Rtm_Edg_t) {
	if edge.big == nil {
		panic("aig: overflow conversion requires overflow storage")
	}
	count := int32(len(edge.big)) + 1
	words := rtmAllocateWords(p, count)
	copy(words, edge.big)
	edge.big = words
}
func Rtm_ObjGetOne2(p *Rtm_Man_t, edge *Rtm_Edg_t, i int32) Rtm_Init_t {
	rtmLatchIndex(edge, i)
	return Rtm_Init_t(Rtm_InitGetTwo(edge.big, i))
}
func Rtm_ObjGetFirst2(p *Rtm_Man_t, edge *Rtm_Edg_t) Rtm_Init_t { return Rtm_ObjGetOne2(p, edge, 0) }
func Rtm_ObjGetLast2(p *Rtm_Man_t, edge *Rtm_Edg_t) Rtm_Init_t {
	return Rtm_ObjGetOne2(p, edge, edge.nLats-1)
}
func Rtm_ObjRemFirst2(p *Rtm_Man_t, edge *Rtm_Edg_t) Rtm_Init_t {
	value := Rtm_ObjGetFirst2(p, edge)
	for i := range edge.big {
		edge.big[i] >>= 2
		if i+1 < len(edge.big) {
			edge.big[i] |= (edge.big[i+1] & 3) << 30
		}
	}
	edge.nLats--
	return value
}
func Rtm_ObjRemLast2(p *Rtm_Man_t, edge *Rtm_Edg_t) Rtm_Init_t {
	value := Rtm_ObjGetLast2(p, edge)
	edge.nLats--
	Rtm_InitSetTwo(edge.big, edge.nLats, 0)
	return value
}
func Rtm_ObjAddFirst2(p *Rtm_Man_t, edge *Rtm_Edg_t, value Rtm_Init_t) {
	rtmLatchValue(value)
	if int(Rtm_InitWordsNum(edge.nLats+1)) > len(edge.big) {
		Rtm_ObjTransferToBigger(p, edge)
	}
	carry := uint32(value)
	for i := range edge.big {
		next := edge.big[i] >> 30
		edge.big[i] = edge.big[i]<<2 | carry
		carry = next
	}
	edge.nLats++
}
func Rtm_ObjAddLast2(p *Rtm_Man_t, edge *Rtm_Edg_t, value Rtm_Init_t) {
	rtmLatchValue(value)
	if int(Rtm_InitWordsNum(edge.nLats+1)) > len(edge.big) {
		Rtm_ObjTransferToBigger(p, edge)
	}
	Rtm_InitSetTwo(edge.big, edge.nLats, int32(value))
	edge.nLats++
}
func Rtm_ObjGetOne(p *Rtm_Man_t, edge *Rtm_Edg_t, i int32) Rtm_Init_t {
	if edge.big != nil {
		return Rtm_ObjGetOne2(p, edge, i)
	}
	return Rtm_ObjGetOne1(edge, i)
}
func Rtm_ObjGetFirst(p *Rtm_Man_t, edge *Rtm_Edg_t) Rtm_Init_t { return Rtm_ObjGetOne(p, edge, 0) }
func Rtm_ObjGetLast(p *Rtm_Man_t, edge *Rtm_Edg_t) Rtm_Init_t {
	return Rtm_ObjGetOne(p, edge, edge.nLats-1)
}
func Rtm_ObjRemFirst(p *Rtm_Man_t, edge *Rtm_Edg_t) Rtm_Init_t {
	var value Rtm_Init_t
	if edge.big != nil {
		value = Rtm_ObjRemFirst2(p, edge)
		if edge.nLats <= 10 {
			Rtm_ObjTransferToSmall(p, edge)
		}
	} else {
		value = Rtm_ObjRemFirst1(edge)
	}
	return value
}
func Rtm_ObjRemLast(p *Rtm_Man_t, edge *Rtm_Edg_t) Rtm_Init_t {
	var value Rtm_Init_t
	if edge.big != nil {
		value = Rtm_ObjRemLast2(p, edge)
		if edge.nLats <= 10 {
			Rtm_ObjTransferToSmall(p, edge)
		}
	} else {
		value = Rtm_ObjRemLast1(edge)
	}
	return value
}
func Rtm_ObjAddFirst(p *Rtm_Man_t, edge *Rtm_Edg_t, value Rtm_Init_t) {
	if edge.big == nil && edge.nLats == 10 {
		Rtm_ObjTransferToBig(p, edge)
	}
	if edge.big != nil {
		Rtm_ObjAddFirst2(p, edge, value)
	} else {
		Rtm_ObjAddFirst1(edge, value)
	}
}
func Rtm_ObjAddLast(p *Rtm_Man_t, edge *Rtm_Edg_t, value Rtm_Init_t) {
	if edge.big == nil && edge.nLats == 10 {
		Rtm_ObjTransferToBig(p, edge)
	}
	if edge.big != nil {
		Rtm_ObjAddLast2(p, edge, value)
	} else {
		Rtm_ObjAddLast1(edge, value)
	}
}
func Rtm_PrintEdge(p *Rtm_Man_t, edge *Rtm_Edg_t) {
	fmt.Printf("%d :", edge.nLats)
	for i := int32(0); i < edge.nLats; i++ {
		fmt.Printf(" %d", Rtm_ObjGetOne(p, edge, i))
	}
	fmt.Println()
}
