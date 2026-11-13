package framgr

// Fra_ParamsDefault initializes the combinational FRAIG configuration.
func Fra_ParamsDefault(p *Fra_Par_t) {
	*p = Fra_Par_t{
		nSimWords: 32, dSimSatur: 0.005, MaxScore: 25, fDoSparse: 1,
		dActConeRatio: 0.3, dActConeBumpMax: 10,
		nBTLimitNode: 100, nBTLimitMiter: 500000, fConeBias: 1,
	}
}

// Fra_ParamsDefaultSeq initializes the sequential FRAIG configuration.
func Fra_ParamsDefaultSeq(p *Fra_Par_t) {
	*p = Fra_Par_t{
		nSimWords: 1, dSimSatur: 0.005, MaxScore: 25, fDoSparse: 1,
		dActConeRatio: 0.3, dActConeBumpMax: 10,
		nBTLimitNode: 10000000, nBTLimitMiter: 500000, nFramesK: 1,
	}
}
