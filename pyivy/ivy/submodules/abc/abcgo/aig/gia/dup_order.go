package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManDupRemapLiterals(lits *Vec_Int_t, p *Gia_Man_t) {
	for i, lit := range vec.Vec_IntArray(lits) {
		if lit < 0 {
			continue
		}
		id := lit >> 1
		if id >= p.nObjs {
			panic("gia: invalid literal to remap")
		}
		mapped := p.pObjs[id].Value
		if mapped == ^uint32(0) {
			vec.Vec_IntWriteEntry(lits, i, -1)
		} else {
			vec.Vec_IntWriteEntry(lits, i, int32(mapped)^lit&1)
		}
	}
}
func Gia_ManDupRemapEquiv(dst, src *Gia_Man_t) {
	if src.PReprs == nil {
		return
	}
	if dst.PReprs != nil || dst.PNexts != nil {
		panic("gia: destination representatives already allocated")
	}
	mapping := make([]int32, src.nObjs)
	for id, obj := range src.pObjs {
		if obj.Value == ^uint32(0) {
			mapping[id] = -1
		} else {
			mapping[id] = int32(obj.Value >> 1)
		}
	}
	if err := dst.TransferRepresentativesFrom(src, mapping); err != nil {
		panic(err)
	}
}
func Gia_ManDupRemapCis(dst, src *Gia_Man_t) {
	if Gia_ManCiNum(dst) != Gia_ManCiNum(src) {
		panic("gia: input count mismatch")
	}
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		obj := Gia_ManCi(src, i)
		if obj.Value == ^uint32(0) || obj.Value&1 != 0 {
			panic("gia: invalid mapped input")
		}
		id := int32(obj.Value >> 1)
		mapped := Gia_ManObj(dst, id)
		vec.Vec_IntWriteEntry(dst.vCis, int(i), id)
		Gia_ObjSetCioId(mapped, i)
	}
}
func Gia_ManDupRemapCos(dst, src *Gia_Man_t) {
	if Gia_ManCoNum(dst) != Gia_ManCoNum(src) {
		panic("gia: output count mismatch")
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		obj := Gia_ManCo(src, i)
		if obj.Value == ^uint32(0) || obj.Value&1 != 0 {
			panic("gia: invalid mapped output")
		}
		id := int32(obj.Value >> 1)
		mapped := Gia_ManObj(dst, id)
		vec.Vec_IntWriteEntry(dst.vCos, int(i), id)
		Gia_ObjSetCioId(mapped, i)
	}
}

// Gia_ManDupOrderDfs_rec maps a source object to a destination literal.
func Gia_ManDupOrderDfs_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if obj.Value != ^uint32(0) {
		return int32(obj.Value)
	}
	a, b, c, kind := src.Gate(obj.id)
	if kind == 1 {
		obj.Value = uint32(Gia_ManAppendCi(dst))
		return int32(obj.Value)
	}
	if kind < 2 || kind > 6 {
		panic("gia: invalid DFS duplication gate")
	}
	literal := func(lit int32) int32 { return Gia_ManDupOrderDfs_rec(dst, src, src.pObjs[lit>>1]) ^ (lit & 1) }
	var result int32
	switch kind {
	case 2:
		result = Gia_ManAppendAnd(dst, literal(a), literal(b))
	case 3:
		result = Gia_ManAppendXorReal(dst, literal(a), literal(b))
	case 4:
		result = Gia_ManAppendMuxReal(dst, literal(a), literal(b), literal(c))
	case 5:
		result = Gia_ManAppendBuf(dst, literal(a))
	case 6:
		result = Gia_ManAppendCo(dst, literal(a))
	}
	obj.Value = uint32(result)
	return result
}
func Gia_ManDupOrderDfs(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManFillValue(src)
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		Gia_ManDupOrderDfs_rec(dst, src, Gia_ManCo(src, i))
	}
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		obj := Gia_ManCi(src, i)
		if obj.Value == ^uint32(0) {
			obj.Value = uint32(Gia_ManAppendCi(dst))
		}
	}
	Gia_ManDupRemapCis(dst, src)
	Gia_ManDupRemapEquiv(dst, src)
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	return dst
}
func Gia_ManDupAbs(src *Gia_Man_t, ppiToFlop, flopToPpi *Vec_Int_t) *Gia_Man_t {
	if vec.Vec_IntSize(flopToPpi) != int(Gia_ManRegNum(src)) {
		panic("gia: invalid abstraction mapping")
	}
	Gia_ManFillValue(src)
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManPiNum(src); i++ {
		Gia_ManPi(src, i).Value = uint32(Gia_ManAppendCi(dst))
	}
	for index, flop := range vec.Vec_IntArray(ppiToFlop) {
		if flop < 0 || flop >= Gia_ManRegNum(src) || vec.Vec_IntEntry(flopToPpi, int(flop)) != int32(index) {
			panic("gia: inconsistent abstraction mapping")
		}
		obj := Gia_ManRo(src, flop)
		if obj.Value != ^uint32(0) {
			panic("gia: duplicate abstracted register")
		}
		obj.Value = uint32(Gia_ManAppendCi(dst))
	}
	for flop, ppi := range vec.Vec_IntArray(flopToPpi) {
		obj := Gia_ManRo(src, int32(flop))
		if ppi >= 0 {
			if obj.Value == ^uint32(0) {
				panic("gia: missing abstracted register")
			}
			continue
		}
		obj.Value = uint32(Gia_ManAppendCi(dst))
	}
	roots := make([]*Gia_Obj_t, 0, Gia_ManCoNum(src))
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		roots = append(roots, Gia_ManPo(src, i))
	}
	for flop, ppi := range vec.Vec_IntArray(flopToPpi) {
		if ppi < 0 {
			roots = append(roots, Gia_ManRi(src, int32(flop)))
		}
	}
	for _, obj := range roots {
		Gia_ManDupOrderDfs_rec(dst, src, Gia_ObjFanin0(obj))
	}
	for _, obj := range roots {
		obj.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj)))
	}
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src)-int32(vec.Vec_IntSize(ppiToFlop)))
	return dst
}
func Gia_ManDupOutputGroup(src *Gia_Man_t, start, stop int32) *Gia_Man_t {
	if start < 0 || stop < start || stop > Gia_ManCoNum(src) {
		panic("gia: invalid output range")
	}
	Gia_ManFillValue(src)
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for i := start; i < stop; i++ {
		Gia_ManDupOrderDfs_rec(dst, src, Gia_ManCo(src, i))
	}
	return dst
}
func Gia_ManDupOutputVec(src *Gia_Man_t, present *Vec_Int_t) *Gia_Man_t {
	if Gia_ManRegNum(src) != 0 || vec.Vec_IntSize(present) != int(Gia_ManPoNum(src)) {
		panic("gia: invalid output selection")
	}
	selected := vec.Vec_IntAlloc(vec.Vec_IntSize(present))
	for i, value := range vec.Vec_IntArray(present) {
		if value != 0 {
			vec.Vec_IntPush(selected, int32(i))
		}
	}
	return Gia_ManDupSelectedOutputs(src, selected)
}
func Gia_ManDupSelectedOutputs(src *Gia_Man_t, selected *Vec_Int_t) *Gia_Man_t {
	if Gia_ManRegNum(src) != 0 {
		panic("gia: selected outputs require a combinational graph")
	}
	Gia_ManFillValue(src)
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManPiNum(src); i++ {
		Gia_ManPi(src, i).Value = uint32(Gia_ManAppendCi(dst))
	}
	for _, i := range vec.Vec_IntArray(selected) {
		if i < 0 || i >= Gia_ManPoNum(src) {
			panic("gia: selected output out of range")
		}
		Gia_ManDupOrderDfs_rec(dst, src, Gia_ObjFanin0(Gia_ManPo(src, i)))
	}
	for _, i := range vec.Vec_IntArray(selected) {
		obj := Gia_ManPo(src, i)
		obj.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj)))
	}
	return dst
}

func Gia_ManDupOrderDfsReverse(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManFillValue(src)
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for i := Gia_ManCoNum(src); i > 0; i-- {
		Gia_ManDupOrderDfs_rec(dst, src, Gia_ManCo(src, i-1))
	}
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		obj := Gia_ManCi(src, i)
		if obj.Value == ^uint32(0) {
			obj.Value = uint32(Gia_ManAppendCi(dst))
		}
	}
	Gia_ManDupRemapCis(dst, src)
	Gia_ManDupRemapCos(dst, src)
	Gia_ManDupRemapEquiv(dst, src)
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	return dst
}
func Gia_ManDupOrderAiger(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManFillValue(src)
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(dst))
	}
	// Preserve even dangling nodes, as the C topological loop does.
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_ManDupOrderDfs_rec(dst, src, obj)
		}
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		Gia_ManDupOrderDfs_rec(dst, src, Gia_ManCo(src, i))
	}
	Gia_ManDupRemapEquiv(dst, src)
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	return dst
}
