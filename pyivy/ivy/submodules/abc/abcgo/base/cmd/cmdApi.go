package cmd

import (
	"fmt"
	"strconv"
	"strings"
)

/**CFile****************************************************************

  FileName    [cmdApi.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [External procedures of the command package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cmdApi.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/mainInt.h"
//#include "cmdInt.h"

//ABC_NAMESPACE_IMPL_START

func Cmd_CommandNameToId(sName string) Cmd_CommandId_t {
    if !(int32(strings.Compare(sName, "alias")) != 0) {
        return CMD_COMMAND_ALIAS; }
    if !(int32(strings.Compare(sName, "source")) != 0) {
        return CMD_COMMAND_SOURCE; }
    if !(int32(strings.Compare(sName, "set")) != 0) {
        return CMD_COMMAND_SET; }
    if !(int32(strings.Compare(sName, "read_blif")) != 0) {
        return CMD_COMMAND_READ_BLIF; }
    if !(int32(strings.Compare(sName, "read_aiger")) != 0) {
        return CMD_COMMAND_READ_AIGER; }
    if !(int32(strings.Compare(sName, "strash")) != 0) {
        return CMD_COMMAND_STRASH; }
    if !(int32(strings.Compare(sName, "print_stats")) != 0) {
        return CMD_COMMAND_PRINT_STATS; }
    if !(int32(strings.Compare(sName, "pdr")) != 0) {
        return CMD_COMMAND_PDR; }
    if !(int32(strings.Compare(sName, "write_aiger_cex")) != 0) {
        return CMD_COMMAND_WRITE_AIGER_CEX; }
    return CMD_COMMAND_UNKNOWN;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cmd_CommandIsDefined(pAbc *Abc_Frame_t, sName string) int32 {
    return st__is_member( pAbc.tCommands, sName );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cmd_CommandAdd(pAbc *Abc_Frame_t, sGroup string, sName string, pFunc Cmd_CommandFuncType, fChanges int32) {
    char * key;
    var value string;
    var pCommand *Abc_Command;
    var fStatus int32;

    key = sName;
    if st__delete( pAbc.tCommands, &key, &value ) {
        // delete existing definition for this command 
        fmt.Fprintf(pAbc.Err, "Cmd warning: redefining '%s'\n", sName);
        CmdCommandFree( (Abc_Command *)value );
    }

    // create the new command
    pCommand = new(Abc_Command);
    pCommand.sName   = Extra_UtilStrsav( sName );
    pCommand.sGroup  = Extra_UtilStrsav( sGroup );
    pCommand.Id      = Cmd_CommandNameToId( sName );
    pCommand.pFunc   = pFunc;
    pCommand.fChange = fChanges;
    fStatus = st__insert( pAbc.tCommands, pCommand.sName, (char *)pCommand );
    assert(!(fStatus != 0), "!fStatus");  // the command should not be in the table
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cmd_CommandHandleSpecial(pAbc *Abc_Frame_t, sCommand string) int32 {
    var pNtk *Abc_Ntk_t  = Abc_FrameReadNtk(pAbc);
    var piCountNew int32;
    if (pNtk != nil) && Abc_NtkHasMapping(pNtk) {
        piCountNew = Abc_NtkCiNum(pNtk);
    } else {
        piCountNew = 0;
    }
    var piCount int32;
    var poCountNew int32;
    if (pNtk != nil) && Abc_NtkHasMapping(pNtk) {
        poCountNew = Abc_NtkCoNum(pNtk);
    } else {
        poCountNew = 0;
    }
    var poCount int32;
    var ndCountNew int32;
    if (pNtk != nil) && Abc_NtkHasMapping(pNtk) {
        ndCountNew = Abc_NtkNodeNum(pNtk);
    } else {
        ndCountNew = 0;
    }
    var ndCount int32;
    var AreaNew float64;
    if (pNtk != nil) && Abc_NtkHasMapping(pNtk) {
        AreaNew = Abc_NtkGetMappedArea(pNtk);
    } else {
        AreaNew = 0;
    }
    var Area float64;
    var DepthNew int32;
    if (pNtk != nil) && Abc_NtkHasMapping(pNtk) {
        DepthNew = Abc_NtkLevel(pNtk);
    } else {
        DepthNew = 0;
    }
    var Depth int32;
    if strstr(sCommand, "#PS") {
        fmt.Printf("pi=%d ", piCountNew);
        fmt.Printf("po=%d ", poCountNew);
        fmt.Printf("fn=%d ", ndCountNew);
        fmt.Printf("ma=%.1f ", AreaNew);
        fmt.Printf("de=%d ", DepthNew);
        fmt.Printf("\n");
        return 1;
    }
    if strstr(sCommand, "#CEC") {
        //int proofStatus = Abc_NtkVerifyUsingCec(pNtk);
        var proofStatus int32  = 1;
        // -1 (undecided), 0 (different), 1 (equivalent)
        fmt.Printf("proofStatus=%d\n", proofStatus);
        return 1;
    }
    if strstr(sCommand, "#ASSERT") {
        var Status int32;
        var pNumb string   = strrchr( (char *)sCommand, '=' );
        if strstr(sCommand, "_PI_") {
            if pNumb {
                piCount = strconv.Atoi(pNumb+1);
            } else {
                piCount = 0;
            }
            if strstr( sCommand, "==" ) {
                Status = piCountNew == piCount;
            } else if strstr( sCommand, "<=" ) {
                Status = piCountNew <= piCount;
            } else { return 0; }
        
        } else if strstr(sCommand, "_PO_") {
            if pNumb {
                poCount = strconv.Atoi(pNumb+1);
            } else {
                poCount = 0;
            }
            if strstr( sCommand, "==" ) {
                Status = poCountNew == poCount;
            } else if strstr( sCommand, "<=" ) {
                Status = poCountNew <= poCount;
            } else { return 0; }
        
        } else if strstr(sCommand, "_NODE_") {
            if pNumb {
                ndCount = strconv.Atoi(pNumb+1);
            } else {
                ndCount = 0;
            }
            if strstr( sCommand, "==" ) {
                Status = ndCountNew == ndCount;
            } else if strstr( sCommand, "<=" ) {
                Status = ndCountNew <= ndCount;
            } else { return 0; }
        
        } else if strstr(sCommand, "_AREA_") {
            var Eplison float64  = 1.0;
            if pNumb {
                Area = atof(pNumb+1);
            } else {
                Area = 0;
            }
            if strstr( sCommand, "==" ) {
                Status = AreaNew >= Area - Eplison && AreaNew <= Area + Eplison;
            } else if strstr( sCommand, "<=" ) {
                Status = AreaNew <= Area + Eplison;
            } else { return 0; }
        
        } else if strstr(sCommand, "_DEPTH_") {
            if pNumb {
                Depth = strconv.Atoi(pNumb+1);
            } else {
                Depth = 0;
            }
            if strstr( sCommand, "==" ) {
                Status = DepthNew == Depth;
            } else if strstr( sCommand, "<=" ) {
                Status = DepthNew <= Depth;
            } else { return 0; }
        
        } else { return 0; }
        fmt.Printf("%s\n", Status ? "succeeded" : "failed");
        return 1;
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cmd_CommandExecute(pAbc *Abc_Frame_t, sCommand string) int32 {
    var fStatus, argc, loop int32;
    char * sCommandNext;
    var argv *string;

    if !pAbc.fAutoexac && !pAbc.fSource { 
        Cmd_HistoryAddCommand(pAbc, sCommand); }
    sCommandNext = sCommand;
    for 
    {
        if sCommandNext[0] == '#' && (Cmd_CommandHandleSpecial( pAbc, sCommandNext ) != 0) {
            break; }
        sCommandNext = CmdSplitLine( pAbc, sCommandNext, &argc, &argv );
        loop = 0;
        fStatus = CmdApplyAlias( pAbc, &argc, &argv, &loop );
        if fStatus == 0 { 
            fStatus = CmdCommandDispatch( pAbc, &argc, &argv ); }
        CmdFreeArgv( argc, argv );
    
if !( fStatus == 0 && *sCommandNext != '\0' ) { break }
} 
    
    return fStatus;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
