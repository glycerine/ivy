package abc

// Port of src/opt/nwk/nwkCheck.c.

import "fmt"

// NwkConsistencyObjectKind is the subset of Nwk_Type_t observed by the
// consistency checker. The checker distinguishes logic nodes from every other
// kind only when looking for duplicated fanins.
type NwkConsistencyObjectKind uint8

const (
	NwkConsistencyNone NwkConsistencyObjectKind = iota
	NwkConsistencyCI
	NwkConsistencyCO
	NwkConsistencyNode
	NwkConsistencyLatch
	NwkConsistencyVoid
)

// NwkConsistencyObject is an owned-by-the-caller view of one network object.
// Relationships use pointer identity, as the source does.
type NwkConsistencyObject struct {
	ID      int
	Kind    NwkConsistencyObjectKind
	Fanins  []*NwkConsistencyObject
	Fanouts []*NwkConsistencyObject
}

// NwkConsistencyGraph preserves the source manager's sparse object array.
// Nil entries are unused slots and are skipped.
type NwkConsistencyGraph struct {
	Objects []*NwkConsistencyObject
}

// NwkConsistencyIssueKind identifies one diagnostic emitted by Nwk_ManCheck.
type NwkConsistencyIssueKind uint8

const (
	NwkDuplicateFanin NwkConsistencyIssueKind = iota + 1
	NwkMissingReverseFanout
	NwkMissingReverseFanin
)

// NwkConsistencyIssue is an owned diagnostic. IDs are copied so a report does
// not retain or expose mutable graph objects.
type NwkConsistencyIssue struct {
	Kind      NwkConsistencyIssueKind
	ObjectID  int
	RelatedID int
}

// NwkConsistencyLimits bounds a check. Zero selects a conservative default.
// Stop is polled deterministically and must not mutate the graph or reenter the
// checker.
type NwkConsistencyLimits struct {
	MaxObjects       int
	MaxRelationships int
	MaxIssues        int
	Stop             func() bool
}

const (
	defaultNwkCheckMaxObjects       = 1 << 20
	defaultNwkCheckMaxRelationships = 1 << 24
	defaultNwkCheckMaxIssues        = 1 << 20
)

// NwkConsistencyErrorCode classifies malformed input and resource failures.
type NwkConsistencyErrorCode uint8

const (
	NwkCheckInvalidInput NwkConsistencyErrorCode = iota + 1
	NwkCheckLimitExceeded
	NwkCheckCancelled
)

// NwkConsistencyError is returned atomically: a failed check never returns a
// partial issue list.
type NwkConsistencyError struct {
	Code    NwkConsistencyErrorCode
	Object  int
	Message string
}

func (e *NwkConsistencyError) Error() string {
	if e.Object >= 0 {
		return fmt.Sprintf("nwk consistency check: object %d: %s", e.Object, e.Message)
	}
	return "nwk consistency check: " + e.Message
}

func nwkCheckLimit(value, fallback int) (int, bool) {
	if value == 0 {
		return fallback, true
	}
	return value, value > 0
}

func nwkContainsIdentity(objects []*NwkConsistencyObject, target *NwkConsistencyObject, poll func() error) (bool, error) {
	for _, object := range objects {
		if err := poll(); err != nil {
			return false, err
		}
		if object == target {
			return true, nil
		}
	}
	return false, nil
}

// CheckNwkConsistency preserves Nwk_ManCheck's diagnostic ordering and
// relationship rules. Duplicated fanins are reported only for logic nodes.
// Missing reverse fanouts are deliberately suppressed when the fanin already
// has at least 100 fanouts, matching the source's high-fanout shortcut.
//
// The C function prints diagnostics and always returns one. This Go operation
// instead returns owned structured diagnostics. Nil relationship entries,
// invalid kinds, limits and cancellation are errors rather than assertion,
// dereference or partial-output hazards.
func CheckNwkConsistency(graph *NwkConsistencyGraph, limits NwkConsistencyLimits) ([]NwkConsistencyIssue, error) {
	if graph == nil {
		return nil, &NwkConsistencyError{Code: NwkCheckInvalidInput, Object: -1, Message: "nil graph"}
	}
	maxObjects, ok := nwkCheckLimit(limits.MaxObjects, defaultNwkCheckMaxObjects)
	if !ok {
		return nil, &NwkConsistencyError{Code: NwkCheckInvalidInput, Object: -1, Message: "negative object limit"}
	}
	maxRelationships, ok := nwkCheckLimit(limits.MaxRelationships, defaultNwkCheckMaxRelationships)
	if !ok {
		return nil, &NwkConsistencyError{Code: NwkCheckInvalidInput, Object: -1, Message: "negative relationship limit"}
	}
	maxIssues, ok := nwkCheckLimit(limits.MaxIssues, defaultNwkCheckMaxIssues)
	if !ok {
		return nil, &NwkConsistencyError{Code: NwkCheckInvalidInput, Object: -1, Message: "negative issue limit"}
	}
	if len(graph.Objects) > maxObjects {
		return nil, &NwkConsistencyError{Code: NwkCheckLimitExceeded, Object: -1, Message: "object limit exceeded"}
	}
	poll := func() error {
		if limits.Stop != nil && limits.Stop() {
			return &NwkConsistencyError{Code: NwkCheckCancelled, Object: -1, Message: "cancelled"}
		}
		return nil
	}

	relationships := 0
	for _, object := range graph.Objects {
		if err := poll(); err != nil {
			return nil, err
		}
		if object == nil {
			continue
		}
		if object.Kind > NwkConsistencyVoid {
			return nil, &NwkConsistencyError{Code: NwkCheckInvalidInput, Object: object.ID, Message: "invalid object kind"}
		}
		if len(object.Fanins) > maxRelationships-relationships {
			return nil, &NwkConsistencyError{Code: NwkCheckLimitExceeded, Object: object.ID, Message: "relationship limit exceeded"}
		}
		relationships += len(object.Fanins)
		if len(object.Fanouts) > maxRelationships-relationships {
			return nil, &NwkConsistencyError{Code: NwkCheckLimitExceeded, Object: object.ID, Message: "relationship limit exceeded"}
		}
		relationships += len(object.Fanouts)
		for _, related := range object.Fanins {
			if err := poll(); err != nil {
				return nil, err
			}
			if related == nil {
				return nil, &NwkConsistencyError{Code: NwkCheckInvalidInput, Object: object.ID, Message: "nil fanin"}
			}
		}
		for _, related := range object.Fanouts {
			if err := poll(); err != nil {
				return nil, err
			}
			if related == nil {
				return nil, &NwkConsistencyError{Code: NwkCheckInvalidInput, Object: object.ID, Message: "nil fanout"}
			}
		}
	}

	issues := make([]NwkConsistencyIssue, 0)
	appendIssue := func(issue NwkConsistencyIssue) error {
		if len(issues) == maxIssues {
			return &NwkConsistencyError{Code: NwkCheckLimitExceeded, Object: issue.ObjectID, Message: "issue limit exceeded"}
		}
		issues = append(issues, issue)
		return nil
	}
	for _, object := range graph.Objects {
		if err := poll(); err != nil {
			return nil, err
		}
		if object == nil || object.Kind != NwkConsistencyNode {
			continue
		}
		for left := 0; left < len(object.Fanins); left++ {
			for right := left + 1; right < len(object.Fanins); right++ {
				if err := poll(); err != nil {
					return nil, err
				}
				if object.Fanins[left] == object.Fanins[right] {
					if err := appendIssue(NwkConsistencyIssue{Kind: NwkDuplicateFanin, ObjectID: object.ID, RelatedID: object.Fanins[left].ID}); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	for _, object := range graph.Objects {
		if err := poll(); err != nil {
			return nil, err
		}
		if object == nil {
			continue
		}
		for _, fanin := range object.Fanins {
			if err := poll(); err != nil {
				return nil, err
			}
			if len(fanin.Fanouts) < 100 {
				found, err := nwkContainsIdentity(fanin.Fanouts, object, poll)
				if err != nil {
					return nil, err
				}
				if !found {
					if err := appendIssue(NwkConsistencyIssue{Kind: NwkMissingReverseFanout, ObjectID: object.ID, RelatedID: fanin.ID}); err != nil {
						return nil, err
					}
				}
			}
		}
		for _, fanout := range object.Fanouts {
			if err := poll(); err != nil {
				return nil, err
			}
			found, err := nwkContainsIdentity(fanout.Fanins, object, poll)
			if err != nil {
				return nil, err
			}
			if !found {
				if err := appendIssue(NwkConsistencyIssue{Kind: NwkMissingReverseFanin, ObjectID: object.ID, RelatedID: fanout.ID}); err != nil {
					return nil, err
				}
			}
		}
	}
	return issues, nil
}
