package util

/**CFile****************************************************************

  FileName    [utilSignal.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName []

  Synopsis    []

  Author      [Baruch Sterin]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2011.]

  Revision    [$Id: utilSignal.c,v 1.00 2011/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>

//#include "abc_global.h"
//#include "utilSignal.h"

#ifdef _MSC_VER
#define unlink _unlink
#else
#include <unistd.h>
#endif

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

func Util_SignalSystem(cmd string) int32 {
    return system(cmd);
}

// int tmpFile(const char* prefix, const char* suffix, char** out_name);


func Util_SignalTmpFile(prefix string, suffix string, out_name *string) int32 {
    return tmpFile(prefix, suffix, out_name);
}

func Util_SignalTmpFileRemove(fname string, fLeave int32) {
    if ! (fLeave != 0) {
        unlink(fname);
    }
}

//ABC_NAMESPACE_IMPL_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
