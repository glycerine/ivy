package dau

import "github.com/glycerine/abcgo/aig/aig"

// dauCore.c is ABC's unchanged Xyz example-manager template. It contains no
// unmapping algorithm; the only defined behavior is allocating and releasing
// this pair of borrowed references.
type Xyz_ParTry_t struct{ Par int32 }
type Xyz_ManTry_t struct {
	pPar *Xyz_ParTry_t
	pAig *aig.Graph
}

func Xyz_ManTryAlloc(graph *aig.Graph, params *Xyz_ParTry_t) *Xyz_ManTry_t {
	return &Xyz_ManTry_t{pPar: params, pAig: graph}
}
func Xyz_ManTryFree(p *Xyz_ManTry_t) {
	if p != nil {
		p.pPar = nil
		p.pAig = nil
	}
}

// Xyz_ManPerform reports successful manager creation and leaves the graph
// untouched, exactly as the original template does.
func Xyz_ManPerform(graph *aig.Graph, params *Xyz_ParTry_t) int32 {
	p := Xyz_ManTryAlloc(graph, params)
	Xyz_ManTryFree(p)
	return 1
}
