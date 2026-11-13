package fraig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"math/bits"
)

// fraigNodeSimComplement reports the phase in which an edge's simulation is stored.
func fraigNodeSimComplement(node *Fraig_Node_t) bool {
	return (Fraig_Regular(node).fInv != 0) != Fraig_IsComplement(node)
}

func fraigNodeSimulationStorage(p *Fraig_Man_t, node *Fraig_Node_t) {
	storage := make([]uint32, p.nWordsRand+p.nWordsDyna)
	node.puSimR = storage[:p.nWordsRand]
	node.puSimD = storage[p.nWordsRand:]
}

func Fraig_NodeCreateConst(p *Fraig_Man_t) *Fraig_Node_t {
	node := &Fraig_Node_t{Num: Fraig_NodeVecReadSize(p.vNodes), NumPi: -1, nRefs: 1, fInv: 1}
	Fraig_NodeVecPush(p.vNodes, node)
	fraigNodeSimulationStorage(p, node)
	node.nOnes = uint32(p.nWordsRand) * 32
	Fraig_HashTableLookupF0(p, node)
	return node
}

func Fraig_NodeCreatePi(p *Fraig_Man_t) *Fraig_Node_t {
	node := &Fraig_Node_t{Num: Fraig_NodeVecReadSize(p.vNodes), NumPi: Fraig_NodeVecReadSize(p.vInputs), nRefs: 1}
	fraigNodeSimulationStorage(p, node)
	Fraig_NodeVecPush(p.vNodes, node)
	Fraig_NodeVecPush(p.vInputs, node)
	start := util.Abc_Clock()
	for i := int32(0); i < p.nWordsRand; i++ {
		node.puSimR[i] = aig.Aig_ManRandom(0)
		if i == 0 {
			node.puSimR[i] <<= 1
		}
		node.uHashR ^= node.puSimR[i] * uint32(s_FraigPrimes[i])
		node.nOnes += uint32(bits.OnesCount32(node.puSimR[i]))
	}
	for i := int32(0); i < p.iWordStart; i++ {
		node.puSimD[i] = aig.Aig_ManRandom(0)
		node.uHashD ^= node.puSimD[i] * uint32(s_FraigPrimes[i])
	}
	p.timeSims += util.Abc_Clock() - start
	assert(Fraig_HashTableLookupF(p, node) == nil, "new primary input has a unique simulation")
	return node
}

func Fraig_NodeCreate(p *Fraig_Man_t, p1, p2 *Fraig_Node_t) *Fraig_Node_t {
	n1, n2 := Fraig_Regular(p1), Fraig_Regular(p2)
	node := &Fraig_Node_t{p1: p1, p2: p2, Num: Fraig_NodeVecReadSize(p.vNodes), NumPi: -1}
	n1.nRefs++
	n2.nRefs++
	Fraig_NodeVecPush(p.vNodes, node)
	node.Level = n1.Level
	if n2.Level > node.Level {
		node.Level = n2.Level
	}
	node.Level++
	if fraigNodeSimComplement(p1) && fraigNodeSimComplement(p2) {
		node.fInv = 1
	}
	node.fFailTfo = n1.fFailTfo | n2.fFailTfo
	start := util.Abc_Clock()
	fraigNodeSimulationStorage(p, node)
	Fraig_NodeSimulate(node, 0, p.nWordsRand, 1)
	Fraig_NodeSimulate(node, 0, p.iWordStart, 0)
	for _, word := range node.puSimR {
		node.nOnes += uint32(bits.OnesCount32(word))
	}
	if node.fInv != 0 {
		node.nOnes = uint32(p.nWordsRand)*32 - node.nOnes
	}
	p.timeSims += util.Abc_Clock() - start
	Fraig_NodeAddFaninFanout(n1, node)
	Fraig_NodeAddFaninFanout(n2, node)
	return node
}

// Fraig_NodeSimulate incrementally computes the normalized AND simulation and hash.
func Fraig_NodeSimulate(node *Fraig_Node_t, start, stop, useRand int32) {
	assert(!Fraig_IsComplement(node), "node must be regular")
	left, right := Fraig_Regular(node.p1), Fraig_Regular(node.p2)
	dst, a, b := node.puSimD, left.puSimD, right.puSimD
	if useRand != 0 {
		dst, a, b = node.puSimR, left.puSimR, right.puSimR
	}
	c1, c2 := fraigNodeSimComplement(node.p1), fraigNodeSimComplement(node.p2)
	var hash uint32
	for i := start; i < stop; i++ {
		x, y := a[i], b[i]
		if c1 {
			x = ^x
		}
		if c2 {
			y = ^y
		}
		value := x & y
		if node.fInv != 0 {
			value = ^value
		}
		dst[i] = value
		hash ^= value * uint32(s_FraigPrimes[i])
	}
	if useRand != 0 {
		node.uHashR ^= hash
	} else {
		node.uHashD ^= hash
	}
}
