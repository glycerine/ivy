package io

import "github.com/glycerine/abcgo/base/abc"

func Io_WriteBlifMv(n *abc.Network, filename string) error {
	return WriteBLIFFile(n, filename, true, true)
}
