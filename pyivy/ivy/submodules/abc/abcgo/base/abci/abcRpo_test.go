package abci

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/bool/rpo"
)

func TestRpoTruthTableTextLoadAndWrite(t *testing.T) {
	input := filepath.Join(t.TempDir(), "truth.txt")
	if err := os.WriteFile(input, []byte("A\n5\n"), 0600); err != nil {
		t.Fatal(err)
	}
	store := Abc_RpoTtStoreLoad(input, -1)
	if store == nil {
		t.Fatal("text table load failed")
	}
	defer Abc_RpoTtStoreFree(store, -1)
	if store.nVars != 2 || store.nFuncs != 2 {
		t.Fatalf("loaded vars/functions = %d/%d, want 2/2", store.nVars, store.nFuncs)
	}
	if got := []int32{Abc_TruthGetHex(&store.pFuncs[0][0], 0), Abc_TruthGetHex(&store.pFuncs[1][0], 0)}; !reflect.DeepEqual(got, []int32{10, 5}) {
		t.Fatalf("loaded truth nibbles = %v, want [10 5]", got)
	}
	output := filepath.Join(t.TempDir(), "roundtrip.txt")
	Abc_RpoTtStoreWrite(output, store, 0)
	contents, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "A\n5\n" {
		t.Fatalf("roundtrip text = %q, want %q", contents, "A\n5\n")
	}
}

func TestRpoBinaryLoadAndUint32Layout(t *testing.T) {
	input := filepath.Join(t.TempDir(), "truth.bin")
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], 0x0123456789abcdef)
	if err := os.WriteFile(input, data[:], 0600); err != nil {
		t.Fatal(err)
	}
	store := Abc_RpoTtStoreLoad(input, 6)
	if store == nil {
		t.Fatal("binary table load failed")
	}
	defer Abc_RpoTtStoreFree(store, 6)
	if store.nFuncs != 1 || store.nWords != 1 {
		t.Fatalf("binary store funcs/words = %d/%d, want 1/1", store.nFuncs, store.nWords)
	}
	got := rpoTruth32(store.pFuncs[0], 6)
	want := []uint32{0x89abcdef, 0x01234567}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RPO words = %#v, want %#v", got, want)
	}
}

func TestRpoTruth32UsesSingleWordForAtMostFiveVars(t *testing.T) {
	got := rpoTruth32([]word{0xfedcba9876543210}, 5)
	if !reflect.DeepEqual(got, []uint32{0x76543210}) {
		t.Fatalf("5-variable RPO words = %#v", got)
	}
}

func TestRpoBorrowedStoreAndFactorization(t *testing.T) {
	buffer := []word{0xaaaaaaaaaaaaaaaa}
	store := Abc_RpoTruthStoreAlloc2(6, 1, buffer)
	if store == nil {
		t.Fatal("borrowed store allocation failed")
	}
	if store.pFuncs[0][0] != buffer[0] {
		t.Fatal("borrowed store did not view caller buffer")
	}
	factor := rpo.Rpo_Factorize(rpoTruth32(store.pFuncs[0], store.nVars), store.nVars, 1000, 0)
	if factor == nil {
		t.Fatal("RPO factorizer did not recognize a variable truth table")
	}
	rpo.Lit_Free(factor)
	Abc_RpoTtStoreFree(store, -1)
	if buffer[0] != 0xaaaaaaaaaaaaaaaa {
		t.Fatal("freeing a borrowed store changed caller-owned data")
	}
}
