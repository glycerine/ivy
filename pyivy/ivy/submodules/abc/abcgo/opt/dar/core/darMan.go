package core

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
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
	if err := g.Check(); err != nil {
		return nil, err
	}
	m := &Manager{Graph: g, Params: p, library: Dar_LibPrepare(p.SubgraphsMax)}
	if p.Power {
		var err error
		m.probabilities, err = SimulateActivity(p.Context, g, DefaultActivityOptions())
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

// Dar_ManStop releases temporary search state; the caller retains its graph.
func Dar_ManStop(m *Manager) {
	if m.Params.Verbose {
		Dar_ManPrintStats(m)
	}
	m.library = nil
	m.probabilities = nil
}

// Dar_ManPrintStats is darMan.c's.
func Dar_ManPrintStats(m *Manager) {
	g, st := m.Graph, &m.Stats
	nodes := int(aig.Aig_ManNodeNum(g))
	gain := st.NodesInit - nodes
	fmt.Printf("Tried = %8d. Beg = %8d. End = %8d. Gain = %6d. (%6.2f %%).  Cut mem = %d MB\n",
		st.NodesTried, st.NodesInit, nodes, gain, 100.0*float64(gain)/float64(st.NodesInit), st.CutMemUsed)
	fmt.Printf("Cuts = %8d. Tried = %8d. Used = %8d. Bad = %5d. Skipped = %5d. Ave = %.2f.\n",
		st.CutsAll, st.CutsTried, st.CutsUsed, st.CutsBad, st.CutsSkipped, float32(st.CutsUsed)/float32(nodes))
	fmt.Printf("Bufs = %5d. BufMax = %5d. BufReplace = %6d. BufFix = %6d.  Levels = %4d.\n",
		aig.Aig_ManBufNum(g), aig.Aig_ManBufMax(g), aig.Aig_ManBufReplaces(g), aig.Aig_ManBufFixes(g), aig.Aig_ManLevels(g))
	other := st.TimeTotal - st.TimeCuts - st.TimeEval
	util.Abc_PrintTime(1, "Cuts  ", st.TimeCuts.Nanoseconds())
	util.Abc_PrintTime(1, "Eval  ", st.TimeEval.Nanoseconds())
	util.Abc_PrintTime(1, "Other ", other.Nanoseconds())
	util.Abc_PrintTime(1, "TOTAL ", st.TimeTotal.Nanoseconds())
}
