package espresso

func new_PLA() pPLA       { return &PLA_t{} }
func PLA_labels(pla pPLA) { pla.label = make([]string, cube.size) }

// free_PLA drops the owned covers and linked metadata, allowing Go to reclaim
// them while preserving the explicit lifecycle used by the original callers.
func free_PLA(pla pPLA) {
	if pla != nil {
		*pla = PLA_t{}
	}
}
