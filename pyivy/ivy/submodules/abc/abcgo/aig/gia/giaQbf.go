package gia

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

/**CFile****************************************************************

  FileName    [giaQbf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [QBF solver.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 18, 2014.]

  Revision    [$Id: giaQbf.c,v 1.00 2014/10/18 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"
//#include "misc/extra/extra.h"
//#include "sat/glucose/AbcGlucose.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

    
type Qbf_Man_t struct {
    pGia      *Gia_Man_t;           // original miter
    nPars             int32;          // parameter variables
    nVars             int32;          // functional variables
    fVerbose             int32;       // verbose flag
    // internal variables
    iParVarBeg             int32;     // SAT var ID of the first par variable in the ver solver
    pSatVer     *sat_solver;        // verification instance
    pSatSyn     *sat_solver;        // synthesis instance
    pSatSynG *bmcg_sat_solver;       // synthesis instance
    vValues      *Vec_Int_t;        // variable values
    vParMap      *Vec_Int_t;        // parameter mapping
    vLits      *Vec_Int_t;          // literals for the SAT solver
    clkStart         abctime;       // global timeout
    clkSat         abctime;         // SAT solver time
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Generating QBF miter to solve the induction problem.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GenCollectFlopIndexes(pStr string, nLutNum int32, nLutSize int32, nFlops int32) *Vec_Int_t {
    var nDups int32;
    var pTemp string;
    var vFlops *Vec_Int_t;
    assert(nLutSize * nLutNum <= nFlops, "nLutSize * nLutNum <= nFlops");
    if pStr == nil {
        return Vec_IntStartNatural( nLutNum * nLutSize ); }
    vFlops = Vec_IntAlloc( nLutNum * nLutSize );
    pTemp = strtok( pStr, ", " );
    for  pTemp  {
        var iFlop int32  = strconv.Atoi(pTemp);
        if iFlop >= nFlops {
            fmt.Printf("Flop index (%d) exceeds the number of flops (%d).\n", iFlop, nFlops);
            break;
        }
        Vec_IntPush( vFlops, iFlop );
        pTemp = strtok( nil, ", " );
    }
    if Vec_IntSize(vFlops) != nLutNum * nLutSize {
        fmt.Printf("Gia_GenCollectFlopIndexes: Expecting %d flop indexes (instead of %d).\n", nLutNum * nLutSize, Vec_IntSize(vFlops));
        Vec_IntFree( vFlops );
        return nil;
    }
    nDups = Vec_IntCountDuplicates(vFlops);
    if nDups != 0 {
        fmt.Printf("Gia_GenCollectFlopIndexes: There are %d duplicated flops in the list.\n", nDups);
        Vec_IntFree( vFlops );
        return nil;
    }
    return vFlops;
}
func Gia_GenCreateMux_rec(pNew *Gia_Man_t, pCtrl *int32, nCtrl int32, vData *Vec_Int_t, Shift int32) int32 {
    var iLit0, iLit1 int32;
    if nCtrl == 0 {
        return Vec_IntEntry( vData, Shift ); }
    iLit0 = Gia_GenCreateMux_rec( pNew, pCtrl, nCtrl-1, vData, Shift );
    iLit1 = Gia_GenCreateMux_rec( pNew, pCtrl, nCtrl-1, vData, Shift + (1<<(nCtrl-1)) );
    return Gia_ManHashMux( pNew, pCtrl[nCtrl-1], iLit1, iLit0 );
}
func Gia_GenCreateMuxes(p *Gia_Man_t, pNew *Gia_Man_t, vFlops *Vec_Int_t, nLutNum int32, nLutSize int32, vParLits *Vec_Int_t, fUseRi int32) *Vec_Int_t {
    var vLits *Vec_Int_t  = Vec_IntAlloc( nLutNum );
    var i int32;
    var k int32;
    var iMux int32;
    var iFlop int32;
    var pCtrl [16]int32;
    // add MUXes for each group of flops
    assert(Vec_IntSize(vFlops) == nLutNum * nLutSize, "Vec_IntSize(vFlops) == nLutNum * nLutSize");
    for  i = 0; i < nLutNum; i++  {
        for  k = 0; k < nLutSize; k++  {
            iFlop = Vec_IntEntry(vFlops, i * nLutSize + k);
            assert(iFlop >= 0 && iFlop < Gia_ManRegNum(p), "iFlop >= 0 && iFlop < Gia_ManRegNum(p)");
            if fUseRi != 0 {
                pCtrl[k] = Gia_ManRi(p, iFlop).Value;
            } else {
                pCtrl[k] = Gia_ManRo(p, iFlop).Value; }
        }
        iMux = Gia_GenCreateMux_rec( pNew, pCtrl, nLutSize, vParLits, i * (1 << nLutSize) );
        Vec_IntPush( vLits, iMux );
    }
    return vLits;
}
func Gia_GenQbfMiter(p *Gia_Man_t, nFrames int32, nLutNum int32, nLutSize int32, pStr string, fUseOut int32, fVerbose int32) *Gia_Man_t {
    var pObj *Gia_Obj_t; 
    var pTemp, pNew *Gia_Man_t;
    var i int32;
    var iMiter int32;
    var iLut0 int32;
    var iLut1 int32;
    var nPars int32 = nLutNum * (1 << nLutSize);
    var vLits0, vLits1, vParLits *Vec_Int_t;
    var vFlops *Vec_Int_t  = Gia_GenCollectFlopIndexes( pStr, nLutNum, nLutSize, Gia_ManRegNum(p) );
    // collect parameter literals (data vars)
    vParLits = Vec_IntAlloc( nPars );
    for  i = 0; i < nPars; i++  {
        Vec_IntPush( vParLits, i ? Abc_Var2Lit(i+1, 0) : 1 );
    }
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    for  i = 0; i < nPars; i++  {
        Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ObjFanin0Copy(pObj);
    }
    vLits0 = Gia_GenCreateMuxes( p, pNew, vFlops, nLutNum, nLutSize, vParLits, 0 );
    vLits1 = Gia_GenCreateMuxes( p, pNew, vFlops, nLutNum, nLutSize, vParLits, 1 );
    // create miter output
    //iMiter = Gia_ManHashAnd( pNew, Vec_IntEntry(vLits0, 0), Abc_LitNot(Vec_IntEntry(vLits1, 0)) );
    ///////////////////////////////////////////////////////////////////////////
    iLut0 = Vec_IntEntry(vLits0, 0);
    iLut1 = Vec_IntEntry(vLits1, 0);
    if fUseOut != 0 {
        var pObjPoLast *Gia_Obj_t  = Gia_ManPo( p, Gia_ManPoNum(p)-1 );
        var iOut int32  = Abc_LitNotCond( Gia_ObjFanin0Copy(pObjPoLast), 0 );
        iLut1 = Gia_ManHashAnd( pNew, iLut1, Abc_LitNot(iOut) );
    }
    iMiter = Gia_ManHashAnd( pNew, iLut0, Abc_LitNot(iLut1) );
    ///////////////////////////////////////////////////////////////////////////
    iMiter = Gia_ManHashAnd( pNew, Abc_LitNot(iMiter), Abc_Var2Lit(1, 0) );
    Gia_ManAppendCo( pNew, iMiter );
    // cleanup
    Vec_IntFree( vLits0 );
    Vec_IntFree( vLits1 );
    Vec_IntFree( vFlops );
    Vec_IntFree( vParLits );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Generate miter for the encoding problem.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Gen2CreateMux_rec(pNew *Gia_Man_t, pCtrl *int32, nCtrl int32, vData *Vec_Int_t, Shift int32) int32 {
    var iLit0, iLit1 int32;
    if nCtrl == 0 {
        return Vec_IntEntry( vData, Shift ); }
    iLit0 = Gia_Gen2CreateMux_rec( pNew, pCtrl, nCtrl-1, vData, Shift );
    iLit1 = Gia_Gen2CreateMux_rec( pNew, pCtrl, nCtrl-1, vData, Shift + (1<<(nCtrl-1)) );
    return Gia_ManHashMux( pNew, pCtrl[nCtrl-1], iLit1, iLit0 );
}
func Gia_Gen2CreateMuxes(pNew *Gia_Man_t, nLutSize int32, nLutNum int32, vPLits *Vec_Int_t, vXLits *Vec_Int_t) *Vec_Int_t {
    var vLits *Vec_Int_t  = Vec_IntAlloc( nLutNum );
    var i, iMux int32;
    // add MUXes for each group of flops
    assert(Vec_IntSize(vPLits) == nLutNum * (1 << nLutSize), "Vec_IntSize(vPLits) == nLutNum * (1 << nLutSize)");
    assert(Vec_IntSize(vXLits) == nLutSize, "Vec_IntSize(vXLits) == nLutSize");
    for  i = 0; i < nLutNum; i++  {
        iMux = Gia_Gen2CreateMux_rec( pNew, Vec_IntArray(vXLits), nLutSize, vPLits, i * (1 << nLutSize) );
        Vec_IntPush( vLits, iMux );
    }
    return vLits;
}
func Gia_Gen2CreateMiter(nLutSize int32, nLutNum int32) *Gia_Man_t {
    // |<-- PVars(0)-->|...|<-- PVars(nLutNum-1)-->|<-- XVars-->|<-- YVars-->|
    var vPLits *Vec_Int_t   = Vec_IntAlloc( nLutNum * (1 << nLutSize) );
    var vXLits *Vec_Int_t   = Vec_IntAlloc( nLutSize );
    var vYLits *Vec_Int_t   = Vec_IntAlloc( nLutSize );
    var vXYLits *Vec_Int_t  = Vec_IntAlloc( nLutSize );
    var vXRes, vYRes, vXYRes *Vec_Int_t;
    var vXYRes2 *Vec_Int_t  = Vec_IntAlloc( 2 * nLutNum );
    var pTemp *Gia_Man_t;
    var pNew *Gia_Man_t = Gia_ManStart( 1000 ); var i, k, v, Cond, Res int32;
    pNew.pName = Abc_UtilStrsav( "homoqbf" );
    Gia_ManHashAlloc( pNew );
    for  i = 0; i < nLutNum * (1 << nLutSize); i++  {
        Vec_IntPush( vPLits, Gia_ManAppendCi(pNew) );
    }
    for  i = 0; i < nLutSize; i++  {
        Vec_IntPush( vXLits, Gia_ManAppendCi(pNew) );
    }
    for  i = 0; i < nLutSize; i++  {
        Vec_IntPush( vYLits, Gia_ManAppendCi(pNew) );
    }
    for  i = 0; i < nLutSize; i++  {
        Vec_IntPush( vXYLits, Abc_LitNot(Gia_ManHashAnd(pNew, Vec_IntEntry(vXLits, i), Vec_IntEntry(vYLits, i))) );
    }
    vXRes  = Gia_Gen2CreateMuxes( pNew, nLutSize, nLutNum, vPLits, vXLits );
    vYRes  = Gia_Gen2CreateMuxes( pNew, nLutSize, nLutNum, vPLits, vYLits );
    vXYRes = Gia_Gen2CreateMuxes( pNew, nLutSize, nLutNum, vPLits, vXYLits );
    for  i = 0; i < nLutNum; i++  {
        Vec_IntPush( vXYRes2, Vec_IntEntry(vXYRes, i) );
        Vec_IntPush( vXYRes2, Abc_LitNot(Gia_ManHashAnd(pNew, Vec_IntEntry(vXRes, i), Vec_IntEntry(vYRes, i))) );
    }
    Res = Gia_ManHashDualMiter( pNew, vXYRes2 );
    // uniqueness of codes
    for  i = 0; i < (1 << nLutSize); i++  {
        var vCondA *Vec_Int_t  = Vec_IntAlloc( nLutNum );
        var vCondB *Vec_Int_t  = Vec_IntAlloc( nLutNum );
        for  v = 0; v < nLutNum; v++  {
            Vec_IntPush( vCondA, Vec_IntEntry(vPLits, v*(1 << nLutSize)+i) );
        }
        for  k = i+1; k < (1 << nLutSize); k++  {
            Vec_IntClear( vCondB );
            for  v = 0; v < nLutNum; v++  {
                Vec_IntPush( vCondB, Vec_IntEntry(vCondA, v) );
                Vec_IntPush( vCondB, Vec_IntEntry(vPLits, v*(1 << nLutSize)+k) );
            }
            Cond = Gia_ManHashDualMiter( pNew, vCondB );
            Res = Gia_ManHashOr( pNew, Res, Abc_LitNot(Cond) );
        }
        Vec_IntFree( vCondA );
        Vec_IntFree( vCondB );
    }
    Gia_ManAppendCo( pNew, Abc_LitNot(Res) );
    Gia_ManHashStop( pNew );
    Vec_IntFree( vPLits );
    Vec_IntFree( vXLits );
    Vec_IntFree( vYLits );
    Vec_IntFree( vXYLits );
    Vec_IntFree( vXRes );
    Vec_IntFree( vYRes );
    Vec_IntFree( vXYRes );
    Vec_IntFree( vXYRes2 );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    fmt.Printf("Generated QBF miter with %d parameters, %d functional variables, and %d AIG nodes.\n", nLutNum * (1 << nLutSize), 2*nLutSize, Gia_ManAndNum(pNew));
    return pNew;
}
func Gia_Gen2CodeOne(nLutSize int32, nLutNum int32, vCode *Vec_Int_t, x int32) int32 {
    var k, Code int32;
    for  k = 0; k < nLutNum; k++  {
        if Vec_IntEntry(vCode, k*(1 << nLutSize)+x) {
            Code |= (1 << k); }
    }
    return Code;
}
func Gia_Gen2CodeOneP(nLutSize int32, nLutNum int32, vCode *Vec_Int_t, x int32) *word {
    // inlining of macro ABC_CALLOC
    var pRes *word  = ((word *) calloc((size_t)(Abc_Bit6WordNum(nLutNum)), sizeof(word)));
    var k int32;
    for  k = 0; k < nLutNum; k++  {
        if Vec_IntEntry(vCode, k*(1 << nLutSize)+x) {
            Abc_InfoSetBit( (unsigned *)pRes, k ); }
    }
    return pRes;
}
func Gia_Gen2CodePrint(nLutSize int32, nLutNum int32, vCode *Vec_Int_t) {
    // |<-- PVars(0)-->|...|<-- PVars(nLutNum-1)-->|
    var i int32;
    var n int32;
    var nPairs int32 = 16;
    fmt.Printf("%d-input %d-output code table:\n", nLutSize, nLutNum);
    for  i = 0; i < (1 << nLutSize); i++  {
        var CodeX *word   = Gia_Gen2CodeOneP( nLutSize, nLutNum, vCode, i );
        fmt.Printf("%3d  ", i);
        Extra_PrintBinary( os.Stdout, (unsigned *)&i, nLutSize );
        fmt.Printf("  -->  ");
        if nLutNum <= 16 {
            fmt.Printf("%5d  ", int32(CodeX[0])); }
        Extra_PrintBinary( os.Stdout, (unsigned *)CodeX, nLutNum );
        fmt.Printf("\n");
        // inlining of macro ABC_FREE
        CodeX = nil;
    }
    // create several different pairs
    srand( time.Now().Unix() );
    fmt.Printf("Simulation of the encoding with %d random pairs:\n", nPairs);
    for  n = 0; n < nPairs; n++  {
        var MaskIn uint32  = Abc_InfoMask( nLutSize );
        var NumX int32;
        var NumY int32;
        var NumXY int32;
        var nWords int32 = Abc_Bit6WordNum(nLutNum);
        var CodeX, CodeY, CodeXY *word;
        // inlining of macro ABC_CALLOC
        var CodeXCodeY *word  = ((word *) calloc((size_t)(nWords), sizeof(word)));
        for  NumX == NumY  {
            NumX = rand() % (1 << nLutSize);
            NumY = rand() % (1 << nLutSize);
            NumXY = MaskIn & ~(NumX & NumY);
        }
        CodeX  = Gia_Gen2CodeOneP( nLutSize, nLutNum, vCode, NumX );
        CodeY  = Gia_Gen2CodeOneP( nLutSize, nLutNum, vCode, NumY );
        CodeXY = Gia_Gen2CodeOneP( nLutSize, nLutNum, vCode, NumXY );
        Abc_TtAnd( CodeXCodeY, CodeX, CodeY, nWords, 1 );
        if nLutNum < 64*nWords {
            CodeXCodeY[nWords-1] &= Abc_Tt6Mask(nLutNum % 64); }            

        fmt.Printf("%2d :", n);
        fmt.Printf(" x =%3d ", NumX);
        Extra_PrintBinary( os.Stdout,(unsigned *) &NumX, nLutSize );
        fmt.Printf(" y =%3d ", NumY);
        Extra_PrintBinary( os.Stdout, (unsigned *)&NumY, nLutSize );
        fmt.Printf(" nand =%3d ", NumXY);
        Extra_PrintBinary( os.Stdout, (unsigned *)&NumXY, nLutSize );
        fmt.Printf("  ");

        fmt.Printf(" c(x) = ");
        Extra_PrintBinary( os.Stdout, (unsigned *)CodeX, nLutNum );
        fmt.Printf(" c(y) = ");
        Extra_PrintBinary( os.Stdout, (unsigned *)CodeY, nLutNum );
        fmt.Printf(" c(nand) = ");
        Extra_PrintBinary( os.Stdout, (unsigned *)CodeXY, nLutNum );
        fmt.Printf("  nand(c(x),c(y)) = ");
        Extra_PrintBinary( os.Stdout, (unsigned *)CodeXCodeY, nLutNum );
        fmt.Printf("  ");

        fmt.Printf("%s", Abc_TtEqual(CodeXCodeY, CodeXY, nWords) ? "yes" : "no");
        fmt.Printf("\n");

        // inlining of macro ABC_FREE
        CodeX = nil;
        // inlining of macro ABC_FREE
        CodeY = nil;
        // inlining of macro ABC_FREE
        CodeXY = nil;
        // inlining of macro ABC_FREE
        CodeXCodeY = nil;
    }
}
func Gia_Gen2CodeTest() {
    var i int32;
    var nLutSize int32 = 1;
    var nLutNum int32 = 2;
    var vCode *Vec_Int_t  = Vec_IntAlloc( (1 << nLutSize) * nLutNum );
    srand( time.Now().Unix() );
    for  i = 0; i < (1 << nLutSize) * nLutNum; i++  {
        Vec_IntPush( vCode, rand() & 1 );
    }
    Gia_Gen2CodePrint( nLutSize, nLutNum, vCode );
    Vec_IntFree( vCode );
}

/**Function*************************************************************

  Synopsis    [Naive way to enumerate SAT assignments.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatEnum(pGia *Gia_Man_t, nConfLimit int32, nTimeOut int32, fVerbose int32) int32 {
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    var vLits *Vec_Int_t;
    var i, iLit, iParVarBeg, Iter int32;
    var nSolutions, RetValue int32;
    var clkStart abctime  = Abc_Clock();
    pCnf = (Cnf_Dat_t *)Mf_ManGenerateCnf( pGia, 8, 0, 1, 0, 0 );
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    iParVarBeg = pCnf.nVars - Gia_ManPiNum(pGia);// - 1;
    Cnf_DataFree( pCnf );
    // iterate through the SAT assignment
    vLits = Vec_IntAlloc( Gia_ManPiNum(pGia) );
    for  Iter = 1 ; ; Iter++  {
        var status int32  = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)nConfLimit, 0, 0, 0 );
        if status == l_False { RetValue =  1; break; }
        if status == l_Undef { RetValue =  0; break; }
        nSolutions++;
        // extract SAT assignment
        Vec_IntClear( vLits );
        for  i = 0; i < Gia_ManPiNum(pGia); i++  {
            Vec_IntPush( vLits, Abc_Var2Lit(iParVarBeg+i, sat_solver_var_value(pSat, iParVarBeg+i)) );
        }
        if fVerbose != 0 {
            fmt.Printf("%5d : ", Iter);
            // inlining of macro Vec_IntForEachEntry
            for  i = 0; (i < Vec_IntSize(vLits)) && (((iLit) = Vec_IntEntry(vLits, i)), 1); i++  {
                fmt.Printf("%d", !Abc_LitIsCompl(iLit));
            }
            fmt.Printf("\n");
        }
        // add clause
        if !sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) ) { RetValue =  1; break; }
        if (nTimeOut != 0) && (Abc_Clock() - clkStart)/int64(time.Second) >= nTimeOut { RetValue = 0; break; }
    }
    sat_solver_delete( pSat );
    Vec_IntFree( vLits );
    if (nTimeOut != 0) && (Abc_Clock() - clkStart)/int64(time.Second) >= nTimeOut {
        fmt.Printf("Enumerated %d assignments when timeout (%d sec) was reached.  ", nSolutions, nTimeOut);
    } else if (nConfLimit != 0) && !(RetValue != 0) {
        fmt.Printf("Enumerated %d assignments when conflict limit (%d) was reached.  ", nSolutions, nConfLimit);
    } else { 
        fmt.Printf("Enumerated the complete set of %d assignments.  ", nSolutions); }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clkStart );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Dumps the original problem in QDIMACS format.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_QbfDumpFile(pGia *Gia_Man_t, nPars int32) {
    // original problem: \exists p \forall x \exists y.  M(p,x,y)
    // negated problem:  \forall p \exists x \exists y. !M(p,x,y)
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( pGia, 8, 0, 1, 0, 0 );
    var vVarMap, vForAlls, vExists *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var pFileName string;
    var i, Entry int32;
    // create var map
    vVarMap = Vec_IntStart( pCnf.nVars );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((pObj) = Gia_ManCi(pGia, i)); i++  {
        if i < nPars {
            Vec_IntWriteEntry( vVarMap, pCnf.pVarNums[Gia_ManCiIdToId(pGia, i)], 1 ); }
    }
    // create various maps
    vForAlls = Vec_IntAlloc( nPars );
    vExists = Vec_IntAlloc( Gia_ManCiNum(pGia) - nPars );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vVarMap)) && (((Entry) = Vec_IntEntry(vVarMap, i)), 1); i++  {
        if Entry != 0 {
            Vec_IntPush( vForAlls, i );
        } else {
            Vec_IntPush( vExists, i ); }
    }
    // generate CNF
    pFileName = Extra_FileNameGenericAppend( pGia.pSpec, ".qdimacs" );
    Cnf_DataWriteIntoFile( pCnf, pFileName, 0, vForAlls, vExists );
    Cnf_DataFree( pCnf );
    Vec_IntFree( vForAlls );
    Vec_IntFree( vExists );
    Vec_IntFree( vVarMap );
    fmt.Printf("The 2QBF formula was written into file \"%s\".\n", pFileName);
}

/**Function*************************************************************

  Synopsis    [Generate one SAT assignment of the problem.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_QbfAlloc(pGia *Gia_Man_t, nPars int32, fGlucose int32, fVerbose int32) *Qbf_Man_t {
    
    var pCnf *Cnf_Dat_t;
    Gia_ObjFlipFaninC0( Gia_ManPo(pGia, 0) );
    pCnf = (Cnf_Dat_t *)Mf_ManGenerateCnf( pGia, 8, 0, 1, 0, 0 );
    Gia_ObjFlipFaninC0( Gia_ManPo(pGia, 0) );
    // inlining of macro ABC_CALLOC
    p := &Qbf_Man_t{};
    p.clkStart   = Abc_Clock();
    p.pGia       = pGia;
    p.nPars      = nPars;
    p.nVars      = Gia_ManPiNum(pGia) - nPars;
    p.fVerbose   = fVerbose;
    p.iParVarBeg = pCnf.nVars - Gia_ManPiNum(pGia);// - 1;
    p.pSatVer    = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    p.pSatSyn    = sat_solver_new();
    if fGlucose != 0 {
        p.pSatSynG = bmcg_sat_solver_start();
    } else {
        p.pSatSynG = nil;
    } 
    p.vValues    = Vec_IntAlloc( Gia_ManPiNum(pGia) );
    p.vParMap    = Vec_IntStartFull( nPars );
    p.vLits      = Vec_IntAlloc( nPars );
    sat_solver_setnvars( p.pSatSyn, nPars );
    if p.pSatSynG { bmcg_sat_solver_set_nvars( p.pSatSynG, nPars ); }
    Cnf_DataFree( pCnf );
    return p;
}
func Gia_QbfFree(p *Qbf_Man_t) {
    sat_solver_delete( p.pSatVer );
    sat_solver_delete( p.pSatSyn );
    if p.pSatSynG { bmcg_sat_solver_stop( p.pSatSynG ); }
    Vec_IntFree( p.vLits );
    Vec_IntFree( p.vValues );
    Vec_IntFree( p.vParMap );
    // inlining of macro ABC_FREE
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Create and add one cofactor.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Create and add one cofactor.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/*
int Gia_QbfAddCofactor( Qbf_Man_t * p, Gia_Man_t * pCof )
{
    Cnf_Dat_t * pCnf = (Cnf_Dat_t *)Mf_ManGenerateCnf( pCof, 8, 0, 1, 0, 0 );
    int i, iFirstVar = sat_solver_nvars(p->pSatSyn) + pCnf->nVars - Gia_ManPiNum(pCof);// - 1;
    pCnf->pMan = NULL;
    Cnf_DataLift( pCnf, sat_solver_nvars(p->pSatSyn) );
    for ( i = 0; i < pCnf->nClauses; i++ )
        if ( !sat_solver_addclause( p->pSatSyn, pCnf->pClauses[i], pCnf->pClauses[i+1] ) )
        {
            Cnf_DataFree( pCnf );
            return 0;
        }
    Cnf_DataFree( pCnf );
    // add connection clauses
    for ( i = 0; i < Gia_ManPiNum(p->pGia); i++ )
        if ( !sat_solver_add_buffer( p->pSatSyn, i, iFirstVar+i, 0 ) )
            return 0;
    return 1;
}
*/
func Cnf_SpecialDataLift(p *Cnf_Dat_t, nVarsPlus int32, firstPiVar int32, lastPiVar int32) {
    var v, var int32;
    for  v = 0; v < p.nLiterals; v++  {
        var = p.pClauses[0][v] / 2;
        if var < firstPiVar || var >= lastPiVar {
            p.pClauses[0][v] += 2*nVarsPlus;
        } else { 
            p.pClauses[0][v] -= 2*firstPiVar; }
    }
}

func Gia_QbfAddCofactor(p *Qbf_Man_t, pCof *Gia_Man_t) int32 {
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( pCof, 8, 0, 1, 0, 0 );
    var i, useold int32;
    var iFirstVar int32;
    if useold != 0 {
        iFirstVar = sat_solver_nvars(p.pSatSyn) + pCnf.nVars - Gia_ManPiNum(pCof);
    } else {
        iFirstVar = pCnf.nVars - Gia_ManPiNum(pCof);
    } //-1   
    pCnf.pMan = nil;
    
    if useold != 0 {    
        Cnf_DataLift( pCnf, sat_solver_nvars(p.pSatSyn) );
    } else {
        Cnf_SpecialDataLift( pCnf, sat_solver_nvars(p.pSatSyn), iFirstVar, iFirstVar + Gia_ManPiNum(p.pGia) ); }

    for  i = 0; i < pCnf.nClauses; i++  {
        if !sat_solver_addclause( p.pSatSyn, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
            Cnf_DataFree( pCnf );
            return 0;
        }
    }
    Cnf_DataFree( pCnf );
    // add connection clauses
    if useold != 0 {
           for  i = 0; i < Gia_ManPiNum(p.pGia); i++  {
            if !sat_solver_add_buffer( p.pSatSyn, i, iFirstVar+i, 0 ) {
                return 0; }
           } }
    return 1;
}
func Gia_QbfAddCofactorG(p *Qbf_Man_t, pCof *Gia_Man_t) int32 {
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( pCof, 8, 0, 1, 0, 0 );
    var i int32;
    var iFirstVar int32 = pCnf.nVars - Gia_ManPiNum(pCof); //-1   
    pCnf.pMan = nil;
    Cnf_SpecialDataLift( pCnf, bmcg_sat_solver_varnum(p.pSatSynG), iFirstVar, iFirstVar + Gia_ManPiNum(p.pGia) );
    for  i = 0; i < pCnf.nClauses; i++  {
        if !bmcg_sat_solver_addclause( p.pSatSynG, pCnf.pClauses[i], pCnf.pClauses[i+1]-pCnf.pClauses[i] ) {
            Cnf_DataFree( pCnf );
            return 0;
        }
    }
    Cnf_DataFree( pCnf );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Extract SAT assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_QbfOnePattern(p *Qbf_Man_t, vValues *Vec_Int_t) {
    var i int32;
    Vec_IntClear( vValues );
    for  i = 0; i < p.nPars; i++  {
        Vec_IntPush( vValues, p.pSatSynG ? bmcg_sat_solver_read_cex_varvalue(p.pSatSynG, i) : sat_solver_var_value(p.pSatSyn, i) );
    }
}
func Gia_QbfPrint(p *Qbf_Man_t, vValues *Vec_Int_t, Iter int32) {
    fmt.Printf("%5d : ", Iter);
    assert(Vec_IntSize(vValues) == p.nVars, "Vec_IntSize(vValues) == p.nVars");
    Vec_IntPrintBinary( vValues ); fmt.Printf("  ");
    fmt.Printf("Var =%7d  ", p.pSatSynG ? bmcg_sat_solver_varnum(p.pSatSynG)      : sat_solver_nvars(p.pSatSyn));
    fmt.Printf("Cla =%7d  ", p.pSatSynG ? bmcg_sat_solver_clausenum(p.pSatSynG)   : sat_solver_nclauses(p.pSatSyn));
    fmt.Printf("Conf =%9d  ", p.pSatSynG ? bmcg_sat_solver_conflictnum(p.pSatSynG) : sat_solver_nconflicts(p.pSatSyn));
    Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
}

/**Function*************************************************************

  Synopsis    [Generate one SAT assignment in terms of functional vars.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_QbfVerify(p *Qbf_Man_t, vValues *Vec_Int_t) int32 {
    var i, Entry, RetValue int32;
    assert(Vec_IntSize(vValues) == p.nPars, "Vec_IntSize(vValues) == p.nPars");
    Vec_IntClear( p.vLits );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vValues)) && (((Entry) = Vec_IntEntry(vValues, i)), 1); i++  {
        Vec_IntPush( p.vLits, Abc_Var2Lit(p.iParVarBeg+i, !Entry) );
    }
    RetValue = sat_solver_solve( p.pSatVer, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits), 0, 0, 0, 0 );
    if RetValue == l_True {
        Vec_IntClear( vValues );
        for  i = 0; i < p.nVars; i++  {
            Vec_IntPush( vValues, sat_solver_var_value(p.pSatVer, p.iParVarBeg+p.nPars+i) );
        }
    }
    if RetValue == l_True {
        return 1;
    } else {
        return 0;
    }
}

/**Function*************************************************************

  Synopsis    [Constraint learning.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_QbfAddSpecialConstr(p *Qbf_Man_t) {
    var i int32;
    var status int32;
    var Lits [2]int32;
    for  i = 0; i < 4*11; i++  {
        if i % 4 == 0 {
            assert(Vec_IntEntry(p.vParMap, i) == -1, "Vec_IntEntry(p.vParMap, i) == -1");
            Vec_IntWriteEntry( p.vParMap, i, (i % 4) == 3 );
            Lits[0] = Abc_Var2Lit( i, (i % 4) != 3 );
            status = sat_solver_addclause( p.pSatSyn, Lits, Lits+1 );
            assert((status != 0), "status");
        }
    }
}
func Gia_QbfLearnConstraint(p *Qbf_Man_t, vValues *Vec_Int_t) {
    var i int32;
    var status int32;
    var Entry int32;
    var Lits [2]int32;
    assert(Vec_IntSize(vValues) == p.nPars, "Vec_IntSize(vValues) == p.nPars");
    fmt.Printf("  Pattern   ");
    Vec_IntPrintBinary( vValues );
    fmt.Printf("\n");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vValues)) && (((Entry) = Vec_IntEntry(vValues, i)), 1); i++  {
        Lits[0] = Abc_Var2Lit( i, Entry );
        status = sat_solver_solve( p.pSatSyn, Lits, Lits+1, 0, 0, 0, 0 );
        fmt.Printf("  Var =%4d ", i);
        if status != l_True {
            fmt.Printf("UNSAT\n");
            Lits[0] = Abc_LitNot(Lits[0]);
            status = sat_solver_addclause( p.pSatSyn, Lits, Lits+1 );
            assert((status != 0), "status");
            continue;
        }
        Gia_QbfOnePattern( p, p.vLits );
        Vec_IntPrintBinary( p.vLits );
        fmt.Printf("\n");
    }
    assert(Vec_IntSize(vValues) == p.nPars, "Vec_IntSize(vValues) == p.nPars");
}

/**Function*************************************************************

  Synopsis    [Performs QBF solving using an improved algorithm.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_QbfSolve(pGia *Gia_Man_t, nPars int32, nIterLimit int32, nConfLimit int32, nTimeOut int32, nEncVars int32, fGlucose int32, fVerbose int32) int32 {
    var p *Qbf_Man_t  = Gia_QbfAlloc( pGia, nPars, fGlucose, fVerbose );
    var pCof *Gia_Man_t;
    var i, status, RetValue int32;
    var clk abctime;
//    Gia_QbfAddSpecialConstr( p );
    if fVerbose != 0 {
        fmt.Printf("Solving QBF for \"%s\" with %d parameters, %d variables and %d AIG nodes.\n", Gia_ManName(pGia), p.nPars, p.nVars, Gia_ManAndNum(pGia)); }
    assert(Gia_ManRegNum(pGia) == 0, "Gia_ManRegNum(pGia) == 0");
    Vec_IntFill( p.vValues, nPars, 0 );
    for  i = 0; (Gia_QbfVerify(p, p.vValues) != 0); i++  {
        // generate next constraint
        assert(Vec_IntSize(p.vValues) == p.nVars, "Vec_IntSize(p.vValues) == p.nVars");
        pCof = Gia_QbfCofactor( pGia, nPars, p.vValues, p.vParMap );
        if p.pSatSynG {
            status = Gia_QbfAddCofactorG( p, pCof );
        } else {
            status = Gia_QbfAddCofactor( p, pCof );
        }
        Gia_ManStop( pCof );
        if status == 0       { RetValue =  1; break; }
        // synthesize next assignment
        clk = Abc_Clock();
        if p.pSatSynG {
            status = bmcg_sat_solver_solve( p.pSatSynG, nil, 0 );
        } else {
            status = sat_solver_solve( p.pSatSyn, nil, nil, (ABC_INT64_T)nConfLimit, 0, 0, 0 ); }
        p.clkSat += Abc_Clock() - clk;
        if fVerbose != 0 {
            Gia_QbfPrint( p, p.vValues, i ); }
        if status == l_False { RetValue =  1; break; }
        if status == l_Undef { RetValue = -1; break; }
        // extract SAT assignment
        Gia_QbfOnePattern( p, p.vValues );
        assert(Vec_IntSize(p.vValues) == p.nPars, "Vec_IntSize(p.vValues) == p.nPars");
        // examine variables
//        Gia_QbfLearnConstraint( p, p->vValues );
//        Vec_IntPrintBinary( p->vValues ); printf( "\n" );
        if (nIterLimit != 0) && i+1 == nIterLimit { RetValue = -1; break; }
        if (nTimeOut != 0) && (Abc_Clock() - p.clkStart)/int64(time.Second) >= nTimeOut { RetValue = -1; break; }
    }
    if RetValue == 0 {
        var nZeros int32  = Vec_IntCountZero( p.vValues );
        fmt.Printf("Parameters: ");
        assert(Vec_IntSize(p.vValues) == nPars, "Vec_IntSize(p.vValues) == nPars");
        Vec_IntPrintBinary( p.vValues );
        fmt.Printf("  Statistics: 0=%d 1=%d\n", nZeros, Vec_IntSize(p.vValues) - nZeros);
        if nEncVars != 0 {
            var nBits int32  = Vec_IntSize(p.vValues)/(1 << nEncVars);
            assert(Vec_IntSize(p.vValues) == (1 << nEncVars) * nBits, "Vec_IntSize(p.vValues) == (1 << nEncVars) * nBits");
            Gia_Gen2CodePrint( nEncVars, nBits, p.vValues );
        }
    }
    if RetValue == -1 && (nTimeOut != 0) && (Abc_Clock() - p.clkStart)/int64(time.Second) >= nTimeOut {
        fmt.Printf("The problem timed out after %d sec.  ", nTimeOut);
    } else if RetValue == -1 && (nConfLimit != 0) {
        fmt.Printf("The problem aborted after %d conflicts.  ", nConfLimit);
    } else if RetValue == -1 && (nIterLimit != 0) {
        fmt.Printf("The problem aborted after %d iterations.  ", nIterLimit);
    } else if RetValue == 1 {
        fmt.Printf("The problem is UNSAT after %d iterations.  ", i);
    } else { 
        fmt.Printf("The problem is SAT after %d iterations.  ", i); }
    if fVerbose != 0 {
        fmt.Printf("\n");
        Abc_PrintTime( 1, "SAT  ", p.clkSat );
        Abc_PrintTime( 1, "Other", Abc_Clock() - p.clkStart - p.clkSat );
        Abc_PrintTime( 1, "TOTAL", Abc_Clock() - p.clkStart );
    
    } else {
        Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart ); }
    Gia_QbfFree( p );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

