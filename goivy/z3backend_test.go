//go:build !tinygo

package goivy

import "testing"

type recordingZ3Backend struct {
	back     BACK
	contexts int
	interps  int
	solvers  int
}

func (b *recordingZ3Backend) Name() BACK {
	return b.back
}

func (b *recordingZ3Backend) NewZ3Context() *Z3Context {
	b.contexts++
	return newCGoZ3Context(b)
}

func (b *recordingZ3Backend) NewInterpolationZ3Context() *Z3Context {
	b.interps++
	return newCGoInterpolationZ3Context(b)
}

func (b *recordingZ3Backend) NewZ3Solver(ctx *Z3Context) *Z3Solver {
	b.solvers++
	return newCGoZ3Solver(ctx)
}

func TestNewSolverUsesConfiguredZ3BackendInterface(t *testing.T) {
	backend := &recordingZ3Backend{back: Recording}
	cfg := NewConfig()
	cfg.BackendName = Recording
	cfg.Backend = backend

	mod := NewModule()
	mod.Cfg = cfg

	s := NewSolver(mod, nil)
	defer s.z3u.Close()
	defer s.tr.Ctx.Close()

	if cfg.BackendName != "recording" {
		t.Fatalf("ResolveBackend did not fill BackendName from injected backend: %q", cfg.BackendName)
	}
	if backend.contexts != 2 {
		t.Fatalf("configured backend created %d contexts, want 2", backend.contexts)
	}
	if s.backend != backend {
		t.Fatalf("solver backend = %T, want injected recording backend", s.backend)
	}
	if s.tr.Ctx.backend != backend {
		t.Fatalf("translator context backend = %T, want injected recording backend", s.tr.Ctx.backend)
	}
	if s.z3u.Ctx.backend != backend {
		t.Fatalf("z3 utils context backend = %T, want injected recording backend", s.z3u.Ctx.backend)
	}

	_ = s.tr.Ctx.NewZ3Solver()
	if backend.solvers != 1 {
		t.Fatalf("ctx.NewZ3Solver used backend %d times, want 1", backend.solvers)
	}

	itpTr := s.NewTranslatorWithInterpolation()
	defer itpTr.Ctx.Close()
	if backend.interps != 1 {
		t.Fatalf("NewTranslatorWithInterpolation used backend %d times, want 1", backend.interps)
	}
}
