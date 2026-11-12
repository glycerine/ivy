package fraig

/**CFile****************************************************************

  FileName    [fraigNode.c]

  PackageName [FRAIG: Functionally reduced AND-INV graphs.]

  Synopsis    [Implementation of the FRAIG node.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - October 1, 2004]

  Revision    [$Id: fraigNode.c,v 1.3 2005/07/08 01:01:32 alanmi Exp $]

***********************************************************************/

//#include "fraigInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// returns the complemented attribute of the node
// Original macro:
// #define Fraig_NodeIsSimComplement(p) (Fraig_IsComplement(p)? !(Fraig_Regular(p)->fInv) : (p)->fInv)


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates the constant 1 node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeCreateConst(p *Fraig_Man_t) *Fraig_Node_t {
    var pNode *Fraig_Node_t;

    // create the node
    pNode = (Fraig_Node_t *)Fraig_MemFixedEntryFetch( p.mmNodes );
    memset( pNode, 0, sizeof(Fraig_Node_t) );

    // assign the number and add to the array of nodes
    pNode.Num   = p.vNodes.nSize;
    Fraig_NodeVecPush( p.vNodes,  pNode );
    pNode.NumPi = -1;  // this is not a PI, so its number is -1
    pNode.Level =  0;  // just like a PI, it has 0 level
    pNode.nRefs =  1;  // it is a persistent node, which comes referenced
    pNode.fInv  =  1;  // the simulation info is complemented

    // create the simulation info
    pNode.puSimR = (unsigned *)Fraig_MemFixedEntryFetch( p.mmSims );
    pNode.puSimD = pNode.puSimR + p.nWordsRand;
    memset( pNode.puSimR, 0, sizeof(unsigned) * p.nWordsRand );
    memset( pNode.puSimD, 0, sizeof(unsigned) * p.nWordsDyna );

    // count the number of ones in the simulation vector
    pNode.nOnes = p.nWordsRand * sizeof(unsigned) * 8;

    // insert it into the hash table
    Fraig_HashTableLookupF0( p, pNode );
    return pNode;
}

/**Function*************************************************************

  Synopsis    [Creates a primary input node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeCreatePi(p *Fraig_Man_t) *Fraig_Node_t {
    var pNode, pNodeRes *Fraig_Node_t;
    var i int32;
    var clk abctime;

    // create the node
    pNode = (Fraig_Node_t *)Fraig_MemFixedEntryFetch( p.mmNodes );
    memset( pNode, 0, sizeof(Fraig_Node_t) );
    pNode.puSimR = (unsigned *)Fraig_MemFixedEntryFetch( p.mmSims );
    pNode.puSimD = pNode.puSimR + p.nWordsRand;
    memset( pNode.puSimD, 0, sizeof(unsigned) * p.nWordsDyna );

    // assign the number and add to the array of nodes
    pNode.Num   = p.vNodes.nSize;
    Fraig_NodeVecPush( p.vNodes,  pNode );

    // assign the PI number and add to the array of primary inputs
    pNode.NumPi = p.vInputs.nSize;   
    Fraig_NodeVecPush( p.vInputs, pNode );

    pNode.Level =  0;  // PI has 0 level
    pNode.nRefs =  1;  // it is a persistent node, which comes referenced
    pNode.fInv  =  0;  // the simulation info of the PI is not complemented

    // derive the simulation info for the new node
clk = Abc_Clock();
    // set the random simulation info for the primary input
    pNode.uHashR = 0;
    for  i = 0; i < p.nWordsRand; i++  {
        // generate the simulation info
        pNode.puSimR[i] = FRAIG_RANDOM_UNSIGNED;
        // for reasons that take very long to explain, it makes sense to have (0000000...) 
        // pattern in the set (this helps if we need to return the counter-examples)
        if i == 0 {
            pNode.puSimR[i] <<= 1; }
        // compute the hash key
        pNode.uHashR ^= pNode.puSimR[i] * s_FraigPrimes[i];
    }
    // count the number of ones in the simulation vector
    pNode.nOnes = Fraig_BitStringCountOnes( pNode.puSimR, p.nWordsRand );

    // set the systematic simulation info for the primary input
    pNode.uHashD = 0;
    for  i = 0; i < p.iWordStart; i++  {
        // generate the simulation info
        pNode.puSimD[i] = FRAIG_RANDOM_UNSIGNED;
        // compute the hash key
        pNode.uHashD ^= pNode.puSimD[i] * s_FraigPrimes[i];
    }
p.timeSims += Abc_Clock() - clk;

    // insert it into the hash table
    pNodeRes = Fraig_HashTableLookupF( p, pNode );
    assert(pNodeRes == nil, "pNodeRes == nil");
    // add to the runtime of simulation
    return pNode;
}

/**Function*************************************************************

  Synopsis    [Creates a new node.]

  Description [This procedure should be called to create the constant
  node and the PI nodes first.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeCreate(p *Fraig_Man_t, p1 *Fraig_Node_t, p2 *Fraig_Node_t) *Fraig_Node_t {
    var pNode *Fraig_Node_t;
    var clk abctime;

    // create the node
    pNode = (Fraig_Node_t *)Fraig_MemFixedEntryFetch( p.mmNodes );
    memset( pNode, 0, sizeof(Fraig_Node_t) );

    // assign the children
    // inlining of macro Fraig_Ref
    // inlining of macro Fraig_Regular
    pNode.p1  = p1;  ;  ((Fraig_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) .nRefs++;
    // inlining of macro Fraig_Ref
    // inlining of macro Fraig_Regular
    pNode.p2  = p2;  ;  ((Fraig_Node_t *)((ABC_PTRUINT_T)(p2) & ~01)) .nRefs++;

    // assign the number and add to the array of nodes
    pNode.Num = p.vNodes.nSize;
    Fraig_NodeVecPush( p.vNodes,  pNode );

    // assign the PI number
    pNode.NumPi = -1;

    // compute the level of this node
    // inlining of macro Fraig_Regular
    pNode.Level = 1 + Abc_MaxInt(((Fraig_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) .Level, ((Fraig_Node_t *)((ABC_PTRUINT_T)(p2) & ~01)) .Level);
    // inlining of macro Fraig_NodeIsSimComplement
    // inlining of macro Fraig_IsComplement
    // inlining of macro Fraig_Regular
    pNode.fInv  = (((int32(((ABC_PTRUINT_T) (p1) & 01))))? !(((Fraig_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) .fInv) : (p1).fInv) & (((int32(((ABC_PTRUINT_T) (p2) & 01))))? !(((Fraig_Node_t *)((ABC_PTRUINT_T)(p2) & ~01)) .fInv) : (p2).fInv);
    // inlining of macro Fraig_Regular
    pNode.fFailTfo = ((Fraig_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) .fFailTfo | ((Fraig_Node_t *)((ABC_PTRUINT_T)(p2) & ~01)) .fFailTfo;

    // derive the simulation info 
clk = Abc_Clock();
    // allocate memory for the simulation info
    pNode.puSimR = (unsigned *)Fraig_MemFixedEntryFetch( p.mmSims );
    pNode.puSimD = pNode.puSimR + p.nWordsRand;
    // derive random simulation info
    pNode.uHashR = 0;
    Fraig_NodeSimulate( pNode, 0, p.nWordsRand, 1 );
    // derive dynamic simulation info
    pNode.uHashD = 0;
    Fraig_NodeSimulate( pNode, 0, p.iWordStart, 0 );
    // count the number of ones in the random simulation info
    pNode.nOnes = Fraig_BitStringCountOnes( pNode.puSimR, p.nWordsRand );
    if pNode.fInv {
        pNode.nOnes = p.nWordsRand * 32 - pNode.nOnes; }
    // add to the runtime of simulation
p.timeSims += Abc_Clock() - clk;

#ifdef FRAIG_ENABLE_FANOUTS
    // create the fanout info
    Fraig_NodeAddFaninFanout( Fraig_Regular(p1), pNode );
    Fraig_NodeAddFaninFanout( Fraig_Regular(p2), pNode );
#endif
    return pNode;
}


/**Function*************************************************************

  Synopsis    [Simulates the node.]

  Description [Simulates the random or dynamic simulation info through 
  the node. Uses phases of the children to determine their real simulation
  info. Uses phase of the node to determine the way its simulation info 
  is stored. The resulting info is guaranteed to be 0 for the first pattern.]
  
  SideEffects [This procedure modified the hash value of the simulation info.]

  SeeAlso     []

***********************************************************************/
func Fraig_NodeSimulate(pNode *Fraig_Node_t, iWordStart int32, iWordStop int32, fUseRand int32) {
    var pSims, pSims1, pSims2 *uint32;
    var uHash uint32;
    var fCompl, fCompl1, fCompl2, i int32;

    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");

    // get hold of the simulation information
    if fUseRand != 0 {
        pSims = pNode.puSimR;
    } else {
        pSims = pNode.puSimD;
    }
    if fUseRand != 0 {
        // inlining of macro Fraig_Regular
        pSims1 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .puSimR;
    } else {
        // inlining of macro Fraig_Regular
        pSims1 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .puSimD;
    }
    if fUseRand != 0 {
        // inlining of macro Fraig_Regular
        pSims2 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .puSimR;
    } else {
        // inlining of macro Fraig_Regular
        pSims2 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .puSimD;
    }

    // get complemented attributes of the children using their random info
    fCompl  = pNode.fInv;
    // inlining of macro Fraig_NodeIsSimComplement
    
    // inlining of macro Fraig_IsComplement
    if (int32(((ABC_PTRUINT_T) (pNode.p1) & 01)) != 0) {
        // inlining of macro Fraig_Regular
        fCompl1 = !(((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .fInv);
    } else {
        fCompl1 = (pNode.p1).fInv;
    }
    // inlining of macro Fraig_NodeIsSimComplement
    
    // inlining of macro Fraig_IsComplement
    if (int32(((ABC_PTRUINT_T) (pNode.p2) & 01)) != 0) {
        // inlining of macro Fraig_Regular
        fCompl2 = !(((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .fInv);
    } else {
        fCompl2 = (pNode.p2).fInv;
    }

    // simulate
    uHash = 0;
    if fCompl1 && fCompl2 {
        if fCompl {
            for  i = iWordStart; i < iWordStop; i++  {
                pSims[i] = (pSims1[i] | pSims2[i]);
                uHash ^= pSims[i] * s_FraigPrimes[i];
            }
        } else {
            for  i = iWordStart; i < iWordStop; i++  {
                pSims[i] = ~(pSims1[i] | pSims2[i]);
                uHash ^= pSims[i] * s_FraigPrimes[i];
            } }
    
    } else if fCompl1 && !fCompl2 {
        if fCompl {
            for  i = iWordStart; i < iWordStop; i++  {
                pSims[i] = (pSims1[i] | ~pSims2[i]);
                uHash ^= pSims[i] * s_FraigPrimes[i];
            }
        } else {
            for  i = iWordStart; i < iWordStop; i++  {
                pSims[i] = (~pSims1[i] & pSims2[i]);
                uHash ^= pSims[i] * s_FraigPrimes[i];
            } }
    
    } else if !fCompl1 && fCompl2 {
        if fCompl {
            for  i = iWordStart; i < iWordStop; i++  {
                pSims[i] = (~pSims1[i] | pSims2[i]);
                uHash ^= pSims[i] * s_FraigPrimes[i];
            }
        } else {
            for  i = iWordStart; i < iWordStop; i++  {
                pSims[i] = (pSims1[i] & ~pSims2[i]);
                uHash ^= pSims[i] * s_FraigPrimes[i];
            } }
    
    } else { // if ( !fCompl1 && !fCompl2 )
    
        if fCompl {
            for  i = iWordStart; i < iWordStop; i++  {
                pSims[i] = ~(pSims1[i] & pSims2[i]);
                uHash ^= pSims[i] * s_FraigPrimes[i];
            }
        } else {
            for  i = iWordStart; i < iWordStop; i++  {
                pSims[i] = (pSims1[i] & pSims2[i]);
                uHash ^= pSims[i] * s_FraigPrimes[i];
            } }
    }

    if fUseRand != 0 {
        pNode.uHashR ^= uHash;
    } else {
        pNode.uHashD ^= uHash; }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

