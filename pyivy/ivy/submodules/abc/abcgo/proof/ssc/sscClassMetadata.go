package ssc

import "github.com/glycerine/abcgo/aig/gia"

// publish synchronizes the GIA's representative and class links with the
// working partition. Constant candidates point at zero without class links.
func (c *Classes) publish() {
	if c.owner == nil {
		return
	}
	for id := range c.Repr {
		gia.Gia_ObjUnsetRepr(c.owner, int32(id))
		gia.Gia_ObjSetNext(c.owner, int32(id), 0)
	}
	for _, group := range c.Groups {
		for i := 1; i < len(group); i++ {
			gia.Gia_ObjSetRepr(c.owner, group[i], group[0])
			if group[0] != 0 {
				gia.Gia_ObjSetNext(c.owner, group[i-1], group[i])
			}
		}
	}
}
