package bsat

/**************************************************************************************************
MiniSat -- Copyright (c) 2005, Niklas Sorensson
http://www.cs.chalmers.se/Cs/Research/FormalMethods/MiniSat/

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

func satBoolNot(x int32) int32 {
	if x == 0 {
		return 1
	}
	return 0
}

func sat_solver_add_const(pSat *sat_solver, iVar int32, fCompl int32) int32 {
	var Lits [1]lit
	var Cid int32
	assert(iVar >= 0, "iVar >= 0")

	Lits[0] = toLitCond(iVar, fCompl)
	Cid = sat_solver_addclause(pSat, Lits[:1])
	assert((Cid != 0), "Cid")
	return 1
}
func sat_solver_add_buffer(pSat *sat_solver, iVarA int32, iVarB int32, fCompl int32) int32 {
	var Lits [2]lit
	var Cid int32
	assert(iVarA >= 0 && iVarB >= 0, "iVarA >= 0 && iVarB >= 0")

	Lits[0] = toLitCond(iVarA, 0)
	Lits[1] = toLitCond(iVarB, satBoolNot(fCompl))
	Cid = sat_solver_addclause(pSat, Lits[:2])
	if Cid == 0 {
		return 0
	}
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarA, 1)
	Lits[1] = toLitCond(iVarB, fCompl)
	Cid = sat_solver_addclause(pSat, Lits[:2])
	if Cid == 0 {
		return 0
	}
	assert((Cid != 0), "Cid")
	return 2
}
func sat_solver_add_buffer_enable(pSat *sat_solver, iVarA int32, iVarB int32, iVarEn int32, fCompl int32) int32 {
	var Lits [3]lit
	var Cid int32
	assert(iVarA >= 0 && iVarB >= 0 && iVarEn >= 0, "iVarA >= 0 && iVarB >= 0 && iVarEn >= 0")

	Lits[0] = toLitCond(iVarA, 0)
	Lits[1] = toLitCond(iVarB, satBoolNot(fCompl))
	Lits[2] = toLitCond(iVarEn, 1)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarA, 1)
	Lits[1] = toLitCond(iVarB, fCompl)
	Lits[2] = toLitCond(iVarEn, 1)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")
	return 2
}
func sat_solver_add_and(pSat *sat_solver, iVar int32, iVar0 int32, iVar1 int32, fCompl0 int32, fCompl1 int32, fCompl int32) int32 {
	var Lits [3]lit
	var Cid int32

	Lits[0] = toLitCond(iVar, satBoolNot(fCompl))
	Lits[1] = toLitCond(iVar0, fCompl0)
	Cid = sat_solver_addclause(pSat, Lits[:2])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVar, satBoolNot(fCompl))
	Lits[1] = toLitCond(iVar1, fCompl1)
	Cid = sat_solver_addclause(pSat, Lits[:2])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVar, fCompl)
	Lits[1] = toLitCond(iVar0, satBoolNot(fCompl0))
	Lits[2] = toLitCond(iVar1, satBoolNot(fCompl1))
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")
	return 3
}
func sat_solver_add_xor(pSat *sat_solver, iVarA int32, iVarB int32, iVarC int32, fCompl int32) int32 {
	var Lits [3]lit
	var Cid int32
	assert(iVarA >= 0 && iVarB >= 0 && iVarC >= 0, "iVarA >= 0 && iVarB >= 0 && iVarC >= 0")

	Lits[0] = toLitCond(iVarA, satBoolNot(fCompl))
	Lits[1] = toLitCond(iVarB, 1)
	Lits[2] = toLitCond(iVarC, 1)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarA, satBoolNot(fCompl))
	Lits[1] = toLitCond(iVarB, 0)
	Lits[2] = toLitCond(iVarC, 0)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarA, fCompl)
	Lits[1] = toLitCond(iVarB, 1)
	Lits[2] = toLitCond(iVarC, 0)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarA, fCompl)
	Lits[1] = toLitCond(iVarB, 0)
	Lits[2] = toLitCond(iVarC, 1)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")
	return 4
}
func sat_solver_add_mux(pSat *sat_solver, iVarZ int32, iVarC int32, iVarT int32, iVarE int32, iComplC int32, iComplT int32, iComplE int32, iComplZ int32) int32 {
	var Lits [3]lit
	var Cid int32
	assert(iVarC >= 0 && iVarT >= 0 && iVarE >= 0 && iVarZ >= 0, "iVarC >= 0 && iVarT >= 0 && iVarE >= 0 && iVarZ >= 0")

	Lits[0] = toLitCond(iVarC, 1^iComplC)
	Lits[1] = toLitCond(iVarT, 1^iComplT)
	Lits[2] = toLitCond(iVarZ, iComplZ)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarC, 1^iComplC)
	Lits[1] = toLitCond(iVarT, 0^iComplT)
	Lits[2] = toLitCond(iVarZ, 1^iComplZ)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarC, 0^iComplC)
	Lits[1] = toLitCond(iVarE, 1^iComplE)
	Lits[2] = toLitCond(iVarZ, 0^iComplZ)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarC, 0^iComplC)
	Lits[1] = toLitCond(iVarE, 0^iComplE)
	Lits[2] = toLitCond(iVarZ, 1^iComplZ)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")

	if iVarT == iVarE {
		return 4
	}

	Lits[0] = toLitCond(iVarT, 0^iComplT)
	Lits[1] = toLitCond(iVarE, 0^iComplE)
	Lits[2] = toLitCond(iVarZ, 1^iComplZ)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarT, 1^iComplT)
	Lits[1] = toLitCond(iVarE, 1^iComplE)
	Lits[2] = toLitCond(iVarZ, 0^iComplZ)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")
	return 6
}
func sat_solver_add_mux41(pSat *sat_solver, iVarZ int32, iVarC0 int32, iVarC1 int32, iVarD0 int32, iVarD1 int32, iVarD2 int32, iVarD3 int32) int32 {
	var Lits [4]lit
	var Cid int32
	assert(iVarC0 >= 0 && iVarC1 >= 0 && iVarD0 >= 0 && iVarD1 >= 0 && iVarD2 >= 0 && iVarD3 >= 0 && iVarZ >= 0, "iVarC0 >= 0 && iVarC1 >= 0 && iVarD0 >= 0 && iVarD1 >= 0 && iVarD2 >= 0 && iVarD3 >= 0 && iVarZ >= 0")

	Lits[0] = toLitCond(iVarD0, 1)
	Lits[1] = toLitCond(iVarC0, 0)
	Lits[2] = toLitCond(iVarC1, 0)
	Lits[3] = toLitCond(iVarZ, 0)
	Cid = sat_solver_addclause(pSat, Lits[:4])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarD1, 1)
	Lits[1] = toLitCond(iVarC0, 1)
	Lits[2] = toLitCond(iVarC1, 0)
	Lits[3] = toLitCond(iVarZ, 0)
	Cid = sat_solver_addclause(pSat, Lits[:4])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarD2, 1)
	Lits[1] = toLitCond(iVarC0, 0)
	Lits[2] = toLitCond(iVarC1, 1)
	Lits[3] = toLitCond(iVarZ, 0)
	Cid = sat_solver_addclause(pSat, Lits[:4])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarD3, 1)
	Lits[1] = toLitCond(iVarC0, 1)
	Lits[2] = toLitCond(iVarC1, 1)
	Lits[3] = toLitCond(iVarZ, 0)
	Cid = sat_solver_addclause(pSat, Lits[:4])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarD0, 0)
	Lits[1] = toLitCond(iVarC0, 0)
	Lits[2] = toLitCond(iVarC1, 0)
	Lits[3] = toLitCond(iVarZ, 1)
	Cid = sat_solver_addclause(pSat, Lits[:4])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarD1, 0)
	Lits[1] = toLitCond(iVarC0, 1)
	Lits[2] = toLitCond(iVarC1, 0)
	Lits[3] = toLitCond(iVarZ, 1)
	Cid = sat_solver_addclause(pSat, Lits[:4])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarD2, 0)
	Lits[1] = toLitCond(iVarC0, 0)
	Lits[2] = toLitCond(iVarC1, 1)
	Lits[3] = toLitCond(iVarZ, 1)
	Cid = sat_solver_addclause(pSat, Lits[:4])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarD3, 0)
	Lits[1] = toLitCond(iVarC0, 1)
	Lits[2] = toLitCond(iVarC1, 1)
	Lits[3] = toLitCond(iVarZ, 1)
	Cid = sat_solver_addclause(pSat, Lits[:4])
	assert((Cid != 0), "Cid")
	return 8
}
func sat_solver_add_xor_and(pSat *sat_solver, iVarF int32, iVarA int32, iVarB int32, iVarC int32) int32 {
	// F = (a (+) b) * c
	var Lits [4]lit
	var Cid int32
	assert(iVarF >= 0 && iVarA >= 0 && iVarB >= 0 && iVarC >= 0, "iVarF >= 0 && iVarA >= 0 && iVarB >= 0 && iVarC >= 0")

	Lits[0] = toLitCond(iVarF, 1)
	Lits[1] = toLitCond(iVarA, 1)
	Lits[2] = toLitCond(iVarB, 1)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarF, 1)
	Lits[1] = toLitCond(iVarA, 0)
	Lits[2] = toLitCond(iVarB, 0)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarF, 1)
	Lits[1] = toLitCond(iVarC, 0)
	Cid = sat_solver_addclause(pSat, Lits[:2])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarF, 0)
	Lits[1] = toLitCond(iVarA, 1)
	Lits[2] = toLitCond(iVarB, 0)
	Lits[3] = toLitCond(iVarC, 1)
	Cid = sat_solver_addclause(pSat, Lits[:4])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarF, 0)
	Lits[1] = toLitCond(iVarA, 0)
	Lits[2] = toLitCond(iVarB, 1)
	Lits[3] = toLitCond(iVarC, 1)
	Cid = sat_solver_addclause(pSat, Lits[:4])
	assert((Cid != 0), "Cid")
	return 5
}
func sat_solver_add_constraint(pSat *sat_solver, iVar int32, iVar2 int32, fCompl int32) int32 {
	var Lits [2]lit
	var Cid int32
	assert(iVar >= 0, "iVar >= 0")

	Lits[0] = toLitCond(iVar, fCompl)
	Lits[1] = toLitCond(iVar2, 0)
	Cid = sat_solver_addclause(pSat, Lits[:2])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVar, fCompl)
	Lits[1] = toLitCond(iVar2, 1)
	Cid = sat_solver_addclause(pSat, Lits[:2])
	assert((Cid != 0), "Cid")
	return 2
}

func sat_solver_add_half_sorter(pSat *sat_solver, iVarA int32, iVarB int32, iVar0 int32, iVar1 int32) int32 {
	var Lits [3]lit
	var Cid int32

	Lits[0] = toLitCond(iVarA, 0)
	Lits[1] = toLitCond(iVar0, 1)
	Cid = sat_solver_addclause(pSat, Lits[:2])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarA, 0)
	Lits[1] = toLitCond(iVar1, 1)
	Cid = sat_solver_addclause(pSat, Lits[:2])
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarB, 0)
	Lits[1] = toLitCond(iVar0, 1)
	Lits[2] = toLitCond(iVar1, 1)
	Cid = sat_solver_addclause(pSat, Lits[:3])
	assert((Cid != 0), "Cid")
	return 3
}

//ABC_NAMESPACE_HEADER_END
