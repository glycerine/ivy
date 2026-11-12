package ivy

import (
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/opt/rwt"
)

type Dec_Graph_t = dec.Dec_Graph_t
type Dec_Node_t = dec.Dec_Node_t
type Rwt_Man_t = rwt.Rwt_Man_t
type Rwt_Node_t = rwt.Rwt_Node_t

var (
	Dec_GraphComplement         = dec.Dec_GraphComplement
	Dec_GraphIsComplement       = dec.Dec_GraphIsComplement
	Dec_GraphIsConst            = dec.Dec_GraphIsConst
	Dec_GraphIsVar              = dec.Dec_GraphIsVar
	Dec_GraphNode               = dec.Dec_GraphNode
	Dec_GraphNodeNum            = dec.Dec_GraphNodeNum
	Dec_GraphVar                = dec.Dec_GraphVar
	Rwt_ManAddTimeCuts          = rwt.Rwt_ManAddTimeCuts
	Rwt_ManAddTimeTotal         = rwt.Rwt_ManAddTimeTotal
	Rwt_ManAddTimeUpdate        = rwt.Rwt_ManAddTimeUpdate
	Rwt_ManCanonInfo            = rwt.Rwt_ManCanonInfo
	Rwt_ManClassSubgraphs       = rwt.Rwt_ManClassSubgraphs
	Rwt_ManNodeDecomposition    = rwt.Rwt_ManNodeDecomposition
	Rwt_ManReadCompl            = rwt.Rwt_ManReadCompl
	Rwt_ManReadCurrentLeaves    = rwt.Rwt_ManReadCurrentLeaves
	Rwt_ManReadDecs             = rwt.Rwt_ManReadDecs
	Rwt_ManReadLeaves           = rwt.Rwt_ManReadLeaves
	Rwt_ManRecordCut            = rwt.Rwt_ManRecordCut
	Rwt_ManRecordNodeConsidered = rwt.Rwt_ManRecordNodeConsidered
	Rwt_ManStart                = rwt.Rwt_ManStart
	Rwt_ManStop                 = rwt.Rwt_ManStop
	Rwt_ManStoreBest            = rwt.Rwt_ManStoreBest
	Rwt_Not                     = rwt.Rwt_Not
	Rwt_IsComplement            = rwt.Rwt_IsComplement
)
