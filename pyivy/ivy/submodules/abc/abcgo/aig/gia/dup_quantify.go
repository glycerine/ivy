package gia

// Gia_ManDupCofactorVar_rec hashes a cone after its leaves are mapped.
func Gia_ManDupCofactorVar_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) {
	if obj.Value != ^uint32(0) {
		return
	}
	a, b, c, kind := src.Gate(obj.id)
	read := func(lit int32) int32 {
		fanin := src.pObjs[lit>>1]
		Gia_ManDupCofactorVar_rec(dst, src, fanin)
		return int32(fanin.Value) ^ (lit & 1)
	}
	var value int32
	switch kind {
	case 2:
		value = Gia_ManHashAnd(dst, read(a), read(b))
	case 3:
		value = Gia_ManHashXor(dst, read(a), read(b))
	case 4:
		value = Gia_ManHashMux(dst, read(a), read(b), read(c))
	case 5:
		value = read(a)
	default:
		panic("gia: invalid cofactor cone")
	}
	obj.Value = uint32(value)
}
func giaQuantStart(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManFillValue(src)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	Gia_ManHashAlloc(out)
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	return out
}
func giaQuantOutputs(dst, src *Gia_Man_t) []int32 {
	result := make([]int32, Gia_ManCoNum(src))
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		obj := Gia_ManCo(src, i)
		Gia_ManDupCofactorVar_rec(dst, src, Gia_ObjFanin0(obj))
		result[i] = Gia_ObjFanin0Copy(obj)
	}
	return result
}
func Gia_ManDupCofactorVar(src *Gia_Man_t, variable, value int32) *Gia_Man_t {
	if variable < 0 || variable >= Gia_ManPiNum(src) || (value != 0 && value != 1) {
		panic("gia: invalid cofactor variable")
	}
	out := giaQuantStart(src)
	Gia_ManPi(src, variable).Value = uint32(value)
	outputs := giaQuantOutputs(out, src)
	for _, lit := range outputs {
		Gia_ManAppendCo(out, lit)
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	out.nConstrs = src.nConstrs
	return giaCleanComposed(out)
}
func Gia_ManDupMux(variable int32, positive, negative *Gia_Man_t) *Gia_Man_t {
	if Gia_ManRegNum(positive) != 0 || Gia_ManRegNum(negative) != 0 || Gia_ManCoNum(positive) != 1 || Gia_ManCoNum(negative) != 1 || Gia_ManCiNum(positive) != Gia_ManCiNum(negative) || variable < 0 || variable >= Gia_ManCiNum(negative) {
		panic("gia: invalid cofactor mux interface")
	}
	out := giaQuantStart(negative)
	neg := giaQuantOutputs(out, negative)[0]
	Gia_ManFillValue(positive)
	positive.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(positive); i++ {
		Gia_ManCi(positive, i).Value = Gia_ManCi(negative, i).Value
	}
	pos := giaQuantOutputs(out, positive)[0]
	ctrl := int32(Gia_ManCi(negative, variable).Value)
	Gia_ManAppendCo(out, Gia_ManHashMux(out, ctrl, pos, neg))
	return giaCleanComposed(out)
}
func giaQuantify(src *Gia_Man_t, variable int32, universal, allOutputs bool) *Gia_Man_t {
	if variable < 0 || variable >= Gia_ManPiNum(src) || !allOutputs && Gia_ManRegNum(src) != 0 {
		panic("gia: invalid quantified variable")
	}
	out := giaQuantStart(src)
	Gia_ManPi(src, variable).Value = 0
	negative := giaQuantOutputs(out, src)
	Gia_ManPi(src, variable).Value = 1
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			obj.Value = ^uint32(0)
		}
	}
	positive := giaQuantOutputs(out, src)
	for i := range positive {
		lit := positive[i]
		if universal {
			if i == 0 {
				lit = Gia_ManHashAnd(out, negative[i], lit)
			}
		} else {
			lit = Gia_ManHashOr(out, negative[i], lit)
		}
		Gia_ManAppendCo(out, lit)
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	out.nConstrs = src.nConstrs
	return giaCleanComposed(out)
}
func Gia_ManDupExist(src *Gia_Man_t, variable int32) *Gia_Man_t {
	if Gia_ManPoNum(src) != 1 {
		panic("gia: existential abstraction expects one output")
	}
	return giaQuantify(src, variable, false, false)
}
func Gia_ManDupUniv(src *Gia_Man_t, variable int32) *Gia_Man_t {
	return giaQuantify(src, variable, true, false)
}
func Gia_ManDupExist2(src *Gia_Man_t, variable int32) *Gia_Man_t {
	return giaQuantify(src, variable, false, true)
}

// Gia_ManDupCofactorObj constrains the selected internal object to value and
// guards every output with the corresponding original condition.
func Gia_ManDupCofactorObj(src *Gia_Man_t, id, value int32) *Gia_Man_t {
	if Gia_ManRegNum(src) != 0 || id <= 0 || id >= src.nObjs || Gia_ObjIsCand(src.pObjs[id]) == 0 || (value != 0 && value != 1) {
		panic("gia: invalid internal cofactor")
	}
	out := giaQuantStart(src)
	guard := int32(-1)
	if Gia_ObjIsCi(src.pObjs[id]) != 0 {
		guard = int32(src.pObjs[id].Value) ^ (value ^ 1)
		src.pObjs[id].Value = uint32(value)
	}
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a, b, c, kind := src.Gate(obj.id)
		read := func(lit int32) int32 { return Gia_ObjLitCopy(src, lit) }
		var result int32
		switch kind {
		case 2:
			result = Gia_ManHashAnd(out, read(a), read(b))
		case 3:
			result = Gia_ManHashXor(out, read(a), read(b))
		case 4:
			result = Gia_ManHashMux(out, read(a), read(b), read(c))
		case 5:
			result = read(a)
		default:
			panic("gia: invalid internal cofactor gate")
		}
		obj.Value = uint32(result)
		if obj.id == id {
			guard = result ^ (value ^ 1)
			obj.Value = uint32(value)
		}
	}
	if guard < 0 {
		panic("gia: missing internal cofactor")
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		Gia_ManAppendCo(out, Gia_ManHashAnd(out, giaMappedEdge(src, Gia_ManCo(src, i)), guard))
	}
	return giaCleanComposed(out)
}
func Gia_ManDupBlock(src *Gia_Man_t, block int32) *Gia_Man_t {
	if block <= 0 || Gia_ManCiNum(src)%block != 0 || Gia_ManCoNum(src)%block != 0 {
		panic("gia: invalid block size")
	}
	Gia_ManFillValue(src)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		value := int32(0)
		if i%block == 0 {
			value = Gia_ManAppendCi(out)
		}
		Gia_ManCi(src, i).Value = uint32(value)
	}
	Gia_ManHashAlloc(out)
	giaComposeNodes(out, src, true)
	for i := int32(0); i < Gia_ManCoNum(src); i += block {
		giaDupMappedOutput(out, src, Gia_ManCo(src, i))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src)/block)
	return giaCleanComposed(out)
}
