package mpm

func Mpm_ManSetParsDefault(p *Mpm_Par_t) {
	*p = Mpm_Par_t{NNumCuts: 8, DelayTarget: -1}
}
