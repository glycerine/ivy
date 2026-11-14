package wln

func wlnBool(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

/**CFile****************************************************************

  FileName    [wlc.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlc.h,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/gia/gia.h"
//#include "misc/vec/vecHash.h"
//#include "misc/extra/extra.h"
//#include "misc/util/utilNam.h"
//#include "misc/util/utilTruth.h"
//#include "aig/miniaig/abcOper.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Wln_Ntk_t struct {
	pName       string               // model name
	pSpec       string               // input file name
	fSmtLib     int32                // derived from SMT-LIB
	vCis        Vec_Int_t            // combinational inputs
	vCos        Vec_Int_t            // combinational outputs
	vFfs        Vec_Int_t            // flops
	vTypes      Vec_Int_t            // object types
	vFanins     []Wln_Vec_t          // object fanins (exceptions: const, select)
	vRanges     Vec_Int_t            // object ranges
	pRanges     *Hash_IntMan_t       // object ranges
	vNameIds    Vec_Int_t            // object name IDs
	vInstIds    Vec_Int_t            // object name IDs
	pManName    *Abc_Nam_t           // object names
	vSigns      Vec_Str_t            // object signedness
	nTravIds    int32                // counter of traversal IDs
	vTravIds    Vec_Int_t            // trav IDs of the objects
	vCopies     Vec_Int_t            // object first bits
	vBits       Vec_Int_t            // object mapping into AIG nodes
	vLevels     Vec_Int_t            // object levels
	vRefs       Vec_Int_t            // object reference counters
	vFanout     Vec_Int_t            // static fanout
	vFaninAttrs Vec_Int_t            // static fanin attributes
	vFaninLists Vec_Int_t            // static fanin attributes
	vTables     *Vec_Ptr_t           // tables
	nObjs       [ABC_OPER_LAST]int32 // counter of objects of each type
	nAnds       [ABC_OPER_LAST]int32 // counter of AND gates after blasting
}

func Wln_NtkObjNum(p *Wln_Ntk_t) int32 { return Vec_IntSize(&p.vTypes) }
func Wln_NtkCiNum(p *Wln_Ntk_t) int32  { return Vec_IntSize(&p.vCis) }
func Wln_NtkCoNum(p *Wln_Ntk_t) int32  { return Vec_IntSize(&p.vCos) }
func Wln_NtkFfNum(p *Wln_Ntk_t) int32  { return Vec_IntSize(&p.vFfs) }
func Wln_NtkPiNum(p *Wln_Ntk_t) int32  { return Wln_NtkCiNum(p) - Wln_NtkFfNum(p) }
func Wln_NtkPoNum(p *Wln_Ntk_t) int32  { return Wln_NtkCoNum(p) - Wln_NtkFfNum(p) }

func Wln_NtkCi(p *Wln_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vCis, i) }
func Wln_NtkCo(p *Wln_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vCos, i) }
func Wln_NtkFf(p *Wln_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vFfs, i) }

func Wln_ObjType(p *Wln_Ntk_t, i int32) int32   { return Vec_IntEntry(&p.vTypes, i) }
func Wln_ObjIsNone(p *Wln_Ntk_t, i int32) int32 { return wlnBool(Wln_ObjType(p, i) == ABC_OPER_NONE) }
func Wln_ObjIsCi(p *Wln_Ntk_t, i int32) int32   { return wlnBool(Wln_ObjType(p, i) == ABC_OPER_CI) }
func Wln_ObjIsCo(p *Wln_Ntk_t, i int32) int32   { return wlnBool(Wln_ObjType(p, i) == ABC_OPER_CO) }
func Wln_ObjIsCio(p *Wln_Ntk_t, i int32) int32 {
	return wlnBool(Wln_ObjType(p, i) == ABC_OPER_CI || Wln_ObjType(p, i) == ABC_OPER_CO)
}
func Wln_ObjIsFon(p *Wln_Ntk_t, i int32) int32   { return wlnBool(Wln_ObjType(p, i) == ABC_OPER_FON) }
func Wln_ObjIsFf(p *Wln_Ntk_t, i int32) int32    { return wlnBool(Wln_ObjType(p, i) == ABC_OPER_DFFRSE) }
func Wln_ObjIsConst(p *Wln_Ntk_t, i int32) int32 { return wlnBool(Wln_ObjType(p, i) == ABC_OPER_CONST) }
func Wln_ObjIsSlice(p *Wln_Ntk_t, i int32) int32 { return wlnBool(Wln_ObjType(p, i) == ABC_OPER_SLICE) }
func Wln_ObjIsRotate(p *Wln_Ntk_t, i int32) int32 {
	return wlnBool(Wln_ObjType(p, i) == ABC_OPER_SHIFT_ROTL || Wln_ObjType(p, i) == ABC_OPER_SHIFT_ROTR)
}
func Wln_ObjIsTable(p *Wln_Ntk_t, i int32) int32 { return wlnBool(Wln_ObjType(p, i) == ABC_OPER_TABLE) }

func Wln_ObjFaninNum(p *Wln_Ntk_t, i int32) int32 { return p.vFanins[i].nSize }
func Wln_ObjFanins(p *Wln_Ntk_t, i int32) []int32 { return p.vFanins[i].fanins() }
func Wln_ObjFanin(p *Wln_Ntk_t, i, f int32) int32 { return Wln_ObjFanins(p, i)[f] }
func Wln_ObjSetFanin(p *Wln_Ntk_t, i, f, v int32) { Wln_ObjFanins(p, i)[f] = v }
func Wln_ObjFanin0(p *Wln_Ntk_t, i int32) int32   { return Wln_ObjFanin(p, i, 0) }
func Wln_ObjFanin1(p *Wln_Ntk_t, i int32) int32   { return Wln_ObjFanin(p, i, 1) }
func Wln_ObjFanin2(p *Wln_Ntk_t, i int32) int32   { return Wln_ObjFanin(p, i, 2) }

func Wln_ObjRangeId(p *Wln_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vRanges, i) }
func Wln_ObjRangeEnd(p *Wln_Ntk_t, i int32) int32 {
	return Hash_IntObjData0(p.pRanges, Wln_ObjRangeId(p, i))
}
func Wln_ObjRangeBeg(p *Wln_Ntk_t, i int32) int32 {
	return Hash_IntObjData1(p.pRanges, Wln_ObjRangeId(p, i))
}
func Wln_ObjRangeIsReversed(p *Wln_Ntk_t, i int32) int32 {
	return wlnBool(Wln_ObjRangeEnd(p, i) < Wln_ObjRangeBeg(p, i))
}
func Wln_ObjRange(p *Wln_Ntk_t, i int32) int32 {
	return 1 + Abc_AbsInt(Wln_ObjRangeEnd(p, i)-Wln_ObjRangeBeg(p, i))
}

func Wln_ObjIsSigned(p *Wln_Ntk_t, i int32) int32 { return int32(Vec_StrEntry(&p.vSigns, i)) }
func Wln_ObjSetSigned(p *Wln_Ntk_t, i int32)      { Vec_StrSetEntry(&p.vSigns, i, byte(1)) }
func Wln_ObjIsSignedFanin0(p *Wln_Ntk_t, i int32) int32 {
	if p.fSmtLib != 0 {
		return Wln_ObjIsSigned(p, i)
	}
	return Wln_ObjIsSigned(p, Wln_ObjFanin0(p, i))
}
func Wln_ObjIsSignedFanin1(p *Wln_Ntk_t, i int32) int32 {
	if p.fSmtLib != 0 {
		return Wln_ObjIsSigned(p, i)
	}
	return Wln_ObjIsSigned(p, Wln_ObjFanin1(p, i))
}
func Wln_ObjIsSignedFanin01(p *Wln_Ntk_t, i int32) int32 {
	return wlnBool(Wln_ObjIsSignedFanin0(p, i) != 0 && Wln_ObjIsSignedFanin1(p, i) != 0)
}
func Wln_ObjSign(p *Wln_Ntk_t, i int32) int32 {
	return Abc_Var2Lit(Wln_ObjRange(p, i), Wln_ObjIsSigned(p, i))
}

func Wln_NtkCleanCopy(p *Wln_Ntk_t)                 { Vec_IntFill(&p.vCopies, Vec_IntCap(&p.vTypes), 0) }
func Wln_NtkHasCopy(p *Wln_Ntk_t) int32             { return wlnBool(Vec_IntSize(&p.vCopies) > 0) }
func Wln_ObjSetCopy(p *Wln_Ntk_t, i int32, c int32) { Vec_IntWriteEntry(&p.vCopies, i, c) }
func Wln_ObjCopy(p *Wln_Ntk_t, i int32) int32       { return Vec_IntEntry(&p.vCopies, i) }

func Wln_NtkCleanLevel(p *Wln_Ntk_t)                 { Vec_IntFill(&p.vLevels, Vec_IntCap(&p.vTypes), 0) }
func Wln_NtkHasLevel(p *Wln_Ntk_t) int32             { return wlnBool(Vec_IntSize(&p.vLevels) > 0) }
func Wln_ObjSetLevel(p *Wln_Ntk_t, i int32, l int32) { Vec_IntWriteEntry(&p.vLevels, i, l) }
func Wln_ObjLevel(p *Wln_Ntk_t, i int32) int32       { return Vec_IntEntry(&p.vLevels, i) }

func Wln_NtkCleanNameId(p *Wln_Ntk_t)                 { Vec_IntFill(&p.vNameIds, Vec_IntCap(&p.vTypes), 0) }
func Wln_NtkHasNameId(p *Wln_Ntk_t) int32             { return wlnBool(Vec_IntSize(&p.vNameIds) > 0) }
func Wln_ObjSetNameId(p *Wln_Ntk_t, i int32, n int32) { Vec_IntWriteEntry(&p.vNameIds, i, n) }
func Wln_ObjNameId(p *Wln_Ntk_t, i int32) int32       { return Vec_IntEntry(&p.vNameIds, i) }

func Wln_NtkCleanInstId(p *Wln_Ntk_t)                 { Vec_IntFill(&p.vInstIds, Vec_IntCap(&p.vTypes), 0) }
func Wln_NtkHasInstId(p *Wln_Ntk_t) int32             { return wlnBool(Vec_IntSize(&p.vInstIds) > 0) }
func Wln_ObjSetInstId(p *Wln_Ntk_t, i int32, n int32) { Vec_IntWriteEntry(&p.vInstIds, i, n) }
func Wln_ObjInstId(p *Wln_Ntk_t, i int32) int32       { return Vec_IntEntry(&p.vInstIds, i) }

func Wln_NtkCleanRefs(p *Wln_Ntk_t)                 { Vec_IntFill(&p.vRefs, Vec_IntCap(&p.vTypes), 0) }
func Wln_NtkHasRefs(p *Wln_Ntk_t) int32             { return wlnBool(Vec_IntSize(&p.vRefs) > 0) }
func Wln_ObjSetRefs(p *Wln_Ntk_t, i int32, n int32) { Vec_IntWriteEntry(&p.vRefs, i, n) }
func Wln_ObjRefs(p *Wln_Ntk_t, i int32) int32       { return Vec_IntEntry(&p.vRefs, i) }
func Wln_ObjRefsInc(p *Wln_Ntk_t, i int32) int32 {
	v := Vec_IntEntryP(&p.vRefs, i)
	old := *v
	*v++
	return old
}
func Wln_ObjRefsDec(p *Wln_Ntk_t, i int32) int32         { v := Vec_IntEntryP(&p.vRefs, i); *v--; return *v }
func Wln_ObjRefsFaninInc(p *Wln_Ntk_t, i int32, k int32) { Wln_ObjRefsInc(p, Wln_ObjFanin(p, i, k)) }
func Wln_ObjRefsFaninDec(p *Wln_Ntk_t, i int32, k int32) { Wln_ObjRefsDec(p, Wln_ObjFanin(p, i, k)) }

func Wln_ObjFanoutNum(p *Wln_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vRefs, i) }
func Wln_ObjFanouts(p *Wln_Ntk_t, i int32) []int32 {
	return wlnIntArray(&p.vFanout)[Vec_IntEntry(&p.vFanout, i):]
}
func Wln_ObjFanout(p *Wln_Ntk_t, i int32, f int32) int32       { return Wln_ObjFanouts(p, i)[f] }
func Wln_ObjSetFanout(p *Wln_Ntk_t, i int32, f int32, v int32) { Wln_ObjFanouts(p, i)[f] = v }

func Wln_NtkIncrementTravId(p *Wln_Ntk_t) {
	old := p.nTravIds
	p.nTravIds++
	if old == 0 {
		Vec_IntFill(&p.vTravIds, Vec_IntCap(&p.vTypes), 0)
	}
}
func Wln_ObjSetTravIdCurrent(p *Wln_Ntk_t, i int32)  { Vec_IntWriteEntry(&p.vTravIds, i, p.nTravIds) }
func Wln_ObjSetTravIdPrevious(p *Wln_Ntk_t, i int32) { Vec_IntWriteEntry(&p.vTravIds, i, p.nTravIds-1) }
func Wln_ObjIsTravIdCurrent(p *Wln_Ntk_t, i int32) int32 {
	return wlnBool((Vec_IntEntry(&p.vTravIds, i) == p.nTravIds))
}
func Wln_ObjIsTravIdPrevious(p *Wln_Ntk_t, i int32) int32 {
	return wlnBool((Vec_IntEntry(&p.vTravIds, i) == p.nTravIds-1))
}
func Wln_ObjCheckTravId(p *Wln_Ntk_t, i int32) int32 {
	if Wln_ObjIsTravIdCurrent(p, i) != 0 {
		return 1
	}
	Wln_ObjSetTravIdCurrent(p, i)
	return 0
}

func Wln_ObjCioId(p *Wln_Ntk_t, i int32) int32 {
	assert((Wln_ObjIsCio(p, i) != 0), "Wln_ObjIsCio(p, i)")
	return Wln_ObjFanin1(p, i)
}
func Wln_ObjIsPi(p *Wln_Ntk_t, i int32) int32 {
	return wlnBool(Wln_ObjIsCi(p, i) != 0 && Wln_ObjCioId(p, i) < Wln_NtkPiNum(p))
}
func Wln_ObjIsPo(p *Wln_Ntk_t, i int32) int32 {
	return wlnBool(Wln_ObjIsCo(p, i) != 0 && Wln_ObjCioId(p, i) < Wln_NtkPoNum(p))
}
func Wln_ObjIsRo(p *Wln_Ntk_t, i int32) int32 {
	return wlnBool(Wln_ObjIsCi(p, i) != 0 && Wln_ObjCioId(p, i) >= Wln_NtkPiNum(p))
}
func Wln_ObjIsRi(p *Wln_Ntk_t, i int32) int32 {
	return wlnBool(Wln_ObjIsCo(p, i) != 0 && Wln_ObjCioId(p, i) >= Wln_NtkPoNum(p))
}
func Wln_ObjRoToRi(p *Wln_Ntk_t, i int32) int32 {
	assert((Wln_ObjIsRo(p, i) != 0), "Wln_ObjIsRo(p, i)")
	return Wln_NtkCo(p, Wln_NtkCoNum(p)-Wln_NtkCiNum(p)+Wln_ObjCioId(p, i))
}
func Wln_ObjRiToRo(p *Wln_Ntk_t, i int32) int32 {
	assert((Wln_ObjIsRi(p, i) != 0), "Wln_ObjIsRi(p, i)")
	return Wln_NtkCi(p, Wln_NtkCiNum(p)-Wln_NtkCoNum(p)+Wln_ObjCioId(p, i))
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Wln_NtkForEachObj( p, i )                                                   \
//     for ( i = 1; i < Wln_NtkObjNum(p); i++ )

// Original macro:
// #define Wln_NtkForEachObjReverse( p, i )                                            \
//     for ( i = Wln_NtkObjNum(p) - 1; i >  0; i-- )

// Original macro:
// #define Wln_NtkForEachObjInternal( p, i )                                           \
//     for ( i = 1; i < Wln_NtkObjNum(p); i++ ) if ( Wln_ObjIsCio(p, i) ) {} else

// Original macro:
// #define Wln_NtkForEachPi( p, iPi, i )                                               \
//     for ( i = 0; (i < Wln_NtkPiNum(p)) && (((iPi) = Wln_NtkCi(p, i)), 1); i++ )

// Original macro:
// #define Wln_NtkForEachPo( p, iPo, i )                                               \
//     for ( i = 0; (i < Wln_NtkPoNum(p)) && (((iPo) = Wln_NtkCo(p, i)), 1); i++ )

// Original macro:
// #define Wln_NtkForEachCi( p, iCi, i )                                               \
//     for ( i = 0; (i < Wln_NtkCiNum(p)) && (((iCi) = Wln_NtkCi(p, i)), 1); i++ )

// Original macro:
// #define Wln_NtkForEachCo( p, iCo, i )                                               \
//     for ( i = 0; (i < Wln_NtkCoNum(p)) && (((iCo) = Wln_NtkCo(p, i)), 1); i++ )

// Original macro:
// #define Wln_NtkForEachFf( p, iFf, i )                                               \
//     for ( i = 0; (i < Wln_NtkFfNum(p)) && (((iFf) = Wln_NtkFf(p, i)), 1); i++ )

// Original macro:
// #define Wln_ObjForEachFanin( p, iObj, iFanin, i )                                   \
//     for ( i = 0; (i < Wln_ObjFaninNum(p, iObj)) && (((iFanin) = Wln_ObjFanin(p, iObj, i)), 1); i++ ) if ( !iFanin ) {} else

// Original macro:
// #define Wln_ObjForEachFaninReverse( pObj, iFanin, i )                               \
//     for ( i = Wln_ObjFaninNum(p, iObj) - 1; (i >= 0) && (((iFanin) = Wln_ObjFanin(p, iObj, i)), 1); i-- ) if ( !iFanin ) {} else

// Original macro:
// #define Wln_ObjForEachFanoutStatic( p, iObj, iFanout, i )                           \
//     for ( i = 0; (i < Wln_ObjRefs(p, iObj)) && (((iFanout) = Wln_ObjFanout(p, iObj, i)), 1); i++ )

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== wlcNdr.c ========================================================*/
// extern Wln_Ntk_t *    Wln_ReadNdr( char * pFileName );

// extern void           Wln_WriteNdr( Wln_Ntk_t * pNtk, char * pFileName );

// extern Wln_Ntk_t *    Wln_NtkFromNdr( void * pData, int fDump );

// extern void *         Wln_NtkToNdr( Wln_Ntk_t * pNtk );

/*=== wlcNtk.c ========================================================*/
// extern Wln_Ntk_t *    Wln_NtkAlloc( char * pName, int nObjsMax );

// extern void           Wln_NtkFree( Wln_Ntk_t * p );

// extern int            Wln_NtkMemUsage( Wln_Ntk_t * p );

// extern void           Wln_NtkPrint( Wln_Ntk_t * p );

// extern Wln_Ntk_t *    Wln_NtkDupDfs( Wln_Ntk_t * p );

// extern int            Wln_NtkIsAcyclic( Wln_Ntk_t * p );

// extern void           Wln_NtkCreateRefs( Wln_Ntk_t * p );

// extern void           Wln_NtkStartFaninMap( Wln_Ntk_t * p, Vec_Int_t * vFaninMap, int nMulti );

// extern void           Wln_NtkStartFanoutMap( Wln_Ntk_t * p, Vec_Int_t * vFanoutMap, Vec_Int_t * vFanoutNums, int nMulti );

// extern void           Wln_NtkStaticFanoutStart( Wln_Ntk_t * p );

// extern void           Wln_NtkStaticFanoutStop( Wln_Ntk_t * p );

// extern void           Wln_NtkStaticFanoutTest( Wln_Ntk_t * p );

/*=== wlcObj.c ========================================================*/
// extern char *         Wln_ObjName( Wln_Ntk_t * p, int iObj );

// extern char *         Wln_ObjConstString( Wln_Ntk_t * p, int iObj );

// extern void           Wln_ObjUpdateType( Wln_Ntk_t * p, int iObj, int Type );

// extern void           Wln_ObjSetConst( Wln_Ntk_t * p, int iObj, int NameId );

// extern void           Wln_ObjSetSlice( Wln_Ntk_t * p, int iObj, int SliceId );

// extern void           Wln_ObjAddFanin( Wln_Ntk_t * p, int iObj, int i );

// extern int            Wln_ObjAddFanins( Wln_Ntk_t * p, int iObj, Vec_Int_t * vFanins );

// extern int            Wln_ObjAlloc( Wln_Ntk_t * p, int Type, int Signed, int End, int Beg );

// extern int            Wln_ObjClone( Wln_Ntk_t * pNew, Wln_Ntk_t * p, int iObj );

// extern int            Wln_ObjCreateCo( Wln_Ntk_t * p, int iFanin );

// extern void           Wln_ObjPrint( Wln_Ntk_t * p, int iObj );

/*=== wlcRetime.c ========================================================*/
// extern Vec_Int_t *    Wln_NtkRetime( Wln_Ntk_t * p, int fSkipSimple, int fVerbose );

// extern void           Wln_NtkRetimeCreateDelayInfo( Wln_Ntk_t * pNtk );

/*=== wlcWriteVer.c ========================================================*/
// extern void           Wln_WriteVer( Wln_Ntk_t * p, char * pFileName );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
