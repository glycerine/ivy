package acec

// Adder-tree phase analysis, ported from src/proof/acec/acecPa.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Pas_ManVerifyPhaseOne_rec(p *Gia_Man_t, o *Gia_Obj_t) int32 {
	if gia.Gia_ObjIsTravIdCurrent(p, o) != 0 {
		return int32(o.Value)
	}
	gia.Gia_ObjSetTravIdCurrent(p, o)
	if gia.Gia_ObjIsAnd(o) == 0 || gia.Gia_ObjIsXor(o) != 0 {
		panic("acec: expected ordinary AND")
	}
	a := Pas_ManVerifyPhaseOne_rec(p, gia.Gia_ObjFanin0(o))
	b := Pas_ManVerifyPhaseOne_rec(p, gia.Gia_ObjFanin1(o))
	if gia.Gia_ObjFaninC0(o) != 0 {
		a = 0xff & ^a
	}
	if gia.Gia_ObjFaninC1(o) != 0 {
		b = 0xff & ^b
	}
	o.Value = uint32(a & b)
	return int32(o.Value)
}
func Pas_ManVerifyPhaseOne(p *Gia_Man_t, adds *Vec_Int_t, box int32, phase *Vec_Bit_t) {
	off := int(6 * box)
	if vec.Vec_IntEntry(adds, off+2) <= 0 {
		return
	}
	gia.Gia_ManIncrementTravId(p)
	truths := [3]uint32{0xaa, 0xcc, 0xf0}
	for i := 0; i < 3; i++ {
		id := vec.Vec_IntEntry(adds, off+i)
		if id == 0 {
			continue
		}
		o := gia.Gia_ManObj(p, id)
		o.Value = truths[i]
		if vec.Vec_BitEntry(phase, int(id)) != 0 {
			o.Value = 0xff & ^o.Value
		}
		gia.Gia_ObjSetTravIdCurrent(p, o)
	}
	vals := [2]int32{}
	for i := 0; i < 2; i++ {
		id := vec.Vec_IntEntry(adds, off+3+i)
		vals[i] = Pas_ManVerifyPhaseOne_rec(p, gia.Gia_ManObj(p, id))
		if vec.Vec_BitEntry(phase, int(id)) != 0 {
			vals[i] = 0xff & ^vals[i]
		}
	}
	if vals[0] != 0x96 {
		fmt.Printf("Fadd %d sum is wrong.\n", box)
	}
	if vals[1] != 0xe8 {
		fmt.Printf("Fadd %d carry is wrong.\n", box)
	}
}
func Pas_ManVerifyPhase(p *Gia_Man_t, adds, order *Vec_Int_t, phase *Vec_Bit_t) {
	for i := 0; i < vec.Vec_IntSize(order); i++ {
		Pas_ManVerifyPhaseOne(p, adds, vec.Vec_IntEntry(order, i), phase)
	}
}
func Pas_ManPhase_rec(p *Gia_Man_t, adds, mapping *Vec_Int_t, o *Gia_Obj_t, phase int32, phases, constants *Vec_Bit_t) {
	if o == gia.Gia_ManConst0(p) {
		panic("acec: constant phase node")
	}
	if gia.Gia_ObjIsTravIdCurrent(p, o) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(p, o)
	if phase != 0 {
		vec.Vec_BitWriteEntry(phases, int(gia.Gia_ObjId(p, o)), int(phase))
	}
	if gia.Gia_ObjIsAnd(o) == 0 {
		return
	}
	box := vec.Vec_IntEntry(mapping, int(gia.Gia_ObjId(p, o)))
	if box == -1 {
		return
	}
	off := int(6 * box)
	x := vec.Vec_IntEntry(adds, off+3)
	sign := vec.Vec_IntEntry(adds, off+5)
	xp := (sign >> 3) & 1
	if vec.Vec_IntEntry(adds, off+2) == 0 && phase != 0 {
		vec.Vec_BitWriteEntry(constants, int(box), 1)
	}
	for i := 0; i < 3; i++ {
		id := vec.Vec_IntEntry(adds, off+i)
		current := ((sign >> i) & 1) ^ phase
		xp ^= current
		if id != 0 {
			Pas_ManPhase_rec(p, adds, mapping, gia.Gia_ManObj(p, id), current, phases, constants)
		}
	}
	vec.Vec_BitWriteEntry(phases, int(x), int(xp))
}
func Pas_ManPhase(p *Gia_Man_t, adds, mapping, roots *Vec_Int_t, constants **Vec_Bit_t) *Vec_Bit_t {
	phases := vec.Vec_BitStart(vec.Vec_IntSize(mapping))
	c := vec.Vec_BitStart(vec.Vec_IntSize(adds) / 6)
	gia.Gia_ManIncrementTravId(p)
	for i := 0; i < vec.Vec_IntSize(roots); i++ {
		Pas_ManPhase_rec(p, adds, mapping, gia.Gia_ManObj(p, vec.Vec_IntEntry(roots, i)), 1, phases, c)
	}
	*constants = c
	return phases
}
func Pas_ManComputeCuts(p *Gia_Man_t, adds, order, ins, outs *Vec_Int_t) int32 {
	unique := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	used := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	mapping := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	roots := vec.Vec_IntAlloc(100)
	for i := 0; i < vec.Vec_IntSize(order); i++ {
		box := vec.Vec_IntEntry(order, i)
		off := int(6 * box)
		carry := vec.Vec_IntEntry(adds, off+4)
		if vec.Vec_BitEntry(unique, int(carry)) != 0 {
			fmt.Printf("Carry %d participates more than once.\n", carry)
		}
		vec.Vec_BitWriteEntry(unique, int(carry), 1)
		vec.Vec_IntWriteEntry(mapping, int(carry), box)
		for k := 0; k < 3; k++ {
			vec.Vec_BitWriteEntry(used, int(vec.Vec_IntEntry(adds, off+k)), 1)
		}
	}
	for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
		if vec.Vec_BitEntry(unique, int(i)) != 0 && vec.Vec_BitEntry(used, int(i)) == 0 {
			vec.Vec_IntPush(roots, i)
		}
	}
	vec.Vec_IntPrint(roots)
	if vec.Vec_IntSize(roots) > 0 {
		var constants *Vec_Bit_t
		phase := Pas_ManPhase(p, adds, mapping, roots, &constants)
		Pas_ManVerifyPhase(p, adds, order, phase)
		fmt.Print("Outputs: ")
		for i := 0; i < vec.Vec_IntSize(outs); i++ {
			id := vec.Vec_IntEntry(outs, i)
			fmt.Printf("%d(%d) ", id, vec.Vec_BitEntry(phase, int(id)))
		}
		fmt.Print("\nInputs: ")
		n := 0
		for i := 0; i < vec.Vec_IntSize(ins); i++ {
			id := vec.Vec_IntEntry(ins, i)
			v := vec.Vec_BitEntry(phase, int(id))
			fmt.Printf("%d(%d) ", id, v)
			n += v
		}
		fmt.Printf("   Compl = %d\n", n)
	}
	return int32(vec.Vec_IntSize(roots))
}
func Pas_ManComputeCutsTest(p *Gia_Man_t) {
	start := time.Now()
	adds := Ree_ManComputeCuts(p, nil, 1)
	var ins, outs *Vec_Int_t
	order := Gia_PolynCoreOrder(p, adds, nil, &ins, &outs)
	n := Ree_ManCountFadds(adds)
	fmt.Printf("Detected %d FAs and %d HAs.  Collected %d adders.  Time = %v\n", n, int32(vec.Vec_IntSize(adds)/6)-n, vec.Vec_IntSize(order), time.Since(start))
	start = time.Now()
	trees := Pas_ManComputeCuts(p, adds, order, ins, outs)
	fmt.Printf("Detected %d adder trees. Time = %v\n", trees, time.Since(start))
}
