package cas

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [casDec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CASCADE: Decomposition of shared BDDs into a LUT cascade.]

  Synopsis    [BDD-based decomposition with encoding.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Spring 2002.]

  Revision    [$Id: casDec.c,v 1.0 2002/01/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <string.h>
//#include <stdlib.h>

//#include "bdd/extrab/extraBdd.h"
//#include "cas.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                      type definitions                            ///
////////////////////////////////////////////////////////////////////////

typedef struct
{
    int nIns;                // the number of LUT variables
    int nInsP;               // the number of inputs coming from the previous LUT
    int nCols;               // the number of columns in this LUT
    int nMulti;              // the column multiplicity, [log2(nCols)]
    int nSimple;             // the number of outputs implemented as direct connections to inputs of the previous block
    int Level;               // the starting level in the ADD in this LUT

//  DdNode ** pbVarsIn[32];  // the BDDs of the elementary input variables
//  DdNode ** pbVarsOut[32]; // the BDDs of the elementary output variables 

//  char * pNamesIn[32];     // the names of input variables
//  char * pNamesOut[32];    // the names of output variables

    DdNode ** pbCols;        // the array of columns represented by BDDs
    DdNode ** pbCodes;       // the array of codes (in terms of pbVarsOut)
    DdNode ** paNodes;       // the array of starting ADD nodes on the next level (also referenced)

    DdNode * bRelation;      // the relation after encoding

    // the relation depends on the three groups of variables:
    // (1) variables on top represent the outputs of the previous cascade
    // (2) variables in the middle represent the primary inputs
    // (3) variables below (CVars) represent the codes
    //
    // the replacement is done after computing the relation
}
 LUT;


////////////////////////////////////////////////////////////////////////
///                      static functions                            ///
////////////////////////////////////////////////////////////////////////

// the LUT-2-BLIF writing function
// void WriteLUTSintoBLIFfile( FILE * pFile, DdManager * dd, LUT ** pLuts, int nLuts, DdNode ** bCVars, char ** pNames, int nNames, char * FileName );


// the function to write a DD (BDD or ADD) as a network of MUXES
// extern void WriteDDintoBLIFfile( FILE * pFile, DdNode * Func, char * OutputName, char * Prefix, char ** InputNames );

// extern void WriteDDintoBLIFfileReorder( DdManager * dd, FILE * pFile, DdNode * Func, char * OutputName, char * Prefix, char ** InputNames );


////////////////////////////////////////////////////////////////////////
///                      static varibles                             ///
////////////////////////////////////////////////////////////////////////

var s_LutSize int32 = 15;
var s_nFuncVars int32; 

var s_EncodingTime long;

var s_EncSearchTime long;
var s_EncComputeTime long;

////////////////////////////////////
// temporary output variables
//FILE * pTable;
//long s_ReadingTime;
//long s_RemappingTime;
////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                      debugging macros                            ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define PRB_(f)       printf( #f " = " ); Cudd_bddPrint(dd,f); printf( "\n" )

// Original macro:
// #define PRK(f,n)     Cudd_PrintKMap(stdout,dd,(f),Cudd_Not(f),(n),NULL,0); printf( "K-map for function" #f "\n\n" )

// Original macro:
// #define PRK2(f,g,n)  Cudd_PrintKMap(stdout,dd,(f),(g),(n),NULL,0); printf( "K-map for function <" #f ", " #g ">\n\n" ) 



////////////////////////////////////////////////////////////////////////
///                     EXTERNAL FUNCTIONS                           ///
////////////////////////////////////////////////////////////////////////

int CreateDecomposedNetwork( DdManager * dd, DdNode * aFunc, char ** pNames, int nNames, char * FileName, int nLutSize, int fCheck, int fVerbose )
// aFunc is a 0-1 ADD for the given function
// pNames (nNames) are the input variable names
// FileName is the name of the output file for the LUT network
// dynamic variable reordering should be disabled when this function is running
{
    static LUT * pLuts[MAXINPUTS];   // the LUT cascade
    static int Profile[MAXINPUTS];   // the profile filled in with the info about the BDD width
    static int Permute[MAXINPUTS];   // the array to store a temporary permutation of variables

    var p *LUT;               // the current LUT
    var i, v int32;

    var bCVars [32]*DdNode;   // these are variables for the codes

    var nVarsRem int32;          // the number of remaining variables
    var PrevMulti int32;         // column multiplicity on the previous level
    var fLastLut int32;          // flag to signal the last LUT
    var nLuts int32;
    var nLutsTotal int32;
    var nLutOutputs int32;
    var nLutOutputsOrig int32;

    var clk1 abctime;

    s_LutSize = nLutSize;

    s_nFuncVars = nNames;

    // get the profile
    clk1 = Abc_Clock();
    Extra_ProfileWidth( dd, aFunc, Profile, -1 );


//  for ( v = 0; v < nNames; v++ )
//      printf( "Level = %2d, Width = %2d\n", v+1, Profile[v] );


//printf( "\n" );

    // mark-up the LUTs
    // assuming that the manager has exactly nNames vars (new vars have not been introduced yet)
    nVarsRem  = nNames;     // the number of remaining variables
    PrevMulti = 0;          // column multiplicity on the previous level
    fLastLut  = 0;
    nLuts     = 0;
    for
    {
        p = (LUT*) ABC_ALLOC( char, sizeof(LUT) );
        memset( p, 0, sizeof(LUT) );

        if nVarsRem + PrevMulti <= s_LutSize { // this is the last LUT
        
            p.nIns   = nVarsRem + PrevMulti;
            p.nInsP  = PrevMulti;
            p.nCols  = 2;
            p.nMulti = 1;
            p.Level  = nNames-nVarsRem;

            nVarsRem  = 0;
            PrevMulti = 1;

            fLastLut  = 1;
        
        } else { // this is not the last LUT
        
            p.nIns   = s_LutSize;
            p.nInsP  = PrevMulti;
            p.nCols  = Profile[nNames-(nVarsRem-(s_LutSize-PrevMulti))];
            p.nMulti = Abc_Base2Log(p.nCols);
            p.Level  = nNames-nVarsRem;

            nVarsRem  = nVarsRem-(s_LutSize-PrevMulti);
            PrevMulti = p.nMulti;
        }
        
        if p.nMulti >= s_LutSize {
            fmt.Printf("The LUT size is too small\n");
            return 0;
        }

        nLutOutputsOrig += p.nMulti;


//if ( fVerbose )
//printf( "Stage %2d: In = %3d, InP = %3d, Cols = %5d, Multi = %2d, Level = %2d\n", 
//       nLuts+1, p->nIns, p->nInsP, p->nCols, p->nMulti, p->Level );


        // there should be as many columns, codes, and nodes, as there are columns on this level
        p.pbCols  = (DdNode **) ABC_ALLOC( char, p.nCols * sizeof(DdNode *) );
        p.pbCodes = (DdNode **) ABC_ALLOC( char, p.nCols * sizeof(DdNode *) );
        p.paNodes = (DdNode **) ABC_ALLOC( char, p.nCols * sizeof(DdNode *) );

        pLuts[nLuts] = p;
        nLuts++;
    
if !( !(fLastLut != 0) ) { break }
}
    


//if ( fVerbose )
//printf( "The number of cascades = %d\n", nLuts );


//fprintf( pTable, "%d ", nLuts );


    // add the new variables at the bottom
    for  i = 0; i < s_LutSize; i++  {
        bCVars[i] = Cudd_bddNewVar(dd);
    }

    // for each LUT - assign the LUT and encode the columns
    s_EncodingTime = 0;
    for  i = 0; i < nLuts; i++  {
        var RetValue int32;
        var bVars [32]*DdNode;    
        var nVars int32;
        var bVarsInCube *DdNode;
        var bVarsCCube *DdNode;
        var bVarsCube *DdNode;
        var CutLevel int32;

        p = pLuts[i];

        // compute the columns of this LUT starting from the given set of nodes with the given codes
        // (these codes have been remapped to depend on the topmost variables in the manager)
        // for the first LUT, start with the constant 1 BDD
        CutLevel = p.Level + p.nIns - p.nInsP;
        if i == 0 {
            RetValue = Extra_bddNodePathsUnderCutArray( 
                            dd, &aFunc, &(b1), 1, 
                            p.paNodes, p.pbCols, CutLevel );
        } else {
            RetValue = Extra_bddNodePathsUnderCutArray( 
                            dd, pLuts[i-1].paNodes, pLuts[i-1].pbCodes, pLuts[i-1].nCols, 
                            p.paNodes, p.pbCols, CutLevel ); }
        assert(RetValue == p.nCols, "RetValue == p.nCols");
        // at this point, we have filled out p->paNodes[] and p->pbCols[] of this LUT
        // pLuts[i-1]->paNodes depended on normal vars
        // pLuts[i-1]->pbCodes depended on the topmost variables 
        // the resulting p->paNodes depend on normal ADD nodes
        // the resulting p->pbCols depend on normal vars and topmost variables in the manager

        // perform the encoding

        // create the cube of these variables
        // collect the topmost variables of the manager
        nVars = p.nInsP;
        for  v = 0; v < nVars; v++  {
            bVars[v] = dd.vars[ dd.invperm[v] ];
        }
        bVarsCCube  = Extra_bddBitsToCube( dd, (1<<nVars)-1, nVars, bVars, 1 );    Cudd_Ref( bVarsCCube );

        // collect the primary input variables involved in this LUT
        nVars = p.nIns - p.nInsP;
        for  v = 0; v < nVars; v++  {
            bVars[v] = dd.vars[ dd.invperm[p.Level+v] ];
        }
        bVarsInCube = Extra_bddBitsToCube( dd, (1<<nVars)-1, nVars, bVars, 1 );    Cudd_Ref( bVarsInCube );

        // get the cube
        bVarsCube   = Cudd_bddAnd( dd, bVarsInCube, bVarsCCube );              Cudd_Ref( bVarsCube );
        Cudd_RecursiveDeref( dd, bVarsInCube );
        Cudd_RecursiveDeref( dd, bVarsCCube );

        // get the encoding relation
        if i == nLuts -1 {
            var bVar *DdNode;
            assert(p.nMulti == 1, "p.nMulti == 1");
            assert(p.nCols == 2, "p.nCols == 2");
            // inlining of macro Cudd_IsConstant
            // inlining of macro Cudd_Regular
            assert(((((DdNode *)((ptruint)(p.paNodes[0]) & ~01))).index == CUDD_CONST_INDEX), "/*__starting_point_inline_macro:Cudd_IsConstant__*/((/*__starting_point_inline_macro:Cudd_Regular__*/((DdNode *)((ptruint)(p.paNodes[0]) & ~01))).index == CUDD_CONST_INDEX)");
            // inlining of macro Cudd_IsConstant
            // inlining of macro Cudd_Regular
            assert(((((DdNode *)((ptruint)(p.paNodes[1]) & ~01))).index == CUDD_CONST_INDEX), "/*__starting_point_inline_macro:Cudd_IsConstant__*/((/*__starting_point_inline_macro:Cudd_Regular__*/((DdNode *)((ptruint)(p.paNodes[1]) & ~01))).index == CUDD_CONST_INDEX)");

            if p.paNodes[0] == a1 {
                bVar = bCVars[0];
            } else {
                // inlining of macro Cudd_Not
                bVar = ((DdNode *)((ptrint)(bCVars[0]) ^ 01));
            }
            p.bRelation = Cudd_bddIte( dd, bVar, p.pbCols[0], p.pbCols[1] );  Cudd_Ref( p.bRelation );
        
        } else {
            var clk2 abctime  = Abc_Clock();
//          p->bRelation = PerformTheEncoding( dd, p->pbCols, p->nCols, bVarsCube, bCVars, p->nMulti, &p->nSimple );  Cudd_Ref( p->bRelation );
            p.bRelation = Extra_bddEncodingNonStrict( dd, p.pbCols, p.nCols, bVarsCube, bCVars, p.nMulti, &p.nSimple );  Cudd_Ref( p.bRelation );
            s_EncodingTime += Abc_Clock() - clk2;
        }

        // update the number of LUT outputs
        nLutOutputs += (p.nMulti - p.nSimple);
        nLutsTotal += p.nMulti;

//if ( fVerbose )
//printf( "Stage %2d: Simple = %d\n", i+1, p->nSimple );

if fVerbose {
fmt.Printf("Stage %3d: In = %3d  InP = %3d  Cols = %5d  Multi = %2d  Simple = %2d  Level = %3d\n", i+1, p.nIns, p.nInsP, p.nCols, p.nMulti, p.nSimple, p.Level); }

        // get the codes from the relation (these are not necessarily cubes)
        {
            var c int32;
            for  c = 0; c < p.nCols; c++  {
                p.pbCodes[c] = Cudd_bddAndAbstract( dd, p.bRelation, p.pbCols[c], bVarsCube ); Cudd_Ref( p.pbCodes[c] );
            }
        }

        Cudd_RecursiveDeref( dd, bVarsCube );

        // remap the codes to depend on the topmost varibles of the manager
        // useful as a preparation for the next step
        {
            var pbTemp **DdNode;
            var k, v int32;

            pbTemp = (DdNode **) ABC_ALLOC( char, p.nCols * sizeof(DdNode *) );

            // create the identical permutation
            for  v = 0; v < dd.size; v++  {
                Permute[v] = v;
            }

            // use the topmost variables of the manager 
            // to represent the previous level codes
            for  v = 0; v < p.nMulti; v++  {
                Permute[bCVars[v].index] = dd.invperm[v];
            }

            Extra_bddPermuteArray( dd, p.pbCodes, pbTemp, p.nCols, Permute );
            // the array pbTemp comes already referenced

            // deref the old codes and assign the new ones
            for  k = 0; k < p.nCols; k++  {
                Cudd_RecursiveDeref( dd, p.pbCodes[k] );
                p.pbCodes[k] = pbTemp[k];
            }
            ABC_FREE( pbTemp );
        }
    } 
    if fVerbose {
    fmt.Printf("LUTs: Total = %5d. Final = %5d. Simple = %5d. (%6.2f %%)  ", nLutsTotal, nLutOutputs, nLutsTotal-nLutOutputs, 100.0*(nLutsTotal-nLutOutputs)/nLutsTotal); }
    if fVerbose {
    fmt.Printf("Memory = %6.2f MB\n", 1.0*nLutOutputs*(1<<nLutSize)/(1<<20)); }
//  printf( "\n" );

//fprintf( pTable, "%d ", nLutOutputsOrig );
//fprintf( pTable, "%d ", nLutOutputs );

    if fVerbose {
    fmt.Printf("Pure decomposition time   = %.2f sec\n", float32((Abc_Clock() - clk1 - s_EncodingTime))/float32((int64(time.Second))));
    fmt.Printf("Encoding time             = %.2f sec\n", float32((s_EncodingTime))/float32((int64(time.Second))));
//  printf( "Encoding search time      = %.2f sec\n", (float)(s_EncSearchTime)/(float)(CLOCKS_PER_SEC) );
//  printf( "Encoding compute time     = %.2f sec\n", (float)(s_EncComputeTime)/(float)(CLOCKS_PER_SEC) );
    }


//fprintf( pTable, "%.2f ", (float)(s_ReadingTime)/(float)(CLOCKS_PER_SEC) );
//fprintf( pTable, "%.2f ", (float)(Abc_Clock() - clk1 - s_EncodingTime)/(float)(CLOCKS_PER_SEC) );
//fprintf( pTable, "%.2f ", (float)(s_EncodingTime)/(float)(CLOCKS_PER_SEC) );
//fprintf( pTable, "%.2f ", (float)(s_RemappingTime)/(float)(CLOCKS_PER_SEC) );


    // write LUTs into the BLIF file
    clk1 = Abc_Clock();
    if fCheck {
        var pFile *os.File;
        // start the file
        pFile = os.OpenFile(FileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        fmt.Fprintf(pFile, ".model %s\n", FileName);

        fmt.Fprintf(pFile, ".inputs");
        for  i = 0; i < nNames; i++  {
            fmt.Fprintf(pFile, " %s", pNames[i]);
        }
        fmt.Fprintf(pFile, "\n");
        fmt.Fprintf(pFile, ".outputs F");
        fmt.Fprintf(pFile, "\n");

        // write the DD into the file
        WriteLUTSintoBLIFfile( pFile, dd, pLuts, nLuts, bCVars, pNames, nNames, FileName );

        fmt.Fprintf(pFile, ".end\n");
        _ = (pFile).Close();
        if fVerbose {
        fmt.Printf("Output file writing time  = %.2f sec\n", float32((Abc_Clock() - clk1))/float32((int64(time.Second)))); }
    }


    // updo the LUT cascade
    for  i = 0; i < nLuts; i++  {
        p = pLuts[i];
        for  v = 0; v < p.nCols; v++  {
            Cudd_RecursiveDeref( dd, p.pbCols[v] );
            Cudd_RecursiveDeref( dd, p.pbCodes[v] );
            Cudd_RecursiveDeref( dd, p.paNodes[v] );
        }
        Cudd_RecursiveDeref( dd, p.bRelation );

        ABC_FREE( p.pbCols );
        ABC_FREE( p.pbCodes );
        ABC_FREE( p.paNodes );
        ABC_FREE( p );
    }

    return 1;
}
 
var spStatic_WriteLUTSintoBLIFfile_pNamesLocalIn [MAXINPUTS]string;
var spStatic_WriteLUTSintoBLIFfile_pNamesLocalOut [MAXINPUTS]string;
var spStatic_WriteLUTSintoBLIFfile_Buffer string // was [100]byte in C

func WriteLUTSintoBLIFfile(pFile *os.File, dd *DdManager, pLuts **LUT, nLuts int32, bCVars **DdNode, pNames *string, nNames int32, FileName string) {
    var i, v, o int32;
    
    
    
    var bCube, bCof, bFunc *DdNode;
    var p *LUT;

    // go through all the LUTs
    for  i = 0; i < nLuts; i++  {
        // get the pointer to the LUT
        p = pLuts[i];

        if i == nLuts -1 {
            assert(p.nMulti == 1, "p.nMulti == 1");
        }


        fmt.Fprintf(pFile, "#----------------- LUT #%d ----------------------\n", i);


        // fill in the names for the current LUT

        // write the outputs of the previous LUT
        if i != 0 {
        for  v = 0; v < p.nInsP; v++  {
            spStatic_WriteLUTSintoBLIFfile_Buffer = fmt.Sprintf("LUT%02d_%02d", i-1, v);
            spStatic_WriteLUTSintoBLIFfile_pNamesLocalIn[dd.invperm[v]] = Extra_UtilStrsav( spStatic_WriteLUTSintoBLIFfile_Buffer );
        } }
        // write the primary inputs of the current LUT
        for  v = 0; v < p.nIns - p.nInsP; v++  {
            spStatic_WriteLUTSintoBLIFfile_pNamesLocalIn[dd.invperm[p.Level+v]] = Extra_UtilStrsav( pNames[dd.invperm[p.Level+v]] );
        }
        // write the outputs of the current LUT
        for  v = 0; v < p.nMulti; v++  {
            spStatic_WriteLUTSintoBLIFfile_Buffer = fmt.Sprintf("LUT%02d_%02d", i, v);
            if i != nLuts - 1 {
                spStatic_WriteLUTSintoBLIFfile_pNamesLocalOut[v] = Extra_UtilStrsav( spStatic_WriteLUTSintoBLIFfile_Buffer );
            } else { 
                spStatic_WriteLUTSintoBLIFfile_pNamesLocalOut[v] = Extra_UtilStrsav( "F" ); }
        }


        // write LUT outputs

        // get the prefix
        spStatic_WriteLUTSintoBLIFfile_Buffer = fmt.Sprintf("L%02d_", i);

        // get the cube of encoding variables
        bCube = Extra_bddBitsToCube( dd, (1<<p.nMulti)-1, p.nMulti, bCVars, 1 );   Cudd_Ref( bCube );

        // write each output of the LUT
        for  o = 0; o < p.nMulti; o++  {
            // get the cofactor of this output
            bCof = Cudd_Cofactor( dd, p.bRelation, bCVars[o] );  Cudd_Ref( bCof );
            // quantify the remaining variables to get the function
            bFunc = Cudd_bddExistAbstract( dd, bCof, bCube );     Cudd_Ref( bFunc );
            Cudd_RecursiveDeref( dd, bCof );
            
            // write BLIF
            spStatic_WriteLUTSintoBLIFfile_Buffer = fmt.Sprintf("L%02d_%02d_", i, o);

//          WriteDDintoBLIFfileReorder( dd, pFile, bFunc, pNamesLocalOut[o], Buffer, pNamesLocalIn );
            // does not work well; the advantage is marginal (30%), the run time is huge...

            WriteDDintoBLIFfile( pFile, bFunc, spStatic_WriteLUTSintoBLIFfile_pNamesLocalOut[o], spStatic_WriteLUTSintoBLIFfile_Buffer, spStatic_WriteLUTSintoBLIFfile_pNamesLocalIn );
            Cudd_RecursiveDeref( dd, bFunc );
        }
        Cudd_RecursiveDeref( dd, bCube );

        // clean up the previous local names
        for  v = 0; v < dd.size; v++  {
            if spStatic_WriteLUTSintoBLIFfile_pNamesLocalIn[v] {
                ABC_FREE( spStatic_WriteLUTSintoBLIFfile_pNamesLocalIn[v] ); }
            spStatic_WriteLUTSintoBLIFfile_pNamesLocalIn[v] = nil;
        }
        for  v = 0; v < p.nMulti; v++  {
            ABC_FREE( spStatic_WriteLUTSintoBLIFfile_pNamesLocalOut[v] );
        }
    }
}


////////////////////////////////////////////////////////////////////////
///                           END OF FILE                            ///
////////////////////////////////////////////////////////////////////////




//ABC_NAMESPACE_IMPL_END

