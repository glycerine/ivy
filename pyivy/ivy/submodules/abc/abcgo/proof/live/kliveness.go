/**CFile****************************************************************

  FileName    [kliveness.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis        [Main implementation module of the algorithm k-Liveness    ]
              [invented by Koen Claessen, Niklas Sorensson. Implements]
              [the code for 'kcs'. 'kcs' pre-processes based on switch]
        [and then runs the (absorber circuit >> pdr) loop  ]

  Author      [Sayak Ray]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 31, 2012.]

***********************************************************************/

package live

import (
	"fmt"
	"strings"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	abc "github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/pdr"
)

// Definition of some macros pertaining to modes/switches
const (
	SIMPLE_kCS                                    = 0
	kCS_WITH_SAFETY_INVARIANTS                    = 1
	kCS_WITH_DISCOVER_MONOTONE_SIGNALS            = 2
	kCS_WITH_SAFETY_AND_DCS_INVARIANTS            = 3
	kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS = 4
)

func readLiveSignal_0(pAig *Aig_Man_t, liveIndex_0 int32) Aig_Obj_t {
	return driverOld(pAig, aig.Aig_ManCo(pAig, liveIndex_0))
}

func readLiveSignal_k(pAig *Aig_Man_t, liveIndex_k int32) Aig_Obj_t {
	return driverOld(pAig, aig.Aig_ManCo(pAig, liveIndex_k))
}

func introduceAbsorberLogic(pAig *Aig_Man_t, pLiveIndex_0, pLiveIndex_k *int32, nonFirstIteration int32) *Aig_Man_t {
	var pPOut Aig_Obj_t = aigNull
	var pPOutCo Aig_Obj_t = aigNull
	var pFirstAbsorberOr, pSecondAbsorberOr Aig_Obj_t
	var loCreated, loCopied int32

	//****************************************************************
	// Step1: create the new manager
	//****************************************************************
	pNewAig := aig.Aig_ManStart(aig.Aig_ManObjNumMax(pAig))
	pNewAig.Name = fmt.Sprintf("%s_%s", pAig.Name, "kCS")
	pNewAig.Spec = ""

	//****************************************************************
	// reading the signal pIn, and pOut
	//****************************************************************

	pPIn := readLiveSignal_0(pAig, *pLiveIndex_0)
	if *pLiveIndex_k == -1 {
		pPOut = aigNull
	} else {
		pPOut = readLiveSignal_k(pAig, *pLiveIndex_k)
	}

	//****************************************************************
	// Step 2: map constant nodes
	//****************************************************************
	aigSetData(pAig, aig.Aig_ManConst1(pAig), aig.Aig_ManConst1(pNewAig))

	//****************************************************************
	// Step 3: create true PIs
	//****************************************************************
	for i := int32(0); i < saigPiNum(pAig); i++ {
		aigSetData(pAig, aig.Aig_ManCi(pAig, i), aig.Aig_ObjCreateCi(pNewAig))
	}

	//****************************************************************
	// Step 5: create register outputs
	//****************************************************************
	for i := int32(0); i < saigRegNum(pAig); i++ {
		loCopied++
		aigSetData(pAig, saig.Saig_ManLo(pAig, i), aig.Aig_ObjCreateCi(pNewAig))
	}

	//****************************************************************
	// Step 6: create "D" register output for the ABSORBER logic
	//****************************************************************
	loCreated++
	pObjAbsorberLo := aig.Aig_ObjCreateCi(pNewAig)

	nRegCount := loCreated + loCopied

	//********************************************************************
	// Step 7: create internal nodes
	//********************************************************************
	nodes, _ := aigManNodes(pAig)
	for _, pObj := range nodes {
		aigSetData(pAig, pObj, aig.Aig_And(pNewAig, aig.Aig_ObjChild0Copy(pAig, pObj), aig.Aig_ObjChild1Copy(pAig, pObj)))
	}

	//****************************************************************
	// Step 8: create the two OR gates of the OBSERVER logic
	//****************************************************************
	if nonFirstIteration == 0 {
		pPInNewArg := copyOf(pAig, pPIn)

		pFirstAbsorberOr = aig.Aig_Or(pNewAig, aig.Aig_Not(pPInNewArg), pObjAbsorberLo)
		pSecondAbsorberOr = aig.Aig_Or(pNewAig, pPInNewArg, aig.Aig_Not(pObjAbsorberLo))
	} else {
		pPInNewArg := copyOf(pAig, pPIn)
		pPOutNewArg := copyOf(pAig, pPOut)

		pFirstAbsorberOr = aig.Aig_Or(pNewAig, aig.Aig_Not(pPOutNewArg), pObjAbsorberLo)
		pSecondAbsorberOr = aig.Aig_Or(pNewAig, pPInNewArg, aig.Aig_Not(pObjAbsorberLo))
	}

	//********************************************************************
	// Step 9: create primary outputs
	//********************************************************************
	var i int32
	for i = 0; i < saigPoNum(pAig); i++ {
		pObj := aig.Aig_ManCo(pAig, i)
		aigSetData(pAig, pObj, aig.Aig_ObjCreateCo(pNewAig, aig.Aig_ObjChild0Copy(pAig, pObj)))
		if i == *pLiveIndex_k {
			pPOutCo = aigData(pAig, pObj)
		}
	}

	//create new po
	if nonFirstIteration == 0 {
		pPOutCo = aig.Aig_ObjCreateCo(pNewAig, pSecondAbsorberOr)

		*pLiveIndex_k = i
	} else {
		aig.Aig_ObjPatchFanin0(pNewAig, pPOutCo, pSecondAbsorberOr)
	}

	for i := int32(0); i < saigRegNum(pAig); i++ {
		aig.Aig_ObjCreateCo(pNewAig, aig.Aig_ObjChild0Copy(pAig, saig.Saig_ManLi(pAig, i)))
	}

	//create new li
	aig.Aig_ObjCreateCo(pNewAig, pFirstAbsorberOr)

	aig.Aig_ManSetRegNum(pNewAig, nRegCount)
	aig.Aig_ManCleanup(pNewAig)

	return pNewAig
}

func modifyAigToApplySafetyInvar(pAig *Aig_Man_t, csTarget, safetyInvarPO int32) {
	pObjPOSafetyInvar := aig.Aig_ManCo(pAig, safetyInvarPO)
	pObjSafetyInvar := driverOld(pAig, pObjPOSafetyInvar)
	pObjPOCSTarget := aig.Aig_ManCo(pAig, csTarget)
	pObjCSTarget := driverOld(pAig, pObjPOCSTarget)

	pObjCSTargetNew := aig.Aig_And(pAig, pObjSafetyInvar, pObjCSTarget)
	aig.Aig_ObjPatchFanin0(pAig, pObjPOCSTarget, pObjCSTargetNew)
}

func flipConePdr(pAig *Aig_Man_t, directive int, targetCSPropertyIndex, safetyInvariantPOIndex, absorberCount int32) int32 {
	var RetValue int32
	var Pars pdr.Pdr_Par_t
	pPars := &Pars

	if directive == kCS_WITH_SAFETY_INVARIANTS || directive == kCS_WITH_SAFETY_AND_DCS_INVARIANTS || directive == kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS {
		modifyAigToApplySafetyInvar(pAig, targetCSPropertyIndex, safetyInvariantPOIndex)
	}

	pAigDupl := pAig
	pAig = aig.Aig_ManDupSimple(pAigDupl)

	for i := int32(0); i < saigPoNum(pAig); i++ {
		pObjTargetPo := aig.Aig_ManCo(pAig, i)
		aig.Aig_ObjChild0Flip(pAig, pObjTargetPo)
	}

	pdr.Pdr_ManSetDefaultParams(pPars)
	pPars.FVerbose = 1
	pPars.FNotVerbose = 1
	pPars.FSolveAll = 1
	pAig.SetSequentialCounterexamples(nil)

	aig.Aig_ManCleanup(pAig)

	pdr.Pdr_ManSolve(pAig, pPars)

	if vSeqModelVec := pAig.SequentialCounterexamples(); vSeqModelVec != nil {
		if cexIsNull(vec.Vec_PtrEntry(vSeqModelVec, int(targetCSPropertyIndex))) {
			RetValue = 1
		} else {
			RetValue = 0
		}
	} else {
		RetValue = -1
		exit0()
	}

	for i := int32(0); i < saigPoNum(pAig); i++ {
		pObjTargetPo := aig.Aig_ManCo(pAig, i)
		aig.Aig_ObjChild0Flip(pAig, pObjTargetPo)
	}

	aig.Aig_ManStop(pAig)
	return RetValue
}

func collectSafetyInvariantPOIndex(pNtk *Abc_Ntk_t) int32 {
	for i := int32(0); i < abc.Abc_NtkPoNum(pNtk); i++ {
		if strings.Contains(abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, i)), "csSafetyInvar_") {
			return i
		}
	}
	return -1
}

func collectUserGivenDisjunctiveMonotoneSignals(pNtk *Abc_Ntk_t) [][]int32 {
	monotoneVector := make([][]int32, 0)
	for i := int32(0); i < abc.Abc_NtkPoNum(pNtk); i++ {
		if strings.Contains(abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, i)), "csLevel1Stabil_") {
			monotoneVector = append(monotoneVector, createSingletonIntVector(i))
		}
	}

	if len(monotoneVector) > 0 {
		return monotoneVector
	}
	return nil
}

func deallocateMasterBarrierDisjunctInt(vMasterBarrierDisjunctsArg [][]int32) {}

// getVecOfVecFairness reads nothing useful in C either (an unfinished
// reader of a fairness file); it is not called.
func getVecOfVecFairness() [][]int32 {
	return make([][]int32, 0)
}

func Abc_CommandCS_kLiveness(pAbc *Abc_Frame_t, argc int32, argv []string) (status int32) {
	var pAig, pAigCS, pAigCSNew *Aig_Man_t
	var pNtkTemp *Abc_Ntk_t
	var absorberCount int32
	absorberLimit := int32(500)
	var RetValue int32
	liveIndex_0, liveIndex_k := int32(-1), int32(-1)
	fVerbose := 1
	directive := -1
	safetyInvariantPO := int32(-1)
	var vMasterBarrierDisjuncts [][]int32
	var pWorkingAig *Aig_Man_t
	defer recoverExit(&status)

	pNtk := frameNtk(pAbc)

	/*************************************************
	Extraction of Command Line Argument
	*************************************************/
	if argc == 1 {
		directive = SIMPLE_kCS
	} else {
		extra.Extra_UtilGetoptReset()
		for {
			c := extra.Extra_UtilGetopt(argc, argv, "cmCgh")
			if c == -1 {
				break
			}
			switch c {
			case 'c':
				directive = kCS_WITH_SAFETY_INVARIANTS
			case 'm':
				directive = kCS_WITH_DISCOVER_MONOTONE_SIGNALS
			case 'C':
				directive = kCS_WITH_SAFETY_AND_DCS_INVARIANTS
			case 'g':
				directive = kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS
			default: // 'h' and unknown options
				goto usage
			}
		}
	}
	/*************************************************
	Extraction of Command Line Argument Ends
	*************************************************/

	// C dereferenced the missing network: CBUGS_FIXED.md #52
	if pNtk == nil {
		util.Abc_Print(-1, "Empty network.\n")
		return 1
	}

	if !(abc.Abc_NtkIsStrash(pNtk) != 0) {
		fmt.Printf("The input network was not strashed, strashing....\n")
		pNtkTemp = Abc_NtkStrash(pNtk, 0, 0, 0)
		pAig = Abc_NtkToDar(pNtkTemp, 0, 1)
	} else {
		pAig = Abc_NtkToDar(pNtk, 0, 1)
		pNtkTemp = pNtk
	}

	// collectCSSignals needs a "csLiveTarget_" output (it asserts one;
	// C crashed without it: CBUGS_FIXED.md #47)
	{
		fTarget := false
		for i := int32(0); i < abc.Abc_NtkPoNum(pNtk); i++ {
			if strings.Contains(abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, i)), "csLiveTarget_") {
				fTarget = true
			}
		}
		if !fTarget {
			fmt.Printf("The network has no liveness target output (named \"csLiveTarget_...\").\n")
			aig.Aig_ManStop(pAig)
			return 1
		}
	}

	// -c, -C and -g need a "csSafetyInvar_" output (flipConePdr reads it;
	// C crashed without it: CBUGS_FIXED.md #56)
	if (directive == kCS_WITH_SAFETY_INVARIANTS || directive == kCS_WITH_SAFETY_AND_DCS_INVARIANTS || directive == kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS) && collectSafetyInvariantPOIndex(pNtkTemp) == -1 {
		fmt.Printf("The network has no safety invariant output (named \"csSafetyInvar_...\").\n")
		aig.Aig_ManStop(pAig)
		return 1
	}
	// -C and -g need the window and pending outputs (arenaViolation.c reads
	// them; CBUGS_FIXED.md #56)
	if (directive == kCS_WITH_SAFETY_AND_DCS_INVARIANTS || directive == kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS) &&
		(collectWindowBeginSignalWithDSC(pNtk, pAig) == -1 || collectWithinWindowSignalWithDSC(pNtk, pAig) == -1 || collectPendingSignalWithDSC(pNtk, pAig) == -1) {
		fmt.Printf("The network has no window or pending outputs (named \"windowBegins_...\", \"withinWindow_...\" and \"pendingSignal...\").\n")
		aig.Aig_ManStop(pAig)
		return 1
	}

	if directive == kCS_WITH_SAFETY_INVARIANTS {
		safetyInvariantPO = collectSafetyInvariantPOIndex(pNtkTemp)
	}

	if directive == kCS_WITH_DISCOVER_MONOTONE_SIGNALS {
		beginTime := util.Abc_Clock()
		vMasterBarrierDisjuncts = findDisjunctiveMonotoneSignals(pNtk)
		endTime := util.Abc_Clock()
		time_spent := float64(endTime-beginTime) / 1e9
		fmt.Printf("pre-processing time = %f\n", time_spent)
		return 0
	}

	if directive == kCS_WITH_SAFETY_AND_DCS_INVARIANTS {
		safetyInvariantPO = collectSafetyInvariantPOIndex(pNtkTemp)

		beginTime := util.Abc_Clock()
		vMasterBarrierDisjuncts = findDisjunctiveMonotoneSignals(pNtk)
		endTime := util.Abc_Clock()
		time_spent := float64(endTime-beginTime) / 1e9
		fmt.Printf("pre-processing time = %f\n", time_spent)
	}

	if directive == kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS {
		safetyInvariantPO = collectSafetyInvariantPOIndex(pNtkTemp)

		beginTime := util.Abc_Clock()
		vMasterBarrierDisjuncts = collectUserGivenDisjunctiveMonotoneSignals(pNtk)
		endTime := util.Abc_Clock()
		time_spent := float64(endTime-beginTime) / 1e9
		fmt.Printf("pre-processing time = %f\n", time_spent)
	}

	// -C and -g need at least one barrier (createArenaLO returns NULL
	// without one; CBUGS_FIXED.md #56)
	if (directive == kCS_WITH_SAFETY_AND_DCS_INVARIANTS || directive == kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS) &&
		len(vMasterBarrierDisjuncts) == 0 {
		fmt.Printf("No barriers (monotone disjunctions of the hint_ or csLevel1Stabil_ outputs) were found.\n")
		aig.Aig_ManStop(pAig)
		return 1
	}

	if directive == kCS_WITH_SAFETY_AND_DCS_INVARIANTS || directive == kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS {
		pWorkingAig = generateWorkingAigWithDSC(pAig, pNtk, &liveIndex_0, vMasterBarrierDisjuncts)
		pAigCS = introduceAbsorberLogic(pWorkingAig, &liveIndex_0, &liveIndex_k, 0)
	} else {
		pWorkingAig = generateWorkingAig(pAig, pNtk, &liveIndex_0)
		pAigCS = introduceAbsorberLogic(pWorkingAig, &liveIndex_0, &liveIndex_k, 0)
	}

	aig.Aig_ManStop(pWorkingAig)

	for absorberCount = 1; absorberCount < absorberLimit; absorberCount++ {
		RetValue = flipConePdr(pAigCS, directive, liveIndex_k, safetyInvariantPO, absorberCount)

		if RetValue == 1 {
			util.Abc_Print(1, "k = %d, Property proved\n", absorberCount)
			break
		} else if RetValue == 0 {
			if fVerbose != 0 {
				util.Abc_Print(1, "k = %d, Property DISPROVED\n", absorberCount)
			}
		} else if RetValue == -1 {
			util.Abc_Print(1, "Property UNDECIDED with k = %d.\n", absorberCount)
		}

		pAigCSNew = introduceAbsorberLogic(pAigCS, &liveIndex_0, &liveIndex_k, absorberCount)
		aig.Aig_ManStop(pAigCS)
		pAigCS = pAigCSNew
	}

	aig.Aig_ManStop(pAigCS)
	aig.Aig_ManStop(pAig)

	deallocateMasterBarrierDisjunctInt(vMasterBarrierDisjuncts)
	return 0

usage:
	fmt.Printf("usage: kcs [-cmgCh]\n")
	fmt.Printf("\timplements Claessen-Sorensson's k-Liveness algorithm\n")
	fmt.Printf("\t-c : verification with constraints, looks for POs prefixed with csSafetyInvar_\n")
	fmt.Printf("\t-m : discovers monotone signals\n")
	fmt.Printf("\t-g : verification with user-supplied barriers, looks for POs prefixed with csLevel1Stabil_\n")
	fmt.Printf("\t-C : verification with discovered monotone signals\n")
	fmt.Printf("\t-h : print usage\n")
	return 1
}

func Abc_CommandNChooseK(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pAig *Aig_Man_t

	pNtk := frameNtk(pAbc)

	/*************************************************
	Extraction of Command Line Argument
	*************************************************/
	if argc != 1 {
		extra.Extra_UtilGetoptReset()
		for {
			c := extra.Extra_UtilGetopt(argc, argv, "cmCgh")
			if c == -1 {
				break
			}
			switch c {
			case 'c', 'm', 'C', 'g':
			default: // 'h' and unknown options
				goto usage
			}
		}
	}
	/*************************************************
	Extraction of Command Line Argument Ends
	*************************************************/

	// C dereferenced the missing network: CBUGS_FIXED.md #52
	if pNtk == nil {
		util.Abc_Print(-1, "Empty network.\n")
		return 1
	}

	if !(abc.Abc_NtkIsStrash(pNtk) != 0) {
		fmt.Printf("The input network was not strashed, strashing....\n")
		pNtkTemp := Abc_NtkStrash(pNtk, 0, 0, 0)
		pAig = Abc_NtkToDar(pNtkTemp, 0, 1)
	} else {
		pAig = Abc_NtkToDar(pNtk, 0, 1)
	}

	/**********************Code for generation of nCk outputs**/
	{
		fmt.Printf("Enter parameterizedCombK = ")
		parameterizedCombK, ok := scanfInt(pAbc)
		if !ok {
			fmt.Printf("\nFailed to read integer!\n")
			return 0
		}
		fmt.Printf("\n")

		// at most all registers can be chosen (C's countCombination did not
		// return and Aig_ManLo read past the registers: CBUGS_FIXED.md #57)
		if parameterizedCombK > aig.Aig_ManRegNum(pAig) {
			fmt.Printf("Cannot choose %d out of %d registers.\n", parameterizedCombK, aig.Aig_ManRegNum(pAig))
			aig.Aig_ManStop(pAig)
			return 1
		}

		pAigCombStabil := generateGeneralDisjunctiveTester(pNtk, pAig, parameterizedCombK)
		aig.Aig_ManPrintStats(pAigCombStabil)
		pNtkCombStabil := Abc_NtkFromAigPhase(pAigCombStabil)
		pNtkCombStabil.Name = pAigCombStabil.Name
		if abc.Abc_NtkCheck(pNtkCombStabil) == 0 {
			fmt.Printf("Abc_NtkCreateCone(): Network check has failed.\n")
		}
		abcmain.Abc_FrameSetCurrentNetwork(pAbc, pNtkCombStabil)

		aig.Aig_ManStop(pAigCombStabil)
		aig.Aig_ManStop(pAig)

		return 1
	}

usage:
	fmt.Printf("usage: nck [-cmgCh]\n")
	fmt.Printf("\tgenerates combinatorial signals for stabilization\n")
	fmt.Printf("\t-h : print usage\n")
	return 1
}
