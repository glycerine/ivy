package abci

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	baseabc "github.com/glycerine/abcgo/base/abc"
	baseio "github.com/glycerine/abcgo/base/io"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/sat/bmc"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// C's atof accepts a numeric prefix and ignores trailing option text.
var abcFloatPrefix = regexp.MustCompile(`(?i)^[+-]?(?:inf(?:inity)?|nan(?:\([^)]*\))?|0x(?:[0-9a-f]+\.?[0-9a-f]*|\.[0-9a-f]+)(?:p[+-]?[0-9]+)?|(?:[0-9]+\.?[0-9]*|\.[0-9]+)(?:e[+-]?[0-9]+)?)`)

func atof(value string) float64 {
	prefix := abcFloatPrefix.FindString(strings.TrimLeft(value, " \t\n\r\f\v"))
	if prefix == "" {
		return 0
	}
	unsigned := strings.TrimLeft(strings.ToLower(prefix), "+-")
	if strings.HasPrefix(unsigned, "nan") {
		return math.NaN()
	}
	if strings.HasPrefix(unsigned, "inf") {
		if strings.HasPrefix(prefix, "-") {
			return math.Inf(-1)
		}
		return math.Inf(1)
	}
	if strings.HasPrefix(unsigned, "0x") && !strings.ContainsAny(prefix, "pP") {
		prefix += "p0"
	}
	result, _ := strconv.ParseFloat(prefix, 64)
	return result
}

func abcTransferGiaCex(frame *Abc_Frame_t) {
	witness := abcFrameGia(frame).TakeSequentialCounterexample()
	Abc_FrameReplaceCex(frame, &witness)
}

// The source bridge assigns 106 to a netlist and 107 to an abstraction.
func abcSendBridge(w io.Writer, graph *Gia_Man_t, kind int32) error {
	netlist, err := gia.Gia_ManToBridgeNetlist(graph)
	if err != nil {
		return err
	}
	if util.Gia_ManToBridgeAbsNetlist(w, netlist, kind) == 0 {
		return fmt.Errorf("could not write bridge netlist")
	}
	return nil
}

func abcGiaFromOwnedAig(graph *aig.Graph) *Gia_Man_t {
	if graph == nil {
		return nil
	}
	defer aig.Aig_ManStop(graph)
	return gia.Gia_ManFromAig(graph)
}

func abcAsGia(value any) *Gia_Man_t {
	if value == nil {
		return nil
	}
	return value.(*gia.Gia_Man_t)
}

func Abc_NtkConvertBb2Wb(input, output string, sequential, verbose int32) {
	file, err := os.Open(input)
	if err != nil {
		fmt.Printf("Input file %q cannot be opened.\n", input)
		return
	}
	file.Close()
	file, err = os.Create(output)
	if err != nil {
		fmt.Printf("Output file %q cannot be opened.\n", output)
		return
	}
	file.Close()
	network := baseio.Io_ReadNetlist(input, baseio.Io_ReadFileType(input), 1)
	if network == nil {
		fmt.Printf("Reading input file %q has failed.\n", input)
		return
	}
	defer baseabc.Abc_NtkDelete(network)
	if err := baseio.Io_WriteBlif(network, output, 1, 1, sequential); err != nil {
		fmt.Println(err)
	}
}

func abcSpecBMC(graph *aig.Graph, start, frames, conflicts, verbose int32) *Abc_Cex_t {
	bmc.Saig_BmcPerform(graph, start, frames, 2000, 0, conflicts, 2000000, verbose, 0, nil, 0, 0)
	witness := graph.SequentialCounterexample()
	graph.ClearSequentialCounterexample()
	return witness
}

func abcCheck75(manager *ifpkg.If_Man_t, truth []uint32, vars, leaves int32, structure string) int32 {
	if len(truth) == 0 {
		return 0
	}
	return ifpkg.If_CutPerformCheck75(manager, &truth[0], vars, leaves, structure)
}
