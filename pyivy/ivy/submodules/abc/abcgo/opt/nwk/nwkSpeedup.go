package nwk

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

func Aig_ManSpeedupNode_rec(g *Aig_Man_t, e Aig_Obj_t, v *Vec_Ptr_t) int32 {
	n := g.Node(e)
	if aig.Aig_ObjIsTravIdCurrent(g, e) {
		return 1
	}
	if n.Kind == aig.Input {
		return 0
	}
	aig.Aig_ObjSetTravIdCurrent(g, e)
	if n.Kind == aig.Constant {
		return 1
	}
	if Aig_ManSpeedupNode_rec(g, n.Fanins[0], v) == 0 || Aig_ManSpeedupNode_rec(g, n.Fanins[1], v) == 0 {
		return 0
	}
	vec.Vec_PtrPush(v, e)
	return 1
}
func Aig_ManSpeedupNode(p *Nwk_Man_t, g *Aig_Man_t, o *Nwk_Obj_t, leaves, times *Vec_Ptr_t) {
	leafNodes, timeNodes := vectorNodes(leaves), vectorNodes(times)
	seen := map[int]bool{}
	for _, n := range leafNodes {
		e := n.Copy.(aig.Edge)
		if seen[e.NodeID()] {
			return
		}
		seen[e.NodeID()] = true
	}
	if len(timeNodes) > 16 {
		panic("nwk: speedup cofactor limit exceeded")
	}
	boundary := map[int]aig.Edge{0: g.Const1()}
	for _, n := range leafNodes {
		e := n.Copy.(aig.Edge)
		boundary[e.NodeID()] = aig.Edge(e.NodeID() << 1)
	}
	root := o.Copy.(aig.Edge)
	var collect func(aig.Edge) bool
	nodes := []aig.Edge{}
	visited := map[int]bool{}
	collect = func(e aig.Edge) bool {
		if _, ok := boundary[e.NodeID()]; ok {
			return true
		}
		if visited[e.NodeID()] {
			return true
		}
		visited[e.NodeID()] = true
		n := g.Node(e)
		if n.Kind == aig.Input {
			return false
		}
		if !collect(n.Fanins[0]) || !collect(n.Fanins[1]) {
			return false
		}
		nodes = append(nodes, aig.Edge(e.NodeID()<<1))
		return true
	}
	if !collect(root) {
		return
	}
	cofactors := make([]aig.Edge, 1<<len(timeNodes))
	for assignment := range cofactors {
		mapping := map[int]aig.Edge{}
		for id, e := range boundary {
			mapping[id] = e
		}
		for k, n := range timeNodes {
			e := n.Copy.(aig.Edge)
			value := g.Const0()
			if assignment>>k&1 != 0 {
				value = g.Const1()
			}
			mapping[e.NodeID()] = value
		}
		translate := func(e aig.Edge) aig.Edge {
			r := mapping[e.NodeID()]
			if e.Complemented() {
				r = r.Not()
			}
			return r
		}
		for _, e := range nodes {
			n := g.Node(e)
			a, b := translate(n.Fanins[0]), translate(n.Fanins[1])
			r := g.AND(a, b)
			if n.Kind == aig.ExclusiveOr {
				r = g.XOR(a, b)
			}
			mapping[e.NodeID()] = r
		}
		cofactors[assignment] = translate(root)
	}
	for k, n := range timeNodes {
		step := 1 << k
		control := aig.Edge(n.Copy.(aig.Edge).NodeID() << 1)
		for i := 0; i < len(cofactors); i += 2 * step {
			cofactors[i] = g.Mux(control, cofactors[i+step], cofactors[i])
		}
	}
	representative := aig.Edge(root.NodeID() << 1)
	alternative := aig.Edge(cofactors[0].NodeID() << 1)
	if representative == alternative {
		return
	}
	if aig.Aig_ObjEquiv(g, representative) == aig.InvalidEdge && aig.Aig_ObjEquiv(g, alternative) == aig.InvalidEdge && aig.Aig_ObjCheckTfi(g, alternative, representative) == 0 {
		aig.Aig_ObjSetEquiv(g, representative, alternative)
	}
}
func Nwk_ManDelayTraceTCEdges(p *Nwk_Man_t, o *Nwk_Obj_t, delta float32, useLibrary int32) uint32 {
	var result uint32
	for pin, index := range pinAssignment(o, true) {
		delay := float32(1)
		if useLibrary != 0 {
			delay = pinDelay(o, pin)
		}
		if o.Required < o.Fanins[index].Arrival+delay+delta {
			result |= 1 << uint(index)
		}
	}
	return result
}
func Nwk_ManSpeedup(p *Nwk_Man_t, useLibrary, percentage, degree, verbose, veryVerbose int32) *Aig_Man_t {
	oldLib, oldTiming := p.Library, p.Timing
	if useLibrary == 0 {
		p.Library = nil
		if p.Timing != nil {
			p.Timing = tim.Tim_ManDup(p.Timing, 1)
		}
	}
	defer func() { p.Library = oldLib; p.Timing = oldTiming }()
	arrival := Nwk_ManDelayTraceLut(p)
	delta := float32(1)
	if useLibrary != 0 {
		delta = arrival * float32(percentage) / 100
	}
	critical := map[*Nwk_Obj_t]uint32{}
	for _, o := range p.Objects {
		if o != nil && o.Type == NWK_OBJ_NODE && o.Slack < delta {
			critical[o] = Nwk_ManDelayTraceTCEdges(p, o, delta, useLibrary)
		}
	}
	g := Nwk_ManStrash(p)
	worked := 0
	for _, o := range p.Objects {
		if o == nil || o.Type != NWK_OBJ_NODE || o.Slack >= delta {
			continue
		}
		times, leaves := []*Nwk_Obj_t{}, []*Nwk_Obj_t{}
		push := func(a []*Nwk_Obj_t, n *Nwk_Obj_t) []*Nwk_Obj_t {
			if findNode(a, n) < 0 {
				return append(a, n)
			}
			return a
		}
		for i, f := range o.Fanins {
			if f.Type == NWK_OBJ_CI {
				leaves = push(leaves, f)
				continue
			}
			for k, n := range f.Fanins {
				leaves = push(leaves, n)
				if critical[o]>>i&1 != 0 && critical[f]>>k&1 != 0 {
					times = push(times, n)
				}
			}
		}
		if len(times) == 0 || len(times) > int(degree) {
			continue
		}
		sort.SliceStable(times, func(i, j int) bool { return times[i].Slack > times[j].Slack })
		Aig_ManSpeedupNode(p, g, o, nodesVector(leaves), nodesVector(times))
		worked++
	}
	for _, e := range g.LiveEdges() {
		alternative := aig.Aig_ObjEquiv(g, e)
		if alternative != aig.InvalidEdge && g.Node(alternative).Refs > 0 {
			aig.Aig_ObjSetEquiv(g, e, aig.InvalidEdge)
		}
	}
	if verbose != 0 {
		fmt.Fprintf(writer(p), "Speedup: %d candidate nodes.\n", worked)
	}
	g = aig.Aig_ManDupDfs(g)
	aig.Aig_ManChoiceLevel(g)
	return g
}
