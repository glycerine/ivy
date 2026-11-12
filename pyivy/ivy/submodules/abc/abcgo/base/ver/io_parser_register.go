package ver

import (
	"github.com/glycerine/abcgo/base/abc"
	abcinput "github.com/glycerine/abcgo/base/io"
)

func init() {
	abcinput.RegisterVerilogParser(func(filename string, check int32) *abc.Design {
		return Ver_ParseFile(filename, nil, check, 1)
	})
}
