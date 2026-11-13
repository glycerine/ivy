package cec

import "github.com/glycerine/abcgo/aig/gia"

// cec4SimulateNode implements the combinational node propagation used by
// Cec4_ObjSimAnd, Cec4_ObjSimXor and Cec4_ObjSimCo.
func cec4SimulateNode(s *cec3Simulation, graph *gia.Gia_Man_t, id int32) {
	a, b, c, kind := graph.Gate(id)
	switch kind {
	case 2:
		s.andFanins(id, a>>1, b>>1, a&1 != 0, b&1 != 0)
	case 3:
		dst, left, right := s.object(id), s.object(a>>1), s.object(b>>1)
		complemented := (a^b)&1 != 0
		for i := range dst {
			dst[i] = left[i] ^ right[i]
			if complemented {
				dst[i] = ^dst[i]
			}
		}
	case 4:
		dst, control, yes, no := s.object(id), s.object(a>>1), s.object(b>>1), s.object(c>>1)
		for i := range dst {
			v, y, n := control[i], yes[i], no[i]
			if a&1 != 0 {
				v = ^v
			}
			if b&1 != 0 {
				y = ^y
			}
			if c&1 != 0 {
				n = ^n
			}
			dst[i] = (v & y) | (^v & n)
		}
	case 6:
		s.copyDriver(id, a>>1, a&1 != 0)
	default:
		panic("cec4: unsupported simulation node")
	}
}
