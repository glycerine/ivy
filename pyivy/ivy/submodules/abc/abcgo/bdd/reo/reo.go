// Package reo implements sifting-based shared decision-diagram reordering.
// Algorithms and cost definitions follow Alan Mishchenko's REO engine (2002).
// Go uses canonical diagram rebuilds in place of pointer-tag mutation.
package reo

import (
	"fmt"
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"io"
)

type DdManager = cudd.DdManager
type DdNode = cudd.DdNode
type Minimize int
type reo_min_type = Minimize

const (
	REO_MINIMIZE_NODES Minimize = iota
	REO_MINIMIZE_WIDTH
	REO_MINIMIZE_APL
)
const (
	REO_REORDER_LIMIT = 1.15
	REO_QUAL_PAR      = 3
	REO_CONST_LEVEL   = 30000
	REO_TOPREF_UNDEF  = 30000
	REO_CHUNK_SIZE    = unitChunkSize
	REO_COST_EPSILON  = 1e-7
	REO_HIGH_VALUE    = 10000000
	REO_ENABLE        = 1
	REO_DISABLE       = 0
)

// Profile records shared DAG size, total frontier width, and the sum of expected
// decision counts under uniformly distributed inputs (one unit per root).
type Profile struct {
	Nodes, Width           int
	APL                    float64
	LevelNodes, LevelWidth []int
	LevelAPL               []float64
}
type Statistics struct {
	Before, After     Profile
	Swaps, Iterations int
}

// Manager is reusable, but callers must serialize calls that mutate it or the
// associated CUDD manager. The returned inverse mapping restores original inputs.
type Manager struct {
	unitStorage
	criterion             Minimize
	remapUp, verify       bool
	hasADD                bool
	iterations            int
	output                io.Writer
	stats                 Statistics
	dd                    *DdManager
	original, roots       []*DdNode
	support, slots, order []int
	current               Profile
	planes                []reo_plane
	units                 map[*DdNode]*reo_unit
	nodes                 map[*reo_unit]*DdNode
}
type reo_man = Manager

func New(maxVariables, maxNodes int) *Manager {
	if maxVariables < 0 || maxNodes < 0 {
		panic("reo: negative capacity")
	}
	return &Manager{iterations: 1, unitStorage: unitStorage{nMemChunksAlloc: int32((10*maxNodes+1)/unitChunkSize + 1)}}
}
func (p *Manager) SetMinimization(kind Minimize) {
	if kind < REO_MINIMIZE_NODES || kind > REO_MINIMIZE_APL {
		panic("reo: invalid minimization criterion")
	}
	p.criterion = kind
}
func (p *Manager) SetRemapping(enabled bool)    { p.remapUp = enabled }
func (p *Manager) SetVerification(enabled bool) { p.verify = enabled }
func (p *Manager) SetIterations(n int) {
	if n < 0 {
		panic("reo: negative iterations")
	}
	p.iterations = n
}
func (p *Manager) SetOutput(w io.Writer)  { p.output = w }
func (p *Manager) Statistics() Statistics { return cloneStatistics(p.stats) }
func (p *Manager) Close() {
	p.stop()
	p.dd = nil
	p.original = nil
	p.roots = nil
	p.units = nil
	p.nodes = nil
	p.planes = nil
	p.support = nil
	p.slots = nil
	p.order = nil
}
func (p *Manager) printProfile(label string, profile Profile) {
	if p.output != nil {
		fmt.Fprintf(p.output, "%s: nodes=%d width=%d apl=%.6f\n", label, profile.Nodes, profile.Width, profile.APL)
	}
}
