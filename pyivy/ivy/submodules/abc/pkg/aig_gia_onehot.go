package abc

// Exactly-one violation generation from src/aig/gia/giaDup.c:Gia_ManOneHot.

// NewGiaOneHot builds a graph whose outputs are all false exactly when one
// of variables inputs is true. The first skips CIs are ignored. Pairwise group
// overlap outputs detect multiple bits; the final output detects no bits.
// These are violation outputs, not positive environment assumptions. At least
// one variable is required. MaxMappingEntries bounds power-of-two scratch.
func NewGiaOneHot(skips, variables int, options GiaCofactorOptions) (*GiaManager, error) {
	if skips < 0 || variables < 1 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	base, err := checkedSize(2, 1, skips, options.MaxObjects)
	if err != nil {
		return nil, err
	}
	if _, err = checkedSize(base, 1, variables, options.MaxObjects); err != nil {
		return nil, err
	}
	slots := 1
	for slots < variables {
		if slots > options.MaxMappingEntries/2 {
			return nil, ErrLimit
		}
		slots *= 2
	}
	if slots > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	g.Name = "onehot"
	for i := 0; i < skips; i++ {
		if stopped() {
			return nil, ErrCancelled
		}
		if _, err = g.AppendCI(); err != nil {
			return nil, err
		}
	}
	temporary := make([]GiaLit, slots)
	for i := 0; i < variables; i++ {
		if stopped() {
			return nil, ErrCancelled
		}
		temporary[i], err = g.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	for shift := 1; shift < slots; shift *= 2 {
		for start := 0; start < slots; start += 2 * shift {
			if stopped() {
				return nil, ErrCancelled
			}
			overlap, err := g.HashAnd(temporary[start], temporary[start+shift])
			if err != nil {
				return nil, err
			}
			if overlap != GiaConst0 {
				if _, err = g.AppendCO(overlap); err != nil {
					return nil, err
				}
			}
			temporary[start], err = g.HashOr(temporary[start], temporary[start+shift])
			if err != nil {
				return nil, err
			}
		}
	}
	g.HashStop()
	if stopped() {
		return nil, ErrCancelled
	}
	if _, err = g.AppendCO(temporary[0].Not()); err != nil {
		return nil, err
	}
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}
