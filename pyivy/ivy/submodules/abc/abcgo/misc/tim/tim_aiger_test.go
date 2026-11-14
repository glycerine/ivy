package tim

import (
	"bytes"
	"testing"
)

func TestAigerTimingRoundTripPreservesManagerState(t *testing.T) {
	p := Tim_ManStart(4, 3)
	Tim_ManCreateBox(p, 0, 2, 2, 1, 2, 1)
	Tim_ManInitPiArrival(p, 0, 7.5)
	Tim_ManInitPoRequired(p, 2, 19.25)
	p.nTravIds = 9
	p.fUseTravId = false
	p.vDelayTables = [][]float32{{0, 2, 1.5, 0.5, 0.25, 1}}
	p.vBoxes[0].iCopy = 6
	encoded, err := p.MarshalAigerTiming(true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalAigerTiming(encoded)
	if err != nil {
		t.Fatal(err)
	}
	encodedAgain, err := got.MarshalAigerTiming(true)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, encodedAgain) {
		t.Fatal("timing manager serialization is not deterministic")
	}
	if got.nCis != p.nCis || got.nCos != p.nCos || got.nTravIds != p.nTravIds || got.fUseTravId != p.fUseTravId {
		t.Fatalf("manager header changed: got (%d,%d,%d,%v), want (%d,%d,%d,%v)", got.nCis, got.nCos, got.nTravIds, got.fUseTravId, p.nCis, p.nCos, p.nTravIds, p.fUseTravId)
	}
	if len(got.vBoxes) != 1 || got.vBoxes[0].iCopy != 6 || got.vBoxes[0].nInputs != 2 || got.vBoxes[0].nOutputs != 1 || !got.vBoxes[0].fBlack {
		t.Fatalf("box metadata changed: %+v", got.vBoxes)
	}
	if got.pCis[0].timeArr != 7.5 || got.pCos[2].timeReq != 19.25 || got.vDelayTables[0][3] != 0.5 {
		t.Fatal("terminal timing or delay data changed")
	}
}

func TestAigerTimingRejectsMalformedPayload(t *testing.T) {
	if _, err := UnmarshalAigerTiming([]byte("not TIM")); err == nil {
		t.Fatal("unknown payload accepted")
	}
	p := Tim_ManStart(0, 0)
	encoded, err := p.MarshalAigerTiming(true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := UnmarshalAigerTiming(encoded[:len(encoded)-1]); err == nil {
		t.Fatal("truncated payload accepted")
	}
}
