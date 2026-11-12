package if_

import (
	"github.com/glycerine/abcgo/misc/vec"
	"strconv"
)

// Gia_ManFromIfGetConfig appends one packed programmable-cell configuration.
// A DSD configuration starts with a nibble permutation followed by truth bits.
func Gia_ManFromIfGetConfig(configs *Vec_Int_t, p *If_Man_t, cut *If_Cut_t, result int32, text *Vec_Str_t) {
	if p == nil || p.pIfDsdMan == nil || Vec_IntSize(configs) < 2 {
		panic("if: missing cell configuration")
	}
	manager := p.pIfDsdMan
	config := If_DsdManGetFuncConfig(manager, If_CutDsdLit(p, cut))
	n := If_DsdManVarNum(manager)
	bits := If_DsdManTtBitNum(manager)
	permBits := If_DsdManPermBitNum(manager)
	size := Vec_IntEntry(configs, 1)
	if n <= 0 || n > 16 || bits < 0 || len(config) < 1+int((bits+63)/64) || size <= 0 || bits+permBits >= 32*size {
		panic("if: invalid cell configuration dimensions")
	}
	perInput := permBits / n
	perm := If_CutDsdPerm(p, cut)
	packed := make([]uint32, size)
	set := func(bit int32) { packed[bit>>5] |= uint32(1) << uint(bit&31) }
	for bit := int32(0); bit < bits; bit++ {
		if config[1+bit/64]>>uint(bit&63)&1 != 0 {
			set(bit)
		}
	}
	for v := int32(0); v < n; v++ {
		dsdVar := int32(config[0] >> uint(v*4) & 15)
		if dsdVar >= int32(cut.nLeaves) {
			panic("if: cell permutation outside cut")
		}
		lit := int32(perm[dsdVar])
		pin := lit >> 1
		if pin >= int32(cut.nLeaves) {
			panic("if: cell pin outside cut")
		}
		lit ^= If_CutLeaf(p, cut, pin).iCopy & 1
		for bit := int32(0); bit < perInput; bit++ {
			if lit>>uint(bit)&1 != 0 {
				set(bits + v*perInput + bit)
			}
		}
	}
	if (If_CutDsdLit(p, cut)&1)^int32(cut.fCompl)^(result&1) != 0 {
		set(bits + permBits)
	}
	for _, value := range packed {
		Vec_IntPush(configs, int32(value))
	}
	vec.Vec_IntAddToEntry(configs, 0, 1)
	if text == nil {
		return
	}
	appendText := func(s string) {
		for i := range s {
			vec.Vec_StrPush(text, s[i])
		}
	}
	bitChar := func(bit int32) byte { return '0' + byte(packed[bit>>5]>>uint(bit&31)&1) }
	appendText(strconv.Itoa(int(result>>1)) + " ")
	for i := int32(0); i < bits; i++ {
		vec.Vec_StrPush(text, bitChar(i))
	}
	appendText("  ")
	for v := int32(0); v < n; v++ {
		for i := int32(0); i < perInput; i++ {
			vec.Vec_StrPush(text, bitChar(bits+v*perInput+i))
			if i == 0 {
				appendText(" ")
			}
		}
		appendText("  ")
	}
	vec.Vec_StrPush(text, bitChar(bits+permBits))
	appendText("\n")
}
