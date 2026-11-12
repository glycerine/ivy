package gia

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [giaSatLE.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Mapping with edges.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSatLE.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/extra/extra.h"
//#include "misc/tim/tim.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Cut printing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// returns 1 if Cut can represent LUT (Cut is equal or is contained in LUT)
func Sle_ManDeriveInit(p *Sle_Man_t) {
    var vEdges *Vec_Int_t;
    var i, iObj, iFanin, iEdge int32;
    if !Gia_ManHasMapping(p.pGia) {
        return; }
    // derive initial state
    Vec_IntClear( p.vPolars );
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        var nFanins int32;
        var pFanins *int32;
        var pCut *int32;
        var pList *int32;
        var iFound int32 = -1;
        if !Gia_ObjIsLut(p.pGia, iObj) {
            continue; }
        Vec_IntPush( p.vPolars, iObj ); // node var
        nFanins = Gia_ObjLutSize( p.pGia, iObj );
        pFanins = Gia_ObjLutFanins( p.pGia, iObj );
        // find cut
        pList = Sle_ManList( p, iObj );
        // inlining of macro Sle_ForEachCut
        for  i, pCut = 0, pList + 1; i <  pList[0]; i++, pCut += Sle_CutSize(pCut) + 1  { // cuts with unit-cut

            if Sle_ManCheckContained( Sle_CutLeaves(pCut), Sle_CutSize(pCut), pFanins, nFanins ) != 0 {
                iFound = i;
                break;
            }
        }
        if iFound == -1 {
            fmt.Printf("Cannot find the following cut at node %d: {", iObj);
            for  i = 0; i < nFanins; i++  {
                fmt.Printf(" %d", pFanins[i]);
            }
            fmt.Printf(" }\n");
            Sle_ManPrintCuts( p.pGia, p.vCuts, iObj );
            fflush( os.Stdout );
        }
        assert(iFound >= 0, "iFound >= 0");
        Vec_IntPush( p.vPolars, Vec_IntEntry(p.vCutFirst, iObj) + iFound ); // cut var
        // check if the cut contains only primary inputs - if so, its delay is equal to 1
        if Sle_ManCutHasPisOnly(pCut, p.vMask) != 0 {
            Vec_IntPush( p.vPolars, Vec_IntEntry(p.vDelayFirst, iObj) ); } // delay var
    }
    }
    Vec_IntSort( p.vPolars, 0 );
    // find zero-delay edges
    if !p.pGia.vEdge1 {
        return; }
    vEdges = Gia_ManEdgeToArray( p.pGia );
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vEdges)) && (((iFanin) = Vec_IntEntry(vEdges, i)), 1) && (((iObj) = Vec_IntEntry(vEdges, i+1)), 1); i += 2  {
        assert(iFanin < iObj, "iFanin < iObj");
        assert(Gia_ObjIsLut(p.pGia, iFanin), "Gia_ObjIsLut(p.pGia, iFanin)");
        assert(Gia_ObjIsLut(p.pGia, iObj), "Gia_ObjIsLut(p.pGia, iObj)");
        assert(Gia_ObjIsAnd(Gia_ManObj(p.pGia, iFanin)), "Gia_ObjIsAnd(Gia_ManObj(p.pGia, iFanin))");
        assert(Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)), "Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj))");
        // find edge
        iEdge = Vec_IntFind( Vec_WecEntry(p.vCutFanins, iObj), iFanin );
        if iEdge < 0 {
            continue; }
        assert(iEdge >= 0, "iEdge >= 0");
        Vec_IntPush( p.vPolars, Vec_IntEntry(p.vEdgeFirst, iObj) + iEdge ); // edge
    }
    Vec_IntFree( vEdges );
}

/**Function*************************************************************

  Synopsis    [Derive CNF.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManDeriveCnf(p *Sle_Man_t, nBTLimit int32, fDynamic int32) {
    var nTimeOut int32;
    var i, iObj, value int32;
    var vArray *Vec_Int_t;

    // start the SAT solver
    p.pSat = sat_solver_new();
    sat_solver_setnvars( p.pSat, p.nVarsTotal );
    sat_solver_set_resource_limits( p.pSat, nBTLimit, 0, 0, 0 );
    sat_solver_set_runtime_limit( p.pSat, nTimeOut ? nTimeOut * int64(time.Second) + Abc_Clock(): 0 );
    sat_solver_set_random( p.pSat, 1 );
    sat_solver_set_polarity( p.pSat, Vec_IntArray(p.vPolars), Vec_IntSize(p.vPolars) );
    //sat_solver_set_var_activity( p->pSat, NULL, p->nVarsTotal );

    // set drivers to be mapped
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && (((iObj) = Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i))), 1); i++  {
        if Vec_BitEntry(p.vMask, iObj) { // internal node
        
            Vec_IntFill( p.vLits, 1, Abc_Var2Lit(iObj, 0) ); // pos lit
            value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
            assert((value != 0), "value");
        }
    }

    // add cover clauses and edge-to-cut clauses
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        var e, iEdge, nEdges, Entry int32;
        var iCutVar0 int32   = Vec_IntEntry( p.vCutFirst, iObj );
        var iEdgeVar0 int32  = Vec_IntEntry( p.vEdgeFirst, iObj );
        var pCut *int32;
        var pList *int32 = Sle_ManList( p, iObj );
        var vCutFans *Vec_Int_t  = Vec_WecEntry( p.vCutFanins, iObj );
        assert(iCutVar0 > 0 && iEdgeVar0 > 0, "iCutVar0 > 0 && iEdgeVar0 > 0");
        // node requires one of the cuts
        Vec_IntFill( p.vLits, 1, Abc_Var2Lit(iObj, 1) ); // neg lit
        for  i = 0; i < Sle_ListCutNum(pList); i++  {
            Vec_IntPush( p.vLits, Abc_Var2Lit(iCutVar0 + i, 0) );
        }
        value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
        assert((value != 0), "value");
        // cuts are mutually exclusive
        for  i = 0; i < Sle_ListCutNum(pList); i++  {
            for  e = i+1; e < Sle_ListCutNum(pList); e++  {
                Vec_IntFillTwo( p.vLits, 2, Abc_Var2Lit(iCutVar0 + i, 1), Abc_Var2Lit(iCutVar0 + e, 1) );
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                assert((value != 0), "value");
            }
        }
        // cut requires fanin nodes
        Vec_WecInit( p.vEdgeCuts, Vec_IntSize(vCutFans) );
        // inlining of macro Sle_ForEachCut
        for  i, pCut = 0, pList + 1; i <  pList[0]; i++, pCut += Sle_CutSize(pCut) + 1  { // cuts with unit-cut

        
            var nSize int32    = Sle_CutSize(pCut);
            var k int32;
            var pC *int32 = Sle_CutLeaves(pCut);
            assert(nSize > 1, "nSize > 1");
            for  k = 0; k < nSize; k++  {
                if !Vec_BitEntry(p.vMask, pC[k]) {
                    continue; }
                Vec_IntFillTwo( p.vLits, 2, Abc_Var2Lit(iCutVar0 + i, 1), Abc_Var2Lit(pC[k], 0) );
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                assert((value != 0), "value");
                // find the edge ID between pC[k] and iObj
                iEdge = Vec_IntEntry(p.vObjMap, pC[k]);
                if iEdge == -1 {
                    Vec_IntWriteEntry( p.vObjMap, pC[k], (iEdge = nEdges++) );
                    Vec_WecPush( p.vFanoutEdges, pC[k], iEdgeVar0 + iEdge );
                }
                Vec_WecPush( p.vEdgeCuts, iEdge, iCutVar0 + i );
                p.nCutClas++;
            }
            // cut requires the node
            Vec_IntFillTwo( p.vLits, 2, Abc_Var2Lit(iCutVar0 + i, 1), Abc_Var2Lit(iObj, 0) );
            value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
            assert((value != 0), "value");
        }
        assert(nEdges == Vec_IntSize(vCutFans), "nEdges == Vec_IntSize(vCutFans)");

        // edge requires one of the fanout cuts
        // inlining of macro Vec_WecForEachLevel
        for  e = 0; (e < Vec_WecSize(p.vEdgeCuts)) && (((vArray) = Vec_WecEntry(p.vEdgeCuts, e)), 1); e++  {
            assert(Vec_IntSize(vArray) > 0, "Vec_IntSize(vArray) > 0");
            Vec_IntFill( p.vLits, 1, Abc_Var2Lit(iEdgeVar0 + e, 1) ); // neg lit (edge)
            // inlining of macro Vec_IntForEachEntry
            for  i = 0; (i < Vec_IntSize(vArray)) && (((Entry) = Vec_IntEntry(vArray, i)), 1); i++  {
                Vec_IntPush( p.vLits, Abc_Var2Lit(Entry, 0) );
            } // pos lit (cut)
            value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
            assert((value != 0), "value");
            p.nEdgeClas++;
        }

        // clean object map
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vCutFans)) && (((Entry) = Vec_IntEntry(vCutFans, i)), 1); i++  {
            Vec_IntWriteEntry( p.vObjMap, Entry, -1 );
        }
    }
    }

    // mutual exclusivity of edges
    // inlining of macro Vec_WecForEachLevel
    for  iObj = 0; (iObj < Vec_WecSize(p.vFanoutEdges)) && (((vArray) = Vec_WecEntry(p.vFanoutEdges, iObj)), 1); iObj++  {
        var j, k, EdgeJ, EdgeK int32;
        var iEdgeVar0 int32  = Vec_IntEntry( p.vEdgeFirst, iObj );
        var vCutFans *Vec_Int_t  = Vec_WecEntry( p.vCutFanins, iObj );
        // add fanin edges
        for  i = 0; i < Vec_IntSize(vCutFans); i++  {
            Vec_IntPush( vArray, iEdgeVar0 + i );
        }
        // generate pairs
        if fDynamic != 0 {
            continue; }
        // inlining of macro Vec_IntForEachEntry
        for  j = 0; (j < Vec_IntSize(vArray)) && (((EdgeJ) = Vec_IntEntry(vArray, j)), 1); j++  {
            // inlining of macro Vec_IntForEachEntryStart
            for  k = j+1; (k < Vec_IntSize(vArray)) && (((EdgeK) = Vec_IntEntry(vArray, k)), 1); k++  {
                Vec_IntFillTwo( p.vLits, 2, Abc_Var2Lit(EdgeJ, 1), Abc_Var2Lit(EdgeK, 1) );
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                assert((value != 0), "value");
            }
        }
        p.nEdgeClas2 += Vec_IntSize(vArray) * (Vec_IntSize(vArray) - 1) / 2;
    }

    // add delay clauses
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        var e, iFanin int32;
        var iEdgeVar0 int32  = Vec_IntEntry( p.vEdgeFirst, iObj );
        var iDelayVar0 int32  = Vec_IntEntry( p.vDelayFirst, iObj );
        var vCutFans *Vec_Int_t  = Vec_WecEntry( p.vCutFanins, iObj );
        // check if the node has cuts containing only primary inputs
        var pCut *int32;
        var pList *int32 = Sle_ManList( p, iObj );
        // inlining of macro Sle_ForEachCut
        for  i, pCut = 0, pList + 1; i <  pList[0]; i++, pCut += Sle_CutSize(pCut) + 1  { // cuts with unit-cut

            if Sle_ManCutHasPisOnly(pCut, p.vMask) != 0 {
                Vec_IntFill( p.vLits, 1, Abc_Var2Lit(iDelayVar0, 0) ); // pos lit
//                Vec_IntFillTwo( p->vLits, 2, Abc_Var2Lit(iObj, 1), Abc_Var2Lit(iDelayVar0, 0) );
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                assert((value != 0), "value");
                break;
            }
        }
//        if ( i < Sle_ListCutNum(pList) )
//            continue;
        // create delay requirements for each cut fanin of this node
        // inlining of macro Vec_IntForEachEntry
        for  e = 0; (e < Vec_IntSize(vCutFans)) && (((iFanin) = Vec_IntEntry(vCutFans, e)), 1); e++  {
            var d int32;
            var iDelayVarIn int32 = Vec_IntEntry( p.vDelayFirst, iFanin );
            for  d = 0; d < p.nLevels; d++  {
                Vec_IntClear( p.vLits );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iObj, 1) );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iFanin, 1) );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iDelayVarIn + d, 1) );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iEdgeVar0 + e, 0) );
                if d < p.nLevels-1 {
                    Vec_IntPush( p.vLits, Abc_Var2Lit(iDelayVar0 + d+1, 0) ); }
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                assert((value != 0), "value");

                Vec_IntClear( p.vLits );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iObj, 1) );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iFanin, 1) );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iDelayVarIn + d, 1) );
                if d < p.nLevels-1 {
                    Vec_IntPush( p.vLits, Abc_Var2Lit(iEdgeVar0 + e, 1) ); }
                Vec_IntPush( p.vLits, Abc_Var2Lit(iDelayVar0 + d, 0) );
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                assert((value != 0), "value");
            }
            p.nDelayClas += 2*p.nLevels;
        }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Add edge compatibility constraints.]

  Description [Returns 1 if constraints have been added.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManAddEdgeConstraints(p *Sle_Man_t, nEdges int32) int32 {
    var vArray *Vec_Int_t;
    var vTemp *Vec_Int_t  = Vec_IntAlloc( 100 );
    var value, iObj, nAdded int32;
    assert(nEdges == 1 || nEdges == 2, "nEdges == 1 || nEdges == 2");
    // inlining of macro Vec_WecForEachLevel
    for  iObj = 0; (iObj < Vec_WecSize(p.vFanoutEdges)) && (((vArray) = Vec_WecEntry(p.vFanoutEdges, iObj)), 1); iObj++  {
        var j, k, EdgeJ, EdgeK int32;
        // check if they are incompatible
        Vec_IntClear( vTemp );
        // inlining of macro Vec_IntForEachEntry
        for  j = 0; (j < Vec_IntSize(vArray)) && (((EdgeJ) = Vec_IntEntry(vArray, j)), 1); j++  {
            if sat_solver_var_value(p.pSat, EdgeJ) {
                Vec_IntPush( vTemp, EdgeJ ); }
        }
        if Vec_IntSize(vTemp) <= nEdges {
            continue; }
        nAdded++;
        if nEdges == 1 {
            // generate pairs
            // inlining of macro Vec_IntForEachEntry
            for  j = 0; (j < Vec_IntSize(vTemp)) && (((EdgeJ) = Vec_IntEntry(vTemp, j)), 1); j++  {
                // inlining of macro Vec_IntForEachEntryStart
                for  k = j+1; (k < Vec_IntSize(vTemp)) && (((EdgeK) = Vec_IntEntry(vTemp, k)), 1); k++  {
                    Vec_IntFillTwo( p.vLits, 2, Abc_Var2Lit(EdgeJ, 1), Abc_Var2Lit(EdgeK, 1) );
                    value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                    assert((value != 0), "value");
                }
            }
            p.nEdgeClas2 += Vec_IntSize(vTemp) * (Vec_IntSize(vTemp) - 1) / 2;
        
        } else if nEdges == 2 {
            var m, EdgeM int32;
            // generate triples
            // inlining of macro Vec_IntForEachEntry
            for  j = 0; (j < Vec_IntSize(vTemp)) && (((EdgeJ) = Vec_IntEntry(vTemp, j)), 1); j++  {
                // inlining of macro Vec_IntForEachEntryStart
                for  k = j+1; (k < Vec_IntSize(vTemp)) && (((EdgeK) = Vec_IntEntry(vTemp, k)), 1); k++  {
                    // inlining of macro Vec_IntForEachEntryStart
                    for  m = k+1; (m < Vec_IntSize(vTemp)) && (((EdgeM) = Vec_IntEntry(vTemp, m)), 1); m++  {
                        Vec_IntFillTwo( p.vLits, 2, Abc_Var2Lit(EdgeJ, 1), Abc_Var2Lit(EdgeK, 1) );
                        Vec_IntPush( p.vLits, Abc_Var2Lit(EdgeM, 1) );
                        value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                        assert((value != 0), "value");
                    }
                }
            }
            p.nEdgeClas2 += Vec_IntSize(vTemp) * (Vec_IntSize(vTemp) - 1) * (Vec_IntSize(vTemp) - 2) / 6;
        
        } else { assert(false, "0"); }
    }
    Vec_IntFree( vTemp );
    //printf( "Added clauses to %d nodes.\n", nAdded );
    return nAdded;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManDeriveResult(p *Sle_Man_t, vEdge2 *Vec_Int_t, vMapping *Vec_Int_t) {
    var vMapTemp *Vec_Int_t;
    var iObj int32;
    // create mapping
    Vec_IntFill( vMapping, Gia_ManObjNum(p.pGia), 0 );
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        var i int32;
        var iCut int32;
        var iCutVar0 int32 = Vec_IntEntry( p.vCutFirst, iObj );
        var pCut *int32;
        var pCutThis *int32 = nil;
        var pList *int32 = Sle_ManList( p, iObj );
        if !sat_solver_var_value(p.pSat, iObj) {
            continue; }
        // inlining of macro Sle_ForEachCut
        for  iCut, pCut = 0, pList + 1; iCut <  pList[0]; iCut++, pCut += Sle_CutSize(pCut) + 1  { // cuts with unit-cut

            if sat_solver_var_value(p.pSat, iCutVar0 + iCut) {
                assert(pCutThis == nil, "pCutThis == nil");
                pCutThis = pCut;
            }
        }
        assert(pCutThis != nil, "pCutThis != nil");
        Vec_IntWriteEntry( vMapping, iObj, Vec_IntSize(vMapping) );
        Vec_IntPush( vMapping, Sle_CutSize(pCutThis) );
        for  i = 0; i < Sle_CutSize(pCutThis); i++  {
            Vec_IntPush( vMapping, Sle_CutLeaves(pCutThis)[i] );
        }
        Vec_IntPush( vMapping, iObj );
    }
    }
    vMapTemp = p.pGia.vMapping;
    p.pGia.vMapping = vMapping;
    // collect edges
    Vec_IntClear( vEdge2 );
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        var i int32;
        var iFanin int32;
        var iEdgeVar0 int32 = Vec_IntEntry( p.vEdgeFirst, iObj );
        var vCutFans *Vec_Int_t  = Vec_WecEntry( p.vCutFanins, iObj );
        //int * pCut, * pList  = Sle_ManList( p, iObj );
        // int iCutVar0  = Vec_IntEntry( p->vCutFirst, iObj );
        if !sat_solver_var_value(p.pSat, iObj) {
            continue; }
        //for ( i = 0; i < Sle_ListCutNum(pList); i++ )
        //    printf( "%d", sat_solver_var_value(p->pSat, iCutVar0 + i) );
        //printf( "\n" );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vCutFans)) && (((iFanin) = Vec_IntEntry(vCutFans, i)), 1); i++  {
            if sat_solver_var_value(p.pSat, iFanin) && sat_solver_var_value(p.pSat, iEdgeVar0 + i) {
                // verify edge
                var pFanins *int32  = Gia_ObjLutFanins( p.pGia, iObj );
                var k int32;
                var nFanins int32 = Gia_ObjLutSize( p.pGia, iObj );
                for  k = 0; k < nFanins; k++  {
                    //printf( "%d ", pFanins[k] );
                    if pFanins[k] == iFanin {
                        break; }
                }
                //printf( "\n" );
                if k == nFanins {
//                    printf( "Cannot find LUT with input %d at node %d.\n", iFanin, iObj );
                    continue; }
                Vec_IntPushTwo( vEdge2, iFanin, iObj );
            }
        }
    }
    }
    p.pGia.vMapping = vMapTemp;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManExplore(pGia *Gia_Man_t, nBTLimit int32, DelayInit int32, fDynamic int32, fTwoEdges int32, fVerbose int32) {
    var fVeryVerbose int32;
    var clk abctime  = Abc_Clock();
    var vEdges2 *Vec_Int_t   = Vec_IntAlloc(1000);
    var vMapping *Vec_Int_t  = Vec_IntAlloc(1000);
    var i, iLut, nConfs, status, Delay, iFirstVar int32;
    var DelayStart int32;
    if (DelayInit != 0) || !Gia_ManHasMapping(pGia) {
        DelayStart = DelayInit;
    } else {
        DelayStart = Gia_ManLutLevel(pGia, nil);
    }
    var p *Sle_Man_t  = Sle_ManAlloc( pGia, DelayStart, fVerbose );
    if fVerbose != 0 {
        fmt.Printf("Running solver with %d conflicts, %d initial delay, and %d edges. Dynamic constraints = %s.\n", nBTLimit, DelayInit, 1+fTwoEdges, fDynamic?"yes":"no"); }
    Sle_ManMarkupVariables( p );
    if fVerbose != 0 {
        fmt.Printf("Vars:  Total = %d.  Node = %d. Cut = %d. Edge = %d. Delay = %d.\n", p.nVarsTotal, p.nNodeVars, p.nCutVars, p.nEdgeVars, p.nDelayVars); }
    Sle_ManDeriveInit( p );
    Sle_ManDeriveCnf( p, nBTLimit, fDynamic || fTwoEdges );
    if fVerbose != 0 {
        fmt.Printf("Clas:  Total = %d.  Cut = %d. Edge = %d. EdgeEx = %d. Delay = %d.\n", sat_solver_nclauses(p.pSat), p.nCutClas, p.nEdgeClas, p.nEdgeClas2, p.nDelayClas); }
    //Sat_SolverWriteDimacs( p->pSat, "temp.cnf", NULL, NULL, 0 );
    for  Delay = p.nLevels; Delay >= 0; Delay--  {
        // we constrain COs, although it would be fine to constrain only POs
        if Delay < p.nLevels {
            // inlining of macro Gia_ManForEachCoDriverId
            for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && (((iLut) = Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i))), 1); i++  {
                if Vec_BitEntry(p.vMask, iLut) { // internal node
                
                    iFirstVar = Vec_IntEntry( p.vDelayFirst, iLut );
                    if !sat_solver_push(p.pSat, Abc_Var2Lit(iFirstVar + Delay, 1)) {
                        break; }
                }
            }
            if i < Gia_ManCoNum(p.pGia) {
                if fVerbose != 0 {
                    fmt.Printf("Proved UNSAT for delay %d.  ", Delay);
                    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
                }
                break;
            }
        }
        // solve with assumptions
        //clk = Abc_Clock();
        nConfs = sat_solver_nconflicts( p.pSat );
        for {
            p.nSatCalls++;
            status = sat_solver_solve_internal( p.pSat );
            if status != l_True {
                break; }
            if !(Sle_ManAddEdgeConstraints(p, 1+fTwoEdges) != 0) {
                break; }
        }
        nConfs = sat_solver_nconflicts( p.pSat ) - nConfs;
        if status == l_True {
            if fVerbose != 0 {
                var nNodes, nEdges int32;
                for  i = 0; i < p.nNodeVars; i++  {
                    nNodes += sat_solver_var_value(p.pSat, i);
                }
                for  i = 0; i < p.nEdgeVars; i++  {
                    nEdges += sat_solver_var_value(p.pSat, p.nNodeVars + p.nCutVars + i);
                }
                fmt.Printf("Solution with delay %2d, node count %5d, and edge count %5d exists. Conf = %8d.  ", Delay, nNodes, nEdges, nConfs);
                Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            // save the result
            Sle_ManDeriveResult( p, vEdges2, vMapping );
            if fVeryVerbose != 0 {
                fmt.Printf("Nodes:  ");
                for  i = 0; i < p.nNodeVars; i++  {
                    if sat_solver_var_value(p.pSat, i) {
                        fmt.Printf("%d ", i); }
                }
                fmt.Printf("\n");
                fmt.Printf("\n");

                Vec_IntPrint( p.vCutFirst );
                fmt.Printf("Cuts:   ");
                for  i = 0; i < p.nCutVars; i++  {
                    if sat_solver_var_value(p.pSat, p.nNodeVars + i) {
                        fmt.Printf("%d ", p.nNodeVars + i); }
                }
                fmt.Printf("\n");
                fmt.Printf("\n");

                Vec_IntPrint( p.vEdgeFirst );
                fmt.Printf("Edges:  ");
                for  i = 0; i < p.nEdgeVars; i++  {
                    if sat_solver_var_value(p.pSat, p.nNodeVars + p.nCutVars + i) {
                        fmt.Printf("%d ", p.nNodeVars + p.nCutVars + i); }
                }
                fmt.Printf("\n");
                fmt.Printf("\n");

                Vec_IntPrint( p.vDelayFirst );
                fmt.Printf("Delays: ");
                for  i = 0; i < p.nDelayVars; i++  {
                    if sat_solver_var_value(p.pSat, p.nNodeVars + p.nCutVars + p.nEdgeVars + i) {
                        fmt.Printf("%d ", p.nNodeVars + p.nCutVars + p.nEdgeVars + i); }
                }
                fmt.Printf("\n");
                fmt.Printf("\n");
            }
        
        } else {
            if fVerbose != 0 {
                if status == l_False {
                    fmt.Printf("Proved UNSAT for delay %d. Conf = %8d.  ", Delay, nConfs);
                } else {
                    fmt.Printf("Resource limit reached for delay %d. Conf = %8d.  ", Delay, nConfs); }
                Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            break;
        }
    }
    if fVerbose != 0 {
        fmt.Printf("Clas:  Total = %d.  Cut = %d. Edge = %d. EdgeEx = %d. Delay = %d.  Calls = %d.\n", sat_solver_nclauses(p.pSat), p.nCutClas, p.nEdgeClas, p.nEdgeClas2, p.nDelayClas, p.nSatCalls); }
    if Vec_IntSize(vMapping) > 0 {
        Gia_ManEdgeFromArray( p.pGia, vEdges2 );
        Vec_IntFree( vEdges2 );
        Vec_IntFreeP( &p.pGia.vMapping );
        p.pGia.vMapping = vMapping;
    
    } else {
        Vec_IntFree( vEdges2 );
        Vec_IntFree( vMapping );
    }
    Vec_IntFreeP( &p.pGia.vPacking );
    Sle_ManStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

