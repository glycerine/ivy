package gia

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/miniaig"
	abcmain "github.com/glycerine/abcgo/base/main"
)

// Gia_MiniAigVerifyToFiles imports a MiniAIG, runs the frame's sequential
// correspondence commands, and writes the resulting miter and reduced graph.
func Gia_MiniAigVerifyToFiles(frame *abcmain.Frame, input, miterPath, reducedPath string) error {
	if frame == nil {
		return fmt.Errorf("gia: missing frame")
	}
	p := miniaig.Mini_AigLoad(input)
	if p == nil {
		return fmt.Errorf("gia: cannot load MiniAIG %q", input)
	}
	if err := abcmain.Abc_FrameGiaInputMiniAig(frame, p); err != nil {
		return err
	}
	if abcmain.Cmd_CommandExecute(frame, "&ps; &scorr; &ps") != 0 {
		return fmt.Errorf("gia: sequential correspondence command failed")
	}
	equivs, err := abcmain.Abc_FrameReadMiniAigEquivClasses(frame)
	if err != nil {
		return err
	}
	miter := Gia_MiniAigMiter(p, equivs)
	if err = Gia_AigerWrite(miter, miterPath, 0, 0, 0); err != nil {
		return err
	}
	fmt.Printf("Dumped miter AIG in file %q.\n", miterPath)
	reduced := Gia_MiniAigReduce(p, equivs)
	if err = Gia_AigerWrite(reduced, reducedPath, 0, 0, 0); err != nil {
		return err
	}
	fmt.Printf("Dumped reduced AIG in file %q.\n", reducedPath)
	return nil
}
func Gia_MiniAigVerify(frame *abcmain.Frame, input string) error {
	return Gia_MiniAigVerifyToFiles(frame, input, "mini_aig_miter.aig", "mini_aig_reduced.aig")
}
