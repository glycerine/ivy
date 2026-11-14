package gia

// Gia_ManTransferTiming moves source timing/name/configuration ownership.
// ABC checks self-transfer only after the non-box metadata groups.
func Gia_ManTransferTiming(dst, src *Gia_Man_t) {
	if src.vCiArrs != nil || src.vCoReqs != nil || src.vCoArrs != nil || src.vCoAttrs != nil {
		dst.vCiArrs, src.vCiArrs = src.vCiArrs, nil
		dst.vCoReqs, src.vCoReqs = src.vCoReqs, nil
		dst.vCoArrs, src.vCoArrs = src.vCoArrs, nil
		dst.vCoAttrs, src.vCoAttrs = src.vCoAttrs, nil
		dst.And2Delay = src.And2Delay
	}
	if src.VInArrs != nil || src.VOutReqs != nil {
		dst.VInArrs, src.VInArrs = src.VInArrs, nil
		dst.VOutReqs, src.VOutReqs = src.VOutReqs, nil
		dst.DefInArrs, dst.DefOutReqs, dst.And2Delay = src.DefInArrs, src.DefOutReqs, src.And2Delay
	}
	if src.VNamesIn != nil || src.VNamesOut != nil {
		dst.VNamesIn, src.VNamesIn = src.VNamesIn, nil
		dst.VNamesOut, src.VNamesOut = src.VNamesOut, nil
	}
	// The Go string representation cannot distinguish a C NULL pointer from
	// a non-NULL pointer to an empty string unless configs are also present.
	if src.VConfigs != nil || src.PCellStr != "" {
		dst.VConfigs, src.VConfigs = src.VConfigs, nil
		dst.PCellStr, src.PCellStr = src.PCellStr, ""
	}
	if src.PManTime == nil || dst == src {
		return
	}
	dst.PManTime, src.PManTime = src.PManTime, nil
	dst.PAigExtra, src.PAigExtra = src.PAigExtra, nil
	dst.vRegClasses, src.vRegClasses = src.vRegClasses, nil
	dst.vRegInits, src.vRegInits = src.vRegInits, nil
	dst.nAnd2Delay, src.nAnd2Delay = src.nAnd2Delay, 0
}
