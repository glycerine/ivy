package io

import spFmt "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ioWriteAiger.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write binary AIGER format developed by
  Armin Biere, Johannes Kepler University (http://fmv.jku.at/)]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - December 16, 2006.]

  Revision    [$Id: ioWriteAiger.c,v 1.00 2006/12/16 00:00:00 alanmi Exp $]

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


#ifdef _WIN32
#define vsnprintf _vsnprintf
#endif

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/*
    The following is taken from the AIGER format description, 
    which can be found at http://fmv.jku.at/aiger
*/


/*
         The AIGER And-Inverter Graph (AIG) Format Version 20061129
         ----------------------------------------------------------
              Armin Biere, Johannes Kepler University, 2006

  This report describes the AIG file format as used by the AIGER library.
  The purpose of this report is not only to motivate and document the
  format, but also to allow independent implementations of writers and
  readers by giving precise and unambiguous definitions.

  ...

Introduction

  The name AIGER contains as one part the acronym AIG of And-Inverter
  Graphs and also if pronounced in German sounds like the name of the
  'Eiger', a mountain in the Swiss alps.  This choice should emphasize the
  origin of this format. It was first openly discussed at the Alpine
  Verification Meeting 2006 in Ascona as a way to provide a simple, compact
  file format for a model checking competition affiliated to CAV 2007.

  ...

Binary Format Definition

  The binary format is semantically a subset of the ASCII format with a
  slightly different syntax.  The binary format may need to reencode
  literals, but translating a file in binary format into ASCII format and
  then back in to binary format will result in the same file.

  The main differences of the binary format to the ASCII format are as
  follows.  After the header the list of input literals and all the
  current state literals of a latch can be omitted.  Furthermore the
  definitions of the AND gates are binary encoded.  However, the symbol
  table and the comment section are as in the ASCII format.

  The header of an AIGER file in binary format has 'aig' as format
  identifier, but otherwise is identical to the ASCII header.  The standard
  file extension for the binary format is therefore '.aig'. 
  
  A header for the binary format is still in ASCII encoding:

    aig M I L O A

  Constants, variables and literals are handled in the same way as in the
  ASCII format.  The first simplifying restriction is on the variable
  indices of inputs and latches.  The variable indices of inputs come first,
  followed by the pseudo-primary inputs of the latches and then the variable
  indices of all LHS of AND gates:

    input variable indices        1,          2,  ... ,  I
    latch variable indices      I+1,        I+2,  ... ,  (I+L)
    AND variable indices      I+L+1,      I+L+2,  ... ,  (I+L+A) == M

  The corresponding unsigned literals are

    input literals                2,          4,  ... ,  2*I
    latch literals            2*I+2,      2*I+4,  ... ,  2*(I+L)
    AND literals          2*(I+L)+2,  2*(I+L)+4,  ... ,  2*(I+L+A) == 2*M
                    
  All literals have to be defined, and therefore 'M = I + L + A'.  With this
  restriction it becomes possible that the inputs and the current state
  literals of the latches do not have to be listed explicitly.  Therefore,
  after the header only the list of 'L' next state literals follows, one per
  latch on a single line, and then the 'O' outputs, again one per line.

  In the binary format we assume that the AND gates are ordered and respect
  the child parent relation.  AND gates with smaller literals on the LHS
  come first.  Therefore we can assume that the literals on the right-hand
  side of a definition of an AND gate are smaller than the LHS literal.
  Furthermore we can sort the literals on the RHS, such that the larger
  literal comes first.  A definition thus consists of three literals
    
      lhs rhs0 rhs1

  with 'lhs' even and 'lhs > rhs0 >= rhs1'.  Also the variable indices are
  pairwise different to avoid combinational self loops.  Since the LHS
  indices of the definitions are all consecutive (as even integers),
  the binary format does not have to keep 'lhs'.  In addition, we can use
  the order restriction and only write the differences 'delta0' and 'delta1'
  instead of 'rhs0' and 'rhs1', with

      delta0 = lhs - rhs0,  delta1 = rhs0 - rhs1
  
  The differences will all be strictly positive, and in practice often very
  small.  We can take advantage of this fact by the simple little-endian
  encoding of unsigned integers of the next section.  After the binary delta
  encoding of the RHSs of all AND gates, the optional symbol table and
  optional comment section start in the same format as in the ASCII case.

  ...

*/

func Io_ObjMakeLit(Var int32, fCompl int32) uint32 { return (Var << 1) | fCompl;                   }
func Io_ObjAigerNum(pObj *Abc_Obj_t) uint32 { return uint32((ABC_PTRINT_T))pObj.pCopy;  }
func Io_ObjSetAigerNum(pObj *Abc_Obj_t, Num uint32) { pObj.pCopy = (Abc_Obj_t *)(ABC_PTRINT_T)Num;     }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Adds one unsigned AIG edge to the output buffer.]

  Description [This procedure is a slightly modified version of Armin Biere's
  procedure "void encode (FILE * file, unsigned x)" ]
  
  SideEffects [Returns the current writing position.]

  SeeAlso     []

***********************************************************************/
int Io_WriteAigerEncode( unsigned char * pBuffer, int Pos, unsigned x )
{
    unsigned char ch;
    for x & ~0x7f {
        ch = (x & 0x7f) | 0x80;
//        putc (ch, file);
        pBuffer[Pos++] = ch;
        x >>= 7;
    }
    ch = x;
//    putc (ch, file);
    pBuffer[Pos++] = ch;
    return Pos;
}

/**Function*************************************************************

  Synopsis    [Create the array of literals to be written.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteAigerLiterals(pNtk *Abc_Ntk_t) *Vec_Int_t {
    var vLits *Vec_Int_t;
    var pObj, pDriver *Abc_Obj_t;
    var i int32;
    vLits = Vec_IntAlloc( Abc_NtkCoNum(pNtk) );
    // inlining of macro Abc_NtkForEachLatchInput
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)); i++  {                                                  if !(Abc_ObjIsLatch(Abc_NtkBox(pNtk, i)) && (((pObj) = Abc_ObjFanin0(Abc_NtkBox(pNtk, i))), 1)) {} else {
        pDriver = Abc_ObjFanin0(pObj);
        Vec_IntPush( vLits, Io_ObjMakeLit( Io_ObjAigerNum(pDriver), Abc_ObjFaninC0(pObj) ^ (Io_ObjAigerNum(pDriver) == 0) ) );
    }
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        pDriver = Abc_ObjFanin0(pObj);
        Vec_IntPush( vLits, Io_ObjMakeLit( Io_ObjAigerNum(pDriver), Abc_ObjFaninC0(pObj) ^ (Io_ObjAigerNum(pDriver) == 0) ) );
    }
    return vLits;
}

/**Function*************************************************************

  Synopsis    [Creates the binary encoded array of literals.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteEncodeLiterals(vLits *Vec_Int_t) *Vec_Str_t {
    var vBinary *Vec_Str_t;
    var Pos, Lit, LitPrev, Diff, i int32;
    vBinary = Vec_StrAlloc( 2 * Vec_IntSize(vLits) );
    LitPrev = Vec_IntEntry( vLits, 0 );
    Pos = Io_WriteAigerEncode( (unsigned char *)Vec_StrArray(vBinary), Pos, LitPrev ); 
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(vLits)) && (((Lit) = Vec_IntEntry(vLits, i)), 1); i++  {
        Diff = Lit - LitPrev;
        if Lit < LitPrev {
            Diff = -Diff;
        } else {
            Diff = Diff;
        }
        Diff = (Diff << 1) | int32((Lit < LitPrev));
        Pos = Io_WriteAigerEncode( (unsigned char *)Vec_StrArray(vBinary), Pos, Diff );
        LitPrev = Lit;
        if Pos + 10 > vBinary.nCap {
            Vec_StrGrow( vBinary, vBinary.nCap+1 ); }
    }
    vBinary.nSize = Pos;
/*
    // verify
    {
        extern Vec_Int_t * Io_WriteDecodeLiterals( char ** ppPos, int nEntries );
        char * pPos = Vec_StrArray( vBinary );
        Vec_Int_t * vTemp = Io_WriteDecodeLiterals( &pPos, Vec_IntSize(vLits) );
        for ( i = 0; i < Vec_IntSize(vLits); i++ )
        {
            int Entry1 = Vec_IntEntry(vLits,i);
            int Entry2 = Vec_IntEntry(vTemp,i);
            assert( Entry1 == Entry2 );
        }
    }
*/
    return vBinary;
}

/**Function*************************************************************

  Synopsis    [Writes the AIG in the binary AIGER format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteAiger_old(pNtk *Abc_Ntk_t, pFileName string, fWriteSymbols int32, fCompact int32) {
    var pProgress *ProgressBar;
    var pFile *os.File;
    var pObj, pDriver, pLatch *Abc_Obj_t;
    var i, nNodes, nBufferSize, Pos, fExtended int32;
    unsigned char * pBuffer;
    var uLit0, uLit1, uLit uint32;

    fExtended = Abc_NtkConstrNum(pNtk);

    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else
        if !Abc_LatchIsInit0(pObj) {
            if !(fCompact != 0) {
                fExtended = 1;
                break;
            }
            spFmt.Fprintf(os.Stdout, "Io_WriteAiger(): Cannot write AIGER format with non-0 latch init values. Run \"zero\".\n");
            return;
        }
    }

    // start the output stream
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        spFmt.Fprintf(os.Stdout, "Io_WriteAiger(): Cannot open the output file \"%s\".\n", pFileName);
        return;
    }

    // set the node numbers to be used in the output file
    nNodes = 0;
    Io_ObjSetAigerNum( Abc_AigConst1(pNtk), nNodes++ );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Io_ObjSetAigerNum( pObj, nNodes++ );
    }
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        Io_ObjSetAigerNum( pObj, nNodes++ ); }
    }

    // write the header "M I L O A" where M = I + L + A
    spFmt.Fprintf(pFile, "aig%s %d %d %d %d %d", fCompact? "2" : "", Abc_NtkPiNum(pNtk) + Abc_NtkLatchNum(pNtk) + Abc_NtkNodeNum(pNtk), Abc_NtkPiNum(pNtk), Abc_NtkLatchNum(pNtk), fExtended ? 0 : Abc_NtkPoNum(pNtk), Abc_NtkNodeNum(pNtk));
    // write the extended header "B C J F"
    if fExtended != 0 {
        spFmt.Fprintf(pFile, " %d %d", Abc_NtkPoNum(pNtk) - Abc_NtkConstrNum(pNtk), Abc_NtkConstrNum(pNtk)); }
    spFmt.Fprintf(pFile, "\n");

    // if the driver node is a constant, we need to complement the literal below
    // because, in the AIGER format, literal 0/1 is represented as number 0/1
    // while, in ABC, constant 1 node has number 0 and so literal 0/1 will be 1/0

    Abc_NtkInvertConstraints( pNtk );
    if !(fCompact != 0) {
        // write latch drivers
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pLatch) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pLatch) {} else {
            pObj = Abc_ObjFanin0(pLatch);
            pDriver = Abc_ObjFanin0(pObj);
            uLit = Io_ObjMakeLit( Io_ObjAigerNum(pDriver), Abc_ObjFaninC0(pObj) ^ (Io_ObjAigerNum(pDriver) == 0) );
            if Abc_LatchIsInit0(pLatch) {
                spFmt.Fprintf(pFile, "%d\n", uLit);
            } else if Abc_LatchIsInit1(pLatch) {
                spFmt.Fprintf(pFile, "%d 1\n", uLit);
            } else {
                // Both None and DC are written as 'uninitialized' e.g. a free boolean value
                assert(Abc_LatchIsInitNone(pLatch) || Abc_LatchIsInitDc(pLatch), "Abc_LatchIsInitNone(pLatch) || Abc_LatchIsInitDc(pLatch)");
                spFmt.Fprintf(pFile, "%d %d\n", uLit, Io_ObjMakeLit( Io_ObjAigerNum(Abc_ObjFanout0(pLatch)), 0 ));
            }
        }
        }
        // write PO drivers
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
            pDriver = Abc_ObjFanin0(pObj);
            spFmt.Fprintf(pFile, "%d\n", Io_ObjMakeLit( Io_ObjAigerNum(pDriver), Abc_ObjFaninC0(pObj) ^ (Io_ObjAigerNum(pDriver) == 0) ));
        }
    
    } else {
        var vLits *Vec_Int_t  = Io_WriteAigerLiterals( pNtk );
        var vBinary *Vec_Str_t  = Io_WriteEncodeLiterals( vLits );
        fwrite( Vec_StrArray(vBinary), 1, Vec_StrSize(vBinary), pFile );
        Vec_StrFree( vBinary );
        Vec_IntFree( vLits );
    }
    Abc_NtkInvertConstraints( pNtk );

    // write the nodes into the buffer
    Pos = 0;
    nBufferSize = 6 * Abc_NtkNodeNum(pNtk) + 100; // skeptically assuming 3 chars per one AIG edge
    pBuffer = ABC_ALLOC( unsigned char, nBufferSize );
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        uLit  = Io_ObjMakeLit( Io_ObjAigerNum(pObj), 0 );
        uLit0 = Io_ObjMakeLit( Io_ObjAigerNum(Abc_ObjFanin0(pObj)), Abc_ObjFaninC0(pObj) );
        uLit1 = Io_ObjMakeLit( Io_ObjAigerNum(Abc_ObjFanin1(pObj)), Abc_ObjFaninC1(pObj) );
        if uLit0 > uLit1 {
            var Temp uint32  = uLit0;
            uLit0 = uLit1;
            uLit1 = Temp;
        }
        assert(uLit1 < uLit, "uLit1 < uLit");
        Pos = Io_WriteAigerEncode( pBuffer, Pos, uint32((uLit  - uLit1)) );
        Pos = Io_WriteAigerEncode( pBuffer, Pos, uint32((uLit1 - uLit0)) );
        if Pos > nBufferSize - 10 {
            spFmt.Printf("Io_WriteAiger(): AIGER generation has failed because the allocated buffer is too small.\n");
            _ = (pFile).Close();
            return;
        }
    }
    }
    assert(Pos < nBufferSize, "Pos < nBufferSize");
    Extra_ProgressBarStop( pProgress );

    // write the buffer
    fwrite( pBuffer, 1, Pos, pFile );
    ABC_FREE( pBuffer );

    // write the symbol table
    if fWriteSymbols != 0 {
        // write PIs
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
            spFmt.Fprintf(pFile, "i%d %s\n", i, Abc_ObjName(pObj));
        }
        // write latches
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
            spFmt.Fprintf(pFile, "l%d %s\n", i, Abc_ObjName(Abc_ObjFanout0(pObj))); }
        }
        // write POs
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
            if !(fExtended != 0) {
                spFmt.Fprintf(pFile, "o%d %s\n", i, Abc_ObjName(pObj));
            } else if i < Abc_NtkPoNum(pNtk) - Abc_NtkConstrNum(pNtk) {
                spFmt.Fprintf(pFile, "b%d %s\n", i, Abc_ObjName(pObj));
            } else {
                spFmt.Fprintf(pFile, "c%d %s\n", i - (Abc_NtkPoNum(pNtk) - Abc_NtkConstrNum(pNtk)), Abc_ObjName(pObj)); }
        }
    }

    // write the comment
    spFmt.Fprintf(pFile, "c\n");
    if pNtk.pName && len(pNtk.pName) > 0 {
        spFmt.Fprintf(pFile, ".model %s\n", pNtk.pName); }
    spFmt.Fprintf(pFile, "This file was produced by ABC on %s\n", Extra_TimeStamp());
    spFmt.Fprintf(pFile, "For information about AIGER format, refer to %s\n", "http://fmv.jku.at/aiger");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Writes the AIG in the binary AIGER format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteAigerGz(pNtk *Abc_Ntk_t, pFileName string, fWriteSymbols int32) {
    var pProgress *ProgressBar;
    var pFile gzFile;
    var pObj, pDriver, pLatch *Abc_Obj_t;
    var i, nNodes, Pos, nBufferSize, fExtended int32;
    unsigned char * pBuffer;
    var uLit0, uLit1, uLit uint32;

    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // start the output stream
    pFile = gzopen( pFileName, "wb" ); // if pFileName doesn't end in ".gz" then this acts as a passthrough to fopen
    if pFile == nil {
        spFmt.Fprintf(os.Stdout, "Io_WriteAigerGz(): Cannot open the output file \"%s\".\n", pFileName);
        return;
    }

    fExtended = Abc_NtkConstrNum(pNtk);

    // set the node numbers to be used in the output file
    nNodes = 0;
    Io_ObjSetAigerNum( Abc_AigConst1(pNtk), nNodes++ );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Io_ObjSetAigerNum( pObj, nNodes++ );
    }
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        Io_ObjSetAigerNum( pObj, nNodes++ ); }
    }

    // write the header "M I L O A" where M = I + L + A
    gzprintf( pFile, "aig %u %u %u %u %u", 
              Abc_NtkPiNum(pNtk) + Abc_NtkLatchNum(pNtk) + Abc_NtkNodeNum(pNtk), 
              Abc_NtkPiNum(pNtk),
              Abc_NtkLatchNum(pNtk),
              fExtended ? 0 : Abc_NtkPoNum(pNtk),
              Abc_NtkNodeNum(pNtk) );
    // write the extended header "B C J F"
    if fExtended != 0 {
        gzprintf( pFile, " %u %u", Abc_NtkPoNum(pNtk) - Abc_NtkConstrNum(pNtk), Abc_NtkConstrNum(pNtk) ); }
    gzprintf( pFile, "\n" ); 

    // if the driver node is a constant, we need to complement the literal below
    // because, in the AIGER format, literal 0/1 is represented as number 0/1
    // while, in ABC, constant 1 node has number 0 and so literal 0/1 will be 1/0

    // write latch drivers
    Abc_NtkInvertConstraints( pNtk );
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pLatch) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pLatch) {} else {
        pObj = Abc_ObjFanin0(pLatch);
        pDriver = Abc_ObjFanin0(pObj);
        uLit = Io_ObjMakeLit( Io_ObjAigerNum(pDriver), Abc_ObjFaninC0(pObj) ^ (Io_ObjAigerNum(pDriver) == 0) );
        if Abc_LatchIsInit0(pLatch) {
            gzprintf( pFile, "%u\n", uLit );
        } else if Abc_LatchIsInit1(pLatch) {
            gzprintf( pFile, "%u 1\n", uLit );
        } else {
            // Both None and DC are written as 'uninitialized' e.g. a free boolean value
            assert(Abc_LatchIsInitNone(pLatch) || Abc_LatchIsInitDc(pLatch), "Abc_LatchIsInitNone(pLatch) || Abc_LatchIsInitDc(pLatch)");
            gzprintf( pFile, "%u %u\n", uLit, Io_ObjMakeLit( Io_ObjAigerNum(Abc_ObjFanout0(pLatch)), 0 ) );
        }
    }
    }
    // write PO drivers
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        pDriver = Abc_ObjFanin0(pObj);
        gzprintf( pFile, "%u\n", Io_ObjMakeLit( Io_ObjAigerNum(pDriver), Abc_ObjFaninC0(pObj) ^ (Io_ObjAigerNum(pDriver) == 0) ) );
    }
    Abc_NtkInvertConstraints( pNtk );

    // write the nodes into the buffer
    Pos = 0;
    nBufferSize = 6 * Abc_NtkNodeNum(pNtk) + 100; // skeptically assuming 3 chars per one AIG edge
    pBuffer = ABC_ALLOC( unsigned char, nBufferSize );
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        uLit  = Io_ObjMakeLit( Io_ObjAigerNum(pObj), 0 );
        uLit0 = Io_ObjMakeLit( Io_ObjAigerNum(Abc_ObjFanin0(pObj)), Abc_ObjFaninC0(pObj) );
        uLit1 = Io_ObjMakeLit( Io_ObjAigerNum(Abc_ObjFanin1(pObj)), Abc_ObjFaninC1(pObj) );
        if uLit0 > uLit1 {
            var Temp uint32  = uLit0;
            uLit0 = uLit1;
            uLit1 = Temp;
        }
        assert(uLit1 < uLit, "uLit1 < uLit");
        Pos = Io_WriteAigerEncode( pBuffer, Pos, uLit  - uLit1 );
        Pos = Io_WriteAigerEncode( pBuffer, Pos, uLit1 - uLit0 );
        if Pos > nBufferSize - 10 {
            spFmt.Printf("Io_WriteAiger(): AIGER generation has failed because the allocated buffer is too small.\n");
            gzclose( pFile );
            return;
        }
    }
    }
    assert(Pos < nBufferSize, "Pos < nBufferSize");
    Extra_ProgressBarStop( pProgress );

    // write the buffer
    gzwrite(pFile, pBuffer, Pos);
    ABC_FREE( pBuffer );

    // write the symbol table
    if fWriteSymbols != 0 {
        // write PIs
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
            gzprintf( pFile, "i%d %s\n", i, Abc_ObjName(pObj) );
        }
        // write latches
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
            gzprintf( pFile, "l%d %s\n", i, Abc_ObjName(Abc_ObjFanout0(pObj)) ); }
        }
        // write POs
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
            if !(fExtended != 0) {
                gzprintf( pFile, "o%d %s\n", i, Abc_ObjName(pObj) );
            } else if i < Abc_NtkPoNum(pNtk) - Abc_NtkConstrNum(pNtk) {
                gzprintf( pFile, "b%d %s\n", i, Abc_ObjName(pObj) );
            } else {
                gzprintf( pFile, "c%d %s\n", i - (Abc_NtkPoNum(pNtk) - Abc_NtkConstrNum(pNtk)), Abc_ObjName(pObj) ); }
        }
    }

    // write the comment
    gzprintf( pFile, "c\n" );
    if pNtk.pName && len(pNtk.pName) > 0 {
        gzprintf( pFile, ".model %s\n", pNtk.pName ); }
    gzprintf( pFile, "This file was produced by ABC on %s\n", Extra_TimeStamp() );
    gzprintf( pFile, "For information about AIGER format, refer to %s\n", "http://fmv.jku.at/aiger" );
    gzclose( pFile );
}

 
/**Function*************************************************************

  Synopsis    [Procedure to write data into BZ2 file.]

  Description [Based on the vsnprintf() man page.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
typedef struct bz2file {
  FILE   * f;
  BZFILE * b;
  char   * buf;
  int      nBytes;
  int      nBytesMax;
}
 bz2file;

int fprintfBz2Aig( bz2file * b, char * fmt, ... ) {
    if b.b {
        var newBuf string;
        var bzError int32;
        va_list ap;
        for {
            va_start(ap,fmt);
            b.nBytes = vsnprintf(b.buf,b.nBytesMax,fmt,ap);
            va_end(ap);
            if b.nBytes > -1 && b.nBytes < b.nBytesMax {
                break; }
            if b.nBytes > -1 {
                b.nBytesMax = b.nBytes + 1;
            } else {
                b.nBytesMax *= 2; }
            if (newBuf = ABC_REALLOC( char,b.buf,b.nBytesMax )) == nil {
                return -1;
            } else {
                b.buf = newBuf; }
        }
        BZ2_bzWrite( &bzError, b.b, b.buf, b.nBytes );
        if bzError == BZ_IO_ERROR {
            spFmt.Fprintf(os.Stdout, "Ioa_WriteBlif(): I/O error writing to compressed stream.\n");
            return -1;
        }
        return b.nBytes;
    } else {
        var n int32;
        va_list ap;
        va_start(ap,fmt);
        n = vfprintf( b.f, fmt, ap);
        va_end(ap);
        return n;
    }
}

/**Function*************************************************************

  Synopsis    [Writes the AIG in the binary AIGER format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteAiger(pNtk *Abc_Ntk_t, pFileName string, fWriteSymbols int32, fCompact int32, fUnique int32) {
    var pProgress *ProgressBar;
//    FILE * pFile;
    var pObj, pDriver, pLatch *Abc_Obj_t;
    var i, nNodes, nBufferSize, bzError, Pos, fExtended int32;
    unsigned char * pBuffer;
    var uLit0, uLit1, uLit uint32;
    var b bz2file;

    // define unique writing
    if fUnique != 0 {
        fWriteSymbols = 0;
        fCompact = 0;
    }

    fExtended = Abc_NtkConstrNum(pNtk);

    // check that the network is valid
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else
        if !Abc_LatchIsInit0(pObj) {
            if !(fCompact != 0) {
                fExtended = 1;
                break;
            }
            spFmt.Fprintf(os.Stdout, "Io_WriteAiger(): Cannot write AIGER format with non-0 latch init values. Run \"zero\".\n");
            return;
        }
    }

    // write the GZ file
    if !strncmp(pFileName+len(pFileName)-3,".gz",3) {
        Io_WriteAigerGz( pNtk, pFileName, fWriteSymbols );
        return;
    }

    memset(&b,0,sizeof(b));
    b.nBytesMax = (1<<12);
    b.buf = ABC_ALLOC( char,b.nBytesMax );

    // start the output stream
    b.f = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666); 
    if b.f == nil {
        spFmt.Fprintf(os.Stdout, "Ioa_WriteBlif(): Cannot open the output file \"%s\".\n", pFileName);
        ABC_FREE(b.buf);
        return;
    }
    if !strncmp(pFileName+len(pFileName)-4,".bz2",4) {
        b.b = BZ2_bzWriteOpen( &bzError, b.f, 9, 0, 0 );
        if bzError != BZ_OK {
            BZ2_bzWriteClose( &bzError, b.b, 0, nil, nil );
            spFmt.Fprintf(os.Stdout, "Ioa_WriteBlif(): Cannot start compressed stream.\n");
            _ = (b.f).Close();
            ABC_FREE(b.buf);
            return;
        }
    }

    // set the node numbers to be used in the output file
    nNodes = 0;
    Io_ObjSetAigerNum( Abc_AigConst1(pNtk), nNodes++ );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Io_ObjSetAigerNum( pObj, nNodes++ );
    }
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        Io_ObjSetAigerNum( pObj, nNodes++ ); }
    }

    // write the header "M I L O A" where M = I + L + A
    fprintfBz2Aig( &b, "aig%s %u %u %u %u %u", 
        fCompact? "2" : "",
        Abc_NtkPiNum(pNtk) + Abc_NtkLatchNum(pNtk) + Abc_NtkNodeNum(pNtk), 
        Abc_NtkPiNum(pNtk),
        Abc_NtkLatchNum(pNtk),
        fExtended ? 0 : Abc_NtkPoNum(pNtk),
        Abc_NtkNodeNum(pNtk) );
    // write the extended header "B C J F"
    if fExtended != 0 {
        fprintfBz2Aig( &b, " %u %u", Abc_NtkPoNum(pNtk) - Abc_NtkConstrNum(pNtk), Abc_NtkConstrNum(pNtk) ); }
    fprintfBz2Aig( &b, "\n" );

    // if the driver node is a constant, we need to complement the literal below
    // because, in the AIGER format, literal 0/1 is represented as number 0/1
    // while, in ABC, constant 1 node has number 0 and so literal 0/1 will be 1/0

    Abc_NtkInvertConstraints( pNtk );
    if !(fCompact != 0) {
        // write latch drivers
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pLatch) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pLatch) {} else {
            pObj = Abc_ObjFanin0(pLatch);
            pDriver = Abc_ObjFanin0(pObj);
            uLit = Io_ObjMakeLit( Io_ObjAigerNum(pDriver), Abc_ObjFaninC0(pObj) ^ (Io_ObjAigerNum(pDriver) == 0) );
            if Abc_LatchIsInit0(pLatch) {
                fprintfBz2Aig( &b, "%u\n", uLit );
            } else if Abc_LatchIsInit1(pLatch) {
                fprintfBz2Aig( &b, "%u 1\n", uLit );
            } else {
                // Both None and DC are written as 'uninitialized' e.g. a free boolean value
                assert(Abc_LatchIsInitNone(pLatch) || Abc_LatchIsInitDc(pLatch), "Abc_LatchIsInitNone(pLatch) || Abc_LatchIsInitDc(pLatch)");
                fprintfBz2Aig( &b, "%u %u\n", uLit, Io_ObjMakeLit( Io_ObjAigerNum(Abc_ObjFanout0(pLatch)), 0 ) );
            }
        }
        }
        // write PO drivers
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
            pDriver = Abc_ObjFanin0(pObj);
            fprintfBz2Aig( &b, "%u\n", Io_ObjMakeLit( Io_ObjAigerNum(pDriver), Abc_ObjFaninC0(pObj) ^ (Io_ObjAigerNum(pDriver) == 0) ) );
        }
    
    } else {
        var vLits *Vec_Int_t  = Io_WriteAigerLiterals( pNtk );
        var vBinary *Vec_Str_t  = Io_WriteEncodeLiterals( vLits );
        if !b.b {
            fwrite( Vec_StrArray(vBinary), 1, Vec_StrSize(vBinary), b.f );
        } else {
            BZ2_bzWrite( &bzError, b.b, Vec_StrArray(vBinary), Vec_StrSize(vBinary) );
            if bzError == BZ_IO_ERROR {
                spFmt.Fprintf(os.Stdout, "Io_WriteAiger(): I/O error writing to compressed stream.\n");
                _ = (b.f).Close();
                ABC_FREE(b.buf);
                Vec_StrFree( vBinary );
                return;
            }
        }
        Vec_StrFree( vBinary );
        Vec_IntFree( vLits );
    }
    Abc_NtkInvertConstraints( pNtk );

    // write the nodes into the buffer
    Pos = 0;
    nBufferSize = 6 * Abc_NtkNodeNum(pNtk) + 100; // skeptically assuming 3 chars per one AIG edge
    pBuffer = ABC_ALLOC( unsigned char, nBufferSize );
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        uLit  = Io_ObjMakeLit( Io_ObjAigerNum(pObj), 0 );
        uLit0 = Io_ObjMakeLit( Io_ObjAigerNum(Abc_ObjFanin0(pObj)), Abc_ObjFaninC0(pObj) );
        uLit1 = Io_ObjMakeLit( Io_ObjAigerNum(Abc_ObjFanin1(pObj)), Abc_ObjFaninC1(pObj) );
        if uLit0 > uLit1 {
            var Temp uint32  = uLit0;
            uLit0 = uLit1;
            uLit1 = Temp;
        }
        assert(uLit1 < uLit, "uLit1 < uLit");
        Pos = Io_WriteAigerEncode( pBuffer, Pos, uint32((uLit  - uLit1)) );
        Pos = Io_WriteAigerEncode( pBuffer, Pos, uint32((uLit1 - uLit0)) );
        if Pos > nBufferSize - 10 {
            spFmt.Printf("Io_WriteAiger(): AIGER generation has failed because the allocated buffer is too small.\n");
            _ = (b.f).Close();
            ABC_FREE(b.buf);
            Extra_ProgressBarStop( pProgress );
            return;
        }
    }
    }
    assert(Pos < nBufferSize, "Pos < nBufferSize");
    Extra_ProgressBarStop( pProgress );

    // write the buffer
    if !b.b {
        fwrite( pBuffer, 1, Pos, b.f );
    } else {
        BZ2_bzWrite( &bzError, b.b, pBuffer, Pos );
        if bzError == BZ_IO_ERROR {
            spFmt.Fprintf(os.Stdout, "Io_WriteAiger(): I/O error writing to compressed stream.\n");
            _ = (b.f).Close();
            ABC_FREE(b.buf);
            return;
        }
    }
    ABC_FREE( pBuffer );

    // write the symbol table
    if fWriteSymbols != 0 {
        // write PIs
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
            fprintfBz2Aig( &b, "i%d %s\n", i, Abc_ObjName(pObj) );
        }
        // write latches
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
            fprintfBz2Aig( &b, "l%d %s\n", i, Abc_ObjName(Abc_ObjFanout0(pObj)) ); }
        }
        // write POs
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
            if !(fExtended != 0) {
                fprintfBz2Aig( &b, "o%d %s\n", i, Abc_ObjName(pObj) );
            } else if i < Abc_NtkPoNum(pNtk) - Abc_NtkConstrNum(pNtk) {
                fprintfBz2Aig( &b, "b%d %s\n", i, Abc_ObjName(pObj) );
            } else {
                fprintfBz2Aig( &b, "c%d %s\n", i - (Abc_NtkPoNum(pNtk) - Abc_NtkConstrNum(pNtk)), Abc_ObjName(pObj) ); }
        }
    }

    // write the comment
    fprintfBz2Aig( &b, "c" );
    if !(fUnique != 0) {
        if pNtk.pName && len(pNtk.pName) > 0 {
            fprintfBz2Aig( &b, "\n%s%c", pNtk.pName, '\0' ); }
        fprintfBz2Aig( &b, "\nThis file was written by ABC on %s\n", Extra_TimeStamp() );
        fprintfBz2Aig( &b, "For information about AIGER format, refer to %s\n", "http://fmv.jku.at/aiger" );
    }

    // close the file
    if b.b {
        BZ2_bzWriteClose( &bzError, b.b, 0, nil, nil );
        if bzError == BZ_IO_ERROR {
            spFmt.Fprintf(os.Stdout, "Io_WriteAiger(): I/O error closing compressed stream.\n");
            _ = (b.f).Close();
            ABC_FREE(b.buf);
            return;
        }
    }
    _ = (b.f).Close();
    ABC_FREE(b.buf);
}

//ABC_NAMESPACE_IMPL_END

//#include "aig/gia/giaAig.h"
//#include "aig/saig/saig.h"

//ABC_NAMESPACE_IMPL_START

/**Function*************************************************************

  Synopsis    [Writes the AIG in the binary AIGER format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteAigerCex(pCex *Abc_Cex_t, pNtk *Abc_Ntk_t, pG any, pFileName string) {
    extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );
    var pFile *os.File;
    var pAig *Aig_Man_t;
    var pObj, pObj2 *Aig_Obj_t;
    var pGia *Gia_Man_t  = (Gia_Man_t *)pG;
    var k, f, b int32;
    assert(pCex != nil, "pCex != nil");

    // derive AIG
    if pNtk != nil && 
         Abc_NtkPiNum(pNtk)    == pCex.nPis && 
         Abc_NtkLatchNum(pNtk) == pCex.nRegs {
        pAig = Abc_NtkToDar( pNtk, 0, 1 );
    
    } else if pGia != nil && 
         Gia_ManPiNum(pGia)  == pCex.nPis && 
         Gia_ManRegNum(pGia) == pCex.nRegs {
        pAig = Gia_ManToAigSimple( pGia );
    
    } else {
        spFmt.Printf("AIG parameters do not match those of the CEX.\n");
        return;
    }

    // create output file
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    spFmt.Fprintf(pFile, "1\n");
    b = pCex.nRegs;
    for  k = 0; k < pCex.nRegs; k++  {
        spFmt.Fprintf(pFile, "0");
    }
    spFmt.Fprintf(pFile, " ");
    Aig_ManCleanMarkA( pAig );
    Aig_ManConst1(pAig).fMarkA = 1;
    for  f = 0; f <= pCex.iFrame; f++  {
        for  k = 0; k < pCex.nPis; k++  {
            spFmt.Fprintf(pFile, "%d", Abc_InfoHasBit(pCex.pData, b));
            Aig_ManCi( pAig, k ).fMarkA = Abc_InfoHasBit(pCex.pData, b++);
        }
        spFmt.Fprintf(pFile, " ");
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, k)), 1); k++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.fMarkA = (Aig_ObjFanin0(pObj).fMarkA ^ Aig_ObjFaninC0(pObj)) &
                           (Aig_ObjFanin1(pObj).fMarkA ^ Aig_ObjFaninC1(pObj)); }
        }
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, k)), 1); k++  {
            pObj.fMarkA = (Aig_ObjFanin0(pObj).fMarkA ^ Aig_ObjFaninC0(pObj));
        }
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  k = 0; (k < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, k)), 1); k++  {
            spFmt.Fprintf(pFile, "%d", pObj.fMarkA);
        }
        spFmt.Fprintf(pFile, " ");
        // inlining of macro Saig_ManForEachLi
        for  k = 0; (k < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, k+Saig_ManPoNum(pAig))), 1); k++  {
            spFmt.Fprintf(pFile, "%d", pObj.fMarkA);
        }
        spFmt.Fprintf(pFile, "\n");
        if f == pCex.iFrame {
            break; }
        // inlining of macro Saig_ManForEachLi
        for  k = 0; (k < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, k+Saig_ManPoNum(pAig))), 1); k++  {
            spFmt.Fprintf(pFile, "%d", pObj.fMarkA);
        }
        spFmt.Fprintf(pFile, " ");
        // inlining of macro Saig_ManForEachLiLo
        for  k = 0; (k < Saig_ManRegNum(pAig)) && (((pObj) = Saig_ManLi(pAig, k)), 1)            && (((pObj2)=Saig_ManLo(pAig, k)), 1); k++  {
            pObj2.fMarkA = pObj.fMarkA;
        }
    }  
    assert(b == pCex.nBits, "b == pCex.nBits");
    _ = (pFile).Close();
    Aig_ManCleanMarkA( pAig );
    Aig_ManStop( pAig );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

