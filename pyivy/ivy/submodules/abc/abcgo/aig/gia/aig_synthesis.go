package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/opt/dar"
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
