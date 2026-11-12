package bsat

import "math"

func sat_solver_set_var_activity(s *sat_solver, variables []int32, n int32) {
	if n < 0 || n > s.size || variables != nil && int32(len(variables)) < n {
		panic("invalid activity variable list")
	}
	clear(s.activity)
	switch s.VarActType {
	case 0:
		s.var_inc = 1 << 5
		s.var_decay = ^uint64(0)
	case 1:
		s.var_inc = math.Float64bits(1)
	default:
		panic("unsupported priority activity mode")
	}
	for i := int32(0); i < n; i++ {
		v := i
		if variables != nil {
			v = variables[i]
		}
		if v < 0 || v >= s.size {
			panic("invalid priority variable")
		}
		if s.VarActType == 0 {
			s.activity[v] = s.var_inc * uint64(n-i)
		} else {
			s.activity[v] = math.Float64bits(float64(n - i))
		}
		if s.orderpos[v] != -1 {
			order_update(s, v)
		}
	}
}
func act_var_rescale(s *sat_solver) {
	switch s.VarActType {
	case 0:
		for i := range s.activity {
			s.activity[i] >>= 19
		}
		s.var_inc >>= 19
		s.var_inc = uint64(max(uint32(s.var_inc), 1<<4))
	case 1:
		for i, v := range s.activity {
			s.activity[i] = math.Float64bits(math.Float64frombits(v) * 1e-100)
		}
		s.var_inc = math.Float64bits(math.Float64frombits(s.var_inc) * 1e-100)
	case 2:
		for i, v := range s.activity {
			s.activity[i] = Xdbl_Div(v, 200)
		}
		s.var_inc = Xdbl_Div(s.var_inc, 200)
	default:
		panic("invalid variable activity mode")
	}
}
func act_var_bump(s *sat_solver, v int32) {
	switch s.VarActType {
	case 0:
		s.activity[v] += s.var_inc
		if uint32(s.activity[v])&0x80000000 != 0 {
			act_var_rescale(s)
		}
	case 1:
		value := math.Float64frombits(s.activity[v]) + math.Float64frombits(s.var_inc)
		s.activity[v] = math.Float64bits(value)
		if value > 1e100 {
			act_var_rescale(s)
		}
	case 2:
		s.activity[v] = Xdbl_Add(s.activity[v], s.var_inc)
		if s.activity[v] > 0x014c924d692ca61b {
			act_var_rescale(s)
		}
	default:
		panic("invalid variable activity mode")
	}
	if s.orderpos[v] != -1 {
		order_update(s, v)
	}
}
func act_var_bump_global(s *sat_solver, v int32) {
	if s.pGlobalVars == nil || s.pGlobalVars[v] == 0 {
		return
	}
	satBumpScaled(s, v, 3, false)
}
func act_var_bump_factor(s *sat_solver, v int32) {
	if s.factors == nil {
		return
	}
	satBumpScaled(s, v, s.factors[v], true)
}
func satBumpScaled(s *sat_solver, v int32, factor float64, asFloat bool) {
	switch s.VarActType {
	case 0:
		delta := int32(uint32(s.var_inc) * 3)
		if asFloat {
			delta = int32(float32(uint32(s.var_inc)) * float32(factor))
		}
		s.activity[v] += uint64(int64(delta))
		if s.activity[v]&0x80000000 != 0 {
			act_var_rescale(s)
		}
	case 1:
		value := math.Float64frombits(s.activity[v]) + math.Float64frombits(s.var_inc)*factor
		s.activity[v] = math.Float64bits(value)
		if value > 1e100 {
			act_var_rescale(s)
		}
	case 2:
		s.activity[v] = Xdbl_Add(s.activity[v], Xdbl_Mul(s.var_inc, Xdbl_FromDouble(factor)))
		if s.activity[v] > 0x014c924d692ca61b {
			act_var_rescale(s)
		}
	default:
		panic("invalid variable activity mode")
	}
	if s.orderpos[v] != -1 {
		order_update(s, v)
	}
}
func act_var_decay(s *sat_solver) {
	switch s.VarActType {
	case 0:
		s.var_inc += s.var_inc >> 4
	case 1:
		s.var_inc = math.Float64bits(math.Float64frombits(s.var_inc) * math.Float64frombits(s.var_decay))
	case 2:
		s.var_inc = Xdbl_Mul(s.var_inc, s.var_decay)
	default:
		panic("invalid variable activity mode")
	}
}
func act_clause_rescale(s *sat_solver) {
	if s.ClaActType == 0 {
		for i := range veci_begin(&s.act_clas) {
			s.act_clas.ptr[i] = int32(uint32(s.act_clas.ptr[i]) >> 14)
		}
		s.cla_inc = max(s.cla_inc>>14, 1<<10)
	} else {
		for i, v := range veci_begin(&s.act_clas) {
			s.act_clas.ptr[i] = int32(math.Float32bits(math.Float32frombits(uint32(v)) * 1e-20))
		}
		s.cla_inc = uint32(float32(s.cla_inc) * 1e-20)
	}
}
func act_clause_bump(s *sat_solver, c *clause) {
	id := clause_id(c)
	if id < 0 || id >= s.act_clas.size {
		panic("invalid clause activity ID")
	}
	if s.ClaActType == 0 {
		value := uint32(s.act_clas.ptr[id]) + s.cla_inc
		s.act_clas.ptr[id] = int32(value)
		if value&0x80000000 != 0 {
			act_clause_rescale(s)
		}
	} else {
		value := math.Float32frombits(uint32(s.act_clas.ptr[id])) + float32(s.cla_inc)
		s.act_clas.ptr[id] = int32(math.Float32bits(value))
		if value > 1e20 {
			act_clause_rescale(s)
		}
	}
}
func act_clause_decay(s *sat_solver) {
	if s.ClaActType == 0 {
		s.cla_inc += s.cla_inc >> 10
	} else {
		s.cla_inc *= s.cla_decay
	}
}
func (s *sat_solver_t) bumpVariable(v int32) { act_var_bump(s, v) }
func (s *sat_solver_t) bumpClause(c *clause) { act_clause_bump(s, c) }
