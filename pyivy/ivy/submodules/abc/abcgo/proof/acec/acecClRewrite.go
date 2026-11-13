package acec

// Adder-output rewriting from src/proof/acec/acecCl.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Acec_RewriteTop(p *Gia_Man_t, box *Acec_Box_t) *Vec_Int_t {
	result := vec.Vec_IntAlloc(int(gia.Gia_ManCoNum(p)) + 1)
	driver := gia.Gia_ObjFaninId0p(p, gia.Gia_ManCo(p, 0))
	start := 0
	for ; start < vec.Vec_WecSize(box.vRootLits); start++ {
		level := vec.Vec_WecEntry(box.vRootLits, start)
		if vec.Vec_IntEntry(level, 0)>>1 == driver {
			break
		}
	}
	if start >= int(gia.Gia_ManCoNum(p)) {
		panic("acec: missing first output in arithmetic box")
	}
	count := 0
	for i := start; i < vec.Vec_WecSize(box.vRootLits); i++ {
		level := vec.Vec_WecEntry(box.vRootLits, i)
		n := vec.Vec_IntSize(level)
		if n == 0 || n > 3 {
			panic("acec: invalid adder level")
		}
		if n == 1 {
			vec.Vec_IntPush(result, vec.Vec_IntEntry(level, 0))
			continue
		}
		var in [3]int32
		copy(in[:], vec.Vec_IntArray(level))
		var out [2]int32
		Acec_InsertFadd(p, in[:], out[:])
		vec.Vec_IntPush(result, out[0])
		if i+1 < vec.Vec_WecSize(box.vRootLits) {
			vec.Vec_IntPush(vec.Vec_WecEntry(box.vRootLits, i+1), out[1])
		} else {
			vec.Vec_IntPush(vec.Vec_WecPushLevel(box.vRootLits), out[1])
		}
		count++
	}
	if vec.Vec_IntSize(result) < int(gia.Gia_ManCoNum(p)) {
		panic("acec: insufficient rewritten outputs")
	}
	vec.Vec_IntShrink(result, int(gia.Gia_ManCoNum(p)))
	fmt.Printf("Added %d adders for replace CLAs.  ", count)
	return result
}
func Acec_RewriteReplace(p *Gia_Man_t, result *Vec_Int_t) *Gia_Man_t {
	if int(gia.Gia_ManCoNum(p)) != vec.Vec_IntSize(result) {
		panic("acec: rewritten output count mismatch")
	}
	dst := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	dst.CopyIdentityFrom(p)
	gia.Gia_ManFillValue(p)
	gia.Gia_ManConst0(p).Value = 0
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		gia.Gia_ManCi(p, i).Value = uint32(gia.Gia_ManAppendCi(dst))
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) != 0 {
			o.Value = uint32(acecCopyLogic(dst, p, o))
		}
	}
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		lit := vec.Vec_IntEntry(result, int(i))
		repr := gia.Gia_ManObj(p, lit>>1)
		gia.Gia_ManCo(p, i).Value = uint32(gia.Gia_ManAppendCo(dst, int32(repr.Value)))
	}
	gia.Gia_ManSetPhase(p)
	gia.Gia_ManSetPhase(dst)
	for i := int32(0); i < gia.Gia_ManCoNum(dst); i++ {
		o := gia.Gia_ManCo(dst, i)
		if gia.Gia_ObjPhase(o) != gia.Gia_ObjPhase(gia.Gia_ManCo(p, i)) {
			gia.Gia_ObjFlipFaninC0(o)
		}
	}
	return gia.Gia_ManCleanup(dst)
}
func Acec_ManDecla(p *Gia_Man_t, booth, verbose int32) *Gia_Man_t {
	start := time.Now()
	var ignore *Vec_Bit_t
	if booth != 0 {
		ignore = Acec_BoothFindPPG(p)
	}
	box := Acec_DeriveBox(p, ignore, 0, 0, verbose)
	if box == nil {
		fmt.Println("Cannot find arithmetic boxes.")
		return gia.Gia_ManDup(p)
	}
	result := Acec_RewriteTop(p, box)
	dst := Acec_RewriteReplace(p, result)
	fmt.Printf("Time = %v\n", time.Since(start))
	return dst
}
