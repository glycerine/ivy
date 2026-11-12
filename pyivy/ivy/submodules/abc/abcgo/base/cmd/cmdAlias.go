package cmd

import "fmt"

/**CFile****************************************************************

  FileName    [cmdAlias.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures dealing with aliases in the command package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cmdAlias.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "cmdInt.h"

//ABC_NAMESPACE_IMPL_START


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
func CmdCommandAliasAdd(pAbc *Abc_Frame_t, sName string, argc int32, argv *string) {
    var pAlias *Abc_Alias;
    var fStatus, i int32;

    pAlias = new(Abc_Alias);
    pAlias.sName = Extra_UtilStrsav(sName);
    pAlias.argc = argc;
    pAlias.argv = ABC_ALLOC(char *, pAlias.argc);
    for i = 0; i < argc; i++ { 
        pAlias.argv[i] = Extra_UtilStrsav(argv[i]);
    }
    fStatus = st__insert( pAbc.tAliases, pAlias.sName, (char *) pAlias );
    assert(!fStatus, "!fStatus");  
}

/**Function********************************************************************

  Synopsis    [required]

  Description [optional]

  SideEffects [required]

  SeeAlso     [optional]

******************************************************************************/
func CmdCommandAliasPrint(pAbc *Abc_Frame_t, pAlias *Abc_Alias) {
    var i int32;
    fmt.Fprintf(pAbc.Out, "%-15s", pAlias.sName);
    for i = 0; i < pAlias.argc; i++ { 
        fmt.Fprintf(pAbc.Out, " %s", pAlias.argv[i]);
    }
    fmt.Fprintf(pAbc.Out, "\n");
}

/**Function********************************************************************

  Synopsis    [required]

  Description [optional]

  SideEffects [required]

  SeeAlso     [optional]

******************************************************************************/
func CmdCommandAliasLookup(pAbc *Abc_Frame_t, sCommand string) string {
  var pAlias *Abc_Alias;
  var value string;
  if ! st__lookup( pAbc.tAliases, sCommand, &value) { 
    return sCommand; }
  pAlias = (Abc_Alias *) value;
  return pAlias.argv[0];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdCommandAliasFree(pAlias *Abc_Alias) {
    CmdFreeArgv( pAlias.argc, pAlias.argv );
    ABC_FREE(pAlias.sName);    
    ABC_FREE(pAlias);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

