package pla

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [plaRead.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SOP manager.]

  Synopsis    [Scalable SOP transformations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - March 18, 2015.]

  Revision    [$Id: plaRead.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "pla.h"

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
func Pla_ReadFile(pFileName string, ppLimit *string) string {
    var pBuffer string;
    var nFileSize, RetValue int32;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open input file.\n");
        return nil;
    }
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile );  
    // move the file current reading position to the beginning
    rewind( pFile ); 
    // load the contents of the file into memory
    pBuffer = ABC_ALLOC( char, nFileSize + 16 );
    pBuffer[0] = '\n';
    RetValue = fread( pBuffer+1, nFileSize, 1, pFile );
    _ = (pFile).Close();
    // terminate the string with '\0'
    pBuffer[nFileSize + 1] = '\n';
    pBuffer[nFileSize + 2] = '\0';
    *ppLimit = pBuffer + nFileSize + 3;
    return pBuffer;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pla_ReadPlaRemoveComments(pBuffer string, pLimit string) {
    var pTemp string; 
    for  pTemp = pBuffer; pTemp < pLimit; pTemp++  {
        if *pTemp == '#' {
            for  *pTemp && *pTemp != '\n'  {
                *pTemp++ = ' ';
            } }
    }
}
func Pla_ReadPlaHeader(pBuffer string, pLimit string, pnIns *int32, pnOuts *int32, pnCubes *int32, pType *int32) int32 {
    var pTemp string; 
    *pType = PLA_FILE_FD;
    *pnIns = *pnOuts = *pnCubes = -1;
    for  pTemp = pBuffer; pTemp < pLimit; pTemp++  {
        if *pTemp != '.' {
            continue; }
        if !strncmp(pTemp, ".i ", 3) {
            *pnIns = strconv.Atoi(pTemp + 3);
        } else if !strncmp(pTemp, ".o ", 3) {
            *pnOuts = strconv.Atoi(pTemp + 3);
        } else if !strncmp(pTemp, ".p ", 3) {
            *pnCubes = strconv.Atoi(pTemp + 3);
        } else if !strncmp(pTemp, ".e ", 3) {
            break;
        } else if !strncmp(pTemp, ".type ", 6) {
            var Buffer [100]byte;
            *pType = PLA_FILE_NONE;
            sscanf( pTemp+6, "%s", Buffer );
            if !strcmp(Buffer, "f") {
                *pType = PLA_FILE_F;
            } else if !strcmp(Buffer, "fr") {
                *pType = PLA_FILE_FR;
            } else if !strcmp(Buffer, "fd") {
                *pType = PLA_FILE_FD;
            } else if !strcmp(Buffer, "fdr") {
                *pType = PLA_FILE_FDR; }
        }
    }
    if *pnIns <= 0 {
        fmt.Printf("The number of inputs (.i) should be positive.\n"); }
    if *pnOuts <= 0 {
        fmt.Printf("The number of outputs (.o) should be positive.\n"); }
    return *pnIns > 0 && *pnOuts > 0;
}
func Pla_ReadPlaBody(pBuffer string, pLimit string, Type Pla_File_t) *Vec_Str_t {
    var pTemp string;
    var vLits *Vec_Str_t;
    vLits = Vec_StrAlloc( 10000 );
    for  pTemp = pBuffer; pTemp < pLimit; pTemp++  {
        if *pTemp == '.' {
            for  *pTemp && *pTemp != '\n'  {
                pTemp++;
            } }
        if *pTemp == '0' {
            Vec_StrPush( vLits, byte(PLA_LIT_ZERO) );
        } else if *pTemp == '1' {
            Vec_StrPush( vLits, byte(PLA_LIT_ONE) );
        } else if *pTemp == '-' || *pTemp == '2' {
            Vec_StrPush( vLits, byte(PLA_LIT_DASH) );
        } else if *pTemp == '~' { // no meaning
        
            if Type == PLA_FILE_F || Type == PLA_FILE_FD {
                Vec_StrPush( vLits, byte(PLA_LIT_ZERO) );
            } else if Type == PLA_FILE_FR {
                Vec_StrPush( vLits, byte(PLA_LIT_DASH) );
            } else if Type == PLA_FILE_FDR {
                Vec_StrPush( vLits, byte(PLA_LIT_FULL) );
            } else { assert(false, "0"); }
        }
    }
    return vLits;
}
func Pla_ReadAddBody(p *Pla_Man_t, vLits *Vec_Str_t) {
    var pCubeIn, pCubeOut *word; 
    var i, k, Lit, Count int32; 
    var nCubesReal int32  = Vec_StrSize(vLits) / (p.nIns + p.nOuts);
    assert(Vec_StrSize(vLits) % (p.nIns + p.nOuts) == 0, "Vec_StrSize(vLits) % (p.nIns + p.nOuts) == 0");
    if nCubesReal != Pla_ManCubeNum(p) {
        fmt.Printf("Warning: Declared number of cubes (%d) differs from the actual (%d).\n", Pla_ManCubeNum(p), nCubesReal);
        if nCubesReal < Pla_ManCubeNum(p) {
            Vec_IntShrink( &p.vCubes, nCubesReal );
        } else {
            assert(nCubesReal > Pla_ManCubeNum(p), "nCubesReal > Pla_ManCubeNum(p)");
            Vec_IntFillNatural( &p.vCubes, nCubesReal );
            Vec_WrdFillExtra( &p.vInBits,  nCubesReal * p.nInWords,  0 );
            Vec_WrdFillExtra( &p.vOutBits, nCubesReal * p.nOutWords, 0 );
        }
    }
    // inlining of macro Pla_ForEachCubeInOut
    for  i = 0; (i < Pla_ManCubeNum(p))  && (((pCubeIn) = Pla_CubeIn(p, i)), 1)  && (((pCubeOut) = Pla_CubeOut(p, i)), 1); i++  {
        // inlining of macro Pla_CubeForEachLit
        for  k = 0; (k < p.nIns)  && (((Lit) = Pla_CubeGetLit(pCubeIn, k)), 1); k++  {
            Pla_CubeSetLit( pCubeIn, k, (Pla_Lit_t)Vec_StrEntry(vLits, Count++) );
        }
        // inlining of macro Pla_CubeForEachLit
        for  k = 0; (k < p.nOuts)  && (((Lit) = Pla_CubeGetLit(pCubeOut, k)), 1); k++  {
            Pla_CubeSetLit( pCubeOut, k, (Pla_Lit_t)Vec_StrEntry(vLits, Count++) );
        }
    }
    assert(Count == Vec_StrSize(vLits), "Count == Vec_StrSize(vLits)");
}
func Pla_ReadPla(pFileName string) *Pla_Man_t {   
    var p *Pla_Man_t;
    var vLits *Vec_Str_t;
    var nIns, nOuts, nCubes, Type int32;
    var pBuffer, pLimit string;   
    pBuffer = Pla_ReadFile( pFileName, &pLimit );
    if pBuffer == nil {
        return nil; }
    Pla_ReadPlaRemoveComments( pBuffer, pLimit );
    if Pla_ReadPlaHeader( pBuffer, pLimit, &nIns, &nOuts, &nCubes, &Type ) != 0 {
        vLits = Pla_ReadPlaBody( pBuffer, pLimit, (Pla_File_t)Type );
        if Vec_StrSize(vLits) % (nIns + nOuts) == 0 {
            if nCubes == -1 {
                nCubes = Vec_StrSize(vLits) / (nIns + nOuts); }
            p = Pla_ManAlloc( pFileName, nIns, nOuts, nCubes );
            p.Type = (Pla_File_t)Type;
            Pla_ReadAddBody( p, vLits );
            Vec_StrFree( vLits );
            ABC_FREE( pBuffer );
            return p;
        }
        fmt.Printf("Literal count is incorrect (in = %d; out = %d; lit = %d).\n", nIns, nOuts, Vec_StrSize(vLits));
        Vec_StrFree( vLits );
    }
    ABC_FREE( pBuffer );
    return nil;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

