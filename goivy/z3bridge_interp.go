package goivy

import (
	"sync/atomic"

	"github.com/glycerine/ivy/goivy/smt"
)

// NewTranslatorWithInterpolation creates a Translator backed by an
// interpolation-capable Z3 context. Use this translator when you need
// to compute Craig interpolants. The interpolation translator owns its
// own private cache (different Z3 context, so cannot share with the
// main solver).
func (s *Solver) NewTranslatorWithInterpolation() *Translator {
	cache := &Z3SessionCache{
		Ctx:            smt.NewInterpolationZ3Context(),
		z3CheckCounter: &atomic.Int64{},
	}
	cache.resetMaps()
	return &Translator{
		s:     s,
		cache: cache,
		Ctx:   cache.Ctx,
	}
}
