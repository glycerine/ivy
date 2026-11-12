// Delay-oriented K-LUT mapping and exact-area recovery ported from ivyFastMap.c.
package ivy

import (
	"fmt"
	"slices"
	"time"
)

const IVY_INFINITY int32 = 10000

type Ivy_SuppMan_t struct {
	nLimit   int32
	supports []Ivy_Supp_t
	vLuts    [][]*Ivy_Obj_t
	elapsed  time.Duration
}
type Ivy_Supp_t struct {
	fMark, fMark2, fMark3 bool
	nRefs, Delay, DelayR  int32
	pArray                []int32
}

func ivySupportManager(p *Ivy_Man_t) *Ivy_SuppMan_t {
	m, ok := p.pData.(*Ivy_SuppMan_t)
	if !ok || m == nil {
		panic("ivy: fast mapping is not initialized")
	}
	return m
}
func Ivy_ObjSupp(p *Ivy_Man_t, n *Ivy_Obj_t) *Ivy_Supp_t {
	n = Ivy_Regular(n)
	m := ivySupportManager(p)
	if n == nil || n.owner != p || n.recycled || n.Id < 0 || int(n.Id) >= len(m.supports) {
		panic("ivy: invalid mapped object")
	}
	return &m.supports[n.Id]
}
func Ivy_ObjSuppStart(p *Ivy_Man_t, n *Ivy_Obj_t) *Ivy_Supp_t {
	s := Ivy_ObjSupp(p, n)
	*s = Ivy_Supp_t{pArray: []int32{n.Id}}
	return s
}
func ivyFastMapOrder(p *Ivy_Man_t) []*Ivy_Obj_t {
	var nodes []*Ivy_Obj_t
	seen := map[*Ivy_Obj_t]bool{}
	var visit func(*Ivy_Obj_t)
	visit = func(n *Ivy_Obj_t) {
		if n == nil || seen[n] || Ivy_ObjIsNode(n) == 0 {
			return
		}
		seen[n] = true
		visit(Ivy_ObjFanin0(n))
		visit(Ivy_ObjFanin1(n))
		nodes = append(nodes, n)
	}
	for _, n := range p.vObjs {
		visit(n)
	}
	return nodes
}
func Ivy_FastMapPerform(p *Ivy_Man_t, limit, recovery, verbose int32) {
	if limit < 2 || limit > 32 || Ivy_ManCheck(p) == 0 || Ivy_ManLatchNum(p) != 0 || Ivy_ManBufNum(p) != 0 {
		panic("ivy: fast mapping requires a valid combinational buffer-free graph and 2..32 inputs")
	}
	if p.pData != nil {
		if _, ok := p.pData.(*Ivy_SuppMan_t); !ok {
			panic("ivy: manager data already in use")
		}
	}
	start := time.Now()
	m := &Ivy_SuppMan_t{nLimit: limit, supports: make([]Ivy_Supp_t, len(p.vObjs))}
	p.pData = m
	Ivy_ObjSuppStart(p, p.pConst1)
	for _, n := range p.vPis {
		Ivy_ObjSuppStart(p, n)
	}
	nodes := ivyFastMapOrder(p)
	phase := time.Now()
	for _, n := range nodes {
		Ivy_FastMapNode(p, n, limit)
	}
	delay, area := Ivy_FastMapDelay(p), Ivy_FastMapArea(p)
	if verbose != 0 {
		Ivy_FastMapPrint(p, delay, area, time.Since(phase), "Delay oriented mapping")
	}
	if recovery != 0 {
		for pass := 0; pass < 3; pass++ {
			phase = time.Now()
			Ivy_FastMapRequired(p, delay, 0)
			if pass == 1 {
				for _, n := range nodes {
					Ivy_FastMapNodeArea(p, n, limit)
				}
			} else {
				Ivy_FastMapRecover(p, limit)
			}
			delay, area = Ivy_FastMapDelay(p), Ivy_FastMapArea(p)
			if verbose != 0 {
				Ivy_FastMapPrint(p, delay, area, time.Since(phase), fmt.Sprintf("Area recovery %d", pass+1))
			}
		}
	}
	m.elapsed = time.Since(start)
}
func Ivy_FastMapStop(p *Ivy_Man_t) {
	if p.pData == nil {
		return
	}
	ivySupportManager(p)
	p.pData = nil
}
func Ivy_FastMapPrint(p *Ivy_Man_t, delay, area int32, elapsed time.Duration, label string) {
	fmt.Printf("%s : Delay = %3d. Area = %6d. Time = %s\n", label, delay, area, elapsed)
}
func Ivy_FastMapDelay(p *Ivy_Man_t) int32 {
	var delay int32
	for _, output := range p.vPos {
		n := Ivy_ObjFanin0(output)
		if Ivy_ObjIsNode(n) != 0 {
			delay = max(delay, Ivy_ObjSupp(p, n).Delay)
		}
	}
	return delay
}
func Ivy_FastMapArea_rec(p *Ivy_Man_t, n *Ivy_Obj_t, luts *[][]*Ivy_Obj_t) int32 {
	s := Ivy_ObjSupp(p, n)
	if s.fMark || len(s.pArray) == 1 {
		return 0
	}
	s.fMark = true
	area := int32(1)
	for _, id := range s.pArray {
		area += Ivy_FastMapArea_rec(p, Ivy_ManObj(p, id), luts)
	}
	for len(*luts) <= int(s.Delay) {
		*luts = append(*luts, nil)
	}
	(*luts)[s.Delay] = append((*luts)[s.Delay], n)
	return area
}
func Ivy_FastMapArea(p *Ivy_Man_t) int32 {
	m := ivySupportManager(p)
	m.vLuts = nil
	var area int32
	defer func() {
		for i := range m.supports {
			m.supports[i].fMark = false
		}
	}()
	for _, output := range p.vPos {
		area += Ivy_FastMapArea_rec(p, Ivy_ObjFanin0(output), &m.vLuts)
	}
	return area
}
func Ivy_ObjIsNodeInt1(n *Ivy_Obj_t) int32 { return ivyBool(Ivy_ObjIsNode(n) != 0 && n.nRefs == 1) }
func Ivy_ObjIsNodeInt2(n *Ivy_Obj_t) int32 { return ivyBool(Ivy_ObjIsNode(n) != 0 && n.nRefs <= 2) }
func Vec_IntRemoveDup(values []int32, size int32) int32 {
	if size < 0 || int(size) > len(values) {
		panic("ivy: invalid duplicate-removal size")
	}
	unique := slices.Compact(values[:size])
	return int32(len(unique))
}
func Ivy_FastMapMerge(a, b, out *Ivy_Supp_t, limit int32) int32 {
	// Build independently so an oversized/aliased merge never corrupts old cuts.
	merged := make([]int32, 0, len(a.pArray)+len(b.pArray))
	i, j := 0, 0
	for i < len(a.pArray) || j < len(b.pArray) {
		var id int32
		if j == len(b.pArray) || (i < len(a.pArray) && a.pArray[i] < b.pArray[j]) {
			id = a.pArray[i]
			i++
		} else {
			id = b.pArray[j]
			j++
		}
		if len(merged) == 0 || merged[len(merged)-1] != id {
			merged = append(merged, id)
		}
		if len(merged) > int(limit) {
			return 0
		}
	}
	out.pArray = merged
	return 1
}
func Ivy_FastMapNode(p *Ivy_Man_t, n *Ivy_Obj_t, limit int32) {
	a, b := Ivy_ObjSupp(p, Ivy_ObjFanin0(n)), Ivy_ObjSupp(p, Ivy_ObjFanin1(n))
	s := Ivy_ObjSupp(p, n)
	s.fMark = false
	if a.Delay == b.Delay {
		s.Delay = max(int32(1), a.Delay)
	} else if a.Delay > b.Delay {
		s.Delay = a.Delay
		b = &Ivy_Supp_t{pArray: []int32{Ivy_ObjFaninId1(n)}}
	} else {
		s.Delay = b.Delay
		a = &Ivy_Supp_t{pArray: []int32{Ivy_ObjFaninId0(n)}}
	}
	if Ivy_FastMapMerge(a, b, s, limit) == 0 {
		s.Delay++
		s.pArray = []int32{Ivy_ObjFaninId0(n), Ivy_ObjFaninId1(n)}
	}
}
func Ivy_FastMapNodeDelay(p *Ivy_Man_t, n *Ivy_Obj_t) int32 {
	var delay int32
	for _, id := range Ivy_ObjSupp(p, n).pArray {
		delay = max(delay, Ivy_ObjSupp(p, Ivy_ManObj(p, id)).Delay)
	}
	return delay + 1
}
func Ivy_FastMapReadSupp(p *Ivy_Man_t, n *Ivy_Obj_t, leaves *[]int32) {
	*leaves = Ivy_ObjSupp(p, n).pArray
}
func Ivy_FastMapRequired_rec(p *Ivy_Man_t, n, root *Ivy_Obj_t, required int32) {
	s := Ivy_ObjSupp(p, n)
	if n != root && (s.nRefs > 0 || Ivy_ObjIsCi(n) != 0 || Ivy_ObjIsConst1(n) != 0) {
		return
	}
	if n.pFanin0 != nil {
		Ivy_FastMapRequired_rec(p, Ivy_ObjFanin0(n), root, required)
	}
	if n.pFanin1 != nil {
		Ivy_FastMapRequired_rec(p, Ivy_ObjFanin1(n), root, required)
	}
	s.DelayR = required
}
func Ivy_FastMapRequired(p *Ivy_Man_t, delay, intermediate int32) {
	m := ivySupportManager(p)
	for i := range m.supports {
		m.supports[i].DelayR = IVY_INFINITY
		m.supports[i].nRefs = 0
	}
	for _, output := range p.vPos {
		s := Ivy_ObjSupp(p, Ivy_ObjFanin0(output))
		s.DelayR = delay
		s.nRefs++
	}
	for level := len(m.vLuts) - 1; level >= 0; level-- {
		for _, n := range m.vLuts[level] {
			s := Ivy_ObjSupp(p, n)
			for _, id := range s.pArray {
				fanin := Ivy_ObjSupp(p, Ivy_ManObj(p, id))
				fanin.DelayR = min(fanin.DelayR, s.DelayR-1)
				fanin.nRefs++
			}
		}
	}
	if intermediate != 0 {
		for level := len(m.vLuts) - 1; level >= 0; level-- {
			for _, n := range m.vLuts[level] {
				Ivy_FastMapRequired_rec(p, n, n, Ivy_ObjSupp(p, n).DelayR)
			}
		}
	}
}
func Ivy_FastMapNodeRef(p *Ivy_Man_t, n *Ivy_Obj_t) int32 {
	area := int32(1)
	for _, id := range Ivy_ObjSupp(p, n).pArray {
		child := Ivy_ManObj(p, id)
		s := Ivy_ObjSupp(p, child)
		if s.nRefs < 0 {
			panic("ivy: negative LUT references")
		}
		old := s.nRefs
		s.nRefs++
		if old == 0 && len(s.pArray) > 1 {
			area += Ivy_FastMapNodeRef(p, child)
		}
	}
	return area
}
func Ivy_FastMapNodeDeref(p *Ivy_Man_t, n *Ivy_Obj_t) int32 {
	area := int32(1)
	for _, id := range Ivy_ObjSupp(p, n).pArray {
		child := Ivy_ManObj(p, id)
		s := Ivy_ObjSupp(p, child)
		if s.nRefs <= 0 {
			panic("ivy: unbalanced LUT dereference")
		}
		s.nRefs--
		if s.nRefs == 0 && len(s.pArray) > 1 {
			area += Ivy_FastMapNodeDeref(p, child)
		}
	}
	return area
}
func Ivy_FastMapNodeAreaRefed(p *Ivy_Man_t, n *Ivy_Obj_t) int32 {
	if Ivy_ObjIsCi(n) != 0 || Ivy_ObjIsConst1(n) != 0 {
		return 0
	}
	if Ivy_ObjSupp(p, n).nRefs <= 0 {
		panic("ivy: referenced LUT expected")
	}
	area := Ivy_FastMapNodeDeref(p, n)
	if Ivy_FastMapNodeRef(p, n) != area {
		panic("ivy: LUT ref/deref area mismatch")
	}
	return area
}
func Ivy_FastMapNodeAreaDerefed(p *Ivy_Man_t, n *Ivy_Obj_t) int32 {
	if Ivy_ObjIsCi(n) != 0 || Ivy_ObjIsConst1(n) != 0 {
		return 0
	}
	if Ivy_ObjSupp(p, n).nRefs != 0 {
		panic("ivy: unreferenced LUT expected")
	}
	area := Ivy_FastMapNodeRef(p, n)
	if Ivy_FastMapNodeDeref(p, n) != area {
		panic("ivy: LUT ref/deref area mismatch")
	}
	return area
}
func ivyFastMapNodeArea(p *Ivy_Man_t, n *Ivy_Obj_t, limit, threshold int32, exact bool) {
	s := Ivy_ObjSupp(p, n)
	var before int32
	referenced := s.nRefs != 0
	if exact {
		if referenced {
			before = Ivy_FastMapNodeAreaRefed(p, n)
			Ivy_FastMapNodeDeref(p, n)
		} else {
			before = Ivy_FastMapNodeAreaDerefed(p, n)
		}
	}
	old, delay := slices.Clone(s.pArray), Ivy_FastMapNodeDelay(p, n)
	a, b := Ivy_ObjFanin0(n), Ivy_ObjFanin1(n)
	sa, sb := Ivy_ObjSupp(p, a), Ivy_ObjSupp(p, b)
	if a.nRefs > threshold && sa.Delay < s.DelayR {
		sa = &Ivy_Supp_t{pArray: []int32{a.Id}}
	}
	if b.nRefs > threshold && sb.Delay < s.DelayR {
		sb = &Ivy_Supp_t{pArray: []int32{b.Id}}
	}
	if Ivy_FastMapMerge(sa, sb, s, limit) == 0 {
		s.pArray = []int32{a.Id, b.Id}
	}
	s.Delay = Ivy_FastMapNodeDelay(p, n)
	reject := s.Delay > s.DelayR
	if exact {
		refs := s.nRefs
		s.nRefs = 0
		after := Ivy_FastMapNodeAreaDerefed(p, n)
		s.nRefs = refs
		reject = reject || after > before
	}
	if reject {
		s.pArray = old
		s.Delay = delay
	}
	if exact && referenced {
		Ivy_FastMapNodeRef(p, n)
	}
}
func Ivy_FastMapNodeArea2(p *Ivy_Man_t, n *Ivy_Obj_t, limit int32) {
	ivyFastMapNodeArea(p, n, limit, 1, false)
}
func Ivy_FastMapNodeArea(p *Ivy_Man_t, n *Ivy_Obj_t, limit int32) {
	ivyFastMapNodeArea(p, n, limit, 2, true)
}

func Ivy_FastMapCutCost(p *Ivy_Man_t, front []*Ivy_Obj_t) int32 {
	var cost int32
	for _, n := range front {
		cost += ivyBool(Ivy_ObjSupp(p, n).nRefs == 0)
	}
	return cost
}
func Ivy_FastMapMark_rec(p *Ivy_Man_t, n *Ivy_Obj_t) {
	if Ivy_ObjIsTravIdCurrent(p, n) != 0 {
		return
	}
	if Ivy_ObjIsNode(n) == 0 {
		panic("ivy: LUT frontier does not separate its cone")
	}
	Ivy_FastMapMark_rec(p, Ivy_ObjFanin0(n))
	Ivy_FastMapMark_rec(p, Ivy_ObjFanin1(n))
	Ivy_ObjSetTravIdCurrent(p, n)
}
func Ivy_FastMapNodeWillGrow(p *Ivy_Man_t, n *Ivy_Obj_t) int32 {
	return ivyBool(Ivy_ObjIsTravIdCurrent(p, Ivy_ObjFanin0(n)) == 0 && Ivy_ObjIsTravIdCurrent(p, Ivy_ObjFanin1(n)) == 0)
}
func Ivy_FastMapNodeFaninCost(p *Ivy_Man_t, n *Ivy_Obj_t) int32 {
	cost := -ivyBool(Ivy_ObjSupp(p, n).nRefs == 0)
	for _, fanin := range []*Ivy_Obj_t{Ivy_ObjFanin0(n), Ivy_ObjFanin1(n)} {
		if Ivy_ObjIsTravIdCurrent(p, fanin) == 0 && Ivy_ObjSupp(p, fanin).nRefs == 0 {
			cost++
		}
	}
	return cost
}
func Ivy_FastMapNodeFaninUpdate(p *Ivy_Man_t, n *Ivy_Obj_t, front *[]*Ivy_Obj_t) {
	index := slices.Index(*front, n)
	if index < 0 {
		panic("ivy: missing LUT frontier node")
	}
	*front = slices.Delete(*front, index, index+1)
	for _, fanin := range []*Ivy_Obj_t{Ivy_ObjFanin0(n), Ivy_ObjFanin1(n)} {
		if Ivy_ObjIsTravIdCurrent(p, fanin) == 0 {
			Ivy_ObjSetTravIdCurrent(p, fanin)
			*front = append(*front, fanin)
		}
	}
}
func ivyFastMapCompact(p *Ivy_Man_t, front *[]*Ivy_Obj_t, mode int) int32 {
	for _, n := range *front {
		if Ivy_ObjIsCi(n) != 0 || Ivy_ObjIsConst1(n) != 0 {
			continue
		}
		if mode == 0 && Ivy_FastMapNodeWillGrow(p, n) != 0 {
			continue
		}
		cost := Ivy_FastMapNodeFaninCost(p, n)
		if (mode == 1 && cost < 0) || (mode != 1 && cost <= 0) {
			Ivy_FastMapNodeFaninUpdate(p, n, front)
			return 1
		}
	}
	return 0
}
func Ivy_FastMapNodeFaninCompact0(p *Ivy_Man_t, n *Ivy_Obj_t, limit int32, front *[]*Ivy_Obj_t) int32 {
	return ivyFastMapCompact(p, front, 0)
}
func Ivy_FastMapNodeFaninCompact1(p *Ivy_Man_t, n *Ivy_Obj_t, limit int32, front *[]*Ivy_Obj_t) int32 {
	return ivyFastMapCompact(p, front, 1)
}
func Ivy_FastMapNodeFaninCompact2(p *Ivy_Man_t, n *Ivy_Obj_t, limit int32, front *[]*Ivy_Obj_t) int32 {
	return ivyFastMapCompact(p, front, 2)
}
func Ivy_FastMapNodeFaninCompact_int(p *Ivy_Man_t, n *Ivy_Obj_t, limit int32, front *[]*Ivy_Obj_t) int32 {
	if Ivy_FastMapNodeFaninCompact0(p, n, limit, front) != 0 {
		return 1
	}
	if len(*front) < int(limit) && (Ivy_FastMapNodeFaninCompact1(p, n, limit, front) != 0 || Ivy_FastMapNodeFaninCompact2(p, n, limit, front) != 0) {
		return 1
	}
	if len(*front) > int(limit) {
		panic("ivy: LUT frontier exceeds its limit")
	}
	return 0
}
func Ivy_FastMapNodeFaninCompact(p *Ivy_Man_t, n *Ivy_Obj_t, limit int32, front *[]*Ivy_Obj_t) {
	for Ivy_FastMapNodeFaninCompact_int(p, n, limit, front) != 0 {
	}
}
func Ivy_FastMapNodePrepare(p *Ivy_Man_t, n *Ivy_Obj_t, limit int32, front, old *[]*Ivy_Obj_t) {
	*front = (*front)[:0]
	*old = (*old)[:0]
	Ivy_ManIncrementTravId(p)
	for _, id := range Ivy_ObjSupp(p, n).pArray {
		fanin := Ivy_ManObj(p, id)
		*front = append(*front, fanin)
		*old = append(*old, fanin)
		Ivy_ObjSetTravIdCurrent(p, fanin)
	}
	Ivy_FastMapMark_rec(p, n)
}
func Ivy_FastMapNodeUpdate(p *Ivy_Man_t, n *Ivy_Obj_t, front []*Ivy_Obj_t) {
	Ivy_FastMapNodeDeref(p, n)
	s := Ivy_ObjSupp(p, n)
	s.pArray = s.pArray[:0]
	for _, fanin := range front {
		s.pArray = append(s.pArray, fanin.Id)
	}
	// Later support merges require canonical sorted leaves.
	slices.Sort(s.pArray)
	s.pArray = slices.Compact(s.pArray)
	Ivy_FastMapNodeRef(p, n)
}
func Ivy_FastMapNodeRecover(p *Ivy_Man_t, n *Ivy_Obj_t, limit int32, front, old *[]*Ivy_Obj_t) {
	s := Ivy_ObjSupp(p, n)
	previousDelay := Ivy_FastMapNodeDelay(p, n)
	s.Delay = previousDelay
	if s.nRefs == 0 {
		return
	}
	before := Ivy_FastMapNodeAreaRefed(p, n)
	Ivy_FastMapNodePrepare(p, n, limit, front, old)
	Ivy_FastMapNodeDeref(p, n)
	Ivy_FastMapNodeFaninCompact(p, n, limit, front)
	Ivy_FastMapNodeRef(p, n)
	Ivy_FastMapNodeUpdate(p, n, *front)
	s.Delay = Ivy_FastMapNodeDelay(p, n)
	if Ivy_FastMapNodeAreaRefed(p, n) > before || s.Delay > s.DelayR {
		Ivy_FastMapNodeUpdate(p, n, *old)
		s.Delay = previousDelay
	}
}
func ivyFastMapNodeRecoverAlternative(p *Ivy_Man_t, n *Ivy_Obj_t, limit int32, front, old *[]*Ivy_Obj_t, checkDelay bool) {
	s := Ivy_ObjSupp(p, n)
	previousDelay := s.Delay
	if checkDelay {
		previousDelay = Ivy_FastMapNodeDelay(p, n)
		s.Delay = previousDelay
	}
	active := s.nRefs != 0
	var before int32
	if active {
		before = Ivy_FastMapNodeAreaRefed(p, n)
	} else {
		before = Ivy_FastMapNodeAreaDerefed(p, n)
	}
	if before == 1 {
		return
	}
	if !active {
		s.nRefs = 1
		Ivy_FastMapNodeRef(p, n)
	}
	Ivy_FastMapNodePrepare(p, n, limit, front, old)
	Ivy_FastMapNodeFaninCompact(p, n, limit, front)
	Ivy_FastMapNodeUpdate(p, n, *front)
	if checkDelay {
		s.Delay = Ivy_FastMapNodeDelay(p, n)
	}
	if Ivy_FastMapNodeAreaRefed(p, n) > before || (checkDelay && s.Delay > s.DelayR) {
		Ivy_FastMapNodeUpdate(p, n, *old)
		s.Delay = previousDelay
	}
	if !active {
		s.nRefs = 0
		Ivy_FastMapNodeDeref(p, n)
	}
}
func Ivy_FastMapNodeRecover2(p *Ivy_Man_t, n *Ivy_Obj_t, limit int32, front, old *[]*Ivy_Obj_t) {
	ivyFastMapNodeRecoverAlternative(p, n, limit, front, old, false)
}
func Ivy_FastMapNodeRecover4(p *Ivy_Man_t, n *Ivy_Obj_t, limit int32, front, old *[]*Ivy_Obj_t) {
	ivyFastMapNodeRecoverAlternative(p, n, limit, front, old, true)
}
func Ivy_FastMapRecover(p *Ivy_Man_t, limit int32) {
	var front, old []*Ivy_Obj_t
	Ivy_ManCleanTravId(p)
	for _, n := range ivyFastMapOrder(p) {
		Ivy_FastMapNodeRecover(p, n, limit, &front, &old)
	}
}
