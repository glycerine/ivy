package reo

import (
	"fmt"
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"sort"
)

func cloneProfile(p Profile) Profile {
	p.LevelNodes = append([]int(nil), p.LevelNodes...)
	p.LevelWidth = append([]int(nil), p.LevelWidth...)
	p.LevelAPL = append([]float64(nil), p.LevelAPL...)
	return p
}
func cloneStatistics(s Statistics) Statistics {
	s.Before = cloneProfile(s.Before)
	s.After = cloneProfile(s.After)
	return s
}
func profileRoots(roots []*DdNode, slots []int) Profile {
	n := len(slots)
	result := Profile{LevelNodes: make([]int, n+1), LevelWidth: make([]int, n+1), LevelAPL: make([]float64, n+1)}
	levels := map[int]int{}
	for i, v := range slots {
		levels[v] = i
	}
	nodeLevel := func(node *DdNode) int {
		if cudd.Cudd_IsConstant(node) {
			return n
		}
		return levels[int(node.Index)]
	}
	top := map[*DdNode]int{}
	var visit func(*DdNode, int)
	visit = func(edge *DdNode, start int) {
		node := cudd.Cudd_Regular(edge)
		old, seen := top[node]
		if seen {
			if start < old {
				top[node] = start
			}
			return
		}
		top[node] = start
		if !cudd.Cudd_IsConstant(node) {
			level := nodeLevel(node)
			visit(node.E, level+1)
			visit(node.T, level+1)
		}
	}
	for _, root := range roots {
		visit(root, 0)
	}
	nodes := make([]*DdNode, 0, len(top))
	for node, first := range top {
		level := nodeLevel(node)
		result.LevelNodes[level]++
		result.Nodes++
		for i := first; i <= level; i++ {
			result.LevelWidth[i]++
			result.Width++
		}
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodeLevel(nodes[i]) < nodeLevel(nodes[j]) })
	weights := map[*DdNode]float64{}
	for _, root := range roots {
		weights[cudd.Cudd_Regular(root)]++
	}
	for _, node := range nodes {
		if cudd.Cudd_IsConstant(node) {
			continue
		}
		weight := weights[node]
		result.APL += weight
		result.LevelAPL[nodeLevel(node)] += weight
		weights[cudd.Cudd_Regular(node.E)] += weight / 2
		weights[cudd.Cudd_Regular(node.T)] += weight / 2
	}
	return result
}
func reoProfileNodesStart(p *Manager)  { p.current = profileRoots(p.roots, p.slots) }
func reoProfileAplStart(p *Manager)    { reoProfileNodesStart(p) }
func reoProfileWidthStart(p *Manager)  { reoProfileNodesStart(p) }
func reoProfileWidthStart2(p *Manager) { reoProfileNodesStart(p) }
func reoProfileAplPrint(p *Manager)    { p.printProfile("APL", p.current) }
func reoProfileNodesPrint(p *Manager)  { p.printProfile("NODES", p.current) }
func reoProfileWidthPrint(p *Manager)  { p.printProfile("WIDTH", p.current) }
func reoProfileWidthVerifyLevel(plane *reo_plane, level int) {
	for u := plane.pHead; u != nil; u = u.Next {
		if int(u.TopRef) > level {
			panic(fmt.Sprintf("reo: invalid top reference at level %d", level))
		}
	}
}
