package fxu

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [fxuCreate.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Create matrix from covers and covers from matrix.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: fxuCreate.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fxuInt.h"
//#include "fxu.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void        Fxu_CreateMatrixAddCube( Fxu_Matrix * p, Fxu_Cube * pCube, char * pSopCube, Vec_Int_t * vFanins, int * pOrder );

// static int         Fxu_CreateMatrixLitCompare( int * ptrX, int * ptrY );

// static void        Fxu_CreateCoversNode( Fxu_Matrix * p, Fxu_Data_t * pData, int iNode, Fxu_Cube * pCubeFirst, Fxu_Cube * pCubeNext );

// static Fxu_Cube *  Fxu_CreateCoversFirstCube( Fxu_Matrix * p, Fxu_Data_t * pData, int iNode );

var s_pLits *int32;

// extern int         Fxu_PreprocessCubePairs( Fxu_Matrix * p, Vec_Ptr_t * vCovers, int nPairsTotal, int nPairsMax );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates the sparse matrix from the array of SOPs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_CreateMatrix(pData *Fxu_Data_t) *Fxu_Matrix {
    var p *Fxu_Matrix;
    var pVar *Fxu_Var;
    var pCubeFirst, pCubeNew *Fxu_Cube;
    var pCube1, pCube2 *Fxu_Cube;
    var vFanins *Vec_Int_t;
    var pSopCover string;
    var pSopCube string;
    var pOrder *int32;
    var nBitsMax int32;
    var i, v, c int32;
    var nCubesTotal int32;
    var nPairsTotal int32;
    var nPairsStore int32;
    var nCubes int32;
    var iCube, iPair int32;
    var nFanins int32;

    // collect all sorts of statistics
    nCubesTotal =  0;
    nPairsTotal =  0;
    nPairsStore =  0;
    nBitsMax    = -1; 
    for  i = 0; i < pData.nNodesOld; i++  {
        if pSopCover = (char *)pData.vSops.pArray[i] {
            nCubes       = Abc_SopGetCubeNum( pSopCover );
            nFanins      = Abc_SopGetVarNum( pSopCover );
            assert(nFanins > 1 && nCubes > 0, "nFanins > 1 && nCubes > 0");

            nCubesTotal += nCubes;
            nPairsTotal += nCubes * (nCubes - 1) / 2;
            nPairsStore += nCubes * nCubes;
            if nBitsMax < nFanins {
                nBitsMax = nFanins; }
        }
    }
    if nBitsMax <= 0 {
        fmt.Printf("The current network does not have SOPs to perform extraction.\n");
        return nil;
    }

    if nPairsStore > 50000000 {
        fmt.Printf("The problem is too large to be solved by \"fxu\" (%d cubes and %d cube pairs)\n", nCubesTotal, nPairsStore);
        return nil;
    }

    // start the matrix
    p = Fxu_MatrixAllocate();
    // create the column labels 
    // inlining of macro ABC_ALLOC
    p.ppVars = ((Fxu_Var * *) malloc(sizeof(Fxu_Var *) * (size_t)(2 * (pData.nNodesOld + pData.nNodesExt))));
    for  i = 0; i < 2 * pData.nNodesOld; i++  {
        p.ppVars[i] = Fxu_MatrixAddVar( p );
    }

    // allocate storage for all cube pairs at once
    // inlining of macro ABC_ALLOC
    p.pppPairs = ((Fxu_Pair ** *) malloc(sizeof(Fxu_Pair **) * (size_t)(nCubesTotal + 100)));
    // inlining of macro ABC_ALLOC
    p.ppPairs  = ((Fxu_Pair * *) malloc(sizeof(Fxu_Pair *) * (size_t)(nPairsStore + 100)));
    memset( p.ppPairs, 0, sizeof(Fxu_Pair *) * nPairsStore );
    iCube = 0;
    iPair = 0;
    for  i = 0; i < pData.nNodesOld; i++  {
        if pSopCover = (char *)pData.vSops.pArray[i] {
            // get the number of cubes
            nCubes = Abc_SopGetCubeNum( pSopCover );
            // get the new var in the matrix
            pVar = p.ppVars[2*i+1];
            // assign the pair storage
            pVar.nCubes     = nCubes;
            if nCubes > 0 {
                pVar.ppPairs    = p.pppPairs + iCube;
                pVar.ppPairs[0] = p.ppPairs  + iPair;
                for  v = 1; v < nCubes; v++  {
                    pVar.ppPairs[v] = pVar.ppPairs[v-1] + nCubes;
                }
            }
            // update
            iCube += nCubes;
            iPair += nCubes * nCubes;
        }
    }
    assert(iCube == nCubesTotal, "iCube == nCubesTotal");
    assert(iPair == nPairsStore, "iPair == nPairsStore");


    // allocate room for the reordered literals
    // inlining of macro ABC_ALLOC
    pOrder = ((int *) malloc(sizeof(int) * (size_t)(nBitsMax)));
    // create the rows
    for  i = 0; i < pData.nNodesOld; i++  {
    if pSopCover = (char *)pData.vSops.pArray[i] {
        // get the new var in the matrix
        pVar = p.ppVars[2*i+1];
        // here we sort the literals of the cover
        // in the increasing order of the numbers of the corresponding nodes
        // because literals should be added to the matrix in this order
        vFanins = (Vec_Int_t *)pData.vFanins.pArray[i];
        s_pLits = vFanins.pArray;
        // start the variable order
        nFanins = Abc_SopGetVarNum( pSopCover );
        for  v = 0; v < nFanins; v++  {
            pOrder[v] = v;
        }
        // reorder the fanins
        qsort( (void *)pOrder, (size_t)nFanins, sizeof(int),(int (*)(void *, void *))Fxu_CreateMatrixLitCompare);
        assert(s_pLits[ pOrder[0] ] < s_pLits[ pOrder[nFanins-1] ], "s_pLits[ pOrder[0] ] < s_pLits[ pOrder[nFanins-1] ]");
        // create the corresponding cubes in the matrix
        pCubeFirst = nil;
        c = 0;
        // inlining of macro Abc_SopForEachCube
        for  pSopCube = (pSopCover); *pSopCube; pSopCube += (nFanins) + 3  {
            // create the cube
            pCubeNew = Fxu_MatrixAddCube( p, pVar, c++ );
            Fxu_CreateMatrixAddCube( p, pCubeNew, pSopCube, vFanins, pOrder );
            if pCubeFirst == nil {
                pCubeFirst = pCubeNew; }
            pCubeNew.pFirst = pCubeFirst;
        }
        // set the first cube of this var
        pVar.pFirst = pCubeFirst;
        // create the divisors without preprocessing
        if nPairsTotal <= pData.nPairsMax {
            for  pCube1 = pCubeFirst; pCube1; pCube1 = pCube1.pNext  {
                for  pCube2 = pCube1? pCube1.pNext: nil; (pCube2 != nil); pCube2 = pCube2.pNext  {
                    Fxu_MatrixAddDivisor( p, pCube1, pCube2 );
                }
            }
        }
    }
    }
    // inlining of macro ABC_FREE
    ((pOrder) ? (free((char *) (pOrder)), (pOrder) = 0) : 0);

    // consider the case when cube pairs should be preprocessed
    // before adding them to the set of divisors
//    if ( pData->fVerbose )
//        printf( "The total number of cube pairs is %d.\n", nPairsTotal );
    if nPairsTotal > 10000000 {
        fmt.Printf("The total number of cube pairs of the network is more than 10,000,000.\n");
        fmt.Printf("Command \"fx\" takes a long time to run in such cases. It is suggested\n");
        fmt.Printf("that the user changes the network by reducing the size of logic node and\n");
        fmt.Printf("consequently the number of cube pairs to be processed by this command.\n");
        fmt.Printf("It can be achieved as follows: \"st; if -K <num>\" or \"st; renode -s -K <num>\"\n");
        fmt.Printf("as a proprocessing step, while selecting <num> as approapriate.\n");
        return nil;
    }
    if nPairsTotal > pData.nPairsMax {
        if !Fxu_PreprocessCubePairs( p, pData.vSops, nPairsTotal, pData.nPairsMax ) {
            return nil; } }
//    if ( pData->fVerbose )
//        printf( "Only %d best cube pairs will be used by the fast extract command.\n", pData->nPairsMax );

    if p.lVars.nItems > 1000000 {
        fmt.Printf("The total number of variables is more than 1,000,000.\n");
        fmt.Printf("Command \"fx\" takes a long time to run in such cases. It is suggested\n");
        fmt.Printf("that the user changes the network by reducing the size of logic node and\n");
        fmt.Printf("consequently the number of cube pairs to be processed by this command.\n");
        fmt.Printf("It can be achieved as follows: \"st; if -K <num>\" or \"st; renode -s -K <num>\"\n");
        fmt.Printf("as a proprocessing step, while selecting <num> as approapriate.\n");
        return nil;
    }


    // add the var pairs to the heap
    Fxu_MatrixComputeSingles( p, pData.fUse0, pData.nSingleMax );

    // print stats
    if pData.fVerbose {
        var Density float64;
        Density = (float64(p.nEntries)) / p.lVars.nItems / p.lCubes.nItems;
        fmt.Fprintf(os.Stdout, "Matrix: [vars x cubes] = [%d x %d]  ", p.lVars.nItems, p.lCubes.nItems);
        fmt.Fprintf(os.Stdout, "Lits = %d  Density = %.5f%%\n", p.nEntries, Density);
        fmt.Fprintf(os.Stdout, "1-cube divs = %6d. (Total = %6d)  ", p.lSingles.nItems, p.nSingleTotal);
        fmt.Fprintf(os.Stdout, "2-cube divs = %6d. (Total = %6d)", p.nDivsTotal, nPairsTotal);
        fmt.Fprintf(os.Stdout, "\n");
    }
//    Fxu_MatrixPrint( stdout, p );
    return p;
}

/**Function*************************************************************

  Synopsis    [Adds one cube with literals to the matrix.]

  Description [Create the cube and literals in the matrix corresponding 
  to the given cube in the SOP cover. Co-singleton transform is performed here.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_CreateMatrixAddCube(p *Fxu_Matrix, pCube *Fxu_Cube, pSopCube string, vFanins *Vec_Int_t, pOrder *int32) {
    var pVar *Fxu_Var;
    var Value, i int32;
    // add literals to the matrix
    // inlining of macro Abc_CubeForEachVar
    for  i = 0; (pSopCube[i] != ' ') && (Value = pSopCube[i]); i++  {
        Value = pSopCube[pOrder[i]];
        if Value == '0' {
            pVar = p.ppVars[ 2 * vFanins.pArray[pOrder[i]] + 1 ];  // CST
            Fxu_MatrixAddLiteral( p, pCube, pVar );
        
        } else if Value == '1' {
            pVar = p.ppVars[ 2 * vFanins.pArray[pOrder[i]] ];  // CST
            Fxu_MatrixAddLiteral( p, pCube, pVar );
        }
    }
}


/**Function*************************************************************

  Synopsis    [Creates the new array of Sop covers from the sparse matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_CreateCovers(p *Fxu_Matrix, pData *Fxu_Data_t) {
    var pCube, pCubeFirst, pCubeNext *Fxu_Cube;
    var pSopCover string;
    var iNode, n int32;

    // get the first cube of the first internal node 
    pCubeFirst = Fxu_CreateCoversFirstCube( p, pData, 0 );

    // go through the internal nodes
    for  n = 0; n < pData.nNodesOld; n++  {
    if pSopCover = (char *)pData.vSops.pArray[n] {
        // get the number of this node
        iNode = n;
        // get the next first cube
        pCubeNext = Fxu_CreateCoversFirstCube(  p, pData, iNode + 1 );
        // check if there any new variables in these cubes
        for  pCube = pCubeFirst; pCube != pCubeNext; pCube = pCube.pNext  {
            if pCube.lLits.pTail && pCube.lLits.pTail.iVar >= 2 * pData.nNodesOld {
                break; }
        }
        if pCube != pCubeNext {
            Fxu_CreateCoversNode( p, pData, iNode, pCubeFirst, pCubeNext ); }
        // update the first cube
        pCubeFirst = pCubeNext;
    }
    }

    // add the covers for the extracted nodes
    for  n = 0; n < pData.nNodesNew; n++  {
        // get the number of this node
        iNode = pData.nNodesOld + n;
        // get the next first cube
        pCubeNext = Fxu_CreateCoversFirstCube( p, pData, iNode + 1 );
        // the node should be added
        Fxu_CreateCoversNode( p, pData, iNode, pCubeFirst, pCubeNext );
        // update the first cube
        pCubeFirst = pCubeNext;
    }
}

/**Function*************************************************************

  Synopsis    [Create Sop covers for one node that has changed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_CreateCoversNode(p *Fxu_Matrix, pData *Fxu_Data_t, iNode int32, pCubeFirst *Fxu_Cube, pCubeNext *Fxu_Cube) {
    var vInputsNew *Vec_Int_t;
    var pSopCover, pSopCube string;
    var pVar *Fxu_Var;
    var pCube *Fxu_Cube;
    var pLit *Fxu_Lit;
    var iNum, nCubes, v int32;

    // collect positive polarity variable in the cubes between pCubeFirst and pCubeNext
    // inlining of macro Fxu_MatrixRingVarsStart
    ((p).ppTailVars  = &((p).pOrderVars));
((p).pOrderVars  = nil);;
    for  pCube = pCubeFirst; pCube != pCubeNext; pCube = pCube.pNext  {
        for  pLit = pCube.lLits.pHead; (pLit != nil); pLit = pLit.pHNext  {
            pVar = p.ppVars[ 2 * (pLit.pVar.iVar/2) + 1 ];
            if pVar.pOrder == nil {
                // inlining of macro Fxu_MatrixRingVarsAdd
                (*((p).ppTailVars ) = pVar );
((p).ppTailVars  = &(pVar).pOrder );
((pVar).pOrder  = (Fxu_Var *)1);; }
        }
    }
    // inlining of macro Fxu_MatrixRingVarsStop
    ;

    // collect the variable numbers
    vInputsNew = Vec_IntAlloc( 4 );
    // inlining of macro Fxu_MatrixForEachVarInRing
    if (p).pOrderVars {    for  pVar = (p).pOrderVars;          pVar != (Fxu_Var *)1;          pVar = pVar.pOrder  {
        Vec_IntPush( vInputsNew, pVar.iVar / 2 );
    } }
    Fxu_MatrixRingVarsUnmark( p );

    // sort the vars by their number
    Vec_IntSort( vInputsNew, 0 );

    // mark the vars with their numbers in the sorted array
    for  v = 0; v < vInputsNew.nSize; v++  {
        p.ppVars[ 2 * vInputsNew.pArray[v] + 0 ].lLits.nItems = v; // hack - reuse lLits.nItems
        p.ppVars[ 2 * vInputsNew.pArray[v] + 1 ].lLits.nItems = v; // hack - reuse lLits.nItems
    }

    // count the number of cubes
    nCubes = 0;
    for  pCube = pCubeFirst; pCube != pCubeNext; pCube = pCube.pNext  {
        if pCube.lLits.nItems {
            nCubes++; }
    }

    // allocate room for the new cover
    pSopCover = Abc_SopStart( pData.pManSop, nCubes, vInputsNew.nSize );
    // set the correct polarity of the cover
    if iNode < pData.nNodesOld && Abc_SopGetPhase( (char *)pData.vSops.pArray[iNode] ) == 0 {
        Abc_SopComplement( pSopCover ); }

    // add the cubes
    nCubes = 0;
    for  pCube = pCubeFirst; pCube != pCubeNext; pCube = pCube.pNext  {
        if pCube.lLits.nItems == 0 {
            continue; }
        // get hold of the SOP cube
        pSopCube = pSopCover + nCubes * (vInputsNew.nSize + 3);
        // insert literals
        for  pLit = pCube.lLits.pHead; (pLit != nil); pLit = pLit.pHNext  {
            iNum = pLit.pVar.lLits.nItems; // hack - reuse lLits.nItems
            assert(iNum < vInputsNew.nSize, "iNum < vInputsNew.nSize");
            if pLit.pVar.iVar / 2 < pData.nNodesOld {
                pSopCube[iNum] = (pLit.pVar.iVar & 1)? '0' : '1';   // reverse CST
            } else {
                pSopCube[iNum] = (pLit.pVar.iVar & 1)? '1' : '0'; }   // no CST
        }
        // count the cube
        nCubes++;
    }
    assert(nCubes == Abc_SopGetCubeNum(pSopCover), "nCubes == Abc_SopGetCubeNum(pSopCover)");

    // set the new cover and the array of fanins
    pData.vSopsNew.pArray[iNode]   = pSopCover;
    pData.vFaninsNew.pArray[iNode] = vInputsNew;
}


/**Function*************************************************************

  Synopsis    [Adds the var to storage.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_CreateCoversFirstCube(p *Fxu_Matrix, pData *Fxu_Data_t, iVar int32) *Fxu_Cube {
    var v int32;
    for  v = iVar; v < pData.nNodesOld + pData.nNodesNew; v++  {
        if p.ppVars[ 2*v + 1 ].pFirst {
            return p.ppVars[ 2*v + 1 ].pFirst; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Compares the vars by their number.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_CreateMatrixLitCompare(ptrX *int32, ptrY *int32) int32 {
    return s_pLits[*ptrX] - s_pLits[*ptrY];
} 

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
//ABC_NAMESPACE_IMPL_END

