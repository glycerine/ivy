package dar

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
)

func Dar_ManStart(g *aig.Graph, params *RewriteParams) (*Manager, error) {
	p := DefaultRewriteParams()
	if params != nil {
		p = *params
	}
	if g == nil {
		return nil, fmt.Errorf("dar: nil graph")
	}
	if p.CutsMax < 2 || p.SubgraphsMax < 1 {
		return nil, fmt.Errorf("dar: invalid cut or subgraph limit")
	}
	if p.Power {
		return nil, fmt.Errorf("dar: power-driven rewriting requires switching activity and is not available")
	}
	if err := g.Check(); err != nil {
		return nil, err
	}
	return &Manager{Graph: g, Params: p, library: Dar_LibPrepare(p.SubgraphsMax)}, nil
}

// Dar_ManStop releases temporary search state; the caller retains its graph.
func Dar_ManStop(m *Manager) { m.library = nil }
