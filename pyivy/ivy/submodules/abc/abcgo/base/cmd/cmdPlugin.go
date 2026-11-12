package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/aig/ioa"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const pluginBinariesKey = "cmd:plugin-binaries"

func pluginBinaries(f *Abc_Frame_t) map[string]string {
	if value, ok := f.Resource(pluginBinariesKey).(map[string]string); ok {
		return value
	}
	value := make(map[string]string)
	f.SetResource(pluginBinariesKey, value)
	return value
}
func Abc_GetBinaryName(f *Abc_Frame_t, argc int32, argv []string) string {
	args := commandArgs(argc, argv)
	if len(args) == 0 {
		return ""
	}
	return pluginBinaries(f)[args[0]]
}
func Abc_ManReadFile(filename string) []byte {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil
	}
	return data
}
func tokenTail(data []byte, token string) (string, bool) {
	index := bytes.Index(data, []byte(token))
	if index < 0 {
		return "", false
	}
	return string(data[index+len(token):]), true
}
func Abc_ManReadBinary(filename, token string) []int32 {
	tail, ok := tokenTail(Abc_ManReadFile(filename), token)
	if !ok {
		return nil
	}
	result := make([]int32, 0)
	for _, c := range tail {
		if c == '0' || c == '?' {
			result = append(result, 0)
		} else if c == '1' {
			result = append(result, 1)
		} else if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
			break
		}
	}
	return result
}
func Abc_ManReadInteger(filename, token string) int32 {
	tail, ok := tokenTail(Abc_ManReadFile(filename), token)
	if !ok {
		return -1
	}
	tail = strings.TrimSpace(tail)
	end := 0
	for end < len(tail) && (tail[end] >= '0' && tail[end] <= '9' || end == 0 && (tail[end] == '-' || tail[end] == '+')) {
		end++
	}
	value, err := strconv.ParseInt(tail[:end], 10, 32)
	if err != nil {
		return 0
	}
	return int32(value)
}
func Abc_ManReadStatus(filename, token string) int32 {
	tail, ok := tokenTail(Abc_ManReadFile(filename), token)
	if !ok {
		return -1
	}
	tail = strings.TrimSpace(tail)
	if strings.HasPrefix(tail, "proved") {
		return 1
	}
	if strings.HasPrefix(tail, "failed") {
		return 0
	}
	return -1
}

// DecodePluginAIG decodes the plugin's little-endian six-bit text encoding.
func DecodePluginAIG(text string) (*ioa.Model, error) {
	space := strings.IndexByte(text, ' ')
	if space < 0 {
		return nil, fmt.Errorf("missing encoded AIG byte count")
	}
	size, err := strconv.Atoi(text[:space])
	if err != nil || size < 0 {
		return nil, fmt.Errorf("invalid encoded AIG byte count")
	}
	encoded := text[space+1:]
	groups := size / 3
	if size%3 != 0 {
		groups++
	}
	if groups > len(encoded)/4 {
		return nil, fmt.Errorf("truncated encoded AIG")
	}
	decoded := make([]byte, groups*3)
	for i := 0; i < groups; i++ {
		var v [4]byte
		for j := range v {
			c := encoded[4*i+j]
			if c < '0' || c > '0'+63 {
				return nil, fmt.Errorf("invalid encoded AIG character")
			}
			v[j] = c - '0'
		}
		decoded[3*i] = v[0] | v[1]<<6
		decoded[3*i+1] = v[1]>>2 | v[2]<<4
		decoded[3*i+2] = v[2]>>4 | v[3]<<2
	}
	return ioa.Read(decoded[:size])
}
func Abc_ManReadAig(filename, token string) *ioa.Model {
	tail, ok := tokenTail(Abc_ManReadFile(filename), token)
	if !ok {
		return nil
	}
	line := strings.TrimLeft(tail, " ")
	if end := strings.IndexAny(line, "\r\n"); end >= 0 {
		line = line[:end]
	}
	model, err := DecodePluginAIG(line)
	if err != nil {
		return nil
	}
	return model
}

// PluginCounterexample records initial latch bits and one complete PI vector
// per frame. Output is the first failing PO in the final frame, as in ABC.
type PluginCounterexample struct {
	Bits                             []int32
	Registers, Inputs, Frame, Output int
}

func Abc_ManExpandCex(model *ioa.Model, bits []int32) []int32 {
	expanded, err := expandPluginCex(model, bits)
	if err != nil {
		return nil
	}
	return expanded
}
func expandPluginCex(model *ioa.Model, bits []int32) ([]int32, error) {
	registers := len(model.Latches)
	if len(bits) < registers {
		return nil, fmt.Errorf("counterexample lacks register bits")
	}
	used := make(map[uint32]bool)
	for _, gate := range model.ANDs {
		used[gate.Left>>1] = true
		used[gate.Right>>1] = true
	}
	for _, lit := range model.Outputs {
		used[lit>>1] = true
	}
	for _, latch := range model.Latches {
		used[latch.Next>>1] = true
	}
	active := 0
	for _, lit := range model.Inputs {
		if used[lit>>1] {
			active++
		}
	}
	if active == 0 {
		return nil, fmt.Errorf("cannot expand counterexample with no used inputs")
	}
	if (len(bits)-registers)%active != 0 {
		return nil, fmt.Errorf("partial counterexample frame")
	}
	frames := (len(bits) - registers) / active
	result := make([]int32, registers, registers+frames*len(model.Inputs))
	offset := registers
	for frame := 0; frame < frames; frame++ {
		for _, lit := range model.Inputs {
			value := int32(0)
			if used[lit>>1] {
				value = bits[offset]
				offset++
			}
			result = append(result, value)
		}
	}
	return result, nil
}
func verifyPluginCex(model *ioa.Model, bits []int32) (*PluginCounterexample, error) {
	if err := model.Check(); err != nil {
		return nil, err
	}
	registers, inputs := len(model.Latches), len(model.Inputs)
	if len(bits) < registers {
		return nil, fmt.Errorf("counterexample lacks register bits")
	}
	if inputs == 0 {
		return nil, fmt.Errorf("counterexample has no input frames")
	}
	if (len(bits)-registers)%inputs != 0 {
		var err error
		bits, err = expandPluginCex(model, bits)
		if err != nil {
			return nil, err
		}
	}
	frames := (len(bits) - registers) / inputs
	if frames == 0 {
		return nil, fmt.Errorf("empty counterexample")
	}
	state := append([]int32(nil), bits[:registers]...)
	values := make([]bool, int(model.MaxVar)+1)
	litValue := func(lit uint32) bool { return values[lit>>1] != (lit&1 != 0) }
	for frame := 0; frame < frames; frame++ {
		for i, lit := range model.Inputs {
			values[lit>>1] = bits[registers+frame*inputs+i] != 0
		}
		for i, latch := range model.Latches {
			values[latch.Literal>>1] = state[i] != 0
		}
		for _, gate := range model.ANDs {
			values[gate.Literal>>1] = litValue(gate.Left) && litValue(gate.Right)
		}
		if frame == frames-1 {
			for output, lit := range model.Outputs {
				if litValue(lit) {
					return &PluginCounterexample{Bits: append([]int32(nil), bits...), Registers: registers, Inputs: inputs, Frame: frame, Output: output}, nil
				}
			}
		}
		for i, latch := range model.Latches {
			state[i] = 0
			if litValue(latch.Next) {
				state[i] = 1
			}
		}
	}
	return nil, fmt.Errorf("counterexample does not fail an output")
}

func Cmd_CommandAbcPlugIn(f *Abc_Frame_t, argc int32, argv []string) int32 {
	args := commandArgs(argc, argv)
	if len(args) == 0 {
		return 1
	}
	binary := Abc_GetBinaryName(f, argc, argv)
	if binary == "" {
		fmt.Fprintln(f.Err, "Plugin binary is not registered.")
		return 1
	}
	value, err := f.Export("gia-aiger")
	provider, ok := value.(ioa.AigerProvider)
	if err != nil || !ok {
		if len(args) == 2 && args[1] == "-h" {
			command := exec.Command(binary, "-abc", args[0], "-h")
			command.Stdout = f.Out
			command.Stderr = f.Err
			command.Run()
		}
		fmt.Fprintln(f.Err, "Current AIG does not exist or cannot be exported.")
		return 1
	}
	model, err := provider.AigerModel()
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	directory, err := os.MkdirTemp("", "abc-plugin-")
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	leave := args[len(args)-1] == "!"
	if leave {
		args = args[:len(args)-1]
		fmt.Fprintf(f.Out, "Plugin files retained in %s\n", directory)
	} else {
		defer os.RemoveAll(directory)
	}
	input, output := filepath.Join(directory, "input.aig"), filepath.Join(directory, "output.txt")
	if err := ioa.Ioa_WriteAiger(model, input, 0, 0); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	command := exec.Command(binary, append([]string{"-abc", "-input=" + input, "-output=" + output}, args...)...)
	command.Stdin = f.In
	command.Stdout = f.Out
	command.Stderr = f.Err
	if err := command.Run(); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	data, err := os.ReadFile(output)
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	if len(data) == 0 {
		return 0
	}
	f.Status = Abc_ManReadStatus(output, "result:")
	f.NFrames = Abc_ManReadInteger(output, "bug-free-depth:")
	if classes := Abc_ManReadBinary(output, "abstraction:"); classes != nil {
		if err := f.Import("gia-flop-classes", classes); err != nil {
			fmt.Fprintln(f.Err, err)
			return 1
		}
	}
	if bits := Abc_ManReadBinary(output, "counter-example:"); bits != nil {
		witness, err := verifyPluginCex(model, bits)
		if err != nil {
			fmt.Fprintln(f.Err, err)
			return 1
		}
		if err := f.Import("gia-counterexample", witness); err != nil {
			fmt.Fprintln(f.Err, err)
			return 1
		}
		fmt.Fprintf(f.Out, "Returned counter-example successfully verified in ABC (frame %d, output %d).\n", witness.Frame, witness.Output)
	}
	if _, present := tokenTail(data, "aig:"); present {
		replacement := Abc_ManReadAig(output, "aig:")
		if replacement == nil {
			fmt.Fprintln(f.Err, "Invalid plugin AIG.")
			return 1
		}
		if err := f.Import("gia-aiger", replacement); err != nil {
			fmt.Fprintln(f.Err, err)
			return 1
		}
	}
	return 0
}
func Cmd_CommandAbcLoadPlugIn(f *Abc_Frame_t, argc int32, argv []string) int32 {
	args := commandArgs(argc, argv)
	pathSearch, verbose := false, false
	var positional []string
	for _, arg := range args[1:] {
		switch arg {
		case "-p":
			pathSearch = !pathSearch
		case "-v":
			verbose = !verbose
		case "-h":
			return usage(f, "load_plugin [-pv] binary section")
		default:
			positional = append(positional, arg)
		}
	}
	if len(positional) != 2 {
		return usage(f, "load_plugin [-pv] binary section")
	}
	binary := positional[0]
	var err error
	if pathSearch {
		binary, err = exec.LookPath(binary)
	} else {
		binary, err = filepath.Abs(binary)
		if err == nil {
			_, err = os.Stat(binary)
		}
	}
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	command := exec.Command(binary, "-abc", "-list-commands")
	command.Stderr = f.Err
	data, err := command.Output()
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	var names []string
	for scanner.Scan() {
		name := strings.TrimSuffix(scanner.Text(), "\r")
		if name == "" {
			continue
		}
		if strings.ContainsAny(name, " \t;#") {
			fmt.Fprintf(f.Err, "Invalid plugin command name %q\n", name)
			return 1
		}
		names = append(names, name)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	for _, name := range names {
		pluginBinaries(f)[name] = binary
		Cmd_CommandAdd(f, positional[1], name, Cmd_CommandAbcPlugIn, 1)
		if verbose {
			fmt.Fprintf(f.Out, "Creating command %s with binary %s\n", name, binary)
		}
	}
	return 0
}
func init() {
	extensionInitializers = append(extensionInitializers, func(f *Abc_Frame_t) { Cmd_CommandAdd(f, "Various", "load_plugin", Cmd_CommandAbcLoadPlugIn, 0) })
}

// CounterexampleData lets graph owners consume verified witnesses without
// importing the command package into their own engine layer.
func (witness *PluginCounterexample) CounterexampleData() (bits []int32, registers, inputs, frame, output int) {
	if witness == nil {
		return nil, 0, 0, 0, 0
	}
	return append([]int32(nil), witness.Bits...), witness.Registers, witness.Inputs, witness.Frame, witness.Output
}
