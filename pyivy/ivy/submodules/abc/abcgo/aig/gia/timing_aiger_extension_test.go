package gia

import (
	"testing"

	"github.com/glycerine/abcgo/misc/tim"
)

func TestAigerRoundTripPreservesTimManagerAndExtraGraph(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b))
	extra := Gia_ManStart(4)
	x := Gia_ManAppendCi(extra)
	Gia_ManAppendCo(extra, x^1)
	manager := tim.Tim_ManStart(2, 1)
	tim.Tim_ManCreateBox(manager, 0, 1, 1, 1, 1, 0)
	tim.Tim_ManInitPiArrival(manager, 0, 2.75)
	Gia_ManSetBoxData(p, extra, manager)
	if err := p.SetAigerExtension('t', []byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	defer Gia_ManStopTimingData(p)
	defer Gia_ManStop(p)

	data, err := MarshalAiger(p, AigerWriteOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseAiger(data, AigerReadOptions{SkipStrash: true, Check: true})
	if err != nil {
		t.Fatal(err)
	}
	defer Gia_ManStopTimingData(got)
	defer Gia_ManStop(got)
	gotManager, gotExtra := Gia_ManBoxData(got)
	tm, ok := gotManager.(*tim.Tim_Man_t)
	if !ok || tm == nil || gotExtra == nil {
		t.Fatalf("round trip lost hierarchy manager: %T, %p", gotManager, gotExtra)
	}
	if tim.Tim_ManBoxNum(tm) != 1 || tim.Tim_ManGetCiArrival(tm, 0) != 2.75 || Gia_ManCompare(extra, gotExtra) == 0 {
		t.Fatal("round trip changed TIM state or auxiliary Boolean graph")
	}
	if got.AigerExtension('t')[0] != 1 || got.AigerExtension('t')[2] != 3 {
		t.Fatal("round trip dropped the separate AIGER timing extension")
	}
}

func TestGiaManStopTimingDataReleasesOwnedManagers(t *testing.T) {
	p := Gia_ManStart(1)
	extra := Gia_ManStart(1)
	timing := tim.Tim_ManStart(1, 1)
	Gia_ManSetBoxData(p, extra, timing)
	Gia_ManStopTimingData(p)
	if tim.Tim_ManCiNum(timing) != 0 || Gia_ManObjNum(extra) != 0 {
		t.Fatal("timing attachments were not released")
	}
	if manager, graph := Gia_ManBoxData(p); manager != nil || graph != nil {
		t.Fatal("stopped timing attachments remain associated")
	}
	Gia_ManStop(p)
}
