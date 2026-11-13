package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCubePairScores(t *testing.T) {
	cubes := vec.Vec_WecAlloc(2)
	for _, v := range []int32{0, 2, 4} {
		vec.Vec_WecPush(cubes, 0, v)
	}
	for _, v := range []int32{0, 2, 6} {
		vec.Vec_WecPush(cubes, 1, v)
	}
	result := vec.Vec_IntArray(Div_CubePairs(cubes, 4, 1))
	if len(result) != 4 || result[0] != int32(DIV_AND) || result[1] != 0 || result[2] != 2 {
		t.Fatalf("best divisor %v", result)
	}
}
