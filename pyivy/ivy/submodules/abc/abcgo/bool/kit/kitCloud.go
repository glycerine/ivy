package kit

import "fmt"

/**CFile****************************************************************

  FileName    [kitCloud.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Computation kit.]

  Synopsis    [Procedures using BDD package CLOUD.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Dec 6, 2006.]

  Revision    [$Id: kitCloud.c,v 1.00 2006/12/06 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// internal representation of the function to be decomposed
   
type Kit_Mux_t struct {
    v uint32; /* C bitfield: 5 bits; review packing and truncation */          // variable 
    t uint32; /* C bitfield: 12 bits; review packing and truncation */          // then edge
    e uint32; /* C bitfield: 12 bits; review packing and truncation */          // else edge
    c uint32; /* C bitfield: 1 bits; review packing and truncation */          // complemented attr of else edge
    i uint32; /* C bitfield: 1 bits; review packing and truncation */          // complemented attr of top node
}

func Kit_Mux2Int(m Kit_Mux_t) int32 { union { var x Kit_Mux_t; var y int32; }
 v; v.x = m; return v.y;  }
func Kit_Int2Mux(m int32) Kit_Mux_t { union { var x Kit_Mux_t; var y int32; }
 v; v.y = m; return v.x;  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Derive BDD from the truth table for 5 variable functions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Kit_TruthToCloud5_rec_uVars [5]uint32 = [5]uint32{ 0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00, 0xFFFF0000 };
func Kit_TruthToCloud5_rec(dd *CloudManager, uTruth uint32, nVars int32, nVarsAll int32) *CloudNode {
    
    var pCof0, pCof1 *CloudNode;
    var uCof0, uCof1 uint32;
    assert(nVars <= 5, "nVars <= 5");
    if uTruth == 0 {
        return dd.zero; }
    if uTruth == ~0 {
        return dd.one; }
    if nVars == 1 {
        if uTruth == spStatic_Kit_TruthToCloud5_rec_uVars[0] {
            return dd.vars[nVarsAll-1]; }
        if uTruth == ~spStatic_Kit_TruthToCloud5_rec_uVars[0] {
            return Cloud_Not(dd.vars[nVarsAll-1]); }
        assert(false, "0");
    }
//    Count++;
    assert(nVars > 1, "nVars > 1");
    uCof0 = uTruth & ~spStatic_Kit_TruthToCloud5_rec_uVars[nVars-1];
    uCof1 = uTruth &  spStatic_Kit_TruthToCloud5_rec_uVars[nVars-1];
    uCof0 |= uCof0 << (1<<(nVars-1));
    uCof1 |= uCof1 >> (1<<(nVars-1));
    if uCof0 == uCof1 {
        return Kit_TruthToCloud5_rec( dd, uCof0, nVars - 1, nVarsAll ); }
    if uCof0 == ~uCof1 {
        pCof0 = Kit_TruthToCloud5_rec( dd, uCof0, nVars - 1, nVarsAll );
        pCof1 = Cloud_Not( pCof0 );
    
    } else {
        pCof0 = Kit_TruthToCloud5_rec( dd, uCof0, nVars - 1, nVarsAll );
        pCof1 = Kit_TruthToCloud5_rec( dd, uCof1, nVars - 1, nVarsAll );
    }
    return Cloud_MakeNode( dd, nVarsAll - nVars, pCof1, pCof0 );
}

/**Function********************************************************************

  Synopsis    [Compute BDD for the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func Kit_TruthToCloud_rec(dd *CloudManager, pTruth *uint32, nVars int32, nVarsAll int32) *CloudNode {
    var pCof0, pCof1 *CloudNode;
    var pTruth0, pTruth1 *uint32;
    if nVars <= 5 {
        return Kit_TruthToCloud5_rec( dd, pTruth[0], nVars, nVarsAll ); }
    if Kit_TruthIsConst0(pTruth, nVars) {
        return dd.zero; }
    if Kit_TruthIsConst1(pTruth, nVars) {
        return dd.one; }
//    Count++;
    pTruth0 = pTruth;
    pTruth1 = pTruth + Kit_TruthWordNum(nVars-1);
    if Kit_TruthIsEqual( pTruth0, pTruth1, nVars - 1 ) {
        return Kit_TruthToCloud_rec( dd, pTruth0, nVars - 1, nVarsAll ); }
    if Kit_TruthIsOpposite( pTruth0, pTruth1, nVars - 1 ) {
        pCof0 = Kit_TruthToCloud_rec( dd, pTruth0, nVars - 1, nVarsAll );
        pCof1 = Cloud_Not( pCof0 );
    
    } else {
        pCof0 = Kit_TruthToCloud_rec( dd, pTruth0, nVars - 1, nVarsAll );
        pCof1 = Kit_TruthToCloud_rec( dd, pTruth1, nVars - 1, nVarsAll );
    }
    return Cloud_MakeNode( dd, nVarsAll - nVars, pCof1, pCof0 );
}

/**Function********************************************************************

  Synopsis    [Compute BDD for the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func Kit_TruthToCloud(dd *CloudManager, pTruth *uint32, nVars int32) *CloudNode {
    var pRes *CloudNode;
    pRes = Kit_TruthToCloud_rec( dd, pTruth, nVars, nVars );
//    printf( "%d/%d  ", Count, Cloud_DagSize(dd, pRes) );
    return pRes;
}

/**Function********************************************************************

  Synopsis    [Transforms the array of BDDs into the integer array.]

  Description []
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func Kit_CreateCloud(dd *CloudManager, pFunc *CloudNode, vNodes *Vec_Int_t) int32 {
    var Mux Kit_Mux_t;
    var nNodes, i int32;
    // collect BDD nodes
    nNodes = Cloud_DagCollect( dd, pFunc );
    if nNodes >= (1<<12) { // because in Kit_Mux_t edge is 12 bit
        return 0; }
    assert(nNodes == Cloud_DagSize( dd, pFunc ), "nNodes == Cloud_DagSize( dd, pFunc )");
    assert(nNodes < dd.nNodesLimit, "nNodes < dd.nNodesLimit");
    Vec_IntClear( vNodes );
    Vec_IntPush( vNodes, 0 ); // const1 node
    dd.ppNodes[0].s = 0;
    for  i = 1; i < nNodes; i++  {
        dd.ppNodes[i].s = i;
        Mux.v = dd.ppNodes[i].v;
        Mux.t = dd.ppNodes[i].t.s;
        Mux.e = Cloud_Regular(dd.ppNodes[i].e).s;
        Mux.c = Cloud_IsComplement(dd.ppNodes[i].e); 
        if i == nNodes - 1 {
            Mux.i = Cloud_IsComplement(pFunc);
        } else {
            Mux.i = 0;
        }
        // put the MUX into the array
        Vec_IntPush( vNodes, Kit_Mux2Int(Mux) );
    }
    assert(Vec_IntSize(vNodes) == nNodes, "Vec_IntSize(vNodes) == nNodes");
    // reset signatures
    for  i = 0; i < nNodes; i++  {
        dd.ppNodes[i].s = dd.nSignCur;
    }
    return 1;
}

/**Function********************************************************************

  Synopsis    [Transforms the array of BDDs into the integer array.]

  Description []
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func Kit_CreateCloudFromTruth(dd *CloudManager, pTruth *uint32, nVars int32, vNodes *Vec_Int_t) int32 {
    var pFunc *CloudNode;
    Cloud_Restart( dd );
    pFunc = Kit_TruthToCloud( dd, pTruth, nVars );
    Vec_IntClear( vNodes );
    return Kit_CreateCloud( dd, pFunc, vNodes );
}

/**Function*************************************************************

  Synopsis    [Computes composition of truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_CloudToTruth(vNodes *Vec_Int_t, nVars int32, vStore *Vec_Ptr_t, fInv int32) *uint32 {
    var pThis, pFan0, pFan1 *uint32;
    var Mux Kit_Mux_t;
    var i, Entry int32;
    assert(Vec_IntSize(vNodes) <= Vec_PtrSize(vStore), "Vec_IntSize(vNodes) <= Vec_PtrSize(vStore)");
    pThis = (unsigned *)Vec_PtrEntry( vStore, 0 );
    Kit_TruthFill( pThis, nVars );
    Vec_IntForEachEntryStart( vNodes, Entry, i, 1 )
    {
        Mux = Kit_Int2Mux(Entry);
        assert(int32(Mux.e) < i && int32(Mux.t) < i && int32(Mux.v) < nVars, "int32(Mux.e) < i && int32(Mux.t) < i && int32(Mux.v) < nVars");          
        pFan0 = (unsigned *)Vec_PtrEntry( vStore, Mux.e );
        pFan1 = (unsigned *)Vec_PtrEntry( vStore, Mux.t );
        pThis = (unsigned *)Vec_PtrEntry( vStore, i );
        Kit_TruthMuxVarPhase( pThis, pFan0, pFan1, nVars, fInv? Mux.v : nVars-1-Mux.v, Mux.c );
    } 
    // complement the result
    if Mux.i {
        Kit_TruthNot( pThis, pThis, nVars ); }
    return pThis;
}

/**Function*************************************************************

  Synopsis    [Computes composition of truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_TruthCompose(dd *CloudManager, pTruth *uint32, nVars int32, pInputs **uint32, nVarsAll int32, vStore *Vec_Ptr_t, vNodes *Vec_Int_t) *uint32 {
    var pFunc *CloudNode;
    var pThis, pFan0, pFan1 *uint32;
    var Mux Kit_Mux_t;
    var i, Entry, RetValue int32;
    // derive BDD from truth table
    Cloud_Restart( dd );
    pFunc = Kit_TruthToCloud( dd, pTruth, nVars );
    // convert it into nodes
    RetValue = Kit_CreateCloud( dd, pFunc, vNodes );
    if RetValue == 0 {
        fmt.Printf("Kit_TruthCompose(): Internal failure!!!\n"); }
    // verify the result
//    pFan0 = Kit_CloudToTruth( vNodes, nVars, vStore, 0 );
//    if ( !Kit_TruthIsEqual( pTruth, pFan0, nVars ) )
//        printf( "Failed!\n" );
    // compute truth table from the BDD
    assert(Vec_IntSize(vNodes) <= Vec_PtrSize(vStore), "Vec_IntSize(vNodes) <= Vec_PtrSize(vStore)");
    pThis = (unsigned *)Vec_PtrEntry( vStore, 0 );
    Kit_TruthFill( pThis, nVarsAll );
    Vec_IntForEachEntryStart( vNodes, Entry, i, 1 )
    {
        Mux = Kit_Int2Mux(Entry);
        pFan0 = (unsigned *)Vec_PtrEntry( vStore, Mux.e );
        pFan1 = (unsigned *)Vec_PtrEntry( vStore, Mux.t );
        pThis = (unsigned *)Vec_PtrEntry( vStore, i );
        Kit_TruthMuxPhase( pThis, pFan0, pFan1, pInputs[nVars-1-Mux.v], nVarsAll, Mux.c );
    }
    // complement the result
    if Mux.i {
        Kit_TruthNot( pThis, pThis, nVarsAll ); }
    return pThis;
}

/**Function********************************************************************

  Synopsis    [Compute BDD for the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func Kit_TruthCofSupports(vBddDir *Vec_Int_t, vBddInv *Vec_Int_t, nVars int32, vMemory *Vec_Int_t, puSupps *uint32) {
    var Mux Kit_Mux_t;
    var puSuppAll *uint32;
    var pThis *uint32  = nil; // Suppress "might be used uninitialized"
    var pFan0, pFan1 *uint32;
    var i, v, Var, Entry, nSupps int32;
    nSupps = 2 * nVars;

    // extend storage
    if Vec_IntSize( vMemory ) < nSupps * Vec_IntSize(vBddDir) {
        Vec_IntGrow( vMemory, nSupps * Vec_IntSize(vBddDir) ); }
    puSuppAll = (unsigned *)Vec_IntArray( vMemory );
    // clear storage for the const node
    memset( puSuppAll, 0, sizeof(unsigned) * nSupps );
    // compute supports from nodes
    Vec_IntForEachEntryStart( vBddDir, Entry, i, 1 )
    {
        Mux = Kit_Int2Mux(Entry);
        Var = nVars - 1 - Mux.v;
        pFan0 = puSuppAll + nSupps * Mux.e;
        pFan1 = puSuppAll + nSupps * Mux.t;
        pThis = puSuppAll + nSupps * i;
        for  v = 0; v < nSupps; v++  {
            pThis[v] = pFan0[v] | pFan1[v] | (1<<Var);
        }
        assert(pFan0[2*Var + 0] == pFan0[2*Var + 1], "pFan0[2*Var + 0] == pFan0[2*Var + 1]");
        assert(pFan1[2*Var + 0] == pFan1[2*Var + 1], "pFan1[2*Var + 0] == pFan1[2*Var + 1]");
        pThis[2*Var + 0] = pFan0[2*Var + 0];// | pFan0[2*Var + 1];
        pThis[2*Var + 1] = pFan1[2*Var + 0];// | pFan1[2*Var + 1];
    }
    // copy the result
    memcpy( puSupps, pThis, sizeof(unsigned) * nSupps );
    // compute the inverse

    // extend storage
    if Vec_IntSize( vMemory ) < nSupps * Vec_IntSize(vBddInv) {
        Vec_IntGrow( vMemory, nSupps * Vec_IntSize(vBddInv) ); }
    puSuppAll = (unsigned *)Vec_IntArray( vMemory );
    // clear storage for the const node
    memset( puSuppAll, 0, sizeof(unsigned) * nSupps );
    // compute supports from nodes
    Vec_IntForEachEntryStart( vBddInv, Entry, i, 1 )
    {
        Mux = Kit_Int2Mux(Entry);
//        Var = nVars - 1 - Mux.v;
        Var = Mux.v;
        pFan0 = puSuppAll + nSupps * Mux.e;
        pFan1 = puSuppAll + nSupps * Mux.t;
        pThis = puSuppAll + nSupps * i;
        for  v = 0; v < nSupps; v++  {
            pThis[v] = pFan0[v] | pFan1[v] | (1<<Var);
        }
        assert(pFan0[2*Var + 0] == pFan0[2*Var + 1], "pFan0[2*Var + 0] == pFan0[2*Var + 1]");
        assert(pFan1[2*Var + 0] == pFan1[2*Var + 1], "pFan1[2*Var + 0] == pFan1[2*Var + 1]");
        pThis[2*Var + 0] = pFan0[2*Var + 0];// | pFan0[2*Var + 1];
        pThis[2*Var + 1] = pFan1[2*Var + 0];// | pFan1[2*Var + 1];
    }

    // merge supports
    for  Var = 0; Var < nSupps; Var++  {
        puSupps[Var] = (puSupps[Var] & Kit_BitMask(Var/2)) | (pThis[Var] & ~Kit_BitMask(Var/2));
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

