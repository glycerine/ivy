package int

// Interpolant validation helpers ported from proof/int/intInter.c.
import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Inter_ManDupExpand duplicates an interpolant and appends unused inputs to
// match the comparison graph, preserving the source input order.
func Inter_ManDupExpand(interpolant, other *aig.Graph) *aig.Graph {
	if interpolant == nil || other == nil || interpolant.InputCount() > other.InputCount() {
		panic("interpolant has too many inputs")
	}
	result := aig.Aig_ManDupSimple(interpolant)
	for result.InputCount() < other.InputCount() {
		result.NewInput()
	}
	return result
}

func verifyInterpolant(lower, upper, interpolant *aig.Graph, name string) {
	aig.Aig_ManFlipFirstPo(upper)
	lowerValid := Inter_ManCheckContainment(lower, Inter_ManDupExpand(interpolant, lower))
	upperValid := Inter_ManCheckContainment(Inter_ManDupExpand(interpolant, upper), upper)
	if lowerValid == 1 && upperValid == 1 {
		fmt.Printf("%s is correct.\n", name)
	}
	if lowerValid == 0 {
		fmt.Printf("Property A => %s fails.\n", name)
	}
	if upperValid == 0 {
		fmt.Printf("Property %s => !B fails.\n", name)
	}
	if lowerValid < 0 || upperValid < 0 {
		fmt.Printf("Could not verify %s: containment check was undecided.\n", name)
	}
}

func Inter_ManVerifyInterpolant1(manager *bsat.Inta_Man_t, proof *bsat.Sto_Man_t, interpolant *aig.Graph) {
	verifyInterpolant(bsat.Inta_ManDeriveClauses(manager, proof, 1), bsat.Inta_ManDeriveClauses(manager, proof, 0), interpolant, "Im")
}

func Inter_ManVerifyInterpolant2(manager *bsat.Intb_Man_t, proof *bsat.Sto_Man_t, interpolant *aig.Graph) {
	verifyInterpolant(bsat.Intb_ManDeriveClauses(manager, proof, 1), bsat.Intb_ManDeriveClauses(manager, proof, 0), interpolant, "Ip")
}
