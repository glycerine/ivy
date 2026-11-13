// Package dar exposes the DAR optimization and choice APIs.
package dar

import (
	context "context"
	aig "github.com/glycerine/abcgo/aig/aig"
	core "github.com/glycerine/abcgo/opt/dar/core"
	"github.com/glycerine/abcgo/proof/cec"
	"github.com/glycerine/abcgo/proof/dch"
	io "io"
)

type ActivityOptions = core.ActivityOptions
type CanonicalTable = core.CanonicalTable
type ChoiceEngine = core.ChoiceEngine
type ChoiceOptions = core.ChoiceOptions
type CompressOptions = core.CompressOptions
type Cut = core.Cut
type CutEnumerator = core.CutEnumerator
type CutNode = core.CutNode
type CutRoot = core.CutRoot
type CutSlot = core.CutSlot
type Dar_Cut_t = core.Dar_Cut_t
type Dar_Man_t = core.Dar_Man_t
type Dar_RefPar_t = core.Dar_RefPar_t
type Dar_RwrPar_t = core.Dar_RwrPar_t
type Library = core.Library
type Manager = core.Manager
type NPNMatch = core.NPNMatch
type PartitionedChoiceEngine = core.PartitionedChoiceEngine
type PreparedLibrary = core.PreparedLibrary
type RefactorParams = core.RefactorParams
type RewriteParams = core.RewriteParams
type RewriteStats = core.RewriteStats
type SATSynthesizer = core.SATSynthesizer

func Balance(ctx context.Context, g *aig.Graph, updateLevel bool) (*aig.Graph, error) {
	return core.Balance(ctx, g, updateLevel)
}
func Compress(g *aig.Graph, p CompressOptions) (*aig.Graph, error)  { return core.Compress(g, p) }
func Compress2(g *aig.Graph, p CompressOptions) (*aig.Graph, error) { return core.Compress2(g, p) }
func Dar_ArrayAlloc(columns, rows, size int) [][]byte {
	return core.Dar_ArrayAlloc(columns, rows, size)
}
func Dar_BalanceBuildSuper(g *aig.Graph, leaves []aig.Edge) aig.Edge {
	return core.Dar_BalanceBuildSuper(g, leaves)
}
func Dar_BalanceCone(g *aig.Graph, root aig.Edge) []aig.Edge  { return core.Dar_BalanceCone(g, root) }
func Dar_BalancePrintStats(w io.Writer, g *aig.Graph) error   { return core.Dar_BalancePrintStats(w, g) }
func Dar_BalanceUniqify(leaves []aig.Edge) ([]aig.Edge, bool) { return core.Dar_BalanceUniqify(leaves) }
func Dar_CutCheckDominance(a, b Cut) bool                     { return core.Dar_CutCheckDominance(a, b) }
func Dar_CutMerge(a, b Cut) (Cut, bool)                       { return core.Dar_CutMerge(a, b) }
func Dar_CutSuppMinimize(c *Cut) bool                         { return core.Dar_CutSuppMinimize(c) }
func Dar_CutTruth(c, a, b Cut, complementA, complementB bool) uint16 {
	return core.Dar_CutTruth(c, a, b, complementA, complementB)
}
func Dar_Factorial(n int) int { return core.Dar_Factorial(n) }
func Dar_LibAddNode(lib *Library, first, second int, complementFirst, complementSecond bool) {
	core.Dar_LibAddNode(lib, first, second, complementFirst, complementSecond)
}
func Dar_LibAlloc(capacity int) *Library            { return core.Dar_LibAlloc(capacity) }
func Dar_LibPrepare(subgraphs int) *PreparedLibrary { return core.Dar_LibPrepare(subgraphs) }
func Dar_LibPrintStats(lib *Library) string         { return core.Dar_LibPrintStats(lib) }
func Dar_LibRead() *Library                         { return core.Dar_LibRead() }
func Dar_LibReadNodes() []uint32                    { return core.Dar_LibReadNodes() }
func Dar_LibReadOuts() []uint32                     { return core.Dar_LibReadOuts() }
func Dar_LibReadPrios() []uint32                    { return core.Dar_LibReadPrios() }
func Dar_LibReturnClass(truth uint32) int           { return core.Dar_LibReturnClass(truth) }
func Dar_LibStart()                                 { core.Dar_LibStart() }
func Dar_LibStop()                                  { core.Dar_LibStop() }
func Dar_ManBalance(g *aig.Graph, updateLevel bool) *aig.Graph {
	return core.Dar_ManBalance(g, updateLevel)
}
func Dar_ManBalanceXor(g *aig.Graph, exor, updateLevel, verbose bool) *aig.Graph {
	return core.Dar_ManBalanceXor(g, exor, updateLevel, verbose)
}
func Dar_ManChoice(g *aig.Graph, balance, updateLevel, construct bool, conflictMax, levelMax int32, verbose bool) (*aig.Graph, error) {
	return core.Dar_ManChoice(g, balance, updateLevel, construct, conflictMax, levelMax, verbose)
}
func Dar_ManChoiceNew(g *aig.Graph, p *ChoiceOptions) (*aig.Graph, error) {
	if p == nil || p.Engine != nil {
		return core.Dar_ManChoiceNew(g, p)
	}
	params := p.DchParams
	if params == nil {
		defaults := dch.DefaultParams()
		defaults.Power, defaults.LightSynth = p.Power, p.Light
		params = &defaults
	}
	return cec.Dar_ManChoiceNew(g, p, params)
}
func Dar_ManChoiceNewAig(g *aig.Graph, p *ChoiceOptions) (*aig.Graph, error) {
	return core.Dar_ManChoiceNewAig(g, p)
}
func Dar_ManChoiceSynthesis(g *aig.Graph, balance, updateLevel, power, verbose bool) ([]*aig.Graph, error) {
	return core.Dar_ManChoiceSynthesis(g, balance, updateLevel, power, verbose)
}
func Dar_ManChoiceWithFraig(g *aig.Graph, balance, updateLevel, construct bool, conflictMax, levelMax int32, verbose bool, fraig aig.PartitionFraigFunc) (*aig.Graph, error) {
	return core.Dar_ManChoiceWithFraig(g, balance, updateLevel, construct, conflictMax, levelMax, verbose, fraig)
}
func Dar_ManCompress(g *aig.Graph, balance, updateLevel, power, verbose bool) (*aig.Graph, error) {
	return core.Dar_ManCompress(g, balance, updateLevel, power, verbose)
}
func Dar_ManCompress2(g *aig.Graph, balance, updateLevel, fanout, power, verbose bool) (*aig.Graph, error) {
	return core.Dar_ManCompress2(g, balance, updateLevel, fanout, power, verbose)
}
func Dar_ManComputeCutSlots(g *aig.Graph, limit int, skipTruthMin bool) (map[aig.Edge][]CutSlot, error) {
	return core.Dar_ManComputeCutSlots(g, limit, skipTruthMin)
}
func Dar_ManComputeCuts(g *aig.Graph, limit int, skipTruthMin bool) (map[aig.Edge][]Cut, error) {
	return core.Dar_ManComputeCuts(g, limit, skipTruthMin)
}
func Dar_ManDefaultRefParams(p *RefactorParams) { core.Dar_ManDefaultRefParams(p) }
func Dar_ManDefaultRwrParams(p *RewriteParams)  { core.Dar_ManDefaultRwrParams(p) }
func Dar_ManRefactor(g *aig.Graph, params *RefactorParams) error {
	return core.Dar_ManRefactor(g, params)
}
func Dar_ManRewrite(g *aig.Graph, params *RewriteParams) error { return core.Dar_ManRewrite(g, params) }
func Dar_ManRewriteDefault(g *aig.Graph) (*aig.Graph, error)   { return core.Dar_ManRewriteDefault(g) }
func Dar_ManRwsat(g *aig.Graph, balance, verbose bool) (*aig.Graph, error) {
	return core.Dar_ManRwsat(g, balance, verbose)
}
func Dar_ManStart(g *aig.Graph, params *RewriteParams) (*Manager, error) {
	return core.Dar_ManStart(g, params)
}
func Dar_ManStop(m *Manager) { core.Dar_ManStop(m) }
func Dar_NewChoiceSynthesis(g *aig.Graph, p *ChoiceOptions) ([]*aig.Graph, error) {
	return core.Dar_NewChoiceSynthesis(g, p)
}
func Dar_NewChoiceSynthesisGuard(g *aig.Graph) bool { return core.Dar_NewChoiceSynthesisGuard(g) }
func Dar_NewCompress(g *aig.Graph, balance, updateLevel, power, verbose bool) (*aig.Graph, error) {
	return core.Dar_NewCompress(g, balance, updateLevel, power, verbose)
}
func Dar_NewCompress2(g *aig.Graph, balance, updateLevel, fanout, power, light, verbose bool) (*aig.Graph, error) {
	return core.Dar_NewCompress2(g, balance, updateLevel, fanout, power, light, verbose)
}
func Dar_Permutations(n int) [][]int { return core.Dar_Permutations(n) }
func Dar_Permutations_rec(result [][]int, nFact, n int, values []int) {
	core.Dar_Permutations_rec(result, nFact, n, values)
}
func Dar_Truth4VarNPN() *CanonicalTable { return core.Dar_Truth4VarNPN() }
func Dar_TruthPermute(truth uint32, permutation []int, variables int, reverse bool) uint32 {
	return core.Dar_TruthPermute(truth, permutation, variables, reverse)
}
func Dar_TruthPermute_int(minterms []int, count int, permutation []int, variables int, result []int) {
	core.Dar_TruthPermute_int(minterms, count, permutation, variables, result)
}
func Dar_TruthPolarize(truth uint32, polarity, variables int) uint32 {
	return core.Dar_TruthPolarize(truth, polarity, variables)
}
func Dar_WordCountOnes(word uint32) int32                     { return core.Dar_WordCountOnes(word) }
func DefaultActivityOptions() ActivityOptions                 { return core.DefaultActivityOptions() }
func DefaultChoiceOptions() ChoiceOptions                     { return core.DefaultChoiceOptions() }
func DefaultRefactorParams() RefactorParams                   { return core.DefaultRefactorParams() }
func DefaultRewriteParams() RewriteParams                     { return core.DefaultRewriteParams() }
func NewCutEnumerator(g *aig.Graph, limit int) *CutEnumerator { return core.NewCutEnumerator(g, limit) }
func RewriteForSAT(ctx context.Context, g *aig.Graph, balance bool, progress io.Writer) (*aig.Graph, error) {
	return core.RewriteForSAT(ctx, g, balance, progress)
}
func Rwsat(ctx context.Context, g *aig.Graph, balance, verbose bool) (*aig.Graph, error) {
	return core.Rwsat(ctx, g, balance, verbose)
}
func SimulateActivity(ctx context.Context, g *aig.Graph, p ActivityOptions) ([]float32, error) {
	return core.SimulateActivity(ctx, g, p)
}
