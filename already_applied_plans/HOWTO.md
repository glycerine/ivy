how to check a proof

~~~

jaten@jbook ~/goivy (master) $ source /Users/jaten/pyivy/venv/bin/activate

(venv) jaten@jbook ~/goivy (master) $ ivy isolate=cf_live  /Users/jaten/go/src/github.com/glycerine/goivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy
Traceback (most recent call last):
  File "/Users/jaten/pyivy/venv/bin/ivy", line 8, in <module>
    sys.exit(main())
  File "/Users/jaten/pyivy/ivy/ivy/ivy.py", line 14, in main
    ui_main_loop(ivy_init())
  File "/Users/jaten/pyivy/ivy/ivy/tk_ui.py", line 319, in ui_main_loop
    ivy_ui.ui.add(art)
  File "/Users/jaten/pyivy/ivy/ivy/tk_ui.py", line 100, in add
    gw = tk_ag_ui(tk,art,frame)
  File "/Users/jaten/pyivy/ivy/ivy/tk_ui.py", line 229, in __init__
    self.rebuild()
  File "/Users/jaten/pyivy/ivy/ivy/tk_ui.py", line 256, in rebuild
    self.create_elements(self.g.as_cy_elements(dot_layout))
  File "/Users/jaten/pyivy/ivy/ivy/ivy_art.py", line 443, in as_cy_elements
    return dot_layout(render_rg(self),edge_labels=True)
  File "/Users/jaten/pyivy/ivy/ivy/dot_layout.py", line 234, in dot_layout
    g.layout(prog='dot')
  File "/Users/jaten/pyivy/ivy/ivy/ivy_graphviz.py", line 137, in layout
    self.g =  pydot.dot_parser.parse_dot_data(txt)[0]
AttributeError: module 'pydot' has no attribute 'dot_parser'
(venv) jaten@jbook ~/goivy (master) $ pip3 install "pydot==1.4.2"
Collecting pydot==1.4.2
  Downloading pydot-1.4.2-py2.py3-none-any.whl.metadata (8.0 kB)
Requirement already satisfied: pyparsing>=2.1.4 in /Users/jaten/pyivy/venv/lib/python3.10/site-packages/pyparsing-3.2.3-py3.10.egg (from pydot==1.4.2) (3.2.3)
Downloading pydot-1.4.2-py2.py3-none-any.whl (21 kB)
Installing collected packages: pydot
  Attempting uninstall: pydot
    Found existing installation: pydot 4.0.0
    Uninstalling pydot-4.0.0:
      Successfully uninstalled pydot-4.0.0
Successfully installed pydot-1.4.2

[notice] A new release of pip is available: 25.1.1 -> 26.0.1
[notice] To update, run: pip install --upgrade pip
(venv) jaten@jbook ~/goivy (master) $ ivy isolate=cf_live  /Users/jaten/go/src/github.com/glycerine/goivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy

(venv) jaten@jbook ~/goivy (master) $ ivy_check isolate=cf_live  /Users/jaten/go/src/github.com/glycerine/goivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy

 ivy_check: profiling option set to false. 

Isolate cf_live:

    The following properties are treated as axioms: 

    The following properties are treated as properties:
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 33: index.spec.prop4
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 4: index.spec.transitivity
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 5: index.spec.antisymmetry
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 6: index.spec.totality
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 11: index.spec.prop1
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 33: lclock.spec.prop4
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 4: lclock.spec.transitivity
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 5: lclock.spec.antisymmetry
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 6: lclock.spec.totality
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 11: lclock.spec.prop1
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 33: tar_clock.spec.prop4
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 4: tar_clock.spec.transitivity
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 5: tar_clock.spec.antisymmetry
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 6: tar_clock.spec.totality
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 11: tar_clock.spec.prop1
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 33: tar_cf_clock.spec.prop4
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 4: tar_cf_clock.spec.transitivity
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 5: tar_cf_clock.spec.antisymmetry
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 6: tar_cf_clock.spec.totality
        /Users/jaten/pyivy/ivy/ivy/include/1.8/order.ivy: line 11: tar_cf_clock.spec.prop1
        /Users/jaten/go/src/github.com/glycerine/goivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy: line 1694: cf_live.cf_liveness
        /Users/jaten/go/src/github.com/glycerine/goivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy: line 1896: rfn.abs.lemma1(M:mem_type)

    The following properties are treated as conjectures: 
...

~~~
