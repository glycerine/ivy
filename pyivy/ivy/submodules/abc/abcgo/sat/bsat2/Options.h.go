package bsat2

import (
	"flag"
	"fmt"
	"io"
	"math"
)

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

type Options struct {
	flags      *flag.FlagSet
	validators []func() error
}

func NewOptions(name string, output io.Writer) *Options {
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.SetOutput(output)
	return &Options{flags: f}
}
func (o *Options) Bool(name string, value bool, help string) *bool {
	return o.flags.Bool(name, value, help)
}
func (o *Options) String(name, value, help string) *string { return o.flags.String(name, value, help) }
func (o *Options) Int(name string, value, minimum, maximum int, help string) *int {
	p := o.flags.Int(name, value, help)
	o.validators = append(o.validators, func() error {
		if *p < minimum || *p > maximum {
			return fmt.Errorf("%s outside [%d,%d]", name, minimum, maximum)
		}
		return nil
	})
	return p
}
func (o *Options) Float64(name string, value, minimum, maximum float64, help string) *float64 {
	p := o.flags.Float64(name, value, help)
	o.validators = append(o.validators, func() error {
		if math.IsNaN(*p) || *p < minimum || *p > maximum {
			return fmt.Errorf("%s outside [%g,%g]", name, minimum, maximum)
		}
		return nil
	})
	return p
}
func (o *Options) PrintHelp() { o.flags.PrintDefaults() }
