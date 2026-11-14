package cba

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [cbaWriteBlif.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Verilog parser.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: cbaWriteBlif.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cba.h"
//#include "cbaPrs.h"
//#include "map/mio/mio.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Writing parser state into a file.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Prs_ManWriteBlifArray(pFile *os.File, p *Prs_Ntk_t, vFanins *Vec_Int_t) {
	var i, NameId int32
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vFanins); i++ {
		NameId = Vec_IntEntry(vFanins, i)
		fmt.Fprintf(pFile, " %s", Prs_NtkStr(p, NameId))
	}
	fmt.Fprintf(pFile, "\n")
}
func Prs_ManWriteBlifLines(pFile *os.File, p *Prs_Ntk_t) {
	var vBox *Vec_Int_t
	var i, k, FormId, ActId int32
	// inlining of macro Prs_NtkForEachBox
	for i = 0; i < Prs_NtkBoxNum(p); i++ {
		vBox = Prs_BoxSignals(p, i)
		var NtkId int32 = Prs_BoxNtk(p, i)
		assert(Prs_BoxIONum(p, i) > 0, "Prs_BoxIONum(p, i) > 0")
		assert(Vec_IntSize(vBox)%2 == 0, "Vec_IntSize(vBox) % 2 == 0")
		if NtkId == -1 { // latch

			fmt.Fprintf(pFile, ".latch")
			fmt.Fprintf(pFile, " %s", Prs_NtkStr(p, Vec_IntEntry(vBox, 1)))
			fmt.Fprintf(pFile, " %s", Prs_NtkStr(p, Vec_IntEntry(vBox, 3)))
			fmt.Fprintf(pFile, " %c\n", '0'+Prs_BoxName(p, i))

		} else if Prs_BoxIsNode(p, i) != 0 { // node

			fmt.Fprintf(pFile, ".names")
			// inlining of macro Vec_IntForEachEntryDouble
			for k = 0; k+1 < Vec_IntSize(vBox); k += 2 {
				FormId = Vec_IntEntry(vBox, k)
				ActId = Vec_IntEntry(vBox, k+1)
				fmt.Fprintf(pFile, " %s", Prs_NtkStr(p, ActId))
			}
			fmt.Fprintf(pFile, "\n%s", Prs_NtkSop(p, NtkId))

		} else { // box

			fmt.Fprintf(pFile, ".subckt")
			fmt.Fprintf(pFile, " %s", Prs_NtkStr(p, NtkId))
			// inlining of macro Vec_IntForEachEntryDouble
			for k = 0; k+1 < Vec_IntSize(vBox); k += 2 {
				FormId = Vec_IntEntry(vBox, k)
				ActId = Vec_IntEntry(vBox, k+1)
				fmt.Fprintf(pFile, " %s=%s", Prs_NtkStr(p, FormId), Prs_NtkStr(p, ActId))
			}
			fmt.Fprintf(pFile, "\n")
		}
	}
}
func Prs_ManWriteBlifNtk(pFile *os.File, p *Prs_Ntk_t) {
	// write header
	fmt.Fprintf(pFile, ".model %s\n", Prs_NtkStr(p, p.iModuleName))
	if Vec_IntSize(&p.vInouts) != 0 {
		fmt.Fprintf(pFile, ".inouts")
	}
	if Vec_IntSize(&p.vInouts) != 0 {
		Prs_ManWriteBlifArray(pFile, p, &p.vInouts)
	}
	fmt.Fprintf(pFile, ".inputs")
	Prs_ManWriteBlifArray(pFile, p, &p.vInputs)
	fmt.Fprintf(pFile, ".outputs")
	Prs_ManWriteBlifArray(pFile, p, &p.vOutputs)
	// write objects
	Prs_ManWriteBlifLines(pFile, p)
	fmt.Fprintf(pFile, ".end\n\n")
}
func Prs_ManWriteBlif(pFileName string, vPrs *Vec_Ptr_t) {
	var pNtk *Prs_Ntk_t = Prs_ManRoot(vPrs)
	var pFile *os.File
	var i int32
	var err error
	pFile, err = os.Create(pFileName)
	if err != nil {
		fmt.Printf("Cannot open output file \"%s\".\n", pFileName)
		return
	}
	fmt.Fprintf(pFile, "# Design \"%s\" written by ABC on %s\n\n", Prs_NtkStr(pNtk, pNtk.iModuleName), Extra_TimeStamp())
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vPrs); i++ {
		pNtk, _ = Vec_PtrEntry(vPrs, i).(*Prs_Ntk_t)
		Prs_ManWriteBlifNtk(pFile, pNtk)
	}
	_ = (pFile).Close()
}

/*
*Function*************************************************************

	Synopsis    [Write elaborated design.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cba_ManWriteBlifLines(pFile *os.File, p *Cba_Ntk_t) {
	var k, iObj, iFon int32
	// inlining of macro Cba_NtkForEachBox
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Cba_ObjIsBox(p, iObj) == 0 {
		} else {
			if Cba_ObjIsBoxUser(p, iObj) != 0 {
				var pNtk *Cba_Ntk_t = Cba_ObjNtk(p, iObj)
				fmt.Fprintf(pFile, ".subckt")
				fmt.Fprintf(pFile, " %s", Cba_NtkName(pNtk))
				// inlining of macro Cba_ObjForEachFinFon
				for k = 0; k < Cba_ObjFinNum(p, iObj); k++ {
					iFon = Cba_ObjFinFon(p, iObj, k)
					fmt.Fprintf(pFile, " %s=%s", Cba_ObjNameStr(pNtk, Cba_NtkPi(pNtk, k)), Cba_FonNameStr(p, iFon))
				}
				// inlining of macro Cba_ObjForEachFon
				for k = 0; k < Cba_ObjFonNum(p, iObj); k++ {
					iFon = Cba_ObjFon(p, iObj, k)
					fmt.Fprintf(pFile, " %s=%s", Cba_ObjNameStr(pNtk, Cba_NtkPo(pNtk, k)), Cba_FonNameStr(p, iFon))
				}
				fmt.Fprintf(pFile, "\n")

			} else if Cba_ObjIsGate(p, iObj) != 0 {
				var pGateName string = Abc_NamStr(p.pDesign.pMods, Cba_ObjNtkId(p, iObj))
				var pLib *Mio_Library_t
				if lib, ok := Abc_FrameReadLibGen().(*Mio_Library_t); ok {
					pLib = lib
				}
				var pGate *Mio_Gate_t = Mio_LibraryReadGateByName(pLib, pGateName, "")
				fmt.Fprintf(pFile, ".gate %s", pGateName)
				// inlining of macro Cba_ObjForEachFinFon
				for k = 0; k < Cba_ObjFinNum(p, iObj); k++ {
					iFon = Cba_ObjFinFon(p, iObj, k)
					fmt.Fprintf(pFile, " %s=%s", Mio_GateReadPinName(pGate, k), Cba_FonNameStr(p, iFon))
				}
				// inlining of macro Cba_ObjForEachFon
				for k = 0; k < Cba_ObjFonNum(p, iObj); k++ {
					iFon = Cba_ObjFon(p, iObj, k)
					fmt.Fprintf(pFile, " %s=%s", Mio_GateReadOutName(pGate), Cba_FonNameStr(p, iFon))
				}
				fmt.Fprintf(pFile, "\n")

			} else {
				fmt.Fprintf(pFile, ".names")
				// inlining of macro Cba_ObjForEachFinFon
				for k = 0; k < Cba_ObjFinNum(p, iObj); k++ {
					iFon = Cba_ObjFinFon(p, iObj, k)
					fmt.Fprintf(pFile, " %s", Cba_FonNameStr(p, iFon))
				}
				fmt.Fprintf(pFile, " %s", Cba_FonNameStr(p, Cba_ObjFon0(p, iObj)))
				fmt.Fprintf(pFile, "\n%s", Cba_NtkSop(p, Cba_ObjFunc(p, iObj)))
			}
		}
	}
}
func Cba_ManWriteBlifArray(pFile *os.File, p *Cba_Ntk_t, vObjs *Vec_Int_t) {
	var iObj, i int32
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vObjs); i++ {
		iObj = Vec_IntEntry(vObjs, i)
		fmt.Fprintf(pFile, " %s", Cba_ObjNameStr(p, iObj))
	}
	fmt.Fprintf(pFile, "\n")
}
func Cba_ManWriteBlifNtk(pFile *os.File, p *Cba_Ntk_t) {
	// write header
	fmt.Fprintf(pFile, ".model %s\n", Cba_NtkName(p))
	fmt.Fprintf(pFile, ".inputs")
	Cba_ManWriteBlifArray(pFile, p, &p.vInputs)
	fmt.Fprintf(pFile, ".outputs")
	Cba_ManWriteBlifArray(pFile, p, &p.vOutputs)
	// write objects
	Cba_ManWriteBlifLines(pFile, p)
	fmt.Fprintf(pFile, ".end\n\n")
}
func Cba_ManWriteBlif(pFileName string, p *Cba_Man_t) {
	var pFile *os.File
	var pNtk *Cba_Ntk_t
	var i int32
	var err error
	// check the library
	if p.pMioLib != nil && p.pMioLib != Abc_FrameReadLibGen() {
		fmt.Printf("Genlib library used in the mapped design is not longer a current library.\n")
		return
	}
	pFile, err = os.OpenFile(pFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("Cannot open output file \"%s\".\n", pFileName)
		return
	}
	fmt.Fprintf(pFile, "# Design \"%s\" written via CBA package in ABC on %s\n\n", Cba_ManName(p), Extra_TimeStamp())
	//    Cba_ManAssignInternWordNames( p );
	// inlining of macro Cba_ManForEachNtk
	for i = 1; i <= Cba_ManNtkNum(p); i++ {
		pNtk = Cba_ManNtk(p, i)
		Cba_ManWriteBlifNtk(pFile, pNtk)
	}
	_ = (pFile).Close()
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
