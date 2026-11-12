package cov

import (
	"sync"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

var managerRegistry sync.Map // map[*abc.Network]*Cov_Man_t; guarded for concurrent networks

/**CFile****************************************************************

  FileName    [cov.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Mapping into network of SOPs/ESOPs.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cov.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "covInt.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Cov_Man_t = Cov_Man_t_

type Abc_Ntk_t = abc.Network
type Abc_Obj_t = abc.Object
type Vec_Int_t = vec.Vec_Int_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Vec_Str_t = vec.Vec_Str_t

// storage for node information
type Cov_Obj_t struct {
	pCover [3]*Min_Cube_t // pos/neg/esop
	vSupp  *Vec_Int_t     // computed support (all nodes except CIs)
}

// storage for additional information
type Cov_Man_t_ struct {
	// general characteristics
	nFaninMax  int32      // the number of vars
	nCubesMax  int32      // the limit on the number of cubes in intermediate covers
	nWords     int32      // the number of words
	vFanCounts *Vec_Int_t // fanout counts
	vObjStrs   *Vec_Ptr_t // object structures (legacy API view)
	pMemory    any        // owned Go storage for object structures
	pManMin    *Min_Man_t // the cube manager
	fUseEsop   int32      // enables ESOPs
	fUseSop    int32      // enables SOPs
	// arrays to map local variables
	vComTo0 *Vec_Int_t // mapping of common variables into first fanin
	vComTo1 *Vec_Int_t // mapping of common variables into second fanin
	vPairs0 *Vec_Int_t // the first var in each pair of common vars
	vPairs1 *Vec_Int_t // the second var in each pair of common vars
	vTriv0  *Vec_Int_t // trivial support of the first node
	vTriv1  *Vec_Int_t // trivial support of the second node
	// statistics
	nSupps     int32 // supports created
	nSuppsMax  int32 // the maximum number of supports
	nBoundary  int32 // the boundary size
	nNodes     int32 // the number of nodes processed
	objectInfo []*Cov_Obj_t
	network    *abc.Network
}

func Abc_ObjGetStr(pObj *Abc_Obj_t) *Cov_Obj_t {
	if pObj == nil || pObj.Network() == nil {
		return nil
	}
	value, ok := managerRegistry.Load(pObj.Network())
	if !ok {
		return nil
	}
	p := value.(*Cov_Man_t)
	if pObj.ID < 0 || pObj.ID >= len(p.objectInfo) {
		return nil
	}
	return p.objectInfo[pObj.ID]
}

func Abc_ObjSetSupp(pObj *Abc_Obj_t, vVec *Vec_Int_t) { Abc_ObjGetStr(pObj).vSupp = vVec }
func Abc_ObjGetSupp(pObj *Abc_Obj_t) *Vec_Int_t       { return Abc_ObjGetStr(pObj).vSupp }

func Abc_ObjSetCover2(pObj *Abc_Obj_t, pCov *Min_Cube_t) { Abc_ObjGetStr(pObj).pCover[2] = pCov }
func Abc_ObjGetCover2(pObj *Abc_Obj_t) *Min_Cube_t       { return Abc_ObjGetStr(pObj).pCover[2] }

func Abc_ObjSetCover(pObj *Abc_Obj_t, pCov *Min_Cube_t, Pol int32) {
	Abc_ObjGetStr(pObj).pCover[Pol] = pCov
}
func Abc_ObjGetCover(pObj *Abc_Obj_t, Pol int32) *Min_Cube_t { return Abc_ObjGetStr(pObj).pCover[Pol] }

func Vec_IntAlloc(n int32) *Vec_Int_t                { return vec.Vec_IntAlloc(int(n)) }
func Vec_IntStart(n int32) *Vec_Int_t                { return vec.Vec_IntStart(int(n)) }
func Vec_IntSize(v *Vec_Int_t) int32                 { return int32(vec.Vec_IntSize(v)) }
func Vec_IntEntry(v *Vec_Int_t, i int32) int32       { return vec.Vec_IntEntry(v, int(i)) }
func Vec_IntWriteEntry(v *Vec_Int_t, i, x int32)     { vec.Vec_IntWriteEntry(v, int(i), x) }
func Vec_IntPush(v *Vec_Int_t, x int32)              { vec.Vec_IntPush(v, x) }
func Vec_IntClear(v *Vec_Int_t)                      { vec.Vec_IntClear(v) }
func Vec_IntFill(v *Vec_Int_t, n, x int32)           { vec.Vec_IntFill(v, int(n), x) }
func Vec_IntFree(v *Vec_Int_t)                       { vec.Vec_IntFree(v) }
func Vec_PtrAlloc(n int32) *Vec_Ptr_t                { return vec.Vec_PtrAlloc(int(n)) }
func Vec_PtrStart(n int32) *Vec_Ptr_t                { return vec.Vec_PtrStart(int(n)) }
func Vec_PtrSize(v *Vec_Ptr_t) int32                 { return int32(vec.Vec_PtrSize(v)) }
func Vec_PtrEntry(v *Vec_Ptr_t, i int32) any         { return vec.Vec_PtrEntry(v, int(i)) }
func Vec_PtrWriteEntry(v *Vec_Ptr_t, i int32, x any) { vec.Vec_PtrWriteEntry(v, int(i), x) }
func Vec_PtrPush(v *Vec_Ptr_t, x any)                { vec.Vec_PtrPush(v, x) }
func Vec_PtrFree(v *Vec_Ptr_t)                       { vec.Vec_PtrFree(v) }
func Vec_StrAlloc(n int32) *Vec_Str_t                { return vec.Vec_StrAlloc(int(n)) }
func Vec_StrSize(v *Vec_Str_t) int32                 { return int32(vec.Vec_StrSize(v)) }
func Vec_StrArray(v *Vec_Str_t) []byte               { return vec.Vec_StrArray(v) }
func Vec_StrPush(v *Vec_Str_t, x byte)               { vec.Vec_StrPush(v, x) }
func Vec_StrClear(v *Vec_Str_t)                      { vec.Vec_StrClear(v) }
func Vec_StrFree(v *Vec_Str_t)                       { vec.Vec_StrFree(v) }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== covBuild.c ==========================================================*/
// extern Abc_Ntk_t * Abc_NtkCovDerive( Cov_Man_t * p, Abc_Ntk_t * pNtk );

// extern Abc_Ntk_t * Abc_NtkCovDeriveClean( Cov_Man_t * p, Abc_Ntk_t * pNtk );

// extern Abc_Ntk_t * Abc_NtkCovDeriveRegular( Cov_Man_t * p, Abc_Ntk_t * pNtk );

/*=== covCore.c ===========================================================*/
// extern Abc_Ntk_t * Abc_NtkSopEsopCover( Abc_Ntk_t * pNtk, int nFaninMax, int nCubesMax, int fUseEsop, int fUseSop, int fUseInvs, int fVerbose );

/*=== covMan.c ============================================================*/
// extern Cov_Man_t * Cov_ManAlloc( Abc_Ntk_t * pNtk, int nFaninMax, int nCubesMax );

// extern void        Cov_ManFree( Cov_Man_t * p );

// extern void        Abc_NodeCovDropData( Cov_Man_t * p, Abc_Obj_t * pObj );

/*=== covTest.c ===========================================================*/
// extern Abc_Ntk_t * Abc_NtkCovTestSop( Abc_Ntk_t * pNtk );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
