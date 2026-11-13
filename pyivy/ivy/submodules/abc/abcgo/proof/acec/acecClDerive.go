package acec

// Mirror substitution and literal polarity from src/proof/acec/acecCl.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Acec_ManDerive_rec(dst, p *Gia_Man_t, node int32, mirrors *Vec_Int_t) {
	objID := node
	mirror := vec.Vec_IntEntry(mirrors, int(node))
	if mirror >= 0 {
		objID = mirror >> 1
		if objID != node {
			// A replacement cone may itself reference the original node.
			vec.Vec_IntWriteEntry(mirrors, int(node), -1)
			defer vec.Vec_IntWriteEntry(mirrors, int(node), mirror)
		}
	}
	obj := gia.Gia_ManObj(p, objID)
	if obj.Value == ^uint32(0) {
		if gia.Gia_ObjIsAnd(obj) == 0 {
			panic("acec: expected AND in mirrored cone")
		}
		Acec_ManDerive_rec(dst, p, gia.Gia_ObjFaninId0(obj, objID), mirrors)
		Acec_ManDerive_rec(dst, p, gia.Gia_ObjFaninId1(obj, objID), mirrors)
		if gia.Gia_ObjIsXor(obj) != 0 {
			obj.Value = uint32(gia.Gia_ManAppendXorReal(dst, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj)))
		} else {
			obj.Value = uint32(gia.Gia_ManAppendAnd(dst, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj)))
		}
	}
	if objID != node {
		gia.Gia_ManObj(p, node).Value = obj.Value ^ uint32(mirror&1)
	}
}
func Acec_ManDerive(p *Gia_Man_t, mirrors *Vec_Int_t) *Gia_Man_t {
	if gia.Gia_ManHasMuxData(p) {
		panic("acec: mirror derivation requires no mux data")
	}
	gia.Gia_ManFillValue(p)
	dst := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	dst.CopyIdentityFrom(p)
	gia.Gia_ManConst0(p).Value = 0
	gia.Gia_ManHashAlloc(dst)
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		gia.Gia_ManCi(p, i).Value = uint32(gia.Gia_ManAppendCi(dst))
	}
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		o := gia.Gia_ManCo(p, i)
		Acec_ManDerive_rec(dst, p, gia.Gia_ObjFaninId0p(p, o), mirrors)
	}
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		o := gia.Gia_ManCo(p, i)
		o.Value = uint32(gia.Gia_ManAppendCo(dst, gia.Gia_ObjFanin0Copy(o)))
	}
	gia.Gia_ManHashStop(dst)
	gia.Gia_ManSetRegNum(dst, gia.Gia_ManRegNum(p))
	return dst
}
func Acec_DetectLitPolarity(p *Gia_Man_t, node, leaf int32) int32 {
	if node < leaf {
		return -1
	}
	if node == leaf {
		return node << 1
	}
	obj := gia.Gia_ManObj(p, node)
	if gia.Gia_ObjIsAnd(obj) == 0 {
		return -1
	}
	a := Acec_DetectLitPolarity(p, gia.Gia_ObjFaninId0(obj, node), leaf)
	b := Acec_DetectLitPolarity(p, gia.Gia_ObjFaninId1(obj, node), leaf)
	if a != -1 {
		a ^= gia.Gia_ObjFaninC0(obj)
	}
	if b != -1 {
		b ^= gia.Gia_ObjFaninC1(obj)
	}
	if a == -1 {
		return b
	}
	if b == -1 {
		return a
	}
	if a != b {
		panic("acec: inconsistent leaf polarity")
	}
	fmt.Printf("Problem for leaf %d\n", leaf)
	return a
}
