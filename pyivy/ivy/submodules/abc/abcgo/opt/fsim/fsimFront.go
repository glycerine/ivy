package fsim

/**CFile****************************************************************

  FileName    [fsimFront.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast sequential AIG simulator.]

  Synopsis    [Simulation frontier.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: fsimFront.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fsimInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManStoreNum(p *Fsim_Man_t, Num int32) {
    var x uint32  = uint32(Num);
    assert(Num >= 0, "Num >= 0");
    for  x & ~0x7f  {
        *p.pDataCur++ = (x & 0x7f) | 0x80;
        x >>= 7;
    }
    *p.pDataCur++ = x;
    assert(p.pDataCur - p.pDataAig < p.nDataAig, "p.pDataCur - p.pDataAig < p.nDataAig");
}

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManRestoreNum(p *Fsim_Man_t) int32 {
    var ch, i, x int32;
    for  i = 0; (ch = *p.pDataCur++) & 0x80; i++  {
        x |= (ch & 0x7f) << (7 * i);
    }
    assert(p.pDataCur - p.pDataAig < p.nDataAig, "p.pDataCur - p.pDataAig < p.nDataAig");
    return x | (ch << (7 * i));
}

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManStoreObj(p *Fsim_Man_t, pObj *Fsim_Obj_t) {
    if p.pDataAig2 {
        *p.pDataCur2++ = pObj.iNode;
        *p.pDataCur2++ = pObj.iFan0;
        *p.pDataCur2++ = pObj.iFan1;
        return;
    }
    if pObj.iFan0 && pObj.iFan1 { // and
    
        assert(pObj.iNode, "pObj.iNode");
        assert(pObj.iNode >= p.iNodePrev, "pObj.iNode >= p.iNodePrev");
        assert((pObj.iNode << 1) > pObj.iFan0, "(pObj.iNode << 1) > pObj.iFan0");
        assert(pObj.iFan0 > pObj.iFan1, "pObj.iFan0 > pObj.iFan1");
        Fsim_ManStoreNum( p, ((pObj.iNode - p.iNodePrev) << 2) | 3 );
        Fsim_ManStoreNum( p, (pObj.iNode << 1) - pObj.iFan0 );
        Fsim_ManStoreNum( p, pObj.iFan0 - pObj.iFan1 );
        p.iNodePrev = pObj.iNode;
    
    } else if !pObj.iFan0 && !pObj.iFan1 { // ci
    
        assert(pObj.iNode, "pObj.iNode");
        assert(pObj.iNode >= p.iNodePrev, "pObj.iNode >= p.iNodePrev");
        Fsim_ManStoreNum( p, ((pObj.iNode - p.iNodePrev) << 2) | 1 );
        p.iNodePrev = pObj.iNode;
    
    } else { // if ( !pObj->iFan0 && pObj->iFan1 ) // co
    
        assert(pObj.iNode == 0, "pObj.iNode == 0");
        assert(pObj.iFan0 != 0, "pObj.iFan0 != 0");
        assert(pObj.iFan1 == 0, "pObj.iFan1 == 0");
        assert(((p.iNodePrev << 1) | 1) >= pObj.iFan0, "((p.iNodePrev << 1) | 1) >= pObj.iFan0");
        Fsim_ManStoreNum( p, (((p.iNodePrev << 1) | 1) - pObj.iFan0) << 1 );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManRestoreObj(p *Fsim_Man_t, pObj *Fsim_Obj_t) int32 {
    var iValue int32  = Fsim_ManRestoreNum( p );
    if (iValue & 3) == 3 { // and
    
        pObj.iNode = (iValue >> 2) + p.iNodePrev;
        pObj.iFan0 = (pObj.iNode << 1) - Fsim_ManRestoreNum( p );
        pObj.iFan1 = pObj.iFan0 - Fsim_ManRestoreNum( p );
        p.iNodePrev = pObj.iNode;
    
    } else if (iValue & 3) == 1 { // ci
    
        pObj.iNode = (iValue >> 2) + p.iNodePrev;
        pObj.iFan0 = 0;
        pObj.iFan1 = 0;
        p.iNodePrev = pObj.iNode;
    
    } else { // if ( (iValue & 1) == 0 ) // co
    
        pObj.iNode = 0;
        pObj.iFan0 = ((p.iNodePrev << 1) | 1) - (iValue >> 1);
        pObj.iFan1 = 0;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Determine the frontier.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManFrontFindNext(p *Fsim_Man_t, pFront string) int32 {
    assert(p.iNumber < (1 << 30) - p.nFront, "p.iNumber < (1 << 30) - p.nFront");
    for {
        if p.iNumber % p.nFront == 0 {
            p.iNumber++; }
        if pFront[p.iNumber % p.nFront] == 0 {
            pFront[p.iNumber % p.nFront] = 1;
            return p.iNumber;
        }
        p.iNumber++;
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Verifies the frontier.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManVerifyFront(p *Fsim_Man_t) {
    var pObj *Fsim_Obj_t;
    var pFans0, pFans1 *int32;  // representation of fanins
    var pFrontToId *int32;        // mapping of nodes into frontier variables
    var i, iVar0, iVar1 int32;
    // inlining of macro ABC_ALLOC
    pFans0 = ((int *) malloc(sizeof(int) * (size_t)(p.nObjs)));
    // inlining of macro ABC_ALLOC
    pFans1 = ((int *) malloc(sizeof(int) * (size_t)(p.nObjs)));
    pFans0[0] = pFans1[0] = 0;
    pFans0[1] = pFans1[1] = 0;
    // inlining of macro ABC_CALLOC
    pFrontToId = ((int *) calloc((size_t)(p.nFront), sizeof(int)));
    if Aig_ObjRefs(Aig_ManConst1(p.pAig)) {
        pFrontToId[1] = 1; }
    // inlining of macro Fsim_ManForEachObj
    for  i = 2, p.pDataCur = p.pDataAig, p.iNodePrev = 0, pObj = &p.Obj;        i < p.nObjs && (Fsim_ManRestoreObj( p, pObj ) != 0); i++  {
        if pObj.iNode {
            pFrontToId[pObj.iNode % p.nFront] = i; }
        iVar0 = Fsim_Lit2Var(pObj.iFan0);
        iVar1 = Fsim_Lit2Var(pObj.iFan1);
        pFans0[i] = Fsim_Var2Lit(pFrontToId[iVar0 % p.nFront], Fsim_LitIsCompl(pObj.iFan0));
        pFans1[i] = Fsim_Var2Lit(pFrontToId[iVar1 % p.nFront], Fsim_LitIsCompl(pObj.iFan1));
    }
    for  i = 0; i < p.nObjs; i++  {
        assert(pFans0[i] == p.pFans0[i], "pFans0[i] == p.pFans0[i]");
        assert(pFans1[i] == p.pFans1[i], "pFans1[i] == p.pFans1[i]");
    }
    pFrontToId = nil;
    pFans0 = nil;
    pFans1 = nil;
}

/**Function*************************************************************

  Synopsis    [Determine the frontier.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManFront(p *Fsim_Man_t, fCompressAig int32) {
    var Obj Fsim_Obj_t;
    var pObj *Fsim_Obj_t = &Obj;
    var pFront string;    // places used for the frontier
    var pIdToFront *int32; // mapping of nodes into frontier places
    var i, iVar0, iVar1, nCrossCut, nCrossCutMax int32;
    // start the frontier
    // inlining of macro ABC_CALLOC
    pFront = ((char *) calloc((size_t)(p.nFront), sizeof(char)));
    // inlining of macro ABC_ALLOC
    pIdToFront = ((int *) malloc(sizeof(int) * (size_t)(p.nObjs)));
    pIdToFront[0] = -1;
    pIdToFront[1] = -1;
    // add constant node
    p.iNumber = 1;
    if p.pRefs[1] {
        pIdToFront[1] = Fsim_ManFrontFindNext( p, pFront );
        nCrossCut = 1;
    }
    // allocate room for data
    if fCompressAig != 0 {
        p.nDataAig = p.nObjs * 6;
        // inlining of macro ABC_ALLOC
        p.pDataAig = ((unsigned char *) malloc(sizeof(unsigned char) * (size_t)(p.nDataAig)));
        p.pDataCur = p.pDataAig;
        p.iNodePrev = 0;
    
    } else {
        // inlining of macro ABC_ALLOC
        p.pDataAig2 = ((int *) malloc(sizeof(int) * (size_t)(3 * p.nObjs)));
        p.pDataCur2 = p.pDataAig2 + 6;
    }
    // iterate through the objects
    for  i = 2; i < p.nObjs; i++  {
        if p.pFans0[i] == 0 { // ci
        
            // store node
            pIdToFront[i] = Fsim_ManFrontFindNext( p, pFront );
            pObj.iNode = pIdToFront[i];
            pObj.iFan0 = 0;
            pObj.iFan1 = 0;
            Fsim_ManStoreObj( p, pObj );
            // handle CIs without fanout
            if p.pRefs[i] == 0 {
                pFront[pIdToFront[i] % p.nFront] = 0;
                pIdToFront[i] = -1;
            }
        
        } else if p.pFans1[i] == 0 { // co
        
            assert(p.pRefs[i] == 0, "p.pRefs[i] == 0");
            // get the fanin
            iVar0 = Fsim_Lit2Var(p.pFans0[i]);
            assert(pIdToFront[iVar0] > 0, "pIdToFront[iVar0] > 0");
            // store node
            pObj.iNode = 0;
            pObj.iFan0 = Fsim_Var2Lit(pIdToFront[iVar0], Fsim_LitIsCompl(p.pFans0[i]));
            pObj.iFan1 = 0;
            Fsim_ManStoreObj( p, pObj );
            // deref the fanin
            if --p.pRefs[iVar0] == 0 {
                pFront[pIdToFront[iVar0] % p.nFront] = 0;
                pIdToFront[iVar0] = -1;
                nCrossCut--;
            }
        
        } else {
            // get the fanins
            iVar0 = Fsim_Lit2Var(p.pFans0[i]);
            assert(pIdToFront[iVar0] > 0, "pIdToFront[iVar0] > 0");
            iVar1 = Fsim_Lit2Var(p.pFans1[i]);
            assert(pIdToFront[iVar1] > 0, "pIdToFront[iVar1] > 0");
            // store node
            pIdToFront[i] = Fsim_ManFrontFindNext( p, pFront );
            pObj.iNode = pIdToFront[i];
            pObj.iFan0 = Fsim_Var2Lit(pIdToFront[iVar0], Fsim_LitIsCompl(p.pFans0[i]));
            pObj.iFan1 = Fsim_Var2Lit(pIdToFront[iVar1], Fsim_LitIsCompl(p.pFans1[i]));
            Fsim_ManStoreObj( p, pObj );
            // deref the fanins
            if --p.pRefs[iVar0] == 0 {
                pFront[pIdToFront[iVar0] % p.nFront] = 0;
                pIdToFront[iVar0] = -1;
                nCrossCut--;
            }
            if --p.pRefs[iVar1] == 0 {
                pFront[pIdToFront[iVar1] % p.nFront] = 0;
                pIdToFront[iVar1] = -1;
                nCrossCut--;
            }
            // handle nodes without fanout (choice nodes)
            if p.pRefs[i] == 0 {
                pFront[pIdToFront[i] % p.nFront] = 0;
                pIdToFront[i] = -1;
            }
        }
        if p.pRefs[i] {
            if nCrossCutMax < ++nCrossCut {
                nCrossCutMax = nCrossCut; } }
    }
    assert(p.pDataAig2 == nil || p.pDataCur2 - p.pDataAig2 == (3 * p.nObjs), "p.pDataAig2 == nil || p.pDataCur2 - p.pDataAig2 == (3 * p.nObjs)");
    assert(nCrossCut == 0, "nCrossCut == 0");
    assert(nCrossCutMax == p.nCrossCutMax, "nCrossCutMax == p.nCrossCutMax");
    for  i = 0; i < p.nFront; i++  {
        assert(pFront[i] == 0, "pFront[i] == 0");
    }
    pFront = nil;
    pIdToFront = nil;
//    Fsim_ManVerifyFront( p );
    p.pFans0 = nil;
    p.pFans1 = nil;
    p.pRefs = nil;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

