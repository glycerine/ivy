package cut

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [cutPre22.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Precomputes truth tables for the 2x2 macro cell.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cutPre22.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cutInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const CUT_CELL_MVAR =  9

type Cut_Cell_t = Cut_Cell_t_;
type Cut_CMan_t = Cut_CMan_t_;

type Cut_Cell_t_ struct {
    pNext        *Cut_Cell_t;                        // pointer to the next cell in the table
    pNextVar        *Cut_Cell_t;                     // pointer to the next cell of this support size
    pParent        *Cut_Cell_t;                      // pointer to the cell used to derive this one
    nUsed                int32;                        // the number of times the cell is used
    Box [4]byte;                       // functions in the boxes
    nVars uint32; /* C bitfield: 4 bits; review packing and truncation */          // the number of variables
    CrossBar0 uint32; /* C bitfield: 4 bits; review packing and truncation */          // the variable set equal
    CrossBar1 uint32; /* C bitfield: 4 bits; review packing and truncation */          // the variable set equal
    CrossBarPhase uint32; /* C bitfield: 2 bits; review packing and truncation */          // the phase of the cross bar (0, 1, or 2)
    CanonPhase uint32; /* C bitfield: 18 bits; review packing and truncation */          // the canonical phase
    CanonPerm [CUT_CELL_MVAR+3]byte;   // semicanonical permutation
    Store [2*CUT_CELL_MVAR]int16;       // minterm counts in the cofactors
    uTruth [1<<(CUT_CELL_MVAR-5)]uint32; // the current truth table
}

type Cut_CMan_t_ struct {
    // storage for canonical cells
    pMem   *Extra_MmFixed_t;
    tTable          *st__table;
    pSameVar [CUT_CELL_MVAR+1]*Cut_Cell_t;
    // elementary truth tables
    uInputs [CUT_CELL_MVAR][1<<(CUT_CELL_MVAR-5)]uint32;
    // temporary truth tables
    uTemp1 [22][1<<(CUT_CELL_MVAR-5)]uint32;
    uTemp2 [22][1<<(CUT_CELL_MVAR-5)]uint32;
    uTemp3 [22][1<<(CUT_CELL_MVAR-5)]uint32;
    uFinal [1<<(CUT_CELL_MVAR-5)]uint32;
    puAux [1<<(CUT_CELL_MVAR-5)]uint32;
    // statistical variables
    nTotal                int32;
    nGood                int32;
    nVarCounts [CUT_CELL_MVAR+1]int32;
    nSymGroups [CUT_CELL_MVAR+1]int32;
    nSymGroupsE [CUT_CELL_MVAR+1]int32;
    timeCanon            abctime;
    timeSupp            abctime;
    timeTable            abctime;
    nCellFound                int32;
    nCellNotFound                int32;
}

// NP-classes of functions of 3 variables (22)
//static char * s_NP3[22] = {
//    " 0\n",                          // 00    const 0            // 0 vars
//    " 1\n",                          // 01    const 1            // 0 vars
//    "1 1\n",                         // 02    a                  // 1 vars
//    "11 1\n",                        // 03    ab                 // 2 vars
//    "11 0\n",                        // 04    (ab)'              // 2 vars
//    "10 1\n01 1\n",                  // 05    a<+>b              // 2 vars
//    "111 1\n",                       // 06 0s abc                // 3 vars
//    "111 0\n",                       // 07    (abc)'             //
//    "11- 1\n1-1 1\n",                // 08 1p a(b+c)             //
//    "11- 0\n1-1 0\n",                // 09    (a(b+c))'          //
//    "111 1\n100 1\n010 1\n001 1\n",  // 10 2s a<+>b<+>c          //
//    "10- 0\n1-0 0\n011 0\n",         // 11 3p a<+>bc             //
//    "101 1\n110 1\n",                // 12 4p a(b<+>c)           //
//    "101 0\n110 0\n",                // 13    (a(b<+>c))'        //
//    "11- 1\n1-1 1\n-11 1\n",         // 14 5s ab+bc+ac           //
//    "111 1\n000 1\n",                // 15 6s abc+a'b'c'         //
//    "111 0\n000 0\n",                // 16    (abc+a'b'c')'      //
//    "11- 1\n-11 1\n0-1 1\n",         // 17 7  ab+bc+a'c          //
//    "011 1\n101 1\n110 1\n",         // 18 8s a'bc+ab'c+abc'     //
//    "011 0\n101 0\n110 0\n",         // 19    (a'bc+ab'c+abc')'  //
//    "100 1\n-11 1\n",                // 20 9p ab'c'+bc           //
//    "100 0\n-11 0\n"                 // 21    (ab'c'+bc)'        //
//};

// NP-classes of functions of 3 variables (22)
var s_NP3Names [22]string = [22]string{
    "   const 0            ",
    "   const 1            ",
    "   a                  ",
    "   ab                 ",
    "   (ab)'              ",
    "   a<+>b              ",
    "0s abc                ",
    "   (abc)'             ",
    "1p a(b+c)             ",
    "   (a(b+c))'          ",
    "2s a<+>b<+>c          ",
    "3p a<+>bc             ",
    "4p a(b<+>c)           ",
    "   (a(b<+>c))'        ",
    "5s ab+bc+ac           ",
    "6s abc+a'b'c'         ",
    "   (abc+a'b'c')'      ",
    "7  ab+bc+a'c          ",
    "8s a'bc+ab'c+abc'     ",
    "   (a'bc+ab'c+abc')'  ",
    "9p ab'c'+bc           ",
    "   (ab'c'+bc)'        ",
};

// the number of variables in each function
//static int s_NP3VarNums[22] = { 0, 0, 1, 2, 2, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3 };

// NPN classes of functions of exactly 3 inputs (10)
var s_NPNe3 [10]int32 = [10]int32{ 6, 8, 10, 11, 12, 14, 15, 17, 18, 20 };

// NPN classes of functions of exactly 3 inputs that are symmetric (5)
//static int s_NPNe3s[10] = { 6, 10, 14, 15, 18 };

// NPN classes of functions of exactly 3 inputs (4)
//static int s_NPNe3p[10] = { 8, 11, 12, 20 };

// static Cut_CMan_t * Cut_CManStart();

// static void Cut_CManStop( Cut_CMan_t * p );

// static void Cut_CellTruthElem( unsigned * InA, unsigned * InB, unsigned * InC, unsigned * pOut, int nVars, int Type );

// static void Cut_CellCanonicize( Cut_CMan_t * p, Cut_Cell_t * pCell );

// static int  Cut_CellTableLookup( Cut_CMan_t * p, Cut_Cell_t * pCell );

// static void Cut_CellSuppMin( Cut_Cell_t * pCell );

// static void Cut_CellCrossBar( Cut_Cell_t * pCell );



var s_pCMan *Cut_CMan_t = nil;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Start the precomputation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CellLoad() {
    var pFile *os.File;
    var pFileName string  = "cells22_daomap_iwls.txt";
    var pString [1000]byte;
    var p *Cut_CMan_t;
    var pCell *Cut_Cell_t;
    var Length int32; //, i;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\".\n", pFileName);
        return;
    }
   // start the manager
    p = Cut_CManStart();
    // load truth tables
    for  fgets(pString, 1000, pFile)  {
        Length = strlen(pString);
        pString[Length--] = 0;
        if Length == 0 {
            continue; }
        // derive the cell
        pCell = (Cut_Cell_t *)Extra_MmFixedEntryFetch( p.pMem );
        *pCell = Cut_Cell_t{};
        pCell.nVars = Abc_Base2Log(Length*4);
        pCell.nUsed = 1;
//        Extra_TruthCopy( pCell->uTruth, pTruth, nVars );
        Extra_ReadHexadecimal( pCell.uTruth, pString, pCell.nVars );
        Cut_CellSuppMin( pCell );
/*
        // set the elementary permutation
        for ( i = 0; i < (int)pCell->nVars; i++ )
            pCell->CanonPerm[i] = i;
        // canonicize
        pCell->CanonPhase = Extra_TruthSemiCanonicize( pCell->uTruth, p->puAux, pCell->nVars, pCell->CanonPerm, pCell->Store );
*/
        // add to the table
        p.nTotal++;

//        Extra_PrintHexadecimal( stdout, pCell->uTruth, pCell->nVars ); printf( "\n" );
//        if ( p->nTotal == 500 )
//            break;

        if !(Cut_CellTableLookup( p, pCell ) != 0) { // new cell
            p.nGood++; }
    }
    fmt.Printf("Read %d cells from file \"%s\". Added %d cells to the table.\n", p.nTotal, pFileName, p.nGood);
    _ = (pFile).Close();
//    return p;
}

/**Function*************************************************************

  Synopsis    [Precomputes truth tables for the 2x2 macro cell.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CellPrecompute() {
    var p *Cut_CMan_t;
    var pCell, pTemp *Cut_Cell_t;
    var i1, i2, i3, i, j, k, c int32;
    var clk abctime  = Abc_Clock(); //, clk2 = Abc_Clock();

    p = Cut_CManStart();

    // precompute truth tables
    for  i = 0; i < 22; i++  {
        Cut_CellTruthElem( p.uInputs[0], p.uInputs[1], p.uInputs[2], p.uTemp1[i], 9, i );
    }
    for  i = 0; i < 22; i++  {
        Cut_CellTruthElem( p.uInputs[3], p.uInputs[4], p.uInputs[5], p.uTemp2[i], 9, i );
    }
    for  i = 0; i < 22; i++  {
        Cut_CellTruthElem( p.uInputs[6], p.uInputs[7], p.uInputs[8], p.uTemp3[i], 9, i );
    }
/*
        if ( k == 8 && ((i1 == 6 && i2 == 14 && i3 == 20) || (i1 == 20 && i2 == 6 && i3 == 14)) )
        {
            Extra_PrintBinary( stdout, &pCell->CanonPhase, pCell->nVars+1 ); printf( " : " );
            for ( i = 0; i < pCell->nVars; i++ )
                printf( "%d=%d/%d  ", pCell->CanonPerm[i], pCell->Store[2*i], pCell->Store[2*i+1] );
            Extra_PrintHexadecimal( stdout, pCell->uTruth, pCell->nVars );
            printf( "\n" );
        }
*/
/*
    // go through symmetric roots
    for ( k = 0; k < 5; k++ )
    for ( i1 =  0; i1 < 22; i1++ )
    for ( i2 = i1; i2 < 22; i2++ )
    for ( i3 = i2; i3 < 22; i3++ )
    {
        // derive the cell
        pCell = (Cut_Cell_t *)Extra_MmFixedEntryFetch( p->pMem );
        memset( pCell, 0, sizeof(Cut_Cell_t) );
        pCell->nVars  = 9;
        pCell->Box[0] = s_NPNe3s[k];
        pCell->Box[1] = i1;
        pCell->Box[2] = i2;
        pCell->Box[3] = i3;
        // fill in the truth table
        Cut_CellTruthElem( p->uTemp1[i1], p->uTemp2[i2], p->uTemp3[i3], pCell->uTruth, 9, s_NPNe3s[k] );
        // canonicize
        Cut_CellCanonicize( pCell );

        // add to the table
        p->nTotal++;
        if ( Cut_CellTableLookup( p, pCell ) ) // already exists
            Extra_MmFixedEntryRecycle( p->pMem, (char *)pCell );
        else
            p->nGood++;
    }

    // go through partially symmetric roots
    for ( k = 0; k < 4; k++ )
    for ( i1 = 0;  i1 < 22; i1++ )
    for ( i2 = 0;  i2 < 22; i2++ )
    for ( i3 = i2; i3 < 22; i3++ )
    {
        // derive the cell
        pCell = (Cut_Cell_t *)Extra_MmFixedEntryFetch( p->pMem );
        memset( pCell, 0, sizeof(Cut_Cell_t) );
        pCell->nVars  = 9;
        pCell->Box[0] = s_NPNe3p[k];
        pCell->Box[1] = i1;
        pCell->Box[2] = i2;
        pCell->Box[3] = i3;
        // fill in the truth table
        Cut_CellTruthElem( p->uTemp1[i1], p->uTemp2[i2], p->uTemp3[i3], pCell->uTruth, 9, s_NPNe3p[k] );
        // canonicize
        Cut_CellCanonicize( pCell );

        // add to the table
        p->nTotal++;
        if ( Cut_CellTableLookup( p, pCell ) ) // already exists
            Extra_MmFixedEntryRecycle( p->pMem, (char *)pCell );
        else
            p->nGood++;
    }

    // go through non-symmetric functions
    for ( i1 = 0; i1 < 22; i1++ )
    for ( i2 = 0; i2 < 22; i2++ )
    for ( i3 = 0; i3 < 22; i3++ )
    {
        // derive the cell
        pCell = (Cut_Cell_t *)Extra_MmFixedEntryFetch( p->pMem );
        memset( pCell, 0, sizeof(Cut_Cell_t) );
        pCell->nVars  = 9;
        pCell->Box[0] = 17;
        pCell->Box[1] = i1;
        pCell->Box[2] = i2;
        pCell->Box[3] = i3;
        // fill in the truth table
        Cut_CellTruthElem( p->uTemp1[i1], p->uTemp2[i2], p->uTemp3[i3], pCell->uTruth, 9, 17 );
        // canonicize
        Cut_CellCanonicize( pCell );

        // add to the table
        p->nTotal++;
        if ( Cut_CellTableLookup( p, pCell ) ) // already exists
            Extra_MmFixedEntryRecycle( p->pMem, (char *)pCell );
        else
            p->nGood++;
    }
*/

    // go through non-symmetric functions
    for  k = 0; k < 10; k++  {
    for  i1 = 0; i1 < 22; i1++  {
    for  i2 = 0; i2 < 22; i2++  {
    for  i3 = 0; i3 < 22; i3++  {
        // derive the cell
        pCell = (Cut_Cell_t *)Extra_MmFixedEntryFetch( p.pMem );
        *pCell = Cut_Cell_t{};
        pCell.nVars  = 9;
        pCell.Box[0] = s_NPNe3[k];
        pCell.Box[1] = i1;
        pCell.Box[2] = i2;
        pCell.Box[3] = i3;
        // set the elementary permutation
        for  i = 0; i < int32(pCell.nVars); i++  {
            pCell.CanonPerm[i] = i;
        }
        // fill in the truth table
        Cut_CellTruthElem( p.uTemp1[i1], p.uTemp2[i2], p.uTemp3[i3], pCell.uTruth, 9, s_NPNe3[k] );
        // minimize the support
        Cut_CellSuppMin( pCell );

        // canonicize
        pCell.CanonPhase = Extra_TruthSemiCanonicize( pCell.uTruth, p.puAux, pCell.nVars, pCell.CanonPerm, pCell.Store );

        // add to the table
        p.nTotal++;
        if Cut_CellTableLookup( p, pCell ) != 0 { // already exists
            Extra_MmFixedEntryRecycle( p.pMem, (char *)pCell );
        } else {
            p.nGood++;
            p.nVarCounts[pCell.nVars]++;

            if pCell.nVars {
            for  i = 0; i < int32(pCell.nVars)-1; i++  {
                if pCell.Store[2*i] != pCell.Store[2*(i+1)] { // i and i+1 cannot be symmetric
                    continue; }
                // i and i+1 can be symmetric
                // find the end of this group
                for  j = i+1; j < int32(pCell.nVars); j++  {
                    if pCell.Store[2*i] != pCell.Store[2*j] { 
                        break; }
                }

                if pCell.Store[2*i] == pCell.Store[2*i+1] {
                    p.nSymGroupsE[j-i]++;
                } else {
                    p.nSymGroups[j-i]++; }
                i = j - 1;
            } }
/*
            if ( pCell->nVars == 3 )
            {
                Extra_PrintBinary( stdout, pCell->uTruth, 32 ); printf( "\n" );
                for ( i = 0; i < (int)pCell->nVars; i++ )
                    printf( "%d=%d/%d  ", pCell->CanonPerm[i], pCell->Store[2*i], pCell->Store[2*i+1] );
                printf( "\n" );
            }
*/
        }
    }
    }
    }
    }

    fmt.Printf("BASIC: Total = %d. Good = %d. Entry = %d. ", int32(p.nTotal), int32(p.nGood), int32(sizeof(Cut_Cell_t)));
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    fmt.Printf("Cells:  ");
    for  i = 0; i <= 9; i++  {
        fmt.Printf("%d=%d ", i, p.nVarCounts[i]);
    }
    fmt.Printf("\nDiffs:  ");
    for  i = 0; i <= 9; i++  {
        fmt.Printf("%d=%d ", i, p.nSymGroups[i]);
    }
    fmt.Printf("\nEquals: ");
    for  i = 0; i <= 9; i++  {
        fmt.Printf("%d=%d ", i, p.nSymGroupsE[i]);
    }
    fmt.Printf("\n");

    // continue adding new cells using support
    for  k = CUT_CELL_MVAR; k > 3; k--  {
        for  pTemp = p.pSameVar[k]; pTemp; pTemp = pTemp.pNextVar  {
        for  i1 = 0; i1 < k; i1++  {
        for  i2 = i1+1; i2 < k; i2++  {
        for  c = 0; c < 3; c++  {
            // derive the cell
            pCell = (Cut_Cell_t *)Extra_MmFixedEntryFetch( p.pMem );
            *pCell = Cut_Cell_t{};
            pCell.nVars   = pTemp.nVars;
            pCell.pParent = pTemp;
            // set the elementary permutation
            for  i = 0; i < int32(pCell.nVars); i++  {
                pCell.CanonPerm[i] = i;
            }
            // fill in the truth table
            Extra_TruthCopy( pCell.uTruth, pTemp.uTruth, pTemp.nVars );
            // create the cross-bar
            pCell.CrossBar0 = i1;
            pCell.CrossBar1 = i2;
            pCell.CrossBarPhase = c;
            Cut_CellCrossBar( pCell );
            // minimize the support
//clk2 = Abc_Clock();
            Cut_CellSuppMin( pCell );
//p->timeSupp += Abc_Clock() - clk2;
            // canonicize
//clk2 = Abc_Clock();
            pCell.CanonPhase = Extra_TruthSemiCanonicize( pCell.uTruth, p.puAux, pCell.nVars, pCell.CanonPerm, pCell.Store );
//p->timeCanon += Abc_Clock() - clk2;

            // add to the table
//clk2 = Abc_Clock();
            p.nTotal++;
            if Cut_CellTableLookup( p, pCell ) != 0 { // already exists
                Extra_MmFixedEntryRecycle( p.pMem, (char *)pCell );
            } else {
                p.nGood++;
                p.nVarCounts[pCell.nVars]++;

                for  i = 0; i < int32(pCell.nVars)-1; i++  {
                    if pCell.Store[2*i] != pCell.Store[2*(i+1)] { // i and i+1 cannot be symmetric
                        continue; }
                    // i and i+1 can be symmetric
                    // find the end of this group
                    for  j = i+1; j < int32(pCell.nVars); j++  {
                        if pCell.Store[2*i] != pCell.Store[2*j] { 
                            break; }
                    }

                    if pCell.Store[2*i] == pCell.Store[2*i+1] {
                        p.nSymGroupsE[j-i]++;
                    } else {
                        p.nSymGroups[j-i]++; }
                    i = j - 1;
                }
/*
                if ( pCell->nVars == 3 )
                {
                    Extra_PrintBinary( stdout, pCell->uTruth, 32 ); printf( "\n" );
                    for ( i = 0; i < (int)pCell->nVars; i++ )
                        printf( "%d=%d/%d  ", pCell->CanonPerm[i], pCell->Store[2*i], pCell->Store[2*i+1] );
                    printf( "\n" );
                }
*/
            }
//p->timeTable += Abc_Clock() - clk2;
        }
        }
        }
        }

        fmt.Printf("VAR %d: Total = %d. Good = %d. Entry = %d. ", k, p.nTotal, p.nGood, int32(sizeof(Cut_Cell_t)));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        fmt.Printf("Cells:  ");
        for  i = 0; i <= 9; i++  {
            fmt.Printf("%d=%d ", i, p.nVarCounts[i]);
        }
        fmt.Printf("\nDiffs:  ");
        for  i = 0; i <= 9; i++  {
            fmt.Printf("%d=%d ", i, p.nSymGroups[i]);
        }
        fmt.Printf("\nEquals: ");
        for  i = 0; i <= 9; i++  {
            fmt.Printf("%d=%d ", i, p.nSymGroupsE[i]);
        }
        fmt.Printf("\n");
    }
//    printf( "\n" );
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Supp "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSupp)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Canon"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeCanon)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Table"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTable)))/(float64(int64(time.Second))));;
//    Cut_CManStop( p );
}

/**Function*************************************************************

  Synopsis    [Check the table.]

  Description [Returns 1 if such a truth table already exists.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CellTableLookup(p *Cut_CMan_t, pCell *Cut_Cell_t) int32 {
    var pSlot **Cut_Cell_t;
    var pTemp *Cut_Cell_t;
    var Hash uint32;
    Hash = Extra_TruthHash( pCell.uTruth, Extra_TruthWordNum( pCell.nVars ) );
    if ! st__find_or_add( p.tTable, (char *)(ABC_PTRUINT_T)Hash, (char ***)&pSlot ) {
        *pSlot = nil; }
    for  pTemp = *pSlot; pTemp; pTemp = pTemp.pNext  {
        if pTemp.nVars != pCell.nVars {
            continue; }
        if Extra_TruthIsEqual(pTemp.uTruth, pCell.uTruth, pCell.nVars) {
            return 1; }
    }
    // the entry is new
    pCell.pNext = *pSlot;
    *pSlot = pCell;
    // add it to the variable support list
    pCell.pNextVar = p.pSameVar[pCell.nVars];
    p.pSameVar[pCell.nVars] = pCell;
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Cut_CellSuppMin_uTemp [1<<(CUT_CELL_MVAR-5)]uint32;
func Cut_CellSuppMin(pCell *Cut_Cell_t) {
    
    var pIn, pOut, pTemp *uint32;
    var i, k, Counter, Temp int32;

    // go backward through the support variables and remove redundant
    for  k = pCell.nVars - 1; k >= 0; k--  {
        if !Extra_TruthVarInSupport(pCell.uTruth, pCell.nVars, k) {
            // shift all the variables above this one
            Counter = 0;
            pIn = pCell.uTruth; pOut = spStatic_Cut_CellSuppMin_uTemp;
            for  i = k; i < int32(pCell.nVars) - 1; i++  {
                Extra_TruthSwapAdjacentVars( pOut, pIn, pCell.nVars, i );
                pTemp = pIn; pIn = pOut; pOut = pTemp;
                // swap the support vars
                Temp = pCell.CanonPerm[i]; 
                pCell.CanonPerm[i] = pCell.CanonPerm[i+1];
                pCell.CanonPerm[i+1] = Temp;
                Counter++;
            }
            // return the function back into its place
            if Counter & 1 {
                Extra_TruthCopy( pOut, pIn, pCell.nVars ); }
            // remove one variable
            pCell.nVars--;
//            Extra_PrintBinary( stdout, pCell->uTruth, (1<<pCell->nVars) ); printf( "\n" );
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Cut_CellCrossBar_uTemp0 [1<<(CUT_CELL_MVAR-5)]uint32;
var spStatic_Cut_CellCrossBar_uTemp1 [1<<(CUT_CELL_MVAR-5)]uint32;
func Cut_CellCrossBar(pCell *Cut_Cell_t) {
    
    
    Extra_TruthCopy( spStatic_Cut_CellCrossBar_uTemp0, pCell.uTruth, pCell.nVars );
    Extra_TruthCopy( spStatic_Cut_CellCrossBar_uTemp1, pCell.uTruth, pCell.nVars );
    if pCell.CanonPhase == 0 {
        Extra_TruthCofactor0( spStatic_Cut_CellCrossBar_uTemp0, pCell.nVars, pCell.CrossBar0 );
        Extra_TruthCofactor0( spStatic_Cut_CellCrossBar_uTemp0, pCell.nVars, pCell.CrossBar1 );
        Extra_TruthCofactor1( spStatic_Cut_CellCrossBar_uTemp1, pCell.nVars, pCell.CrossBar0 );
        Extra_TruthCofactor1( spStatic_Cut_CellCrossBar_uTemp1, pCell.nVars, pCell.CrossBar1 );
    
    } else if pCell.CanonPhase == 1 {
        Extra_TruthCofactor1( spStatic_Cut_CellCrossBar_uTemp0, pCell.nVars, pCell.CrossBar0 );
        Extra_TruthCofactor0( spStatic_Cut_CellCrossBar_uTemp0, pCell.nVars, pCell.CrossBar1 );
        Extra_TruthCofactor0( spStatic_Cut_CellCrossBar_uTemp1, pCell.nVars, pCell.CrossBar0 );
        Extra_TruthCofactor1( spStatic_Cut_CellCrossBar_uTemp1, pCell.nVars, pCell.CrossBar1 );
    
    } else if pCell.CanonPhase == 2 {
        Extra_TruthCofactor0( spStatic_Cut_CellCrossBar_uTemp0, pCell.nVars, pCell.CrossBar0 );
        Extra_TruthCofactor1( spStatic_Cut_CellCrossBar_uTemp0, pCell.nVars, pCell.CrossBar1 );
        Extra_TruthCofactor1( spStatic_Cut_CellCrossBar_uTemp1, pCell.nVars, pCell.CrossBar0 );
        Extra_TruthCofactor0( spStatic_Cut_CellCrossBar_uTemp1, pCell.nVars, pCell.CrossBar1 );
    
    } else { assert(false, "0"); }
    Extra_TruthMux( pCell.uTruth, spStatic_Cut_CellCrossBar_uTemp0, spStatic_Cut_CellCrossBar_uTemp1, pCell.nVars, pCell.CrossBar0 );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CellTruthElem(InA *uint32, InB *uint32, InC *uint32, pOut *uint32, nVars int32, Type int32) {
    var nWords int32  = Extra_TruthWordNum( nVars );
    var i int32;

    assert(Type < 22, "Type < 22");
    switch  Type  {
    // " 0\n",                         // 00   const 0
    case 0:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = 0;
        return;
    // " 1\n",                         // 01   const 1
    case 1:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = 0xFFFFFFFF;
        return;
    // "1 1\n",                        // 02   a
    case 2:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = InA[i];
        return;
    // "11 1\n",                       // 03   ab
    case 3:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = InA[i] & InB[i];
        return;
    // "11 0\n",                       // 04   (ab)'
    case 4:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = ~(InA[i] & InB[i]);
        return;
    // "10 1\n01 1\n",                 // 05   a<+>b
    case 5:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = InA[i] ^ InB[i];
        return;
    // "111 1\n",                      // 06 + abc
    case 6:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = InA[i] & InB[i] & InC[i];
        return;
    // "111 0\n",                      // 07   (abc)'
    case 7:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = ~(InA[i] & InB[i] & InC[i]);
        return;
    // "11- 1\n1-1 1\n",               // 08 + a(b+c)
    case 8:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = InA[i] & (InB[i] | InC[i]);
        return;
    // "11- 0\n1-1 0\n",               // 09   (a(b+c))'
    case 9:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = ~(InA[i] & (InB[i] | InC[i]));
        return;
    // "111 1\n100 1\n010 1\n001 1\n", // 10 + a<+>b<+>c
    case 10:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = InA[i] ^ InB[i] ^ InC[i];
        return;
    // "10- 0\n1-0 0\n011 0\n",        // 11 + a<+>bc
    case 11:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = InA[i] ^ (InB[i] & InC[i]);
        return;
    // "101 1\n110 1\n",               // 12 + a(b<+>c)
    case 12:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = InA[i] & (InB[i] ^ InC[i]);
        return;
    // "101 0\n110 0\n",               // 13   (a(b<+>c))'
    case 13:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = ~(InA[i] & (InB[i] ^ InC[i]));
        return;
    // "11- 1\n1-1 1\n-11 1\n",        // 14 + ab+bc+ac
    case 14:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = (InA[i] & InB[i]) | (InB[i] & InC[i]) | (InA[i] & InC[i]);
        return;
    // "111 1\n000 1\n",               // 15 + abc+a'b'c'
    case 15:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = (InA[i] & InB[i] & InC[i]) | (~InA[i] & ~InB[i] & ~InC[i]);
        return;
    // "111 0\n000 0\n",               // 16   (abc+a'b'c')'
    case 16:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = ~((InA[i] & InB[i] & InC[i]) | (~InA[i] & ~InB[i] & ~InC[i]));
        return;
    // "11- 1\n-11 1\n0-1 1\n",        // 17 + ab+bc+a'c
    case 17:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = (InA[i] & InB[i]) | (InB[i] & InC[i]) | (~InA[i] & InC[i]);
        return;
    // "011 1\n101 1\n110 1\n",        // 18 + a'bc+ab'c+abc'
    case 18:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = (~InA[i] & InB[i] & InC[i]) | (InA[i] & ~InB[i] & InC[i]) | (InA[i] & InB[i] & ~InC[i]);
        return;
    // "011 0\n101 0\n110 0\n",        // 19   (a'bc+ab'c+abc')'
    case 19:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = ~((~InA[i] & InB[i] & InC[i]) | (InA[i] & ~InB[i] & InC[i]) | (InA[i] & InB[i] & ~InC[i]));
        return;
    // "100 1\n-11 1\n",               // 20 + ab'c'+bc
    case 20:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = (InA[i] & ~InB[i] & ~InC[i]) | (InB[i] & InC[i]);
        return;
    // "100 0\n-11 0\n"                // 21   (ab'c'+bc)'
    case 21:
        for ( i = 0; i < nWords; i++ )
            pOut[i] = ~((InA[i] & ~InB[i] & ~InC[i]) | (InB[i] & InC[i]));
        return;
    }
}


/**Function*************************************************************

  Synopsis    [Start the precomputation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CManStart() *Cut_CMan_t {
    var p *Cut_CMan_t;
    var i, k int32;
    // start the manager
    assert(sizeof(unsigned) == 4, "sizeof(unsigned) == 4");
    // inlining of macro ABC_ALLOC
    p = ((Cut_CMan_t *) malloc(sizeof(Cut_CMan_t) * (size_t)(1)));
    *p = Cut_CMan_t{};
    // start the table and the memory manager
    p.tTable = st__init_table( st__ptrcmp, st__ptrhash);;
    p.pMem = Extra_MmFixedStart( sizeof(Cut_Cell_t) );
    // set elementary truth tables
    for  k = 0; k < CUT_CELL_MVAR; k++  {
        for  i = 0; i < (1<<CUT_CELL_MVAR); i++  {
            if i & (1 << k) {
                p.uInputs[k][i>>5] |= (1 << (i&31)); }
        }
    }
    s_pCMan = p;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CManStop(p *Cut_CMan_t) {
    st__free_table( p.tTable );
    Extra_MmFixedStop( p.pMem );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CellIsRunning() int32 {
    return s_pCMan != nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CellDumpToFile() {
    var pFile *os.File;
    var p *Cut_CMan_t  = s_pCMan;
    var pTemp *Cut_Cell_t;
    var pFileName string  = "celllib22.txt";
    var NumUsed [10][5]int32 = [10][5]int32{{0}};
    var BoxUsed [22][5]int32 = [22][5]int32{{0}};
    var i, k, Counter int32;
    var clk abctime  = Abc_Clock();

    if p == nil {
        fmt.Printf("Cut_CellDumpToFile: Cell manager is not defined.\n");
        return;
    }

    // count the number of cells used
    for  k = CUT_CELL_MVAR; k >= 0; k--  {
        for  pTemp = p.pSameVar[k]; pTemp; pTemp = pTemp.pNextVar  {
            if pTemp.nUsed == 0 {
                NumUsed[k][0]++;
            } else if pTemp.nUsed < 10 {
                NumUsed[k][1]++;
            } else if pTemp.nUsed < 100 {
                NumUsed[k][2]++;
            } else if pTemp.nUsed < 1000 {
                NumUsed[k][3]++;
            } else { 
                NumUsed[k][4]++; }

            for  i = 0; i < 4; i++  {
                if pTemp.nUsed == 0 {
                    BoxUsed[ int32(pTemp.Box[i]) ][0]++;
                } else if pTemp.nUsed < 10 {
                    BoxUsed[ int32(pTemp.Box[i]) ][1]++;
                } else if pTemp.nUsed < 100 {
                    BoxUsed[ int32(pTemp.Box[i]) ][2]++;
                } else if pTemp.nUsed < 1000 {
                    BoxUsed[ int32(pTemp.Box[i]) ][3]++;
                } else { 
                    BoxUsed[ int32(pTemp.Box[i]) ][4]++; }
            }
        }
    }

    fmt.Printf("Functions found = %10d.  Functions not found = %10d.\n", p.nCellFound, p.nCellNotFound);
    for  k = 0; k <= CUT_CELL_MVAR; k++  {
        fmt.Printf("%3d  : ", k);
        for  i = 0; i < 5; i++  { 
            fmt.Printf("%8d ", NumUsed[k][i]);
        }
        fmt.Printf("\n");
    }
    fmt.Printf("Box usage:\n");
    for  k = 0; k < 22; k++  {
        fmt.Printf("%3d  : ", k);
        for  i = 0; i < 5; i++  { 
            fmt.Printf("%8d ", BoxUsed[k][i]);
        }
        fmt.Printf("  %s", s_NP3Names[k]);
        fmt.Printf("\n");
    }

    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cut_CellDumpToFile: Cannout open output file.\n");
        return;
    }

    Counter = 0;
    for  k = 0; k <= CUT_CELL_MVAR; k++  {
        for  pTemp = p.pSameVar[k]; pTemp; pTemp = pTemp.pNextVar  {
            if pTemp.nUsed > 0 {
                Extra_PrintHexadecimal( pFile, pTemp.uTruth, (k <= 5? 5 : k) );
                fmt.Fprintf(pFile, "\n");
                Counter++;
            }
        }
        fmt.Fprintf(pFile, "\n");
    }
    _ = (pFile).Close();

    fmt.Printf("Library composed of %d functions is written into file \"%s\".  ", Counter, pFileName);

    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}


/**Function*************************************************************

  Synopsis    [Looks up if the given function exists in the hash table.]

  Description [If the function exists, returns 1, meaning that it can be
  implemented using two levels of 3-input LUTs. If the function does not 
  exist, return 0.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CellTruthLookup(pTruth *uint32, nVars int32) int32 {
    var p *Cut_CMan_t  = s_pCMan;
    var pTemp *Cut_Cell_t;
    var Cell Cut_Cell_t;
    var pCell *Cut_Cell_t = &Cell;
    var Hash uint32;
    var i int32;

    // cell manager is not defined
    if p == nil {
        fmt.Printf("Cut_CellTruthLookup: Cell manager is not defined.\n");
        return 0;
    }

    // canonicize
    *pCell = Cut_Cell_t{};
    pCell.nVars = nVars;
    Extra_TruthCopy( pCell.uTruth, pTruth, nVars );
    Cut_CellSuppMin( pCell );
    // set the elementary permutation
    for  i = 0; i < int32(pCell.nVars); i++  {
        pCell.CanonPerm[i] = i;
    }
    // canonicize
    pCell.CanonPhase = Extra_TruthSemiCanonicize( pCell.uTruth, p.puAux, pCell.nVars, pCell.CanonPerm, pCell.Store );


    // check if the cell exists
    Hash = Extra_TruthHash( pCell.uTruth, Extra_TruthWordNum(pCell.nVars) );
    if st__lookup( p.tTable, (char *)(ABC_PTRUINT_T)Hash, (char **)&pTemp ) {
        for  ; pTemp; pTemp = pTemp.pNext  {
            if pTemp.nVars != pCell.nVars {
                continue; }
            if Extra_TruthIsEqual(pTemp.uTruth, pCell.uTruth, pCell.nVars) {
                pTemp.nUsed++;
                p.nCellFound++;
                return 1;
            }
        }
    }
    p.nCellNotFound++;
    return 0;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

