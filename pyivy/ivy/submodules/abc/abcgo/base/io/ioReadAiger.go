package io

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [ioReadAiger.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to read binary AIGER format developed by
  Armin Biere, Johannes Kepler University (http://fmv.jku.at/)]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - December 16, 2006.]

  Revision    [$Id: ioReadAiger.c,v 1.00 2006/12/16 00:00:00 alanmi Exp $]

***********************************************************************/

// The code in this file is developed in collaboration with Mark Jarvin of Toronto.

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/bzlib/bzlib.h"
//#include "misc/zlib/zlib.h"
//#include "ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Extracts one unsigned AIG edge from the input buffer.]

  Description [This procedure is a slightly modified version of Armin Biere's
  procedure "unsigned decode (FILE * file)". ]
  
  SideEffects [Updates the current reading position.]

  SeeAlso     []

***********************************************************************/
func Io_ReadAigerDecode(ppPos *string) uint32 {
    var x, i uint32;
    unsigned char ch;

//    while ((ch = getnoneofch (file)) & 0x80)
    for (ch = *(*ppPos)++) & 0x80 {
        x |= (ch & 0x7f) << (7 * i++);
    }

    return x | (ch << (7 * i));
}

/**Function*************************************************************

  Synopsis    [Decodes the encoded array of literals.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteDecodeLiterals(ppPos *string, nEntries int32) *Vec_Int_t {
    var vLits *Vec_Int_t;
    var Lit, LitPrev, Diff, i int32;
    vLits = Vec_IntAlloc( nEntries );
    LitPrev = Io_ReadAigerDecode( ppPos );
    Vec_IntPush( vLits, LitPrev );
    for  i = 1; i < nEntries; i++  {
//        Diff = Lit - LitPrev;
//        Diff = (Lit < LitPrev)? -Diff : Diff;
//        Diff = ((2 * Diff) << 1) | (int)(Lit < LitPrev);
        Diff = Io_ReadAigerDecode( ppPos );
        if Diff & 1 {
            Diff = -(Diff >> 1);
        } else {
            Diff = Diff >> 1;
        }
        Lit  = Diff + LitPrev;
        Vec_IntPush( vLits, Lit );
        LitPrev = Lit;
    }
    return vLits;
}


/**Function*************************************************************

  Synopsis    [Reads the file into a character buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
typedef struct buflist {
  char buf[1<<20];
  int nBuf;
  struct buflist * next;
}
 buflist;

func Ioa_ReadLoadFileBz2Aig(pFileName string, pFileSize *int32) string {
    var pFile *os.File;
    var nFileSize int32;
    var pContents string;
    var b *BZFILE;
    var bzError int32;
    struct buflist * pNext;
    var bufHead *buflist = nil;
    var buf *buflist = nil;
    var RetValue int32;

    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Ioa_ReadLoadFileBz2(): The file is unavailable (absent or open).\n");
        return nil;
    }
    b = BZ2_bzReadOpen(&bzError,pFile,0,0,nil,0);
    if bzError != BZ_OK {
        fmt.Printf("Ioa_ReadLoadFileBz2(): BZ2_bzReadOpen() failed with error %d.\n", bzError);
        return nil;
    }
    for {
        if !(bufHead != nil) {
            buf = bufHead = new(buflist);
        } else {
            buf = buf.next = new(buflist); }
        nFileSize += buf.nBuf = BZ2_bzRead(&bzError,b,buf.buf,1<<20);
        buf.next = nil;
    
if !(bzError == BZ_OK) { break }
} 
    if bzError == BZ_STREAM_END {
        // we're okay
        var p string;
        var nBytes int32;
        BZ2_bzReadClose(&bzError,b);
        p = pContents = ABC_ALLOC( char, nFileSize + 10 );
        buf = bufHead;
        for {
            memcpy(p+nBytes,buf.buf,(size_t)buf.nBuf);
            nBytes += buf.nBuf;
//        } while((buf = buf->next));
            pNext = buf.next;
            ABC_FREE( buf );
        
if !((buf = pNext)) { break }
} 
    } else if bzError == BZ_DATA_ERROR_MAGIC {
        // not a BZIP2 file
        BZ2_bzReadClose(&bzError,b);
        fseek( pFile, 0, SEEK_END );
        nFileSize = ftell( pFile );
        if nFileSize == 0 {
            fmt.Printf("Ioa_ReadLoadFileBz2(): The file is empty.\n");
            return nil;
        }
        pContents = ABC_ALLOC( char, nFileSize + 10 );
        rewind( pFile );
        RetValue = fread( pContents, nFileSize, 1, pFile );
    } else { 
        // Some other error.
        fmt.Printf("Ioa_ReadLoadFileBz2(): Unable to read the compressed BLIF.\n");
        return nil;
    }
    _ = (pFile).Close();
    // finish off the file with the spare .end line
    // some benchmarks suddenly break off without this line
//    strcpy( pContents + nFileSize, "\n.end\n" );
    *pFileSize = nFileSize;
    return pContents;
}

/**Function*************************************************************

  Synopsis    [Reads the file into a character buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ioa_ReadLoadFileGzAig(pFileName string, pFileSize *int32) string {
    int READ_BLOCK_SIZE = 100000;
    var pFile gzFile;
    var pContents string;
    var amtRead int32;
    var readBlock int32;
    var nFileSize int32 = READ_BLOCK_SIZE;
    pFile = gzopen( pFileName, "rb" ); // if pFileName doesn't end in ".gz" then this acts as a passthrough to fopen
    pContents = ABC_ALLOC( char, nFileSize );        
    readBlock = 0;
    for (amtRead = gzread(pFile, pContents + readBlock * READ_BLOCK_SIZE, READ_BLOCK_SIZE)) == READ_BLOCK_SIZE {
        //printf("%d: read %d bytes\n", readBlock, amtRead);
        nFileSize += READ_BLOCK_SIZE;
        pContents = ABC_REALLOC(char, pContents, nFileSize);
        ++readBlock;
    }
    //printf("%d: read %d bytes\n", readBlock, amtRead);
    assert(amtRead != -1, "amtRead != -1"); // indicates a zlib error
    nFileSize -= (READ_BLOCK_SIZE - amtRead);
    gzclose(pFile);
    *pFileSize = nFileSize;
    return pContents;
}


/**Function*************************************************************

  Synopsis    [Reads the AIG in the binary AIGER format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadAiger(pFileName string, fCheck int32) *Abc_Ntk_t {
    var pProgress *ProgressBar;
    var pFile *os.File;
    var vNodes, vTerms *Vec_Ptr_t;
    var vLits *Vec_Int_t  = nil;
    var pObj, pNode0, pNode1 *Abc_Obj_t;
    var pNtkNew *Abc_Ntk_t;
    var nTotal, nInputs, nOutputs, nLatches, nAnds int32;
    var nBad, nConstr, nJust, nFair int32;
    var nFileSize int32 = -1;
    var iTerm int32;
    var nDigits int32;
    var i int32;
    var pContents string;
    var pDrivers string = nil;
    var pSymbols string;
    var pCur string;
    var pName string;
    var pType string;
    var uLit0, uLit1, uLit uint32;
    var RetValue int32;

    // read the file into the buffer
    if !strncmp(pFileName+len(pFileName)-4,".bz2",4) {
        pContents = Ioa_ReadLoadFileBz2Aig( pFileName, &nFileSize );
    } else if !strncmp(pFileName+len(pFileName)-3,".gz",3) {
        pContents = Ioa_ReadLoadFileGzAig( pFileName, &nFileSize );
    } else {
//        pContents = Ioa_ReadLoadFile( pFileName );
        nFileSize = Extra_FileSize( pFileName );
        pFile = os.Open(pFileName);
        pContents = ABC_ALLOC( char, nFileSize );
        RetValue = fread( pContents, nFileSize, 1, pFile );
        _ = (pFile).Close();
    }


    // check if the input file format is correct
    if strncmp(pContents, "aig", 3) != 0 || (pContents[3] != ' ' && pContents[3] != '2') {
        fmt.Fprintf(os.Stdout, "Wrong input file format.\n");
        ABC_FREE( pContents );
        return nil;
    }

    // read the parameters (M I L O A + B C J F)
    pCur = pContents;         for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of objects
    nTotal = strconv.Atoi(pCur);    for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of inputs
    nInputs = strconv.Atoi(pCur);   for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of latches
    nLatches = strconv.Atoi(pCur);  for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of outputs
    nOutputs = strconv.Atoi(pCur);  for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of nodes
    nAnds = strconv.Atoi(pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
    if *pCur == ' ' {
//        assert( nOutputs == 0 );
        // read the number of properties
        pCur++;
        nBad = strconv.Atoi(pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
        nOutputs += nBad;
    }
    if *pCur == ' ' {
        // read the number of properties
        pCur++;
        nConstr = strconv.Atoi(pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
        nOutputs += nConstr;
    }
    if *pCur == ' ' {
        // read the number of properties
        pCur++;
        nJust = strconv.Atoi(pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
        nOutputs += nJust;
    }
    if *pCur == ' ' {
        // read the number of properties
        pCur++;
        nFair = strconv.Atoi(pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
        nOutputs += nFair;
    }
    if *pCur != '\n' {
        fmt.Fprintf(os.Stdout, "The parameter line is in a wrong format.\n");
        ABC_FREE( pContents );
        return nil;
    }
    pCur++;

    // check the parameters
    if nTotal != nInputs + nLatches + nAnds {
        fmt.Fprintf(os.Stdout, "The number of objects does not match.\n");
        ABC_FREE( pContents );
        return nil;
    }
    if nJust || (nFair != 0) {
        fmt.Fprintf(os.Stdout, "Reading AIGER files with liveness properties is currently not supported.\n");
        ABC_FREE( pContents );
        return nil;
    }

    if nConstr {
        if nConstr == 1 {
            fmt.Fprintf(os.Stdout, "Warning: The last output is interpreted as a constraint.\n");
        } else {
            fmt.Fprintf(os.Stdout, "Warning: The last %d outputs are interpreted as constraints.\n", nConstr); }
    }

    // allocate the empty AIG
    pNtkNew = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    pName = Extra_FileNameGeneric( pFileName );
    pNtkNew.pName = Extra_UtilStrsav( pName );
    pNtkNew.pSpec = Extra_UtilStrsav( pFileName );
    ABC_FREE( pName );
    pNtkNew.nConstrs = nConstr;

    // prepare the array of nodes
    vNodes = Vec_PtrAlloc( 1 + nInputs + nLatches + nAnds );
    Vec_PtrPush( vNodes, Abc_ObjNot( Abc_AigConst1(pNtkNew) ) );

    // create the PIs
    for  i = 0; i < nInputs; i++  {
        pObj = Abc_NtkCreatePi(pNtkNew);    
        Vec_PtrPush( vNodes, pObj );
    }
    // create the POs
    for  i = 0; i < nOutputs; i++  {
        pObj = Abc_NtkCreatePo(pNtkNew);   
    }
    // create the latches
    nDigits = Abc_Base10Log( nLatches );
    for  i = 0; i < nLatches; i++  {
        pObj = Abc_NtkCreateLatch(pNtkNew);
        Abc_LatchSetInit0( pObj );
        pNode0 = Abc_NtkCreateBi(pNtkNew);
        pNode1 = Abc_NtkCreateBo(pNtkNew);
        Abc_ObjAddFanin( pObj, pNode0 );
        Abc_ObjAddFanin( pNode1, pObj );
        Vec_PtrPush( vNodes, pNode1 );
        // assign names to latch and its input
//        Abc_ObjAssignName( pObj, Abc_ObjNameDummy("_L", i, nDigits), NULL );
//        printf( "Creating latch %s with input %d and output %d.\n", Abc_ObjName(pObj), pNode0->Id, pNode1->Id );
    } 
    

    if pContents[3] == ' ' { // standard AIGER
    
        // remember the beginning of latch/PO literals
        pDrivers = pCur;
        // scroll to the beginning of the binary data
        for  i = 0; i < nLatches + nOutputs;  {
            if *pCur++ == '\n' {
                i++; }
        }
    
    } else { // modified AIGER
     
        vLits = Io_WriteDecodeLiterals( &pCur, nLatches + nOutputs );
    }

    // create the AND gates
    pProgress = Extra_ProgressBarStart( os.Stdout, nAnds );
    for  i = 0; i < nAnds; i++  {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        uLit = ((i + 1 + nInputs + nLatches) << 1);
        uLit1 = uLit  - Io_ReadAigerDecode( &pCur );
        uLit0 = uLit1 - Io_ReadAigerDecode( &pCur );
//        assert( uLit1 > uLit0 );
        pNode0 = Abc_ObjNotCond( (Abc_Obj_t *)Vec_PtrEntry(vNodes, uLit0 >> 1), uLit0 & 1 );
        pNode1 = Abc_ObjNotCond( (Abc_Obj_t *)Vec_PtrEntry(vNodes, uLit1 >> 1), uLit1 & 1 );
        assert(Vec_PtrSize(vNodes) == i + 1 + nInputs + nLatches, "Vec_PtrSize(vNodes) == i + 1 + nInputs + nLatches");
        Vec_PtrPush( vNodes, Abc_AigAnd((Abc_Aig_t *)pNtkNew.pManFunc, pNode0, pNode1) );
    }
    Extra_ProgressBarStop( pProgress );

    // remember the place where symbols begin
    pSymbols = pCur;

    // read the latch driver literals
    pCur = pDrivers;
    if pContents[3] == ' ' { // standard AIGER
    
        // inlining of macro Abc_NtkForEachLatchInput
        for  i = 0; (i < Vec_PtrSize((pNtkNew).vBoxes)); i++  {                                                  if !(Abc_ObjIsLatch(Abc_NtkBox(pNtkNew, i)) && (((pObj) = Abc_ObjFanin0(Abc_NtkBox(pNtkNew, i))), 1)) {} else {
            uLit0 = strconv.Atoi(pCur);  for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
            if *pCur == ' ' { // read initial value
            
                var Init int32;
                pCur++;
                Init = strconv.Atoi(pCur);
                if Init == 0 {
                    Abc_LatchSetInit0( Abc_NtkBox(pNtkNew, i) );
                } else if Init == 1 {
                    Abc_LatchSetInit1( Abc_NtkBox(pNtkNew, i) );
                } else {
                    assert(Init == Abc_Var2Lit(1+Abc_NtkPiNum(pNtkNew)+i, 0), "Init == Abc_Var2Lit(1+Abc_NtkPiNum(pNtkNew)+i, 0)"); 
                    // unitialized value of the latch is the latch literal according to http://fmv.jku.at/hwmcc11/beyond1.pdf
                    Abc_LatchSetInitDc( Abc_NtkBox(pNtkNew, i) );
                }
                for  *pCur != ' ' && *pCur != '\n'  { pCur++; }
            }
            if *pCur != '\n' {
                fmt.Fprintf(os.Stdout, "The initial value of latch number %d is not recongnized.\n", i);
                return nil;
            }
            pCur++;

            pNode0 = Abc_ObjNotCond( (Abc_Obj_t *)Vec_PtrEntry(vNodes, uLit0 >> 1), (uLit0 & 1) );//^ (uLit0 < 2) );
            Abc_ObjAddFanin( pObj, pNode0 );
        }
        }
        // read the PO driver literals
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtkNew)) && (((pObj) = Abc_NtkPo(pNtkNew, i)), 1); i++  {
            uLit0 = strconv.Atoi(pCur);  for  *pCur++ != '\n'  {; }
            pNode0 = Abc_ObjNotCond( (Abc_Obj_t *)Vec_PtrEntry(vNodes, uLit0 >> 1), (uLit0 & 1) );//^ (uLit0 < 2) );
            Abc_ObjAddFanin( pObj, pNode0 );
        }
    
    } else {
        // read the latch driver literals
        // inlining of macro Abc_NtkForEachLatchInput
        for  i = 0; (i < Vec_PtrSize((pNtkNew).vBoxes)); i++  {                                                  if !(Abc_ObjIsLatch(Abc_NtkBox(pNtkNew, i)) && (((pObj) = Abc_ObjFanin0(Abc_NtkBox(pNtkNew, i))), 1)) {} else {
            uLit0 = Vec_IntEntry( vLits, i );
            pNode0 = Abc_ObjNotCond( (Abc_Obj_t *)Vec_PtrEntry(vNodes, uLit0 >> 1), (uLit0 & 1) );
            Abc_ObjAddFanin( pObj, pNode0 );
        }
        }
        // read the PO driver literals
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtkNew)) && (((pObj) = Abc_NtkPo(pNtkNew, i)), 1); i++  {
            uLit0 = Vec_IntEntry( vLits, i+Abc_NtkLatchNum(pNtkNew) );
            pNode0 = Abc_ObjNotCond( (Abc_Obj_t *)Vec_PtrEntry(vNodes, uLit0 >> 1), (uLit0 & 1) );
            Abc_ObjAddFanin( pObj, pNode0 );
        }
        Vec_IntFree( vLits );
    }
 
    // read the names if present
    pCur = pSymbols;
    if pCur < pContents + nFileSize && *pCur != 'c' {
        var Counter int32;
        for  pCur < pContents + nFileSize && *pCur != 'c'  {
            // get the terminal type
            pType = pCur;
            if *pCur == 'i' {
                vTerms = pNtkNew.vPis;
            } else if *pCur == 'l' {
                vTerms = pNtkNew.vBoxes;
            } else if *pCur == 'o' || *pCur == 'b' || *pCur == 'c' || *pCur == 'j' || *pCur == 'f' {
                vTerms = pNtkNew.vPos;
            } else {
//                fprintf( stdout, "Wrong terminal type.\n" );
                return nil;
            }
            // get the terminal number
            iTerm = strconv.Atoi(++pCur);  for  *pCur++ != ' '  {; }
            // get the node
            if iTerm >= Vec_PtrSize(vTerms) {
                fmt.Fprintf(os.Stdout, "The number of terminal is out of bound.\n");
                return nil;
            }
            pObj = (Abc_Obj_t *)Vec_PtrEntry( vTerms, iTerm );
            if *pType == 'l' {
                pObj = Abc_ObjFanout0(pObj); }
            // assign the name
            pName = pCur;          for  *pCur++ != '\n'  {; }
            // assign this name 
            *(pCur-1) = 0;
            Abc_ObjAssignName( pObj, pName, nil );
            if *pType == 'l' {
                Abc_ObjAssignName( Abc_ObjFanin0(pObj), Abc_ObjName(pObj), "L" );
                Abc_ObjAssignName( Abc_ObjFanin0(Abc_ObjFanin0(pObj)), Abc_ObjName(pObj), "_in" );
            }
            // mark the node as named
            pObj.pCopy = (Abc_Obj_t *)Abc_ObjName(pObj);
        } 

        // assign the remaining names
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtkNew)) && (((pObj) = Abc_NtkPi(pNtkNew, i)), 1); i++  {
            if pObj.pCopy { continue; }
            Abc_ObjAssignName( pObj, Abc_ObjName(pObj), nil );
            Counter++;
        }
        // inlining of macro Abc_NtkForEachLatchOutput
        for  i = 0; (i < Vec_PtrSize((pNtkNew).vBoxes)); i++  {                                                  if !(Abc_ObjIsLatch(Abc_NtkBox(pNtkNew, i)) && (((pObj) = Abc_ObjFanout0(Abc_NtkBox(pNtkNew, i))), 1)) {} else {
            if pObj.pCopy { continue; }
            Abc_ObjAssignName( pObj, Abc_ObjName(pObj), nil );
            Abc_ObjAssignName( Abc_ObjFanin0(pObj), Abc_ObjName(pObj), "L" );
            Abc_ObjAssignName( Abc_ObjFanin0(Abc_ObjFanin0(pObj)), Abc_ObjName(pObj), "_in" );
            Counter++;
        }
        }
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtkNew)) && (((pObj) = Abc_NtkPo(pNtkNew, i)), 1); i++  {
            if pObj.pCopy { continue; }
            Abc_ObjAssignName( pObj, Abc_ObjName(pObj), nil );
            Counter++;
        }
//        if ( Counter )
//            printf( "Io_ReadAiger(): Added %d default names for nameless I/O/register objects.\n", Counter );
    
    } else {
//        printf( "Io_ReadAiger(): I/O/register names are not given. Generating short names.\n" );
        Abc_NtkShortNames( pNtkNew );
    }

    // read the name of the model if given
    pCur = pSymbols;
    if pCur + 1 < pContents + nFileSize && *pCur == 'c' {
        pCur++;
        if *pCur == 'n' {
            pCur++;
            // read model name
            if len(pCur) > 0 {
                ABC_FREE( pNtkNew.pName );
                pNtkNew.pName = Extra_UtilStrsav( pCur );
            }
        }
    }

    // skipping the comments
    ABC_FREE( pContents );
    Vec_PtrFree( vNodes );

    // remove the extra nodes
    Abc_AigCleanup( (Abc_Aig_t *)pNtkNew.pManFunc );

    // update polarity of the additional outputs
    if nBad || nConstr || nJust || (nFair != 0) {
        Abc_NtkInvertConstraints( pNtkNew ); }

    // check the result
    if (fCheck != 0) && !Abc_NtkCheckRead( pNtkNew ) {
        fmt.Printf("Io_ReadAiger: The network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

