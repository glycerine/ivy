package gia

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
)

// Nf_Man_t exposes the mapper state needed by SAT window extraction.
type Nf_Man_t struct {
	matcher *nfMatcher
	refs    *nfMappingRefs
}

func Nf_ManSetDefaultPars(pars *Jf_Par_t) { nfDefaultParameters(pars) }

// Nf_ManPerformMapping uses the current ABC genlib and returns nil when it
// lacks the constant, buffer and inverter cells required by the mapper.
func Nf_ManPerformMapping(g *Gia_Man_t, pars *Jf_Par_t) *Gia_Man_t {
	if g == nil || pars == nil {
		return nil
	}
	cells := mio.Mio_CollectRootsNewDefault2(NF_LEAF_MAX, nil, pars.FVerbose)
	if len(cells) < 4 {
		return nil
	}
	library := nfPrepareLibrary(cells, pars.FPinFilter != 0, pars.FPinPerm != 0, pars.FPinQuick != 0)
	defer vec.Vec_MemFree(library.truth.memory)
	defer vec.Vec_MemHashFree(library.truth.memory)
	return nfMap(g, pars, library)
}

func Nf_ManExtractWindow(man any, roots *Vec_Int_t, cuts, objCuts *Vec_Wec_t, solution, gates *Vec_Int_t, areas *Vec_Wrd_t, invArea *word, startVar, nVars int32) int32 {
	p, ok := man.(*Nf_Man_t)
	if !ok || p == nil || p.matcher == nil || p.refs == nil {
		panic("gia: invalid Nf window manager")
	}
	return p.matcher.nfExtractWindow(p.refs, roots, cuts, objCuts, solution, gates, areas, invArea, startVar, nVars)
}
