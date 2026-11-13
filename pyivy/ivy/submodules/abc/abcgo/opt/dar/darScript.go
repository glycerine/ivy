package dar

import (
	"context"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"io"
	"os"
)

// SATSynthesizer implements aig.PartSATSynthesizer without an import cycle.
// Each call returns an independent graph with unchanged CI/CO ordering.
type SATSynthesizer struct {
	Context  context.Context
	Balance  bool
	Progress io.Writer
}

func (s SATSynthesizer) SynthesizePart(g *aig.Graph) (*aig.Graph, error) {
	return RewriteForSAT(s.Context, g, s.Balance, s.Progress)
}

// RewriteForSAT follows ABC's rwsat sequence: optional balance, rewrite,
// refactor, optional balance, rewrite, compacting between each rewriting pass.
func RewriteForSAT(ctx context.Context, g *aig.Graph, balance bool, progress io.Writer) (*aig.Graph, error) {
	if g == nil {
		return nil, fmt.Errorf("dar: nil graph")
	}
	if err := g.Check(); err != nil {
		return nil, err
	}
	result := aig.Aig_ManDupDfs(g)
	report := func(stage string) error {
		if progress != nil {
			err := writeProgress(progress, "%s: %d nodes\n", stage, result.NodeCount())
			return err
		}
		if ctx != nil {
			return ctx.Err()
		}
		return nil
	}
	if err := report("Starting"); err != nil {
		return nil, err
	}
	rwr := DefaultRewriteParams()
	rwr.Context = ctx
	rwr.UpdateLevel = false
	ref := DefaultRefactorParams()
	ref.Context = ctx
	ref.UpdateLevel = false
	if balance {
		var err error
		result, err = Balance(ctx, result, false)
		if err != nil {
			return nil, err
		}
		if err = report("Balance"); err != nil {
			return nil, err
		}
	}
	if err := Dar_ManRewrite(result, &rwr); err != nil {
		return nil, err
	}
	result = aig.Aig_ManDupDfs(result)
	if err := report("Rewrite"); err != nil {
		return nil, err
	}
	if err := Dar_ManRefactor(result, &ref); err != nil {
		return nil, err
	}
	result = aig.Aig_ManDupDfs(result)
	if err := report("Refactor"); err != nil {
		return nil, err
	}
	if balance {
		var err error
		result, err = Balance(ctx, result, false)
		if err != nil {
			return nil, err
		}
		if err = report("Balance"); err != nil {
			return nil, err
		}
	}
	if err := Dar_ManRewrite(result, &rwr); err != nil {
		return nil, err
	}
	result = aig.Aig_ManDupDfs(result)
	if err := report("Rewrite"); err != nil {
		return nil, err
	}
	return result, result.Check()
}
func Rwsat(ctx context.Context, g *aig.Graph, balance, verbose bool) (*aig.Graph, error) {
	var progress io.Writer
	if verbose {
		progress = os.Stdout
	}
	return RewriteForSAT(ctx, g, balance, progress)
}
func Dar_ManRwsat(g *aig.Graph, balance, verbose bool) (*aig.Graph, error) {
	ctx := context.Background()
	if g != nil && !g.Time2Quit.IsZero() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, g.Time2Quit)
		defer cancel()
	}
	return Rwsat(ctx, g, balance, verbose)
}
func Dar_ManRewriteDefault(g *aig.Graph) (*aig.Graph, error) {
	if g == nil {
		return nil, fmt.Errorf("dar: nil graph")
	}
	result := aig.Aig_ManDupDfs(g)
	if err := Dar_ManRewrite(result, nil); err != nil {
		return nil, err
	}
	return aig.Aig_ManDupDfs(result), nil
}

// CompressOptions controls the original compress/compress2 orchestration.
type CompressOptions struct {
	Context                                    context.Context
	Balance, UpdateLevel, Fanout, Power, Light bool
	Progress                                   io.Writer
}
type synthesisStage struct {
	kind               string
	zeros, ignoreLevel bool
}

func compressStages(g *aig.Graph, p CompressOptions, stages []synthesisStage) (*aig.Graph, error) {
	if g == nil {
		return nil, fmt.Errorf("dar: nil graph")
	}
	if err := g.Check(); err != nil {
		return nil, err
	}
	result := aig.Aig_ManDupDfs(g)
	for _, stage := range stages {
		if p.Context != nil {
			if err := p.Context.Err(); err != nil {
				return nil, err
			}
		}
		var err error
		switch stage.kind {
		case "balance":
			result, err = Balance(p.Context, result, p.UpdateLevel)
		case "rewrite":
			r := DefaultRewriteParams()
			r.Context = p.Context
			r.UpdateLevel = p.UpdateLevel && !stage.ignoreLevel
			r.Fanout = p.Fanout
			r.Power = p.Power
			r.UseZeros = stage.zeros
			err = Dar_ManRewrite(result, &r)
		case "refactor":
			r := DefaultRefactorParams()
			r.Context = p.Context
			r.UpdateLevel = p.UpdateLevel
			r.UseZeros = stage.zeros
			err = Dar_ManRefactor(result, &r)
		}
		if err != nil {
			return nil, err
		}
		result = aig.Aig_ManDupDfs(result)
		if p.Progress != nil {
			if err = writeProgress(p.Progress, "%s: %d nodes\n", stage.kind, result.NodeCount()); err != nil {
				return nil, err
			}
		}
	}
	return result, result.Check()
}
func Compress(g *aig.Graph, p CompressOptions) (*aig.Graph, error) {
	stages := []synthesisStage{{kind: "rewrite"}, {kind: "refactor"}}
	if p.Balance {
		stages = append(stages, synthesisStage{kind: "balance"})
	}
	stages = append(stages, synthesisStage{kind: "rewrite", zeros: true})
	return compressStages(g, p, stages)
}
func Compress2(g *aig.Graph, p CompressOptions) (*aig.Graph, error) { return compress2(g, p, false) }
func compress2(g *aig.Graph, p CompressOptions, newVariant bool) (*aig.Graph, error) {
	var stages []synthesisStage
	if !p.Light {
		stages = append(stages, synthesisStage{kind: "rewrite", ignoreLevel: !newVariant}, synthesisStage{kind: "refactor"})
	}
	stages = append(stages, synthesisStage{kind: "balance"})
	if !p.Light {
		stages = append(stages, synthesisStage{kind: "rewrite"})
	}
	stages = append(stages, synthesisStage{kind: "rewrite", zeros: true})
	if !p.Light && p.Balance {
		stages = append(stages, synthesisStage{kind: "balance"})
	}
	stages = append(stages, synthesisStage{kind: "refactor", zeros: true})
	if !p.Light {
		stages = append(stages, synthesisStage{kind: "rewrite", zeros: true})
	}
	if p.Balance {
		stages = append(stages, synthesisStage{kind: "balance"})
	}
	return compressStages(g, p, stages)
}
func Dar_ManCompress(g *aig.Graph, balance, updateLevel, power, verbose bool) (*aig.Graph, error) {
	p := CompressOptions{Balance: balance, UpdateLevel: updateLevel, Power: power, Fanout: true}
	if verbose {
		p.Progress = os.Stdout
	}
	return Compress(g, p)
}
func Dar_ManCompress2(g *aig.Graph, balance, updateLevel, fanout, power, verbose bool) (*aig.Graph, error) {
	p := CompressOptions{Balance: balance, UpdateLevel: updateLevel, Fanout: fanout, Power: power}
	if verbose {
		p.Progress = os.Stdout
	}
	return Compress2(g, p)
}
func Dar_NewCompress(g *aig.Graph, balance, updateLevel, power, verbose bool) (*aig.Graph, error) {
	return Dar_ManCompress(g, balance, updateLevel, power, verbose)
}
func Dar_NewCompress2(g *aig.Graph, balance, updateLevel, fanout, power, light, verbose bool) (*aig.Graph, error) {
	p := CompressOptions{Balance: balance, UpdateLevel: updateLevel, Fanout: fanout, Power: power, Light: light}
	if verbose {
		p.Progress = os.Stdout
	}
	return compress2(g, p, true)
}

// Dar_ManChoiceSynthesis preserves the original, compress, and compress2
// versions as independent graphs with identical ordered interfaces.
func Dar_ManChoiceSynthesis(g *aig.Graph, balance, updateLevel, power, verbose bool) ([]*aig.Graph, error) {
	if g == nil {
		return nil, fmt.Errorf("dar: nil graph")
	}
	original := aig.Aig_ManDupDfs(g)
	first, err := Dar_ManCompress(original, balance, updateLevel, power, verbose)
	if err != nil {
		return nil, err
	}
	second, err := Dar_ManCompress2(first, balance, updateLevel, true, power, verbose)
	if err != nil {
		return nil, err
	}
	return []*aig.Graph{original, first, second}, nil
}
func Dar_ManChoice(g *aig.Graph, balance, updateLevel, construct bool, conflictMax, levelMax int32, verbose bool) (*aig.Graph, error) {
	versions, err := Dar_ManChoiceSynthesis(g, balance, updateLevel, false, verbose)
	if err != nil {
		return nil, err
	}
	verbosity := int32(0)
	if verbose {
		verbosity = 1
	}
	var result *aig.Graph
	if construct {
		result = aig.Aig_ManChoiceConstructive(versions, verbosity)
	} else {
		versions[0], versions[len(versions)-1] = versions[len(versions)-1], versions[0]
		result = aig.Aig_ManChoicePartitioned(versions, 300, conflictMax, levelMax, verbosity)
	}
	if err := result.Check(); err != nil {
		return nil, err
	}
	return result, nil
}

// Dar_NewChoiceSynthesisGuard detects the source's high-fanout threshold used
// to disable costly level maintenance during choice synthesis.
func Dar_NewChoiceSynthesisGuard(g *aig.Graph) bool {
	count := 0
	for _, e := range g.LiveEdges() {
		n := g.Node(e)
		if n.Kind == aig.And || n.Kind == aig.ExclusiveOr {
			if n.Refs > 1000 {
				count += n.Refs / 1000
			}
		}
	}
	return count > 10
}

func writeProgress(w io.Writer, format string, args ...any) error {
	text := fmt.Sprintf(format, args...)
	n, err := io.WriteString(w, text)
	if err == nil && n != len(text) {
		return io.ErrShortWrite
	}
	return err
}
