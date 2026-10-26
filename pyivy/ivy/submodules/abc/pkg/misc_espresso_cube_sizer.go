package abc

// Binary cube sizing and minimizer defaults from src/misc/espresso/cubehack.c.

import (
	"slices"
	"sync"
)

type EspressoMinimizeFlags struct {
	Summary, Trace                    bool
	RemoveEssential, ForceIrredundant bool
	UnwrapOnset, SingleExpand         bool
	PositiveOutput, RecomputeOnset    bool
	UseSuperGasp, UseRandomOrder      bool
}

func DefaultEspressoMinimizeFlags() EspressoMinimizeFlags {
	return EspressoMinimizeFlags{
		RemoveEssential:  true,
		ForceIrredundant: true,
		UnwrapOnset:      true,
	}
}

func newEspressoBinaryLayout(inputCount int) (*EspressoCubeLayout, error) {
	if inputCount < 0 {
		return nil, ErrDimension
	}
	if inputCount > espressoSetMaximumBits/2 {
		return nil, ErrLimit
	}
	return NewEspressoCubeLayout(slices.Repeat([]int{2}, inputCount), inputCount)
}

// EspressoCubeSizer is an owned replacement for cubehack.c's process-global
// layout and hidden 100-input allocation cache.
type EspressoCubeSizer struct {
	mu     sync.Mutex
	layout *EspressoCubeLayout
}

func NewEspressoCubeSizer() *EspressoCubeSizer { return &EspressoCubeSizer{} }

// CautiousDefine installs an exact all-binary layout, reusing an equal current
// value. Allocation capacity is deliberately unobservable in the Go API.
func (sizer *EspressoCubeSizer) CautiousDefine(inputCount int) (*EspressoCubeLayout, error) {
	if sizer == nil {
		return nil, ErrDimension
	}
	sizer.mu.Lock()
	defer sizer.mu.Unlock()
	if sizer.layout != nil && sizer.layout.binaryVariables == inputCount && len(sizer.layout.partSizes) == inputCount {
		return sizer.layout.Clone()
	}
	layout, err := newEspressoBinaryLayout(inputCount)
	if err != nil {
		return nil, err
	}
	sizer.layout = layout
	return layout.Clone()
}

// Define preserves define_cube_size's logical result. The native <=100
// over-allocation is only a memory-reuse strategy and is not exposed.
func (sizer *EspressoCubeSizer) Define(inputCount int) (*EspressoCubeLayout, error) {
	return sizer.CautiousDefine(inputCount)
}

func (sizer *EspressoCubeSizer) Current() (*EspressoCubeLayout, error) {
	if sizer == nil {
		return nil, ErrDimension
	}
	sizer.mu.Lock()
	defer sizer.mu.Unlock()
	if sizer.layout == nil {
		return nil, nil
	}
	return sizer.layout.Clone()
}

// Undefine represents undefine_cube_size by releasing the owned cached value.
func (sizer *EspressoCubeSizer) Undefine() error {
	if sizer == nil {
		return ErrDimension
	}
	sizer.mu.Lock()
	sizer.layout = nil
	sizer.mu.Unlock()
	return nil
}
