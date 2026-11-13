package ssw

import "github.com/glycerine/abcgo/aig/aig"

func sswInputFrame(p *Ssw_Sml_t, obj aig.Edge, frame int32) []uint32 {
	if p.pAig.Node(obj).Kind != aig.Input {
		panic("ssw: expected input")
	}
	return sswFrameWords(p, obj, frame)
}
func Ssw_SmlAssignRandom(p *Ssw_Sml_t, obj aig.Edge) {
	if p.pAig.Node(obj).Kind != aig.Input || obj.Complemented() {
		panic("ssw: expected regular input")
	}
	words := Ssw_ObjSim(p, int32(obj.NodeID()))
	for i := range words {
		words[i] = Ssw_ObjRandomSim()
	}
	for frame := int32(0); frame < p.nFrames; frame++ {
		words[p.nWordsFrame*frame] <<= 1
	}
}
func Ssw_SmlAssignRandomFrame(p *Ssw_Sml_t, obj aig.Edge, frame int32) {
	words := sswInputFrame(p, obj, frame)
	for i := range words {
		words[i] = Ssw_ObjRandomSim()
	}
}
func Ssw_SmlObjAssignConst(p *Ssw_Sml_t, obj aig.Edge, value, frame int32) {
	words := sswInputFrame(p, obj, frame)
	var word uint32
	if value != 0 {
		word = ^word
	}
	for i := range words {
		words[i] = word
	}
}
func Ssw_SmlObjAssignConstWord(p *Ssw_Sml_t, obj aig.Edge, value, frame, index int32) {
	var word uint32
	if value != 0 {
		word = ^word
	}
	sswInputFrame(p, obj, frame)[index] = word
}
func Ssw_SmlObjSetWord(p *Ssw_Sml_t, obj aig.Edge, word uint32, index, frame int32) {
	sswInputFrame(p, obj, frame)[index] = word
}
