package wln

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Wln_WriteTableOne writes one source-style Verilog lookup table.
func Wln_WriteTableOne(f *os.File, nFans, nOuts int32, table []uint64, id int32) {
	assert(nOuts > 0 && nOuts <= 64 && 64%nOuts == 0, "nOuts > 0 && nOuts <= 64 && (64 % nOuts) == 0")
	fmt.Fprintf(f, "module table%d(ind, val);\n", id)
	fmt.Fprintf(f, "  input  [%d:0] ind;\n", nFans-1)
	fmt.Fprintf(f, "  output [%d:0] val;\n", nOuts-1)
	fmt.Fprintf(f, "  reg    [%d:0] val;\n", nOuts-1)
	fmt.Fprint(f, "  always @(ind)\n  begin\n    case (ind)\n")
	for m := int32(0); m < 1<<nFans; m++ {
		value := uint32((table[(nOuts*m)>>6] >> ((nOuts * m) & 63)) & util.Abc_Tt6Mask(nOuts))
		fmt.Fprintf(f, "      %d'h%x: val = %d'h%x;\n", nFans, m, nOuts, value)
	}
	fmt.Fprint(f, "    endcase\n  end\nendmodule\n\n")
}

func Wln_WriteTables(f *os.File, p *Wln_Ntk_t) {
	if p.vTables == nil || vec.Vec_PtrSize(p.vTables) == 0 {
		return
	}
	nodes := vec.Vec_IntStart(vec.Vec_PtrSize(p.vTables))
	for id := int32(1); id < Wln_NtkObjNum(p); id++ {
		if Wln_ObjType(p, id) == ABC_OPER_TABLE {
			vec.Vec_IntWriteEntry(nodes, int(Wln_ObjFanin1(p, id)), id)
		}
	}
	for i := 0; i < vec.Vec_PtrSize(p.vTables); i++ {
		table := vec.Vec_PtrEntry(p.vTables, i).([]uint64)
		node := vec.Vec_IntEntry(nodes, i)
		fanin := Wln_ObjFanin0(p, node)
		Wln_WriteTableOne(f, Wln_ObjRange(p, fanin), Wln_ObjRange(p, node), table, int32(i))
	}
}

func Wln_WriteAddPos(p *Wln_Ntk_t) {
	for id := int32(1); id < Wln_NtkObjNum(p); id++ {
		if Wln_ObjIsCio(p, id) == 0 {
			Wln_ObjCreateCo(p, id)
		}
	}
}

func Wln_WriteVerIntVec(f *os.File, p *Wln_Ntk_t, v *Vec_Int_t, start int32) {
	lineLength, count := start, int32(0)
	for i := int32(0); i < Vec_IntSize(v); i++ {
		name := Wln_ObjName(p, Vec_IntEntry(v, i))
		added := int32(len(name)) + 2
		if count != 0 && lineLength+added+3 > 70 {
			fmt.Fprint(f, "\n   ")
			lineLength, count = start, 0
		}
		suffix := ","
		if i == Vec_IntSize(v)-1 {
			suffix = ""
		}
		fmt.Fprintf(f, " %s%s", name, suffix)
		lineLength += added
		count++
	}
}
