package abci

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [abcLog.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Log file printing.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcLog.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "aig/gia/gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/*
    Log file format (Jiang, Mon, 28 Sep 2009;  updated by Alan in Jan 2011)

    <result> <bug_free_depth>  <engine_name> <0-based_output_num> <0-based_frame>
    <INIT_STATE> : default is empty line.
    <TRACE> : default is empty line
   
    <result> is one of the following: "snl_SAT", "snl_UNSAT", "snl_UNK", "snl_ABORT".
    <bug_free_depth> is the number of timeframes exhaustively explored without counter-examples
    <0-based_output_num> only need to be given if the problem is SAT.
    <0-based_frame> only need to be given if the problem is SAT and <0-based_frame> is different from <bug_free_depth>.   
    <INIT_STATE>  : initial state
    <TRACE> : input vector

    <INIT_STATE>and <TRACE> are strings of 0/1/- ( - means don't care). The length is equivalent to #input*#<cyc>.
*/


         



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkWriteLogFile(pFileName string, pCex *Abc_Cex_t, Status int32, nFrames int32, pCommand string) {
    var pFile *os.File;
    var i int32;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open log file for writing \"%s\".\n", pFileName);
        return;
    }
    // write <result>
    if Status == 1 {
        fmt.Fprintf(pFile, "snl_UNSAT");
    } else if Status == 0 {
        fmt.Fprintf(pFile, "snl_SAT");
    } else if Status == -1 {
        fmt.Fprintf(pFile, "snl_UNK");
    } else { 
        fmt.Printf("Abc_NtkWriteLogFile(): Cannot recognize solving status.\n"); }
    fmt.Fprintf(pFile, " ");
    // write <bug_free_depth>
    fmt.Fprintf(pFile, "%d", nFrames);
    fmt.Fprintf(pFile, " ");
    // write <engine_name>
    fmt.Fprintf(pFile, "%s", pCommand ? pCommand : "unknown");
    if (pCex != nil) && Status == 0 {
        fmt.Fprintf(pFile, " %d", pCex.iPo); }
    // write <cyc>
    if (pCex != nil) && pCex.iFrame != nFrames {
        fmt.Fprintf(pFile, " %d", pCex.iFrame); }
    fmt.Fprintf(pFile, "\n");
    // write <INIT_STATE>
    if pCex == nil {
        fmt.Fprintf(pFile, "NULL");
    } else {
        for  i = 0; i < pCex.nRegs; i++  {
            fmt.Fprintf(pFile, "%d", Abc_InfoHasBit(pCex.pData,i));
        }
    }
    fmt.Fprintf(pFile, "\n");
    // write <TRACE>
    if pCex == nil {
        fmt.Fprintf(pFile, "NULL");
    } else {
        assert(pCex.nBits - pCex.nRegs == pCex.nPis * (pCex.iFrame + 1), "pCex.nBits - pCex.nRegs == pCex.nPis * (pCex.iFrame + 1)");
        for  i = pCex.nRegs; i < pCex.nBits; i++  {
            fmt.Fprintf(pFile, "%d", Abc_InfoHasBit(pCex.pData,i));
        }
    }
    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkReadLogFile(pFileName string, ppCex **Abc_Cex_t, pnFrames *int32) int32 {
    var pFile *os.File;
    var pCex *Abc_Cex_t;
    var vNums *Vec_Int_t;
    var Buffer [1000]byte;
    var pToken string;
    var RetValue string;
    var c int32;
    var nRegs int32 = -1;
    var nFrames int32 = -1;
    var iPo int32 = -1;
    var Status int32 = -1;
    var nFrames2 int32 = -1;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open log file for reading \"%s\".\n", pFileName);
        return -1;
    }
    RetValue = fgets( Buffer, 1000, pFile );
    if !strncmp( Buffer, "snl_UNSAT", len("snl_UNSAT") ) {
        Status = 1;
        nFrames = strconv.Atoi(Buffer + len("snl_UNSAT")); 
    
    } else if !strncmp( Buffer, "snl_SAT", len("snl_SAT") ) {
        Status = 0;
//        nFrames = atoi( Buffer + strlen("snl_SAT") ); 
        pToken  = strtok( Buffer + len("snl_SAT"), " \t\n" );
        nFrames = strconv.Atoi(pToken); 
        pToken  = strtok( nil, " \t\n" );
        pToken  = strtok( nil, " \t\n" );
        if pToken != nil {
            iPo     = strconv.Atoi(pToken); 
            pToken  = strtok( nil, " \t\n" );
            if pToken {
                nFrames2 = strconv.Atoi(pToken); } 
        }
//        else
//            printf( "Warning! The current status is SAT but the current CEX is not given.\n"  );
    
    } else if !strncmp( Buffer, "snl_UNK", len("snl_UNK") ) {
        Status = -1;
        nFrames = strconv.Atoi(Buffer + len("snl_UNK")); 
    
    } else {
        fmt.Printf("Unrecognized status.\n");
    }
    // found regs till the new line
    vNums = Vec_IntAlloc( 100 );
    for  (c = fgetc(pFile)) != -1  {
        if c == '\n' {
            break; }
        if c == '0' || c == '1' {
            Vec_IntPush( vNums, c - '0' ); }
    }
    nRegs = Vec_IntSize(vNums);
    // skip till the new line
    for  (c = fgetc(pFile)) != -1  {
        if c == '0' || c == '1' {
            Vec_IntPush( vNums, c - '0' ); }
    }
    _ = (pFile).Close();
    if Vec_IntSize(vNums) {
        var iFrameCex int32;
        if nFrames2 == -1 {
            iFrameCex = nFrames;
        } else {
            iFrameCex = nFrames2;
        }
        if nRegs < 0 {
            fmt.Printf("Cannot read register number.\n");
            Vec_IntFree( vNums );
            return -1;
        }
        if Vec_IntSize(vNums)-nRegs == 0 {
            fmt.Printf("Cannot read counter example.\n");
            Vec_IntFree( vNums );
            return -1;
        }
        if (Vec_IntSize(vNums)-nRegs) % (iFrameCex + 1) != 0 {
            fmt.Printf("Incorrect number of bits.\n");
            Vec_IntFree( vNums );
            return -1;
        }
        pCex = Abc_CexAlloc( nRegs, (Vec_IntSize(vNums)-nRegs)/(iFrameCex + 1), iFrameCex + 1 );
        pCex.iPo    = iPo;
        pCex.iFrame = iFrameCex;
        assert(Vec_IntSize(vNums) == pCex.nBits, "Vec_IntSize(vNums) == pCex.nBits");
        for  c = 0; c < pCex.nBits; c++  {
            if Vec_IntEntry(vNums, c) {
                Abc_InfoSetBit( pCex.pData, c ); }
        }
        Vec_IntFree( vNums );
        if ppCex {
            *ppCex = pCex;
        } else {
            pCex = nil; }
    
    } else {                                                         
        // corner case of seq circuit with no PIs             
        var iFrameCex int32;
        if nFrames2 == -1 {
            iFrameCex = nFrames;
        } else {
            iFrameCex = nFrames2;
        }
        pCex = Abc_CexAlloc( 0, 0, iFrameCex + 1 );           
        pCex.iFrame = iFrameCex;                             
        pCex.iPo    = iPo;                                   
        if ppCex {                                          
            *ppCex = pCex; }                                    
        Vec_IntFree( vNums );                                 
    }                                                         
    if pnFrames != nil {
        *pnFrames = nFrames; }
    return Status;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

