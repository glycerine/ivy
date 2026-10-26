package abc

// Initialized unrolling/reconstruction from src/aig/gia/giaEnable.c:
// Gia_ManUnrollAndCofactor.

type GiaUnrollCofactorOptions struct {
	GiaUnrollOptions
	FanoutThreshold   int
	MaxCofactorPasses int
}

// UnrollAndCofactor builds initialized timeframes, selects structural
// sequential candidates above the fanout threshold and Shannon-reconstructs
// their frame copies. It preserves the unrolled Boolean function. Outputs
// remain frame-major, including interleaved constraints; global constraint
// count is zero. Initial nil matches C's zero initialization. Source scratch
// is preserved. C's 200 selected-frame-signal cap is retained.
func (m *GiaManager) UnrollAndCofactor(options GiaUnrollCofactorOptions) (*GiaManager, error) {
	if options.FanoutThreshold < 0 || options.MaxCofactorPasses < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 4 << 20
	}
	frames, err := m.UnrollWithOptions(options.GiaUnrollOptions)
	if err != nil {
		return nil, err
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	signals, err := m.SequentialSignalsAboveFanout(options.FanoutThreshold, options.MaxObjects)
	if err != nil {
		return nil, err
	}
	// UnrollWithOptions already cleans and remaps its object copies. Compose
	// with identity on that cleaned graph rather than borrowing Value scratch.
	graph := frames.Graph()
	if graph.ObjectCount() > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	identity := make([]GiaLit, graph.ObjectCount())
	for i := range identity {
		identity[i] = GiaLit(i) << 1
	}
	selected, err := TransferGiaFrameSignals(m, frames, graph, identity, signals, options.MaxMappingEntries)
	if err != nil {
		return nil, err
	}
	if stopped() {
		return nil, ErrCancelled
	}
	return graph.CofactorRebuildAll(selected, GiaMultiCofactorOptions{
		GiaCofactorOptions: GiaCofactorOptions{MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries, Stop: options.Stop},
		MaxPasses:          options.MaxCofactorPasses,
	})
}
