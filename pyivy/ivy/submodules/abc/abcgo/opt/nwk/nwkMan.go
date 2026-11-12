package nwk

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/misc/tim"
	"strings"
	"sync"
)

func Nwk_ManAlloc() *Nwk_Man_t { return &Nwk_Man_t{Hop: hop.Hop_ManStart()} }
func Nwk_ManFree(p *Nwk_Man_t) {
	if p == nil {
		return
	}
	if p.Hop != nil {
		hop.Hop_ManStop(p.Hop)
	}
	if p.Timing != nil {
		tim.Tim_ManStop(p.Timing)
	}
	*p = Nwk_Man_t{}
}
func Nwk_ManPrintLutSizes(p *Nwk_Man_t, lib *If_LibLut_t) {
	sizes := map[int]int{}
	for _, o := range p.Objects {
		if o != nil && o.Type == NWK_OBJ_NODE {
			sizes[len(o.Fanins)]++
		}
	}
	fmt.Fprint(writer(p), "LUTs by size: ")
	maxSize := int(Nwk_ManGetFaninMax(p))
	if lib != nil {
		maxSize = int(lib.LutMax)
	}
	for i := 0; i <= maxSize; i++ {
		fmt.Fprintf(writer(p), "%d:%d ", i, sizes[i])
	}
}

var bestNetwork struct {
	sync.Mutex
	name                string
	valid               bool
	depth, flops, nodes int32
}

func Nwk_ManCompareAndSaveBest(p *Nwk_Man_t, context any) int32 {
	bestNetwork.Lock()
	defer bestNetwork.Unlock()
	if p == nil {
		bestNetwork.valid = false
		bestNetwork.name = ""
		return 0
	}
	depth, flops, nodes := Nwk_ManLevel(p), Nwk_ManLatchNum(p), Nwk_ManNodeNum(p)
	if !bestNetwork.valid || bestNetwork.name != p.Name || depth < bestNetwork.depth || depth == bestNetwork.depth && (flops < bestNetwork.flops || flops == bestNetwork.flops && nodes < bestNetwork.nodes) {
		bestNetwork.valid = true
		bestNetwork.name = p.Name
		bestNetwork.depth = depth
		bestNetwork.flops = flops
		bestNetwork.nodes = nodes
		return 1
	}
	return 0
}
func Nwk_FileNameGeneric(name string) string {
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		return name[:i]
	}
	return name
}
func Nwk_ManPrintStats(p *Nwk_Man_t, lib *If_LibLut_t, save, dump, power int32, context *Ntl_Man_t) {
	if save != 0 {
		Nwk_ManCompareAndSaveBest(p, context)
	}
	p.Library = lib
	fmt.Fprintf(writer(p), "%-15s : pi = %5d  po = %5d  ci = %5d  co = %5d  node = %5d  edge = %5d  aig = %6d  lev = %3d  delay = %5.2f  ", p.Name, Nwk_ManPiNum(p), Nwk_ManPoNum(p), Nwk_ManCiNum(p), Nwk_ManCoNum(p), Nwk_ManNodeNum(p), Nwk_ManGetTotalFanins(p), Nwk_ManGetAigNodeNum(p), Nwk_ManLevel(p), Nwk_ManDelayTraceLut(p))
	if power != 0 {
		fmt.Fprintf(writer(p), "power = %7.2f  ", Nwl_ManComputeTotalSwitching(p))
	}
	Nwk_ManPrintLutSizes(p, lib)
	fmt.Fprintln(writer(p))
}
