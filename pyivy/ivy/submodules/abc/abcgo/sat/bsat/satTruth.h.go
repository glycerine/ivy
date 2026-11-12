package bsat

/**CFile****************************************************************

  FileName    [satTruth.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Truth table computation package.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: satTruth.h,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

type Tru_Man_t = Tru_Man_t_

func Tru_ManEqual(a, b []word, n int32) int32 {
	for i, v := range a[:n] {
		if v != b[i] {
			return 0
		}
	}
	return 1
}
func Tru_ManEqual0(a []word, n int32) int32 {
	for _, v := range a[:n] {
		if v != 0 {
			return 0
		}
	}
	return 1
}
func Tru_ManEqual1(a []word, n int32) int32 {
	for _, v := range a[:n] {
		if v != ^word(0) {
			return 0
		}
	}
	return 1
}
func Tru_ManCopy(out, in []word, n int32) []word { copy(out[:n], in[:n]); return out }
func Tru_ManClear(out []word, n int32) []word    { clear(out[:n]); return out }
func Tru_ManFill(out []word, n int32) []word {
	for i := range out[:n] {
		out[i] = ^word(0)
	}
	return out
}
func Tru_ManNot(out []word, n int32) []word {
	for i := range out[:n] {
		out[i] = ^out[i]
	}
	return out
}
func Tru_ManAnd(out, in []word, n int32) []word {
	for i := range out[:n] {
		out[i] &= in[i]
	}
	return out
}
func Tru_ManOr(out, in []word, n int32) []word {
	for i := range out[:n] {
		out[i] |= in[i]
	}
	return out
}
func Tru_ManCopyNot(out, in []word, n int32) []word {
	for i := range out[:n] {
		out[i] = ^in[i]
	}
	return out
}
func Tru_ManAndNot(out, in []word, n int32) []word {
	for i := range out[:n] {
		out[i] &= ^in[i]
	}
	return out
}
func Tru_ManOrNot(out, in []word, n int32) []word {
	for i := range out[:n] {
		out[i] |= ^in[i]
	}
	return out
}
func Tru_ManCopyNotCond(out, in []word, n, complement int32) []word {
	if complement != 0 {
		return Tru_ManCopyNot(out, in, n)
	}
	return Tru_ManCopy(out, in, n)
}
func Tru_ManAndNotCond(out, in []word, n, complement int32) []word {
	if complement != 0 {
		return Tru_ManAndNot(out, in, n)
	}
	return Tru_ManAnd(out, in, n)
}
func Tru_ManOrNotCond(out, in []word, n, complement int32) []word {
	if complement != 0 {
		return Tru_ManOrNot(out, in, n)
	}
	return Tru_ManOr(out, in, n)
}
