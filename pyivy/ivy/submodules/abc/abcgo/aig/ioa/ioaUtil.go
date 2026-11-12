package ioa

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ioaUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to read binary AIGER format developed by
  Armin Biere, Johannes Kepler University (http://fmv.jku.at/)]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - December 16, 2006.]

  Revision    [$Id: ioaUtil.c,v 1.00 2006/12/16 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioa.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the file size.]

  Description [The file should be closed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ioa_FileSize(pFileName string) int32 {
    var pFile *os.File;
    var nFileSize int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Ioa_FileSize(): The file is unavailable (absent or open).\n");
        return 0;
    }
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile ); 
    _ = (pFile).Close();
    return nFileSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ioa_FileNameGeneric(FileName string) string {
    var pDot, pRes string;
    pRes = Abc_UtilStrsav( FileName );
    if pDot = strrchr( pRes, '.' ) {
        *pDot = 0; }
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Returns the composite name of the file.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ioa_FileNameGenericAppend_Buffer [1000]byte;
func Ioa_FileNameGenericAppend(pBase string, pSuffix string) string {
    
    var pDot string;
    if pBase == nil {
        strcpy( spStatic_Ioa_FileNameGenericAppend_Buffer, pSuffix );
        return spStatic_Ioa_FileNameGenericAppend_Buffer;
    }
    strcpy( spStatic_Ioa_FileNameGenericAppend_Buffer, pBase );
    if pDot = strrchr( spStatic_Ioa_FileNameGenericAppend_Buffer, '.' ) {
        *pDot = 0; }
    strcat( spStatic_Ioa_FileNameGenericAppend_Buffer, pSuffix );
    // find the last occurrance of slash
    for  pDot = spStatic_Ioa_FileNameGenericAppend_Buffer + strlen(spStatic_Ioa_FileNameGenericAppend_Buffer) - 1; pDot >= spStatic_Ioa_FileNameGenericAppend_Buffer; pDot--  {    
        if !((*pDot >= '0' && *pDot <= '9') ||
              (*pDot >= 'a' && *pDot <= 'z') ||
              (*pDot >= 'A' && *pDot <= 'Z') || 
               *pDot == '_' || *pDot == '.') {
               break; }
    }
    return pDot + 1;
}

/**Function*************************************************************

  Synopsis    [Returns the time stamp.]

  Description [The file should be closed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ioa_TimeStamp_Buffer [100]byte;
func Ioa_TimeStamp() string {
    
    var TimeStamp string;
    var ltime int64;
    // get the current time
    time( &ltime );
    TimeStamp = asctime( localtime( &ltime ) );
    TimeStamp[ len(TimeStamp) - 1 ] = 0;
    strcpy( spStatic_Ioa_TimeStamp_Buffer, TimeStamp );
    return spStatic_Ioa_TimeStamp_Buffer;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

