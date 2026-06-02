all:
	rm -rf ./connection_mgmt.cpp ./connection_mgmt.h connection_mgmt.dSYM connection_mgmt.dsc
	XTRACE_OFF=1 ivy2cpp target=test build=true connection_mgmt.ivy 
	./connection_mgmt

py:
	rm -rf ./connection_mgmt.cpp ./connection_mgmt.h connection_mgmt.dSYM connection_mgmt.dsc
	XTRACE_OFF=1 ivy_to_cpp target=test build=true connection_mgmt.ivy
	./connection_mgmt
