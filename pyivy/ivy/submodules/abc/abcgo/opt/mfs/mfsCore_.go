package mfs

// The upstream mfsCore_.c is an obsolete duplicate of mfsCore.c, excluded by
// module.make. Its pass variants share the native window/solver implementation.
// PowerResub preserves the source's three thresholded passes.
func Abc_NtkMfsPowerResub(p *Mfs_Man_t) {
	for pass := 0; pass < 3; pass++ {
		threshold := float32(0.35)
		if pass >= 2 {
			threshold = 0.2
		}
		for target, v := range p.Network.Nodes {
			if target == 0 || v.Input || v.Deleted || len(v.Fanins) < 2 || len(v.Fanins) > 8 || (p.Params.DepthMax > 0 && v.Level > p.Params.DepthMax) {
				continue
			}
			if !Abc_WinNode(p, target) {
				continue
			}
			Abc_MfsCreateSolverResub(p, nil, false)
			for i := 0; i < len(p.Network.Nodes[target].Fanins); i++ {
				f := p.Network.Nodes[target].Fanins[i]
				if p.probabilities[f] >= threshold && Abc_NtkMfsSolveSatResub(p, target, i, false, false) {
					// Continue the source pass after success, but prepare a fresh
					// projection for the changed target before the next fanin.
					if !Abc_WinNode(p, target) {
						break
					}
					Abc_MfsCreateSolverResub(p, nil, false)
				}
			}
		}
	}
}
