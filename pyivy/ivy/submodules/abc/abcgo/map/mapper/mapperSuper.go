package mapper

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [mapperSuper.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperSuper.c,v 1.6 2005/01/23 06:59:44 alanmi Exp $]

***********************************************************************/

//#include "mapperInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int           Map_LibraryReadFile( Map_SuperLib_t * pLib, FILE * pFile );

// static Map_Super_t * Map_LibraryReadGate( Map_SuperLib_t * pLib, char * pBuffer, int nVars );

// static int           Map_LibraryTruthVerify( Map_SuperLib_t * pLib, Map_Super_t * pGate );

// static void          Map_LibraryComputeTruth( Map_SuperLib_t * pLib, char * pFormula, unsigned uTruthRes[] );

// static void          Map_LibraryComputeTruth_rec( Map_SuperLib_t * pLib, char * pFormula, unsigned uTruthsIn[][2], unsigned uTruthRes[] );

// static void          Map_LibraryPrintClasses( Map_SuperLib_t * p );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Reads the supergate library from file.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryRead(pLib *Map_SuperLib_t, pFileName string) int32 {
	var pFile *os.File
	var Status int32
	// read the beginning of the file
	assert(pLib.pGenlib == nil, "pLib.pGenlib == nil")
	pFile = os.Open(pFileName)
	if pFile == nil {
		fmt.Printf("Cannot open input file \"%s\".\n", pFileName)
		return 0
	}
	Status = Map_LibraryReadFile(pLib, pFile)
	_ = (pFile).Close()
	//    Map_LibraryPrintClasses( pLib );
	return Status
}

/*
*Function*************************************************************

	Synopsis    [Reads the library file.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryReadFile(pLib *Map_SuperLib_t, pFile *os.File) int32 {
	var pProgress *ProgressBar
	var pBuffer [5000]byte
	var pFileGen *os.File
	var pGate *Map_Super_t
	var pTemp string = nil // Suppress "might be used uninitialized"
	var pLibName string
	var nCounter, nGatesTotal int32
	var uCanon [2]uint32
	var RetValue int32

	// skip empty and comment lines
	for fgets(pBuffer, 2000, pFile) != nil {
		// skip leading spaces
		for pTemp = pBuffer; *pTemp == ' ' || *pTemp == '\r' || *pTemp == '\n'; pTemp++ {
		}
		// skip comment lines and empty lines
		if *pTemp != 0 && *pTemp != '#' {
			break
		}
	}

	// get the genlib file name
	pLibName = strtok(pTemp, " \t\r\n")
	if int32(strings.Compare(pLibName, "GATE")) == 0 {
		fmt.Printf("The input file \"%s\" looks like a genlib file and not a supergate library file.\n", pLib.pName)
		return 0
	}
	pFileGen = os.Open(pLibName)
	if pFileGen == nil {
		fmt.Printf("Cannot open the genlib file \"%s\".\n", pLibName)
		return 0
	}
	_ = (pFileGen).Close()

	// read the genlib library
	pLib.pGenlib = Mio_LibraryRead(pLibName, nil, 0, 0)
	if pLib.pGenlib == nil {
		fmt.Printf("Cannot read genlib file \"%s\".\n", pLibName)
		return 0
	}

	// read the number of variables
	RetValue = fscanf(pFile, "%d\n", &pLib.nVarsMax)
	if pLib.nVarsMax < 2 || pLib.nVarsMax > 10 {
		fmt.Printf("Suspicious number of variables (%d).\n", pLib.nVarsMax)
		return 0
	}

	// read the number of gates
	RetValue = fscanf(pFile, "%d\n", &nGatesTotal)
	if nGatesTotal < 1 || nGatesTotal > 10000000 {
		fmt.Printf("Suspicious number of gates (%d).\n", nGatesTotal)
		return 0
	}

	// read the lines
	nCounter = 0
	pProgress = Extra_ProgressBarStart(os.Stdout, nGatesTotal)
	for fgets(pBuffer, 5000, pFile) != nil {
		for pTemp = pBuffer; *pTemp == ' ' || *pTemp == '\r' || *pTemp == '\n'; pTemp++ {
		}
		if pTemp[0] == '\x00' {
			continue
		}
		// get the gate
		pGate = Map_LibraryReadGate(pLib, pTemp, pLib.nVarsMax)
		assert(pGate.Num == nCounter+1, "pGate.Num == nCounter + 1")
		// count the number of parentheses in the formula - this is the number of gates
		for pTemp = pGate.pFormula; *pTemp; pTemp++ {
			pGate.nGates += (*pTemp == '(')
		}
		// verify the truth table
		assert((Map_LibraryTruthVerify(pLib, pGate) != 0), "Map_LibraryTruthVerify(pLib, pGate)")

		// find the N-canonical form of this supergate
		pGate.nPhases = Map_CanonComputeSlow(pLib.uTruths, pLib.nVarsMax, pLib.nVarsMax, pGate.uTruth, pGate.uPhases, uCanon)
		// add the supergate into the table by its N-canonical table
		Map_SuperTableInsertC(pLib.tTableC, uCanon, pGate)
		// update the progress bar
		nCounter++
		Extra_ProgressBarUpdate(pProgress, nCounter, nil)
	}
	Extra_ProgressBarStop(pProgress)
	pLib.nSupersAll = nCounter
	if nCounter != nGatesTotal {
		fmt.Printf("The number of gates read (%d) is different what the file says (%d).\n", nGatesTotal, nCounter)
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryReadGate(pLib *Map_SuperLib_t, pBuffer string, nVars int32) *Map_Super_t {
	var pGate *Map_Super_t
	var pTemp string
	var i int32

	// start and clean the gate
	pGate = &Map_SuperStruct_t_{}

	// read the number
	pTemp = strtok(pBuffer, " ")
	pGate.Num = strconv.Atoi(pTemp)

	// read the signature
	pTemp = strtok(nil, " ")
	if pLib.nVarsMax < 6 {
		pGate.uTruth[0] = Extra_ReadBinary(pTemp)
		pGate.uTruth[1] = 0

	} else {
		pGate.uTruth[0] = Extra_ReadBinary(pTemp + 32)
		pTemp[32] = 0
		pGate.uTruth[1] = Extra_ReadBinary(pTemp)
	}

	// read the max delay
	pTemp = strtok(nil, " ")
	pGate.tDelayMax.Rise = float32(atof(pTemp))
	pGate.tDelayMax.Fall = pGate.tDelayMax.Rise

	// read the pin-to-pin delay
	for i = 0; i < nVars; i++ {
		pTemp = strtok(nil, " ")
		pGate.tDelaysR[i].Rise = float32(atof(pTemp))
		pGate.tDelaysF[i].Fall = pGate.tDelaysR[i].Rise
	}

	// read the area
	pTemp = strtok(nil, " ")
	pGate.Area = float32(atof(pTemp))

	// the rest is the gate name
	pTemp = strtok(nil, " \r\n")
	if len(pTemp) == 0 {
		fmt.Printf("A gate name is empty.\n")
	}

	// save the gate name
	pGate.pFormula = Extra_MmFlexEntryFetch(pLib.mmForms, len(pTemp)+1)
	strcpy(pGate.pFormula, pTemp)

	// the rest is the gate name
	pTemp = strtok(nil, " \n\x00")
	if pTemp != nil {
		fmt.Printf("The following trailing symbols found \"%s\".\n", pTemp)
	}
	return pGate
}

/*
*Function*************************************************************

	Synopsis    [Performs one step of parsing the formula into parts.]

	Description [This function will eventually be replaced when the
	tree-supergate library representation will become standard.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryReadFormulaStep(pFormula string, pStrings []string, pnStrings *int32) string {
	var pName, pPar1, pPar2, pCur string
	var nStrings, CountPars int32

	// skip leading spaces
	for pName = pFormula; *pName && *pName == ' '; pName++ {
	}
	assert(*pName, "*pName")
	// find the first opening parenthesis
	for pPar1 = pName; *pPar1 && *pPar1 != '('; pPar1++ {
	}
	if *pPar1 == 0 {
		*pnStrings = 0
		return pName
	}
	// overwrite it with space
	assert(*pPar1 == '(', "*pPar1 == '('")
	*pPar1 = 0
	// find the corresponding closing parenthesis
	for CountPars, pPar2 = 1, pPar1+1; *pPar2 && (CountPars != 0); pPar2++ {
		if *pPar2 == '(' {
			CountPars++
		} else if *pPar2 == ')' {
			CountPars--
		}
	}
	pPar2--
	assert(CountPars == 0, "CountPars == 0")
	// overwrite it with space
	assert(*pPar2 == ')', "*pPar2 == ')'")
	*pPar2 = 0
	// save the intervals between the commas
	nStrings = 0
	pCur = pPar1 + 1
	for {
		// save the current string
		pStrings[nStrings] = pCur
		nStrings++
		// find the beginning of the next string
		for CountPars = 0; *pCur && ((CountPars != 0) || *pCur != ','); pCur++ {
			if *pCur == '(' {
				CountPars++
			} else if *pCur == ')' {
				CountPars--
			}
		}
		if *pCur == 0 {
			break
		}
		assert(*pCur == ',', "*pCur == ','")
		*pCur = 0
		pCur++
	}
	// save the results and return
	*pnStrings = nStrings
	return pName
}

/*
*Function*************************************************************

	Synopsis    [Verifies the truth table of the supergate.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryTruthVerify(pLib *Map_SuperLib_t, pGate *Map_Super_t) int32 {
	var uTruthRes [2]uint32
	Map_LibraryComputeTruth(pLib, pGate.pFormula, uTruthRes)
	if uTruthRes[0] != pGate.uTruth[0] || uTruthRes[1] != pGate.uTruth[1] {
		return 0
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Derives the functionality of the supergate.]

	Description [This procedure is useful for verification the supergate
	library. The truth table derived by this procedure should be the same
	as the one contained in the original supergate file.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryComputeTruth(pLib *Map_SuperLib_t, pFormula string, uTruthRes []uint32) {
	var Buffer [1000]byte
	strcpy(Buffer, pFormula)
	Map_LibraryComputeTruth_rec(pLib, Buffer, pLib.uTruths, uTruthRes)
}

/*
*Function*************************************************************

	Synopsis    [Derives the functionality of the supergate.]

	Description [This procedure is useful for verification the supergate
	library. The truth table derived by this procedure should be the same
	as the one contained in the original supergate file.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryComputeTruth_rec(pLib *Map_SuperLib_t, pFormula string, uTruthsIn [][2]uint32, uTruthRes []uint32) {
	var pMioGate *Mio_Gate_t
	var pGateName string
	var pStrings [6]string
	var uTruthsFanins [6][2]uint32
	var nStrings, i int32

	// perform one step parsing of the formula
	// detect the root gate name, the next-step strings, and their number
	pGateName = Map_LibraryReadFormulaStep(pFormula, pStrings, &nStrings)
	if nStrings == 0 { // elementary variable

		assert(pGateName[0]-'a' < pLib.nVarsMax, "pGateName[0] - 'a' < pLib.nVarsMax")
		uTruthRes[0] = uTruthsIn[pGateName[0]-'a'][0]
		uTruthRes[1] = uTruthsIn[pGateName[0]-'a'][1]
		return
	}
	// derive the functionality of the fanins
	for i = 0; i < nStrings; i++ {
		Map_LibraryComputeTruth_rec(pLib, pStrings[i], uTruthsIn, uTruthsFanins[i])
	}
	// get the root supergate
	pMioGate = Mio_LibraryReadGateByName(pLib.pGenlib, pGateName, nil)
	if pMioGate == nil {
		fmt.Printf("A supergate contains gate \"%s\" that is not in \"%s\".\n", pGateName, Mio_LibraryReadName(pLib.pGenlib))
	}
	// derive the functionality of the output of the supergate
	Mio_DeriveTruthTable(pMioGate, uTruthsFanins, nStrings, pLib.nVarsMax, uTruthRes)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryPrintSupergate(pGate *Map_Super_t) {
	fmt.Printf("%5d : ", pGate.nUsed)
	fmt.Printf("%5d   ", pGate.Num)
	fmt.Printf("A = %5.2f   ", pGate.Area)
	fmt.Printf("D = %5.2f/%5.2f/%5.2f   ", pGate.tDelayMax.Rise, pGate.tDelayMax.Fall, pGate.tDelayMax.Worst)
	fmt.Printf("%s", pGate.pFormula)
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    [Prints N-classes of supergates.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryPrintClasses(p *Map_SuperLib_t) {
	/*
	   st__generator * gen;
	   Map_Super_t * pSuper, * pSuper2;
	   unsigned Key, uTruth;
	   int Counter = 0;
	   // copy all the supergates into one array
	   st__foreach_item( p->tSuplib, gen, (char **)&Key, (char **)&pSuper )

	   	{
	   	    for ( pSuper2 = pSuper; pSuper2; pSuper2 = pSuper2->pNext )
	   	    {
	   	        uTruth = pSuper2->Phase;
	   	        Extra_PrintBinary( stdout, &uTruth, 5 );
	   	        printf( "  %5d   ",  pSuper2->Num );
	   	        printf( "%s",    pSuper2->pFormula );
	   	        printf( "\n" );
	   	    }
	   	    printf( "\n" );
	   	    if ( ++ Counter == 100 )
	   	        break;
	   	}
	*/
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
