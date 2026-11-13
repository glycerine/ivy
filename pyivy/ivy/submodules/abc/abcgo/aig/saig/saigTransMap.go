package saig

import "github.com/glycerine/abcgo/aig/aig"

// frameMappings replaces the two C manager data pointers with owned, typed
// tables. Entries contain object IDs; -1 indicates no mapping.
type frameMappings struct {
	size     int
	one, two []int32
}

func frameMaps(g *aig.Graph, create bool) *frameMappings {
	if g == nil {
		panic("saig: nil graph")
	}
	if g.Data == nil && create {
		g.Data = &frameMappings{size: len(g.Nodes())}
	}
	maps, ok := g.Data.(*frameMappings)
	if !ok {
		panic("saig: graph data is already in use")
	}
	return maps
}
func newFrameTable(length int) []int32 {
	v := make([]int32, length)
	for i := range v {
		v[i] = -1
	}
	return v
}
func (m *frameMappings) offset(e aig.Edge, frame int32) int {
	if e.Complemented() || e.NodeID() >= m.size || frame < 0 {
		panic("saig: invalid frame map index")
	}
	return int(frame)*m.size + e.NodeID()
}
func Saig_ManStartMap1(g *aig.Graph, frames int32) {
	if frames <= 0 {
		panic("saig: invalid frame count")
	}
	m := frameMaps(g, true)
	if m.one != nil {
		panic("saig: frame map already exists")
	}
	m.one = newFrameTable(m.size * int(frames))
}
func Saig_ManStopMap1(g *aig.Graph) {
	m := frameMaps(g, false)
	if m.one == nil {
		panic("saig: missing frame map")
	}
	m.one = nil
	if m.two == nil {
		g.Data = nil
	}
}
func Saig_ManHasMap1(g *aig.Graph) int32 {
	if g == nil || g.Data == nil {
		return 0
	}
	m, ok := g.Data.(*frameMappings)
	if ok && m.one != nil {
		return 1
	}
	return 0
}
func Saig_ManSetMap1(g *aig.Graph, old aig.Edge, frame int32, new aig.Edge) {
	if new.Complemented() {
		panic("saig: complemented frame mapping")
	}
	m := frameMaps(g, false)
	m.one[m.offset(old, frame)] = int32(new.NodeID())
}
func Saig_ManGetMap1(g *aig.Graph, old aig.Edge, frame int32) int32 {
	m := frameMaps(g, false)
	return m.one[m.offset(old, frame)]
}
func Saig_ManStartMap2(g *aig.Graph, frames int32) {
	if frames <= 0 {
		panic("saig: invalid frame count")
	}
	m := frameMaps(g, true)
	if m.two != nil {
		panic("saig: frame map already exists")
	}
	m.two = newFrameTable(2 * m.size * int(frames))
}
func Saig_ManStopMap2(g *aig.Graph) {
	m := frameMaps(g, false)
	if m.two == nil {
		panic("saig: missing frame map")
	}
	m.two = nil
	if m.one == nil {
		g.Data = nil
	}
}
func Saig_ManHasMap2(g *aig.Graph) int32 {
	if g == nil || g.Data == nil {
		return 0
	}
	m, ok := g.Data.(*frameMappings)
	if ok && m.two != nil {
		return 1
	}
	return 0
}
func Saig_ManSetMap2(g *aig.Graph, old aig.Edge, frame int32, new aig.Edge, newFrame int32) {
	if new.Complemented() {
		panic("saig: complemented frame mapping")
	}
	m := frameMaps(g, false)
	i := 2 * m.offset(old, frame)
	m.two[i], m.two[i+1] = int32(new.NodeID()), newFrame
}
func Saig_ManGetMap2(g *aig.Graph, old aig.Edge, frame int32, newFrame *int32) int32 {
	m := frameMaps(g, false)
	i := 2 * m.offset(old, frame)
	*newFrame = m.two[i+1]
	return m.two[i]
}
