Avoid the use directories named "testdata", because the Go
fuzzer stores seeds/data there and then testdata/ could get wiped out;
so prefere test_vectors/ instead of testdata/ every time.
