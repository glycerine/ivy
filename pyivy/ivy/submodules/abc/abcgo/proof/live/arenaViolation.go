package live

import "github.com/glycerine/abcgo/aig/aig"

func createArenaLO(p *aig.Graph, barriers [][]int32) []aig.Edge {
	if len(barriers) == 0 {
		return nil
	}
	result := make([]aig.Edge, len(barriers))
	for i := range result {
		result[i] = aig.Aig_ObjCreateCi(p)
	}
	return result
}
func createArenaLi(p *aig.Graph, barriers [][]int32, signals []aig.Edge) []aig.Edge {
	if len(barriers) == 0 {
		return nil
	}
	if len(signals) < len(barriers) {
		panic("live: arena next-state count")
	}
	result := make([]aig.Edge, len(barriers))
	for i, e := range signals[:len(barriers)] {
		result[i] = aig.Aig_ObjCreateCo(p, e)
	}
	return result
}
func createMonotoneBarrierLO(p *aig.Graph, barriers [][]int32) []aig.Edge {
	return createArenaLO(p, barriers)
}
func driverToPoNew(p *aig.Graph, po aig.Edge) aig.Edge {
	node := p.Node(po)
	if node.Kind != aig.Output {
		panic("live: expected PO")
	}
	driver := node.Fanins[0]
	copy := aig.Aig_ObjCopy(p, aig.Aig_Regular(driver))
	if copy == aig.InvalidEdge {
		panic("live: missing output copy")
	}
	return copy ^ aig.Edge(driver&1)
}
func collectBarrierDisjunctions(old, new *aig.Graph, barriers [][]int32) []aig.Edge {
	if len(barriers) == 0 {
		return nil
	}
	signals := make([]aig.Edge, len(barriers))
	for i, group := range barriers {
		if len(group) == 0 {
			panic("live: empty barrier disjunction")
		}
		terms := make([]aig.Edge, len(group))
		for j, index := range group {
			if index < 0 || int(index) >= trueOutputs(old) {
				panic("live: barrier output out of range")
			}
			terms[j] = copiedDriver(old, int(index))
		}
		signals[i] = disjoin(new, terms)
	}
	return signals
}
func Aig_Xor(p *aig.Graph, a, b aig.Edge) aig.Edge { return aig.Aig_Exor(p, a, b) }
func createArenaViolation(old, new *aig.Graph, begin, within aig.Edge, barriers [][]int32, previous []aig.Edge, next, monotone *[]aig.Edge) aig.Edge {
	return createArenaViolationMode(old, new, begin, within, barriers, previous, next, monotone, true)
}
func createArenaViolationMode(old, new *aig.Graph, begin, within aig.Edge, barriers [][]int32, previous []aig.Edge, next, monotone *[]aig.Edge, source bool) aig.Edge {
	if next == nil || monotone == nil {
		panic("live: missing arena output vector")
	}
	if source && len(*monotone) != 0 {
		panic("live: source arena output vector must be empty")
	}
	signals := collectBarrierDisjunctions(old, new, barriers)
	if source && signals == nil {
		panic("live: source requires nonempty barriers")
	}
	if len(previous) < len(signals) {
		panic("live: barrier latch count")
	}
	*monotone = append((*monotone)[:0], signals...)
	violation := aig.Edge(1)
	for i, signal := range signals {
		*next = append(*next, aig.Aig_Or(new, aig.Aig_And(new, signal, begin), previous[i]))
		violation = aig.Aig_Or(new, violation, aig.Aig_And(new, aig.Aig_Exor(new, signal, previous[i]), within))
	}
	return violation
}
func createConstrained0LiveConeWithDSC(old, new *aig.Graph, signals []aig.Edge) aig.Edge {
	return createConstrained0LiveCone(old, new, signals)
}
func collectCSSignalsWithDSC(n *Network, p *aig.Graph) []aig.Edge { return collectCSSignals(n, p) }
func collectWindowBeginSignalWithDSC(n *Network, p *aig.Graph) int32 {
	return findOutput(sourceNetwork(n, p), "windowBegins_")
}
func collectWithinWindowSignalWithDSC(n *Network, p *aig.Graph) int32 {
	return findOutput(sourceNetwork(n, p), "withinWindow_")
}
func collectPendingSignalWithDSC(n *Network, p *aig.Graph) int32 {
	return findPendingSignal(sourceNetwork(n, p))
}
func createMonotoneBarriers(p *aig.Graph, signals, previous []aig.Edge, pending, oldPending aig.Edge) aig.Edge {
	if len(previous) < len(signals) {
		panic("live: monotone barrier latch count")
	}
	guard := aig.Aig_And(p, pending, oldPending)
	result := aig.Edge(0)
	for i, signal := range signals {
		result = aig.Aig_And(p, result, aig.Aig_Or(p, aig.Aig_And(p, guard, previous[i]).Not(), signal))
	}
	return result
}

// createAndGateForMonotonicityVerification retains the C argument order.
func createAndGateForMonotonicityVerification(p *aig.Graph, signals, previous []aig.Edge, pendingLo, pendingSignal aig.Edge) aig.Edge {
	return createMonotoneBarriers(p, signals, previous, pendingSignal, pendingLo)
}
func createNewAigWith0LivePoWithDSC(p *aig.Graph, signals []aig.Edge, index *int32, windowBegin, withinWindow, pendingIndex int32, barriers [][]int32) *aig.Graph {
	return createNewAigWith0LivePoWithDSCMode(p, signals, index, windowBegin, withinWindow, pendingIndex, barriers, true)
}
func createNewAigWith0LivePoWithDSCMode(p *aig.Graph, signals []aig.Edge, index *int32, windowBegin, withinWindow, pendingIndex int32, barriers [][]int32, source bool) *aig.Graph {
	if source && len(barriers) == 0 {
		panic("live: source requires nonempty barriers")
	}
	if windowBegin < 0 || withinWindow < 0 || pendingIndex < 0 {
		panic("live: arena window/pending output missing")
	}
	c := newGraphCopy(p, "_0Live")
	if source {
		c = newSourceGraphCopy(p, "_0Live")
	}
	c.copyInputs()
	arenaLatches := createArenaLO(c.destination, barriers)
	violationOld := aig.Aig_ObjCreateCi(c.destination)
	pendingOld := aig.Aig_ObjCreateCi(c.destination)
	monotoneOld := createMonotoneBarrierLO(c.destination, barriers)
	c.copySourceNodes()
	for _, e := range signals {
		c.signal(e)
	}
	// Populate copied PO drivers before the source-compatible barrier collector.
	for i := 0; i < trueOutputs(p); i++ {
		c.signal(p.Outputs()[i])
	}
	begin, within := c.signal(p.Outputs()[windowBegin]), c.signal(p.Outputs()[withinWindow])
	var barrierNext, barrierSignals []aig.Edge
	violation := createArenaViolationMode(p, c.destination, begin, within, barriers, arenaLatches, &barrierNext, &barrierSignals, source)
	pending := c.signal(p.Outputs()[pendingIndex])
	monotone := createMonotoneBarriers(c.destination, barrierSignals, monotoneOld, pending, pendingOld)
	for i := 0; i < propertyOutputs(p); i++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[i]))
	}
	property := aig.Aig_And(c.destination, aig.Aig_Or(c.destination, createConstrained0LiveConeWithDSC(p, c.destination, signals), violationOld), monotone)
	if index != nil {
		*index = int32(propertyOutputs(p))
	}
	aig.Aig_ObjCreateCo(c.destination, property)
	c.appendConstraints()
	for r := 0; r < registers(p); r++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[trueOutputs(p)+r]))
	}
	createArenaLi(c.destination, barriers, barrierNext)
	aig.Aig_ObjCreateCo(c.destination, aig.Aig_Or(c.destination, violation, violationOld))
	aig.Aig_ObjCreateCo(c.destination, pending)
	for _, e := range barrierSignals {
		aig.Aig_ObjCreateCo(c.destination, e)
	}
	aig.Aig_ManSetRegNum(c.destination, int32(registers(p)+2*len(barriers)+2))
	c.setConstraints()
	return c.destination
}
func generateWorkingAigWithDSC(p *aig.Graph, n *Network, index *int32, barriers [][]int32) *aig.Graph {
	return createNewAigWith0LivePoWithDSC(p, collectCSSignalsWithDSC(n, p), index, collectWindowBeginSignalWithDSC(n, p), collectWithinWindowSignalWithDSC(n, p), collectPendingSignalWithDSC(n, p), barriers)
}
