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

#ifndef ABC__base__cmd__cmd_h
#define ABC__base__cmd__cmd_h


////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

typedef struct MvCommand    Abc_Command;  // one command
typedef struct MvAlias      Abc_Alias;    // one alias

typedef enum Cmd_CommandId_t_
{
    CMD_COMMAND_UNKNOWN = 0,
    CMD_COMMAND_ALIAS,
    CMD_COMMAND_SOURCE,
    CMD_COMMAND_SET,
    CMD_COMMAND_READ_BLIF,
    CMD_COMMAND_READ_AIGER,
    CMD_COMMAND_STRASH,
    CMD_COMMAND_PRINT_STATS,
    CMD_COMMAND_PDR,
    CMD_COMMAND_WRITE_AIGER_CEX
} Cmd_CommandId_t;

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== cmd.c ===========================================================*/
extern void        Cmd_Init( Abc_Frame_t * pAbc );
extern void        Cmd_End( Abc_Frame_t * pAbc );
/*=== cmdApi.c ========================================================*/
typedef int (*Cmd_CommandFuncType)(Abc_Frame_t*, int, char**);
extern int         Cmd_CommandIsDefined( Abc_Frame_t * pAbc, const char * sName );
extern void        Cmd_CommandAdd( Abc_Frame_t * pAbc, const char * sGroup, const char * sName, Cmd_CommandFuncType pFunc, int fChanges );
extern ABC_DLL int Cmd_CommandExecute( Abc_Frame_t * pAbc, const char * sCommand );
/*=== cmdFlag.c ========================================================*/
extern char *      Cmd_FlagReadByName( Abc_Frame_t * pAbc, char * flag );
extern void        Cmd_FlagDeleteByName( Abc_Frame_t * pAbc, const char * key );
extern void        Cmd_FlagUpdateValue( Abc_Frame_t * pAbc, const char * key, char * value );
/*=== cmdHist.c ========================================================*/
extern void          Cmd_HistoryAddCommand( Abc_Frame_t * pAbc, const char * command );
extern void        Cmd_HistoryRead( Abc_Frame_t * p );
extern void        Cmd_HistoryWrite( Abc_Frame_t * p, int Limit );
extern void        Cmd_HistoryPrint( Abc_Frame_t * p, int Limit );
/*=== cmdLoad.c ========================================================*/
extern int         CmdCommandLoad( Abc_Frame_t * pAbc, int argc, char ** argv );



ABC_NAMESPACE_HEADER_END



#endif

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
