package fracomb

import "testing"

// The pigeonhole miter and limits were also run through ABC's Fra_FraigSat:
// both conflict and propagation limit 1 return undecided, while 100000 proves
// UNSAT. Its solver2 path asserts on the first zero-ID CNF clause.
func TestSATResourceLimitsMatchSource(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		conflicts, propagations int64
		flip                    int32
		want                    int32
	}{
		{"conflict exhaustion", 1, 0, 1, -1},
		{"propagation exhaustion", 0, 1, 0, -1},
		{"conflict budget", 100000, 0, 1, 1},
		{"propagation budget", 0, 100000, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := cecPigeonholeMiter()
			got := Fra_FraigSat(g, tc.conflicts, tc.propagations, 0, 0, 0, tc.flip, 0, 0, 0)
			if got != tc.want || g.Data != nil {
				t.Fatalf("status=%d want=%d model=%v", got, tc.want, g.Data)
			}
		})
	}
}
