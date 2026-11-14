package ver

import (
	"os"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/st"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [ver.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 19, 2006.]

  Revision    [$Id: ver.h,v 1.00 2006/08/19 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include "base/abc/abc.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Ver_Man_t = Ver_Man_t_
type Ver_Stream_t = Ver_Stream_t_

type Ver_Man_t_ struct {
	// internal parameters
	fMapped    int32 // mapped verilog
	fUseMemMan int32 // allocate memory manager in the networks
	fCheck     int32 // checks network for currectness
	// input file stream
	pFileName string
	pReader   *Ver_Stream_t
	fNameLast int32
	pProgress *extra.ProgressBar
	// current design
	pDesign      *abc.Design
	pGateLib     *abc.Design
	pGenlib      any
	tName2Suffix *st.Table
	// error handling
	Output    *os.File
	fTopLevel int32
	fError    int32
	sError    [2000]byte
	// intermediate structures
	vNames   *vec.Vec_Ptr_t
	vStackFn *vec.Vec_Ptr_t
	vStackOp *vec.Vec_Int_t
	vPerm    *vec.Vec_Int_t
	boxBundles map[*abc.Object]*vec.Vec_Ptr_t
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== verCore.c ========================================================*/
// extern Abc_Des_t *    Ver_ParseFile( char * pFileName, Abc_Des_t * pGateLib, int fCheck, int fUseMemMan );

// extern void           Ver_ParsePrintErrorMessage( Ver_Man_t * p );

/*=== verFormula.c ========================================================*/
// extern void *         Ver_FormulaParser( char * pFormula, void * pMan, Vec_Ptr_t * vNames, Vec_Ptr_t * vStackFn, Vec_Int_t * vStackOp, char * pErrorMessage );

// extern void *         Ver_FormulaReduction( char * pFormula, void * pMan, Vec_Ptr_t * vNames, char * pErrorMessage );

/*=== verParse.c ========================================================*/
// extern int            Ver_ParseSkipComments( Ver_Man_t * p );

// extern char *         Ver_ParseGetName( Ver_Man_t * p );

/*=== verStream.c ========================================================*/
// extern Ver_Stream_t * Ver_StreamAlloc( char * pFileName );

// extern void           Ver_StreamFree( Ver_Stream_t * p );

// extern char *         Ver_StreamGetFileName( Ver_Stream_t * p );

// extern int            Ver_StreamGetFileSize( Ver_Stream_t * p );

// extern int            Ver_StreamGetCurPosition( Ver_Stream_t * p );

// extern int            Ver_StreamGetLineNumber( Ver_Stream_t * p );

// extern int            Ver_StreamIsOkey( Ver_Stream_t * p );

// extern char           Ver_StreamScanChar( Ver_Stream_t * p );

// extern char           Ver_StreamPopChar( Ver_Stream_t * p );

// extern void           Ver_StreamSkipChars( Ver_Stream_t * p, char * pCharsToSkip );

// extern void           Ver_StreamSkipToChars( Ver_Stream_t * p, char * pCharsToStop );

// extern char *         Ver_StreamGetWord( Ver_Stream_t * p, char * pCharsToStop );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
