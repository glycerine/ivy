package cmd

/**CFile****************************************************************

  FileName    [cmd.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [External declarations of the command package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cmd.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

type Abc_Command =    MvCommand;  // one command
type Abc_Alias =      MvAlias;    // one alias

 type Cmd_CommandId_t int32
type Cmd_CommandId_t_ = Cmd_CommandId_t
const 
(
    CMD_COMMAND_UNKNOWN Cmd_CommandId_t = 0

    CMD_COMMAND_ALIAS Cmd_CommandId_t = CMD_COMMAND_UNKNOWN + 1

    CMD_COMMAND_SOURCE Cmd_CommandId_t = CMD_COMMAND_ALIAS + 1

    CMD_COMMAND_SET Cmd_CommandId_t = CMD_COMMAND_SOURCE + 1

    CMD_COMMAND_READ_BLIF Cmd_CommandId_t = CMD_COMMAND_SET + 1

    CMD_COMMAND_READ_AIGER Cmd_CommandId_t = CMD_COMMAND_READ_BLIF + 1

    CMD_COMMAND_STRASH Cmd_CommandId_t = CMD_COMMAND_READ_AIGER + 1

    CMD_COMMAND_PRINT_STATS Cmd_CommandId_t = CMD_COMMAND_STRASH + 1

    CMD_COMMAND_PDR Cmd_CommandId_t = CMD_COMMAND_PRINT_STATS + 1

    CMD_COMMAND_WRITE_AIGER_CEX Cmd_CommandId_t = CMD_COMMAND_PDR + 1

)

 

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== cmd.c ===========================================================*/
// extern void        Cmd_Init( Abc_Frame_t * pAbc );

// extern void        Cmd_End( Abc_Frame_t * pAbc );

/*=== cmdApi.c ========================================================*/
typedef int (*Cmd_CommandFuncType)(Abc_Frame_t*, int, char**);
// extern int         Cmd_CommandIsDefined( Abc_Frame_t * pAbc, const char * sName );

// extern void        Cmd_CommandAdd( Abc_Frame_t * pAbc, const char * sGroup, const char * sName, Cmd_CommandFuncType pFunc, int fChanges );

extern ABC_DLL int Cmd_CommandExecute( Abc_Frame_t * pAbc, char * sCommand );
/*=== cmdFlag.c ========================================================*/
// extern char *      Cmd_FlagReadByName( Abc_Frame_t * pAbc, char * flag );

// extern void        Cmd_FlagDeleteByName( Abc_Frame_t * pAbc, const char * key );

// extern void        Cmd_FlagUpdateValue( Abc_Frame_t * pAbc, const char * key, char * value );

/*=== cmdHist.c ========================================================*/
// extern void          Cmd_HistoryAddCommand( Abc_Frame_t * pAbc, const char * command );

// extern void        Cmd_HistoryRead( Abc_Frame_t * p );

// extern void        Cmd_HistoryWrite( Abc_Frame_t * p, int Limit );

// extern void        Cmd_HistoryPrint( Abc_Frame_t * p, int Limit );

/*=== cmdLoad.c ========================================================*/
// extern int         CmdCommandLoad( Abc_Frame_t * pAbc, int argc, char ** argv );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
