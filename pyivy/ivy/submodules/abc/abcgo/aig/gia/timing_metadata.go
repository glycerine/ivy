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
	if src.vInArrs != nil || src.vOutReqs != nil {
		dst.vInArrs, src.vInArrs = src.vInArrs, nil
		dst.vOutReqs, src.vOutReqs = src.vOutReqs, nil
		dst.DefInArrs, dst.DefOutReqs, dst.And2Delay = src.DefInArrs, src.DefOutReqs, src.And2Delay
	}
	if src.vNamesIn != nil || src.vNamesOut != nil {
		dst.vNamesIn, src.vNamesIn = src.vNamesIn, nil
		dst.vNamesOut, src.vNamesOut = src.vNamesOut, nil
	}
	// The Go string representation cannot distinguish a C NULL pointer from
	// a non-NULL pointer to an empty string unless configs are also present.
	if src.vConfigs != nil || src.pCellStr != "" {
		dst.vConfigs, src.vConfigs = src.vConfigs, nil
		dst.pCellStr, src.pCellStr = src.pCellStr, ""
	}
	if src.pManTime == nil || dst == src {
		return
	}
	dst.pManTime, src.pManTime = src.pManTime, nil
	dst.pAigExtra, src.pAigExtra = src.pAigExtra, nil
	dst.vRegClasses, src.vRegClasses = src.vRegClasses, nil
	dst.vRegInits, src.vRegInits = src.vRegInits, nil
	dst.nAnd2Delay, src.nAnd2Delay = src.nAnd2Delay, 0
}
