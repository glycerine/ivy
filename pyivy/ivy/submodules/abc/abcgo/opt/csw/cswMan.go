package csw

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func cswTruthWords(n int32) int {
	if n <= 5 {
		return 1
	}
	return 1 << uint(n-5)
}
func Csw_ManStart(graph *aig.Graph, nCutsMax, nLeafMax, verbose int32) *Csw_Man_t {
	if graph == nil || nCutsMax < 2 || nLeafMax < 1 || nLeafMax > 16 {
		panic("csw: need at least two cuts and one through sixteen leaves")
	}
	if err := graph.Check(); err != nil {
		panic(err)
	}
	p := &Csw_Man_t{pManAig: graph, pManRes: aig.Aig_ManStartFrom(graph), nCutsMax: nCutsMax, nLeafMax: nLeafMax, fVerbose: verbose, pEquiv: make([]aig.Edge, aig.Aig_ManObjNumMax(graph))}
	for i := range p.pEquiv {
		p.pEquiv[i] = aig.InvalidEdge
	}
	p.pEquiv[0] = p.pManRes.Const1()
	outInputs := p.pManRes.Inputs()
	for i, ci := range graph.Inputs() {
		p.pEquiv[ci.NodeID()] = outInputs[i]
	}
	p.ensureObject(int(aig.Aig_ManObjNumMax(p.pManRes)) - 1)
	size := util.Abc_PrimeCudd(uint32(aig.Aig_ManNodeNum(graph) * nCutsMax / 2))
	p.pTable = make([]*Csw_Cut_t, size)
	for i := range p.puTemp {
		p.puTemp[i] = make([]uint32, cswTruthWords(nLeafMax))
	}
	return p
}
func Csw_ManStop(p *Csw_Man_t) {
	if p == nil {
		return
	}
	if p.fVerbose != 0 {
		before, after := aig.Aig_ManNodeNum(p.pManAig), aig.Aig_ManNodeNum(p.pManRes)
		percent := float64(0)
		if before != 0 {
			percent = 100 * float64(before-after) / float64(before)
		}
		fmt.Printf("Beg = %7d. End = %7d. (%6.2f %%) Try = %7d. Cuts = %8d.\n", before, after, percent, p.nNodesTried, Csw_TableCountCuts(p))
		fmt.Printf("Triv0 = %6d. Triv1 = %6d. Triv2 = %6d. Cut-replace = %6d.\n", p.nNodesTriv0, p.nNodesTriv1, p.nNodesTriv2, p.nNodesCuts)
		for _, s := range []struct {
			name string
			ns   int64
		}{{"Cuts", p.timeCuts}, {"Hashing", p.timeHash}, {"Other", p.timeOther}, {"TOTAL", p.timeTotal}} {
			fmt.Printf("%s = %.2f sec\n", s.name, float64(s.ns)/1e9)
		}
	}
	*p = Csw_Man_t{}
}
