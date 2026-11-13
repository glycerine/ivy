package ssw

import "github.com/glycerine/abcgo/aig/aig"

func sswFrameWords(p *Ssw_Sml_t, obj aig.Edge, frame int32) []uint32 {
	if obj.Complemented() || frame < 0 || frame >= p.nFrames {
		panic("ssw: invalid simulation frame or edge")
	}
	start := frame * p.nWordsFrame
	return Ssw_ObjSim(p, int32(obj.NodeID()))[start : start+p.nWordsFrame]
}
func Ssw_SmlNodesCompareInFrame(p *Ssw_Sml_t, a, b aig.Edge, af, bf int32) int32 {
	x, y := sswFrameWords(p, a, af), sswFrameWords(p, b, bf)
	for i := range x {
		if x[i] != y[i] {
			return 0
		}
	}
	return 1
}
func Ssw_SmlNodeCopyFanin(p *Ssw_Sml_t, obj aig.Edge, frame int32) {
	node := p.pAig.Node(obj)
	if node.Kind != aig.Output {
		panic("ssw: expected output")
	}
	fanin := node.Fanins[0]
	src := sswFrameWords(p, fanin&^1, frame)
	dst := sswFrameWords(p, obj, frame)
	invert := p.pAig.Node(fanin).Phase != fanin.Complemented()
	for i, w := range src {
		if invert {
			w = ^w
		}
		dst[i] = w
	}
}
func Ssw_SmlNodeTransferNext(p *Ssw_Sml_t, out, in aig.Edge, frame int32) {
	if p.pAig.Node(out).Kind != aig.Output || p.pAig.Node(in).Kind != aig.Input {
		panic("ssw: expected output and input")
	}
	copy(sswFrameWords(p, in, frame+1), sswFrameWords(p, out, frame))
}
func Ssw_SmlNodeTransferFirst(p *Ssw_Sml_t, out, in aig.Edge) {
	if p.pAig.Node(out).Kind != aig.Output || p.pAig.Node(in).Kind != aig.Input || p.nFrames < 2 {
		panic("ssw: expected output, input, and at least two frames")
	}
	copy(sswFrameWords(p, in, 0), sswFrameWords(p, out, p.nFrames-1))
}

// Ssw_SmlNodeSimulate stores each node's simulation relative to its phase.
func Ssw_SmlNodeSimulate(p *Ssw_Sml_t, obj aig.Edge, frame int32) {
	node := p.pAig.Node(obj)
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer {
		panic("ssw: expected internal node")
	}
	a := node.Fanins[0]
	x := sswFrameWords(p, a&^1, frame)
	dst := sswFrameWords(p, obj, frame)
	invertA := p.pAig.Node(a).Phase != a.Complemented()
	var y []uint32
	var invertB bool
	if node.Kind != aig.Buffer {
		b := node.Fanins[1]
		y = sswFrameWords(p, b&^1, frame)
		invertB = p.pAig.Node(b).Phase != b.Complemented()
	}
	for i := range dst {
		value := x[i]
		if invertA {
			value = ^value
		}
		if node.Kind != aig.Buffer {
			right := y[i]
			if invertB {
				right = ^right
			}
			if node.Kind == aig.ExclusiveOr {
				value ^= right
			} else {
				value &= right
			}
		}
		if node.Phase {
			value = ^value
		}
		dst[i] = value
	}
}
