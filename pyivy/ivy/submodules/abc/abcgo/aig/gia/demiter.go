package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

func Gia_ManSortByValue(a, b *Gia_Obj_t) int32 {
	x, y := Gia_Regular(a).Value, Gia_Regular(b).Value
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}
func Gia_ManDupDemiter(src *Gia_Man_t, verbose int32) *Gia_Man_t {
	if Gia_ManPoNum(src) != 1 {
		panic("gia: demiter expects one primary output")
	}
	po := Gia_ManPo(src, 0)
	super := vec.Vec_IntAlloc(16)
	Gia_ManDupWithConstrCollectAnd_rec(src, Gia_ObjFanin0(po), super, 1)
	if vec.Vec_IntSize(super) < 2 {
		panic("gia: miter is not decomposable")
	}
	mode := "AND"
	if Gia_ObjFaninC0(po) != 0 {
		mode = "OR"
	}
	fmt.Printf("The miter is %s-decomposable into %d parts.\n", mode, vec.Vec_IntSize(super))
	Gia_ManLevelNum(src)
	lits := append([]int32(nil), vec.Vec_IntArray(super)...)
	sort.SliceStable(lits, func(i, j int) bool { return Gia_ObjLevelId(src, lits[i]>>1) < Gia_ObjLevelId(src, lits[j]>>1) })
	out := giaQuantStart(src)
	giaComposeNodes(out, src, true)
	for _, lit := range lits {
		Gia_ManAppendCo(out, Gia_ObjLitCopy(src, lit)^Gia_ObjFaninC0(po))
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManRi(src, i))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	Gia_ManHashStop(out)
	return giaCleanComposed(out)
}
func Gia_ManDupDemiterOrderXors2(p *Gia_Man_t, xors *Vec_Int_t) {
	ids := vec.Vec_IntArray(xors)
	sizes := make(map[int32]int32, len(ids))
	for _, id := range ids {
		sizes[id] = Gia_ManSuppSize(p, []int32{id}, 1)
	}
	sort.SliceStable(ids, func(i, j int) bool { return sizes[ids[i]] < sizes[ids[j]] })
}
func Gia_ManDupDemiterFindMin(supports *Vec_Wec_t, takenInputs, takenOutputs *Vec_Int_t) int32 {
	best := int32(-1)
	minimum := int(^uint(0) >> 1)
	for i := 0; i < vec.Vec_WecSize(supports); i++ {
		if vec.Vec_IntEntry(takenOutputs, i) != 0 {
			continue
		}
		n := 0
		for _, id := range vec.Vec_IntArray(vec.Vec_WecEntry(supports, i)) {
			if vec.Vec_IntEntry(takenInputs, int(id)) == 0 {
				n++
			}
		}
		if n < minimum {
			minimum = n
			best = int32(i)
		}
	}
	return best
}
func Gia_ManDupDemiterOrderXors(p *Gia_Man_t, xors *Vec_Int_t) {
	supports := Gia_ManCreateNodeSupps(p, xors, 0)
	takenInputs := vec.Vec_IntStart(int(Gia_ManCiNum(p)))
	takenOutputs := vec.Vec_IntStart(vec.Vec_IntSize(xors))
	order := make([]int32, 0, vec.Vec_IntSize(xors))
	for range vec.Vec_IntSize(xors) {
		i := Gia_ManDupDemiterFindMin(supports, takenInputs, takenOutputs)
		if i < 0 {
			panic("gia: incomplete demiter order")
		}
		order = append(order, vec.Vec_IntEntry(xors, int(i)))
		vec.Vec_IntWriteEntry(takenOutputs, int(i), 1)
		for _, id := range vec.Vec_IntArray(vec.Vec_WecEntry(supports, int(i))) {
			vec.Vec_IntWriteEntry(takenInputs, int(id), 1)
		}
	}
	copy(vec.Vec_IntArray(xors), order)
}
func giaMarkCone(p *Gia_Man_t, id int32, mark1 bool) {
	obj := Gia_ManObj(p, id)
	if mark1 {
		if obj.fMark1 != 0 {
			return
		}
		obj.fMark1 = 1
	} else {
		if obj.fMark0 != 0 {
			return
		}
		obj.fMark0 = 1
	}
	if Gia_ObjIsAnd(obj) == 0 {
		return
	}
	for _, fanin := range giaObjectFanins(p, obj) {
		giaMarkCone(p, fanin, mark1)
	}
}
func Gia_ManSetMark0Dfs_rec(p *Gia_Man_t, id int32) { giaMarkCone(p, id, false) }
func Gia_ManSetMark1Dfs_rec(p *Gia_Man_t, id int32) { giaMarkCone(p, id, true) }
func giaCountMark(p *Gia_Man_t, id int32, mark1 bool) int32 {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return 0
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	var count int32
	if mark1 {
		count = int32(obj.fMark1)
	} else {
		count = int32(obj.fMark0)
	}
	if Gia_ObjIsAnd(obj) != 0 {
		for _, fanin := range giaObjectFanins(p, obj) {
			count += giaCountMark(p, fanin, mark1)
		}
	}
	return count
}
func Gia_ManCountMark0Dfs_rec(p *Gia_Man_t, id int32) int32 { return giaCountMark(p, id, false) }
func Gia_ManCountMark1Dfs_rec(p *Gia_Man_t, id int32) int32 { return giaCountMark(p, id, true) }
func Gia_ManCountMark0Dfs(p *Gia_Man_t, id int32) int32 {
	Gia_ManIncrementTravId(p)
	return giaCountMark(p, id, false)
}
func Gia_ManCountMark1Dfs(p *Gia_Man_t, id int32) int32 {
	Gia_ManIncrementTravId(p)
	return giaCountMark(p, id, true)
}
func Gia_ManDecideWhereToAdd(p *Gia_Man_t, parts []*Vec_Int_t, fanins []*Gia_Obj_t) int32 {
	if len(parts) != 2 || len(fanins) != 2 || vec.Vec_IntSize(parts[0]) != vec.Vec_IntSize(parts[1]) {
		panic("gia: invalid demiter partitions")
	}
	a, b := int32(1), int32(0)
	if vec.Vec_IntSize(parts[0]) > 0 {
		a = Gia_ManCountMark0Dfs(p, Gia_ObjId(p, fanins[0])) + Gia_ManCountMark1Dfs(p, Gia_ObjId(p, fanins[1]))
		b = Gia_ManCountMark0Dfs(p, Gia_ObjId(p, fanins[1])) + Gia_ManCountMark1Dfs(p, Gia_ObjId(p, fanins[0]))
	}
	return giaBool(a < b)
}
func Gia_ManCollectDfs_rec(p *Gia_Man_t, id int32, nodes *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsAnd(obj) == 0 {
		return
	}
	for _, fanin := range giaObjectFanins(p, obj) {
		Gia_ManCollectDfs_rec(p, fanin, nodes)
	}
	vec.Vec_IntPush(nodes, id)
}
func giaCollectDemiterReach(p *Gia_Man_t, second int32, halves bool) *Vec_Int_t {
	if second != 0 && second != 1 {
		panic("gia: invalid demiter side")
	}
	nodes := vec.Vec_IntAlloc(int(p.nObjs))
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p))
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		chosen := i%2 == second
		if halves {
			chosen = (i < Gia_ManCoNum(p)/2) != (second != 0)
		}
		if chosen {
			obj := Gia_ManCo(p, i)
			Gia_ManCollectDfs_rec(p, Gia_ObjFaninId0(obj, obj.id), nodes)
		}
	}
	return nodes
}
func Gia_ManCollectReach(p *Gia_Man_t, odd int32) *Vec_Int_t {
	return giaCollectDemiterReach(p, odd, false)
}
func Gia_ManCollectReach2(p *Gia_Man_t, second int32) *Vec_Int_t {
	return giaCollectDemiterReach(p, second, true)
}
func giaDemiterSides(p *Gia_Man_t, first, second **Gia_Man_t, halves bool) int32 {
	if first == nil || second == nil || Gia_ManRegNum(p) != 0 || Gia_ManCoNum(p)%2 != 0 {
		panic("gia: invalid demiter interface")
	}
	outputs := [2]*Vec_Int_t{vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		side := i % 2
		if halves && i < Gia_ManCoNum(p)/2 {
			side = 0
		} else if halves {
			side = 1
		}
		vec.Vec_IntPush(outputs[side], i)
	}
	*first = Gia_ManDupSelectedOutputs(p, outputs[0])
	*second = Gia_ManDupSelectedOutputs(p, outputs[1])
	return 1
}
func Gia_ManDemiterDual(p *Gia_Man_t, first, second **Gia_Man_t) int32 {
	return giaDemiterSides(p, first, second, false)
}
func Gia_ManDemiterTwoWords(p *Gia_Man_t, first, second **Gia_Man_t) int32 {
	return giaDemiterSides(p, first, second, true)
}
func Gia_ManDupHalfSeq(src *Gia_Man_t, second int32) *Gia_Man_t {
	Gia_ManFillValue(src)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		out.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManPiNum(src); i++ {
		Gia_ManPi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	split := Gia_ManRegNum(src) / 2
	if second == 0 {
		for i := split; i < Gia_ManRegNum(src); i++ {
			Gia_ManRo(src, i).Value = uint32(Gia_ManAppendCi(out))
		}
		for i := int32(0); i < split; i++ {
			Gia_ManRo(src, i).Value = uint32(Gia_ManAppendCi(out))
		}
	} else {
		for i := int32(0); i < Gia_ManRegNum(src); i++ {
			Gia_ManRo(src, i).Value = uint32(Gia_ManAppendCi(out))
		}
	}
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		Gia_ManDupOrderDfs_rec(out, src, Gia_ManPo(src, i))
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		if (i >= split) == (second != 0) {
			Gia_ManDupOrderDfs_rec(out, src, Gia_ManRi(src, i))
		}
	}
	regs := split
	if second != 0 {
		regs = Gia_ManRegNum(src) - split
	}
	Gia_ManSetRegNum(out, regs)
	return out
}

func Gia_ManCollectTopXors_rec(p *Gia_Man_t, obj *Gia_Obj_t, xors *Vec_Int_t) {
	if Gia_ObjRecognizeExor(obj, nil, nil) != 0 || Gia_ObjIsAnd(obj) == 0 {
		vec.Vec_IntPushUnique(xors, obj.id)
		return
	}
	for _, lit := range []int32{Gia_ObjFaninLit0(obj, obj.id), Gia_ObjFaninLit1(obj, obj.id)} {
		if lit&1 != 0 {
			vec.Vec_IntPushUnique(xors, lit>>1)
		} else {
			Gia_ManCollectTopXors_rec(p, Gia_ManObj(p, lit>>1), xors)
		}
	}
}
func Gia_ManCollectTopXors(p *Gia_Man_t) *Vec_Int_t {
	if Gia_ManCoNum(p) == 0 {
		return nil
	}
	xors := vec.Vec_IntAlloc(16)
	if Gia_ManCoNum(p) == 1 {
		obj := Gia_ManCo(p, 0)
		if Gia_ObjFaninC0(obj) != 0 {
			Gia_ManCollectTopXors_rec(p, Gia_ObjFanin0(obj), xors)
		} else {
			vec.Vec_IntPush(xors, Gia_ObjFaninId0(obj, obj.id))
		}
	} else {
		for i := int32(0); i < Gia_ManCoNum(p); i++ {
			obj := Gia_ManCo(p, i)
			id := Gia_ObjFaninId0(obj, obj.id)
			if id > 0 {
				vec.Vec_IntPush(xors, id)
			}
		}
	}
	Gia_ManDupDemiterOrderXors(p, xors)
	vec.Vec_IntReverseOrder(xors)
	Gia_ManCleanMark01(p)
	parts := []*Vec_Int_t{vec.Vec_IntAlloc(16), vec.Vec_IntAlloc(16)}
	for _, id := range vec.Vec_IntArray(xors) {
		obj := Gia_ManObj(p, id)
		var a, b *Gia_Obj_t
		if Gia_ObjRecognizeExor(obj, &a, &b) == 0 {
			a, b = obj, Gia_ManConst0(p)
		} else {
			a, b = Gia_Regular(a), Gia_Regular(b)
		}
		fanins := []*Gia_Obj_t{a, b}
		flip := Gia_ManDecideWhereToAdd(p, parts, fanins)
		first, second := fanins[flip], fanins[1-flip]
		vec.Vec_IntPush(parts[0], first.id)
		vec.Vec_IntPush(parts[1], second.id)
		Gia_ManSetMark0Dfs_rec(p, first.id)
		Gia_ManSetMark1Dfs_rec(p, second.id)
	}
	Gia_ManCleanMark01(p)
	order := vec.Vec_IntAlloc(2 * vec.Vec_IntSize(parts[0]))
	for i := 0; i < vec.Vec_IntSize(parts[0]); i++ {
		vec.Vec_IntPush(order, vec.Vec_IntEntry(parts[0], i))
		vec.Vec_IntPush(order, vec.Vec_IntEntry(parts[1], i))
	}
	vec.Vec_IntReverseOrder(order)
	return order
}
func Gia_ManDemiterToDual(src *Gia_Man_t) *Gia_Man_t {
	order := Gia_ManCollectTopXors(src)
	if order == nil {
		fmt.Println("Cannot demiter because the top-most gate is an AND-gate.")
		return nil
	}
	if vec.Vec_IntSize(order)%2 != 0 {
		panic("gia: invalid demiter order")
	}
	nodes := vec.Vec_IntAlloc(int(src.nObjs))
	Gia_ManIncrementTravId(src)
	Gia_ManCollectAnds(src, vec.Vec_IntArray(order), int32(vec.Vec_IntSize(order)), nodes, nil)
	out := giaDupMappedStart(src)
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	for _, id := range vec.Vec_IntArray(nodes) {
		giaAppendMappedNode(out, src, Gia_ManObj(src, id))
	}
	po := Gia_ManCo(src, 0)
	if Gia_ObjFaninId0(po, po.id) == 0 {
		Gia_ManAppendCo(out, 0)
		Gia_ManAppendCo(out, Gia_ObjFaninC0(po))
	} else {
		Gia_ManSetPhase(src)
		for _, id := range vec.Vec_IntArray(order) {
			obj := Gia_ManObj(src, id)
			Gia_ManAppendCo(out, int32(obj.Value)^int32(obj.fPhase))
		}
	}
	return out
}
