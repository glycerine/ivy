package abc

// Port of src/map/amap/amapMan.c and src/map/amap/amapCore.c.

import (
	"fmt"
	"math"
)

const amapInitialVectorCapacity = 100

// AMapParams contains the mapper controls initialized by
// Amap_ManSetDefaultParams. Zero values are not the source defaults; use
// DefaultAMapParams when the caller wants the native configuration.
type AMapParams struct {
	FlowIterations int
	AreaIterations int
	MaximumCuts    int
	UseMUXes       bool
	UseXORs        bool
	FreeInverters  bool
	Epsilon        float32
	AreaDelayRatio float32
	Verbose        bool
}

// DefaultAMapParams returns the exact active defaults in amapCore.c.
func DefaultAMapParams() AMapParams {
	return AMapParams{
		FlowIterations: 1,
		AreaIterations: 4,
		MaximumCuts:    500,
		UseXORs:        true,
		Epsilon:        0.001,
	}
}

// SetDefaultAMapParams is the checked in-place form of
// Amap_ManSetDefaultParams. Assignment replaces every prior field, matching
// the source memset followed by explicit initialization.
func SetDefaultAMapParams(params *AMapParams) error {
	if params == nil {
		return ErrDimension
	}
	*params = DefaultAMapParams()
	return nil
}

func validateAMapParams(params AMapParams) error {
	if params.FlowIterations < 0 || params.AreaIterations < 0 || params.MaximumCuts <= 0 {
		return ErrDimension
	}
	if params.Epsilon <= 0 || math.IsNaN(float64(params.Epsilon)) || math.IsInf(float64(params.Epsilon), 0) {
		return ErrDimension
	}
	if params.AreaDelayRatio < 0 || math.IsNaN(float64(params.AreaDelayRatio)) || math.IsInf(float64(params.AreaDelayRatio), 0) {
		return ErrDimension
	}
	return nil
}

// AMapLibrary is the part of Amap_Lib_t consumed by amapCore.c. The complete
// gate and rule representation belongs to the remaining AMAP source units.
type AMapLibrary struct {
	Name         string
	InverterArea float64
	HasXOR       bool
	HasMUX       bool
	NodeCount    int
}

func (library AMapLibrary) valid() bool {
	return library.NodeCount >= 0 && library.InverterArea >= 0 &&
		!math.IsNaN(library.InverterArea) && !math.IsInf(library.InverterArea, 0)
}

// AMapOutput is the owned Go representation of Amap_Out_t. Type is -1 for an
// input, 0 for an internal gate, and 1 for an output.
type AMapOutput struct {
	Name   string
	Type   int16
	Fanins []int
}

func cloneAMapOutputs(outputs []AMapOutput) ([]AMapOutput, error) {
	copyOutputs := make([]AMapOutput, len(outputs))
	for index, output := range outputs {
		if output.Type < -1 || output.Type > 1 {
			return nil, fmt.Errorf("%w: AMAP output type %d", ErrInvariant, output.Type)
		}
		copyOutputs[index] = output
		copyOutputs[index].Fanins = append([]int(nil), output.Fanins...)
	}
	return copyOutputs, nil
}

// AMapObject and AMapCut are the owned entries needed by the manager storage
// created in amapMan.c. Their detailed fields are supplied by the graph and cut
// source units; this lifecycle slice intentionally treats them as opaque.
type AMapObject struct{}
type AMapCut struct{}

// AMapManagerOptions adds checked allocation and synchronous cancellation to
// the native lifecycle. A zero limit selects a conservative implementation
// default.
type AMapManagerOptions struct {
	MaxNodes            int
	MaxTemporaryEntries int
	Stop                func() bool
}

func normalizeAMapManagerOptions(options AMapManagerOptions) (AMapManagerOptions, error) {
	if options.MaxNodes < 0 || options.MaxTemporaryEntries < 0 {
		return options, ErrDimension
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 1 << 26
	}
	if options.MaxTemporaryEntries == 0 {
		options.MaxTemporaryEntries = 1 << 26
	}
	return options, nil
}

// AMapManager owns all arrays allocated by Amap_ManStart and the temporary
// matching arrays allocated by Amap_ManTest. It must not be copied.
type AMapManager struct {
	params  AMapParams
	library AMapLibrary

	epsilonInternal float32
	inverterArea    float64
	useXOR          bool
	useMUX          bool

	primaryInputs  []*AMapObject
	primaryOutputs []*AMapObject
	objects        []*AMapObject
	temporaryInts  []int
	cutBuffers     [3][]*AMapCut
	temporaryObjs  []*AMapObject

	temporaryMatches []int
	temporaryCuts    []*AMapCut
	// primaryInputCuts is pCutsPi. It is populated by the graph/cut port, but
	// belongs to this manager and must be released by Close.
	primaryInputCuts []AMapCut
	objectChunkSize  int

	options AMapManagerOptions
	closed  bool
}

// NewAMapManager implements Amap_ManStart. nodeCapacity is the expected object
// count passed to Aig_MmFixedStart: it selects max(nodeCapacity/8, 8) entries
// per allocator chunk. All eight source work vectors, including vObjs, retain
// their independent native initial capacity of 100.
func NewAMapManager(nodeCapacity int, options AMapManagerOptions) (*AMapManager, error) {
	options, err := normalizeAMapManagerOptions(options)
	if err != nil || nodeCapacity < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if nodeCapacity > options.MaxNodes {
		return nil, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	chunkSize := nodeCapacity / 8
	if chunkSize < 8 {
		chunkSize = 8
	}
	manager := &AMapManager{
		epsilonInternal: 0.01,
		primaryInputs:   make([]*AMapObject, 0, amapInitialVectorCapacity),
		primaryOutputs:  make([]*AMapObject, 0, amapInitialVectorCapacity),
		objects:         make([]*AMapObject, 0, amapInitialVectorCapacity),
		temporaryInts:   make([]int, 0, amapInitialVectorCapacity),
		temporaryObjs:   make([]*AMapObject, 0, amapInitialVectorCapacity),
		objectChunkSize: chunkSize,
		options:         options,
	}
	for index := range manager.cutBuffers {
		manager.cutBuffers[index] = make([]*AMapCut, 0, amapInitialVectorCapacity)
	}
	return manager, nil
}

func (manager *AMapManager) check() error {
	if manager == nil {
		return ErrDimension
	}
	if manager.closed {
		return ErrInvariant
	}
	if manager.options.MaxNodes <= 0 || manager.options.MaxTemporaryEntries <= 0 {
		return ErrInvariant
	}
	return nil
}

func (manager *AMapManager) stopped() bool {
	return manager.options.Stop != nil && manager.options.Stop()
}

// Close implements Amap_ManStop. It releases all retained references and is
// idempotent so deferred cleanup is safe on every orchestration failure path.
func (manager *AMapManager) Close() error {
	if manager == nil {
		return ErrDimension
	}
	if manager.closed {
		return nil
	}
	manager.primaryInputs = nil
	manager.primaryOutputs = nil
	manager.objects = nil
	manager.temporaryInts = nil
	for index := range manager.cutBuffers {
		manager.cutBuffers[index] = nil
	}
	manager.temporaryObjs = nil
	manager.temporaryMatches = nil
	manager.temporaryCuts = nil
	manager.primaryInputCuts = nil
	manager.library = AMapLibrary{}
	manager.params = AMapParams{}
	manager.closed = true
	return nil
}

// Closed reports lifecycle state without exposing manager storage.
func (manager *AMapManager) Closed() bool { return manager != nil && manager.closed }

// AMapManagerSnapshot is an owned diagnostic view used by subsequent mapping
// units and by standalone lifecycle tests.
type AMapManagerSnapshot struct {
	EpsilonInternal   float32
	InverterArea      float64
	UseXOR            bool
	UseMUX            bool
	ObjectCapacity    int
	ObjectChunkSize   int
	InputCapacity     int
	OutputCapacity    int
	TemporaryCapacity int
	CutCapacities     [3]int
	MatchSlots        int
	CutSlots          int
}

func (manager *AMapManager) Snapshot() (AMapManagerSnapshot, error) {
	if err := manager.check(); err != nil {
		return AMapManagerSnapshot{}, err
	}
	snapshot := AMapManagerSnapshot{
		EpsilonInternal:   manager.epsilonInternal,
		InverterArea:      manager.inverterArea,
		UseXOR:            manager.useXOR,
		UseMUX:            manager.useMUX,
		ObjectCapacity:    cap(manager.objects),
		ObjectChunkSize:   manager.objectChunkSize,
		InputCapacity:     cap(manager.primaryInputs),
		OutputCapacity:    cap(manager.primaryOutputs),
		TemporaryCapacity: cap(manager.temporaryInts),
		MatchSlots:        len(manager.temporaryMatches),
		CutSlots:          len(manager.temporaryCuts),
	}
	for index := range manager.cutBuffers {
		snapshot.CutCapacities[index] = cap(manager.cutBuffers[index])
	}
	return snapshot, nil
}

// AMapPipeline supplies the three algorithms called by Amap_ManTest but
// implemented in amapGraph.c, amapMatch.c and amapOutput.c. Concrete typed
// callbacks keep this orchestration independently testable without a runtime
// interface or a global library singleton.
type AMapPipeline struct {
	Create  func(*AMapManager, *AigManager) error
	Map     func(*AMapManager) error
	Produce func(*AMapManager) ([]AMapOutput, error)
}

func (pipeline AMapPipeline) valid() bool {
	return pipeline.Create != nil && pipeline.Map != nil && pipeline.Produce != nil
}

// RunAMapCore implements the non-timing behavior of Amap_ManTest. The library
// is explicit instead of process-global, errors replace printing, and output is
// deeply owned. Manager cleanup runs after success, failure, or cancellation.
func RunAMapCore(aig *AigManager, params AMapParams, library *AMapLibrary, pipeline AMapPipeline, options AMapManagerOptions) ([]AMapOutput, error) {
	if aig == nil || library == nil || !pipeline.valid() {
		return nil, ErrDimension
	}
	if err := aig.Check(); err != nil {
		return nil, err
	}
	if err := validateAMapParams(params); err != nil {
		return nil, err
	}
	if !library.valid() {
		return nil, ErrDimension
	}
	options, err := normalizeAMapManagerOptions(options)
	if err != nil {
		return nil, err
	}
	if library.NodeCount > options.MaxTemporaryEntries/2 {
		return nil, ErrLimit
	}
	manager, err := NewAMapManager(aig.AndCount(), options)
	if err != nil {
		return nil, err
	}
	defer manager.Close()
	manager.params = params
	manager.library = *library
	if params.FreeInverters {
		manager.inverterArea = 0
	} else {
		manager.inverterArea = library.InverterArea
	}
	manager.useMUX = params.UseMUXes && library.HasMUX
	manager.useXOR = params.UseXORs && library.HasXOR
	manager.temporaryCuts = make([]*AMapCut, 2*library.NodeCount)
	manager.temporaryMatches = make([]int, 2*library.NodeCount)

	if manager.stopped() {
		return nil, ErrCancelled
	}
	if err = pipeline.Create(manager, aig); err != nil {
		return nil, err
	}
	if manager.stopped() {
		return nil, ErrCancelled
	}
	if err = pipeline.Map(manager); err != nil {
		return nil, err
	}
	if manager.stopped() {
		return nil, ErrCancelled
	}
	outputs, err := pipeline.Produce(manager)
	if err != nil {
		return nil, err
	}
	if manager.stopped() {
		return nil, ErrCancelled
	}
	return cloneAMapOutputs(outputs)
}
