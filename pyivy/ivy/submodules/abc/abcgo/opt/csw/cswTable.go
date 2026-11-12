package csw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bool/kit"
	"slices"
)

func Csw_CutHash(c *Csw_Cut_t) uint32 {
	primes := [16]uint32{1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459, 1499, 1559, 1607, 1657, 1709, 1759}
	if len(c.pFanins) > 16 {
		panic("csw: cut exceeds sixteen leaves")
	}
	var hash uint32
	for i, id := range c.pFanins {
		hash ^= uint32(id) * primes[i]
	}
	return hash
}
func Csw_TableCountCuts(p *Csw_Man_t) int32 {
	count := int32(0)
	for _, head := range p.pTable {
		for c := head; c != nil; c = c.pNext {
			count++
		}
	}
	return count
}
func Csw_TableCutInsert(p *Csw_Man_t, c *Csw_Cut_t) {
	bucket := int(Csw_CutHash(c) % uint32(len(p.pTable)))
	c.pNext = p.pTable[bucket]
	p.pTable[bucket] = c
}
func Csw_TableCutLookup(p *Csw_Man_t, c *Csw_Cut_t) aig.Edge {
	bucket := int(Csw_CutHash(c) % uint32(len(p.pTable)))
	for old := p.pTable[bucket]; old != nil; old = old.pNext {
		if old.uSign != c.uSign || !slices.Equal(old.pFanins, c.pFanins) {
			continue
		}
		n := int32(len(c.pFanins))
		phase := (old.truth[0] ^ c.truth[0]) & 1
		equal := kit.Kit_TruthIsEqual(old.truth, c.truth, n) != 0
		if phase != 0 {
			equal = kit.Kit_TruthIsOpposite(old.truth, c.truth, n) != 0
		}
		if equal {
			e := aig.Edge(old.iNode) << 1
			if phase != 0 {
				e = e.Not()
			}
			return e
		}
	}
	return aig.InvalidEdge
}
