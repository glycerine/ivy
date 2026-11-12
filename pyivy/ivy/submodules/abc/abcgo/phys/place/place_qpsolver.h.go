// Quadratic placement solver interface, ported from Philip Chong's qps header.
package place

import (
	"fmt"
	"math"
)

type qps_float_t = float32

type qps_problem_t struct {
	num_cells                                           int32
	connect                                             []int32
	edge_weight                                         []float32
	x, y                                                []float32
	fixed                                               []int32
	f                                                   float32
	cog_num                                             int32
	cog_list                                            []int32
	cog_x, cog_y, area                                  []float32
	loop_num                                            int32
	loop_list                                           []int32
	loop_max, loop_penalty                              []float32
	loop_k, loop_done, loop_fail                        int32
	max_x, max_y                                        float32
	max_enable, max_done                                int32
	priv_ii                                             []int32
	priv_n                                              int32
	priv_cp, priv_g, priv_h, priv_xi, priv_tp, priv_tp2 []float32
	priv_pcgt                                           []float32
	priv_lt, priv_mxl, priv_mxh, priv_myl, priv_myh     []float32
	priv_f, priv_fmax, priv_fprev, priv_fopt, priv_eps  float32
	priv_pn                                             int32
	edges                                               []qpEdge
	groups                                              []qpGroup
	loops                                               [][]int32
	initialized                                         bool
	solveErr                                            error
}
type qpEdge struct {
	a, b   int32
	weight float32
}
type qpGroup struct {
	members   []int32
	dependent int32
	total     float32
}

// QPProblem retains float32 arithmetic and ABC's original cost/constraint model.
type QPProblem = qps_problem_t

// QPEdge is an undirected connection whose squared Euclidean length is weighted.
type QPEdge struct {
	A, B   int32
	Weight float32
}

// NewQPProblem copies positions and constructs symmetric connectivity from edges.
func NewQPProblem(x, y []float32, fixed []bool, edges []QPEdge) (*QPProblem, error) {
	if len(x) > math.MaxInt32/2 || len(x) != len(y) || len(fixed) != len(x) {
		return nil, fmt.Errorf("QP input lengths differ")
	}
	p := &QPProblem{num_cells: int32(len(x)), x: append([]float32(nil), x...), y: append([]float32(nil), y...), fixed: make([]int32, len(x))}
	rows := make([][]int, len(x))
	for i, e := range edges {
		if e.A < 0 || e.B < 0 || int(e.A) >= len(x) || int(e.B) >= len(x) || e.Weight < 0 || !qpFinite(e.Weight) {
			return nil, fmt.Errorf("invalid QP edge")
		}
		rows[e.A] = append(rows[e.A], i)
		if e.B != e.A {
			rows[e.B] = append(rows[e.B], i)
		}
	}
	for v, row := range rows {
		if fixed[v] {
			p.fixed[v] = 1
		}
		for _, i := range row {
			e := edges[i]
			other := e.A
			if other == int32(v) {
				other = e.B
			}
			p.connect = append(p.connect, other)
			p.edge_weight = append(p.edge_weight, e.Weight)
		}
		p.connect = append(p.connect, -1)
		p.edge_weight = append(p.edge_weight, -1)
	}
	if err := qpValidateGraph(p); err != nil {
		return nil, err
	}
	return p, nil
}
func (p *QPProblem) SetAreas(area []float32) error {
	if len(area) != int(p.num_cells) {
		return fmt.Errorf("QP area length differs")
	}
	for _, a := range area {
		if a < 0 || !qpFinite(a) {
			return fmt.Errorf("invalid QP area")
		}
	}
	p.area = append(p.area[:0], area...)
	return nil
}
func (p *QPProblem) AddCenterOfGravity(cells []int32, x, y float32) error {
	if len(cells) == 0 || !qpFinite(x) || !qpFinite(y) {
		return fmt.Errorf("invalid COG constraint")
	}
	for _, v := range cells {
		if v < 0 || v >= p.num_cells {
			return fmt.Errorf("COG index outside QP")
		}
	}
	p.cog_list = append(p.cog_list, cells...)
	p.cog_list = append(p.cog_list, -1)
	p.cog_x = append(p.cog_x, x)
	p.cog_y = append(p.cog_y, y)
	p.cog_num++
	return nil
}

// AddLoop bounds the sum of squared edge lengths around the closed cell walk.
func (p *QPProblem) AddLoop(cells []int32, maxSquaredLength, penalty float32) error {
	if len(cells) == 0 || maxSquaredLength < 0 || penalty < 0 || !qpFinite(maxSquaredLength) || !qpFinite(penalty) {
		return fmt.Errorf("invalid loop constraint")
	}
	for _, v := range cells {
		if v < 0 || v >= p.num_cells {
			return fmt.Errorf("loop index outside QP")
		}
	}
	p.loop_list = append(p.loop_list, cells...)
	p.loop_list = append(p.loop_list, -1)
	p.loop_max = append(p.loop_max, maxSquaredLength)
	p.loop_penalty = append(p.loop_penalty, penalty)
	p.loop_num++
	return nil
}
func (p *QPProblem) SetBounds(x, y float32) error {
	if x < 0 || y < 0 || !qpFinite(x) || !qpFinite(y) {
		return fmt.Errorf("invalid QP bounds")
	}
	p.max_x, p.max_y, p.max_enable = x, y, 1
	return nil
}
func (p *QPProblem) Solve() error {
	if err := qpValidateGraph(p); err != nil {
		return err
	}
	if err := qpValidateConstraints(p); err != nil {
		return err
	}
	bounds := p.max_enable != 0
	qps_init(p)
	if bounds {
		p.max_enable = 1
	}
	defer qps_clean(p)
	qps_solve(p)
	if p.solveErr != nil {
		return p.solveErr
	}
	if p.loop_fail != 0 {
		return fmt.Errorf("QP constraint optimization did not converge (failure %d)", p.loop_fail)
	}
	return nil
}
func (p *QPProblem) Coordinates() ([]float32, []float32) {
	return append([]float32(nil), p.x...), append([]float32(nil), p.y...)
}
func (p *QPProblem) Objective() float32 { return p.f }
func (p *QPProblem) FailureCode() int32 { return p.loop_fail }
