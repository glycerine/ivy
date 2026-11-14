package gia

import "sync/atomic"

// Gia_Obj_t retains the source representation: fanins are backward index
// differences and complemented-edge flags are separate bits.
type objectStore struct {
	objects *[]*Gia_Obj_t
	// Manager-backed objects retain a view of the live companion table. The
	// pointer to its slice header survives late initialization and growth.
	muxes *[]uint32
}

type Gia_Obj_t struct {
	pair         atomic.Pointer[Gia_Obj_t]
	complemented bool
	store        *objectStore
	id           int32  // stable object index in its owning manager
	iDiff0       uint32 /* C bitfield: 29 bits; review packing and truncation */ // the diff of the first fanin
	fCompl0      uint32 /* C bitfield: 1 bits; review packing and truncation */  // the complemented attribute
	fMark0       uint32 /* C bitfield: 1 bits; review packing and truncation */  // first user-controlled mark
	fTerm        uint32 /* C bitfield: 1 bits; review packing and truncation */  // terminal node (CI/CO)

	iDiff1  uint32 /* C bitfield: 29 bits; review packing and truncation */ // the diff of the second fanin
	fCompl1 uint32 /* C bitfield: 1 bits; review packing and truncation */  // the complemented attribute
	fMark1  uint32 /* C bitfield: 1 bits; review packing and truncation */  // second user-controlled mark
	fPhase  uint32 /* C bitfield: 1 bits; review packing and truncation */  // value under 000 pattern

	Value uint32 // application-specific value
}

const objectNoFanin uint32 = (1 << 29) - 1

func giaBool(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
func Gia_ObjIsTerm(p *Gia_Obj_t) int32        { return giaBool(p.fTerm != 0) }
func Gia_ObjIsAndOrConst0(p *Gia_Obj_t) int32 { return giaBool(p.fTerm == 0) }
func Gia_ObjIsCi(p *Gia_Obj_t) int32          { return giaBool(p.fTerm != 0 && p.iDiff0 == objectNoFanin) }
func Gia_ObjIsCo(p *Gia_Obj_t) int32          { return giaBool(p.fTerm != 0 && p.iDiff0 != objectNoFanin) }
func Gia_ObjIsAnd(p *Gia_Obj_t) int32         { return giaBool(p.fTerm == 0 && p.iDiff0 != objectNoFanin) }
func Gia_ObjIsXor(p *Gia_Obj_t) int32         { return giaBool(Gia_ObjIsAnd(p) != 0 && p.iDiff0 < p.iDiff1) }
func Gia_ObjIsBuf(p *Gia_Obj_t) int32 {
	return giaBool(p.iDiff0 == p.iDiff1 && p.iDiff0 != objectNoFanin && p.fTerm == 0 && p.fCompl0 == p.fCompl1)
}
func Gia_ObjIsAndNotBuf(p *Gia_Obj_t) int32 {
	return giaBool(Gia_ObjIsAnd(p) != 0 && Gia_ObjIsBuf(p) == 0)
}
func Gia_ObjIsCand(p *Gia_Obj_t) int32 { return giaBool(Gia_ObjIsAnd(p) != 0 || Gia_ObjIsCi(p) != 0) }
func Gia_ObjIsConst0(p *Gia_Obj_t) int32 {
	return giaBool(p.iDiff0 == objectNoFanin && p.iDiff1 == objectNoFanin)
}

func Gia_ObjCioId(p *Gia_Obj_t) int32 {
	if p.fTerm == 0 {
		panic("gia: CIO id requested for nonterminal")
	}
	return int32(p.iDiff1)
}
func Gia_ObjSetCioId(p *Gia_Obj_t, id int32) {
	if p.fTerm == 0 || id < 0 || uint32(id) > objectNoFanin {
		panic("gia: invalid CIO id")
	}
	p.iDiff1 = uint32(id)
}
func Gia_ObjValue(p *Gia_Obj_t) int32              { return int32(p.Value) }
func Gia_ObjSetValue(p *Gia_Obj_t, value int32)    { p.Value = uint32(value) }
func Gia_ObjPhase(p *Gia_Obj_t) int32              { return int32(p.fPhase) }
func Gia_ObjFaninC0(p *Gia_Obj_t) int32            { return int32(p.fCompl0) }
func Gia_ObjFaninC1(p *Gia_Obj_t) int32            { return int32(p.fCompl1) }
func Gia_ObjFaninId0(p *Gia_Obj_t, id int32) int32 { return id - int32(p.iDiff0) }
func Gia_ObjFaninId1(p *Gia_Obj_t, id int32) int32 { return id - int32(p.iDiff1) }
func Gia_ObjFaninLit0(p *Gia_Obj_t, id int32) int32 {
	return Gia_ObjFaninId0(p, id)<<1 | Gia_ObjFaninC0(p)
}
func Gia_ObjFaninLit1(p *Gia_Obj_t, id int32) int32 {
	return Gia_ObjFaninId1(p, id)<<1 | Gia_ObjFaninC1(p)
}
func Gia_ObjFlipFaninC0(p *Gia_Obj_t) {
	if Gia_ObjIsCo(p) == 0 {
		panic("gia: expected CO")
	}
	p.fCompl0 ^= 1
}

func Gia_ObjFanin0(p *Gia_Obj_t) *Gia_Obj_t { return p.fanin(p.iDiff0) }
func Gia_ObjFanin1(p *Gia_Obj_t) *Gia_Obj_t { return p.fanin(p.iDiff1) }
func (p *Gia_Obj_t) fanin(diff uint32) *Gia_Obj_t {
	if p.store == nil || diff == 0 || diff > uint32(p.id) {
		panic("gia: invalid fanin")
	}
	return (*p.store.objects)[p.id-int32(diff)]
}

// Complemented edge handles replace C's tagged pointer bit. They are stable
// paired views; callers unwrap with Gia_Regular before accessing object fields.
func Gia_Regular(p *Gia_Obj_t) *Gia_Obj_t {
	if p != nil && p.complemented {
		return p.pair.Load()
	}
	return p
}
func Gia_IsComplement(p *Gia_Obj_t) int32 { return giaBool(p != nil && p.complemented) }
func Gia_Not(p *Gia_Obj_t) *Gia_Obj_t {
	if p == nil {
		panic("gia: cannot complement a nil object")
	}
	if p.complemented {
		return p.pair.Load()
	}
	if inverse := p.pair.Load(); inverse != nil {
		return inverse
	}
	inverse := &Gia_Obj_t{complemented: true}
	inverse.pair.Store(p)
	if p.pair.CompareAndSwap(nil, inverse) {
		return inverse
	}
	return p.pair.Load()
}
func Gia_NotCond(p *Gia_Obj_t, condition int32) *Gia_Obj_t {
	if condition&1 != 0 {
		return Gia_Not(p)
	}
	return p
}

func Gia_ObjFanin0Copy(p *Gia_Obj_t) int32 { return int32(Gia_ObjFanin0(p).Value) ^ int32(p.fCompl0) }
func Gia_ObjFanin1Copy(p *Gia_Obj_t) int32 { return int32(Gia_ObjFanin1(p).Value) ^ int32(p.fCompl1) }
