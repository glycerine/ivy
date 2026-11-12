package int2

// Int2_ManSetDefaultParams resets p and applies the interpolation defaults.
func Int2_ManSetDefaultParams(p *Int2_ManPars_t) {
	if p == nil {
		panic("int2: nil parameter block")
	}
	*p = Int2_ManPars_t{
		nFramesS:  1,
		nFramesK:  1,
		iFrameMax: -1,
	}
}
