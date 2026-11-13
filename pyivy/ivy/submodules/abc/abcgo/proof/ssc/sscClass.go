package ssc

import (
	"encoding/binary"
	"github.com/glycerine/abcgo/aig/gia"
)

func Ssc_GiaClassesInit(p *gia.Gia_Man_t, pivot []bool) *Classes {
	s := NewSimulation(p, 1)
	Ssc_GiaSavePiPattern(s, pivot)
	Ssc_GiaSimRound(s)
	c := &Classes{Repr: make([]int32, p.ObjectCount()), Phase: make([]bool, p.ObjectCount())}
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
	var out [][]int32
	var splits int32
	for _, g := range c.Groups {
		groups := map[string]int{}
		var parts [][]int32
		for _, id := range g {
			buf := make([]byte, 8*s.Words)
			flip := s.Values[id][0]&1 != 0
			for w, x := range s.Values[id] {
				if flip {
					x = ^x
				}
				binary.LittleEndian.PutUint64(buf[w*8:], x)
			}
			key := string(buf)
			index, ok := groups[key]
			if !ok {
				index = len(parts)
				groups[key] = index
				parts = append(parts, nil)
			}
			parts[index] = append(parts[index], id)
		}
		if len(g) > 0 && g[0] == 0 {
			// The source counts each node leaving the constant class,
			// even when several share a newly discovered signature.
			splits += int32(len(g) - len(parts[0]))
		} else {
			splits += int32(len(parts) - 1)
		}
		out = append(out, parts...)
	}
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
