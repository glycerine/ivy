package abc

// Fast sequential AIG simulation from src/opt/fsim/{fsim.h,fsimInt.h,
// fsimFront.c,fsimMan.c,fsimSim.c}.

import (
	"fmt"
	"math"
	"math/bits"
	"slices"
	"time"
)

// FsimBuildOptions bounds construction of the frontier representation. A zero
// limit selects a conservative package default.
type FsimBuildOptions struct {
	Compress        bool
	MaxObjects      int
	MaxFrontier     int
	MaxEncodedBytes int
	Stop            func() bool
}

// FsimRunOptions controls deterministic random patterns and resource use.
// Seed zero selects the exact reset Aig_ManRandom sequence; nonzero seeds use
// an owned deterministic extension for independently reproducible campaigns.
type FsimRunOptions struct {
	Seed      uint64
	MaxWords  int
	MaxFrames int
	MaxBytes  int
	Stop      func() bool
}

// FsimObject is one owned entry in the frontier-encoded AIG. A CI has both
// fanins zero, a CO has Node and Fanin1 zero, and an AND has all fields nonzero.
type FsimObject struct {
	Node, Fanin0, Fanin1 int
}

// FsimManagerStatistics describes the immutable manager representation.
type FsimManagerStatistics struct {
	PrimaryInputs, PrimaryOutputs int
	Inputs, Outputs, Nodes        int
	Objects, MaximumCrossCut      int
	Frontier, MaximumFrontierID   int
	CompressedBytes               int
}

// FsimResult is the owned result of a packed simulation campaign.
type FsimResult struct {
	Counterexample *Counterexample
	Frames         int
	Patterns       int
	TimedOut       bool
	Property       int
	Pattern        int
}

// FsimManager owns the source's DFS object order and frontier representation.
// It does not retain borrowed mutable slices from callers.
type FsimManager struct {
	graph                         *AigManager
	nPis, nPos, nCis, nCos, nodes int
	fans0, fans1, refs            []int
	cisToPIO, latchOutputs        []int
	latchInputs                   []int
	maximumCrossCut, frontier     int
	constantFront, maximumID      int
	records                       []FsimObject
	compressed                    []byte
}

func normalizeFsimBuildOptions(options FsimBuildOptions) (FsimBuildOptions, error) {
	if options.MaxObjects < 0 || options.MaxFrontier < 0 || options.MaxEncodedBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 24
	}
	if options.MaxFrontier == 0 {
		options.MaxFrontier = 1 << 24
	}
	if options.MaxEncodedBytes == 0 {
		options.MaxEncodedBytes = 1 << 28
	}
	return options, nil
}

func fsimVar2Lit(variable int, complement bool) int {
	if complement {
		return variable<<1 | 1
	}
	return variable << 1
}
func fsimLit2Var(literal int) int     { return literal >> 1 }
func fsimLitIsCompl(literal int) bool { return literal&1 != 0 }
func fsimLitNot(literal int) int      { return literal ^ 1 }
func fsimLitNotIf(literal int, c bool) int {
	if c {
		return literal ^ 1
	}
	return literal
}
func fsimLitRegular(literal int) int { return literal &^ 1 }

// NewFsimManager implements Fsim_ManCreate and Fsim_ManFront without mutating
// AigManager scratch fields. The returned manager owns all derived storage.
func NewFsimManager(graph *AigManager, options FsimBuildOptions) (*FsimManager, error) {
	options, err := normalizeFsimBuildOptions(options)
	if err != nil {
		return nil, err
	}
	if graph == nil {
		return nil, ErrDimension
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	manager := &FsimManager{
		graph: graph, nPis: graph.PrimaryInputCount(), nPos: graph.PrimaryOutputCount(),
		nCis: graph.InputCount(), nCos: graph.OutputCount(),
		fans0: []int{0, 0}, fans1: []int{0, 0}, refs: []int{0, int(graph.objects[0].RefCount)},
	}
	for _, object := range graph.objects {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if object.Type == AigAND {
			manager.nodes++
		} else if object.Type == AigBuffer || object.Type == AigXOR || object.Type == AigVoid {
			return nil, fmt.Errorf("%w: FSIM requires a buffer-free AND AIG", ErrUnsupported)
		}
	}
	expected := manager.nCis + manager.nCos + manager.nodes + 2
	if expected > options.MaxObjects {
		return nil, ErrLimit
	}
	mapping := make([]int, len(graph.objects))
	for index := range mapping {
		mapping[index] = -1
	}
	mapping[0] = 1
	crossCut := 0
	if manager.refs[1] != 0 {
		crossCut = 1
		manager.maximumCrossCut = 1
	}
	ciPIO := make([]int, len(graph.objects))
	for index, id := range graph.cis {
		ciPIO[id] = index
	}
	var create func(AigID) (int, error)
	create = func(id AigID) (int, error) {
		if options.Stop != nil && options.Stop() {
			return 0, ErrCancelled
		}
		if uint64(id) >= uint64(len(graph.objects)) || graph.objects[id].Type == AigNone {
			return 0, ErrInvalidLiteral
		}
		if mapping[id] >= 0 {
			return mapping[id], nil
		}
		object := graph.objects[id]
		fanin0, fanin1 := 0, 0
		switch object.Type {
		case AigAND:
			left, createErr := create(object.Fanin0.ID())
			if createErr != nil {
				return 0, createErr
			}
			right, createErr := create(object.Fanin1.ID())
			if createErr != nil {
				return 0, createErr
			}
			if left == right || manager.refs[left] == 0 || manager.refs[right] == 0 {
				return 0, ErrInvariant
			}
			manager.refs[left]--
			if manager.refs[left] == 0 {
				crossCut--
			}
			manager.refs[right]--
			if manager.refs[right] == 0 {
				crossCut--
			}
			fanin0 = fsimVar2Lit(left, object.Fanin0.Complemented())
			fanin1 = fsimVar2Lit(right, object.Fanin1.Complemented())
			if fanin0 < fanin1 {
				fanin0, fanin1 = fanin1, fanin0
			}
			if len(graph.equivs) != 0 && graph.equivs[id] != InvalidAigID {
				if _, createErr = create(graph.equivs[id]); createErr != nil {
					return 0, createErr
				}
			}
		case AigCO:
			fanin, createErr := create(object.Fanin0.ID())
			if createErr != nil {
				return 0, createErr
			}
			if manager.refs[fanin] == 0 {
				return 0, ErrInvariant
			}
			manager.refs[fanin]--
			if manager.refs[fanin] == 0 {
				crossCut--
			}
			fanin0 = fsimVar2Lit(fanin, object.Fanin0.Complemented())
		case AigCI:
			manager.cisToPIO = append(manager.cisToPIO, ciPIO[id])
		default:
			return 0, ErrUnsupported
		}
		if len(manager.fans0) >= options.MaxObjects {
			return 0, ErrLimit
		}
		internal := len(manager.fans0)
		manager.fans0 = append(manager.fans0, fanin0)
		manager.fans1 = append(manager.fans1, fanin1)
		manager.refs = append(manager.refs, int(object.RefCount))
		mapping[id] = internal
		if object.RefCount != 0 {
			crossCut++
			manager.maximumCrossCut = max(manager.maximumCrossCut, crossCut)
		}
		return internal, nil
	}
	for _, id := range graph.cis {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if graph.objects[id].RefCount == 0 {
			if _, err = create(id); err != nil {
				return nil, err
			}
		}
	}
	for _, id := range graph.cos {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if _, err = create(id); err != nil {
			return nil, err
		}
	}
	if len(manager.fans0) != expected || len(manager.cisToPIO) != manager.nCis || crossCut != 0 {
		return nil, fmt.Errorf("%w: incomplete FSIM DFS", ErrInvariant)
	}
	// Fsim_ManCreate_rec consumes the temporary references while deriving the
	// DFS order, then Fsim_ManCreate restores them for frontier allocation.
	for id, internal := range mapping {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if internal >= 0 {
			manager.refs[internal] = int(graph.objects[id].RefCount)
		}
	}
	manager.latchOutputs = make([]int, graph.registers)
	manager.latchInputs = make([]int, graph.registers)
	for index := 0; index < graph.registers; index++ {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		manager.latchOutputs[index] = mapping[graph.cis[manager.nPis+index]]
		manager.latchInputs[index] = mapping[graph.cos[manager.nPos+index]]
	}
	manager.frontier = 1 + manager.maximumCrossCut*11/10
	if manager.frontier < manager.maximumCrossCut+2 {
		// Slot zero is a literal sentinel and an AND result is allocated before
		// its fanins are released. The native 10% heuristic can therefore loop
		// forever for cuts below ten; reserve the required transient slot.
		manager.frontier = manager.maximumCrossCut + 2
	}
	if manager.frontier > options.MaxFrontier {
		return nil, ErrLimit
	}
	if err = manager.buildFrontier(options); err != nil {
		return nil, err
	}
	if err = manager.Verify(); err != nil {
		return nil, err
	}
	return manager, nil
}

func (manager *FsimManager) nextFrontier(used []bool, number *int, stop func() bool) (int, error) {
	for attempts := 0; attempts <= 2*len(used)+1; attempts++ {
		if stop != nil && stop() {
			return 0, ErrCancelled
		}
		if *number >= math.MaxInt32-manager.frontier {
			return 0, ErrLimit
		}
		if *number%manager.frontier == 0 {
			*number++
		}
		place := *number % manager.frontier
		if !used[place] {
			used[place] = true
			return *number, nil
		}
		*number++
	}
	return 0, fmt.Errorf("%w: no free frontier slot size=%d used=%v number=%d", ErrInvariant, len(used), used, *number)
}

func (manager *FsimManager) buildFrontier(options FsimBuildOptions) error {
	used := make([]bool, manager.frontier)
	idToFront := make([]int, len(manager.fans0))
	for index := range idToFront {
		idToFront[index] = -1
	}
	refs := slices.Clone(manager.refs)
	number := 1
	crossCut, maximum := 0, 0
	if refs[1] != 0 {
		front, err := manager.nextFrontier(used, &number, options.Stop)
		if err != nil {
			return err
		}
		manager.constantFront = front
		idToFront[1] = front
		crossCut = 1
	}
	manager.records = make([]FsimObject, 0, len(manager.fans0)-2)
	for id := 2; id < len(manager.fans0); id++ {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		var record FsimObject
		if manager.fans0[id] == 0 {
			front, err := manager.nextFrontier(used, &number, options.Stop)
			if err != nil {
				return err
			}
			idToFront[id] = front
			record.Node = front
			if refs[id] == 0 {
				used[front%manager.frontier] = false
				idToFront[id] = -1
			}
		} else if manager.fans1[id] == 0 {
			fanin := fsimLit2Var(manager.fans0[id])
			if fanin < 0 || fanin >= len(idToFront) || idToFront[fanin] <= 0 || refs[id] != 0 {
				return fmt.Errorf("%w: CO %d fanin=%d front=%v refs=%d", ErrInvariant, id, fanin, idToFront, refs[id])
			}
			record.Fanin0 = fsimVar2Lit(idToFront[fanin], fsimLitIsCompl(manager.fans0[id]))
			refs[fanin]--
			if refs[fanin] == 0 {
				used[idToFront[fanin]%manager.frontier] = false
				idToFront[fanin] = -1
				crossCut--
			}
		} else {
			left, right := fsimLit2Var(manager.fans0[id]), fsimLit2Var(manager.fans1[id])
			if left >= len(idToFront) || right >= len(idToFront) || idToFront[left] <= 0 || idToFront[right] <= 0 {
				return fmt.Errorf("%w: AND %d fanins=%d/%d front=%v", ErrInvariant, id, left, right, idToFront)
			}
			front, err := manager.nextFrontier(used, &number, options.Stop)
			if err != nil {
				return err
			}
			idToFront[id] = front
			record.Node = front
			record.Fanin0 = fsimVar2Lit(idToFront[left], fsimLitIsCompl(manager.fans0[id]))
			record.Fanin1 = fsimVar2Lit(idToFront[right], fsimLitIsCompl(manager.fans1[id]))
			for _, fanin := range []int{left, right} {
				refs[fanin]--
				if refs[fanin] == 0 {
					used[idToFront[fanin]%manager.frontier] = false
					idToFront[fanin] = -1
					crossCut--
				}
			}
			if refs[id] == 0 {
				used[front%manager.frontier] = false
				idToFront[id] = -1
			}
		}
		manager.records = append(manager.records, record)
		if refs[id] != 0 {
			crossCut++
			maximum = max(maximum, crossCut)
		}
	}
	if crossCut != 0 || maximum != manager.maximumCrossCut {
		return fmt.Errorf("%w: frontier cross-cut final=%d maximum=%d want=%d", ErrInvariant, crossCut, maximum, manager.maximumCrossCut)
	}
	manager.maximumID = number
	if options.Compress {
		encoded, err := EncodeFsimObjects(manager.records, options.MaxEncodedBytes)
		if err != nil {
			return err
		}
		manager.compressed = encoded
	}
	return nil
}

func appendFsimNumber(destination []byte, value, maximum int) ([]byte, error) {
	if value < 0 {
		return nil, ErrDimension
	}
	for {
		if len(destination) == maximum {
			return nil, ErrLimit
		}
		part := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			part |= 0x80
		}
		destination = append(destination, part)
		if value == 0 {
			return destination, nil
		}
	}
}

// EncodeFsimObjects implements the compressed Fsim_ManStoreObj stream.
func EncodeFsimObjects(objects []FsimObject, maximumBytes int) ([]byte, error) {
	if maximumBytes < 0 {
		return nil, ErrDimension
	}
	if maximumBytes == 0 {
		maximumBytes = 1 << 28
	}
	data := make([]byte, 0, min(maximumBytes, len(objects)*3))
	previous := 0
	for _, object := range objects {
		var err error
		switch {
		case object.Fanin0 != 0 && object.Fanin1 != 0:
			if object.Node <= 0 || object.Node < previous || object.Node<<1 <= object.Fanin0 || object.Fanin0 <= object.Fanin1 {
				return nil, ErrInvariant
			}
			data, err = appendFsimNumber(data, (object.Node-previous)<<2|3, maximumBytes)
			if err == nil {
				data, err = appendFsimNumber(data, object.Node<<1-object.Fanin0, maximumBytes)
			}
			if err == nil {
				data, err = appendFsimNumber(data, object.Fanin0-object.Fanin1, maximumBytes)
			}
			previous = object.Node
		case object.Fanin0 == 0 && object.Fanin1 == 0:
			if object.Node <= 0 || object.Node < previous {
				return nil, ErrInvariant
			}
			data, err = appendFsimNumber(data, (object.Node-previous)<<2|1, maximumBytes)
			previous = object.Node
		case object.Node == 0 && object.Fanin0 != 0 && object.Fanin1 == 0:
			if (previous<<1 | 1) < object.Fanin0 {
				return nil, ErrInvariant
			}
			data, err = appendFsimNumber(data, ((previous<<1|1)-object.Fanin0)<<1, maximumBytes)
		default:
			return nil, ErrInvariant
		}
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func restoreFsimNumber(data []byte, position *int) (int, error) {
	value := uint64(0)
	for shift := uint(0); ; shift += 7 {
		if *position >= len(data) {
			return 0, ErrFormat
		}
		if shift >= 35 {
			return 0, ErrLimit
		}
		part := data[*position]
		*position++
		value |= uint64(part&0x7f) << shift
		if part&0x80 == 0 {
			if value > math.MaxInt32 {
				return 0, ErrLimit
			}
			return int(value), nil
		}
	}
}

// DecodeFsimObjects implements the checked Fsim_ManRestoreObj stream.
func DecodeFsimObjects(data []byte, count int) ([]FsimObject, error) {
	if count < 0 {
		return nil, ErrDimension
	}
	// Every encoded object occupies at least one byte. Check this before using
	// the caller-controlled count as a slice capacity so malformed input cannot
	// request an allocation larger than the stream can possibly describe.
	if count > len(data) {
		return nil, ErrFormat
	}
	objects := make([]FsimObject, 0, count)
	position, previous := 0, 0
	for len(objects) < count {
		value, err := restoreFsimNumber(data, &position)
		if err != nil {
			return nil, err
		}
		var object FsimObject
		switch value & 3 {
		case 3:
			first, firstErr := restoreFsimNumber(data, &position)
			second, secondErr := restoreFsimNumber(data, &position)
			if firstErr != nil {
				return nil, firstErr
			}
			if secondErr != nil {
				return nil, secondErr
			}
			object.Node = previous + value>>2
			object.Fanin0 = object.Node<<1 - first
			object.Fanin1 = object.Fanin0 - second
			previous = object.Node
		case 1:
			object.Node = previous + value>>2
			previous = object.Node
		default:
			object.Fanin0 = (previous<<1 | 1) - value>>1
		}
		objects = append(objects, object)
	}
	if position != len(data) {
		return nil, ErrInvariant
	}
	canonical, err := EncodeFsimObjects(objects, math.MaxInt32)
	if err != nil || !slices.Equal(canonical, data) {
		return nil, ErrFormat
	}
	return objects, nil
}

func (manager *FsimManager) iterationObjects() ([]FsimObject, error) {
	if len(manager.compressed) == 0 {
		return manager.records, nil
	}
	return DecodeFsimObjects(manager.compressed, len(manager.records))
}

// Verify implements Fsim_ManVerifyFront and validates the compressed stream.
func (manager *FsimManager) Verify() error {
	if manager == nil || manager.frontier < 2 || len(manager.records) != len(manager.fans0)-2 {
		if manager == nil {
			return fmt.Errorf("%w: nil FSIM manager", ErrInvariant)
		}
		return fmt.Errorf("%w: frontier=%d records=%d fans=%d", ErrInvariant, manager.frontier, len(manager.records), len(manager.fans0))
	}
	objects, err := manager.iterationObjects()
	if err != nil {
		return err
	}
	frontToID := make([]int, manager.frontier)
	if manager.constantFront != 0 {
		frontToID[manager.constantFront%manager.frontier] = 1
	}
	for index, object := range objects {
		id := index + 2
		if object.Node != 0 {
			frontToID[object.Node%manager.frontier] = id
		}
		left, right := fsimLit2Var(object.Fanin0), fsimLit2Var(object.Fanin1)
		got0 := fsimVar2Lit(frontToID[left%manager.frontier], fsimLitIsCompl(object.Fanin0))
		got1 := fsimVar2Lit(frontToID[right%manager.frontier], fsimLitIsCompl(object.Fanin1))
		if got0 != manager.fans0[id] || got1 != manager.fans1[id] {
			return fmt.Errorf("%w: frontier object %d", ErrInvariant, id)
		}
	}
	return nil
}

// Statistics returns an immutable manager summary.
func (manager *FsimManager) Statistics() FsimManagerStatistics {
	if manager == nil {
		return FsimManagerStatistics{}
	}
	return FsimManagerStatistics{
		PrimaryInputs: manager.nPis, PrimaryOutputs: manager.nPos, Inputs: manager.nCis,
		Outputs: manager.nCos, Nodes: manager.nodes, Objects: len(manager.fans0),
		MaximumCrossCut: manager.maximumCrossCut, Frontier: manager.frontier,
		MaximumFrontierID: manager.maximumID, CompressedBytes: len(manager.compressed),
	}
}

// Objects returns an owned frontier record snapshot.
func (manager *FsimManager) Objects() []FsimObject {
	if manager == nil {
		return nil
	}
	return slices.Clone(manager.records)
}

// CompressedData returns an owned compressed stream snapshot.
func (manager *FsimManager) CompressedData() []byte {
	if manager == nil {
		return nil
	}
	return slices.Clone(manager.compressed)
}

// Close implements the observable ownership effect of Fsim_ManDelete. It is
// idempotent; subsequent use fails validation rather than retaining storage.
func (manager *FsimManager) Close() {
	if manager == nil {
		return
	}
	*manager = FsimManager{}
}

type fsimRandom struct {
	native    ABCRandom
	state     uint64
	useNative bool
}

func newFsimRandom(seed uint64) fsimRandom {
	if seed == 0 {
		// The source resets Aig_ManRandom before every campaign. ABCRandom is
		// the owned, exact implementation of that generator; keeping it as the
		// default preserves native patterns without process-global state.
		return fsimRandom{useNative: true}
	}
	return fsimRandom{state: seed}
}
func (random *fsimRandom) word() uint32 {
	if random.useNative {
		return random.native.Uint32(false)
	}
	x := random.state
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	random.state = x
	return uint32(x * 2685821657736338717)
}

func normalizeFsimRunOptions(options FsimRunOptions) (FsimRunOptions, error) {
	if options.MaxWords < 0 || options.MaxFrames < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxWords == 0 {
		options.MaxWords = 1 << 16
	}
	if options.MaxFrames == 0 {
		options.MaxFrames = 1 << 20
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 1 << 30
	}
	return options, nil
}

func fsimCheckedWords(rows, words, maximum int) error {
	if rows < 0 || words < 0 || maximum < 0 {
		return ErrDimension
	}
	if rows != 0 && words > maximum/4/rows {
		return ErrLimit
	}
	return nil
}

func fsimRead(data []uint32, words, row int) []uint32 {
	return data[row*words : (row+1)*words]
}

// Simulate implements Fsim_ManSimulate over this manager. The source's global
// RNG and graph-attached witness are replaced by an explicit seed and result.
// A zero time limit means no deadline instead of the source's accidental
// immediate timeout; positive limits retain the source's per-frame polling.
func (manager *FsimManager) Simulate(parameters FsimSimulationParams, options FsimRunOptions) (FsimResult, error) {
	options, err := normalizeFsimRunOptions(options)
	if err != nil {
		return FsimResult{}, err
	}
	if manager == nil || manager.graph == nil {
		return FsimResult{}, ErrInvariant
	}
	if parameters.Words <= 0 || parameters.Iterations < 0 || parameters.TimeLimit < 0 {
		return FsimResult{}, ErrDimension
	}
	if parameters.Words > options.MaxWords || parameters.Iterations > options.MaxFrames {
		return FsimResult{}, ErrLimit
	}
	rows := manager.frontier + manager.nCis + manager.nCos
	if err = fsimCheckedWords(rows, parameters.Words, options.MaxBytes); err != nil {
		return FsimResult{}, err
	}
	historyWords := manager.nPis * parameters.Words
	if parameters.Iterations != 0 && historyWords > options.MaxBytes/4/parameters.Iterations {
		return FsimResult{}, ErrLimit
	}
	objects, err := manager.iterationObjects()
	if err != nil {
		return FsimResult{}, err
	}
	front := make([]uint32, manager.frontier*parameters.Words)
	cis := make([]uint32, manager.nCis*parameters.Words)
	cos := make([]uint32, manager.nCos*parameters.Words)
	history := make([]uint32, 0, historyWords*parameters.Iterations)
	random := newFsimRandom(options.Seed)
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	fillInputs := func(initial bool) error {
		for ci, pio := range manager.cisToPIO {
			if stopped() {
				return ErrCancelled
			}
			row := fsimRead(cis, parameters.Words, ci)
			if pio < manager.nPis {
				for word := parameters.Words - 1; word >= 0; word-- {
					if stopped() {
						return ErrCancelled
					}
					row[word] = random.word()
				}
			} else if initial {
				clear(row)
			} else {
				copy(row, fsimRead(cos, parameters.Words, manager.nPos+pio-manager.nPis))
			}
		}
		return nil
	}
	recordHistory := func() error {
		frame := make([]uint32, historyWords)
		for ci, pio := range manager.cisToPIO {
			if stopped() {
				return ErrCancelled
			}
			if pio < manager.nPis {
				copy(fsimRead(frame, parameters.Words, pio), fsimRead(cis, parameters.Words, ci))
			}
		}
		history = append(history, frame...)
		return nil
	}
	if err = fillInputs(true); err != nil {
		return FsimResult{}, err
	}
	var deadline time.Time
	if parameters.TimeLimit > 0 {
		deadline = time.Now().Add(time.Duration(parameters.TimeLimit) * time.Second)
	}
	result := FsimResult{Patterns: parameters.Words * 32, Property: -1, Pattern: -1}
	for frame := 0; frame < parameters.Iterations; frame++ {
		if stopped() {
			return FsimResult{}, ErrCancelled
		}
		if err = recordHistory(); err != nil {
			return FsimResult{}, err
		}
		if manager.constantFront != 0 {
			row := fsimRead(front, parameters.Words, manager.constantFront%manager.frontier)
			for word := range row {
				if stopped() {
					return FsimResult{}, ErrCancelled
				}
				row[word] = math.MaxUint32
			}
		}
		ciIndex, coIndex := 0, 0
		for _, object := range objects {
			if stopped() {
				return FsimResult{}, ErrCancelled
			}
			switch {
			case object.Fanin0 == 0:
				copy(fsimRead(front, parameters.Words, object.Node%manager.frontier), fsimRead(cis, parameters.Words, ciIndex))
				ciIndex++
			case object.Fanin1 == 0:
				destination := fsimRead(cos, parameters.Words, coIndex)
				source := fsimRead(front, parameters.Words, fsimLit2Var(object.Fanin0)%manager.frontier)
				for word := range destination {
					if stopped() {
						return FsimResult{}, ErrCancelled
					}
					if fsimLitIsCompl(object.Fanin0) {
						destination[word] = ^source[word]
					} else {
						destination[word] = source[word]
					}
				}
				coIndex++
			default:
				destination := fsimRead(front, parameters.Words, object.Node%manager.frontier)
				left := fsimRead(front, parameters.Words, fsimLit2Var(object.Fanin0)%manager.frontier)
				right := fsimRead(front, parameters.Words, fsimLit2Var(object.Fanin1)%manager.frontier)
				for word := range destination {
					if stopped() {
						return FsimResult{}, ErrCancelled
					}
					leftWord, rightWord := left[word], right[word]
					if fsimLitIsCompl(object.Fanin0) {
						leftWord = ^leftWord
					}
					if fsimLitIsCompl(object.Fanin1) {
						rightWord = ^rightWord
					}
					destination[word] = leftWord & rightWord
				}
			}
		}
		if ciIndex != manager.nCis || coIndex != manager.nCos {
			return FsimResult{}, ErrInvariant
		}
		result.Frames = frame + 1
		if parameters.CheckMiter {
			found := false
			for output := 0; output < manager.nPos && !found; output++ {
				if stopped() {
					return FsimResult{}, ErrCancelled
				}
				row := fsimRead(cos, parameters.Words, output)
				for word := parameters.Words - 1; word >= 0; word-- {
					if stopped() {
						return FsimResult{}, ErrCancelled
					}
					if row[word] != 0 {
						result.Property = output
						result.Pattern = word*32 + bits.TrailingZeros32(row[word])
						found = true
						break
					}
				}
			}
			if found {
				counterexample, cexErr := NewCounterexample(manager.graph.registers, manager.nPis, frame+1)
				if cexErr != nil {
					return FsimResult{}, cexErr
				}
				counterexample.Property = result.Property
				for oldFrame := 0; oldFrame <= frame; oldFrame++ {
					if stopped() {
						return FsimResult{}, ErrCancelled
					}
					frameWords := history[oldFrame*historyWords : (oldFrame+1)*historyWords]
					for input := 0; input < manager.nPis; input++ {
						if stopped() {
							return FsimResult{}, ErrCancelled
						}
						row := fsimRead(frameWords, parameters.Words, input)
						value := row[result.Pattern/32]&(uint32(1)<<uint(result.Pattern%32)) != 0
						_ = counterexample.SetInput(oldFrame, input, value)
					}
				}
				result.Counterexample = counterexample
				return result, nil
			}
		}
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			result.TimedOut = true
			break
		}
		if frame+1 < parameters.Iterations {
			if err = fillInputs(false); err != nil {
				return FsimResult{}, err
			}
		}
	}
	return result, nil
}

// FsimSimulate is the owned top-level replacement for Fsim_ManSimulate.
func (graph *AigManager) FsimSimulate(parameters FsimSimulationParams, build FsimBuildOptions, run FsimRunOptions) (FsimResult, error) {
	// fCompressAig belongs to the native simulation parameter record. The
	// explicit build option may additionally request compression, but must not
	// silently disable a request carried by the source-shaped parameters.
	build.Compress = build.Compress || parameters.CompressAIG
	manager, err := NewFsimManager(graph, build)
	if err != nil {
		return FsimResult{}, err
	}
	defer manager.Close()
	return manager.Simulate(parameters, run)
}
