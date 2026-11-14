package aig

import "fmt"

//#include "aig.h"

//ABC_NAMESPACE_IMPL_START

func Aig_ProcedureTest() {
	var p *Aig_Man_t
	var pA, pB, pC Edge
	var pFunc Edge
	var pFunc2 Edge

	p = Aig_ManStart(1000)

	pA = Aig_IthVar(p, 0)
	pB = Aig_IthVar(p, 1)
	pC = Aig_IthVar(p, 2)

	pFunc = Aig_Mux(p, pA, pB, pC)
	pFunc2 = Aig_And(p, pA, pB)

	Aig_ObjCreateCo(p, pFunc)
	Aig_ObjCreateCo(p, pFunc2)

	Aig_ManSetRegNum(p, 1)

	Aig_ManCleanup(p)

	if Aig_ManCheck(p) == 0 {
		fmt.Printf("Check has failed\n")
	}

	if err := Aig_ManDumpBlif(p, "aig_test_file.blif", nil, nil); err != nil {
		fmt.Println(err)
	}
	Aig_ManStop(p)
}

//ABC_NAMESPACE_IMPL_END
