package abcmain

// Abc_FrameHandle replaces the old header's untyped void* frame handles. The
// public functions in abcapis.go retain the legacy names with typed Go frames.
type Abc_FrameHandle = *Frame
