package abc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Abc_NtkRetimeWithClassesAig retimes an AIG and transfers its classes.
func Abc_NtkRetimeWithClassesAig(graph *aig.Graph, classes *vec.Vec_Int_t, newClasses **vec.Vec_Int_t, verbose int32) *aig.Graph {
	result, updated := gia.Gia_ManRetimeClassesAig(graph, classes, verbose)
	*newClasses = updated
	return result
}
func Abc_NtkRetimeWithClassesNtk(n *Network, classes *vec.Vec_Int_t, newClasses **vec.Vec_Int_t, verbose int32) *Network {
	strashed := Abc_NtkStrash(n, 0, 1, 0)
	graph := Abc_NtkToDar(strashed, 0, 1)
	retimed := Abc_NtkRetimeWithClassesAig(graph, classes, newClasses, verbose)
	rebuilt := Abc_NtkFromDarSeqSweep(strashed, retimed)
	result := Abc_NtkToLogic(rebuilt)
	Abc_NtkDelete(rebuilt)
	Abc_NtkDelete(strashed)
	aig.Aig_ManStop(retimed)
	aig.Aig_ManStop(graph)
	return result
}
func Abc_NtkCRetime(n *Network, verbose int32) *Network {
	old := Abc_NtkDup(n)
	controls, classes, counts := abcClassifyLatchControls(old)
	if counts.opposite != 0 {
		fmt.Printf("Opposite phase enable is present in %d flops (out of %d).\n", counts.opposite, Abc_NtkLatchNum(old))
	}
	if verbose != 0 {
		fmt.Printf("CountN = %4d. Count2 = %4d. Count1 = %4d. Count0 = %4d. Ctrls = %d.\n", counts.notMux, counts.other, counts.opposite, counts.enabled, vec.Vec_PtrSize(controls))
	}
	abcExposeLatchControls(old, controls, verbose)
	var updated *vec.Vec_Int_t
	result := Abc_NtkRetimeWithClassesNtk(old, classes, &updated, verbose)
	Abc_NtkTransformBack(old, result, controls, updated)
	Abc_NtkDelete(old)
	return result
}
