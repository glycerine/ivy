package espresso

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

var espressoRunMutex sync.Mutex

// Run executes the original Espresso subcommands. Arguments exclude argv[0].
// PLA output is written to output; optional algorithm tracing and diagnostics
// retain their traditional process-standard-output behavior. Like Espresso's
// shared cube context, direct package algorithms must not run concurrently.
func Run(args []string, input io.Reader, output io.Writer) (err error) {
	espressoRunMutex.Lock()
	defer espressoRunMutex.Unlock()
	savedCube, savedCdata := cube, cdata
	savedDebug, savedFilename := debug, filename
	savedNames, savedTimes, savedCalls := total_name, total_time, total_calls
	flags := []*bool{&verbose_debug, &echo_comments, &echo_unknown_commands, &force_irredundant, &skip_make_sparse, &kiss, &pos, &print_solution, &recompute_onset, &remove_essential, &single_expand, &summary, &trace, &unwrap_onset, &use_random_order, &use_super_gasp, &debug_exact_minimization}
	savedFlags := make([]bool, len(flags))
	for i, p := range flags {
		savedFlags[i] = *p
	}
	defer func() {
		cube, cdata = savedCube, savedCdata
		debug, filename = savedDebug, savedFilename
		total_name, total_time, total_calls = savedNames, savedTimes, savedCalls
		for i, p := range flags {
			*p = savedFlags[i]
		}
	}()
	defer func() {
		if value := recover(); value != nil {
			err = fmt.Errorf("%v", value)
		}
		setdown_cube()
		cube = cube_struct{}
		sf_cleanup()
	}()
	if input == nil {
		input = os.Stdin
	}
	if output == nil {
		output = os.Stdout
	}
	cube = cube_struct{}
	set_espresso_flags()
	debug = 0
	verbose_debug = false
	print_solution = true
	skip_make_sparse = false
	kiss = false
	echo_comments = true
	echo_unknown_commands = true
	init_runtimeSummary()
	started := ptime()
	command := option_table[0]
	outType, inputType := int32(F_type), int32(FD_type)
	strategy := int32(0)
	first, last := int32(-1), int32(-1)
	lookup := func(name string) (commandOption, bool) {
		for _, entry := range option_table {
			if name == entry.name {
				return entry, true
			}
		}
		return commandOption{}, false
	}
	outputKind := func(name string) (int32, bool) {
		for _, entry := range pla_types {
			if strings.TrimPrefix(entry.key, "-") == name && entry.key != "" {
				return entry.value, true
			}
		}
		return 0, false
	}
	// Normalize the compatibility spellings before parsing traditional options.
	normalized := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-do", "-out":
			if i+1 >= len(args) {
				return fmt.Errorf("missing argument after %s", arg)
			}
			prefix := "-D"
			if arg == "-out" {
				prefix = "-o"
			}
			normalized = append(normalized, prefix+args[i+1])
			i++
		default:
			if _, ok := esp_opt_table[strings.TrimPrefix(arg, "-")]; ok && strings.HasPrefix(arg, "-") {
				normalized = append(normalized, "-e"+arg[1:])
			} else {
				normalized = append(normalized, arg)
			}
		}
	}
	var paths []string
	for i := 0; i < len(normalized); i++ {
		arg := normalized[i]
		if arg == "--" {
			paths = append(paths, normalized[i+1:]...)
			break
		}
		if arg == "--help" || arg == "-h" {
			usageTo(output)
			return nil
		}
		switch arg {
		case "-fdr":
			inputType = FDR_type
			continue
		case "-fr":
			inputType = FR_type
			continue
		case "-f":
			inputType = F_type
			continue
		}
		if arg == "-" || !strings.HasPrefix(arg, "-") {
			paths = append(paths, normalized[i:]...)
			break
		}
		for pos := 1; pos < len(arg); pos++ {
			option := arg[pos]
			value := ""
			if strings.ContainsRune("DS eorv", rune(option)) && option != ' ' {
				if pos+1 < len(arg) {
					value = arg[pos+1:]
					pos = len(arg)
				} else {
					if i+1 >= len(normalized) {
						return fmt.Errorf("missing argument for -%c", option)
					}
					i++
					value = normalized[i]
				}
			}
			switch option {
			case 'D':
				entry, ok := lookup(value)
				if !ok {
					return fmt.Errorf("bad subcommand %q", value)
				}
				command = entry
			case 'o':
				kind, ok := outputKind(value)
				if !ok {
					return fmt.Errorf("bad output type %q", value)
				}
				outType = kind
			case 'e':
				entry, ok := esp_opt_table[value]
				if !ok {
					return fmt.Errorf("bad espresso option %q", value)
				}
				*entry.variable = entry.value
			case 'd':
				debug = debug_table[""]
				trace, summary = true, true
			case 'v':
				value, ok := debug_table[value]
				if !ok {
					return fmt.Errorf("bad debug type")
				}
				verbose_debug = true
				debug |= value
			case 't':
				trace = true
			case 's':
				summary = true
			case 'x':
				print_solution = false
			case 'S':
				n, e := strconv.ParseInt(value, 10, 32)
				if e != nil {
					return fmt.Errorf("bad strategy %q", value)
				}
				strategy = int32(n)
			case 'r':
				if _, e := fmt.Sscanf(value, "%d-%d", &first, &last); e != nil {
					return fmt.Errorf("bad output range %q", value)
				}
			default:
				return fmt.Errorf("unknown option -%c", option)
			}
		}
	}
	if len(paths) > int(command.num_plas) {
		return fmt.Errorf("trailing arguments on command line")
	}
	if summary || trace {
		fmt.Fprintf(output, "# espresso %s\n# %s\n", strings.Join(args, " "), VERSION)
	}
	stdin := bufio.NewReader(input)
	var lastReader *bufio.Reader
	var opened []*os.File
	defer func() {
		for _, file := range opened {
			file.Close()
		}
	}()
	load := func(index int) (pPLA, error) {
		reader, name := stdin, "(stdin)"
		if index < len(paths) && paths[index] != "-" {
			name = paths[index]
			file, e := os.Open(name)
			if e != nil {
				return nil, e
			}
			opened = append(opened, file)
			reader = bufio.NewReader(file)
		}
		needsDC, needsOff := command.needs_dcset, command.needs_offset
		if command.key == KEY_echo {
			needsDC = outType&D_type != 0
			needsOff = outType&R_type != 0
		}
		var pla pPLA
		if read_pla(reader, needsDC, needsOff, inputType, &pla) < 0 {
			return nil, fmt.Errorf("unable to find PLA on %s", name)
		}
		pla.filename = name
		filename = name
		lastReader = reader
		return pla, nil
	}
	pla, e := load(0)
	if e != nil {
		return e
	}
	var second pPLA
	if command.num_plas == 2 {
		second, e = load(1)
		if e != nil {
			return e
		}
	}
	if summary || trace {
		PLA_summary(pla)
		if second != nil {
			PLA_summary(second)
		}
	}
	if command.key == KEY_MANY_ESPRESSO {
		for {
			pla.F = espresso(pla.F, pla.D, pla.R)
			if print_solution {
				fprint_pla(output, pla, outType)
			}
			kind := pla.pla_type
			setdown_cube()
			cube = cube_struct{}
			if read_pla(lastReader, true, true, kind, &pla) < 0 {
				break
			}
		}
		return nil
	}
	var cost cost_t
	rangeBounds := func(limit int32) {
		if first < 0 || first >= limit {
			first = 0
		}
		if last < 0 || last >= limit {
			last = limit - 1
		}
		if first > last {
			fatal("range starts after its end")
		}
	}
	verifyResult := func(original pcover) error {
		start := ptime()
		failed := verify(pla.F, original, pla.D)
		totals(start, VERIFY_TIME, pla.F, &cost)
		if failed {
			pla.F = original
			print_solution = false
			check_consistency(pla)
			return fmt.Errorf("cover verification failed")
		}
		return nil
	}
	switch command.key {
	case KEY_ESPRESSO:
		original := sf_save(pla.F)
		pla.F = espresso(pla.F, pla.D, pla.R)
		if e = verifyResult(original); e != nil {
			return e
		}
	case KEY_exact, KEY_qm:
		original := sf_save(pla.F)
		pla.F = minimize_exact(pla.F, pla.D, pla.R, command.key == KEY_exact)
		if e = verifyResult(original); e != nil {
			return e
		}
	case KEY_simplify:
		pla.F = simplify(cube1list(pla.F))
	case KEY_so:
		if strategy < 0 || strategy > 1 {
			strategy = 0
		}
		so_espresso(pla, strategy)
	case KEY_so_both:
		if strategy < 0 || strategy > 1 {
			strategy = 0
		}
		so_both_espresso(pla, strategy)
	case KEY_expand:
		pla.F = expand(pla.F, pla.R, false)
	case KEY_irred:
		pla.F = irredundant(pla.F, pla.D)
	case KEY_reduce:
		pla.F = reduce(pla.F, pla.D)
	case KEY_essen:
		for row := int32(0); row < pla.F.count; row++ {
			p := GETSET(pla.F, row)
			SET(p, RELESSEN)
			RESET(p, NONESSEN)
		}
		essential(&pla.F, &pla.D)
	case KEY_super_gasp:
		pla.F = super_gasp(pla.F, pla.D, pla.R, &cost)
	case KEY_gasp:
		pla.F = last_gasp(pla.F, pla.D, pla.R, &cost)
	case KEY_make_sparse:
		pla.F = make_sparse(pla.F, pla.D, pla.R)
	case KEY_primes:
		pla.F = primes_consensus(cube2list(pla.F, pla.D))
	case KEY_map:
		mapCover(pla.F)
		print_solution = false
	case KEY_opo:
		phase_assignment(pla, strategy)
	case KEY_opoall:
		rangeBounds(cube.part_size[cube.output])
		opoall(pla, first, last, strategy)
	case KEY_pair:
		find_optimal_pairing(pla, strategy)
	case KEY_pairall:
		pair_all(pla, strategy)
	case KEY_echo:
	case KEY_taut:
		word := " not "
		if tautology(cube1list(pla.F)) {
			word = " "
		}
		fmt.Fprintf(output, "ON-set is%sa tautology\n", word)
		print_solution = false
	case KEY_contain:
		pla.F = sf_contain(pla.F)
	case KEY_intersect:
		pla.F = cv_intersect(pla.F, second.F)
	case KEY_union:
		pla.F = sf_union(sf_contain(pla.F), sf_contain(second.F))
	case KEY_disjoint:
		pla.F = make_disjoint(pla.F)
	case KEY_dsharp:
		pla.F = cv_dsharp(pla.F, second.F)
	case KEY_sharp:
		pla.F = cv_sharp(pla.F, second.F)
	case KEY_lexsort:
		pla.F = lex_sort(pla.F)
	case KEY_stats:
		if !summary {
			PLA_summary(pla)
		}
		print_solution = false
	case KEY_minterms:
		rangeBounds(cube.num_vars)
		pla.F = sf_dupl(unravel_range(pla.F, first, last))
	case KEY_d1merge:
		rangeBounds(cube.num_vars)
		for v := first; v <= last; v++ {
			pla.F = d1merge(pla.F, v)
		}
	case KEY_d1merge_in:
		for v := int32(0); v < cube.num_binary_vars; v++ {
			pla.F = d1merge(pla.F, v)
		}
	case KEY_PLA_verify, KEY_verify:
		failed := false
		if command.key == KEY_PLA_verify {
			failed = PLA_verify(pla, second)
		} else {
			failed = verify(second.F, pla.F, pla.D)
		}
		if failed {
			return fmt.Errorf("PLA comparison failed; the PLA's are not equivalent")
		}
		fmt.Fprintln(output, "PLA's compared equal")
		return nil
	case KEY_check:
		check_consistency(pla)
		print_solution = false
	case KEY_mapdc:
		map_dcset(pla)
		outType = FD_type
	case KEY_equiv:
		find_equiv_outputs(pla)
		print_solution = false
	case KEY_separate:
		pla.F = complement(cube2list(pla.D, pla.R))
	case KEY_xor:
		pla.F = sf_contain(sf_join(cv_intersect(pla.F, second.R), cv_intersect(second.F, pla.R)))
	case KEY_fsm:
		verbose := int32(0)
		if summary {
			verbose = 1
		}
		disassemble_fsm(pla, verbose)
		print_solution = false
	case KEY_test:
		t := sf_join(pla.D, pla.R)
		empty := new_cover(0)
		pla.F = complement(cube1list(t))
		pla.F = expand(pla.F, t, false)
		pla.F = irredundant(pla.F, empty)
		t = sf_join(pla.F, pla.R)
		pla.D = expand(pla.D, t, false)
		pla.D = irredundant(pla.D, empty)
	default:
		return fmt.Errorf("unsupported command %s", command.name)
	}
	if trace {
		runtimeSummary()
	}
	if summary || trace {
		print_trace(pla.F, command.name, ptime()-started)
	}
	if print_solution {
		start := ptime()
		fprint_pla(output, pla, outType)
		totals(start, WRITE_TIME, pla.F, &cost)
	}
	return nil
}

// EspressoMain is a thin process adapter without os.Exit, allowing embedding.
func EspressoMain(argv []string) int {
	args := argv
	if len(args) > 0 {
		args = args[1:]
	}
	if err := Run(args, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
func runtimeSummary() {
	total := int64(1)
	for _, elapsed := range total_time {
		total += elapsed
	}
	for i, calls := range total_calls {
		if calls != 0 {
			percent := 100 * total_time[i]
			fmt.Printf("# %s\t%2d call(s) for %s (%2d.%01d%%)\n", total_name[i], calls, print_time(total_time[i]), percent/total, (10*(percent%total))/total)
		}
	}
}
func init_runtimeSummary() {
	total_time = [TIME_COUNT]int64{}
	total_calls = [TIME_COUNT]int32{}
	total_name = [TIME_COUNT]string{"READ       ", "WRITE      ", "COMPL      ", "REDUCE     ", "EXPAND     ", "ESSEN      ", "IRRED      ", "REDUCE_GASP", "EXPAND_GASP", "IRRED_GASP ", "MV_REDUCE  ", "RAISE_IN   ", "VERIFY     ", "PRIMES     ", "MINCOV     ", ""}
}
func subcommands() {
	names := make([]string, len(option_table))
	for i, entry := range option_table {
		names[i] = entry.name
	}
	fmt.Println("                " + strings.Join(names, ", "))
}
func usageTo(w io.Writer) {
	fmt.Fprintf(w, "%s\n\nSYNOPSIS: espresso [options] [file]\n\n", VERSION)
	fmt.Fprintln(w, "  -Dcmd     Select subcommand (default ESPRESSO)")
	fmt.Fprintln(w, "  -eopt     Select Espresso option: fast, ness, nirr, nunwrap, onset, pos, strong, eat, eatdots, kiss, random")
	fmt.Fprintln(w, "  -otype    Select output: f, fd, fr, fdr, pleasure, eqntott, kiss, cons")
	fmt.Fprintln(w, "  -rn-m     Select input/output range")
	fmt.Fprintln(w, "  -Sn       Select subcommand strategy")
	fmt.Fprintln(w, "  -d -vtype Enable debugging; -s summary; -t trace; -x suppress solution")
	names := make([]string, len(option_table))
	for i, entry := range option_table {
		names[i] = entry.name
	}
	fmt.Fprintln(w, "  commands: "+strings.Join(names, ", "))
}
func usage() { usageTo(os.Stdout) }
