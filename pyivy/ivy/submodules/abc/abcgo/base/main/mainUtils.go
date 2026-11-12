package main

import "fmt"

/**CFile****************************************************************

  FileName    [mainUtils.c]

  SystemName  [ABC: Logic synthesis and verification system.]
 
  PackageName [The main package.]

  Synopsis    [Miscellaneous utilities.]
 
  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: mainUtils.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "mainInt.h"

#ifdef ABC_USE_READLINE
#include <readline/readline.h>
#include <readline/history.h>
#endif

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static char * DateReadFromDateString( char * datestr );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_UtilsGetVersion_Version string // was [1000]byte in C

func Abc_UtilsGetVersion(pAbc *Abc_Frame_t) string {
    
    spStatic_Abc_UtilsGetVersion_Version = fmt.Sprintf("%s (compiled %s %s)", ABC_VERSION, __DATE__, __TIME__);
    return spStatic_Abc_UtilsGetVersion_Version;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_UtilsGetUsersInput_Prompt string // was [5000]byte in C

func Abc_UtilsGetUsersInput(pAbc *Abc_Frame_t) string {
    
    spStatic_Abc_UtilsGetUsersInput_Prompt = fmt.Sprintf("abc %02d> ", pAbc.nSteps);
#ifdef ABC_USE_READLINE
    {
    static char * line = NULL;
    if (line != NULL) ABC_FREE(line);
    line = readline(Prompt);  
    if (line == NULL){ printf("***EOF***\n"); exit(0); }
    add_history(line);
    return line;
    }
#else
    {
    char * pRetValue;
    fprintf( pAbc->Out, "%s", Prompt );
    pRetValue = fgets( Prompt, 5000, stdin );
    return Prompt;
    }
#endif
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_UtilsPrintHello(pAbc *Abc_Frame_t) {
    fmt.Fprintf(pAbc.Out, "%s\n", pAbc.sVersion);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_UtilsPrintUsage(pAbc *Abc_Frame_t, ProgName string) {
    fmt.Fprintf(pAbc.Err, "\n");
    fmt.Fprintf(pAbc.Err, "usage: %s [-c cmd] [-q cmd] [-C cmd] [-Q cmd] [-f script] [-h] [-o file] [-s] [-t type] [-T type] [-x] [-b] [file]\n", ProgName);
    fmt.Fprintf(pAbc.Err, "    -c cmd\texecute commands `cmd'\n");
    fmt.Fprintf(pAbc.Err, "    -q cmd\texecute commands `cmd' quietly\n");
    fmt.Fprintf(pAbc.Err, "    -C cmd\texecute commands `cmd', then continue in interactive mode\n");
    fmt.Fprintf(pAbc.Err, "    -Q cmd\texecute commands `cmd' quietly, then continue in interactive mode\n");
    fmt.Fprintf(pAbc.Err, "    -F script\texecute commands from a script file and echo commands\n");
    fmt.Fprintf(pAbc.Err, "    -f script\texecute commands from a script file\n");
    fmt.Fprintf(pAbc.Err, "    -h\t\tprint the command usage\n");
    fmt.Fprintf(pAbc.Err, "    -o file\tspecify output filename to store the result\n");
    fmt.Fprintf(pAbc.Err, "    -s\t\tdo not read any initialization file\n");
    fmt.Fprintf(pAbc.Err, "    -t type\tspecify input type (blif_mv (default), blif_mvs, blif, or none)\n");
    fmt.Fprintf(pAbc.Err, "    -T type\tspecify output type (blif_mv (default), blif_mvs, blif, or none)\n");
    fmt.Fprintf(pAbc.Err, "    -x\t\tequivalent to '-t none -T none'\n");
    fmt.Fprintf(pAbc.Err, "    -b\t\trunning in bridge mode\n");
    fmt.Fprintf(pAbc.Err, "\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_UtilsSource(pAbc *Abc_Frame_t) {  
#ifdef WIN32
    if ( Cmd_CommandExecute(pAbc, "source abc.rc") )
    {
        if ( Cmd_CommandExecute(pAbc, "source ..\\abc.rc") == 0 )
            printf( "Loaded \"abc.rc\" from the parent directory.\n" );
        else if ( Cmd_CommandExecute(pAbc, "source ..\\..\\abc.rc") == 0 )
            printf( "Loaded \"abc.rc\" from the grandparent directory.\n" );
    }
#else

#if 0
    {
        char * sPath1, * sPath2;

         // If .rc is present in both the home and current directories, then read
         // it from the home directory.  Otherwise, read it from wherever it's located.
        sPath1 = Extra_UtilFileSearch(".rc", "~/", "r");
        sPath2 = Extra_UtilFileSearch(".rc", ".",  "r");
  
        if ( sPath1 && sPath2 ) {
            /* ~/.rc == .rc : Source the file only once */
            (void) Cmd_CommandExecute(pAbc, "source -s ~/.rc");
        }
        else {
            if (sPath1) {
                (void) Cmd_CommandExecute(pAbc, "source -s ~/.rc");
            }
            if (sPath2) {
                (void) Cmd_CommandExecute(pAbc, "source -s .rc");
            }
        }
        if ( sPath1 ) ABC_FREE(sPath1);
        if ( sPath2 ) ABC_FREE(sPath2);
    
        /* execute the abc script which can be open with the "open_path" */
        Cmd_CommandExecute( pAbc, "source -s abc.rc" );
    }
#endif

    {
        char * sPath1, * sPath2;
        char * home;

         // If .rc is present in both the home and current directories, then read
         // it from the home directory.  Otherwise, read it from wherever it's located.
        home = getenv("HOME");
        if (home){
            char * sPath3 = ABC_ALLOC(char, strlen(home) + 2);
            (void) sprintf(sPath3, "%s/", home);
            sPath1 = Extra_UtilFileSearch(".abc.rc", sPath3, "r");
            ABC_FREE(sPath3);
        }else
            sPath1 = NULL;

        sPath2 = Extra_UtilFileSearch(".abc.rc", ".",  "r");

        if ( sPath1 && sPath2 ) {
            /* ~/.rc == .rc : Source the file only once */
            char *tmp_cmd = ABC_ALLOC(char, strlen(sPath1)+12);
            (void) sprintf(tmp_cmd, "source -s %s", sPath1);
            // (void) Cmd_CommandExecute(pAbc, "source -s ~/.abc.rc");
            (void) Cmd_CommandExecute(pAbc, tmp_cmd);
            ABC_FREE(tmp_cmd);
        }
        else {
            if (sPath1) {
                char *tmp_cmd = ABC_ALLOC(char, strlen(sPath1)+12);
                (void) sprintf(tmp_cmd, "source -s %s", sPath1);
                // (void) Cmd_CommandExecute(pAbc, "source -s ~/.abc.rc");
                (void) Cmd_CommandExecute(pAbc, tmp_cmd);
                ABC_FREE(tmp_cmd);
            }
            if (sPath2) {
                char *tmp_cmd = ABC_ALLOC(char, strlen(sPath2)+12);
                (void) sprintf(tmp_cmd, "source -s %s", sPath2);
                // (void) Cmd_CommandExecute(pAbc, "source -s .abc.rc");
                (void) Cmd_CommandExecute(pAbc, tmp_cmd);
                ABC_FREE(tmp_cmd);
            }
        }
        if ( sPath1 ) ABC_FREE(sPath1);
        if ( sPath2 ) ABC_FREE(sPath2);

        /* execute the abc script which can be open with the "open_path" */
        Cmd_CommandExecute( pAbc, "source -s abc.rc" );
    }
    
#endif //WIN32
}

/**Function********************************************************************

  Synopsis    [Returns the date in a brief format assuming its coming from
  the program `date'.]

  Description [optional]

  SideEffects []

******************************************************************************/
var spStatic_DateReadFromDateString_result [25]byte;
func DateReadFromDateString(datestr string) string {
  
  var day [10]byte;
  var month [10]byte;
  var zone [10]byte;
  var at string;
  var date int32;
  var hour int32;
  var minute int32;
  var second int32;
  var year int32;

  if sscanf(datestr, "%s %s %2d %2d:%2d:%2d %s %4d",
               day, month, &date, &hour, &minute, &second, zone, &year) == 8 {
    if hour >= 12 {
      if hour >= 13 { hour -= 12; }
      at = "PM";
    
    } else {
      if hour == 0 { hour = 12; }
      at = "AM";
    }
    (void) sprintf(spStatic_DateReadFromDateString_result, "%d-%3s-%02d at %d:%02d %s", 
                       date, month, year % 100, hour, minute, at);
    return spStatic_DateReadFromDateString_result;
  
  } else {
    return datestr;
  }
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
