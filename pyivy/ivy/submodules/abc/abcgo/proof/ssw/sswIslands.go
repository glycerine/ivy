// Go port of sswIslands.c by Alan Mishchenko, UC Berkeley (2008).
package ssw

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
)

func sswIslandMatch(g *aig.Graph, e aig.Edge) (aig.Edge, bool) {
	match, ok := g.ObjectData(e).(aig.Edge)
	return match, ok
}

func Ssw_CreatePair(p0, p1 *aig.Graph, pairs *vec.Vec_Int_t, a, b aig.Edge) {
	p0.SetObjectData(a, b)
	p1.SetObjectData(b, a)
	vec.Vec_IntPushTwo(pairs, int32(a.NodeID()), int32(b.NodeID()))
}

func Ssw_MatchingStart(p0, p1 *aig.Graph, pairs *vec.Vec_Int_t) {
	aig.Aig_ManCleanData(p0)
	aig.Aig_ManCleanData(p1)
	entries := vec.Vec_IntArray(pairs)
	if len(entries)%2 != 0 {
		panic("ssw: odd number of matching IDs")
	}
	for i := 0; i < len(entries); i += 2 {
		a, b := aig.Aig_ManObj(p0, entries[i]), aig.Aig_ManObj(p1, entries[i+1])
		if _, ok := sswIslandMatch(p0, a); ok {
			panic("ssw: duplicate left match")
		}
		if _, ok := sswIslandMatch(p1, b); ok {
			panic("ssw: duplicate right match")
		}
		p0.SetObjectData(a, b)
		p1.SetObjectData(b, a)
	}
	check := func(a, b aig.Edge) {
		x, ok := sswIslandMatch(p0, a)
		y, ok2 := sswIslandMatch(p1, b)
		if !ok || !ok2 || x != b || y != a {
			panic("ssw: missing corresponding constant or PI")
		}
	}
	check(p0.Const1(), p1.Const1())
	pi := int(saig.Saig_ManPiNum(p0))
	if pi != int(saig.Saig_ManPiNum(p1)) {
		panic("ssw: PI count mismatch")
	}
	for i := 0; i < pi; i++ {
		check(p0.Inputs()[i], p1.Inputs()[i])
	}
	for _, e := range p0.Outputs() {
		if _, ok := sswIslandMatch(p0, e); ok {
			panic("ssw: output was matched")
		}
	}
	for _, e := range p1.Outputs() {
		if _, ok := sswIslandMatch(p1, e); ok {
			panic("ssw: output was matched")
		}
	}
	for _, e := range p0.Inputs()[pi:] {
		if other, ok := sswIslandMatch(p0, e); ok && !saig.Saig_ObjIsLo(p1, other) {
			fmt.Println("Mismatch between LO pairs.")
		}
	}
	for _, e := range p1.Inputs()[pi:] {
		if other, ok := sswIslandMatch(p1, e); ok && !saig.Saig_ObjIsLo(p0, other) {
			fmt.Println("Mismatch between LO pairs.")
		}
	}
}

// Ssw_MatchingExtendOne collects matched neighbors of the unmatched region,
// following register-input/output connections as a single edge.
func Ssw_MatchingExtendOne(g *aig.Graph, out *vec.Vec_Ptr_t) {
	vec.Vec_PtrClear(out)
	nodes, inputs, outputs := g.Nodes(), g.Inputs(), g.Outputs()
	pi, po := int(saig.Saig_ManPiNum(g)), int(saig.Saig_ManPoNum(g))
	fanouts := make([][]aig.Edge, len(nodes))
	for id, n := range nodes {
		e := aig.Edge(id) << 1
		if aig.Aig_ObjIsNone(g, e) {
			continue
		}
		switch n.Kind {
		case aig.And, aig.ExclusiveOr:
			for _, fan := range n.Fanins {
				fanouts[fan.NodeID()] = append(fanouts[fan.NodeID()], e)
			}
		case aig.Buffer, aig.Output:
			fanouts[n.Fanins[0].NodeID()] = append(fanouts[n.Fanins[0].NodeID()], e)
		}
	}
	seen := make(map[int]bool)
	push := func(e aig.Edge, noConst bool) {
		e &= ^aig.Edge(1)
		if noConst && e.NodeID() == 0 {
			return
		}
		if _, ok := sswIslandMatch(g, e); ok && !seen[e.NodeID()] {
			seen[e.NodeID()] = true
			vec.Vec_PtrPush(out, e)
		}
	}
	for id, n := range nodes {
		e := aig.Edge(id) << 1
		if aig.Aig_ObjIsNone(g, e) || n.Kind != aig.Input && n.Kind != aig.And && n.Kind != aig.ExclusiveOr && n.Kind != aig.Buffer {
			continue
		}
		if _, ok := sswIslandMatch(g, e); ok {
			continue
		}
		if saig.Saig_ObjIsLo(g, e) {
			for i, input := range inputs[pi:] {
				if input == e {
					push(g.Node(outputs[po+i]).Fanins[0], true)
					break
				}
			}
		}
		if n.Kind == aig.And || n.Kind == aig.ExclusiveOr {
			push(n.Fanins[0], false)
			push(n.Fanins[1], false)
		} else if n.Kind == aig.Buffer {
			push(n.Fanins[0], false)
		}
		for _, fan := range fanouts[id] {
			if saig.Saig_ObjIsPo(g, fan) {
				continue
			}
			if saig.Saig_ObjIsLi(g, fan) {
				for i, output := range outputs[po:] {
					if output == fan {
						fan = inputs[pi+i]
						break
					}
				}
			}
			push(fan, false)
		}
	}
}

func Ssw_MatchingCountUnmached(g *aig.Graph) int32 {
	var count int32
	for id, n := range g.Nodes() {
		e := aig.Edge(id) << 1
		if aig.Aig_ObjIsNone(g, e) {
			continue
		}
		if n.Kind == aig.Input || n.Kind == aig.And || n.Kind == aig.ExclusiveOr || n.Kind == aig.Buffer {
			if _, ok := sswIslandMatch(g, e); !ok {
				count++
			}
		}
	}
	return count
}

func Ssw_MatchingExtend(p0, p1 *aig.Graph, distance, verbose int32) {
	if verbose != 0 {
		fmt.Printf("Extending islands by %d steps:\n", distance)
	}
	a, b := vec.Vec_PtrAlloc(1000), vec.Vec_PtrAlloc(1000)
	for d := int32(0); d < distance; d++ {
		Ssw_MatchingExtendOne(p0, a)
		Ssw_MatchingExtendOne(p1, b)
		for i, list := range []*vec.Vec_Ptr_t{a, b} {
			from, to := p0, p1
			if i == 1 {
				from, to = p1, p0
			}
			for _, item := range vec.Vec_PtrArray(list) {
				e := item.(aig.Edge)
				other, ok := sswIslandMatch(from, e)
				if !ok || saig.Saig_ObjIsPi(to, other) {
					continue
				}
				back, ok := sswIslandMatch(to, other)
				if !ok || back != e {
					panic("ssw: asymmetric match")
				}
				from.SetObjectData(e, nil)
				to.SetObjectData(other, nil)
			}
		}
	}
}

func Ssw_MatchingComplete(p0, p1 *aig.Graph) {
	inputs, outputs := p0.Inputs(), p0.Outputs()
	pi, po := int(saig.Saig_ManPiNum(p0)), int(saig.Saig_ManPoNum(p0))
	var added []aig.Edge
	for i, e := range inputs[pi:] {
		if _, ok := sswIslandMatch(p0, e); !ok {
			newInput := p1.NewInput()
			p0.SetObjectData(e, newInput)
			p1.SetObjectData(newInput, e)
			added = append(added, outputs[po+i])
		}
	}
	copyLit := func(e aig.Edge) aig.Edge {
		v, ok := sswIslandMatch(p0, e)
		if !ok {
			panic("ssw: unmapped fanin")
		}
		return v ^ (e & 1)
	}
	for _, e := range aig.Aig_ManDfsAll(p0) {
		n := p0.Node(e)
		if n.Kind != aig.And && n.Kind != aig.ExclusiveOr && n.Kind != aig.Buffer {
			continue
		}
		if _, ok := sswIslandMatch(p0, e); ok {
			continue
		}
		var v aig.Edge
		switch n.Kind {
		case aig.And:
			v = p1.AND(copyLit(n.Fanins[0]), copyLit(n.Fanins[1]))
		case aig.ExclusiveOr:
			v = p1.XOR(copyLit(n.Fanins[0]), copyLit(n.Fanins[1]))
		case aig.Buffer:
			v = copyLit(n.Fanins[0])
		}
		p0.SetObjectData(e, v)
		p1.SetObjectData(v, e)
	}
	for _, e := range added {
		p1.AddOutput(copyLit(p0.Node(e).Fanins[0]))
	}
	aig.Aig_ManSetRegNum(p1, aig.Aig_ManRegNum(p1)+int32(len(added)))
}

func Ssw_MatchingPairs(p0, p1 *aig.Graph) *vec.Vec_Int_t {
	if p0.InputCount() != p1.InputCount() || p0.OutputCount() != p1.OutputCount() || aig.Aig_ManRegNum(p0) != aig.Aig_ManRegNum(p1) || p0.NodeCount() != p1.NodeCount() {
		panic("ssw: incomplete matching")
	}
	pairs := vec.Vec_IntAlloc(2 * p0.NodeCount())
	for id, n := range p0.Nodes() {
		e := aig.Edge(id) << 1
		if aig.Aig_ObjIsNone(p0, e) || n.Kind == aig.Output {
			continue
		}
		other, ok := sswIslandMatch(p0, e)
		if !ok {
			panic("ssw: missing match")
		}
		vec.Vec_IntPushTwo(pairs, int32(id), int32(other.NodeID()))
	}
	return pairs
}

func Ssw_MatchingMiter(miter, p0, p1 *aig.Graph, pairs *vec.Vec_Int_t) *vec.Vec_Int_t {
	result := vec.Vec_IntAlloc(vec.Vec_IntSize(pairs))
	entries := vec.Vec_IntArray(pairs)
	for i := 0; i < len(entries); i += 2 {
		a, b := aig.Aig_ManObj(p0, entries[i]), aig.Aig_ManObj(p1, entries[i+1])
		x, ok := sswIslandMatch(p0, a)
		y, ok2 := sswIslandMatch(p1, b)
		if !ok || !ok2 {
			panic("ssw: missing miter copy")
		}
		if x == y || aig.Aig_ObjIsNone(miter, x) || aig.Aig_ObjIsNone(miter, y) {
			continue
		}
		if x.Complemented() || y.Complemented() || x.NodeID() >= y.NodeID() {
			panic("ssw: invalid miter mapping")
		}
		vec.Vec_IntPushTwo(result, int32(x.NodeID()), int32(y.NodeID()))
	}
	return result
}

func Saig_StrSimPerformMatching_hack(p0, p1 *aig.Graph) *vec.Vec_Int_t {
	pairs := vec.Vec_IntAlloc(100)
	for id, n := range p0.Nodes() {
		e := aig.Edge(id) << 1
		if !aig.Aig_ObjIsNone(p0, e) && n.Kind != aig.Output {
			vec.Vec_IntPushTwo(pairs, int32(id), int32(id))
		}
	}
	return pairs
}
