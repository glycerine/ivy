package if_

import (
	"reflect"
	"testing"
)

func TestLpkParameterInitializationMatchesSource(t *testing.T) {
	p := If_Par_t_{nLutSize: 9, nCutsMax: 3, fAreaOnly: 1, pTimesArr: []float32{99}, fBidec: 1}
	If_ManSetLpkPars(&p, 4)
	want := If_Par_t_{nLutSize: 4, nCutsMax: 16, DelayTarget: -1, Epsilon: 0.005, fArea: 1, fTruth: 1}
	if !reflect.DeepEqual(p, want) {
		t.Fatalf("got %+v, want %+v", p, want)
	}
}

func TestLpkIndexedArrivalsSurviveRestart(t *testing.T) {
	var p If_Par_t
	If_ManSetLpkPars(&p, 4)
	m := If_ManStart(&p)
	defer If_ManStop(m)
	p.SetArrivalTimes(make([]float32, 32))
	p.SetArrivalTime(0, 2.25)
	p.SetArrivalTime(1, 4.5)
	p.SetArrivalTime(31, 9)
	p.SetAreaOnly(true)
	If_ManRestart(m)
	if got := m.Parameters(); got != &p {
		t.Fatal("restart changed parameter identity")
	}
	if p.pTimesArr[0] != 2.25 || p.pTimesArr[1] != 4.5 || p.pTimesArr[31] != 9 || p.fAreaOnly != 1 {
		t.Fatalf("restart changed mapping configuration: %+v", p)
	}
	If_ManCreateCi(m)
	If_ManCreateCi(m)
	p.SetArrivalTime(1, 7)
	if p.pTimesArr[0] != 2.25 || p.pTimesArr[1] != 7 || p.pTimesArr[31] != 9 {
		t.Fatalf("indexed update changed other slots: %v", p.pTimesArr)
	}
	p.SetAreaOnly(false)
	if p.fAreaOnly != 0 {
		t.Fatal("failed to clear area-only flag")
	}
}
func TestLpkIndexedArrivalRejectsOutOfRange(t *testing.T) {
	for _, i := range []int{-1, 32} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("index %d should panic", i)
				}
			}()
			var p If_Par_t
			p.SetArrivalTimes(make([]float32, 32))
			p.SetArrivalTime(i, 1)
		}()
	}
}

func TestLpkIndexedArrivalsReachMapping(t *testing.T) {
	var p If_Par_t
	If_ManSetLpkPars(&p, 4)
	m := If_ManStart(&p)
	defer If_ManStop(m)
	If_ManSetupSetAll(m, 1000)
	p.SetArrivalTimes(make([]float32, 32))
	for pass := 0; pass < 2; pass++ {
		If_ManRestart(m)
		a, b := If_ManCreateCi(m), If_ManCreateCi(m)
		p.SetArrivalTime(0, float32(pass+2))
		p.SetArrivalTime(1, float32(pass+5))
		If_ManSetupCiCutSets(m)
		root := If_ManCreateAnd(m, a, b)
		If_ManCreateCo(m, root)
		p.SetAreaOnly(true)
		If_ManPerformMappingComb(m)
		if If_ObjArrTime(a) != float32(pass+2) || If_ObjArrTime(b) != float32(pass+5) {
			t.Fatalf("pass %d arrivals %v/%v", pass, If_ObjArrTime(a), If_ObjArrTime(b))
		}
	}
}
