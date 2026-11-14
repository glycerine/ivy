package gia

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Gia_ManDupMuxesTest(p *Gia_Man_t) *Gia_Man_t {
	coarse := Gia_ManDupMuxes(p, 2)
	result := Gia_ManDupNoMuxes(coarse, 0)
	Gia_ManPrintMuxStats(p)
	Gia_ManPrintMuxStats(coarse)
	Gia_ManPrintMuxStats(result)
	Gia_ManStop(coarse)
	return result
}

func Gia_ManMuxRestructure(src *Gia_Man_t) *Gia_Man_t {
	if src.PMuxes == nil || Gia_ManHasChoices(src) != 0 || Gia_ManHasMapping(src) != 0 {
		panic("gia: MUX restructuring requires unmapped real gates without choices")
	}
	src.Refs = nil
	Gia_ManCreateRefs(src)
	used := make([]bool, src.nObjs)
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	dst.EnableMuxes()
	Gia_ManHashStart(dst)
	src.pObjs[0].Value = 0
	for _, obj := range src.pObjs[1:] {
		a, b, c, kind := src.Gate(obj.id)
		var lit int32
		switch kind {
		case 1:
			lit = Gia_ManAppendCi(dst)
		case 2:
			lit = Gia_ManHashAnd(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
		case 3:
			lit = Gia_ManHashXor(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
		case 4:
			// Source restructuring shares an intermediate selector: on the then
			// branch it equals the inner control; otherwise it equals the
			// original else result. Preserve a noncanonical then-edge phase.
			bc, bt, be, bk := src.Gate(b >> 1)
			cc, _, _, ck := src.Gate(c >> 1)
			if bk == 4 && ck == 4 && !used[b>>1] && !used[c>>1] && bc>>1 == cc>>1 {
				selector := Gia_ManHashMux(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, bc), Gia_ObjLitCopy(src, c))
				inner := Gia_ManHashMux(dst, selector, Gia_ObjLitCopy(src, bt), Gia_ObjLitCopy(src, be)) ^ (b & 1)
				lit = Gia_ManHashMux(dst, Gia_ObjLitCopy(src, a), inner, selector)
				used[b>>1], used[c>>1], used[obj.id] = true, true, true
			} else {
				lit = Gia_ManHashMux(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b), Gia_ObjLitCopy(src, c))
			}

		case 5:
			lit = Gia_ManAppendBuf(dst, Gia_ObjLitCopy(src, a))
		case 6:
			lit = Gia_ManAppendCo(dst, Gia_ObjLitCopy(src, a))
		default:
			panic("gia: invalid MUX restructuring object")
		}
		obj.Value = uint32(lit)
	}
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, src.nRegs)
	giaCopyInterfaceMetadata(dst, src)
	result := Gia_ManCleanup(dst)
	Gia_ManStop(dst)
	return result
}
func Gia_ManDupMuxRestructure(src *Gia_Man_t) *Gia_Man_t {
	coarse := Gia_ManDupMuxes(src, 2)
	middle := Gia_ManMuxRestructure(coarse)
	Gia_ManStop(coarse)
	result := Gia_ManDupNoMuxes(middle, 0)
	Gia_ManStop(middle)
	return result
}

func giaMuxReference(p *Gia_Man_t, id int32, increment bool) int32 {
	stack := []int32{id}
	var count int32
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if Gia_ObjIsMuxId(p, id) == 0 {
			continue
		}
		obj := p.pObjs[id]
		if increment {
			if Gia_ObjRefInc(p, obj) != 0 {
				continue
			}
		} else {
			if Gia_ObjRefDec(p, obj) != 0 {
				continue
			}
		}
		count++
		stack = append(stack, Gia_ObjFaninId2(p, id), Gia_ObjFaninId1(obj, id), Gia_ObjFaninId0(obj, id))
	}
	return count
}
func Gia_MuxRef_rec(p *Gia_Man_t, id int32) int32   { return giaMuxReference(p, id, true) }
func Gia_MuxDeref_rec(p *Gia_Man_t, id int32) int32 { return giaMuxReference(p, id, false) }
func giaMuxRootReference(p *Gia_Man_t, id int32, increment bool) int32 {
	if Gia_ObjIsMuxId(p, id) == 0 {
		panic("gia: expected real MUX root")
	}
	obj := p.pObjs[id]
	return 1 + giaMuxReference(p, Gia_ObjFaninId0(obj, id), increment) + giaMuxReference(p, Gia_ObjFaninId1(obj, id), increment) + giaMuxReference(p, Gia_ObjFaninId2(p, id), increment)
}
func Gia_MuxRef(p *Gia_Man_t, id int32) int32   { return giaMuxRootReference(p, id, true) }
func Gia_MuxDeref(p *Gia_Man_t, id int32) int32 { return giaMuxRootReference(p, id, false) }
func Gia_MuxMffcSize(p *Gia_Man_t, id int32) int32 {
	if Gia_ObjIsMuxId(p, id) == 0 {
		return 0
	}
	count := Gia_MuxDeref(p, id)
	if Gia_MuxRef(p, id) != count {
		panic("gia: MUX references did not restore")
	}
	return count
}
func giaMuxStructure(p *Gia_Man_t, id int32, first bool, print bool, digits int) string {
	if !first && (Gia_ObjIsMuxId(p, id) == 0 || p.Refs[id] > 0) {
		if print {
			return fmt.Sprintf("<%02d>", Gia_ObjLevelId(p, id))
		}
		return ""
	}
	obj := p.pObjs[id]
	control := Gia_ObjFaninId2(p, id)
	var name string
	if Gia_ObjIsMuxId(p, control) != 0 && p.Refs[control] == 0 {
		name = giaMuxStructure(p, control, false, print, digits)
	} else if print {
		name = fmt.Sprintf("%d<%d>", control, Gia_ObjLevelId(p, control))
	} else {
		name = fmt.Sprintf("%0*d", digits, control)
	}
	then, otherwise := Gia_ObjFaninId1(obj, id), Gia_ObjFaninId0(obj, id)
	if Gia_ObjFaninC2(p, obj) != 0 {
		then, otherwise = otherwise, then
	}
	prefix := "[("
	if print {
		prefix = " [("
	}
	return prefix + name + ")" + giaMuxStructure(p, then, false, print, digits) + "|" + giaMuxStructure(p, otherwise, false, print, digits) + "]"
}
func Gia_MuxStructPrint_rec(p *Gia_Man_t, id, first int32) {
	fmt.Print(giaMuxStructure(p, id, first != 0, true, 0))
}
func Gia_MuxStructPrint(p *Gia_Man_t, id int32) {
	count := Gia_MuxDeref(p, id)
	Gia_MuxStructPrint_rec(p, id, 1)
	if Gia_MuxRef(p, id) != count {
		panic("gia: MUX print changed references")
	}
	fmt.Println()
}
func Gia_MuxStructDump_rec(p *Gia_Man_t, id, first int32, text *Vec_Str_t, digits int32) {
	value := giaMuxStructure(p, id, first != 0, false, int(digits))
	vec.Vec_StrPushBuffer(text, []byte(value), len(value))
}
func Gia_MuxStructDump(p *Gia_Man_t, id int32, text *Vec_Str_t, countDigits, idDigits int32) int32 {
	count := Gia_MuxDeref(p, id)
	vec.Vec_StrClear(text)
	vec.Vec_StrPrintNumStar(text, count, int(countDigits))
	Gia_MuxStructDump_rec(p, id, 1, text, idDigits)
	vec.Vec_StrPush(text, 0)
	if Gia_MuxRef(p, id) != count {
		panic("gia: MUX dump changed references")
	}
	return count
}
func Gia_ManMuxCompare(a, b *string) int32 { return int32(-strings.Compare(*a, *b)) }
func Gia_ManMuxCountOne(text string) int32 { return int32(strings.Count(text, "[")) }

type Mux_Man_t struct {
	pGia  *Gia_Man_t
	names []string
	ids   map[string]int
	tops  [][]int32
}

func Mux_ManAlloc(p *Gia_Man_t) *Mux_Man_t {
	return &Mux_Man_t{pGia: p, names: []string{""}, ids: make(map[string]int), tops: [][]int32{nil}}
}
func Mux_ManFree(p *Mux_Man_t) {
	if p != nil {
		*p = Mux_Man_t{}
	}
}
func Gia_ManMuxProfile(p *Mux_Man_t, width int32) int32 {
	counts := make([]int, 1000)
	for i := 1; i < len(p.tops); i++ {
		value := len(p.tops[i])
		if width == 0 {
			value, _ = strconv.Atoi(strings.Fields(p.names[i])[0])
		}
		counts[min(value, 999)]++
	}
	var distinct, total, maximum int
	for value, count := range counts {
		if count != 0 {
			distinct++
			total += count
			maximum = max(maximum, count)
			fmt.Printf(" %d=%d", value, count)
		}
	}
	if distinct == 0 {
		return 0
	}
	fmt.Printf("\nSummary: Max=%d Ave=%.2f\n", maximum, float64(total)/float64(distinct))
	return 1
}
func Gia_ManMuxProfiling(src *Gia_Man_t) {
	started := time.Now()
	p := Gia_ManDupMuxes(src, 2)
	defer Gia_ManStop(p)
	manager := Mux_ManAlloc(p)
	Gia_ManLevelNum(p)
	Gia_ManCreateRefs(p)
	for _, id := range Vec_IntArray(p.vCos) {
		Gia_ObjRefFanin0Inc(p, p.pObjs[id])
	}
	text := vec.Vec_StrAlloc(1000)
	fans := Gia_ManFirstFanouts(p)
	for _, obj := range p.pObjs {
		id := obj.id
		if Gia_ObjIsMuxId(p, id) == 0 {
			continue
		}
		if p.Refs[id] == 1 && Gia_ObjIsMuxId(p, Vec_IntEntry(fans, int(id))) != 0 {
			continue
		}
		if Gia_MuxStructDump(p, id, text, 3, int32(len(strconv.Itoa(int(p.nObjs))))) == 1 {
			continue
		}
		key := strings.TrimSuffix(string(vec.Vec_StrArray(text)), "\x00")
		index, found := manager.ids[key]
		if !found {
			index = len(manager.names)
			manager.ids[key] = index
			manager.names = append(manager.names, key)
			manager.tops = append(manager.tops, nil)
		}
		manager.tops[index] = append(manager.tops[index], id)
	}
	var total int
	for _, top := range manager.tops {
		total += len(top)
	}
	fmt.Printf("MUX structure profile for AIG %q: Total MUXes=%d Total trees=%d Unique trees=%d Time=%s\n", src.pName, Gia_ManMuxNum(p), total, len(manager.tops)-1, time.Since(started))
	if Gia_ManMuxProfile(manager, 0) != 0 {
		Gia_ManMuxProfile(manager, 1)
		for i := 1; i < min(len(manager.tops), 10); i++ {
			fmt.Printf("%5d: Occur=%4d %s\n", i, len(manager.tops[i]), manager.names[i])
		}
	}
}

// Object, incoming-edge and outgoing-edge classes used by the source's
// one-level TFI/TFO signatures. Complement classes remain reserved as in C.
type Gia_ObjType_t int32

const (
	C0 Gia_ObjType_t = iota
	PO
	PI
	FF
	XOR
	MUX
	AND
	iC0
	iC1
	iPI
	iFF
	iXOR
	iMUX
	iAND
	iANDn
	iANDp
	oPO
	oFF
	oXOR
	oMUXc
	oMUXd
	oAND
	oANDn
	oANDp
	GIA_END
)

var GIA_TYPE_STRINGS = []string{"C0", "PO", "PI", "FF", "XOR", "MUX", "AND", "iC0", "iC1", "iPI", "iFF", "iXOR", "iMUX", "iAND", "iANDn", "iANDp", "oPO", "oFF", "oXOR", "oMUXc", "oMUXd", "oAND", "oANDn", "oANDp", "GIA_END"}

func Gia_ManProfileStructuresTest(p *Gia_Man_t) {
	for i := Gia_ObjType_t(0); i < GIA_END; i++ {
		fmt.Printf("%d = %s\n", i, GIA_TYPE_STRINGS[i])
	}
}
func Gia_ManEncodeObj(p *Gia_Man_t, id int32) int32 {
	obj := p.pObjs[id]
	if Gia_ObjIsRi(p, obj) != 0 {
		panic("gia: RI has no standalone profile code")
	}
	switch {
	case Gia_ObjIsConst0(obj) != 0:
		return int32(C0)
	case Gia_ObjIsPo(p, obj) != 0:
		return int32(PO)
	case Gia_ObjIsPi(p, obj) != 0:
		return int32(PI)
	case Gia_ObjIsCi(obj) != 0:
		return int32(FF)
	case Gia_ObjIsXor(obj) != 0:
		return int32(XOR)
	case Gia_ObjIsMux(p, obj) != 0:
		return int32(MUX)
	default:
		return int32(AND)
	}
}
func Gia_ManEncodeFanin(p *Gia_Man_t, literal int32) int32 {
	obj := p.pObjs[literal>>1]
	switch {
	case Gia_ObjIsConst0(obj) != 0:
		return int32(iC0)
	case Gia_ObjIsPi(p, obj) != 0:
		return int32(iPI)
	case Gia_ObjIsCi(obj) != 0:
		return int32(iFF)
	case Gia_ObjIsXor(obj) != 0:
		return int32(iXOR)
	case Gia_ObjIsMux(p, obj) != 0:
		return int32(iMUX)
	default:
		return int32(iAND)
	}
}
func Gia_ManEncodeFanout(p *Gia_Man_t, obj *Gia_Obj_t, index int32) Gia_ObjType_t {
	switch {
	case Gia_ObjIsPo(p, obj) != 0:
		return oPO
	case Gia_ObjIsCo(obj) != 0:
		return oFF
	case Gia_ObjIsXor(obj) != 0:
		return oXOR
	case Gia_ObjIsMux(p, obj) != 0:
		if index == 2 {
			return oMUXc
		}
		return oMUXd
	default:
		return oAND
	}
}
func Gia_ManProfileCollect(p *Gia_Man_t, id int32, code, offsets, array *Vec_Int_t) {
	Vec_IntClear(array)
	for _, value := range Vec_IntArray(code)[Vec_IntEntry(offsets, int(id)):Vec_IntEntry(offsets, int(id)+1)] {
		Vec_IntPush(array, value)
	}
}
func Gia_ManProfilePrintOne(p *Gia_Man_t, id int32, array *Vec_Int_t) {
	obj := p.pObjs[id]
	if Gia_ObjIsRi(p, obj) != 0 {
		return
	}
	fanins := int(Gia_ObjFaninNum(p, obj))
	if Gia_ObjIsRo(p, obj) != 0 {
		fanins = 1
	}
	values := Vec_IntArray(array)
	if len(values) == 0 {
		return
	}
	fmt.Printf("%6d: ", id)
	for i := 0; i < 3; i++ {
		name := ""
		if i < fanins {
			name = GIA_TYPE_STRINGS[values[1+i]]
		}
		fmt.Printf(" %5s", name)
	}
	fmt.Printf(" -> %5s ->", GIA_TYPE_STRINGS[values[0]])
	for i := 1 + fanins; i < len(values); {
		end := i + 1
		for end < len(values) && values[end] == values[i] {
			end++
		}
		fmt.Printf(" %d x %s", end-i, GIA_TYPE_STRINGS[values[i]])
		i = end
	}
	fmt.Println()
}
func giaProfileKey(values []int32) string {
	data := make([]byte, 4*len(values))
	for i, value := range values {
		binary.LittleEndian.PutUint32(data[i*4:], uint32(value))
	}
	return string(data)
}
func Gia_ManProfileHash(p *Gia_Man_t, code, offsets *Vec_Int_t) *Vec_Int_t {
	result := Vec_IntAlloc(int(p.nObjs))
	ids := map[string]int32{"": 0}
	for id := int32(0); id < p.nObjs; id++ {
		key := giaProfileKey(Vec_IntArray(code)[Vec_IntEntry(offsets, int(id)):Vec_IntEntry(offsets, int(id)+1)])
		index, found := ids[key]
		if !found {
			index = int32(len(ids))
			ids[key] = index
		}
		Vec_IntPush(result, index)
	}
	return result
}
func giaProfileCodes(p *Gia_Man_t) (*Vec_Int_t, *Vec_Int_t) {
	code, offsets := Vec_IntAlloc(5*int(p.nObjs)), Vec_IntAlloc(int(p.nObjs)+1)
	for _, obj := range p.pObjs {
		Vec_IntPush(offsets, int32(Vec_IntSize(code)))
		if Gia_ObjIsRi(p, obj) != 0 {
			continue
		}
		Vec_IntPush(code, Gia_ManEncodeObj(p, obj.id))
		a, b, c, kind := p.Gate(obj.id)
		switch kind {
		case 4:
			Vec_IntPush(code, Gia_ManEncodeFanin(p, a&^1))
			if a&1 != 0 {
				b, c = c, b
			}
			Vec_IntPush(code, Gia_ManEncodeFanin(p, b))
			Vec_IntPush(code, Gia_ManEncodeFanin(p, c))
		case 2, 3, 5:
			x, y := Gia_ManEncodeFanin(p, a), Gia_ManEncodeFanin(p, b)
			Vec_IntPush(code, min(x, y))
			Vec_IntPush(code, max(x, y))
		case 6:
			Vec_IntPush(code, Gia_ManEncodeFanin(p, a))
		case 1:
			if Gia_ObjIsRo(p, obj) != 0 {
				Vec_IntPush(code, Gia_ManEncodeFanin(p, Gia_ObjFaninLit0p(p, Gia_ObjRoToRi(p, obj))))
			}
		}
		var fanouts []int32
		for i := int32(0); i < Gia_ObjFanoutNum(p, obj); i++ {
			fanout := Gia_ObjFanout(p, obj, i)
			fanouts = append(fanouts, int32(Gia_ManEncodeFanout(p, fanout, Gia_ObjWhatFanin(p, fanout, obj))))
		}
		sort.Slice(fanouts, func(i, j int) bool { return fanouts[i] < fanouts[j] })
		for _, value := range fanouts {
			Vec_IntPush(code, value)
		}
	}
	Vec_IntPush(offsets, int32(Vec_IntSize(code)))
	return code, offsets
}
func Gia_ManProfileStructuresInt(p *Gia_Man_t, limit, verbose int32) {
	if p.PMuxes == nil {
		panic("gia: profiling requires real-gate manager")
	}
	Gia_ManStaticFanoutStart(p)
	defer Gia_ManStaticFanoutStop(p)
	code, offsets := giaProfileCodes(p)
	array := Vec_IntAlloc(100)
	if verbose != 0 {
		for _, obj := range p.pObjs {
			Gia_ManProfileCollect(p, obj.id, code, offsets, array)
			Gia_ManProfilePrintOne(p, obj.id, array)
		}
	}
	classes := Gia_ManProfileHash(p, code, offsets)
	type entry struct{ id, count, first int32 }
	groups := make([]entry, int(Vec_IntFindMax(classes))+1)
	for id, class := range Vec_IntArray(classes) {
		if Gia_ObjIsRi(p, p.pObjs[id]) != 0 {
			continue
		}
		if groups[class].count == 0 {
			groups[class].first = int32(id)
			groups[class].id = class
		}
		groups[class].count++
	}
	sort.SliceStable(groups, func(i, j int) bool { return groups[i].count > groups[j].count })
	fmt.Printf("Showing TFI/node/TFO structures that appear at least %d times.\n", limit)
	for _, group := range groups {
		if group.count < limit || group.count == 0 {
			break
		}
		fmt.Printf("Class %d: weight=%d first obj=", group.id, group.count)
		Gia_ManProfileCollect(p, group.first, code, offsets, array)
		Gia_ManProfilePrintOne(p, group.first, array)
	}
}
func Gia_ManProfileStructures(p *Gia_Man_t, limit, verbose int32) {
	if p.PMuxes != nil {
		Gia_ManProfileStructuresInt(p, limit, verbose)
		return
	}
	coarse := Gia_ManDupMuxes(p, 2)
	defer Gia_ManStop(coarse)
	Gia_ManProfileStructuresInt(coarse, limit, verbose)
}

func Gia_ManMarkTfi_rec(p *Gia_Man_t, obj *Gia_Obj_t) {
	stack := []int32{obj.id}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
			continue
		}
		Gia_ObjSetTravIdCurrentId(p, id)
		obj := p.pObjs[id]
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		stack = append(stack, giaObjectFanins(p, obj)...)
	}
}
func Gia_ManFindSharedInputs(p *Gia_Man_t) *Vec_Int_t {
	counts := make([]int32, Gia_ManCiNum(p))
	for _, id := range Vec_IntArray(p.vCos) {
		Gia_ManIncrementTravId(p)
		Gia_ManMarkTfi_rec(p, Gia_ObjFanin0(p.pObjs[id]))
		for i, ci := range Vec_IntArray(p.vCis) {
			if Gia_ObjIsTravIdCurrentId(p, ci) != 0 {
				counts[i]++
			}
		}
	}
	result := Vec_IntAlloc(len(counts))
	for i, count := range counts {
		if count == Gia_ManCoNum(p) {
			Vec_IntPush(result, int32(i))
		}
	}
	if Vec_IntSize(result) == 0 || Vec_IntSize(result) > 10 {
		return nil
	}
	return result
}
func Gia_ManFindCofs(src *Gia_Man_t, inputs *Vec_Int_t, out **Gia_Man_t) *Vec_Wec_t {
	if Vec_IntSize(inputs) >= int(Gia_ManCiNum(src)) || Vec_IntSize(inputs) > 10 {
		panic("gia: invalid cofactor input count")
	}
	count := 1 << uint(Vec_IntSize(inputs))
	cofactors := vec.Vec_WecStart(count)
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	Gia_ManHashAlloc(dst)
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	for _, id := range Vec_IntArray(src.vCis) {
		src.pObjs[id].Value = uint32(Gia_ManAppendCi(dst))
	}
	for assignment := 0; assignment < count; assignment++ {
		for i, index := range Vec_IntArray(inputs) {
			Gia_ManCi(src, index).Value = uint32(assignment >> uint(i) & 1)
		}
		for _, obj := range src.pObjs {
			a, b, c, kind := src.Gate(obj.id)
			var lit int32
			switch kind {
			case 2, 5:
				lit = Gia_ManHashAnd(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
			case 3:
				lit = Gia_ManHashXor(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
			case 4:
				lit = Gia_ManHashMux(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b), Gia_ObjLitCopy(src, c))
			default:
				continue
			}
			obj.Value = uint32(lit)
		}
		layer := vec.Vec_WecEntry(cofactors, assignment)
		for _, id := range Vec_IntArray(src.vCos) {
			Vec_IntPush(layer, Gia_ObjFanin0Copy(src.pObjs[id]))
		}
	}
	Gia_ManSetRegNum(dst, src.nRegs)
	if out != nil {
		*out = dst
	}
	return cofactors
}
func Gia_ManFindEquivClasses(cofactors *Vec_Wec_t) *Vec_Int_t {
	result := Vec_IntAlloc(vec.Vec_WecSize(cofactors))
	ids := make(map[string]int32)
	for i := 0; i < vec.Vec_WecSize(cofactors); i++ {
		key := giaProfileKey(Vec_IntArray(vec.Vec_WecEntry(cofactors, i)))
		id, found := ids[key]
		if !found {
			id = int32(len(ids))
			ids[key] = id
		}
		Vec_IntPush(result, id)
	}
	return result
}
func Gia_ManFindMuxTree_rec(p *Gia_Man_t, controls []int32, count int32, data *Vec_Int_t, shift int32) int32 {
	if count == 0 {
		return Vec_IntEntry(data, int(shift))
	}
	count--
	a := Gia_ManFindMuxTree_rec(p, controls, count, data, shift)
	b := Gia_ManFindMuxTree_rec(p, controls, count, data, shift+(1<<uint(count)))
	return Gia_ManHashMux(p, controls[count], b, a)
}
func Gia_ManFindDerive(p *Gia_Man_t, outputs int32, inputs *Vec_Int_t, cofactors *Vec_Wec_t, classes *Vec_Int_t) {
	values := Vec_IntFindMax(classes) + 1
	count := int32(bits.Len32(uint32(values - 1)))
	if count >= int32(Vec_IntSize(inputs)) || Vec_IntSize(classes) != vec.Vec_WecSize(cofactors) || Vec_IntSize(classes) != 1<<uint(Vec_IntSize(inputs)) {
		panic("gia: invalid cofactor compression dimensions")
	}
	leaves := make([]int32, Vec_IntSize(inputs))
	for i, index := range Vec_IntArray(inputs) {
		leaves[i] = Gia_ObjToLit(p, Gia_ManCi(p, index))
	}
	controls := make([]int32, count)
	var memory []uint32
	for bit := int32(0); bit < count; bit++ {
		truth := make([]uint32, max(1, (Vec_IntSize(classes)+31)/32))
		for assignment, class := range Vec_IntArray(classes) {
			if class>>uint(bit)&1 != 0 {
				truth[assignment>>5] |= 1 << uint(assignment&31)
			}
		}
		// Kit's <=5-variable truth convention repeats the low pattern to 32 bits.
		if len(leaves) < 5 {
			period := 1 << uint(len(leaves))
			for i := period; i < 32; i++ {
				if truth[0]>>uint(i%period)&1 != 0 {
					truth[0] |= 1 << uint(i)
				}
			}
		}
		controls[bit] = kit.Kit_TruthToGia(p, truth, int32(len(leaves)), &memory, leaves, 1)
	}
	first := make([]int, values)
	for i := range first {
		first[i] = -1
	}
	for assignment, class := range Vec_IntArray(classes) {
		if first[class] < 0 {
			first[class] = assignment
		}
	}
	data := Vec_IntAlloc(1 << uint(count))
	for output := int32(0); output < outputs; output++ {
		Vec_IntClear(data)
		for _, assignment := range first {
			Vec_IntPush(data, Vec_IntEntry(vec.Vec_WecEntry(cofactors, assignment), int(output)))
		}
		for Vec_IntSize(data) < 1<<uint(count) {
			Vec_IntPush(data, 0)
		}
		Gia_ManAppendCo(p, Gia_ManFindMuxTree_rec(p, controls, count, data, 0))
	}
}
func Gia_ManCofStructure(src *Gia_Man_t) *Gia_Man_t {
	inputs := Gia_ManFindSharedInputs(src)
	if inputs == nil {
		return nil
	}
	// If all inputs are selected there is no nontrivial decomposition boundary.
	if Vec_IntSize(inputs) >= int(Gia_ManCiNum(src)) {
		return nil
	}
	var dst *Gia_Man_t
	cofactors := Gia_ManFindCofs(src, inputs, &dst)
	classes := Gia_ManFindEquivClasses(cofactors)
	if bits.Len32(uint32(Vec_IntFindMax(classes))) >= Vec_IntSize(inputs) {
		Gia_ManStop(dst)
		return nil
	}
	Gia_ManFindDerive(dst, Gia_ManCoNum(src), inputs, cofactors, classes)
	Gia_ManHashStop(dst)
	giaCopyInterfaceMetadata(dst, src)
	result := Gia_ManCleanup(dst)
	Gia_ManStop(dst)
	return result
}
