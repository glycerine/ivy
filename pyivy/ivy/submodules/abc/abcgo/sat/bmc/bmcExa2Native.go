package bmc

// Exa_ManExactSynthesis2 is the BSAT variant of two-input exact synthesis.
// The source variant does not write a BLIF file or impose a runtime deadline.
func Exa_ManExactSynthesis2(pars *Bmc_EsPar_t) {
	exaManExactSynthesis(pars, false, false)
}
