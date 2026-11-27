package dsd

// APIs of the DSD node (dsdApi.c). The node's type is one of the
// following: constant 1 node, the buffer (or the elementary variable), OR
// gate, EXOR gate, or PRIME function. Dsd_NodeReadFunc() is the global
// function of the node; Dsd_NodeReadSupp() its support; Dsd_NodeReadDecs()
// the decomposition nodes of its formal inputs. A complemented handle reads
// its regular node.
func Dsd_NodeReadType(p *Dsd_Node_t) Dsd_Type_t        { return Dsd_Regular(p).Type }
func Dsd_NodeReadFunc(p *Dsd_Node_t) *DdNode           { return Dsd_Regular(p).G }
func Dsd_NodeReadSupp(p *Dsd_Node_t) *DdNode           { return Dsd_Regular(p).S }
func Dsd_NodeReadDecs(p *Dsd_Node_t) []*Dsd_Node_t     { return Dsd_Regular(p).pDecs }
func Dsd_NodeReadDec(p *Dsd_Node_t, i int) *Dsd_Node_t { return Dsd_Regular(p).pDecs[i] }
func Dsd_NodeReadDecsNum(p *Dsd_Node_t) int            { return len(Dsd_Regular(p).pDecs) }
func Dsd_NodeReadMark(p *Dsd_Node_t) int               { return Dsd_Regular(p).Mark }

// Dsd_NodeSetMark sets the integer mark in the given DSD node. The mark is
// guaranteed to persist as long as the calls to the decomposition are not
// performed.
func Dsd_NodeSetMark(p *Dsd_Node_t, mark int) { Dsd_Regular(p).Mark = mark }

// APIs of the DSD manager. The root may have the complemented attribute.
func Dsd_ManagerReadRoot(m *Dsd_Manager_t, i int) *Dsd_Node_t  { return m.roots[i] }
func Dsd_ManagerReadInput(m *Dsd_Manager_t, i int) *Dsd_Node_t { return m.inputs[i] }
func Dsd_ManagerReadConst1(m *Dsd_Manager_t) *Dsd_Node_t       { return m.pConst1 }
func Dsd_ManagerReadDd(m *Dsd_Manager_t) *DdManager            { return m.dd }
