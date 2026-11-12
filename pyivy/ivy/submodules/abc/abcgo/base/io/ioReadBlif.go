package io

import "github.com/glycerine/abcgo/base/abc"

func Io_ReadBlif(filename string, check int32) *abc.Network { return Io_ReadBlifMv(filename, 0, check) }
