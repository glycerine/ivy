package abc

// Native packed/mapping2 conversion from aig/gia/giaSplit.c:
// Spl_ManToWecMapping and Spl_ManFromWecMapping.

// AlternateNetwork converts all native fanin widths, unlike the four-input
// edge-engine Network adapter. Empty records and the constant-zero record are
// dropped, as in C. Signed packed footers are not retained by mapping2.
func (m *GiaLUTMapping) AlternateNetwork(o GiaLUTAnalysisOptions) (*GiaLUTNetwork, error) {
	return m.alternateNetwork(o, 0)
}

// PackedMapping emits nonempty records in ascending root order, with root-ID
// footers. Empty records, stale packed records and previous footer metadata are
// absent. MaxEntries counts the header and every record integer; zero is one
// million. MaxVisits separately bounds copied fanin slots. Source is unchanged.
func (n *GiaLUTNetwork) PackedMapping(maxEntries int, o GiaLUTAnalysisOptions) (*GiaLUTMapping, error) {
	if n == nil || len(n.fanins) == 0 || len(n.fanins[0]) != 0 || maxEntries < 0 || o.MaxObjects < 0 || o.MaxVisits < 0 {
		return nil, ErrDimension
	}
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxVisits == 0 {
		o.MaxVisits = 1 << 20
	}
	capEntries := min(maxEntries, int(^uint32(0)>>1))
	if len(n.fanins) > o.MaxObjects || len(n.fanins) > capEntries {
		return nil, ErrLimit
	}
	stopped := func() bool { return o.Stop != nil && o.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	total, slots := len(n.fanins), 0
	for root, fanins := range n.fanins {
		if stopped() {
			return nil, ErrCancelled
		}
		if len(fanins) == 0 {
			continue
		}
		var e error
		total, e = checkedSize(total, 1, len(fanins)+2, capEntries)
		if e != nil {
			return nil, e
		}
		if len(fanins) > o.MaxVisits-slots {
			return nil, ErrLimit
		}
		slots += len(fanins)
		for _, f := range fanins {
			if f >= GiaID(root) {
				return nil, ErrDimension
			}
		}
	}
	packed := make([]int, len(n.fanins), total)
	for root, fanins := range n.fanins {
		if stopped() {
			return nil, ErrCancelled
		}
		if len(fanins) == 0 {
			continue
		}
		packed[root] = len(packed)
		packed = append(packed, len(fanins))
		for _, f := range fanins {
			if stopped() {
				return nil, ErrCancelled
			}
			packed = append(packed, int(f))
		}
		packed = append(packed, root)
	}
	if stopped() {
		return nil, ErrCancelled
	}
	return &GiaLUTMapping{objects: len(n.fanins), packed: packed}, nil
}

type GiaPackedUnitRemapPassResult struct {
	Mapping *GiaLUTMapping
	Pass    GiaUnitRemapPassResult
}

// RemapUnitPass performs checked packed -> mapping2 -> unit pass -> packed
// conversion on owned storage. Native packing/footer metadata is discarded by
// this conversion, as in Gia_ManLutSat. Unit search supports native four/six-input feasibility.
func (mapping *GiaLUTMapping) RemapUnitPass(m *GiaManager, delayStart int, o GiaUnitRemapPassOptions) (GiaPackedUnitRemapPassResult, error) {
	fail := func(e error) (GiaPackedUnitRemapPassResult, error) { return GiaPackedUnitRemapPassResult{}, e }
	if o.MaxPackedEntries < 0 {
		return fail(ErrDimension)
	}
	analysis := o.Search.Analysis
	prior := analysis.Stop
	analysis.Stop = func() bool { return o.Search.Stop != nil && o.Search.Stop() || prior != nil && prior() }
	n, e := mapping.AlternateNetwork(analysis)
	if e != nil {
		return fail(e)
	}
	pass, e := n.RemapUnitPass(m, delayStart, o)
	if e != nil {
		return fail(e)
	}
	packed, e := pass.Network.PackedMapping(o.MaxPackedEntries, analysis)
	if e != nil {
		return fail(e)
	}
	return GiaPackedUnitRemapPassResult{Mapping: packed, Pass: pass}, nil
}
