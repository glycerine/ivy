package ssw

// Partitioned signal correspondence, ported from sswPart.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"time"
)

func Ssw_SignalCorrespondencePart(g *aig.Graph, pars *Ssw_Pars_t) *aig.Graph {
	if pars.FConstrs != 0 {
		fmt.Println("Cannot use partitioned computation with constraints.")
		return nil
	}
	started := time.Now()
	partSize, verbose := pars.NPartSize, pars.FVerbose
	pars.NPartSize = 0
	pars.FVerbose = 0
	defer func() { pars.NPartSize = partSize; pars.FVerbose = verbose }()
	var partitions [][]int32
	if g.ClockDomains != nil {
		for _, domain := range g.ClockDomains {
			if partSize != 0 && len(domain) > int(partSize) {
				aig.Aig_ManPartDivide(&partitions, domain, partSize, pars.NOverSize)
			} else {
				partitions = append(partitions, append([]int32(nil), domain...))
			}
		}
	} else {
		partitions = aig.Aig_ManRegPartitionSimple(g, partSize, pars.NOverSize)
	}
	aig.Aig_ManReprStart(g, int32(len(g.Nodes())))
	for i, part := range partitions {
		var countPis, countRegs int32
		var mapping []int32
		subgraph := aig.Aig_ManRegCreatePart(g, part, &countPis, &countRegs, &mapping)
		if g.OneHotGroups != nil {
			subgraph.OneHotInputGroups = aig.Aig_ManRegProjectOnehots(g, subgraph, g.OneHotGroups, verbose)
		}
		if countPis > 0 {
			reduced := Ssw_SignalCorrespondence(subgraph, pars)
			classes := aig.Aig_TransferMappedClasses(g, subgraph, mapping)
			if verbose != 0 {
				fmt.Printf("%3d : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d. It = %3d. Cl = %5d.\n", i, len(part), subgraph.InputCount()-len(part), countPis, countRegs, aig.Aig_ManNodeNum(subgraph), pars.nIters, classes)
			}
			aig.Aig_ManStop(reduced)
		}
		aig.Aig_ManStop(subgraph)
	}
	result := aig.Aig_ManDupRepr(g, 0)
	aig.Aig_ManSeqCleanup(result)
	if verbose != 0 {
		fmt.Printf("Total time: %s\n", time.Since(started))
	}
	return result
}
