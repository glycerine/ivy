package ioa

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ioaWriteAiger.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write binary AIGER format developed by
  Armin Biere, Johannes Kepler University (http://fmv.jku.at/)]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - December 16, 2006.]

  Revision    [$Id: ioaWriteAiger.c,v 1.00 2006/12/16 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioa.h"

//ABC_NAMESPACE_IMPL_START


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

func Ioa_ObjMakeLit(Var int32, fCompl int32) int32 { return (Var << 1) | fCompl;  }
func Ioa_ObjAigerNum(pObj *Aig_Obj_t) int32 { return pObj.iData;          }
func Ioa_ObjSetAigerNum(pObj *Aig_Obj_t, Num uint32) { pObj.iData = Num;           }

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
int Ioa_WriteAigerEncode( unsigned char * pBuffer, int Pos, unsigned x )
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

  Synopsis    [Adds one unsigned AIG edge to the output buffer.]

  Description [This procedure is a slightly modified version of Armin Biere's
  procedure "void encode (FILE * file, unsigned x)" ]
  
  SideEffects [Returns the current writing position.]

  SeeAlso     []

***********************************************************************/
func Ioa_WriteAigerEncodeStr(vStr *Vec_Str_t, x uint32) {
    unsigned char ch;
    for x & ~0x7f {
        ch = (x & 0x7f) | 0x80;
//        putc (ch, file);
//        pBuffer[Pos++] = ch;
        Vec_StrPush( vStr, ch );
        x >>= 7;
    }
    ch = x;
//    putc (ch, file);
//    pBuffer[Pos++] = ch;
    Vec_StrPush( vStr, ch );
}

/**Function*************************************************************

  Synopsis    [Create the array of literals to be written.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ioa_WriteAigerLiterals(pMan *Aig_Man_t) *Vec_Int_t {
    var vLits *Vec_Int_t;
    var pObj, pDriver *Aig_Obj_t;
    var i int32;
    vLits = Vec_IntAlloc( Aig_ManCoNum(pMan) );
    // inlining of macro Aig_ManForEachLiSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCoNum(pMan)-Aig_ManRegNum(pMan); (i < Vec_PtrSize(pMan.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
        pDriver = Aig_ObjFanin0(pObj);
        Vec_IntPush( vLits, Ioa_ObjMakeLit( Ioa_ObjAigerNum(pDriver), Aig_ObjFaninC0(pObj) ^ (Ioa_ObjAigerNum(pDriver) == 0) ) );
    }
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(pMan)-Aig_ManRegNum(pMan)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
        pDriver = Aig_ObjFanin0(pObj);
        Vec_IntPush( vLits, Ioa_ObjMakeLit( Ioa_ObjAigerNum(pDriver), Aig_ObjFaninC0(pObj) ^ (Ioa_ObjAigerNum(pDriver) == 0) ) );
    }
    return vLits;
}

/**Function*************************************************************

  Synopsis    [Creates the binary encoded array of literals.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ioa_WriteEncodeLiterals(vLits *Vec_Int_t) *Vec_Str_t {
    var vBinary *Vec_Str_t;
    var Pos, Lit, LitPrev, Diff, i int32;
    vBinary = Vec_StrAlloc( 2 * Vec_IntSize(vLits) );
    LitPrev = Vec_IntEntry( vLits, 0 );
    Pos = Ioa_WriteAigerEncode( (unsigned char *)Vec_StrArray(vBinary), Pos, LitPrev ); 
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(vLits)) && (((Lit) = Vec_IntEntry(vLits, i)), 1); i++  {
        Diff = Lit - LitPrev;
        if Lit < LitPrev {
            Diff = -Diff;
        } else {
            Diff = Diff;
        }
        Diff = (Diff << 1) | int32((Lit < LitPrev));
        Pos = Ioa_WriteAigerEncode( (unsigned char *)Vec_StrArray(vBinary), Pos, Diff );
        LitPrev = Lit;
        if Pos + 10 > vBinary.nCap {
            Vec_StrGrow( vBinary, vBinary.nCap+1 ); }
    }
    vBinary.nSize = Pos;
/*
    // verify
    {
        extern Vec_Int_t * Ioa_WriteDecodeLiterals( char ** ppPos, int nEntries );
        char * pPos = Vec_StrArray( vBinary );
        Vec_Int_t * vTemp = Ioa_WriteDecodeLiterals( &pPos, Vec_IntSize(vLits) );
        for ( i = 0; i < Vec_IntSize(vLits); i++ )
        {
            int Entry1 = Vec_IntEntry(vLits,i);
            int Entry2 = Vec_IntEntry(vTemp,i);
            assert( Entry1 == Entry2 );
        }
        Vec_IntFree( vTemp );
    }
*/
    return vBinary;
}

/**Function*************************************************************

  Synopsis    [Writes the AIG in into the memory buffer.]

  Description [The resulting buffer constains the AIG in AIGER format. 
  The returned size (pnSize) gives the number of bytes in the buffer. 
  The resulting buffer should be deallocated by the user.]
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ioa_WriteAigerIntoMemoryStr(pMan *Aig_Man_t) *Vec_Str_t {
    var vBuffer *Vec_Str_t;
    var pObj, pDriver *Aig_Obj_t;
    var nNodes, i, uLit, uLit0, uLit1 int32; 
    // set the node numbers to be used in the output file
    nNodes = 0;
    Ioa_ObjSetAigerNum( Aig_ManConst1(pMan), nNodes++ );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCis, i)), 1); i++  {
        Ioa_ObjSetAigerNum( pObj, nNodes++ );
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Ioa_ObjSetAigerNum( pObj, nNodes++ ); }
    }

    // write the header "M I L O A" where M = I + L + A
/*
    fprintf( pFile, "aig%s %u %u %u %u %u\n", 
        fCompact? "2" : "",
        Aig_ManCiNum(pMan) + Aig_ManNodeNum(pMan), 
        Aig_ManCiNum(pMan) - Aig_ManRegNum(pMan),
        Aig_ManRegNum(pMan),
        Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan),
        Aig_ManNodeNum(pMan) );
*/
    vBuffer = Vec_StrAlloc( 3*Aig_ManObjNum(pMan) );
    Vec_StrPrintStr( vBuffer, "aig " );
    Vec_StrPrintNum( vBuffer, Aig_ManCiNum(pMan) + Aig_ManNodeNum(pMan) );
    Vec_StrPrintStr( vBuffer, " " );
    Vec_StrPrintNum( vBuffer, Aig_ManCiNum(pMan) - Aig_ManRegNum(pMan) );
    Vec_StrPrintStr( vBuffer, " " );
    Vec_StrPrintNum( vBuffer, Aig_ManRegNum(pMan) );
    Vec_StrPrintStr( vBuffer, " " );
    Vec_StrPrintNum( vBuffer, Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan) );
    Vec_StrPrintStr( vBuffer, " " );
    Vec_StrPrintNum( vBuffer, Aig_ManNodeNum(pMan) );
    Vec_StrPrintStr( vBuffer, "\n" );

    // write latch drivers
    // inlining of macro Aig_ManForEachLiSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCoNum(pMan)-Aig_ManRegNum(pMan); (i < Vec_PtrSize(pMan.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
        pDriver = Aig_ObjFanin0(pObj);
        uLit    = Ioa_ObjMakeLit( Ioa_ObjAigerNum(pDriver), Aig_ObjFaninC0(pObj) ^ (Ioa_ObjAigerNum(pDriver) == 0) );
//        fprintf( pFile, "%u\n", uLit );
        Vec_StrPrintNum( vBuffer, uLit );
        Vec_StrPrintStr( vBuffer, "\n" );
    }

    // write PO drivers
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(pMan)-Aig_ManRegNum(pMan)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
        pDriver = Aig_ObjFanin0(pObj);
        uLit    = Ioa_ObjMakeLit( Ioa_ObjAigerNum(pDriver), Aig_ObjFaninC0(pObj) ^ (Ioa_ObjAigerNum(pDriver) == 0) );
//        fprintf( pFile, "%u\n", uLit );
        Vec_StrPrintNum( vBuffer, uLit );
        Vec_StrPrintStr( vBuffer, "\n" );
    }
    // write the nodes into the buffer
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        uLit  = Ioa_ObjMakeLit( Ioa_ObjAigerNum(pObj), 0 );
        uLit0 = Ioa_ObjMakeLit( Ioa_ObjAigerNum(Aig_ObjFanin0(pObj)), Aig_ObjFaninC0(pObj) );
        uLit1 = Ioa_ObjMakeLit( Ioa_ObjAigerNum(Aig_ObjFanin1(pObj)), Aig_ObjFaninC1(pObj) );
        assert(uLit0 != uLit1, "uLit0 != uLit1");
        if uLit0 > uLit1 {
            var Temp int32  = uLit0;
            uLit0 = uLit1;
            uLit1 = Temp;
        }
//        Pos = Ioa_WriteAigerEncode( pBuffer, Pos, uLit  - uLit1 );
//        Pos = Ioa_WriteAigerEncode( pBuffer, Pos, uLit1 - uLit0 );
        Ioa_WriteAigerEncodeStr( vBuffer, uLit  - uLit1 );
        Ioa_WriteAigerEncodeStr( vBuffer, uLit1 - uLit0 );
    }
    }
    Vec_StrPrintStr( vBuffer, "c" );
    return vBuffer;
}

/**Function*************************************************************

  Synopsis    [Writes the AIG in into the memory buffer.]

  Description [The resulting buffer constains the AIG in AIGER format. 
  The returned size (pnSize) gives the number of bytes in the buffer. 
  The resulting buffer should be deallocated by the user.]
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ioa_WriteAigerIntoMemory(pMan *Aig_Man_t, pnSize *int32) string {
    var pBuffer string;
    var vBuffer *Vec_Str_t;
    vBuffer = Ioa_WriteAigerIntoMemoryStr( pMan );
    if pMan.pName {
        Vec_StrPrintStr( vBuffer, "n" );
        Vec_StrPrintStr( vBuffer, pMan.pName );
        Vec_StrPush( vBuffer, 0 );
    }
    // prepare the return values
    *pnSize = Vec_StrSize( vBuffer );
    pBuffer = Vec_StrReleaseArray( vBuffer );
    Vec_StrFree( vBuffer );
    return pBuffer;
}

/**Function*************************************************************

  Synopsis    [This procedure is used to test the above procedure.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ioa_WriteAigerBufferTest(pMan *Aig_Man_t, pFileName string, fWriteSymbols int32, fCompact int32) {
    var pFile *os.File;
    var pBuffer string;
    var nSize int32;
    if Aig_ManCoNum(pMan) == 0 {
        fmt.Printf("AIG cannot be written because it has no POs.\n");
        return;
    }
    // start the output stream
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Ioa_WriteAiger(): Cannot open the output file \"%s\".\n", pFileName);
        return;
    }
    // write the buffer
    pBuffer = Ioa_WriteAigerIntoMemory( pMan, &nSize );
    fwrite( pBuffer, 1, nSize, pFile );
    // inlining of macro ABC_FREE
    ((pBuffer) ? (free((char *) (pBuffer)), (pBuffer) = 0) : 0);
    // write the comment
//    fprintf( pFile, "c" );
//    if ( pMan->pName )
//        fprintf( pFile, "n%s%c", pMan->pName, '\0' );
    fmt.Fprintf(pFile, "\nThis file was produced by the IOA package in ABC on %s\n", Ioa_TimeStamp());
    fmt.Fprintf(pFile, "For information about AIGER format, refer to %s\n", "http://fmv.jku.at/aiger");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Writes the AIG in the binary AIGER format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ioa_WriteAiger(pMan *Aig_Man_t, pFileName string, fWriteSymbols int32, fCompact int32) {
//    Bar_Progress_t * pProgress;
    var pFile *os.File;
    var pObj, pDriver *Aig_Obj_t;
    var i, nNodes, nBufferSize, Pos int32;
    unsigned char * pBuffer;
    var uLit0, uLit1, uLit uint32;

    if Aig_ManCoNum(pMan) == 0 {
        fmt.Printf("AIG cannot be written because it has no POs.\n");
        return;
    }

//    assert( Aig_ManIsStrash(pMan) );
    // start the output stream
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Ioa_WriteAiger(): Cannot open the output file \"%s\".\n", pFileName);
        return;
    }
/*
    Aig_ManForEachLatch( pMan, pObj, i )
        if ( !Aig_LatchIsInit0(pObj) )
        {
            fprintf( stdout, "Ioa_WriteAiger(): Cannot write AIGER format with non-0 latch init values. Run \"zero\".\n" );
            return;
        }
*/
    // set the node numbers to be used in the output file
    nNodes = 0;
    Ioa_ObjSetAigerNum( Aig_ManConst1(pMan), nNodes++ );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCis, i)), 1); i++  {
        Ioa_ObjSetAigerNum( pObj, nNodes++ );
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Ioa_ObjSetAigerNum( pObj, nNodes++ ); }
    }

    // write the header "M I L O A" where M = I + L + A
    fmt.Fprintf(pFile, "aig%s %d %d %d %d %d", fCompact? "2" : "", Aig_ManCiNum(pMan) + Aig_ManNodeNum(pMan), Aig_ManCiNum(pMan) - Aig_ManRegNum(pMan), Aig_ManRegNum(pMan), Aig_ManConstrNum(pMan) ? 0 : Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan), Aig_ManNodeNum(pMan));
    // write the extended header "B C J F"
    if Aig_ManConstrNum(pMan) {
        fmt.Fprintf(pFile, " %d %d", Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan) - Aig_ManConstrNum(pMan), Aig_ManConstrNum(pMan)); }
    fmt.Fprintf(pFile, "\n"); 

    // if the driver node is a constant, we need to complement the literal below
    // because, in the AIGER format, literal 0/1 is represented as number 0/1
    // while, in ABC, constant 1 node has number 0 and so literal 0/1 will be 1/0

    Aig_ManInvertConstraints( pMan );
    if !(fCompact != 0) {
        // write latch drivers
        // inlining of macro Aig_ManForEachLiSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCoNum(pMan)-Aig_ManRegNum(pMan); (i < Vec_PtrSize(pMan.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
            pDriver = Aig_ObjFanin0(pObj);
            fmt.Fprintf(pFile, "%d\n", Ioa_ObjMakeLit( Ioa_ObjAigerNum(pDriver), Aig_ObjFaninC0(pObj) ^ (Ioa_ObjAigerNum(pDriver) == 0) ));
        }

        // write PO drivers
        // inlining of macro Aig_ManForEachPoSeq
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Aig_ManCoNum(pMan)-Aig_ManRegNum(pMan)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
            pDriver = Aig_ObjFanin0(pObj);
            fmt.Fprintf(pFile, "%d\n", Ioa_ObjMakeLit( Ioa_ObjAigerNum(pDriver), Aig_ObjFaninC0(pObj) ^ (Ioa_ObjAigerNum(pDriver) == 0) ));
        }
    
    } else {
        var vLits *Vec_Int_t  = Ioa_WriteAigerLiterals( pMan );
        var vBinary *Vec_Str_t  = Ioa_WriteEncodeLiterals( vLits );
        fwrite( Vec_StrArray(vBinary), 1, Vec_StrSize(vBinary), pFile );
        Vec_StrFree( vBinary );
        Vec_IntFree( vLits );
    }
    Aig_ManInvertConstraints( pMan );

    // write the nodes into the buffer
    Pos = 0;
    nBufferSize = 6 * Aig_ManNodeNum(pMan) + 100; // skeptically assuming 3 chars per one AIG edge
    // inlining of macro ABC_ALLOC
    pBuffer = ((unsigned char *) malloc(sizeof(unsigned char) * (size_t)(nBufferSize)));
//    pProgress = Bar_ProgressStart( stdout, Aig_ManObjNumMax(pMan) );
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
//        Bar_ProgressUpdate( pProgress, i, NULL );
        uLit  = Ioa_ObjMakeLit( Ioa_ObjAigerNum(pObj), 0 );
        uLit0 = Ioa_ObjMakeLit( Ioa_ObjAigerNum(Aig_ObjFanin0(pObj)), Aig_ObjFaninC0(pObj) );
        uLit1 = Ioa_ObjMakeLit( Ioa_ObjAigerNum(Aig_ObjFanin1(pObj)), Aig_ObjFaninC1(pObj) );
        assert(uLit0 != uLit1, "uLit0 != uLit1");
        if uLit0 > uLit1 {
            var Temp int32  = uLit0;
            uLit0 = uLit1;
            uLit1 = Temp;
        }
        Pos = Ioa_WriteAigerEncode( pBuffer, Pos, uLit  - uLit1 );
        Pos = Ioa_WriteAigerEncode( pBuffer, Pos, uLit1 - uLit0 );
        if Pos > nBufferSize - 10 {
            fmt.Printf("Ioa_WriteAiger(): AIGER generation has failed because the allocated buffer is too small.\n");
            _ = (pFile).Close();
            return;
        }
    }
    }
    assert(Pos < nBufferSize, "Pos < nBufferSize");
//    Bar_ProgressStop( pProgress );

    // write the buffer
    fwrite( pBuffer, 1, Pos, pFile );
    // inlining of macro ABC_FREE
    ((pBuffer) ? (free((char *) (pBuffer)), (pBuffer) = 0) : 0);
/*
    // write the symbol table
    if ( fWriteSymbols )
    {
        int bads;
        // write PIs
        Aig_ManForEachPiSeq( pMan, pObj, i )
            fprintf( pFile, "i%d %s\n", i, Aig_ObjName(pObj) );
        // write latches
        Aig_ManForEachLoSeq( pMan, pObj, i )
            fprintf( pFile, "l%d %s\n", i, Aig_ObjName(Aig_ObjFanout0(pObj)) );
        // write POs
        bads = Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan) - Aig_ManConstrNum(pMan);
        Aig_ManForEachPoSeq( pMan, pObj, i )
            if ( !Aig_ManConstrNum(pMan) )
                fprintf( pFile, "o%d %s\n", i, Aig_ObjName(pObj) );
            else if ( i < bads )
                fprintf( pFile, "b%d %s\n", i, Aig_ObjName(pObj) );
            else
                fprintf( pFile, "c%d %s\n", i - bads, Aig_ObjName(pObj) );
    }
*/
    // write the comment
    fmt.Fprintf(pFile, "c");
    if pMan.pName {
        fmt.Fprintf(pFile, "n%s%c", pMan.pName, '\0'); }
    fmt.Fprintf(pFile, "\nThis file was produced by the IOA package in ABC on %s\n", Ioa_TimeStamp());
    fmt.Fprintf(pFile, "For information about AIGER format, refer to %s\n", "http://fmv.jku.at/aiger");
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

