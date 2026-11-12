package if_

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"math/rand"
	"testing"
)

func TestLogCounterNativeIndependentSum(t *testing.T) {
	r := rand.New(rand.NewSource(239))
	for trial := 0; trial < 500; trial++ {
		times := make([]int32, 64)
		var n int32
		var sum uint64
		for k := 0; k < 32; k++ {
			num := int32(r.Intn(30))
			sum += 1 << uint(num)
			got := If_LogCounterAdd(times, &n, num, 0)
			want := int32(bits.Len64(sum - 1))
			if got != want {
				t.Fatal("log counter", got, want)
			}
			if If_LogCounter64Eval(sum) != want {
				t.Fatal("word counter")
			}
		}
	}
}
func TestLogCounterNativePinProfilesAndAig(t *testing.T) {
	times := make([]int32, 16)
	profiles := make([]uint64, 16)
	var n int32
	for i := int32(0); i < 4; i++ {
		If_LogCounterPinDelays(times, &n, profiles, 0, If_CutPinDelayInit(i), 4, 0)
	}
	if n != 1 || times[0] != 2 {
		t.Fatal("balanced counter")
	}
	for i := int32(0); i < 4; i++ {
		if If_CutPinDelayGet(profiles[0], i) != 3 {
			t.Fatal("pin profile")
		}
	}
	for _, xor := range []int32{0, 1} {
		aig := vec.Vec_IntAlloc(16)
		times := make([]int32, 16)
		lits := make([]int32, 16)
		n = 0
		for i := int32(0); i < 4; i++ {
			If_LogCounterAddAig(times, &n, lits, 0, i*2, aig, 4, xor, xor)
		}
		if n != 1 {
			t.Fatal("unbalanced gate collection")
		}
		vec.Vec_IntPush(aig, lits[0]&1)
		truth := If_AigVerifyArray(aig, 4)
		for assignment := 0; assignment < 64; assignment++ {
			want := assignment&15 == 15
			if xor != 0 {
				want = bits.OnesCount(uint(assignment&15))&1 != 0
			}
			if (truth>>uint(assignment)&1 != 0) != want {
				t.Fatal("counter AIG truth", xor, assignment)
			}
		}
	}
}
func TestLogCounterNativeWideShift(t *testing.T) {
	var start int32
	count := If_LogCounter32Add(0, &start, 0)
	count = If_LogCounter32Add(count, &start, 100)
	if count != 1<<16 || start>>1 != 84 || start&1 == 0 || If_LogCounter32Eval(count, start) != 101 {
		t.Fatal("large shift lost sticky information")
	}
}
