package abcmain

func Abc_Start() { Abc_FrameGetGlobalFrame() }
func Abc_Stop() {
	if frame := Abc_FrameReadGlobalFrame(); frame != nil {
		Abc_FrameDeallocate(frame)
	}
}
