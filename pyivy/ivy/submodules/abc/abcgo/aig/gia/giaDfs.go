package gia

// DFS/support ordering APIs ported from giaDfs.c. Native iterative walks avoid
// call-stack depth limits and include real MUX controls in graph dependencies.
import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

type giaDfsFrame struct {
	id       int32
	expanded bool
}

func giaDfsCollect(p *Gia_Man_t, root int32, result *Vec_Int_t, inputs, nodes bool) int32 {
	stack := []giaDfsFrame{{id: root}}
	var count int32
	for len(stack) > 0 {
		frame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		obj := Gia_ManObj(p, frame.id)
		if frame.expanded {
			if nodes {
				if result != nil {
					Vec_IntPush(result, frame.id)
				}
				count++
			}
			continue
		}
		if Gia_ObjIsCo(obj) != 0 {
			stack = append(stack, giaDfsFrame{id: Gia_ObjFaninId0(obj, frame.id)})
			continue
		}
		if Gia_ObjIsTravIdCurrentId(p, frame.id) != 0 {
			continue
		}
		Gia_ObjSetTravIdCurrentId(p, frame.id)
		if frame.id == 0 {
			continue
		}
		if Gia_ObjIsCi(obj) != 0 {
			if inputs && result != nil {
				Vec_IntPush(result, frame.id)
			}
			continue
		}
		if Gia_ObjIsAnd(obj) == 0 {
			panic("gia: unsupported DFS object")
		}
		stack = append(stack, giaDfsFrame{id: frame.id, expanded: true})
		fanins := giaObjectFanins(p, obj)
		for i := len(fanins) - 1; i >= 0; i-- {
			stack = append(stack, giaDfsFrame{id: fanins[i]})
		}
	}
	return count
}
func Gia_ManCollectCis_rec(p *Gia_Man_t, obj *Gia_Obj_t, support *Vec_Int_t) {
	giaDfsCollect(p, Gia_ObjId(p, obj), support, true, false)
}
func Gia_ManCollectCis(p *Gia_Man_t, roots []int32, count int32, support *Vec_Int_t) {
	Vec_IntClear(support)
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrentId(p, 0)
	for _, root := range roots[:count] {
		Gia_ManCollectCis_rec(p, Gia_ManObj(p, root), support)
	}
}
func Gia_ManCollectAnds_rec(p *Gia_Man_t, id int32, nodes *Vec_Int_t) {
	giaDfsCollect(p, id, nodes, false, true)
}

// CollectAnds deliberately shares the caller's current traversal epoch.
func Gia_ManCollectAnds(p *Gia_Man_t, roots []int32, count int32, nodes, leaves *Vec_Int_t) {
	Gia_ObjSetTravIdCurrentId(p, 0)
	if leaves != nil {
		for _, leaf := range Vec_IntArray(leaves) {
			Gia_ObjSetTravIdCurrentId(p, leaf)
		}
	}
	Vec_IntClear(nodes)
	for _, root := range roots[:count] {
		obj := Gia_ManObj(p, root)
		if Gia_ObjIsCo(obj) != 0 || Gia_ObjIsAnd(obj) != 0 {
			Gia_ManCollectAnds_rec(p, root, nodes)
		}
	}
}
func Gia_ManCollectAndsAll(p *Gia_Man_t) *Vec_Int_t {
	result := Vec_IntAlloc(int(Gia_ManAndNum(p)))
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			Vec_IntPush(result, obj.id)
		}
	}
	return result
}
func Gia_ManCollectNodesCis_rec(p *Gia_Man_t, obj *Gia_Obj_t, nodes *Vec_Int_t) {
	giaDfsCollect(p, Gia_ObjId(p, obj), nodes, true, true)
}
func Gia_ManCollectNodesCis(p *Gia_Man_t, roots []int32, count int32) *Vec_Int_t {
	result := Vec_IntAlloc(10000)
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrentId(p, 0)
	for _, root := range roots[:count] {
		Gia_ManCollectNodesCis_rec(p, Gia_ManObj(p, root), result)
	}
	return result
}
func Gia_ManCollectTest(p *Gia_Man_t) {
	start := time.Now()
	result := Vec_IntAlloc(100)
	Gia_ManIncrementTravId(p)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		root := Gia_ObjId(p, Gia_ManCo(p, i))
		Gia_ManCollectAnds(p, []int32{root}, 1, result, nil)
	}
	fmt.Printf("DFS from each output = %9.2f sec\n", time.Since(start).Seconds())
}
func Gia_ManConeSize_rec(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return giaDfsCollect(p, Gia_ObjId(p, obj), nil, false, true)
}
func Gia_ManConeSize(p *Gia_Man_t, roots []int32, count int32) int32 {
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrentId(p, 0)
	var total int32
	for _, root := range roots[:count] {
		total += Gia_ManConeSize_rec(p, Gia_ManObj(p, root))
	}
	return total
}
func Gia_ManLevelize(p *Gia_Man_t) *Vec_Vec_t {
	levels := Gia_ManLevelNum(p)
	result := vec.Vec_VecStart(int(levels + 1))
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			vec.Vec_VecPush(result, int(Gia_ObjLevel(p, obj)), obj)
		}
	}
	return result
}
func Gia_ManLevelizeR(p *Gia_Man_t) *Vec_Wec_t {
	levels := Gia_ManLevelRNum(p)
	result := Vec_WecStart(int(levels + 1))
	for _, obj := range p.pObjs {
		level := Gia_ObjLevel(p, obj)
		if obj.id == 0 || Gia_ObjIsCo(obj) == 0 && level == 0 {
			continue
		}
		Vec_WecPush(result, int(level), obj.id)
	}
	return result
}
func Gia_ManOrderReverse(p *Gia_Man_t) *Vec_Int_t {
	buckets := make([][]int32, 1)
	add := func(id, level int32) {
		if level < 0 {
			panic("gia: negative level")
		}
		for len(buckets) <= int(level) {
			buckets = append(buckets, nil)
		}
		buckets[level] = append(buckets[level], id)
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && Gia_ObjLevel(p, obj) <= 0 {
			panic("gia: reverse ordering requires positive AND levels")
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		add(obj.id, Gia_ObjLevel(p, Gia_ObjFanin0(obj)))
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsCo(obj) == 0 {
			add(obj.id, Gia_ObjLevel(p, obj))
		}
	}
	result := Vec_IntAlloc(len(p.pObjs))
	for i := len(buckets) - 1; i >= 0; i-- {
		for _, id := range buckets[i] {
			Vec_IntPush(result, id)
		}
	}
	return result
}
func Gia_ManCollectSeq_rec(p *Gia_Man_t, id int32, roots, objects *Vec_Int_t) {
	// Explicit postorder stack preserves source RO-before-appended-RI scheduling.
	stack := []giaDfsFrame{{id: id}}
	for len(stack) > 0 {
		frame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if frame.expanded {
			Vec_IntPush(objects, frame.id)
			continue
		}
		if Gia_ObjIsTravIdCurrentId(p, frame.id) != 0 {
			continue
		}
		Gia_ObjSetTravIdCurrentId(p, frame.id)
		obj := Gia_ManObj(p, frame.id)
		if Gia_ObjIsCi(obj) != 0 {
			if Gia_ObjIsRo(p, obj) != 0 {
				Vec_IntPush(roots, Gia_ObjId(p, Gia_ObjRoToRi(p, obj)))
			}
			Vec_IntPush(objects, frame.id)
			continue
		}
		if Gia_ObjIsCo(obj) == 0 && Gia_ObjIsAnd(obj) == 0 {
			panic("gia: invalid sequential cone object")
		}
		stack = append(stack, giaDfsFrame{id: frame.id, expanded: true})
		fanins := giaObjectFanins(p, obj)
		for i := len(fanins) - 1; i >= 0; i-- {
			stack = append(stack, giaDfsFrame{id: fanins[i]})
		}
	}
}
func Gia_ManCollectSeq(p *Gia_Man_t, outputs []int32, count int32) *Vec_Int_t {
	roots := Vec_IntAlloc(100)
	for _, index := range outputs[:count] {
		Vec_IntPush(roots, Gia_ObjId(p, Gia_ManPo(p, index)))
	}
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrentId(p, 0)
	result := Vec_IntAlloc(1000)
	Vec_IntPush(result, 0)
	for i := 0; i < Vec_IntSize(roots); i++ {
		Gia_ManCollectSeq_rec(p, Vec_IntEntry(roots, i), roots, result)
	}
	return result
}
func Gia_ManCollectSeqTest(p *Gia_Man_t) {
	start := time.Now()
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		if i%10000 == 0 {
			fmt.Printf("%8d finished...\r", i)
		}
		Gia_ManCollectSeq(p, []int32{i}, 1)
	}
	fmt.Printf("Time = %.2f sec\n", time.Since(start).Seconds())
}
func Gia_ManCollectTfi_rec(p *Gia_Man_t, id int32, nodes *Vec_Int_t) {
	giaDfsCollect(p, id, nodes, false, true)
}
func Gia_ManCollectTfi(p *Gia_Man_t, roots, nodes *Vec_Int_t) {
	Vec_IntClear(nodes)
	Gia_ManIncrementTravId(p)
	for _, root := range Vec_IntArray(roots) {
		Gia_ManCollectTfi_rec(p, root, nodes)
	}
}
func Gia_ManCollectTfo_rec(p *Gia_Man_t, id int32, nodes *Vec_Int_t) {
	stack := []giaDfsFrame{{id: id}}
	for len(stack) > 0 {
		frame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if frame.expanded {
			Vec_IntPush(nodes, frame.id)
			continue
		}
		if Gia_ObjIsTravIdCurrentId(p, frame.id) != 0 {
			continue
		}
		Gia_ObjSetTravIdCurrentId(p, frame.id)
		obj := Gia_ManObj(p, frame.id)
		if Gia_ObjIsCo(obj) != 0 {
			continue
		}
		if Gia_ObjIsAnd(obj) == 0 {
			panic("gia: forward cone root must be an AND")
		}
		stack = append(stack, giaDfsFrame{id: frame.id, expanded: true})
		for i := Gia_ObjFanoutNumId(p, frame.id) - 1; i >= 0; i-- {
			stack = append(stack, giaDfsFrame{id: Gia_ObjFanoutId(p, frame.id, i)})
		}
	}
}
func Gia_ManCollectTfo(p *Gia_Man_t, roots, nodes *Vec_Int_t) {
	Vec_IntClear(nodes)
	Gia_ManIncrementTravId(p)
	for _, root := range Vec_IntArray(roots) {
		Gia_ManCollectTfo_rec(p, root, nodes)
	}
}
