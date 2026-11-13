package gia

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/dch"
)

// Gia_ManChoiceMiterFromAigs implements DAR's conversion of three ordered
// synthesis snapshots to GIA followed by Gia_ManChoiceMiter. The caller owns
// the returned miter; the input AIGs remain owned by their caller.
func Gia_ManChoiceMiterFromAigs(versions []*aig.Graph) (*Gia_Man_t, error) {
	if len(versions) != 3 {
		return nil, fmt.Errorf("gia: DAR choice path requires three snapshots")
	}
	snapshots := make([]*Gia_Man_t, 0, len(versions))
	defer func() {
		for _, graph := range snapshots {
			Gia_ManStop(graph)
		}
	}()
	for i, graph := range versions {
		if graph == nil {
			return nil, fmt.Errorf("gia: nil DAR snapshot %d", i)
		}
		if err := graph.Check(); err != nil {
			return nil, fmt.Errorf("gia: DAR snapshot %d: %w", i, err)
		}
		if i != 0 && (graph.InputCount() != versions[0].InputCount() || graph.OutputCount() != versions[0].OutputCount() || aig.Aig_ManRegNum(graph) != aig.Aig_ManRegNum(versions[0])) {
			return nil, fmt.Errorf("gia: incompatible DAR snapshot interfaces")
		}
		snapshots = append(snapshots, Gia_ManFromAig(graph))
	}
	return Gia_ManChoiceMiter(snapshots), nil
}

// giaSourceDCHChoiceEngine is the source's default branch: GIA miter,
// stride-three AIG export, then DCH. CEC branches are supplied explicitly by
// the higher proof/cec layer to avoid a GIA-to-CEC import cycle.
type giaSourceDCHChoiceEngine struct{ params *dch.Params }

func (e giaSourceDCHChoiceEngine) ComputeChoices(versions []*aig.Graph) (*aig.Graph, error) {
	if e.params.UseGia || e.params.UseNew {
		return nil, fmt.Errorf("gia: CEC choice mode requires an explicit CEC engine")
	}
	miter, err := Gia_ManChoiceMiterFromAigs(versions)
	if err != nil {
		return nil, err
	}
	defer Gia_ManStop(miter)
	input := Gia_ManToAigSkip(miter, 3)
	defer aig.Aig_ManStop(input)
	return dch.ComputeChoices(input, e.params)
}
