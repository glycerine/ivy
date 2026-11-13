package abs

import "github.com/glycerine/abcgo/misc/vec"

// Saig_ManCbaFindReason records the visited frame nodes for later shrinking.
func Saig_ManCbaFindReason(p *Saig_ManCba_t) *vec.Vec_Int_t {
	ref := &Saig_RefMan_t{pAig: p.pAig, pFrames: p.pFrames, pCex: p.pCex, vMapPiF2A: p.vMapPiF2A}
	result, visited := oldRefinementReason(ref)
	p.phase, p.visited = ref.phase, visited
	if len(p.pFrames.Outputs()) > 0 {
		out := p.pFrames.Node(p.pFrames.Outputs()[0]).Fanins[0]
		if out.NodeID() == p.pFrames.Const1().NodeID() {
			p.visited[out.NodeID()] = true
		}
	}
	return result
}
