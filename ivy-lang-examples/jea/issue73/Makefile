all:
	rm -rf ./ord.cpp ./ord.h ord.dSYM ord.dsc ord
	XTRACE_OFF=1 ivy2cpp target=test build=true ord.ivy 
	./ord

py:
	rm -rf ./ord.cpp ./ord.h ord.dSYM ord.dsc ord
	XTRACE_OFF=1 ivy_to_cpp target=test build=true ord.ivy
	./ord
