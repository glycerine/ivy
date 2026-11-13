package int2

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

type Gia_Man_t = gia.Gia_Man_t
type Gia_Obj_t = gia.Gia_Obj_t
type Vec_Int_t = vec.Vec_Int_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Cnf_Dat_t = cnf.CNF
type abctime = int64

/**CFile****************************************************************

  FileName    [int2Int.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Dec 1, 2013.]

  Revision    [$Id: int2Int.h,v 1.00 2013/12/01 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/gia/gia.h"
//#include "sat/bsat/satSolver.h"
//#include "sat/cnf/cnf.h"
//#include "int2.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// interpolation manager

type Int2_Man_t struct {
	// parameters
	pPars *Int2_ManPars_t // parameters
	// GIA managers
	pGia     *Gia_Man_t // original manager
	pGiaPref *Gia_Man_t // prefix manager
	pGiaSuff *Gia_Man_t // suffix manager
	// subset of the manager
	vSuffCis *Vec_Int_t // suffix CIs
	vSuffCos *Vec_Int_t // suffix COs
	vPrefCos *Vec_Int_t // suffix POs
	vStack   *Vec_Int_t // temporary stack
	// preimages
	vImageOne  *Vec_Int_t // latest preimage
	vImagesAll *Vec_Int_t // cumulative preimage
	// variable maps
	vMapFrames *Vec_Ptr_t // mapping of GIA IDs into frame IDs
	vMapPref   *Vec_Int_t // mapping of flop inputs into SAT variables
	vMapSuff   *Vec_Int_t // mapping of flop outputs into SAT variables
	// initial minimization
	vAssign *Vec_Int_t // assignment of PIs in pGiaSuff
	vPrio   *Vec_Int_t // priority of PIs in pGiaSuff
	// SAT solving
	pSatPref *sat_solver // prefix solver
	pSatSuff *sat_solver // suffix solver
	// runtime
	timeSatPref abctime
	timeSatSuff abctime
	timeOther   abctime
	timeTotal   abctime
}

// Int2_ManCreate is unsupported because the C constructor references an undeclared manager.
func Int2_ManCreate(_ *Gia_Man_t, _ *Int2_ManPars_t) *Int2_Man_t { return nil }

// Int2_GoCreate is a Go-only reconstruction, not an executable C port.
func Int2_GoCreate(pGia *Gia_Man_t, pPars *Int2_ManPars_t) *Int2_Man_t {
	if pGia == nil || pPars == nil {
		panic("int2: nil manager or parameter block")
	}
	p := &Int2_Man_t{pPars: pPars, pGia: pGia}
	p.pGiaPref = gia.Gia_ManStart(10000)
	gia.Gia_ManHashAlloc(p.pGiaPref)
	p.vSuffCis = vec.Vec_IntAlloc(int(gia.Gia_ManCiNum(pGia)))
	p.vSuffCos = vec.Vec_IntAlloc(int(gia.Gia_ManCoNum(pGia)))
	p.vPrefCos = vec.Vec_IntAlloc(int(gia.Gia_ManCoNum(pGia)))
	p.vStack = vec.Vec_IntAlloc(10000)
	p.vImageOne = vec.Vec_IntAlloc(1000)
	p.vImagesAll = vec.Vec_IntAlloc(1000)
	p.vMapFrames = vec.Vec_PtrAlloc(100)
	p.vMapPref = vec.Vec_IntAlloc(int(gia.Gia_ManRegNum(pGia)))
	p.vMapSuff = vec.Vec_IntAlloc(int(gia.Gia_ManRegNum(pGia)))
	p.vAssign = vec.Vec_IntAlloc(int(gia.Gia_ManCiNum(pGia)))
	p.vPrio = vec.Vec_IntAlloc(int(gia.Gia_ManCiNum(pGia)))
	return p
}
func Int2_ManStop(p *Int2_Man_t) {
	if p == nil {
		return
	}
	gia.Gia_ManStopP(&p.pGiaPref)
	gia.Gia_ManStopP(&p.pGiaSuff)
	for i := 0; i < vec.Vec_PtrSize(p.vMapFrames); i++ {
		if frame, ok := vec.Vec_PtrEntry(p.vMapFrames, i).(*vec.Vec_Int_t); ok {
			vec.Vec_IntFree(frame)
		}
	}
	vec.Vec_PtrFreeP(&p.vMapFrames)
	for _, vector := range []**vec.Vec_Int_t{
		&p.vSuffCis, &p.vSuffCos, &p.vPrefCos, &p.vStack,
		&p.vImageOne, &p.vImagesAll, &p.vMapPref, &p.vMapSuff,
		&p.vAssign, &p.vPrio,
	} {
		vec.Vec_IntFreeP(vector)
	}
	// SAT solver state is Go-owned; clearing the references releases it.
	p.pSatPref = nil
	p.pSatSuff = nil
	*p = Int2_Man_t{}
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== int2Bmc.c =============================================================*/
// extern int           Int2_ManCheckInit( Gia_Man_t * p );

// extern Gia_Man_t *   Int2_ManDupInit( Gia_Man_t * p, int fVerbose );

// extern sat_solver *  Int2_ManSetupBmcSolver( Gia_Man_t * p, int nFrames );

// extern void          Int2_ManCreateFrames( Int2_Man_t * p, int iFrame, Vec_Int_t * vPrefCos );

// extern int           Int2_ManCheckBmc( Int2_Man_t * p, Vec_Int_t * vCube );

/*=== int2Refine.c =============================================================*/
// extern Vec_Int_t *   Int2_ManRefineCube( Gia_Man_t * p, Vec_Int_t * vAssign, Vec_Int_t * vPrio );

/*=== int2Util.c ============================================================*/
// extern Gia_Man_t *   Int2_ManProbToGia( Gia_Man_t * p, Vec_Int_t * vSop );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
