package io

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
)

func IoCommandReadGig(frame *abcmain.Frame, args []string) int32 {
	if len(args) != 2 || strings.HasPrefix(args[1], "-") {
		fmt.Fprintln(frame.Err, "usage: &read_gig [-h] <file>")
		return 1
	}
	f, err := os.Open(args[1])
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	f.Close()
	abcmain.Abc_FrameUpdateGia(frame, gia.Gia_ManReadGig(args[1]))
	return 0
}

func transformSFToPLA(input io.Reader, output io.Writer) error {
	scan := bufio.NewScanner(input)
	scan.Buffer(make([]byte, 4096), 1_000_000)
	fmt.Fprintln(output, ".type fd")
	started := false
	for scan.Scan() {
		line := scan.Text()
		if strings.Contains(line, "END_SDF") {
			break
		}
		if strings.Contains(line, "SDF") {
			if !scan.Scan() {
				return fmt.Errorf("io: missing SDF dimensions")
			}
			fields := strings.Fields(scan.Text())
			if len(fields) < 3 {
				return fmt.Errorf("io: invalid SDF dimensions")
			}
			for i, name := range []string{".i", ".o", ".p"} {
				n, err := strconv.Atoi(fields[i])
				if err != nil {
					return fmt.Errorf("io: invalid SDF dimension: %w", err)
				}
				fmt.Fprintf(output, "%s %d\n", name, n)
			}
			started = true
		} else if started {
			fmt.Fprintln(output, line)
		}
	}
	if err := scan.Err(); err != nil {
		return err
	}
	if !started {
		return fmt.Errorf("io: missing SDF section")
	}
	_, err := fmt.Fprintln(output, ".e")
	return err
}

func Io_TransformSF2PLA(inputName, outputName string) {
	input, err := os.Open(inputName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer input.Close()
	output, err := os.Create(outputName)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = transformSFToPLA(input, output)
	closeErr := output.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		fmt.Println(err)
	}
}

func IoCommandReadSF(frame *abcmain.Frame, args []string) int32 {
	if len(args) != 2 || strings.HasPrefix(args[1], "-") {
		fmt.Fprintln(frame.Err, "usage: read_sf [-h] <file>")
		return 1
	}
	f, err := os.Open(args[1])
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	defer f.Close()
	var pla strings.Builder
	if err := transformSFToPLA(f, &pla); err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	doc, err := parsePLADocument(args[1], pla.String())
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	net, err := plaDocumentNetwork(doc, 0, 0, 0, false, true)
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	network := abc.Abc_NtkToLogic(net)
	abc.Abc_NtkDelete(net)
	if network == nil {
		return 1
	}
	network.Name, network.Spec = extra.Extra_FileNameGeneric(args[1]), args[1]
	return ioAcceptNetwork(frame, network)
}
