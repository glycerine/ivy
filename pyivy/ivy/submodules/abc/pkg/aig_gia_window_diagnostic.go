package abc

// Whole-graph window and owned cut diagnostics from aig/gia/giaSatLut.c:
// Sbl_ManWindow, Sbl_ManPrintCut and Sbl_ManFindAndPrintCut.

import (
	"fmt"
	"math/bits"
	"strings"
)

// GiaFullWindow is the deterministic whole-graph boundary used by the
// diagnostic Sbl_ManWindow path. Roots retain CO order and duplicates, leaves
// retain CI order, and Ands retain object order.
type GiaFullWindow struct {
	Roots, Leaves, Ands []GiaID
}

// GiaRemapPassStats is the deterministic replacement for the pass summary and
// wall-clock runtime printouts. Logical work is retained; unstable elapsed-time
// sampling belongs in benchmarks rather than library results.
type GiaRemapPassStats struct {
	Tried, Improved, Repeated, Small, Large int
	DiscoveryBuilds, ReferenceRefreshes     int
	SolverBuilds, WindowResets              int
}

// Stats summarizes a completed whole-network remapping pass.
func (r GiaUnitRemapPassResult) Stats() GiaRemapPassStats {
	s := GiaRemapPassStats{
		Tried: len(r.Attempts), Improved: r.Improvements,
		DiscoveryBuilds: r.DiscoveryBuilds, ReferenceRefreshes: r.ReferenceRefreshes,
		SolverBuilds: r.SolverBuilds, WindowResets: r.WindowResets,
	}
	for _, attempt := range r.Attempts {
		switch attempt.Reason {
		case "repeated":
			s.Repeated++
		case "small":
			s.Small++
		case "large":
			s.Large++
		}
	}
	return s
}

// Format returns an owned stable summary instead of writing global stdout.
func (s GiaRemapPassStats) Format() string {
	return fmt.Sprintf("Tried=%d Improved=%d Repeated=%d Small=%d Large=%d DiscoveryBuilds=%d ReferenceRefreshes=%d SolverBuilds=%d WindowResets=%d\n", s.Tried, s.Improved, s.Repeated, s.Small, s.Large, s.DiscoveryBuilds, s.ReferenceRefreshes, s.SolverBuilds, s.WindowResets)
}

// FullWindow returns the complete graph window without borrowing manager
// storage or mutating object scratch. MaxObjects bounds aggregate output and
// MaxVisits bounds inspected objects and terminals.
func (m *GiaManager) FullWindow(o GiaLUTAnalysisOptions) (GiaFullWindow, error) {
	fail := func(e error) (GiaFullWindow, error) { return GiaFullWindow{}, e }
	if m == nil || len(m.objects) == 0 || o.MaxObjects < 0 || o.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxVisits == 0 {
		o.MaxVisits = 1 << 20
	}
	total, e := checkedSize(0, 1, len(m.cis)+len(m.cos)+m.AndCount(), o.MaxObjects)
	if e != nil {
		return fail(e)
	}
	_ = total
	visits := 0
	charge := func() error {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		if visits == o.MaxVisits {
			return ErrLimit
		}
		visits++
		return nil
	}
	r := GiaFullWindow{
		Roots:  make([]GiaID, 0, len(m.cos)),
		Leaves: make([]GiaID, 0, len(m.cis)),
		Ands:   make([]GiaID, 0, m.AndCount()),
	}
	for _, id := range m.cis {
		if e = charge(); e != nil {
			return fail(e)
		}
		r.Leaves = append(r.Leaves, id)
	}
	for id, object := range m.objects {
		if e = charge(); e != nil {
			return fail(e)
		}
		if object.IsAnd() {
			r.Ands = append(r.Ands, GiaID(id))
		}
	}
	for _, id := range m.cos {
		if e = charge(); e != nil {
			return fail(e)
		}
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return fail(err)
		}
		r.Roots = append(r.Roots, driver.ID())
	}
	if o.Stop != nil && o.Stop() {
		return fail(ErrCancelled)
	}
	return r, nil
}

// FormatGiaWindowCut returns the native diagnostic spelling and population.
// The result is owned and deterministic; no process-global output is used.
func FormatGiaWindowCut(c GiaWindowCut) (string, int) {
	var out strings.Builder
	out.WriteString("{ ")
	count := 0
	for half, word := range c.Leaves {
		count += bits.OnesCount64(word)
		for word != 0 {
			bit := bits.TrailingZeros64(word)
			fmt.Fprintf(&out, "i%d ", 64*half+bit)
			word &= word - 1
		}
	}
	out.WriteByte(' ')
	for half, word := range c.Nodes {
		count += bits.OnesCount64(word)
		for word != 0 {
			bit := bits.TrailingZeros64(word)
			fmt.Fprintf(&out, "n%d ", 64*half+bit)
			word &= word - 1
		}
	}
	out.WriteString("};\n")
	return out.String(), count
}
