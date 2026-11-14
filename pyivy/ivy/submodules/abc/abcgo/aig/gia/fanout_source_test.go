package gia

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDynamicFanoutSourceRemovalOrders(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("fixture uses GNU linker section collection")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	src := filepath.Clean(filepath.Join(dir, "../../../src"))
	binary := filepath.Join(t.TempDir(), "fanout")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	args := []string{"-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src, filepath.Join(dir, "test_vectors/dynamic_fanout.c"), "-Wl,--gc-sections", "-o", binary}
	if out, err := exec.CommandContext(ctx, cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile source fanout fixture: %v: %s", err, out)
	}
	want, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run source fanout fixture: %v: %s", err, want)
	}
	var got strings.Builder
	orders := [][3]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}
	for perm, order := range orders {
		p := Gia_ManStart(16)
		a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		roots := []int32{Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, c)}
		Gia_ManFanoutStart(p)
		snapshot := func(step int) {
			fmt.Fprintf(&got, "perm %d step %d", perm, step)
			for _, word := range p.pFanData[:5*p.nObjs] {
				fmt.Fprintf(&got, " %d", word)
			}
			got.WriteByte('\n')
		}
		snapshot(0)
		for i, index := range order {
			Gia_ObjRemoveFanout(p, Gia_ManObj(p, a>>1), Gia_ManObj(p, roots[index]>>1))
			snapshot(i + 1)
		}
		for i := range order {
			Gia_ObjAddFanout(p, Gia_ManObj(p, a>>1), Gia_ManObj(p, roots[order[2-i]]>>1))
			snapshot(i + 4)
		}
	}
	if got.String() != string(want) {
		t.Fatalf("Go fanout snapshots:\n%sC snapshots:\n%s", got.String(), want)
	}
}
