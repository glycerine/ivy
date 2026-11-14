package mpm

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestDsdLookupUsesTruthComplementBit(t *testing.T) {
	truths := vec.Vec_WrdAlloc(2)
	vec.Vec_WrdPush(truths, 0xaaaaaaaaaaaaaaaa)
	vec.Vec_WrdPush(truths, 0x8888888888888888)
	p := &Mpm_Man_t{pHash: vec.Hsh_WrdManHashArrayStart(truths, 1), vConfgRes: vec.Vec_IntAlloc(2)}
	vec.Vec_IntPush(p.vConfgRes, 123)
	vec.Vec_IntPush(p.vConfgRes, 456)
	for _, tc := range []struct {
		truth uint64
		want  int32
	}{
		{0xaaaaaaaaaaaaaaaa, 123},
		{0x5555555555555555, 123 ^ (1 << 16)},
		{0x8888888888888888, 456},
		{0xcccccccccccccccc, -1},
	} {
		if got := Mpm_CutCheckDsd6(p, tc.truth); got != tc.want {
			t.Errorf("truth=%016x got=%d want=%d", tc.truth, got, tc.want)
		}
	}
}

func TestDsdPrecomputeReadsSourceDataset(t *testing.T) {
	dir := t.TempDir()
	data := make([]byte, 24)
	binary.NativeEndian.PutUint64(data[8:], 0xaaaaaaaaaaaaaaaa)
	binary.NativeEndian.PutUint32(data[20:], 1<<17)
	if err := os.WriteFile(filepath.Join(dir, "dsdfuncs6.dat"), data, 0600); err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Error(err)
		}
	})
	p := &Mpm_Man_t{}
	Mpm_ManPrecomputePerms(p)
	if len(p.pDsd6) != 595 || vec.Vec_WrdSize(p.vPerm6) != 595*720 || vec.Vec_IntSize(p.vMap2Perm) != 1<<18 {
		t.Fatal("DSD schedules have wrong size")
	}
	if got := Mpm_CutCheckDsd6(p, 0xaaaaaaaaaaaaaaaa); got != 1<<17 {
		t.Fatalf("variable class = %d, want %d", got, 1<<17)
	}
	if got := Mpm_CutCheckDsd6(p, 0x5555555555555555); got != (1<<17)|(1<<16) {
		t.Fatalf("complement class = %d", got)
	}
}
