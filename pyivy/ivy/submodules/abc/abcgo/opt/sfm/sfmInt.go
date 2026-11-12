package sfm

func Sfm_NtkPiNum(p *Sfm_Ntk_t) int               { return p.Pis }
func Sfm_NtkPoNum(p *Sfm_Ntk_t) int               { return p.Pos }
func Sfm_NtkNodeNum(p *Sfm_Ntk_t) int             { return len(p.Fanins) - p.Pis - p.Pos }
func Sfm_ObjIsPi(p *Sfm_Ntk_t, id int) bool       { return id < p.Pis }
func Sfm_ObjIsPo(p *Sfm_Ntk_t, id int) bool       { return id >= len(p.Fanins)-p.Pos }
func Sfm_ObjIsNode(p *Sfm_Ntk_t, id int) bool     { return id >= p.Pis && id < len(p.Fanins)-p.Pos }
func Sfm_ObjIsFixed(p *Sfm_Ntk_t, id int) bool    { return p.Fixed[id] }
func Sfm_ObjAddsLevel(p *Sfm_Ntk_t, id int) bool  { return !p.Empty[id] }
func Sfm_ObjFaninNum(p *Sfm_Ntk_t, id int) int    { return len(p.Fanins[id]) }
func Sfm_ObjFanoutNum(p *Sfm_Ntk_t, id int) int   { return len(p.Fanouts[id]) }
func Sfm_ObjFanin(p *Sfm_Ntk_t, id, pin int) int  { return p.Fanins[id][pin] }
func Sfm_ObjFanout(p *Sfm_Ntk_t, id, pin int) int { return p.Fanouts[id][pin] }
func Sfm_ObjLevel(p *Sfm_Ntk_t, id int) int       { return p.Levels[id] }
func Sfm_ObjLevelR(p *Sfm_Ntk_t, id int) int      { return p.Reverse[id] }
func Sfm_ObjSatVar(p *Sfm_Ntk_t, id int) int {
	v, ok := p.idVar[id]
	if !ok {
		panic("sfm: object not in SAT window")
	}
	return v
}
func truthWords(vars int) int {
	if vars <= 6 {
		return 1
	}
	return 1 << uint(vars-6)
}
func bit(truth []uint64, a int) bool { return truth[a>>6]&(uint64(1)<<uint(a&63)) != 0 }
func set(truth []uint64, a int)      { truth[a>>6] |= uint64(1) << uint(a&63) }
func stretch(t uint64, vars int) uint64 {
	if vars >= 6 {
		return t
	}
	n := 1 << uint(vars)
	t &= (uint64(1) << uint(n)) - 1
	for size := n; size < 64; size *= 2 {
		t |= t << uint(size)
	}
	return t
}
