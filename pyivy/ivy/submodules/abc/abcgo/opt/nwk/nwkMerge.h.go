package nwk

const NWK_MAX_LIST = 16

type Nwk_LMPars_t struct {
	MaxLutSize, MaxSuppSize, MaxDistance, MaxLevelDiff, MaxFanout int32
	UseDiffSupp, UseTfiTfo, VeryVerbose, Verbose                  bool
}
type Nwk_Vrt_t struct {
	ID, Prev, Next int
	Edges          []int
}
type Nwk_Edg_t struct{ Node1, Node2 int }
type Nwk_Grf_t struct {
	MaxVertices    int
	Edges          []Nwk_Edg_t
	Vertices       []*Nwk_Vrt_t
	Pairs          *Vec_Int_t
	MapIDToLUT     []int32
	Lists1, Lists2 [NWK_MAX_LIST + 1]int
	edgeSet        map[Nwk_Edg_t]bool
	hashSize       int
}
