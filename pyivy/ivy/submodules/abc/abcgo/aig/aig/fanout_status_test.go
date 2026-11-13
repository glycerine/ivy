package aig

import "testing"

func TestFanoutStatusTracksLifecycle(t *testing.T) {
	g := NewGraph()
	if Aig_ManHasFanout(nil) || Aig_ManHasFanout(g) { t.Fatal("unexpected active fanout") }
	Aig_ManFanoutStart(g)
	if !Aig_ManHasFanout(g) { t.Fatal("missing active fanout") }
	Aig_ManFanoutStop(g)
	if Aig_ManHasFanout(g) { t.Fatal("stopped fanout still active") }
}
