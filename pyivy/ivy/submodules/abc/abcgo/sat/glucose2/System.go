package glucose2

/****************************************************************************************[System.h]
Copyright (c) 2003-2006, Niklas Een, Niklas Sorensson
Copyright (c) 2007-2010, Niklas Sorensson

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
	"os"
	"strconv"
	"strings"
)

// CPUTime returns process user CPU seconds on Linux, matching the source's
// getrusage ru_utime. /proc exports USER_HZ ticks (100 on Go-supported Linux
// architectures). Unsupported hosts return zero instead of wall-clock time.
func CPUTime() float64 {
	data, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		return 0
	}
	end := strings.LastIndexByte(string(data), ')')
	if end < 0 {
		return 0
	}
	fields := strings.Fields(string(data[end+1:]))
	if len(fields) <= 11 {
		return 0
	}
	ticks, err := strconv.ParseFloat(fields[11], 64)
	if err != nil {
		return 0
	}
	return ticks / 100
}
