// Quadratic connectivity and partition constraints from place_genqp.c.
package place

import "fmt"

var g_place_qpProb *qps_problem_t

func splitPenalty(pins int32) float32 {
	if pins > 1 {
		return 1 + CLIQUE_PENALTY/float32(pins-1)
	}
	return 1 + CLIQUE_PENALTY
}

// constructQuadraticProblem retains previous movable coordinates during incremental placement.
func constructQuadraticProblem() error {
	n := int(g_place_numCells)
	x, y, fixed, area := make([]float32, n), make([]float32, n), make([]bool, n), make([]float32, n)
	for i := 0; i < n; i++ {
		c := g_place_concreteCells[i]
		if c == nil {
			fixed[i] = true
			continue
		}
		area[i] = getCellArea(c)
		fixed[i] = c.m_fixed || c.m_parent.m_pad
		x[i], y[i] = c.m_x, c.m_y
		if !fixed[i] {
			if g_place_qpProb != nil && i < len(g_place_qpProb.x) {
				x[i], y[i] = g_place_qpProb.x[i], g_place_qpProb.y[i]
			} else {
				x[i] = g_place_coreBounds.x + g_place_coreBounds.w/2
				y[i] = g_place_coreBounds.y + g_place_coreBounds.h/2
			}
		}
	}
	var edges []QPEdge
	for _, net := range g_place_concreteNets {
		if net == nil || net.m_numTerms > IGNORE_NETSIZE {
			continue
		}
		weight := net.m_weight / splitPenalty(net.m_numTerms)
		for i, a := range net.m_terms {
			for _, b := range net.m_terms[i+1:] {
				if a == nil || b == nil || a.m_id < 0 || b.m_id < 0 || int(a.m_id) >= n || int(b.m_id) >= n || g_place_concreteCells[a.m_id] != a || g_place_concreteCells[b.m_id] != b {
					return fmt.Errorf("net %d contains an unregistered cell", net.m_id)
				}
				if a != b {
					edges = append(edges, QPEdge{a.m_id, b.m_id, weight})
				}
			}
		}
	}
	p, err := NewQPProblem(x, y, fixed, edges)
	if err != nil {
		return err
	}
	if err = p.SetAreas(area); err != nil {
		return err
	}
	g_place_qpProb = p
	return nil
}

type reverseCOG struct {
	x, y  float32
	part  *Partition
	delta float32
}

// generateCoGConstraints includes every nonempty positive-area leaf; unlike the C
// implementation it does not accidentally discard the final leaf constraint.
func generateCoGConstraints(reverse []reverseCOG) int32 {
	p := g_place_qpProb
	p.cog_num = 0
	p.cog_list = nil
	p.cog_x = nil
	p.cog_y = nil
	var visit func(*Partition)
	visit = func(part *Partition) {
		if part == nil {
			return
		}
		if !part.m_leaf {
			visit(part.m_sub1)
			visit(part.m_sub2)
			return
		}
		var members []int32
		var area float32
		for _, c := range part.m_members {
			if c != nil && !c.m_fixed && !c.m_parent.m_pad {
				members = append(members, c.m_id)
				area += getCellArea(c)
			}
		}
		if len(members) == 0 || area <= 0 {
			return
		}
		x, y := part.m_bounds.x+part.m_bounds.w/2, part.m_bounds.y+part.m_bounds.h/2
		index := int(p.cog_num)
		_ = p.AddCenterOfGravity(members, x, y)
		if index < len(reverse) {
			reverse[index] = reverseCOG{x: x, y: y, part: part}
		}
	}
	visit(g_place_rootPartition)
	return p.cog_num
}

func solveQuadraticProblem(useCOG bool) error {
	if g_place_qpProb == nil {
		return fmt.Errorf("quadratic problem has not been constructed")
	}
	p := g_place_qpProb
	p.cog_num = 0
	p.cog_list = nil
	p.cog_x = nil
	p.cog_y = nil
	if useCOG {
		generateCoGConstraints(nil)
	}
	if err := p.Solve(); err != nil {
		return err
	}
	for i, c := range g_place_concreteCells {
		if c != nil && i < len(p.x) {
			c.m_x, c.m_y = p.x[i], p.y[i]
		}
	}
	return nil
}
