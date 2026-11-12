package sim

import "github.com/glycerine/abcgo/aig/aig"

// Sim_ComputeSwitching estimates independent-pattern toggle probability 2p(1-p),
// not a temporal transition count. Like C it rounds the sample count to words.
func Sim_ComputeSwitching(info []uint32, words int) float32 {
	if words <= 0 {
		panic("switching requires samples")
	}
	ones := float32(Sim_UtilCountOnes(info, words))
	total := float32(32 * words)
	return 2 * ones / total * (total - ones) / total
}
func Sim_NtkComputeSwitching(g *aig.Graph, patterns int) []float32 {
	if patterns <= 0 {
		panic("switching requires samples")
	}
	order := graphOrder(g)
	words := SIM_NUM_WORDS(patterns)
	info := Sim_UtilInfoAlloc(int(aig.Aig_ManObjNumMax(g)), words, true)
	r := newRandom()
	for _, e := range g.Inputs() {
		fillRandom(info[e.NodeID()], r)
	}
	result := make([]float32, len(info))
	for _, e := range order {
		Sim_UtilSimulateNodeOne(g, e, info, words, 0)
		result[e.NodeID()] = Sim_ComputeSwitching(info[e.NodeID()], words)
	}
	return result
}
