package gia

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [gia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: gia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
   
type Slv_Man_t struct {
    vCis      Vec_Int_t;
    vCos      Vec_Int_t;
    vFanins      Vec_Int_t;
    vFanoutN      Vec_Int_t;
    vFanout0      Vec_Int_t;
    vValues      Vec_Str_t;
    vCopies      Vec_Int_t;
}

func Slv_ManObjNum(p *Slv_Man_t) int32 { return Vec_IntSize(&p.vFanins)/2;                       }

func Slv_ObjFaninLit(p *Slv_Man_t, iObj int32, f int32) int32 { return Vec_IntEntry(&p.vFanins, Abc_Var2Lit(iObj, f));  }
func Slv_ObjFanin(p *Slv_Man_t, iObj int32, f int32) int32 { return Abc_Lit2Var(Slv_ObjFaninLit(p, iObj, f));         }
func Slv_ObjFaninC(p *Slv_Man_t, iObj int32, f int32) int32 { return Abc_LitIsCompl(Slv_ObjFaninLit(p, iObj, f));      }

func Slv_ObjIsCi(p *Slv_Man_t, iObj int32) int32 { return !Slv_ObjFaninLit(p, iObj, 0) &&  Slv_ObjFaninLit(p, iObj, 1);  }
func Slv_ObjIsCo(p *Slv_Man_t, iObj int32) int32 { return  Slv_ObjFaninLit(p, iObj, 0) && !Slv_ObjFaninLit(p, iObj, 1);  }
func Slv_ObjIsAnd(p *Slv_Man_t, iObj int32) int32 { return  Slv_ObjFaninLit(p, iObj, 0) &&  Slv_ObjFaninLit(p, iObj, 1);  }

func Slv_ObjFanout0(p *Slv_Man_t, iObj int32) int32 { return Vec_IntEntry(&p.vFanout0, iObj);      }
func Slv_ObjSetFanout0(p *Slv_Man_t, iObj int32, iLit int32) { Vec_IntWriteEntry(&p.vFanout0, iObj, iLit);  }

func Slv_ObjNextFanout(p *Slv_Man_t, iLit int32) int32 { return Vec_IntEntry(&p.vFanoutN, iLit);      }
func Slv_ObjSetNextFanout(p *Slv_Man_t, iLit int32, iLitF int32) { Vec_IntWriteEntry(&p.vFanoutN, iLit, iLitF); }

func Slv_ObjCopyLit(p *Slv_Man_t, iObj int32) int32 { return Vec_IntEntry(&p.vCopies, iObj);       }
func Slv_ObjSetCopyLit(p *Slv_Man_t, iObj int32, iLit int32) { Vec_IntWriteEntry(&p.vCopies, iObj, iLit);   }

// Original macro:
// #define Slv_ManForEachObj( p, iObj )                                  \
//     for ( iObj = 1; iObj < Slv_ManObjNum(p); iObj++ )


// Original macro:
// #define Slv_ObjForEachFanout( p, iObj, iFanLit )                      \
//     for ( iFanLit = Slv_ObjFanout0(p, iObj); iFanLit; iFanLit = Slv_ObjNextFanout(p, iFanLit) )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Slv_ManAlloc(nObjs int32) *Slv_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Slv_Man_t  = &Slv_Man_t{};
    Vec_IntGrow( &p.vCis,     100 );
    Vec_IntGrow( &p.vCos,     100 );
    Vec_IntGrow( &p.vFanins,  2*nObjs );
    Vec_IntGrow( &p.vFanoutN, 2*nObjs );
    Vec_IntGrow( &p.vFanout0, nObjs );
    Vec_StrGrow( &p.vValues,  nObjs );
    // constant node
    Vec_IntFill( &p.vFanins,  2, 0 );
    Vec_IntFill( &p.vFanoutN, 2, 0 );
    Vec_IntFill( &p.vFanout0, 1, 0 );
    return p;
}
func Slv_ManFree(p *Slv_Man_t) {
    Vec_IntErase( &p.vCis     );
    Vec_IntErase( &p.vCos     );
    Vec_IntErase( &p.vFanins  );
    Vec_IntErase( &p.vFanoutN );
    Vec_IntErase( &p.vFanout0 );
    Vec_StrErase( &p.vValues  );
    Vec_IntErase( &p.vCopies  );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Slv_ManPrintFanouts(p *Slv_Man_t) {
    var iObj, iFanLit int32;
    // inlining of macro Slv_ManForEachObj
    for  iObj = 1; iObj < Slv_ManObjNum(p); iObj++  {
        fmt.Printf("Fanouts of node %d: ", iObj);
        // inlining of macro Slv_ObjForEachFanout
        for  iFanLit = Slv_ObjFanout0(p, iObj); (iFanLit != 0); iFanLit = Slv_ObjNextFanout(p, iFanLit)  {
            fmt.Printf("%d ", Abc_Lit2Var(iFanLit));
        }
        fmt.Printf("\n");
    }
}
func Slv_ManAppendObj(p *Slv_Man_t, iLit0 int32, iLit1 int32) int32 {
    var iObj int32  = Slv_ManObjNum(p);
    Vec_StrPush( &p.vValues,  0 );
    Vec_IntPush( &p.vFanout0, 0 );
    Vec_IntPushTwo( &p.vFanoutN, 0, 0 );
    if !(iLit0 != 0) {  // primary input
        assert(!(iLit1 != 0), "!iLit1");
iLit1 = Vec_IntSize(&p.vCis)+1;
Vec_IntPush(&p.vCis, iObj);
    } else if !(iLit1 != 0) {  // primary output
        assert((iLit0 != 0), "iLit0");
Vec_IntPush(&p.vCos, iObj);
    } else {
        Slv_ObjSetNextFanout( p, Abc_Var2Lit(iObj, 0), Slv_ObjFanout0(p, Abc_Lit2Var(iLit0)) );
        Slv_ObjSetNextFanout( p, Abc_Var2Lit(iObj, 1), Slv_ObjFanout0(p, Abc_Lit2Var(iLit1)) );
        Slv_ObjSetFanout0( p, Abc_Lit2Var(iLit0), Abc_Var2Lit(iObj, 0) );
        Slv_ObjSetFanout0( p, Abc_Lit2Var(iLit1), Abc_Var2Lit(iObj, 1) );
    }
    Vec_IntPushTwo( &p.vFanins, iLit0, iLit1 );
    return Abc_Var2Lit( iObj, 0 );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Slv_ManFromGia(p *Gia_Man_t) *Slv_Man_t {
    var pObj *Gia_Obj_t; var i int32;
    var pNew *Slv_Man_t  = Slv_ManAlloc( Gia_ManObjNum(p) );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Slv_ManAppendObj( pNew, Abc_LitNotCond(Gia_ObjFanin0(pObj).Value, Gia_ObjFaninC0(pObj)), Abc_LitNotCond(Gia_ObjFanin1(pObj).Value, Gia_ObjFaninC1(pObj)) );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Slv_ManAppendObj( pNew, Abc_LitNotCond(Gia_ObjFanin0(pObj).Value, Gia_ObjFaninC0(pObj)), 0 );
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Slv_ManAppendObj( pNew, 0, 0 );
        } else {
            assert(false, "0"); }
    }
    assert(Gia_ManObjNum(p) == Slv_ManObjNum(pNew), "Gia_ManObjNum(p) == Slv_ManObjNum(pNew)");
    return pNew;
}
func Slv_ManToGia(p *Slv_Man_t) *Gia_Man_t {
    var iObj int32;
    var pNew *Gia_Man_t  = Gia_ManStart( Slv_ManObjNum(p) ); 
    Vec_IntFill( &p.vCopies,  Slv_ManObjNum(p), 0 );
    // inlining of macro Slv_ManForEachObj
    for  iObj = 1; iObj < Slv_ManObjNum(p); iObj++  {
        if Slv_ObjIsCi(p, iObj) != 0 {
            Slv_ObjSetCopyLit( p, iObj, Gia_ManAppendCi(pNew) );
        } else if Slv_ObjIsCo(p, iObj) != 0 {
            var iLit0 int32  = Abc_LitNotCond( Slv_ObjCopyLit(p, Slv_ObjFanin(p, iObj, 0)), Slv_ObjFaninC(p, iObj, 0) );
            Slv_ObjSetCopyLit( p, iObj, Gia_ManAppendCo(pNew, iLit0) );
        
        } else if Slv_ObjIsAnd(p, iObj) != 0 {
            var iLit0 int32  = Abc_LitNotCond( Slv_ObjCopyLit(p, Slv_ObjFanin(p, iObj, 0)), Slv_ObjFaninC(p, iObj, 0) );
            var iLit1 int32  = Abc_LitNotCond( Slv_ObjCopyLit(p, Slv_ObjFanin(p, iObj, 1)), Slv_ObjFaninC(p, iObj, 1) );
            Slv_ObjSetCopyLit( p, iObj, Gia_ManAppendAnd(pNew, iLit0, iLit1) );
        
        } else { assert(false, "0"); }
    }
    assert(Gia_ManObjNum(pNew) == Slv_ManObjNum(p), "Gia_ManObjNum(pNew) == Slv_ManObjNum(p)");
    return pNew;
}

func Slv_ManToAig(pGia *Gia_Man_t) *Gia_Man_t {
    var p *Slv_Man_t  = Slv_ManFromGia( pGia );
    var pNew *Gia_Man_t  = Slv_ManToGia( p );
    pNew.pName = Abc_UtilStrsav( pGia.pName );
    pNew.pSpec = Abc_UtilStrsav( pGia.pSpec );
    Slv_ManPrintFanouts( p );
    Slv_ManFree( p );
    return pNew;
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCofPisVars(p *Gia_Man_t, nVars int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t; var i, m int32;
    pNew = Gia_ManStart( 1000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Gia_ManAppendCi( pNew );
    }
    Gia_ManHashStart( pNew );
    for  m = 0; m < (1 << nVars); m++  {
        Gia_ManFillValue( p );
        Gia_ManConst0(p).Value = 0;
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            if i < nVars {
                pObj.Value = (m >> i) & 1;
            } else {
                pObj.Value = Gia_ObjToLit(pNew, Gia_ManCi(pNew, i)); }
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        }
    }
    Gia_ManHashStop( pNew );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStructExperiment(p *Gia_Man_t) {
    extern int Cec_ManVerifyTwo( Gia_Man_t * p0, Gia_Man_t * p1, int fVerbose );
    var pTemp, pUsed *Gia_Man_t;
    var vGias *Vec_Ptr_t  = Vec_PtrAlloc( 100 );
    var pObj *Gia_Obj_t; var i, k int32;
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        var iFan0 int32  = Gia_ObjFaninId0p(p, pObj);
        pTemp = Gia_ManDupAndCones( p, &iFan0, 1, 1 );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vGias)) && (((pUsed) = (Gia_Man_t *)Vec_PtrEntry(vGias, k)), 1); k++  {
            if Gia_ManCiNum(pTemp) == Gia_ManCiNum(pUsed) && Cec_ManVerifyTwo(pTemp, pUsed, 0) == 1 {
                // inlining of macro ABC_SWAP
                { var t any  = Vec_PtrArray(vGias)[0]; Vec_PtrArray(vGias)[0] = Vec_PtrArray(vGias)[k]; Vec_PtrArray(vGias)[k] = t; };
                break;
            
            } else {
                // inlining of macro ABC_FREE
                ((pTemp.pCexComb) ? (free((char *) (pTemp.pCexComb)), (pTemp.pCexComb) = 0) : 0); }
        }

        fmt.Printf("\nOut %6d : ", i);
        if k == Vec_PtrSize(vGias) {
            fmt.Printf("Equiv to none    ");
        } else {
            fmt.Printf("Equiv to %6d  ", k); }
        Gia_ManPrintStats( pTemp, nil );

        if k == Vec_PtrSize(vGias) {
            Vec_PtrPush( vGias, pTemp );
        } else {
            Gia_ManStop( pTemp ); }
    }
    fmt.Printf("\nComputed %d classes.\n\n", Vec_PtrSize(vGias));
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vGias)) && (((pTemp) = (Gia_Man_t *)Vec_PtrEntry(vGias, i)), 1); i++  {
        Gia_ManStop( pTemp );
    }
    Vec_PtrFree( vGias );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_EnumFirstUnused(pUsed *int32, nVars int32) int32 {
    var i int32;
    for  i = 0; i < nVars; i++  {
        if pUsed[i] == 0 {
            return i; }
    }
    return -1;
}
func Gia_EnumPerms_rec(pUsed *int32, nVars int32, pPerm *int32, nPerm int32, pCount *int32, pFile *os.File, nLogVars int32) {
    var i, k, New int32;
    if nPerm == nVars {
        if pFile != nil {
            for  i = 0; i < nLogVars; i++  {
                fmt.Fprintf(pFile, "%c", '0' + ((*pCount) >> (nLogVars-1-i) & 1));
            }
            fmt.Fprintf(pFile, " ");
            for  i = 0; i < nVars; i++  {
            for  k = 0; k < nVars; k++  {
                fmt.Fprintf(pFile, "%c", '0' + (pPerm[i] == k));
            }
            }
            fmt.Fprintf(pFile, "\n");
        
        } else {
            if *pCount < 20 {
                fmt.Printf("%5d : ", (*pCount));
                for  i = 0; i < nVars; i += 2  {
                    fmt.Printf("%d %d  ", pPerm[i], pPerm[i+1]);
                }
                fmt.Printf("\n");
            }
        }
        (*pCount)++;
        return;
    }
    New = Gia_EnumFirstUnused( pUsed, nVars );
    assert(New >= 0, "New >= 0");
    pPerm[nPerm] = New;
    assert(pUsed[New] == 0, "pUsed[New] == 0");
    pUsed[New] = 1;
    // try remaining ones
    for  i = 0; i < nVars; i++  {
        if pUsed[i] == 1 {
            continue; }
        pPerm[nPerm+1] = i;
        assert(pUsed[i] == 0, "pUsed[i] == 0");
        pUsed[i] = 1;
        Gia_EnumPerms_rec( pUsed, nVars, pPerm, nPerm+2, pCount, pFile, nLogVars );
        assert(pUsed[i] == 1, "pUsed[i] == 1");
        pUsed[i] = 0;
    }
    assert(pUsed[New] == 1, "pUsed[New] == 1");
    pUsed[New] = 0;
}
func Gia_EnumPerms(nVars int32) {
    var nLogVars, Count int32;
    // inlining of macro ABC_CALLOC
    var pUsed *int32  = ((int *) calloc((size_t)(nVars), sizeof(int)));
    // inlining of macro ABC_CALLOC
    var pPerm *int32  = ((int *) calloc((size_t)(nVars), sizeof(int)));
    var pFile *os.File  = os.OpenFile("pairset.pla", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    assert(nVars % 2 == 0, "nVars % 2 == 0");

    fmt.Printf("Printing sets of pairs for %d objects:\n", nVars);
    Gia_EnumPerms_rec( pUsed, nVars, pPerm, 0, &Count, nil, -1 ); 
    if Count > 20 {
        fmt.Printf("...\n"); }
    fmt.Printf("Finished enumerating %d sets of pairs.\n", Count);

    nLogVars = Abc_Base2Log( Count );
    fmt.Printf("Need %d variables to encode %d sets.\n", nLogVars, Count);
    Count = 0;
    fmt.Fprintf(pFile, ".i %d\n", nLogVars);
    fmt.Fprintf(pFile, ".o %d\n", nVars*nVars);
    Gia_EnumPerms_rec( pUsed, nVars, pPerm, 0, &Count, pFile, nLogVars );   
    fmt.Fprintf(pFile, ".e\n");
    _ = (pFile).Close();
    fmt.Printf("Finished dumping file \"%s\".\n", "pairset.pla");

    // inlining of macro ABC_FREE
    ((pUsed) ? (free((char *) (pUsed)), (pUsed) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pPerm) ? (free((char *) (pPerm)), (pPerm) = 0) : 0);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

