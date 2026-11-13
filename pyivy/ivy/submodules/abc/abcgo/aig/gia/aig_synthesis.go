package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	dar "github.com/glycerine/abcgo/opt/dar/core"
	"github.com/glycerine/abcgo/proof/dch"
	"os"
)

func Gia_ManCompress2Result(p *Gia_Man_t, updateLevel, verbose bool) (*Gia_Man_t, error) {
	if p == nil {
		return nil, fmt.Errorf("gia: nil compression graph")
	}
	if p.pManTime != nil && p.vLevels == nil {
		Gia_ManLevelWithBoxes(p)
	}
	source := Gia_ManToAig(p, 0)
	defer aig.Aig_ManStop(source)
	optimized, err := dar.Dar_ManCompress2(source, true, updateLevel, true, false, verbose)
	if err != nil {
		return nil, err
	}
	defer aig.Aig_ManStop(optimized)
	result := Gia_ManFromAig(optimized)
	Gia_ManTransferTiming(result, p)
	return result, nil
}
func Gia_ManCompress2(p *Gia_Man_t, updateLevel, verbose int32) *Gia_Man_t {
	result, err := Gia_ManCompress2Result(p, updateLevel != 0, verbose != 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return result
}

func Gia_ManPerformDchResult(p *Gia_Man_t, options *dar.ChoiceOptions) (*Gia_Man_t, error) {
	if p == nil || options == nil {
		return nil, fmt.Errorf("gia: missing graph or choice options")
	}
	if p.pManTime != nil && p.vLevels == nil {
		Gia_ManLevelWithBoxes(p)
	}
	source := Gia_ManDup(p)
	defer Gia_ManStop(source)
	converted := Gia_ManToAig(source, 0)
	defer aig.Aig_ManStop(converted)
	optimized, err := dar.Dar_ManChoiceNew(converted, options)
	if err != nil {
		return nil, err
	}
	defer aig.Aig_ManStop(optimized)
	result := Gia_ManFromAigChoices(optimized)
	Gia_ManTransferTiming(result, p)
	return result, nil
}
func Gia_ManPerformDch(p *Gia_Man_t, parameters any) *Gia_Man_t {
	var options *dar.ChoiceOptions
	var params *dch.Params
	switch value := parameters.(type) {
	case *dar.ChoiceOptions:
		options = value
	case *dch.Params:
		params = value
	case nil:
		defaults := dch.DefaultParams()
		params = &defaults
	default:
		fmt.Fprintln(os.Stderr, "gia: unsupported choice parameter type")
		return nil
	}
	if params != nil {
		if params.UseGia || params.UseNew {
			fmt.Fprintln(os.Stderr, "gia: requested CEC choice engine is not yet ported; provide an explicit ChoiceOptions engine")
			return nil
		}
		options = &dar.ChoiceOptions{Balance: true, UpdateLevel: true, Power: params.Power, Light: params.LightSynth, Engine: dch.Engine{Params: params}}
		if params.Verbose {
			options.Progress = os.Stdout
		}
	}
	result, err := Gia_ManPerformDchResult(p, options)
	if params != nil && options != nil {
		params.SynthDuration = options.SynthesisDuration
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return result
}
