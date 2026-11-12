package extra

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [extraUtilUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Old SIS utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilUtil.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <string.h>
//#include "extra.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const EXTRA_RLIMIT_DATA_DEFAULT = 67108864  // assume 64MB by default 

/*  File   : getopt.c
 *  Author : Henry Spencer, University of Toronto
 *  Updated: 28 April 1984
 *
 *  Changes: (R Rudell)
 *  changed index() to strchr();
 *  added getopt_reset() to reset the getopt argument parsing
 *
 *  Purpose: get option letter from argv.
 */

var globalUtilOptarg string;        // Global argument pointer (util_optarg)
var globalUtilOptind int32;    // Global argv index (util_optind)

var pScanStr string;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [getSoftDataLimit()]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_GetSoftDataLimit() int32 {
    return EXTRA_RLIMIT_DATA_DEFAULT;
}

/**Function*************************************************************

  Synopsis    [util_getopt_reset()]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_UtilGetoptReset() {
    globalUtilOptarg = 0;
    globalUtilOptind = 0;
    pScanStr = 0;
}

/**Function*************************************************************

  Synopsis    [util_getopt()]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_UtilGetopt(argc int32, argv []string, optstring string) int32 {
    register int c;
    register char *place;

    globalUtilOptarg = nil;

    if pScanStr == nil || *pScanStr == '\0' {
        if globalUtilOptind == 0 { 
            globalUtilOptind++; }
        if globalUtilOptind >= argc { 
            return -1; }
        place = argv[globalUtilOptind];
        if place[0] != '-' || place[1] == '\0' { 
            return -1; }
        globalUtilOptind++;
        if place[1] == '-' && place[2] == '\0' { 
            return -1; }
        pScanStr = place+1;
    }

    c = *pScanStr++;
    place = strchr(optstring, c);
    if place == nil || c == ':' {
        (void) fmt.Fprintf(os.Stderr, "%s: unknown option %c\n", argv[0], c);
        return '?';
    }
    if *++place == ':' {
        if *pScanStr != '\0' {
            globalUtilOptarg = pScanStr;
            pScanStr = nil;
         
        } else {
            if globalUtilOptind >= argc {
                (void) fmt.Fprintf(os.Stderr, "%s: %c requires an argument\n", argv[0], c);
                return '?';
            }
            globalUtilOptarg = argv[globalUtilOptind];
            globalUtilOptind++;
        }
    }
    return c;
}

/**Function*************************************************************

  Synopsis    [util_print_time()]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Extra_UtilPrintTime_s [40]byte;
func Extra_UtilPrintTime(t long) string {
    

    (void) sprintf(spStatic_Extra_UtilPrintTime_s, "%ld.%02ld sec", t/1000, (t%1000)/10);
    return spStatic_Extra_UtilPrintTime_s;
}


/**Function*************************************************************

  Synopsis    [Extra_UtilStrsav()]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_UtilStrsav(s string) string {
    if s == nil {  /* added 7/95, for robustness */
       return nil;
    
    } else {
       // inlining of macro ABC_ALLOC
       return strcpy(((char *) malloc(sizeof(char) * (size_t)(len(s)+1))), s);
    }
}

/**Function*************************************************************

  Synopsis    [util_tilde_expand()]

  Description [The code contributed by Niklas Sorensson.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_UtilTildeExpand(fname string) string {
    return Extra_UtilStrsav( fname );
/*
    int         n_tildes = 0;
    const char* home;
    char*       expanded;
    int         length;
    int         i, j, k;

    for (i = 0; i < (int)strlen(fname); i++)
        if (fname[i] == '~') n_tildes++;

    home     = getenv("HOME");
    length   = n_tildes * strlen(home) + strlen(fname);
    expanded = ABC_ALLOC(char, length + 1);

    j = 0;
    for (i = 0; i < (int)strlen(fname); i++){
        if (fname[i] == '~'){
            for (k = 0; k < (int)strlen(home); k++)
                expanded[j++] = home[k];
        }else
            expanded[j++] = fname[i];
    }

    expanded[j] = '\0';
    return expanded; 
*/
}

/**Function*************************************************************

  Synopsis    [check_file()]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_UtilCheckFile(filename string, mode string) int32 {
    var fp *os.File;
    var got_file int32;

    if int32(strings.Compare(mode, "x")) == 0 {
    mode = "r";
    }
    fp = fopen(filename, mode);
    got_file = (fp != 0);
    if fp != 0 {
    (void) fclose(fp);
    }
    return got_file;
}

/**Function*************************************************************

  Synopsis    [util_file_search()]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
char * Extra_UtilFileSearch(char *file, char *path, char *mode)
//char *file;            // file we're looking for 
//char *path;            // search path, colon separated 
//char *mode;            // "r", "w", or "x" 
{
    var quit int32;
    var buffer, filename, save_path, cp string;

    if path == 0 || int32(strings.Compare(path, "")) == 0 {
    path = ".";        /* just look in the current directory */
    }

    save_path = path = Extra_UtilStrsav(path);
    quit = 0;
    for {
    cp = strchr(path, ':');
    if cp != 0 {
        *cp = '\0';
    } else {
        quit = 1;
    }

    /* cons up the filename out of the path and file name */
    if int32(strings.Compare(path, ".")) == 0 {
        buffer = Extra_UtilStrsav(file);
    } else {
        // inlining of macro ABC_ALLOC
        buffer = ((char *) malloc(sizeof(char) * (size_t)(len(path) + len(file) + 4)));
        (void) sprintf(buffer, "%s/%s", path, file);
    }
    filename = Extra_UtilTildeExpand(buffer);
    // inlining of macro ABC_FREE
    ((buffer) ? (free((char *) (buffer)), (buffer) = 0) : 0);

    /* see if we can access it */
    if Extra_UtilCheckFile(filename, mode) != 0 {
        // inlining of macro ABC_FREE
        ((save_path) ? (free((char *) (save_path)), (save_path) = 0) : 0);
        return filename;
    }
    // inlining of macro ABC_FREE
    ((filename) ? (free((char *) (filename)), (filename) = 0) : 0);
    path = ++cp;
    
if !(! (quit != 0)) { break }
}  

    // inlining of macro ABC_FREE
    ((save_path) ? (free((char *) (save_path)), (save_path) = 0) : 0);
    return 0;
}

/**Function*************************************************************

  Synopsis    [MMout_of_memory()]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/* MMout_of_memory -- out of memory for lazy people, flush and exit */
func Extra_UtilMMout_Of_Memory(size long) {
    (void) fflush(os.Stdout);
    (void) fmt.Fprintf(os.Stderr, "\nout of memory allocating %d bytes\n", uint32( size));
    assert(false, "0");
    os.Exit(int(1));
}

/**Function*************************************************************

  Synopsis    [MMoutOfMemory()]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
void (*Extra_UtilMMoutOfMemory)( long size ) = (void (*)( long size ))Extra_UtilMMout_Of_Memory;


/**Function*************************************************************

  Synopsis    [util_cpu_time()]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_CpuTime() abctime {
    return Abc_Clock();
}

/**Function*************************************************************

  Synopsis    [util_cpu_time()]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
#if defined(NT) || defined(NT64) || defined(WIN32) || defined(__EMSCRIPTEN__)
double Extra_CpuTimeDouble()
{
    return 1.0*Abc_Clock()/CLOCKS_PER_SEC;
}
#else

ABC_NAMESPACE_IMPL_END

#include <sys/time.h>
#include <sys/resource.h>
#include <unistd.h>

ABC_NAMESPACE_IMPL_START

double Extra_CpuTimeDouble()
{
    struct rusage ru;
    getrusage(RUSAGE_SELF, &ru);
    return (double)ru.ru_utime.tv_sec + (double)ru.ru_utime.tv_usec / 1000000; 
}
#endif

/**Function*************************************************************

  Synopsis    [Testing memory leaks.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_MemTest() {
//    ABC_ALLOC( char, 1002 );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
