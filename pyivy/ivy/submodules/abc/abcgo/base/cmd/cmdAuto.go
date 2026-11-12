package cmd

import (
	"bufio"
	"fmt"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/sat/satoko"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// AutoEvaluator returns source-compatible MF-mapped-CNF SAT conflict cost.
// The graph/SAT integration installs the evaluator; the command package owns no
// substitute CNF mapper, since changing CNF changes the optimization objective.
type AutoEvaluator func(*ioa.Model, satoko.Options) (int64, error)

var autoBackend struct {
	sync.RWMutex
	evaluate AutoEvaluator
}

func SetAutoEvaluator(evaluate AutoEvaluator) {
	autoBackend.Lock()
	autoBackend.evaluate = evaluate
	autoBackend.Unlock()
}
func currentAutoEvaluator() AutoEvaluator {
	autoBackend.RLock()
	defer autoBackend.RUnlock()
	return autoBackend.evaluate
}

type AutoParameter struct {
	Symbol byte
	Value  float64
	Toggle bool
}
type AutoSetting struct {
	Options satoko.Options
	Text    string
}
type AutoResult struct {
	Setting AutoSetting
	Cost    int64
}

func ParseAutoParameters(reader io.Reader) ([][]AutoParameter, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	var rows [][]AutoParameter
	line := 0
	for scanner.Scan() {
		line++
		text := scanner.Text()
		if text == "" || strings.ContainsRune(" \t\r#", rune(text[0])) {
			continue
		}
		var row []AutoParameter
		for i := 0; i < len(text); {
			if strings.ContainsRune(" \t\r", rune(text[i])) {
				i++
				continue
			}
			symbol := text[i]
			i++
			parameter := AutoParameter{Symbol: symbol}
			if symbol >= 'a' && symbol <= 'z' {
				parameter.Toggle = true
			} else if symbol >= 'A' && symbol <= 'Z' {
				start := i
				for i < len(text) && (text[i] >= '0' && text[i] <= '9' || text[i] == '.') {
					i++
				}
				if start == i {
					return nil, fmt.Errorf("line %d: %c must be immediately followed by a number", line, symbol)
				}
				value, err := strconv.ParseFloat(text[start:i], 32)
				if err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
					return nil, fmt.Errorf("line %d: invalid numeric parameter", line)
				}
				parameter.Value = value
			} else {
				return nil, fmt.Errorf("line %d: invalid parameter symbol %q", line, symbol)
			}
			row = append(row, parameter)
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("empty autotuner configuration")
	}
	return rows, nil
}
func Cmd_ReadParamChoices(filename string) [][]AutoParameter {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()
	rows, err := ParseAutoParameters(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return rows
}
func Cmd_DeriveConvertIntoString(argc int32, argv []string) string {
	return strings.Join(commandArgs(argc, argv), " ") + " "
}
func deriveAutoOptions(args []string) (satoko.Options, error) {
	options := satoko.DefaultOptions()
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if arg == "-v" {
			options.Verbose = !options.Verbose
			continue
		}
		if len(arg) != 2 || arg[0] != '-' || i+1 == len(args) {
			return options, fmt.Errorf("invalid autotuner option %q", arg)
		}
		i++
		value, err := strconv.ParseFloat(args[i], 64)
		if err != nil || value < 0 || math.IsInf(value, 0) || math.IsNaN(value) {
			return options, fmt.Errorf("invalid value for %s", arg)
		}
		if strings.ContainsRune("CPFGHIJKLOQTU", rune(arg[1])) && value > math.MaxUint32 {
			return options, fmt.Errorf("value out of range for %s", arg)
		}
		switch arg[1] {
		case 'C':
			options.ConflictLimit = int64(value)
		case 'P':
			options.PropagationLimit = int64(value)
		case 'D':
			options.RestartFactor = value
		case 'E':
			options.BlockFactor = value
		case 'F':
			options.FirstBlockRestart = uint32(value)
		case 'G':
			options.RestartWindow = uint32(value)
		case 'H':
			options.TrailWindow = uint32(value)
		case 'I':
			options.FirstReduce = uint32(value)
		case 'J':
			options.ReduceIncrement = uint32(value)
		case 'K':
			options.SpecialReduceIncrement = uint32(value)
		case 'L':
			options.FreezeLBD = uint32(value)
		case 'M':
			options.LearntRatio = float32(value / 100)
		case 'N':
			options.GarbageRatio = float32(value / 100)
		case 'O':
			options.BinaryResolutionSize = uint32(value)
		case 'Q':
			options.BinaryResolutionLBD = uint32(value)
		case 'R':
			options.ClauseDecay = float32(value)
		case 'S':
			options.VariableDecay = value
		case 'T':
			options.VariableLimit = uint64(value)
		case 'U':
			options.VariableRescale = uint32(value)
		default:
			return options, fmt.Errorf("unknown autotuner option %s", arg)
		}
	}
	return options, nil
}
func Cmd_DeriveOptionFromSettings(argc int32, argv []string) *satoko.Options {
	options, err := deriveAutoOptions(commandArgs(argc, argv))
	if err != nil {
		return nil
	}
	return &options
}
func parameterArgs(parameter AutoParameter) []string {
	args := []string{"-" + string(parameter.Symbol)}
	if !parameter.Toggle {
		value := parameter.Value
		number := strconv.FormatFloat(value, 'f', 3, 64)
		if value == math.Trunc(value) {
			number = strconv.FormatFloat(value, 'f', 0, 64)
		}
		args = append(args, number)
	}
	return args
}
func CreateAutoSettings(rows [][]AutoParameter) ([]AutoSetting, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("empty autotuner configuration")
	}
	args := []string{"autotuner"}
	for _, parameter := range rows[0] {
		args = append(args, parameterArgs(parameter)...)
	}
	var settings []AutoSetting
	var visit func(int, []string) error
	visit = func(index int, args []string) error {
		if len(args) > 100 {
			return fmt.Errorf("more than 100 autotuner arguments")
		}
		if index == len(rows) {
			options, err := deriveAutoOptions(args)
			if err != nil {
				return err
			}
			settings = append(settings, AutoSetting{Options: options, Text: strings.Join(args, " ") + " "})
			return nil
		}
		row := rows[index]
		if len(row) == 1 && row[0].Toggle {
			if err := visit(index+1, args); err != nil {
				return err
			}
		}
		for _, parameter := range row {
			next := append(append([]string(nil), args...), parameterArgs(parameter)...)
			if err := visit(index+1, next); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(1, args); err != nil {
		return nil, err
	}
	return settings, nil
}
func Cmf_CreateOptions(rows [][]AutoParameter) []AutoSetting {
	settings, err := CreateAutoSettings(rows)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return settings
}
func Cmd_RunAutoTunerPrintOptions(options *satoko.Options) {
	fmt.Printf("-C %d -V %.3f -W %.3f", options.ConflictLimit, options.VariableDecay, options.ClauseDecay)
	if options.Verbose {
		fmt.Print(" -v")
	}
	fmt.Println()
}
func readAutoFiles(filename string, out io.Writer) ([]*ioa.Model, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	var models []*ioa.Model
	for scanner.Scan() {
		name := scanner.Text()
		if name == "" || strings.ContainsRune(" \t\r#", rune(name[0])) {
			continue
		}
		name = strings.TrimRight(name, " \t\r")
		model, err := ioa.Ioa_ReadAiger(name, 1)
		if err != nil {
			fmt.Fprintf(out, "Cannot read AIG from file %q: %v\n", name, err)
			continue
		}
		models = append(models, model)
	}
	return models, scanner.Err()
}
func Cmd_ReadFiles(filename string) []*ioa.Model {
	models, err := readAutoFiles(filename, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return models
}
func evaluateAuto(models []*ioa.Model, options satoko.Options, cores int, evaluate AutoEvaluator) (int64, error) {
	if evaluate == nil {
		return 0, fmt.Errorf("autotuner requires the source-compatible GIA MF SAT-cost evaluator")
	}
	if cores < 1 || cores > 101 {
		return 0, fmt.Errorf("autotuner cores must be between 1 and 101")
	}
	workers := cores - 1
	if workers < 1 {
		workers = 1
	}
	if workers > len(models) {
		workers = len(models)
	}
	jobs := make(chan *ioa.Model)
	var group sync.WaitGroup
	var mu sync.Mutex
	var cost int64
	var firstErr error
	for i := 0; i < workers; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for model := range jobs {
				value, err := evaluate(model, options)
				mu.Lock()
				if err != nil && firstErr == nil {
					firstErr = err
				}
				if value < 0 && firstErr == nil {
					firstErr = fmt.Errorf("negative SAT conflict cost")
				}
				if value > math.MaxInt64-cost {
					if firstErr == nil {
						firstErr = fmt.Errorf("SAT conflict cost overflow")
					}
				} else if value >= 0 {
					cost += value
				}
				mu.Unlock()
			}
		}()
	}
	for _, model := range models {
		jobs <- model
	}
	close(jobs)
	group.Wait()
	return cost, firstErr
}
func Cmd_RunAutoTunerEvalSimple(models []*ioa.Model, options *satoko.Options) (int64, error) {
	return evaluateAuto(models, *options, 1, currentAutoEvaluator())
}
func Cmd_RunAutoTunerEval(models []*ioa.Model, options *satoko.Options, cores int32) (int64, error) {
	return evaluateAuto(models, *options, int(cores), currentAutoEvaluator())
}
func RunAutoTuner(config, fileList string, cores int, out io.Writer, evaluate AutoEvaluator) (AutoResult, error) {
	file, err := os.Open(config)
	if err != nil {
		return AutoResult{}, err
	}
	rows, parseErr := ParseAutoParameters(file)
	closeErr := file.Close()
	if parseErr != nil {
		return AutoResult{}, parseErr
	}
	if closeErr != nil {
		return AutoResult{}, closeErr
	}
	settings, err := CreateAutoSettings(rows)
	if err != nil {
		return AutoResult{}, err
	}
	models, err := readAutoFiles(fileList, out)
	if err != nil {
		return AutoResult{}, err
	}
	if len(models) == 0 {
		return AutoResult{}, fmt.Errorf("no readable autotuner benchmarks")
	}
	best := AutoResult{Cost: math.MaxInt64}
	start := time.Now()
	for _, setting := range settings {
		fmt.Fprintf(out, "Evaluating settings: %20s...\n", setting.Text)
		cost, err := evaluateAuto(models, setting.Options, cores, evaluate)
		if err != nil {
			return AutoResult{}, err
		}
		fmt.Fprintf(out, "Cost = %6d.\n", cost)
		if best.Setting.Text == "" || cost < best.Cost {
			best = AutoResult{Setting: setting, Cost: cost}
		}
	}
	fmt.Fprintf(out, "The best settings are: %20s\nBest cost = %6d. Total time = %.2f sec\n", best.Setting.Text, best.Cost, time.Since(start).Seconds())
	return best, nil
}
func Cmd_RunAutoTuner(config, fileList string, cores int32) {
	if _, err := RunAutoTuner(config, fileList, int(cores), os.Stdout, currentAutoEvaluator()); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
func CmdCommandAutoTuner(f *Abc_Frame_t, argc int32, argv []string) int32 {
	args := commandArgs(argc, argv)
	config, list := "", ""
	cores := 3
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-N", "-C", "-F":
			option := args[i]
			i++
			if i == len(args) {
				return usage(f, "autotuner -C config -F filelist [-N cores]")
			}
			switch option {
			case "-C":
				config = args[i]
			case "-F":
				list = args[i]
			case "-N":
				value, err := strconv.Atoi(args[i])
				if err != nil {
					return 1
				}
				cores = value
			}
		case "-v":
		default:
			return usage(f, "autotuner -C config -F filelist [-N cores]")
		}
	}
	if config == "" || list == "" {
		return usage(f, "autotuner -C config -F filelist [-N cores]")
	}
	evaluate := currentAutoEvaluator()
	if local, ok := f.Resource("cmd:auto-evaluator").(AutoEvaluator); ok {
		evaluate = local
	}
	if _, err := RunAutoTuner(config, list, cores, f.Out, evaluate); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	return 0
}
func init() {
	extensionInitializers = append(extensionInitializers, func(f *Abc_Frame_t) { Cmd_CommandAdd(f, "Various", "autotuner", CmdCommandAutoTuner, 0) })
}
