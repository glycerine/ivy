package bac

// Binary BAC serialization, ported from base/bac/bacBac.c.
import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func BacManReadBacLine(vOut *vec.Vec_Str_t, pos *int32) (string, bool) {
	line, next, ok := bacReadHeaderLine(vec.Vec_StrArray(vOut), int(*pos))
	*pos = int32(next)
	return line, ok
}
func BacManReadBacNameAndNums(line string) (string, [4]int32, bool) {
	return bacParseHeaderNumbers(line)
}
func Bac_ManReadBacVecStr(vOut *vec.Vec_Str_t, pos *int32, p *vec.Vec_Str_t, nSize int32) {
	data := vec.Vec_StrArray(vOut)
	assert(nSize >= 0 && *pos >= 0 && int64(*pos)+int64(nSize) <= int64(len(data)), "BAC string vector fits file")
	vec.Vec_StrFill(p, int(nSize), 0)
	copy(vec.Vec_StrArray(p), data[int(*pos):int(*pos+nSize)])
	*pos += nSize
	assert(vec.Vec_StrSize(p) == vec.Vec_StrCap(p), "BAC string vector fills capacity")
}
func Bac_ManReadBacVecInt(vOut *vec.Vec_Str_t, pos *int32, p *vec.Vec_Int_t, nSize int32) {
	data := vec.Vec_StrArray(vOut)
	assert(nSize >= 0 && nSize%4 == 0 && *pos >= 0 && int64(*pos)+int64(nSize) <= int64(len(data)), "BAC integer vector fits file")
	vec.Vec_IntFill(p, int(nSize/4), 0)
	for i := range vec.Vec_IntArray(p) {
		vec.Vec_IntArray(p)[i] = int32(binary.NativeEndian.Uint32(data[int(*pos)+4*i:]))
	}
	*pos += nSize
	assert(vec.Vec_IntSize(p) == vec.Vec_IntCap(p), "BAC integer vector fills capacity")
}
func Bac_ManReadBacNtk(vOut *vec.Vec_Str_t, pos *int32, ntk *Bac_Ntk_t) {
	Bac_ManReadBacVecStr(vOut, pos, &ntk.vType, int32(vec.Vec_StrCap(&ntk.vType)))
	Bac_ManReadBacVecInt(vOut, pos, &ntk.vFanin, 4*int32(vec.Vec_StrCap(&ntk.vType)))
	Bac_ManReadBacVecInt(vOut, pos, &ntk.vInfo, 12*int32(vec.Vec_IntCap(&ntk.vInfo)/3))
	for i, typ := range vec.Vec_StrArray(&ntk.vType) {
		if typ>>1 == byte(BAC_OBJ_PI) {
			vec.Vec_IntPush(&ntk.vInputs, int32(i))
		}
		if typ>>1 == byte(BAC_OBJ_PO) {
			vec.Vec_IntPush(&ntk.vOutputs, int32(i))
		}
	}
	assert(vec.Vec_IntSize(&ntk.vInputs) == vec.Vec_IntCap(&ntk.vInputs), "BAC PI count matches header")
	assert(vec.Vec_IntSize(&ntk.vOutputs) == vec.Vec_IntCap(&ntk.vOutputs), "BAC PO count matches header")
}
func Bac_ManReadBacInt(vOut *vec.Vec_Str_t) *Bac_Man_t {
	pos := int32(0)
	var line string
	for {
		var ok bool
		line, ok = BacManReadBacLine(vOut, &pos)
		if !ok {
			return nil
		}
		if len(line) == 0 || line[0] != '#' {
			break
		}
	}
	name, nums, ok := BacManReadBacNameAndNums(line)
	if !ok {
		return nil
	}
	nNtks := nums[0]
	assert(nNtks > 0, "BAC network count is positive")
	man := Bac_ManAlloc(name, nNtks)
	for i := int32(1); i <= nNtks; i++ {
		line, ok = BacManReadBacLine(vOut, &pos)
		if !ok {
			Bac_ManFree(man)
			return nil
		}
		name, nums, ok = BacManReadBacNameAndNums(line)
		if !ok {
			Bac_ManFree(man)
			return nil
		}
		assert(nums[0] >= 0 && nums[1] >= 0 && nums[2] >= 0, "BAC network sizes are nonnegative")
		id := util.Abc_NamStrFindOrAdd(man.pStrs, name, nil)
		ntk := Bac_ManNtk(man, i)
		Bac_NtkAlloc(ntk, id, nums[0], nums[1], nums[2])
		vec.Vec_IntFill(&ntk.vInfo, int(3*nums[3]), -1)
	}
	for i := int32(1); i <= nNtks; i++ {
		Bac_ManReadBacNtk(vOut, &pos, Bac_ManNtk(man, i))
	}
	assert(int(pos) == vec.Vec_StrSize(vOut), "BAC file was fully consumed")
	return man
}
func Bac_ManReadBac(filename string) *Bac_Man_t {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Cannot open file %q for reading.\n", filename)
		return nil
	}
	v := vec.Vec_StrAllocArrayCopy(data, len(data))
	man := Bac_ManReadBacInt(v)
	if man != nil {
		man.pSpec = filename
	}
	return man
}
func Bac_ManWriteBacNtk(out *vec.Vec_Str_t, ntk *Bac_Ntk_t) {
	data := vec.Vec_StrArray(&ntk.vType)
	vec.Vec_StrPushBuffer(out, data, len(data))
	for _, values := range [][]int32{vec.Vec_IntArray(&ntk.vFanin)[:len(data)], vec.Vec_IntArray(&ntk.vInfo)[:3*(vec.Vec_IntSize(&ntk.vInfo)/3)]} {
		for _, value := range values {
			var buf [4]byte
			binary.NativeEndian.PutUint32(buf[:], uint32(value))
			vec.Vec_StrPushBuffer(out, buf[:], 4)
		}
	}
}
func Bac_ManWriteBacInt(out *vec.Vec_Str_t, man *Bac_Man_t) {
	vec.Vec_StrPrintStr(out, fmt.Sprintf("# Design \"%s\" written by ABC on %s\n", Bac_ManName(man), extra.Extra_TimeStamp()))
	vec.Vec_StrPrintStr(out, fmt.Sprintf("%s %d \n", Bac_ManName(man), Bac_ManNtkNum(man)))
	for i := int32(1); i <= Bac_ManNtkNum(man); i++ {
		ntk := Bac_ManNtk(man, i)
		vec.Vec_StrPrintStr(out, fmt.Sprintf("%s %d %d %d %d \n", Bac_NtkName(ntk), Bac_NtkPiNum(ntk), Bac_NtkPoNum(ntk), Bac_NtkObjNum(ntk), Bac_NtkInfoNum(ntk)))
	}
	for i := int32(1); i <= Bac_ManNtkNum(man); i++ {
		Bac_ManWriteBacNtk(out, Bac_ManNtk(man, i))
	}
}
func Bac_ManWriteBac(filename string, man *Bac_Man_t) {
	assert(man.pMioLib == nil, "BAC binary writer requires no library")
	out := vec.Vec_StrAlloc(10000)
	Bac_ManWriteBacInt(out, man)
	if vec.Vec_StrSize(out) > 0 {
		if err := os.WriteFile(filename, vec.Vec_StrArray(out), 0666); err != nil {
			fmt.Printf("Cannot open file %q for writing.\n", filename)
		}
	}
}
