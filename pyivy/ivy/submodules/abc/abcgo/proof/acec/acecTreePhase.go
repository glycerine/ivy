package acec

// Adder-tree polarity propagation and verification from acecTree.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Acec_TreeVerifyPhaseOne_rec(p *Gia_Man_t, o *Gia_Obj_t) int32 {
	if gia.Gia_ObjIsTravIdCurrent(p, o) != 0 {
		return int32(o.Value)
	}
	gia.Gia_ObjSetTravIdCurrent(p, o)
	if gia.Gia_ObjIsAnd(o) == 0 || gia.Gia_ObjIsXor(o) != 0 {
		panic("acec: expected ordinary AND")
	}
	a, b := Acec_TreeVerifyPhaseOne_rec(p, gia.Gia_ObjFanin0(o)), Acec_TreeVerifyPhaseOne_rec(p, gia.Gia_ObjFanin1(o))
	if gia.Gia_ObjFaninC0(o) != 0 {
		a = 0xff & ^a
	}
	if gia.Gia_ObjFaninC1(o) != 0 {
		b = 0xff & ^b
	}
	o.Value = uint32(a & b)
	return int32(o.Value)
}
func Acec_TreeVerifyPhaseOne(p *Gia_Man_t, adds *Vec_Int_t, box int32) {
	off := int(6 * box)
	full := vec.Vec_IntEntry(adds, off+2) > 0
	flip := int32(0)
	if !full {
		flip = Acec_SignBit2(adds, box, 2)
	}
	truths := [3]uint32{0xaa, 0xcc, 0xf0}
	gia.Gia_ManIncrementTravId(p)
	for k := 0; k < 3; k++ {
		id := vec.Vec_IntEntry(adds, off+k)
		if id == 0 {
			continue
		}
		o := gia.Gia_ManObj(p, id)
		o.Value = truths[k]
		if Acec_SignBit2(adds, box, int32(k))^flip != 0 {
			o.Value = 0xff & ^o.Value
		}
		gia.Gia_ObjSetTravIdCurrent(p, o)
	}
	values := [2]int32{}
	for k := 0; k < 2; k++ {
		id := vec.Vec_IntEntry(adds, off+3+k)
		values[k] = Acec_TreeVerifyPhaseOne_rec(p, gia.Gia_ManObj(p, id))
		if Acec_SignBit2(adds, box, int32(3+k))^flip != 0 {
			values[k] = 0xff & ^values[k]
		}
	}
	kind := "Hadd"
	want := [2]int32{0x66, 0x88}
	if full {
		kind = "Fadd"
		want = [2]int32{0x96, 0xe8}
	}
	for k, label := range []string{"sum", "carry"} {
		if values[k] != want[k] {
			fmt.Printf("%s %d %s %d is wrong.\n", kind, box, label, vec.Vec_IntEntry(adds, off+3+k))
		}
	}
}
func Acec_TreeVerifyPhases(p *Gia_Man_t, adds *Vec_Int_t, boxes *Vec_Wec_t) {
	for i := 0; i < vec.Vec_WecSize(boxes); i++ {
		for _, box := range vec.Vec_IntArray(vec.Vec_WecEntry(boxes, i)) {
			Acec_TreeVerifyPhaseOne(p, adds, box)
		}
	}
}
func Acec_TreeVerifyPhases2(p *Gia_Man_t, adds *Vec_Int_t, boxes *Vec_Wec_t) {
	phase, roots := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p))), vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; i < vec.Vec_WecSize(boxes); i++ {
		for _, box := range vec.Vec_IntArray(vec.Vec_WecEntry(boxes, i)) {
			for k := int32(3); k < 5; k++ {
				id := vec.Vec_IntEntry(adds, int(6*box+k))
				vec.Vec_BitWriteEntry(roots, int(id), 1)
				vec.Vec_BitWriteEntry(phase, int(id), int(Acec_SignBit2(adds, box, k)))
			}
		}
	}
	for i := 0; i < vec.Vec_WecSize(boxes); i++ {
		for _, box := range vec.Vec_IntArray(vec.Vec_WecEntry(boxes, i)) {
			for k := int32(0); k < 3; k++ {
				id := vec.Vec_IntEntry(adds, int(6*box+k))
				if vec.Vec_BitEntry(roots, int(id)) != 0 && int32(vec.Vec_BitEntry(phase, int(id))) != Acec_SignBit2(adds, box, k) {
					fmt.Printf("Phase of input %d=%d is mismatched in box %d=(%d,%d).\n", k, id, box, vec.Vec_IntEntry(adds, int(6*box+3)), vec.Vec_IntEntry(adds, int(6*box+4)))
				}
			}
		}
	}
}
func Acec_TreeVerifyConnections(p *Gia_Man_t, adds *Vec_Int_t, boxes *Vec_Wec_t) {
	counts := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	for i := 0; i < vec.Vec_WecSize(boxes); i++ {
		for _, box := range vec.Vec_IntArray(vec.Vec_WecEntry(boxes, i)) {
			for k := 3; k < 5; k++ {
				vec.Vec_IntWriteEntry(counts, int(vec.Vec_IntEntry(adds, int(6*box)+k)), 0)
			}
		}
	}
	for i := 0; i < vec.Vec_WecSize(boxes); i++ {
		for _, box := range vec.Vec_IntArray(vec.Vec_WecEntry(boxes, i)) {
			for k := 0; k < 3; k++ {
				id := vec.Vec_IntEntry(adds, int(6*box)+k)
				if vec.Vec_IntEntry(counts, int(id)) != -1 {
					vec.Vec_IntAddToEntry(counts, int(id), 1)
				}
			}
		}
	}
	fmt.Printf("The adder tree has %d internal cut points. ", vec.Vec_IntCountLarger(counts, -1))
	n := vec.Vec_IntCountLarger(counts, 1)
	if n == 0 {
		fmt.Println("There is no internal fanouts.")
		return
	}
	fmt.Printf("These %d points have more than one fanout:\n", n)
	for id, fanout := range vec.Vec_IntArray(counts) {
		if fanout > 1 {
			fmt.Printf("Node %d(lev %d) has fanout %d.\n", id, gia.Gia_ObjLevelId(p, int32(id)), fanout)
		}
	}
}
func Acec_TreeCarryMap(p *Gia_Man_t, adds *Vec_Int_t, boxes *Vec_Wec_t) *Vec_Int_t {
	m := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	for i := 0; i < vec.Vec_WecSize(boxes); i++ {
		for _, box := range vec.Vec_IntArray(vec.Vec_WecEntry(boxes, i)) {
			vec.Vec_IntWriteEntry(m, int(vec.Vec_IntEntry(adds, int(6*box+4))), box)
		}
	}
	return m
}
func Acec_TreePhases_rec(p *Gia_Man_t, adds, m *Vec_Int_t, node, phase int32, visited *Vec_Bit_t) {
	if node == 0 {
		panic("acec: zero carry node")
	}
	box := vec.Vec_IntEntry(m, int(node))
	if box == -1 {
		return
	}
	if node != vec.Vec_IntEntry(adds, int(6*box+4)) {
		panic("acec: carry map mismatch")
	}
	if vec.Vec_BitEntry(visited, int(box)) != 0 {
		return
	}
	vec.Vec_BitWriteEntry(visited, int(box), 1)
	xphase := Acec_SignBit(adds, box, 3)
	if vec.Vec_IntEntry(adds, int(6*box+2)) == 0 {
		phase ^= Acec_SignBit(adds, box, 2)
		xphase ^= phase
		Acec_SignSetBit2(adds, box, 2, phase)
	}
	for k := int32(0); k < 3; k++ {
		id := vec.Vec_IntEntry(adds, int(6*box+k))
		if id == 0 {
			continue
		}
		v := Acec_SignBit(adds, box, k) ^ phase
		xphase ^= v
		Acec_TreePhases_rec(p, adds, m, id, v, visited)
		Acec_SignSetBit2(adds, box, k, v)
	}
	Acec_SignSetBit2(adds, box, 3, xphase)
	Acec_SignSetBit2(adds, box, 4, phase)
}
