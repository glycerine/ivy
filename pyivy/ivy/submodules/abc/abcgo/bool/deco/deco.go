// Package deco exposes the standalone decomposition inline API using the
// authoritative bool/dec graph implementation.
package deco

import (
	"github.com/glycerine/abcgo/bool/dec"
	"iter"
)

type Dec_Edge_t = dec.Dec_Edge_t
type Dec_Node_t = dec.Dec_Node_t
type Dec_Graph_t = dec.Dec_Graph_t

// Dec_Man_t carries factoring scratch and NPN metadata present in deco.h.
// The header defines no manager methods; callers own these Go buffers.
type Dec_Man_t struct {
	FactoringMemory               any
	Cubes, Literals               []int32
	CanonicalTruths               []uint16
	Phases, Permutations, Classes []byte
}

// C bitfield assignment keeps exactly 30 index bits and one phase bit.
func Dec_EdgeCreate(node, compl int32) Dec_Edge_t {
	return dec.Dec_EdgeCreate(int32(uint32(node)&0x3fffffff), compl&1)
}
func Dec_EdgeToInt(e Dec_Edge_t) uint32                      { return (e.Node&0x3fffffff)<<1 | dec.Dec_EdgeToInt(e)&1 }
func Dec_IntToEdge(e uint32) Dec_Edge_t                      { return Dec_EdgeCreate(int32(e>>1), int32(e&1)) }
func Dec_EdgeToInt_(e Dec_Edge_t) uint32                     { return Dec_EdgeToInt(e) }
func Dec_IntToEdge_(e uint32) Dec_Edge_t                     { return Dec_IntToEdge(e) }
func Dec_GraphCreate(leaves int32) *Dec_Graph_t              { return dec.Dec_GraphCreate(leaves) }
func Dec_Factor(sop string) *Dec_Graph_t                    { return dec.Dec_Factor(sop) }
func Dec_GraphCreateConst0() *Dec_Graph_t                    { return dec.Dec_GraphCreateConst0() }
func Dec_GraphCreateConst1() *Dec_Graph_t                    { return dec.Dec_GraphCreateConst1() }
func Dec_GraphCreateLeaf(i, n, c int32) *Dec_Graph_t         { return dec.Dec_GraphCreateLeaf(i, n, c) }
func Dec_GraphFree(g *Dec_Graph_t)                           { dec.Dec_GraphFree(g) }
func Dec_GraphIsConst(g *Dec_Graph_t) int32                  { return dec.Dec_GraphIsConst(g) }
func Dec_GraphIsConst0(g *Dec_Graph_t) int32                 { return dec.Dec_GraphIsConst0(g) }
func Dec_GraphIsConst1(g *Dec_Graph_t) int32                 { return dec.Dec_GraphIsConst1(g) }
func Dec_GraphIsComplement(g *Dec_Graph_t) int32             { return dec.Dec_GraphIsComplement(g) }
func Dec_GraphComplement(g *Dec_Graph_t)                     { dec.Dec_GraphComplement(g) }
func Dec_GraphLeaveNum(g *Dec_Graph_t) int32                 { return dec.Dec_GraphLeaveNum(g) }
func Dec_GraphNodeNum(g *Dec_Graph_t) int32                  { return dec.Dec_GraphNodeNum(g) }
func Dec_GraphNode(g *Dec_Graph_t, i int32) *Dec_Node_t      { return dec.Dec_GraphNode(g, i) }
func Dec_GraphNodeLast(g *Dec_Graph_t) *Dec_Node_t           { return dec.Dec_GraphNodeLast(g) }
func Dec_GraphNodeInt(g *Dec_Graph_t, n *Dec_Node_t) int32   { return dec.Dec_GraphNodeInt(g, n) }
func Dec_GraphIsVar(g *Dec_Graph_t) int32                    { return dec.Dec_GraphIsVar(g) }
func Dec_GraphNodeIsVar(g *Dec_Graph_t, n *Dec_Node_t) int32 { return dec.Dec_GraphNodeIsVar(g, n) }
func Dec_GraphVar(g *Dec_Graph_t) *Dec_Node_t {
	if Dec_GraphIsVar(g) == 0 {
		panic("deco: graph root is not a variable")
	}
	return dec.Dec_GraphVar(g)
}
func Dec_GraphVarInt(g *Dec_Graph_t) int32 {
	if Dec_GraphIsVar(g) == 0 {
		panic("deco: graph root is not a variable")
	}
	return dec.Dec_GraphVarInt(g)
}
func Dec_GraphSetRoot(g *Dec_Graph_t, e Dec_Edge_t)  { dec.Dec_GraphSetRoot(g, normalizeEdge(e)) }
func Dec_GraphAppendNode(g *Dec_Graph_t) *Dec_Node_t { return dec.Dec_GraphAppendNode(g) }
func Dec_GraphAddNodeAnd(g *Dec_Graph_t, a, b Dec_Edge_t) Dec_Edge_t {
	return dec.Dec_GraphAddNodeAnd(g, normalizeEdge(a), normalizeEdge(b))
}
func Dec_GraphAddNodeOr(g *Dec_Graph_t, a, b Dec_Edge_t) Dec_Edge_t {
	return dec.Dec_GraphAddNodeOr(g, normalizeEdge(a), normalizeEdge(b))
}
func Dec_GraphAddNodeXor(g *Dec_Graph_t, a, b Dec_Edge_t, kind int32) Dec_Edge_t {
	return dec.Dec_GraphAddNodeXor(g, normalizeEdge(a), normalizeEdge(b), kind)
}
func Dec_GraphAddNodeMux(g *Dec_Graph_t, c, t, e Dec_Edge_t, kind int32) Dec_Edge_t {
	return dec.Dec_GraphAddNodeMux(g, normalizeEdge(c), normalizeEdge(t), normalizeEdge(e), kind)
}

// Header iterator macros become range-compatible sequences, preserving indices
// and early loop termination without hiding loop bodies in callbacks.
func Dec_GraphForEachLeaf(g *Dec_Graph_t) iter.Seq2[int32, *Dec_Node_t] {
	return func(yield func(int32, *Dec_Node_t) bool) {
		for i := int32(0); i < Dec_GraphLeaveNum(g); i++ {
			if !yield(i, Dec_GraphNode(g, i)) {
				return
			}
		}
	}
}
func Dec_GraphForEachNode(g *Dec_Graph_t) iter.Seq2[int32, *Dec_Node_t] {
	return func(yield func(int32, *Dec_Node_t) bool) {
		for i := Dec_GraphLeaveNum(g); i < Dec_GraphLeaveNum(g)+Dec_GraphNodeNum(g); i++ {
			if !yield(i, Dec_GraphNode(g, i)) {
				return
			}
		}
	}
}
func Dec_NodeSetLevel(n *Dec_Node_t, level uint32) { n.SetLevel(level) }
func Dec_NodeLevel(n *Dec_Node_t) uint32           { return n.Level & 0x3fff }

func normalizeEdge(e Dec_Edge_t) Dec_Edge_t { return Dec_IntToEdge(Dec_EdgeToInt(e)) }
