// Activity simulation and power scoring ported from ABC giaSwitch.c,
// aigMffc.c and darLib.c (Alan Mishchenko).
package core
import (
	"context"
	"fmt"
	"math"
	"math/bits"

	"github.com/glycerine/abcgo/aig/aig"
)

// ActivityOptions mirrors giaSwitch.c's bit-parallel sequential simulation.
// Frames before Prefix settle initially-zero registers and are not counted.
type ActivityOptions struct {
	Words, Frames, Prefix int
	RandomFactor          int // -1: 3/8 transition mask; 0: 1/2; n>0: 1/2^(n+1)
	Transitions           bool
}

func DefaultActivityOptions() ActivityOptions {
	return ActivityOptions{Words: 10, Frames: 48, Prefix: 16}
}

// SimulateActivity returns per-node probabilities with stable graph IDs. Ones
// mode follows ABC's 32-bit word stream exactly; transition mode compares packed
// consecutive 16-bit frames. Unlike the C GIA projection, constant-one and CO
// entries describe the actual AIG signal, including their polarity.
func SimulateActivity(ctx context.Context, g *aig.Graph, p ActivityOptions) ([]float32, error) {
	if g == nil {
		return nil, fmt.Errorf("dar: nil activity graph")
	}
	if p.Words < 1 || p.Frames < 1 || p.Prefix < 0 || p.Prefix >= p.Frames || p.RandomFactor < -1 || p.RandomFactor > 31 || p.Words > math.MaxInt/32/(p.Frames-p.Prefix) {
		return nil, fmt.Errorf("dar: invalid activity dimensions")
	}
	if err := g.Check(); err != nil {
		return nil, err
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	count := int(aig.Aig_ManObjNumMax(g))
	if count > math.MaxInt/(p.Words*4) {
		return nil, fmt.Errorf("dar: activity storage overflow")
	}
	storage := make([]uint32, count*p.Words)
	data := func(id int) []uint32 { return storage[id*p.Words : (id+1)*p.Words] }
	inputs, outputs := g.Inputs(), g.Outputs()
	registers := int(aig.Aig_ManRegNum(g))
	primary := len(inputs) - registers
	// Use an explicit local random state: simulations are repeatable and concurrent.
	random := aig.NewAigRandomSource()
	random.Uint32() // Gia_ManRandom(1) advances once.
	for _, e := range inputs[:primary] {
		randomActivityWords(data(e.NodeID()), random, 0, false)
	}
	order := activityOrder(g)
	counts := make([]uint64, count)
	value := func(e aig.Edge, w int) uint32 {
		v := data(e.NodeID())[w]
		if e.Complemented() {
			v = ^v
		}
		return v
	}
	for frame := 0; frame < p.Frames; frame++ {
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		for _, e := range order {
			n := g.Node(e)
			words := data(e.NodeID())
			switch n.Kind {
			case aig.Constant:
				for w := range words {
					words[w] = ^uint32(0)
				}
			case aig.Input:
			case aig.And:
				for w := range words {
					words[w] = value(n.Fanins[0], w) & value(n.Fanins[1], w)
				}
			case aig.ExclusiveOr:
				for w := range words {
					words[w] = value(n.Fanins[0], w) ^ value(n.Fanins[1], w)
				}
			case aig.Buffer, aig.Output:
				for w := range words {
					words[w] = value(n.Fanins[0], w)
				}
			}
			if frame >= p.Prefix {
				for _, v := range words {
					if p.Transitions {
						counts[e.NodeID()] += uint64(2 * bits.OnesCount32((v^(v>>16))&0xffff))
					} else {
						counts[e.NodeID()] += uint64(bits.OnesCount32(v))
					}
				}
			}
		}
		if frame+1 == p.Frames {
			break
		}
		for _, e := range inputs[:primary] {
			randomActivityWords(data(e.NodeID()), random, p.RandomFactor, p.Transitions)
		}
		// Update all register outputs simultaneously from the last register COs.
		// Their stored vectors are independent of the input storage being replaced.
		for i, e := range inputs[primary:] {
			words := data(e.NodeID())
			out := outputs[len(outputs)-registers+i]
			for w := range words {
				v := value(out, w)
				if p.Transitions {
					words[w] = (words[w] << 16) | (v & 0xffff)
				} else {
					words[w] = v
				}
			}
		}
	}
	denominator := float32(32 * p.Words * (p.Frames - p.Prefix))
	probabilities := make([]float32, count)
	for _, e := range g.LiveEdges() {
		probabilities[e.NodeID()] = float32(counts[e.NodeID()]) / denominator
	}
	return probabilities, nil
}
func activityOrder(g *aig.Graph) []aig.Edge {
	seen := make(map[int]bool)
	var order []aig.Edge
	var visit func(aig.Edge)
	visit = func(e aig.Edge) {
		e = aig.Edge(e.NodeID() << 1)
		if seen[e.NodeID()] {
			return
		}
		seen[e.NodeID()] = true
		n := g.Node(e)
		switch n.Kind {
		case aig.And, aig.ExclusiveOr:
			visit(n.Fanins[0])
			visit(n.Fanins[1])
		case aig.Buffer, aig.Output:
			visit(n.Fanins[0])
		}
		order = append(order, e)
	}
	for _, e := range g.LiveEdges() {
		visit(e)
	}
	return order
}
func randomActivityWords(words []uint32, r *aig.AigRandomSource, factor int, shift bool) {
	if factor == 0 && !shift {
		for w := len(words) - 1; w >= 0; w-- {
			words[w] = r.Uint32()
		}
		return
	}
	var mask uint32
	if factor == -1 {
		mask = (r.Uint32() & r.Uint32()) | (r.Uint32() & r.Uint32() & r.Uint32())
	} else {
		mask = r.Uint32()
		for i := 0; i < factor; i++ {
			mask &= r.Uint32()
		}
	}
	for w := len(words) - 1; w >= 0; w-- {
		if shift {
			words[w] = (words[w] << 16) | ((words[w] ^ mask) & 0xffff)
		} else {
			words[w] ^= mask
		}
	}
}

func edgeProbability(probabilities []float32, e aig.Edge) float32 {
	p := probabilities[e.NodeID()]
	if e.Complemented() {
		p = 1 - p
	}
	return p
}

// mffcPower matches Aig_NodeDeref_rec's edge-weighted power sum, using the
// current cut-protected MFFC labels without changing graph references.
func mffcPower(g *aig.Graph, root aig.Edge, probabilities []float32) float32 {
	remaining := make(map[int]int)
	var visit func(aig.Edge) float32
	visit = func(e aig.Edge) float32 {
		n := g.Node(e)
		if n.Kind == aig.Input || n.Kind == aig.Constant {
			return 0
		}
		var power float32
		fanins := n.Fanins[:1]
		if n.Kind == aig.And || n.Kind == aig.ExclusiveOr {
			fanins = n.Fanins[:]
		}
		for _, f := range fanins {
			var child float32
			refs, ok := remaining[f.NodeID()]
			if !ok {
				refs = g.Node(f).Refs
			}
			refs--
			remaining[f.NodeID()] = refs
			if refs == 0 && aig.Aig_ObjIsTravIdCurrent(g, f) {
				child = visit(f)
			}
			power = float32(float64(power) + (float64(child) + switchingPower64(probabilities[f.NodeID()])))
		}
		return power
	}
	return visit(root)
}
func switchingPower64(p float32) float64 { return 2 * float64(p) * (1 - float64(p)) }

// evaluateLibraryCandidate preserves Dar_LibEval_rec's original accounting:
// cut terminals/reused nodes return their one-probability, repeated DAG visits
// return zero, and each charged node adds its two fanin switching terms.
func evaluateLibraryCandidate(g *aig.Graph, class *preparedClass, existing []aig.Edge, index int, initial []float32, powerMode bool) (int, float32) {
	seen := make([]bool, len(existing))
	probabilities := append([]float32(nil), initial...)
	var evaluate func(int) (int, float32)
	evaluate = func(index int) (int, float32) {
		if seen[index] {
			return 0, 0
		}
		seen[index] = true
		if index < 4 {
			return 0, probabilities[index]
		}
		if e := existing[index]; e != aig.InvalidEdge && !aig.Aig_ObjIsTravIdCurrent(g, e) {
			return 0, probabilities[index]
		}
		n := class.nodes[index-4]
		a, b := int(n.fanins[0]>>1), int(n.fanins[1]>>1)
		area0, power0 := evaluate(a)
		area1, power1 := evaluate(b)
		if !powerMode {
			return 1 + area0 + area1, 0
		}
		pa, pb := float64(probabilities[a]), float64(probabilities[b])
		if n.fanins[0]&1 != 0 {
			pa = 1 - pa
		}
		if n.fanins[1]&1 != 0 {
			pb = 1 - pb
		}
		probabilities[index] = float32(pa * pb)
		power := float32(float64(power0) + switchingPower64(probabilities[a]) + float64(power1) + switchingPower64(probabilities[b]))
		return 1 + area0 + area1, power
	}
	return evaluate(index)
}

// New structural-hash nodes use the original Aig_ObjCreate independence model;
// previously measured node probabilities are retained throughout the run.
func (m *Manager) extendProbabilities() {
	if !m.Params.Power {
		return
	}
	old := len(m.probabilities)
	count := int(aig.Aig_ManObjNumMax(m.Graph))
	if count <= old {
		return
	}
	m.probabilities = append(m.probabilities, make([]float32, count-old)...)
	for _, e := range activityOrder(m.Graph) {
		if e.NodeID() < old {
			continue
		}
		n := m.Graph.Node(e)
		var p float32
		switch n.Kind {
		case aig.Constant:
			p = 1
		case aig.And:
			a, b := float64(edgeProbability(m.probabilities, n.Fanins[0])), float64(edgeProbability(m.probabilities, n.Fanins[1]))
			p = float32(a * b)
		case aig.ExclusiveOr:
			a, b := float64(edgeProbability(m.probabilities, n.Fanins[0])), float64(edgeProbability(m.probabilities, n.Fanins[1]))
			p = float32(a*(1-b) + (1-a)*b)
		case aig.Buffer, aig.Output:
			p = edgeProbability(m.probabilities, n.Fanins[0])
		}
		m.probabilities[e.NodeID()] = p
	}
}
