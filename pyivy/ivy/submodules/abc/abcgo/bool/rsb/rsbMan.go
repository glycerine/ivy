package rsb

func Rsb_ManAlloc(leaves, divisors, decompositions, verbose int32) *Rsb_Man_t {
	if leaves < 0 || leaves > 20 || divisors < 0 || divisors > 200 {
		panic("rsb: invalid manager dimensions")
	}
	return &Rsb_Man_t{leafMax: leaves, decMax: decompositions, verbose: verbose != 0, cexes: make([]uint64, 0, divisors+150), patterns: make([]uint32, 0, rsbWords(leaves)), fanins: make([]int32, 0, 4)}
}
func Rsb_ManFree(p *Rsb_Man_t) {
	if p != nil {
		*p = Rsb_Man_t{}
	}
}
func Rsb_ManGetFanins(p *Rsb_Man_t) []int32    { return p.fanins }
func Rsb_ManGetFaninsOld(p *Rsb_Man_t) []int32 { return p.faninsOld }
func Rsb_ManSetFaninsOld(p *Rsb_Man_t, indices []int32) {
	p.faninsOld = append(p.faninsOld[:0], indices...)
}
