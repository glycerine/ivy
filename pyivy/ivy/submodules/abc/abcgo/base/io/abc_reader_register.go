package io

import "github.com/glycerine/abcgo/base/abc"

func init() {
	abc.RegisterNetworkReader(func(filename string, check int32) *abc.Network {
		return Io_Read(filename, Io_ReadFileType(filename), check, 0)
	})
}
