package gia

import (
	"bytes"
	abcmain "github.com/glycerine/abcgo/base/main"
	"strings"
	"testing"
)

type deepDispatcher struct {
	calls   int
	fail    bool
	batchOK bool
	command string
}

func (d *deepDispatcher) Execute(f *abcmain.Frame, command string) int32 {
	d.calls++
	d.command = command
	d.batchOK = abcmain.Abc_FrameIsBatchMode() != 0
	if d.fail {
		return 1
	}
	current := f.Resource("pGia").(*Gia_Man_t)
	next := Gia_ManStart(8)
	a := Gia_ManAppendCi(next)
	Gia_ManAppendCi(next)
	Gia_ManAppendCo(next, a)
	if Gia_ManCiNum(current) != 2 {
		panic("unexpected input graph")
	}
	f.SetResource("pGia", next)
	return 0
}
func TestDeepSynthesisFrameAndFailure(t *testing.T) {
	old := abcmain.Abc_FrameReadGlobalFrame()
	f := abcmain.NewFrame()
	abcmain.Abc_FrameSetGlobalFrame(f)
	defer abcmain.Abc_FrameSetGlobalFrame(old)
	defer f.Close()
	var errors bytes.Buffer
	f.Err = &errors
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b))
	f.SetResource("pGia", p)
	d := &deepDispatcher{}
	f.Dispatcher = d
	out := Gia_ManDeepSynOne(5, 0, 0, 0, 0, 0)
	if out == nil || Gia_ManAndNum(out) != 0 || d.calls != 7 || !d.batchOK || abcmain.Abc_FrameIsBatchMode() != 0 || !strings.Contains(d.command, "&mfs") {
		t.Fatalf("result %v calls %d batch %v", out, d.calls, d.batchOK)
	}
	d.fail = true
	errors.Reset()
	if Gia_ManDeepSynOne(5, 0, 0, 0, 0, 0) != nil || abcmain.Abc_FrameIsBatchMode() != 0 || !strings.Contains(errors.String(), "did not work") {
		t.Fatal("failure was not propagated or batch mode not restored")
	}
}
