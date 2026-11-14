package gia

import "fmt"

func giaCountDomain(classes *Vec_Int_t, domain int32) int32 {
	if classes == nil {
		return 0
	}
	return int32(Vec_IntCountEntry(classes, domain))
}
func giaSweepReport(domain, before, after int32, kinds [3]int32) {
	kinds[2] = before - after - kinds[0] - kinds[1]
	fmt.Printf("Domain %2d : %5d -> %5d :  EqConst =%4d.  EqFlop =%4d.  Dangling =%4d.  Unused =%4d.\n", domain, before, after, kinds[0], kinds[1], max(0, kinds[2]), max(0, -kinds[2]))
}
func giaSweepReduceBoxes(original, graph, collapsed *Gia_Man_t, verbose int32, sequential int32) (*Gia_Man_t, [3]int32) {
	var kinds [3]int32
	reprs := Gia_ManFraigSelectReprs(graph, collapsed, verbose, kinds[:])
	Gia_ManStop(collapsed)
	Gia_ManTransferTiming(original, graph)
	reduced := Gia_ManFraigReduceGia(graph, reprs)
	Gia_ManTransferTiming(reduced, original)
	Gia_ManStop(graph)
	result := Gia_ManDupWithBoxes(reduced, sequential)
	Gia_ManStop(reduced)
	return result, kinds
}
func giaSweepNormalize(graph *Gia_Man_t) *Gia_Man_t {
	result := Gia_ManDupNormalize(graph, 0)
	Gia_ManTransferTiming(result, graph)
	Gia_ManStop(graph)
	return result
}

func Gia_ManSweepWithBoxesAndDomains(p *Gia_Man_t, sequentialParameters any, constant, equivalent, verbose, verboseEquivs int32) *Gia_Man_t {
	if Gia_ManRegNum(p) != 0 || p.PAigExtra == nil || p.vRegClasses == nil || Vec_IntFindMax(p.vRegClasses) <= 1 {
		panic("gia: invalid multi-domain box sweep")
	}
	domains := Vec_IntFindMax(p.vRegClasses)
	graph := Gia_ManDupUnnormalize(p)
	if graph == nil {
		return nil
	}
	Gia_ManTransferTiming(graph, p)
	for domain := int32(1); domain <= domains; domain++ {
		before := giaCountDomain(graph.vRegClasses, domain)
		if before < 2 {
			continue
		}
		collapsed := Gia_ManDupCollapse(graph, graph.PAigExtra, nil, 1)
		Gia_ManSweepComputeOneDomainEquivs(collapsed, graph.vRegClasses, domain, sequentialParameters, constant, equivalent, verbose)
		var kinds [3]int32
		graph, kinds = giaSweepReduceBoxes(p, graph, collapsed, verbose, 1)
		if verboseEquivs != 0 {
			giaSweepReport(domain, before, giaCountDomain(graph.vRegClasses, domain), kinds)
		}
	}
	return giaSweepNormalize(graph)
}

func Gia_ManSweepWithBoxes(p *Gia_Man_t, combinationalParameters, sequentialParameters any, constant, equivalent, verbose, verboseEquivs int32) *Gia_Man_t {
	if Gia_ManRegNum(p) != 0 || p.PAigExtra == nil {
		panic("gia: invalid boxed sweep")
	}
	if combinationalParameters == nil && Gia_ManClockDomainNum(p) > 1 {
		return Gia_ManSweepWithBoxesAndDomains(p, sequentialParameters, constant, equivalent, verbose, verboseEquivs)
	}
	graph := Gia_ManDupUnnormalize(p)
	if graph == nil {
		return nil
	}
	Gia_ManTransferTiming(graph, p)
	before := giaCountDomain(graph.vRegClasses, 1)
	sequential := int32(1)
	if combinationalParameters != nil {
		sequential = 0
	}
	collapsed := Gia_ManDupCollapse(graph, graph.PAigExtra, nil, sequential)
	switch {
	case combinationalParameters != nil:
		Gia_ManFraigSweepPerform(collapsed, combinationalParameters)
	case sequentialParameters != nil:
		engine, ok := sequentialParameters.(SweepCorrespondence)
		if !ok {
			panic("gia: invalid sequential correspondence engine")
		}
		engine.SweepCorrespondence(collapsed)
	default:
		Gia_ManSeqCleanupClasses(collapsed, constant, equivalent, verbose)
	}
	var kinds [3]int32
	graph, kinds = giaSweepReduceBoxes(p, graph, collapsed, verbose, sequential)
	if verboseEquivs != 0 {
		giaSweepReport(1, before, giaCountDomain(graph.vRegClasses, 1), kinds)
	}
	return giaSweepNormalize(graph)
}
