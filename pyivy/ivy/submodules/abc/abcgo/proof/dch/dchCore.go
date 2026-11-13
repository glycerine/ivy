package dch

import (
	"fmt"
	"time"

	aig "github.com/glycerine/abcgo/aig/aig"
)

func (m *Manager) computeEquivalences() error {
	// Dch_ClassesStart requires unallocated representative storage; otherwise
	// starting a new sweep would silently discard the caller's existing links.
	if aig.Aig_ManHasReprs(m.graph) {
		panic("dch: representative storage already exists")
	}
	started := time.Now()
	simStart := time.Now()
	m.random = aig.NewAigRandomSource()
	m.classes = createClasses(m.graph, m.order, m.params.Words, m.random)
	m.stats.Simulation = time.Since(simStart)
	m.stats.Candidates = Dch_ClassesLitNum(m.classes)
	if err := m.sweep(); err != nil {
		return err
	}
	// Only proved links survive. Never publish simulation-only candidates.
	aig.Aig_ManReprStart(m.graph, int32(m.graph.NodeCount()))
	for id, representative := range m.proven {
		aig.Aig_ObjSetRepr(m.graph, aig.Edge(id)<<1, representative)
	}
	m.stats.Total = time.Since(started)
	return nil
}
func ComputeEquivalences(g *aig.Graph, p *Params) error {
	manager, err := newManager(g, p)
	if err != nil {
		return err
	}
	return manager.computeEquivalences()
}
func ComputeChoices(g *aig.Graph, p *Params) (*aig.Graph, error) {
	result, _, err := ComputeChoicesWithStats(g, p)
	return result, err
}
func ComputeChoicesWithStats(g *aig.Graph, p *Params) (*aig.Graph, Statistics, error) {
	started := time.Now()
	manager, err := newManager(g, p)
	if err != nil {
		return nil, Statistics{}, err
	}
	if err = manager.computeEquivalences(); err != nil {
		return nil, manager.stats, err
	}
	result, err := DeriveChoiceAig(g, manager.params.SkipRedundantSupport)
	manager.stats.Total = time.Since(started)
	if manager.params.Verbose && err == nil {
		fmt.Printf("DCH: candidates=%d proofs=%d sat=%d counterexamples=%d unknown=%d choices=%d\n", manager.stats.Candidates, manager.stats.Proofs, manager.stats.SATCalls, manager.stats.Counterexamples, manager.stats.Unknown, aig.Aig_ManChoiceNum(result))
	}
	return result, manager.stats, err
}
