package espresso

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var line_length_error bool
var lineno int32

func readPLAByte(fp *bufio.Reader) (byte, error) {
	b, err := fp.ReadByte()
	if err == nil && b == '\n' {
		lineno++
	}
	return b, err
}
func unreadPLAByte(fp *bufio.Reader, b byte) {
	if err := fp.UnreadByte(); err != nil {
		panic(err)
	}
	if b == '\n' {
		lineno--
	}
}
func skip_line(input *bufio.Reader, output io.Writer, echo bool) {
	for {
		b, err := readPLAByte(input)
		if err != nil || b == '\n' {
			break
		}
		if echo {
			fmt.Fprintf(output, "%c", b)
		}
	}
	if echo {
		fmt.Fprintln(output)
	}
}
func get_word(fp *bufio.Reader) string {
	var b strings.Builder
	for {
		ch, err := readPLAByte(fp)
		if err != nil {
			return b.String()
		}
		if unicode.IsSpace(rune(ch)) {
			if b.Len() > 0 {
				if ch == '\n' {
					unreadPLAByte(fp, ch)
				}
				return b.String()
			}
			continue
		}
		b.WriteByte(ch)
	}
}
func plaInt(fp *bufio.Reader, context string) int32 {
	word := get_word(fp)
	n, err := strconv.ParseInt(word, 10, 32)
	if err != nil {
		fatal("error reading " + context)
	}
	return int32(n)
}
func ensurePLACovers(pla pPLA) {
	if pla.F == nil {
		pla.F = new_cover(10)
		pla.D = new_cover(10)
		pla.R = new_cover(10)
	}
}

// cubeInputChar ignores field separators, retaining the source's permissive
// handling of product terms that span lines (with one warning per input).
func cubeInputChar(fp *bufio.Reader) (byte, error) {
	for {
		b, err := readPLAByte(fp)
		if err != nil {
			return 0, err
		}
		switch b {
		case ' ', '|', '\t', '\r':
			continue
		case '\n':
			if !line_length_error {
				fmt.Fprintln(os.Stderr, "product term(s) span more than one line (warning only)")
			}
			line_length_error = true
			continue
		}
		return b, nil
	}
}
func read_cube(fp *bufio.Reader, pla pPLA) {
	cf, cr, cd := new_cube(), new_cube(), new_cube()
	saveF, saveD, saveR := false, false, false
	bad := func() {
		fmt.Fprintf(os.Stderr, "(warning): input line #%d ignored\n", lineno)
		skip_line(fp, os.Stdout, true)
	}
	for v := int32(0); v < cube.num_binary_vars; v++ {
		ch, err := cubeInputChar(fp)
		if err != nil {
			bad()
			return
		}
		switch ch {
		case '2', '-':
			set_insert(cf, 2*v)
			set_insert(cf, 2*v+1)
		case '0':
			set_insert(cf, 2*v)
		case '1':
			set_insert(cf, 2*v+1)
		case '?':
		default:
			bad()
			return
		}
	}
	for v := cube.num_binary_vars; v < cube.num_vars-1; v++ {
		if cube.part_size[v] < 0 {
			token := get_word(fp)
			switch token {
			case "-", "ANY":
				if !kiss || v != cube.num_vars-2 {
					set_or(cf, cf, cube.var_mask[v])
				}
			case "~":
			default:
				variable, offset := v, int32(0)
				if kiss && v == cube.num_vars-2 {
					variable = v - 1
					offset = cube.last_part[variable] - cube.first_part[variable] + 1
				}
				found := false
				for i := cube.first_part[variable]; i <= cube.last_part[variable]; i++ {
					if pla.label[i] == "" {
						pla.label[i] = token
					}
					if pla.label[i] == token {
						set_insert(cf, i+offset)
						found = true
						break
					}
				}
				if !found {
					fatal(fmt.Sprintf("declared size of variable %d (counting from variable 0) is too small", v))
				}
			}
		} else {
			for i := cube.first_part[v]; i <= cube.last_part[v]; i++ {
				ch, err := cubeInputChar(fp)
				if err != nil {
					bad()
					return
				}
				switch ch {
				case '1':
					set_insert(cf, i)
				case '0':
				default:
					bad()
					return
				}
			}
		}
	}
	if cube.output == -1 {
		pla.F = sf_addset(pla.F, cf)
		return
	}
	if kiss {
		saveF, saveR = true, true
		set_xor(cr, cf, cube.var_mask[cube.num_vars-2])
	} else {
		set_copy(cr, cf)
	}
	set_copy(cd, cf)
	for i := cube.first_part[cube.output]; i <= cube.last_part[cube.output]; i++ {
		ch, err := cubeInputChar(fp)
		if err != nil {
			bad()
			return
		}
		switch ch {
		case '4', '1':
			if pla.pla_type&F_type != 0 {
				set_insert(cf, i)
				saveF = true
			}
		case '3', '0':
			if pla.pla_type&R_type != 0 {
				set_insert(cr, i)
				saveR = true
			}
		case '2', '-':
			if pla.pla_type&D_type != 0 {
				set_insert(cd, i)
				saveD = true
			}
		case '~':
		default:
			bad()
			return
		}
	}
	if saveF {
		pla.F = sf_addset(pla.F, cf)
	}
	if saveD {
		pla.D = sf_addset(pla.D, cd)
	}
	if saveR {
		pla.R = sf_addset(pla.R, cr)
	}
}
func requirePLASize(command string) {
	if cube.fullset == nil {
		fatal("PLA size must be declared before ." + command)
	}
}
func appendSymbolic(head **symbolic_t, node *symbolic_t) {
	if *head == nil {
		*head = node
		return
	}
	tail := *head
	for tail.next != nil {
		tail = tail.next
	}
	tail.next = node
}
func parse_pla(fp *bufio.Reader, pla pPLA) {
	lineno = 1
	line_length_error = false
	for {
		ch, err := readPLAByte(fp)
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			return
		}
		if unicode.IsSpace(rune(ch)) {
			continue
		}
		if ch == '#' {
			unreadPLAByte(fp, ch)
			skip_line(fp, os.Stdout, echo_comments)
			continue
		}
		if ch != '.' {
			unreadPLAByte(fp, ch)
			if cube.fullset == nil {
				if echo_comments {
					fmt.Print("#")
				}
				skip_line(fp, os.Stdout, echo_comments)
				continue
			}
			ensurePLACovers(pla)
			read_cube(fp, pla)
			continue
		}
		command := get_word(fp)
		switch command {
		case "i":
			if cube.fullset != nil {
				fmt.Fprintln(os.Stderr, "extra .i ignored")
				skip_line(fp, os.Stdout, false)
				continue
			}
			cube.num_binary_vars = plaInt(fp, ".i")
			if cube.num_binary_vars < 0 || cube.num_binary_vars > 16368 {
				fatal("invalid binary variable count")
			}
			cube.num_vars = cube.num_binary_vars + 1
			cube.part_size = make([]int32, cube.num_vars)
		case "o":
			if cube.fullset != nil {
				fmt.Fprintln(os.Stderr, "extra .o ignored")
				skip_line(fp, os.Stdout, false)
				continue
			}
			if cube.part_size == nil {
				fatal(".o cannot appear before .i")
			}
			cube.part_size[cube.num_vars-1] = plaInt(fp, ".o")
			cube_setup()
			PLA_labels(pla)
		case "mv":
			if cube.fullset != nil {
				fmt.Fprintln(os.Stderr, "extra .mv ignored")
				skip_line(fp, os.Stdout, false)
				continue
			}
			if cube.part_size != nil {
				fatal("cannot mix .i and .mv")
			}
			cube.num_vars = plaInt(fp, ".mv")
			cube.num_binary_vars = plaInt(fp, ".mv")
			if cube.num_binary_vars < 0 || cube.num_vars < cube.num_binary_vars || cube.num_vars > 32736 {
				fatal("invalid variable counts in .mv")
			}
			cube.part_size = make([]int32, cube.num_vars)
			for v := cube.num_binary_vars; v < cube.num_vars; v++ {
				cube.part_size[v] = plaInt(fp, ".mv")
			}
			cube_setup()
			PLA_labels(pla)
		case "p":
			plaInt(fp, ".p")
		case "e", "end":
			if cube.fullset != nil {
				ensurePLACovers(pla)
			}
			return
		case "kiss":
			kiss = true
		case "type":
			word := get_word(fp)
			found := false
			for _, entry := range pla_types {
				if strings.TrimPrefix(entry.key, "-") == word && entry.key != "" {
					pla.pla_type = entry.value
					found = true
					break
				}
			}
			if !found {
				fatal("unknown type in .type command")
			}
		case "ilb":
			requirePLASize(command)
			if pla.label == nil {
				PLA_labels(pla)
			}
			for v := int32(0); v < cube.num_binary_vars; v++ {
				word := get_word(fp)
				pla.label[cube.first_part[v]+1] = word
				pla.label[cube.first_part[v]] = word + ".bar"
			}
		case "ob":
			requirePLASize(command)
			if pla.label == nil {
				PLA_labels(pla)
			}
			if cube.output < 0 {
				fatal(".ob requires an output variable")
			}
			for i := cube.first_part[cube.output]; i <= cube.last_part[cube.output]; i++ {
				pla.label[i] = get_word(fp)
			}
		case "label":
			requirePLASize(command)
			if pla.label == nil {
				PLA_labels(pla)
			}
			word := get_word(fp)
			number, err := strconv.ParseInt(strings.TrimPrefix(word, "var="), 10, 32)
			if err != nil || !strings.HasPrefix(word, "var=") || number < 0 || number >= int64(cube.num_vars) {
				fatal("Error reading labels")
			}
			v := int32(number)
			for i := cube.first_part[v]; i <= cube.last_part[v]; i++ {
				pla.label[i] = get_word(fp)
			}
		case "symbolic", "symbolic-output":
			var node *symbolic_t
			if !read_symbolic(fp, pla, &node) {
				fatal("error reading ." + command)
			}
			if command == "symbolic" {
				appendSymbolic(&pla.symbolic, node)
			} else {
				appendSymbolic(&pla.symbolic_output, node)
			}
		case "phase":
			requirePLASize(command)
			if pla.phase != nil {
				fmt.Fprintln(os.Stderr, "extra .phase ignored")
				skip_line(fp, os.Stdout, false)
				continue
			}
			if cube.output < 0 {
				fatal(".phase requires an output variable")
			}
			word := get_word(fp)
			if len(word) != int(cube.part_size[cube.output]) {
				fatal("incorrect number of phase bits")
			}
			pla.phase = set_save(cube.fullset)
			for index, ch := range word {
				if ch == '0' {
					set_remove(pla.phase, cube.first_part[cube.output]+int32(index))
				} else if ch != '1' {
					fatal("only 0 or 1 allowed in phase description")
				}
			}
		case "pair":
			if pla.pair != nil {
				fmt.Fprintln(os.Stderr, "extra .pair ignored")
				skip_line(fp, os.Stdout, false)
				continue
			}
			count := plaInt(fp, ".pair")
			if count < 0 || count > cube.num_binary_vars/2 {
				fatal("invalid .pair count")
			}
			pair := &pair_t{cnt: count, var1: make([]int32, count), var2: make([]int32, count)}
			for i := int32(0); i < count; i++ {
				for _, target := range []*int32{&pair.var1[i], &pair.var2[i]} {
					word := strings.Trim(get_word(fp), "()")
					var variable, part int32
					if !label_index(pla, word, &variable, &part) || variable < 0 || variable >= cube.num_binary_vars {
						fatal("syntax error in .pair")
					}
					*target = variable + 1
				}
			}
			pla.pair = pair
		default:
			if echo_unknown_commands {
				fmt.Printf(".%s ", command)
			}
			skip_line(fp, os.Stdout, echo_unknown_commands)
		}
	}
}
func read_symbolic(fp *bufio.Reader, pla pPLA, result **symbolic_t) bool {
	node := &symbolic_t{}
	var tail *symbolic_list_t
	for {
		word := get_word(fp)
		if word == ";" {
			break
		}
		if word == "" {
			return false
		}
		var variable, part int32
		if !label_index(pla, word, &variable, &part) {
			return false
		}
		p := &symbolic_list_t{variable: variable, pos: part}
		if tail == nil {
			node.symbolic_list = p
		} else {
			tail.next = p
		}
		tail = p
		node.symbolic_list_length++
	}
	var labelTail *symbolic_label_t
	for {
		word := get_word(fp)
		if word == ";" {
			break
		}
		if word == "" {
			return false
		}
		p := &symbolic_label_t{label: word}
		if labelTail == nil {
			node.symbolic_label = p
		} else {
			labelTail.next = p
		}
		labelTail = p
		node.symbolic_label_length++
	}
	*result = node
	return true
}
func label_index(pla pPLA, word string, variable, part *int32) bool {
	if len(pla.label) == 0 || pla.label[0] == "" {
		number, err := strconv.ParseInt(word, 10, 32)
		if err != nil {
			return false
		}
		*variable = int32(number)
		*part = int32(number)
		return true
	}
	for v := int32(0); v < cube.num_vars; v++ {
		size := cube.last_part[v] - cube.first_part[v] + 1
		for i := int32(0); i < size; i++ {
			if pla.label[cube.first_part[v]+i] == word {
				*variable = v
				*part = i
				return true
			}
		}
	}
	return false
}
