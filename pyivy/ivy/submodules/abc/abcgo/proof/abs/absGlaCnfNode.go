package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Ga2_ManAddToAbsOneStatic(p *Ga2_Man_t, obj *gia.Gia_Obj_t, frame, useID int32) {
	out := Ga2_ObjFindOrAddLit(p, obj, frame)
	if out <= 1 {
		panic("abs: invalid output literal")
	}
	isConst := gia.Gia_ObjIsConst0(obj) != 0
	isRo := gia.Gia_ObjIsRo(p.pGia, obj) != 0
	if !isConst && !isRo && gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: invalid abstracted object")
	}
	proofID := int32(-1)
	if useID != 0 {
		proofID = gia.Gia_ObjId(p.pGia, obj)
	}
	if isConst || frame == 0 && isRo {
		p.pSat.AddClause(proofID, bsat.Literal(out^1))
		return
	}
	vec.Vec_IntClear(p.vLits)
	leaves := Ga2_ObjLeaves(p.pGia, obj)
	sourceFrame := frame
	if isRo {
		sourceFrame--
	}
	static := true
	for i := 0; i < vec.Vec_IntSize(leaves); i++ {
		leaf := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(leaves, i))
		lit := Ga2_ObjFindOrAddLit(p, leaf, sourceFrame)
		vec.Vec_IntPush(p.vLits, lit)
		if lit < 2 {
			static = false
		}
	}
	if static || isRo {
		Ga2_ManCnfAddStatic(p.pSat, Ga2_ObjCnf0(p, obj), Ga2_ObjCnf1(p, obj), vec.Vec_IntArray(p.vLits), out, proofID)
	} else {
		truth := Ga2_ObjComputeTruthSpecial(p.pGia, obj, leaves, p.vLits)
		Ga2_ManCnfAddDynamic(p, int32(truth&0xffff), vec.Vec_IntArray(p.vLits), out, gia.Gia_ObjId(p.pGia, obj))
	}
}

// Ga2_ManAddToAbsOneDynamic simplifies an original abstraction cut before
// allocating SAT variables and emitting its clauses.
func Ga2_ManAddToAbsOneDynamic(p *Ga2_Man_t, obj *gia.Gia_Obj_t, frame int32) {
	if !Ga2_ObjIsAbs0(p, obj) {
		panic("abs: node is outside original abstraction")
	}
	isConst := gia.Gia_ObjIsConst0(obj) != 0
	isRo := gia.Gia_ObjIsRo(p.pGia, obj) != 0
	if !isConst && !isRo && gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: invalid abstracted object")
	}
	if isConst || frame == 0 && isRo {
		Ga2_ObjAddLit(p, obj, frame, 0)
		return
	}
	if isRo {
		ri := gia.Gia_ObjRoToRi(p.pGia, obj)
		lit := Ga2_ObjFindOrAddLit(p, gia.Gia_ObjFanin0(ri), frame-1) ^ gia.Gia_ObjFaninC0(ri)
		Ga2_ObjAddLit(p, obj, frame, lit)
		return
	}
	vec.Vec_IntClear(p.vLits)
	leaves := Ga2_ObjLeaves(p.pGia, obj)
	for i := 0; i < vec.Vec_IntSize(leaves); i++ {
		leaf := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(leaves, i))
		var lit int32
		if Ga2_ObjIsAbs0(p, leaf) {
			lit = Ga2_ObjFindLit(p, leaf, frame)
			if lit < 0 {
				panic("abs: missing abstract cut literal")
			}
		} else if Ga2_ObjIsLeaf0(p, leaf) {
			lit = Ga2_ObjFindLit(p, leaf, frame)
			if lit == -1 {
				lit = GA2_BIG_NUM + 2*int32(i)
			}
		} else {
			panic("abs: invalid original cut leaf")
		}
		vec.Vec_IntPush(p.vLits, lit)
	}
	truth := Ga2_ObjComputeTruthSpecial(p.pGia, obj, leaves, p.vLits)
	if truth == 0 || truth == ^uint32(0) {
		lit := int32(0)
		if truth != 0 {
			lit = 1
		}
		Ga2_ObjAddLit(p, obj, frame, lit)
		return
	}
	if truth == 0xaaaaaaaa || truth == 0x55555555 {
		lit := vec.Vec_IntEntry(p.vLits, 0)
		if lit >= GA2_BIG_NUM {
			leaf := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(leaves, int((lit-GA2_BIG_NUM)/2)))
			if Ga2_ObjFindLit(p, leaf, frame) != -1 {
				panic("abs: expected unassigned cut literal")
			}
			lit = Ga2_ObjFindOrAddLit(p, leaf, frame)
		}
		if truth == 0x55555555 {
			lit ^= 1
		}
		Ga2_ObjAddLit(p, obj, frame, lit)
		return
	}
	size := vec.Vec_IntSize(p.vLits)
	if size <= 1 || size >= 6 {
		panic("abs: invalid minimized cut size")
	}
	for i := 0; i < size; i++ {
		lit := vec.Vec_IntEntry(p.vLits, i)
		if lit >= GA2_BIG_NUM {
			leaf := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(leaves, int((lit-GA2_BIG_NUM)/2)))
			if Ga2_ObjFindLit(p, leaf, frame) != -1 {
				panic("abs: expected unassigned cut literal")
			}
			lit = Ga2_ObjFindOrAddLit(p, leaf, frame)
			vec.Vec_IntWriteEntry(p.vLits, i, lit)
		}
		if lit >= 10000000 {
			panic("abs: SAT literal too large")
		}
	}
	if size == 5 {
		vec.Vec_IntClear(p.vLits)
		for i := 0; i < vec.Vec_IntSize(leaves); i++ {
			leaf := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(leaves, i))
			vec.Vec_IntPush(p.vLits, Ga2_ObjFindOrAddLit(p, leaf, frame))
		}
		lit := Ga2_ObjFindOrAddLit(p, obj, frame)
		Ga2_ManCnfAddStatic(p.pSat, Ga2_ObjCnf0(p, obj), Ga2_ObjCnf1(p, obj), vec.Vec_IntArray(p.vLits), lit, -1)
		return
	}
	if p.pPars.FSkipHash == 0 {
		for vec.Vec_IntSize(p.vLits) < 4 {
			vec.Vec_IntPush(p.vLits, GA2_BIG_NUM)
		}
		vec.Vec_IntPush(p.vLits, int32(truth))
		slot := Saig_ManBmcLookup(p, vec.Vec_IntArray(p.vLits))
		if *slot == 0 {
			*slot = Ga2_ObjFindOrAddLit(p, obj, frame)
			vec.Vec_IntShrink(p.vLits, size)
			Ga2_ManCnfAddDynamic(p, int32(truth&0xffff), vec.Vec_IntArray(p.vLits), *slot, -1)
		} else {
			Ga2_ObjAddLit(p, obj, frame, *slot)
		}
	} else {
		lit := Ga2_ObjFindOrAddLit(p, obj, frame)
		Ga2_ManCnfAddDynamic(p, int32(truth&0xffff), vec.Vec_IntArray(p.vLits), lit, -1)
	}
}

// Ga2_ManAddAbsClauses encodes the original abstraction followed by additions.
func Ga2_ManAddAbsClauses(p *Ga2_Man_t, frame int32) {
	for i := 0; i < vec.Vec_IntSize(p.vValues) && i < int(p.LimAbs); i++ {
		obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(p.vValues, i))
		Ga2_ManAddToAbsOneDynamic(p, obj, frame)
	}
	for i := int(p.LimAbs); i < vec.Vec_IntSize(p.vAbs); i++ {
		obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(p.vAbs, i))
		Ga2_ManAddToAbsOneStatic(p, obj, frame, 1)
	}
}
