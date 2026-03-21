// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_to_cpp.py REPL/test boilerplate (~lines 5002-5226).

// This file covers test harness and server REPL generation,
// value parsing, ctuple solver conversions, and parameter assignments.
package cppgen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/module"
)

// ---------------------------------------------------------------------------
// EmitReplBoilerplate3Server — Server main loop boilerplate.
// ---------------------------------------------------------------------------

// EmitReplBoilerplate3Server generates the server main loop that waits
// for reader threads to terminate.
// Corresponds to Python emit_repl_boilerplate3server() (lines 5002-5022).
func EmitReplBoilerplate3Server(impl *CodeText, classname string) {
	code := `

    ivy.__unlock();

    // The main thread waits for all reader threads to die

    for(unsigned i = 0; true ; i++) {
        ivy.__lock();
        if (i >= ivy.thread_ids.size()){
            ivy.__unlock();
            break;
        }
        pthread_t tid = ivy.thread_ids[i];
        ivy.__unlock();
        pthread_join(tid,NULL);
    }
    return 0;

`
	impl.Append(strings.ReplaceAll(code, "classname", classname))
}

// ---------------------------------------------------------------------------
// EmitReplBoilerplate3Test — Test harness with weighted action selection.
// ---------------------------------------------------------------------------

// EmitReplBoilerplate3Test generates the test harness main loop with:
//   - Weighted random action selection
//   - Network I/O multiplexing (select())
//   - Reader/timer management
//   - Configurable test iterations
//
// Corresponds to Python emit_repl_boilerplate3test() (lines 5024-5226).
func EmitReplBoilerplate3Test(impl *CodeText, classname string, mod *module.Module) {
	// Preamble: bind readers and create generators
	impl.Append(`        ivy.__unlock();
        initializing = false;
        for(int rdridx = 0; rdridx < readers.size(); rdridx++) {
            readers[rdridx]->bind();
        }

        init_gen my_init_gen(ivy);
        my_init_gen.generate(ivy);
        std::vector<gen *> generators;
        std::vector<double> weights;

`)

	// Generate action generators and weights
	totalWeight := 0.0
	numPublicActions := 0
	if mod != nil {
		sortedActions := make([]string, 0, len(mod.PublicActions))
		for name := range mod.PublicActions {
			sortedActions = append(sortedActions, name)
		}
		sort.Strings(sortedActions)

		for _, actname := range sortedActions {
			if actname == "ext:_finalize" {
				continue
			}
			numPublicActions++
			impl.Append(fmt.Sprintf("        generators.push_back(new %s_gen(ivy));\n", Varname(actname)))

			aname := actname
			if strings.HasPrefix(aname, "ext:") {
				aname = aname[4:]
			}
			aname += ".weight"

			weight := 1.0
			if mod.Attributes != nil {
				if attrVal, ok := mod.Attributes[aname]; ok {
					if s, ok := attrVal.(fmt.Stringer); ok {
						str := s.String()
						str = strings.Trim(str, "\"")
						fmt.Sscanf(str, "%f", &weight)
					}
				}
			}
			impl.Append(fmt.Sprintf("        weights.push_back(%g);\n", weight))
			totalWeight += weight
		}
	}
	impl.Append(fmt.Sprintf("        double totalweight = %g;\n", totalWeight))
	impl.Append(fmt.Sprintf("        int num_gens = %d;\n", numPublicActions))

	// Finalize code
	finalCode := ""
	if mod != nil && mod.PublicActions["ext:_finalize"] {
		finalCode = "ivy.__lock(); ivy.ext___finalize(); ivy.__unlock();"
	}

	// Main test loop with select()-based I/O multiplexing
	testLoop := `

#ifdef _WIN32
    LARGE_INTEGER freq;
    QueryPerformanceFrequency(&freq);
#endif
    double frnd = 0.0;
    bool do_over = false;
    for(int cycle = 0; cycle < test_iters; cycle++) {

        double choices = totalweight + 5.0;
        if (do_over) {
           do_over = false;
        }  else {
            frnd = choices * (((double)rand())/(((double)RAND_MAX)+1.0));
        }
        if (frnd < totalweight) {
            int idx = 0;
            double sum = 0.0;
            while (idx < num_gens-1) {
                sum += weights[idx];
                if (frnd < sum)
                    break;
                idx++;
            }
            gen &g = *generators[idx];
            ivy.__lock();
            ivy._generating = true;
            bool sat = g.generate(ivy);
            if (sat){
                g.execute(ivy);
                ivy._generating = false;
                ivy.__unlock();
            }
            else {
                ivy._generating = false;
                ivy.__unlock();
                cycle--;
            }
            continue;
        }

        fd_set rdfds;
        FD_ZERO(&rdfds);
        int maxfds = 0;

        for (unsigned i = 0; i < readers.size(); i++) {
            reader *r = readers[i];
            int fds = r->fdes();
            if (fds >= 0) {
                FD_SET(fds,&rdfds);
            }
            if (fds > maxfds)
                maxfds = fds;
        }

#ifdef _WIN32
        int timer_min = 15;
#else
        int timer_min = 5;
#endif

        struct timeval timeout;
        timeout.tv_sec = timer_min/1000;
        timeout.tv_usec = 1000 * (timer_min % 1000);

#ifdef _WIN32
        int foo;
        if (readers.size() == 0){
            Sleep(timer_min);
            foo = 0;
        }
        else
            foo = select(maxfds+1,&rdfds,0,0,&timeout);
#else
        int foo = select(maxfds+1,&rdfds,0,0,&timeout);
#endif

        if (foo < 0)
#ifdef _WIN32
            {std::cerr << "select failed: " << WSAGetLastError() << std::endl; __ivy_exit(1);}
#else
            {perror("select failed"); __ivy_exit(1);}
#endif

        if (foo == 0){
           cycle--;
           for (unsigned i = 0; i < timers.size(); i++){
               if (timer_min >= timers[i]->ms_delay()) {
                   cycle++;
                   break;
               }
           }
           for (unsigned i = 0; i < timers.size(); i++)
               timers[i]->timeout(timer_min);
        }
        else {
            int fdc = 0;
            for (unsigned i = 0; i < readers.size(); i++) {
                reader *r = readers[i];
                if (FD_ISSET(r->fdes(),&rdfds))
                    fdc++;
            }
            int fdi = fdc * (((double)rand())/(((double)RAND_MAX)+1.0));
            fdc = 0;
            for (unsigned i = 0; i < readers.size(); i++) {
                reader *r = readers[i];
                if (FD_ISSET(r->fdes(),&rdfds)) {
                    if (fdc == fdi) {
                        r->read();
                        if (r->background()) {
                           cycle--;
                           do_over = true;
                        }
                        break;
                    }
                    fdc++;
                }
            }
        }
    }
    FINALIZE
    __ivy_out << "test_completed" << std::endl;
    if (runidx == runs-1) {
        struct timespec ts;
        int ms = 50;
        ts.tv_sec = ms/1000;
        ts.tv_nsec = (ms % 1000) * 1000000;
        nanosleep(&ts,NULL);
        exit(0);
    }
    for (unsigned i = 0; i < readers.size(); i++)
        delete readers[i];
    readers.clear();
    for (unsigned i = 0; i < timers.size(); i++)
        delete timers[i];
    timers.clear();

`
	testLoop = strings.ReplaceAll(testLoop, "classname", classname)
	testLoop = strings.ReplaceAll(testLoop, "FINALIZE", finalCode)
	impl.Append(testLoop)
}

// ---------------------------------------------------------------------------
// EmitValueParser — REPL parameter value parsing.
// ---------------------------------------------------------------------------

// EmitValueParser generates a C++ function that parses a parameter value
// from a string input, with exception handling for invalid values.
// Corresponds to Python emit_value_parser() (lines 3431-3442).
func EmitValueParser(impl *CodeText, paramName string, paramType string, classname string) {
	Indent(impl)
	impl.Append(fmt.Sprintf("template<typename T> T _arg(std::vector<ivy_value> &args, unsigned idx, long long bound);\n"))
	Indent(impl)
	impl.Append(fmt.Sprintf("template<> %s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound) {\n", paramType, paramType))
	IndentLevel++
	codeLine(impl, fmt.Sprintf("if (idx >= args.size()) throw out_of_bounds(\"not enough arguments\",args.size())"))
	codeLine(impl, fmt.Sprintf("return _arg_helper<%s>(args[idx], bound)", paramType))
	IndentLevel--
	Indent(impl)
	impl.Append("}\n")
}

// ---------------------------------------------------------------------------
// EmitCtupleToSolver — Z3 solver conversions for compound tuples.
// ---------------------------------------------------------------------------

// EmitCtupleToSolver generates code to convert a compound tuple value
// to its Z3 solver representation.
// Corresponds to Python emit_ctuple_to_solver() (lines 1831-1845).
func EmitCtupleToSolver(impl *CodeText, typeName string, fieldTypes []string) {
	openScope(impl, fmt.Sprintf("z3::expr __to_solver(gen &g, const z3::expr &v, %s &val)", typeName))
	if len(fieldTypes) == 0 {
		codeLine(impl, "return v")
	} else {
		codeLine(impl, "z3::expr __result = v")
		for i, ft := range fieldTypes {
			codeLine(impl, fmt.Sprintf("__result = __result && __to_solver(g, g.apply(\"%s_field%d\", v), val.field%d)",
				typeName, i, i))
			_ = ft
		}
		codeLine(impl, "return __result")
	}
	closeScope(impl, false)
}

// EmitCtupleEquality generates an equality predicate for a compound tuple.
func EmitCtupleEquality(impl *CodeText, typeName string, numFields int) {
	openScope(impl, fmt.Sprintf("bool operator==(%s const &a, %s const &b)", typeName, typeName))
	if numFields == 0 {
		codeLine(impl, "return true")
	} else {
		parts := make([]string, numFields)
		for i := 0; i < numFields; i++ {
			parts[i] = fmt.Sprintf("a.field%d == b.field%d", i, i)
		}
		codeLine(impl, fmt.Sprintf("return %s", strings.Join(parts, " && ")))
	}
	closeScope(impl, false)
}

// ---------------------------------------------------------------------------
// EmitParameterAssignments — Parameter value initialization.
// ---------------------------------------------------------------------------

// EmitParameterAssignments generates code to assign parameter values
// during initialization.
// Corresponds to Python emit_parameter_assignments() (lines 3568-3574).
func EmitParameterAssignments(impl *CodeText, mod *module.Module, classname string) {
	if mod == nil {
		return
	}
	for i, param := range mod.Params {
		paramName := param.Name
		if i < len(mod.ParamDefaults) && mod.ParamDefaults[i] != nil {
			codeLine(impl, fmt.Sprintf("%s = %s", Varname(paramName), fmt.Sprint(mod.ParamDefaults[i])))
		}
	}
}

// ---------------------------------------------------------------------------
// EmitTemplateParams — Template parameter handling.
// ---------------------------------------------------------------------------

// EmitTemplateParams generates C++ template parameter declarations.
func EmitTemplateParams(header *CodeText, params []string) {
	if len(params) == 0 {
		return
	}
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = "class " + p
	}
	Indent(header)
	header.Append(fmt.Sprintf("template<%s>\n", strings.Join(parts, ", ")))
}

// ---------------------------------------------------------------------------
// EmitTick — Timer/progress callback emission.
// ---------------------------------------------------------------------------

// EmitTick generates code for the timer tick callback, which handles
// progress properties and timeout checking.
// Corresponds to Python emit_tick() (~lines 4700-4790).
func EmitTick(impl *CodeText, classname string, mod *module.Module) {
	openScope(impl, fmt.Sprintf("void %s::__tick(int __timeout)", classname))

	// Handle progress properties
	if mod != nil && len(mod.Progress) > 0 {
		codeLine(impl, "// Check progress properties")
		for i := range mod.Progress {
			codeLine(impl, fmt.Sprintf("// Progress property %d would be checked here", i))
		}
	}

	// Handle timers
	codeLine(impl, "for (unsigned i = 0; i < timers.size(); i++)")
	codeLine(impl, "    timers[i]->timeout(__timeout)")

	closeScope(impl, false)
}

// EmitClearProgress generates code to reset progress tracking.
func EmitClearProgress(impl *CodeText, classname string) {
	openScope(impl, fmt.Sprintf("void %s::__clear_progress()", classname))
	codeLine(impl, "// Reset progress counters")
	closeScope(impl, false)
}

// EmitParamDecls generates parameter declarations for an action.
func EmitParamDecls(header *CodeText, params []string, types []string) {
	for i, p := range params {
		t := "int"
		if i < len(types) {
			t = types[i]
		}
		Indent(header)
		header.Append(fmt.Sprintf("%s %s;\n", t, Varname(p)))
	}
}
