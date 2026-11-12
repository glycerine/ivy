package bsat

func sat_solver_lit_removable(s *sat_solver, x, minl int32) int32 {
	if satLiteralRemovable(&s.satAssignments, &s.Mem, x, uint32(minl)) {
		return 1
	}
	return 0
}

func sat_solver_analyze_final(s *sat_solver, h int32, skip int32) {
	result := satAnalyzeFinal(&s.satAssignments, &s.Mem, h, s.root_level, skip != 0)
	veci_resize(&s.conf_final, 0)
	for _, l := range result {
		veci_push(&s.conf_final, l)
	}
}

func sat_solver_analyze(s *sat_solver, h int32, learnt *veci) {
	result := satAnalyze(&s.satAssignments, &s.Mem, h, s, &s.stats)
	veci_resize(learnt, 0)
	for _, l := range result {
		veci_push(learnt, l)
	}
}
