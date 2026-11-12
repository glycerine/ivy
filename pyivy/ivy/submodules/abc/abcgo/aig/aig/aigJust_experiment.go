package aig

import (
	"fmt"
	"time"
)

// Aig_ManJustExperiment exercises rare-value justification and packs resulting
// support assignments into the simulation patterns used by ABC's experiment.
func Aig_ManJustExperiment(p *Graph) {
	start := time.Now()
	pack := Aig_ManPackStart(p)
	nodes := Aig_ManPackConstNodes(pack)
	var support []int32
	counts := [2]int{}
	failed := [2]int{}
	literals := 0
	defer func() {
		for id := range p.nodes {
			p.nodes[id].MarkA = false
			p.nodes[id].MarkB = false
		}
		Aig_ManPackStop(pack)
	}()
	for _, id := range nodes {
		e := Edge(id) << 1
		target := int32(1)
		if p.node(e).Phase {
			target = 0
		}
		if Aig_ObjFindSatAssign(p, e, target, &support) {
			counts[target]++
			literals += len(support)
			Aig_ManPackAddPattern(pack, support)
		} else {
			failed[target]++
		}
	}
	successes := counts[0] + counts[1]
	percent, average := 0.0, 0.0
	if len(p.outputs) > 0 {
		percent = 100 * float64(successes) / float64(len(p.outputs))
	}
	if successes > 0 {
		average = float64(literals) / float64(successes)
	}
	fmt.Printf("PO =%6d. C0 =%6d. C0f =%6d. C1 =%6d. C1f =%6d. (%6.2f %%) Ave =%4.1f T=%s\n", len(p.outputs), counts[0], failed[0], counts[1], failed[1], percent, average, time.Since(start))
}
