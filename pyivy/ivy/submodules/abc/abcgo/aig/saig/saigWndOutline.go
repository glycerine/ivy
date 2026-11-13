package saig

import (
	"sort"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func windowFanouts(g *aig.Graph) map[aig.Edge][]aig.Edge {
	fanouts := make(map[aig.Edge][]aig.Edge)
	for _, e := range g.LiveEdges() {
		n := g.Node(e)
		count := 0
		switch n.Kind {
		case aig.And, aig.ExclusiveOr:
			count = 2
		case aig.Buffer, aig.Output:
			count = 1
		}
		for i := 0; i < count; i++ {
			f := aig.Aig_Regular(n.Fanins[i])
			fanouts[f] = append(fanouts[f], e)
		}
	}
	return fanouts
}

// Saig_ManWindowOutline marks and collects the window around pivot, crossing
// each latch boundary without consuming distance.
func Saig_ManWindowOutline(g *aig.Graph, pivot aig.Edge, distance int32) *vec.Vec_Ptr_t {
	if g == nil || !g.ValidEdge(pivot) || distance < 0 {
		panic("saig: invalid window pivot")
	}
	aig.Aig_ManIncrementTravId(g)
	fanouts := windowFanouts(g)
	dists := make([]int32, len(g.Nodes()))
	var nodes []aig.Edge
	var visit func(aig.Edge, int32)
	visit = func(e aig.Edge, d int32) {
		e = aig.Aig_Regular(e)
		if d == 0 || dists[e.NodeID()] >= d {
			return
		}
		dists[e.NodeID()] = d
		collected := aig.Aig_ObjIsTravIdCurrent(g, e)
		aig.Aig_ObjSetTravIdCurrent(g, e)
		if e == g.Const1() || Saig_ObjIsPo(g, e) {
			return
		}
		if Saig_ObjIsLi(g, e) {
			lo := Saig_ObjLiToLo(g, e)
			if !aig.Aig_ObjIsTravIdCurrent(g, lo) {
				visit(lo, d)
			}
			visit(aig.Aig_ObjFanin0(g, e), d-1)
			return
		}
		if !collected {
			nodes = append(nodes, e)
		}
		if Saig_ObjIsPi(g, e) {
			return
		}
		if Saig_ObjIsLo(g, e) {
			li := Saig_ObjLoToLi(g, e)
			if !aig.Aig_ObjIsTravIdCurrent(g, li) {
				visit(li, d)
			}
			for _, f := range fanouts[e] {
				visit(f, d-1)
			}
			return
		}
		n := g.Node(e)
		if n.Kind != aig.And && n.Kind != aig.ExclusiveOr && n.Kind != aig.Buffer {
			panic("saig: unsupported window object")
		}
		visit(n.Fanins[0], d-1)
		if n.Kind != aig.Buffer {
			visit(n.Fanins[1], d-1)
		}
		for _, f := range fanouts[e] {
			visit(f, d-1)
		}
	}
	visit(pivot, distance)
	sort.Slice(nodes, func(i, j int) bool { return nodes[i] < nodes[j] })
	result := vec.Vec_PtrAlloc(len(nodes))
	for _, e := range nodes {
		vec.Vec_PtrPush(result, e)
	}
	return result
}
func Saig_ObjHasUnlabeledFanout(g *aig.Graph, e aig.Edge) aig.Edge {
	for _, f := range windowFanouts(g)[aig.Aig_Regular(e)] {
		if Saig_ObjIsPo(g, f) || !aig.Aig_ObjIsTravIdCurrent(g, f) {
			return f
		}
	}
	return aig.InvalidEdge
}
func Saig_ManWindowCollectPis(g *aig.Graph, nodes *vec.Vec_Ptr_t) *vec.Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(0)
	seen := map[aig.Edge]bool{g.Const1(): true}
	add := func(e aig.Edge) {
		e = aig.Aig_Regular(e)
		if !seen[e] {
			seen[e] = true
			vec.Vec_PtrPush(result, e)
		}
	}
	for _, item := range vec.Vec_PtrArray(nodes) {
		e := item.(aig.Edge)
		if Saig_ObjIsPi(g, e) {
			add(e)
			continue
		}
		n := g.Node(e)
		var fanins []aig.Edge
		if Saig_ObjIsLo(g, e) {
			fanins = []aig.Edge{aig.Aig_ObjFanin0(g, Saig_ObjLoToLi(g, e))}
		} else if n.Kind == aig.Buffer {
			fanins = []aig.Edge{n.Fanins[0]}
		} else {
			fanins = n.Fanins[:]
		}
		for _, f := range fanins {
			f = aig.Aig_Regular(f)
			if !aig.Aig_ObjIsTravIdCurrent(g, f) {
				add(f)
			}
		}
	}
	return result
}
func Saig_ManWindowCollectPos(g *aig.Graph, nodes *vec.Vec_Ptr_t, pointers **vec.Vec_Ptr_t) *vec.Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(0)
	if pointers != nil {
		*pointers = vec.Vec_PtrAlloc(0)
	}
	fanouts := windowFanouts(g)
	for _, item := range vec.Vec_PtrArray(nodes) {
		e := item.(aig.Edge)
		for _, f := range fanouts[e] {
			if Saig_ObjIsPo(g, f) || !aig.Aig_ObjIsTravIdCurrent(g, f) {
				vec.Vec_PtrPush(result, e)
				if pointers != nil {
					vec.Vec_PtrPush(*pointers, f)
				}
				break
			}
		}
	}
	return result
}
