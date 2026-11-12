// Preconditioned conjugate-gradient placement solver, ported from
// place_qpsolver.c (Philip Chong). COG equalities eliminate one variable per
// disjoint group; loops and bounds use projected Lagrange multipliers.
package place

import (
	"fmt"
	"math"
)

const (
	QPS_TOL              = 1e-3
	QPS_EPS              = QPS_TOL * QPS_TOL
	QPS_MAX_TOL          = .1
	QPS_LOOP_TOL         = 1e-3
	QPS_RELAX_ITER       = 180
	QPS_MAX_ITER         = 200
	QPS_STEPSIZE_RETRIES = 2
	QPS_MINSTEP          = 1e-6
	QPS_DEC_CHANGE       = .01
	QPS_PRECON_EPS       = 1e-9
)

func qpAbs(x float32) float32 { return float32(math.Abs(float64(x))) }
func qpFinite(x float32) bool { return !math.IsNaN(float64(x)) && !math.IsInf(float64(x), 0) }
func qpLists(data []int32, count, n int32) ([][]int32, error) {
	if count < 0 {
		return nil, fmt.Errorf("negative constraint count")
	}
	rows := make([][]int32, 0, count)
	start := 0
	for i := int32(0); i < count; i++ {
		end := start
		for end < len(data) && data[end] >= 0 {
			if data[end] >= n {
				return nil, fmt.Errorf("cell index outside QP")
			}
			end++
		}
		if end == len(data) {
			return nil, fmt.Errorf("missing -1 list terminator")
		}
		if data[end] != -1 {
			return nil, fmt.Errorf("invalid list terminator")
		}
		rows = append(rows, append([]int32(nil), data[start:end]...))
		start = end + 1
	}
	return rows, nil
}
func qpValidateGraph(p *qps_problem_t) error {
	if p == nil || p.num_cells < 0 || int64(p.num_cells)*2 > math.MaxInt32 {
		return fmt.Errorf("invalid QP size")
	}
	n := int(p.num_cells)
	if len(p.x) < n || len(p.y) < n || len(p.fixed) < n {
		return fmt.Errorf("short QP coordinate/fixed array")
	}
	for i := 0; i < n; i++ {
		if !qpFinite(p.x[i]) || !qpFinite(p.y[i]) {
			return fmt.Errorf("nonfinite QP coordinate")
		}
	}
	rows, err := qpLists(p.connect, p.num_cells, p.num_cells)
	if err != nil {
		return err
	}
	offset := 0
	for _, row := range rows {
		for range row {
			if offset >= len(p.edge_weight) || !qpFinite(p.edge_weight[offset]) || p.edge_weight[offset] < 0 {
				return fmt.Errorf("invalid QP edge weight")
			}
			offset++
		}
		offset++
	}
	return nil
}
func qpValidateConstraints(p *qps_problem_t) error {
	groups, err := qpLists(p.cog_list, p.cog_num, p.num_cells)
	if err != nil {
		return err
	}
	if len(p.cog_x) < len(groups) || len(p.cog_y) < len(groups) || len(groups) > 0 && len(p.area) < int(p.num_cells) {
		return fmt.Errorf("short COG array")
	}
	used := map[int32]bool{}
	for i, members := range groups {
		if !qpFinite(p.cog_x[i]) || !qpFinite(p.cog_y[i]) {
			return fmt.Errorf("nonfinite COG target")
		}
		dependent := false
		var total float32
		local := map[int32]bool{}
		for _, v := range members {
			if local[v] {
				return fmt.Errorf("duplicate COG member")
			}
			local[v] = true
			if !qpFinite(p.area[v]) || p.area[v] < 0 {
				return fmt.Errorf("invalid COG area")
			}
			total += p.area[v]
			if p.fixed[v] == 0 {
				if used[v] {
					return fmt.Errorf("overlapping movable COG groups")
				}
				used[v] = true
				dependent = dependent || p.area[v] > 0
			}
		}
		if !qpFinite(total) {
			return fmt.Errorf("COG total area overflows float32")
		}
		if !dependent {
			return fmt.Errorf("COG group requires a positive-area movable cell")
		}
	}
	loops, err := qpLists(p.loop_list, p.loop_num, p.num_cells)
	if err != nil {
		return err
	}
	if len(p.loop_max) < len(loops) || len(p.loop_penalty) < len(loops) {
		return fmt.Errorf("short loop constraint array")
	}
	for i, row := range loops {
		if len(row) == 0 || !qpFinite(p.loop_max[i]) || !qpFinite(p.loop_penalty[i]) || p.loop_max[i] < 0 || p.loop_penalty[i] < 0 {
			return fmt.Errorf("invalid loop constraint")
		}
	}
	if p.max_enable != 0 && (!qpFinite(p.max_x) || !qpFinite(p.max_y) || p.max_x < 0 || p.max_y < 0) {
		return fmt.Errorf("invalid QP bounds")
	}
	return nil
}
func qps_init(p *qps_problem_t) {
	if err := qpValidateGraph(p); err != nil {
		panic(err)
	}
	p.edges = nil
	pr := 0
	for i := int32(0); i < p.num_cells; i++ {
		for p.connect[pr] >= 0 {
			j := p.connect[pr]
			if j > i {
				p.edges = append(p.edges, qpEdge{i, j, p.edge_weight[pr]})
			}
			pr++
		}
		pr++
	}
	p.priv_ii = make([]int32, p.num_cells)
	p.priv_tp = make([]float32, 2*p.num_cells)
	p.priv_tp2 = make([]float32, 2*p.num_cells)
	p.max_enable = 0
	p.priv_fopt = 0
	p.initialized = true
	p.solveErr = nil
}
func qps_settp(p *qps_problem_t) {
	for i := p.num_cells - 1; i >= 0; i-- {
		index := p.priv_ii[i]
		if index >= 0 {
			p.priv_tp[2*i], p.priv_tp[2*i+1] = p.priv_cp[index], p.priv_cp[index+1]
		} else if index == -1 {
			p.priv_tp[2*i], p.priv_tp[2*i+1] = p.x[i], p.y[i]
		}
	}
	for i := p.num_cells - 1; i >= 0; i-- {
		index := p.priv_ii[i]
		if index >= -1 {
			continue
		}
		group := p.groups[-index-2]
		var rx, ry float32
		for _, v := range group.members {
			if v != i {
				rx -= p.area[v] * p.priv_tp[2*v]
				ry -= p.area[v] * p.priv_tp[2*v+1]
			}
		}
		rx += p.cog_x[-index-2] * group.total
		ry += p.cog_y[-index-2] * group.total
		p.priv_tp[2*i], p.priv_tp[2*i+1] = rx/p.area[i], ry/p.area[i]
	}
}
func qpLoopEnergy(p *qps_problem_t, points []float32, i int) float32 {
	members := p.loops[i]
	var value float32
	previous := members[0]
	for _, next := range members[1:] {
		dx, dy := points[2*previous]-points[2*next], points[2*previous+1]-points[2*next+1]
		value += dx*dx + dy*dy
		previous = next
	}
	first := members[0]
	dx, dy := points[2*previous]-points[2*first], points[2*previous+1]-points[2*first+1]
	return value + (dx*dx + dy*dy)
}
func qps_func(p *qps_problem_t) float32 {
	var value float32
	for _, e := range p.edges {
		dx, dy := p.priv_tp[2*e.b]-p.priv_tp[2*e.a], p.priv_tp[2*e.b+1]-p.priv_tp[2*e.a+1]
		value += e.weight * (dx*dx + dy*dy)
	}
	p.f = value
	for i := range p.loops {
		residual := qpLoopEnergy(p, p.priv_tp, i) - p.loop_max[i]
		p.priv_lt[i] = residual
		value += p.loop_penalty[i] * residual
	}
	if p.max_enable != 0 {
		for i := p.num_cells - 1; i >= 0; i-- {
			x, y := p.priv_tp[2*i], p.priv_tp[2*i+1]
			value += p.priv_mxl[i] * (-x)
			value += p.priv_mxh[i] * (x - p.max_x)
			value += p.priv_myl[i] * (-y)
			value += p.priv_myh[i] * (y - p.max_y)
		}
	}
	return value
}
func qpAddGradient(all []float32, a, b int32, weight float32, points []float32) {
	dx, dy := weight*(points[2*a]-points[2*b]), weight*(points[2*a+1]-points[2*b+1])
	all[2*a] += dx
	all[2*b] -= dx
	all[2*a+1] += dy
	all[2*b+1] -= dy
}
func qps_dfunc(p *qps_problem_t, d []float32) {
	clear(p.priv_tp2)
	for _, e := range p.edges {
		qpAddGradient(p.priv_tp2, e.a, e.b, 2*e.weight, p.priv_tp)
	}
	for i, row := range p.loops {
		previous := row[0]
		for _, next := range row[1:] {
			qpAddGradient(p.priv_tp2, previous, next, 2*p.loop_penalty[i], p.priv_tp)
			previous = next
		}
		qpAddGradient(p.priv_tp2, previous, row[0], 2*p.loop_penalty[i], p.priv_tp)
	}
	if p.max_enable != 0 {
		for i := p.num_cells - 1; i >= 0; i-- {
			p.priv_tp2[2*i] += p.priv_mxh[i] - p.priv_mxl[i]
			p.priv_tp2[2*i+1] += p.priv_myh[i] - p.priv_myl[i]
		}
	}
	clear(d[:p.priv_n])
	for i := p.num_cells - 1; i >= 0; i-- {
		index := p.priv_ii[i]
		if index >= 0 {
			d[index] += p.priv_tp2[2*i]
			d[index+1] += p.priv_tp2[2*i+1]
		} else if index < -1 {
			for _, v := range p.groups[-index-2].members {
				other := p.priv_ii[v]
				if other >= 0 {
					weight := p.area[v] / p.area[i]
					d[other] -= p.priv_tp2[2*i] * weight
					d[other+1] -= p.priv_tp2[2*i+1] * weight
				}
			}
		}
	}
}
func qps_linmin(p *qps_problem_t, dgg float32, h []float32) {
	clear(p.priv_tp)
	for j := p.num_cells - 1; j >= 0; j-- {
		index := p.priv_ii[j]
		if index >= 0 {
			p.priv_tp[2*j], p.priv_tp[2*j+1] = h[index], h[index+1]
		} else if index < -1 {
			for _, v := range p.groups[-index-2].members {
				other := p.priv_ii[v]
				if other >= 0 {
					weight := p.area[v] / p.area[j]
					p.priv_tp[2*j] -= h[other] * weight
					p.priv_tp[2*j+1] -= h[other+1] * weight
				}
			}
		}
	}
	var curvature float32
	for _, e := range p.edges {
		dx, dy := p.priv_tp[2*e.b]-p.priv_tp[2*e.a], p.priv_tp[2*e.b+1]-p.priv_tp[2*e.a+1]
		curvature += e.weight * (dx*dx + dy*dy)
	}
	for i := range p.loops {
		curvature += p.loop_penalty[i] * qpLoopEnergy(p, p.priv_tp, i)
	}
	if curvature <= 0 || !qpFinite(curvature) {
		p.solveErr = fmt.Errorf("QP has a singular or nonfinite search direction")
		return
	}
	step := (dgg / curvature) / 2
	if !qpFinite(step) {
		p.solveErr = fmt.Errorf("nonfinite QP line-search step")
		return
	}
	for j := p.priv_n - 1; j >= 0; j-- {
		if !qpFinite(p.priv_cp[j] + step*h[j]) {
			p.solveErr = fmt.Errorf("nonfinite QP line-search point")
			return
		}
	}
	for j := p.priv_n - 1; j >= 0; j-- {
		p.priv_cp[j] += step * h[j]
	}
}
func qps_cgmin(p *qps_problem_t) {
	qps_settp(p)
	previous := qps_func(p)
	p.priv_f = previous
	qps_dfunc(p, p.priv_g)
	var dgg float32
	for j := p.priv_n - 1; j >= 0; j-- {
		p.priv_g[j] = -p.priv_g[j]
		p.priv_h[j] = p.priv_g[j] * p.priv_pcgt[j]
		dgg += p.priv_g[j] * p.priv_h[j]
	}
	for iteration := int32(0); iteration < 2*p.priv_n; iteration++ {
		if dgg == 0 {
			return
		}
		qps_linmin(p, dgg, p.priv_h)
		qps_settp(p)
		p.priv_f = qps_func(p)
		if p.solveErr != nil {
			return
		}
		if !qpFinite(p.priv_f) {
			p.solveErr = fmt.Errorf("nonfinite QP objective")
			return
		}
		if qpAbs(p.priv_f-previous) <= (qpAbs(p.priv_f)+qpAbs(previous)+QPS_EPS)*QPS_TOL/2 {
			return
		}
		previous = p.priv_f
		qps_dfunc(p, p.priv_xi)
		gg := dgg
		dgg = 0
		for j := p.priv_n - 1; j >= 0; j-- {
			dgg += p.priv_xi[j] * p.priv_xi[j] * p.priv_pcgt[j]
		}
		gamma := dgg / gg
		for j := p.priv_n - 1; j >= 0; j-- {
			p.priv_g[j] = -p.priv_xi[j]
			p.priv_h[j] = p.priv_g[j]*p.priv_pcgt[j] + gamma*p.priv_h[j]
		}
	}
}

func qpPrepareVariables(p *qps_problem_t) {
	if err := qpValidateConstraints(p); err != nil {
		panic(err)
	}
	for i := range p.priv_ii {
		p.priv_ii[i] = 0
		if p.fixed[i] != 0 {
			p.priv_ii[i] = -1
		}
	}
	rows, _ := qpLists(p.cog_list, p.cog_num, p.num_cells)
	p.groups = make([]qpGroup, len(rows))
	for i, members := range rows {
		group := qpGroup{members: members, dependent: -1}
		var largest float32 = -1
		for _, v := range members {
			group.total += p.area[v]
			if p.fixed[v] == 0 && p.area[v] > largest {
				largest = p.area[v]
				group.dependent = v
			}
		}
		p.groups[i] = group
		p.priv_ii[group.dependent] = -2 - int32(i)
	}
	p.priv_n = 0
	for i := p.num_cells - 1; i >= 0; i-- {
		if p.priv_ii[i] == 0 {
			p.priv_ii[i] = 2 * p.priv_n
			p.priv_n++
		}
	}
	p.priv_n *= 2
	p.priv_cp = make([]float32, p.priv_n)
	p.priv_g = make([]float32, p.priv_n)
	p.priv_h = make([]float32, p.priv_n)
	p.priv_xi = make([]float32, p.priv_n)
	for i := p.num_cells - 1; i >= 0; i-- {
		if index := p.priv_ii[i]; index >= 0 {
			p.priv_cp[index], p.priv_cp[index+1] = p.x[i], p.y[i]
		}
	}
	p.loops, _ = qpLists(p.loop_list, p.loop_num, p.num_cells)
	p.priv_lt = make([]float32, p.loop_num)
	degree := make([]float32, p.num_cells)
	for _, e := range p.edges {
		degree[e.a] += e.weight
		degree[e.b] += e.weight
	}
	for i, row := range p.loops {
		weight := 2 * p.loop_penalty[i]
		for _, v := range row {
			degree[v] += weight
		}
	}
	p.priv_pcgt = make([]float32, p.priv_n)
	for i := int32(0); i < p.num_cells; i++ {
		if index := p.priv_ii[i]; index >= 0 {
			p.priv_pcgt[index] += degree[i]
			p.priv_pcgt[index+1] += degree[i]
		}
	}
	for i, value := range p.priv_pcgt {
		if qpAbs(value) < QPS_PRECON_EPS || qpAbs(value) > 1/QPS_PRECON_EPS {
			p.priv_pcgt[i] = 1
		} else {
			p.priv_pcgt[i] = 1 / value
		}
	}
	if p.max_enable != 0 {
		p.max_enable = 1
		p.priv_mxl = make([]float32, p.num_cells)
		p.priv_mxh = make([]float32, p.num_cells)
		p.priv_myl = make([]float32, p.num_cells)
		p.priv_myh = make([]float32, p.num_cells)
	}
}
func qps_estopt(p *qps_problem_t) float32 {
	estimate := p.priv_f
	if p.max_enable != 0 {
		temporary := make([]float32, 2*p.num_cells)
		for i, group := range p.groups {
			for _, v := range group.members {
				temporary[2*v], temporary[2*v+1] = p.cog_x[i], p.cog_y[i]
			}
		}
		old := p.priv_tp
		p.priv_tp = temporary
		estimate = qps_func(p)
		p.priv_tp = old
		// qps_func also writes base cost and loop residuals: restore these for
		// the current point before updating constraint multipliers.
		qps_func(p)
		term := (p.max_x*p.max_x + p.max_y*p.max_y) * float32(p.num_cells)
		for _, fixed := range p.fixed[:p.num_cells] {
			if fixed != 0 {
				estimate += term
			}
		}
	}
	if p.loop_num != 0 {
		estimate *= 8
	}
	return estimate
}
func qpUpdateLoopMultipliers(p *qps_problem_t) {
	p.loop_done = 1
	if p.loop_num == 0 {
		return
	}
	p.loop_k++
	var norm float32
	for i := p.loop_num - 1; i >= 0; i-- {
		residual := p.priv_lt[i]
		if residual > 0 || p.loop_penalty[i] > 0 {
			norm += residual * residual
		}
		if qpAbs(residual) < QPS_LOOP_TOL {
			continue
		}
		tolerance := float32(QPS_LOOP_TOL) * p.loop_max[i]
		if residual > tolerance || (p.loop_k < QPS_RELAX_ITER && p.loop_penalty[i]*residual < -tolerance) {
			p.loop_done = 0
		}
	}
	if p.loop_done != 0 {
		return
	}
	if norm <= 0 || !qpFinite(norm) {
		p.loop_fail |= 1
		return
	}
	step := p.priv_eps * (p.priv_fopt - p.priv_f) / norm
	for i := p.loop_num - 1; i >= 0; i-- {
		p.loop_penalty[i] = max(float32(0), p.loop_penalty[i]+step*p.priv_lt[i])
	}
}
func qpUpdateBounds(p *qps_problem_t) {
	p.max_done = 1
	if p.max_enable == 0 {
		return
	}
	var norm float32
	for i := p.num_cells - 1; i >= 0; i-- {
		x, y := p.priv_tp[2*i], p.priv_tp[2*i+1]
		for k, residual := range []float32{-x, x - p.max_x, -y, y - p.max_y} {
			norm += residual * residual
			multiplier := []float32{p.priv_mxl[i], p.priv_mxh[i], p.priv_myl[i], p.priv_myh[i]}[k]
			if residual > QPS_TOL || (p.max_enable < QPS_RELAX_ITER && multiplier*residual < -QPS_MAX_TOL) {
				p.max_done = 0
			}
		}
	}
	if p.max_done == 0 {
		if norm <= 0 || !qpFinite(norm) {
			p.loop_fail |= 1
			return
		}
		step := p.priv_eps * (p.priv_fopt - p.priv_f) / norm
		for i := p.num_cells - 1; i >= 0; i-- {
			x, y := p.priv_tp[2*i], p.priv_tp[2*i+1]
			p.priv_mxl[i] = max(float32(0), p.priv_mxl[i]-step*x)
			p.priv_mxh[i] = max(float32(0), p.priv_mxh[i]+step*(x-p.max_x))
			p.priv_myl[i] = max(float32(0), p.priv_myl[i]-step*y)
			p.priv_myh[i] = max(float32(0), p.priv_myh[i]+step*(y-p.max_y))
		}
	}
	p.max_enable++
}
func qps_solve_inner(p *qps_problem_t) {
	qps_cgmin(p)
	if p.solveErr != nil {
		p.loop_fail |= 1
		p.loop_done, p.max_done = 1, 1
		return
	}
	if p.max_enable != 0 || p.loop_num != 0 {
		if p.max_enable == 1 || p.loop_num != 0 && p.loop_k == 0 {
			p.priv_eps = 2
			p.priv_fmax, p.priv_fprev = p.priv_f, p.priv_f
			p.priv_fopt = qps_estopt(p)
			if p.priv_fopt <= p.priv_f {
				p.priv_fopt = p.priv_f + max(float32(1), qpAbs(p.priv_f))
			}
			p.priv_pn = 0
		} else {
			if p.priv_f < p.priv_fprev && p.priv_fprev-p.priv_f > QPS_DEC_CHANGE*qpAbs(p.priv_fprev) {
				old := p.priv_pn
				p.priv_pn++
				if old >= QPS_STEPSIZE_RETRIES {
					p.priv_eps /= 2
					p.priv_pn = 0
				}
			}
			p.priv_fprev = p.priv_f
			p.priv_fmax = max(p.priv_fmax, p.priv_f)
			if p.priv_f >= p.priv_fopt {
				p.priv_fopt = p.priv_fmax + max(float32(1), qpAbs(p.priv_fmax))
			}
		}
	}
	qpUpdateLoopMultipliers(p)
	qpUpdateBounds(p)
	if (p.loop_num != 0 || p.max_enable != 0) && (p.loop_k >= QPS_MAX_ITER || p.priv_eps < QPS_MINSTEP) {
		p.loop_fail |= 1
	}
	if p.loop_fail != 0 {
		p.loop_done, p.max_done = 1, 1
	}
}
func qps_solve(p *qps_problem_t) {
	if p == nil || !p.initialized {
		panic("qps_init must precede qps_solve")
	}
	p.solveErr = nil
	p.loop_k, p.loop_fail = 0, 0
	qpPrepareVariables(p)
	if p.max_enable != 0 {
		if err := qpPrepareBoxStart(p); err != nil {
			p.solveErr = err
			p.loop_fail = 1
			p.loop_done, p.max_done = 1, 1
			return
		}
	}
	for iteration := 0; ; iteration++ {
		qps_solve_inner(p)
		if p.loop_done != 0 && p.max_done != 0 {
			break
		}
		if iteration+1 >= QPS_MAX_ITER {
			p.loop_fail |= 1
			p.loop_done, p.max_done = 1, 1
			break
		}
	}
	if p.loop_fail != 0 && (p.loop_num != 0 || p.max_enable != 0) {
		if qpRecoverConstraints(p) {
			p.loop_fail = 0
			p.loop_done, p.max_done = 1, 1
		}
	}
	qps_settp(p)
	qps_func(p)
	for i := p.num_cells - 1; i >= 0; i-- {
		p.x[i], p.y[i] = p.priv_tp[2*i], p.priv_tp[2*i+1]
	}
}
func qps_clean(p *qps_problem_t) {
	if p == nil {
		return
	}
	p.priv_ii = nil
	p.priv_cp = nil
	p.priv_g = nil
	p.priv_h = nil
	p.priv_xi = nil
	p.priv_tp = nil
	p.priv_tp2 = nil
	p.priv_pcgt = nil
	p.priv_lt = nil
	p.priv_mxl = nil
	p.priv_mxh = nil
	p.priv_myl = nil
	p.priv_myh = nil
	p.edges = nil
	p.groups = nil
	p.loops = nil
	p.initialized = false
}

// The original subgradient step estimate can collapse before feasible loop/box
// constraints converge. Coordinate maximization of the same quadratic dual is
// a recovery path: each multiplier has a monotone residual, bracketed and solved
// with the existing preconditioned CG inner solver. No objective is substituted.
func qpDualCoordinate(p *qps_problem_t, multiplier *float32, residual func() float32, tolerance float32) bool {
	solve := func(value float32) (float32, bool) {
		*multiplier = value
		p.solveErr = nil
		qps_cgmin(p)
		r := residual()
		return r, p.solveErr == nil && qpFinite(r)
	}
	current := *multiplier
	r := residual()
	if r <= tolerance && (current == 0 || r >= -tolerance) {
		return true
	}
	var low, high float32
	if r > tolerance {
		low = current
		high = max(float32(1), current*2)
		bracketed := false
		for i := 0; i < 60; i++ {
			value, ok := solve(high)
			if !ok {
				return false
			}
			if value <= 0 {
				bracketed = true
				break
			}
			low = high
			high *= 2
			if !qpFinite(high) {
				return false
			}
		}
		if !bracketed {
			return false
		}
	} else {
		high = current
		value, ok := solve(0)
		if !ok {
			return false
		}
		if value <= tolerance {
			return true
		}
	}
	for i := 0; i < 50; i++ {
		mid := low + (high-low)/2
		value, ok := solve(mid)
		if !ok {
			return false
		}
		if qpAbs(value) <= tolerance {
			return true
		}
		if value > 0 {
			low = mid
		} else {
			high = mid
		}
		if mid == low && mid == high {
			break
		}
	}
	_, ok := solve(high)
	return ok
}
func qpDualConverged(p *qps_problem_t) bool {
	for i, residual := range p.priv_lt {
		tolerance := max(float32(QPS_LOOP_TOL), float32(QPS_LOOP_TOL)*p.loop_max[i])
		if residual > tolerance || p.loop_penalty[i] > 0 && residual < -tolerance {
			return false
		}
	}
	if p.max_enable != 0 {
		for i := int32(0); i < p.num_cells; i++ {
			x, y := p.priv_tp[2*i], p.priv_tp[2*i+1]
			residuals := [4]float32{-x, x - p.max_x, -y, y - p.max_y}
			multipliers := [4]float32{p.priv_mxl[i], p.priv_mxh[i], p.priv_myl[i], p.priv_myh[i]}
			for j, r := range residuals {
				if r > QPS_TOL || multipliers[j] > 0 && r < -QPS_TOL {
					return false
				}
			}
		}
	}
	return true
}
func qpRecoverConstraints(p *qps_problem_t) bool {
	clear(p.loop_penalty[:p.loop_num])
	clear(p.priv_mxl)
	clear(p.priv_mxh)
	clear(p.priv_myl)
	clear(p.priv_myh)
	for i := p.num_cells - 1; i >= 0; i-- {
		if index := p.priv_ii[i]; index >= 0 {
			p.priv_cp[index], p.priv_cp[index+1] = p.x[i], p.y[i]
		}
	}
	p.solveErr = nil
	qps_cgmin(p)
	if p.solveErr != nil {
		return false
	}
	for pass := 0; pass < QPS_MAX_ITER; pass++ {
		if qpDualConverged(p) {
			p.loop_fail = 0
			p.loop_done, p.max_done = 1, 1
			return true
		}
		for i := range p.loops {
			index := i
			tolerance := max(float32(QPS_LOOP_TOL), float32(QPS_LOOP_TOL)*p.loop_max[i])
			if !qpDualCoordinate(p, &p.loop_penalty[i], func() float32 { return p.priv_lt[index] }, tolerance) {
				return false
			}
		}
		if p.max_enable != 0 {
			for i := int32(0); i < p.num_cells; i++ {
				cell := i
				multipliers := [4]*float32{&p.priv_mxl[i], &p.priv_mxh[i], &p.priv_myl[i], &p.priv_myh[i]}
				for side, multiplier := range multipliers {
					direction := side
					residual := func() float32 {
						x, y := p.priv_tp[2*cell], p.priv_tp[2*cell+1]
						return [4]float32{-x, x - p.max_x, -y, y - p.max_y}[direction]
					}
					if !qpDualCoordinate(p, multiplier, residual, QPS_TOL) {
						return false
					}
				}
			}
		}
	}
	return qpDualConverged(p)
}

// A feasible box/COG starting point avoids unbounded dual directions caused by
// the translation nullspace of an unanchored wirelength graph. Disjoint COG
// groups admit a common coordinate for all movable members once fixed moments
// are removed. This changes the starting point, not the optimization problem.
func qpPrepareBoxStart(p *qps_problem_t) error {
	x, y := append([]float32(nil), p.x...), append([]float32(nil), p.y...)
	for i := int32(0); i < p.num_cells; i++ {
		if p.fixed[i] != 0 {
			if x[i] < 0 || x[i] > p.max_x || y[i] < 0 || y[i] > p.max_y {
				return fmt.Errorf("fixed cell lies outside QP bounds")
			}
			continue
		}
		x[i] = min(max(x[i], 0), p.max_x)
		y[i] = min(max(y[i], 0), p.max_y)
	}
	for i, group := range p.groups {
		targetX, targetY := float64(p.cog_x[i])*float64(group.total), float64(p.cog_y[i])*float64(group.total)
		var movableArea float64
		for _, v := range group.members {
			if p.fixed[v] != 0 {
				targetX -= float64(p.area[v]) * float64(x[v])
				targetY -= float64(p.area[v]) * float64(y[v])
			} else {
				movableArea += float64(p.area[v])
			}
		}
		cx, cy := float32(targetX/movableArea), float32(targetY/movableArea)
		if !qpFinite(cx) || !qpFinite(cy) || cx < -QPS_TOL || cx > p.max_x+QPS_TOL || cy < -QPS_TOL || cy > p.max_y+QPS_TOL {
			return fmt.Errorf("COG target cannot satisfy QP bounds")
		}
		cx = min(max(cx, 0), p.max_x)
		cy = min(max(cy, 0), p.max_y)
		for _, v := range group.members {
			if p.fixed[v] == 0 {
				x[v], y[v] = cx, cy
			}
		}
	}
	copy(p.x, x)
	copy(p.y, y)
	for i := p.num_cells - 1; i >= 0; i-- {
		if index := p.priv_ii[i]; index >= 0 {
			p.priv_cp[index], p.priv_cp[index+1] = p.x[i], p.y[i]
		}
	}
	return nil
}
