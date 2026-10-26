package abc

// Derived from ABC src/base/abc/abcBarBuf.c, abcLib.c, abcUtil.c,
// abcShow.c, and module.make.

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// ABCUtilityOptions bounds utility work and returned text.
type ABCUtilityOptions struct {
	MaxOperations int
	MaxTextBytes  int
	Stop          func() bool
}

func (o ABCUtilityOptions) normalized() ABCUtilityOptions {
	if o.MaxOperations <= 0 {
		o.MaxOperations = 1 << 24
	}
	if o.MaxTextBytes <= 0 {
		o.MaxTextBytes = 1 << 24
	}
	return o
}

func abcUtilityTick(o ABCUtilityOptions, work *int, amount int) error {
	if amount < 0 || *work > o.MaxOperations-amount {
		return ErrLimit
	}
	*work += amount
	if o.Stop != nil && o.Stop() {
		return ErrCancelled
	}
	return nil
}

// ABCNetworkStats is the owned counterpart of the frequently used statistics
// and count helpers in abcUtil.c.
type ABCNetworkStats struct {
	Objects, PIs, POs, CIs, COs, Nodes, Nets, Latches, Boxes int
	Edges, Cubes, Literals, Buffers, MultiReferenced         int
	FaninMaximum, FanoutMaximum, TotalFanins                 int
	LevelMaximum, CrossCutMaximum                            int
}

func abcUtilitySOP(o ABCNetworkObject) (ABCSOP, bool) {
	if o.Function == "" || o.Type != ABCObjectNode {
		return ABCSOP{}, false
	}
	sop, err := ParseABCSOP(o.Function, ABCMetadataOptions{})
	return sop, err == nil
}

// ABCComputeNetworkStats computes deterministic structural and SOP counts.
func ABCComputeNetworkStats(network *ABCNetwork, options ABCUtilityOptions) (ABCNetworkStats, error) {
	options = options.normalized()
	if network == nil {
		return ABCNetworkStats{}, ErrDimension
	}
	if err := network.Validate(); err != nil {
		return ABCNetworkStats{}, err
	}
	stats := ABCNetworkStats{PIs: len(network.pis), POs: len(network.pos), CIs: len(network.cis), COs: len(network.cos), Boxes: len(network.boxes), Edges: network.edges}
	work := 0
	for _, object := range network.objects {
		if !object.Alive {
			continue
		}
		if err := abcUtilityTick(options, &work, len(object.Fanins)+len(object.Fanouts)+1); err != nil {
			return ABCNetworkStats{}, err
		}
		stats.Objects++
		if object.Type == ABCObjectNode {
			stats.Nodes++
			stats.TotalFanins += len(object.Fanins)
			stats.FaninMaximum = max(stats.FaninMaximum, len(object.Fanins))
			stats.FanoutMaximum = max(stats.FanoutMaximum, len(object.Fanouts))
			if len(object.Fanouts) > 1 {
				stats.MultiReferenced++
			}
			if len(object.Fanins) == 1 {
				stats.Buffers++
			}
		}
		if object.Type == ABCObjectNet {
			stats.Nets++
		}
		if object.Type == ABCObjectLatch {
			stats.Latches++
		}
		if sop, ok := abcUtilitySOP(object); ok {
			stats.Cubes += len(sop.Cubes)
			stats.Literals += sop.LiteralCount()
		}
	}
	remaining := options.MaxOperations - work
	if remaining <= 0 {
		return ABCNetworkStats{}, ErrLimit
	}
	levels, maximum, err := ABCNetworkLevels(network, ABCAnalysisOptions{MaxVisited: remaining, Stop: options.Stop})
	if err != nil {
		return ABCNetworkStats{}, err
	}
	_ = levels
	stats.LevelMaximum = maximum
	stats.CrossCutMaximum, err = ABCCrossCut(network, options)
	if err != nil {
		return ABCNetworkStats{}, err
	}
	return stats, nil
}

type abcCrossCutFrame struct {
	id        int
	next      int
	decrement int
	entered   bool
}

// ABCCrossCut returns the source Abc_NtkCrossCut value. It preserves the
// source's CO-by-CO recursive event order while using an iterative stack.
// Repeated fanout edges are counted with multiplicity; CIs are not in the cut.
func ABCCrossCut(network *ABCNetwork, options ABCUtilityOptions) (int, error) {
	options = options.normalized()
	if network == nil {
		return 0, ErrDimension
	}
	if err := network.Validate(); err != nil {
		return 0, err
	}
	// ABC normally sweeps dangling logic before this calculation. Restricting
	// reference counts to CO cones gives the same result on swept networks and
	// keeps retained-but-unobservable nodes from corrupting the cut balance.
	active := make([]bool, len(network.objects))
	pending := slices.Clone(network.cos)
	for len(pending) != 0 {
		id := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if active[id] {
			continue
		}
		active[id] = true
		for _, edge := range network.objects[id].Fanins {
			pending = append(pending, edge.ID)
		}
	}
	activeFanouts := make([]int, len(network.objects))
	for id := range network.objects {
		if !active[id] {
			continue
		}
		for _, fanout := range network.objects[id].Fanouts {
			if active[fanout] {
				activeFanouts[id]++
			}
		}
	}
	visited := make([]bool, len(network.objects))
	visits := make([]int, len(network.objects))
	current, peak, work := 0, 0, 0
	for _, root := range network.cos {
		stack := []abcCrossCutFrame{{id: root}}
		returned, returnValue := false, false
		for len(stack) != 0 {
			frame := &stack[len(stack)-1]
			if returned {
				if returnValue {
					frame.decrement++
				}
				returned = false
			}
			object := &network.objects[frame.id]
			if !frame.entered {
				if err := abcUtilityTick(options, &work, 1); err != nil {
					return 0, err
				}
				if abcIsCI(object.Type) {
					stack = stack[:len(stack)-1]
					returned, returnValue = true, false
					continue
				}
				if visited[frame.id] {
					visits[frame.id]++
					returnValue = visits[frame.id] == activeFanouts[frame.id]
					stack = stack[:len(stack)-1]
					returned = true
					continue
				}
				visited[frame.id] = true
				frame.entered = true
			}
			if frame.next < len(object.Fanins) {
				child := object.Fanins[frame.next].ID
				frame.next++
				stack = append(stack, abcCrossCutFrame{id: child})
				continue
			}
			current++
			peak = max(peak, current)
			current -= frame.decrement
			if current < 0 {
				return 0, ErrInvariant
			}
			visits[frame.id]++
			returnValue = visits[frame.id] == activeFanouts[frame.id]
			stack = stack[:len(stack)-1]
			returned = true
		}
		current-- // The source removes each CO after visiting its cone.
		if current < 0 {
			return 0, ErrInvariant
		}
	}
	if current != 0 {
		return 0, ErrInvariant
	}
	return peak, nil
}

// ABCFanoutCounts returns source-ID-indexed fanout multiplicities.
func ABCFanoutCounts(network *ABCNetwork) ([]int, error) {
	if network == nil {
		return nil, ErrDimension
	}
	if err := network.Validate(); err != nil {
		return nil, err
	}
	result := make([]int, len(network.objects))
	for _, object := range network.objects {
		if object.Alive {
			result[object.ID] = len(object.Fanouts)
		}
	}
	return result, nil
}

// ABCCollectObjects returns live IDs in stable source-object order.
func ABCCollectObjects(network *ABCNetwork) ([]int, error) {
	if network == nil {
		return nil, ErrDimension
	}
	if err := network.Validate(); err != nil {
		return nil, err
	}
	result := make([]int, 0, network.ObjectCount())
	for _, object := range network.objects {
		if object.Alive {
			result = append(result, object.ID)
		}
	}
	return result, nil
}

// ABCCIIDs returns CI object IDs in source terminal order.
func ABCCIIDs(network *ABCNetwork) ([]int, error) {
	if network == nil {
		return nil, ErrDimension
	}
	if err := network.Validate(); err != nil {
		return nil, err
	}
	return slices.Clone(network.cis), nil
}

// ABCOrderCIsCOs restores the source PI/BO and PO/BI partitions.
func ABCOrderCIsCOs(network *ABCNetwork) error {
	if network == nil {
		return ErrDimension
	}
	backup := network.Clone()
	pis, bos, pos, bis := make([]int, 0), make([]int, 0), make([]int, 0), make([]int, 0)
	for _, object := range network.objects {
		if !object.Alive {
			continue
		}
		switch object.Type {
		case ABCObjectPI:
			pis = append(pis, object.ID)
		case ABCObjectBO:
			bos = append(bos, object.ID)
		case ABCObjectPO:
			pos = append(pos, object.ID)
		case ABCObjectBI:
			bis = append(bis, object.ID)
		}
	}
	network.pis, network.pos = pis, pos
	network.cis = append(slices.Clone(pis), bos...)
	network.cos = append(slices.Clone(pos), bis...)
	if err := network.Validate(); err != nil {
		network.restore(backup)
		return err
	}
	return nil
}

// ABCBarrierRequest identifies an existing edge to isolate with a barrier
// buffer. The request is stable-ID based and supports parallel edges by taking
// the first remaining occurrence, as InsertBetween does.
type ABCBarrierRequest struct{ Fanin, Fanout int }

// ABCInsertBarrierBuffers inserts ordinary buffer nodes transactionally.
func ABCInsertBarrierBuffers(network *ABCNetwork, requests []ABCBarrierRequest, options ABCUtilityOptions) ([]int, error) {
	options = options.normalized()
	if network == nil {
		return nil, ErrDimension
	}
	backup := network.Clone()
	result := make([]int, 0, len(requests))
	work := 0
	for _, request := range requests {
		if err := abcUtilityTick(options, &work, 1); err != nil {
			network.restore(backup)
			return nil, err
		}
		id, err := network.InsertBetween(request.Fanin, request.Fanout, ABCObjectNode)
		if err == nil {
			err = network.SetNodeFunction(id, "1 1\n")
		}
		if err != nil {
			network.restore(backup)
			return nil, err
		}
		result = append(result, id)
	}
	if err := abcUtilityTick(options, &work, 0); err != nil {
		network.restore(backup)
		return nil, err
	}
	if err := network.Validate(); err != nil {
		network.restore(backup)
		return nil, err
	}
	return result, nil
}

// ABCRemoveBarrierBuffers splices exact one-input buffer nodes and is atomic.
func ABCRemoveBarrierBuffers(network *ABCNetwork, buffers []int, options ABCUtilityOptions) error {
	options = options.normalized()
	if network == nil {
		return ErrDimension
	}
	backup := network.Clone()
	work := 0
	seen := make(map[int]bool, len(buffers))
	for _, id := range buffers {
		if err := abcUtilityTick(options, &work, 1); err != nil {
			network.restore(backup)
			return err
		}
		if seen[id] {
			network.restore(backup)
			return ErrDimension
		}
		seen[id] = true
		object, err := network.Object(id)
		if err != nil || object.Type != ABCObjectNode || len(object.Fanins) != 1 {
			network.restore(backup)
			return ErrDimension
		}
		sop, parseErr := ParseABCSOP(object.Function, ABCMetadataOptions{})
		if parseErr != nil || !sop.IsBuffer() {
			network.restore(backup)
			return ErrDimension
		}
		for len(network.objects[id].Fanouts) > 0 {
			if err := abcUtilityTick(options, &work, 1); err != nil {
				network.restore(backup)
				return err
			}
			fanout := network.objects[id].Fanouts[0]
			if err := network.PatchFanin(fanout, id, object.Fanins[0]); err != nil {
				network.restore(backup)
				return err
			}
		}
		if err := network.DeleteObject(id); err != nil {
			network.restore(backup)
			return err
		}
	}
	if err := abcUtilityTick(options, &work, 0); err != nil {
		network.restore(backup)
		return err
	}
	if err := network.Validate(); err != nil {
		network.restore(backup)
		return err
	}
	return nil
}

// ABCDesign is an owned, deterministic counterpart of Abc_Des_t. Dependencies
// name instantiated models because the current concrete network deliberately
// has no generic pData pointer.
type ABCDesign struct {
	Name         string
	modules      []*ABCNetwork
	dependencies [][]string
	byName       map[string]int
	closed       bool
}

func NewABCDesign(name string) (*ABCDesign, error) {
	if name == "" || strings.IndexByte(name, 0) >= 0 {
		return nil, ErrDimension
	}
	return &ABCDesign{Name: name, byName: make(map[string]int)}, nil
}

func (d *ABCDesign) check() error {
	if d == nil || d.closed || d.byName == nil || len(d.modules) != len(d.dependencies) {
		return ErrInvariant
	}
	for i, module := range d.modules {
		if module == nil || module.closed || module.Name == "" {
			return ErrInvariant
		}
		index, present := d.byName[module.Name]
		if !present || index != i {
			return ErrInvariant
		}
	}
	return nil
}

func (d *ABCDesign) AddModel(network *ABCNetwork, dependencies []string) error {
	if err := d.check(); err != nil {
		return err
	}
	if network == nil || network.Name == "" || strings.IndexByte(network.Name, 0) >= 0 {
		return ErrDimension
	}
	if _, exists := d.byName[network.Name]; exists {
		return ErrDimension
	}
	if err := network.Validate(); err != nil {
		return err
	}
	deps := slices.Clone(dependencies)
	for _, name := range deps {
		if name == "" || strings.IndexByte(name, 0) >= 0 {
			return ErrDimension
		}
	}
	copyNetwork := network.Clone()
	d.byName[copyNetwork.Name] = len(d.modules)
	d.modules = append(d.modules, copyNetwork)
	d.dependencies = append(d.dependencies, deps)
	return nil
}

func (d *ABCDesign) FindModel(name string) (*ABCNetwork, bool) {
	if d == nil || d.closed {
		return nil, false
	}
	index, ok := d.byName[name]
	if !ok {
		return nil, false
	}
	return d.modules[index].Clone(), true
}

func (d *ABCDesign) TopModels() ([]string, error) {
	if err := d.check(); err != nil {
		return nil, err
	}
	referenced := make(map[string]bool)
	for _, deps := range d.dependencies {
		for _, name := range deps {
			if _, ok := d.byName[name]; !ok {
				return nil, ErrDimension
			}
			referenced[name] = true
		}
	}
	result := make([]string, 0, len(d.modules))
	for _, module := range d.modules {
		if !referenced[module.Name] {
			result = append(result, module.Name)
		}
	}
	return result, nil
}

func (d *ABCDesign) DeriveRoot() (*ABCNetwork, error) {
	if err := d.check(); err != nil {
		return nil, err
	}
	if len(d.modules) != 1 {
		return nil, ErrDimension
	}
	return d.modules[0].Clone(), nil
}

func (d *ABCDesign) Clone() (*ABCDesign, error) {
	if err := d.check(); err != nil {
		return nil, err
	}
	clone, _ := NewABCDesign(d.Name)
	for i, module := range d.modules {
		if err := clone.AddModel(module, d.dependencies[i]); err != nil {
			return nil, err
		}
	}
	return clone, nil
}

func (d *ABCDesign) Close() {
	if d == nil || d.closed {
		return
	}
	for _, module := range d.modules {
		module.Close()
	}
	d.modules, d.dependencies, d.byName, d.closed = nil, nil, nil, true
}

func (d *ABCDesign) Summary(options ABCUtilityOptions) (string, error) {
	options = options.normalized()
	if err := d.check(); err != nil {
		return "", err
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "design %s models %d\n", d.Name, len(d.modules))
	if builder.Len() > options.MaxTextBytes {
		return "", ErrLimit
	}
	for i, module := range d.modules {
		stats, err := ABCComputeNetworkStats(module, options)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&builder, "%d %s objects=%d nodes=%d latches=%d deps=%s\n", i+1, module.Name, stats.Objects, stats.Nodes, stats.Latches, strings.Join(d.dependencies[i], ","))
		if builder.Len() > options.MaxTextBytes {
			return "", ErrLimit
		}
	}
	return builder.String(), nil
}

func abcDOTQuote(text string) string {
	return strconv.Quote(text)
}

// ABCDOT returns deterministic Graphviz bytes instead of using fixed files or
// spawning a viewer.
func ABCDOT(network *ABCNetwork, includeFunctions bool, options ABCUtilityOptions) (string, error) {
	options = options.normalized()
	if network == nil {
		return "", ErrDimension
	}
	if err := network.Validate(); err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString("digraph G {\n  rankdir=LR;\n")
	work := 0
	for _, object := range network.objects {
		if !object.Alive {
			continue
		}
		label := object.Name
		if label == "" {
			label = fmt.Sprintf("%d", object.ID)
		}
		if includeFunctions && object.Function != "" {
			label += "\\n" + strings.TrimSpace(object.Function)
		}
		fmt.Fprintf(&builder, "  n%d [label=%s];\n", object.ID, abcDOTQuote(label))
		for _, edge := range object.Fanins {
			style := "solid"
			if edge.Complemented {
				style = "dashed"
			}
			fmt.Fprintf(&builder, "  n%d -> n%d [style=%s];\n", edge.ID, object.ID, style)
		}
		if builder.Len() > options.MaxTextBytes {
			return "", ErrLimit
		}
		if err := abcUtilityTick(options, &work, len(object.Fanins)+1); err != nil {
			return "", err
		}
	}
	builder.WriteString("}\n")
	if builder.Len() > options.MaxTextBytes {
		return "", ErrLimit
	}
	return builder.String(), nil
}

// ABCDOTFilename applies the source filename sanitization without I/O.
func ABCDOTFilename(name string) (string, error) {
	if name == "" || strings.IndexByte(name, 0) >= 0 {
		return "", ErrDimension
	}
	var builder strings.Builder
	for _, r := range name + ".dot" {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '.' {
			builder.WriteRune(r)
		} else {
			builder.WriteByte('_')
		}
	}
	return builder.String(), nil
}

// ABCFlopDependencyDOT returns the latch dependency graph as owned text.
func ABCFlopDependencyDOT(network *ABCNetwork, options ABCUtilityOptions) (string, error) {
	options = options.normalized()
	triples, err := ABCLatchTriples(network)
	if err != nil {
		return "", err
	}
	boToLatch := make(map[int]int, len(triples))
	for i, triple := range triples {
		boToLatch[triple.Output] = i
	}
	var builder strings.Builder
	builder.WriteString("digraph G {\n  edge [dir=forward];\n")
	work := 0
	for i, triple := range triples {
		if work >= options.MaxOperations {
			return "", ErrLimit
		}
		support, supportErr := ABCNetworkSupport(network, []int{triple.Input}, ABCAnalysisOptions{MaxVisited: options.MaxOperations - work, Stop: options.Stop})
		if supportErr != nil {
			return "", supportErr
		}
		piCount := 0
		deps := make([]int, 0)
		for _, id := range support {
			if network.objects[id].Type == ABCObjectPI {
				piCount++
			} else if dependency, ok := boToLatch[id]; ok {
				deps = append(deps, dependency)
			}
		}
		sort.Ints(deps)
		fmt.Fprintf(&builder, "  %d [label=%s];\n", i, abcDOTQuote(fmt.Sprintf("%d(%d)", i, piCount)))
		for _, dependency := range deps {
			fmt.Fprintf(&builder, "  %d -> %d;\n", dependency, i)
		}
		if builder.Len() > options.MaxTextBytes {
			return "", ErrLimit
		}
		if err := abcUtilityTick(options, &work, len(support)+1); err != nil {
			return "", err
		}
	}
	builder.WriteString("}\n")
	if builder.Len() > options.MaxTextBytes {
		return "", ErrLimit
	}
	return builder.String(), nil
}

// ABCNetworkReport returns deterministic structural statistics.
func ABCNetworkReport(network *ABCNetwork, options ABCUtilityOptions) (string, error) {
	stats, err := ABCComputeNetworkStats(network, options)
	if err != nil {
		return "", err
	}
	result := fmt.Sprintf("name=%s objects=%d pi=%d po=%d node=%d net=%d latch=%d edge=%d cube=%d lit=%d buf=%d level=%d crosscut=%d\n",
		network.Name, stats.Objects, stats.PIs, stats.POs, stats.Nodes, stats.Nets, stats.Latches, stats.Edges, stats.Cubes, stats.Literals, stats.Buffers, stats.LevelMaximum, stats.CrossCutMaximum)
	if len(result) > options.normalized().MaxTextBytes {
		return "", ErrLimit
	}
	return result, nil
}

// ABCBaseModuleSources is the exact source inventory in base/abc/module.make.
func ABCBaseModuleSources() []string {
	return []string{
		"abcAig.c", "abcBarBuf.c", "abcBlifMv.c", "abcCheck.c", "abcDfs.c",
		"abcFanio.c", "abcFanOrder.c", "abcFunc.c", "abcHie.c", "abcHieCec.c",
		"abcHieGia.c", "abcHieNew.c", "abcLatch.c", "abcLib.c", "abcMinBase.c",
		"abcNames.c", "abcNetlist.c", "abcNtk.c", "abcObj.c", "abcRefs.c",
		"abcShow.c", "abcSop.c", "abcUtil.c",
	}
}
