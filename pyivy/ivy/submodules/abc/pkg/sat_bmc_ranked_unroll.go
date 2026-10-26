package abc

// Ranked incremental unrolling from src/sat/bmc/bmcUnroll.c.

import "slices"

const giaRankNull = 0x7fff

type GiaRankedUnrollOptions struct {
	MaxObjects, MaxRank, MaxMappingEntries int // zero: 1M, 32767, 4M
	Stop                                   func() bool
}

type GiaRankedUnrollStats struct {
	Frames, SourceObjects, ReachableObjects    int
	MaxTent, MaxRank, ExtraRankSlots           int
	TargetObjects, TargetInputs, TargetOutputs int
}

type giaRankedNode struct {
	present                               bool
	tent, rankMax, rankCur                int
	fanin0, fanin1, fanin2                GiaID
	rankDiff0, rankDiff1, rankDiff2       int
	complement0, complement1, complement2 bool
	piIndex                               int
	isPI, isPO, isBuffer, isXor, isMux    bool
	values                                []GiaLit
}

type GiaRankedUnroller struct {
	source     *GiaManager
	target     *GiaManager
	nodes      []giaRankedNode
	byTent     [][]GiaID
	piLiterals []GiaLit
	options    GiaRankedUnrollOptions
	frames     int
	reachable  int
	extraSlots int
	maxRank    int
	stopped    bool
}

type giaRankSetupEvent struct {
	id, child GiaID
	tent      int
	kind      uint8 // 0 enter, 1 update child rank, 2 exit
}

func NewGiaRankedUnroller(source *GiaManager, options GiaRankedUnrollOptions) (*GiaRankedUnroller, error) {
	if source == nil || options.MaxObjects < 0 || options.MaxRank < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxRank == 0 {
		options.MaxRank = giaRankNull
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 4 << 20
	}
	if options.MaxRank > giaRankNull || len(source.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	if err := source.Check(); err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	// Own an immutable source snapshot so later caller edits cannot alter an
	// incremental unrolling already in progress.
	source = source.Clone()
	source.HashStop()
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name, target.Spec = source.Name, source.Spec
	if err = target.HashStart(); err != nil {
		return nil, err
	}
	u := &GiaRankedUnroller{source: source, target: target, nodes: make([]giaRankedNode, len(source.objects)), options: options}
	if err = u.setup(); err != nil {
		return nil, err
	}
	return u, nil
}

func (u *GiaRankedUnroller) checkStop() error {
	if u.options.Stop != nil && u.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (u *GiaRankedUnroller) setup() error {
	source := u.source
	tents := make([]int, len(source.objects))
	ranks := make([]int, len(source.objects))
	for i := range tents {
		tents[i] = -1
	}
	tents[0] = 0
	order := []GiaID{0}
	limits := []int{1}
	roToRI := make([]GiaID, len(source.objects))
	for i := range roToRI {
		roToRI[i] = InvalidGiaID
	}
	for register := 0; register < source.registers; register++ {
		roToRI[source.cis[source.PrimaryInputCount()+register]] = source.cos[source.PrimaryOutputCount()+register]
	}
	var roots []GiaID
	setupRoot := func(root GiaID, tent int, nextRoots *[]GiaID) error {
		stack := []giaRankSetupEvent{{id: root, tent: tent}}
		for len(stack) > 0 {
			if err := u.checkStop(); err != nil {
				return err
			}
			event := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			switch event.kind {
			case 1:
				if tents[event.child] < 0 {
					return ErrInvariant
				}
				difference := max(0, event.tent-tents[event.child]-1)
				if difference > ranks[event.child] {
					ranks[event.child] = difference
				}
				continue
			case 2:
				order = append(order, event.id)
				continue
			}
			if tents[event.id] >= 0 {
				continue
			}
			tents[event.id] = event.tent
			object := source.objects[event.id]
			stack = append(stack, giaRankSetupEvent{id: event.id, tent: event.tent, kind: 2})
			if object.IsAnd() {
				first, _ := object.Fanin0(event.id)
				second, _ := object.Fanin1(event.id)
				if control, mux := source.MuxControl(event.id); mux {
					stack = append(stack, giaRankSetupEvent{child: control.ID(), tent: event.tent, kind: 1}, giaRankSetupEvent{id: control.ID(), tent: event.tent})
				}
				stack = append(stack,
					giaRankSetupEvent{child: second.ID(), tent: event.tent, kind: 1}, giaRankSetupEvent{id: second.ID(), tent: event.tent},
					giaRankSetupEvent{child: first.ID(), tent: event.tent, kind: 1}, giaRankSetupEvent{id: first.ID(), tent: event.tent},
				)
			} else if object.IsCO() {
				first, _ := object.Fanin0(event.id)
				stack = append(stack, giaRankSetupEvent{child: first.ID(), tent: event.tent, kind: 1}, giaRankSetupEvent{id: first.ID(), tent: event.tent})
			} else if ri := roToRI[event.id]; ri != InvalidGiaID {
				*nextRoots = append(*nextRoots, ri)
			}
		}
		return nil
	}
	for _, po := range source.cos[:source.PrimaryOutputCount()] {
		if err := setupRoot(po, 0, &roots); err != nil {
			return err
		}
	}
	for tent := 1; len(roots) > 0; tent++ {
		if tent >= u.options.MaxRank {
			return ErrLimit
		}
		limits = append(limits, len(order))
		next := make([]GiaID, 0)
		for _, root := range roots {
			if err := setupRoot(root, tent, &next); err != nil {
				return err
			}
		}
		roots = next
	}
	limits = append(limits, len(order))
	u.byTent = make([][]GiaID, len(limits)-1)
	piIndex := make([]int, len(source.objects))
	for i := range piIndex {
		piIndex[i] = -1
	}
	for index, id := range source.cis[:source.PrimaryInputCount()] {
		piIndex[id] = index
	}
	isPO := make([]bool, len(source.objects))
	for _, id := range source.cos[:source.PrimaryOutputCount()] {
		isPO[id] = true
	}
	for tent := range u.byTent {
		u.byTent[tent] = slices.Clone(order[limits[tent]:limits[tent+1]])
		for _, id := range u.byTent[tent] {
			object := source.objects[id]
			node := giaRankedNode{present: true, tent: tent, rankMax: ranks[id] + 1, rankCur: giaRankNull, piIndex: piIndex[id], isPI: piIndex[id] >= 0, isPO: isPO[id]}
			if node.rankMax > u.options.MaxRank {
				return ErrLimit
			}
			node.values = make([]GiaLit, node.rankMax)
			for i := range node.values {
				node.values[i] = InvalidGiaLit
			}
			setFanin := func(slot int, literal GiaLit) {
				difference := max(0, tent-tents[literal.ID()]-1)
				switch slot {
				case 0:
					node.fanin0, node.rankDiff0, node.complement0 = literal.ID(), difference, literal.Complemented()
				case 1:
					node.fanin1, node.rankDiff1, node.complement1 = literal.ID(), difference, literal.Complemented()
				case 2:
					node.fanin2, node.rankDiff2, node.complement2 = literal.ID(), difference, literal.Complemented()
				}
			}
			if object.IsAnd() {
				first, _ := object.Fanin0(id)
				second, _ := object.Fanin1(id)
				setFanin(0, first)
				setFanin(1, second)
				node.isBuffer = object.IsBuffer()
				if control, mux := source.MuxControl(id); mux {
					node.isMux = true
					setFanin(2, control)
				} else {
					node.isXor = object.IsXor()
				}
			} else if object.IsCO() {
				first, _ := object.Fanin0(id)
				setFanin(0, first)
			} else if ri := roToRI[id]; ri != InvalidGiaID {
				setFanin(0, GiaLit(ri)<<1)
			}
			u.nodes[id] = node
			u.reachable++
			u.extraSlots += node.rankMax - 1
			u.maxRank = max(u.maxRank, node.rankMax-1)
		}
	}
	u.nodes[0] = giaRankedNode{present: true, rankMax: 1, rankCur: 0, values: []GiaLit{GiaConst0}}
	for register := 0; register < source.registers; register++ {
		ri := source.cos[source.PrimaryOutputCount()+register]
		if u.nodes[ri].present {
			if err := u.setValue(ri, GiaConst0); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *GiaRankedUnroller) setValue(id GiaID, value GiaLit) error {
	node := &u.nodes[id]
	if !node.present || node.rankMax < 1 || !value.Valid() {
		return ErrInvariant
	}
	node.rankCur = ((node.rankCur + 1) & giaRankNull) % node.rankMax
	node.values[node.rankCur] = value
	return nil
}

func (u *GiaRankedUnroller) faninValue(id GiaID, difference int, complement bool) (GiaLit, error) {
	if uint64(id) >= uint64(len(u.nodes)) {
		return InvalidGiaLit, ErrInvariant
	}
	node := &u.nodes[id]
	if !node.present || node.rankCur < 0 || node.rankCur >= node.rankMax || difference < 0 || difference >= node.rankMax {
		return InvalidGiaLit, ErrInvariant
	}
	value := node.values[(node.rankCur+node.rankMax-difference)%node.rankMax]
	if !value.Valid() {
		return InvalidGiaLit, ErrInvariant
	}
	return value.NotIf(complement), nil
}

func (u *GiaRankedUnroller) clone() *GiaRankedUnroller {
	result := *u
	result.target = u.target.Clone()
	result.nodes = make([]giaRankedNode, len(u.nodes))
	for i, node := range u.nodes {
		result.nodes[i] = node
		result.nodes[i].values = slices.Clone(node.values)
	}
	result.piLiterals = slices.Clone(u.piLiterals)
	return &result
}

func (u *GiaRankedUnroller) appendFrame() error {
	if err := u.checkStop(); err != nil {
		return err
	}
	primaryInputs := u.source.PrimaryInputCount()
	for i := 0; i < primaryInputs; i++ {
		literal, err := u.target.AppendCI()
		if err != nil {
			return err
		}
		u.piLiterals = append(u.piLiterals, literal)
	}
	maximumTent := min(u.frames, len(u.byTent)-1)
	for tent := maximumTent; tent >= 0; tent-- {
		for _, id := range u.byTent[tent] {
			if err := u.checkStop(); err != nil {
				return err
			}
			node := &u.nodes[id]
			var value GiaLit
			var err error
			switch {
			case node.isPI:
				frame := u.frames - node.tent
				index := frame*primaryInputs + node.piIndex
				if frame < 0 || node.piIndex < 0 || index < 0 || index >= len(u.piLiterals) {
					return ErrInvariant
				}
				value = u.piLiterals[index]
			case u.source.objects[id].IsAnd():
				first, firstErr := u.faninValue(node.fanin0, node.rankDiff0, node.complement0)
				if firstErr != nil {
					return firstErr
				}
				if node.isBuffer {
					value = first
					break
				}
				second, secondErr := u.faninValue(node.fanin1, node.rankDiff1, node.complement1)
				if secondErr != nil {
					return secondErr
				}
				switch {
				case node.isMux:
					control, controlErr := u.faninValue(node.fanin2, node.rankDiff2, node.complement2)
					if controlErr != nil {
						return controlErr
					}
					value, err = u.target.HashMux(control, second, first)
				case node.isXor:
					value, err = u.target.HashXor(first, second)
				default:
					value, err = u.target.HashAnd(first, second)
				}
			case u.source.objects[id].IsCI():
				value, err = u.faninValue(node.fanin0, node.rankDiff0, node.complement0)
			default:
				value, err = u.faninValue(node.fanin0, node.rankDiff0, node.complement0)
			}
			if err != nil {
				return err
			}
			if err = u.setValue(id, value); err != nil {
				return err
			}
			if node.isPO {
				if _, err = u.target.AppendCO(value); err != nil {
					return err
				}
			}
		}
	}
	u.frames++
	return nil
}

// AppendFrames atomically adds count frames to the incremental unrolling.
func (u *GiaRankedUnroller) AppendFrames(count int) error {
	if u == nil || count < 0 {
		return ErrDimension
	}
	if u.stopped || u.source == nil || u.target == nil {
		return ErrInvariant
	}
	if count == 0 {
		return nil
	}
	if _, err := checkedSize(u.frames*len(u.source.objects), count, len(u.source.objects), u.options.MaxMappingEntries); err != nil {
		return ErrLimit
	}
	trial := u.clone()
	for i := 0; i < count; i++ {
		if err := trial.appendFrame(); err != nil {
			return err
		}
	}
	*u = *trial
	return nil
}

func (u *GiaRankedUnroller) Frames() int {
	if u == nil || u.stopped {
		return 0
	}
	return u.frames
}

func (u *GiaRankedUnroller) RankProfile() []int {
	if u == nil || u.stopped {
		return nil
	}
	profile := make([]int, u.maxRank+1)
	for _, node := range u.nodes {
		if node.present {
			profile[node.rankMax-1]++
		}
	}
	return profile
}

func (u *GiaRankedUnroller) Statistics() GiaRankedUnrollStats {
	if u == nil || u.stopped {
		return GiaRankedUnrollStats{}
	}
	return GiaRankedUnrollStats{
		Frames: u.frames, SourceObjects: len(u.source.objects), ReachableObjects: u.reachable,
		MaxTent: len(u.byTent) - 1, MaxRank: u.maxRank, ExtraRankSlots: u.extraSlots,
		TargetObjects: len(u.target.objects), TargetInputs: len(u.target.cis), TargetOutputs: len(u.target.cos),
	}
}

// Snapshot returns an independent combinational graph. Cleanup matches the
// one-shot native Unr_ManUnroll wrapper.
func (u *GiaRankedUnroller) Snapshot(cleanup bool) (*GiaManager, error) {
	if u == nil || u.stopped || u.target == nil {
		return nil, ErrInvariant
	}
	result := u.target.Clone()
	result.HashStop()
	if !cleanup {
		if err := result.Check(); err != nil {
			return nil, err
		}
		return result, nil
	}
	cleaned, _, err := result.Cleanup()
	return cleaned, err
}

func (u *GiaRankedUnroller) Stop() {
	if u == nil || u.stopped {
		return
	}
	u.source, u.target, u.nodes, u.byTent, u.piLiterals = nil, nil, nil, nil, nil
	u.stopped = true
}

func (m *GiaManager) UnrollRanked(frames int, options GiaRankedUnrollOptions) (*GiaManager, error) {
	if frames < 1 {
		return nil, ErrDimension
	}
	unroller, err := NewGiaRankedUnroller(m, options)
	if err != nil {
		return nil, err
	}
	defer unroller.Stop()
	if err = unroller.AppendFrames(frames); err != nil {
		return nil, err
	}
	return unroller.Snapshot(true)
}
