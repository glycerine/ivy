package abci

import abcinput "github.com/glycerine/abcgo/base/io"

func init() { abcinput.RegisterDarCNFWriter(Abc_NtkDarToCnf) }
