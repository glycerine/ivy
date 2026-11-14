package mpm

func Mpm_ManSetParsDefault(p *Mpm_Par_t) {
	*p = Mpm_Par_t{nNumCuts: 8, DelayTarget: -1}
}
