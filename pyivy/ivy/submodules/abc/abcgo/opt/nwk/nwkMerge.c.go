package nwk

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"sort"
)

func Nwk_ManGraphAlloc(max int32) *Nwk_Grf_t {
	if max < 0 {
		panic("nwk: negative graph capacity")
	}
	return &Nwk_Grf_t{MaxVertices: int(max), Pairs: vec.Vec_IntAlloc(0), edgeSet: map[Nwk_Edg_t]bool{}, hashSize: int(util.Abc_PrimeCudd(uint32(3 * max)))}
}
func Nwk_ManGraphFree(p *Nwk_Grf_t) {
	p.Edges = nil
	p.Vertices = nil
	p.edgeSet = nil
	p.Pairs = nil
	p.MapIDToLUT = nil
}
func Nwk_ManGraphReportMemoryUsage(p *Nwk_Grf_t) {
	fmt.Printf("Graph: %d vertices, %d edges.\n", len(p.Vertices)-1, len(p.Edges))
}
func Nwk_ManGraphHashEdge(p *Nwk_Grf_t, a, b int32) {
	if a == b {
		return
	}
	if a < 0 || b < 0 {
		panic("nwk: negative merge vertex")
	}
	if a > b {
		a, b = b, a
	}
	edge := Nwk_Edg_t{int(a), int(b)}
	if !p.edgeSet[edge] {
		p.edgeSet[edge] = true
		p.Edges = append(p.Edges, edge)
	}
}
func Nwk_ManGraphListAdd(p *Nwk_Grf_t, head *int, v *Nwk_Vrt_t) {
	v.Prev = 0
	v.Next = *head
	if *head != 0 {
		p.Vertices[*head].Prev = v.ID
	}
	*head = v.ID
}
func Nwk_ManGraphListDelete(p *Nwk_Grf_t, head *int, v *Nwk_Vrt_t) {
	if v.Prev != 0 {
		p.Vertices[v.Prev].Next = v.Next
	}
	if v.Next != 0 {
		p.Vertices[v.Next].Prev = v.Prev
	}
	if *head == v.ID {
		*head = v.Next
	}
	v.Prev = 0
	v.Next = 0
}
func graphList(p *Nwk_Grf_t, v *Nwk_Vrt_t) *int {
	if len(v.Edges) == 1 {
		return &p.Lists1[min(NWK_MAX_LIST, len(p.Vertices[v.Edges[0]].Edges))]
	}
	return &p.Lists2[min(NWK_MAX_LIST, len(v.Edges))]
}
func Nwk_ManGraphListInsert(p *Nwk_Grf_t, v *Nwk_Vrt_t) {
	if len(v.Edges) > 0 {
		Nwk_ManGraphListAdd(p, graphList(p, v), v)
	}
}
func Nwk_ManGraphListExtract(p *Nwk_Grf_t, v *Nwk_Vrt_t) {
	Nwk_ManGraphListDelete(p, graphList(p, v), v)
}
func Nwk_ManGraphPrepare(p *Nwk_Grf_t) {
	ids := map[int]bool{}
	for _, e := range p.Edges {
		ids[e.Node1] = true
		ids[e.Node2] = true
	}
	ordered := []int{}
	for id := range ids {
		ordered = append(ordered, id)
	}
	sort.Ints(ordered)
	mapping := map[int]int{}
	p.Vertices = make([]*Nwk_Vrt_t, len(ordered)+1)
	p.MapIDToLUT = make([]int32, len(ordered)+1)
	p.Lists1 = [NWK_MAX_LIST + 1]int{}
	p.Lists2 = [NWK_MAX_LIST + 1]int{}
	for i, id := range ordered {
		mapping[id] = i + 1
		p.MapIDToLUT[i+1] = int32(id)
		p.Vertices[i+1] = &Nwk_Vrt_t{ID: i + 1}
	}
	buckets := make([][]Nwk_Edg_t, max(p.hashSize, 1))
	for _, e := range p.Edges {
		key := uint32(741457*uint32(e.Node1)+4256249*uint32(e.Node2)) % uint32(len(buckets))
		buckets[key] = append(buckets[key], e)
	}
	for _, bucket := range buckets {
		for i := len(bucket) - 1; i >= 0; i-- {
			e := bucket[i]
			a, b := mapping[e.Node1], mapping[e.Node2]
			p.Vertices[a].Edges = append(p.Vertices[a].Edges, b)
			p.Vertices[b].Edges = append(p.Vertices[b].Edges, a)
		}
	}
	for _, v := range p.Vertices[1:] {
		Nwk_ManGraphListInsert(p, v)
	}
	vec.Vec_IntClear(p.Pairs)
}
func Nwk_ManGraphSortPairs(p *Nwk_Grf_t) {
	pairs := [][2]int32{}
	a := vec.Vec_IntArray(p.Pairs)
	for i := 0; i < len(a); i += 2 {
		pairs = append(pairs, [2]int32{a[i], a[i+1]})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i][0] < pairs[j][0] })
	vec.Vec_IntClear(p.Pairs)
	for _, pair := range pairs {
		vec.Vec_IntPush(p.Pairs, pair[0])
		vec.Vec_IntPush(p.Pairs, pair[1])
	}
}
func Nwk_ManGraphCheckLists(p *Nwk_Grf_t) {
	for i := 1; i <= NWK_MAX_LIST; i++ {
		for id := p.Lists1[i]; id != 0; id = p.Vertices[id].Next {
			v := p.Vertices[id]
			if len(v.Edges) != 1 || min(NWK_MAX_LIST, len(p.Vertices[v.Edges[0]].Edges)) != i {
				panic("nwk: invalid singleton list")
			}
		}
		for id := p.Lists2[i]; id != 0; id = p.Vertices[id].Next {
			if min(NWK_MAX_LIST, len(p.Vertices[id].Edges)) != i {
				panic("nwk: invalid degree list")
			}
		}
	}
}
func Nwk_ManGraphVertexRemoveEdge(v, other *Nwk_Vrt_t) {
	for i, id := range v.Edges {
		if id == other.ID {
			v.Edges = append(v.Edges[:i], v.Edges[i+1:]...)
			return
		}
	}
	panic("nwk: absent graph edge")
}
func Nwk_ManGraphUpdate(p *Nwk_Grf_t, a, b *Nwk_Vrt_t) {
	Nwk_ManGraphListExtract(p, a)
	Nwk_ManGraphListExtract(p, b)
	for _, pair := range [][2]*Nwk_Vrt_t{{a, b}, {b, a}} {
		v, other := pair[0], pair[1]
		for _, id := range v.Edges {
			changed := p.Vertices[id]
			if changed == other {
				continue
			}
			Nwk_ManGraphListExtract(p, changed)
			if len(changed.Edges) > 1 {
				for _, neighbor := range changed.Edges {
					leaf := p.Vertices[neighbor]
					if leaf == v || len(leaf.Edges) > 1 {
						continue
					}
					Nwk_ManGraphListExtract(p, leaf)
					saved := changed.Edges
					changed.Edges = changed.Edges[:len(changed.Edges)-1]
					Nwk_ManGraphListInsert(p, leaf)
					changed.Edges = saved
				}
			}
			Nwk_ManGraphVertexRemoveEdge(changed, v)
			if len(changed.Edges) > 0 {
				Nwk_ManGraphListInsert(p, changed)
			}
		}
	}
	x, y := p.MapIDToLUT[a.ID], p.MapIDToLUT[b.ID]
	if x > y {
		x, y = y, x
	}
	vec.Vec_IntPush(p.Pairs, x)
	vec.Vec_IntPush(p.Pairs, y)
}
func Nwk_ManGraphListLength(p *Nwk_Grf_t, head int32) int32 {
	var n int32
	for id := int(head); id != 0; id = p.Vertices[id].Next {
		n++
	}
	return n
}
func Nwk_ManGraphListFindMinEdge(p *Nwk_Grf_t, v *Nwk_Vrt_t) *Nwk_Vrt_t {
	var best *Nwk_Vrt_t
	for _, id := range v.Edges {
		n := p.Vertices[id]
		if best == nil || len(n.Edges) < len(best.Edges) {
			best = n
		}
	}
	return best
}
func Nwk_ManGraphListFindMin(p *Nwk_Grf_t, head int32) *Nwk_Vrt_t {
	var best *Nwk_Vrt_t
	cost := int(^uint(0) >> 1)
	count := 0
	for id := int(head); id != 0 && count < 10000; id = p.Vertices[id].Next {
		v := p.Vertices[id]
		for _, neighbor := range v.Edges {
			if len(p.Vertices[neighbor].Edges) < cost {
				cost = len(p.Vertices[neighbor].Edges)
				best = v
			}
		}
		count++
	}
	return best
}
func Nwk_ManGraphSolve(p *Nwk_Grf_t) {
	Nwk_ManGraphPrepare(p)
	for {
		var a, b *Nwk_Vrt_t
		for i := 1; i <= NWK_MAX_LIST; i++ {
			if p.Lists1[i] != 0 {
				a = p.Vertices[p.Lists1[i]]
				b = p.Vertices[a.Edges[0]]
				break
			}
		}
		if a == nil {
			for i := 2; i <= NWK_MAX_LIST; i++ {
				if p.Lists2[i] != 0 {
					a = Nwk_ManGraphListFindMin(p, int32(p.Lists2[i]))
					b = Nwk_ManGraphListFindMinEdge(p, a)
					break
				}
			}
		}
		if a == nil {
			break
		}
		Nwk_ManGraphUpdate(p, a, b)
	}
	Nwk_ManGraphSortPairs(p)
}
func Nwk_ManLutMergeReadGraph(path string) *Nwk_Grf_t {
	f, e := os.Open(path)
	if e != nil {
		return nil
	}
	defer f.Close()
	var label string
	var n, m int32
	if _, e = fmt.Fscan(f, &label, &n); e != nil {
		return nil
	}
	if _, e = fmt.Fscan(f, &label, &m); e != nil {
		return nil
	}
	p := Nwk_ManGraphAlloc(n)
	for {
		var a, b int32
		if _, e = fmt.Fscan(f, &label, &a, &b); e != nil {
			break
		}
		Nwk_ManGraphHashEdge(p, a, b)
	}
	return p
}
func Nwk_ManLutMergeGraphTest(path string) int32 {
	p := Nwk_ManLutMergeReadGraph(path)
	if p == nil {
		return 0
	}
	Nwk_ManGraphSolve(p)
	Nwk_ManGraphReportMemoryUsage(p)
	return 1
}

func Nwk_ManMarkFanins_rec(o *Nwk_Obj_t, minimum int32) {
	if o.Type != NWK_OBJ_NODE || Nwk_ObjIsTravIdCurrent(o) != 0 {
		return
	}
	Nwk_ObjSetTravIdCurrent(o)
	if o.Level < minimum {
		return
	}
	for _, n := range o.Fanins {
		Nwk_ManMarkFanins_rec(n, minimum)
	}
}
func Nwk_ManMarkFanouts_rec(o *Nwk_Obj_t, maximum, fanoutMaximum int32) {
	if o.Type != NWK_OBJ_NODE || Nwk_ObjIsTravIdCurrent(o) != 0 {
		return
	}
	Nwk_ObjSetTravIdCurrent(o)
	if o.Level > maximum || int32(len(o.Fanouts)) > fanoutMaximum {
		return
	}
	for _, n := range o.Fanouts {
		Nwk_ManMarkFanouts_rec(n, maximum, fanoutMaximum)
	}
}
func Nwk_ManCollectCircle(start, next *Vec_Ptr_t, maximum int32) {
	vec.Vec_PtrClear(next)
	for _, o := range vectorNodes(start) {
		for _, n := range o.Fanins {
			if n.Type != NWK_OBJ_NODE || Nwk_ObjIsTravIdCurrent(n) != 0 {
				continue
			}
			Nwk_ObjSetTravIdCurrent(n)
			vec.Vec_PtrPush(next, n)
		}
		for _, n := range o.Fanouts {
			if n.Type != NWK_OBJ_NODE || Nwk_ObjIsTravIdCurrent(n) != 0 {
				continue
			}
			Nwk_ObjSetTravIdCurrent(n)
			if int32(len(n.Fanouts)) <= maximum {
				vec.Vec_PtrPush(next, n)
			}
		}
	}
}
func Nwk_ManCollectNonOverlapCands(o *Nwk_Obj_t, start, next, candidates *Vec_Ptr_t, pars *Nwk_LMPars_t) {
	vec.Vec_PtrClear(candidates)
	if pars.MaxSuppSize-int32(len(o.Fanins)) <= 1 {
		return
	}
	if pars.MaxDistance <= 0 {
		panic("nwk: non-overlap search requires positive distance")
	}
	vec.Vec_PtrClear(start)
	vec.Vec_PtrPush(start, o)
	Nwk_ManIncrementTravId(o.Manager)
	Nwk_ObjSetTravIdCurrent(o)
	for distance := int32(1); distance <= pars.MaxDistance; distance++ {
		Nwk_ManCollectCircle(start, next, pars.MaxFanout)
		start, next = next, start
		for _, n := range vectorNodes(start) {
			vec.Vec_PtrPush(candidates, n)
		}
	}
	Nwk_ManIncrementTravId(o.Manager)
	if pars.UseTfiTfo {
		Nwk_ObjSetTravIdCurrent(o)
	} else {
		Nwk_ObjSetTravIdPrevious(o)
		Nwk_ManMarkFanins_rec(o, o.Level-pars.MaxDistance)
		Nwk_ObjSetTravIdPrevious(o)
		Nwk_ManMarkFanouts_rec(o, o.Level+pars.MaxDistance, pars.MaxFanout)
	}
	keep := []*Nwk_Obj_t{}
	for _, n := range vectorNodes(candidates) {
		if Nwk_ObjIsTravIdCurrent(n) != 0 || int32(len(o.Fanins)+len(n.Fanins)) > pars.MaxSuppSize || absLevel(o.Level-n.Level) > pars.MaxLevelDiff {
			continue
		}
		keep = append(keep, n)
	}
	vec.Vec_PtrClear(candidates)
	for _, n := range keep {
		vec.Vec_PtrPush(candidates, n)
	}
}
func absLevel(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}
func Nwk_ManCountTotalFanins(o, candidate *Nwk_Obj_t) int32 {
	count := int32(len(o.Fanins))
	for _, n := range candidate.Fanins {
		if !n.MarkC {
			count++
		}
	}
	return count
}
func Nwk_ManCollectOverlapCands(o *Nwk_Obj_t, candidates *Vec_Ptr_t, pars *Nwk_LMPars_t) {
	for _, n := range o.Fanins {
		n.MarkC = true
	}
	defer func() {
		for _, n := range o.Fanins {
			n.MarkC = false
		}
	}()
	vec.Vec_PtrClear(candidates)
	Nwk_ManIncrementTravId(o.Manager)
	Nwk_ObjSetTravIdCurrent(o)
	for _, fanin := range o.Fanins {
		if fanin.Type != NWK_OBJ_NODE || int32(len(fanin.Fanouts)) > pars.MaxFanout {
			continue
		}
		for _, n := range fanin.Fanouts {
			if n.Type != NWK_OBJ_NODE || Nwk_ObjIsTravIdCurrent(n) != 0 {
				continue
			}
			Nwk_ObjSetTravIdCurrent(n)
			if absLevel(o.Level-n.Level) > pars.MaxLevelDiff || Nwk_ManCountTotalFanins(o, n) > pars.MaxSuppSize {
				continue
			}
			vec.Vec_PtrPush(candidates, n)
		}
	}
}
func Nwk_ManLutMerge(p *Nwk_Man_t, options any) *Vec_Int_t {
	pars, ok := options.(*Nwk_LMPars_t)
	if !ok || pars == nil {
		panic("nwk: LUT merge parameters required")
	}
	if pars.MaxLutSize < 0 || pars.MaxSuppSize < 0 || pars.MaxDistance < 0 || pars.MaxLevelDiff < 0 || pars.MaxFanout < 0 {
		panic("nwk: negative LUT merge limit")
	}
	graph := Nwk_ManGraphAlloc(Nwk_ManNodeNum(p))
	start, next, overlap, disjoint := vec.Vec_PtrAlloc(0), vec.Vec_PtrAlloc(0), vec.Vec_PtrAlloc(0), vec.Vec_PtrAlloc(0)
	for _, o := range p.Objects {
		if o == nil || o.Type != NWK_OBJ_NODE || int32(len(o.Fanins)) > pars.MaxLutSize {
			continue
		}
		Nwk_ManCollectOverlapCands(o, overlap, pars)
		vec.Vec_PtrClear(disjoint)
		if pars.UseDiffSupp {
			Nwk_ManCollectNonOverlapCands(o, start, next, disjoint, pars)
		}
		for _, v := range []*Vec_Ptr_t{overlap, disjoint} {
			for _, n := range vectorNodes(v) {
				Nwk_ManGraphHashEdge(graph, o.ID, n.ID)
			}
		}
	}
	Nwk_ManGraphSolve(graph)
	if pars.Verbose {
		fmt.Fprintf(writer(p), "Mergable LUTs = %d. Pairs = %d.\n", len(graph.Vertices)-1, vec.Vec_IntSize(graph.Pairs)/2)
	}
	return graph.Pairs
}
