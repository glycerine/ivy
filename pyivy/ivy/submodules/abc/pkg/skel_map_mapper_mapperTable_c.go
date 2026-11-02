package abc

// Hand-checked skeleton of src/map/mapper/mapperTable.c.

// Map_SuperTableCreate
// C: Map_HashTable_t * Map_SuperTableCreate( Map_SuperLib_t * pLib )
func (pLib *MapSuperLib) MapSuperTableCreate() *MapHashTable {
	panic("TODO, stubbed")
}

// Map_SuperTableFree
// C: void Map_SuperTableFree( Map_HashTable_t * p )
func (p *MapHashTable) MapSuperTableFree() {
	panic("TODO, stubbed")
}

// Map_SuperTableInsertC
// C: int Map_SuperTableInsertC( Map_HashTable_t * p, unsigned uTruthC[], Map_Super_t * pGate )
func (p *MapHashTable) MapSuperTableInsertC(uTruthC []uint32, pGate *MapSuper) int {
	panic("TODO, stubbed")
}

// Map_SuperTableInsert
// C: int Map_SuperTableInsert( Map_HashTable_t * p, unsigned uTruth[], Map_Super_t * pGate, unsigned uPhase )
func (p *MapHashTable) MapSuperTableInsert(uTruth []uint32, pGate *MapSuper, uPhase uint32) int {
	panic("TODO, stubbed")
}

// Map_SuperTableLookupC
// C: Map_Super_t * Map_SuperTableLookupC( Map_SuperLib_t * p, unsigned uTruth[] )
func (p *MapSuperLib) MapSuperTableLookupC(uTruth []uint32) *MapSuper {
	panic("TODO, stubbed")
}

// Map_SuperTableLookup (puPhase is an out parameter)
// C: Map_Super_t * Map_SuperTableLookup( Map_HashTable_t * p, unsigned uTruth[], unsigned * puPhase )
func (p *MapHashTable) MapSuperTableLookup(uTruth []uint32, puPhase *uint32) *MapSuper {
	panic("TODO, stubbed")
}

// Map_SuperTableResize (static)
// C: void Map_SuperTableResize( Map_HashTable_t * p )
func (p *MapHashTable) mapSuperTableResize() {
	panic("TODO, stubbed")
}

// Map_SuperTableCompareSupergates
// C: int Map_SuperTableCompareSupergates( Map_Super_t ** ppS1, Map_Super_t ** ppS2 )
func MapSuperTableCompareSupergates(ppS1 **MapSuper, ppS2 **MapSuper) int {
	panic("TODO, stubbed")
}

// Map_SuperTableCompareGatesInList
// C: int Map_SuperTableCompareGatesInList( Map_Super_t ** ppS1, Map_Super_t ** ppS2 )
func MapSuperTableCompareGatesInList(ppS1 **MapSuper, ppS2 **MapSuper) int {
	panic("TODO, stubbed")
}

// Map_SuperTableSortSupergates
// C: void Map_SuperTableSortSupergates( Map_HashTable_t * p, int nSupersMax )
func (p *MapHashTable) MapSuperTableSortSupergates(nSupersMax int) {
	panic("TODO, stubbed")
}

// Map_SuperTableSortSupergatesByDelay
// C: void Map_SuperTableSortSupergatesByDelay( Map_HashTable_t * p, int nSupersMax )
func (p *MapHashTable) MapSuperTableSortSupergatesByDelay(nSupersMax int) {
	panic("TODO, stubbed")
}
