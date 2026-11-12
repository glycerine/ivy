package amap

import "fmt"

// Vec_IntCheckWithMask finds a packed rule by its low 16-bit fanin literal.
// The high 16 bits contain the resulting node ID.
func Vec_IntCheckWithMask(v *Vec_Int_t, entry int32) int32 {
	if v == nil {
		return -1
	}
	for _, packed := range Vec_IntArray(v) {
		if packed&0xffff == entry&0xffff {
			return packed >> 16
		}
	}
	return -1
}

// Vec_IntPushOrderWithMask keeps a packed rule vector ordered by the fanin
// literal, independently of the high bits carrying the result node ID.
func Vec_IntPushOrderWithMask(v *Vec_Int_t, entry int32) {
	if v == nil {
		panic("amap: nil rule vector")
	}
	items := Vec_IntArray(v)
	key := entry & 0xffff
	index := 0
	for index < len(items) && items[index]&0xffff <= key {
		index++
	}
	vecIntInsert(v, index, entry)
}

func Amap_LibFindNode(lib *Amap_Lib_t, fan0, fan1 int32, xor int32) int32 {
	if lib == nil {
		return -1
	}
	rules := lib.vRules
	if xor != 0 {
		rules = lib.vRulesX
	}
	if rules == nil || fan0 < 0 || fan0 >= Vec_PtrSize(rules) {
		return -1
	}
	v, ok := Vec_PtrEntry(rules, fan0).(*Vec_Int_t)
	if !ok {
		return -1
	}
	return Vec_IntCheckWithMask(v, fan1)
}

func Amap_LibFindMux(lib *Amap_Lib_t, fan0, fan1, fan2 int32) int32 {
	if lib == nil || lib.vRules3 == nil {
		return -1
	}
	for i := int32(0); i+3 < Vec_IntSize(lib.vRules3); i += 4 {
		if Vec_IntEntry(lib.vRules3, i) == fan0 && Vec_IntEntry(lib.vRules3, i+1) == fan1 && Vec_IntEntry(lib.vRules3, i+2) == fan2 {
			return Vec_IntEntry(lib.vRules3, i+3)
		}
	}
	return -1
}

func Amap_LibCreateObj(lib *Amap_Lib_t) *Amap_Nod_t {
	if lib == nil {
		panic("amap: nil library")
	}
	if lib.nNodes == lib.nNodesAlloc {
		capacity := max(256, lib.nNodesAlloc*2)
		grown := make([]Amap_Nod_t, lib.nNodes, capacity)
		copy(grown, lib.pNodes)
		lib.pNodes = grown
		lib.nNodesAlloc = int32(capacity)
	}
	if int(lib.nNodes) >= len(lib.pNodes) {
		lib.pNodes = append(lib.pNodes, Amap_Nod_t{})
	} else {
		lib.pNodes[lib.nNodes] = Amap_Nod_t{}
	}
	node := &lib.pNodes[lib.nNodes]
	node.Id = uint32(lib.nNodes)
	lib.nNodes++
	if lib.vRules == nil {
		lib.vRules = Vec_PtrAlloc(16)
	}
	if lib.vRulesX == nil {
		lib.vRulesX = Vec_PtrAlloc(16)
	}
	Vec_PtrPush(lib.vRules, Vec_IntAlloc(8))
	Vec_PtrPush(lib.vRules, Vec_IntAlloc(8))
	Vec_PtrPush(lib.vRulesX, Vec_IntAlloc(8))
	Vec_PtrPush(lib.vRulesX, Vec_IntAlloc(8))
	return node
}

func Amap_LibCreateVar(lib *Amap_Lib_t) int32 {
	assert(lib != nil && len(lib.pNodes) == 0 && lib.nNodes == 0, "empty library node table")
	lib.nNodesAlloc = 256
	lib.pNodes = make([]Amap_Nod_t, 0, int(lib.nNodesAlloc))
	node := Amap_LibCreateObj(lib)
	node.Type = AMAP_OBJ_PI
	node.nSuppSize = 1
	return 0
}

func Amap_LibCreateNode(lib *Amap_Lib_t, fan0, fan1 int32, xor int32) int32 {
	if fan0 < fan1 {
		fan0, fan1 = fan1, fan0
	}
	node := Amap_LibCreateObj(lib)
	if xor != 0 {
		node.Type = AMAP_OBJ_XOR
	} else {
		node.Type = AMAP_OBJ_AND
	}
	node.nSuppSize = Amap_LibNod(lib, Abc_Lit2Var(fan0)).nSuppSize + Amap_LibNod(lib, Abc_Lit2Var(fan1)).nSuppSize
	node.iFan0, node.iFan1 = int16(fan0), int16(fan1)
	if lib.fVerbose != 0 {
		op := ' '
		if xor != 0 {
			op = 'x'
		}
		fmt.Printf("Creating node %5d %c : iFan0 = %5d%c iFan1 = %5d%c\n", node.Id, op,
			Abc_Lit2Var(fan0), polarity(Abc_LitIsCompl(fan0)), Abc_Lit2Var(fan1), polarity(Abc_LitIsCompl(fan1)))
	}
	rules := lib.vRules
	if xor != 0 {
		rules = lib.vRulesX
	}
	insertPackedRule(rules, fan0, fan1, int32(node.Id))
	if fan0 != fan1 {
		insertPackedRule(rules, fan1, fan0, int32(node.Id))
	}
	return int32(node.Id)
}

func insertPackedRule(rules *Vec_Ptr_t, index, key, node int32) {
	if index < 0 || index >= Vec_PtrSize(rules) {
		panic(fmt.Sprintf("amap: rule fanin literal %d outside table of %d", index, Vec_PtrSize(rules)))
	}
	v, ok := Vec_PtrEntry(rules, index).(*Vec_Int_t)
	if !ok || node > 0xffff || key < 0 || key > 0xffff {
		panic("amap: invalid packed rule")
	}
	Vec_IntPushOrderWithMask(v, node<<16|key)
}

func polarity(complement int32) byte {
	if complement != 0 {
		return '-'
	}
	return '+'
}

func Amap_LibCreateMux(lib *Amap_Lib_t, fan0, fan1, fan2 int32) int32 {
	node := Amap_LibCreateObj(lib)
	node.Type = AMAP_OBJ_MUX
	node.nSuppSize = Amap_LibNod(lib, Abc_Lit2Var(fan0)).nSuppSize +
		Amap_LibNod(lib, Abc_Lit2Var(fan1)).nSuppSize + Amap_LibNod(lib, Abc_Lit2Var(fan2)).nSuppSize
	node.iFan0, node.iFan1, node.iFan2 = int16(fan0), int16(fan1), int16(fan2)
	if lib.fVerbose != 0 {
		fmt.Printf("Creating node %5d m : iFan0 = %5d%c iFan1 = %5d%c iFan2 = %5d%c\n", node.Id,
			Abc_Lit2Var(fan0), polarity(Abc_LitIsCompl(fan0)), Abc_Lit2Var(fan1), polarity(Abc_LitIsCompl(fan1)),
			Abc_Lit2Var(fan2), polarity(Abc_LitIsCompl(fan2)))
	}
	if lib.vRules3 == nil {
		lib.vRules3 = Vec_IntAlloc(16)
	}
	for _, value := range []int32{fan0, fan1, fan2, int32(node.Id)} {
		Vec_IntPush(lib.vRules3, value)
	}
	return int32(node.Id)
}

// Amap_LibLookupTableAlloc copies packed vectors into native slices and adds
// the zero terminator expected by the tight matcher loops.
func Amap_LibLookupTableAlloc(vectors *Vec_Ptr_t, verbose int32) [][]int32 {
	if vectors == nil {
		return nil
	}
	result := make([][]int32, Vec_PtrSize(vectors))
	for i := int32(0); i < Vec_PtrSize(vectors); i++ {
		entries, ok := Vec_PtrEntry(vectors, i).(*Vec_Int_t)
		if !ok || entries == nil {
			continue
		}
		result[i] = append(append([]int32(nil), Vec_IntArray(entries)...), 0)
		if verbose != 0 {
			fmt.Printf("%d : ", i)
			for _, value := range Vec_IntArray(entries) {
				fmt.Printf("%d(%d) ", value&0xffff, value>>16)
			}
			fmt.Println()
		}
	}
	return result
}

func vecIntInsert(v *Vec_Int_t, index int, value int32) {
	items := Vec_IntArray(v)
	if index < 0 || index > len(items) {
		panic("amap: invalid rule insertion index")
	}
	Vec_IntPush(v, 0)
	items = Vec_IntArray(v)
	copy(items[index+1:], items[index:len(items)-1])
	items[index] = value
}
