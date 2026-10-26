package abc

import "testing"

func TestPdrInvariantRejectsInvalidCertificates(t *testing.T) {
	cube, err := NewPdrCube([]PdrLit{0}, nil)
	if err != nil {
		t.Fatal(err)
	}
	inv, err := NewPdrInvariant([]*PdrCube{cube})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name      string
		bad, next uint8
		init      LatchInit
		phase     PdrInvariantPhase
		status    PdrInvariantStatus
	}{
		{"valid", 204, 0, LatchInitZero, PdrInvariantComplete, PdrInvariantValid},
		{"initiation", 204, 0, LatchInitOne, PdrInvariantInitiation, PdrInvariantInvalid},
		{"induction", 204, 255, LatchInitZero, PdrInvariantInduction, PdrInvariantInvalid},
		{"safety", 51, 0, LatchInitZero, PdrInvariantSafety, PdrInvariantInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := bmcTruthModel(t, bmcTruthCase{bad: tc.bad, next0: tc.next, constraint: 255, initial: [2]LatchInit{tc.init, LatchInitZero}})
			r, err := inv.Verify(m, 0, PdrVerifyOptions{})
			if err != nil || r.Status != tc.status || r.Phase != tc.phase {
				t.Fatal(r, err)
			}
		})
	}
	safe := bmcTruthModel(t, bmcTruthCase{constraint: 255})
	for _, limit := range []int{1, 2} {
		r, err := inv.Verify(safe, 0, PdrVerifyOptions{MaxQueries: limit})
		if err != nil || r.Status != PdrInvariantUnknown || !r.QueryLimit || r.Queries != uint64(limit) {
			t.Fatal(r, err)
		}
	}
	r, err := inv.Verify(safe, 0, PdrVerifyOptions{SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || r.Status != PdrInvariantUnknown || r.Reason != SatCancelled {
		t.Fatal(r, err)
	}
	for _, property := range []int{-1, 1} {
		if _, err := inv.Verify(safe, property, PdrVerifyOptions{}); err == nil {
			t.Fatal("invalid property accepted")
		}
	}
	out, err := NewPdrCube([]PdrLit{4}, nil)
	if err != nil {
		t.Fatal(err)
	}
	wrong, err := NewPdrInvariant([]*PdrCube{out})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wrong.Verify(safe, 0, PdrVerifyOptions{}); err == nil {
		t.Fatal("out of range state accepted")
	}
	empty, err := NewPdrCube(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	impossible, err := NewPdrInvariant([]*PdrCube{empty})
	if err != nil {
		t.Fatal(err)
	}
	vacuous := bmcTruthModel(t, bmcTruthCase{bad: 255, constraint: 0})
	r, err = impossible.Verify(vacuous, 0, PdrVerifyOptions{})
	if err != nil || r.Status != PdrInvariantValid {
		t.Fatal("vacuous model", r, err)
	}
	r, err = impossible.Verify(safe, 0, PdrVerifyOptions{})
	if err != nil || r.Status != PdrInvariantInvalid || r.Phase != PdrInvariantInitiation {
		t.Fatal("false certificate", r, err)
	}
}
func TestPdrTransitionMasks(t *testing.T) {
	// Include constants, complements and shared next-state functions, with
	// fixed and unknown initial states. Sparse bad masks produce deeper paths
	// and proofs that broad random output masks seldom exercise.
	for n0 := 0; n0 < 256; n0++ {
		for _, n1 := range []uint8{0, 255, uint8(n0), ^uint8(n0), 170, 204, 240} {
			for _, bad := range []uint8{0, 3, 192} {
				pdrVerifyTruthCase(t, bmcTruthCase{bad: bad, next0: uint8(n0), next1: n1, constraint: 255}, false)
			}
		}
	}
}
func TestPdrCancellationAndResourceLimits(t *testing.T) {
	m := bmcTruthModel(t, bmcTruthCase{constraint: 255, next0: 170, next1: 204})
	for _, tc := range []struct {
		options PdrOptions
		reason  PdrStopReason
	}{
		{PdrOptions{MaxVariables: 1}, PdrResourceLimit},
		{PdrOptions{MaxFrameVariables: 1}, PdrResourceLimit},
		{PdrOptions{MaxQueries: 1}, PdrQueryLimit},
		{PdrOptions{SAT: SatLimits{Stop: func() bool { return true }}}, PdrCancelled},
	} {
		r, err := m.Prove(tc.options)
		if err != nil || r.Status != PdrUnknown || r.Reason != tc.reason || r.Invariant != nil || r.Counterexample != nil {
			t.Fatal(r, err)
		}
	}
	r, err := m.Prove(PdrOptions{})
	if err != nil || r.Status != PdrProved {
		t.Fatal("fresh run after interruption", r, err)
	}
}
