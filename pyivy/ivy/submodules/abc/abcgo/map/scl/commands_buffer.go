package scl

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
)

// TopologicalTimingNetwork preserves the concrete host graph while assigning
// objects in topological order; a plain clone is insufficient for topo.
type TopologicalTimingNetwork interface{ SclCloneTopological() (TimingNetwork, error) }

func Scl_CommandTopo(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "v", "")
	if e != nil || len(o.args) != 0 {
		return sclCommandError(f, fmt.Errorf("usage: topo [-vh]"))
	}
	net, ok := f.Network().(TopologicalTimingNetwork)
	if !ok {
		return sclCommandError(f, fmt.Errorf("current network does not provide a topological clone"))
	}
	result, e := net.SclCloneTopological()
	if e != nil {
		return sclCommandError(f, e)
	}
	if result == nil {
		return sclCommandError(f, fmt.Errorf("topological clone returned no network"))
	}
	f.ReplaceCurrentNetwork(result)
	return 0
}
func Scl_CommandUnBuffer(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "iv", "")
	if e != nil || len(o.args) != 0 {
		return sclCommandError(f, fmt.Errorf("usage: unbuffer [-ivh]"))
	}
	net, ok := f.Network().(TimingNetwork)
	if !ok {
		return sclCommandError(f, fmt.Errorf("current network does not provide standard-cell topology"))
	}
	var result TimingNetwork
	if o.flags['i'] {
		result, e = Abc_SclUnBufferPhase(net, o.flags['v'])
	} else {
		result, e = Abc_SclUnBufferPerform(net, o.flags['v'])
	}
	if e != nil {
		return sclCommandError(f, e)
	}
	if result == nil {
		return sclCommandError(f, fmt.Errorf("unbuffer returned no network"))
	}
	f.ReplaceCurrentNetwork(result)
	return 0
}
func Scl_CommandBuffer(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "sbpcvw", "GSN")
	if e != nil || len(o.args) != 0 {
		return sclCommandError(f, fmt.Errorf("usage: buffer [-GSN num] [-sbpcvwh]"))
	}
	lib, net, e := sclMappedNetwork(f)
	if e != nil {
		return sclCommandError(f, e)
	}
	if !Abc_SclHasDelayInfo(lib) {
		return sclCommandError(f, fmt.Errorf("library delay info is not available"))
	}
	pars := SC_BusPars{GainRatio: 300, Slew: int(Abc_SclComputeAverageSlew(lib)), Degree: 10, SizeOnly: o.flags['s'], AddBuffers: !o.flags['b'], BufferPIs: o.flags['p'], UseWireLoads: o.flags['c'], Verbose: o.flags['v'], VeryVerbose: o.flags['w']}
	for key, dst := range map[byte]*int{'G': &pars.GainRatio, 'S': &pars.Slew, 'N': &pars.Degree} {
		if v, ok := o.numbers[key]; ok {
			if v != float32(int(v)) {
				return sclCommandError(f, fmt.Errorf("-%c requires an integer", key))
			}
			*dst = int(v)
		}
	}
	if !pars.SizeOnly && !pars.AddBuffers && !hasPhases(net) {
		return sclCommandError(f, fmt.Errorf("fanin phase information is not available"))
	}
	result, e := Abc_SclBufferingPerform(net, lib, &pars)
	if e != nil {
		return sclCommandError(f, e)
	}
	if result == nil {
		return sclCommandError(f, fmt.Errorf("buffering returned no network"))
	}
	f.ReplaceCurrentNetwork(result)
	return 0
}

// BufferOld remains an unregistered compatibility entry point, as in scl.c.
func Scl_CommandBufferOld(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "aixpdv", "NMR")
	if e != nil || len(o.args) != 0 {
		return sclCommandError(f, fmt.Errorf("usage: _buffer [-NMR num] [-aixpdvh]"))
	}
	lib, net, e := sclMappedNetwork(f)
	if e != nil {
		return sclCommandError(f, e)
	}
	if !Abc_SclHasDelayInfo(lib) {
		return sclCommandError(f, fmt.Errorf("library delay info is not available"))
	}
	minimum, maximum, root := 6, 14, 0
	for key, dst := range map[byte]*int{'N': &minimum, 'M': &maximum, 'R': &root} {
		if v, ok := o.numbers[key]; ok {
			if v != float32(int(v)) {
				return sclCommandError(f, fmt.Errorf("-%c requires an integer", key))
			}
			*dst = int(v)
		}
	}
	var result TimingNetwork
	if o.flags['i'] {
		if !hasPhases(net) {
			return sclCommandError(f, fmt.Errorf("fanin phase information is not available"))
		}
		result, e = Abc_SclBufferPhase(net, o.flags['v'])
	} else if o.flags['a'] {
		result, e = Abc_SclPerformBuffering(net, root, maximum, o.flags['x'], o.flags['v'])
	} else {
		result, e = Abc_SclBufPerform(net, minimum, maximum, o.flags['p'], o.flags['d'], o.flags['v'])
	}
	if e != nil {
		return sclCommandError(f, e)
	}
	if result == nil {
		return sclCommandError(f, fmt.Errorf("buffering returned no network"))
	}
	f.ReplaceCurrentNetwork(result)
	return 0
}
