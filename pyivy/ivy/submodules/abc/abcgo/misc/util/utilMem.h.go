package util

// Recorder state is protected by memoryRecordMu. Interrupt remains a legacy
// integer flag; command-layer cancellation should be carried by contexts.
var s_vAllocs *Vec_Mem_t
var s_vFrees *Vec_Mem_t
var s_fInterrupt int32
