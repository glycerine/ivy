package vec

import "testing"

func TestSetPackingPagesAndHandles(t *testing.T) {
	p := Vec_SetAlloc(9)
	defer Vec_SetFree(p)
	var handles []int32
	for i := int32(0); i < 400; i++ {
		h := Vec_SetAppend(p, []int32{i, -i, i + 7}, 3)
		handles = append(handles, h)
	}
	if p.iPage == 0 || Vec_SetEntryNum(p) != 400 {
		t.Fatal(p.iPage, Vec_SetEntryNum(p))
	}
	for i, h := range handles {
		for j, want := range []int32{int32(i), -int32(i), int32(i) + 7} {
			if got := Vec_SetReadInt(p, h, int32(j)); got != want {
				t.Fatal(i, j, got, want)
			}
		}
	}
	Vec_SetWriteInt(p, handles[0], 1, -123)
	if Vec_SetReadInt(p, handles[0], 1) != -123 {
		t.Fatal("write")
	}
	entry := Vec_SetEntry(p, handles[0])
	for i := 0; i < 1000; i++ {
		Vec_SetFetchH(p, 64)
	}
	if entry[0] != Vec_SetEntry(p, handles[0])[0] {
		t.Fatal("unstable storage")
	}
	if Vec_ReportMemory(p) <= float64(Vec_SetMemoryAll(p)) {
		t.Fatal("memory accounting")
	}
}
func TestSetRestartShrinkAndShadow(t *testing.T) {
	p := Vec_SetAlloc(9)
	h := Vec_SetAppend(p, []int32{1, 2}, 2)
	limit := Vec_SetHandCurrent(p)
	Vec_SetAppend(p, []int32{3, 4}, 2)
	count := Vec_SetEntryNum(p)
	Vec_SetShrink(p, limit)
	if Vec_SetHandCurrent(p) != limit || Vec_SetEntryNum(p) != count {
		t.Fatal("shrink changed entry count")
	}
	Vec_SetShrinkS(p, 2)
	if Vec_SetAppendS(p, 2) != h {
		t.Fatal("shadow placement")
	}
	if Vec_SetHandCurrentS(p) != limit {
		t.Fatal("shadow limit")
	}
	Vec_SetShrinkLimits(p)
	if Vec_SetHandCurrent(p) != limit {
		t.Fatal("commit limits")
	}
	pages := p.pPages[0]
	Vec_SetRestart(p)
	if Vec_SetEntryNum(p) != 0 || Vec_SetHandCurrent(p) != 2 || Vec_SetLimitS(p.pPages[0]) != -1 {
		t.Fatal("restart")
	}
	if &pages[0] != &p.pPages[0][0] {
		t.Fatal("restart reallocates")
	}
}
func TestSetShadowAcrossPages(t *testing.T) {
	p := Vec_SetAlloc(9)
	var live []int32
	for i := 0; i < 20; i++ {
		live = append(live, Vec_SetFetchH(p, 400))
	}
	Vec_SetShrinkS(p, 2)
	for _, h := range live {
		if got := Vec_SetAppendS(p, 100); got != h {
			t.Fatal(got, h)
		}
	}
	if Vec_SetHandCurrent(p) != Vec_SetHandCurrentS(p) {
		t.Fatal("shadow traversal")
	}
	Vec_SetShrinkLimits(p)
}
func TestSetStringAndFetch(t *testing.T) {
	p := Vec_SetAlloc(9)
	h := Vec_SetHandCurrent(p)
	s := Vec_SetStrsav(p, "hello world")
	if s != "hello world" {
		t.Fatal(s)
	}
	words := Vec_SetEntry(p, h)
	for i, b := range append([]byte(s), 0) {
		if byte(words[i/8]>>uint(8*(i%8))) != b {
			t.Fatal("string packing", i)
		}
	}
	buf := Vec_SetFetch(p, 17)
	if len(buf) != 3 {
		t.Fatal(len(buf))
	}
	buf[0] = 42
	Vec_SetRestart(p)
	if s != "hello world" {
		t.Fatal("string lifetime")
	}
}
func TestSetInvalidSizes(t *testing.T) {
	for _, f := range []func(){func() { Vec_SetAlloc(8) }, func() { Vec_SetAppend(Vec_SetAlloc(9), nil, 1024) }, func() { Vec_SetAppend(Vec_SetAlloc(9), []int32{1}, 2) }, func() { Vec_SetAppendS(Vec_SetAlloc(9), 2) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid operation accepted")
				}
			}()
			f()
		}()
	}
}
