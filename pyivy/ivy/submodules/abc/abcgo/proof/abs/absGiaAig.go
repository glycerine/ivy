package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
)

// absGiaToAigSimple matches Gia_ManToAigSimple's source assertion at the
// package's C call sites. The Linux Makefile's default -g -O flags do not
// define NDEBUG, so its assert that each converted object keeps its ID is
// active. The shared Go exporter retains the actual mapped edges for callers
// that rely on release-build behavior; ABS validates the debug-source contract
// locally before using the graph.
func absGiaToAigSimple(p *gia.Gia_Man_t) *aig.Graph {
	graph := gia.Gia_ManToAigSimple(p)
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		mapped := gia.Gia_ObjValue(gia.Gia_ManObj(p, id))
		if mapped < 0 || int32(uint32(mapped)>>1) != id {
			panic("gia: simple export changed object ID")
		}
	}
	return graph
}
