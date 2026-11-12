package main

import "fmt"

/**CFile****************************************************************

  FileName    [libSupport.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [The main package.]

  Synopsis    [Support for external libaries.]

  Author      [Mike Case]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: libSupport.c,v 1.1 2005/09/06 19:59:51 casem Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <string.h>

//#include "base/abc/abc.h"
//#include "mainInt.h"

//ABC_NAMESPACE_IMPL_START

#if defined(ABC_NO_DYNAMIC_LINKING)
#define WIN32
#endif

#ifndef WIN32
# include <sys/types.h>
# include <dirent.h>
# include <dlfcn.h>
#endif

// fix by Paddy O'Brien  on Sep 22, 2009
#ifdef __CYGWIN__
#ifndef RTLD_LOCAL
#define RTLD_LOCAL 0
#endif
#endif 


const MAX_LIBS = 256
var libHandles [MAX_LIBS+1]any; // will be null terminated

typedef void (*lib_init_end_func) (Abc_Frame_t * pAbc);

////////////////////////////////////////////////////////////////////////////////////////////////////
// This will find all the ABC library extensions in the current directory and load them all.
////////////////////////////////////////////////////////////////////////////////////////////////////
func open_libs() {
    var curr_lib int32;

#ifdef WIN32
//    printf("Warning: open_libs WIN32 not implemented.\n");
#else
    DIR* dirp;
    struct dirent* dp;
    char *env, *init_p, *p;
    int done;

    env = getenv ("ABC_LIB_PATH");
    if (env == NULL) {
//    printf("Warning: ABC_LIB_PATH not defined. Looking into the current directory.\n");
      init_p = ABC_ALLOC( char, (2*sizeof(char)) );
      init_p[0]='.'; init_p[1] = 0;
    } else {
      init_p = ABC_ALLOC( char, ((strlen(env)+1)*sizeof(char)) );
      strcpy (init_p, env);
    }

    // Extract directories and read libraries
    done = 0;
    p = init_p;
    while (!done) {
      char *endp = strchr (p,':');
      if (endp == NULL) done = 1; // last directory in the list
      else *endp = 0; // end of string

      dirp = opendir(p);
      if (dirp == NULL) {
//      printf("Warning: directory in ABC_LIB_PATH does not exist (%s).\n", p);
        continue;
      }

      while ((dp = readdir(dirp)) != NULL) {
        if ((strncmp("libabc_", dp->d_name, 7) == 0) &&
            (strcmp(".so", dp->d_name + strlen(dp->d_name) - 3) == 0)) {

          // make sure we don't overflow the handle array
          if (curr_lib >= MAX_LIBS) {
            printf("Warning: maximum number of ABC libraries (%d) exceeded.  Not loading %s.\n",
                   MAX_LIBS,
                   dp->d_name);
          }
          
          // attempt to load it
          else {
            char* szPrefixed = ABC_ALLOC( char, ((strlen(dp->d_name) + strlen(p) + 2) * 
                                      sizeof(char)) );
            sprintf(szPrefixed, "%s/", p);
            strcat(szPrefixed, dp->d_name);
            libHandles[curr_lib] = dlopen(szPrefixed, RTLD_NOW | RTLD_LOCAL);
            
            // did the load succeed?
            if (libHandles[curr_lib] != 0) {
              printf("Loaded ABC library: %s (Abc library extension #%d)\n", szPrefixed, curr_lib);
              curr_lib++;
            } else {
              printf("Warning: failed to load ABC library %s:\n\t%s\n", szPrefixed, dlerror());
            }
            
            ABC_FREE(szPrefixed);
          }
        }
      }
      closedir(dirp);
      p = endp+1;
    }

    ABC_FREE(init_p);
#endif
    
    // null terminate the list of handles
    libHandles[curr_lib] = 0;    
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// This will close all open ABC library extensions
////////////////////////////////////////////////////////////////////////////////////////////////////
func close_libs() {
#ifdef WIN32
    printf("Warning: close_libs WIN32 not implemented.\n");
#else
    int i;
    for (i = 0; libHandles[i] != 0; i++) {
        if (dlclose(libHandles[i]) != 0) {
            printf("Warning: failed to close library %d\n", i);
        }
        libHandles[i] = 0;
    }
#endif
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// This will get a pointer to a function inside of an open library
////////////////////////////////////////////////////////////////////////////////////////////////////
func get_fnct_ptr(lib_num int32, sym_name string) any {
#ifdef WIN32
    printf("Warning: get_fnct_ptr WIN32 not implemented.\n");
    return 0;
#else
    return dlsym(libHandles[lib_num], sym_name);
#endif
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// This will call an initialization function in every open library.
////////////////////////////////////////////////////////////////////////////////////////////////////
func call_inits(pAbc *Abc_Frame_t) {
    var i int32;
    lib_init_end_func init_func;
    for i = 0; libHandles[i] != 0; i++ {
        init_func = (lib_init_end_func) get_fnct_ptr(i, "abc_init");
        if init_func == 0 {
            fmt.Printf("Warning: Failed to initialize library %d.\n", i);
        } else {
            (*init_func)(pAbc);
        }
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// This will call a shutdown function in every open library.
////////////////////////////////////////////////////////////////////////////////////////////////////
func call_ends(pAbc *Abc_Frame_t) {
    var i int32;
    lib_init_end_func end_func;
    for i = 0; libHandles[i] != 0; i++ {
        end_func = (lib_init_end_func) get_fnct_ptr(i, "abc_end");
        if end_func == 0 {
            fmt.Printf("Warning: Failed to end library %d.\n", i);
        } else {
            (*end_func)(pAbc);
        }
    }
}

func Libs_Init(pAbc *Abc_Frame_t) {
    open_libs();
    call_inits(pAbc);
}

func Libs_End(pAbc *Abc_Frame_t) {
    call_ends(pAbc);

    // It's good practice to close our libraries at this point, but this can mess up any backtrace printed by Valgind.
    //    close_libs();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
//ABC_NAMESPACE_IMPL_END

