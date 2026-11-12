package fxu

import "io"

// Fxu_Data_t carries binary SOP/fanin vectors and original FXU controls. Counts
// of old/new nodes derive from the vector lengths, rather than cached fields.
type Fxu_Data_t struct {
	OnlyS, OnlyD, Use0, UseCompl, Verbose, VeryVerbose        bool
	NodesExt, SingleMax, PairsMax, WeightMin, LiteralCountMax int32
	CanonicalDivisors                                         bool // retained source setting; original FXU does not consult it
	SOPs                                                      []string
	Fanins                                                    [][]int32
	NewSOPs                                                   []string
	NewFanins                                                 [][]int32
	Log                                                       io.Writer
	LastError                                                 error
}

func Abc_NtkSetDefaultFxParams(p *Fxu_Data_t) {
	*p = Fxu_Data_t{NodesExt: 1000000, SingleMax: 20000, PairsMax: 30000, LiteralCountMax: 4, UseCompl: true}
}
func NewParams() *Fxu_Data_t              { p := new(Fxu_Data_t); Abc_NtkSetDefaultFxParams(p); return p }
func (p *Fxu_Data_t) OldNodeCount() int32 { return int32(len(p.SOPs)) }
func (p *Fxu_Data_t) NewNodeCount() int32 {
	if len(p.NewSOPs) < len(p.SOPs) {
		return 0
	}
	return int32(len(p.NewSOPs) - len(p.SOPs))
}

// FXUNetwork keeps the algorithm independent of ABC's concrete graph package.
// A network adapter owns conversion to SCC-free SOPs and graph reconstruction.
type FXUNetwork interface {
	FXUCovers() ([]string, [][]int32, error)
	ReplaceFXUCovers([]string, [][]int32) error
}

func Abc_NtkFxuFreeInfo(p *Fxu_Data_t) {
	p.SOPs = nil
	p.Fanins = nil
	p.NewSOPs = nil
	p.NewFanins = nil
	p.LastError = nil
}
func Abc_NtkFastExtract(network FXUNetwork, p *Fxu_Data_t) int32 {
	sops, fans, err := network.FXUCovers()
	if err != nil {
		p.LastError = err
		return 0
	}
	p.SOPs = sops
	p.Fanins = fans
	count, err := Extract(p)
	if err != nil {
		p.LastError = err
		return 0
	}
	if count == 0 {
		return 0
	}
	if err := network.ReplaceFXUCovers(p.NewSOPs, p.NewFanins); err != nil {
		p.LastError = err
		return 0
	}
	return 1
}
