package dsc

// DSC_MAX_VAR bounds truth-table variable counts handled by this implementation.
const DSC_MAX_VAR = 16

// DSC_MAX_STR is the maximum expression buffer size used by the C algorithm.
// Go string results do not require fixed-size buffers, but the value remains
// exported for source compatibility and caller sizing where needed.
const DSC_MAX_STR = 4 * DSC_MAX_VAR
