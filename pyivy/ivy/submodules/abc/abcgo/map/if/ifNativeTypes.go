package if_

const (
	IFN_INS = 11
	IFN_WRD = 1 << (IFN_INS - 6)
	IFN_PAR = 1024
)

type Ifn_Obj_t struct {
	Type, nFanins, iFirst, Var uint32
	Fanins                     [IFN_INS]int32
}

// If_DsdMan_t_ owns the caches and scratch vectors used by cut decomposition.
// Vector fields that were embedded C structs remain values; variable-length
// truth/configuration data is represented by Go slices.
type If_DsdMan_t_ struct {
	pStore                                                string
	nVars, LutSize, nWords, nBins                         int32
	pBins                                                 []int32
	pMem                                                  *Mem_Flex_t
	vObjs                                                 Vec_Ptr_t
	vNexts, vTruths                                       Vec_Int_t
	vTemp1, vTemp2                                        *Vec_Int_t
	pTtElems                                              [][]word
	vTtMem                                                [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Mem_t
	vTtDecs                                               [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Ptr_t
	vIsops                                                [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Wec_t
	vIsopComplement                                       [IF_MAX_FUNC_LUTSIZE + 1][]bool
	pSched                                                [IF_MAX_FUNC_LUTSIZE][]int32
	nTtBits, nConfigWords                                 int32
	vConfigs                                              *Vec_Wrd_t
	pTtGia                                                any
	vCover                                                *Vec_Int_t
	pSat                                                  any
	pCellStr                                              string
	nObjsPrev, fNewAsUseless                              int32
	nUniqueHits, nUniqueMisses                            int32
	timeDsd, timeCanon, timeCheck, timeCheck2, timeVerify abctime
}

// Ifn_Ntk_t_ is a bounded local network for truth-table matching. Fixed-size
// records are retained because the source limits this network to IFN_INS inputs.
type Ifn_Ntk_t_ struct {
	nInps, nObjs                               int32
	Nodes                                      [2 * IFN_INS]Ifn_Obj_t
	pConstr                                    [IFN_INS * IFN_INS]int32
	nConstr                                    int32
	nVars, nWords, nParsVNum, nParsVIni, nPars int32
	pTruth                                     []word
	Values                                     [IFN_PAR]int32
	pTtElems                                   [IFN_INS * IFN_WRD]word
	pTtObjs                                    [2 * IFN_INS * IFN_WRD]word
}
