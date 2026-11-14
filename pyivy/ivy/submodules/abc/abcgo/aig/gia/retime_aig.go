package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManRetimeClassesAig retimes an AIG and returns its new register classes.
func Gia_ManRetimeClassesAig(graph *aig.Graph, classes *vec.Vec_Int_t, verbose int32) (*aig.Graph, *vec.Vec_Int_t) {
	if graph == nil || classes == nil || vec.Vec_IntSize(classes) != int(graph.RegisterCount()) {
		panic("gia: invalid retiming classes")
	}
	original := Gia_ManFromAigSimple(graph)
	original.VFlopClasses = classes
	retimed := Gia_ManRetimeForward(original, 10, verbose)
	resultClasses := retimed.VFlopClasses
	retimed.VFlopClasses = nil
	result := Gia_ManToAig(retimed, 0)
	if retimed != original {
		Gia_ManStop(retimed)
	}
	original.VFlopClasses = nil
	Gia_ManStop(original)
	return result, resultClasses
}
