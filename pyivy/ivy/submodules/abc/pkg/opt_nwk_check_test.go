package abc

// Go-only tests for src/opt/nwk/nwkCheck.c.

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"
)

func linkNwkObjects(from, to *NwkConsistencyObject) {
	to.Fanins = append(to.Fanins, from)
	from.Fanouts = append(from.Fanouts, to)
}

func TestCheckNwkConsistencyValidSparseNetwork(t *testing.T) {
	ci0 := &NwkConsistencyObject{ID: 0, Kind: NwkConsistencyCI}
	ci1 := &NwkConsistencyObject{ID: 1, Kind: NwkConsistencyCI}
	node := &NwkConsistencyObject{ID: 2, Kind: NwkConsistencyNode}
	co := &NwkConsistencyObject{ID: 3, Kind: NwkConsistencyCO}
	linkNwkObjects(ci0, node)
	linkNwkObjects(ci1, node)
	linkNwkObjects(node, co)
	issues, err := CheckNwkConsistency(&NwkConsistencyGraph{Objects: []*NwkConsistencyObject{ci0, nil, ci1, node, co}}, NwkConsistencyLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("unexpected issues: %+v", issues)
	}
}

func TestCheckNwkConsistencyExactDiagnosticsAndOrder(t *testing.T) {
	a := &NwkConsistencyObject{ID: 10, Kind: NwkConsistencyCI}
	b := &NwkConsistencyObject{ID: 11, Kind: NwkConsistencyCI}
	n := &NwkConsistencyObject{ID: 12, Kind: NwkConsistencyNode, Fanins: []*NwkConsistencyObject{a, a, a, b}}
	x := &NwkConsistencyObject{ID: 13, Kind: NwkConsistencyCO}
	a.Fanouts = []*NwkConsistencyObject{n, x}
	x.Fanins = []*NwkConsistencyObject{b}
	want := []NwkConsistencyIssue{
		{Kind: NwkDuplicateFanin, ObjectID: 12, RelatedID: 10},
		{Kind: NwkDuplicateFanin, ObjectID: 12, RelatedID: 10},
		{Kind: NwkDuplicateFanin, ObjectID: 12, RelatedID: 10},
		{Kind: NwkMissingReverseFanin, ObjectID: 10, RelatedID: 13},
		{Kind: NwkMissingReverseFanout, ObjectID: 12, RelatedID: 11},
		{Kind: NwkMissingReverseFanout, ObjectID: 13, RelatedID: 11},
	}
	got, err := CheckNwkConsistency(&NwkConsistencyGraph{Objects: []*NwkConsistencyObject{a, b, n, x}}, NwkConsistencyLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("issues\n got: %+v\nwant: %+v", got, want)
	}
}

func TestCheckNwkConsistencySourceSpecialCases(t *testing.T) {
	a := &NwkConsistencyObject{ID: 1, Kind: NwkConsistencyCI}
	co := &NwkConsistencyObject{ID: 2, Kind: NwkConsistencyCO, Fanins: []*NwkConsistencyObject{a, a}}
	// Duplicate fanins on a non-node are intentionally not diagnosed.
	for i := 0; i < 100; i++ {
		a.Fanouts = append(a.Fanouts, &NwkConsistencyObject{ID: 100 + i, Kind: NwkConsistencyCO, Fanins: []*NwkConsistencyObject{a}})
	}
	issues, err := CheckNwkConsistency(&NwkConsistencyGraph{Objects: []*NwkConsistencyObject{a, co}}, NwkConsistencyLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("high-fanout shortcut or non-node duplicate mismatch: %+v", issues)
	}
	a.Fanouts = a.Fanouts[:99]
	issues, err = CheckNwkConsistency(&NwkConsistencyGraph{Objects: []*NwkConsistencyObject{a, co}}, NwkConsistencyLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 2 || issues[0].Kind != NwkMissingReverseFanout || issues[1].Kind != NwkMissingReverseFanout {
		t.Fatalf("threshold must expose both source relationships: %+v", issues)
	}
}

func TestCheckNwkConsistencyErrorsLimitsCancellationAndOwnership(t *testing.T) {
	_, err := CheckNwkConsistency(nil, NwkConsistencyLimits{})
	var checkErr *NwkConsistencyError
	if !errors.As(err, &checkErr) || checkErr.Code != NwkCheckInvalidInput {
		t.Fatalf("nil graph error = %v", err)
	}
	bad := &NwkConsistencyObject{ID: 4, Kind: NwkConsistencyNode, Fanins: []*NwkConsistencyObject{nil}}
	_, err = CheckNwkConsistency(&NwkConsistencyGraph{Objects: []*NwkConsistencyObject{bad}}, NwkConsistencyLimits{})
	if !errors.As(err, &checkErr) || checkErr.Code != NwkCheckInvalidInput {
		t.Fatalf("nil relationship error = %v", err)
	}
	bad.Fanins = nil
	bad.Kind = 99
	_, err = CheckNwkConsistency(&NwkConsistencyGraph{Objects: []*NwkConsistencyObject{bad}}, NwkConsistencyLimits{})
	if !errors.As(err, &checkErr) || checkErr.Code != NwkCheckInvalidInput {
		t.Fatalf("kind error = %v", err)
	}
	a := &NwkConsistencyObject{ID: 1, Kind: NwkConsistencyCI}
	b := &NwkConsistencyObject{ID: 2, Kind: NwkConsistencyNode, Fanins: []*NwkConsistencyObject{a}}
	graph := &NwkConsistencyGraph{Objects: []*NwkConsistencyObject{a, b}}
	_, err = CheckNwkConsistency(graph, NwkConsistencyLimits{MaxObjects: 1})
	if !errors.As(err, &checkErr) || checkErr.Code != NwkCheckLimitExceeded {
		t.Fatalf("object limit error = %v", err)
	}
	b.Fanouts = []*NwkConsistencyObject{a}
	_, err = CheckNwkConsistency(graph, NwkConsistencyLimits{MaxRelationships: 1})
	if !errors.As(err, &checkErr) || checkErr.Code != NwkCheckLimitExceeded {
		t.Fatalf("relationship limit error = %v", err)
	}
	b.Fanouts = nil
	negativeLimits := []struct {
		name   string
		limits NwkConsistencyLimits
	}{
		{"objects", NwkConsistencyLimits{MaxObjects: -1}},
		{"relationships", NwkConsistencyLimits{MaxRelationships: -1}},
		{"issues", NwkConsistencyLimits{MaxIssues: -1}},
	}
	for _, test := range negativeLimits {
		_, err = CheckNwkConsistency(graph, test.limits)
		if !errors.As(err, &checkErr) || checkErr.Code != NwkCheckInvalidInput {
			t.Fatalf("negative %s limit error = %v", test.name, err)
		}
	}
	calls := 0
	issues, err := CheckNwkConsistency(graph, NwkConsistencyLimits{Stop: func() bool { calls++; return calls == 2 }})
	if issues != nil || !errors.As(err, &checkErr) || checkErr.Code != NwkCheckCancelled {
		t.Fatalf("atomic cancellation = (%+v, %v)", issues, err)
	}
	issues, err = CheckNwkConsistency(graph, NwkConsistencyLimits{})
	if err != nil || len(issues) != 1 {
		t.Fatalf("missing issue = (%+v, %v)", issues, err)
	}
	b.ID = 22
	a.ID = 11
	if issues[0].ObjectID != 2 || issues[0].RelatedID != 1 {
		t.Fatalf("report retained mutable objects: %+v", issues)
	}
}

func TestCheckNwkConsistencyAtomicMidTraversalFailures(t *testing.T) {
	a := &NwkConsistencyObject{ID: 1, Kind: NwkConsistencyCI}
	node := &NwkConsistencyObject{ID: 2, Kind: NwkConsistencyNode, Fanins: []*NwkConsistencyObject{a, a, a}}
	graph := &NwkConsistencyGraph{Objects: []*NwkConsistencyObject{node}}

	issues, err := CheckNwkConsistency(graph, NwkConsistencyLimits{MaxIssues: 1})
	var checkErr *NwkConsistencyError
	if issues != nil || !errors.As(err, &checkErr) || checkErr.Code != NwkCheckLimitExceeded {
		t.Fatalf("atomic issue limit = (%+v, %v)", issues, err)
	}

	// The first duplicate pair is recorded at poll six. Cancelling at poll
	// seven exercises the next inner-loop iteration and must discard it.
	calls := 0
	issues, err = CheckNwkConsistency(graph, NwkConsistencyLimits{Stop: func() bool {
		calls++
		return calls == 7
	}})
	if issues != nil || !errors.As(err, &checkErr) || checkErr.Code != NwkCheckCancelled {
		t.Fatalf("atomic inner-loop cancellation = (%+v, %v), polls=%d", issues, err, calls)
	}
}

func modelNwkConsistency(objects []*NwkConsistencyObject) []NwkConsistencyIssue {
	result := make([]NwkConsistencyIssue, 0)
	for _, object := range objects {
		if object == nil || object.Kind != NwkConsistencyNode {
			continue
		}
		for i, fanin := range object.Fanins {
			for _, later := range object.Fanins[i+1:] {
				if fanin == later {
					result = append(result, NwkConsistencyIssue{Kind: NwkDuplicateFanin, ObjectID: object.ID, RelatedID: fanin.ID})
				}
			}
		}
	}
	for _, object := range objects {
		if object == nil {
			continue
		}
		for _, fanin := range object.Fanins {
			found := false
			for _, candidate := range fanin.Fanouts {
				found = found || candidate == object
			}
			if len(fanin.Fanouts) < 100 && !found {
				result = append(result, NwkConsistencyIssue{Kind: NwkMissingReverseFanout, ObjectID: object.ID, RelatedID: fanin.ID})
			}
		}
		for _, fanout := range object.Fanouts {
			found := false
			for _, candidate := range fanout.Fanins {
				found = found || candidate == object
			}
			if !found {
				result = append(result, NwkConsistencyIssue{Kind: NwkMissingReverseFanin, ObjectID: object.ID, RelatedID: fanout.ID})
			}
		}
	}
	return result
}

func runNwkConsistencyModel(t *testing.T, seed int64, executions int) {
	t.Helper()
	rng := rand.New(rand.NewSource(seed))
	for execution := 0; execution < executions; execution++ {
		count := 1 + rng.Intn(10)
		objects := make([]*NwkConsistencyObject, count)
		for i := range objects {
			objects[i] = &NwkConsistencyObject{ID: execution*16 + i, Kind: NwkConsistencyObjectKind(rng.Intn(int(NwkConsistencyVoid) + 1))}
		}
		for _, object := range objects {
			for i := 0; i < rng.Intn(8); i++ {
				object.Fanins = append(object.Fanins, objects[rng.Intn(count)])
			}
			for i := 0; i < rng.Intn(8); i++ {
				object.Fanouts = append(object.Fanouts, objects[rng.Intn(count)])
			}
		}
		got, err := CheckNwkConsistency(&NwkConsistencyGraph{Objects: objects}, NwkConsistencyLimits{})
		if err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
		want := modelNwkConsistency(objects)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("execution %d mismatch\n got: %+v\nwant: %+v", execution, got, want)
		}
	}
}

func TestCheckNwkConsistencyModelCampaign(t *testing.T) {
	runNwkConsistencyModel(t, 0x4e574b, 5000)
}

func FuzzCheckNwkConsistency(f *testing.F) {
	f.Add(int64(1), uint16(32))
	f.Add(int64(-1), uint16(128))
	f.Fuzz(func(t *testing.T, seed int64, executions uint16) {
		count := int(executions%200) + 1
		runNwkConsistencyModel(t, seed, count)
	})
}
