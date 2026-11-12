package glucose2

/***************************************************************************************[Options.h]
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
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

type IntRange struct{ Begin, End int32 }
type Int64Range struct{ Begin, End int64 }
type DoubleRange struct {
	Begin, End                   float64
	BeginInclusive, EndInclusive bool
}
type OptionInfo struct{ Name, Description, Category, TypeName string }
type Option interface {
	Info() OptionInfo
	Parse(string) (bool, error)
	Help(io.Writer, bool) error
}

func (o OptionInfo) Info() OptionInfo             { return o }
func optionValue(arg, name string) (string, bool) { return strings.CutPrefix(arg, "-"+name+"=") }
func optionHelp(w io.Writer, o OptionInfo, value any, verbose bool) error {
	if _, err := fmt.Fprintf(w, "  -%-12s = %-8s (default: %v)\n", o.Name, o.TypeName, value); err != nil {
		return err
	}
	if verbose {
		_, err := fmt.Fprintf(w, "\n        %s\n\n", o.Description)
		return err
	}
	return nil
}

type IntOption struct {
	OptionInfo
	Value int32
	Range IntRange
}

func NewIntOption(category, name, description string, value int32, ranges ...IntRange) *IntOption {
	r := IntRange{math.MinInt32, math.MaxInt32}
	if len(ranges) > 0 {
		r = ranges[0]
	}
	o := &IntOption{OptionInfo{name, description, category, "<int32>"}, value, r}
	DefaultOptions.Add(o)
	return o
}
func (o *IntOption) Parse(arg string) (bool, error) {
	text, ok := optionValue(arg, o.Name)
	if !ok {
		return false, nil
	}
	n, err := strconv.ParseInt(text, 10, 32)
	if err != nil {
		return true, err
	}
	if n < int64(o.Range.Begin) || n > int64(o.Range.End) {
		return true, fmt.Errorf("option %s outside [%d,%d]", o.Name, o.Range.Begin, o.Range.End)
	}
	o.Value = int32(n)
	return true, nil
}
func (o *IntOption) Help(w io.Writer, verbose bool) error {
	return optionHelp(w, o.OptionInfo, o.Value, verbose)
}

type Int64Option struct {
	OptionInfo
	Value int64
	Range Int64Range
}

func NewInt64Option(category, name, description string, value int64, ranges ...Int64Range) *Int64Option {
	r := Int64Range{math.MinInt64, math.MaxInt64}
	if len(ranges) > 0 {
		r = ranges[0]
	}
	o := &Int64Option{OptionInfo{name, description, category, "<int64>"}, value, r}
	DefaultOptions.Add(o)
	return o
}
func (o *Int64Option) Parse(arg string) (bool, error) {
	text, ok := optionValue(arg, o.Name)
	if !ok {
		return false, nil
	}
	n, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return true, err
	}
	if n < o.Range.Begin || n > o.Range.End {
		return true, fmt.Errorf("option %s outside [%d,%d]", o.Name, o.Range.Begin, o.Range.End)
	}
	o.Value = n
	return true, nil
}
func (o *Int64Option) Help(w io.Writer, verbose bool) error {
	return optionHelp(w, o.OptionInfo, o.Value, verbose)
}

type DoubleOption struct {
	OptionInfo
	Value float64
	Range DoubleRange
}

func NewDoubleOption(category, name, description string, value float64, ranges ...DoubleRange) *DoubleOption {
	r := DoubleRange{Begin: math.Inf(-1), End: math.Inf(1)}
	if len(ranges) > 0 {
		r = ranges[0]
	}
	o := &DoubleOption{OptionInfo{name, description, category, "<double>"}, value, r}
	DefaultOptions.Add(o)
	return o
}
func (o *DoubleOption) Parse(arg string) (bool, error) {
	text, ok := optionValue(arg, o.Name)
	if !ok {
		return false, nil
	}
	n, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return true, err
	}
	r := o.Range
	if math.IsNaN(n) || n < r.Begin || n > r.End || n == r.Begin && !r.BeginInclusive || n == r.End && !r.EndInclusive {
		return true, fmt.Errorf("option %s outside allowed range", o.Name)
	}
	o.Value = n
	return true, nil
}
func (o *DoubleOption) Help(w io.Writer, verbose bool) error {
	return optionHelp(w, o.OptionInfo, o.Value, verbose)
}

type BoolOption struct {
	OptionInfo
	Value bool
}

func NewBoolOption(category, name, description string, value bool) *BoolOption {
	o := &BoolOption{OptionInfo{name, description, category, "<bool>"}, value}
	DefaultOptions.Add(o)
	return o
}
func (o *BoolOption) Parse(arg string) (bool, error) {
	switch arg {
	case "-" + o.Name:
		o.Value = true
		return true, nil
	case "-no-" + o.Name:
		o.Value = false
		return true, nil
	}
	return false, nil
}
func (o *BoolOption) Help(w io.Writer, verbose bool) error {
	return optionHelp(w, o.OptionInfo, o.Value, verbose)
}

type StringOption struct {
	OptionInfo
	Value string
}

func NewStringOption(category, name, description, value string) *StringOption {
	o := &StringOption{OptionInfo{name, description, category, "<string>"}, value}
	DefaultOptions.Add(o)
	return o
}
func (o *StringOption) Parse(arg string) (bool, error) {
	value, ok := optionValue(arg, o.Name)
	if ok {
		o.Value = value
	}
	return ok, nil
}
func (o *StringOption) Help(w io.Writer, verbose bool) error {
	return optionHelp(w, o.OptionInfo, o.Value, verbose)
}
