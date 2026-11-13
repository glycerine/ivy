package ssc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"sort"
)

func Ssc_GiaClassesInit(p *gia.Gia_Man_t, pivot []bool) *Classes {
	if p.HasRepresentatives() {
		panic("ssc: representative storage already exists")
	}
	s := NewSimulation(p, 1)
	Ssc_GiaSavePiPattern(s, pivot)
	Ssc_GiaSimRound(s)
	p.StartRepresentatives()
	c := &Classes{owner: p, Repr: make([]int32, p.ObjectCount()), Phase: make([]bool, p.ObjectCount())}
	group := []int32{0}
	for id := 0; id < p.ObjectCount(); id++ {
		c.Repr[id] = -1
		c.Phase[id] = s.Values[id][0]&1 != 0
		value := int32(0)
		if c.Phase[id] {
			value = 1
		}
		gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(p, int32(id)), value)
		if _, _, _, k := p.Gate(int32(id)); k >= 1 && k <= 5 {
			group = append(group, int32(id))
		}
	}
	c.Groups = [][]int32{group}
	c.reindex()
	return c
}
func (c *Classes) reindex() {
	for i := range c.Repr {
		c.Repr[i] = -1
	}
	for _, g := range c.Groups {
		for _, id := range g[1:] {
			c.Repr[id] = g[0]
		}
	}
	c.publish()
}
func Ssc_GiaSimAreEqual(s *Simulation, a, b int32) bool {
	x, y := s.Values[a], s.Values[b]
	flip := (x[0]^y[0])&1 != 0
	for i := range x {
		v := y[i]
		if flip {
			v = ^v
		}
		if x[i] != v {
			return false
		}
	}
	return true
}
func Ssc_GiaSimIsConst0(s *Simulation, id int32) bool { return Ssc_GiaSimAreEqual(s, 0, id) }
func Ssc_GiaClassesRefine(c *Classes, s *Simulation) int32 {
	// C first refines existing nonconstant classes, then hashes nodes that
	// just left the constant class and refines each resulting hash bucket.
	var out [][]int32
	var moved []int32
	var splits int32
	partition := func(g []int32) [][]int32 {
		var parts [][]int32
		for len(g) > 0 {
			group := []int32{g[0]}
			var rest []int32
			for _, id := range g[1:] {
				if Ssc_GiaSimAreEqual(s, g[0], id) {
					group = append(group, id)
				} else {
					rest = append(rest, id)
				}
			}
			parts = append(parts, group)
			g = rest
		}
		return parts
	}
	for _, g := range c.Groups {
		if g[0] == 0 {
			constants := []int32{0}
			for _, id := range g[1:] {
				if Ssc_GiaSimIsConst0(s, id) {
					constants = append(constants, id)
				} else {
					moved = append(moved, id)
				}
			}
			out = append(out, constants)
			continue
		}
		parts := partition(g)
		splits += int32(len(parts) - 1)
		out = append(out, parts...)
	}
	if len(moved) != 0 {
		primes := [...]uint64{1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147}
		tableSize := uint32(util.Abc_PrimeCudd(uint32(100 + len(moved)/3)))
		buckets := make(map[uint32][]int32)
		var keys []uint32
		for _, id := range moved {
			var hash uint32
			flip := s.Values[id][0]&1 != 0
			for i, word := range s.Values[id] {
				if flip {
					word = ^word
				}
				hash ^= uint32(word * primes[i&15])
			}
			key := hash % tableSize
			if len(buckets[key]) == 0 {
				keys = append(keys, key)
			}
			buckets[key] = append(buckets[key], id)
		}
		for _, key := range keys {
			out = append(out, partition(buckets[key])...)
		}
		splits += int32(len(moved))
	}
	sort.Slice(out, func(i, j int) bool { return out[i][0] < out[j][0] })
	c.Groups = out
	c.reindex()
	return splits
}
func Ssc_GiaSimClassRefineOneBit(c *Classes, values []bool, head int32) int32 {
	for i, g := range c.Groups {
		if g[0] != head {
			continue
		}
		var same, diff []int32
		reference := values[head] != c.Phase[head]
		for _, id := range g {
			if (values[id] != c.Phase[id]) == reference {
				same = append(same, id)
			} else {
				diff = append(diff, id)
			}
		}
		if len(diff) == 0 {
			return 0
		}
		c.Groups[i] = same
		c.Groups = append(c.Groups, diff)
		c.reindex()
		return 1
	}
	panic("ssc: class head not found")
}
func Ssc_GiaClassesCheckPairs(c *Classes, pairs [][2]int32) bool {
	for _, pair := range pairs {
		if c.Repr[pair[1]] == pair[0] {
			return false
		}
	}
	return true
}
