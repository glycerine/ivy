package abci

// Three-valued primitives from abcXsim.c.
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abc"
)

const (
	XVS0 = int32(abc.ABC_INIT_ZERO)
	XVS1 = int32(abc.ABC_INIT_ONE)
	XVSX = int32(abc.ABC_INIT_DC)
)

func Abc_XsimInv(value int32) int32 {
	if value == XVS0 {
		return XVS1
	}
	if value == XVS1 {
		return XVS0
	}
	if value != XVSX {
		panic("abci: invalid X simulation value")
	}
	return XVSX
}
func Abc_XsimAnd(a, b int32) int32 {
	if a == XVS0 || b == XVS0 {
		return XVS0
	}
	if a == XVSX || b == XVSX {
		return XVSX
	}
	if a != XVS1 || b != XVS1 {
		panic("abci: invalid X simulation value")
	}
	return XVS1
}
func Abc_XsimRand2() int32 {
	if gia.Gia_ManRandom(0)&1 != 0 {
		return XVS1
	}
	return XVS0
}
func Abc_XsimRand3() int32 {
	for {
		v := int32(gia.Gia_ManRandom(0) & 3)
		if v != 0 {
			return v
		}
	}
}
