package ssw

import "time"

func sswDeadlineExpired(p *Ssw_Pars_t) bool {
	return p != nil && !p.deadline.IsZero() && !time.Now().Before(p.deadline)
}
func sswSatStart(p *Ssw_Pars_t) *Ssw_Sat_t {
	s := Ssw_SatStart(p.fPolarFlip)
	s.pSat.SetDeadline(p.deadline)
	return s
}
