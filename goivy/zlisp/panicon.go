package zlisp

func panicOn(err error) {
	if err != nil {
		panic(err)
	}
}
