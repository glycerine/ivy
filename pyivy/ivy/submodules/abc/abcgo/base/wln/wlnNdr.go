package wln

// NDR import/export ported from src/base/wln/wlnNdr.c.
import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Wln_NtkToNdr(p *Wln_Ntk_t) *miniaig.Ndr_Data_t {
	design := miniaig.Ndr_Create(1)
	mod := miniaig.Ndr_AddModule(design, 1)
	add := func(id int32, typ int, inputs, outputs []int, fn any) {
		miniaig.Ndr_AddObject(design, mod, typ, 0, int(Wln_ObjRangeEnd(p, id)), int(Wln_ObjRangeBeg(p, id)), int(Wln_ObjIsSigned(p, id)), len(inputs), inputs, len(outputs), outputs, fn)
	}
	for i := int32(0); i < Wln_NtkPiNum(p); i++ {
		id := Wln_NtkCi(p, i)
		add(id, ABC_OPER_CI, nil, []int{int(id)}, nil)
	}
	for id := int32(1); id < Wln_NtkObjNum(p); id++ {
		if Wln_ObjIsCio(p, id) != 0 {
			continue
		}
		var fanins []int
		for _, f := range Wln_ObjFanins(p, id)[:Wln_ObjFaninNum(p, id)] {
			fanins = append(fanins, int(f))
		}
		var fn any
		if Wln_ObjIsConst(p, id) != 0 {
			fn = Wln_ObjConstString(p, id)
		}
		add(id, int(Wln_ObjType(p, id)), fanins, []int{int(id)}, fn)
	}
	for i := int32(0); i < Wln_NtkPoNum(p); i++ {
		id := Wln_NtkCo(p, i)
		add(id, ABC_OPER_CO, []int{int(Wln_ObjFanin0(p, id))}, nil, nil)
	}
	return design
}
func Wln_WriteNdr(p *Wln_Ntk_t, name string) {
	d := Wln_NtkToNdr(p)
	miniaig.Ndr_Write(name, d)
	miniaig.Ndr_Delete(d)
	fmt.Printf("Dumped the current design into file \"%s\".\n", name)
}
func Wln_NtkToNdrTest(p *Wln_Ntk_t) {
	d := Wln_NtkToNdr(p)
	names := make([]string, Wln_NtkObjNum(p)+1)
	for i := int32(1); i < Wln_NtkObjNum(p); i++ {
		names[i] = Wln_ObjName(p, i)
	}
	miniaig.Ndr_WriteVerilog("", d, names)
	miniaig.Ndr_Write("test.ndr", d)
	miniaig.Ndr_Delete(d)
}
func Ndr_ObjGetRange(p *miniaig.Ndr_Data_t, obj int32, end, beg *int32) int32 {
	var data []int
	n := miniaig.Ndr_ObjReadArray(p, int(obj), miniaig.NDR_RANGE, &data)
	*end = 0
	*beg = 0
	if n == 0 {
		return 0
	}
	*end = int32(data[0])
	*beg = *end
	if n > 1 {
		*beg = int32(data[1])
	}
	if n == 3 {
		return 1
	}
	return 0
}
func Ndr_NtkPrintObjects(p *Wln_Ntk_t) {
	fmt.Println("Node IDs and their fanins:")
	for id := int32(1); id < Wln_NtkObjNum(p); id++ {
		fmt.Printf("%5d = ", id)
		k := int32(0)
		for ; k < Wln_ObjFaninNum(p, id); k++ {
			f := Wln_ObjFanin(p, id, k)
			if f != 0 {
				fmt.Printf("%5d ", f)
			}
		}
		for ; k < 4; k++ {
			fmt.Print("      ")
		}
		fmt.Printf("    Name Id %d ", Wln_ObjNameId(p, id))
		if Wln_ObjIsPi(p, id) != 0 {
			fmt.Print("  pi  ")
		}
		if Wln_ObjIsPo(p, id) != 0 {
			fmt.Print("  po  ")
		}
		fmt.Println()
	}
}
func wlnNdrObjects(p *miniaig.Ndr_Data_t, mod int) []int {
	var out []int
	for obj := mod + 1; obj < miniaig.Ndr_DataEnd(p, mod); obj += miniaig.Ndr_DataSize(p, obj) {
		if miniaig.Ndr_DataType(p, obj) == miniaig.NDR_OBJECT {
			out = append(out, obj)
		}
	}
	return out
}
func Wln_NtkCheckIntegrity(p *miniaig.Ndr_Data_t) {
	names := make(map[int]int)
	for _, obj := range wlnNdrObjects(p, 2) {
		id := miniaig.Ndr_ObjReadBody(p, obj, miniaig.NDR_OUTPUT)
		typ := miniaig.Ndr_ObjReadBody(p, obj, miniaig.NDR_OPERTYPE)
		if id == -1 {
			if typ != ABC_OPER_CO {
				fmt.Printf("Internal object %d of type %s has no output name.\n", obj, miniaig.Abc_OperName(typ))
			}
			continue
		}
		if names[id] > 0 {
			fmt.Printf("Output name %d is used more than once (obj %d and obj %d).\n", id, names[id], obj)
		}
		names[id] = obj
	}
	for _, obj := range wlnNdrObjects(p, 2) {
		typ := miniaig.Ndr_ObjReadBody(p, obj, miniaig.NDR_OPERTYPE)
		var in []int
		miniaig.Ndr_ObjReadArray(p, obj, miniaig.NDR_INPUT, &in)
		for i, id := range in {
			if names[id] == 0 && !(typ == ABC_OPER_DFFRSE && i >= 5 && i <= 7) {
				fmt.Printf("Input name %d appearing as fanin %d of obj %d is not used as output name in any object.\n", id, i, obj)
			}
		}
	}
}
func wlnNdrString(words []int) string {
	data := make([]byte, 4*len(words))
	for i, w := range words {
		binary.LittleEndian.PutUint32(data[4*i:], uint32(w))
	}
	for i, c := range data {
		if c == 0 {
			return string(data[:i])
		}
	}
	return string(data)
}
func Wln_NtkFromNdr(p *miniaig.Ndr_Data_t, dump int32) *Wln_Ntk_t {
	const mod = 2
	objects := wlnNdrObjects(p, mod)
	q := Wln_NtkAlloc("top", int32(miniaig.Ndr_DataObjNum(p, mod)))
	Wln_NtkCheckIntegrity(p)
	Wln_NtkCleanNameId(q)
	Wln_NtkCleanInstId(q)
	var consts []string
	for phase := 0; phase < 3; phase++ {
		for _, obj := range objects {
			typ := miniaig.Ndr_ObjReadBody(p, obj, miniaig.NDR_OPERTYPE)
			if phase == 0 && typ != ABC_OPER_CI || phase == 1 && (typ == ABC_OPER_CI || typ == ABC_OPER_CO) || phase == 2 && typ != ABC_OPER_CO {
				continue
			}
			var end, beg int32
			signed := Ndr_ObjGetRange(p, int32(obj), &end, &beg)
			id := Wln_ObjAlloc(q, int32(typ), signed, end, beg)
			name := int32(miniaig.Ndr_ObjReadBody(p, obj, miniaig.NDR_OUTPUT))
			inst := int32(miniaig.Ndr_ObjReadBody(p, obj, miniaig.NDR_NAME))
			var inputs []int
			miniaig.Ndr_ObjReadArray(p, obj, miniaig.NDR_INPUT, &inputs)
			if phase == 1 {
				assert(typ != miniaig.ABC_OPER_DFF, "NDR DFF type")
				if Wln_ObjIsSlice(q, id) != 0 {
					Wln_ObjSetSlice(q, id, vec.Hash_Int2ManInsert(q.pRanges, end, beg, 0))
				} else if Wln_ObjIsConst(q, id) != 0 {
					consts = append(consts, wlnNdrString(miniaig.Ndr_ObjReadBodyP(p, obj, miniaig.NDR_FUNCTION)))
				}
				for _, f := range inputs {
					Wln_ObjAddFanin(q, id, int32(f))
				}
				Wln_ObjSetNameId(q, id, name)
				if inst > 0 {
					Wln_ObjSetInstId(q, id, inst)
				}
				if typ == miniaig.ABC_OPER_ARI_SMUL {
					assert(Wln_ObjFaninNum(q, id) == 2, "signed multiplier fanins")
					Wln_ObjSetSigned(q, Wln_ObjFanin0(q, id))
					Wln_ObjSetSigned(q, Wln_ObjFanin1(q, id))
				}
			} else if phase == 0 {
				Wln_ObjSetNameId(q, id, name)
				if inst > 0 {
					Wln_ObjSetInstId(q, id, inst)
				}
			} else {
				assert(len(inputs) == 1 && name == -1 && inst == -1, "NDR output")
				Wln_ObjAddFanin(q, id, int32(inputs[0]))
			}
		}
	}
	if vec.Vec_IntCountPositive(&q.vInstIds) == 0 {
		vec.Vec_IntErase(&q.vInstIds)
	}
	inv := vec.Vec_IntInvert(&q.vNameIds, 0)
	for id := int32(1); id < Wln_NtkObjNum(q); id++ {
		for k := int32(0); k < Wln_ObjFaninNum(q, id); k++ {
			f := Wln_ObjFanin(q, id, k)
			if f != 0 {
				Wln_ObjSetFanin(q, id, k, int32(vec.Vec_IntEntry(inv, int(f))))
			}
		}
	}
	maxName := vec.Vec_IntFindMax(&q.vNameIds)
	digits := util.Abc_Base10Log(uint32(maxName + 1))
	q.pManName = util.Abc_NamStart(maxName+1, 10)
	for i := int32(1); i <= maxName; i++ {
		var found int32
		id := util.Abc_NamStrFindOrAdd(q.pManName, fmt.Sprintf("s%0*d", digits, i), &found)
		assert(found == 0 && id == i, "NDR name IDs")
	}
	k := 0
	for id := int32(1); id < Wln_NtkObjNum(q); id++ {
		if Wln_ObjIsConst(q, id) != 0 {
			Wln_ObjSetConst(q, id, util.Abc_NamStrFindOrAdd(q.pManName, consts[k], nil))
			k++
		}
	}
	assert(k == len(consts), "NDR constant count")
	Wln_WriteVer(q, "temp_ndr.v")
	fmt.Printf("Dumped design \"%s\" into file \"temp_ndr.v\".\n", q.pName)
	if Wln_NtkIsAcyclic(q) == 0 {
		Wln_NtkFree(q)
		return nil
	}
	result := Wln_NtkDupDfs(q)
	Wln_NtkFree(q)
	return result
}
func Wln_ReadNdr(name string) *Wln_Ntk_t {
	data := miniaig.Ndr_Read(name)
	var q *Wln_Ntk_t
	if data != nil {
		q = Wln_NtkFromNdr(data, 0)
	}
	if q != nil {
		miniaig.Ndr_Delete(data)
		return q
	}
	miniaig.Ndr_Delete(data)
	return q
}
func Wln_ReadNdrTest() {
	p := Wln_ReadNdr(`D:\temp\brijesh\for_alan_dff_warning\work_fir_filter_fir_filter_proc_out.ndr`)
	Wln_WriteVer(p, "test__test.v")
	Wln_NtkPrint(p)
	Wln_NtkStaticFanoutTest(p)
	Wln_NtkFree(p)
}
func Wln_NtkRetimeTest(name string, skip, dump, verbose int32) {
	data := miniaig.Ndr_Read(name)
	var p *Wln_Ntk_t
	if data != nil {
		p = Wln_NtkFromNdr(data, dump)
	}
	miniaig.Ndr_Delete(data)
	if p == nil {
		fmt.Println("Retiming network is not available.")
		return
	}
	Wln_NtkRetimeCreateDelayInfo(p)
	Wln_NtkRetime(p, skip, verbose)
	Wln_NtkFree(p)
}
