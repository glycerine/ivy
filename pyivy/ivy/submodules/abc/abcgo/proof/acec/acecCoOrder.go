package acec

// Adder ordering from src/proof/acec/acecCo.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_PolynCoreNonXors_rec(p *Gia_Man_t, o *Gia_Obj_t, pairs *Vec_Int_t) {
	var a, b *Gia_Obj_t
	if gia.Gia_ObjRecognizeExor(o, &a, &b) == 0 {
		return
	}
	a = gia.Gia_Regular(a)
	b = gia.Gia_Regular(b)
	Gia_PolynCoreNonXors_rec(p, a, pairs)
	Gia_PolynCoreNonXors_rec(p, b, pairs)
	vec.Vec_IntPushTwo(pairs, gia.Gia_ObjId(p, a), gia.Gia_ObjId(p, b))
}
func Gia_PolynAddHaRoots(p *Gia_Man_t) *Vec_Int_t {
	out, pairs := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	Gia_PolynCoreNonXors_rec(p, gia.Gia_ObjFanin0(gia.Gia_ManCo(p, gia.Gia_ManCoNum(p)-1)), pairs)
	gia.Gia_ManSetPhase(p)
	for i := 0; i < vec.Vec_IntSize(pairs); i += 2 {
		a, b := vec.Vec_IntEntry(pairs, i), vec.Vec_IntEntry(pairs, i+1)
		a = a<<1 | gia.Gia_ObjPhase(gia.Gia_ManObj(p, a))
		b = b<<1 | gia.Gia_ObjPhase(gia.Gia_ManObj(p, b))
		vec.Vec_IntPush(out, gia.Gia_ManAppendAnd(p, a, b)>>1)
	}
	vec.Vec_IntReverseOrder(out)
	return out
}
func Gia_PolynComputeMap(adds *Vec_Int_t, n int32) *Vec_Wec_t {
	m := vec.Vec_WecStart(int(n))
	for i := 0; 6*i < vec.Vec_IntSize(adds); i++ {
		a, b := vec.Vec_IntEntry(adds, 6*i+3), vec.Vec_IntEntry(adds, 6*i+4)
		vec.Vec_WecPush(m, int(a), int32(i))
		vec.Vec_WecPush(m, int(a), b)
		vec.Vec_WecPush(m, int(b), int32(i))
		vec.Vec_WecPush(m, int(b), a)
	}
	return m
}
func Gia_PolynCoreOrder_int(p *Gia_Man_t, adds *Vec_Int_t, m *Vec_Wec_t, roots *Vec_Int_t, ins **Vec_Int_t) *Vec_Int_t {
	order := vec.Vec_IntAlloc(1000)
	marked := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for _, id := range vec.Vec_IntArray(roots) {
		vec.Vec_BitWriteEntry(marked, int(id), 1)
	}
	for {
		found := false
		for i := 0; i < vec.Vec_IntSize(roots); i++ {
			a := vec.Vec_IntEntry(roots, i)
			level := vec.Vec_WecEntry(m, int(a))
			for j := 0; j < vec.Vec_IntSize(level); j += 2 {
				box, b := vec.Vec_IntEntry(level, j), vec.Vec_IntEntry(level, j+1)
				if vec.Vec_BitEntry(marked, int(b)) == 0 {
					continue
				}
				vec.Vec_IntPush(order, box)
				vec.Vec_BitWriteEntry(marked, int(a), 0)
				vec.Vec_BitWriteEntry(marked, int(b), 0)
				vec.Vec_IntRemove(roots, a)
				vec.Vec_IntRemove(roots, b)
				for k := 0; k < 3; k++ {
					id := vec.Vec_IntEntry(adds, int(6*box)+k)
					vec.Vec_BitWriteEntry(marked, int(id), 1)
					vec.Vec_IntPushUnique(roots, id)
				}
				found = true
				break
			}
			if found {
				break
			}
		}
		if !found {
			break
		}
	}
	if ins != nil {
		*ins = vec.Vec_IntAlloc(vec.Vec_BitSize(marked))
		for i := 1; i < vec.Vec_BitSize(marked); i++ {
			if vec.Vec_BitEntry(marked, i) != 0 {
				vec.Vec_IntPush(*ins, int32(i))
			}
		}
	}
	return order
}
func Gia_PolynCoreOrder(p *Gia_Man_t, adds, extra *Vec_Int_t, ins, outs **Vec_Int_t) *Vec_Int_t {
	m := Gia_PolynComputeMap(adds, gia.Gia_ManObjNum(p))
	roots := vec.Vec_IntAlloc(int(gia.Gia_ManCoNum(p)))
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		vec.Vec_IntPush(roots, gia.Gia_ObjFaninId0p(p, gia.Gia_ManCo(p, i)))
	}
	if extra != nil {
		vec.Vec_IntAppend(roots, extra)
	}
	if outs != nil {
		*outs = vec.Vec_IntDup(roots)
	}
	order := Gia_PolynCoreOrder_int(p, adds, m, roots, ins)
	fmt.Printf("Collected %d boxes.\n", vec.Vec_IntSize(order))
	return order
}
func Gia_PolyCollectRoots_rec(adds *Vec_Int_t, m *Vec_Wec_t, marks *Vec_Bit_t, box int32, roots *Vec_Int_t) {
	for k := 0; k < 3; k++ {
		carry := vec.Vec_IntEntry(adds, int(6*box)+k)
		if carry == 0 {
			continue
		}
		level := vec.Vec_WecEntry(m, int(carry))
		for i := 0; i < vec.Vec_IntSize(level); i += 2 {
			index, sum := vec.Vec_IntEntry(level, i), vec.Vec_IntEntry(level, i+1)
			if vec.Vec_IntEntry(adds, int(6*index)+4) == carry && vec.Vec_BitEntry(marks, int(sum)) == 0 {
				vec.Vec_IntPush(roots, sum)
				Gia_PolyCollectRoots_rec(adds, m, marks, index, roots)
			}
		}
	}
}
func Gia_PolyCollectRoots(adds *Vec_Int_t, m *Vec_Wec_t, marks *Vec_Bit_t, box int32, roots *Vec_Int_t) {
	vec.Vec_IntClear(roots)
	vec.Vec_IntPush(roots, vec.Vec_IntEntry(adds, int(6*box)+3))
	vec.Vec_IntPush(roots, vec.Vec_IntEntry(adds, int(6*box)+4))
	Gia_PolyCollectRoots_rec(adds, m, marks, box, roots)
}
