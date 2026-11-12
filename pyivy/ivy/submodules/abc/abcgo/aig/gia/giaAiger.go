package gia

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [giaAiger.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Procedures to read/write binary AIGER format developed by
  Armin Biere, Johannes Kepler University (http://fmv.jku.at/)]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaAiger.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/tim/tim.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START

const XAIG_VERBOSE = 0

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_FileFixName(pFileName string) {
    var pName string;
    for  pName = pFileName; *pName; pName++  {
        if *pName == '>' {
            *pName = '\\'; }
    }
}
func Gia_FileNameGeneric(FileName string) string {
    var pDot, pRes string;
    pRes = Abc_UtilStrsav( FileName );
    if pDot = strrchr( pRes, '.' ) {
        *pDot = 0; }
    return pRes;
}
func Gia_FileSize(pFileName string) int32 {
    var pFile *os.File;
    var nFileSize int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Gia_FileSize(): The file is unavailable (absent or open).\n");
        return 0;
    }
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile ); 
    _ = (pFile).Close();
    return nFileSize;
}
func Gia_FileWriteBufferSize(pFile *os.File, nSize int32) {
    unsigned char Buffer[5];
    Gia_AigerWriteInt( Buffer, nSize );
    fwrite( Buffer, 1, 4, pFile );
}

/**Function*************************************************************

  Synopsis    [Create the array of literals to be written.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_AigerCollectLiterals(p *Gia_Man_t) *Vec_Int_t {
    var vLits *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    vLits = Vec_IntAlloc( Gia_ManPoNum(p) );
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        Vec_IntPush( vLits, Gia_ObjFaninLit0p(p, pObj) );
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Vec_IntPush( vLits, Gia_ObjFaninLit0p(p, pObj) );
    }
    return vLits;
}
Vec_Int_t * Gia_AigerReadLiterals( unsigned char ** ppPos, int nEntries )
{
    var vLits *Vec_Int_t;
    var Lit, LitPrev, Diff, i int32;
    vLits = Vec_IntAlloc( nEntries );
    LitPrev = Gia_AigerReadUnsigned( ppPos );
    Vec_IntPush( vLits, LitPrev );
    for  i = 1; i < nEntries; i++  {
//        Diff = Lit - LitPrev;
//        Diff = (Lit < LitPrev)? -Diff : Diff;
//        Diff = ((2 * Diff) << 1) | (int)(Lit < LitPrev);
        Diff = Gia_AigerReadUnsigned( ppPos );
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
func Gia_AigerWriteLiterals(vLits *Vec_Int_t) *Vec_Str_t {
    var vBinary *Vec_Str_t;
    var Pos, Lit, LitPrev, Diff, i int32;
    vBinary = Vec_StrAlloc( 2 * Vec_IntSize(vLits) );
    LitPrev = Vec_IntEntry( vLits, 0 );
    Pos = Gia_AigerWriteUnsignedBuffer( (unsigned char *)Vec_StrArray(vBinary), Pos, LitPrev ); 
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(vLits)) && (((Lit) = Vec_IntEntry(vLits, i)), 1); i++  {
        Diff = Lit - LitPrev;
        if Lit < LitPrev {
            Diff = -Diff;
        } else {
            Diff = Diff;
        }
        Diff = (Diff << 1) | int32((Lit < LitPrev));
        Pos = Gia_AigerWriteUnsignedBuffer( (unsigned char *)Vec_StrArray(vBinary), Pos, Diff );
        LitPrev = Lit;
        if Pos + 10 > vBinary.nCap {
            Vec_StrGrow( vBinary, vBinary.nCap+1 ); }
    }
    vBinary.nSize = Pos;
/*
    // verify
    {
        extern Vec_Int_t * Gia_AigerReadLiterals( char ** ppPos, int nEntries );
        char * pPos = Vec_StrArray( vBinary );
        Vec_Int_t * vTemp = Gia_AigerReadLiterals( &pPos, Vec_IntSize(vLits) );
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

  Synopsis    [Reads the AIG in the binary AIGER format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_AigerReadFromMemory(pContents string, nFileSize int32, fGiaSimple int32, fSkipStrash int32, fCheck int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var vLits *Vec_Int_t = nil;
    var vPoTypes *Vec_Int_t = nil;
    var vNodes *Vec_Int_t;
    var vDrivers *Vec_Int_t;
    var vInits *Vec_Int_t = nil;
    var iObj, iNode0, iNode1, fHieOnly int32;
    var nTotal, nInputs, nOutputs, nLatches, nAnds, i int32;
    var nBad, nConstr, nJust, nFair int32;
    unsigned char * pDrivers, * pSymbols, * pCur;
    var uLit0, uLit1, uLit uint32;

    // read the parameters (M I L O A + B C J F)
    pCur = (unsigned char *)pContents;         for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of objects
    nTotal = strconv.Atoi((char *)pCur);    for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of inputs
    nInputs = strconv.Atoi((char *)pCur);   for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of latches
    nLatches = strconv.Atoi((char *)pCur);  for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of outputs
    nOutputs = strconv.Atoi((char *)pCur);  for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of nodes
    nAnds = strconv.Atoi((char *)pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
    if *pCur == ' ' {
//        assert( nOutputs == 0 );
        // read the number of properties
        pCur++;
        nBad = strconv.Atoi((char *)pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
        nOutputs += nBad;
    }
    if *pCur == ' ' {
        // read the number of properties
        pCur++;
        nConstr = strconv.Atoi((char *)pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
        nOutputs += nConstr;
    }
    if *pCur == ' ' {
        // read the number of properties
        pCur++;
        nJust = strconv.Atoi((char *)pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
        nOutputs += nJust;
    }
    if *pCur == ' ' {
        // read the number of properties
        pCur++;
        nFair = strconv.Atoi((char *)pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
        nOutputs += nFair;
    }
    if *pCur != '\n' {
        fmt.Fprintf(os.Stdout, "The parameter line is in a wrong format.\n");
        return nil;
    }
    pCur++;

    // check the parameters
    if nTotal != nInputs + nLatches + nAnds {
        fmt.Fprintf(os.Stdout, "The number of objects does not match.\n");
        return nil;
    }
    if nJust || (nFair != 0) {
        fmt.Fprintf(os.Stdout, "Reading AIGER files with liveness properties is currently not supported.\n");
        return nil;
    }

    if nConstr {
        if nConstr == 1 {
            fmt.Fprintf(os.Stdout, "Warning: The last output is interpreted as a constraint.\n");
        } else {
            fmt.Fprintf(os.Stdout, "Warning: The last %d outputs are interpreted as constraints.\n", nConstr); }
    }

    // allocate the empty AIG
    pNew = Gia_ManStart( nTotal + nLatches + nOutputs + 1 );
    pNew.nConstrs = nConstr;
    pNew.fGiaSimple = fGiaSimple;

    // prepare the array of nodes
    vNodes = Vec_IntAlloc( 1 + nTotal );
    Vec_IntPush( vNodes, 0 );

    // create the PIs
    for  i = 0; i < nInputs + nLatches; i++  {
        iObj = Gia_ManAppendCi(pNew);    
        Vec_IntPush( vNodes, iObj );
    }

    // remember the beginning of latch/PO literals
    pDrivers = pCur;
    if pContents[3] == ' ' { // standard AIGER
    
        // scroll to the beginning of the binary data
        for  i = 0; i < nLatches + nOutputs;  {
            if *pCur++ == '\n' {
                i++; }
        }
    
    } else { // modified AIGER
    
        vLits = Gia_AigerReadLiterals( &pCur, nLatches + nOutputs );
    }

    // create the AND gates
    if !(fGiaSimple != 0) && !(fSkipStrash != 0) {
        Gia_ManHashAlloc( pNew ); }
    for  i = 0; i < nAnds; i++  {
        uLit = ((i + 1 + nInputs + nLatches) << 1);
        uLit1 = uLit  - Gia_AigerReadUnsigned( &pCur );
        uLit0 = uLit1 - Gia_AigerReadUnsigned( &pCur );
//        assert( uLit1 > uLit0 );
        iNode0 = Abc_LitNotCond( Vec_IntEntry(vNodes, uLit0 >> 1), uLit0 & 1 );
        iNode1 = Abc_LitNotCond( Vec_IntEntry(vNodes, uLit1 >> 1), uLit1 & 1 );
        assert(Vec_IntSize(vNodes) == i + 1 + nInputs + nLatches, "Vec_IntSize(vNodes) == i + 1 + nInputs + nLatches");
        if !(fGiaSimple != 0) && (fSkipStrash != 0) {
            if iNode0 == iNode1 {
                Vec_IntPush( vNodes, Gia_ManAppendBuf(pNew, iNode0) );
            } else {
                Vec_IntPush( vNodes, Gia_ManAppendAnd(pNew, iNode0, iNode1) ); }
        
        } else {
            Vec_IntPush( vNodes, Gia_ManHashAnd(pNew, iNode0, iNode1) ); }
    }
    if !(fGiaSimple != 0) && !(fSkipStrash != 0) {
        Gia_ManHashStop( pNew ); }

    // remember the place where symbols begin
    pSymbols = pCur;

    // read the latch driver literals
    vDrivers = Vec_IntAlloc( nLatches + nOutputs );
    if pContents[3] == ' ' { // standard AIGER
    
        vInits = Vec_IntAlloc( nLatches );
        pCur = pDrivers;
        for  i = 0; i < nLatches; i++  {
            uLit0 = strconv.Atoi((char *)pCur);   
            for  *pCur != ' ' && *pCur != '\n'  { 
                pCur++;
            }
            if *pCur == ' ' {
                pCur++;
                Vec_IntPush( vInits, strconv.Atoi((char *)pCur) );
                for  *pCur++ != '\n'  {; }
            
            } else {
                pCur++;
                Vec_IntPush( vInits, 0 );
            }
            iNode0 = Abc_LitNotCond( Vec_IntEntry(vNodes, uLit0 >> 1), (uLit0 & 1) );
            Vec_IntPush( vDrivers, iNode0 );
        }
        // read the PO driver literals
        for  i = 0; i < nOutputs; i++  {
            uLit0 = strconv.Atoi((char *)pCur);   for  *pCur++ != '\n'  {; }
            iNode0 = Abc_LitNotCond( Vec_IntEntry(vNodes, uLit0 >> 1), (uLit0 & 1) );
            Vec_IntPush( vDrivers, iNode0 );
        }

    
    } else {
        // read the latch driver literals
        for  i = 0; i < nLatches; i++  {
            uLit0 = Vec_IntEntry( vLits, i );
            iNode0 = Abc_LitNotCond( Vec_IntEntry(vNodes, uLit0 >> 1), (uLit0 & 1) );
            Vec_IntPush( vDrivers, iNode0 );
        }
        // read the PO driver literals
        for  i = 0; i < nOutputs; i++  {
            uLit0 = Vec_IntEntry( vLits, i+nLatches );
            iNode0 = Abc_LitNotCond( Vec_IntEntry(vNodes, uLit0 >> 1), (uLit0 & 1) );
            Vec_IntPush( vDrivers, iNode0 );
        }
        Vec_IntFree( vLits );
    }

    // create the POs
    for  i = 0; i < nOutputs; i++  {
        Gia_ManAppendCo( pNew, Vec_IntEntry(vDrivers, nLatches + i) );
    }
    for  i = 0; i < nLatches; i++  {
        Gia_ManAppendCo( pNew, Vec_IntEntry(vDrivers, i) );
    }
    Vec_IntFree( vDrivers );

    // create the latches
    Gia_ManSetRegNum( pNew, nLatches );

    // read signal names if they are of the special type
    pCur = pSymbols;
    if pCur < (unsigned char *)pContents + nFileSize && *pCur != 'c' {
        var fBreakUsed int32;
        unsigned char * pCurOld = pCur;
        pNew.vUserPiIds = Vec_IntStartFull( nInputs );
        pNew.vUserPoIds = Vec_IntStartFull( nOutputs );
        pNew.vUserFfIds = Vec_IntStartFull( nLatches );
        for  pCur < (unsigned char *)pContents + nFileSize && *pCur != 'c'  {
            var iTerm int32;
            var pType string  = (char *)pCur;
            // check terminal type
            if *pCur != 'i' && *pCur != 'o' && *pCur != 'l' {
//                fprintf( stdout, "Wrong terminal type.\n" );
                fBreakUsed = 1;
                break;
            }
            // get terminal number
            iTerm = strconv.Atoi((char *)++pCur);  for  *pCur++ != ' '  {; }
            // skip spaces
            for  *pCur == ' '  {
                pCur++;
            }
            // decode the user numbers:
            // flops are named: @l<num>
            // PIs are named: @i<num>
            // POs are named: @o<num>
            if *pCur++ != '@' {
                fBreakUsed = 1;
                break;
            }
            if *pCur == 'i' && *pType == 'i' {
                Vec_IntWriteEntry( pNew.vUserPiIds, iTerm, strconv.Atoi((char *)pCur+1) );
            } else if *pCur == 'o' && *pType == 'o' {
                Vec_IntWriteEntry( pNew.vUserPoIds, iTerm, strconv.Atoi((char *)pCur+1) );
            } else if *pCur == 'l' && *pType == 'l' {
                Vec_IntWriteEntry( pNew.vUserFfIds, iTerm, strconv.Atoi((char *)pCur+1) );
            } else {
                fmt.Fprintf(os.Stdout, "Wrong name format.\n");
                fBreakUsed = 1;
                break;
            }
            // skip digits
            for  *pCur++ != '\n'  {; }
        }
        // in case of abnormal termination, remove the arrays
        if fBreakUsed != 0 {
            unsigned char * pName;
            var Entry, nInvars, nConstr, iTerm int32;

            var vPoNames *Vec_Int_t  = Vec_IntStartFull( nOutputs );

            Vec_IntFreeP( &pNew.vUserPiIds );
            Vec_IntFreeP( &pNew.vUserPoIds );
            Vec_IntFreeP( &pNew.vUserFfIds );

            // try to figure out signal names
            fBreakUsed = 0;
            pCur = (unsigned char *)pCurOld;
            for  pCur < (unsigned char *)pContents + nFileSize && *pCur != 'c'  {
                // get the terminal type
                if *pCur == 'i' || *pCur == 'l' {
                    // skip till the end of the line
                    for  *pCur++ != '\n'  {; }
                    *(pCur-1) = 0;
                    continue;
                }
                if *pCur != 'o' {
//                    fprintf( stdout, "Wrong terminal type.\n" );
                    fBreakUsed = 1;
                    break;
                }
                // get the terminal number
                iTerm = strconv.Atoi((char *)++pCur);  for  *pCur++ != ' '  {; }
                // get the node
                if iTerm < 0 || iTerm >= nOutputs {
                    fmt.Fprintf(os.Stdout, "The output number (%d) is out of range.\n", iTerm);
                    fBreakUsed = 1;
                    break;
                }
                if Vec_IntEntry(vPoNames, iTerm) != ~0 {
                    fmt.Fprintf(os.Stdout, "The output number (%d) is listed twice.\n", iTerm);
                    fBreakUsed = 1;
                    break;
                }

                // get the name
                pName = pCur;          for  *pCur++ != '\n'  {; }
                *(pCur-1) = 0;
                // assign the name
                Vec_IntWriteEntry( vPoNames, iTerm, pName - (unsigned char *)pContents );
            } 

            // check that all names are assigned
            if !(fBreakUsed != 0) {
                nInvars = nConstr = 0;
                vPoTypes = Vec_IntStart( Gia_ManPoNum(pNew) );
                // inlining of macro Vec_IntForEachEntry
                for  i = 0; (i < Vec_IntSize(vPoNames)) && (((Entry) = Vec_IntEntry(vPoNames, i)), 1); i++  {
                    if Entry == ~0 {
                        continue; }
                    if strncmp( pContents+Entry, "constraint:", 11 ) == 0 {
                        Vec_IntWriteEntry( vPoTypes, i, 1 );
                        nConstr++;
                    }
                    if strncmp( pContents+Entry, "invariant:", 10 ) == 0 {
                        Vec_IntWriteEntry( vPoTypes, i, 2 );
                        nInvars++;
                    }
                }
                if nConstr {
                    fmt.Fprintf(os.Stdout, "Recognized and added %d constraints.\n", nConstr); }
                if nInvars {
                    fmt.Fprintf(os.Stdout, "Recognized and skipped %d invariants.\n", nInvars); }
                if nConstr == 0 && nInvars == 0 {
                    Vec_IntFreeP( &vPoTypes ); }
            }
            Vec_IntFree( vPoNames );
        }
    }


    // check if there are other types of information to read
    if pCur + 1 < (unsigned char *)pContents + nFileSize && *pCur == 'c' {
        var fVerbose int32  = XAIG_VERBOSE;
        var vStr *Vec_Str_t;
        unsigned char * pCurTemp;
        pCur++;
        // skip new line if present
//        if ( *pCur == '\n' )
//            pCur++;
        for  pCur < (unsigned char *)pContents + nFileSize  {
            // read extra AIG
            if *pCur == 'a' {
                pCur++;
                vStr = Vec_StrStart( Gia_AigerReadInt(pCur) );             pCur += 4;
                memcpy( Vec_StrArray(vStr), pCur, (size_t)Vec_StrSize(vStr) );
                pCur += Vec_StrSize(vStr);
                pNew.pAigExtra = Gia_AigerReadFromMemory( Vec_StrArray(vStr), Vec_StrSize(vStr), 0, 0, 0 );
                Vec_StrFree( vStr );
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"a\".\n"); }
            
            // read number of constraints
            } else if *pCur == 'c' {
                pCur++;
                assert(Gia_AigerReadInt(pCur) == 4, "Gia_AigerReadInt(pCur) == 4");                     pCur += 4;
                pNew.nConstrs = Gia_AigerReadInt( pCur );                 pCur += 4;
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"c\".\n"); }
            
            // read delay information
            } else if *pCur == 'd' {
                pCur++;
                assert(Gia_AigerReadInt(pCur) == 4, "Gia_AigerReadInt(pCur) == 4");                     pCur += 4;
                pNew.nAnd2Delay = Gia_AigerReadInt(pCur);                 pCur += 4;
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"d\".\n"); }
            
            } else if *pCur == 'i' {
                pCur++;
                nInputs = Gia_AigerReadInt(pCur)/4;                        pCur += 4;
                pNew.vInArrs  = Vec_FltStart( nInputs );
                memcpy( Vec_FltArray(pNew.vInArrs),  pCur, (size_t)4*nInputs );   pCur += 4*nInputs;
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"i\".\n"); }
            
            } else if *pCur == 'o' {
                pCur++;
                nOutputs = Gia_AigerReadInt(pCur)/4;                       pCur += 4;
                pNew.vOutReqs  = Vec_FltStart( nOutputs );
                memcpy( Vec_FltArray(pNew.vOutReqs),  pCur, (size_t)4*nOutputs ); pCur += 4*nOutputs;
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"o\".\n"); }
            
            // read equivalence classes
            } else if *pCur == 'e' {
                extern Gia_Rpr_t * Gia_AigerReadEquivClasses( unsigned char ** ppPos, int nSize );
                pCur++;
                pCurTemp = pCur + Gia_AigerReadInt(pCur) + 4;              pCur += 4;
                pNew.pReprs = Gia_AigerReadEquivClasses( &pCur, Gia_ManObjNum(pNew) );
                pNew.pNexts = Gia_ManDeriveNexts( pNew );
                assert(pCur == pCurTemp, "pCur == pCurTemp");
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"e\".\n"); }
            
            // read flop classes
            } else if *pCur == 'f' {
                pCur++;
                assert(Gia_AigerReadInt(pCur) == 4*Gia_ManRegNum(pNew), "Gia_AigerReadInt(pCur) == 4*Gia_ManRegNum(pNew)");   pCur += 4;
                pNew.vFlopClasses  = Vec_IntStart( Gia_ManRegNum(pNew) );
                memcpy( Vec_IntArray(pNew.vFlopClasses),  pCur, (size_t)4*Gia_ManRegNum(pNew) );   pCur += 4*Gia_ManRegNum(pNew);
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"f\".\n"); }
            
            // read gate classes
            } else if *pCur == 'g' {
                pCur++;
                assert(Gia_AigerReadInt(pCur) == 4*Gia_ManObjNum(pNew), "Gia_AigerReadInt(pCur) == 4*Gia_ManObjNum(pNew)");   pCur += 4;
                pNew.vGateClasses  = Vec_IntStart( Gia_ManObjNum(pNew) );
                memcpy( Vec_IntArray(pNew.vGateClasses),  pCur, (size_t)4*Gia_ManObjNum(pNew) );   pCur += 4*Gia_ManObjNum(pNew);
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"g\".\n"); }
            
            // read hierarchy information
            } else if *pCur == 'h' {
                pCur++;
                vStr = Vec_StrStart( Gia_AigerReadInt(pCur) );          pCur += 4;
                memcpy( Vec_StrArray(vStr), pCur, (size_t)Vec_StrSize(vStr) );
                pCur += Vec_StrSize(vStr);
                pNew.pManTime = Tim_ManLoad( vStr, 1 );
                Vec_StrFree( vStr );
                fHieOnly = 1;
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"h\".\n"); }
            
            // read packing
            } else if *pCur == 'k' {
                extern Vec_Int_t * Gia_AigerReadPacking( unsigned char ** ppPos, int nSize );
                var nSize int32;
                pCur++;
                nSize = Gia_AigerReadInt(pCur);
                pCurTemp = pCur + nSize + 4;                            pCur += 4;
                pNew.vPacking = Gia_AigerReadPacking( &pCur, nSize ); 
                assert(pCur == pCurTemp, "pCur == pCurTemp");
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"k\".\n"); }
            
            // read mapping
            } else if *pCur == 'm' {
                extern int * Gia_AigerReadMapping( unsigned char ** ppPos, int nSize );
                extern int * Gia_AigerReadMappingSimple( unsigned char ** ppPos, int nSize );
                extern Vec_Int_t * Gia_AigerReadMappingDoc( unsigned char ** ppPos, int nObjs );
                var nSize int32;
                pCur++;
                nSize = Gia_AigerReadInt(pCur);
                pCurTemp = pCur + nSize + 4;           pCur += 4;
                pNew.vMapping = Gia_AigerReadMappingDoc( &pCur, Gia_ManObjNum(pNew) );
                assert(pCur == pCurTemp, "pCur == pCurTemp");
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"m\".\n"); }
            
            // read model name
            } else if *pCur == 'n' {
                pCur++;
                if (*pCur >= 'a' && *pCur <= 'z') || (*pCur >= 'A' && *pCur <= 'Z') || (*pCur >= '0' && *pCur <= '9') {
                    // inlining of macro ABC_FREE
                    ((pNew.pName) ? (free((char *) (pNew.pName)), (pNew.pName) = 0) : 0);
                    pNew.pName = Abc_UtilStrsav( (char *)pCur );       pCur += len(pNew.pName) + 1;
                
                } else {
                    pCurTemp = pCur + Gia_AigerReadInt(pCur) + 4;       pCur += 4;
                    // inlining of macro ABC_FREE
                    ((pNew.pName) ? (free((char *) (pNew.pName)), (pNew.pName) = 0) : 0);
                    pNew.pName = Abc_UtilStrsav( (char *)pCur );       pCur += len(pNew.pName) + 1;
                    assert(pCur == pCurTemp, "pCur == pCurTemp");
                }
            
            // read placement
            } else if *pCur == 'p' {
                var pPlacement *Gia_Plc_t;
                pCur++;
                pCurTemp = pCur + Gia_AigerReadInt(pCur) + 4;           pCur += 4;
                // inlining of macro ABC_ALLOC
                pPlacement = ((Gia_Plc_t *) malloc(sizeof(Gia_Plc_t) * (size_t)(Gia_ManObjNum(pNew))));
                memcpy( pPlacement, pCur, (size_t)4*Gia_ManObjNum(pNew) );      pCur += 4*Gia_ManObjNum(pNew);
                assert(pCur == pCurTemp, "pCur == pCurTemp");
                pNew.pPlacement = pPlacement;
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"p\".\n"); }
            
            // read register classes
            } else if *pCur == 'r' {
                var i, nRegs int32;
                pCur++;
                pCurTemp = pCur + Gia_AigerReadInt(pCur) + 4;           pCur += 4;
                nRegs = Gia_AigerReadInt(pCur);                         pCur += 4;
                //nRegs = (pCurTemp - pCur)/4;
                pNew.vRegClasses = Vec_IntAlloc( nRegs );
                for  i = 0; i < nRegs; i++  {
                    Vec_IntPush( pNew.vRegClasses, Gia_AigerReadInt(pCur) );
pCur += 4;
                }
                assert(pCur == pCurTemp, "pCur == pCurTemp");
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"r\".\n"); }
            
            // read register inits
            } else if *pCur == 's' {
                var i, nRegs int32;
                pCur++;
                pCurTemp = pCur + Gia_AigerReadInt(pCur) + 4;           pCur += 4;
                nRegs = Gia_AigerReadInt(pCur);                         pCur += 4;
                pNew.vRegInits = Vec_IntAlloc( nRegs );
                for  i = 0; i < nRegs; i++  {
                    Vec_IntPush( pNew.vRegInits, Gia_AigerReadInt(pCur) );
pCur += 4;
                }
                assert(pCur == pCurTemp, "pCur == pCurTemp");
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"s\".\n"); }
            
            // read configuration data
            } else if *pCur == 'b' {
                var nSize int32;
                pCur++;
                nSize = Gia_AigerReadInt(pCur);
                pCurTemp = pCur + nSize + 4;                            pCur += 4;
                pNew.pCellStr = Abc_UtilStrsav( (char*)pCur );         pCur += len((char*)pCur) + 1;
                nSize = nSize - len(pNew.pCellStr) - 1;
                assert(nSize % 4 == 0, "nSize % 4 == 0");
                pNew.vConfigs = Vec_IntAlloc(nSize / 4);
//                memcpy(Vec_IntArray(pNew->vConfigs), pCur, (size_t)nSize);      pCur += nSize;
                for  i = 0; i < nSize / 4; i++  {
                    Vec_IntPush( pNew.vConfigs, Gia_AigerReadInt(pCur) );
pCur += 4;
                }
                assert(pCur == pCurTemp, "pCur == pCurTemp");
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"b\".\n"); }
            
            // read choices
            } else if *pCur == 'q' {
                var i, nPairs, iRepr, iNode int32;
                assert(!Gia_ManHasChoices(pNew), "!Gia_ManHasChoices(pNew)");
                // inlining of macro ABC_CALLOC
                pNew.pSibls = ((int *) calloc((size_t)(Gia_ManObjNum(pNew)), sizeof(int)));
                pCur++;
                pCurTemp = pCur + Gia_AigerReadInt(pCur) + 4;           pCur += 4;
                nPairs = Gia_AigerReadInt(pCur);                        pCur += 4;
                for  i = 0; i < nPairs; i++  {
                    iRepr = Gia_AigerReadInt(pCur);                     pCur += 4;
                    iNode = Gia_AigerReadInt(pCur);                     pCur += 4;
                    pNew.pSibls[iRepr] = iNode;
                    assert(iRepr > iNode, "iRepr > iNode");
                }
                assert(pCur == pCurTemp, "pCur == pCurTemp");
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"q\".\n"); }
            
            // read switching activity
            } else if *pCur == 'u' { 
                unsigned char * pSwitching;
                pCur++;
                pCurTemp = pCur + Gia_AigerReadInt(pCur) + 4;           pCur += 4;
                // inlining of macro ABC_ALLOC
                pSwitching = ((unsigned char *) malloc(sizeof(unsigned char) * (size_t)(Gia_ManObjNum(pNew))));
                memcpy( pSwitching, pCur, (size_t)Gia_ManObjNum(pNew) );        pCur += Gia_ManObjNum(pNew);
                assert(pCur == pCurTemp, "pCur == pCurTemp");
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"s\".\n"); }
            
            // read timing manager
            } else if *pCur == 't' {
                pCur++;
                vStr = Vec_StrStart( Gia_AigerReadInt(pCur) );          pCur += 4;
                memcpy( Vec_StrArray(vStr), pCur, (size_t)Vec_StrSize(vStr) );  pCur += Vec_StrSize(vStr);
                pNew.pManTime = Tim_ManLoad( vStr, 0 );
                Vec_StrFree( vStr );
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"t\".\n"); }
            
            // read object classes
            } else if *pCur == 'v' {
                pCur++;
                pNew.vObjClasses = Vec_IntStart( Gia_AigerReadInt(pCur)/4 ); pCur += 4;
                memcpy( Vec_IntArray(pNew.vObjClasses), pCur, (size_t)4*Vec_IntSize(pNew.vObjClasses) );
                pCur += 4*Vec_IntSize(pNew.vObjClasses);
                if fVerbose != 0 { fmt.Printf("Finished reading extension \"v\".\n"); }
            
            // read edge information
            } else if *pCur == 'w' {
                var vPairs *Vec_Int_t;
                var i, nPairs int32;
                pCur++;
                pCurTemp = pCur + Gia_AigerReadInt(pCur) + 4;           pCur += 4;
                nPairs = Gia_AigerReadInt(pCur);                        pCur += 4;
                vPairs = Vec_IntAlloc( 2*nPairs );
                for  i = 0; i < 2*nPairs; i++  {
                    Vec_IntPush( vPairs, Gia_AigerReadInt(pCur) );
pCur += 4;
                }
                assert(pCur == pCurTemp, "pCur == pCurTemp");
                if fSkipStrash != 0 {
                    Gia_ManEdgeFromArray( pNew, vPairs );
                    if fVerbose != 0 { fmt.Printf("Finished reading extension \"w\".\n"); }
                
                } else {
                    fmt.Printf("Cannot read extension \"w\" because AIG is rehashed. Use \"&r -s <file.aig>\".\n"); }
                Vec_IntFree( vPairs );
            
            } else { break; }
        }
    }

    // skipping the comments
    Vec_IntFree( vNodes );

    // update polarity of the additional outputs
    if nBad || nConstr || nJust || (nFair != 0) {
        Gia_ManInvertConstraints( pNew ); }

    // clean the PO drivers
    if vPoTypes != nil {
        pNew = Gia_ManDupWithConstraints( pTemp = pNew, vPoTypes );
        Gia_ManStop( pTemp );
        Vec_IntFreeP( &vPoTypes );
    }

    if !(fGiaSimple != 0) && !(fSkipStrash != 0) && Gia_ManHasDangling(pNew) {
        var pManTime *Tim_Man_t;
        var pAigExtra *Gia_Man_t;
        var vFlopMap, vGateMap, vObjMap, vRegClasses, vRegInits *Vec_Int_t;
        vRegClasses = pNew.vRegClasses;  pNew.vRegClasses    = nil;
        vRegInits   = pNew.vRegInits;    pNew.vRegInits      = nil;
        vFlopMap    = pNew.vFlopClasses; pNew.vFlopClasses   = nil;
        vGateMap    = pNew.vGateClasses; pNew.vGateClasses   = nil;
        vObjMap     = pNew.vObjClasses;  pNew.vObjClasses    = nil;
        pManTime = (Tim_Man_t *)pNew.pManTime; pNew.pManTime = nil;
        pAigExtra   = pNew.pAigExtra;    pNew.pAigExtra      = nil;
        pNew = Gia_ManCleanup( pTemp = pNew );
        if (vGateMap || vObjMap) && (Gia_ManObjNum(pNew) < Gia_ManObjNum(pTemp)) {
            fmt.Printf("Cleanup removed objects after reading. Old gate/object abstraction maps are invalid!\n"); }
        Gia_ManStop( pTemp );
        pNew.vRegClasses  = vRegClasses;
        pNew.vRegInits    = vRegInits;
        pNew.vFlopClasses = vFlopMap;
        pNew.vGateClasses = vGateMap;
        pNew.vObjClasses  = vObjMap;
        pNew.pManTime     = pManTime;
        pNew.pAigExtra    = pAigExtra;
    }

    if fHieOnly != 0 {
//        Tim_ManPrint( (Tim_Man_t *)pNew->pManTime );
        if Abc_FrameReadLibBox() == nil {
            fmt.Printf("Warning: Creating unit-delay box delay tables because box library is not available.\n"); }
        Tim_ManCreate( (Tim_Man_t *)pNew.pManTime, Abc_FrameReadLibBox(), pNew.vInArrs, pNew.vOutReqs );
    }
    Vec_FltFreeP( &pNew.vInArrs );
    Vec_FltFreeP( &pNew.vOutReqs );
/*
    // check the result
    if ( fCheck && !Gia_ManCheck( pNew ) )
    {
        printf( "Gia_AigerRead: The network check has failed.\n" );
        Gia_ManStop( pNew );
        return NULL;
    }
*/

    if (vInits != nil) && Vec_IntSum(vInits) {
        // inlining of macro ABC_ALLOC
        var pInit string  = ((char *) malloc(sizeof(char) * (size_t)(Vec_IntSize(vInits) + 1)));
        var pObj *Gia_Obj_t;
        var i int32;
        assert(Vec_IntSize(vInits) == Gia_ManRegNum(pNew), "Vec_IntSize(vInits) == Gia_ManRegNum(pNew)");
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(pNew)) && ((pObj) = Gia_ManCi(pNew, Gia_ManPiNum(pNew)+i)); i++  {
            if Vec_IntEntry(vInits, i) == 0 {
                pInit[i] = '0';
            } else if Vec_IntEntry(vInits, i) == 1 {
                pInit[i] = '1';
            } else {
                assert(Vec_IntEntry(vInits, i) == Abc_Var2Lit(Gia_ObjId(pNew, pObj), 0), "Vec_IntEntry(vInits, i) == Abc_Var2Lit(Gia_ObjId(pNew, pObj), 0)");
                // unitialized value of the latch is the latch literal according to http://fmv.jku.at/hwmcc11/beyond1.pdf
                pInit[i] = 'X';
            }
        }
        pInit[i] = 0;
        pNew = Gia_ManDupZeroUndc( pTemp = pNew, pInit, 0, fGiaSimple, 1 );
        pNew.nConstrs = pTemp.nConstrs; pTemp.nConstrs = 0;
        Gia_ManStop( pTemp );
        // inlining of macro ABC_FREE
        ((pInit) ? (free((char *) (pInit)), (pInit) = 0) : 0);
    }
    Vec_IntFreeP( &vInits );
    if !(fGiaSimple != 0) && !(fSkipStrash != 0) && pNew.vMapping {
        Abc_Print( 0, "Structural hashing enabled while reading AIGER invalidated the mapping.  Consider using \"&r -s\".\n" );
        Vec_IntFreeP( &pNew.vMapping );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Reads the AIG in the binary AIGER format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_AigerRead(pFileName string, fGiaSimple int32, fSkipStrash int32, fCheck int32) *Gia_Man_t {
    var pFile *os.File;
    var pNew *Gia_Man_t;
    var pName, pContents string;
    var nFileSize int32;
    var RetValue int32;

    // read the file into the buffer
    Gia_FileFixName( pFileName );
    nFileSize = Gia_FileSize( pFileName );
    pFile = os.Open(pFileName);
    // inlining of macro ABC_ALLOC
    pContents = ((char *) malloc(sizeof(char) * (size_t)(nFileSize)));
    RetValue = fread( pContents, nFileSize, 1, pFile );
    _ = (pFile).Close();

    pNew = Gia_AigerReadFromMemory( pContents, nFileSize, fGiaSimple, fSkipStrash, fCheck );
    // inlining of macro ABC_FREE
    ((pContents) ? (free((char *) (pContents)), (pContents) = 0) : 0);
    if pNew != nil {
        // inlining of macro ABC_FREE
        ((pNew.pName) ? (free((char *) (pNew.pName)), (pNew.pName) = 0) : 0);
        pName = Gia_FileNameGeneric( pFileName );
        pNew.pName = Abc_UtilStrsav( pName );
        // inlining of macro ABC_FREE
        ((pName) ? (free((char *) (pName)), (pName) = 0) : 0);

        assert(pNew.pSpec == nil, "pNew.pSpec == nil");
        pNew.pSpec = Abc_UtilStrsav( pFileName );
    }
    return pNew;
}



/**Function*************************************************************

  Synopsis    [Writes the AIG in into the memory buffer.]

  Description [The resulting buffer constains the AIG in AIGER format. 
  The resulting buffer should be deallocated by the user.]
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_AigerWriteIntoMemoryStr(p *Gia_Man_t) *Vec_Str_t {
    var vBuffer *Vec_Str_t;
    var pObj *Gia_Obj_t;
    var nNodes, i, uLit, uLit0, uLit1 int32; 
    // set the node numbers to be used in the output file
    Gia_ManConst0(p).Value = nNodes++;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = nNodes++;
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = nNodes++; }
    }

    // write the header "M I L O A" where M = I + L + A
    vBuffer = Vec_StrAlloc( 3*Gia_ManObjNum(p) );
    Vec_StrPrintStr( vBuffer, "aig " );
    Vec_StrPrintNum( vBuffer, Gia_ManCandNum(p) );
    Vec_StrPrintStr( vBuffer, " " );
    Vec_StrPrintNum( vBuffer, Gia_ManPiNum(p) );
    Vec_StrPrintStr( vBuffer, " " );
    Vec_StrPrintNum( vBuffer, Gia_ManRegNum(p) );
    Vec_StrPrintStr( vBuffer, " " );
    Vec_StrPrintNum( vBuffer, Gia_ManPoNum(p) );
    Vec_StrPrintStr( vBuffer, " " );
    Vec_StrPrintNum( vBuffer, Gia_ManAndNum(p) );
    Vec_StrPrintStr( vBuffer, "\n" );

    // write latch drivers
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        uLit = Abc_Var2Lit( Gia_ObjValue(Gia_ObjFanin0(pObj)), Gia_ObjFaninC0(pObj) );
        Vec_StrPrintNum( vBuffer, uLit );
        Vec_StrPrintStr( vBuffer, "\n" );
    }

    // write PO drivers
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        uLit = Abc_Var2Lit( Gia_ObjValue(Gia_ObjFanin0(pObj)), Gia_ObjFaninC0(pObj) );
        Vec_StrPrintNum( vBuffer, uLit );
        Vec_StrPrintStr( vBuffer, "\n" );
    }
    // write the nodes into the buffer
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        uLit  = Abc_Var2Lit( Gia_ObjValue(pObj), 0 );
        uLit0 = Abc_Var2Lit( Gia_ObjValue(Gia_ObjFanin0(pObj)), Gia_ObjFaninC0(pObj) );
        uLit1 = Abc_Var2Lit( Gia_ObjValue(Gia_ObjFanin1(pObj)), Gia_ObjFaninC1(pObj) );
        assert(uLit0 != uLit1, "uLit0 != uLit1");
        if uLit0 > uLit1 {
            var Temp int32  = uLit0;
            uLit0 = uLit1;
            uLit1 = Temp;
        }
        Gia_AigerWriteUnsigned( vBuffer, uLit  - uLit1 );
        Gia_AigerWriteUnsigned( vBuffer, uLit1 - uLit0 );
    }
    }
    Vec_StrPrintStr( vBuffer, "c" );
    return vBuffer;
}

/**Function*************************************************************

  Synopsis    [Writes the AIG in into the memory buffer.]

  Description [The resulting buffer constains the AIG in AIGER format.
  The CI/CO/AND nodes are assumed to be ordered according to some rule.
  The resulting buffer should be deallocated by the user.]
  
  SideEffects [Note that in vCos, PIs are order first, followed by latches!]

  SeeAlso     []

***********************************************************************/
func Gia_AigerWriteIntoMemoryStrPart(p *Gia_Man_t, vCis *Vec_Int_t, vAnds *Vec_Int_t, vCos *Vec_Int_t, nRegs int32) *Vec_Str_t {
    var vBuffer *Vec_Str_t;
    var pObj *Gia_Obj_t;
    var nNodes, i, uLit, uLit0, uLit1 int32; 
    // set the node numbers to be used in the output file
    Gia_ManConst0(p).Value = nNodes++;
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCis)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vCis,i))); i++  {
        assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)");
        pObj.Value = nNodes++;
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vAnds)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vAnds,i))); i++  {
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        pObj.Value = nNodes++;
    }

    // write the header "M I L O A" where M = I + L + A
    vBuffer = Vec_StrAlloc( 3*Gia_ManObjNum(p) );
    Vec_StrPrintStr( vBuffer, "aig " );
    Vec_StrPrintNum( vBuffer, Vec_IntSize(vCis) + Vec_IntSize(vAnds) );
    Vec_StrPrintStr( vBuffer, " " );
    Vec_StrPrintNum( vBuffer, Vec_IntSize(vCis) - nRegs );
    Vec_StrPrintStr( vBuffer, " " );
    Vec_StrPrintNum( vBuffer, nRegs );
    Vec_StrPrintStr( vBuffer, " " );
    Vec_StrPrintNum( vBuffer, Vec_IntSize(vCos) - nRegs );
    Vec_StrPrintStr( vBuffer, " " );
    Vec_StrPrintNum( vBuffer, Vec_IntSize(vAnds) );
    Vec_StrPrintStr( vBuffer, "\n" );

    // write latch drivers
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCos)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vCos,i))); i++  {
        assert(Gia_ObjIsCo(pObj), "Gia_ObjIsCo(pObj)");
        if i < Vec_IntSize(vCos) - nRegs {
            continue; }
        uLit = Abc_Var2Lit( Gia_ObjValue(Gia_ObjFanin0(pObj)), Gia_ObjFaninC0(pObj) );
        Vec_StrPrintNum( vBuffer, uLit );
        Vec_StrPrintStr( vBuffer, "\n" );
    }
    // write output drivers
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCos)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vCos,i))); i++  {
        assert(Gia_ObjIsCo(pObj), "Gia_ObjIsCo(pObj)");
        if i >= Vec_IntSize(vCos) - nRegs {
            continue; }
        uLit = Abc_Var2Lit( Gia_ObjValue(Gia_ObjFanin0(pObj)), Gia_ObjFaninC0(pObj) );
        Vec_StrPrintNum( vBuffer, uLit );
        Vec_StrPrintStr( vBuffer, "\n" );
    }

    // write the nodes into the buffer
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vAnds)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vAnds,i))); i++  {
        uLit  = Abc_Var2Lit( Gia_ObjValue(pObj), 0 );
        uLit0 = Abc_Var2Lit( Gia_ObjValue(Gia_ObjFanin0(pObj)), Gia_ObjFaninC0(pObj) );
        uLit1 = Abc_Var2Lit( Gia_ObjValue(Gia_ObjFanin1(pObj)), Gia_ObjFaninC1(pObj) );
        assert(uLit0 != uLit1, "uLit0 != uLit1");
        if uLit0 > uLit1 {
            var Temp int32  = uLit0;
            uLit0 = uLit1;
            uLit1 = Temp;
        }
        Gia_AigerWriteUnsigned( vBuffer, uLit  - uLit1 );
        Gia_AigerWriteUnsigned( vBuffer, uLit1 - uLit0 );
    }
    Vec_StrPrintStr( vBuffer, "c" );
    return vBuffer;
}

/**Function*************************************************************

  Synopsis    [Writes the AIG in the binary AIGER format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_AigerWrite(pInit *Gia_Man_t, pFileName string, fWriteSymbols int32, fCompact int32, fWriteNewLine int32) {
    var fVerbose int32  = XAIG_VERBOSE;
    var pFile *os.File;
    var p *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vStrExt *Vec_Str_t;
    var i, nBufferSize, Pos int32;
    unsigned char * pBuffer;
    var uLit0, uLit1, uLit uint32;
    assert(pInit.nXors == 0 && pInit.nMuxes == 0, "pInit.nXors == 0 && pInit.nMuxes == 0");

    if Gia_ManCoNum(pInit) == 0 {
        fmt.Printf("AIG cannot be written because it has no POs.\n");
        return;
    }

    // start the output stream
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Gia_AigerWrite(): Cannot open the output file \"%s\".\n", pFileName);
        return;
    }

    // create normalized AIG
    if !Gia_ManIsNormalized(pInit) {
//        printf( "Gia_AigerWrite(): Normalizing AIG for writing.\n" );
        p = Gia_ManDupNormalize( pInit, 0 );
        Gia_ManTransferMapping( p, pInit );
        Gia_ManTransferPacking( p, pInit );
        Gia_ManTransferTiming( p, pInit );
        p.nConstrs   = pInit.nConstrs;
    
    } else {
        p = pInit; }

    // write the header "M I L O A" where M = I + L + A
    fmt.Fprintf(pFile, "aig%s %d %d %d %d %d", fCompact? "2" : "", Gia_ManCiNum(p) + Gia_ManAndNum(p), Gia_ManPiNum(p), Gia_ManRegNum(p), Gia_ManConstrNum(p) ? 0 : Gia_ManPoNum(p), Gia_ManAndNum(p));
    // write the extended header "B C J F"
    if Gia_ManConstrNum(p) {
        fmt.Fprintf(pFile, " %d %d", Gia_ManPoNum(p) - Gia_ManConstrNum(p), Gia_ManConstrNum(p)); }
    fmt.Fprintf(pFile, "\n"); 

    Gia_ManInvertConstraints( p );
    if !(fCompact != 0) {
        // write latch drivers
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
            fmt.Fprintf(pFile, "%d\n", Gia_ObjFaninLit0p(p, pObj));
        }
        // write PO drivers
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            fmt.Fprintf(pFile, "%d\n", Gia_ObjFaninLit0p(p, pObj));
        }
    
    } else {
        var vLits *Vec_Int_t  = Gia_AigerCollectLiterals( p );
        var vBinary *Vec_Str_t  = Gia_AigerWriteLiterals( vLits );
        fwrite( Vec_StrArray(vBinary), 1, Vec_StrSize(vBinary), pFile );
        Vec_StrFree( vBinary );
        Vec_IntFree( vLits );
    }
    Gia_ManInvertConstraints( p );

    // write the nodes into the buffer
    Pos = 0;
    nBufferSize = 8 * Gia_ManAndNum(p) + 100; // skeptically assuming 3 chars per one AIG edge
    // inlining of macro ABC_ALLOC
    pBuffer = ((unsigned char *) malloc(sizeof(unsigned char) * (size_t)(nBufferSize)));
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        uLit  = Abc_Var2Lit( i, 0 );
        uLit0 = Gia_ObjFaninLit0( pObj, i );
        uLit1 = Gia_ObjFaninLit1( pObj, i );
        assert(p.fGiaSimple || Gia_ManBufNum(p) || uLit0 < uLit1, "p.fGiaSimple || Gia_ManBufNum(p) || uLit0 < uLit1");
        Pos = Gia_AigerWriteUnsignedBuffer( pBuffer, Pos, uLit  - uLit1 );
        Pos = Gia_AigerWriteUnsignedBuffer( pBuffer, Pos, uLit1 - uLit0 );
        if Pos > nBufferSize - 10 {
            fmt.Printf("Gia_AigerWrite(): AIGER generation has failed because the allocated buffer is too small.\n");
            _ = (pFile).Close();
            if p != pInit {
                Gia_ManStop( p ); }
            return;
        }
    }
    }
    assert(Pos < nBufferSize, "Pos < nBufferSize");

    // write the buffer
    fwrite( pBuffer, 1, Pos, pFile );
    // inlining of macro ABC_FREE
    ((pBuffer) ? (free((char *) (pBuffer)), (pBuffer) = 0) : 0);

    // write the symbol table
    if p.vNamesIn && p.vNamesOut {
        assert(Vec_PtrSize(p.vNamesIn)  == Gia_ManCiNum(p), "Vec_PtrSize(p.vNamesIn)  == Gia_ManCiNum(p)");
        assert(Vec_PtrSize(p.vNamesOut) == Gia_ManCoNum(p), "Vec_PtrSize(p.vNamesOut) == Gia_ManCoNum(p)");
        // write PIs
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            fmt.Fprintf(pFile, "i%d %s\n", i, (char *)Vec_PtrEntry(p.vNamesIn, i));
        }
        // write latches
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            fmt.Fprintf(pFile, "l%d %s\n", i, (char *)Vec_PtrEntry(p.vNamesIn, Gia_ManPiNum(p) + i));
        }
        // write POs
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            fmt.Fprintf(pFile, "o%d %s\n", i, (char *)Vec_PtrEntry(p.vNamesOut, i));
        }
    }

    // write the comment
    if fWriteNewLine != 0 { 
        fmt.Fprintf(pFile, "c\n");
    } else {
        fmt.Fprintf(pFile, "c"); }

    // write additional AIG
    if p.pAigExtra {
        fmt.Fprintf(pFile, "a");
        vStrExt = Gia_AigerWriteIntoMemoryStr( p.pAigExtra );
        Gia_FileWriteBufferSize( pFile, Vec_StrSize(vStrExt) );
        fwrite( Vec_StrArray(vStrExt), 1, Vec_StrSize(vStrExt), pFile );
        Vec_StrFree( vStrExt );
        if fVerbose != 0 { fmt.Printf("Finished writing extension \"a\".\n"); }
    }
    // write constraints
    if p.nConstrs {
        fmt.Fprintf(pFile, "c");
        Gia_FileWriteBufferSize( pFile, 4 );
        Gia_FileWriteBufferSize( pFile, p.nConstrs );
    }
    // write timing information
    if p.nAnd2Delay {
        fmt.Fprintf(pFile, "d");
        Gia_FileWriteBufferSize( pFile, 4 );
        Gia_FileWriteBufferSize( pFile, p.nAnd2Delay );
    }
    if p.pManTime {
        var pTimes *float32;
        pTimes = Tim_ManGetArrTimes( (Tim_Man_t *)p.pManTime );
        if pTimes != nil {
            fmt.Fprintf(pFile, "i");
            Gia_FileWriteBufferSize( pFile, 4*Tim_ManPiNum((Tim_Man_t *)p.pManTime) );
            fwrite( pTimes, 1, 4*Tim_ManPiNum((Tim_Man_t *)p.pManTime), pFile );
            // inlining of macro ABC_FREE
            ((pTimes) ? (free((char *) (pTimes)), (pTimes) = 0) : 0);
            if fVerbose != 0 { fmt.Printf("Finished writing extension \"i\".\n"); }
        }
        pTimes = Tim_ManGetReqTimes( (Tim_Man_t *)p.pManTime );
        if pTimes != nil {
            fmt.Fprintf(pFile, "o");
            Gia_FileWriteBufferSize( pFile, 4*Tim_ManPoNum((Tim_Man_t *)p.pManTime) );
            fwrite( pTimes, 1, 4*Tim_ManPoNum((Tim_Man_t *)p.pManTime), pFile );
            // inlining of macro ABC_FREE
            ((pTimes) ? (free((char *) (pTimes)), (pTimes) = 0) : 0);
            if fVerbose != 0 { fmt.Printf("Finished writing extension \"o\".\n"); }
        }
    }
    // write equivalences
    if p.pReprs && p.pNexts {
        extern Vec_Str_t * Gia_WriteEquivClasses( Gia_Man_t * p );
        fmt.Fprintf(pFile, "e");
        vStrExt = Gia_WriteEquivClasses( p );
        Gia_FileWriteBufferSize( pFile, Vec_StrSize(vStrExt) );
        fwrite( Vec_StrArray(vStrExt), 1, Vec_StrSize(vStrExt), pFile );
        Vec_StrFree( vStrExt );
    }
    // write flop classes
    if p.vFlopClasses {
        fmt.Fprintf(pFile, "f");
        Gia_FileWriteBufferSize( pFile, 4*Gia_ManRegNum(p) );
        assert(Vec_IntSize(p.vFlopClasses) == Gia_ManRegNum(p), "Vec_IntSize(p.vFlopClasses) == Gia_ManRegNum(p)");
        fwrite( Vec_IntArray(p.vFlopClasses), 1, 4*Gia_ManRegNum(p), pFile );
    }
    // write gate classes
    if p.vGateClasses {
        fmt.Fprintf(pFile, "g");
        Gia_FileWriteBufferSize( pFile, 4*Gia_ManObjNum(p) );
        assert(Vec_IntSize(p.vGateClasses) == Gia_ManObjNum(p), "Vec_IntSize(p.vGateClasses) == Gia_ManObjNum(p)");
        fwrite( Vec_IntArray(p.vGateClasses), 1, 4*Gia_ManObjNum(p), pFile );
    }
    // write hierarchy info
    if p.pManTime {
        fmt.Fprintf(pFile, "h");
        vStrExt = Tim_ManSave( (Tim_Man_t *)p.pManTime, 1 );
        Gia_FileWriteBufferSize( pFile, Vec_StrSize(vStrExt) );
        fwrite( Vec_StrArray(vStrExt), 1, Vec_StrSize(vStrExt), pFile );
        Vec_StrFree( vStrExt );
        if fVerbose != 0 { fmt.Printf("Finished writing extension \"h\".\n"); }
    }
    // write packing
    if p.vPacking {
        extern Vec_Str_t * Gia_WritePacking( Vec_Int_t * vPacking );
        fmt.Fprintf(pFile, "k");
        vStrExt = Gia_WritePacking( p.vPacking );
        Gia_FileWriteBufferSize( pFile, Vec_StrSize(vStrExt) );
        fwrite( Vec_StrArray(vStrExt), 1, Vec_StrSize(vStrExt), pFile );
        Vec_StrFree( vStrExt );
        if fVerbose != 0 { fmt.Printf("Finished writing extension \"k\".\n"); }
    }
    // write edges
    if p.vEdge1 {
        var vPairs *Vec_Int_t  = Gia_ManEdgeToArray( p );
        var i int32;
        fmt.Fprintf(pFile, "w");
        Gia_FileWriteBufferSize( pFile, 4*(Vec_IntSize(vPairs)+1) );
        Gia_FileWriteBufferSize( pFile, Vec_IntSize(vPairs)/2 );
        for  i = 0; i < Vec_IntSize(vPairs); i++  {
            Gia_FileWriteBufferSize( pFile, Vec_IntEntry(vPairs, i) );
        }
        Vec_IntFree( vPairs );
    }
    // write mapping
    if Gia_ManHasMapping(p) {
        extern Vec_Str_t * Gia_AigerWriteMapping( Gia_Man_t * p );
        extern Vec_Str_t * Gia_AigerWriteMappingSimple( Gia_Man_t * p );
        extern Vec_Str_t * Gia_AigerWriteMappingDoc( Gia_Man_t * p );
        fmt.Fprintf(pFile, "m");
        vStrExt = Gia_AigerWriteMappingDoc( p );
        Gia_FileWriteBufferSize( pFile, Vec_StrSize(vStrExt) );
        fwrite( Vec_StrArray(vStrExt), 1, Vec_StrSize(vStrExt), pFile );
        Vec_StrFree( vStrExt );
        if fVerbose != 0 { fmt.Printf("Finished writing extension \"m\".\n"); }
    }
    // write placement
    if p.pPlacement {
        fmt.Fprintf(pFile, "p");
        Gia_FileWriteBufferSize( pFile, 4*Gia_ManObjNum(p) );
        fwrite( p.pPlacement, 1, 4*Gia_ManObjNum(p), pFile );
    }
    // write register classes
    if p.vRegClasses {
        var i int32;
        fmt.Fprintf(pFile, "r");
        Gia_FileWriteBufferSize( pFile, 4*(Vec_IntSize(p.vRegClasses)+1) );
        Gia_FileWriteBufferSize( pFile, Vec_IntSize(p.vRegClasses) );
        for  i = 0; i < Vec_IntSize(p.vRegClasses); i++  {
            Gia_FileWriteBufferSize( pFile, Vec_IntEntry(p.vRegClasses, i) );
        }
    }
    // write register inits
    if p.vRegInits {
        var i int32;
        fmt.Fprintf(pFile, "s");
        Gia_FileWriteBufferSize( pFile, 4*(Vec_IntSize(p.vRegInits)+1) );
        Gia_FileWriteBufferSize( pFile, Vec_IntSize(p.vRegInits) );
        for  i = 0; i < Vec_IntSize(p.vRegInits); i++  {
            Gia_FileWriteBufferSize( pFile, Vec_IntEntry(p.vRegInits, i) );
        }
    }
    // write configuration data
    if p.vConfigs {
        fmt.Fprintf(pFile, "b");
        assert(p.pCellStr != nil, "p.pCellStr != nil");
        Gia_FileWriteBufferSize( pFile, 4*Vec_IntSize(p.vConfigs) + len(p.pCellStr) + 1 );
        fwrite( p.pCellStr, 1, len(p.pCellStr) + 1, pFile );
//        fwrite( Vec_IntArray(p->vConfigs), 1, 4*Vec_IntSize(p->vConfigs), pFile );
        for  i = 0; i < Vec_IntSize(p.vConfigs); i++  {
            Gia_FileWriteBufferSize( pFile, Vec_IntEntry(p.vConfigs, i) );
        }
    }
    // write choices
    if Gia_ManHasChoices(p) {
        var i, nPairs int32;
        fmt.Fprintf(pFile, "q");
        for  i = 0; i < Gia_ManObjNum(p); i++  {
            nPairs += (Gia_ObjSibl(p, i) > 0);
        }
        Gia_FileWriteBufferSize( pFile, 4*(nPairs * 2 + 1) );
        Gia_FileWriteBufferSize( pFile, nPairs );
        for  i = 0; i < Gia_ManObjNum(p); i++  {
            if Gia_ObjSibl(p, i) {
                assert(i > Gia_ObjSibl(p, i), "i > Gia_ObjSibl(p, i)");
                Gia_FileWriteBufferSize( pFile, i );
                Gia_FileWriteBufferSize( pFile, Gia_ObjSibl(p, i) );
            }
        }
        if fVerbose != 0 { fmt.Printf("Finished writing extension \"q\".\n"); }
    }
    // write switching activity
    if p.pSwitching {
        fmt.Fprintf(pFile, "u");
        Gia_FileWriteBufferSize( pFile, Gia_ManObjNum(p) );
        fwrite( p.pSwitching, 1, Gia_ManObjNum(p), pFile );
    }
/*
    // write timing information
    if ( p->pManTime )
    {
        fprintf( pFile, "t" );
        vStrExt = Tim_ManSave( (Tim_Man_t *)p->pManTime, 0 );
        Gia_FileWriteBufferSize( pFile, Vec_StrSize(vStrExt) );
        fwrite( Vec_StrArray(vStrExt), 1, Vec_StrSize(vStrExt), pFile );
        Vec_StrFree( vStrExt );
    }
*/
    // write object classes
    if p.vObjClasses {
        fmt.Fprintf(pFile, "v");
        Gia_FileWriteBufferSize( pFile, 4*Gia_ManObjNum(p) );
        assert(Vec_IntSize(p.vObjClasses) == Gia_ManObjNum(p), "Vec_IntSize(p.vObjClasses) == Gia_ManObjNum(p)");
        fwrite( Vec_IntArray(p.vObjClasses), 1, 4*Gia_ManObjNum(p), pFile );
    }
    // write name
    if p.pName {
        fmt.Fprintf(pFile, "n");
        Gia_FileWriteBufferSize( pFile, len(p.pName)+1 );
        fwrite( p.pName, 1, len(p.pName), pFile );
        fmt.Fprintf(pFile, "%c", '\0');
    }
    // write comments
    if fWriteNewLine != 0 {
        fmt.Fprintf(pFile, "c\n"); }
    fmt.Fprintf(pFile, "\nThis file was produced by the GIA package in ABC on %s\n", Gia_TimeStamp());
    fmt.Fprintf(pFile, "For information about AIGER format, refer to %s\n", "http://fmv.jku.at/aiger");
    _ = (pFile).Close();
    if p != pInit {
        Gia_ManTransferTiming( pInit, p );
        Gia_ManStop( p );
    }
}

/**Function*************************************************************

  Synopsis    [Writes the AIG in the binary AIGER format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_DumpAiger(p *Gia_Man_t, pFilePrefix string, iFileNum int32, nFileNumDigits int32) {
    var Buffer string // was [100]byte in C

    Buffer = fmt.Sprintf("%s%0*d.aig", pFilePrefix, nFileNumDigits, iFileNum);
    Gia_AigerWrite( p, Buffer, 0, 0, 0 );
}

/**Function*************************************************************

  Synopsis    [Writes the AIG in the binary AIGER format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_AigerWriteSimple(pInit *Gia_Man_t, pFileName string) {
    var pFile *os.File;
    var vStr *Vec_Str_t;
    if Gia_ManPoNum(pInit) == 0 {
        fmt.Printf("Gia_AigerWriteSimple(): AIG cannot be written because it has no POs.\n");
        return;
    }
    // start the output stream
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Gia_AigerWriteSimple(): Cannot open the output file \"%s\".\n", pFileName);
        return;
    }
    // write the buffer
    vStr = Gia_AigerWriteIntoMemoryStr( pInit );
    fwrite( Vec_StrArray(vStr), 1, Vec_StrSize(vStr), pFile );
    Vec_StrFree( vStr );
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

