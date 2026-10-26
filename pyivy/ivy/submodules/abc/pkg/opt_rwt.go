package abc

// DAG-aware rewriting library from src/opt/rwt/rwt.h, rwtDec.c,
// rwtMan.c, rwtUtil.c, and module.make.

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"slices"
	"time"
)

const (
	rwtFunctionCount    = 1 << 16
	rwtClassCount       = 222
	rwtDefaultNodeLimit = 1 << 18
)

// RwtOptions bounds construction. Precompute preserves the source mode, which
// creates only the five elementary nodes because its precomputation call is
// disabled in rwtMan.c.
type RwtOptions struct {
	Precompute bool
	MaxNodes   int
	MaxClasses int
	Stop       func() bool
}

// RwtLink is the portable form of a complemented Rwt_Node_t pointer.
type RwtLink struct {
	Node       int
	Complement bool
}

// RwtNode is an owned snapshot of one rewriting-library node.
type RwtNode struct {
	ID, TraversalID int
	Truth           uint16
	Volume, Level   int
	Used, XOR       bool
	Fan0, Fan1      RwtLink
}

// RwtStats contains both library and client-populated rewriting counters.
type RwtStats struct {
	Considered, Added, Classes                      int
	NodesConsidered, NodesRewritten, NodesGained    int
	CutsGood, CutsBad, Subgraphs                    int
	Scores                                          [rwtClassCount]int
	Start, TruthTime, CutTime, ResynthesisTime      time.Duration
	EvaluationTime, MFFCTime, UpdateTime, TotalTime time.Duration
}

// RwtManager owns the exact static four-input rewriting forest.
type RwtManager struct {
	options                        RwtOptions
	canonical                      []uint16
	phases, permutations, classMap []uint8
	practical                      []bool
	nodes                          []RwtNode
	table                          [][]int
	classes                        [][]int
	mapInverse                     []uint16
	graphs                         []*DecoGraph
	traversalID                    int
	stats                          RwtStats
	graph                          *DecoGraph
	leaves                         []int
	complement                     bool
	closed                         bool
}

func normalizeRwtOptions(options RwtOptions) (RwtOptions, error) {
	if options.MaxNodes < 0 || options.MaxClasses < 0 {
		return options, ErrDimension
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = rwtDefaultNodeLimit
	}
	if options.MaxClasses == 0 {
		options.MaxClasses = rwtClassCount
	}
	if options.MaxClasses < rwtClassCount {
		return options, ErrLimit
	}
	return options, nil
}

func (manager *RwtManager) stopped() bool {
	return manager.options.Stop != nil && manager.options.Stop()
}
func (manager *RwtManager) check() error {
	if manager == nil || manager.closed {
		return ErrInvariant
	}
	return nil
}

// NewRwtManager implements Rwt_ManStart.
func NewRwtManager(options RwtOptions) (*RwtManager, error) {
	options, err := normalizeRwtOptions(options)
	if err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	start := time.Now()
	dec, err := NewDecManager()
	if err != nil {
		return nil, err
	}
	canonical, phases, permutations, classMap, err := dec.NPN4Tables()
	closeErr := dec.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	m := &RwtManager{options: options, canonical: canonical, phases: phases, permutations: permutations,
		classMap: classMap, practical: make([]bool, rwtFunctionCount), table: make([][]int, rwtFunctionCount), traversalID: 1}
	for _, truth := range rwtPracticalClasses {
		m.practical[truth] = true
	}
	for _, truth := range []uint16{0, 0xaaaa, 0xcccc, 0xf0f0, 0xff00} {
		if _, err = m.addVariable(truth, options.Precompute); err != nil {
			m.Close()
			return nil, err
		}
	}
	m.stats.Classes = 5
	if !options.Precompute {
		if err = m.loadStatic(); err == nil {
			err = m.preprocess()
		}
		if err != nil {
			m.Close()
			return nil, err
		}
	}
	m.stats.Start = time.Since(start)
	return m, nil
}

// RwtPrecompute implements Rwt_Precompute. As in the source revision, the
// actual enumeration call is disabled; this validates the precompute-mode
// manager lifecycle and then releases it.
func RwtPrecompute(options RwtOptions) error {
	options.Precompute = true
	manager, err := NewRwtManager(options)
	if err != nil {
		return err
	}
	return manager.Close()
}

func (manager *RwtManager) addVariable(truth uint16, precompute bool) (int, error) {
	if manager.stopped() {
		return -1, ErrCancelled
	}
	if len(manager.nodes) >= manager.options.MaxNodes {
		return -1, ErrLimit
	}
	id := len(manager.nodes)
	manager.nodes = append(manager.nodes, RwtNode{ID: id, Truth: truth, Used: true, Fan0: RwtLink{Node: -1}, Fan1: RwtLink{Node: -1}})
	if precompute {
		manager.table[truth] = append(manager.table[truth], id)
	}
	return id, nil
}

// AddVariable implements Rwt_ManAddVar and atomically adds an elementary node.
func (manager *RwtManager) AddVariable(truth uint16, precompute bool) (int, error) {
	if err := manager.check(); err != nil {
		return -1, err
	}
	if manager.classes != nil || manager.graphs != nil {
		return -1, ErrInvariant
	}
	return manager.addVariable(truth, precompute)
}

func (manager *RwtManager) truthOf(link RwtLink) (uint16, error) {
	if link.Node < 0 || link.Node >= len(manager.nodes) {
		return 0, ErrDimension
	}
	truth := manager.nodes[link.Node].Truth
	if link.Complement {
		truth = ^truth
	}
	return truth, nil
}

// AddNode implements Rwt_ManAddNode with source truth, cost, and list order.
func (manager *RwtManager) AddNode(first, second RwtLink, xor bool, level, volume int) (int, error) {
	if err := manager.check(); err != nil {
		return -1, err
	}
	if manager.classes != nil || manager.graphs != nil {
		return -1, ErrInvariant
	}
	if level < 0 || level >= 1<<6 || volume < 0 || volume >= 1<<8 {
		return -1, ErrDimension
	}
	if manager.stopped() {
		return -1, ErrCancelled
	}
	if len(manager.nodes) >= manager.options.MaxNodes {
		return -1, ErrLimit
	}
	firstTruth, err := manager.truthOf(first)
	if err != nil {
		return -1, err
	}
	secondTruth, err := manager.truthOf(second)
	if err != nil {
		return -1, err
	}
	truth := firstTruth & secondTruth
	if xor {
		// The source dereferences regular pointers in its XOR case. Static XOR
		// records are regular, so reject an impossible encoded complement.
		if first.Complement || second.Complement {
			return -1, ErrInvariant
		}
		truth = firstTruth ^ secondTruth
	}
	id := len(manager.nodes)
	manager.nodes = append(manager.nodes, RwtNode{ID: id, Truth: truth, Volume: volume, Level: level, XOR: xor, Fan0: first, Fan1: second})
	manager.stats.Considered++
	if truth == manager.canonical[truth] {
		if len(manager.table[truth]) == 0 {
			manager.stats.Classes++
		}
		manager.table[truth] = append(manager.table[truth], id)
		manager.stats.Added++
	}
	return id, nil
}

func (manager *RwtManager) incrementTraversal() {
	manager.traversalID++
	if manager.traversalID < 0x8fffffff {
		return
	}
	for i := range manager.nodes {
		manager.nodes[i].TraversalID = 0
	}
	manager.traversalID = 1
}

// IncrementTraversalID implements Rwt_ManIncTravId.
func (manager *RwtManager) IncrementTraversalID() error {
	if err := manager.check(); err != nil {
		return err
	}
	manager.incrementTraversal()
	return nil
}

func (manager *RwtManager) visitVolume(id int, volume *int) error {
	if id < 0 || id >= len(manager.nodes) {
		return ErrInvariant
	}
	n := &manager.nodes[id]
	if n.Used || n.TraversalID == manager.traversalID {
		return nil
	}
	n.TraversalID = manager.traversalID
	*volume++
	if n.XOR {
		*volume++
	}
	if err := manager.visitVolume(n.Fan0.Node, volume); err != nil {
		return err
	}
	return manager.visitVolume(n.Fan1.Node, volume)
}

// NodeVolume implements Rwt_ManNodeVolume, including shared-DAG accounting.
func (manager *RwtManager) NodeVolume(first, second int) (int, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	manager.incrementTraversal()
	result := 0
	if err := manager.visitVolume(first, &result); err != nil {
		return 0, err
	}
	if err := manager.visitVolume(second, &result); err != nil {
		return 0, err
	}
	return result, nil
}

func decodeRwtStatic() ([]uint16, error) {
	raw, err := base64.StdEncoding.DecodeString(rwtStaticGzip)
	if err != nil {
		return nil, ErrInvariant
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, ErrInvariant
	}
	data, err := io.ReadAll(io.LimitReader(z, 1<<20))
	closeErr := z.Close()
	if err != nil || closeErr != nil || len(data)%2 != 0 {
		return nil, ErrInvariant
	}
	result := make([]uint16, len(data)/2)
	for i := range result {
		result[i] = binary.LittleEndian.Uint16(data[2*i:])
	}
	return result, nil
}

func (manager *RwtManager) loadStatic() error {
	entries, err := decodeRwtStatic()
	if err != nil {
		return err
	}
	if len(entries) < 2 || entries[len(entries)-2] != 0 || entries[len(entries)-1] != 0 {
		return ErrInvariant
	}
	for i := 0; i+1 < len(entries); i += 2 {
		if manager.stopped() {
			return ErrCancelled
		}
		e0, e1 := uint32(entries[i]), uint32(entries[i+1])
		if e0 == 0 && e1 == 0 {
			break
		}
		xor := e0&1 != 0
		e0 >>= 1
		first, second := RwtLink{Node: int(e0 >> 1), Complement: e0&1 != 0}, RwtLink{Node: int(e1 >> 1), Complement: e1&1 != 0}
		if first.Node >= len(manager.nodes) || second.Node >= len(manager.nodes) {
			return ErrInvariant
		}
		level := 1 + max(manager.nodes[first.Node].Level, manager.nodes[second.Node].Level)
		volume, e := manager.NodeVolume(first.Node, second.Node)
		if e != nil {
			return e
		}
		if _, e = manager.AddNode(first, second, xor, level, 1+volume+rwtBoolToInt(xor)); e != nil {
			return e
		}
	}
	return nil
}

// LoadStatic implements Rwt_ManLoadFromArray. A cancellation or limit leaves
// the manager byte-for-byte equivalent at its public snapshot boundary.
func (manager *RwtManager) LoadStatic() error {
	if err := manager.check(); err != nil {
		return err
	}
	// Array indexes are defined relative to exactly the five elementary
	// nodes. Loading after any construction, or after preprocessing, would
	// silently connect the saved forest to the wrong nodes or leave the
	// class/decomposition indexes stale.
	if len(manager.nodes) != 5 || manager.stats.Considered != 0 || manager.classes != nil || manager.graphs != nil {
		return ErrInvariant
	}
	nodes := slices.Clone(manager.nodes)
	table := make([][]int, len(manager.table))
	for i := range table {
		table[i] = slices.Clone(manager.table[i])
	}
	stats, traversal := manager.stats, manager.traversalID
	if err := manager.loadStatic(); err != nil {
		manager.nodes, manager.table, manager.stats, manager.traversalID = nodes, table, stats, traversal
		return err
	}
	return nil
}

func rwtBoolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (manager *RwtManager) collect(id int, graph *DecoGraph) (DecoEdge, error) {
	n := &manager.nodes[id]
	if n.Used {
		return DecoEdge{Node: n.ID - 1}, nil
	}
	if n.TraversalID == manager.traversalID {
		return DecoEdgeFromUint32(uint32(n.Volume))
	}
	n.TraversalID = manager.traversalID
	left, err := manager.collect(n.Fan0.Node, graph)
	if err != nil {
		return DecoEdge{}, err
	}
	if n.Fan0.Complement {
		left = left.Not()
	}
	right, err := manager.collect(n.Fan1.Node, graph)
	if err != nil {
		return DecoEdge{}, err
	}
	if n.Fan1.Complement {
		right = right.Not()
	}
	var edge DecoEdge
	if n.XOR {
		edge, err = graph.AddXOR(left, right, false)
	} else {
		edge, err = graph.AddAND(left, right)
	}
	if err != nil {
		return DecoEdge{}, err
	}
	encoded, err := edge.Uint32()
	if err != nil || encoded >= 1<<8 {
		return DecoEdge{}, ErrLimit
	}
	n.Volume = int(encoded)
	return edge, nil
}

func (manager *RwtManager) nodeGraph(id int) (*DecoGraph, error) {
	n := manager.nodes[id]
	limits := DecoGraphLimits{MaxLeaves: 4, MaxNodes: 128, Stop: manager.options.Stop}
	if n.Truth == 0 {
		return NewConstantDecoGraph(false, limits)
	}
	if n.Truth == 0x00ff {
		return NewLeafDecoGraph(3, 4, true, limits)
	}
	graph, err := NewDecoGraph(4, limits)
	if err != nil {
		return nil, err
	}
	manager.incrementTraversal()
	root, err := manager.collect(id, graph)
	if err != nil {
		graph.Close()
		return nil, err
	}
	if err = graph.SetRoot(root); err != nil {
		graph.Close()
		return nil, err
	}
	return graph, nil
}

func (manager *RwtManager) preprocess() error {
	manager.mapInverse = make([]uint16, rwtClassCount)
	manager.classes = make([][]int, rwtClassCount)
	for truth, list := range manager.table {
		for _, id := range list {
			if manager.nodes[id].Truth != uint16(truth) {
				return ErrInvariant
			}
			class := int(manager.classMap[truth])
			if class >= rwtClassCount {
				return ErrInvariant
			}
			manager.classes[class] = append(manager.classes[class], id)
			manager.mapInverse[class] = manager.canonical[truth]
		}
	}
	manager.graphs = make([]*DecoGraph, len(manager.nodes))
	for _, class := range manager.classes {
		for _, id := range class {
			if manager.stopped() {
				return ErrCancelled
			}
			graph, err := manager.nodeGraph(id)
			if err != nil {
				return err
			}
			manager.graphs[id] = graph
		}
	}
	return nil
}

// Preprocess implements Rwt_ManPreprocess. The retained class and graph
// indexes are committed only after every decomposition succeeds.
func (manager *RwtManager) Preprocess() error {
	if err := manager.check(); err != nil {
		return err
	}
	if manager.classes != nil || manager.graphs != nil {
		return ErrInvariant
	}
	nodes := slices.Clone(manager.nodes)
	if err := manager.preprocess(); err != nil {
		for _, graph := range manager.graphs {
			if graph != nil {
				graph.Close()
			}
		}
		manager.nodes, manager.classes, manager.mapInverse, manager.graphs = nodes, nil, nil, nil
		return err
	}
	return nil
}

// Node returns an owned node value.
func (manager *RwtManager) Node(id int) (RwtNode, error) {
	if err := manager.check(); err != nil {
		return RwtNode{}, err
	}
	if id < 0 || id >= len(manager.nodes) {
		return RwtNode{}, ErrDimension
	}
	return manager.nodes[id], nil
}
func (manager *RwtManager) Nodes() ([]RwtNode, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	return slices.Clone(manager.nodes), nil
}
func (manager *RwtManager) Classes() ([][]int, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	out := make([][]int, len(manager.classes))
	for i := range out {
		out[i] = slices.Clone(manager.classes[i])
	}
	return out, nil
}
func (manager *RwtManager) Canonical(truth uint16) (uint16, uint8, uint8, uint8, error) {
	if err := manager.check(); err != nil {
		return 0, 0, 0, 0, err
	}
	return manager.canonical[truth], manager.phases[truth], manager.permutations[truth], manager.classMap[truth], nil
}
func (manager *RwtManager) Practical(truth uint16) (bool, error) {
	if err := manager.check(); err != nil {
		return false, err
	}
	return manager.practical[truth], nil
}
func (manager *RwtManager) PracticalTable() ([]bool, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	return slices.Clone(manager.practical), nil
}
func (manager *RwtManager) ClassRepresentative(class int) (uint16, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if class < 0 || class >= len(manager.mapInverse) {
		return 0, ErrDimension
	}
	return manager.mapInverse[class], nil
}
func (manager *RwtManager) Decomposition(id int) (*DecoGraph, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	if id < 0 || id >= len(manager.graphs) || manager.graphs[id] == nil {
		return nil, ErrDimension
	}
	return manager.graphs[id].Clone()
}
func (manager *RwtManager) Stats() (RwtStats, error) {
	if err := manager.check(); err != nil {
		return RwtStats{}, err
	}
	return manager.stats, nil
}
func (manager *RwtManager) AddCutTime(d time.Duration) error {
	if err := manager.check(); err != nil {
		return err
	}
	manager.stats.CutTime += d
	return nil
}
func (manager *RwtManager) AddUpdateTime(d time.Duration) error {
	if err := manager.check(); err != nil {
		return err
	}
	manager.stats.UpdateTime += d
	return nil
}
func (manager *RwtManager) AddTotalTime(d time.Duration) error {
	if err := manager.check(); err != nil {
		return err
	}
	manager.stats.TotalTime += d
	return nil
}
func (manager *RwtManager) Leaves() ([]int, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	return slices.Clone(manager.leaves), nil
}
func (manager *RwtManager) ResultGraph() (*DecoGraph, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	if manager.graph == nil {
		return nil, nil
	}
	return manager.graph.Clone()
}
func (manager *RwtManager) ResultComplement() (bool, error) {
	if err := manager.check(); err != nil {
		return false, err
	}
	return manager.complement, nil
}

// StatsText implements Rwt_ManPrintStats without process-global output.
func (manager *RwtManager) StatsText() (string, error) {
	if err := manager.check(); err != nil {
		return "", err
	}
	used := 0
	for _, score := range manager.stats.Scores {
		if score > 0 {
			used++
		}
	}
	s := manager.stats
	return fmt.Sprintf("Rewriting statistics:\nTotal cuts tries  = %8d.\nBad cuts found    = %8d.\nTotal subgraphs   = %8d.\nUsed NPN classes  = %8d.\nNodes considered  = %8d.\nNodes rewritten   = %8d.\nCalculated gain   = %8d.\nStart        =%9.2f sec\nCuts         =%9.2f sec\nTruth        =%9.2f sec\nResynthesis  =%9.2f sec\n    Mffc     =%9.2f sec\n    Eval     =%9.2f sec\nUpdate       =%9.2f sec\nTOTAL        =%9.2f sec\n",
		s.CutsGood, s.CutsBad, s.Subgraphs, used, s.NodesConsidered, s.NodesRewritten, s.NodesGained,
		s.Start.Seconds(), s.CutTime.Seconds(), s.TruthTime.Seconds(), s.ResynthesisTime.Seconds(),
		s.MFFCTime.Seconds(), s.EvaluationTime.Seconds(), s.UpdateTime.Seconds(), s.TotalTime.Seconds()), nil
}
func (manager *RwtManager) StatsRecord() ([]byte, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	s := manager.stats
	return []byte(fmt.Sprintf("%d %d %d %d\n", s.CutsGood, s.Subgraphs, s.NodesRewritten, s.NodesGained)), nil
}

// Close implements Rwt_ManStop. It releases all owned decomposition graphs.
func (manager *RwtManager) Close() error {
	if err := manager.check(); err != nil {
		return err
	}
	for _, graph := range manager.graphs {
		if graph != nil {
			graph.Close()
		}
	}
	if manager.graph != nil {
		manager.graph.Close()
	}
	manager.canonical = nil
	manager.phases = nil
	manager.permutations = nil
	manager.classMap = nil
	manager.practical = nil
	manager.nodes = nil
	manager.table = nil
	manager.classes = nil
	manager.mapInverse = nil
	manager.graphs = nil
	manager.leaves = nil
	manager.closed = true
	return nil
}

var rwtPracticalClasses = [...]uint16{0x0000, 0x0001, 0x0003, 0x0006, 0x0007, 0x000f, 0x0016, 0x0017, 0x0018, 0x0019, 0x001b, 0x001e, 0x001f, 0x003c, 0x003d, 0x003f, 0x0069, 0x006b, 0x006f, 0x007e, 0x007f, 0x00ff, 0x0116, 0x0118, 0x0119, 0x011a, 0x011b, 0x011e, 0x011f, 0x012c, 0x012d, 0x012f, 0x013c, 0x013d, 0x013e, 0x013f, 0x0168, 0x0169, 0x016f, 0x017f, 0x0180, 0x0181, 0x0182, 0x0183, 0x0186, 0x0189, 0x018b, 0x018f, 0x0198, 0x0199, 0x019b, 0x01a8, 0x01a9, 0x01aa, 0x01ab, 0x01ac, 0x01ad, 0x01ae, 0x01af, 0x01bf, 0x01e9, 0x01ea, 0x01eb, 0x01ee, 0x01ef, 0x01fe, 0x033c, 0x033d, 0x033f, 0x0356, 0x0357, 0x0358, 0x0359, 0x035a, 0x035b, 0x035f, 0x0368, 0x0369, 0x036c, 0x036e, 0x037d, 0x03c0, 0x03c1, 0x03c3, 0x03c7, 0x03cf, 0x03d4, 0x03d5, 0x03d7, 0x03d8, 0x03d9, 0x03dc, 0x03dd, 0x03de, 0x03fc, 0x0660, 0x0661, 0x0666, 0x0669, 0x066f, 0x0676, 0x067e, 0x0690, 0x0696, 0x0697, 0x069f, 0x06b1, 0x06b6, 0x06f0, 0x06f2, 0x06f6, 0x06f9, 0x0776, 0x0778, 0x07b0, 0x07b1, 0x07b4, 0x07bc, 0x07f0, 0x07f2, 0x07f8, 0x0ff0, 0x1683, 0x1696, 0x1698, 0x169e, 0x16e9, 0x178e, 0x17e8, 0x18e7, 0x19e6, 0x1be4, 0x1ee1, 0x3cc3, 0x6996}

const rwtStaticGzip = "H4sIAAAAAAAC/x2XDVCWZbqA7+d53n8CBETFfxQhUJBASxihldDFP1I01hjlqKOETRqyyqSWJxUp3TVadsE0JXWKdc2YJHPUs1aamXWUJbb1pGOuy5ajpjbg5mgacy4+Z6657+d+7v/38/0+PNESgCcGacRHfwgiQtIgjYSHdAvdQtpIG5slvbBHQy98oiEqpFvoPXc20sbWozvoDtJFutgc6cNd75B0kC7SldiQ7qF7SB/pY/OkH/b+0A+f/hAX0n10H92TweQaiG0wDOR+MAwK6T13HtJH+tg8GYp9GAzFZxjEh3QPvefOR/rYPEnAnggJ+CTCiNC5x+4jfc6ePMxsKfAw9VPweRhbCnfJIRmGDJCRyAhkFMTIKPzTYBQxo/BNg1T8U9lsGr5p+KWxpXR8MsiZwf0j3GeQL4OtZ+DzCPkyyDUGn0dhDLkehbGhs8fZR/qcPWQYhEMU517IGBmHXxaMwycbsvAbh192iHCIgmhsMcgYGc9zzYXxPONc6uTS13hicmE88bmQEzqHQxTEQF8YAIPkV8TmhbBlAnvMw3cC5OGfh38eTzmPJ5mHfx5PawLk4z+JehPpcxL1JuI/if4mMf8k8k+kv0nUmETMJJ5YAf5ToICYKTA5pHvoHtJH+th6zmEQCVEQDTEQC3EwAAbJNGKnkWN6SHpSCNOIL4TpIT0MPQwZiewhCj0aYtB7iEXvi4xDDkAOQA6EQeiDZAYzFcEMdlFEziL2UESeImKLiCsirgi/IhkC8TAcEmCEzAzJJJlNfDHMZqfF5JnNjorpo5haxXzaZxNbHGIoerw8hSwmdjY5iuGpkJ6EnoSeKCXEl9BLCXOV0H8JsU8TMwe/khBJnBNlLjVLYS51S6k7l7h5yFJmKKX/UvqfS+w8eiglvpTapfyvKmWGueQphbnUL4V5oXMiJMlCci6AheRaAPPZx3zyLYT57GMh+5vPThaGGIIeD8PRE5BJxCTKIp5XGSwivoy+ypinjL7KyFHGTGX0VkauMt4ci6GMPGXkKCO+TEZDujxLfDnPvpwcz4aIQu8DfdF76A8DYASkyjJ8l1FrGbWWsbtl1FmG/zJ8n8N3WYghMBLGSgX5K4ipJHcFccuRlcRVMm8lfVYQv5y3XyXxlbzlKui3EirIU0HdSliOXsn8lbxZqshXRb6VIenICnJWwUr0Kmavov8q8q4kZxV5qoitIsdK9rmC2VfybFawuypYwR6q6LOKXbxAvhd4HqtD2LIqJH0IIAzCIRJ6QRREQwz0hQEQD0kwEkZDOjnGILMgW9aS/0Vyrg3xEHo4MgpiYACMgFR4TNYz33p6WE/99dR+mZrrQ0RDLAyBkTAWsuBxqSZ/DVQTWwMbQ2eLs41uoQfoYegBejhEQCS2XhCFHsVdjx4NMZxjOEcj42A4JEMmZMlm8m8i/6vk3QyvUmNTSA+DXpyjkFHYevRo9Gj0HtkXBqAPRY7gLgk5EkZDFuds2cKbsI6567ivo2YdO6mjbh17qeOtW8fbsp7a9dSrJ389eRvYfQO+9XyTNLD3BvwbyNUgT8h23l/bybctJKNgOGRCDuTDRO7yZSc5G/HZgW8j7ERvxH8nNRrZdyM73oV9F7Y91NhDjiY+c0189pqwNZGjSabKPs77eEbN3DVjb2beZmZppl4zvTfTUzM1W7h/H98DyBb2coD/Q++z6xZmacH3Q97MR6l3lGf0V+RHISLlGM/tI/xPcD6B/wl6O0Evpzh/xt0pap9ixpPs7RR7/YydnKJeK7Vayd3KXSv9tLL7VuJaqdVOD23EtdFrO3zF/5k28rbj9xXvm3Z2307v7eQ5z5z/YvcdfIY6yHeZPV8hxxXeSFfkGbnG3XV2d51+rvNMr/Gtep3d3cD/VojHpAv/Lt5kP7Hzn5nlZ87d9PALPKCPbnp4QO1u+IX63dTu5pvoF75NlOJXlJougfIhTHwVIJ+EORKhPHkIewSEo0dwH6FGoqcgx0IOPA6T4TfYi2GORCkL+L1GXDREERutRnPmF4R6DD0LcuBxmMh9PnIqPA1L4L+lD3GxxPWmZp8Qc6S/ioUhMBIeh8mwUQYrfrOpOBgAwyEJRsMTMBUK4Un4L1gAi2GDDKPHocQOg3j0YeSLx3eYqpURzJCAfQT2BPImhpgIr8jD2JOxp1AzhTlS1ExYDBskVZVIBr1nkGMMcqzKlWx8spkxC3Kx5TDXeGzjOY9Xs7DVQr3kcc5TW6FR8qlRwLwFaiAMgUSZrGpkOnUL8ZuuCpCNUkS+mapJnlLpMk9ly0Jsi1QkxMgSFSEruF9FvZeR64irgY2wWZVLPbM2YX+H+wPEnlTJkC5f0s+X9HOG8xnFZ5pez6kW+Qbf8zz78zzn8+pt+ZaYDuxXqHVFvQVH4RP4O/wDvperypYu7rvI14VPFz5d+HTho3Ut8PbRt+GeBDpffH1QIjSfN81nT+dAPszCVgv16AeQt+Ge9NIxEqUTkIkSpy0ZjM9gfOJ1tqST41EYo5MhHTIhG3IgX8bqJsnmPgfbRM6TkYW6ERw1k3xzyP00uefoBcga5OfwT/QbyAcyj7t5+pLM13eknDxLyFFOX+X0tZRcFbpcViJXY18Fa/VaWYfcyFxbtAPtUq/PSAPx9fqubKPHRnx26peliTmamGMfve/D/iExR8h1jBmPUfuEbpHj7Okk96dDZMqXzNWK3op/O7W/0s/KP+A8sZf0KNUB39Hn9zoFimGUuo7/dfyvwQ30G8hb2DuJ76RGp/bQR6k77OKOjoUhkA6L4Rl4Cw7BEfgE/gYu/mEQCbGQCmNgFiyHKnVXG/lZr1EP9FF5oPkcmBgYAukwDxbDM7AN3oL34Ah8An8TZdqhG91V2oRBJPSFMTALfgNVyjH8JWVaxDcXJdzclWhzW/qZtTLIvCzxxoM/SaJZoxLNOpVs0lQKegp6mqmVNMO3H/5p5p5kGP6Gwj/d5EA+zMJWC/XoB5C34R56ksogRwY5xjPTeDNQckwN3IBKlWt2qolGSz6x+cRMZ94ikymz6XM2egl1SogvMW/CHlVKnvnYymARPovwWYT/IpMNOZAvS7BVcFeBrOJuNb6riV9N/Ivo68lRg6zGpxqfanyqia8mvpr4zSEawVGbzRtqi4mAIpij6rDXYa/DXs8c9WY73IBK9YYphJWwRe0ww2WHSYIfIYtzudpl4mSXGQTDIUl2m0J52yxUb5sNai+72Yttr8lEL4B1aj+2/djeZ/73OR80U+WguQz35bCJhBh4Fz6GNjlmLsAodZq709ydpt5p6p2lz7P0eJYe24lvp6d2E42eBeXqHPnPkf8CtS9Qu4P+Oqh7JcQ6dYv7W8hu9qOsFKWsRmVbu1RghUtgDUE+r8KxR1trVH8rU/pbObBOxVs1Em9thxtQqdIsW2dDLn551k9qmuVKEf5FVqsqstqgE+7DCD3bsqWEuxLylHBfwn0J90usS2oZttXYVnPeaI1Sm+B16xvVgH8Dtd+wauG2NFr5EK93W47sQW9C/4tlyX4rSQ7Q1yFrqxxGP2XF6S/p6Yw1Vdqsd+Uc+jdWpHxjxUEm5MilEEZdsn4Hj+gO8nVYjeCoDusN9b31NhxS17Fdx3Yd200rTW5a42CPdLKLTmp2sotOdnHPKpS7VrK+a6VqsXO1ZffVvezb0tt+AqbDp5CjetuxuredoWPt+TrOTpF+djo8A1/AD6qf/aMaaGdBuRpqJ0miXagS7ZWwRaXaojPseMmwP4Y2yOJcpnPsBMmxL8gUm7/Z7XchWxdSt5C6hdQtpG4hdQupW2jv1EXkLCJfMf0V29+pEnubnmfXyDx7O9yASlVmX4R39FI70KvtHFltf61W27fVKrtUr6W/tfS3znakxl6jamz+/9m/16/arrxmZ0IttKnX7Bf167aRBmwN2BrsemhTDdj32M/rd+xqvY/4/batW7g/hH7YvqCO21X6lF2gT9u/hrX6DPXPUP8M9dvovY3eL9j35FvqX7Lz5Xs7Ta7Ynfo6PtfxuW6n6pu2ljvs72d7qzwgXjs54ji3JXCSJAI9wjEqwvkdPKIjnK90lPNARTljdH/nFT3I6a3jna9VCr4pTqpOcxolzXFUmvOGykDPQM9Az3LyJMtpAU/lOHfVBKdGJjjb4QZUqokO37n4FDivwE34rZrs7FIFzgRd4DwP/wTXTHOy9TTnAz3TWSgznL7I/XqGc04/Rew8WOQc0IucVr3EGaeXOinynJMlS5076jknRj/npGHbg/yXXs5dhcO70nkGvoAfVIXzI3yqVzpD5QUnU73gFKgXmWsdNaudeKl2PoY22cT8rzkLYYP6gyO6wRkuDc5luA/leruzXu901soeZyrk6j3ORd3kbJV99LaP3prpqZl+mumnmV5ayNVCrkPkOUSeQ+Q56tTCd+ojZ6o5zq6Os6vj7Oo4uzrlXICX9SlnpmnFr935t2rnWbQ7s8w5Z7U+57ypLziZcsFpVRecTnWJWS4xSwczdDBDBzN8z0zXHEtuOb/Xt5xxpst5GNLNfxxb7jmXVTfnbs7a5fPg8nvMTTeWayTgHHAO3K8VEtJNlJsCaaYPd/1cW/dz1+vB+Ax1L6qh7js60U0wKe5mneL+WSe7A00y/mn4prn3gM+MWwDf6XRs6e6vkf+GBJMBWeTNxf4ryHf59eS+qQvcTClwW1WB26mmuQthg5rhDpcZ7mW4L0+5k3SJWy8l7lYzz7XkWfcDs9TVUune15WuMpVuLGw1VW6trHX5G9ndLXXuTvgU7us6fOrwqXOPmQbuG9wndSP3TSHu6ybum7hv4n4f9/u4b+buiFsoh91X4Cb8Vh12t6ojbn99xD0Mi8xHbrJ85KbLJ+4adQpOs6NWaKOPc3Devaq+det0hztLrnC+wt0V/K4Sd5W4G248vKe73HWqi/r/4XzXjdS/uGkiXq3YniO+d8KEe9US7fF3lsffWd5tuAdrVLS3TvX2MqWfN1wGeht1osc3JnGjuMvwnpAMbwDyA/hUjfVSzDhvo2Qhc73JkkPOXO++zvUU51jYavKoMYEa+chJ5JlC7kLyzvS2S7G3W572Bpp5Xoss8KLhjF7szZHF3lpZ4k2R5V4vqfJyZJU3yqz1Nsl6L1mv91Lhc/hfXePFS433nt7oFZvNXl/Z7EXqTd5ss8XbIXX413uT9Hb6b6Ruo3dPN3oxZje9NnmzpMmrh61mL3O95/U3LciD7OFDb5c+TOxxL0FOcv6CXs56L8Iwfdb7k/7Ky5ZzzHIeedEbIpfJfZnY77Bd9W6qm8zTRc93sN/B/jP2bs7av6asEOtM4B8w4X5/08vfpnv7U02c78lAPxNyZKifJAkh7qsUvxU6Vao/0KT7Q+RR/zMZF2KdGe8PkDz/JTUF38IQ/C4gb4mfLCU+f9fBAijz7+kyP8aU+z/opdSp8NeoFdRa5c9QL/nbZb3/R10dYpXZjM9rvi919NEA2/yF0ojcRf53/Hy11/9QN/u10uzvlBZ/OsyCTyFHtfifwXJ90B8hR8l/1F8lf/Vj5YT/uDoZYpX5kvizviMXqHGeXi76YfCkfOvXyGVsV/zdchV5k3MX83b5o2EyPA8b4Q/wF/g7/B8MUl1+PIzVXf4fzR0/wdzF/y4+d7nrJkc3++imn27m6CZHN/13Ey8B319BqooIPpBeQRi8YvoET0u/wJfBwROSEGyUFPTUwJOMYLPOCP6sHwkGmrFBmskKHMkObJUdPKnHB9t0XnBTTUIWIKeRo4h8ReQrRpaSYz61yvB/NqhXz2GrCmbJC9hXB7Wynrga4jYHSfJaMFXq8K0PhpltwTjZFrwuO4IF8Dmkqh3B8/ASfKF3BBvM7qCXvE3/e4PRsp/494k/SO9HiDvG3ckgTs4EP8FIfSZYY9qp1x50qnNBgjkfpJtLnC9xvhzw+UW/FhTKLfJ0MW8X83Yy7x3O95DdwTXpDv5H67AwccJ+gpHaCVtjgrAc6fn3/4ZL9yvUGwAA"
