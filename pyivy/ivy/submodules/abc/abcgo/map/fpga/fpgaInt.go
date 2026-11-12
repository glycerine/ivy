// Package fpga implements ABC's cut-based variable-size LUT mapper.
package fpga

import (
	"context"
	"io"
	"time"
)

const (
	FPGA_MAX_LEAVES               = 6
	FPGA_MAX_LUTSIZE              = 10
	FPGA_FLOAT_LARGE      float32 = 1e20
	FPGA_FLOAT_SMALL      float32 = 1e-20
	FPGA_INT_LARGE                = 10000000
	FPGA_NO_VAR           float32 = -9999
	FPGA_CUTS_MAX_COMPUTE         = 2000
	FPGA_CUTS_MAX_USE             = 1000
)

func FPGA_MASK(n int32) uint32 {
	if n <= 0 {
		return 0
	}
	if n >= 32 {
		return ^uint32(0)
	}
	return 1<<uint(n) - 1
}
func FPGA_NUM_BYTES(n int32) int32 { return (n + 15) / 16 * 16 }
func truthMask(n int) uint64 {
	if n == 6 {
		return ^uint64(0)
	}
	return 1<<uint(1<<uint(n)) - 1
}
func boolInt(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

type Fpga_LutLib_t struct {
	Name              string
	LutMax            int32
	VariablePinDelays bool
	Areas             []float32
	Delays            [][]float32
}
type nodePair struct{ one, two *Fpga_Node_t }
type Fpga_Man_t struct {
	Inputs, Outputs                               []*Fpga_Node_t
	Const1                                        *Fpga_Node_t
	Nodes                                         []*Fpga_Node_t
	Ands, Mapping                                 *Fpga_NodeVec_t
	Library                                       *Fpga_LutLib_t
	Name                                          string
	OutputNames                                   []string
	InputArrivals                                 []float32
	AreaRecovery, Verbose, Switching, LatchPaths  bool
	Latches, ChoiceNodes, Choices                 int32
	DelayTarget, DelayLimit, AreaLimit, TimeLimit float32
	RequiredGlobal, AreaGlobal, Epsilon           float32
	CutLimit                                      int
	Context                                       context.Context
	Output                                        io.Writer
	LastError                                     error
	Stats                                         MappingStatistics
	table                                         map[nodePair]*Fpga_Node_t
}
type MappingStatistics struct {
	Cuts, Nodes, LUTs                       int
	Area, Delay, Switching                  float32
	CutTime, MatchTime, RecoveryTime, Total time.Duration
}
type Fpga_Node_t struct {
	Num, Level, Refs                      int32
	One, Two, NextE, Repr, NextLevel      *Fpga_Node_t
	Phase                                 bool
	Required, EstimatedFanouts, Switching float32
	Cuts, Best, Old                       *Fpga_Cut_t
	Data0, Data1                          any
	Value                                 uint8
	Used                                  bool
	TravID                                int32
	LValue                                int32
	Latches1, Latches2                    int16
	fanouts                               []*Fpga_Node_t
	manager                               *Fpga_Man_t
	regular, neg                          *Fpga_Node_t
}
type Fpga_Cut_t struct {
	Root                 *Fpga_Node_t
	Leaves               []*Fpga_Node_t
	One, Two             *Fpga_Cut_t
	Next                 *Fpga_Cut_t
	Truth                uint64
	Phase, Mark          bool
	Signature            uint32
	Level, Arrival, Flow float32
	PinOrder             []int
	regular, neg         *Fpga_Cut_t
}
type Fpga_NodeVec_t struct{ Nodes []*Fpga_Node_t }
type Fpga_CutTable_t struct {
	Cuts    map[string]*Fpga_Cut_t
	Array   []*Fpga_Cut_t
	Manager *Fpga_Man_t
}

func Fpga_FloatMoreThan(p *Fpga_Man_t, a, b float32) bool { return a > b+p.Epsilon }
func Fpga_FloatLessThan(p *Fpga_Man_t, a, b float32) bool { return a < b-p.Epsilon }
func Fpga_FloatEqual(p *Fpga_Man_t, a, b float32) bool    { return a >= b-p.Epsilon && a <= b+p.Epsilon }
