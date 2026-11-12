// Port of ABC src/aig/miniaig/ndr.h (Alan Mishchenko, UC Berkeley).
package miniaig

func Ndr_ModuleTest() {
	NameIdA := 2
	_ = NameIdA
	NameIdS := 3
	_ = NameIdS
	NameIdC := 4
	_ = NameIdC
	Fanins := [2]int{NameIdA, NameIdC}
	_ = Fanins
	ppNames := []string{"", "add10", "a", "s", "const10"}
	_ = ppNames
	pDesign := Ndr_Create(1)
	_ = pDesign
	ModuleID := Ndr_AddModule(pDesign, 1)
	_ = ModuleID
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 3, 0, 0, 0, nil, 1, []int{NameIdA}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CONST, 0, 3, 0, 0, 0, nil, 1, []int{NameIdC}, "4'b1010")
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_ARI_ADD, 0, 3, 0, 0, 2, Fanins[:], 1, []int{NameIdS}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CO, 0, 3, 0, 0, 1, []int{NameIdS}, 0, nil, nil)
	Ndr_WriteVerilog("", pDesign, ppNames)
	Ndr_Write("add4.ndr", pDesign)
	Ndr_Delete(pDesign)
}
func Ndr_ModuleTestAdder() {
	ppNames := []string{"",
		"a", "b", "s", "co",
		"a0", "a1", "b0", "b1",
		"r0", "s0", "rco",
		"r1", "s1", "add8"}
	_ = ppNames
	FaninA := 1
	_ = FaninA
	FaninB := 2
	_ = FaninB
	FaninS := 3
	_ = FaninS
	FaninCO := 4
	_ = FaninCO
	FaninA0 := 5
	_ = FaninA0
	FaninA1 := 6
	_ = FaninA1
	FaninB0 := 7
	_ = FaninB0
	FaninB1 := 8
	_ = FaninB1
	FaninR0 := 9
	_ = FaninR0
	FaninS0 := 10
	_ = FaninS0
	FaninRCO := 11
	_ = FaninRCO
	FaninR1 := 12
	_ = FaninR1
	FaninS1 := 13
	_ = FaninS1
	Fanins1 := [2]int{FaninA0, FaninB0}
	_ = Fanins1
	Fanins2 := [3]int{FaninA1, FaninB1, FaninRCO}
	_ = Fanins2
	Fanins3 := [4]int{FaninS1, FaninS0}
	_ = Fanins3
	pDesign := Ndr_Create(14)
	_ = pDesign
	ModuleID := Ndr_AddModule(pDesign, 14)
	_ = ModuleID
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 7, 0, 0, 0, nil, 1, []int{FaninA}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 7, 0, 0, 0, nil, 1, []int{FaninB}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_SLICE, 0, 3, 0, 0, 1, []int{FaninA}, 1, []int{FaninA0}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_SLICE, 0, 3, 0, 0, 1, []int{FaninB}, 1, []int{FaninB0}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_SLICE, 0, 7, 4, 0, 1, []int{FaninA}, 1, []int{FaninA1}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_SLICE, 0, 7, 4, 0, 1, []int{FaninB}, 1, []int{FaninB1}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_ARI_ADD, 0, 4, 0, 0, 2, Fanins1[:], 1, []int{FaninR0}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_SLICE, 0, 3, 0, 0, 1, []int{FaninR0}, 1, []int{FaninS0}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_SLICE, 0, 4, 4, 0, 1, []int{FaninR0}, 1, []int{FaninRCO}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_ARI_ADD, 0, 4, 0, 0, 3, Fanins2[:], 1, []int{FaninR1}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_SLICE, 0, 3, 0, 0, 1, []int{FaninR1}, 1, []int{FaninS1}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_SLICE, 0, 4, 4, 0, 1, []int{FaninR1}, 1, []int{FaninCO}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CONCAT, 0, 7, 0, 0, 2, Fanins3[:], 1, []int{FaninS}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CO, 0, 7, 0, 0, 1, []int{FaninS}, 0, nil, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CO, 0, 0, 0, 0, 1, []int{FaninCO}, 0, nil, nil)
	Ndr_WriteVerilog("", pDesign, ppNames)
	Ndr_Write("add8.ndr", pDesign)
	Ndr_Delete(pDesign)
}
func Ndr_ModuleTestHierarchy() {
	ppNames := []string{"",
		"mux21w", "mux41w",
		"sel", "d", "out",
		"d0", "d1", "d2", "d3",
		"sel0", "sel1",
		"t0", "t1",
		"i0", "i1", "i2"}
	_ = ppNames
	FaninSel := 3
	_ = FaninSel
	FaninSel0 := 10
	_ = FaninSel0
	FaninSel1 := 11
	_ = FaninSel1
	FaninD := 4
	_ = FaninD
	FaninD0 := 6
	_ = FaninD0
	FaninD1 := 7
	_ = FaninD1
	FaninD2 := 8
	_ = FaninD2
	FaninD3 := 9
	_ = FaninD3
	FaninT0 := 12
	_ = FaninT0
	FaninT1 := 13
	_ = FaninT1
	FaninOut := 5
	_ = FaninOut
	Fanins1 := [3]int{FaninSel, FaninD1, FaninD0}
	_ = Fanins1
	Fanins3 := [3][3]int{{FaninSel0, FaninD1, FaninD0},
		{FaninSel0, FaninD3, FaninD2},
		{FaninSel1, FaninT1, FaninT0}}
	_ = Fanins3
	pDesign := Ndr_Create(2)
	_ = pDesign
	var Module21, Module41 int
	Module21 = Ndr_AddModule(pDesign, 1)
	Ndr_AddObject(pDesign, Module21, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, []int{FaninSel}, nil)
	Ndr_AddObject(pDesign, Module21, ABC_OPER_CI, 0, 3, 0, 0, 0, nil, 1, []int{FaninD1}, nil)
	Ndr_AddObject(pDesign, Module21, ABC_OPER_CI, 0, 3, 0, 0, 0, nil, 1, []int{FaninD0}, nil)
	Ndr_AddObject(pDesign, Module21, ABC_OPER_BIT_MUX, 0, 3, 0, 0, 3, Fanins1[:], 1, []int{FaninOut}, nil)
	Ndr_AddObject(pDesign, Module21, ABC_OPER_CO, 0, 3, 0, 0, 1, []int{FaninOut}, 0, nil, nil)
	Module41 = Ndr_AddModule(pDesign, 2)
	Ndr_AddObject(pDesign, Module41, ABC_OPER_CI, 0, 1, 0, 0, 0, nil, 1, []int{FaninSel}, nil)
	Ndr_AddObject(pDesign, Module41, ABC_OPER_CI, 0, 15, 0, 0, 0, nil, 1, []int{FaninD}, nil)
	Ndr_AddObject(pDesign, Module41, ABC_OPER_SLICE, 0, 3, 0, 0, 1, []int{FaninD}, 1, []int{FaninD0}, nil)
	Ndr_AddObject(pDesign, Module41, ABC_OPER_SLICE, 0, 7, 4, 0, 1, []int{FaninD}, 1, []int{FaninD1}, nil)
	Ndr_AddObject(pDesign, Module41, ABC_OPER_SLICE, 0, 11, 8, 0, 1, []int{FaninD}, 1, []int{FaninD2}, nil)
	Ndr_AddObject(pDesign, Module41, ABC_OPER_SLICE, 0, 15, 12, 0, 1, []int{FaninD}, 1, []int{FaninD3}, nil)
	Ndr_AddObject(pDesign, Module41, ABC_OPER_SLICE, 0, 0, 0, 0, 1, []int{FaninSel}, 1, []int{FaninSel0}, nil)
	Ndr_AddObject(pDesign, Module41, ABC_OPER_SLICE, 0, 1, 1, 0, 1, []int{FaninSel}, 1, []int{FaninSel1}, nil)
	Ndr_AddObject(pDesign, Module41, Module21, 14, 3, 0, 0, 3, Fanins3[0][:], 1, []int{FaninT0}, nil)
	Ndr_AddObject(pDesign, Module41, Module21, 15, 3, 0, 0, 3, Fanins3[1][:], 1, []int{FaninT1}, nil)
	Ndr_AddObject(pDesign, Module41, Module21, 16, 3, 0, 0, 3, Fanins3[2][:], 1, []int{FaninOut}, nil)
	Ndr_AddObject(pDesign, Module41, ABC_OPER_CO, 0, 3, 0, 0, 1, []int{FaninOut}, 0, nil, nil)
	Ndr_WriteVerilog("", pDesign, ppNames)
	Ndr_Write("mux41w.ndr", pDesign)
	Ndr_Delete(pDesign)
}
func Ndr_ModuleTestMemory() {
	ppNames := []string{"",
		"clk", "raddr", "waddr", "data", "mem_init", "out",
		"read1", "read2",
		"mem_fo1", "mem_fo2", "mem_fi1", "mem_fi2",
		"i_reg1", "i_reg2",
		"i_read1", "i_read2",
		"i_write1", "i_write2", "memtest"}
	_ = ppNames
	NameIdClk := 1
	_ = NameIdClk
	NameIdRaddr := 2
	_ = NameIdRaddr
	NameIdWaddr := 3
	_ = NameIdWaddr
	NameIdData := 4
	_ = NameIdData
	NameIdMemInit := 5
	_ = NameIdMemInit
	NameIdFF1 := 9
	_ = NameIdFF1
	NameIdFF2 := 10
	_ = NameIdFF2
	FaninsFF1 := [2]int{11, 5}
	_ = FaninsFF1
	FaninsFF2 := [2]int{12, 5}
	_ = FaninsFF2
	NameIdWrite1 := 11
	_ = NameIdWrite1
	NameIdWrite2 := 12
	_ = NameIdWrite2
	FaninsWrite1 := [3]int{9, 3, 4}
	_ = FaninsWrite1
	FaninsWrite2 := [3]int{10, 3, 4}
	_ = FaninsWrite2
	NameIdRead1 := 7
	_ = NameIdRead1
	NameIdRead2 := 8
	_ = NameIdRead2
	FaninsRead1 := [2]int{11, 2}
	_ = FaninsRead1
	FaninsRead2 := [2]int{12, 2}
	_ = FaninsRead2
	NameIdComp := 6
	_ = NameIdComp
	FaninsComp := [2]int{7, 8}
	_ = FaninsComp
	pDesign := Ndr_Create(19)
	_ = pDesign
	ModuleID := Ndr_AddModule(pDesign, 19)
	_ = ModuleID
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, []int{NameIdClk}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 8, 0, 0, 0, nil, 1, []int{NameIdRaddr}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 8, 0, 0, 0, nil, 1, []int{NameIdWaddr}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 31, 0, 0, 0, nil, 1, []int{NameIdData}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 16383, 0, 0, 0, nil, 1, []int{NameIdMemInit}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CO, 0, 0, 0, 0, 1, []int{NameIdComp}, 0, nil, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_DFF, 13, 16383, 0, 0, 2, FaninsFF1[:], 1, []int{NameIdFF1}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_DFF, 14, 16383, 0, 0, 2, FaninsFF2[:], 1, []int{NameIdFF2}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_RAMW, 17, 16383, 0, 0, 3, FaninsWrite1[:], 1, []int{NameIdWrite1}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_RAMW, 18, 16383, 0, 0, 3, FaninsWrite2[:], 1, []int{NameIdWrite2}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_RAMR, 15, 31, 0, 0, 2, FaninsRead1[:], 1, []int{NameIdRead1}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_RAMR, 16, 31, 0, 0, 2, FaninsRead2[:], 1, []int{NameIdRead2}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_COMP_NOTEQU, 0, 0, 0, 0, 2, FaninsComp[:], 1, []int{NameIdComp}, nil)
	Ndr_WriteVerilog("", pDesign, ppNames)
	Ndr_Write("memtest.ndr", pDesign)
	Ndr_Delete(pDesign)
}
func Ndr_ModuleTestFlop() {
	ppNames := []string{"", "flop", "data", "clk", "reset", "set", "enable", "async", "sre", "init", "q"}
	_ = ppNames
	NameIdData := 2
	_ = NameIdData
	NameIdClk := 3
	_ = NameIdClk
	NameIdReset := 4
	_ = NameIdReset
	NameIdSet := 5
	_ = NameIdSet
	NameIdEnable := 6
	_ = NameIdEnable
	NameIdAsync := 7
	_ = NameIdAsync
	NameIdSre := 8
	_ = NameIdSre
	NameIdInit := 9
	_ = NameIdInit
	NameIdQ := 10
	_ = NameIdQ
	Fanins := [8]int{NameIdData, NameIdClk, NameIdReset, NameIdSet, NameIdEnable, NameIdAsync, NameIdSre, NameIdInit}
	_ = Fanins
	pDesign := Ndr_Create(1)
	_ = pDesign
	ModuleID := Ndr_AddModule(pDesign, 1)
	_ = ModuleID
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 3, 0, 0, 0, nil, 1, []int{NameIdData}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, []int{NameIdClk}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, []int{NameIdReset}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, []int{NameIdSet}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, []int{NameIdEnable}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, []int{NameIdAsync}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, []int{NameIdSre}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 3, 0, 0, 0, nil, 1, []int{NameIdInit}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_DFFRSE, 0, 3, 0, 0, 8, Fanins[:], 1, []int{NameIdQ}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CO, 0, 3, 0, 0, 1, []int{NameIdQ}, 0, nil, nil)
	Ndr_WriteVerilog("", pDesign, ppNames)
	Ndr_Write("flop.ndr", pDesign)
	Ndr_Delete(pDesign)
}
func Ndr_ModuleTestSelSel() {
	NameIdC := 2
	_ = NameIdC
	NameIdD0 := 3
	_ = NameIdD0
	NameIdD1 := 4
	_ = NameIdD1
	NameIdD2 := 5
	_ = NameIdD2
	NameIdD3 := 6
	_ = NameIdD3
	NameIdOut := 7
	_ = NameIdOut
	Fanins := [8]int{NameIdC, NameIdD0, NameIdD1, NameIdD2, NameIdD3}
	_ = Fanins
	pDesign := Ndr_Create(1)
	_ = pDesign
	ModuleID := Ndr_AddModule(pDesign, 1)
	_ = ModuleID
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 3, 0, 0, 0, nil, 1, []int{NameIdC}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 2, 0, 0, 0, nil, 1, []int{NameIdD0}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 2, 0, 0, 0, nil, 1, []int{NameIdD1}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 2, 0, 0, 0, nil, 1, []int{NameIdD2}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 2, 0, 0, 0, nil, 1, []int{NameIdD3}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_SEL_SEL, 0, 2, 0, 0, 5, Fanins[:], 1, []int{NameIdOut}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CO, 0, 2, 0, 0, 1, []int{NameIdOut}, 0, nil, nil)
	Ndr_Write("sel.ndr", pDesign)
	Ndr_Delete(pDesign)
}
func Ndr_ModuleTestDec() {
	NameIdIn := 2
	_ = NameIdIn
	NameIdOut := 3
	_ = NameIdOut
	pDesign := Ndr_Create(1)
	_ = pDesign
	ModuleID := Ndr_AddModule(pDesign, 1)
	_ = ModuleID
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 1, 0, 0, 0, nil, 1, []int{NameIdIn}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_SEL_DEC, 0, 3, 0, 0, 1, []int{NameIdIn}, 1, []int{NameIdOut}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CO, 0, 3, 0, 0, 1, []int{NameIdOut}, 0, nil, nil)
	Ndr_Write("dec.ndr", pDesign)
	Ndr_Delete(pDesign)
}
func Ndr_ModuleTestAddSub() {
	NameIdInMode := 2
	_ = NameIdInMode
	NameIdInCin := 3
	_ = NameIdInCin
	NameIdInA := 4
	_ = NameIdInA
	NameIdInB := 5
	_ = NameIdInB
	NameIdOut := 6
	_ = NameIdOut
	Fanins := [8]int{2, 3, 4, 5}
	_ = Fanins
	pDesign := Ndr_Create(1)
	_ = pDesign
	ModuleID := Ndr_AddModule(pDesign, 1)
	_ = ModuleID
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, []int{NameIdInMode}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, []int{NameIdInCin}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 2, 0, 0, 0, nil, 1, []int{NameIdInA}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 2, 0, 0, 0, nil, 1, []int{NameIdInB}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_ARI_ADDSUB, 0, 3, 0, 0, 4, Fanins[:], 1, []int{NameIdOut}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CO, 0, 3, 0, 0, 1, []int{NameIdOut}, 0, nil, nil)
	Ndr_Write("addsub.ndr", pDesign)
	Ndr_Delete(pDesign)
}
func Ndr_ModuleTestLut() {
	NameIdIn := 2
	_ = NameIdIn
	NameIdOut := 3
	_ = NameIdOut
	pDesign := Ndr_Create(1)
	_ = pDesign
	ModuleID := Ndr_AddModule(pDesign, 1)
	_ = ModuleID
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 1, 0, 0, 0, nil, 1, []int{NameIdIn}, nil)
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_LUT, 0, 0, 0, 0, 1, []int{NameIdIn}, 1, []int{NameIdOut}, uint64(0x8))
	Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CO, 0, 0, 0, 0, 1, []int{NameIdOut}, 0, nil, nil)
	Ndr_Write("lut_test.ndr", pDesign)
	Ndr_Delete(pDesign)
}
