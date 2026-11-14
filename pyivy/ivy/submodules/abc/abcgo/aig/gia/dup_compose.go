package gia

import "github.com/glycerine/abcgo/misc/vec"

// giaComposeNodes duplicates the source's internal logic using its Value mapping.
func giaComposeNodes(dst, src *Gia_Man_t, hashed bool) {
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a, b, c, kind := src.Gate(obj.id)
		read := func(lit int32) int32 {
			value := src.pObjs[lit>>1].Value
			if value == ^uint32(0) {
				panic("gia: unmapped composition fanin")
			}
			return int32(value) ^ (lit & 1)
		}
		var lit int32
		switch kind {
		case 2:
			if hashed {
				lit = Gia_ManHashAnd(dst, read(a), read(b))
			} else {
				lit = Gia_ManAppendAnd(dst, read(a), read(b))
			}
		case 3:
			if hashed {
				lit = Gia_ManHashXor(dst, read(a), read(b))
			} else {
				lit = Gia_ManAppendXor(dst, read(a), read(b))
			}
		case 4:
			if hashed {
				lit = Gia_ManHashMux(dst, read(a), read(b), read(c))
			} else {
				lit = Gia_ManAppendMux(dst, read(a), read(b), read(c))
			}
		case 5:
			if hashed {
				lit = read(a)
			} else {
				lit = Gia_ManAppendBuf(dst, read(a))
			}
		default:
			panic("gia: unsupported composition gate")
		}
		obj.Value = uint32(lit)
	}
}
func Gia_ManDup2(first, second *Gia_Man_t) *Gia_Man_t {
	if Gia_ManCiNum(first) != Gia_ManCiNum(second) || Gia_ManCoNum(first) != Gia_ManCoNum(second) {
		panic("gia: graph interface mismatch")
	}
	out := Gia_ManStart(first.nObjs + second.nObjs)
	Gia_ManHashAlloc(out)
	Gia_ManFillValue(first)
	Gia_ManFillValue(second)
	first.pObjs[0].Value = 0
	second.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(first); i++ {
		lit := uint32(Gia_ManAppendCi(out))
		Gia_ManCi(first, i).Value = lit
		Gia_ManCi(second, i).Value = lit
	}
	giaComposeNodes(out, first, true)
	giaComposeNodes(out, second, true)
	for _, p := range []*Gia_Man_t{first, second} {
		for i := int32(0); i < Gia_ManCoNum(p); i++ {
			giaDupMappedOutput(out, p, Gia_ManCo(p, i))
		}
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(first))
	Gia_ManHashStop(out)
	return out
}
func Gia_ManDupAppend(dst, src *Gia_Man_t) {
	dst.nRegs = 0
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	if Vec_IntSize(&dst.vHTable) == 0 {
		Gia_ManHashStart(dst)
	}
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	// Source-order appending preserves the original interleaving of I/Os.
	for _, obj := range src.pObjs[1:] {
		_, _, _, kind := src.Gate(obj.id)
		if kind == 1 {
			obj.Value = uint32(Gia_ManAppendCi(dst))
		} else if kind == 6 {
			giaDupMappedOutput(dst, src, obj)
		} else {
			Gia_ManDupOrderDfs_rec(dst, src, obj)
		}
	}
}
func Gia_ManDupAppendShare(dst, src *Gia_Man_t) {
	if Gia_ManCiNum(dst) != Gia_ManCiNum(src) {
		panic("gia: shared graph input count mismatch")
	}
	if Vec_IntSize(&dst.vHTable) == 0 {
		Gia_ManHashStart(dst)
	}
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ObjToLit(dst, Gia_ManCi(dst, i)))
	}
	giaComposeNodes(dst, src, true)
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		giaDupMappedOutput(dst, src, Gia_ManCo(src, i))
	}
}
func Gia_ManDupAppendNew(first, second *Gia_Man_t) *Gia_Man_t {
	if Gia_ManPiNum(first) != Gia_ManPiNum(second) {
		panic("gia: primary input count mismatch")
	}
	out := Gia_ManStart(first.nObjs + second.nObjs)
	out.CopyIdentityFrom(first)
	Gia_ManHashAlloc(out)
	Gia_ManFillValue(first)
	Gia_ManFillValue(second)
	first.pObjs[0].Value = 0
	second.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManPiNum(first); i++ {
		lit := uint32(Gia_ManAppendCi(out))
		Gia_ManPi(first, i).Value = lit
		Gia_ManPi(second, i).Value = lit
	}
	for _, p := range []*Gia_Man_t{first, second} {
		for i := int32(0); i < Gia_ManRegNum(p); i++ {
			Gia_ManRo(p, i).Value = uint32(Gia_ManAppendCi(out))
		}
	}
	giaComposeNodes(out, first, true)
	giaComposeNodes(out, second, true)
	Gia_ManHashStop(out)
	for _, p := range []*Gia_Man_t{first, second} {
		for i := int32(0); i < Gia_ManPoNum(p); i++ {
			giaDupMappedOutput(out, p, Gia_ManPo(p, i))
		}
	}
	for _, p := range []*Gia_Man_t{first, second} {
		for i := int32(0); i < Gia_ManRegNum(p); i++ {
			giaDupMappedOutput(out, p, Gia_ManRi(p, i))
		}
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(first)+Gia_ManRegNum(second))
	return out
}

func Gia_ManDupOntop(first, second *Gia_Man_t) *Gia_Man_t {
	if Gia_ManRegNum(first) != 0 || Gia_ManRegNum(second) != 0 || Gia_ManPoNum(first) != Gia_ManPiNum(second) {
		panic("gia: incompatible cascade interfaces")
	}
	out := giaQuantStart(first)
	giaComposeNodes(out, first, true)
	Gia_ManFillValue(second)
	second.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManPiNum(second); i++ {
		Gia_ManPi(second, i).Value = uint32(giaMappedEdge(first, Gia_ManPo(first, i)))
	}
	giaComposeNodes(out, second, true)
	for i := int32(0); i < Gia_ManCoNum(second); i++ {
		giaDupMappedOutput(out, second, Gia_ManCo(second, i))
	}
	Gia_ManHashStop(out)
	return giaCleanComposed(out)
}
func Gia_ManDupWithNewPo(src, property *Gia_Man_t) *Gia_Man_t {
	if Gia_ManRegNum(property) != 0 || Gia_ManPiNum(property) != Gia_ManRegNum(src) {
		panic("gia: incompatible property interface")
	}
	out := giaQuantStart(src)
	giaComposeNodes(out, src, true)
	Gia_ManFillValue(property)
	property.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManPiNum(property); i++ {
		Gia_ManPi(property, i).Value = Gia_ManRo(src, i).Value
	}
	giaComposeNodes(out, property, true)
	for i := int32(0); i < Gia_ManPoNum(property); i++ {
		giaDupMappedOutput(out, property, Gia_ManPo(property, i))
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManRi(src, i))
	}
	Gia_ManHashStop(out)
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return giaCleanComposed(out)
}

func Gia_ManDupWithConstraints(src *Gia_Man_t, types *Vec_Int_t) *Gia_Man_t {
	if vec.Vec_IntSize(types) != int(Gia_ManPoNum(src)) {
		panic("gia: invalid PO type count")
	}
	out := giaDupMappedStart(src)
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	giaDupMappedNodes(out, src)
	var constraints int32
	for phase := int32(0); phase <= 1; phase++ {
		for i := int32(0); i < Gia_ManPoNum(src); i++ {
			kind := vec.Vec_IntEntry(types, int(i))
			if kind != 0 && kind != 1 {
				panic("gia: invalid PO type")
			}
			if kind != phase {
				continue
			}
			obj := Gia_ManPo(src, i)
			lit := giaMappedEdge(src, obj) ^ phase
			obj.Value = uint32(Gia_ManAppendCo(out, lit))
			if phase == 1 {
				constraints++
			}
		}
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManRi(src, i))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	out.nConstrs = constraints
	return out
}
