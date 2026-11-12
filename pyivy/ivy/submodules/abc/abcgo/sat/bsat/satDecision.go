package bsat

const (
	var0 int32 = 1
	var1 int32 = 0
	varX int32 = satUnassigned
)

func var_level(s *sat_solver, v int32) int32 { return s.levels[v] }
func var_value(s *sat_solver, v int32) int32 { return s.assigns[v] }
func var_polar(s *sat_solver, v int32) int32 { return s.polarity[v] }

func var_set_level(s *sat_solver, v int32, lev int32) { s.levels[v] = lev }
func var_set_value(s *sat_solver, v int32, val int32) { s.assigns[v] = val }
func var_set_polar(s *sat_solver, v int32, pol int32) { s.polarity[v] = pol }

// variable tags
func var_tag(s *sat_solver, v int32) int32 { return s.tags[v] }
func var_set_tag(s *sat_solver, v int32, tag int32) {
	assert(tag > 0 && tag < 16, "tag > 0 && tag < 16")
	if s.tags[v] == 0 {
		veci_push(&s.tagged, v)
	}
	s.tags[v] = tag
}
func var_add_tag(s *sat_solver, v int32, tag int32) {
	assert(tag > 0 && tag < 16, "tag > 0 && tag < 16")
	if s.tags[v] == 0 {
		veci_push(&s.tagged, v)
	}
	s.tags[v] |= tag
}
func solver2_clear_tags(s *sat_solver, start int32) {
	var i int32
	tagged := veci_begin(&s.tagged)
	for i = start; i < veci_size(&s.tagged); i++ {
		s.tags[tagged[i]] = 0
	}
	veci_resize(&s.tagged, start)
}

func sat_solver_get_var_value(s *sat_solver, v int32) int32 {
	if var_value(s, v) == var0 {
		return l_False
	}
	if var_value(s, v) == var1 {
		return l_True
	}
	if var_value(s, v) == varX {
		return l_Undef
	}
	assert(false, "0")
	return 0
}

//=================================================================================================
// Simple helpers:

func sat_solver_dl(s *sat_solver) int32                { return veci_size(&s.trail_lim) }
func sat_solver_read_wlist(s *sat_solver, l lit) *veci { return &s.wlists[l] }

//=================================================================================================
// Variable order functions:

func order_update(s *sat_solver, v int32) {
	satOrderUpdate(veci_begin(&s.order), s.orderpos, s.activity, v)
}

func order_assigned(s *sat_solver, v int32) {
}

func order_unassigned(s *sat_solver, v int32) {
	if s.orderpos[v] == -1 {
		s.orderpos[v] = veci_size(&s.order)
		veci_push(&s.order, v)
		order_update(s, v)
	}
}
func order_select(s *sat_solver, random_var_freq float64) int32 {
	if s.size > 0 && drand(&s.random_seed) < random_var_freq {
		next := irand(&s.random_seed, s.size)
		if var_value(s, next) == varX {
			return next
		}
	}
	for s.order.size > 0 {
		heap, next := satOrderPop(veci_begin(&s.order), s.orderpos, s.activity)
		s.order.size = int32(len(heap))
		if var_value(s, next) == varX {
			return next
		}
	}
	return var_Undef
}
