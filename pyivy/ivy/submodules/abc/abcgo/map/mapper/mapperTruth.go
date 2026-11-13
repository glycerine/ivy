package mapper

import (
	"github.com/glycerine/abcgo/misc/extra"
	"os"
)

// Map_MappingTruths derives the source mapper's canonical cut matches.
func Map_MappingTruths(p *Map_Man_t) {
	progress := extra.Extra_ProgressBarStart(os.Stdout, p.vMapObjs.nSize)
	for i := int32(0); i < p.vMapObjs.nSize; i++ {
		node := p.vMapObjs.pArray[i]
		if Map_NodeIsAnd(node) == 0 {
			continue
		}
		assert(node.pCuts != nil && node.pCuts.nLeaves == 1, "node has trivial cut")
		for phase := 0; phase < 2; phase++ {
			match := &node.pCuts.M[phase]
			match.uPhase = 0
			match.pSupers = p.pSuperLib.pSuperInv
			match.uPhaseBest = uint32(phase)
			match.pSuperBest = p.pSuperLib.pSuperInv
		}
		for cut := node.pCuts.pNext; cut != nil; cut = cut.pNext {
			Map_TruthsCut(p, cut)
		}
		extra.Extra_ProgressBarUpdate(progress, i, "Tables ...")
	}
	extra.Extra_ProgressBarStop(progress)
}

func Map_TruthsCut(p *Map_Man_t, cut *Map_Cut_t) {
	if cut.nLeaves == 1 {
		return
	}
	var truth, canon [2]uint32
	var phases [16]uint8
	Map_TruthsCutOne(p, cut, truth[:])
	for phase := 1; phase >= 0; phase-- {
		Map_CanonComputeFast(p, p.nVarsMax, int32(cut.nLeaves), truth[:], phases[:], canon[:])
		cut.M[phase].pSupers = Map_SuperTableLookupC(p.pSuperLib, canon[:])
		cut.M[phase].uPhase = uint32(phases[0])
		p.nCanons++
		truth[0], truth[1] = ^truth[0], ^truth[1]
	}
}

// Map_TruthsCutOne evaluates the cut cone in source postorder.
func Map_TruthsCutOne(p *Map_Man_t, cut *Map_Cut_t, truth []uint32) {
	for i := 0; i < int(cut.nLeaves); i++ {
		leaf := cut.ppLeaves[i].pCuts
		leaf.fMark = 1
		leaf.M[0].uPhaseBest = p.uTruths[i][0]
		leaf.M[1].uPhaseBest = p.uTruths[i][1]
	}
	assert(cut.fMark == 0, "cut is not a leaf")
	var visited []*Map_Cut_t
	var collect func(*Map_Cut_t)
	collect = func(c *Map_Cut_t) {
		c = Map_CutRegular(c)
		if c.fMark != 0 {
			return
		}
		collect(c.pOne)
		collect(c.pTwo)
		c.fMark = 1
		visited = append(visited, c)
	}
	collect(cut)
	assert(len(visited) > 0, "nontrivial cut")
	cut.nVolume = int32(len(visited))
	for i := 0; i < int(cut.nLeaves); i++ {
		cut.ppLeaves[i].pCuts.fMark = 0
	}
	for _, c := range visited {
		c.fMark = 0
		a, b := Map_CutRegular(c.pOne), Map_CutRegular(c.pTwo)
		for i := 0; i < 2; i++ {
			x, y := a.M[i].uPhaseBest, b.M[i].uPhaseBest
			if Map_CutIsComplement(c.pOne) {
				x = ^x
			}
			if Map_CutIsComplement(c.pTwo) {
				y = ^y
			}
			value := x & y
			if c.Phase != 0 {
				value = ^value
			}
			c.M[i].uPhaseBest = value
		}
	}
	truth[0], truth[1] = cut.M[0].uPhaseBest, cut.M[1].uPhaseBest
}
