package test_test

import (
	"testing"

	abcmain "github.com/glycerine/abcgo/base/main"
	abctest "github.com/glycerine/abcgo/base/test"
)

// These assertions also verify the original C frame alias remains compatible
// with the importable Go frame contract.
var _ func(*abcmain.Frame) = abctest.Test_Init
var _ func(*abcmain.Abc_Frame_t) = abctest.Test_End

func TestEmptyHooksAcceptFrameAndNil(t *testing.T) {
	for _, frame := range []*abcmain.Frame{nil, {}} {
		abctest.Test_Init(frame)
		abctest.Test_End(frame)
	}
}
