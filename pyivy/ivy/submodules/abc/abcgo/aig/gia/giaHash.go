package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
)

func Gia_ManHashOne(a, b, c, size int32) int32 {
	key := uint32(c)*2011 + uint32(a>>1)*7937 + uint32(b>>1)*2971 + uint32(a&1)*911 + uint32(b&1)*353
	return int32(key % uint32(size))
}

// HashFind returns a stable vector slot until the next append/resize. Callers
// reacquire it after allocation, as in the original linked-bucket table.
func Gia_ManHashFind(p *Gia_Man_t, a, b, c int32) *int32 {
	if Vec_IntSize(&p.vHTable) == 0 {
		panic("gia: structural hash is not started")
	}
	slot := Vec_IntEntryP(&p.vHTable, int(Gia_ManHashOne(a, b, c, int32(Vec_IntSize(&p.vHTable)))))
	for *slot != 0 {
		id := *slot
		obj := p.pObjs[id]
		if Gia_ObjFaninLit0(obj, id) == a && Gia_ObjFaninLit1(obj, id) == b && (p.PMuxes == nil || Gia_ObjFaninLit2(p, id) == c) {
			break
		}
		slot = Vec_IntEntryP(&p.vHash, int(id))
	}
	return slot
}
func Gia_ManHashLookupInt(p *Gia_Man_t, a, b int32) int32 {
	if a > b {
		a, b = b, a
	}
	return *Gia_ManHashFind(p, a, b, -1) << 1
}
func Gia_ManHashLookup(p *Gia_Man_t, a, b *Gia_Obj_t) int32 {
	return Gia_ManHashLookupInt(p, Gia_ObjToLit(p, a), Gia_ObjToLit(p, b))
}
func Gia_ManHashAlloc(p *Gia_Man_t) {
	if Vec_IntSize(&p.vHTable) != 0 {
		panic("gia: structural hash already started")
	}
	p.literalHash = nil
	size := p.nObjsAlloc
	if Gia_ManAndNum(p) != 0 {
		size = Gia_ManAndNum(p) + 1000
	}
	Vec_IntFill(&p.vHTable, int(util.Abc_PrimeCudd(uint32(size))), 0)
	Vec_IntFill(&p.vHash, int(p.nObjs), 0)
}
func Gia_ManHashStart(p *Gia_Man_t) {
	Gia_ManHashAlloc(p)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && Gia_ObjIsBuf(obj) == 0 {
			slot := Gia_ManHashFind(p, Gia_ObjFaninLit0(obj, obj.id), Gia_ObjFaninLit1(obj, obj.id), Gia_ObjFaninLit2(p, obj.id))
			if *slot != 0 {
				panic("gia: duplicate object in structural hash")
			}
			*slot = obj.id
		}
	}
}
func Gia_ManHashStop(p *Gia_Man_t) {
	Vec_IntErase(&p.vHTable)
	Vec_IntErase(&p.vHash)
	p.literalHash = nil
}
func Gia_ManHashResize(p *Gia_Man_t) {
	old := append([]int32(nil), Vec_IntArray(&p.vHTable)...)
	if len(old) == 0 {
		panic("gia: cannot resize inactive hash")
	}
	Vec_IntFill(&p.vHTable, int(util.Abc_PrimeCudd(uint32(2*Gia_ManAndNum(p)))), 0)
	for _, head := range old {
		for head != 0 {
			obj := p.pObjs[head]
			next := Vec_IntEntry(&p.vHash, int(head))
			Vec_IntWriteEntry(&p.vHash, int(head), 0)
			slot := Gia_ManHashFind(p, Gia_ObjFaninLit0(obj, head), Gia_ObjFaninLit1(obj, head), Gia_ObjFaninLit2(p, head))
			if *slot != 0 {
				panic("gia: duplicate rehashed gate")
			}
			*slot = head
			head = next
		}
	}
}
func Gia_ManHashProfile(p *Gia_Man_t) {
	fmt.Printf("Table size = %d. Entries = %d. Hits = %d. Misses = %d.\n", Vec_IntSize(&p.vHTable), Gia_ManAndNum(p), p.nHashHit, p.nHashMiss)
}
func giaHashIntern(p *Gia_Man_t, a, b, c int32, appendGate func() int32) int32 {
	if p.nObjs&255 == 0 && 2*Vec_IntSize(&p.vHTable) < int(Gia_ManAndNum(p)) {
		Gia_ManHashResize(p)
	}
	slot := Gia_ManHashFind(p, a, b, c)
	if *slot != 0 {
		p.nHashHit++
		return *slot << 1
	}
	p.nHashMiss++
	literal := appendGate()
	slot = Gia_ManHashFind(p, a, b, c)
	*slot = literal >> 1
	return literal
}
func Gia_ManHashAnd(p *Gia_Man_t, a, b int32) int32 {
	giaCheckFanin(p, a)
	giaCheckFanin(p, b)
	if a < 2 {
		if a == 0 {
			return 0
		}
		return b
	}
	if b < 2 {
		if b == 0 {
			return 0
		}
		return a
	}
	if a == b {
		return a
	}
	if a == b^1 {
		return 0
	}
	if p.fGiaSimple != 0 {
		return Gia_ManAppendAnd(p, a, b)
	}
	if p.fAddStrash != 0 {
		if obj := Gia_ManAddStrash(p, Gia_ObjFromLit(p, a), Gia_ObjFromLit(p, b)); obj != nil {
			return Gia_ObjToLit(p, obj)
		}
	}
	if a > b {
		a, b = b, a
	}
	return giaHashIntern(p, a, b, -1, func() int32 { return Gia_ManAppendAnd(p, a, b) })
}
func Gia_ManHashOr(p *Gia_Man_t, a, b int32) int32 { return Gia_ManHashAnd(p, a^1, b^1) ^ 1 }
func Gia_ManHashAndTry(p *Gia_Man_t, a, b int32) int32 {
	if a < 2 {
		if a == 0 {
			return 0
		}
		return b
	}
	if b < 2 {
		if b == 0 {
			return 0
		}
		return a
	}
	if a == b {
		return a
	}
	if a == b^1 {
		return 0
	}
	if lit := Gia_ManHashLookupInt(p, a, b); lit != 0 {
		return lit
	}
	return -1
}
func Gia_ManHashXorReal(p *Gia_Man_t, a, b int32) int32 {
	if p.fAddStrash != 0 {
		panic("gia: additional strashing is incompatible with real XORs")
	}
	if a < 2 {
		return b ^ a
	}
	if b < 2 {
		return a ^ b
	}
	if a == b {
		return 0
	}
	if a == b^1 {
		return 1
	}
	if a < b {
		a, b = b, a
	}
	phase := (a ^ b) & 1
	a, b = a&^1, b&^1
	return giaHashIntern(p, a, b, -1, func() int32 { return Gia_ManAppendXorReal(p, a, b) }) ^ phase
}
func Gia_ManHashMuxReal(p *Gia_Man_t, control, then, otherwise int32) int32 {
	if p.fAddStrash != 0 {
		panic("gia: additional strashing is incompatible with real MUXs")
	}
	if control < 2 {
		if control == 1 {
			return then
		}
		return otherwise
	}
	if otherwise < 2 {
		if otherwise == 1 {
			return Gia_ManHashOr(p, control^1, then)
		}
		return Gia_ManHashAnd(p, control, then)
	}
	if then < 2 {
		if then == 1 {
			return Gia_ManHashOr(p, control, otherwise)
		}
		return Gia_ManHashAnd(p, control^1, otherwise)
	}
	if otherwise == then {
		return otherwise
	}
	if control == otherwise || control == then^1 {
		return Gia_ManHashAnd(p, otherwise, then)
	}
	if control == then || control == otherwise^1 {
		return Gia_ManHashOr(p, otherwise, then)
	}
	if otherwise>>1 == then>>1 {
		return Gia_ManHashXorReal(p, control, otherwise)
	}
	if otherwise > then {
		otherwise, then = then, otherwise
		control ^= 1
	}
	phase := then & 1
	if phase != 0 {
		otherwise ^= 1
		then ^= 1
	}
	return giaHashIntern(p, otherwise, then, control, func() int32 { return Gia_ManAppendMuxReal(p, control, then, otherwise) }) ^ phase
}
func Gia_ManHashXor(p *Gia_Man_t, a, b int32) int32 {
	if p.fGiaSimple != 0 {
		return Gia_ManHashOr(p, Gia_ManHashAnd(p, a, b^1), Gia_ManHashAnd(p, a^1, b))
	}
	phase := (a ^ b) & 1
	x := Gia_ManHashAnd(p, a&^1, (b&^1)^1)
	y := Gia_ManHashAnd(p, b&^1, (a&^1)^1)
	return Gia_ManHashAnd(p, x^1, y^1) ^ (phase ^ 1)
}
func Gia_ManHashMux(p *Gia_Man_t, control, then, otherwise int32) int32 {
	if p.fGiaSimple != 0 {
		return Gia_ManHashOr(p, Gia_ManHashAnd(p, control, then), Gia_ManHashAnd(p, control^1, otherwise))
	}
	if otherwise > then {
		otherwise, then = then, otherwise
		control ^= 1
	}
	phase := then & 1
	if phase != 0 {
		otherwise ^= 1
		then ^= 1
	}
	x := Gia_ManHashAnd(p, control^1, otherwise)
	y := Gia_ManHashAnd(p, control, then)
	return Gia_ManHashAnd(p, x^1, y^1) ^ (phase ^ 1)
}
func Gia_ManHashMaj(p *Gia_Man_t, a, b, c int32) int32 {
	return Gia_ManHashOr(p, Gia_ManHashAnd(p, a, Gia_ManHashOr(p, b, c)), Gia_ManHashAnd(p, b, c))
}
func giaHashOpposite(a, b *Gia_Obj_t) bool { return a != nil && b != nil && a == Gia_Not(b) }
func Gia_ObjRecognizeMuxTwo(pNode0 *Gia_Obj_t, pNode1 *Gia_Obj_t, ppNodeT **Gia_Obj_t, ppNodeE **Gia_Obj_t) *Gia_Obj_t {
	if !(Gia_IsComplement(pNode0) == 0) {
		panic("gia: invalid structural hashing precondition")
	}
	if !(Gia_IsComplement(pNode1) == 0) {
		panic("gia: invalid structural hashing precondition")
	}

	if Gia_ObjFanin1(pNode0) == Gia_ObjFanin1(pNode1) && (Gia_ObjFaninC1(pNode0) != Gia_ObjFaninC1(pNode1)) {

		if Gia_ObjFaninC1(pNode0) != 0 {
			*ppNodeT = Gia_Not(Gia_ObjChild0(pNode1))
			*ppNodeE = Gia_Not(Gia_ObjChild0(pNode0))
			return Gia_ObjChild1(pNode1)

		} else {
			*ppNodeT = Gia_Not(Gia_ObjChild0(pNode0))
			*ppNodeE = Gia_Not(Gia_ObjChild0(pNode1))
			return Gia_ObjChild1(pNode0)
		}

	} else if Gia_ObjFanin0(pNode0) == Gia_ObjFanin0(pNode1) && (Gia_ObjFaninC0(pNode0) != Gia_ObjFaninC0(pNode1)) {

		if Gia_ObjFaninC0(pNode0) != 0 {
			*ppNodeT = Gia_Not(Gia_ObjChild1(pNode1))
			*ppNodeE = Gia_Not(Gia_ObjChild1(pNode0))
			return Gia_ObjChild0(pNode1)

		} else {
			*ppNodeT = Gia_Not(Gia_ObjChild1(pNode0))
			*ppNodeE = Gia_Not(Gia_ObjChild1(pNode1))
			return Gia_ObjChild0(pNode0)
		}

	} else if Gia_ObjFanin0(pNode0) == Gia_ObjFanin1(pNode1) && (Gia_ObjFaninC0(pNode0) != Gia_ObjFaninC1(pNode1)) {

		if Gia_ObjFaninC0(pNode0) != 0 {
			*ppNodeT = Gia_Not(Gia_ObjChild0(pNode1))
			*ppNodeE = Gia_Not(Gia_ObjChild1(pNode0))
			return Gia_ObjChild1(pNode1)

		} else {
			*ppNodeT = Gia_Not(Gia_ObjChild1(pNode0))
			*ppNodeE = Gia_Not(Gia_ObjChild0(pNode1))
			return Gia_ObjChild0(pNode0)
		}

	} else if Gia_ObjFanin1(pNode0) == Gia_ObjFanin0(pNode1) && (Gia_ObjFaninC1(pNode0) != Gia_ObjFaninC0(pNode1)) {

		if Gia_ObjFaninC1(pNode0) != 0 {
			*ppNodeT = Gia_Not(Gia_ObjChild1(pNode1))
			*ppNodeE = Gia_Not(Gia_ObjChild0(pNode0))
			return Gia_ObjChild0(pNode1)

		} else {
			*ppNodeT = Gia_Not(Gia_ObjChild0(pNode0))
			*ppNodeE = Gia_Not(Gia_ObjChild1(pNode1))
			return Gia_ObjChild1(pNode0)
		}
	}
	if !(false) {
		panic("gia: invalid structural hashing precondition")
	}
	return nil
}

func Gia_ManHashAndP(p *Gia_Man_t, p0 *Gia_Obj_t, p1 *Gia_Obj_t) *Gia_Obj_t {
	return Gia_ObjFromLit(p, Gia_ManHashAnd(p, Gia_ObjToLit(p, p0), Gia_ObjToLit(p, p1)))
}

func Gia_ManAddStrash(p *Gia_Man_t, p0 *Gia_Obj_t, p1 *Gia_Obj_t) *Gia_Obj_t {
	var pNode0, pNode1, pFanA, pFanB, pFanC, pFanD *Gia_Obj_t
	if !(p.fAddStrash != 0) {
		panic("gia: invalid structural hashing precondition")
	}
	pNode0 = Gia_Regular(p0)
	pNode1 = Gia_Regular(p1)
	if Gia_ObjIsAnd(pNode0) == 0 && Gia_ObjIsAnd(pNode1) == 0 {
		return nil
	}
	if Gia_ObjIsAnd(pNode0) != 0 {
		pFanA = Gia_ObjChild0(pNode0)
	} else {
		pFanA = nil
	}
	if Gia_ObjIsAnd(pNode0) != 0 {
		pFanB = Gia_ObjChild1(pNode0)
	} else {
		pFanB = nil
	}
	if Gia_ObjIsAnd(pNode1) != 0 {
		pFanC = Gia_ObjChild0(pNode1)
	} else {
		pFanC = nil
	}
	if Gia_ObjIsAnd(pNode1) != 0 {
		pFanD = Gia_ObjChild1(pNode1)
	} else {
		pFanD = nil
	}
	if Gia_IsComplement(p0) != 0 {
		if pFanA == Gia_Not(p1) || pFanB == Gia_Not(p1) {
			return p1
		}
		if pFanB == p1 {
			return Gia_ManHashAndP(p, Gia_Not(pFanA), pFanB)
		}
		if pFanA == p1 {
			return Gia_ManHashAndP(p, Gia_Not(pFanB), pFanA)
		}

	} else {
		if pFanA == Gia_Not(p1) || pFanB == Gia_Not(p1) {
			return Gia_ManConst0(p)
		}
		if pFanA == p1 || pFanB == p1 {
			return p0
		}
	}
	if Gia_IsComplement(p1) != 0 {
		if pFanC == Gia_Not(p0) || pFanD == Gia_Not(p0) {
			return p0
		}
		if pFanD == p0 {
			return Gia_ManHashAndP(p, Gia_Not(pFanC), pFanD)
		}
		if pFanC == p0 {
			return Gia_ManHashAndP(p, Gia_Not(pFanD), pFanC)
		}

	} else {
		if pFanC == Gia_Not(p0) || pFanD == Gia_Not(p0) {
			return Gia_ManConst0(p)
		}
		if pFanC == p0 || pFanD == p0 {
			return p1
		}
	}
	if Gia_IsComplement(p0) == 0 && Gia_IsComplement(p1) == 0 {
		if giaHashOpposite(pFanA, pFanC) || giaHashOpposite(pFanA, pFanD) || giaHashOpposite(pFanB, pFanC) || giaHashOpposite(pFanB, pFanD) {
			return Gia_ManConst0(p)
		}
		if pFanA == pFanC || pFanB == pFanC {
			return Gia_ManHashAndP(p, p0, pFanD)
		}
		if pFanB == pFanC || pFanB == pFanD {
			return Gia_ManHashAndP(p, pFanA, p1)
		}
		if pFanA == pFanD || pFanB == pFanD {
			return Gia_ManHashAndP(p, p0, pFanC)
		}
		if pFanA == pFanC || pFanA == pFanD {
			return Gia_ManHashAndP(p, pFanB, p1)
		}

	} else if Gia_IsComplement(p0) != 0 && Gia_IsComplement(p1) == 0 {
		if giaHashOpposite(pFanA, pFanC) || giaHashOpposite(pFanA, pFanD) || giaHashOpposite(pFanB, pFanC) || giaHashOpposite(pFanB, pFanD) {
			return p1
		}
		if pFanB == pFanC || pFanB == pFanD {
			return Gia_ManHashAndP(p, Gia_Not(pFanA), p1)
		}
		if pFanA == pFanC || pFanA == pFanD {
			return Gia_ManHashAndP(p, Gia_Not(pFanB), p1)
		}

	} else if Gia_IsComplement(p0) == 0 && Gia_IsComplement(p1) != 0 {
		if giaHashOpposite(pFanC, pFanA) || giaHashOpposite(pFanC, pFanB) || giaHashOpposite(pFanD, pFanA) || giaHashOpposite(pFanD, pFanB) {
			return p0
		}
		if pFanD == pFanA || pFanD == pFanB {
			return Gia_ManHashAndP(p, Gia_Not(pFanC), p0)
		}
		if pFanC == pFanA || pFanC == pFanB {
			return Gia_ManHashAndP(p, Gia_Not(pFanD), p0)
		}

	} else {

		if pFanA == pFanD && giaHashOpposite(pFanB, pFanC) {
			return Gia_Not(pFanA)
		}
		if pFanB == pFanC && giaHashOpposite(pFanA, pFanD) {
			return Gia_Not(pFanB)
		}
		if pFanA == pFanC && giaHashOpposite(pFanB, pFanD) {
			return Gia_Not(pFanA)
		}
		if pFanB == pFanD && giaHashOpposite(pFanA, pFanC) {
			return Gia_Not(pFanB)
		}
	}

	return nil
}

func Gia_ManRehash(src *Gia_Man_t, additional int32) *Gia_Man_t {
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	dst.fAddStrash = additional
	Gia_ManHashAlloc(dst)
	src.pObjs[0].Value = 0
	for _, obj := range src.pObjs[1:] {
		switch {
		case Gia_ObjIsAnd(obj) != 0:
			obj.Value = uint32(Gia_ManHashAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		case Gia_ObjIsCi(obj) != 0:
			obj.Value = uint32(Gia_ManAppendCi(dst))
		case Gia_ObjIsCo(obj) != 0:
			obj.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj)))
		}
	}
	Gia_ManHashStop(dst)
	dst.fAddStrash = 0
	Gia_ManSetRegNum(dst, src.nRegs)
	dst.nConstrs = src.nConstrs
	result := Gia_ManCleanup(dst)
	Gia_ManStop(dst)
	return result
}
func Gia_ManHashAndMulti(p *Gia_Man_t, literals *Vec_Int_t) int32 {
	if Vec_IntSize(literals) == 0 {
		return 0
	}
	for Vec_IntSize(literals) > 1 {
		array := Vec_IntArray(literals)
		count := 0
		for i := 0; i+1 < len(array); i += 2 {
			array[count] = Gia_ManHashAnd(p, array[i], array[i+1])
			count++
		}
		if len(array)&1 != 0 {
			array[count] = array[len(array)-1]
			count++
		}
		Vec_IntShrink(literals, count)
	}
	return Vec_IntEntry(literals, 0)
}
func Gia_ManHashAndMulti2(p *Gia_Man_t, literals *Vec_Int_t) int32 {
	result := int32(1)
	for _, lit := range Vec_IntArray(literals) {
		result = Gia_ManHashAnd(p, result, lit)
	}
	return result
}
func Gia_ManHashDualMiter(p *Gia_Man_t, outputs *Vec_Int_t) int32 {
	result := int32(0)
	array := Vec_IntArray(outputs)
	for i := 0; i+1 < len(array); i += 2 {
		result = Gia_ManHashOr(p, result, Gia_ManHashXor(p, array[i], array[i+1]))
	}
	return result
}
