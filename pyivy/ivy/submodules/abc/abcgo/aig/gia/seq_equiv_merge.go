package gia

import "fmt"

// Gia_ManSeqEquivMerge combines equivalence classes whose part-object Value
// fields point back to object IDs in the full graph.
func Gia_ManSeqEquivMerge(p *Gia_Man_t, parts []*Gia_Man_t) {
	if len(parts) != 2 || p.PReprs != nil || p.PNexts != nil {
		panic("gia: invalid sequential equivalence merge")
	}
	p.StartRepresentatives()
	for _, part := range parts {
		if part == nil || len(part.PReprs) != int(part.nObjs) || len(part.PNexts) != int(part.nObjs) {
			panic("gia: invalid equivalence partition")
		}
		for id := int32(1); id < part.nObjs; id++ {
			repr := Gia_ObjRepr(part, id)
			if repr == GIA_VOID {
				continue
			}
			if repr < 0 || repr >= id {
				panic("gia: noncanonical partition representative")
			}
			left, right := part.pObjs[id].Value, part.pObjs[repr].Value
			if left >= uint32(p.nObjs) || right >= uint32(p.nObjs) {
				panic("gia: invalid partition object map")
			}
			giaMergeRepr(p, int32(left), int32(right))
		}
	}
	giaFlattenReprs(p)
	p.PNexts = Gia_ManDeriveNexts(p)
}
func Gia_ManPrintEquivs(p *Gia_Man_t) {
	if len(p.PReprs) != int(p.nObjs) || len(p.PNexts) != int(p.nObjs) {
		panic("gia: missing equivalence classes")
	}
	fmt.Print("Const0:")
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && Gia_ObjRepr(p, obj.id) == 0 {
			fmt.Printf(" %d", obj.id)
		}
	}
	fmt.Println()
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsHead(p, id) == 0 {
			continue
		}
		fmt.Printf("%d:", id)
		for next := Gia_ObjNext(p, id); next > 0; next = Gia_ObjNext(p, next) {
			fmt.Printf(" %d", next)
		}
		fmt.Println()
	}
}
