package dch

import (
	"fmt"
	"time"

	aig "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func (m *Manager) computeEquivalences() error {
	// Dch_ClassesStart requires unallocated representative storage; otherwise
	// starting a new sweep would silently discard the caller's existing links.
	if aig.Aig_ManHasReprs(m.graph) {
		panic("dch: representative storage already exists")
	}
	started := time.Now()
	simStart := time.Now()
	m.random = &globalRandomSource{reset: true}
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
	if manager.params.Verbose { // Dch_ManStop
		manager.printStats()
		util.Abc_PrintTime(1, "Old choice computation time", time.Since(started).Nanoseconds())
	}
	result, err := DeriveChoiceAig(g, manager.params.SkipRedundantSupport)
	manager.stats.Total = time.Since(started)
	if manager.params.Verbose && err == nil {
		fmt.Printf("STATS:  Ands:%8d  ->%8d.  Reprs:%7d  ->%7d.  Choices =%7d.\n",
			aig.Aig_ManNodeNum(g), aig.Aig_ManNodeNum(result),
			Dch_DeriveChoiceCountReprs(g), Dch_DeriveChoiceCountEquivs(result), aig.Aig_ManChoiceNum(result))
	}
	return result, manager.stats, err
}

// printStats is Dch_ManPrintStats. The port does not split the SAT time by
// outcome or time choice derivation separately; those lines report zero.
func (m *Manager) printStats() {
	total := int(aig.Aig_ManNodeNum(m.graph))
	nodeNum := total / 3
	fmt.Printf("Parameters: Sim words = %d. Conf limit = %d. SAT var max = %d.\n",
		m.params.Words, m.params.ConflictLimit, m.params.SatVarMax)
	fmt.Printf("AIG nodes : Total = %6d. Dangling = %6d. Main = %6d. (%6.2f %%)\n",
		total, total-nodeNum, nodeNum, 100.0*float64(nodeNum)/float64(total))
	fmt.Printf("SAT solver: Vars = %d. Max cone = %d. Recycles = %d.\n", m.nextVar, m.coneMax, m.stats.Recycles)
	fmt.Printf("SAT calls : All = %6d. Unsat = %6d. Sat = %6d. Fail = %6d.\n",
		m.stats.SATCalls, m.stats.SATCalls-m.stats.Counterexamples-m.stats.Unknown, m.stats.Counterexamples, m.stats.Unknown)
	fmt.Printf("Choices   : Lits = %6d. Reprs = %5d. Equivs = %5d. Choices = %5d.\n", m.stats.Candidates, 0, 0, 0)
	fmt.Printf("Choicing runtime statistics:\n")
	tot := m.stats.Total.Nanoseconds()
	sim, sat := m.stats.Simulation.Nanoseconds(), m.stats.Solving.Nanoseconds()
	util.Abc_PrintTimeP(1, "Sim init   ", sim, tot)
	util.Abc_PrintTimeP(1, "Sim SAT    ", 0, tot)
	util.Abc_PrintTimeP(1, "SAT solving", sat, tot)
	util.Abc_PrintTimeP(1, "  sat      ", 0, tot)
	util.Abc_PrintTimeP(1, "  unsat    ", 0, tot)
	util.Abc_PrintTimeP(1, "  undecided", 0, tot)
	util.Abc_PrintTimeP(1, "Choice     ", 0, tot)
	util.Abc_PrintTimeP(1, "Other      ", tot-sim-sat, tot)
	util.Abc_PrintTimeP(1, "TOTAL      ", tot, tot)
	if m.params.SynthDuration != 0 {
		util.Abc_PrintTime(1, "Synthesis  ", m.params.SynthDuration.Nanoseconds())
	}
}
