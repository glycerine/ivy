package if_

// IfifParams is the subset of parameters consumed by the experimental IFIF
// mapper in abcIfif.c.
type IfifParams struct {
	LUTSize, Degree               int32
	Cascade, Verbose, VeryVerbose int32
	WireDelay                     float32
}

func IfifSetLibrary(p *Ifif_Par_t, lib *If_LibLut_t) { p.pLutLib = lib }
func IfifSetParams(p *Ifif_Par_t, values IfifParams) {
	p.nLutSize, p.nDegree = values.LUTSize, values.Degree
	p.fCascade, p.fVerbose, p.fVeryVerbose = values.Cascade, values.Verbose, values.VeryVerbose
	p.DelayWire = values.WireDelay
}
func IfifReadParams(p *Ifif_Par_t) IfifParams {
	return IfifParams{LUTSize: p.nLutSize, Degree: p.nDegree, Cascade: p.fCascade, Verbose: p.fVerbose, VeryVerbose: p.fVeryVerbose, WireDelay: p.DelayWire}
}
func IfifReadLibrary(p *Ifif_Par_t) *If_LibLut_t              { return p.pLutLib }
func IfifSetPinDelay(p *Ifif_Par_t, pin int32, delay float32) { p.pLutDelays[pin] = delay }
func IfifPinDelay(p *Ifif_Par_t, pin int32) float32           { return p.pLutDelays[pin] }
func If_LibLutHasVariablePinDelays(lib *If_LibLut_t) bool     { return lib.fVarPinDelays != 0 }
func If_LibLutDelay(lib *If_LibLut_t, lutSize, pin int32) float32 {
	return lib.pLutDelays[lutSize][pin]
}
