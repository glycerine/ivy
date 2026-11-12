package saig

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
)

// Saig_ManReportUselessRegisters reports registers whose sole use is a primary
// output, or which are driven only by a primary input. The test is based on
// the sequential AIG's terminal reference counts, matching saigScl.c.
func Saig_ManReportUselessRegisters(p *aig.Graph) {
	nRegs := aig.Aig_ManRegNum(p)
	nPis := int32(p.InputCount()) - nRegs
	nPos := int32(p.OutputCount()) - nRegs
	var outputs, inputs int32

	for i := int32(0); i < nPos; i++ {
		po := aig.Aig_ManCo(p, i)
		driver := aig.Aig_ObjFanin0(p, po)
		if isRegisterOutput(p, driver, nPis) && aig.Aig_ObjRefs(p, driver) == 1 {
			outputs++
		}
	}
	for i := int32(0); i < nRegs; i++ {
		li := aig.Aig_ManLi(p, i)
		driver := aig.Aig_ObjFanin0(p, li)
		if isPrimaryInput(p, driver, nPis) && aig.Aig_ObjRefs(p, driver) == 1 {
			inputs++
		}
	}
	if outputs != 0 {
		fmt.Printf("Network has %d (out of %d) registers driving POs.\n", outputs, nRegs)
	}
	if inputs != 0 {
		fmt.Printf("Network has %d (out of %d) registers driven by PIs.\n", inputs, nRegs)
	}
}

// Saig_ManReportComplements counts repeated register-input drivers. It restores
// every temporary mark before returning, including on an invalid pre-marked
// graph, so callers never inherit scratch state from this report.
func Saig_ManReportComplements(p *aig.Graph) int32 {
	nRegs := aig.Aig_ManRegNum(p)
	if nRegs <= 0 {
		panic("Saig_ManReportComplements requires at least one register")
	}
	marked := make([]aig.Edge, 0, nRegs)
	defer func() {
		for _, edge := range marked {
			aig.Aig_ObjClearMarkA(p, edge)
		}
	}()

	var repeated int32
	for i := int32(0); i < nRegs; i++ {
		li := aig.Aig_ManLi(p, i)
		driver := aig.Aig_Regular(aig.Aig_ObjFanin0(p, li))
		if aig.Aig_ObjIsMarkA(p, driver) {
			repeated++
			continue
		}
		aig.Aig_ObjSetMarkA(p, driver)
		marked = append(marked, driver)
	}
	return repeated
}

func isRegisterOutput(p *aig.Graph, e aig.Edge, nPis int32) bool {
	return aig.Aig_ObjIsCi(p, e) && aig.Aig_ObjCioId(p, e) >= nPis
}

func isPrimaryInput(p *aig.Graph, e aig.Edge, nPis int32) bool {
	return aig.Aig_ObjIsCi(p, e) && aig.Aig_ObjCioId(p, e) < nPis
}
