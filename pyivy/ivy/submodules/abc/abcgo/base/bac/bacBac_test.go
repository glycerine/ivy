package bac

import (
	"bytes"
	"encoding/binary"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"reflect"
	"testing"
)

func TestNetworkBinaryLayout(t *testing.T) {
	var n Bac_Ntk_t
	vec.Vec_StrFill(&n.vType, 3, 0)
	// The object kind is stored as an ABC literal; retain its low flag bit.
	copy(vec.Vec_StrArray(&n.vType), []byte{2, 5, 10})
	vec.Vec_IntFill(&n.vFanin, 3, 0)
	copy(vec.Vec_IntArray(&n.vFanin), []int32{-1, 42, -3})
	vec.Vec_IntFill(&n.vInfo, 3, 0)
	copy(vec.Vec_IntArray(&n.vInfo), []int32{7, -9, 11})
	out := vec.Vec_StrAlloc(0)
	Bac_ManWriteBacNtk(out, &n)
	want := []byte{2, 5, 10}
	for _, v := range []int32{-1, 42, -3, 7, -9, 11} {
		var b [4]byte
		binary.NativeEndian.PutUint32(b[:], uint32(v))
		want = append(want, b[:]...)
	}
	if !reflect.DeepEqual(vec.Vec_StrArray(out), want) {
		t.Fatalf("bytes %v want %v", vec.Vec_StrArray(out), want)
	}
	var got Bac_Ntk_t
	vec.Vec_StrGrow(&got.vType, 3)
	vec.Vec_IntGrow(&got.vFanin, 3)
	vec.Vec_IntGrow(&got.vInfo, 3)
	vec.Vec_IntGrow(&got.vInputs, 1)
	vec.Vec_IntGrow(&got.vOutputs, 1)
	pos := int32(0)
	Bac_ManReadBacNtk(out, &pos, &got)
	if int(pos) != len(want) || !reflect.DeepEqual(vec.Vec_IntArray(&got.vFanin), []int32{-1, 42, -3}) || !reflect.DeepEqual(vec.Vec_IntArray(&got.vInfo), []int32{7, -9, 11}) || !reflect.DeepEqual(vec.Vec_IntArray(&got.vInputs), []int32{0}) || !reflect.DeepEqual(vec.Vec_IntArray(&got.vOutputs), []int32{1}) {
		t.Fatal(pos, vec.Vec_IntArray(&got.vFanin), vec.Vec_IntArray(&got.vInfo), vec.Vec_IntArray(&got.vInputs), vec.Vec_IntArray(&got.vOutputs))
	}
}

func TestReadBacFixtureFromABC(t *testing.T) {
	// test_vectors/simple.bac was emitted by ABC's @_read/@_write commands
	// from test_vectors/simple.blif on a little-endian system.
	if binary.NativeEndian.Uint32([]byte{1, 0, 0, 0}) != 1 {
		t.Skip("source fixture uses little-endian integer storage")
	}
	man := Bac_ManReadBac("test_vectors/simple.bac")
	if man == nil {
		t.Fatal("source BAC fixture rejected")
	}
	defer Bac_ManFree(man)
	if Bac_ManNtkNum(man) != 1 {
		t.Fatalf("network count %d", Bac_ManNtkNum(man))
	}
	ntk := Bac_ManNtk(man, 1)
	if Bac_NtkName(ntk) != "top" || Bac_NtkPiNum(ntk) != 2 || Bac_NtkPoNum(ntk) != 1 || Bac_NtkObjNum(ntk) != 7 || Bac_NtkInfoNum(ntk) != 3 {
		t.Fatalf("unexpected network: %s %d %d %d %d", Bac_NtkName(ntk), Bac_NtkPiNum(ntk), Bac_NtkPoNum(ntk), Bac_NtkObjNum(ntk), Bac_NtkInfoNum(ntk))
	}
	before, err := os.ReadFile("test_vectors/simple.bac")
	if err != nil {
		t.Fatal(err)
	}
	out := vec.Vec_StrAlloc(0)
	Bac_ManWriteBacInt(out, man)
	after := vec.Vec_StrArray(out)
	// The timestamped comment is intentionally different between runs.
	before = before[bytes.IndexByte(before, '\n')+1:]
	after = after[bytes.IndexByte(after, '\n')+1:]
	if !bytes.Equal(before, after) {
		t.Fatalf("round-trip bytes differ: got %x, want %x", after, before)
	}
}
