package abc

// Specialized matching compatibility routines from src/map/if/ifMatch2.c.

// BatFunctionTableSetup implements Bat_ManFuncSetupTable. The native routine
// has no state and no side effects.
func BatFunctionTableSetup() {}

// BatFunctionTableSetdown implements Bat_ManFuncSetdownTable. The native
// routine has no state and no side effects.
func BatFunctionTableSetdown() {}

// BatCellFunctionLookup implements Bat_ManCellFuncLookup. The active native
// stub ignores every argument and unconditionally reports success. uintptr
// retains the unused opaque-manager value without introducing an interface.
func BatCellFunctionLookup(manager uintptr, truth []uint32, variableCount, leafCount int, description string) int {
	return 1
}
