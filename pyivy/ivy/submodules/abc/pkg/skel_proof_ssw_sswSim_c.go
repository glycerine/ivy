package abc

// Hand-written skeleton of the type in src/proof/ssw/sswSim.c.

// SswSml is Ssw_Sml_t: sequential simulation manager.
type SswSml struct {
	pAig         *AigManager // the original AIG manager
	nPref        int         // the number of timeframes in the prefix
	nFrames      int         // the number of timeframes
	nWordsFrame  int         // the number of words in each timeframe
	nWordsTotal  int         // the total number of words at a node
	nWordsPref   int         // the number of word in the prefix
	fNonConstOut int         // have seen a non-const-0 output during simulation
	nSimRounds   int         // statistics
	timeSim      int64       // statistics
	pData        []uint32    // unsigned pData[0] : flexible array member, simulation data for the nodes
}
