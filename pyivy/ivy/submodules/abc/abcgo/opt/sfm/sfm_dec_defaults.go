package sfm

// Sfm_ParSetDefault3 initializes parameters for mapped-network decomposition.
func Sfm_ParSetDefault3(p *Sfm_Par_t) {
	*p = Sfm_Par_t{TfoLevMax: 100, TfiLevMax: 100, FanoutMax: 10, MffcMin: 1, MffcMax: 3, VarMax: 6, DecMax: 1, TimeWin: 1}
}
