package acec

// Adder-tree construction ported from acecNorm.c.
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Acec_InsertHadd(p *gia.Gia_Man_t, in, out []int32) {
	out[1] = gia.Gia_ManAppendAnd2(p, in[0], in[1])
	bothZero := gia.Gia_ManAppendAnd2(p, in[0]^1, in[1]^1)
	out[0] = gia.Gia_ManAppendOr2(p, out[1], bothZero) ^ 1
}
func Acec_InsertFadd(p *gia.Gia_Man_t, in, out []int32) {
	var first, second [2]int32
	Acec_InsertHadd(p, in[:2], first[:])
	Acec_InsertHadd(p, []int32{first[0], in[2]}, second[:])
	out[0] = second[0]
	out[1] = gia.Gia_ManAppendOr2(p, first[1], second[1])
}
func Acec_InsertTree(p *gia.Gia_Man_t, leafMap *vec.Vec_Wec_t) *vec.Vec_Int_t {
	roots := vec.Vec_IntAlloc(vec.Vec_WecSize(leafMap) + 5)
	for i := 0; i < vec.Vec_WecSize(leafMap); i++ {
		level := vec.Vec_WecEntry(leafMap, i)
		if vec.Vec_IntSize(level) == 0 {
			vec.Vec_IntPush(roots, 0)
			continue
		}
		for vec.Vec_IntSize(level) > 1 {
			if vec.Vec_IntSize(level) == 2 {
				vec.Vec_IntPush(level, 0)
			}
			var in [3]int32
			for j := range in {
				in[j] = vec.Vec_IntEntry(level, 0)
				vec.Vec_IntDrop(level, 0)
			}
			var out [2]int32
			Acec_InsertFadd(p, in[:], out[:])
			vec.Vec_IntPush(level, out[0])
			if i+1 < vec.Vec_WecSize(leafMap) {
				level = vec.Vec_WecEntry(leafMap, i+1)
			} else {
				level = vec.Vec_WecPushLevel(leafMap)
			}
			vec.Vec_IntPush(level, out[1])
			level = vec.Vec_WecEntry(leafMap, i)
		}
		vec.Vec_IntPush(roots, vec.Vec_IntEntry(level, 0))
	}
	return roots
}

func Acec_InsertBox_rec(dst, src *gia.Gia_Man_t, obj *gia.Gia_Obj_t) int32 {
	if obj.Value != ^uint32(0) {
		return int32(obj.Value)
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("acec: unbound non-AND box leaf")
	}
	Acec_InsertBox_rec(dst, src, gia.Gia_ObjFanin0(obj))
	Acec_InsertBox_rec(dst, src, gia.Gia_ObjFanin1(obj))
	if gia.Gia_ObjIsMux(src, obj) != 0 {
		Acec_InsertBox_rec(dst, src, gia.Gia_ObjFanin2(src, obj))
	}
	lit := acecCopyLogic(dst, src, obj)
	obj.Value = uint32(lit)
	return lit
}
func Acec_BuildTree(dst, src *gia.Gia_Man_t, leafLits *vec.Vec_Wec_t, rootLits *vec.Vec_Int_t) *vec.Vec_Int_t {
	leafMap := vec.Vec_WecStart(vec.Vec_WecSize(leafLits))
	if rootLits != nil {
		for i, lit := range vec.Vec_IntArray(rootLits) {
			if i >= vec.Vec_WecSize(leafMap) {
				vec.Vec_WecPushLevel(leafMap)
			}
			vec.Vec_IntPush(vec.Vec_WecEntry(leafMap, i), lit)
		}
	}
	for rank := 0; rank < vec.Vec_WecSize(leafLits); rank++ {
		for _, lit := range vec.Vec_IntArray(vec.Vec_WecEntry(leafLits, rank)) {
			obj := gia.Gia_ManObj(src, lit>>1)
			copy := Acec_InsertBox_rec(dst, src, obj) ^ (lit & 1)
			vec.Vec_WecPush(leafMap, rank, copy)
		}
	}
	return Acec_InsertTree(dst, leafMap)
}
func Acec_InsertBox(box *Acec_Box_t, all int32) *gia.Gia_Man_t {
	src := box.pGia
	dst := gia.Gia_ManStart(gia.Gia_ManObjNum(src))
	dst.CopyIdentityFrom(src)
	gia.Gia_ManFillValue(src)
	gia.Gia_ManConst0(src).Value = 0
	for i := int32(0); i < gia.Gia_ManCiNum(src); i++ {
		gia.Gia_ManCi(src, i).Value = uint32(gia.Gia_ManAppendCi(dst))
	}
	var roots *vec.Vec_Int_t
	if all != 0 {
		roots = Acec_BuildTree(dst, src, box.vLeafLits, nil)
	} else {
		if box.vShared == nil || box.vUnique == nil {
			panic("acec: missing shared or unique box leaves")
		}
		roots = Acec_BuildTree(dst, src, box.vShared, nil)
		roots = Acec_BuildTree(dst, src, box.vUnique, roots)
	}
	for rank := 0; rank < vec.Vec_WecSize(box.vRootLits); rank++ {
		for k, lit := range vec.Vec_IntArray(vec.Vec_WecEntry(box.vRootLits, rank)) {
			copy := int32(0)
			if k == 0 {
				copy = vec.Vec_IntEntry(roots, rank)
			}
			gia.Gia_ManObj(src, lit>>1).Value = uint32(copy ^ (lit & 1))
		}
	}
	for i := int32(0); i < gia.Gia_ManCoNum(src); i++ {
		Acec_InsertBox_rec(dst, src, gia.Gia_ObjFanin0(gia.Gia_ManCo(src, i)))
	}
	for i := int32(0); i < gia.Gia_ManCoNum(src); i++ {
		obj := gia.Gia_ManCo(src, i)
		obj.Value = uint32(gia.Gia_ManAppendCo(dst, gia.Gia_ObjFanin0Copy(obj)))
	}
	gia.Gia_ManSetRegNum(dst, gia.Gia_ManRegNum(src))
	return dst
}
