package cec

/**CFile****************************************************************

  FileName    [cecIso.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Detection of structural isomorphism.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecIso.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Cec_ManIsoInfo(pStore *uint32, nWords int32, Id int32) *uint32 { return pStore + nWords * Id; }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes simulation info for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIsoSimulate(pObj *Gia_Obj_t, Id int32, pStore *uint32, nWords int32) {
    var pInfo *uint32   = Cec_ManIsoInfo( pStore, nWords, Id );
    var pInfo0 *uint32  = Cec_ManIsoInfo( pStore, nWords, Gia_ObjFaninId0(pObj, Id) );
    var pInfo1 *uint32  = Cec_ManIsoInfo( pStore, nWords, Gia_ObjFaninId1(pObj, Id) );
    var w int32;
    if Gia_ObjFaninC0(pObj) {
        if Gia_ObjFaninC1(pObj) {
            for  w = 0; w < nWords; w++  {
                pInfo[w] = ~(pInfo0[w] | pInfo1[w]);
            }
        } else { 
            for  w = 0; w < nWords; w++  {
                pInfo[w] = ~pInfo0[w] & pInfo1[w];
            } }
    
    } else {
        if Gia_ObjFaninC1(pObj) {
            for  w = 0; w < nWords; w++  {
                pInfo[w] = pInfo0[w] & ~pInfo1[w];
            }
        } else { 
            for  w = 0; w < nWords; w++  {
                pInfo[w] = pInfo0[w] & pInfo1[w];
            } }
    }
}

/**Function*************************************************************

  Synopsis    [Copies simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIsoCopy(IdDest int32, IdSour int32, pStore *uint32, nWords int32) {
    var pInfo0 *uint32  = Cec_ManIsoInfo( pStore, nWords, IdDest );
    var pInfo1 *uint32  = Cec_ManIsoInfo( pStore, nWords, IdSour );
    var w int32;
    for  w = 0; w < nWords; w++  {
        pInfo0[w] = pInfo1[w];
    }
}

/**Function*************************************************************

  Synopsis    [Compares simulation info of two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIsoEqual(Id0 int32, Id1 int32, pStore *uint32, nWords int32) int32 {
    var pInfo0 *uint32  = Cec_ManIsoInfo( pStore, nWords, Id0 );
    var pInfo1 *uint32  = Cec_ManIsoInfo( pStore, nWords, Id1 );
    var w int32;
    for  w = 0; w < nWords; w++  {
        if pInfo0[w] != pInfo1[w] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Generates random simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIsoRandom(Id int32, pStore *uint32, nWords int32) {
    var pInfo0 *uint32  = Cec_ManIsoInfo( pStore, nWords, Id );
    var w int32;
    for  w = 0; w < nWords; w++  {
        pInfo0[w] = Gia_ManRandom( 0 );
    }
}

/**Function*************************************************************

  Synopsis    [Computes hash key of the simuation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Gia_ManIsoHashKey_s_Primes [16]int32 = [16]int32{ 
        1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 
        4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147 };
func Gia_ManIsoHashKey(Id int32, pStore *uint32, nWords int32, nTableSize int32) int32 {
    
    var pInfo0 *uint32  = Cec_ManIsoInfo( pStore, nWords, Id );
    var uHash uint32;
    var i int32;
    for  i = 0; i < nWords; i++  {
        uHash ^= pInfo0[i] * spStatic_Gia_ManIsoHashKey_s_Primes[i & 0xf];
    }
    return int32((uHash % nTableSize));

}

/**Function*************************************************************

  Synopsis    [Adds node to the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIsoTableAdd(p *Gia_Man_t, Id int32, pStore *uint32, nWords int32, pTable *int32, nTableSize int32) {
    var pTemp *Gia_Obj_t;
    var Key int32;
    var Ent int32;
    var Color int32 = Gia_ObjColors( p, Id );
    assert(Color == 1 || Color == 2, "Color == 1 || Color == 2");
    Key = Gia_ManIsoHashKey( Id, pStore, nWords, nTableSize );
    for  Ent = pTable[Key],  pTemp = (Ent ?  Gia_ManObj(p, Ent) : nil);  (pTemp != nil);
          Ent = pTemp.Value, pTemp = (Ent ?  Gia_ManObj(p, Ent) : nil)  {
        if Gia_ObjColors( p, Ent ) != Color {
            continue; }
        if !(Gia_ManIsoEqual( Id, Ent, pStore, nWords ) != 0) {
            continue; }
        // found node with the same color and signature - mark it and do not add new node
        pTemp.fMark0 = 1;
        return;
    }
    // did not find the node with the same color and signature - add new node
    pTemp = Gia_ManObj( p, Id );
    assert(pTemp.Value == 0, "pTemp.Value == 0");
    assert(pTemp.fMark0 == 0, "pTemp.fMark0 == 0");
    pTemp.Value = pTable[Key];
    pTable[Key] = Id;
}

/**Function*************************************************************

  Synopsis    [Extracts equivalence class candidates from one bin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIsoExtractClasses(p *Gia_Man_t, Bin int32, pStore *uint32, nWords int32, vNodesA *Vec_Int_t, vNodesB *Vec_Int_t) int32 {
    var pTemp *Gia_Obj_t;
    var Ent int32;
    Vec_IntClear( vNodesA );
    Vec_IntClear( vNodesB );
    for  Ent = Bin,          pTemp = (Ent ?  Gia_ManObj(p, Ent) : nil);  (pTemp != nil);
          Ent = pTemp.Value, pTemp = (Ent ?  Gia_ManObj(p, Ent) : nil)  {
        if pTemp.fMark0 {
            pTemp.fMark0 = 0;
            continue;
        }
        if Gia_ObjColors( p, Ent ) == 1 {
            Vec_IntPush( vNodesA, Ent );
        } else {
            Vec_IntPush( vNodesB, Ent ); }
    }
    return Vec_IntSize(vNodesA) > 0 && Vec_IntSize(vNodesB) > 0;    
}

/**Function*************************************************************

  Synopsis    [Matches nodes in the extacted classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIsoMatchNodes(pIso *int32, pStore *uint32, nWords int32, vNodesA *Vec_Int_t, vNodesB *Vec_Int_t) {
    var k0, k1, IdA, IdB int32;
    // inlining of macro Vec_IntForEachEntry
    for  k0 = 0; (k0 < Vec_IntSize(vNodesA)) && (((IdA) = Vec_IntEntry(vNodesA, k0)), 1); k0++  {
    // inlining of macro Vec_IntForEachEntry
    for  k1 = 0; (k1 < Vec_IntSize(vNodesB)) && (((IdB) = Vec_IntEntry(vNodesB, k1)), 1); k1++  {
        if Gia_ManIsoEqual( IdA, IdB, pStore, nWords ) != 0 {
            assert(pIso[IdA] == 0, "pIso[IdA] == 0");
            assert(pIso[IdB] == 0, "pIso[IdB] == 0");
            assert(IdA != IdB, "IdA != IdB");
            pIso[IdA] = IdB;
            pIso[IdB] = IdA;
            continue;
        }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Transforms iso into equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManTransformClasses(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    assert(p.pReprs && p.pNexts && p.pIso, "p.pReprs && p.pNexts && p.pIso");
    memset( p.pReprs, 0, sizeof(int) * Gia_ManObjNum(p) );
    memset( p.pNexts, 0, sizeof(int) * Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        p.pReprs[i].iRepr = GIA_VOID;
        if p.pIso[i] && p.pIso[i] < i {
            p.pReprs[i].iRepr = p.pIso[i];
            p.pNexts[p.pIso[i]] = i;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Finds node correspondences in the miter.]

  Description [Assumes that the colors are assigned.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManDetectIsomorphism(p *Gia_Man_t) *int32 {
    var nWords int32  = 2;
    var pObj *Gia_Obj_t;
    var vNodesA, vNodesB *Vec_Int_t;
    var pStore *uint32;
    var Counter uint32;
    var i int32;
    var pIso *int32;
    var pTable *int32;
    var nTableSize int32;
    // start equivalence classes
    // inlining of macro ABC_CALLOC
    pIso = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCo(pObj) {
            assert(Gia_ObjColors(p, i) == 0, "Gia_ObjColors(p, i) == 0");
            continue;
        }
        assert(Gia_ObjColors(p, i), "Gia_ObjColors(p, i)");
        if Gia_ObjColors(p, i) == 3 {
            pIso[i] = i; }
    }
    // start simulation info
    // inlining of macro ABC_ALLOC
    pStore = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(Gia_ManObjNum(p) * nWords)));
    // simulate and create table
    nTableSize = Abc_PrimeCudd( 100 + Gia_ManObjNum(p)/2 );
    // inlining of macro ABC_CALLOC
    pTable = ((int *) calloc((size_t)(nTableSize), sizeof(int)));
    Gia_ManCleanValue( p );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCo(pObj) {
            continue; }
        if pIso[i] == 0 { // simulate
            Gia_ManIsoSimulate( pObj, i, pStore, nWords );
        } else if pIso[i] < i { // copy
            Gia_ManIsoCopy( i, pIso[i], pStore, nWords );
        } else { // generate
            Gia_ManIsoRandom( i, pStore, nWords ); }
        if pIso[i] == 0 {
            Gia_ManIsoTableAdd( p, i, pStore, nWords, pTable, nTableSize ); }
    }
    // create equivalence classes
    vNodesA = Vec_IntAlloc( 100 );
    vNodesB = Vec_IntAlloc( 100 );
    for  i = 0; i < nTableSize; i++  {
        if Gia_ManIsoExtractClasses( p, pTable[i], pStore, nWords, vNodesA, vNodesB ) != 0 {
            Gia_ManIsoMatchNodes( pIso, pStore, nWords, vNodesA, vNodesB ); }
    }
    Vec_IntFree( vNodesA );
    Vec_IntFree( vNodesB );
    // collect info
    Counter = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        Counter += (pIso[i] && pIso[i] < i);
/*
        if ( pIso[i] && pIso[i] < i )
        {
            if ( (Gia_ObjIsHead(p,pIso[i]) && Gia_ObjRepr(p,i)==pIso[i]) || 
                 (Gia_ObjIsClass(p,pIso[i]) && Gia_ObjRepr(p,i)==Gia_ObjRepr(p,pIso[i])) )
                 Abc_Print( 1, "1" );
            else
                Abc_Print( 1, "0" );
        }
*/
    }
    Abc_Print( 1, "Computed %d pairs of structurally equivalent nodes.\n", Counter );
//    p->pIso = pIso;
//    Cec_ManTransformClasses( p );

    pTable = nil;
    pStore = nil;
    return pIso;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

