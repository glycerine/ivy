package ivy

import "fmt"

/**CFile****************************************************************

  FileName    [ivyMulti.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [Constructing multi-input AND/EXOR gates.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyMulti.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Ivy_Eval_t struct {
    Mask uint32; /* C bitfield: 5 bits; review packing and truncation */  // the mask of covered nodes
    Weight uint32; /* C bitfield: 3 bits; review packing and truncation */  // the number of covered nodes
    Cost uint32; /* C bitfield: 4 bits; review packing and truncation */  // the number of overlapping nodes
    Level uint32; /* C bitfield: 12 bits; review packing and truncation */  // the level of this node
    Fan0 uint32; /* C bitfield: 4 bits; review packing and truncation */  // the first fanin
    Fan1 uint32; /* C bitfield: 4 bits; review packing and truncation */  // the second fanin
}

// static Ivy_Obj_t * Ivy_MultiBuild_rec( Ivy_Eval_t * pEvals, int iNum, Ivy_Obj_t ** pArgs, int nArgs, Ivy_Type_t Type );

// static void        Ivy_MultiSort( Ivy_Obj_t ** pArgs, int nArgs );

// static int         Ivy_MultiPushUniqueOrderByLevel( Ivy_Obj_t ** pArray, int nArgs, Ivy_Obj_t * pNode );

// static Ivy_Obj_t * Ivy_MultiEval( Ivy_Obj_t ** pArgs, int nArgs, Ivy_Type_t Type );



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Constructs the well-balanced tree of gates.]

  Description [Disregards levels and possible logic sharing.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_Multi_rec(ppObjs **Ivy_Obj_t, nObjs int32, Type Ivy_Type_t) *Ivy_Obj_t {
    var pObj1, pObj2 *Ivy_Obj_t;
    if nObjs == 1 {
        return ppObjs[0]; }
    pObj1 = Ivy_Multi_rec( ppObjs,           nObjs/2,         Type );
    pObj2 = Ivy_Multi_rec( ppObjs + nObjs/2, nObjs - nObjs/2, Type );
    return Ivy_Oper( pObj1, pObj2, Type );
}

/**Function*************************************************************

  Synopsis    [Constructs a balanced tree while taking sharing into account.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ivy_Multi_NumBits [32]byte = [32]byte{0,1,1,2,1,2,2,3,1,2,2,3,2,3,3,4,1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5};
var spStatic_Ivy_Multi_pEvals [15+15*14/2]Ivy_Eval_t;
var spStatic_Ivy_Multi_pArgs [16]*Ivy_Obj_t;
func Ivy_Multi(pArgsInit **Ivy_Obj_t, nArgs int32, Type Ivy_Type_t) *Ivy_Obj_t {
    
    
    
    var pEva, pEvaBest *Ivy_Eval_t;
    var nArgsNew, nEvals, i, k int32;
    var pTemp *Ivy_Obj_t;

    // consider the case of one argument
    assert(nArgs > 0, "nArgs > 0");
    if nArgs == 1 {
        return pArgsInit[0]; }
    // consider the case of two arguments
    if nArgs == 2 {
        return Ivy_Oper( pArgsInit[0], pArgsInit[1], Type ); }

//Ivy_MultiEval( pArgsInit, nArgs, Type ); printf( "\n" );

    // set the initial ones
    for  i = 0; i < nArgs; i++  {
        spStatic_Ivy_Multi_pArgs[i] = pArgsInit[i];
        pEva = spStatic_Ivy_Multi_pEvals + i;
        pEva.Mask   = (1 << i);
        pEva.Weight = 1;
        pEva.Cost   = 0; 
        pEva.Level  = Ivy_Regular(spStatic_Ivy_Multi_pArgs[i]).Level; 
        pEva.Fan0   = 0; 
        pEva.Fan1   = 0; 
    }

    // find the available nodes
    pEvaBest = spStatic_Ivy_Multi_pEvals;
    nArgsNew = nArgs;
    for  i = 1; i < nArgsNew; i++  {
    for  k = 0; k < i; k++  {
        if pTemp = Ivy_TableLookup(Ivy_ObjCreateGhost(spStatic_Ivy_Multi_pArgs[k], spStatic_Ivy_Multi_pArgs[i], Type, IVY_INIT_NONE)) {
            pEva = spStatic_Ivy_Multi_pEvals + nArgsNew;
            pEva.Mask   = spStatic_Ivy_Multi_pEvals[k].Mask | spStatic_Ivy_Multi_pEvals[i].Mask;
            pEva.Weight = spStatic_Ivy_Multi_NumBits[pEva.Mask];
            pEva.Cost   = spStatic_Ivy_Multi_pEvals[k].Cost + spStatic_Ivy_Multi_pEvals[i].Cost + spStatic_Ivy_Multi_NumBits[spStatic_Ivy_Multi_pEvals[k].Mask & spStatic_Ivy_Multi_pEvals[i].Mask]; 
            // inlining of macro IVY_MAX
            pEva.Level  = 1 + (((spStatic_Ivy_Multi_pEvals[k].Level) > (spStatic_Ivy_Multi_pEvals[i].Level))? (spStatic_Ivy_Multi_pEvals[k].Level) : (spStatic_Ivy_Multi_pEvals[i].Level)); 
            pEva.Fan0   = k; 
            pEva.Fan1   = i; 
//            assert( pEva->Level == (unsigned)Ivy_ObjLevel(pTemp) );
            // compare
            if pEvaBest.Weight < pEva.Weight ||
                 pEvaBest.Weight == pEva.Weight && pEvaBest.Cost > pEva.Cost ||
                 pEvaBest.Weight == pEva.Weight && pEvaBest.Cost == pEva.Cost && pEvaBest.Level > pEva.Level {
                 pEvaBest = pEva; }
            // save the argument
            spStatic_Ivy_Multi_pArgs[nArgsNew++] = pTemp;
            if nArgsNew == 15 {
                goto Outside; }
        }
    }
    }
Outside:

//    printf( "Best = %d.\n", pEvaBest - pEvals );

    // the case of no common nodes
    if nArgsNew == nArgs {
        Ivy_MultiSort( spStatic_Ivy_Multi_pArgs, nArgs );
        return Ivy_MultiBalance_rec( spStatic_Ivy_Multi_pArgs, nArgs, Type );
    }
    // the case of one common node
    if nArgsNew == nArgs + 1 {
        assert(pEvaBest - spStatic_Ivy_Multi_pEvals == nArgs, "pEvaBest - spStatic_Ivy_Multi_pEvals == nArgs");
        k = 0;
        for  i = 0; i < nArgs; i++  {
            if i != int32(pEvaBest.Fan0) && i != int32(pEvaBest.Fan1) {
                spStatic_Ivy_Multi_pArgs[k++] = spStatic_Ivy_Multi_pArgs[i]; }
        }
        spStatic_Ivy_Multi_pArgs[k++] = spStatic_Ivy_Multi_pArgs[nArgs];
        assert(k == nArgs - 1, "k == nArgs - 1");
        nArgs = k;
        Ivy_MultiSort( spStatic_Ivy_Multi_pArgs, nArgs );
        return Ivy_MultiBalance_rec( spStatic_Ivy_Multi_pArgs, nArgs, Type );
    }
    // the case when there is a node that covers everything
    if int32(pEvaBest.Mask) == ((1 << nArgs) - 1) {
        return Ivy_MultiBuild_rec( spStatic_Ivy_Multi_pEvals, pEvaBest - spStatic_Ivy_Multi_pEvals, spStatic_Ivy_Multi_pArgs, nArgsNew, Type ); } 

    // evaluate node pairs
    nEvals = nArgsNew;
    for  i = 1; i < nArgsNew; i++  {
    for  k = 0; k < i; k++  {
        pEva = spStatic_Ivy_Multi_pEvals + nEvals;
        pEva.Mask   = spStatic_Ivy_Multi_pEvals[k].Mask | spStatic_Ivy_Multi_pEvals[i].Mask;
        pEva.Weight = spStatic_Ivy_Multi_NumBits[pEva.Mask];
        pEva.Cost   = spStatic_Ivy_Multi_pEvals[k].Cost + spStatic_Ivy_Multi_pEvals[i].Cost + spStatic_Ivy_Multi_NumBits[spStatic_Ivy_Multi_pEvals[k].Mask & spStatic_Ivy_Multi_pEvals[i].Mask]; 
        // inlining of macro IVY_MAX
        pEva.Level  = 1 + (((spStatic_Ivy_Multi_pEvals[k].Level) > (spStatic_Ivy_Multi_pEvals[i].Level))? (spStatic_Ivy_Multi_pEvals[k].Level) : (spStatic_Ivy_Multi_pEvals[i].Level)); 
        pEva.Fan0   = k; 
        pEva.Fan1   = i; 
        // compare
        if pEvaBest.Weight < pEva.Weight ||
             pEvaBest.Weight == pEva.Weight && pEvaBest.Cost > pEva.Cost ||
             pEvaBest.Weight == pEva.Weight && pEvaBest.Cost == pEva.Cost && pEvaBest.Level > pEva.Level {
             pEvaBest = pEva; }
        // save the argument
        nEvals++;
    }
    }
    assert(pEvaBest - spStatic_Ivy_Multi_pEvals >= nArgsNew, "pEvaBest - spStatic_Ivy_Multi_pEvals >= nArgsNew");

//    printf( "Used (%d, %d).\n", pEvaBest->Fan0, pEvaBest->Fan1 );

    // get the best implementation
    pTemp = Ivy_MultiBuild_rec( spStatic_Ivy_Multi_pEvals, pEvaBest - spStatic_Ivy_Multi_pEvals, spStatic_Ivy_Multi_pArgs, nArgsNew, Type );

    // collect those not covered by EvaBest
    k = 0;
    for  i = 0; i < nArgs; i++  {
        if (pEvaBest.Mask & (1 << i)) == 0 {
            spStatic_Ivy_Multi_pArgs[k++] = spStatic_Ivy_Multi_pArgs[i]; }
    }
    spStatic_Ivy_Multi_pArgs[k++] = pTemp;
    assert(k == nArgs - int32(pEvaBest.Weight) + 1, "k == nArgs - int32(pEvaBest.Weight) + 1");
    nArgs = k;
    Ivy_MultiSort( spStatic_Ivy_Multi_pArgs, nArgs );
    return Ivy_MultiBalance_rec( spStatic_Ivy_Multi_pArgs, nArgs, Type );
}

/**Function*************************************************************

  Synopsis    [Implements multi-input AND/EXOR operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_MultiBuild_rec(pEvals *Ivy_Eval_t, iNum int32, pArgs **Ivy_Obj_t, nArgs int32, Type Ivy_Type_t) *Ivy_Obj_t {
    var pNode0, pNode1 *Ivy_Obj_t;
    if iNum < nArgs {
        return pArgs[iNum]; }
    pNode0 = Ivy_MultiBuild_rec( pEvals, pEvals[iNum].Fan0, pArgs, nArgs, Type );
    pNode1 = Ivy_MultiBuild_rec( pEvals, pEvals[iNum].Fan1, pArgs, nArgs, Type );
    return Ivy_Oper( pNode0, pNode1, Type );
}

/**Function*************************************************************

  Synopsis    [Selection-sorts the nodes in the decreasing over of level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_MultiSort(pArgs **Ivy_Obj_t, nArgs int32) {
    var pTemp *Ivy_Obj_t;
    var i, j, iBest int32;

    for  i = 0; i < nArgs-1; i++  {
        iBest = i;
        for  j = i+1; j < nArgs; j++  {
            if Ivy_Regular(pArgs[j]).Level > Ivy_Regular(pArgs[iBest]).Level {
                iBest = j; }
        }
        pTemp = pArgs[i]; 
        pArgs[i] = pArgs[iBest]; 
        pArgs[iBest] = pTemp;
    }
}

/**Function*************************************************************

  Synopsis    [Inserts a new node in the order by levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_MultiPushUniqueOrderByLevel(pArray **Ivy_Obj_t, nArgs int32, pNode *Ivy_Obj_t) int32 {
    var pNode1, pNode2 *Ivy_Obj_t;
    var i int32;
    // try to find the node in the array
    for  i = 0; i < nArgs; i++  {
        if pArray[i] == pNode {
            return nArgs; }
    }
    // put the node last
    pArray[nArgs++] = pNode;
    // find the place to put the new node
    for  i = nArgs-1; i > 0; i--  {
        pNode1 = pArray[i  ];
        pNode2 = pArray[i-1];
        if Ivy_Regular(pNode1).Level <= Ivy_Regular(pNode2).Level {
            break; }
        pArray[i  ] = pNode2;
        pArray[i-1] = pNode1;
    }
    return nArgs;
}

/**Function*************************************************************

  Synopsis    [Balances the array recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_MultiBalance_rec(pArgs **Ivy_Obj_t, nArgs int32, Type Ivy_Type_t) *Ivy_Obj_t {
    var pNodeNew *Ivy_Obj_t;
    // consider the case of one argument
    assert(nArgs > 0, "nArgs > 0");
    if nArgs == 1 {
        return pArgs[0]; }
    // consider the case of two arguments
    if nArgs == 2 {
        return Ivy_Oper( pArgs[0], pArgs[1], Type ); }
    // get the last two nodes
    pNodeNew = Ivy_Oper( pArgs[nArgs-1], pArgs[nArgs-2], Type );
    // add the new node
    nArgs = Ivy_MultiPushUniqueOrderByLevel( pArgs, nArgs - 2, pNodeNew );
    return Ivy_MultiBalance_rec( pArgs, nArgs, Type );
}

/**Function*************************************************************

  Synopsis    [Implements multi-input AND/EXOR operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_MultiEval(pArgs **Ivy_Obj_t, nArgs int32, Type Ivy_Type_t) *Ivy_Obj_t {
    var pTemp *Ivy_Obj_t;
    var i, k int32;
    var nArgsOld int32  = nArgs;
    for  i = 0; i < nArgs; i++  {
        fmt.Printf("%d[%d] ", i, Ivy_Regular(pArgs[i]).Level);
    }
    for  i = 1; i < nArgs; i++  {
        for  k = 0; k < i; k++  {
            pTemp = Ivy_TableLookup(Ivy_ObjCreateGhost(pArgs[k], pArgs[i], Type, IVY_INIT_NONE));
            if pTemp != nil {
                fmt.Printf("%d[%d]=(%d,%d) ", nArgs, Ivy_Regular(pTemp).Level, k, i);
                pArgs[nArgs++] = pTemp;
            }
        }
    }
    fmt.Printf("     ((%d/%d))    ", nArgsOld, nArgs-nArgsOld);
    return nil;
}



/**Function*************************************************************

  Synopsis    [Old code.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_Multi1(pArgs **Ivy_Obj_t, nArgs int32, Type Ivy_Type_t) *Ivy_Obj_t {
    var pArgsRef [5]*Ivy_Obj_t;
    var pTemp *Ivy_Obj_t;
    var i, k, m, nArgsNew, Counter int32;
 
    
//Ivy_MultiEval( pArgs, nArgs, Type );  printf( "\n" );


    assert(Type == IVY_AND || Type == IVY_EXOR, "Type == IVY_AND || Type == IVY_EXOR");
    assert(nArgs > 0, "nArgs > 0");
    if nArgs == 1 {
        return pArgs[0]; }

    // find the nodes with more than one fanout
    nArgsNew = 0;
    for  i = 0; i < nArgs; i++  {
        if Ivy_ObjRefs( Ivy_Regular(pArgs[i]) ) > 0 {
            pArgsRef[nArgsNew++] = pArgs[i]; }
    }

    // go through pairs
    if nArgsNew >= 2 {
    for  i = 0; i < nArgsNew; i++  {
    for  k = i + 1; k < nArgsNew; k++  {
        if pTemp = Ivy_TableLookup(Ivy_ObjCreateGhost(pArgsRef[i], pArgsRef[k], Type, IVY_INIT_NONE)) {
            Counter++; }
    }
    } }
//    printf( "%d", Counter );
            
    // go through pairs
    if nArgsNew >= 2 {
    for  i = 0; i < nArgsNew; i++  {
    for  k = i + 1; k < nArgsNew; k++  {
        if pTemp = Ivy_TableLookup(Ivy_ObjCreateGhost(pArgsRef[i], pArgsRef[k], Type, IVY_INIT_NONE)) {
            nArgsNew = 0;
            for  m = 0; m < nArgs; m++  {
                if pArgs[m] != pArgsRef[i] && pArgs[m] != pArgsRef[k] {
                    pArgs[nArgsNew++] = pArgs[m]; }
            }
            pArgs[nArgsNew++] = pTemp;
            assert(nArgsNew == nArgs - 1, "nArgsNew == nArgs - 1");
            return Ivy_Multi1( pArgs, nArgsNew, Type );
        }
    }
    } }
    return Ivy_Multi_rec( pArgs, nArgs, Type );
}

/**Function*************************************************************

  Synopsis    [Old code.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_Multi2(pArgs **Ivy_Obj_t, nArgs int32, Type Ivy_Type_t) *Ivy_Obj_t {
    assert(Type == IVY_AND || Type == IVY_EXOR, "Type == IVY_AND || Type == IVY_EXOR");
    assert(nArgs > 0, "nArgs > 0");
    return Ivy_Multi_rec( pArgs, nArgs, Type );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

