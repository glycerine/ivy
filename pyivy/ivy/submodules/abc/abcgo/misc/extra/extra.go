package extra

import (
	"math/bits"
	"slices"
)

/**CFile****************************************************************

  FileName    [extra.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Various reusable software utilities.]

  Description [This library contains a number of operators and
  traversal routines developed to extend the functionality of
  CUDD v.2.3.x, by Fabio Somenzi (http://vlsi.colorado.edu/~fabio/)
  To compile your code with the library, #include "extra.h"
  in your source files and link your project to CUDD and this
  library. Use the library at your own risk and with caution.
  Note that debugging of some operators still continues.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extra.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

// Go requires no platform-specific inline keyword.

/*---------------------------------------------------------------------------*/
/* Nested includes                                                           */
/*---------------------------------------------------------------------------*/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/st/st.h"

//ABC_NAMESPACE_HEADER_START

/*---------------------------------------------------------------------------*/
/* Constant declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

// C uint8/uint16/uint32 aliases map directly to Go built-in integer types.

/*===========================================================================*/
/*     Various Utilities                                                     */
/*===========================================================================*/

/*=== extraUtilBitMatrix.c ================================================================*/

// extern Extra_BitMat_t * Extra_BitMatrixStart( int nSize );

// extern void         Extra_BitMatrixClean( Extra_BitMat_t * p );

// extern void         Extra_BitMatrixStop( Extra_BitMat_t * p );

// extern void         Extra_BitMatrixPrint( Extra_BitMat_t * p );

// extern int          Extra_BitMatrixReadSize( Extra_BitMat_t * p );

// extern void         Extra_BitMatrixInsert1( Extra_BitMat_t * p, int i, int k );

// extern int          Extra_BitMatrixLookup1( Extra_BitMat_t * p, int i, int k );

// extern void         Extra_BitMatrixDelete1( Extra_BitMat_t * p, int i, int k );

// extern void         Extra_BitMatrixInsert2( Extra_BitMat_t * p, int i, int k );

// extern int          Extra_BitMatrixLookup2( Extra_BitMat_t * p, int i, int k );

// extern void         Extra_BitMatrixDelete2( Extra_BitMat_t * p, int i, int k );

// extern void         Extra_BitMatrixOr( Extra_BitMat_t * p, int i, unsigned * pInfo );

// extern void         Extra_BitMatrixOrTwo( Extra_BitMat_t * p, int i, int j );

// extern int          Extra_BitMatrixCountOnesUpper( Extra_BitMat_t * p );

// extern int          Extra_BitMatrixIsDisjoint( Extra_BitMat_t * p1, Extra_BitMat_t * p2 );

// extern int          Extra_BitMatrixIsClique( Extra_BitMat_t * p );

/*=== extraUtilFile.c ========================================================*/

// extern char *       Extra_FileGetSimilarName( char * pFileNameWrong, char * pS1, char * pS2, char * pS3, char * pS4, char * pS5 );

// extern char *       Extra_FileNameExtension( char * FileName );

// extern char *       Extra_FileNameAppend( char * pBase, char * pSuffix );

// extern char *       Extra_FileNameGeneric( char * FileName );

// extern char *       Extra_FileNameGenericAppend( char * pBase, char * pSuffix );

// extern void         Extra_FileNameCorrectPath( char * FileName );

// extern char *       Extra_FileNameWithoutPath( char * FileName );

// extern char *       Extra_FilePathWithoutName( char * FileName );

// extern char *       Extra_FileDesignName( char * pFileName );

// extern int          Extra_FileCheck( char * pFileName );

// extern int          Extra_FileSize( char * pFileName );

// extern char *       Extra_FileRead( FILE * pFile );

// extern char *       Extra_FileRead2( FILE * pFile, FILE * pFile2 );

// extern char *       Extra_FileReadContents( char * pFileName );

// extern char *       Extra_FileReadContents2( char * pFileName, char * pFileName2 );

// extern int          Extra_FileIsType( char * pFileName, char * pS1, char * pS2, char * pS3 );

// extern char *       Extra_TimeStamp();

// extern char *       Extra_StringAppend( char * pStrGiven, char * pStrAdd );

// extern void         Extra_StringClean( char * pStrGiven, char * pCharKeep );

// extern unsigned     Extra_ReadBinary( char * Buffer );

// extern void         Extra_PrintBinary( FILE * pFile, unsigned Sign[], int nBits );

// extern void         Extra_PrintBinary2( FILE * pFile, unsigned Sign[], int nBits );

// extern int          Extra_ReadHex( unsigned Sign[], char * pString, int nDigits );

// extern int          Extra_ReadHexadecimal( unsigned Sign[], char * pString, int nVars );

// extern void         Extra_PrintHexadecimal( FILE * pFile, unsigned Sign[], int nVars );

// extern void         Extra_PrintHexadecimalString( char * pString, unsigned Sign[], int nVars );

// extern void         Extra_PrintHex( FILE * pFile, unsigned * pTruth, int nVars );

// extern void         Extra_PrintHex2( FILE * pFile, unsigned * pTruth, int nVars );

// extern void         Extra_PrintHexReverse( FILE * pFile, unsigned * pTruth, int nVars );

// extern void         Extra_PrintSymbols( FILE * pFile, char Char, int nTimes, int fPrintNewLine );

/*=== extraUtilReader.c ========================================================*/

// extern Extra_FileReader_t * Extra_FileReaderAlloc( char * pFileName,
//     char * pCharsComment, char * pCharsStop, char * pCharsClean );

// extern void         Extra_FileReaderFree( Extra_FileReader_t * p );

// extern char *       Extra_FileReaderGetFileName( Extra_FileReader_t * p );

// extern int          Extra_FileReaderGetFileSize( Extra_FileReader_t * p );

// extern int          Extra_FileReaderGetCurPosition( Extra_FileReader_t * p );

// extern void *       Extra_FileReaderGetTokens( Extra_FileReader_t * p );

// extern int          Extra_FileReaderGetLineNumber( Extra_FileReader_t * p, int iToken );

/*=== extraUtilMemory.c ========================================================*/

// fixed-size-block memory manager
// extern Extra_MmFixed_t *  Extra_MmFixedStart( int nEntrySize );

// extern void        Extra_MmFixedStop( Extra_MmFixed_t * p );

// extern char *      Extra_MmFixedEntryFetch( Extra_MmFixed_t * p );

// extern void        Extra_MmFixedEntryRecycle( Extra_MmFixed_t * p, char * pEntry );

// extern void        Extra_MmFixedRestart( Extra_MmFixed_t * p );

// extern int         Extra_MmFixedReadMemUsage( Extra_MmFixed_t * p );

// extern int         Extra_MmFixedReadMaxEntriesUsed( Extra_MmFixed_t * p );

// flexible-size-block memory manager
// extern Extra_MmFlex_t * Extra_MmFlexStart();

// extern void        Extra_MmFlexStop( Extra_MmFlex_t * p );

// extern void        Extra_MmFlexPrint( Extra_MmFlex_t * p );

// extern char *      Extra_MmFlexEntryFetch( Extra_MmFlex_t * p, int nBytes );

// extern int         Extra_MmFlexReadMemUsage( Extra_MmFlex_t * p );

// hierarchical memory manager
// extern Extra_MmStep_t * Extra_MmStepStart( int nSteps );

// extern void        Extra_MmStepStop( Extra_MmStep_t * p );

// extern char *      Extra_MmStepEntryFetch( Extra_MmStep_t * p, int nBytes );

// extern void        Extra_MmStepEntryRecycle( Extra_MmStep_t * p, char * pEntry, int nBytes );

// extern int         Extra_MmStepReadMemUsage( Extra_MmStep_t * p );

/*=== extraUtilMisc.c ========================================================*/

/* finds the smallest integer larger or equal than the logarithm */
// extern int         Extra_Base2LogDouble( double Num );

/* returns the power of two as a double */
// extern double      Extra_Power2( int Num );

// extern int         Extra_Power3( int Num );

/* the number of combinations of k elements out of n */
// extern int         Extra_NumCombinations( int k, int n  );

// extern int *       Extra_DeriveRadixCode( int Number, int Radix, int nDigits );

/* counts the number of 1s in the bitstring */
// extern int         Extra_CountOnes( unsigned char * pBytes, int nBytes );

/* the factorial of number */
// extern int         Extra_Factorial( int n );

/* the permutation of the given number of elements */
// extern char **     Extra_Permutations( int n );

/* permutation and complementation of a truth table */
// unsigned           Extra_TruthPermute( unsigned Truth, char * pPerms, int nVars, int fReverse );

// unsigned           Extra_TruthPolarize( unsigned uTruth, int Polarity, int nVars );

/* canonical forms of a truth table */
// extern unsigned    Extra_TruthCanonN( unsigned uTruth, int nVars );

// extern unsigned    Extra_TruthCanonNN( unsigned uTruth, int nVars );

// extern unsigned    Extra_TruthCanonP( unsigned uTruth, int nVars );

// extern unsigned    Extra_TruthCanonNP( unsigned uTruth, int nVars );

// extern unsigned    Extra_TruthCanonNPN( unsigned uTruth, int nVars );

/* canonical forms of 4-variable functions */
// extern void        Extra_Truth4VarNPN( unsigned short ** puCanons, char ** puPhases, char ** puPerms, unsigned char ** puMap );

// extern void        Extra_Truth4VarN( unsigned short ** puCanons, char *** puPhases, char ** ppCounters, int nPhasesMax );

/* permutation mapping */
// extern unsigned short Extra_TruthPerm4One( unsigned uTruth, int Phase );

// extern unsigned    Extra_TruthPerm5One( unsigned uTruth, int Phase );

// extern void        Extra_TruthPerm6One( unsigned * uTruth, int Phase, unsigned * uTruthRes );

// extern void        Extra_TruthExpand( int nVars, int nWords, unsigned * puTruth, unsigned uPhase, unsigned * puTruthR );

/* precomputing tables for permutation mapping */
// extern void **     Extra_ArrayAlloc( int nCols, int nRows, int Size );

// extern unsigned short ** Extra_TruthPerm43();

// extern unsigned ** Extra_TruthPerm53();

// extern unsigned ** Extra_TruthPerm54();

/* bubble sort for small number of entries */
// extern void        Extra_BubbleSort( int Order[], int Costs[], int nSize, int fIncreasing );

/* complementation/permutation generation */
// extern int *       Extra_GreyCodeSchedule( int n );

// extern int *       Extra_PermSchedule( int n );

// extern word        Extra_Truth6MinimumExact( word t, int * pComp, int * pPerm );

// extern word        Extra_Truth6MinimumHeuristic( word t );

/*=== extraUtilCanon.c ========================================================*/

/* fast computation of N-canoninical form up to 6 inputs */
// extern int         Extra_TruthCanonFastN( int nVarsMax, int nVarsReal, unsigned * pt, unsigned ** pptRes, char ** ppfRes );

/*=== extraUtilDsd.c ========================================================*/

// extern int         Sdm_ManCanRead();

// extern Sdm_Man_t * Sdm_ManRead();

// extern void        Sdm_ManQuit();

// extern int         Sdm_ManComputeFunc( Sdm_Man_t * p, int iDsdLit0, int iDsdLit1, int * pCut, int uMask, int fXor );

// extern void        Sdm_ManPrintDsdStats( Sdm_Man_t * p, int fVerbose );

// extern int         Sdm_ManReadDsdVarNum( Sdm_Man_t * p, int iDsd );

// extern int         Sdm_ManReadDsdAndNum( Sdm_Man_t * p, int iDsd );

// extern int         Sdm_ManReadDsdClauseNum( Sdm_Man_t * p, int iDsd );

// extern word        Sdm_ManReadDsdTruth( Sdm_Man_t * p, int iDsd );

// extern char *      Sdm_ManReadDsdStr( Sdm_Man_t * p, int iDsd );

// extern void        Sdm_ManReadCnfCosts( Sdm_Man_t * p, int * pCosts, int nCosts );

/*=== extraUtilProgress.c ================================================================*/

// ProgressBar and update implementation are defined in extraUtilProgress.go.

// extern ProgressBar * Extra_ProgressBarStart( FILE * pFile, int nItemsTotal );

// extern void        Extra_ProgressBarStop( ProgressBar * p );

// extern void        Extra_ProgressBarUpdate_int( ProgressBar * p, int nItemsCur, char * pString );

/*=== extraUtilTruth.c ================================================================*/

// Truth tables retain ABC's convention: at least one complete 32-bit word,
// with functions of fewer than five variables repeated across that word.
func Extra_BitWordNum(nBits int32) int32 {
	if nBits < 0 {
		panic("negative bit count")
	}
	return int32((int64(nBits) + 31) / 32)
}
func Extra_TruthWordNum(nVars int32) int32        { return int32(truthWords(nVars)) }
func Extra_TruthSetBit(table []uint32, bit int32) { table[bit>>5] |= 1 << uint(bit&31) }
func Extra_TruthXorBit(table []uint32, bit int32) { table[bit>>5] ^= 1 << uint(bit&31) }
func Extra_TruthHasBit(table []uint32, bit int32) int32 {
	if table[bit>>5]&(1<<uint(bit&31)) != 0 {
		return 1
	}
	return 0
}
func Extra_WordCountOnes(word uint32) int32 { return int32(bits.OnesCount32(word)) }
func Extra_TruthCountOnes(table []uint32, n int32) int32 {
	var count int32
	for _, word := range truthTable(table, n) {
		count += int32(bits.OnesCount32(word))
	}
	return count
}
func Extra_TruthIsEqual(a, b []uint32, n int32) int32 {
	if slices.Equal(truthTable(a, n), truthTable(b, n)) {
		return 1
	}
	return 0
}
func Extra_TruthIsConst0(table []uint32, n int32) int32 {
	for _, word := range truthTable(table, n) {
		if word != 0 {
			return 0
		}
	}
	return 1
}
func Extra_TruthIsConst1(table []uint32, n int32) int32 {
	for _, word := range truthTable(table, n) {
		if word != ^uint32(0) {
			return 0
		}
	}
	return 1
}
func Extra_TruthIsImply(a, b []uint32, n int32) int32 {
	a = truthTable(a, n)
	b = truthTable(b, n)
	for i, value := range a {
		if value&^b[i] != 0 {
			return 0
		}
	}
	return 1
}
func Extra_TruthCopy(out, in []uint32, n int32) { copy(truthTable(out, n), truthTable(in, n)) }
func Extra_TruthClear(out []uint32, n int32)    { clear(truthTable(out, n)) }
func Extra_TruthFill(out []uint32, n int32) {
	for i := range truthTable(out, n) {
		out[i] = ^uint32(0)
	}
}
func Extra_TruthNot(out, in []uint32, n int32) {
	in = slices.Clone(truthTable(in, n))
	out = truthTable(out, n)
	for i, value := range in {
		out[i] = ^value
	}
}
func truthBinaryWords(out, a, b []uint32, n int32, operation byte) {
	out = truthTable(out, n)
	a = slices.Clone(truthTable(a, n))
	b = slices.Clone(truthTable(b, n))
	for i, x := range a {
		y := b[i]
		switch operation {
		case '&':
			out[i] = x & y
		case '|':
			out[i] = x | y
		case '-':
			out[i] = x &^ y
		case '!':
			out[i] = ^(x & y)
		}
	}
}
func Extra_TruthAnd(out, a, b []uint32, n int32)   { truthBinaryWords(out, a, b, n, '&') }
func Extra_TruthOr(out, a, b []uint32, n int32)    { truthBinaryWords(out, a, b, n, '|') }
func Extra_TruthSharp(out, a, b []uint32, n int32) { truthBinaryWords(out, a, b, n, '-') }
func Extra_TruthNand(out, a, b []uint32, n int32)  { truthBinaryWords(out, a, b, n, '!') }
func Extra_TruthAndPhase(out, a, b []uint32, n, compl0, compl1 int32) {
	out = truthTable(out, n)
	a = slices.Clone(truthTable(a, n))
	b = slices.Clone(truthTable(b, n))
	var mask0, mask1 uint32
	if compl0 != 0 {
		mask0 = ^uint32(0)
	}
	if compl1 != 0 {
		mask1 = ^uint32(0)
	}
	for i, x := range a {
		out[i] = (x ^ mask0) & (b[i] ^ mask1)
	}
}

// extern unsigned ** Extra_TruthElementary( int nVars );

// extern void        Extra_TruthSwapAdjacentVars( unsigned * pOut, unsigned * pIn, int nVars, int Start );

// extern void        Extra_TruthStretch( unsigned * pOut, unsigned * pIn, int nVars, int nVarsAll, unsigned Phase );

// extern void        Extra_TruthShrink( unsigned * pOut, unsigned * pIn, int nVars, int nVarsAll, unsigned Phase );

// extern int         Extra_TruthVarInSupport( unsigned * pTruth, int nVars, int iVar );

// extern int         Extra_TruthSupportSize( unsigned * pTruth, int nVars );

// extern int         Extra_TruthSupport( unsigned * pTruth, int nVars );

// extern void        Extra_TruthCofactor0( unsigned * pTruth, int nVars, int iVar );

// extern void        Extra_TruthCofactor1( unsigned * pTruth, int nVars, int iVar );

// extern void        Extra_TruthExist( unsigned * pTruth, int nVars, int iVar );

// extern void        Extra_TruthForall( unsigned * pTruth, int nVars, int iVar );

// extern void        Extra_TruthMux( unsigned * pOut, unsigned * pCof0, unsigned * pCof1, int nVars, int iVar );

// extern void        Extra_TruthChangePhase( unsigned * pTruth, int nVars, int iVar );

// extern int         Extra_TruthMinCofSuppOverlap( unsigned * pTruth, int nVars, int * pVarMin );

// extern void        Extra_TruthCountOnesInCofs( unsigned * pTruth, int nVars, short * pStore );

// extern unsigned    Extra_TruthHash( unsigned * pIn, int nWords );

// extern unsigned    Extra_TruthSemiCanonicize( unsigned * pInOut, unsigned * pAux, int nVars, char * pCanonPerm, short * pStore );

/*=== extraUtilUtil.c ================================================================*/

// extern abctime       Extra_CpuTime();

// extern double        Extra_CpuTimeDouble();

// extern int           Extra_GetSoftDataLimit();

// Extra_UtilGetoptReset is defined in extraUtilUtil.go.
// extern int           Extra_UtilGetopt( int argc, char *argv[], const char *optstring );

// extern char *        Extra_UtilPrintTime( long t );

// extern char *        Extra_UtilStrsav( const char *s );

// extern char *        Extra_UtilTildeExpand( char *fname );

// extern char *        Extra_UtilFileSearch( char *file, char *path, char *mode );

// Extra_UtilMMoutOfMemory is defined in extraUtilUtil.go.

// Option globals are defined in extraUtilUtil.go.

/**AutomaticEnd***************************************************************/

//ABC_NAMESPACE_HEADER_END
