package abs

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestVtaFramePackingRoundTrip(t *testing.T) {
	for _, packed := range [][]int32{
		{0, 2},
		{1, 3, 3},
		{3, 5, 7, 7, 10, 9, 12, 3, 4, 5},
	} {
		input := vec.Vec_IntAllocArrayCopy(packed, len(packed))
		frames := Gia_VtaAbsToFrames(input)
		output := Gia_VtaFramesToAbs(frames)
		if got := vec.Vec_IntArray(output); !reflect.DeepEqual(got, packed) {
			t.Fatalf("round trip: got %v, want %v", got, packed)
		}
		if vec.Vec_PtrSize(frames) != 0 {
			frame := vec.Vec_VecEntryInt(frames, 0)
			if vec.Vec_IntSize(frame) != 0 {
				vec.Vec_IntWriteEntry(frame, 0, -77)
				if vec.Vec_IntArray(input)[int(packed[1])] == -77 {
					t.Fatal("unpacked frame aliases input")
				}
			}
		}
	}
}

func TestVtaFramePackingRejectsInvalidOffsets(t *testing.T) {
	for _, packed := range [][]int32{{}, {1}, {2, 4, 6, 5, 1}, {1, 2, 3}, {1, 3, 4}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("accepted invalid offsets %v", packed)
				}
			}()
			Gia_VtaAbsToFrames(vec.Vec_IntAllocArrayCopy(packed, len(packed)))
		}()
	}
}

func TestVtaWordsWidenAndCollect(t *testing.T) {
	data := []int32{0, 1, 0, 0, -1, 2}
	words := vec.Vec_IntAllocArrayCopy(data, len(data))
	if got := vec.Vec_IntArray(Vta_ManDeriveAbsAll(words, 2)); !reflect.DeepEqual(got, []int32{1, 0, 1}) {
		t.Fatalf("object mask %v", got)
	}
	if width := Vec_IntDoubleWidth(words, 2); width != 4 {
		t.Fatalf("width %d", width)
	}
	if got := vec.Vec_IntArray(words); !reflect.DeepEqual(got, []int32{0, 1, 0, 0, 0, 0, 0, 0, -1, 2, 0, 0}) {
		t.Fatalf("widened %v", got)
	}
	if got := vec.Vec_IntArray(Vta_ManDeriveAbsAll(words, 4)); !reflect.DeepEqual(got, []int32{1, 0, 1}) {
		t.Fatalf("widened object mask %v", got)
	}
}
