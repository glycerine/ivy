package sim

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/bits"
	"math/rand/v2"
)

func Sim_UtilInfoAlloc(size, words int, _ bool) Info {
	if size < 0 || words < 0 {
		panic("negative simulation dimensions")
	}
	data := make([]uint32, size*words)
	r := make(Info, size)
	for i := range r {
		r[i] = data[i*words : (i+1)*words : (i+1)*words]
	}
	return r
}
func Sim_UtilInfoFree(p *Info) { *p = nil }
func Sim_UtilInfoAdd(a, b []uint32, n int) {
	for i := 0; i < n; i++ {
		a[i] |= b[i]
	}
}
func positions(a, b []uint32, n int, news bool) []int {
	var r []int
	for i := 0; i < n; i++ {
		v := a[i] ^ b[i]
		if news {
			v = b[i] &^ a[i]
		}
		for v != 0 {
			j := bits.TrailingZeros32(v)
			r = append(r, 32*i+j)
			v &= v - 1
		}
	}
	return r
}
func Sim_UtilInfoDetectDiffs(a, b []uint32, n int) []int { return positions(a, b, n, false) }
func Sim_UtilInfoDetectNews(a, b []uint32, n int) []int  { return positions(a, b, n, true) }
func Sim_UtilCountOnes(p []uint32, n int) int {
	c := 0
	for _, x := range p[:n] {
		c += bits.OnesCount32(x)
	}
	return c
}
func Sim_UtilCountOnesArray(p Info, n int) []int {
	r := make([]int, len(p))
	for i := range p {
		r[i] = Sim_UtilCountOnes(p[i], n)
	}
	return r
}
func Sim_UtilSetRandom(p []uint32, n int) {
	for i := range p[:n] {
		p[i] = rand.Uint32()
	}
}
func Sim_UtilSetCompl(p []uint32, n int) {
	for i := 0; i < n; i++ {
		p[i] = ^p[i]
	}
}
func Sim_UtilSetConst(p []uint32, n int, one bool) {
	var v uint32
	if one {
		v = ^uint32(0)
	}
	for i := 0; i < n; i++ {
		p[i] = v
	}
}
func Sim_UtilInfoIsEqual(a, b []uint32, n int) bool {
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func Sim_UtilInfoIsImp(a, b []uint32, n int) bool {
	for i := 0; i < n; i++ {
		if a[i]&^b[i] != 0 {
			return false
		}
	}
	return true
}
func Sim_UtilInfoIsClause(a, b []uint32, n int) bool {
	for i := 0; i < n; i++ {
		if ^(a[i] | b[i]) != 0 {
			return false
		}
	}
	return true
}
func edgeWord(info Info, e aig.Edge, w int) uint32 {
	v := info[e.NodeID()][w]
	if e.Complemented() {
		v = ^v
	}
	return v
}
func simulateNode(g *aig.Graph, e aig.Edge, dst, left, right Info, n, off int) {
	node := g.Node(e)
	if node.Kind == aig.Input {
		return
	}
	for w := off; w < off+n; w++ {
		var v uint32
		switch node.Kind {
		case aig.Constant:
			v = ^uint32(0)
		case aig.And:
			v = edgeWord(left, node.Fanins[0], w) & edgeWord(right, node.Fanins[1], w)
		case aig.ExclusiveOr:
			v = edgeWord(left, node.Fanins[0], w) ^ edgeWord(right, node.Fanins[1], w)
		case aig.Buffer, aig.Output:
			v = edgeWord(left, node.Fanins[0], w)
		default:
			panic("unsupported graph node")
		}
		dst[e.NodeID()][w] = v
	}
}
func Sim_UtilSimulateNodeOne(g *aig.Graph, e aig.Edge, info Info, n, off int) {
	simulateNode(g, e, info, info, info, n, off)
}
func Sim_UtilTransferNodeOne(g *aig.Graph, e aig.Edge, info Info, n, off int, shift bool) {
	node := g.Node(e)
	if node.Kind != aig.Output && node.Kind != aig.Buffer {
		panic("transfer requires unary node")
	}
	to := off
	if shift {
		to += n
	}
	for w := 0; w < n; w++ {
		info[e.NodeID()][to+w] = edgeWord(info, node.Fanins[0], off+w)
	}
}
func Sim_UtilSimulate(p *Sim_Man_t, second bool) {
	info := p.Sim0
	if second {
		info = p.Sim1
	}
	for _, e := range p.Order {
		Sim_UtilSimulateNodeOne(p.Graph, e, info, p.Words, 0)
	}
}
func Sim_UtilSimulateNode(p *Sim_Man_t, e aig.Edge, dst, left, right bool) {
	d, a, b := p.Sim0, p.Sim0, p.Sim0
	if dst {
		d = p.Sim1
	}
	if left {
		a = p.Sim1
	}
	if right {
		b = p.Sim1
	}
	simulateNode(p.Graph, e, d, a, b, p.Words, 0)
}
func Sim_UtilInfoFlip(p *Sim_Man_t, e aig.Edge) {
	for i, v := range p.Sim0[e.NodeID()] {
		p.Sim1[e.NodeID()][i] = ^v
	}
}
func Sim_UtilInfoCompare(p *Sim_Man_t, e aig.Edge) bool {
	return Sim_UtilInfoIsEqual(p.Sim0[e.NodeID()], p.Sim1[e.NodeID()], p.Words)
}
func Sim_UtilCountSuppSizes(p *Sim_Man_t, structural bool) int {
	if structural && p.Structural == nil {
		return 0
	}
	if !structural {
		return sum(Sim_UtilCountOnesArray(p.Functional, SIM_NUM_WORDS(len(p.Inputs))))
	}
	c := 0
	for _, e := range p.Outputs {
		c += Sim_UtilCountOnes(p.Structural[e.NodeID()], SIM_NUM_WORDS(len(p.Inputs)))
	}
	return c
}
func Sim_UtilCountAllPairs(info Info, n int, counters []int) int {
	for i, row := range info {
		k := Sim_UtilCountOnes(row, n)
		counters[i] = k * (k - 1) / 2
	}
	return sum(counters)
}
func Sim_UtilCountPairsOne(m Matrix, support []int) int {
	n := 0
	for i, a := range support {
		for _, b := range support[i+1:] {
			if m.Has(a, b) {
				n++
			}
		}
	}
	return n
}
func Sim_UtilCountPairsAll(p *Sym_Man_t) {
	for o, s := range p.Supports {
		p.PairsSym[o] = Sim_UtilCountPairsOne(p.Symmetric[o], s)
		p.PairsNonSym[o] = Sim_UtilCountPairsOne(p.NonSymmetric[o], s)
	}
}
func Sim_UtilMatrsAreDisjoint(p *Sym_Man_t) bool {
	for o, s := range p.Supports {
		for i, a := range s {
			for _, b := range s[i+1:] {
				if p.Symmetric[o].Has(a, b) && p.NonSymmetric[o].Has(a, b) {
					return false
				}
			}
		}
	}
	return true
}
