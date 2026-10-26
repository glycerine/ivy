package abc

// Structural mux recognition from aig/aig/aigUtil.c.

type AigMuxInputs struct {
	Control             AigID
	WhenTrue, WhenFalse AigLit // function of the regular root, including data complements
}

func aigRecognizeMux(g *AigManager, id AigID) (AigMuxInputs, bool) {
	n := g.objects[id]
	if n.Type != AigAND || !n.Fanin0.Complemented() || !n.Fanin1.Complemented() {
		return AigMuxInputs{}, false
	}
	a, b := g.objects[n.Fanin0.ID()], g.objects[n.Fanin1.ID()]
	if a.Type != AigAND || b.Type != AigAND {
		return AigMuxInputs{}, false
	}
	left, right := [2]AigLit{a.Fanin0, a.Fanin1}, [2]AigLit{b.Fanin0, b.Fanin1}
	for _, pair := range [][2]int{{1, 1}, {0, 0}, {0, 1}, {1, 0}} {
		i, j := pair[0], pair[1]
		if left[i] != right[j].Not() {
			continue
		}
		t, e := left[1-i].Not(), right[1-j].Not()
		if left[i].Complemented() {
			t, e = e, t
		}
		return AigMuxInputs{Control: left[i].ID(), WhenTrue: t, WhenFalse: e}, true
	}
	return AigMuxInputs{}, false
}

// RecognizeMux preserves the native common-grandchild matching priority. XOR
// structures are also mux types. Non-mux objects return false without error.
func (g *AigManager) RecognizeMux(id AigID) (AigMuxInputs, bool, error) {
	if g == nil {
		return AigMuxInputs{}, false, ErrDimension
	}
	if e := g.Check(); e != nil {
		return AigMuxInputs{}, false, e
	}
	if _, e := g.Object(id); e != nil {
		return AigMuxInputs{}, false, e
	}
	mux, ok := aigRecognizeMux(g, id)
	return mux, ok, nil
}
