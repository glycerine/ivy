package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ObjEdgeCount(id int32, first, second *Vec_Int_t) int32 {
	var count int32
	if vec.Vec_IntEntry(first, int(id)) > 0 {
		count++
	}
	if vec.Vec_IntEntry(second, int(id)) > 0 {
		count++
	}
	return count
}
func Gia_ObjEdgeAdd(id, next int32, first, second *Vec_Int_t) int32 {
	if vec.Vec_IntEntry(first, int(id)) == 0 {
		vec.Vec_IntWriteEntry(first, int(id), next)
		return 0
	}
	if vec.Vec_IntEntry(second, int(id)) == 0 {
		vec.Vec_IntWriteEntry(second, int(id), next)
		return 0
	}
	return 1
}
func Gia_ObjEdgeRemove(id, next int32, first, second *Vec_Int_t) {
	a, b := vec.Vec_IntEntry(first, int(id)), vec.Vec_IntEntry(second, int(id))
	if a != next && b != next {
		panic("gia: missing assigned edge")
	}
	if a == next {
		vec.Vec_IntWriteEntry(first, int(id), b)
	}
	vec.Vec_IntWriteEntry(second, int(id), 0)
}
func Gia_ObjEdgeClean(id int32, first, second *Vec_Int_t) {
	vec.Vec_IntWriteEntry(first, int(id), 0)
	vec.Vec_IntWriteEntry(second, int(id), 0)
}
func Gia_ManEdgeFromArray(p *Gia_Man_t, edges *Vec_Int_t) {
	data := vec.Vec_IntArray(edges)
	if len(data)%2 != 0 {
		panic("gia: incomplete edge pair")
	}
	p.VEdge1 = vec.Vec_IntStart(int(p.nObjs))
	p.VEdge2 = vec.Vec_IntStart(int(p.nObjs))
	var violations int32
	for i := 0; i < len(data); i += 2 {
		a, b := data[i], data[i+1]
		if a < 0 || a >= b || b >= p.nObjs {
			panic("gia: invalid edge endpoints")
		}
		violations += Gia_ObjEdgeAdd(a, b, p.VEdge1, p.VEdge2)
		violations += Gia_ObjEdgeAdd(b, a, p.VEdge1, p.VEdge2)
	}
	if violations != 0 {
		fmt.Printf("Found %d violations during edge conversion.\n", violations)
	}
}
func Gia_ManEdgeToArray(p *Gia_Man_t) *Vec_Int_t {
	if p.VEdge1 == nil || p.VEdge2 == nil || vec.Vec_IntSize(p.VEdge1) != int(p.nObjs) || vec.Vec_IntSize(p.VEdge2) != int(p.nObjs) {
		panic("gia: invalid assigned edge storage")
	}
	result := vec.Vec_IntAlloc(16)
	for id := int32(0); id < p.nObjs; id++ {
		for _, edges := range []*Vec_Int_t{p.VEdge1, p.VEdge2} {
			fanin := vec.Vec_IntEntry(edges, int(id))
			if fanin != 0 && fanin < id {
				vec.Vec_IntPush(result, fanin)
				vec.Vec_IntPush(result, id)
			}
		}
	}
	return result
}
func Gia_ManConvertPackingToEdges(p *Gia_Man_t) {
	if p.VPacking == nil {
		return
	}
	data := vec.Vec_IntArray(p.VPacking)
	if len(data) == 0 {
		panic("gia: empty packing")
	}
	p.VEdge1 = vec.Vec_IntStart(int(p.nObjs))
	p.VEdge2 = vec.Vec_IntStart(int(p.nObjs))
	count, violations := 0, int32(0)
	add := func(a, b int32) {
		if a < 0 || b < 0 || a >= p.nObjs || b >= p.nObjs {
			panic("gia: invalid packed node")
		}
		violations += Gia_ObjEdgeAdd(a, b, p.VEdge1, p.VEdge2)
		violations += Gia_ObjEdgeAdd(b, a, p.VEdge1, p.VEdge2)
	}
	for pos := 1; pos < len(data); {
		n := int(data[pos])
		pos++
		if n < 1 || n > 3 || pos+n > len(data) {
			panic("gia: malformed packing")
		}
		nodes := data[pos : pos+n]
		pos += n
		count++
		if n == 2 {
			add(nodes[0], nodes[1])
		} else if n == 3 {
			add(nodes[0], nodes[2])
			add(nodes[1], nodes[2])
		}
	}
	if count != int(data[0]) {
		panic("gia: packing group count mismatch")
	}
	if violations != 0 {
		fmt.Printf("Skipped %d illegal edges.\n", violations)
	}
}
func Gia_ObjHaveEdge(p *Gia_Man_t, id, next int32) int32 {
	return giaBool(vec.Vec_IntEntry(p.VEdge1, int(id)) == next || vec.Vec_IntEntry(p.VEdge2, int(id)) == next)
}
func Gia_ObjCheckEdge(p *Gia_Man_t, id, next int32) int32 { return Gia_ObjHaveEdge(p, id, next) }
