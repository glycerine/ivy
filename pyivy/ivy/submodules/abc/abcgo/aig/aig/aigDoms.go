package aig

import (
	"fmt"
	"math/bits"
	"sort"
)

/**CFile****************************************************************

  FileName    [aigDoms.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Computing multi-output dominators.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigDoms.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

// Aig_Dom_t is an ordered multi-output dominator cut. Signature accelerates
// subset tests; Nodes is authoritative and contains distinct ascending IDs.
type Aig_Dom_t struct {
	Signature uint64
	Nodes     []int
}

// Aig_Sto_t owns dominator sets and analysis-only marks/reference counts.
// Direct output edges receive synthetic boundary IDs starting at NodeCount;
// explicit Output nodes use their existing IDs. OutputIDs exposes this mapping.
// The graph must not be structurally edited while this store is in use.
type Aig_Sto_t struct {
	Limit                                                  int
	Graph                                                  *Graph
	Dominators                                             [][]*Aig_Dom_t
	OutputIDs                                              []int
	Times                                                  []int
	DomNodes, Total, FilteredSupersets, FilteredCandidates int
	fanins, fanouts                                        [][]int
	marked                                                 []bool
	refs                                                   []int
}

func Aig_ManDomStart(g *Graph, limit int) *Aig_Sto_t {
	if g == nil || limit <= 0 {
		panic("aig: dominator limit must be positive")
	}
	count := len(g.nodes) + len(g.outputs)
	s := &Aig_Sto_t{
		Limit: limit, Graph: g,
		Dominators: make([][]*Aig_Dom_t, count),
		Times:      make([]int, count),
		fanins:     make([][]int, count), fanouts: make([][]int, count),
		marked: make([]bool, count), refs: make([]int, count),
	}
	for id, n := range g.nodes {
		if g.deleted[id] {
			continue
		}
		switch n.Kind {
		case And, ExclusiveOr:
			s.fanins[id] = []int{n.Fanins[0].NodeID(), n.Fanins[1].NodeID()}
		case Buffer, Output:
			s.fanins[id] = []int{n.Fanins[0].NodeID()}
		}
	}
	for i, e := range g.outputs {
		if !g.valid(e) {
			panic("aig: invalid dominator output")
		}
		id := e.NodeID()
		if g.nodes[id].Kind != Output {
			id = len(g.nodes) + i
			s.fanins[id] = []int{e.NodeID()}
		}
		s.OutputIDs = append(s.OutputIDs, id)
	}
	for id, inputs := range s.fanins {
		for _, f := range inputs {
			if f < 0 || f >= len(g.nodes) || g.deleted[f] {
				panic("aig: invalid dominator fanin")
			}
			s.fanouts[f] = append(s.fanouts[f], id)
			s.refs[f]++
		}
	}

	return s
}
func Aig_ManDomStop(s *Aig_Sto_t) {
	if s != nil {
		*s = Aig_Sto_t{}
	}
}
func domSignature(ids []int) uint64 {
	var signature uint64
	for _, id := range ids {
		if id < 0 {
			panic("aig: negative dominator id")
		}
		signature |= uint64(1) << uint(id%63)
	}
	return signature
}

// NewDominator constructs an ordered set from node IDs and owns its storage.
func NewDominator(ids ...int) *Aig_Dom_t {
	nodes := append([]int(nil), ids...)
	sort.Ints(nodes)
	unique := nodes[:0]
	for _, id := range nodes {
		if len(unique) == 0 || unique[len(unique)-1] != id {
			unique = append(unique, id)
		}
	}
	return newAigDom(unique)
}
func newAigDom(ids []int) *Aig_Dom_t {
	return &Aig_Dom_t{Signature: domSignature(ids), Nodes: append([]int(nil), ids...)}
}
func Aig_ObjAddTriv(s *Aig_Sto_t, id int, doms []*Aig_Dom_t) {
	if id < 0 || id >= len(s.Dominators) || s.Dominators[id] != nil {
		panic("aig: dominator initialized twice")
	}
	s.Dominators[id] = append([]*Aig_Dom_t{newAigDom([]int{id})}, doms...)
}
func Aig_ObjDomVecDup(_ *Aig_Sto_t, doms []*Aig_Dom_t, skip int) []*Aig_Dom_t {
	if skip < 0 || skip > len(doms) {
		panic("aig: invalid dominator skip")
	}
	out := make([]*Aig_Dom_t, 0, len(doms)-skip)
	for _, d := range doms[skip:] {
		out = append(out, newAigDom(d.Nodes))
	}
	return out
}

// Recycle releases the caller's slice; manager-owned sets are released by Stop.
func Aig_ObjDomVecRecycle(_ *Aig_Sto_t, doms *[]*Aig_Dom_t) { *doms = nil }
func Aig_ObjDomPrint(s *Aig_Sto_t, d *Aig_Dom_t, num int) {
	fmt.Printf("%4d : {", num)
	for _, id := range d.Nodes {
		fmt.Printf(" %4d", id)
	}
	for i := len(d.Nodes); i < s.Limit; i++ {
		fmt.Print("     ")
	}
	fmt.Println(" }")
}
func Aig_ObjDomVecPrint(s *Aig_Sto_t, doms []*Aig_Dom_t) {
	for i, d := range doms {
		Aig_ObjDomPrint(s, d, i)
	}
}
func Aig_ManDomPrint(s *Aig_Sto_t) {
	n := len(s.Graph.inputs) - int(s.Graph.nRegs)
	for i, e := range s.Graph.inputs[:n] {
		fmt.Printf("*** PI %4d %4d :\n", i, e.NodeID())
		Aig_ObjDomVecPrint(s, s.Dominators[e.NodeID()])
	}
}
func Aig_ObjDomCheck(d *Aig_Dom_t) bool {
	if d == nil {
		return false
	}
	for i, id := range d.Nodes {
		if id < 0 || i > 0 && d.Nodes[i-1] >= id {
			return false
		}
	}
	return true
}
func Aig_ObjDomCheckDominance(a, b *Aig_Dom_t) bool {
	i, j := 0, 0
	for i < len(a.Nodes) && j < len(b.Nodes) {
		if a.Nodes[i] == b.Nodes[j] {
			i++
			j++
		} else if a.Nodes[i] > b.Nodes[j] {
			j++
		} else {
			return false
		}
	}
	return i == len(a.Nodes)
}
func Aig_ObjDomFilter(s *Aig_Sto_t, doms *[]*Aig_Dom_t, d *Aig_Dom_t) bool {
	d.Signature = domSignature(d.Nodes)
	for i := 0; i < len(*doms); i++ {
		old := (*doms)[i]
		old.Signature = domSignature(old.Nodes)
		if len(old.Nodes) > len(d.Nodes) {
			if i == 0 || old.Signature&d.Signature != d.Signature {
				continue
			}
			if Aig_ObjDomCheckDominance(d, old) {
				copy((*doms)[i:], (*doms)[i+1:])
				*doms = (*doms)[:len(*doms)-1]
				i--
				s.FilteredSupersets++
			}
		} else if old.Signature&d.Signature == old.Signature && Aig_ObjDomCheckDominance(old, d) {
			s.FilteredCandidates++
			return true
		}
	}
	return false
}
func Aig_ObjDomMergeOrdered(a, b, out *Aig_Dom_t, limit int) bool {
	if limit <= 0 || !Aig_ObjDomCheck(a) || !Aig_ObjDomCheck(b) {
		return false
	}
	capacity := len(a.Nodes) + len(b.Nodes)
	if capacity > limit {
		capacity = limit
	}
	merged := make([]int, 0, capacity)
	i, j := 0, 0
	for i < len(a.Nodes) || j < len(b.Nodes) {
		var id int
		if j == len(b.Nodes) || i < len(a.Nodes) && a.Nodes[i] < b.Nodes[j] {
			id = a.Nodes[i]
			i++
		} else if i == len(a.Nodes) || b.Nodes[j] < a.Nodes[i] {
			id = b.Nodes[j]
			j++
		} else {
			id = a.Nodes[i]
			i++
			j++
		}
		if len(merged) == limit {
			return false
		}
		merged = append(merged, id)
	}
	*out = *newAigDom(merged)
	return true
}
func Aig_ObjDomMergeTwo(a, b, out *Aig_Dom_t, limit int) bool {
	return Aig_ObjDomMergeOrdered(a, b, out, limit)
}
func Aig_ObjDomMerge(s *Aig_Sto_t, a, b []*Aig_Dom_t) []*Aig_Dom_t {
	var result []*Aig_Dom_t
	for _, d := range a {
		d.Signature = domSignature(d.Nodes)
	}
	for _, d := range b {
		d.Signature = domSignature(d.Nodes)
	}
	for _, x := range a {
		for _, y := range b {
			if bits.OnesCount64(x.Signature|y.Signature) > s.Limit {
				continue
			}
			d := new(Aig_Dom_t)
			if Aig_ObjDomMergeTwo(x, y, d, s.Limit) && !Aig_ObjDomFilter(s, &result, d) {
				result = append(result, d)
			}
		}
	}
	return result
}
func Aig_ObjDomUnion(s *Aig_Sto_t, dst *[]*Aig_Dom_t, src []*Aig_Dom_t) {
	for i, d := range src {
		if i == 0 {
			continue
		}
		if !Aig_ObjDomFilter(s, dst, d) {
			*dst = append(*dst, newAigDom(d.Nodes))
		}
	}
}
func Aig_ObjDomCompute(s *Aig_Sto_t, e Edge) {
	id := e.NodeID()
	if id >= len(s.Graph.nodes) || s.Graph.deleted[id] {
		panic("aig: invalid dominator object")
	}
	if s.Graph.nodes[id].Kind == And || s.Graph.nodes[id].Kind == ExclusiveOr {
		s.DomNodes++
	}
	var fans []int
	g := s.Graph
	for i, fan := 0, Aig_ObjFanout0Int(g, int32(id)); i < g.node(e).Refs; i, fan = i+1, Aig_ObjFanoutNext(g, fan) {
		f := int(fan >> 1)
		if g.nodes[f].TravID == g.travID {
			if s.Dominators[f] == nil {
				panic("aig: missing fanout dominators; clean dangling logic before analysis")
			}
			fans = append(fans, f)
		}
	}
	if len(fans) == 0 {
		return
	}
	doms := Aig_ObjDomVecDup(s, s.Dominators[fans[0]], 0)
	for _, f := range fans[1:] {
		doms = Aig_ObjDomMerge(s, doms, s.Dominators[f])
	}
	Aig_ObjAddTriv(s, id, doms)
	s.Total += len(s.Dominators[id])
}
func domRegisterCount(g *Graph) int {
	n := int(g.nRegs)
	if n < 0 || n > len(g.inputs) || n > len(g.outputs) {
		panic("aig: invalid register boundary")
	}
	return n
}
func Aig_ManMarkFlopTfi_rec(g *Graph, e Edge) int {
	if e.Complemented() {
		panic("aig: flop TFI requires a regular object")
	}
	id := e.NodeID()
	if g.nodes[id].TravID == g.travID {
		return 0
	}
	g.nodes[id].TravID = g.travID
	n := g.nodes[id]
	switch n.Kind {
	case Constant, Input:
		return 1
	case And, ExclusiveOr:
		return Aig_ManMarkFlopTfi_rec(g, Aig_Regular(n.Fanins[0])) + Aig_ManMarkFlopTfi_rec(g, Aig_Regular(n.Fanins[1]))
	default:
		return Aig_ManMarkFlopTfi_rec(g, Aig_Regular(n.Fanins[0]))
	}
}
func Aig_ManMarkFlopTfi(g *Graph) {
	Aig_ManIncrementTravId(g)
	n := domRegisterCount(g)
	for _, e := range g.outputs[len(g.outputs)-n:] {
		Aig_ManMarkFlopTfi_rec(g, e)
	}
}
func domCompute(g *Graph, limit int, mode string) *Aig_Sto_t {
	if g.fanouts == nil {
		panic("aig: dominator driver requires fanout allocation")
	}
	s := Aig_ManDomStart(g, limit)
	regs := domRegisterCount(g)
	boundaries := s.OutputIDs
	if mode == "flops" {
		boundaries = boundaries[len(boundaries)-regs:]
	}
	for _, id := range boundaries {
		Aig_ObjAddTriv(s, id, nil)
	}
	nodes := Aig_ManDfsReverse(g)
	if mode != "nodes" {
		Aig_ManMarkFlopTfi(g)
	}
	for _, e := range nodes {
		if mode == "nodes" || Aig_ObjIsTravIdCurrent(g, e) {
			Aig_ObjDomCompute(s, e)
		}
	}
	inputs := g.inputs
	if mode == "pis" {
		inputs = inputs[:len(inputs)-regs]
	}
	for _, e := range inputs {
		Aig_ObjDomCompute(s, e)
	}
	return s
}
func Aig_ManComputeDomsFlops(g *Graph, limit int) *Aig_Sto_t { return domCompute(g, limit, "flops") }
func Aig_ManComputeDomsPis(g *Graph, limit int) *Aig_Sto_t   { return domCompute(g, limit, "pis") }
func Aig_ManComputeDomsNodes(g *Graph, limit int) *Aig_Sto_t { return domCompute(g, limit, "nodes") }
func Aig_ObjDomCollect(s *Aig_Sto_t, cut []int) []*Aig_Dom_t {
	var out []*Aig_Dom_t
	for i, id := range cut {
		if id < 0 || id >= len(s.Dominators) {
			panic("aig: invalid dominator cut id")
		}
		doms := s.Dominators[id]
		if len(doms) == 0 {
			continue
		}
		if i == 0 {
			out = Aig_ObjDomVecDup(s, doms, 1)
		} else {
			Aig_ObjDomUnion(s, &out, doms)
		}
	}
	return out
}
func domVolume(s *Aig_Sto_t, id int, seen []bool) int {
	if seen[id] {
		return 0
	}
	seen[id] = true
	if s.marked[id] || len(s.fanins[id]) == 0 {
		return 1
	}
	n := 0
	for _, f := range s.fanins[id] {
		n += domVolume(s, f, seen)
	}
	return n
}
func Aig_ObjDomVolume_rec(s *Aig_Sto_t, id int, seen []bool) int { return domVolume(s, id, seen) }
func Aig_ObjDomVolume(s *Aig_Sto_t, d *Aig_Dom_t) int {
	seen := make([]bool, len(s.fanins))
	n := 0
	for _, id := range d.Nodes {
		n += domVolume(s, id, seen)
	}
	return n
}
func Aig_ObjDomDeref_rec(s *Aig_Sto_t, id int) int { return domDeref(s, id, nil) }
func domDeref(s *Aig_Sto_t, id int, barriers map[int]bool) int {
	if s.refs[id] <= 0 {
		panic("aig: dominator reference underflow")
	}
	s.refs[id]--
	if s.refs[id] > 0 {
		return 0
	}
	if s.marked[id] {
		return 1
	}
	if barriers[id] {
		return 0
	}
	n := 0
	for _, f := range s.fanins[id] {
		n += domDeref(s, f, barriers)
	}
	return n
}
func Aig_ObjDomRef_rec(s *Aig_Sto_t, id int) int { return domRef(s, id, nil) }
func domRef(s *Aig_Sto_t, id int, barriers map[int]bool) int {
	old := s.refs[id]
	s.refs[id]++
	if old > 0 {
		return 0
	}
	if s.marked[id] {
		return 1
	}
	if barriers[id] {
		return 0
	}
	n := 0
	for _, f := range s.fanins[id] {
		n += domRef(s, f, barriers)
	}
	return n
}
func Aig_ObjDomDomed(s *Aig_Sto_t, d *Aig_Dom_t) int {
	// A cut may include related nodes. Their input edges are already removed
	// explicitly, so reaching another cut node must not remove them a second time.
	if !Aig_ObjDomCheck(d) {
		panic("aig: invalid dominated-cone cut")
	}
	barriers := map[int]bool{}
	for _, id := range d.Nodes {
		if id >= len(s.fanins) {
			panic("aig: invalid dominated-cone id")
		}
		if id < len(s.Graph.nodes) && s.Graph.nodes[id].Kind == Input {
			panic("aig: dominated-cone cut cannot contain inputs")
		}
		barriers[id] = true
	}
	before, after := 0, 0
	for _, id := range d.Nodes {
		for _, f := range s.fanins[id] {
			before += domDeref(s, f, barriers)
		}
	}
	for _, id := range d.Nodes {
		for _, f := range s.fanins[id] {
			after += domRef(s, f, barriers)
		}
	}
	if before != after {
		panic("aig: dominator reference restoration failed")
	}
	return before
}
func Aig_ObjDomCollectLos(s *Aig_Sto_t) []int {
	g := s.Graph
	n := domRegisterCount(g)
	ids := make([]int, 0, n)
	for _, e := range g.inputs[len(g.inputs)-n:] {
		id := e.NodeID()
		ids = append(ids, id)
		s.marked[id] = true
	}
	return ids
}
func domOutputDriver(s *Aig_Sto_t, index int) int { id := s.OutputIDs[index]; return s.fanins[id][0] }
func Aig_ObjPoLogicDeref(s *Aig_Sto_t) {
	g := s.Graph
	for i := 0; i < len(g.outputs)-domRegisterCount(g); i++ {
		Aig_ObjDomDeref_rec(s, domOutputDriver(s, i))
	}
}
func Aig_ObjPoLogicRef(s *Aig_Sto_t) {
	g := s.Graph
	for i := 0; i < len(g.outputs)-domRegisterCount(g); i++ {
		Aig_ObjDomRef_rec(s, domOutputDriver(s, i))
	}
}

// DomAnalysis reports the source diagnostic's boundary volume and released LOs.
type DomAnalysis struct {
	Cut           *Aig_Dom_t
	Volume, Domed int
}

func Aig_ObjDomFindGood(s *Aig_Sto_t) []DomAnalysis {
	cut := Aig_ObjDomCollectLos(s)
	doms := Aig_ObjDomCollect(s, cut)
	Aig_ObjPoLogicDeref(s)
	var result []DomAnalysis
	for _, d := range doms {
		result = append(result, DomAnalysis{newAigDom(d.Nodes), Aig_ObjDomVolume(s, d), Aig_ObjDomDomed(s, d)})
	}
	Aig_ObjPoLogicRef(s)
	for i := range s.marked {
		s.marked[i] = false
	}
	return result
}
func Aig_ManComputeDomsTest2(g *Graph, num int) []DomAnalysis {
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	s := Aig_ManComputeDomsFlops(g, num)
	defer Aig_ManDomStop(s)
	return Aig_ObjDomFindGood(s)
}
func Aig_ManComputeDomsTest(g *Graph) {
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	s := Aig_ManComputeDomsPis(g, 1)
	defer Aig_ManDomStop(s)
	Aig_ManDomPrint(s)
}
func Aig_ObjDomCount(s *Aig_Sto_t, e Edge) map[int]int {
	result := map[int]int{}
	for _, fan := range s.fanouts[e.NodeID()] {
		doms := s.Dominators[fan]
		for i, d := range doms {
			if i > 0 && len(d.Nodes) > 0 {
				s.Times[d.Nodes[0]]++
				result[d.Nodes[0]] = s.Times[d.Nodes[0]]
			}
		}
	}
	for id := range result {
		s.Times[id] = 0
	}
	return result
}
func Aig_ManComputeDomsForCofactoring(g *Graph) map[int]map[int]int {
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	s := Aig_ManComputeDomsNodes(g, 1)
	defer Aig_ManDomStop(s)
	result := map[int]map[int]int{}
	for id, n := range g.nodes {
		if g.deleted[id] || n.Kind != Input && n.Kind != And && n.Kind != ExclusiveOr || s.refs[id] < 10 {
			continue
		}
		counts := Aig_ObjDomCount(s, Edge(id)<<1)
		good := map[int]int{}
		for candidate, count := range counts {
			if count > 5 {
				good[candidate] = count
			}
		}
		if len(good) > 0 {
			result[id] = good
		}
	}
	return result
}

// DominatorSets returns deep copies suitable for independent sorting/inspection.
func (s *Aig_Sto_t) DominatorSets(id int) [][]int {
	sets := make([][]int, len(s.Dominators[id]))
	for i, d := range s.Dominators[id] {
		sets[i] = append([]int(nil), d.Nodes...)
	}
	sort.Slice(sets, func(i, j int) bool {
		a, b := sets[i], sets[j]
		for k := 0; k < len(a) && k < len(b); k++ {
			if a[k] != b[k] {
				return a[k] < b[k]
			}
		}
		return len(a) < len(b)
	})
	return sets
}
