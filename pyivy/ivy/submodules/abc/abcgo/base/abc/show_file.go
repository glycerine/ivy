package abc

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	abcmain "github.com/glycerine/abcgo/base/main"
)

// Abc_ShowGetFileName derives a DOT filename using ABC's ASCII sanitization.
func Abc_ShowGetFileName(name string) string {
	data := []byte(name + ".dot")
	for i, c := range data {
		if !('0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '.') {
			data[i] = '_'
		}
	}
	return string(data)
}

// showDotFile runs Graphviz and starts a viewer; the DOT input is removed only
// once Graphviz has successfully generated the PostScript file.
func showDotFile(filename, dot, viewer string) error {
	if _, err := os.Stat(filename); err != nil {
		return fmt.Errorf("cannot open intermediate file %q: %w", filename, err)
	}
	postscript := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".ps"
	cmd := exec.Command(dot, "-Tps", "-o", postscript, filename)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cannot generate %q: %w: %s", postscript, err, strings.TrimSpace(string(output)))
	}
	if _, err := os.Stat(postscript); err != nil {
		return fmt.Errorf("cannot open intermediate file %q: %w", postscript, err)
	}
	if err := os.Remove(filename); err != nil {
		return err
	}
	command := exec.Command(viewer, postscript)
	if err := command.Start(); err != nil {
		return fmt.Errorf("cannot execute %q: %w", viewer, err)
	}
	go command.Wait()
	return nil
}

func Abc_ShowFile(filename string) {
	dot, viewer := "dot", "gv"
	dotFlag, viewerFlag := "dotunix", "gsviewunix"
	if runtime.GOOS == "windows" {
		dot, viewer = "dot.exe", "gsview32.exe"
		dotFlag, viewerFlag = "dotwin", "gsviewwin"
	}
	if configured := abcmain.Abc_FrameReadFlag(dotFlag); configured != "" {
		dot = configured
	}
	if configured := abcmain.Abc_FrameReadFlag(viewerFlag); configured != "" {
		viewer = configured
	}
	if err := showDotFile(filename, dot, viewer); err != nil {
		fmt.Println(err)
	}
}

func Abc_NtkShowFlopDependency(n *Network) {
	if Abc_NtkIsStrash(n) == 0 && Abc_NtkIsLogic(n) == 0 {
		panic("abc: flop dependency requires a logic or AIG network")
	}
	filename := Abc_ShowGetFileName(n.Name)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Cannot open the intermediate file %q.\n", filename)
		return
	}
	if err = WriteFlopDependency(file, n); err == nil {
		err = file.Close()
	} else {
		file.Close()
	}
	if err != nil {
		fmt.Printf("Cannot write intermediate file %q: %v\n", filename, err)
		return
	}
	Abc_ShowFile(filename)
}

// WriteNetworkVisualization renders a copy, preserving the source network.
func WriteNetworkVisualization(w io.Writer, n *Network, gateNames, sequential, reverse bool) error {
	if n == nil || (Abc_NtkIsStrash(n) == 0 && Abc_NtkIsLogic(n) == 0) {
		return fmt.Errorf("abc: visualization requires a logic or AIG network")
	}
	if Abc_NtkGetChoiceNum(n) != 0 {
		return fmt.Errorf("visualization of AIGs with choice nodes is disabled")
	}
	copy := Abc_NtkDup(n)
	defer Abc_NtkDelete(copy)
	if Abc_NtkIsLogic(copy) != 0 && Abc_NtkHasMapping(copy) == 0 && Abc_NtkToSop(copy, -1, int32(^uint32(0)>>1)) == 0 {
		return fmt.Errorf("abc: failed to convert network to SOP")
	}
	copy.Barriers = 0
	return WriteDOT(w, copy, nil, nil, gateNames, reverse, sequential)
}

func Abc_NtkShow(n *Network, gateNames, sequential, reverse int32) {
	filename := Abc_ShowGetFileName(n.Name)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Cannot open the intermediate file %q.\n", filename)
		return
	}
	if err = WriteNetworkVisualization(file, n, gateNames != 0, sequential != 0, reverse != 0); err == nil {
		err = file.Close()
	} else {
		file.Close()
	}
	if err != nil {
		fmt.Printf("Cannot write intermediate file %q: %v\n", filename, err)
		return
	}
	Abc_ShowFile(filename)
}

// WriteCutVisualization renders the containing cone and eligible fanouts.
func WriteCutVisualization(w io.Writer, root *Object, nodeLimit, coneLimit int32) error {
	if root == nil || root.network == nil || Abc_NtkIsStrash(root.network) == 0 || root.Kind != ObjectNode || len(root.fanins) != 2 {
		return fmt.Errorf("abc: cut visualization requires an AIG node")
	}
	if nodeLimit < 2 || coneLimit <= nodeLimit {
		return fmt.Errorf("abc: invalid cut limits")
	}
	const infinity = int32(^uint32(0) >> 1)
	manager := Abc_NtkManCutStart(nodeLimit, coneLimit, 2, infinity)
	defer Abc_NtkManCutStop(manager)
	small := Abc_NodeFindCut(manager, root, 1)
	inside := Abc_NodeConeCollect([]*Object{root}, manager.large, true)
	present := make(map[*Object]bool, len(inside))
	for _, node := range inside {
		present[node] = true
	}
	for _, node := range Abc_NodeCollectTfoCands(manager, root, small, infinity) {
		if !present[node] {
			present[node] = true
			inside = append(inside, node)
		}
	}
	highlight := append(append([]*Object(nil), small...), root)
	return WriteDOT(w, root.network, inside, highlight, false, false, false)
}

func Abc_NodeShowCut(root *Object, nodeLimit, coneLimit int32) {
	filename := Abc_ShowGetFileName(Abc_ObjName(root))
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Cannot open the intermediate file %q.\n", filename)
		return
	}
	if err = WriteCutVisualization(file, root, nodeLimit, coneLimit); err == nil {
		err = file.Close()
	} else {
		file.Close()
	}
	if err != nil {
		fmt.Printf("Cannot write intermediate file %q: %v\n", filename, err)
		return
	}
	Abc_ShowFile(filename)
}
