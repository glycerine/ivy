// Package mfs is a line-by-line port of ABC's src/opt/mfs ("the good old
// minimization with complete don't-cares"). It operates directly on the
// base/abc logic network whose local functions are Hop AIGs, exactly as
// the C code does: windows are computed on the network, strashed into an
// AIG, converted to CNF, and solved with the bsat solver; nodes are then
// either resubstituted (interpolation) or re-decomposed (BDC).
//
// The legacy files mfs_.c, mfsCore_.c and mfsResub_.c and the optional
// mfsGia.c are not part of upstream module.make (all mfsGia.c call sites
// are commented out in mfsCore.c / mfsResub.c), so they are not ported.
package mfs

import (
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/bdc"
	"github.com/glycerine/abcgo/misc/vec"
)

// Mfs_Par_t mirrors the C struct field for field.
type Mfs_Par_t struct {
	// general parameters
	NWinTfoLevs  int32 // the maximum fanout levels
	NFanoutsMax  int32 // the maximum number of fanouts
	NDepthMax    int32 // the maximum number of logic levels
	NWinMax      int32 // the maximum size of the window
	NGrowthLevel int32 // the maximum allowed growth in level
	NBTLimit     int32 // the maximum number of conflicts in one SAT run
	FRrOnly      int32 // perform redundance removal
	FResub       int32 // performs resubstitution
	FArea        int32 // performs optimization for area
	FMoreEffort  int32 // performs high-affort minimization
	FSwapEdge    int32 // performs edge swapping
	FOneHotness  int32 // adds one-hotness conditions
	FDelay       int32 // performs optimization for delay
	FPower       int32 // performs power-aware optimization
	FGiaSat      int32 // use new SAT solver
	FVerbose     int32 // enable basic stats
	FVeryVerbose int32 // enable detailed stats
}

// Hooks for C functions that live in base/abci (which imports this package).
// base/abci installs them in its init function.
var (
	// Abc_NodeIfNodeResyn (abcBidec.c)
	NodeIfNodeResyn func(p *bdc.Bdc_Man_t, pHop *hop.Hop_Man_t, pRoot *hop.Hop_Obj_t, nVars int32, vTruth *vec.Vec_Int_t, puCare []uint32, dProb float32) *hop.Hop_Obj_t
	// Abc_NtkPowerEstimate (abcSpeedup.c)
	NtkPowerEstimate func(pNtk *abc.Network, fProbOne int32) *vec.Vec_Int_t
	// Abc_NtkMfsTotalSwitching (abcPrint.c)
	NtkMfsTotalSwitching func(pNtk *abc.Network) float32
	// Abc_NtkStrash (abcStrash.c), used for the external care set
	NtkStrash func(pNtk *abc.Network, fAllNodes, fCleanup, fRecord int32) *abc.Network
)
