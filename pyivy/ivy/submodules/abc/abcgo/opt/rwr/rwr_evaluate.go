package rwr

import "github.com/glycerine/abcgo/bool/dec"

// rwrEvaluateCut follows Rwr_CutEvaluate. count is the source's
// Dec_GraphToNetworkCount, supplied with the graph after its leaves are set.
// A negative count rejects the candidate. Ties retain the first candidate.
func (p *rwrLibrary) rwrEvaluateCut(truth uint16, fanins []any, saved int32, count func(*dec.Dec_Graph_t) int32) (*dec.Dec_Graph_t, int32) {
	if len(fanins) != 4 {
		panic("rwr: expected four cut fanins")
	}
	nodes := p.classes[p.classMap[truth]]
	p.nSubgraphs += len(nodes)
	bestGain := int32(-1)
	var best *dec.Dec_Graph_t
	for _, id := range nodes {
		graph := p.graphs[id]
		for i, fanin := range fanins {
			dec.Dec_GraphSetLeafFunc(graph, int32(i), fanin)
		}
		added := count(graph)
		if added == -1 {
			continue
		}
		if added < 0 || added > saved {
			panic("rwr: invalid prospective node count")
		}
		gain := saved - added
		if gain > bestGain {
			bestGain, best = gain, graph
			if gain > 0 {
				p.scores[id]++
				p.gains[id] += int16(gain)
				p.added[id] += int16(added)
			}
		}
	}
	return best, bestGain
}
