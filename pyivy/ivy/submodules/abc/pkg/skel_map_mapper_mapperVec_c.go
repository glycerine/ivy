package abc

// Hand-checked skeleton of src/map/mapper/mapperVec.c.

// Map_NodeVecAlloc
// C: Map_NodeVec_t * Map_NodeVecAlloc( int nCap )
func MapNodeVecAlloc(nCap int) *MapNodeVec {
	panic("TODO, stubbed")
}

// Map_NodeVecFree
// C: void Map_NodeVecFree( Map_NodeVec_t * p )
func (p *MapNodeVec) MapNodeVecFree() {
	panic("TODO, stubbed")
}

// Map_NodeVecDup
// C: Map_NodeVec_t * Map_NodeVecDup( Map_NodeVec_t * p )
func (p *MapNodeVec) MapNodeVecDup() *MapNodeVec {
	panic("TODO, stubbed")
}

// Map_NodeVecReadArray
// C: Map_Node_t ** Map_NodeVecReadArray( Map_NodeVec_t * p )
func (p *MapNodeVec) MapNodeVecReadArray() []*MapNode {
	panic("TODO, stubbed")
}

// Map_NodeVecReadSize
// C: int Map_NodeVecReadSize( Map_NodeVec_t * p )
func (p *MapNodeVec) MapNodeVecReadSize() int {
	panic("TODO, stubbed")
}

// Map_NodeVecGrow
// C: void Map_NodeVecGrow( Map_NodeVec_t * p, int nCapMin )
func (p *MapNodeVec) MapNodeVecGrow(nCapMin int) {
	panic("TODO, stubbed")
}

// Map_NodeVecShrink
// C: void Map_NodeVecShrink( Map_NodeVec_t * p, int nSizeNew )
func (p *MapNodeVec) MapNodeVecShrink(nSizeNew int) {
	panic("TODO, stubbed")
}

// Map_NodeVecClear
// C: void Map_NodeVecClear( Map_NodeVec_t * p )
func (p *MapNodeVec) MapNodeVecClear() {
	panic("TODO, stubbed")
}

// Map_NodeVecPush
// C: void Map_NodeVecPush( Map_NodeVec_t * p, Map_Node_t * Entry )
func (p *MapNodeVec) MapNodeVecPush(Entry *MapNode) {
	panic("TODO, stubbed")
}

// Map_NodeVecPushUnique
// C: int Map_NodeVecPushUnique( Map_NodeVec_t * p, Map_Node_t * Entry )
func (p *MapNodeVec) MapNodeVecPushUnique(Entry *MapNode) int {
	panic("TODO, stubbed")
}

// Map_NodeVecPop
// C: Map_Node_t * Map_NodeVecPop( Map_NodeVec_t * p )
func (p *MapNodeVec) MapNodeVecPop() *MapNode {
	panic("TODO, stubbed")
}

// Map_NodeVecRemove
// C: void Map_NodeVecRemove( Map_NodeVec_t * p, Map_Node_t * Entry )
func (p *MapNodeVec) MapNodeVecRemove(Entry *MapNode) {
	panic("TODO, stubbed")
}

// Map_NodeVecWriteEntry
// C: void Map_NodeVecWriteEntry( Map_NodeVec_t * p, int i, Map_Node_t * Entry )
func (p *MapNodeVec) MapNodeVecWriteEntry(i int, Entry *MapNode) {
	panic("TODO, stubbed")
}

// Map_NodeVecReadEntry
// C: Map_Node_t * Map_NodeVecReadEntry( Map_NodeVec_t * p, int i )
func (p *MapNodeVec) MapNodeVecReadEntry(i int) *MapNode {
	panic("TODO, stubbed")
}

// Map_NodeVecSortByLevel
// C: void Map_NodeVecSortByLevel( Map_NodeVec_t * p )
func (p *MapNodeVec) MapNodeVecSortByLevel() {
	panic("TODO, stubbed")
}

// Map_NodeVecCompareLevels (static)
// C: int Map_NodeVecCompareLevels( Map_Node_t ** pp1, Map_Node_t ** pp2 )
func mapNodeVecCompareLevels(pp1 **MapNode, pp2 **MapNode) int {
	panic("TODO, stubbed")
}
