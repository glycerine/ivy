package util

import spTime "time"

import "os"

/**CFile****************************************************************

  FileName    [abc_global.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Global declarations.]

  Synopsis    [Global declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Jan 30, 2009.]

  Revision    [$Id: abc_global.h,v 1.00 2009/01/30 00:00:00 alanmi Exp $]

***********************************************************************/




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

#ifndef ABC_USE_BRIDGE
#define ABC_USE_BRIDGE
#endif

#ifdef _WIN32
#ifndef __MINGW32__
#define inline __inline // compatible with MS VS 6.0
#pragma warning(disable : 4152) // warning C4152: nonstandard extension, function/data pointer conversion in expression
#pragma warning(disable : 4200) // warning C4200: nonstandard extension used : zero-sized array in struct/union
#pragma warning(disable : 4244) // warning C4244: '+=' : conversion from 'int ' to 'unsigned short ', possible loss of data
#pragma warning(disable : 4514) // warning C4514: 'Vec_StrPop' : unreferenced inline function has been removed
#pragma warning(disable : 4710) // warning C4710: function 'Vec_PtrGrow' not inlined
//#pragma warning( disable : 4273 )
#endif
#endif

#ifdef WIN32
  #ifdef WIN32_NO_DLL
    #define ABC_DLLEXPORT
    #define ABC_DLLIMPORT
  #else
    #define ABC_DLLEXPORT __declspec(dllexport)
    #define ABC_DLLIMPORT __declspec(dllimport)
  #endif
#else  /* defined(WIN32) */
#define ABC_DLLIMPORT
#endif /* defined(WIN32) */

#ifndef ABC_DLL
#define ABC_DLL ABC_DLLIMPORT
#endif

#if !defined(___unused)
#if defined(__GNUC__)
#define ___unused __attribute__ ((__unused__))
#else
#define ___unused
#endif
#endif

/*
#ifdef __cplusplus
#error "C++ code"
#else
#error "C code"
#endif
*/

//#include <time.h>
//#include <stdarg.h>
//#include <stdlib.h>
//#include <assert.h>
//#include <stdio.h>
//#include <string.h>

// catch memory leaks in Visual Studio
#ifdef WIN32
  #ifdef _DEBUG
    #define _CRTDBG_MAP_ALLOC
    #include <crtdbg.h>
  #endif
#endif

//#include "abc_namespaces.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

#ifdef ABC_USE_STDINT_H
// If there is stdint.h, assume this is a reasonably-modern platform that
// would also have stddef.h and limits.h
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#if UINTPTR_MAX == 0xFFFFFFFFFFFFFFFF
#define SIZEOF_VOID_P 8
#ifdef _WIN32
#define NT64
#else
#define LIN64
#endif
#elif UINTPTR_MAX == 0xFFFFFFFF
#define SIZEOF_VOID_P 4
#ifdef _WIN32
#define NT
#else
#define LIN
#endif
#else
   #error unsupported platform
#endif

#if ULONG_MAX == 0xFFFFFFFFFFFFFFFF
#define SIZEOF_LONG 8
#elif ULONG_MAX == 0xFFFFFFFF
#define SIZEOF_LONG 4
#else
   #error unsupported platform
#endif

#if UINT_MAX == 0xFFFFFFFFFFFFFFFF
#define SIZEOF_INT 8
#elif UINT_MAX == 0xFFFFFFFF
#define SIZEOF_INT 4
#else
   #error unsupported platform
#endif

#endif

/**
 * Pointer difference type; replacement for ptrdiff_t.
 * This is a signed integral type that is the same size as a pointer.
 * NOTE: This type may be different sizes on different platforms.
 */
#if       defined(__ccdoc__)
typedef platform_dependent_type ABC_PTRDIFF_T;
#elif     defined(ABC_USE_STDINT_H)
typedef ptrdiff_t ABC_PTRDIFF_T;
#elif     defined(LIN64)
typedef long ABC_PTRDIFF_T;
#elif     defined(NT64)
typedef long long ABC_PTRDIFF_T;
#elif     defined(NT) || defined(LIN) || defined(WIN32)
typedef int ABC_PTRDIFF_T;
#else
   #error unknown platform
#endif /* defined(PLATFORM) */

/**
 * Unsigned integral type that can contain a pointer.
 * This is an unsigned integral type that is the same size as a pointer.
 * NOTE: This type may be different sizes on different platforms.
 */
#if       defined(__ccdoc__)
typedef platform_dependent_type ABC_PTRUINT_T;
#elif     defined(ABC_USE_STDINT_H)
typedef uintptr_t ABC_PTRUINT_T;
#elif     defined(LIN64)
typedef unsigned long ABC_PTRUINT_T;
#elif     defined(NT64)
typedef unsigned long long ABC_PTRUINT_T;
#elif     defined(NT) || defined(LIN) || defined(WIN32)
typedef unsigned int ABC_PTRUINT_T;
#else
   #error unknown platform
#endif /* defined(PLATFORM) */

/**
 * Signed integral type that can contain a pointer.
 * This is a signed integral type that is the same size as a pointer.
 * NOTE: This type may be different sizes on different platforms.
 */
#if       defined(__ccdoc__)
typedef platform_dependent_type ABC_PTRINT_T;
#elif     defined(ABC_USE_STDINT_H)
typedef intptr_t ABC_PTRINT_T;
#elif     defined(LIN64)
typedef long ABC_PTRINT_T;
#elif     defined(NT64)
typedef long long ABC_PTRINT_T;
#elif     defined(NT) || defined(LIN) || defined(WIN32)
typedef int ABC_PTRINT_T;
#else
   #error unknown platform
#endif /* defined(PLATFORM) */

/**
 * 64-bit signed integral type.
 */
#if       defined(__ccdoc__)
typedef platform_dependent_type ABC_INT64_T;
#elif     defined(ABC_USE_STDINT_H)
typedef int64_t ABC_INT64_T;
#elif     defined(LIN64)
typedef long ABC_INT64_T;
#elif     defined(NT64) || defined(LIN)
typedef long long ABC_INT64_T;
#elif     defined(WIN32) || defined(NT)
typedef signed __int64 ABC_INT64_T;
#else
   #error unknown platform
#endif /* defined(PLATFORM) */

/**
 * 64-bit unsigned integral type.
 */
#if       defined(__ccdoc__)
typedef platform_dependent_type ABC_UINT64_T;
#elif     defined(ABC_USE_STDINT_H)
typedef uint64_t ABC_UINT64_T;
#elif     defined(LIN64)
typedef unsigned long ABC_UINT64_T;
#elif     defined(NT64) || defined(LIN)
typedef unsigned long long ABC_UINT64_T;
#elif     defined(WIN32) || defined(NT)
typedef unsigned __int64 ABC_UINT64_T;
#else
   #error unknown platform
#endif /* defined(PLATFORM) */

#ifdef LIN
  #define ABC_CONST(number) number ## ULL 
#else // LIN64 and windows
  #define ABC_CONST(number) number
#endif

typedef ABC_UINT64_T word;
typedef ABC_INT64_T iword;

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

#define ABC_INFINITY    (1000000000)

// Original macro:
// #define ABC_SWAP(Type, a, b)  { Type t = a; a = b; b = t; }


// Original macro:
// #define ABC_PRT(a,t)    (Abc_Print(1, "%s =", (a)), Abc_Print(1, "%9.2f sec\n", 1.0*((double)(t))/((double)CLOCKS_PER_SEC)))

// Original macro:
// #define ABC_PRTr(a,t)   (Abc_Print(1, "%s =", (a)), Abc_Print(1, "%9.2f sec\r", 1.0*((double)(t))/((double)CLOCKS_PER_SEC)))

// Original macro:
// #define ABC_PRTn(a,t)   (Abc_Print(1, "%s =", (a)), Abc_Print(1, "%9.2f sec  ", 1.0*((double)(t))/((double)CLOCKS_PER_SEC)))

// Original macro:
// #define ABC_PRTP(a,t,T) (Abc_Print(1, "%s =", (a)), Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*((double)(t))/((double)CLOCKS_PER_SEC), ((double)(T))? 100.0*((double)(t))/((double)(T)) : 0.0))

// Original macro:
// #define ABC_PRM(a,f)    (Abc_Print(1, "%s =", (a)), Abc_Print(1, "%10.3f MB\n",    1.0*((double)(f))/(1<<20)))

// Original macro:
// #define ABC_PRMr(a,f)   (Abc_Print(1, "%s =", (a)), Abc_Print(1, "%10.3f MB\r",    1.0*((double)(f))/(1<<20)))

// Original macro:
// #define ABC_PRMn(a,f)   (Abc_Print(1, "%s =", (a)), Abc_Print(1, "%10.3f MB  ",    1.0*((double)(f))/(1<<20)))

// Original macro:
// #define ABC_PRMP(a,f,F) (Abc_Print(1, "%s =", (a)), Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*((double)(f))/(1<<20)), (((double)(F))? 100.0*((double)(f))/((double)(F)) : 0.0) ) )


// Original macro:
// #define ABC_ALLOC(type, num)     ((type *) malloc(sizeof(type) * (size_t)(num)))

// Original macro:
// #define ABC_CALLOC(type, num)    ((type *) calloc((size_t)(num), sizeof(type)))

// Original macro:
// #define ABC_FALLOC(type, num)    ((type *) memset(malloc(sizeof(type) * (size_t)(num)), 0xff, sizeof(type) * (size_t)(num)))

// Original macro:
// #define ABC_FREE(obj)            ((obj) ? (free((char *) (obj)), (obj) = 0) : 0)

// Original macro:
// #define ABC_REALLOC(type, obj, num) \
//         ((obj) ? ((type *) realloc((char *)(obj), sizeof(type) * (size_t)(num))) : \
//          ((type *) malloc(sizeof(type) * (size_t)(num))))


func Abc_AbsInt(a int32) int32 { if a < 0 {
    return -a;
} else {
    return a;
} }
func Abc_MaxInt(a int32, b int32) int32 { if a > b {
    return a;
} else {
    return b;
} }
func Abc_MinInt(a int32, b int32) int32 { if a < b {
    return a;
} else {
    return b;
} }
func Abc_MaxWord(a word, b word) word { if a > b {
    return a;
} else {
    return b;
} }
func Abc_MinWord(a word, b word) word { if a < b {
    return a;
} else {
    return b;
} }
func Abc_AbsFloat(a float32) float32 { if a < 0 {
    return -a;
} else {
    return a;
} }
func Abc_MaxFloat(a float32, b float32) float32 { if a > b {
    return a;
} else {
    return b;
} }
func Abc_MinFloat(a float32, b float32) float32 { if a < b {
    return a;
} else {
    return b;
} }
func Abc_AbsDouble(a float64) float64 { if a < 0 {
    return -a;
} else {
    return a;
} }
func Abc_MaxDouble(a float64, b float64) float64 { if a > b {
    return a;
} else {
    return b;
} }
func Abc_MinDouble(a float64, b float64) float64 { if a < b {
    return a;
} else {
    return b;
} }

func Abc_Float2Int(Val float32) int32 { union { var x int32; var y float32; }
 v; v.y = Val; return v.x;         }
func Abc_Int2Float(Num int32) float32 { union { var x int32; var y float32; }
 v; v.x = Num; return v.y;         }
func Abc_Dbl2Word(Dbl float64) word { union { var x word; var y float64; }
 v; v.y = Dbl; return v.x;       }
func Abc_Word2Dbl(Num word) float64 { union { var x word; var y float64; }
 v; v.x = Num; return v.y;       }
func Abc_Base2Log(n uint32) int32 { var r int32; if n < 2 { return int32(n); }
 for  r, n = 0, n-1; (n != 0); n >>= 1, r++  {}; return r; }
func Abc_Base10Log(n uint32) int32 { var r int32; if n < 2 { return int32(n); }
 for  r, n = 0, n-1; (n != 0); n /= 10, r++  {}; return r; }
func Abc_Base16Log(n uint32) int32 { var r int32; if n < 2 { return int32(n); }
 for  r, n = 0, n-1; (n != 0); n /= 16, r++  {}; return r; }
func Abc_UtilStrsav(s string) string { if s {
    // inlining of macro ABC_ALLOC
    return strcpy(((char *) malloc(sizeof(char) * (size_t)(len(s)+1))), s);
} else {
    return nil;
}  }
func Abc_BitByteNum(nBits int32) int32 { return (nBits>>3) + ((nBits&7)  > 0);                       }
func Abc_BitWordNum(nBits int32) int32 { return (nBits>>5) + ((nBits&31) > 0);                       }
func Abc_Bit6WordNum(nBits int32) int32 { return (nBits>>6) + ((nBits&63) > 0);                       }
func Abc_TruthByteNum(nVars int32) int32 { if nVars <= 3 {
    return 1;
} else {
    return (1 << (nVars - 3));
}                 }
func Abc_TruthWordNum(nVars int32) int32 { if nVars <= 5 {
    return 1;
} else {
    return (1 << (nVars - 5));
}                 }
func Abc_Truth6WordNum(nVars int32) int32 { if nVars <= 6 {
    return 1;
} else {
    return (1 << (nVars - 6));
}                 }
func Abc_InfoHasBit(p *uint32, i int32) int32 { return (p[(i)>>5] & uint32((1<<((i) & 31)))) > 0;         }
func Abc_InfoSetBit(p *uint32, i int32) { p[(i)>>5] |= uint32((1<<((i) & 31)));                     }
func Abc_InfoXorBit(p *uint32, i int32) { p[(i)>>5] ^= uint32((1<<((i) & 31)));                     }
func Abc_InfoMask(nVar int32) uint32 { return (~uint32(0)) >> (32-nVar);                         }

func Abc_Var2Lit(Var int32, c int32) int32 { assert(Var >= 0 && !(c >> 1), "Var >= 0 && !(c >> 1)"); return Var + Var + c;        }
func Abc_Lit2Var(Lit int32) int32 { assert(Lit >= 0, "Lit >= 0"); return Lit >> 1;                          }
func Abc_LitIsCompl(Lit int32) int32 { assert(Lit >= 0, "Lit >= 0"); return Lit & 1;                           }
func Abc_LitNot(Lit int32) int32 { assert(Lit >= 0, "Lit >= 0"); return Lit ^ 1;                           }
func Abc_LitNotCond(Lit int32, c int32) int32 { assert(Lit >= 0, "Lit >= 0"); return Lit ^ int32((c > 0));                }
func Abc_LitRegular(Lit int32) int32 { assert(Lit >= 0, "Lit >= 0"); return Lit & ~01;                         }
func Abc_Lit2LitV(pMap *int32, Lit int32) int32 { assert(Lit >= 0, "Lit >= 0"); return Abc_Var2Lit( pMap[Abc_Lit2Var(Lit)], Abc_LitIsCompl(Lit) );      }
func Abc_Lit2LitL(pMap *int32, Lit int32) int32 { assert(Lit >= 0, "Lit >= 0"); return Abc_LitNotCond( pMap[Abc_Lit2Var(Lit)], Abc_LitIsCompl(Lit) );   }

func Abc_Ptr2Int(p any) int32 { return int32((ABC_PTRINT_T))p;      }
func Abc_Int2Ptr(i int32) any { return (void *)(ABC_PTRINT_T)i;   }
func Abc_Ptr2Wrd(p any) word { return (word)(ABC_PTRUINT_T)p;    }
func Abc_Wrd2Ptr(i word) any { return (void *)(ABC_PTRUINT_T)i;  }

func Abc_Var2Lit2(Var int32, Att int32) int32 { assert(!(Att >> 2), "!(Att >> 2)"); return (Var << 2) + Att; }
func Abc_Lit2Var2(Lit int32) int32 { assert(Lit >= 0, "Lit >= 0");    return Lit >> 2;         }
func Abc_Lit2Att2(Lit int32) int32 { assert(Lit >= 0, "Lit >= 0");    return Lit & 3;          }
func Abc_Var2Lit3(Var int32, Att int32) int32 { assert(!(Att >> 3), "!(Att >> 3)"); return (Var << 3) + Att; }
func Abc_Lit2Var3(Lit int32) int32 { assert(Lit >= 0, "Lit >= 0");    return Lit >> 3;         }
func Abc_Lit2Att3(Lit int32) int32 { assert(Lit >= 0, "Lit >= 0");    return Lit & 7;          }
func Abc_Var2Lit4(Var int32, Att int32) int32 { assert(!(Att >> 4), "!(Att >> 4)"); return (Var << 4) + Att; }
func Abc_Lit2Var4(Lit int32) int32 { assert(Lit >= 0, "Lit >= 0");    return Lit >> 4;         }
func Abc_Lit2Att4(Lit int32) int32 { assert(Lit >= 0, "Lit >= 0");    return Lit & 15;         }

// time counting
typedef ABC_INT64_T abctime;
func Abc_Clock() abctime {
#if (defined(LIN) || defined(LIN64)) && !(__APPLE__ & __MACH__) && !defined(__MINGW32__)
    struct timespec ts;
    if ( clock_gettime(CLOCK_THREAD_CPUTIME_ID, &ts) < 0 ) 
        return (abctime)-1;
    abctime res = ((abctime) ts.tv_sec) * CLOCKS_PER_SEC;
    res += (((abctime) ts.tv_nsec) * CLOCKS_PER_SEC) / 1000000000;
    return res;
#else
    return (abctime) clock();
#endif
}


// misc printing procedures
type Abc_VerbLevel int32
const 
(
    ABC_PROMPT Abc_VerbLevel   = -2

    ABC_ERROR Abc_VerbLevel    = -1

    ABC_WARNING Abc_VerbLevel  =  0

    ABC_STANDARD Abc_VerbLevel =  1

    ABC_VERBOSE Abc_VerbLevel  =  2

)



#ifdef ABC_USE_BRIDGE

// bridge communication
#define BRIDGE_NETLIST           106
#define BRIDGE_ABS_NETLIST       107
extern int Gia_ManToBridgeText( FILE * pFile, int Size, unsigned char * pBuffer );
extern int Gia_ManToBridgeAbsNetlist( FILE * pFile, void * p, int pkg_type );

// string printing
extern char * vnsprintf(const char* format, va_list args);
extern char * nsprintf(const char* format, ...);

static inline void Abc_Print( int level, const char * format, ... )
{
    extern ABC_DLL int Abc_FrameIsBridgeMode();
    va_list args;
    extern unsigned enable_dbg_outs;  
    if ( !enable_dbg_outs )
        return;

    if ( ! Abc_FrameIsBridgeMode() ){
        if ( level == ABC_ERROR )
            printf( "Error: " );
        else if ( level == ABC_WARNING )
            printf( "Warning: " );
    }else{
        if ( level == ABC_ERROR )
            Gia_ManToBridgeText( stdout, (int)strlen("Error: "), (unsigned char*)"Error: " );
        else if ( level == ABC_WARNING )
            Gia_ManToBridgeText( stdout, (int)strlen("Warning: "), (unsigned char*)"Warning: " );
    }

    va_start( args, format );
    if ( Abc_FrameIsBridgeMode() )
    {
        char * tmp = vnsprintf( format, args );
        Gia_ManToBridgeText( stdout, (int)strlen(tmp), (unsigned char*)tmp );
        free( tmp );
    }
    else
        vprintf( format, args );
    va_end( args );
}

#else

static inline void Abc_Print( int level, const char * format, ... )
{
    va_list args;
    va_start( args, format );
    if ( level == ABC_ERROR )
        printf( "Error: " );
    else if ( level == ABC_WARNING )
        printf( "Warning: " );
    vprintf( format, args );
    va_end( args );
}

#endif


func Abc_PrintInt(i int32) {
    var v3 float64  = float64(i)/1000;
    var v6 float64  = float64(i)/1000000;

    Abc_Print( 1,  "  " );

    if i > -1000 && i < 1000 {
        Abc_Print( 1, " %4d", i );

    } else if v3 > -9.995 && v3 < 9.995 {
        Abc_Print( 1, "%4.2fk", v3 );
    } else if v3 > -99.95 && v3 < 99.95 {
        Abc_Print( 1, "%4.1fk", v3 );
    } else if v3 > -999.5 && v3 < 999.5 {
        Abc_Print( 1, "%4.0fk", v3 );

    } else if v6 > -9.995 && v6 < 9.995 {
        Abc_Print( 1, "%4.2fm", v6 );
    } else if v6 > -99.95 && v6 < 99.95 {
        Abc_Print( 1, "%4.1fm", v6 );
    } else if v6 > -999.5 && v6 < 999.5 {
        Abc_Print( 1, "%4.0fm", v6 ); }
}


func Abc_PrintTime(level int32, pStr string, time abctime) {
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", (pStr));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time)))/(float64(int64(spTime.Second))));;
}

func Abc_PrintTimeP(level int32, pStr string, time abctime, Time abctime) {
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (pStr));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((time)))/(float64(int64(spTime.Second))), (float64((Time)))? 100.0*(float64((time)))/(float64((Time))) : 0.0);;
}

func Abc_PrintMemoryP(level int32, pStr string, mem int32, Mem int32) {
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", (pStr));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((mem)))/(1<<20)), ((float64((Mem)))? 100.0*(float64((mem)))/(float64((Mem))) : 0.0) );;
}

// Returns the next prime >= p
static inline int Abc_PrimeCudd( unsigned int p )
{
    var i, pn int32;
    p--;
    for {
        p++;
        if p&1 {
            pn = 1;
            i = 3;
            for uint32( (i * i)) <= p {
                if p % uint32(i) == 0 {
                    pn = 0;
                    break;
                }
                i += 2;
            }
        
        } else {
            pn = 0; }
    
if !(!(pn != 0)) { break }
} 
    return int32((p));

} // end of Cudd_Prime 

// the returned buffer has 32 unused bytes at the end, filled with zeros
func Abc_FileReadContents(pFileName string, pnFileSize *int32) any {
    var RetValue, nFileSize int32;
    var pBuffer string;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        return nil; }
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );
    nFileSize = ftell( pFile );
    if pnFileSize != nil {
        *pnFileSize = nFileSize; }
    rewind( pFile );
    // load the contents of the file into memory
    // inlining of macro ABC_ALLOC
    pBuffer = ((char *) malloc(sizeof(char) * (size_t)(nFileSize + 32)));
    RetValue = fread( pBuffer, 1, nFileSize, pFile );
    memset( pBuffer + nFileSize, 0, 32 );
    assert(RetValue == nFileSize, "RetValue == nFileSize");
    _ = (pFile).Close();
    return (void *)pBuffer;
}
func Abc_ReverseOrder(pA *int32, nA int32) {
    var i int32;
    for  i = 0; i < nA/2; i++  {
        // inlining of macro ABC_SWAP
         var t int32  = pA[i]; pA[i] = pA[nA-1-i]; pA[nA-1-i] = t; };
}


// sorting
// extern void   Abc_MergeSort( int * pInput, int nSize );

// extern int *  Abc_MergeSortCost( int * pCosts, int nSize );

// extern void   Abc_QuickSort1( word * pData, int nSize, int fDecrease );

// extern void   Abc_QuickSort2( word * pData, int nSize, int fDecrease );

// extern void   Abc_QuickSort3( word * pData, int nSize, int fDecrease );

// extern void   Abc_QuickSortCostData( int * pCosts, int nSize, int fDecrease, word * pData, int * pResult );

// extern int *  Abc_QuickSortCost( int * pCosts, int nSize, int fDecrease );


// extern unsigned Abc_Random( int fReset );

// extern word     Abc_RandomW( int fReset );


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
