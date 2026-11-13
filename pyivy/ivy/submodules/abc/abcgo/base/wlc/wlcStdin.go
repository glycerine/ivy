package wlc

import (
	"fmt"
	"os"
	"strings"
)

/**CFile****************************************************************

  FileName    [wlcStdin.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [stdin processing for STM interface.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcStdin.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Converts a bit-string into a number in a given radix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Report results.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_NtkReport(p *Wlc_Ntk_t, pCex *Abc_Cex_t, pName string, Radix int32) {
    var vNum *Vec_Str_t;
    var i int32;
    var NameId int32;
    var Name int32;
    var Start int32 = -1;
    var nBits int32 = -1;
    assert(pCex.nRegs == 0, "pCex.nRegs == 0");
    // get the name ID
    NameId = Abc_NamStrFind( p.pManName, pName );
    if NameId <= 0 {
        fmt.Printf("Cannot find \"%s\" among nodes of the network.\n", pName);
        return;
    }
    // get the primary input
    Vec_IntForEachEntryTriple( &p.vValues, Name, Start, nBits, i )
        if ( Name == NameId )
            break;
    // find the PI
    if i == Vec_IntSize(&p.vValues) {
        fmt.Printf("Cannot find \"%s\" among primary inputs of the network.\n", pName);
        return;
    }
    // print value
    assert(Radix == 2 || Radix == 10 || Radix == 16, "Radix == 2 || Radix == 10 || Radix == 16");
    vNum = Wlc_ConvertToRadix( pCex.pData, Start, nBits, Radix );
    fmt.Printf("((%s %s%s))\n", pName, Radix == 16 ? "#x" : (Radix == 2 ? "#b" : ""), Vec_StrArray(vNum));
    Vec_StrFree( vNum );
}

/**Function********************************************************************

  Synopsis    [Reads stdin and stops when "(check-sat)" is observed.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Wlc_StdinCollectStop(vInput *Vec_Str_t, pLine string, LineSize int32) int32 {
    var pStr string  = Vec_StrArray(vInput) + Vec_StrSize(vInput) - LineSize;
    if Vec_StrSize(vInput) < LineSize {
        return 0; }
    return !strncmp( pStr, pLine, LineSize );
}
func Wlc_StdinCollectProblem(pDir string) *Vec_Str_t {
    var c int32;
    var DirSize int32 = len(pDir);
    var vInput *Vec_Str_t  = Vec_StrAlloc( 1000 );
    for  (c = fgetc(os.Stdin)) != -1  {
        Vec_StrPush( vInput, byte(c) );
        if c == ')' && (Wlc_StdinCollectStop(vInput, pDir, DirSize) != 0) {
            break; }
    }
    Vec_StrPush( vInput, '\0' );
    return vInput;
}
func Wlc_StdinCollectQuery() *Vec_Str_t {
    var c, Count, fSomeInput int32;
    var vInput *Vec_Str_t  = Vec_StrAlloc( 1000 );
    for  (c = fgetc(os.Stdin)) != -1  {
        Vec_StrPush( vInput, byte(c) );
        if c == '(' {
            Count++, fSomeInput = 1;
        } else if c == ')' {
            Count--; }
        if Count == 0 && (fSomeInput != 0) {
            break; }
    }
    if c == -1 {
        Vec_StrFreeP( &vInput );
    } else {
        Vec_StrPush( vInput, '\0' ); }
    return vInput;
}
func Wlc_StdinProcessSmt(pAbc *Abc_Frame_t, pCmd string) int32 {
    // collect stdin until (check-sat)
    var vInput *Vec_Str_t  = Wlc_StdinCollectProblem( "(check-sat)" );
    // parse input
    var pNtk *Wlc_Ntk_t  = Wlc_ReadSmtBuffer( "top", Vec_StrArray(vInput), Vec_StrArray(vInput) + Vec_StrSize(vInput), 0, 0 );
    Vec_StrFree( vInput );
    // install current network
    Wlc_SetNtk( pAbc, pNtk );
    // execute command
    if Cmd_CommandExecute(pAbc, pCmd) {
        Abc_Print( 1, "Something did not work out with the command \"%s\".\n", pCmd );
        return 0;
    }
    // solver finished
    if Abc_FrameReadProbStatus(pAbc) == -1 {
        fmt.Printf("undecided\n");
    } else if Abc_FrameReadProbStatus(pAbc) == 1 {
        fmt.Printf("unsat\n");
    } else if Abc_FrameReadProbStatus(pAbc) == 0 {
        fmt.Printf("sat\n");
    } else { assert(false, "0"); }
    fflush( os.Stdout );
    // wait for stdin for give directions
    for  (vInput = Wlc_StdinCollectQuery()) != nil  {
        var pName string  = strtok( Vec_StrArray(vInput), " \n\t\r()" );
        // check directive 
        if int32(strings.Compare(pName, "get-value")) != 0 {
            Abc_Print( 1, "ABC is expecting \"get-value\" in a follow-up input of the satisfiable problem.\n" );
            Vec_StrFree( vInput );
            return 0;
        }
        // check status
        if Abc_FrameReadProbStatus(pAbc) != 0 {
            Abc_Print( 1, "ABC received a follow-up input for a problem that is not known to be satisfiable.\n" );
            Vec_StrFree( vInput );
            return 0;
        }
        // get the variable number
        pName = strtok( nil, "() \n\t\r" );
        // get the counter-example
        if Abc_FrameReadCex(pAbc) == nil {
            Abc_Print( 1, "ABC does not have a counter-example available to process a \"get-value\" request.\n" );
            Vec_StrFree( vInput );
            return 0;
        }
        // report value of this variable
        Wlc_NtkReport( (Wlc_Ntk_t *)pAbc.pAbcWlc, (Abc_Cex_t *)Abc_FrameReadCex(pAbc), pName, 16 );
        Vec_StrFree( vInput );
        fflush( os.Stdout );
    }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

