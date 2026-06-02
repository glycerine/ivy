all:
	rm -rf ./issue57.cpp ./issue57.h issue57.dSYM issue57.dsc
	XTRACE_OFF=1 ivy2cpp target=test build=true issue57.ivy 
	./issue57

py:
	rm -rf ./issue57.cpp ./issue57.h issue57.dSYM issue57.dsc
	XTRACE_OFF=1 ivy_to_cpp target=test build=true issue57.ivy
	./issue57
