package espresso

var singleOutputMinimum pcover
var singleOutputPhase pcube

func so_espresso(pla pPLA, strategy int32) {
	singleOutputMinimum = new_cover(pla.F.count)
	if strategy == 0 {
		foreach_output_function(pla, so_do_espresso, so_save)
	} else {
		foreach_output_function(pla, so_do_exact, so_save)
	}
	pla.F = singleOutputMinimum
}
func so_both_espresso(pla pPLA, strategy int32) {
	singleOutputPhase = set_save(cube.fullset)
	singleOutputMinimum = new_cover(pla.F.count)
	if strategy == 0 {
		foreach_output_function(pla, so_both_do_espresso, so_both_save)
	} else {
		foreach_output_function(pla, so_both_do_exact, so_both_save)
	}
	pla.F = singleOutputMinimum
	pla.phase = singleOutputPhase
}
func so_do_espresso(pla pPLA, _ int32) bool {
	skip_make_sparse = true
	pla.F = espresso(pla.F, pla.D, pla.R)
	return true
}
func so_do_exact(pla pPLA, _ int32) bool {
	skip_make_sparse = true
	pla.F = minimize_exact(pla.F, pla.D, pla.R, true)
	return true
}
func so_save(pla pPLA, _ int32) bool {
	singleOutputMinimum = sf_append(singleOutputMinimum, pla.F)
	pla.F = nil
	return true
}
func so_both_do_espresso(pla pPLA, i int32) bool {
	so_do_espresso(pla, i)
	skip_make_sparse = true
	pla.R = espresso(pla.R, pla.D, pla.F)
	return true
}
func so_both_do_exact(pla pPLA, i int32) bool {
	so_do_exact(pla, i)
	skip_make_sparse = true
	pla.R = minimize_exact(pla.R, pla.D, pla.F, true)
	return true
}
func so_both_save(pla pPLA, i int32) bool {
	if pla.F.count > pla.R.count {
		pla.F = pla.R
		set_remove(singleOutputPhase, i+cube.first_part[cube.output])
	}
	pla.R = nil
	return so_save(pla, i)
}
