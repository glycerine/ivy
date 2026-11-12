// Structural hashing ported from ABC ivyTable.c (Alan Mishchenko).
package ivy

import "fmt"

// Go's map owns collision resolution and resizing. The complete original key
// includes both complemented fanin edges, node kind and latch initialization.
type ivyHashKey struct {
	first, second *Ivy_Obj_t
	kind          Ivy_Type_t
	init          Ivy_Init_t
}

func ivyKey(p *Ivy_Man_t, node *Ivy_Obj_t) ivyHashKey {
	if p == nil || p.memoryStopped || node == nil || Ivy_IsComplement(node) != 0 || Ivy_ObjIsHash(node) == 0 {
		panic("invalid structural hash object")
	}
	a, b := node.pFanin0, node.pFanin1
	if a == nil || Ivy_Regular(a).owner != p || Ivy_Regular(a).recycled {
		panic("invalid first hash fanin")
	}
	if b != nil && (Ivy_Regular(b).owner != p || Ivy_Regular(b).recycled || Ivy_ObjId(a) >= Ivy_ObjId(b)) {
		panic("invalid or noncanonical second hash fanin")
	}
	if node.Type != IVY_LATCH && (b == nil || Ivy_ObjId(a) <= 0) {
		panic("binary hash node needs two nonconstant fanins")
	}
	return ivyHashKey{a, b, node.Type, node.Init}
}
func Ivy_TableLookup(p *Ivy_Man_t, node *Ivy_Obj_t) *Ivy_Obj_t {
	if node == nil || Ivy_IsComplement(node) != 0 {
		panic("regular object required for hash lookup")
	}
	if Ivy_ObjIsHash(node) == 0 {
		return nil
	}
	return p.pTable[ivyKey(p, node)]
}
func Ivy_TableInsert(p *Ivy_Man_t, node *Ivy_Obj_t) {
	if node == nil || Ivy_IsComplement(node) != 0 {
		panic("regular object required for hash insertion")
	}
	if Ivy_ObjIsHash(node) == 0 {
		return
	}
	key := ivyKey(p, node)
	if _, found := p.pTable[key]; found {
		panic("duplicate structural hash entry")
	}
	if node.owner != p || node.recycled || node.Id <= 0 || int64(node.Id) >= int64(len(p.vObjs)) || p.vObjs[node.Id] != node {
		panic("hash object is not manager-owned")
	}
	if p.pTable == nil {
		p.pTable = make(map[ivyHashKey]*Ivy_Obj_t)
	}
	p.pTable[key] = node
}
func Ivy_TableDelete(p *Ivy_Man_t, node *Ivy_Obj_t) {
	if node == nil || Ivy_IsComplement(node) != 0 {
		panic("regular object required for hash deletion")
	}
	if Ivy_ObjIsHash(node) == 0 {
		return
	}
	key := ivyKey(p, node)
	if p.pTable[key] != node {
		panic("structural hash entry is absent")
	}
	delete(p.pTable, key)
}
func Ivy_TableUpdate(p *Ivy_Man_t, node *Ivy_Obj_t, id int32) {
	if node == nil || Ivy_IsComplement(node) != 0 {
		panic("regular object required for hash update")
	}
	if Ivy_ObjIsHash(node) == 0 {
		return
	}
	key := ivyKey(p, node)
	if p.pTable[key] != node {
		panic("structural hash entry is absent")
	}
	if id < 0 || int64(id) >= int64(len(p.vObjs)) {
		panic("new hash object ID is out of bounds")
	}
	target := p.vObjs[id]
	if target == nil || target.owner != p || target.recycled {
		panic("new hash object is not live")
	}
	if Ivy_ObjIsHash(target) == 0 || ivyKey(p, target) != key {
		panic("new hash object has different structure")
	}
	for oldKey, entry := range p.pTable {
		if entry == target && oldKey != key {
			panic("new hash object is already indexed")
		}
	}
	p.pTable[key] = target
}
func Ivy_TableCountEntries(p *Ivy_Man_t) int32 { return int32(len(p.pTable)) }
func Ivy_TableResize(p *Ivy_Man_t) {
	if p == nil || p.memoryStopped {
		panic("manager stopped")
	}
	table := make(map[ivyHashKey]*Ivy_Obj_t, 2*len(p.pTable))
	for key, node := range p.pTable {
		table[key] = node
	}
	p.pTable = table
}
func Ivy_TableProfile(p *Ivy_Man_t) { fmt.Printf("Ivy structural hash: %d entries.\n", len(p.pTable)) }
