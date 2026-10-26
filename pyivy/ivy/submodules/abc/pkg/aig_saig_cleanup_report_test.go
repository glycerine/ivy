package abc

import (
	"errors"
	"slices"
	"testing"
)

func TestAigUselessRegisterReport(t *testing.T) {
	g := NewAigManager()
	p0 := aigTestCI(t, g)
	p1 := aigTestCI(t, g)
	r0 := aigTestCI(t, g)
	_ = aigTestCI(t, g)
	aigTestCO(t, g, r0.Not())
	extra, err := g.And(p1, r0)
	if err != nil {
		t.Fatal(err)
	}
	aigTestCO(t, g, extra)
	aigTestCO(t, g, p0.Not())
	aigTestCO(t, g, p1)
	if err = g.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	before := slices.Clone(g.objects)
	report, err := g.UselessRegisterReport(AigSequentialDiagnosticOptions{})
	if err != nil || report != (AigUselessRegisterReport{Registers: 2, RegistersDrivingPrimaryOutput: 0, RegistersDrivenByPrimaryInput: 1}) {
		t.Fatal(report, err)
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("source mutation")
	}
	for _, options := range []AigSequentialDiagnosticOptions{{MaxObjects: -1}, {MaxObjects: 1}} {
		if got, e := g.UselessRegisterReport(options); e == nil || got != (AigUselessRegisterReport{}) {
			t.Fatal(options, got, e)
		}
	}
	calls := 0
	if got, e := g.UselessRegisterReport(AigSequentialDiagnosticOptions{Stop: func() bool { calls++; return calls == 2 }}); !errors.Is(e, ErrCancelled) || got != (AigUselessRegisterReport{}) {
		t.Fatal(calls, got, e)
	}
}

func TestAigComplementedRegisterPairs(t *testing.T) {
	g := NewAigManager()
	p := aigTestCI(t, g)
	for i := 0; i < 3; i++ {
		aigTestCI(t, g)
	}
	for _, driver := range []AigLit{p, p.Not(), p} {
		aigTestCO(t, g, driver)
	}
	if err := g.SetRegisterCount(3); err != nil {
		t.Fatal(err)
	}
	got, err := g.ComplementedRegisterPairs(AigSequentialDiagnosticOptions{})
	if err != nil || got != 2 {
		t.Fatal(got, err)
	}
	if got, err = g.ComplementedRegisterPairs(AigSequentialDiagnosticOptions{MaxObjects: g.ObjectCount() - 1}); !errors.Is(err, ErrLimit) || got != 0 {
		t.Fatal(got, err)
	}
	calls := 0
	if got, err = g.ComplementedRegisterPairs(AigSequentialDiagnosticOptions{Stop: func() bool { calls++; return calls == 2 }}); !errors.Is(err, ErrCancelled) || got != 0 {
		t.Fatal(calls, got, err)
	}
	empty := NewAigManager()
	if got, err = empty.ComplementedRegisterPairs(AigSequentialDiagnosticOptions{}); !errors.Is(err, ErrDimension) || got != 0 {
		t.Fatal(got, err)
	}
}

func FuzzAigSequentialDiagnostics(f *testing.F) {
	f.Add(uint8(0), uint8(1), uint8(2), uint8(3))
	f.Fuzz(func(t *testing.T, po0, po1, ri0, ri1 uint8) {
		g := NewAigManager()
		literals := make([]AigLit, 0, 5)
		literals = append(literals, AigConst1)
		for i := 0; i < 4; i++ {
			literals = append(literals, aigTestCI(t, g))
		}
		selectors := []uint8{po0, po1, ri0, ri1}
		for i, selector := range selectors {
			literal := literals[int(selector)%len(literals)].NotIf(selector&8 != 0)
			selectors[i] = uint8(literal.ID())
			aigTestCO(t, g, literal)
		}
		if err := g.SetRegisterCount(2); err != nil {
			t.Fatal(err)
		}
		report, err := g.UselessRegisterReport(AigSequentialDiagnosticOptions{})
		if err != nil {
			t.Fatal(err)
		}
		wantPO, wantPI := 0, 0
		for _, id := range selectors[:2] {
			if (id == 3 || id == 4) && g.objects[id].RefCount == 1 {
				wantPO++
			}
		}
		for _, id := range selectors[2:] {
			if (id == 1 || id == 2) && g.objects[id].RefCount == 1 {
				wantPI++
			}
		}
		if report.Registers != 2 || report.RegistersDrivingPrimaryOutput != wantPO || report.RegistersDrivenByPrimaryInput != wantPI {
			t.Fatal(selectors, report, wantPO, wantPI)
		}
		duplicates, err := g.ComplementedRegisterPairs(AigSequentialDiagnosticOptions{})
		wantDuplicates := 0
		if selectors[2] == selectors[3] {
			wantDuplicates = 1
		}
		if err != nil || duplicates != wantDuplicates {
			t.Fatal(selectors, duplicates, wantDuplicates, err)
		}
	})
}
