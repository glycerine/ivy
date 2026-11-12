package bsat

import "slices"

/**CFile****************************************************************

  FileName    [satProof.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Proof manipulation procedures.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satProof.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

// ProofGraph stores the original resolution DAG as typed records. Handles are
// word offsets, with two low antecedent bits reserved for root/pivot tags.
type ProofGraph struct {
	nodes map[int32]*satset
	order []int32
	next  int32
}
type Vec_Set_t = ProofGraph // Legacy proof-only arena spelling used by solver2.
type satset struct {
	learnt, mark, partA uint32
	nEnts               int32
	Id                  int32
	pEnts               []int32
}

func NewProofGraph() *ProofGraph { return &ProofGraph{nodes: make(map[int32]*satset), next: 2} }
func (g *ProofGraph) Append(entries []int32) int32 {
	if g.nodes == nil {
		g.nodes = make(map[int32]*satset)
		g.next = 2
	}
	words := Proof_NodeWordNum(int32(len(entries)))
	if g.next > 0x1fffffff-words {
		panic("proof handle overflow")
	}
	for _, entry := range entries {
		if entry < 0 {
			panic("negative proof antecedent")
		}
		if entry&1 == 0 {
			if _, ok := g.nodes[entry>>2]; !ok {
				panic("proof references missing or future node")
			}
		}
	}
	h := g.next
	g.next += words
	g.nodes[h] = &satset{nEnts: int32(len(entries)), pEnts: append([]int32(nil), entries...)}
	g.order = append(g.order, h)
	return h
}
func Proof_NodeRead(g *ProofGraph, h int32) *satset {
	node, ok := g.nodes[h]
	if !ok || h <= 0 {
		panic("invalid proof node handle")
	}
	return node
}
func Proof_NodeWordNum(n int32) int32 {
	if n <= 0 || n > 0x1fffffff {
		panic("invalid proof antecedent count")
	}
	return 1 + (n+1)/2
}
func Proof_ClauseSetEnts(g *ProofGraph, h, n int32) {
	node := Proof_NodeRead(g, h)
	if n <= 0 || int(n) > len(node.pEnts) {
		panic("invalid proof node shrink")
	}
	node.pEnts = node.pEnts[:n]
	node.nEnts = n
}
func Proof_CleanCollected(g *ProofGraph, used []int32) {
	for _, h := range used {
		Proof_NodeRead(g, h).Id = 0
	}
}
func Proof_CollectUsed_iter(g *ProofGraph, h int32, used, stack *[]int32) {
	node := Proof_NodeRead(g, h)
	if node.Id != 0 {
		return
	}
	node.Id = 1
	*stack = append(*stack, h<<1)
	for len(*stack) > 0 {
		item := (*stack)[len(*stack)-1]
		*stack = (*stack)[:len(*stack)-1]
		if item&1 != 0 {
			*used = append(*used, item>>1)
			continue
		}
		*stack = append(*stack, item^1)
		node = Proof_NodeRead(g, item>>1)
		for _, entry := range node.pEnts {
			if entry&1 != 0 {
				continue
			}
			child := Proof_NodeRead(g, entry>>2)
			if child.Id == 0 {
				child.Id = 1
				*stack = append(*stack, (entry>>2)<<1)
			}
		}
	}
}

// ABC sorts this result regardless of its historical fSort parameter. Allocation
// order is topological, so sorting yields a valid compaction traversal.
func Proof_CollectUsedIter(g *ProofGraph, roots []int32, _ int32) []int32 {
	var used, stack []int32
	for _, h := range roots {
		if h >= 0 {
			Proof_CollectUsed_iter(g, h, &used, &stack)
		}
	}
	slices.Sort(used)
	return used
}
func Proof_CollectUsed_rec(g *ProofGraph, h int32, used *[]int32) {
	node := Proof_NodeRead(g, h)
	if node.Id != 0 {
		return
	}
	node.Id = 1
	for _, entry := range node.pEnts {
		if entry&1 == 0 {
			Proof_CollectUsed_rec(g, entry>>2, used)
		}
	}
	*used = append(*used, h)
}
func Proof_CollectUsedRec(g *ProofGraph, roots []int32) []int32 {
	var used []int32
	for _, h := range roots {
		if h >= 0 {
			Proof_CollectUsed_rec(g, h, &used)
		}
	}
	return used
}
func Proof_MarkUsed_rec(g *ProofGraph, h int32) int32 {
	var used, stack []int32
	Proof_CollectUsed_iter(g, h, &used, &stack)
	return int32(len(used))
}
func Proof_MarkUsedRec(g *ProofGraph, roots []int32) int32 {
	var count int32
	for _, h := range roots {
		if h >= 0 {
			count += Proof_MarkUsed_rec(g, h)
		}
	}
	return count
}
func Sat_ProofCheck0(g *ProofGraph) {
	for _, node := range g.nodes {
		if node.Id != 0 {
			panic("unclean proof traversal marks")
		}
	}
}
func proofRoots(data any) []int32 {
	switch roots := data.(type) {
	case []int32:
		return roots
	case *[]int32:
		return *roots
	case *veci:
		return veci_begin(roots)
	default:
		panic("invalid proof roots")
	}
}
func Sat_ProofReduce(g *ProofGraph, data any, pivot int32) int32 {
	if pivot < 1 || pivot > g.next {
		panic("invalid proof bookmark")
	}
	roots := proofRoots(data)
	for _, node := range g.nodes {
		node.Id = 0
	}
	used := Proof_CollectUsedIter(g, roots, 1)
	mapping := make(map[int32]int32, len(used))
	next := int32(2)
	newPivot := int32(0)
	for _, h := range used {
		mapping[h] = next
		if newPivot == 0 && h >= pivot {
			newPivot = next
		}
		next += Proof_NodeWordNum(Proof_NodeRead(g, h).nEnts)
	}
	if newPivot == 0 {
		newPivot = next
	}
	nodes := make(map[int32]*satset, len(used))
	order := make([]int32, 0, len(used))
	for _, h := range used {
		node := Proof_NodeRead(g, h)
		for i, entry := range node.pEnts {
			if entry&1 == 0 {
				remapped, ok := mapping[entry>>2]
				if !ok {
					panic("reachable proof lost antecedent")
				}
				node.pEnts[i] = remapped<<2 | (entry & 2)
			}
		}
		node.Id = 0
		nodes[mapping[h]] = node
		order = append(order, mapping[h])
	}
	for i, h := range roots {
		if h >= 0 {
			roots[i] = mapping[h]
		}
	}
	g.nodes = nodes
	g.order = order
	g.next = next
	return newPivot
}
func Sat_ProofCollectCore(g *ProofGraph, used []int32) []int32 {
	seen := map[int32]bool{}
	var core []int32
	for _, h := range used {
		for _, entry := range Proof_NodeRead(g, h).pEnts {
			if entry&1 != 0 {
				id := entry >> 2
				if !seen[id] {
					seen[id] = true
					core = append(core, id)
				}
			}
		}
	}
	return core
}
func Proof_DeriveCore(g *ProofGraph, root int32) []int32 {
	if root == -1 {
		return nil
	}
	used := Proof_CollectUsedIter(g, []int32{root}, 0)
	core := Sat_ProofCollectCore(g, used)
	Proof_CleanCollected(g, used)
	slices.Sort(core)
	slices.Reverse(core)
	return core
}
