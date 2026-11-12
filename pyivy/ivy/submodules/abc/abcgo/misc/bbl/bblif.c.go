package bbl

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [bblif.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Binary BLIF representation for logic networks.]

  Synopsis    [Main implementation module.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 28, 2009.]

  Revision    [$Id: bblif.c,v 1.00 2009/02/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/util/abc_global.h"
//#include "bblif.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// vector of integers
   
type Vec_Int_t struct {
    nCap         int32;
    nSize         int32;
    pArray        *int32;
}

// vector of characters
   
type Vec_Str_t struct {
    nCap         int32;
    nSize         int32;
    pArray       string;
}

// network object
type Bbl_Obj_t_ struct {
    Id         int32;              // user ID
    Fnc         int32;             // functionality
    fCi uint32; /* C bitfield: 1 bits; review packing and truncation */    // combinational input
    fCo uint32; /* C bitfield: 1 bits; review packing and truncation */    // combinational output
    fBox uint32; /* C bitfield: 1 bits; review packing and truncation */    // subcircuit
    fMark uint32; /* C bitfield: 1 bits; review packing and truncation */    // temporary mark
    nFanins uint32; /* C bitfield: 28 bits; review packing and truncation */    // fanin number
    pFanins [0]int32;      // fanin array
}

// object function
   
type Bbl_Fnc_t struct {
    nWords         int32;          // word number
    pWords [0]int32;       // word array
}

// object function
   
type Bbl_Ent_t struct {
    iFunc         int32;           // function handle
    iNext         int32;           // next entry handle
}

// data manager
type Bbl_Man_t_ struct {
    // data pool
    pName  *Vec_Str_t;           // design name
    pObjs  *Vec_Str_t;           // vector of objects
    pFncs  *Vec_Str_t;           // vector of functions
    // construction
    vId2Obj  *Vec_Int_t;         // mapping user IDs into objects
    vObj2Id  *Vec_Int_t;         // mapping objects into user IDs
    vFaninNums  *Vec_Int_t;      // mapping user IDs into fanin number
    // file contents
    nFileSize         int32;       // file size
    pFileData       string;       // file contents
    // other data
    pEnts  *Vec_Str_t;           // vector of entries
    SopMap [17][17]int32;  // mapping vars x cubes into entry handles
}

func Bbl_ObjIsCi(pObj *Bbl_Obj_t) int32 { return pObj.fCi;                                        } 
func Bbl_ObjIsCo(pObj *Bbl_Obj_t) int32 { return pObj.fCo;                                        } 
func Bbl_ObjIsNode(pObj *Bbl_Obj_t) int32 { return!pObj.fCi && !pObj.fCo;                          } 

func Bbl_ObjFaninNum(pObj *Bbl_Obj_t) int32 { return pObj.nFanins;                                    } 
func Bbl_ObjFanin(pObj *Bbl_Obj_t, i int32) *Bbl_Obj_t { return (Bbl_Obj_t *)(((char *)pObj) - pObj.pFanins[i]); } 

func Bbl_ObjSize(pObj *Bbl_Obj_t) int32 { return sizeof(Bbl_Obj_t) + sizeof(int) * pObj.nFanins;  } 
func Bbl_FncSize(pFnc *Bbl_Fnc_t) int32 { return sizeof(Bbl_Fnc_t) + sizeof(int) * pFnc.nWords;   } 

func Bbl_VecObj(p *Vec_Str_t, h int32) *Bbl_Obj_t { return (Bbl_Obj_t *)(p.pArray + h);                     } 
func Bbl_VecFnc(p *Vec_Str_t, h int32) *Bbl_Fnc_t { return (Bbl_Fnc_t *)(p.pArray + h);                     } 
func Bbl_VecEnt(p *Vec_Str_t, h int32) *Bbl_Ent_t { return (Bbl_Ent_t *)(p.pArray + h);                     } 

func Bbl_ManSop(p *Bbl_Man_t, h int32) string { return (char *)Bbl_VecFnc(p.pFncs, h).pWords;          } 
func Bbl_ManObj(p *Bbl_Man_t, Id int32) *Bbl_Obj_t { return Bbl_VecObj(p.pObjs, p.vId2Obj.pArray[Id]);     } 

// Original macro:
// #define Bbl_ManForEachObj_int( p, pObj, h )                \
//     for ( h = 0; (h < p->nSize) && (pObj = Bbl_VecObj(p,h)); h += Bbl_ObjSize(pObj) )

// Original macro:
// #define Bbl_ManForEachFnc_int( p, pObj, h )                \
//     for ( h = 0; (h < p->nSize) && (pObj = Bbl_VecFnc(p,h)); h += Bbl_FncSize(pObj) )

// Original macro:
// #define Bbl_ObjForEachFanin_int( pObj, pFanin, i )         \
//     for ( i = 0; (i < (int)pObj->nFanins) && (pFanin = Bbl_ObjFanin(pObj,i)); i++ )


// Original macro:
// #define BBLIF_ALLOC(type, num)     ((type *) malloc(sizeof(type) * (num)))

// Original macro:
// #define BBLIF_CALLOC(type, num)     ((type *) calloc((num), sizeof(type)))

// Original macro:
// #define BBLIF_FALLOC(type, num)     ((type *) memset(malloc(sizeof(type) * (num)), 0xff, sizeof(type) * (num)))

// Original macro:
// #define BBLIF_FREE(obj)             ((obj) ? (free((char *) (obj)), (obj) = 0) : 0)

// Original macro:
// #define BBLIF_REALLOC(type, obj, num)    \
//         ((obj) ? ((type *) realloc((char *)(obj), sizeof(type) * (num))) : \
//          ((type *) malloc(sizeof(type) * (num))))


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntAlloc(nCap int32) *Vec_Int_t {
    var p *Vec_Int_t;
    // inlining of macro BBLIF_ALLOC
    p = ((Vec_Int_t *) malloc(sizeof(Vec_Int_t) * (1)));
    if nCap > 0 && nCap < 16 {
        nCap = 16; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        // inlining of macro BBLIF_ALLOC
        p.pArray = ((int *) malloc(sizeof(int) * (p.nCap)));
    } else {
        p.pArray = nil;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given size and cleans it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntStart(nSize int32) *Vec_Int_t {
    var p *Vec_Int_t;
    p = Vec_IntAlloc( nSize );
    p.nSize = nSize;
    memset( p.pArray, 0, sizeof(int) * nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given size and cleans it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntStartNatural(nSize int32) *Vec_Int_t {
    var p *Vec_Int_t;
    var i int32;
    p = Vec_IntAlloc( nSize );
    p.nSize = nSize;
    for  i = 0; i < nSize; i++  {
        p.pArray[i] = i;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates the vector from an integer array of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntAllocArray(pArray *int32, nSize int32) *Vec_Int_t {
    var p *Vec_Int_t;
    // inlining of macro BBLIF_ALLOC
    p = ((Vec_Int_t *) malloc(sizeof(Vec_Int_t) * (1)));
    p.nSize  = nSize;
    p.nCap   = nSize;
    p.pArray = pArray;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntFree(p *Vec_Int_t) {
    // inlining of macro BBLIF_FREE
    ((p.pArray) ? (free((char *) (p.pArray)), (p.pArray) = 0) : 0);
    // inlining of macro BBLIF_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSize(p *Vec_Int_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntEntry(p *Vec_Int_t, i int32) int32 {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntWriteEntry(p *Vec_Int_t, i int32, Entry int32) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntAddToEntry(p *Vec_Int_t, i int32, Addition int32) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] += Addition;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntEntryLast(p *Vec_Int_t) int32 {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.pArray[p.nSize-1];
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntGrow(p *Vec_Int_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    // inlining of macro BBLIF_REALLOC
    
    if p.pArray {
        p.pArray = ((int *) realloc((char *)(p.pArray), sizeof(int) * (nCapMin)));
    } else {
        p.pArray = ((int *) malloc(sizeof(int) * (nCapMin)));
    } 
    assert(p.pArray, "p.pArray");
    p.nCap   = nCapMin;
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntFill(p *Vec_Int_t, nSize int32, Fill int32) {
    var i int32;
    Vec_IntGrow( p, nSize );
    for  i = 0; i < nSize; i++  {
        p.pArray[i] = Fill;
    }
    p.nSize = nSize;
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntFillExtra(p *Vec_Int_t, nSize int32, Fill int32) {
    var i int32;
    if p.nSize >= nSize {
        return; }
    if nSize > 2 * p.nCap {
        Vec_IntGrow( p, nSize );
    } else if nSize > p.nCap {
        Vec_IntGrow( p, 2 * p.nCap ); }
    for  i = p.nSize; i < nSize; i++  {
        p.pArray[i] = Fill;
    }
    p.nSize = nSize;
}

/**Function*************************************************************

  Synopsis    [Returns the entry even if the place not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntGetEntry(p *Vec_Int_t, i int32) int32 {
    Vec_IntFillExtra( p, i + 1, 0 );
    return Vec_IntEntry( p, i );
}

/**Function*************************************************************

  Synopsis    [Inserts the entry even if the place does not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSetEntry(p *Vec_Int_t, i int32, Entry int32) {
    Vec_IntFillExtra( p, i + 1, 0 );
    Vec_IntWriteEntry( p, i, Entry );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntShrink(p *Vec_Int_t, nSizeNew int32) {
    assert(p.nSize >= nSizeNew, "p.nSize >= nSizeNew");
    p.nSize = nSizeNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntClear(p *Vec_Int_t) {
    p.nSize = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntPush(p *Vec_Int_t, Entry int32) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_IntGrow( p, 16 );
        } else {
            Vec_IntGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Entry;
}





/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrAlloc(nCap int32) *Vec_Str_t {
    var p *Vec_Str_t;
    // inlining of macro BBLIF_ALLOC
    p = ((Vec_Str_t *) malloc(sizeof(Vec_Str_t) * (1)));
    if nCap > 0 && nCap < 16 {
        nCap = 16; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        // inlining of macro BBLIF_ALLOC
        p.pArray = ((char *) malloc(sizeof(char) * (p.nCap)));
    } else {
        p.pArray = nil;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates the vector from an array of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrAllocArray(pArray string, nSize int32) *Vec_Str_t {
    var p *Vec_Str_t;
    // inlining of macro BBLIF_ALLOC
    p = ((Vec_Str_t *) malloc(sizeof(Vec_Str_t) * (1)));
    p.nSize  = nSize;
    p.nCap   = nSize;
    p.pArray = pArray;
    return p;
}

/**Fnction*************************************************************

  Synopsis    [Returns a piece of memory.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrFetch(p *Vec_Str_t, nBytes int32) string {
    for  p.nSize + nBytes > p.nCap  {
        // inlining of macro BBLIF_REALLOC
        
        if p.pArray {
            p.pArray = ((char *) realloc((char *)(p.pArray), sizeof(char) * (3 * p.nCap)));
        } else {
            p.pArray = ((char *) malloc(sizeof(char) * (3 * p.nCap)));
        }
        p.nCap *= 3;
    }
    p.nSize += nBytes;
    return p.pArray + p.nSize - nBytes;
}

/**Fnction*************************************************************

  Synopsis    [Write vector into file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrWrite(pFile *os.File, p *Vec_Str_t) {
    fwrite( &p.nSize, sizeof(int), 1, pFile );
    fwrite( p.pArray, sizeof(char), p.nSize, pFile );
}

/**Fnction*************************************************************

  Synopsis    [Write vector into file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrRead(ppStr *string) *Vec_Str_t {
    var p *Vec_Str_t;
    var pStr string  = *ppStr;
    p = Vec_StrAlloc( 0 );
    p.nSize = *(int *)pStr;
    p.pArray = pStr + sizeof(int);
    *ppStr = pStr + sizeof(int) + p.nSize * sizeof(char);
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrSize(p *Vec_Str_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrFree(p *Vec_Str_t) {
    // inlining of macro BBLIF_FREE
    ((p.pArray) ? (free((char *) (p.pArray)), (p.pArray) = 0) : 0);
    // inlining of macro BBLIF_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}





/**Fnction*************************************************************

  Synopsis    [Returns the file size.]

  Description [The file should be closed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManFileSize(pFileName string) int32 {
    var pFile *os.File;
    var nFileSize int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Bbl_ManFileSize(): The file is unavailable (absent or open).\n");
        return 0;
    }
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile ); 
    _ = (pFile).Close();
    return nFileSize;
}

/**Fnction*************************************************************

  Synopsis    [Read data from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManFileRead(pFileName string) string {
    var pFile *os.File;
    var pContents string;
    var nFileSize int32;
    var RetValue int32;
    nFileSize = Bbl_ManFileSize( pFileName );
    pFile = os.Open(pFileName);
    // inlining of macro BBLIF_ALLOC
    pContents = ((char *) malloc(sizeof(char) * (nFileSize)));
    RetValue = fread( pContents, nFileSize, 1, pFile );
    _ = (pFile).Close();
    return pContents;
}



/**Fnction*************************************************************

  Synopsis    [Writes data into file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManDumpBinaryBlif(p *Bbl_Man_t, pFileName string) {
    var pFile *os.File;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    Vec_StrWrite( pFile, p.pName );
    Vec_StrWrite( pFile, p.pObjs );
    Vec_StrWrite( pFile, p.pFncs );
    _ = (pFile).Close();
}

/**Fnction*************************************************************

  Synopsis    [Creates manager after reading.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManReadBinaryBlif(pFileName string) *Bbl_Man_t {
    var p *Bbl_Man_t;
    var pObj *Bbl_Obj_t;
    var pBuffer string;
    var h int32;
    // inlining of macro BBLIF_ALLOC
    p = ((Bbl_Man_t *) malloc(sizeof(Bbl_Man_t) * (1)));
    memset( p, 0, sizeof(Bbl_Man_t) );
    p.nFileSize = Bbl_ManFileSize( pFileName );
    p.pFileData = Bbl_ManFileRead( pFileName );
    // extract three managers
    pBuffer = p.pFileData;
    p.pName = Vec_StrRead( &pBuffer );
    p.pObjs = Vec_StrRead( &pBuffer );
    p.pFncs = Vec_StrRead( &pBuffer );
    assert(pBuffer - p.pFileData == p.nFileSize, "pBuffer - p.pFileData == p.nFileSize");
    // remember original IDs in the objects
    p.vObj2Id = Vec_IntAlloc( 1000 );
    // inlining of macro Bbl_ManForEachObj_int
    for  h = 0; (h < p.pObjs.nSize) && (pObj = Bbl_VecObj(p.pObjs,h)); h += Bbl_ObjSize(pObj)  {
        Vec_IntPush( p.vObj2Id, pObj.Id );
        pObj.Id = Vec_IntSize(p.vObj2Id) - 1; 
    }
    return p;
}

/**Fnction*************************************************************

  Synopsis    [Prints stats of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManPrintStats(p *Bbl_Man_t) {
    var pObj *Bbl_Obj_t;
    var pFnc *Bbl_Fnc_t;
    var h, nFuncs, nNodes, nObjs int32;
    // inlining of macro Bbl_ManForEachObj_int
    for  h = 0; (h < p.pObjs.nSize) && (pObj = Bbl_VecObj(p.pObjs,h)); h += Bbl_ObjSize(pObj)  {
        nObjs++, nNodes += Bbl_ObjIsNode(pObj);
    }
    // inlining of macro Bbl_ManForEachFnc_int
    for  h = 0; (h < p.pFncs.nSize) && (pFnc = Bbl_VecFnc(p.pFncs,h)); h += Bbl_FncSize(pFnc)  {
        nFuncs++;
    }
    fmt.Printf("Total objects = %7d.  Total nodes = %7d. Unique functions = %7d.\n", nObjs, nNodes, nFuncs);
    fmt.Printf("Name manager = %5.2f MB\n", 1.0*Vec_StrSize(p.pName)/(1 << 20));
    fmt.Printf("Objs manager = %5.2f MB\n", 1.0*Vec_StrSize(p.pObjs)/(1 << 20));
    fmt.Printf("Fncs manager = %5.2f MB\n", 1.0*Vec_StrSize(p.pFncs)/(1 << 20));
}

/**Fnction*************************************************************

  Synopsis    [Deletes the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManStop(p *Bbl_Man_t) {
    if p.vId2Obj {    Vec_IntFree( p.vId2Obj ); }
    if p.vObj2Id {    Vec_IntFree( p.vObj2Id ); }
    if p.vFaninNums { Vec_IntFree( p.vFaninNums ); }
    if p.pFileData {
        // inlining of macro BBLIF_FREE
        ((p.pFileData) ? (free((char *) (p.pFileData)), (p.pFileData) = 0) : 0);
        p.pName.pArray = nil;
        p.pObjs.pArray = nil;
        p.pFncs.pArray = nil;
    }
    if p.pEnts {
    Vec_StrFree( p.pEnts ); }
    Vec_StrFree( p.pName );
    Vec_StrFree( p.pObjs );
    Vec_StrFree( p.pFncs );
    // inlining of macro BBLIF_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Fnction*************************************************************

  Synopsis    [Creates manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManStart(pName string) *Bbl_Man_t {
    var p *Bbl_Man_t;
    var nLength int32;
    // inlining of macro BBLIF_ALLOC
    p = ((Bbl_Man_t *) malloc(sizeof(Bbl_Man_t) * (1)));
    memset( p, 0, sizeof(Bbl_Man_t) );
    if pName {
        nLength = 4 * ((len(pName) + 1) / 4 + 1);
    } else {
        nLength = 0;
    }
    p.pName  = Vec_StrAlloc( nLength );
    p.pName.nSize = p.pName.nCap;
    if pName {
        strcpy( p.pName.pArray, pName ); }
    p.pObjs = Vec_StrAlloc( 1 << 16 );
    p.pFncs = Vec_StrAlloc( 1 << 16 );
    p.pEnts = Vec_StrAlloc( 1 << 16 ); p.pEnts.nSize = 1;
    p.vId2Obj = Vec_IntStart( 1 << 10 );
    p.vFaninNums = Vec_IntStart( 1 << 10 );
    return p;
}




/**Function*************************************************************

  Synopsis    [Performs selection sort on the array of cubes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManSortCubes(pCubes *string, nCubes int32, nVars int32) {
    var pTemp string;
    var i, j, best_i int32;
    for  i = 0; i < nCubes-1; i++  {
        best_i = i;
        for j = i+1; j < nCubes; j++ {
            if memcmp( pCubes[j], pCubes[best_i], (size_t)nVars ) < 0 {
                best_i = j; }
        }
        pTemp = pCubes[i]; pCubes[i] = pCubes[best_i]; pCubes[best_i] = pTemp;
    }
}

/**Function*************************************************************

  Synopsis    [Sorts the cubes in the SOP to uniqify them to some extent.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManSortSop(pSop string, nVars int32) string {
    var pCubes *string;
    var pSopNew string;
    var c, Length, nCubes int32;
    Length = len(pSop);
    assert(Length % (nVars + 3) == 0, "Length % (nVars + 3) == 0");
    nCubes = Length / (nVars + 3);
    if nCubes < 2 {
        // inlining of macro BBLIF_ALLOC
        pSopNew = ((char *) malloc(sizeof(char) * (Length + 1)));
        memcpy( pSopNew, pSop, (size_t)(Length + 1) );
        return pSopNew;
    }
    // inlining of macro BBLIF_ALLOC
    pCubes = ((char * *) malloc(sizeof(char *) * (nCubes)));
    for  c = 0; c < nCubes; c++  {
        pCubes[c] = pSop + c * (nVars + 3);
    }
    if nCubes < 300 {
        Bbl_ManSortCubes( pCubes, nCubes, nVars ); }
    // inlining of macro BBLIF_ALLOC
    pSopNew = ((char *) malloc(sizeof(char) * (Length + 1)));
    for  c = 0; c < nCubes; c++  {
        memcpy( pSopNew + c * (nVars + 3), pCubes[c], (size_t)(nVars + 3) );
    }
    // inlining of macro BBLIF_FREE
    ((pCubes) ? (free((char *) (pCubes)), (pCubes) = 0) : 0);
    pSopNew[nCubes * (nVars + 3)] = 0;
    return pSopNew;
}

/**Fnction*************************************************************

  Synopsis    [Saves one entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManCreateEntry(p *Bbl_Man_t, iFunc int32, iNext int32) int32 {
    var pEnt *Bbl_Ent_t;
    pEnt = (Bbl_Ent_t *)Vec_StrFetch( p.pEnts, 2 * sizeof(int) );
    pEnt.iFunc = iFunc;
    pEnt.iNext = iNext;
    return (char *)pEnt - p.pEnts.pArray;
}

/**Function*************************************************************

  Synopsis    [Sorts the cubes in the SOP to uniqify them to some extent.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManSopCheckUnique(p *Bbl_Man_t, pSop string, nVars int32, nCubes int32, iFunc int32) int32 {
    var pFnc *Bbl_Fnc_t;
    var pEnt *Bbl_Ent_t;
    var h int32;
    var Length int32 = len(pSop) + 1;
    var nWords int32  = (Length / 4 + (Length % 4 > 0));
    if nVars > 16 {   nVars = 16; }
    if nCubes > 16 {  nCubes = 16; }
//    if ( nVars == 16 && nCubes == 16 )
//        return iFunc;
    for  h = p.SopMap[nVars][nCubes]; (h != 0); h = pEnt.iNext  {
        pEnt = Bbl_VecEnt( p.pEnts, h );
        pFnc = Bbl_VecFnc( p.pFncs, pEnt.iFunc );
        assert(nVars == 16 || nCubes == 16 || pFnc.nWords == nWords, "nVars == 16 || nCubes == 16 || pFnc.nWords == nWords");
        if pFnc.nWords == nWords && memcmp( pFnc.pWords, pSop, (size_t)Length ) == 0 {
            return pEnt.iFunc; }
    }
    p.SopMap[nVars][nCubes] = Bbl_ManCreateEntry( p, iFunc, p.SopMap[nVars][nCubes] );
    return iFunc;
}

/**Fnction*************************************************************

  Synopsis    [Saves one SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManSaveSop(p *Bbl_Man_t, pSop string, nVars int32) int32 {
    var pFnc *Bbl_Fnc_t;
    var pSopNew string;
    var iFunc int32;
    var Length int32 = len(pSop) + 1;
    var nWords int32  = Length / 4 + (Length % 4 > 0);
    // reorder cubes to semi-canicize SOPs
    pSopNew = Bbl_ManSortSop( pSop, nVars );
    // get the candidate location
    iFunc = Bbl_ManSopCheckUnique( p, pSopNew, nVars, Length / (nVars + 3), Vec_StrSize(p.pFncs) );
//    iFunc = Vec_StrSize(p->pFncs);
    if iFunc == Vec_StrSize(p.pFncs) { // store this SOP
        pFnc = (Bbl_Fnc_t *)Vec_StrFetch( p.pFncs, sizeof(Bbl_Fnc_t) + nWords * sizeof(int) );
        pFnc.pWords[nWords-1] = 0;
        pFnc.nWords = nWords;
        strcpy( (char *)pFnc.pWords, pSopNew );
        assert(iFunc == (char *)pFnc - p.pFncs.pArray, "iFunc == (char *)pFnc - p.pFncs.pArray");
    }
    // inlining of macro BBLIF_FREE
    ((pSopNew) ? (free((char *) (pSopNew)), (pSopNew) = 0) : 0);
    return iFunc;
}

/**Fnction*************************************************************

  Synopsis    [Adds one object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManCreateObject(p *Bbl_Man_t, Type Bbl_Type_t, ObjId int32, nFanins int32, pSop string) {
    var pObj *Bbl_Obj_t;
    if Type == BBL_OBJ_CI && nFanins != 0 {
        fmt.Printf("Attempting to create a combinational input with %d fanins (should be 0).\n", nFanins);
        return;
    }
    if Type == BBL_OBJ_CO && nFanins != 1 {
        fmt.Printf("Attempting to create a combinational output with %d fanins (should be 1).\n", nFanins);
        return;
    }
    pObj = (Bbl_Obj_t *)Vec_StrFetch( p.pObjs, sizeof(Bbl_Obj_t) + nFanins * sizeof(int) );
    memset( pObj, 0, sizeof(Bbl_Obj_t) );
    Vec_IntSetEntry( p.vId2Obj, ObjId, (char *)pObj - p.pObjs.pArray );
    Vec_IntSetEntry( p.vFaninNums, ObjId, 0 );
    pObj.fCi = (Type == BBL_OBJ_CI);
    pObj.fCo = (Type == BBL_OBJ_CO);
    pObj.Id  = ObjId;
    if pSop {
        pObj.Fnc = Bbl_ManSaveSop(p, pSop, nFanins);
    } else {
        pObj.Fnc = -1;
    }
    pObj.nFanins = nFanins;
}

/**Fnction*************************************************************

  Synopsis    [Creates fanin/fanout relationship between two objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManAddFanin(p *Bbl_Man_t, ObjId int32, FaninId int32) {
    var pObj, pFanin *Bbl_Obj_t;
    var iFanin int32;
    pObj   = Bbl_ManObj( p, ObjId );
    if Bbl_ObjIsCi(pObj) != 0 {
        fmt.Printf("Bbl_ManAddFanin(): Cannot add fanin of the combinational input (Id = %d).\n", ObjId);
        return;
    }
    pFanin = Bbl_ManObj( p, FaninId );
    if Bbl_ObjIsCo(pFanin) != 0 {
        fmt.Printf("Bbl_ManAddFanin(): Cannot add fanout of the combinational output (Id = %d).\n", FaninId);
        return;
    }
    iFanin = Vec_IntEntry( p.vFaninNums, ObjId );
    if iFanin >= int32(pObj.nFanins) {
        fmt.Printf("Bbl_ManAddFanin(): Trying to add more fanins to object (Id = %d) than declared (%d).\n", ObjId, pObj.nFanins);
        return;
    }
    assert(iFanin < int32(pObj.nFanins), "iFanin < int32(pObj.nFanins)");
    Vec_IntWriteEntry( p.vFaninNums, ObjId, iFanin+1 );
    pObj.pFanins[iFanin] = (char *)pObj - (char *)pFanin;
}


/**Fnction*************************************************************

  Synopsis    [Returns 1 if the manager was created correctly.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManCheck(p *Bbl_Man_t) int32 {
    var pObj *Bbl_Obj_t;
    var h int32;
    var RetValue int32 = 1;
    // inlining of macro Bbl_ManForEachObj_int
    for  h = 0; (h < p.pObjs.nSize) && (pObj = Bbl_VecObj(p.pObjs,h)); h += Bbl_ObjSize(pObj)  {
        if (Bbl_ObjIsNode(pObj) != 0) && pObj.Fnc == -1 {
            RetValue = 0;
fmt.Printf("Bbl_ManCheck(): Node %d does not have function specified.\n", pObj.Id); }
        if (Bbl_ObjIsCi(pObj) != 0) && pObj.Fnc != -1 {
            RetValue = 0;
fmt.Printf("Bbl_ManCheck(): CI with %d has function specified.\n", pObj.Id); }
        if (Bbl_ObjIsCo(pObj) != 0) && pObj.Fnc != -1 {
            RetValue = 0;
fmt.Printf("Bbl_ManCheck(): CO with %d has function specified.\n", pObj.Id); }
        if Vec_IntEntry(p.vFaninNums, pObj.Id) != int32(pObj.nFanins) {
            RetValue = 0;
fmt.Printf("Bbl_ManCheck(): Object %d has less fanins (%d) than declared (%d).\n", pObj.Id, Vec_IntEntry(p.vFaninNums, pObj.Id), pObj.nFanins); }
    }
    return RetValue;
}


/**Fnction*************************************************************

  Synopsis    [Misc APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ObjIsInput(p *Bbl_Obj_t) int32 { return Bbl_ObjIsCi(p);                      }
func Bbl_ObjIsOutput(p *Bbl_Obj_t) int32 { return Bbl_ObjIsCo(p);                      }
func Bbl_ObjIsLut(p *Bbl_Obj_t) int32 { return Bbl_ObjIsNode(p);                    }
func Bbl_ObjId(p *Bbl_Obj_t) int32 { return p.Id;                               }
func Bbl_ObjIdOriginal(pMan *Bbl_Man_t, p *Bbl_Obj_t) int32 { assert(false, "0"); return Vec_IntEntry(pMan.vObj2Id, p.Id);  }
func Bbl_ObjFaninNumber(p *Bbl_Obj_t) int32 { return Bbl_ObjFaninNum(p);                  }
func Bbl_ObjSop(pMan *Bbl_Man_t, p *Bbl_Obj_t) string { return Bbl_ManSop(pMan, p.Fnc);            }
func Bbl_ObjIsMarked(p *Bbl_Obj_t) int32 { return p.fMark;                            }
func Bbl_ObjMark(p *Bbl_Obj_t) { p.fMark = 1;                               }
func Bbl_ObjFncHandle(p *Bbl_Obj_t) int32 { return p.Fnc;                              }

/**Fnction*************************************************************

  Synopsis    [Returns the name of the design.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManName(p *Bbl_Man_t) string {
    return p.pName.pArray;
}

/**Fnction*************************************************************

  Synopsis    [Returns the maximum handle of the SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManFncSize(p *Bbl_Man_t) int32 {
    return p.pFncs.nSize;
}

/**Fnction*************************************************************

  Synopsis    [Returns the first object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManObjFirst(p *Bbl_Man_t) *Bbl_Obj_t {
    return Bbl_VecObj( p.pObjs, 0 );
}

/**Fnction*************************************************************

  Synopsis    [Returns the next object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManObjNext(p *Bbl_Man_t, pObj *Bbl_Obj_t) *Bbl_Obj_t {
    var pNext string  = (char *)pObj + Bbl_ObjSize(pObj);
    var pEdge string  = p.pObjs.pArray + p.pObjs.nSize;
    return (Bbl_Obj_t *)(pNext < pEdge ? pNext : nil);
}

/**Fnction*************************************************************

  Synopsis    [Returns the first fanin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ObjFaninFirst(p *Bbl_Obj_t) *Bbl_Obj_t {
    if Bbl_ObjFaninNum(p) != 0 {
        return Bbl_ObjFanin( p, 0 );
    } else {
        return nil;
    }
}

/**Fnction*************************************************************

  Synopsis    [Returns the next fanin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ObjFaninNext(p *Bbl_Obj_t, pPrev *Bbl_Obj_t) *Bbl_Obj_t {
    var pFanin *Bbl_Obj_t;
    var i int32;
    // inlining of macro Bbl_ObjForEachFanin_int
    for  i = 0; (i < int32(p.nFanins)) && (pFanin = Bbl_ObjFanin(p,i)); i++  {
        if pFanin == pPrev {
            break; }
    }
    if i < Bbl_ObjFaninNum(p) - 1 {
        return Bbl_ObjFanin( p, i+1 );
    } else {
        return nil;
    }
}

/**Fnction*************************************************************

  Synopsis    [Drives text BLIF file for debugging.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManDumpBlif(p *Bbl_Man_t, pFileName string) {
    var pFile *os.File;
    var pObj, pFanin *Bbl_Obj_t;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pFile, "# Test file written by Bbl_ManDumpBlif() in ABC.\n");
    fmt.Fprintf(pFile, ".model %s\n", Bbl_ManName(p));
    // write objects
    // inlining of macro Bbl_ManForEachObj
    for  pObj = Bbl_ManObjFirst(p); (pObj != nil); pObj = Bbl_ManObjNext(p, pObj)  {
        if Bbl_ObjIsInput(pObj) != 0 {
            fmt.Fprintf(pFile, ".inputs %d\n", Bbl_ObjId(pObj));
        } else if Bbl_ObjIsOutput(pObj) != 0 {
            fmt.Fprintf(pFile, ".outputs %d\n", Bbl_ObjId(pObj));
        } else if Bbl_ObjIsLut(pObj) != 0 {
            fmt.Fprintf(pFile, ".names");
            // inlining of macro Bbl_ObjForEachFanin
            for  pFanin = Bbl_ObjFaninFirst(pObj); (pFanin != nil); pFanin = Bbl_ObjFaninNext(pObj, pFanin)  {
                fmt.Fprintf(pFile, " %d", Bbl_ObjId(pFanin));
            }
            fmt.Fprintf(pFile, " %d\n", Bbl_ObjId(pObj));
            fmt.Fprintf(pFile, "%s", Bbl_ObjSop(p, pObj));
        
        } else { assert(false, "0"); }
    }
    // write output drivers
    // inlining of macro Bbl_ManForEachObj
    for  pObj = Bbl_ManObjFirst(p); (pObj != nil); pObj = Bbl_ManObjNext(p, pObj)  {
        if !(Bbl_ObjIsOutput(pObj) != 0) {
            continue; }
        fmt.Fprintf(pFile, ".names");
        // inlining of macro Bbl_ObjForEachFanin
        for  pFanin = Bbl_ObjFaninFirst(pObj); (pFanin != nil); pFanin = Bbl_ObjFaninNext(pObj, pFanin)  {
            fmt.Fprintf(pFile, " %d", Bbl_ObjId(pFanin));
        }
        fmt.Fprintf(pFile, " %d\n", Bbl_ObjId(pObj));
        fmt.Fprintf(pFile, "1 1\n");
    }
    fmt.Fprintf(pFile, ".end\n");
    _ = (pFile).Close();
}

/**Fnction*************************************************************

  Synopsis    [Converting truth table into an SOP.]

  Description [The truth table is given as a bit-string pTruth
  composed of 2^nVars bits. The result is an SOP derived by
  collecting minterms appearing in the truth table. The SOP is 
  represented as a C-string, as documented in file "bblif.h".
  It is recommended to limit the use of this procedure to Boolean 
  functions up to 6 inputs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManTruthToSop(pTruth *uint32, nVars int32) string {
    var pResult, pTemp string;
    var nMints, nOnes, b, v int32;
    assert(nVars >= 0 && nVars <= 16, "nVars >= 0 && nVars <= 16");
    nMints = (1 << nVars);
    // count the number of ones
    nOnes = 0;
    for  b = 0; b < nMints; b++  {
        nOnes += ((pTruth[b>>5] >> (b&31)) & 1);
    }
    // handle constants
    if nOnes == 0 || nOnes == nMints {
        // inlining of macro BBLIF_ALLOC
        pResult = pTemp = ((char *) malloc(sizeof(char) * (nVars + 4)));        
        for  v = 0; v < nVars; v++  {
            *pTemp++ = '-';
        }
        *pTemp++ = ' ';
        *pTemp++ = nOnes? '1' : '0';
        *pTemp++ = '\n';
        *pTemp++ = 0;
        assert(pTemp - pResult == nVars + 4, "pTemp - pResult == nVars + 4");
        return pResult;
    }
    // inlining of macro BBLIF_ALLOC
    pResult = pTemp = ((char *) malloc(sizeof(char) * (nOnes * (nVars + 3) + 1)));        
    for  b = 0; b < nMints; b++  {
        if ((pTruth[b>>5] >> (b&31)) & 1) == 0 {
            continue; }
        for  v = 0; v < nVars; v++  {
            *pTemp++ = ((b >> v) & 1)? '1' : '0';
        }
        *pTemp++ = ' ';
        *pTemp++ = '1';
        *pTemp++ = '\n';
    }
    *pTemp++ = 0;
    assert(pTemp - pResult == nOnes * (nVars + 3) + 1, "pTemp - pResult == nOnes * (nVars + 3) + 1");
    return pResult;
}

/**Function*************************************************************

  Synopsis    [Allocates the array of truth tables for the given number of vars.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManSopToTruthElem(nVars int32, pVars **uint32) {
    var Masks [5]uint32 = [5]uint32{ 0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00, 0xFFFF0000 };
    var i, k, nWords int32;
    if nVars <= 5 {
        nWords = 1;
    } else {
        nWords = (1 << (nVars - 5));
    }
    for  i = 0; i < nVars; i++  {
        if i < 5 {
            for  k = 0; k < nWords; k++  {
                pVars[i][k] = Masks[i];
            }
        
        } else {
            for  k = 0; k < nWords; k++  {
                if k & (1 << (i-5)) {
                    pVars[i][k] = ~uint32(0);
                } else {
                    pVars[i][k] = 0; }
            }
        }
    }
}

/**Fnction*************************************************************

  Synopsis    [Converting SOP into a truth table.]

  Description [The SOP is represented as a C-string, as documented in 
  file "bblif.h". The truth table is returned as a bit-string composed 
  of 2^nVars bits. For functions of less than 6 variables, the full
  machine word is returned. (The truth table looks as if the function
  had 5 variables.) The use of this procedure should be limited to 
  Boolean functions with no more than 16 inputs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManSopToTruth(pSop string, nVars int32) *uint32 {
    var pTruth *uint32;
    var pCube *uint32;
    var pVars [16]*uint32;
    var nWords int32;
    if nVars <= 5 {
        nWords = 1;
    } else {
        nWords = (1 << (nVars - 5));
    }
    var v, c, w, nCubes, fCompl int32;
    if pSop == nil {
        return nil; }
    if len(pSop) % (nVars + 3) != 0 {
        fmt.Printf("Bbl_ManSopToTruth(): SOP is represented incorrectly.\n");
        return nil;
    }
    // create storage for TTs of the result, elementary variables and the temp cube
    // inlining of macro BBLIF_ALLOC
    pTruth   = ((unsigned *) malloc(sizeof(unsigned) * (nWords)));   
    // inlining of macro BBLIF_ALLOC
    pVars[0] = ((unsigned *) malloc(sizeof(unsigned) * (nWords * (nVars+1))));   
    for  v = 1; v < nVars; v++  {
        pVars[v] = pVars[v-1] + nWords;
    }
    pCube = pVars[v-1] + nWords;
    Bbl_ManSopToTruthElem( nVars, pVars );
    // iterate through the cubes
    memset( pTruth, 0, sizeof(unsigned) * nWords );
    nCubes =  len(pSop) / (nVars + 3);
    for  c = 0; c < nCubes; c++  {
        fCompl = (pSop[nVars+1] == '0');
        memset( pCube, 0xff, sizeof(unsigned) * nWords );
        // iterate through the literals of the cube
        for  v = 0; v < nVars; v++  {
            if pSop[v] == '1' {
                for  w = 0; w < nWords; w++  {
                    pCube[w] &= pVars[v][w];
                }
            } else if pSop[v] == '0' {
                for  w = 0; w < nWords; w++  {
                    pCube[w] &= ~pVars[v][w];
                } }
        }
        // add cube to storage
        for  w = 0; w < nWords; w++  {
            pTruth[w] |= pCube[w];
        }
        // go to the next cube
        pSop += (nVars + 3);
    }
    // inlining of macro BBLIF_FREE
    ((pVars[0]) ? (free((char *) (pVars[0])), (pVars[0]) = 0) : 0);
    if fCompl != 0 {
        for  w = 0; w < nWords; w++  {
            pTruth[w] = ~pTruth[w];
        } }
    return pTruth;
}


/**Fnction*************************************************************

  Synopsis    [Checks the truth table computation.]

  Description [We construct the logic network for the half-adder represnted
  using the BLIF file below]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManTestTruth(pSop string, nVars int32) {
    var pTruth *uint32;
    var pSopNew string;
    pTruth = Bbl_ManSopToTruth( pSop, nVars );
    pSopNew = Bbl_ManTruthToSop( pTruth, nVars );
    fmt.Printf("Old SOP:\n%s\n", pSop);
    fmt.Printf("New SOP:\n%s\n", pSopNew);
    // inlining of macro BBLIF_FREE
    ((pSopNew) ? (free((char *) (pSopNew)), (pSopNew) = 0) : 0);
    // inlining of macro BBLIF_FREE
    ((pTruth) ? (free((char *) (pTruth)), (pTruth) = 0) : 0);
}

/**Fnction*************************************************************

  Synopsis    [This demo shows using the internal to construct a half-adder.]

  Description [We construct the logic network for the half-adder represnted
  using the BLIF file below]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManSimpleDemo() {
/*
    # There are contents of a BLIF file representing a half-adder:
    .model hadder
    .inputs a     // ID = 1
    .inputs b     // ID = 2
    .inputs cin   // ID = 3
    .outputs s    // ID = 4
    .outputs cout // ID = 5
    .names a b cin s_driver   // ID = 6
    100 1
    010 1
    001 1
    111 1
    .names a b cin cout_driver // ID = 7
    -11 1
    1-1 1
    11- 1
    .names s_driver s
    1 1
    .names cout_driver cout
    1 1
    .end
*/
    var p *Bbl_Man_t;
    // start the data manager
    p = Bbl_ManStart( "hadder" );
    // create CIs
    Bbl_ManCreateObject( p, BBL_OBJ_CI, 1, 0, nil ); // a
    Bbl_ManCreateObject( p, BBL_OBJ_CI, 2, 0, nil ); // b
    Bbl_ManCreateObject( p, BBL_OBJ_CI, 3, 0, nil ); // cin
    // create COs
    Bbl_ManCreateObject( p, BBL_OBJ_CO, 4, 1, nil ); // s
    Bbl_ManCreateObject( p, BBL_OBJ_CO, 5, 1, nil ); // cout
    // create internal nodes
    Bbl_ManCreateObject( p, BBL_OBJ_NODE, 6, 3, "100 1\n010 1\n001 1\n111 1\n" ); // s_driver
    Bbl_ManCreateObject( p, BBL_OBJ_NODE, 7, 3, "-11 1\n1-1 1\n11- 1\n" );        // cout_driver
    // add fanins of node 6
    Bbl_ManAddFanin( p, 6, 1 ); // s_driver <- a
    Bbl_ManAddFanin( p, 6, 2 ); // s_driver <- b
    Bbl_ManAddFanin( p, 6, 3 ); // s_driver <- cin
    // add fanins of node 7
    Bbl_ManAddFanin( p, 7, 1 ); // cout_driver <- a
    Bbl_ManAddFanin( p, 7, 2 ); // cout_driver <- b
    Bbl_ManAddFanin( p, 7, 3 ); // cout_driver <- cin
    // add fanins of COs
    Bbl_ManAddFanin( p, 4, 6 ); // s <- s_driver
    Bbl_ManAddFanin( p, 5, 7 ); // cout <- cout_driver
    // sanity check
    Bbl_ManCheck( p );
    // write BLIF file as a sanity check
    Bbl_ManDumpBlif( p, "hadder.blif" );
    // write binary BLIF file
    Bbl_ManDumpBinaryBlif( p, "hadder.bblif" );
    // remove the manager
    Bbl_ManStop( p );


//    Bbl_ManTestTruth( "100 1\n010 1\n001 1\n111 1\n", 3 );
//    Bbl_ManTestTruth( "-11 0\n1-1 0\n11- 0\n", 3 );
//    Bbl_ManTestTruth( "--- 1\n", 3 );
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

