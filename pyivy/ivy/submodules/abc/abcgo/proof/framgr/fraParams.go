package framgr

// Fra_ParamsDefault initializes the combinational FRAIG configuration.
func Fra_ParamsDefault(p *Fra_Par_t) {
	*p = Fra_Par_t{
		NSimWords: 32, DSimSatur: 0.005, MaxScore: 25, FDoSparse: 1,
		DActConeRatio: 0.3, DActConeBumpMax: 10,
		NBTLimitNode: 100, NBTLimitMiter: 500000, FConeBias: 1,
	}
}

// Fra_ParamsDefaultSeq initializes the sequential FRAIG configuration.
func Fra_ParamsDefaultSeq(p *Fra_Par_t) {
	*p = Fra_Par_t{
		NSimWords: 1, DSimSatur: 0.005, MaxScore: 25, FDoSparse: 1,
		DActConeRatio: 0.3, DActConeBumpMax: 10,
		NBTLimitNode: 10000000, NBTLimitMiter: 500000, NFramesK: 1,
	}
}
