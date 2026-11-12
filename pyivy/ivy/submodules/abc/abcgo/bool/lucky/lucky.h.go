package lucky

type permInfo struct {
	varN                int32
	swapArray           []int32
	swapCtr, totalSwaps int32
	flipArray           []int32
	flipCtr, totalFlips int32
}
type PermInfo = permInfo

func NewPermInfo(variables int32) *PermInfo { return setPermInfoPtr(variables) }
