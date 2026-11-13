package fraig

import (
	"testing"
	"unsafe"
)

// C allocates one mmSims entry per node plus three feedback buffers. This
// checks manager ownership/accounting, rather than only testing the pool alone.
func TestManagerSimulationPoolAccountingAndResize(t *testing.T) {
	params := &Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, FFuncRed: 1, FFeedBack: 1}
	p := Fraig_ManCreate(params)
	defer Fraig_ManFree(p)
	entryBytes := int32(4 * (p.nWordsRand + p.nWordsDyna))
	if got, want := Fraig_MemFixedReadMemUsage(p.mmSims), entryBytes*1024; got != want {
		t.Fatalf("initial manager pool bytes = %d, want %d", got, want)
	}
	if got, want := fraigNodePoolMemUsage(p.mmNodes), uintptr(p.mmNodes.chunkEntries)*uintptr(unsafe.Sizeof(Fraig_Node_t{})); got != want {
		t.Fatalf("initial node pool bytes = %d, want %d", got, want)
	}
	for i := 0; i < 1021; i++ {
		Fraig_NodeCreatePi(p)
	}
	if got, want := Fraig_MemFixedReadMemUsage(p.mmSims), entryBytes*2048; got != want {
		t.Fatalf("pool bytes after crossing chunk boundary = %d, want %d", got, want)
	}
	nodeSize := unsafe.Sizeof(Fraig_Node_t{})
	nodeChunks := (uintptr(1022) + uintptr(p.mmNodes.chunkEntries) - 1) / uintptr(p.mmNodes.chunkEntries)
	if got, want := fraigNodePoolMemUsage(p.mmNodes), nodeChunks*uintptr(p.mmNodes.chunkEntries)*nodeSize; got != want {
		t.Fatalf("node pool bytes after growth = %d, want %d", got, want)
	}

	oldWords := p.nWordsDyna
	Fraig_ReallocateSimulationInfo(p)
	if p.nWordsDyna != oldWords*2 {
		t.Fatalf("dynamic words after resize = %d, want %d", p.nWordsDyna, oldWords*2)
	}
	resizedEntryBytes := int32(4 * (p.nWordsRand + p.nWordsDyna))
	if got, want := Fraig_MemFixedReadMemUsage(p.mmSims), resizedEntryBytes*2048; got != want {
		t.Fatalf("resized manager pool bytes = %d, want %d", got, want)
	}
}
