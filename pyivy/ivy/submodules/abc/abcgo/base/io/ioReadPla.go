package io

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

/**CFile****************************************************************

  FileName    [ioReadPla.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedure to read network from file.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioReadPla.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static Abc_Ntk_t * Io_ReadPlaNetwork( Extra_FileReader_t * p, int fZeros, int fBoth, int fOnDc, int fSkipPrepro );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Checks if cubes are distance-1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadPlaPrintCube(p *word, nVars int32) {
    var Symbs [3]byte = [3]byte{'-', '0', '1'}; var v int32;
    for  v = 0; v < nVars; v++  {
        fmt.Printf("%c", Symbs[Abc_TtGetQua(p, v)]);
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Checks if cubes are distance-1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadPlaDistance1(p *word, q *word, nWords int32) int32 {
    var Test word; var c, fFound int32;
    for  c = 0; c < nWords; c++  {
        if p[c] == q[c] {
            continue; }
        if fFound != 0 {
            return 0; }
        // check if the number of 1s is one
//        Test = ((p[c] ^ q[c]) & ((p[c] ^ q[c]) >> 1)) & ABC_CONST(0x5555555555555555); // exactly one 0/1 literal (but may be -/0 or -/1)
        Test = ((p[c] ^ q[c]) | ((p[c] ^ q[c]) >> 1)) & ABC_CONST(0x5555555555555555);
        if !Abc_TtOnlyOneOne(Test) {
            return 0; }
        fFound = 1;
    }
    return fFound;
}
func Io_ReadPlaConsensus(p *word, q *word, nWords int32, piVar *int32) int32 {
    var Test word; var c, fFound int32;
    for  c = 0; c < nWords; c++  {
        if p[c] == q[c] {
            continue; }
        if fFound != 0 {
            return 0; }
        // check if there is exactly one opposite literal (0/1) but may have other diffs (-/0 or -/1)
        Test = ((p[c] ^ q[c]) & ((p[c] ^ q[c]) >> 1)) & ABC_CONST(0x5555555555555555); 
        if !Abc_TtOnlyOneOne(Test) {
            return 0; }
        fFound = 1;
        *piVar = c * 32 + Abc_Tt6FirstBit(Test)/2;
    }
    return fFound;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadPlaMarkIdentical(pCs **word, nCubes int32, nWords int32, vMarks *Vec_Bit_t) {
    var c1, c2 int32;
    Vec_BitFill( vMarks, nCubes, 0 );
    for  c1 = 0; c1 < nCubes; c1++  {
        if !Vec_BitEntry(vMarks, c1) {
            for  c2 = c1 + 1; c2 < nCubes; c2++  {
                if !Vec_BitEntry(vMarks, c2) {
                    if Abc_TtEqual(pCs[c1], pCs[c2], nWords) {
                        Vec_BitWriteEntry( vMarks, c2, 1 ); } }
            } }
    }
}
func Io_ReadPlaMarkContained(pCs **word, nCubes int32, nWords int32, vMarks *Vec_Bit_t) {
    var c1, c2 int32;
    Vec_BitFill( vMarks, nCubes, 0 );
    for  c1 = 0; c1 < nCubes; c1++  {
        if !Vec_BitEntry(vMarks, c1) {
            for  c2 = c1 + 1; c2 < nCubes; c2++  {
                if !Vec_BitEntry(vMarks, c2) {
                    if Abc_TtImply(pCs[c1], pCs[c2], nWords) {
                        Vec_BitWriteEntry( vMarks, c2, 1 );
                    } else if Abc_TtImply(pCs[c2], pCs[c1], nWords) {
                        Vec_BitWriteEntry( vMarks, c1, 1 );
                        break;
                    }
                }
            } }
    }
}
func Io_ReadPlaRemoveMarked(pCs **word, nCubes int32, nWords int32, vMarks *Vec_Bit_t) int32 {
    var c1, c int32;
    for  c1 = c = 0; c1 < nCubes; c1++  {
        if !Vec_BitEntry(vMarks, c1) {
            if c == c1 {
                c++;
            } else {
                Abc_TtCopy( pCs[c++], pCs[c1], nWords, 0 ); }
        }
    }
    return c;
}
func Io_ReadPlaMergeDistance1(pCs **word, nCubes int32, nWords int32, vMarks *Vec_Bit_t) int32 {
    var c1, c2, Res, Counter int32;
    Vec_BitFill( vMarks, nCubes, 0 );
    for  c1 = 0; c1 < nCubes; c1++  {
        if !Vec_BitEntry(vMarks, c1) {
            for  c2 = c1 + 1; c2 < nCubes; c2++  {
                if !Vec_BitEntry(vMarks, c2) {
                    Res = Io_ReadPlaDistance1( pCs[c1], pCs[c2], nWords );
                    if !(Res != 0) {
                        continue; }
                    Abc_TtAnd( pCs[c1], pCs[c1], pCs[c2], nWords, 0 );
                    Vec_BitWriteEntry( vMarks, c2, 1 );
                    Counter++;
                    break;
                }
            } }
    }
    return Counter;
}
func Io_ReadPlaSelfSubsumption(pCs **word, nCubes int32, nWords int32, vMarks *Vec_Bit_t) int32 {
    var c1 int32;
    var c2 int32;
    var Res int32;
    var Counter int32;
    var iVar int32 = -1;
    var Val0 int32;
    var Val1 int32;
    Vec_BitFill( vMarks, nCubes, 0 );
    for  c1 = 0; c1 < nCubes; c1++  {
        if !Vec_BitEntry(vMarks, c1) {
            for  c2 = c1 + 1; c2 < nCubes; c2++  {
                if !Vec_BitEntry(vMarks, c2) {
                    Res = Io_ReadPlaConsensus( pCs[c1], pCs[c2], nWords, &iVar );
                    if !(Res != 0) {
                        continue; }
                    assert(iVar >= 0  && iVar < nWords*32, "iVar >= 0  && iVar < nWords*32");
                    Val0 = Abc_TtGetQua( pCs[c1], iVar );
                    Val1 = Abc_TtGetQua( pCs[c2], iVar );
                    // remove values
                    Abc_TtXorQua( pCs[c1], iVar, Val0 );
                    Abc_TtXorQua( pCs[c2], iVar, Val1 );
                    // check containment
                    if Abc_TtImply(pCs[c1], pCs[c2], nWords) {
                        Abc_TtXorQua( pCs[c1], iVar, Val0 );
                        Vec_BitWriteEntry( vMarks, c2, 1 );
                        Counter++;
                    
                    } else if Abc_TtImply(pCs[c2], pCs[c1], nWords) {
                        Abc_TtXorQua( pCs[c2], iVar, Val1 );
                        Vec_BitWriteEntry( vMarks, c1, 1 );
                        Counter++;
                        break;
                    
                    } else {
                        Abc_TtXorQua( pCs[c1], iVar, Val0 );
                        Abc_TtXorQua( pCs[c2], iVar, Val1 );
                    }

/*
                    printf( "Var = %3d  ", iVar );
                    printf( "Cube0 = %d  ", Abc_TtGetQua(pCs[c1], iVar) );
                    printf( "Cube1 = %d  ", Abc_TtGetQua(pCs[c2], iVar) );
                    printf( "\n" );
                    Io_ReadPlaPrintCube( pCs[c1], 32 * nWords );
                    Io_ReadPlaPrintCube( pCs[c2], 32 * nWords );
                    printf( "\n" );
*/
                    break;
                }
            } }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadPlaCubeSetup(vSop *Vec_Str_t) **word {
    var pSop string = Vec_StrArray( vSop );
    var pCube string;
    var Lit byte;
    var nCubes int32   = Abc_SopGetCubeNum( pSop );
    var nVars int32    = Abc_SopGetVarNum( pSop );
    var nWords int32 = Abc_Bit6WordNum( 2*nVars );
    var c int32;
    var v int32;
    // inlining of macro ABC_ALLOC
    var pCs **word  = ((word * *) malloc(sizeof(word *) * (size_t)(nCubes)));
    // inlining of macro ABC_CALLOC
    pCs[0] = ((word *) calloc((size_t)(nCubes * nWords), sizeof(word)));
    for  c = 1; c < nCubes; c++  {
        pCs[c] = pCs[c-1] + nWords;
    }
    c = 0;
    // inlining of macro Abc_SopForEachCube
    for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
        // inlining of macro Abc_CubeForEachVar
        for  v = 0; (pCube[v] != ' ') && (Lit = pCube[v]); v++  {           
            if Lit == '0' {
                Abc_TtSetBit( pCs[c], Abc_Var2Lit(v,0) );
            } else if Lit == '1' {
                Abc_TtSetBit( pCs[c], Abc_Var2Lit(v,1) ); }
        }
        c++;
    }
    assert(c == nCubes, "c == nCubes");
    return pCs;
}
func Io_ReadPlaCubeSetdown(vSop *Vec_Str_t, pCs **word, nCubes int32, nVars int32) {
    var Symbs [3]byte = [3]byte{'-', '0', '1'}; var c, v int32;
    Vec_StrClear( vSop );
    for  c = 0; c < nCubes; c++  {
        for  v = 0; v < nVars; v++  {
            Vec_StrPush( vSop, Symbs[Abc_TtGetQua(pCs[c], v)] );
        }
        Vec_StrPrintStr( vSop, " 1\n" );
    }
    Vec_StrPush( vSop, 0 );
}
func Io_ReadPlaCubePreprocess(vSop *Vec_Str_t, iCover int32, fVerbose int32) {
    var pCs **word  = Io_ReadPlaCubeSetup( vSop );
    var nCubes int32   = Abc_SopGetCubeNum( Vec_StrArray(vSop) );
    var nVars int32    = Abc_SopGetVarNum( Vec_StrArray(vSop) );
    var nWords int32   = Abc_Bit6WordNum( 2*nVars );
    var nCubesNew, Count, Iter int32;
    var vMarks *Vec_Bit_t  = Vec_BitStart( nCubes );
    if fVerbose != 0 {
        fmt.Printf("Cover %5d : V =%5d  C%d =%5d", iCover, nVars, Iter, nCubes); }

    for 
    {
        Iter++;
        for 
        {
            // remove contained
            Io_ReadPlaMarkContained( pCs, nCubes, nWords, vMarks );
            nCubesNew = Io_ReadPlaRemoveMarked( pCs, nCubes, nWords, vMarks );
            //if ( fVerbose )
            //    printf( "  C =%5d", nCubes - nCubesNew );
            nCubes = nCubesNew;
            // merge distance-1
            Count = Io_ReadPlaMergeDistance1( pCs, nCubes, nWords, vMarks );
        
if !( Count ) { break }
} 
        if fVerbose != 0 {
            fmt.Printf("  C%d =%5d", Iter, nCubes); }
        // try consensus
        //Count = Io_ReadPlaSelfSubsumption( pCs, nCubes, nWords, vMarks );
        if fVerbose != 0 {
            fmt.Printf("%4d", Count); }
    
if !( Count ) { break }
} 

    // translate
    Io_ReadPlaCubeSetdown( vSop, pCs, nCubes, nVars );
    // finalize
    if fVerbose != 0 {
        fmt.Printf("\n"); }
    Vec_BitFree( vMarks );
    pCs[0] = nil;
    pCs = nil;
}

/**Function*************************************************************

  Synopsis    [Reads the network from a PLA file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadPla(pFileName string, fZeros int32, fBoth int32, fOnDc int32, fSkipPrepro int32, fCheck int32) *Abc_Ntk_t {
    var p *Extra_FileReader_t;
    var pNtk *Abc_Ntk_t;

    // start the file
    p = Extra_FileReaderAlloc( pFileName, "#", "\n\r", " \t|" );
//    p = Extra_FileReaderAlloc( pFileName, "", "\n\r", " \t|" );
    if p == nil {
        return nil; }

    // read the network
    pNtk = Io_ReadPlaNetwork( p, fZeros, fBoth, fOnDc, fSkipPrepro );
    Extra_FileReaderFree( p );
    if pNtk == nil {
        return nil; }

    // make sure that everything is okay with the network structure
    if (fCheck != 0) && !Abc_NtkCheckRead( pNtk ) {
        fmt.Printf("Io_ReadPla: The network check has failed.\n");
        Abc_NtkDelete( pNtk );
        return nil;
    }
    return pNtk;
}
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadPlaNetwork(p *Extra_FileReader_t, fZeros int32, fBoth int32, fOnDc int32, fSkipPrepro int32) *Abc_Ntk_t {
    var pProgress *ProgressBar;
    var vTokens *Vec_Ptr_t;
    var pNtk *Abc_Ntk_t;
    var pTermPi, pTermPo, pNode *Abc_Obj_t;
    var ppSops **Vec_Str_t  = nil;
    var Buffer string // was [100]byte in C

    var nInputs int32 = -1;
    var nOutputs int32 = -1;
    var nProducts int32 = -1;
    var pCubeIn, pCubeOut string;
    var i, k, iLine, nDigits, nCubes int32;

    // allocate the empty network
    pNtk = Abc_NtkStartRead( Extra_FileReaderGetFileName(p) );

    // go through the lines of the file
    nCubes = 0;
    pProgress = Extra_ProgressBarStart( os.Stdout, Extra_FileReaderGetFileSize(p) );
    for  vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p)  {
        Extra_ProgressBarUpdate( pProgress, Extra_FileReaderGetCurPosition(p), nil );
        iLine = Extra_FileReaderGetLineNumber( p, 0 );

        // if it is the end of file, quit the loop
        if strncmp( (char *)vTokens.pArray[0], ".e", 2 ) == 0 {
            break; }

        // if it is type directive, ignore it for now
        if strncmp( (char *)vTokens.pArray[0], ".type", 5 ) == 0 {
            continue; }

        // if it is the model name, get the name
        if int32(strings.Compare((char *)vTokens.pArray[0], ".model")) == 0 {
            pNtk.pName = nil;
            pNtk.pName = Extra_UtilStrsav( (char *)vTokens.pArray[1] );
            continue;
        }

        if vTokens.nSize == 1 {
            fmt.Printf("%s (line %d): Wrong number of token.\n", Extra_FileReaderGetFileName(p), iLine);
            Abc_NtkDelete( pNtk );
            Extra_ProgressBarStop( pProgress );
            ppSops = nil;
            return nil;
        }

        if int32(strings.Compare((char *)vTokens.pArray[0], ".i")) == 0 {
            nInputs = strconv.Atoi((char *)vTokens.pArray[1]);
        } else if int32(strings.Compare((char *)vTokens.pArray[0], ".o")) == 0 {
            nOutputs = strconv.Atoi((char *)vTokens.pArray[1]);
        } else if int32(strings.Compare((char *)vTokens.pArray[0], ".p")) == 0 {
            nProducts = strconv.Atoi((char *)vTokens.pArray[1]);
        } else if int32(strings.Compare((char *)vTokens.pArray[0], ".ilb")) == 0 {
            if vTokens.nSize - 1 != nInputs {
                fmt.Printf("Warning: Mismatch between the number of PIs on the .i line (%d) and the number of PIs on the .ilb line (%d).\n", nInputs, vTokens.nSize - 1); }
            for  i = 1; i < vTokens.nSize; i++  {
                Io_ReadCreatePi( pNtk, (char *)vTokens.pArray[i] );
            }
        
        } else if int32(strings.Compare((char *)vTokens.pArray[0], ".ob")) == 0 {
            if vTokens.nSize - 1 != nOutputs {
                fmt.Printf("Warning: Mismatch between the number of POs on the .o line (%d) and the number of POs on the .ob line (%d).\n", nOutputs, vTokens.nSize - 1); }
            for  i = 1; i < vTokens.nSize; i++  {
                Io_ReadCreatePo( pNtk, (char *)vTokens.pArray[i] );
            }
        
        } else {
            // check if the input/output names are given
            if Abc_NtkPiNum(pNtk) == 0 {
                if nInputs == -1 {
                    fmt.Printf("%s: The number of inputs is not specified.\n", Extra_FileReaderGetFileName(p));
                    Abc_NtkDelete( pNtk );
                    Extra_ProgressBarStop( pProgress );
                    ppSops = nil;
                    return nil;
                }
                nDigits = Abc_Base10Log( nInputs );
                for  i = 0; i < nInputs; i++  {
                    Buffer = fmt.Sprintf("x%0*d", nDigits, i);
                    Io_ReadCreatePi( pNtk, Buffer );
                }
            }
            if Abc_NtkPoNum(pNtk) == 0 {
                if nOutputs == -1 {
                    fmt.Printf("%s: The number of outputs is not specified.\n", Extra_FileReaderGetFileName(p));
                    Abc_NtkDelete( pNtk );
                    Extra_ProgressBarStop( pProgress );
                    ppSops = nil;
                    return nil;
                }
                nDigits = Abc_Base10Log( nOutputs );
                for  i = 0; i < nOutputs; i++  {
                    Buffer = fmt.Sprintf("z%0*d", nDigits, i);
                    Io_ReadCreatePo( pNtk, Buffer );
                }
            }
            if Abc_NtkNodeNum(pNtk) == 0 { // first time here
                // create the PO drivers and add them
                // start the SOP covers
                // inlining of macro ABC_ALLOC
                ppSops = ((Vec_Str_t * *) malloc(sizeof(Vec_Str_t *) * (size_t)(nOutputs)));
                // inlining of macro Abc_NtkForEachPo
                for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTermPo) = Abc_NtkPo(pNtk, i)), 1); i++  {
                    ppSops[i] = Vec_StrAlloc( 100 );
                    // create the node
                    pNode = Abc_NtkCreateNode(pNtk);
                    // connect the node to the PO net
                    Abc_ObjAddFanin( Abc_ObjFanin0Ntk(pTermPo), pNode );
                    // connect the node to the PI nets
                    // inlining of macro Abc_NtkForEachPi
                    for  k = 0; (k < Abc_NtkPiNum(pNtk)) && (((pTermPi) = Abc_NtkPi(pNtk, k)), 1); k++  {
                        Abc_ObjAddFanin( pNode, Abc_ObjFanout0Ntk(pTermPi) );
                    }
                }
            }
            // read the cubes
            if vTokens.nSize != 2 {
                fmt.Printf("%s (line %d): Input and output cubes are not specified.\n", Extra_FileReaderGetFileName(p), iLine);
                Abc_NtkDelete( pNtk );
                Extra_ProgressBarStop( pProgress );
                ppSops = nil;
                return nil;
            }
            pCubeIn  = (char *)vTokens.pArray[0];
            pCubeOut = (char *)vTokens.pArray[1];
            if int32(len(pCubeIn)) != nInputs {
                fmt.Printf("%s (line %d): Input cube length (%d) differs from the number of inputs (%d).\n", Extra_FileReaderGetFileName(p), iLine, int32(len(pCubeIn)), nInputs);
                Abc_NtkDelete( pNtk );
                ppSops = nil;
                return nil;
            }
            if int32(len(pCubeOut)) != nOutputs {
                fmt.Printf("%s (line %d): Output cube length (%d) differs from the number of outputs (%d).\n", Extra_FileReaderGetFileName(p), iLine, int32(len(pCubeOut)), nOutputs);
                Abc_NtkDelete( pNtk );
                Extra_ProgressBarStop( pProgress );
                ppSops = nil;
                return nil;
            }
            if fZeros != 0 {
                for  i = 0; i < nOutputs; i++  {
                    if pCubeOut[i] == '0' {
                        Vec_StrPrintStr( ppSops[i], pCubeIn );
                        Vec_StrPrintStr( ppSops[i], " 1\n" );
                    }
                }
            
            } else if fBoth != 0 {
                for  i = 0; i < nOutputs; i++  {
                    if pCubeOut[i] == '0' || pCubeOut[i] == '1' {
                        Vec_StrPrintStr( ppSops[i], pCubeIn );
                        Vec_StrPrintStr( ppSops[i], " 1\n" );
                    }
                }
            
            } else if fOnDc != 0 {
                for  i = 0; i < nOutputs; i++  {
                    if pCubeOut[i] == '-' || pCubeOut[i] == '1' {
                        Vec_StrPrintStr( ppSops[i], pCubeIn );
                        Vec_StrPrintStr( ppSops[i], " 1\n" );
                    }
                }
            
            } else {
                for  i = 0; i < nOutputs; i++  {
                    if pCubeOut[i] == '1' {
                        Vec_StrPrintStr( ppSops[i], pCubeIn );
                        Vec_StrPrintStr( ppSops[i], " 1\n" );
                    }
                }
            }
            nCubes++;
        }
    }
    Extra_ProgressBarStop( pProgress );
    if nProducts != -1 && nCubes != nProducts {
        fmt.Printf("Warning: Mismatch between the number of cubes (%d) and the number on .p line (%d).\n", nCubes, nProducts); }

    // add the SOP covers
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTermPo) = Abc_NtkPo(pNtk, i)), 1); i++  {
        pNode = Abc_ObjFanin0Ntk( Abc_ObjFanin0(pTermPo) );
        if ppSops[i].nSize == 0 {
            Abc_ObjRemoveFanins(pNode);
            pNode.pData = Abc_SopRegister( (Mem_Flex_t *)pNtk.pManFunc, " 0\n" );
            Vec_StrFree( ppSops[i] );
            continue;
        }
        Vec_StrPush( ppSops[i], 0 );
        if !(fSkipPrepro != 0) {
            Io_ReadPlaCubePreprocess( ppSops[i], i, 0 ); }
        pNode.pData = Abc_SopRegister( (Mem_Flex_t *)pNtk.pManFunc, ppSops[i].pArray );
        Vec_StrFree( ppSops[i] );
    }
    ppSops = nil;
    Abc_NtkFinalizeRead( pNtk );
    return pNtk;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_IMPL_END

