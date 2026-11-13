package dch

import (
	"encoding/binary"
	"sort"

	aig "github.com/glycerine/abcgo/aig/aig"
)

// Classes only split: simulation proposes candidates, never proves an
// equivalence. Members and representatives are sorted by stable object ID.
type Classes struct {
	groups [][]aig.Edge
	byNode map[int]int
}
type Dch_Cla_t = Classes

func (c *Classes) index() {
	c.byNode = map[int]int{}
	for i, group := range c.groups {
		for _, node := range group {
			c.byNode[node.NodeID()] = i
		}
	}
}
func (c *Classes) representative(node aig.Edge) aig.Edge {
	i, ok := c.byNode[node.NodeID()]
	if !ok || len(c.groups[i]) < 2 || c.groups[i][0] == node {
		return aig.InvalidEdge
	}
	return c.groups[i][0]
}
func (c *Classes) remove(node aig.Edge) {
	i, ok := c.byNode[node.NodeID()]
	if !ok {
		return
	}
	group := c.groups[i]
	for j, n := range group {
		if n == node {
			c.groups[i] = append(group[:j], group[j+1:]...)
			break
		}
	}
	c.index()
}
func signature(g *aig.Graph, node aig.Edge, sims [][]uint32) string {
	row := sims[node.NodeID()]
	bytes := make([]byte, 4*len(row))
	phase := aig.Aig_ObjPhase(g, node)
	for i, value := range row {
		if phase {
			value = ^value
		}
		binary.LittleEndian.PutUint32(bytes[4*i:], value)
	}
	return string(bytes)
}
func createClasses(g *aig.Graph, order []aig.Edge, words int, sources ...*aig.AigRandomSource) *Classes {
	candidates := []aig.Edge{}
	for _, node := range g.LiveEdges() {
		kind := g.Node(node).Kind
		if kind == aig.Constant || kind == aig.Input || kind == aig.And || kind == aig.ExclusiveOr {
			candidates = append(candidates, node)
		}
	}
	c := &Classes{groups: [][]aig.Edge{candidates}}
	var random *aig.AigRandomSource
	if len(sources) > 0 {
		random = sources[0]
	} else {
		random = aig.NewAigRandomSource()
	}
	random.Uint32() // Aig_ManRandom(1) reset draw is discarded by dchCore.c.
	for pass := 0; pass < 8; pass++ {
		sims := simulateWords(g, order, words, random)
		refined := [][]aig.Edge{}
		for _, group := range c.groups {
			// Dch_ClassesRefine only refines ordinary classes. The constant
			// candidate group is refined later by SAT counterexamples.
			if pass > 0 && len(group) > 0 && group[0].NodeID() == 0 {
				refined = append(refined, group)
				continue
			}
			if len(group) < 2 {
				refined = append(refined, group)
				continue
			}
			buckets := map[string][]aig.Edge{}
			keys := []string{}
			for _, node := range group {
				key := signature(g, node, sims)
				if _, exists := buckets[key]; !exists {
					keys = append(keys, key)
				}
				buckets[key] = append(buckets[key], node)
			}
			for _, key := range keys {
				refined = append(refined, buckets[key])
			}
		}
		c.groups = refined
	}
	for _, group := range c.groups {
		sort.Slice(group, func(i, j int) bool { return group[i].NodeID() < group[j].NodeID() })
	}
	c.index()
	return c
}
func Dch_ClassesLitNum(c *Classes) int {
	n := 0
	for _, g := range c.groups {
		// Constant candidates have their own counter in Dch_Cla_t and are
		// not included in the source's count of ordinary class literals.
		if len(g) > 1 && g[0].NodeID() != 0 {
			n += len(g) - 1
		}
	}
	return n
}
