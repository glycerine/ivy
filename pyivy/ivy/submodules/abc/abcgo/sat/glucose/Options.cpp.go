package glucose

/**************************************************************************************[Options.cc]
Copyright (c) 2008-2010, Niklas Sorensson

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

type OptionSet struct {
	Options           []Option
	Usage, HelpPrefix string
}

var DefaultOptions OptionSet
var ErrHelpRequested = errors.New("help requested")

func (s *OptionSet) Add(option Option) {
	for _, o := range s.Options {
		if o.Info().Name == option.Info().Name {
			panic("duplicate option name")
		}
	}
	s.Options = append(s.Options, option)
}

// Parse preserves argv[0] and positional arguments, like the source in-place
// parser. Help and invalid values return errors rather than exiting a library.
func (s *OptionSet) Parse(argv []string, strict bool) ([]string, error) {
	if len(argv) == 0 {
		return nil, nil
	}
	remaining := []string{argv[0]}
	for _, arg := range argv[1:] {
		if arg == "--"+s.HelpPrefix+"help" || arg == "--"+s.HelpPrefix+"help-verb" {
			return remaining, ErrHelpRequested
		}
		parsed := false
		for _, option := range s.Options {
			ok, err := option.Parse(arg)
			if err != nil {
				return remaining, fmt.Errorf("%s: %w", option.Info().Name, err)
			}
			if ok {
				parsed = true
				break
			}
		}
		if !parsed {
			if strict && strings.HasPrefix(arg, "-") {
				return remaining, fmt.Errorf("unknown flag %q", arg)
			}
			remaining = append(remaining, arg)
		}
	}
	return remaining, nil
}
func (s *OptionSet) PrintUsage(w io.Writer, program string, verbose bool) error {
	if s.Usage != "" {
		if _, err := io.WriteString(w, strings.ReplaceAll(s.Usage, "%s", program)); err != nil {
			return err
		}
	}
	options := append([]Option(nil), s.Options...)
	sort.SliceStable(options, func(i, j int) bool {
		a, b := options[i].Info(), options[j].Info()
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		return a.TypeName < b.TypeName
	})
	previous := ""
	for _, o := range options {
		info := o.Info()
		if info.Category != previous {
			if _, err := fmt.Fprintf(w, "\n%s OPTIONS:\n\n", info.Category); err != nil {
				return err
			}
			previous = info.Category
		}
		if err := o.Help(w, verbose); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "\nHELP OPTIONS:\n\n  --%shelp        Print help message.\n  --%shelp-verb   Print verbose help message.\n", s.HelpPrefix, s.HelpPrefix)
	return err
}
func ParseOptions(argv []string, strict bool) ([]string, error) {
	return DefaultOptions.Parse(argv, strict)
}
func SetUsageHelp(text string)     { DefaultOptions.Usage = text }
func SetHelpPrefixStr(text string) { DefaultOptions.HelpPrefix = text }
func PrintUsage(w io.Writer, program string, verbose bool) error {
	return DefaultOptions.PrintUsage(w, program, verbose)
}
