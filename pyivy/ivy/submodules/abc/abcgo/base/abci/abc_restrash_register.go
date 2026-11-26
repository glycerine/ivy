package abci

import "github.com/glycerine/abcgo/base/abc"

func init() { abc.RestrashHook = Abc_NtkRestrash }
