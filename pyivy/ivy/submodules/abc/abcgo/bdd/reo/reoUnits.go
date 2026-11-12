package reo

// Unit dispenser and plane-list operations ported from src/bdd/reo/reoUnits.c
// (Alan Mishchenko, UC Berkeley, 2002). Storage uses Go-owned slice chunks;
// linked-node addresses remain stable when the chunk directory grows.
func reoUnitsGetNextUnit(p *reo_man) *reo_unit             { return p.unitStorage.get() }
func reoUnitsRecycleUnit(p *reo_man, u *reo_unit)          { p.unitStorage.recycle(u) }
func reoUnitsRecycleUnitList(p *reo_man, plane *reo_plane) { p.unitStorage.recyclePlane(plane) }
func reoUnitsStopDispenser(p *reo_man)                     { p.unitStorage.stop() }
func reoUnitsAddUnitToPlane(plane *reo_plane, u *reo_unit) { addUnitToPlane(plane, u) }
func reoUnitsAddToFreeUnitList(p *reo_man)                 { p.unitStorage.grow() }
