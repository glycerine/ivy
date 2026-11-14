package gia

import "github.com/glycerine/abcgo/misc/vec"

// giaDupMappedNodes copies all internal gates after source CIs have been mapped.
func giaDupMappedNodes(dst, src *Gia_Man_t) {
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a, b, c, kind := src.Gate(obj.id)
		copyLit := func(lit int32) int32 {
			value := src.pObjs[lit>>1].Value
			if value == ^uint32(0) {
				panic("gia: removed input is still used")
			}
			return int32(value) ^ (lit & 1)
		}
		var result int32
		switch kind {
		case 2:
			result = Gia_ManAppendAnd(dst, copyLit(a), copyLit(b))
		case 3:
			result = Gia_ManAppendXorReal(dst, copyLit(a), copyLit(b))
		case 4:
			result = Gia_ManAppendMuxReal(dst, copyLit(a), copyLit(b), copyLit(c))
		case 5:
			result = Gia_ManAppendBuf(dst, copyLit(a))
		default:
			panic("gia: invalid gate in mapped duplication")
		}
		obj.Value = uint32(result)
	}
}
func giaDupMappedStart(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	return dst
}
func giaDupMappedOutput(dst, src *Gia_Man_t, obj *Gia_Obj_t) {
	edge, _, _, _ := src.Gate(obj.id)
	value := src.pObjs[edge>>1].Value
	if value == ^uint32(0) {
		panic("gia: removed input drives an output")
	}
	obj.Value = uint32(Gia_ManAppendCo(dst, int32(value)^(edge&1)))
}
func giaValidatePermutation(perm *Vec_Int_t, n int32) []int32 {
	if vec.Vec_IntSize(perm) != int(n) {
		panic("gia: permutation size mismatch")
	}
	values := vec.Vec_IntArray(perm)
	seen := make([]bool, n)
	for _, v := range values {
		if v < 0 || v >= n || seen[v] {
			panic("gia: invalid permutation")
		}
		seen[v] = true
	}
	return values
}
func Gia_ManDupZero(src *Gia_Man_t) *Gia_Man_t {
	out := Gia_ManStart(1 + Gia_ManCiNum(src) + Gia_ManCoNum(src))
	out.pName = src.pName
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManAppendCi(out)
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		Gia_ManAppendCo(out, 0)
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return out
}
func Gia_ManDupPerm(src *Gia_Man_t, perm *Vec_Int_t) *Gia_Man_t {
	order := giaValidatePermutation(perm, Gia_ManPiNum(src))
	out := giaDupMappedStart(src)
	for _, i := range order {
		Gia_ManPi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		Gia_ManRo(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	giaDupMappedNodes(out, src)
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManCo(src, i))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return out
}
func Gia_ManDupPermFlop(src *Gia_Man_t, perm *Vec_Int_t) *Gia_Man_t {
	order := giaValidatePermutation(perm, Gia_ManRegNum(src))
	out := giaDupMappedStart(src)
	for i := int32(0); i < Gia_ManPiNum(src); i++ {
		Gia_ManPi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	for _, i := range order {
		Gia_ManRo(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	giaDupMappedNodes(out, src)
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManPo(src, i))
	}
	for _, i := range order {
		giaDupMappedOutput(out, src, Gia_ManRi(src, i))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return out
}
func Gia_ManDupSpreadFlop(src *Gia_Man_t, mask *Vec_Int_t) *Gia_Man_t {
	slots := vec.Vec_IntArray(mask)
	count := 0
	for _, v := range slots {
		if v != -1 {
			count++
		}
	}
	if count != int(Gia_ManRegNum(src)) {
		panic("gia: flop mask size mismatch")
	}
	out := giaDupMappedStart(src)
	for i := int32(0); i < Gia_ManPiNum(src); i++ {
		Gia_ManPi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	index := int32(0)
	for _, v := range slots {
		lit := Gia_ManAppendCi(out)
		if v != -1 {
			Gia_ManRo(src, index).Value = uint32(lit)
			index++
		}
	}
	giaDupMappedNodes(out, src)
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManPo(src, i))
	}
	index = 0
	for _, v := range slots {
		if v == -1 {
			Gia_ManAppendCo(out, 0)
		} else {
			giaDupMappedOutput(out, src, Gia_ManRi(src, index))
			index++
		}
	}
	Gia_ManSetRegNum(out, int32(len(slots)))
	return out
}
func Gia_ManDupPermFlopGap(src *Gia_Man_t, mask *Vec_Int_t) *Gia_Man_t {
	perm := vec.Vec_IntAlloc(int(Gia_ManRegNum(src)))
	for _, id := range vec.Vec_IntArray(mask) {
		if id != -1 {
			vec.Vec_IntPush(perm, id)
		}
	}
	interim := Gia_ManDupPermFlop(src, perm)
	out := Gia_ManDupSpreadFlop(interim, mask)
	Gia_ManStop(interim)
	return out
}
func Gia_ManDupLastPis(src *Gia_Man_t, n int32) *Gia_Man_t {
	if Gia_ManRegNum(src) != 0 || n < 0 || n > Gia_ManPiNum(src) {
		panic("gia: invalid retained primary inputs")
	}
	out := giaDupMappedStart(src)
	for i := Gia_ManPiNum(src) - n; i < Gia_ManPiNum(src); i++ {
		Gia_ManPi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	giaDupMappedNodes(out, src)
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManCo(src, i))
	}
	return out
}
func Gia_ManDupRemovePis(src *Gia_Man_t, n int32) *Gia_Man_t {
	if n < 0 || n > Gia_ManPiNum(src) {
		panic("gia: invalid removed primary inputs")
	}
	out := giaDupMappedStart(src)
	keep := Gia_ManPiNum(src) - n
	for i := int32(0); i < keep; i++ {
		Gia_ManPi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		Gia_ManRo(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	giaDupMappedNodes(out, src)
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManCo(src, i))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return out
}

func Gia_ManDupOnsetOffset(src *Gia_Man_t) *Gia_Man_t {
	out := giaDupMappedStart(src)
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	giaDupMappedNodes(out, src)
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		obj := Gia_ManCo(src, i)
		edge, _, _, _ := src.Gate(obj.id)
		lit := int32(src.pObjs[edge>>1].Value) ^ (edge & 1)
		Gia_ManAppendCo(out, lit)
		obj.Value = uint32(Gia_ManAppendCo(out, lit^1))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return out
}

// Gia_ManDupPiPerm randomizes CI-driven edges independently, as the source
// experiment does; it is a graph mutation rather than a bijective permutation.
func Gia_ManDupPiPerm(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManRandom(1)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	Gia_ManHashAlloc(out)
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	count := Gia_ManCiNum(src)
	for i := int32(0); i < count; i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	randomize := func(lit int32) int32 {
		if count == 0 {
			return lit
		}
		place := int32(Gia_ManRandom(0) % uint32(count))
		if lit>>1 <= count {
			return (place+1)<<1 | lit&1
		}
		return lit
	}
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a, b, c, kind := src.Gate(obj.id)
		read := func(lit int32) int32 { return Gia_ObjLitCopy(src, lit) }
		var lit int32
		switch kind {
		case 2:
			lit = Gia_ManHashAnd(out, randomize(read(a)), randomize(read(b)))
		case 3:
			lit = Gia_ManHashXor(out, randomize(read(a)), randomize(read(b)))
		case 4:
			lit = Gia_ManHashMux(out, randomize(read(a)), randomize(read(b)), randomize(read(c)))
		case 5:
			lit = randomize(read(a))
		default:
			panic("gia: invalid random mutation gate")
		}
		obj.Value = uint32(lit)
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManCo(src, i))
	}
	Gia_ManHashStop(out)
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return giaCleanComposed(out)
}

func Gia_ManPermuteInputs(src *Gia_Man_t, ppis, extra int32) *Gia_Man_t {
	n := Gia_ManPiNum(src)
	if ppis < 0 || extra < 0 || ppis+extra > n {
		panic("gia: invalid input groups")
	}
	order := vec.Vec_IntAlloc(int(n))
	for i := int32(0); i < n-ppis-extra; i++ {
		vec.Vec_IntPush(order, i)
	}
	for i := n - extra; i < n; i++ {
		vec.Vec_IntPush(order, i)
	}
	for i := n - ppis - extra; i < n-extra; i++ {
		vec.Vec_IntPush(order, i)
	}
	return Gia_ManDupPerm(src, order)
}
