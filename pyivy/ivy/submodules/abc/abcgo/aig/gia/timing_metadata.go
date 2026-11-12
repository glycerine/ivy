package gia

// Gia_ManTransferTiming moves source timing/name/configuration ownership. A
// self-transfer is a no-op, including metadata processed before box timing.
func Gia_ManTransferTiming(dst, src *Gia_Man_t) {
	if dst == src {
		return
	}
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
	if src.vConfigs != nil || src.pCellStr != "" {
		dst.vConfigs, src.vConfigs = src.vConfigs, nil
		dst.pCellStr, src.pCellStr = src.pCellStr, ""
	}
	if src.pManTime == nil {
		return
	}
	dst.pManTime, src.pManTime = src.pManTime, nil
	dst.pAigExtra, src.pAigExtra = src.pAigExtra, nil
	dst.vRegClasses, src.vRegClasses = src.vRegClasses, nil
	dst.vRegInits, src.vRegInits = src.vRegInits, nil
	dst.nAnd2Delay, src.nAnd2Delay = src.nAnd2Delay, 0
}
