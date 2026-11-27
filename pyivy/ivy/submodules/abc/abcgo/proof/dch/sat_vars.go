package dch

// satVars maps AIG node IDs to SAT variables (always positive) as a dense
// slice, ABC's pSatVars array; 0 marks a node without a variable.
type satVars struct {
	vars []int32
	n    int
}

func (s *satVars) get(id int) (int32, bool) {
	if id < len(s.vars) && s.vars[id] > 0 {
		return s.vars[id], true
	}
	return 0, false
}

// at is id's variable, 0 when it has none.
func (s *satVars) at(id int) int32 {
	v, _ := s.get(id)
	return v
}

func (s *satVars) set(id int, v int32) {
	if v <= 0 {
		panic("dch: SAT variables are positive")
	}
	if id >= len(s.vars) {
		s.vars = append(s.vars, make([]int32, max(id+1, 2*len(s.vars))-len(s.vars))...)
	}
	if s.vars[id] == 0 {
		s.n++
	}
	s.vars[id] = v
}

func (s *satVars) len() int { return s.n }
