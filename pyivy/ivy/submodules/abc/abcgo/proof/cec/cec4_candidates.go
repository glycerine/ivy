package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// cec4Candidates retains candidate IDs between passes, removing entries whose
// representative was cleared since the previous iteration.
type cec4Candidates struct {
	graph       *gia.Gia_Man_t
	ids         []int32
	read, write int
}

func newCec4Candidates(graph *gia.Gia_Man_t) *cec4Candidates {
	c := &cec4Candidates{graph: graph}
	for id := int32(1); id < gia.Gia_ManObjNum(graph); id++ {
		if gia.Gia_ObjRepr(graph, id) != gia.GIA_VOID {
			c.ids = append(c.ids, id)
		}
	}
	for i := range c.ids {
		j := int(util.Abc_Random(0) % uint32(len(c.ids)))
		c.ids[i], c.ids[j] = c.ids[j], c.ids[i]
	}
	return c
}
func (c *cec4Candidates) next() int32 {
	for len(c.ids) > 0 {
		id := c.ids[c.read]
		live := gia.Gia_ObjRepr(c.graph, id) != gia.GIA_VOID
		if live {
			c.ids[c.write] = id
			c.write++
		}
		c.read++
		if c.read == len(c.ids) {
			c.ids = c.ids[:c.write]
			c.read, c.write = 0, 0
		}
		if live {
			return id
		}
	}
	return 0
}
