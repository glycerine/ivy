package amap

import (
	"fmt"
	"time"
)

// Amap_ManSetDefaultParams resets p and installs ABC's source defaults.
func Amap_ManSetDefaultParams(p *Amap_Par_t) {
	if p == nil {
		panic("amap: nil parameters")
	}
	*p = Amap_Par_t{
		NIterFlow: 1,
		NIterArea: 4,
		NCutsMax:  500,
		FUseXors:  1,
		FEpsilon:  0.001,
	}
}

// Amap_ManTest runs technology mapping using the library selected in ABC's
// process frame, preserving the C command's public entry point.
func Amap_ManTest(pAig *Aig_Man_t, pPars *Amap_Par_t) *Vec_Ptr_t {
	resource := Abc_FrameReadLibGen2()
	pLib, ok := resource.(*Amap_Lib_t)
	if !ok || pLib == nil {
		fmt.Println("Library is not available.")
		return nil
	}
	return Amap_ManMapGraph(pAig, pLib, pPars)
}

// Amap_ManMapGraph is the explicit-library form of Amap_ManTest. It avoids
// process-global library state for callers and tests that already own a cell
// library.
func Amap_ManMapGraph(pAig *Aig_Man_t, pLib *Amap_Lib_t, pPars *Amap_Par_t) *Vec_Ptr_t {
	if pAig == nil || pLib == nil {
		return nil
	}
	if pPars == nil {
		pPars = &Amap_Par_t{}
		Amap_ManSetDefaultParams(pPars)
	}
	started := time.Now()
	p := Amap_ManStart(int32(pAig.NodeCount()))
	p.pPars, p.pLib = pPars, pLib
	if pPars.FFreeInvs != 0 {
		p.fAreaInv = 0
	} else if pLib.pGateInv != nil {
		p.fAreaInv = float32(pLib.pGateInv.dArea)
	}
	p.fUseMux = amapBool(pPars.FUseMuxes != 0 && pLib.fHasMux != 0)
	p.fUseXor = amapBool(pPars.FUseXors != 0 && pLib.fHasXor != 0)
	tempSize := max(0, int(pLib.nNodes)*2)
	p.ppCutsTemp = make([]*Amap_Cut_t, tempSize)
	Amap_ManCreate(p, pAig)
	Amap_ManMap(p)
	result := Amap_ManProduceMapped(p)
	Amap_ManStop(p)
	if pPars.FVerbose != 0 {
		fmt.Printf("Total runtime: %.2f sec\n", time.Since(started).Seconds())
	}
	return result
}
