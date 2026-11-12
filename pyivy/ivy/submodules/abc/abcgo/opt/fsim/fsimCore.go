package fsim

import "time"

func Fsim_ManSetDefaultParamsSim(p *Fsim_ParSim_t) {
	*p = Fsim_ParSim_t{Words: 8, Iterations: 32, TimeLimit: time.Minute, Verbose: true}
}
func Fsim_ManSetDefaultParamsSwitch(p *Fsim_ParSwitch_t) {
	*p = Fsim_ParSwitch_t{Words: 1, Iterations: 48, Prefix: 16, ProbabilityOne: true, ProbabilityTransition: true, Verbose: true}
}
