package gia

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestIiffSourceDeriveStubPreservesOutputs(t *testing.T) {
	if got := Gia_ManDeriveMatches(nil, nil, nil, nil, nil, nil); got != 0 {
		t.Fatalf("nil-argument source stub returned%d", got)
	}
	names := vec.Vec_PtrAllocArrayCopy([]any{"keep"}, 1)
	truths := vec.Vec_WrdAllocArrayCopy([]uint64{0x123456789abcdef0}, 1)
	match4 := vec.Vec_IntAllocArrayCopy([]int32{3, 4}, 2)
	configs := vec.Vec_IntAllocArrayCopy([]int32{17, 19}, 2)
	wantNames, wantTruths, wantMatch4, wantConfigs := names, truths, match4, configs
	memories := make([]*Vec_Mem_t, 3)
	matches := make([]*Vec_Int_t, 3)
	for i := range memories {
		memories[i] = vec.Vec_MemAlloc(1, 4)
		vec.Vec_MemPush(memories[i], []uint64{uint64(0x100 + i)})
		matches[i] = vec.Vec_IntAllocArrayCopy([]int32{int32(30 + i)}, 1)
	}
	wantMemories, wantMatches := slices.Clone(memories), slices.Clone(matches)
	if got := Gia_ManDeriveMatches(&names, &truths, &match4, &configs, memories, matches); got != 0 {
		t.Fatalf("populated-argument source stub returned%d", got)
	}
	if names != wantNames || truths != wantTruths || match4 != wantMatch4 || configs != wantConfigs || !slices.Equal(memories, wantMemories) || !slices.Equal(matches, wantMatches) {
		t.Fatal("source stub replaced an output pointer")
	}
	if !slices.Equal(vec.Vec_PtrArray(names), []any{"keep"}) || !slices.Equal(vec.Vec_WrdArray(truths), []uint64{0x123456789abcdef0}) || !slices.Equal(vec.Vec_IntArray(match4), []int32{3, 4}) || !slices.Equal(vec.Vec_IntArray(configs), []int32{17, 19}) {
		t.Fatal("source stub changed an output vector")
	}
	for i, memory := range memories {
		if vec.Vec_MemEntryNum(memory) != 1 || vec.Vec_MemEntrySize(memory) != 1 || vec.Vec_MemReadEntry(memory, 0)[0] != uint64(0x100+i) || !slices.Equal(vec.Vec_IntArray(matches[i]), []int32{int32(30 + i)}) {
			t.Fatal("source stub changed a paged truth or matching table")
		}
	}
}

func TestIiffSourceTestStubNeedsNoFileAndPreservesGraph(t *testing.T) {
	if got := Gia_ManIiffTest("", nil, 0, 0, 0, 0, 0, 0); got != nil {
		t.Fatal("nil-argument source stub returned a graph")
	}
	graph := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
	Gia_ManAppendCo(graph, Gia_ManAppendAnd(graph, a, b^1))
	objects := Gia_ManObjNum(graph)
	directory := t.TempDir()
	filename := filepath.Join(directory, "absent.input")
	for _, flag := range []int32{-1, 0, 1} {
		if got := Gia_ManIiffTest(filename, graph, 4, 8, flag, flag, flag, flag); got != nil {
			t.Fatal("source stub invented a matching result")
		}
	}
	if Gia_ManObjNum(graph) != objects || graph.VMapping != nil || graph.VPacking != nil {
		t.Fatal("source stub modified the borrowed graph")
	}
	for assignment := 0; assignment < 4; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0}
		got, err := graph.Evaluate(inputs)
		if err != nil || len(got) != 1 || got[0] != (inputs[0] && !inputs[1]) {
			t.Fatalf("borrowed graph changed: assignment%d got%v err%v", assignment, got, err)
		}
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 0 {
		t.Fatalf("source stub changed its file directory: entries%d err%v", len(entries), err)
	}
}
