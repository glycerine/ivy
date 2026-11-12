package extra

import (
	"encoding/binary"
	"github.com/glycerine/abcgo/misc/vec"
	"slices"
	"testing"
)

func TestCubeBFSOriginalCounts(t *testing.T) {
	want := []int{1, 10, 64, 385, 2232}
	for _, signature := range []bool{false, true} {
		got, converged := cubeEnumerate(4, 1<<22, signature, false)
		if converged || !slices.Equal(got, want) {
			t.Fatal(signature, got, converged)
		}
	}
	got, converged := cubeEnumerate(5, 20, true, false)
	if converged || !slices.Equal(got, []int{1, 10}) {
		t.Fatal("state limit", got, converged)
	}
}
func TestCubeStateMovesAndPackedWords(t *testing.T) {
	moves := cubeQuarterMoves()
	var identity [24]byte
	for i := range identity {
		identity[i] = byte(i)
	}
	for axis := 0; axis < 3; axis++ {
		state := identity
		for power := 0; power < 4; power++ {
			Abc_StatePerm(state[:], moves[axis][:], state[:])
		}
		if state != identity {
			t.Fatal("quarter turn order")
		}
	}
	data := vec.Vec_IntStart(6)
	Abc_StatePush(data, identity[:], 0)
	for i := 0; i < 6; i++ {
		if uint32(vec.Vec_IntEntry(data, i)) != binary.LittleEndian.Uint32(identity[4*i:]) {
			t.Fatal("packed byte order")
		}
	}
	bits := make([]uint64, 2)
	Abc_DataXorBit(bits, 65)
	if Abc_DataHasBit(bits, 65) != 1 {
		t.Fatal("bit")
	}
	Abc_DataXorBit(bits, 65)
	if bits[1] != 0 {
		t.Fatal("xor")
	}
	for i := int32(0); i < 7; i++ {
		if Abc_DataGetCube(Abc_DataXorCube(0, i, 17), i) != 17 {
			t.Fatal("field")
		}
	}
	signs := map[uint64]bool{Abc_CubeGenerateSign(identity[:]): true}
	for _, move := range moves {
		key := Abc_CubeGenerateSign(move[:])
		if signs[key] {
			t.Fatal("signature collision")
		}
		signs[key] = true
	}
}
