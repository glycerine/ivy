// Package super enumerates technology supergates and AND/inverter trees.
// Algorithms derive from ABC/MVSIS superGate.c (UC Berkeley).
package super

import (
	"context"
	"fmt"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"math"
	"os"
	"sort"
	"time"
)

const SUPER_NO_VAR float32 = -9999
const SUPER_EPSILON float32 = .001
const SUPER_FULL uint32 = ^uint32(0)

func SUPER_MASK(bits int32) uint32 {
	if bits <= 0 {
		return 0
	}
	if bits >= 32 {
		return ^uint32(0)
	}
	return (1 << uint(bits)) - 1
}

type Options struct {
	Inputs, Levels, MaxGates       int
	DelayLimit, AreaLimit          float32
	TimeLimit                      time.Duration
	SkipInputInverters, Exhaustive bool
	Context                        context.Context
	Progress                       io.Writer
}

func DefaultOptions() Options { return Options{Inputs: 5, Levels: 2, SkipInputInverters: true} }

type Statistics struct {
	Tried, Added, Removed, Lookups int64
	TimedOut, LimitReached         bool
	Elapsed                        time.Duration
}
type Super_Gate_t struct {
	Root     *mio.Mio_Gate_t
	Variable int // -1 for a composite
	Truth    uint64
	Fanins   []*Super_Gate_t
	Area     float32
	Delays   []float32
	MaxDelay float32
	serial   int
}
type Super_Man_t struct {
	Name       string
	Options    Options
	Inputs     []*Super_Gate_t
	Gates      []*Super_Gate_t // last frozen generation frontier; output after Write
	Levels     int
	Stats      Statistics
	table      map[uint64][]*Super_Gate_t
	nextSerial int
	started    time.Time
}
type Mio_Library_t = mio.Mio_Library_t
type Mio_Gate_t = mio.Mio_Gate_t
type Vec_Str_t = vec.Vec_Str_t

func finite(value float32) bool { return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0) }
func truthMask(inputs int) uint64 {
	if inputs == 6 {
		return ^uint64(0)
	}
	return (uint64(1) << uint(1<<uint(inputs))) - 1
}
func Super_ManStart() *Super_Man_t {
	return &Super_Man_t{table: make(map[uint64][]*Super_Gate_t), started: time.Now()}
}
func Super_ManStop(p *Super_Man_t) {
	if p != nil {
		p.table = nil
		p.Gates = nil
		p.Inputs = nil
	}
}
func Super_First(p *Super_Man_t, inputs int32) {
	if inputs < 1 || inputs > 6 {
		panic("super: input count must be 1..6")
	}
	p.Options.Inputs = int(inputs)
	p.Levels = 0
	p.Inputs = nil
	p.table = make(map[uint64][]*Super_Gate_t)
	for variable := 0; variable < int(inputs); variable++ {
		g := &Super_Gate_t{Variable: variable, Delays: make([]float32, inputs), serial: p.nextSerial}
		p.nextSerial++
		for pin := range g.Delays {
			g.Delays[pin] = SUPER_NO_VAR
		}
		g.Delays[variable] = 0
		for mint := 0; mint < 1<<uint(inputs); mint++ {
			if mint&(1<<uint(variable)) != 0 {
				g.Truth |= 1 << uint(mint)
			}
		}
		p.Inputs = append(p.Inputs, g)
		Super_AddGateToTable(p, g)
	}
	p.Gates = append([]*Super_Gate_t(nil), p.Inputs...)
}
func Super_AddGateToTable(p *Super_Man_t, g *Super_Gate_t) {
	p.table[g.Truth] = append([]*Super_Gate_t{g}, p.table[g.Truth]...)
	p.Stats.Added++
}
func Super_TranferGatesToArray(p *Super_Man_t) {
	p.Gates = nil
	for _, group := range p.table {
		p.Gates = append(p.Gates, group...)
	}
	sort.SliceStable(p.Gates, func(i, j int) bool { return p.Gates[i].serial < p.Gates[j].serial })
}

// CompareGates updates a Pareto frontier only after the candidate is known to
// survive every existing point. Removed objects remain alive for parent DAGs.
func Super_CompareGates(p *Super_Man_t, truth []uint32, area float32, delays []float32, pins int32) int32 {
	if len(truth) < 2 || int(pins) != p.Options.Inputs || int(pins) > len(delays) {
		panic("super: invalid candidate dimensions")
	}
	key := (uint64(truth[0]) | uint64(truth[1])<<32) & truthMask(p.Options.Inputs)
	if key == 0 || key == truthMask(p.Options.Inputs) {
		return 0
	}
	group := p.table[key]
	remove := make([]bool, len(group))
	for index, g := range group {
		p.Stats.Lookups++
		oldBetter, newBetter := false, false
		if g.Area+SUPER_EPSILON < area {
			oldBetter = true
		} else if area+SUPER_EPSILON < g.Area {
			newBetter = true
		}
		for pin := 0; pin < int(pins); pin++ {
			a, b := g.Delays[pin], delays[pin]
			if a == SUPER_NO_VAR || b == SUPER_NO_VAR {
				continue
			}
			if a+SUPER_EPSILON < b {
				oldBetter = true
			} else if b+SUPER_EPSILON < a {
				newBetter = true
			}
		}
		if !newBetter {
			return 0
		}
		remove[index] = !oldBetter
	}
	kept := group[:0]
	for i, g := range group {
		if remove[i] {
			p.Stats.Removed++
		} else {
			kept = append(kept, g)
		}
	}
	for index := len(kept); index < len(group); index++ {
		group[index] = nil
	}
	if len(kept) == 0 {
		delete(p.table, key)
	} else {
		p.table[key] = kept
	}
	return 1
}
func Super_CreateGateNew(p *Super_Man_t, root *Mio_Gate_t, children []*Super_Gate_t, count int32, truth []uint32, area float32, delays []float32, maximum float32, pins int32) *Super_Gate_t {
	g := &Super_Gate_t{Root: root, Variable: -1, Truth: (uint64(truth[0]) | uint64(truth[1])<<32) & truthMask(p.Options.Inputs), Fanins: append([]*Super_Gate_t(nil), children[:count]...), Area: area, Delays: append([]float32(nil), delays[:pins]...), MaxDelay: maximum, serial: p.nextSerial}
	p.nextSerial++
	return g
}
func (p *Super_Man_t) stopped() bool {
	if p.Options.Context != nil && p.Options.Context.Err() != nil {
		return true
	}
	if p.Options.TimeLimit > 0 && time.Since(p.started) >= p.Options.TimeLimit {
		p.Stats.TimedOut = true
		return true
	}
	return p.Stats.LimitReached
}
func Super_CheckTimeout(p *Super_Man_t) int32 {
	if p.stopped() {
		return 1
	}
	p.Stats.Tried++
	return 0
}
func Super_Compute(p *Super_Man_t, roots []*Mio_Gate_t, count, maxGates, skipInverters int32) *Super_Man_t {
	Super_TranferGatesToArray(p)
	frontier := append([]*Super_Gate_t(nil), p.Gates...)
	sort.SliceStable(frontier, func(i, j int) bool {
		if frontier[i].Area != frontier[j].Area {
			return frontier[i].Area < frontier[j].Area
		}
		return frontier[i].serial < frontier[j].serial
	})
	for _, root := range roots[:count] {
		pins := int(mio.Mio_GateReadPinNum(root))
		if pins < 1 || pins > 6 {
			continue
		}
		if !p.Options.Exhaustive && p.Levels >= 1 && pins >= 4 {
			continue
		}
		candidates := make([]*Super_Gate_t, 0, len(frontier))
		for _, g := range frontier {
			if !p.Options.Exhaustive && p.Levels >= 1 && len(g.Fanins) >= 3 {
				continue
			}
			candidates = append(candidates, g)
		}
		chosen := make([]*Super_Gate_t, pins)
		used := make([]bool, len(candidates))
		rootArea := float32(mio.Mio_GateReadArea(root))
		var enumerate func(int, float32)
		enumerate = func(pin int, area float32) {
			if p.stopped() {
				return
			}
			if pin < pins {
				for index, g := range candidates {
					if used[index] || pins == 1 && skipInverters != 0 && g.Root == nil {
						continue
					}
					nextArea := area + g.Area
					if !finite(nextArea) {
						break
					}
					if p.Options.AreaLimit > 0 && nextArea > p.Options.AreaLimit {
						break
					}
					chosen[pin] = g
					used[index] = true
					enumerate(pin+1, nextArea)
					used[index] = false
					if p.stopped() {
						return
					}
				}
				return
			}
			p.Stats.Tried++
			truths := make([][2]uint32, pins)
			delaysIn := make([][]float32, pins)
			for i, g := range chosen {
				truths[i] = [2]uint32{uint32(g.Truth), uint32(g.Truth >> 32)}
				delaysIn[i] = g.Delays
			}
			delays := make([]float32, p.Options.Inputs)
			var maximum float32
			mio.Mio_DeriveGateDelays(root, delaysIn, int32(pins), int32(p.Options.Inputs), SUPER_NO_VAR, delays, &maximum)
			if !finite(maximum) || p.Options.DelayLimit > 0 && maximum > p.Options.DelayLimit {
				return
			}
			truth := make([]uint32, 2)
			mio.Mio_DeriveTruthTable(root, truths, int32(pins), int32(p.Options.Inputs), truth)
			key := (uint64(truth[0]) | uint64(truth[1])<<32) & truthMask(p.Options.Inputs)
			truth[0], truth[1] = uint32(key), uint32(key>>32)
			if Super_CompareGates(p, truth, area, delays, int32(p.Options.Inputs)) == 0 {
				return
			}
			g := Super_CreateGateNew(p, root, chosen, int32(pins), truth, area, delays, maximum, int32(p.Options.Inputs))
			Super_AddGateToTable(p, g)
			if maxGates > 0 && len(p.table) >= int(maxGates) {
				p.Stats.LimitReached = true
			}
		}
		if p.Options.AreaLimit <= 0 || rootArea <= p.Options.AreaLimit {
			enumerate(0, rootArea)
		}
		if p.stopped() {
			break
		}
	}
	return p
}
func Generate(lib *Mio_Library_t, options Options) (*Super_Man_t, error) {
	if lib == nil || options.Inputs < 1 || options.Inputs > 6 || options.Levels < 0 || options.MaxGates < 0 || int64(options.MaxGates) > 1<<31-1 || options.MaxGates > 0 && options.MaxGates < options.Inputs || !finite(options.DelayLimit) || !finite(options.AreaLimit) || options.DelayLimit < 0 || options.AreaLimit < 0 || options.TimeLimit < 0 {
		return nil, fmt.Errorf("super: invalid library or generation limits")
	}
	var count int32
	roots := mio.Mio_CollectRoots(lib, int32(options.Inputs), options.DelayLimit, 0, &count, 0)
	if len(roots) == 0 && options.Levels > 0 {
		return nil, fmt.Errorf("super: no root gates satisfy input/delay limits")
	}
	for _, root := range roots {
		if !finite(float32(mio.Mio_GateReadArea(root))) || mio.Mio_GateReadArea(root) < 0 {
			return nil, fmt.Errorf("super: invalid root area")
		}
		for pin := int32(0); pin < mio.Mio_GateReadPinNum(root); pin++ {
			delay := mio.Mio_GateReadPinDelay(root, pin)
			if !finite(delay) || delay < 0 {
				return nil, fmt.Errorf("super: invalid root pin delay")
			}
		}
	}
	p := Super_ManStart()
	p.Options = options
	p.Name = mio.Mio_LibraryReadName(lib)
	p.started = time.Now()
	Super_First(p, int32(options.Inputs))
	if options.MaxGates > 0 && len(p.table) >= options.MaxGates {
		p.Stats.LimitReached = true
	}
	for level := 0; level < options.Levels && !p.stopped(); level++ {
		skip := int32(0)
		if options.SkipInputInverters {
			skip = 1
		}
		Super_Compute(p, roots, count, int32(options.MaxGates), skip)
		p.Levels++
		if options.Progress != nil {
			fmt.Fprintf(options.Progress, "Level %d: tried=%d kept=%d\n", p.Levels, p.Stats.Tried, p.Stats.Added-p.Stats.Removed)
		}
	}
	p.Stats.Elapsed = time.Since(p.started)
	Super_TranferGatesToArray(p)
	if options.Context != nil && options.Context.Err() != nil {
		return p, options.Context.Err()
	}
	return p, nil
}
func legacyOptions(inputs, levels, maxGates int32, delay, area float32, seconds, skip, verbose int32) Options {
	o := Options{Inputs: int(inputs), Levels: int(levels), MaxGates: int(maxGates), DelayLimit: delay, AreaLimit: area, TimeLimit: time.Duration(seconds) * time.Second, SkipInputInverters: skip != 0}
	if verbose != 0 {
		o.Progress = os.Stdout
	}
	return o
}
func Super_PrecomputeStr(lib *Mio_Library_t, inputs, levels, maxGates int32, delay, area float32, seconds, skip, verbose int32) *Vec_Str_t {
	p, err := Generate(lib, legacyOptions(inputs, levels, maxGates, delay, area, seconds, skip, verbose))
	if err != nil {
		return nil
	}
	return Super_Write(p)
}
func Super_Precompute(lib *Mio_Library_t, inputs, levels, maxGates int32, delay, area float32, seconds, skip, verbose int32, path string) error {
	p, err := Generate(lib, legacyOptions(inputs, levels, maxGates, delay, area, seconds, skip, verbose))
	if err != nil {
		return err
	}
	data, err := p.MarshalTree()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
