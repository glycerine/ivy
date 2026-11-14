package abci

import "testing"

func TestSaucyBumpActivityRescalesDatabase(t *testing.T) {
	entries := Vec_PtrAlloc(2)
	first := &sim_result{activity: 2e20}
	second := &sim_result{activity: 3e20}
	Vec_PtrPush(entries, first)
	Vec_PtrPush(entries, second)
	s := &saucy{activityInc: 1e20, satCounterExamples: entries}
	bumpActivity(s, first)
	if first.activity != 3 || second.activity != 3 || s.activityInc != 1 {
		t.Fatalf("rescaled activities/inc = %g, %g, %g; want 3, 3, 1", first.activity, second.activity, s.activityInc)
	}
}

func TestSaucyReduceDBDropsLowActivityEntries(t *testing.T) {
	entries := Vec_PtrAlloc(40)
	for i := 0; i < 36; i++ {
		activity := 100.0
		if i < 18 {
			activity = 0
		}
		Vec_PtrPush(entries, &sim_result{activity: activity})
	}
	s := &saucy{activityInc: 1, satCounterExamples: entries}
	reduceDB(s)
	if got := Vec_PtrSize(entries); got > 35 {
		t.Fatalf("retained %d counterexamples, want at most 35", got)
	}
	for i := int32(0); i < Vec_PtrSize(entries); i++ {
		if Vec_PtrEntry(entries, i).(*sim_result).activity < 100 {
			t.Fatalf("low activity entry retained at index %d", i)
		}
	}
}
