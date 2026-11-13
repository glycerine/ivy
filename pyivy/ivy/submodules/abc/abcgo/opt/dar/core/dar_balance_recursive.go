package core

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/tim"
)

// balanceRecursive reconstructs CO cones in source order, preserving sharing.
func balanceRecursive(ctx context.Context, src *aig.Graph, update bool) (*aig.Graph, error) {
	dst := aig.NewGraph()
	dst.Name, dst.Spec, dst.Time2Quit = src.Name, src.Spec, src.Time2Quit
	dst.Assertions = src.Assertions
	dst.FlopNumbers = append([]int32(nil), src.FlopNumbers...)
	dst.BarrierBuffers = src.BarrierBuffers
	var timing *tim.Tim_Man_t
	if src.Timing != nil {
		if m, ok := src.Timing.(*tim.Tim_Man_t); ok {
			timing = tim.Tim_ManDup(m, 0)
			dst.Timing = timing
		} else if c, ok := src.Timing.(interface{ CloneTiming() any }); ok {
			dst.Timing = c.CloneTiming()
		} else {
			return nil, fmt.Errorf("dar: timing metadata must implement CloneTiming or be a timing manager")
		}
	}
	if timing != nil {
		tim.Tim_ManIncrementTravId(timing)
	}
	copies := map[int]aig.Edge{0: 0}
	ciIds := aig.Aig_ManSetCioIds(src)
	defer aig.Aig_ManCleanCioIds(src)
	for _, in := range src.Inputs() {
		out := dst.NewInput()
		level := src.Node(in).Level
		if timing != nil {
			level = int(tim.Tim_ManGetCiArrival(timing, int(ciIds[in])))
		}
		aig.Aig_ObjSetLevel(dst, out, int32(level))
		copies[in.NodeID()] = out
	}
	var visit func(aig.Edge) (aig.Edge, error)
	visits := 0
	visit = func(edge aig.Edge) (aig.Edge, error) {
		visits++
		if visits&255 == 0 && !src.Time2Quit.IsZero() && time.Now().After(src.Time2Quit) {
			return 0, context.DeadlineExceeded
		}
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
		}
		edge = aig.Aig_ObjReal_rec(src, edge)
		complemented := edge.Complemented()
		root := aig.Edge(edge.NodeID() << 1)
		if value, ok := copies[root.NodeID()]; ok {
			if complemented {
				value = value.Not()
			}
			return value, nil
		}
		kind := src.Node(root).Kind
		if kind != aig.And && kind != aig.ExclusiveOr {
			return 0, fmt.Errorf("dar: unsupported balance node %v", kind)
		}
		var leaves []aig.Edge
		var collect func(aig.Edge, bool)
		collect = func(e aig.Edge, first bool) {
			e = aig.Aig_ObjReal_rec(src, e)
			n := src.Node(e)
			if !first && (e.Complemented() || n.Kind != kind || n.Refs > 1 || len(leaves) > 10000) {
				leaves = append(leaves, e)
				return
			}
			collect(n.Fanins[0], false)
			collect(n.Fanins[1], false)
		}
		collect(root, true)
		sort.Slice(leaves, func(i, j int) bool { return leaves[i] < leaves[j] })
		unique := leaves[:0]
		for i := 0; i < len(leaves); i++ {
			if i+1 < len(leaves) && kind == aig.And && leaves[i] == leaves[i+1].Not() {
				unique = nil
				break
			}
			if i+1 < len(leaves) && leaves[i] == leaves[i+1] {
				if kind == aig.ExclusiveOr {
					i++
				}
				continue
			}
			unique = append(unique, leaves[i])
		}
		work := make([]aig.Edge, 0, len(unique))
		for _, leaf := range unique {
			value, err := visit(leaf)
			if err != nil {
				return 0, err
			}
			work = append(work, value)
		}
		value := dst.Const0()
		if len(work) == 1 {
			value = work[0]
		} else if len(work) > 1 {
			value = balanceRecursiveBuild(dst, work, kind, update)
		}
		if value.NodeID()&255 == 0 && !src.Time2Quit.IsZero() && time.Now().After(src.Time2Quit) {
			return 0, context.DeadlineExceeded
		}
		copies[root.NodeID()] = value
		if complemented {
			value = value.Not()
		}
		return value, nil
	}
	outputs := src.Outputs()
	deferOutputs := src.Timing == nil && src.BarrierBuffers > 0 && int(src.BarrierBuffers) <= len(outputs) && int(src.BarrierBuffers) <= len(src.Inputs())
	balancedOutputs := make([]aig.Edge, len(outputs))
	order := make([]int, len(outputs))
	for i := range order {
		order[i] = i
	}
	if deferOutputs {
		n := int(src.BarrierBuffers)
		for i := range order {
			if i < n {
				order[i] = len(outputs) - n + i
			} else {
				order[i] = i - n
			}
		}
	}
	for i, outIndex := range order {
		out := outputs[outIndex]
		value, err := visit(src.Node(out).Fanins[0])
		if err != nil {
			return nil, err
		}
		balancedOutputs[outIndex] = value
		if !deferOutputs {
			dst.AddOutput(value)
		}
		if timing != nil {
			co := src.Node(out)
			tim.Tim_ManSetCoArrival(timing, int(co.CioID), float32(dst.Node(value).Level))
		}
		if src.Timing == nil && i < int(src.BarrierBuffers) {
			ciIndex := len(src.Inputs()) - int(src.BarrierBuffers) + i
			input := dst.Inputs()[ciIndex]
			aig.Aig_ObjSetLevel(dst, input, int32(dst.Node(value).Level))
		}
	}
	if deferOutputs {
		for _, value := range balancedOutputs {
			dst.AddOutput(value)
		}
	}
	aig.Aig_ManCleanup(dst)
	aig.Aig_ManSetRegNum(dst, src.RegisterCount())
	aig.Aig_ManSetConstrNum(dst, aig.Aig_ManConstrNum(src))
	return dst, dst.Check()
}

func balanceRecursiveBuild(g *aig.Graph, work []aig.Edge, kind aig.NodeKind, update bool) aig.Edge {
	sort.Slice(work, func(i, j int) bool {
		a, b := g.Node(work[i]).Level, g.Node(work[j]).Level
		if a != b {
			return a > b
		}
		return work[i].NodeID() > work[j].NodeID()
	})
	for len(work) > 1 {
		right := len(work) - 2
		left := 0
		if update {
			left = right
			for left > 0 && g.Node(work[left-1]).Level == g.Node(work[right]).Level {
				left--
			}
		}
		one, two := work[right+1], work[right]
		if left < right && one.NodeID() != 0 && two.NodeID() != 0 && one.NodeID() != two.NodeID() {
			for i := right; i >= left; i-- {
				candidate := work[i]
				share := candidate.NodeID() == 0 || candidate.NodeID() == one.NodeID()
				if !share {
					ghost := aig.Aig_ObjCreateGhost(g, one, candidate, aig.Aig_Type_t(map[aig.NodeKind]int32{aig.And: 5, aig.ExclusiveOr: 6}[kind]))
					share = aig.Aig_TableLookup(g, ghost) != aig.InvalidEdge
				}
				if share {
					work[i], work[right] = work[right], work[i]
					break
				}
			}
		}
		one, two = work[len(work)-1], work[len(work)-2]
		work = work[:len(work)-2]
		var value aig.Edge
		if kind == aig.ExclusiveOr {
			value = g.NewXor(one, two)
		} else {
			value = g.AND(one, two)
		}
		found := -1
		for i, e := range work {
			if e == value {
				found = i
				break
			}
		}
		if found >= 0 {
			if kind == aig.ExclusiveOr {
				work = append(work[:found], work[found+1:]...)
			}
			continue
		}
		work = append(work, value)
		for i := len(work) - 1; i > 0 && g.Node(work[i]).Level > g.Node(work[i-1]).Level; i-- {
			work[i], work[i-1] = work[i-1], work[i]
		}
	}
	if len(work) == 0 {
		return g.Const0()
	}
	return work[0]
}
