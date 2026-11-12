package fxu

/**CFile****************************************************************

  FileName    [fxuSelect.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Procedures to select the best divisor/complement pair.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: fxuSelect.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fxuInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const MAX_SIZE_LOOKAHEAD =      20

// static int Fxu_MatrixFindComplement( Fxu_Matrix * p, int iVar );


// static Fxu_Double * Fxu_MatrixFindComplementSingle( Fxu_Matrix * p, Fxu_Single * pSingle );

// static Fxu_Single * Fxu_MatrixFindComplementDouble2( Fxu_Matrix * p, Fxu_Double * pDouble );

// static Fxu_Double * Fxu_MatrixFindComplementDouble4( Fxu_Matrix * p, Fxu_Double * pDouble );


// Fxu_Double * Fxu_MatrixFindDouble( Fxu_Matrix * p, 
//      int piVarsC1[], int piVarsC2[], int nVarsC1, int nVarsC2 );

// void Fxu_MatrixGetDoubleVars( Fxu_Matrix * p, Fxu_Double * pDouble, 
//     int piVarsC1[], int piVarsC2[], int * pnVarsC1, int * pnVarsC2 );



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Selects the best pair (Single,Double) and returns their weight.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_Select(p *Fxu_Matrix, ppSingle **Fxu_Single, ppDouble **Fxu_Double) int32 {
    // the top entries
    var pSingles [MAX_SIZE_LOOKAHEAD]*Fxu_Single ;
    var pDoubles [MAX_SIZE_LOOKAHEAD]*Fxu_Double ;
    // the complements
    var pSCompl [MAX_SIZE_LOOKAHEAD]*Fxu_Double ;
    var pDComplS [MAX_SIZE_LOOKAHEAD]*Fxu_Single ;
    var pDComplD [MAX_SIZE_LOOKAHEAD]*Fxu_Double ;
    var pPair *Fxu_Pair;
    var nSingles int32;
    var nDoubles int32;
    var i int32;
    var WeightBest int32;
    var WeightCur int32;
    var iNum, fBestS int32;

    // collect the top entries from the queues
    for  nSingles = 0; nSingles < MAX_SIZE_LOOKAHEAD; nSingles++  {
        pSingles[nSingles] = Fxu_HeapSingleGetMax( p.pHeapSingle );
        if pSingles[nSingles] == nil {
            break; }
    }
    // put them back into the queue
    for  i = 0; i < nSingles; i++  {
        if pSingles[i] {
            Fxu_HeapSingleInsert( p.pHeapSingle, pSingles[i] ); }
    }
        
    // the same for doubles
    // collect the top entries from the queues
    for  nDoubles = 0; nDoubles < MAX_SIZE_LOOKAHEAD; nDoubles++  {
        pDoubles[nDoubles] = Fxu_HeapDoubleGetMax( p.pHeapDouble );
        if pDoubles[nDoubles] == nil {
            break; }
    }
    // put them back into the queue
    for  i = 0; i < nDoubles; i++  {
        if pDoubles[i] {
            Fxu_HeapDoubleInsert( p.pHeapDouble, pDoubles[i] ); }
    }

    // for each single, find the complement double (if any)
    for  i = 0; i < nSingles; i++  {
        if pSingles[i] {
            pSCompl[i] = Fxu_MatrixFindComplementSingle( p, pSingles[i] ); }
    }

    // for each double, find the complement single or double (if any)
    for  i = 0; i < nDoubles; i++  {
        if pDoubles[i] {
            pPair = pDoubles[i].lPairs.pHead;
            if pPair.nLits1 == 1 && pPair.nLits2 == 1 {
                pDComplS[i] = Fxu_MatrixFindComplementDouble2( p, pDoubles[i] );
                pDComplD[i] = nil;
            
//            else if ( pPair->nLits1 == 2 && pPair->nLits2 == 2 )
//            {
//                pDComplS[i] = NULL;
//                pDComplD[i] = Fxu_MatrixFindComplementDouble4( p, pDoubles[i] );
//            }
            } else {
                pDComplS[i] = nil;
                pDComplD[i] = nil;
            }
        }
    }

    // select the best pair
    WeightBest = -1;
    for  i = 0; i < nSingles; i++  {
        WeightCur = pSingles[i].Weight;
        if pSCompl[i] {
            // add the weight of the double
            WeightCur += pSCompl[i].Weight;
            // there is no need to implement this double, so...
            pPair      = pSCompl[i].lPairs.pHead;
            WeightCur += pPair.nLits1 + pPair.nLits2;
        }
        if WeightBest < WeightCur {
            WeightBest = WeightCur;
            *ppSingle = pSingles[i];
            *ppDouble = pSCompl[i];
            fBestS = 1;
            iNum = i;
        }
    }
    for  i = 0; i < nDoubles; i++  {
        WeightCur = pDoubles[i].Weight;
        if pDComplS[i] {
            // add the weight of the single
            WeightCur += pDComplS[i].Weight;
            // there is no need to implement this double, so...
            pPair      = pDoubles[i].lPairs.pHead;
            WeightCur += pPair.nLits1 + pPair.nLits2;
        }
        if WeightBest < WeightCur {
            WeightBest = WeightCur;
            *ppSingle = pDComplS[i];
            *ppDouble = pDoubles[i];
            fBestS = 0;
            iNum = i;
        }
    }
/*
    // print the statistics
    printf( "\n" );
    for ( i = 0; i < nSingles; i++ )
    {
        printf( "Single #%d: Weight = %3d. ", i, pSingles[i]->Weight );
        printf( "Compl: " );
        if ( pSCompl[i] == NULL )
            printf( "None." );
        else
            printf( "D  Weight = %3d  Sum = %3d", 
                pSCompl[i]->Weight, pSCompl[i]->Weight + pSingles[i]->Weight );
        printf( "\n" );
    }
    printf( "\n" );
    for ( i = 0; i < nDoubles; i++ )
    {
        printf( "Double #%d: Weight = %3d. ", i, pDoubles[i]->Weight );
        printf( "Compl: " );
        if ( pDComplS[i] == NULL && pDComplD[i] == NULL )
            printf( "None." );
        else if ( pDComplS[i] )
            printf( "S  Weight = %3d  Sum = %3d", 
                pDComplS[i]->Weight, pDComplS[i]->Weight + pDoubles[i]->Weight );
        else if ( pDComplD[i] )
            printf( "D  Weight = %3d  Sum = %3d", 
                pDComplD[i]->Weight, pDComplD[i]->Weight + pDoubles[i]->Weight );
        printf( "\n" );
    }
    if ( WeightBest == -1 )
        printf( "Selected NONE\n" );
    else
    {
        printf( "Selected = %s.  ", fBestS? "S": "D" );
        printf( "Number = %d.  ", iNum );
        printf( "Weight = %d.\n", WeightBest );
    }
    printf( "\n" );
*/
    return WeightBest;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixFindComplementSingle(p *Fxu_Matrix, pSingle *Fxu_Single) *Fxu_Double {
//    int * pValue2Node = p->pValue2Node;
    var iVar1, iVar2 int32;
    var iVar1C, iVar2C int32;
    // get the variables of this single div
    iVar1  = pSingle.pVar1.iVar;
    iVar2  = pSingle.pVar2.iVar;
    iVar1C = Fxu_MatrixFindComplement( p, iVar1 );
    iVar2C = Fxu_MatrixFindComplement( p, iVar2 );
    if iVar1C == -1 || iVar2C == -1 {
        return nil; }
    assert(iVar1C < iVar2C, "iVar1C < iVar2C");
    return Fxu_MatrixFindDouble( p, &iVar1C, &iVar2C, 1, 1 );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixFindComplementDouble2(p *Fxu_Matrix, pDouble *Fxu_Double) *Fxu_Single {
//    int * pValue2Node = p->pValue2Node;
    var piVarsC1, piVarsC2 [10]int32;
    var nVarsC1, nVarsC2 int32;
    var iVar1, iVar2, iVarTemp int32;
    var iVar1C, iVar2C int32;
    var pSingle *Fxu_Single;

    // get the variables of this double div
    Fxu_MatrixGetDoubleVars( p, pDouble, piVarsC1, piVarsC2, &nVarsC1, &nVarsC2 );
    assert(nVarsC1 == 1, "nVarsC1 == 1");
    assert(nVarsC2 == 1, "nVarsC2 == 1");
    iVar1 = piVarsC1[0];
    iVar2 = piVarsC2[0];
    assert(iVar1 < iVar2, "iVar1 < iVar2");

    iVar1C = Fxu_MatrixFindComplement( p, iVar1 );
    iVar2C = Fxu_MatrixFindComplement( p, iVar2 );
    if iVar1C == -1 || iVar2C == -1 {
        return nil; }
 
    // go through the queque and find this one
//    assert( iVar1C < iVar2C );
    if iVar1C > iVar2C {
        iVarTemp = iVar1C;
        iVar1C = iVar2C;
        iVar2C = iVarTemp;
    }

    // inlining of macro Fxu_MatrixForEachSingle
    for  pSingle = (p).lSingles.pHead;          (pSingle != nil);          pSingle = pSingle.pNext  {
        if pSingle.pVar1.iVar == iVar1C && pSingle.pVar2.iVar == iVar2C {
            return pSingle; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixFindComplementDouble4(p *Fxu_Matrix, pDouble *Fxu_Double) *Fxu_Double {
//    int * pValue2Node = p->pValue2Node;
    var piVarsC1, piVarsC2 [10]int32;
    var nVarsC1, nVarsC2 int32;
    var iVar11, iVar12, iVar21, iVar22 int32;
    var iVar11C, iVar12C, iVar21C, iVar22C int32;
    var RetValue int32;

    // get the variables of this double div
    Fxu_MatrixGetDoubleVars( p, pDouble, piVarsC1, piVarsC2, &nVarsC1, &nVarsC2 );
    assert(nVarsC1 == 2 && nVarsC2 == 2, "nVarsC1 == 2 && nVarsC2 == 2");

    iVar11 = piVarsC1[0];
    iVar12 = piVarsC1[1];
    iVar21 = piVarsC2[0];
    iVar22 = piVarsC2[1];
    assert(iVar11 < iVar21, "iVar11 < iVar21");

    iVar11C = Fxu_MatrixFindComplement( p, iVar11 );
    iVar12C = Fxu_MatrixFindComplement( p, iVar12 );
    iVar21C = Fxu_MatrixFindComplement( p, iVar21 );
    iVar22C = Fxu_MatrixFindComplement( p, iVar22 );
    if iVar11C == -1 || iVar12C == -1 || iVar21C == -1 || iVar22C == -1 {
        return nil; }
    if iVar11C != iVar21 || iVar12C != iVar22 || 
         iVar21C != iVar11 || iVar22C != iVar12 {
         return nil; }

    // a'b' + ab   =>  a'b  + ab'
    // a'b  + ab'  =>  a'b' + ab
    // swap the second pair in each cube
    RetValue    = piVarsC1[1];
    piVarsC1[1] = piVarsC2[1];
    piVarsC2[1] = RetValue;

    return Fxu_MatrixFindDouble( p, piVarsC1, piVarsC2, 2, 2 );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixFindComplement(p *Fxu_Matrix, iVar int32) int32 {
    return iVar ^ 1;
/*
//    int * pValue2Node = p->pValue2Node;
    int iVarC;
    int iNode;
    int Beg, End;

    // get the nodes
    iNode = pValue2Node[iVar];
    // get the first node with the same var
    for ( Beg = iVar; Beg >= 0; Beg-- )
        if ( pValue2Node[Beg] != iNode )
        {
            Beg++;
            break;
        }
    // get the last node with the same var
    for ( End = iVar;          ; End++ )
        if ( pValue2Node[End] != iNode )
        {
            End--;
            break;
        }

    // if one of the vars is not binary, quit
    if ( End - Beg > 1 )
        return -1;

    // get the complements
    if ( iVar == Beg )
        iVarC = End;
    else if ( iVar == End ) 
        iVarC = Beg;
    else
    {
        assert( 0 );
    }
    return iVarC;
*/
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixGetDoubleVars(p *Fxu_Matrix, pDouble *Fxu_Double, piVarsC1 []int32, piVarsC2 []int32, pnVarsC1 *int32, pnVarsC2 *int32) {
    var pPair *Fxu_Pair;
    var pLit1, pLit2 *Fxu_Lit;
    var nLits1, nLits2 int32;

    // get the first pair
    pPair = pDouble.lPairs.pHead;
    // init the parameters
    nLits1 = 0;
    nLits2 = 0;
    pLit1 = pPair.pCube1.lLits.pHead;
    pLit2 = pPair.pCube2.lLits.pHead;
    for {
        if pLit1 && (pLit2 != nil) {
            if pLit1.iVar == pLit2.iVar { // ensure cube-free
                pLit1 = pLit1.pHNext;
                pLit2 = pLit2.pHNext;
            
            } else if pLit1.iVar < pLit2.iVar {
                piVarsC1[nLits1++] = pLit1.iVar;
                 pLit1 = pLit1.pHNext;
           
            } else {
                piVarsC2[nLits2++] = pLit2.iVar;
                pLit2 = pLit2.pHNext;
            }
        
        } else if pLit1 && !(pLit2 != nil) {
            piVarsC1[nLits1++] = pLit1.iVar;
            pLit1 = pLit1.pHNext;
        
        } else if !pLit1 && (pLit2 != nil) {
            piVarsC2[nLits2++] = pLit2.iVar;
            pLit2 = pLit2.pHNext;
        
        } else {
            break; }
    }
    *pnVarsC1 = nLits1;
    *pnVarsC2 = nLits2;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixFindDouble(p *Fxu_Matrix, piVarsC1 []int32, piVarsC2 []int32, nVarsC1 int32, nVarsC2 int32) *Fxu_Double {
    var piVarsC1_, piVarsC2_ [100]int32;
    var nVarsC1_, nVarsC2_, i int32;
    var pDouble *Fxu_Double;
    var pPair *Fxu_Pair;
    var Key uint32;

    assert(nVarsC1 > 0, "nVarsC1 > 0");
    assert(nVarsC2 > 0, "nVarsC2 > 0");
    assert(piVarsC1[0] < piVarsC2[0], "piVarsC1[0] < piVarsC2[0]");

    // get the hash key
    Key = Fxu_PairHashKeyArray( p, piVarsC1, piVarsC2, nVarsC1, nVarsC2 );
    
    // check if the divisor for this pair already exists
    Key %= p.nTableSize;
    // inlining of macro Fxu_TableForEachDouble
    for  pDouble = (p).pTable[Key].pHead;          (pDouble != nil);          pDouble = pDouble.pNext  {
        pPair = pDouble.lPairs.pHead;
        if pPair.nLits1 != nVarsC1 {
            continue; }
        if pPair.nLits2 != nVarsC2 {
            continue; }
        // get the cubes of this divisor
        Fxu_MatrixGetDoubleVars( p, pDouble, piVarsC1_, piVarsC2_, &nVarsC1_, &nVarsC2_ );
        // compare lits of the first cube
        for  i = 0; i < nVarsC1; i++  {
            if piVarsC1[i] != piVarsC1_[i] {
                break; }
        }
        if i != nVarsC1 {
            continue; }
        // compare lits of the second cube
        for  i = 0; i < nVarsC2; i++  {
            if piVarsC2[i] != piVarsC2_[i] {
                break; }
        }
        if i != nVarsC2 {
            continue; }
        return pDouble;
    }
    return nil;
}





/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_SelectSCD(p *Fxu_Matrix, WeightLimit int32, ppVar1 **Fxu_Var, ppVar2 **Fxu_Var) int32 {
//    int * pValue2Node = p->pValue2Node;
    var pVar1 *Fxu_Var;
    var pVar2, pVarTemp *Fxu_Var;
    var pLitV, pLitH *Fxu_Lit;
    var Coin int32;
    var CounterAll int32;
    var CounterTest int32;
    var WeightCur int32;
    var WeightBest int32;

    CounterAll = 0;
    CounterTest = 0;

    WeightBest = -10;
    
    // iterate through the columns in the matrix
    // inlining of macro Fxu_MatrixForEachVariable
    for  pVar1 = (p).lVars.pHead;          (pVar1 != nil);          pVar1 = pVar1.pNext  {
        // start collecting the affected vars
        // inlining of macro Fxu_MatrixRingVarsStart
        ((p).ppTailVars  = &((p).pOrderVars));
((p).pOrderVars  = nil);;

        // go through all the literals of this variable
        for  pLitV = pVar1.lLits.pHead; pLitV; pLitV = pLitV.pVNext  {
            // for this literal, go through all the horizontal literals
            for  pLitH = pLitV.pHNext; (pLitH != nil); pLitH = pLitH.pHNext  {
                // get another variable
                pVar2 = pLitH.pVar;
                CounterAll++;
                // skip the var if it is already used
                if pVar2.pOrder {
                    continue; }
                // skip the var if it belongs to the same node
//                if ( pValue2Node[pVar1->iVar] == pValue2Node[pVar2->iVar] )
//                    continue;
                // collect the var
                // inlining of macro Fxu_MatrixRingVarsAdd
                (*((p).ppTailVars ) = pVar2 );
((p).ppTailVars  = &(pVar2).pOrder );
((pVar2).pOrder  = (Fxu_Var *)1);;
            }
        }
        // stop collecting the selected vars
        // inlining of macro Fxu_MatrixRingVarsStop
        ;

        // iterate through the selected vars
        // inlining of macro Fxu_MatrixForEachVarInRing
        if (p).pOrderVars {    for  pVar2 = (p).pOrderVars;          pVar2 != (Fxu_Var *)1;          pVar2 = pVar2.pOrder  {
            CounterTest++;

            // count the coincidence
            Coin = Fxu_SingleCountCoincidence( p, pVar1, pVar2 );
            assert(Coin > 0, "Coin > 0");

            // get the new weight
            WeightCur = Coin - 2;

            // compare the weights
            if WeightBest < WeightCur {
                WeightBest = WeightCur;
                *ppVar1 = pVar1;
                *ppVar2 = pVar2;
            }
        } }
        // unmark the vars
        // inlining of macro Fxu_MatrixForEachVarInRingSafe
        if (p).pOrderVars {    for  pVar2 = (p).pOrderVars, pVarTemp = ((pVar2 != (Fxu_Var *)1)? pVar2.pOrder: (Fxu_Var *)1);          pVar2 != (Fxu_Var *)1;          pVar2 = pVarTemp, pVarTemp = ((pVar2 != (Fxu_Var *)1)? pVar2.pOrder: (Fxu_Var *)1)  {
            pVar2.pOrder = nil;
        } }
        // inlining of macro Fxu_MatrixRingVarsReset
        ((p).pOrderVars  = nil);
((p).ppTailVars  = nil);;
    }

//    if ( WeightBest == WeightLimit )
//        return -1;
    return WeightBest;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

