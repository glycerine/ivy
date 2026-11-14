package if_

import (
	"unsafe"

	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/opt/dau"
)

// truthWordsView adapts the C API's pointer to a truth table. Callers must
// provide at least nWords aligned 64-bit words, as the original C cast did.
func truthWordsView(p *uint32, nWords int) []uint64 {
	if p == nil || nWords < 0 {
		panic("if: invalid truth-table pointer or size")
	}
	return unsafe.Slice((*uint64)(unsafe.Pointer(p)), nWords)
}

func truthWordsAsUint32(words []uint64) []uint32 {
	out := make([]uint32, len(words)*2)
	for i, word := range words {
		ptr := (*[2]uint32)(unsafe.Pointer(&word))
		out[2*i], out[2*i+1] = ptr[0], ptr[1]
	}
	return out
}

// If_CutPerformCheck75__ applies the DSD-based special-cut filter to a truth
// table with six, seven, or eight leaves.
func If_CutPerformCheck75__(p *If_Man_t, pTruth *uint32, nVars, nLeaves int32, pStr string) int32 {
	if nLeaves < 6 {
		return 1
	}
	if nLeaves > 8 {
		panic("if: check75 supports at most eight leaves")
	}
	if p == nil || p.pPars == nil {
		panic("if: check75 requires manager parameters")
	}
	words := truthWordsView(pTruth, int(truthWordCount(nLeaves)))
	if nLeaves < 8 && If_CutPerformCheck16(p, truthWordsAsUint32(words), nVars, nLeaves, "44") != 0 {
		return 1
	}
	var dsd []byte
	nonDecomposable := dau.Dau_DsdDecompose(words, nLeaves, 0, 0, &dsd)
	text := string(dsd)

	switch nLeaves {
	case 8:
		if nonDecomposable >= 5 {
			return 0
		}
		return bool32(Dau_DsdCheckDecAndExist(text)&(1<<4) != 0)
	case 7:
		if nonDecomposable < 5 {
			if Dau_DsdCheckDecExist(text)&(1<<4) != 0 || Dau_DsdCheckDecAndExist(text)&0x18 != 0 {
				return 1
			}
		}
		return check75Cofactors(words, &dsd)
	case 6:
		if nonDecomposable < 5 {
			if Dau_DsdCheckDecExist(text)&0x18 != 0 || Dau_DsdCheckDecAndExist(text)&0x1c != 0 {
				return 1
			}
		}
		return If_CutPerformCheck07(p, truthWordsAsUint32(words), nVars, nLeaves, pStr)
	default:
		panic("if: unreachable check75 leaf count")
	}
}

func check75Cofactors(truth []uint64, dsd *[]byte) int32 {
	for variable := int32(0); variable < 7; variable++ {
		cofactor0 := append([]uint64(nil), truth...)
		cofactor1 := append([]uint64(nil), truth...)
		abcutil.Abc_TtCofactor0(cofactor0, 2, variable)
		abcutil.Abc_TtCofactor1(cofactor1, 2, variable)
		var candidate []uint64
		if abcutil.Abc_TtSupportSize(cofactor0, 7) < 4 {
			candidate = cofactor1
		} else if abcutil.Abc_TtSupportSize(cofactor1, 7) < 4 {
			candidate = cofactor0
		} else {
			continue
		}
		var minVars int32
		If_Dec7MinimumBase(candidate, nil, 7, &minVars)
		if dau.Dau_DsdDecompose(candidate, minVars, 0, 0, dsd) >= 5 {
			continue
		}
		if Dau_DsdCheckDecExist(string(*dsd))&0x18 != 0 {
			return 1
		}
	}
	return 0
}

func truthWordCount(nLeaves int32) int32 {
	if nLeaves <= 6 {
		return 1
	}
	return 1 << (nLeaves - 6)
}

func bool32(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

// If_CutPerformCheck75 combines the DSD check with the existing cut checks.
func If_CutPerformCheck75(p *If_Man_t, pTruth0 *uint32, nVars, nLeaves int32, pStr string) int32 {
	if nLeaves < 0 || nLeaves > 8 {
		panic("if: check75 supports between zero and eight leaves")
	}
	if p == nil || p.pPars == nil {
		panic("if: check75 requires manager parameters")
	}
	truth := append([]uint64(nil), truthWordsView(pTruth0, 4)...)
	if p.pPars.FCutMin == 0 {
		abcutil.Abc_TtMinimumBase(truth, nil, nLeaves, &nLeaves)
	}
	if nLeaves < 6 {
		return 1
	}
	truth32 := truthWordsAsUint32(truth)
	if nLeaves < 8 && If_CutPerformCheck16(p, truth32, nVars, nLeaves, "44") != 0 {
		return 1
	}
	if p.pPars.FDeriveLuts == 0 && p.pPars.FEnableCheck75 != 0 && nLeaves == 8 {
		var dsd []byte
		if dau.Dau_DsdDecompose(truth, nLeaves, 0, 0, &dsd) >= 5 {
			return 0
		}
		return bool32(Dau_DsdCheckDecAndExist(string(dsd))&(1<<4) != 0)
	}
	if If_CutPerformCheck45(p, truth32, nVars, nLeaves, pStr) != 0 {
		return 1
	}
	if If_CutPerformCheck54(p, truth32, nVars, nLeaves, pStr) != 0 {
		return 1
	}
	return 0
}
