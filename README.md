GoIvy (Ivy in Go/as a Web application backed by a Go server)
======================================================

![an ivy covered place](https://github.com/glycerine/ivy/blob/master/reference/new_ivy_icon_starting_large.png)

This is a port of the Ivy formal verification toolchain to Go and a Web app, based
on the original python and Tk by Oded Padon and Ken McMillan et al 2016. It supports Ivy
language versions 1.6 and up (so 1.6, 1.7, and 1.8). See the Q & A below for
links and references.

A quick aside: Note that this project has nothing to do with Rob Pike's Ivy https://github.com/robpike/ivy calculator project. The repeat of the name Ivy is a mere coincidence. McMillan and Padon were thinking of "[I]nteractive [V]erification", and I doubt they were even aware of Mr. Pike's arbitrary-precision APL-like calculator project in Go.

GoIvy installation steps:

# binary installation (pre-built for amd64)

download a macOS (Sonoma 14.0/amd64) or Linux (Ubuntu 24.04/amd64) release 
from the releases page https://github.com/glycerine/ivy/releases

# building from source

pre-requisites to build from source: 

you should already have installed 

1. go1.26.4 ; get the pre-compiled Go for your platform from https://go.dev/dl/

2. python 3.10.18

3. clang or gnu C++ toolchain; 
to build Z3. I used Apple clang version 15.0.0 (clang-1500.3.9.4) on darwin.

4. emscripten 3.1.50 to produce wasm support, to do ivy-checks in-browser (they can
also be run on the server). emcc --version should show that emcc 3.1.50 is on your path;
emar and emmake and wasm-opt should all be on your $PATH.

Note that the required Z3 fork is vendored in this repo, in ivy/goivy/z3vendor/z3 (Z3 v4.7.1_425291ee2)
and is compiled below (building Z3 takes about 10-20 minutes, be patient).

~~~
export GOPATH=/home/yourUserName/go # where "go install" will put the goivy_check and ivyweb binaries

git clone https://github.com/glycerine/ivy
cd ivy/pyivy
# setup a python venv that knows how to find Ivy python programs
./redo.sh
# install abc and aigtoaig binaries that are shelled out to.
./onetime.sh # same as "make setup" from ivy/goivy

cd ../goivy # same as ivy/goivy
make get_goyacc
cd z3vendor/z3
./compile.sh # build libz3.a
./compile-wasm.sh # build wasm version of Z3
cd ../.. # so you should be in ivy/goivy now, using the goivy/Makefile to:
make z3-build
make build
~~~

binaries built:
~~~
# the main tools
ivyweb      (starts the Ivy web app on 127.0.0.1:8080 )
goivy_check (equivalent to python ivy_check, but 
             faster and with many bug fixes).

# for compositional testing
ivy2cpp
ivy2go
~~~

----

GoIvy author: Jason E. Aten, Ph.D.

License: Gnu Affero General Public License, version 3 (See https://www.gnu.org/licenses/agpl-3.0.html or the enclosed LICENSE file)

-------
original Ivy:

The MIT License for the original python and Tk Ivy (before updates
and fixes authored by me, Jason) can be viewed 
https://github.com/glycerine/ivy/blob/master/pyivy/LICENSE.original.ivy.txt
or here
https://raw.githubusercontent.com/kenmcmil/ivy/refs/heads/master/license.txt
and is as follows. It was committed in 0a6f6e707348f0e075a90440cbda39bbe39c5699 on  Dec 28, 2015 by github account kenmcmil in https://github.com/kenmcmil/ivy/commit/0a6f6e707348f0e075a90440cbda39bbe39c5699.

~~~
IVy-tool

Copyright (c) Microsoft Corporation

All rights reserved. 

MIT License

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
""Software""), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED *AS IS*, WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
~~~

Note that GoIvy and my (Jason E. Aten) updates, fixes, and works derived from and to python Ivy are licensed under the AGPL3 license and not the MIT license.

Questions and Answers
---------------------


# What is Ivy?

Q: Tell me about the history and current status of the Ivy project from Microsoft Research for formal verification. It looks to me like a real breakthough, but it is not widely used. Or perhaps that is a false negative since I have low visibility. Apple hardware engineers used it for broad liveness checking a memory subsystem design. But other uses are less visible.

A: (2026 March 22)

Your intuition about Ivy is well-founded — it really is a significant technical achievement, and the limited public visibility is partly a real gap in industrial adoption and partly a consequence of how serious industrial use tends to stay proprietary. Let me lay out the full picture.

---

**Origins and Design Philosophy**

Ivy was created primarily by Kenneth L. McMillan at Microsoft Research, in collaboration with researchers at Tel Aviv University (Oded Padon, Mooly Sagiv, Sharon Shoham) and UC Berkeley (Aurojit Panda). The seminal paper, "Ivy: Safety Verification by Interactive Generalization," appeared at PLDI 2016. The core motivation was that despite decades of research, the formal verification of infinite-state systems had resisted effective automation. Ivy's key principle is that whenever verification fails, it graphically displays a concrete counterexample to induction, and the user then interactively guides generalization from that counterexample.

The distinguishing intellectual contribution is the commitment to *decidable* logics. Ivy's language is designed to make it practical to reduce all proof obligations to statements in decidable logics. It provides primitives to support modeling in decidable logics and also modular refinement techniques that make it possible to separate the verification effort into local, decidable problems. This is the crux of what makes Ivy different from tools like Coq or Isabelle/HOL: rather than asking users to construct arbitrary proofs in expressive but undecidable logics, Ivy constrains the problem space so that automation remains reliable and predictable.

---

**Technical Capabilities**

Ivy is a multi-modal verification tool supporting modular specification, implementation, and proof. It supports proving safety and liveness properties of parameterized and infinite-state systems via three modes: deductive verification using an SMT solver, abstraction and model checking, and manual proofs using natural deduction. It also supports lightweight formal methods via compositional specification-based testing and bounded model checking. Ivy can extract executable distributed programs by translation to efficient C++ code.

The language design deliberately limits expressiveness to maintain decidability. There are no looping constructs in Ivy, because loops would make important verification problems undecidable. This is a philosophically interesting trade-off: accept a less convenient programming model in exchange for guarantees that your verification conditions can actually be discharged algorithmically.

---

**Notable Applications**

Ivy has been used to create verified implementations of distributed consensus protocols such as Paxos and Raft. It has also been used to create composable specifications of protocols such as the coherent memory interface of the RISC-V machine architecture and the QUIC Internet transport protocol — specifications that were used to rigorously test implementations, revealing significant errors in those implementations.

The QUIC work (published at SIGCOMM 2019) is particularly notable: Ivy's formal specification was used to generate randomized automated testers that exercised real QUIC implementations and found genuine bugs.

---

**The Apple Memory Subsystem Case**

Your information is correct and well-sourced. The inspiration to study Ivy's liveness proof capabilities came specifically from an effort to prove liveness of models of memory systems developed by hardware engineers at Apple, Inc. The Apple engineers use Ivy to prove safety properties of memory subsystem models.

The generic model consists of approximately 1,200 lines of Ivy code. The safety proof consists of 78 invariants comprising approximately 500 additional lines. It was also considered important to prove a liveness property — that every operation issued by a core is eventually retired. That liveness property is harder than safety, and the Apple work drove McMillan to develop new proof methods for liveness by relational rankings, published in 2024.

Apple provided a generic abstract memory system model designed to capture the essential difficulties in proving memory system consistency and liveness without revealing intellectual property. This model has been contributed to the open-source Ivy project, along with the proofs. So the artifact is publicly available even though the underlying Apple silicon design obviously is not.

---

**Why the Adoption Gap is Real, Not Just Low Visibility**

Your skepticism is healthy but the gap is genuine. Several factors explain it:

1. **Steep learning curve for a niche skill set.** Using Ivy effectively requires understanding both the decidable fragment of first-order logic that it targets and the protocol or system being verified. This combination is rare outside of formal methods specialists.

2. **Competition from lighter-weight tools.** For many industrial purposes, model checkers like TLA+/TLC, property checkers in EDA flows (JasperGold, Cadence Formal), or even fuzzing provide "good enough" assurance with shallower investment. Ivy's value proposition requires that you care about *proofs* of correctness, not just high-confidence testing.

3. **Research tool positioning.** McMillan himself has been clear that Ivy is a research tool intended to explore automated proof techniques. The toolchain, documentation, and ecosystem are not at the commercial robustness level of something like Dafny.

4. **McMillan's move to UT Austin.** After spending years at Microsoft Research, McMillan joined the University of Texas at Austin, which may have affected the pace of Ivy's tooling development, though research publication continues actively through 2024.

5. **Industrial use stays confidential.** The Apple case is the exception in being publicly documented (via an academic paper, not an Apple announcement). Other industrial uses, if they exist, would likely never be disclosed.

---

**Current Status (as of early 2026)**

The project remains active on GitHub under `https://github.com/kenmcmil/ivy`, with ongoing academic research. McMillan has continued publishing papers using and extending Ivy — recent work addresses liveness by relational rankings (the technique motivated by the Apple case), counterexample-driven quantifier instantiation for distributed protocols (OOPSLA 2023), and formal analysis of networking algorithms. The academic community continues to cite and build on Ivy, particularly in the distributed systems verification space.

The bottom line: Ivy represents a genuine breakthrough in making formal verification of parameterized infinite-state systems tractable by insisting on decidability. The Apple liveness work is one of the most impressive industrial-scale applications of formal methods to hardware design seen in recent years. But it remains a specialist tool at the frontier of what's practically achievable, not yet a broadly deployed engineering method.

# talks

ETAPS 2020 PhD talk ("How far can you EPR?")https://www.wisdom.weizmann.ac.il/~padon/how-far-can-you-epr-etaps2020.mp4 (direct mp4, not YouTube)

"Paxos Made EPR: Decidable Reasoning about Distributed Protocols"
https://www.youtube.com/watch?v=_suyrSMJeCo



# IVy Formal Verification — Collected Bibliography

**IVy** is a language and tool for specifying, modeling, implementing, and verifying protocols,
developed primarily by Kenneth L. McMillan (Microsoft Research / UT Austin) in collaboration with
researchers at Tel Aviv University and UC Berkeley. This bibliography covers the core papers,
related foundational work, and key industrial applications, in roughly chronological order.

---

## Primary IVy Papers

### 2016

**Padon, O., McMillan, K. L., Panda, A., Sagiv, M., & Shoham, S.**
*Ivy: Safety Verification by Interactive Generalization.*
PLDI 2016: Proceedings of the 37th ACM SIGPLAN Conference on Programming Language Design and Implementation, pp. 614–630. ACM.
[https://dl.acm.org/doi/10.1145/2908080.2908118](https://dl.acm.org/doi/10.1145/2908080.2908118)
> The founding paper. Introduces the interactive counterexample-guided approach to finding inductive invariants, the Relational Modeling Language (RML), and the constraint to decidable fragments of first-order logic.

---

### 2018

**McMillan, K. L., & Padon, O.**
*Deductive Verification in Decidable Fragments with Ivy.*
SAS 2018: Static Analysis Symposium, pp. 43–55. Springer.
[https://wcl.cs.rpi.edu/pilots/library/papers/fp/ivy_sas18.pdf](https://wcl.cs.rpi.edu/pilots/library/papers/fp/ivy_sas18.pdf)
> Surveys progress on Ivy through 2018, including deductive verification, compositional testing, and case studies on distributed consensus protocols and cache-coherent interfaces.

**Taube, M., Losa, G., McMillan, K. L., Padon, O., Sagiv, M., Shoham, S., Wilcox, J. R., & Woos, D.**
*Modularity for Decidability of Deductive Verification with Applications to Distributed Systems.*
PLDI 2018, ACM.
[http://jamesrwilcox.com/modularity-for-decidability.pdf](http://jamesrwilcox.com/modularity-for-decidability.pdf)
> Extends Ivy's modular proof methodology to support verified implementations of Paxos and Raft, demonstrating the approach's scalability to real consensus protocols.

**Padon, O., Hoenicke, J., McMillan, K. L., Podelski, A., Sagiv, M., & Shoham, S.**
*Temporal Prophecy for Proving Temporal Properties of Infinite-State Systems.*
FMCAD 2018, pp. 74–84.
[https://link.springer.com/article/10.1007/s10703-021-00377-1](https://link.springer.com/article/10.1007/s10703-021-00377-1)
> Introduces prophecy variables as a technique for reducing temporal (liveness) verification to safety verification in infinite-state systems, implemented in Ivy.

---
### 2019

**McMillan, K. L., & Zuck, L. D.**
*Formal Specification and Testing of QUIC.*
SIGCOMM 2019, pp. 227–240. ACM.
[https://dl.acm.org/doi/10.1145/3341302.3342087](https://dl.acm.org/doi/10.1145/3341302.3342087)
> Applies Ivy's compositional specification-based testing methodology to the QUIC Internet transport protocol (then under IETF standardization), generating randomized testers that found significant bugs in real implementations.

**McMillan, K. L., & Zuck, L. D.**
*Compositional Testing of Internet Protocols.*
IEEE Secure Development Conference (SecDev) 2019, pp. 161–174.
[https://ieeexplore.ieee.org/document/8900069](https://ieeexplore.ieee.org/document/8900069)
> Companion paper to the SIGCOMM work, focusing on the security and compliance aspects of Network Conformance Testing (NCT) methodology.

---

### 2020

**McMillan, K. L., & Padon, O.**
*Ivy: A Multi-Modal Verification Tool for Distributed Algorithms.*
CAV 2020 (Computer Aided Verification), Part II, LNCS vol. 12225, pp. 190–202. Springer.
[https://link.springer.com/chapter/10.1007/978-3-030-53291-8_12](https://link.springer.com/chapter/10.1007/978-3-030-53291-8_12)
[Open Access via PubMed Central](https://pmc.ncbi.nlm.nih.gov/articles/PMC7363183/)
> Definitive tool paper presenting Ivy's full multi-modal architecture: deductive verification via SMT, eager abstraction model checking, liveness-to-safety tactics, and C++ code extraction.

---

### 2021

**Padon, O., Hoenicke, J., McMillan, K. L., Podelski, A., Sagiv, M., & Shoham, S.**
*Temporal Prophecy for Proving Temporal Properties of Infinite-State Systems.* (Extended version)
Formal Methods in System Design, 57(2), pp. 246–269. Springer.
[https://link.springer.com/article/10.1007/s10703-021-00377-1](https://link.springer.com/article/10.1007/s10703-021-00377-1)
> Journal version of the FMCAD 2018 paper, with extended proofs and additional case studies.

---

### 2022

**Padon, O., Wilcox, J. R., Koenig, J. R., McMillan, K. L., & Aiken, A.**
*Induction Duality: Primal-Dual Search for Invariants.*
POPL 2022: Proceedings of the ACM on Programming Languages, 6(POPL), pp. 1–29.
[https://dl.acm.org/doi/10.1145/3498712](https://dl.acm.org/doi/10.1145/3498712)
> Introduces primal-dual invariant search, a new approach to automated invariant inference that complements Ivy's interactive generalization methodology.

---

### 2023

**Tamir, O., Taube, M., McMillan, K. L., Shoham, S., Howell, J., Gueta, G., & Sagiv, M.**
*Counterexample Driven Quantifier Instantiations with Applications to Distributed Protocols.*
OOPSLA 2023: Proceedings of the ACM on Programming Languages, 7(OOPSLA2), pp. 1878–1904.
[https://dl.acm.org/doi/10.1145/3622849](https://dl.acm.org/doi/10.1145/3622849)
> Advances Ivy's handling of quantified formulas, a core challenge in automated verification of distributed protocols.

**von Hippel, M., McMillan, K. L., Nita-Rotaru, C., & Zuck, L. D.**
*A Formal Analysis of Karn's Algorithm.*
NETYS 2023, pp. 43–61. Springer.
[https://link.springer.com/chapter/10.1007/978-3-031-37765-5_4](https://link.springer.com/chapter/10.1007/978-3-031-37765-5_4)
> Applies Ivy's specification and testing methodology to Karn's retransmission algorithm, a foundational component of TCP congestion control.

**Vick, C., & McMillan, K. L.**
*Synthesizing History and Prophecy Variables for Symbolic Model Checking.*
VMCAI 2023, pp. 320–340. Springer.
[https://link.springer.com/chapter/10.1007/978-3-031-24950-1_15](https://link.springer.com/chapter/10.1007/978-3-031-24950-1_15)

---

### 2024

**McMillan, K. L.**
*Toward Liveness Proofs at Scale.*
CAV 2024: Computer Aided Verification, LNCS. Springer.
[https://link.springer.com/chapter/10.1007/978-3-031-65627-9_13](https://link.springer.com/chapter/10.1007/978-3-031-65627-9_13)
[Semantic Scholar entry](https://www.semanticscholar.org/paper/Toward-Liveness-Proofs-at-Scale-McMillan/23b034f67b555d3415ac0982385f992e28836559)
> Introduces liveness proof by *relational rankings*, directly motivated by Apple hardware engineers' work proving liveness of a memory subsystem model in Ivy. Demonstrates the method on the Apple generic memory model (contributed to the Ivy open-source repository).

---

## Hardware Verification Applications

**McMillan, K. L.**
*Modular Specification and Verification of a Cache-Coherent Interface.*
FMCAD 2016, pp. 109–116. IEEE.
[https://ieeexplore.ieee.org/document/7886665](https://ieeexplore.ieee.org/document/7886665)
> Applies Ivy's compositional methodology to the coherent memory interface of the RISC-V processor architecture; found subtle timing bugs in RTL-level implementations.

---

## Foundational / Closely Related Work

**Padon, O., Hoenicke, J., Losa, G., Podelski, A., Sagiv, M., & Shoham, S.**
*Reducing Liveness to Safety in First-Order Logic.*
PACMPL 2(POPL), Article 26. ACM, 2018.
[https://dl.acm.org/doi/10.1145/3158114](https://dl.acm.org/doi/10.1145/3158114)
> Foundational paper for liveness-to-safety reduction in first-order logic, which underpins Ivy's liveness proof tactics.

**Padon, O.**
*Ivy: Safety Verification by Interactive Generalization.* (PhD Thesis, Tel Aviv University, 2018)
[https://www.wisdom.weizmann.ac.il/~padon/ivy.pdf](https://www.wisdom.weizmann.ac.il/~padon/ivy.pdf) [cached](doc/ivy_paper_diss.pdf)
> Comprehensive treatment of the theory and implementation behind Ivy's interactive generalization approach.

---

## Project Resources

| Resource | URL |
|---|---|
| Microsoft Research project page | [https://www.microsoft.com/en-us/research/project/ivy/](https://www.microsoft.com/en-us/research/project/ivy/) |
| IVy documentation & tutorials | [https://kenmcmil.github.io/ivy/](https://kenmcmil.github.io/ivy/) |
| GitHub repository (microsoft/ivy) | [https://github.com/kenmcmil/ivy](https://github.com/kenmcmil/ivy) |
| QUIC formal specification (in Ivy) | [https://github.com/kenmcmil/ivy/blob/master/doc/examples/quic/README.md](https://github.com/kenmcmil/ivy/blob/master/doc/examples/quic/README.md) |
| Apple generic memory model (contributed to Ivy repo) |  [https://github.com/kenmcmil/ivy/tree/master/doc/examples/apple](https://github.com/kenmcmil/ivy/tree/master/doc/examples/apple) |
| Ken McMillan's home page | [https://mcmil.net/wordpress/](https://mcmil.net/wordpress/) |
| McMillan publication list (dblp) | [https://dblp.org/pid/m/KennethLMcMillan.html](https://dblp.org/pid/m/KennethLMcMillan.html) |

---

## Notes

- The **Apple memory subsystem work** is described in the CAV 2024 paper "Toward Liveness Proofs at Scale." Apple engineers used Ivy to prove *safety* properties of a ~1,200-line memory subsystem model (78 invariants, ~500 lines of proof). McMillan extended Ivy's liveness support to handle their liveness conjecture (that every core operation is eventually retired). The generic model and proofs were contributed to the open-source Ivy repository.

- The **QUIC specification** work (SIGCOMM 2019) is among the highest-impact applications of Ivy's testing methodology outside hardware, and directly influenced the IETF standardization process for HTTP/3.

- The **RISC-V cache coherence** work (FMCAD 2016) predates the Apple engagement and established the pattern of using Ivy specifications to test hardware implementations compositionally.

- McMillan left Microsoft Research for **UT Austin** (Department of Computer Science), where Ivy research continues.

---

*Bibliography compiled March 2026. For the most current publication list, see McMillan's dblp page linked above.*

------------
# Questions

Q: What is first order logic and how does it relate to the logic fragment EPR?

A:

**First-Order Logic (FOL)**

First-order logic extends propositional logic by adding quantifiers and terms, allowing statements about objects in a domain rather than just atomic truth values. Its components are:

- **Terms** — variables, constants, and function symbols applied to terms (e.g., `f(x, c)`)
- **Atoms** — predicate symbols applied to terms (e.g., `P(x, f(y))`)
- **Connectives** — ∧, ∨, ¬, →, ↔
- **Quantifiers** — ∀ and ∃ ranging over domain elements

FOL is highly expressive — you can reason about relations, functions, and structured domains — but this comes at a cost. FOL is **semi-decidable**: validity is recursively enumerable, but unsatisfiability is not, so automated reasoning does not always terminate.

---

**The EPR Fragment**

EPR stands for **Effectively Propositional Reasoning** (also called the **Bernays-Schönfinkel class**). It is the fragment of FOL restricted to formulas of the form:

> ∃\* ∀\* φ

meaning: existential quantifiers followed by universal quantifiers, with **no function symbols** (constants are permitted). This quantifier prefix pattern is key — it is the opposite order from what Skolemization typically produces.

**Why It's Decidable**

Without function symbols, terms cannot nest. Every term is either a variable or a constant, so the **Herbrand universe** — the set of ground terms — is finite. This means the search space for a model is bounded, and satisfiability becomes decidable, essentially reducing to a propositional problem, hence the name "effectively propositional."

**Why the Quantifier Order Matters**

In ∃\*∀\* formulas, Skolemization only introduces constants (not functions), because the existential quantifiers precede all universals and thus depend on nothing. This keeps the Herbrand universe finite. If you allowed ∀∃ alternation, Skolem functions would appear, terms could nest arbitrarily deeply, and decidability would be lost.

**Why EPR Matters in Practice**

Despite its restrictions, EPR is surprisingly expressive for verification tasks:

- Many **software and hardware verification** problems can be encoded in EPR
- **Bounded model checking** naturally fits the fragment
- Systems like **Ivy** (developed at Microsoft Research and Carnegie Mellon) use EPR as their core logic, because its decidability gives strong guarantees for protocol verification
- It sits at a useful sweet spot — expressive enough to capture interesting properties, constrained enough to be mechanically decidable

It is one of the most practically important decidable fragments of FOL in formal methods today.
------

Q: Tell me about propositional logic

A:

**Propositional Logic**

Propositional logic is the simplest and most foundational formal logic. It reasons about **propositions** — statements that are either true or false — and how they combine via logical connectives. There are no quantifiers, no variables ranging over domains, no functions or predicates. Just atomic truth values and the connectives that operate on them.

---

**Syntax**

The basic ingredients are:

- **Atomic propositions** — primitive symbols, e.g., `P`, `Q`, `R`, each assigned a truth value
- **Connectives** — the standard ones being:
  - ¬ (negation)
  - ∧ (conjunction)
  - ∨ (disjunction)
  - → (implication)
  - ↔ (biconditional)
          
Formulas are built recursively from atoms and connectives in the usual way.
          
---
          
**Semantics**
          
The semantics is simple: a **valuation** assigns true or false to each atomic proposition, and the truth value of a compound formula is determined compositionally by the standard truth tables. There is no deeper model theory — everything reduces to truth table evaluation.
          
Key semantic notions:
          
 - **Tautology** — true under all valuations (e.g., `P ∨ ¬P`)
 - **Satisfiable** — true under at least one valuation
 - **Unsatisfiable (contradiction)** — true under no valuation (e.g., `P ∧ ¬P`)
          
---
          
**Decidability**
          
Propositional logic is fully **decidable** — truth table evaluation always terminates. However, the decision problem for satisfiability, known as **SAT**, is **NP-complete**. This means that while a solution can always be verified quickly, finding one may require exponential time in the worst case. SAT was in fact the first problem proven NP-complete (Cook, 1971).

In practice, modern **SAT solvers** handle formulas with millions of variables efficiently through heuristics like DPLL and CDCL, and are a cornerstone of industrial formal verification.
          
---

**Normal Forms**
          
Propositional logic has two important normal forms:
          
 - **CNF (Conjunctive Normal Form)** — a conjunction of clauses, where each clause is a disjunction of literals. This is the standard input format for SAT solvers.
 - **DNF (Disjunctive Normal Form)** — a disjunction of conjunctions of literals.
          
Any propositional formula can be converted to either form, though the conversion to CNF can introduce new variables (Tseitin transformation) to avoid an exponential blowup.
          
---
          
**Relationship to FOL**
          
Propositional logic can be seen as the quantifier-free, function-free, variable-free fragment of FOL. FOL strictly generalizes it — every propositional tautology lifts to a FOL tautology, but FOL can express vastly more. The jump from propositional to first-order logic is where undecidability enters, which is why the decidable fragments of FOL (like EPR) are valuable — they recover some of the tractability of the propositional setting while retaining more expressive power.

------

Q: Give an example please of how FOL can express vastly more, but at the same time be reduced to propositional logic.

Great question — it gets at something deep about the relationship between the two logics.

---

**What FOL Can Express That Propositional Logic Cannot**

Consider a simple FOL statement:

> ∀x (Person(x) → Mortal(x))

This says something general about *all* objects in a domain — an infinite or arbitrary collection. Propositional logic has no mechanism for this. You could write `P → Q` for specific named individuals, but you cannot quantify over an open-ended domain. FOL captures structure and relationships that propositional logic simply cannot express.

---

**How FOL Reduces to Propositional Logic**

The reduction works through **grounding** — instantiating the universal quantifier over a finite, concrete domain. Suppose your domain is just three individuals: `{alice, bob, carol}`. Then:

> ∀x (Person(x) → Mortal(x))

grounds out to the conjunction:

> (Person(alice) → Mortal(alice)) ∧ (Person(bob) → Mortal(bob)) ∧ (Person(carol) → Mortal(carol))

Now each atom like `Person(alice)` is just a proposition — it has no internal structure anymore, it is simply true or false. The entire formula is now a plain propositional formula, and a SAT solver can handle it directly.

This is essentially what happens in **EPR** — because the Herbrand universe is finite, every FOL formula in that fragment can be grounded into a propositional formula of finite size, which is why EPR is decidable.

---

**The Catch**

This only works when the domain is finite and known. In full FOL, the domain may be infinite — you cannot enumerate all instantiations. The grounding process never terminates, which is precisely where undecidability comes from. The power of FOL to reason over arbitrary domains is exactly what makes it escape propositional decidability.

So the tradeoff is sharp: **expressiveness buys you undecidability; bounding the domain buys you back decidability at the cost of generality.**

------

Q: Ivy uses EPR to prove safety of inifinite state systems; how is that possible since EPR needs finite domains?

A: Excellent question — this is one of the most elegant aspects of Ivy's design.

---

**The Apparent Contradiction**

You are right that EPR's decidability rests on a finite Herbrand universe. Yet distributed protocols — the primary target of Ivy — have unbounded numbers of nodes, messages, and quorums. How can a finite-domain logic reason about them?

---

**The Resolution: Reasoning About the Protocol, Not the State Space**

The key insight is that Ivy does not enumerate the state space. Instead it reasons about **universally quantified invariants** over the protocol's transition relation. The question is never "check all reachable states" but rather "does this invariant hold inductively" — meaning:

- It holds in the initial state
- If it holds before a transition, it holds after

This is a **verification condition** expressed as a FOL formula, and Ivy is designed so that these verification conditions fall within EPR. The domain of the verification condition is the set of symbolic actors — nodes, values, quorums — treated as uninterpreted sorts, not as enumerated objects.

---

**Uninterpreted Sorts Are the Bridge**

In Ivy, you declare sorts like `node` or `quorum` without giving them a size. The EPR formula says things like:

> ∀ n:node, q:quorum. member(n, q) → ...

The EPR decision procedure treats these sorts as **arbitrary but fixed** — it does not need to know how many nodes there are. It asks: is there *any* interpretation of the sorts and relations that falsifies the invariant? If not, the invariant holds for all domain sizes simultaneously.

This works because EPR's decidability does not actually require the domain to be small — it requires the **Herbrand universe to be finite given the formula**, which holds when there are no function symbols. The sorts can be arbitrarily large in any specific model; what matters is that the formula's structure prevents unbounded term nesting.

---

**So What Is Finite?**

What is finite is not the domain of the system being verified, but the **syntactic structure of the verification condition**. The absence of function symbols keeps the formula's Herbrand universe finite at the logical level, even though the semantic domain — the actual nodes and quorums in a real deployment — can be arbitrarily large.

---

**The Payoff**

This is what makes Ivy powerful. You get:

- A decidable verification procedure (no false timeouts or incompleteness)
- Correctness guarantees that hold for **all instance sizes** — 3 nodes, 300 nodes, or 3 million
- No need to fix a bound ahead of time, unlike bounded model checking

It is a genuinely clever engineering of the logic — staying within EPR is a discipline imposed on the user (Ivy will reject specifications that escape the fragment), but the reward is a verifier that is both sound and complete for the properties it can express.

------------

# notes on installing python original Ivy

Some of the commented out steps (#) may 
be needed for a completely fresh install
when z3 is not built yet. The z3 build
takes 10 minutes, so we comment out
to avoid that unnecessary delay.

~~~
sudo apt install python3-tk
cd goivy/pyivy
deactivate # exit any old venv first! (if any)
python3 -m venv goivy-venv
source goivy-venv/bin/activate
pip install -r requirements.txt

pip3 install 'setuptools<70.0.0' ## depends on older version
#pip3 install pyyaml # needed at one point.
#pip3 install blake3

cd ivy

# builds z3. very, very slow (10 minutes):
#git submodule update --init --recursive
#python3 build_submodules.py

# uninstall any prior version
pip3 uninstall ms_ivy 

# install as "editable" egg so that
# python source code changes take 
# effect immediately.
pip3 install -e .
~~~

------------

# Ivy Python -> goivy in Go

---

## KEY PYTHON SOURCE FILE PATHS

For reference during implementation:

```
~/ivy/pyivy/ivy/ivy/ivy_logic.py         — higher-level logic IR (1,774 lines)
~/ivy/pyivy/ivy/ivy/ivy_module.py         — module context (412 lines)
~/ivy/pyivy/ivy/ivy/ivy_actions.py        — action semantics (1,687 lines)
~/ivy/pyivy/ivy/ivy/ivy_transrel.py       — transition relations (669 lines)
~/ivy/pyivy/ivy/ivy/ivy_proof.py          — proof checking (1,653 lines)
~/ivy/pyivy/ivy/ivy/ivy_compiler.py       — AST → logic compilation (2,320 lines)
~/ivy/pyivy/ivy/ivy/ivy_isolate.py        — modular verification (2,022 lines)
~/ivy/pyivy/ivy/ivy/ivy_theory.py         — built-in theories (181 lines)
~/ivy/pyivy/ivy/ivy/ivy_interp.py         — symbolic interpreter (642 lines)
~/ivy/pyivy/ivy/ivy/ivy_art.py            — abstract reachability (523 lines)
~/ivy/pyivy/ivy/ivy/ivy_check.py          — verification driver (1,040 lines)
~/ivy/pyivy/ivy/ivy/ivy_mc.py             — model checking (1,772 lines)
~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py    — clause operations (1,635 lines)
~/ivy/pyivy/ivy/ivy/ivy_solver.py         — Z3 solver bridge (1,716 lines)
```

# amusing

~~~
~/goivy $ scc --exclude-dir z3vendor --exclude-dir z3bridge --exclude-dir z3ivy
───────────────────────────────────────────────────────────────────────────────
Language                 Files     Lines   Blanks  Comments     Code Complexity
───────────────────────────────────────────────────────────────────────────────
Go                         393    185933    18638     23666   143629      33607
Markdown                   174     49225    11031         0    38194          0
HTML                        33     25209    10898        22    14289          0
Happy                        6      7162      460         0     6702          0
JavaScript                   5      4541      395       954     3192        548
Python                       5      1463       56       107     1300         42
Plain Text                   3       521       15         0      506          0
CSS                          2       888      115        35      738          0
JSON                         1        10        0         0       10          0
Makefile                     1       136       34        13       89         18
SVG                          1        31        3         4       24          0
gitignore                    1         8        0         0        8          0
───────────────────────────────────────────────────────────────────────────────
Total                      625    275127    41645     24801   208681      34215
───────────────────────────────────────────────────────────────────────────────
Estimated Cost to Develop $7,362,958
Estimated Schedule Effort 32.762021 months
Estimated People Required 26.621735
───────────────────────────────────────────────────────────────────────────────

~~~

# the User Interface data model for Ivy's UI (now ported to Go and typescript in webui/)

See also ivy/already_applied_plans/python_data_model_claude.md

~~~
The Python GUI's Data Model

Yes — there is a clear, coherent three-layer data model. It is not tangled with the
view code. Here it is:

Layer 1: Semantic/Computation Model (ivy_art.py, ivy_interp.py)

ivy_art.AnalysisGraph (ivy_art.py:72) — the ARG itself:
states       list[State]           ivy_art.py:77   — the graph nodes
transitions  list[(State,op,label,State)]  :78   — directed edges
covering     list[(State,State)]   :79            — covering pairs
pvars        predicate variables   :80
state_graphs list                  :81

ivy_interp.State (ivy_interp.py:83) — one ARG node:
domain    module/interpreter       :90
moded, clauses, precond  (from value tuple)  :91–92
expr      action-application expr  :95   (encodes how state was reached)
label     display label            :96
cached_update  transition relation :97
cached_pred    predecessor state   :98
id        int (set when added)     ivy_art.py:254
safe      bool (safety flag)       set in check_bounded_safety
universe  model universe           set from BMC result

State.value (ivy_interp.py:101) is a (moded, clauses, precond) triple where clauses is
the abstract formula representing the state. State.conjs and State.unders are global on
the domain (not per-node) — a design wart explicitly commented at :107–114.

Layer 2: Concept/Proof Model (concept.py, concept_interactive_session.py)

ConceptInteractiveSession (concept_interactive_session.py:26) — proof state for one
concept graph panel:
domain              ConceptDomain   :41   — active predicates
state               formula         :42   — current abstract state formula
axioms              formula         :43
goal_constraints    list[formula]   :44
suppose_constraints list[formula]   :45   — user-added emptiness constraints
abstract_value      dict            :93   — cached result of alpha()
undo_stack          list[(domain,suppose_constraints)]  :47
cache               dict            :51   — formula cache (shallow-copied)

concept.ConceptDomain (concept.py) — the set of predicates the user has chosen to
track:
- concepts: dict of name → Concept (quantifier-free formula)
- combiners: second-order formulas (node/edge combiners like at_least_one, all_to_all,
etc.)
- combinations: the instantiations to evaluate

Layer 3: Render Model (cy_elements.py, cy_render.py, dot_layout.py)

CyElements (cy_elements.py:5) — the serializable intermediate representation, totally
independent of both computation and browser:
elements  list[dict]     :42   — JSON-serializable list
node_id   dict[obj→str]  :43   — object → 'n{i}'
edge_id   dict[(obj,src,tgt)→str]  :44
shape_id  dict[obj→str]  :45

Each element dict has group ('nodes'/'edges'), data (id, label, classes, short_info,
long_info, events, actions), and positional data added later by dot_layout.

The rendering pipeline is a pure function chain:
AnalysisGraph
  → cy_render.render_rg()     (ivy_art.py:as_cy_elements, cy_render.py:276)
  → dot_layout.dot_layout()   (dot_layout.py:136) — adds x,y,width,height,bspline
  → CyElements (now with positions)
  → JSON to JS/browser

dot_layout is a pure function: given CyElements with no positions, it shells out to
graphviz dot, parses the output, and returns CyElements with positions. Entirely
testable without a browser.

cy_render.render_rg() (cy_render.py:276) is also pure: 
given an AnalysisGraph and callback lists, it produces 
CyElements. Testable with just the graph data.

Summary: The Python Data Model in One Diagram

AnalysisGraph (ivy_art.py:72)
  └─ []State (ivy_interp.py:83)
       ├─ clauses   (abstract formula)
       ├─ expr      (how we got here)
       └─ id, safe, universe
  └─ []transition  (State, op, label, State)
  └─ []covering    (State, State)

ConceptInteractiveSession (concept_interactive_session.py:26)
  └─ ConceptDomain (concept.py)
       └─ concepts: dict[name → Concept]
  └─ abstract_value: dict  (output of alpha())
  └─ undo_stack

CyElements (cy_elements.py:5)         ← render-side, serializable
  └─ elements: []dict                 ← sent as JSON to browser

This is a solid data model. 

----

Layer 1 — Semantic model (art.go, goivy package):
- AnalysisGraph { States []*State; Transitions []Transition; Covering []CoveringPair;
PVars []Expr }
- State { ID; Clauses; IsBottom(); Label; ... }
- Transition { Pre *State; Op ActionsAction; Label string; Post *State }

Layer 1b — Typed render intermediates (art_cyrender.go):
- AnalysisGraphState { States []ARGNode; Transitions []ARGTransition; Covering
[]ARGCover }
- ARGNode { ID int; Label string; IsBottom bool; Info string }
- ARGTransition { SourceID; TargetID; Label; IsJoin }
- ARGCover { CoveredID; CoveringID }

Layer 2 — Concept/proof model (webui_graph_model.go, webui_concept_isession.go,
webui_concept_domain.go):
- Graph { Sorts; ConceptSess *ConceptSession; Checks *DisplayCheckboxes; GraphStack }
- DisplayCheckboxes { EdgeDisplayCheckboxes map[string]map[string]*Option;
NodeLabelDisplayCheckboxes ... }
- Full ConceptInteractiveSession with domain, undo/redo stack, suppose constraints

Layer 3 — Cytoscape render (art_cyrender.go):
- CyElements { Elements []CyElement; NodeID map[string]string; EdgeID map[string]string
 }
- CyElement { Group; Data map[string]interface{}; Classes; Position *CyPosition }
~~~

We use both graphviz and cytoscape and persist the graph layout as well.

Enhancement over python note

efaca53ae fix a webui bug, concept graph recompute failed:

Q: okay, does python Ivy do the same? Or maybe python 
Ivy does not expose the same functionality? I'm worried 
we might have diverged from what Python ivy 
does in similar circumstances.

A: (Codex comments:)

Short answer: Python Ivy is inconsistent here.

In `pyivy/ivy/ivy/concept.py`, `get_initial_concept_domain()` classifies symbols by arity only, not by Boolean range. So a binary non-Boolean function is made into `Concept(name, [X,Y], c(X,Y))`, which is not a Boolean formula. See [concept.py](~/ivy/go/src/github.com/glycerine/ivy/pyivy/ivy/ivy/concept.py:450) and [concept.py](~/ivy/go/src/github.com/glycerine/ivy/pyivy/ivy/ivy/concept.py:456).

But Python’s `get_structure_concept_domain()` does the correct thing: it splits relations from functions by checking `c.sort.range == Boolean`; non-Boolean arity-2 functions become ternary concepts `Eq(c(X,Y), Z)`. See [concept.py](~/ivy/go/src/github.com/glycerine/ivy/pyivy/ivy/ivy/concept.py:595) and [concept.py](~/ivy/go/src/github.com/glycerine/ivy/pyivy/ivy/ivy/concept.py:624).

One more wrinkle: Python’s initial domain does not immediately put those arity-2 concepts into `concepts['edges']`; the edge recompute path iterates `concepts['edges']`. The Go webui had eagerly put binary symbols in `edges`, so it hit the bug earlier and harder during Check.

So: yes, Python has the same underlying bad arity-only assumption in one path, but it often does not expose it immediately. The fix I made aligns Go with Python’s more correct `get_structure_concept_domain()` behavior, not with the buggy initial-domain behavior. That seems like the right direction, especially because it preserves actual verification while keeping the concept graph well-typed.

------

emacs-ish mode for the CodeMirror 5 editor is actually pretty similar behavior to emacs now
that we have added some custom key handling and a 2 second time window after Esc is pressed
in which we check for a special follow-on key press.

~~~
Esc >         -- jump to end of buffer
Esc <         -- jump to begining
Esc w         -- copy region from mark to point.
Ctrl-x Ctrl-s -- to save
Ctrl-s        -- to i-search
Ctrl-r        -- to i-search backwards
Esc %         -- for search-and-replace
  Ctrl-g          to cancel search
  Arrow keys      to cancel search
  '.', y, n, !, q to supervise search-and-replace.
~~~

# Propositions versus Predicates

Q: I am new to "logic" because in computer science undergrad 
they taught us truth tables and booleans and skipped alot of 
mathematical navel gazing. Still now that I am working with
Ivy I need to understand the difference between Propositional Logic and Predicate logic. 
Or are they synonyms? Does Ivy use one and not the other? 
Was Dyjkstra talking about one or the other when he invented "weakest pre-conditions"?

A (Claude):

Great set of questions — and the framing of "mathematical navel-gazing" actually misses how *practical* this stuff becomes once you're doing verification. Let me give you the real picture.

## Propositional vs. predicate logic — not synonyms

They're a hierarchy. Predicate logic strictly contains propositional logic.

**Propositional logic** is exactly what you learned: atomic propositions (P, Q, R) that are either true or false, glued together with ∧, ∨, ¬, →, ↔. Truth tables decide everything. The atoms have no internal structure — `P` is just a Boolean, period. SAT solvers like MiniSAT operate here.

**Predicate logic** (a.k.a. first-order logic, FOL) breaks open those atoms. Now your atoms are *predicates applied to terms*: `leader(x)`, `id(x) < id(y)`, `pending(m, n)`. You get:

- **Variables and quantifiers**: ∀x and ∃x ranging over some domain (nodes, messages, integers, …)
- **Function symbols**: `id(x)`, `succ(n)`, `f(g(x))`
- **Equality**: `x = y`
- **A domain of discourse**: the set the variables range over

So "all nodes have a unique id" is `∀x ∀y. id(x) = id(y) → x = y` — there is no way to write this in propositional logic without first fixing a specific finite number of nodes and then expanding to a giant conjunction.

The expressive gap is huge. Propositional logic is *decidable* (truth tables always work, even if slowly). Predicate logic is only *semi-decidable* in general — if a formula is valid, you can find a proof, but if it isn't valid, your prover might run forever.

## Where Ivy sits

Ivy is fundamentally a **first-order logic** tool — predicate logic all the way. The whole point of Ivy's design philosophy makes no sense at the propositional level. Recall from your own work that Ivy targets the **EPR fragment** (Effectively Propositional, also called Bernays-Schönfinkel-Ramsey) and **FAU** (Finite Almost Uninterpreted). These are *fragments of first-order logic* that Ivy carves out specifically because:

1. They're expressive enough to model real distributed protocols with unbounded numbers of nodes, messages, etc. — that's the FOL part you need.
2. They're *decidable* — Z3 always terminates on them with a definite answer. That's the special thing about these fragments versus general FOL.

The name "Effectively Propositional" is telling. EPR formulas have the form ∃ X . ∀ Y . φ where φ is quantifier-free and has no function symbols (only relations and constants). The Skolemization of such a formula produces a *finite* Herbrand universe, so deciding satisfiability reduces to a (potentially huge) propositional SAT problem. EPR is in NEXPTIME — propositional logic is in NP — but the reduction is the reason for the name. So Ivy is using predicate logic, but deliberately staying in a fragment that has a propositional-flavored decision procedure underneath.

When Ivy complains about a "function cycle" or "quantifier alternation" pushing your VC outside the decidable fragment, what's happening is: your formula is valid FOL, but it's drifted out of EPR/FAU into territory where Z3 might loop forever instantiating ground terms (the Herbrand expansion is now infinite). The fragment checker is enforcing the boundary.

## Dijkstra and weakest preconditions

Dijkstra's **wp** ("Predicate Transformer Semantics," 1975-76, and *A Discipline of Programming*, 1976) is squarely in **predicate logic**, and the name itself tells you so. The "P" in "wp" stands for *predicate*.

The setup: given a program statement S and a desired postcondition Q (a predicate on program states), `wp(S, Q)` is the weakest predicate P such that executing S in any state satisfying P guarantees termination in a state satisfying Q. Example:

```
wp(x := x + 1, x > 0)  =  x > -1
wp(x := x + 1, x = 5)  =  x = 4
```

These predicates are formulas over program variables — `x > 0`, `x = 5`, `∀i. 0 ≤ i < n → a[i] > 0` — and program variables range over infinite domains (integers, arrays, …). You cannot express any of this in propositional logic. The moment you write `x > 0` with x being any integer, you've left propositional logic; that's a unary predicate `>` applied to a term `x` and a constant `0`, ranging over an infinite domain.

Dijkstra was extremely deliberate about this. *A Discipline of Programming* leans heavily on quantifiers and predicate calculus precisely because reasoning about loops requires inductive invariants, and inductive invariants over unbounded state need quantifiers. The whole edifice of guarded commands, weakest preconditions, and the calculational style he and Gries and Hoare developed is built on predicate logic.

Ivy generates its VCs using exactly this wp calculus — that's not a coincidence. Read the [Ivy decidability docs page](https://kenmcmil.github.io/ivy/decidability.html) you found earlier with this lens: when they say verification conditions in Ivy are generated from the calculus of weakest liberal preconditions, that's Dijkstra's machinery directly. The "liberal" variant (wlp) just drops the termination requirement — wlp(S, Q) is the weakest predicate ensuring "if S terminates, then Q holds," which is the right notion when you're proving partial correctness via inductive invariants.

## Short mental model for your Ivy work

- **Propositional logic** = Booleans and connectives. Atoms have no structure. Decidable but only handles finite, fixed-size statements.
- **First-order / predicate logic** = quantifiers over a domain, predicates and functions on terms, equality. Expressive enough for real verification. Undecidable in general.
- **EPR, FAU, QFLIA, etc.** = carefully chosen *fragments* of FOL that are decidable. Ivy lives here.
- **Dijkstra's wp** = a function from (program, postcondition predicate) to precondition predicate. Pure predicate logic.

The "navel-gazing" turns out to be load-bearing: every constraint Ivy enforces about quantifier alternation, function cycles, and theory mixing is the tool trying to keep you inside a *decidable fragment of FOL*. Once you internalize that propositional logic was never going to be enough — you can't talk about "every node" without quantifiers — the whole design starts feeling forced by necessity rather than aesthetic preference.

If you want the canonical short read on this, Dijkstra's [EWD472 "Guarded commands, nondeterminacy and formal derivation of programs"](https://www.cs.utexas.edu/~EWD/transcriptions/EWD04xx/EWD472.html) (1975) is the original wp paper and is only a dozen pages. For the FOL-fragments angle, the Padon/McMillan paper [*Deductive Verification in Decidable Fragments with Ivy*](https://www.wisdom.weizmann.ac.il/~padon/ivy-sas18.pdf) is the right anchor — they spell out exactly why each fragment was chosen.

## follow-up: what is a "term" in Predicate logic (aka first-order logic, FOL)?

Q: You say, "Predicate logic (a.k.a. first-order logic, FOL) breaks open those atoms. Now your atoms are predicates applied to terms", but then do not define a "term". What is a "term" in the context of Predicate logic?

A (Claude):

Fair catch — I used "term" as if it were obvious, and it's actually a precise technical concept that matters.

## The definition

In predicate logic, a **term** is anything that *denotes an object* in your domain of discourse. Terms are the "noun phrases" of the logic — the things that refer to elements of the universe you're reasoning about. Predicates, by contrast, are the "verbs" or "adjectives" that say something *about* terms.

Terms are defined inductively (this is the standard textbook definition):

1. **Every variable is a term.** `x`, `y`, `n` — these denote some unspecified object in the domain.
2. **Every constant is a term.** `0`, `nil`, `alice` — these denote specific named objects in the domain.
3. **If `f` is an n-ary function symbol and `t₁, …, tₙ` are terms, then `f(t₁, …, tₙ)` is a term.** Functions take terms and produce new terms.
4. Nothing else is a term.

That's it. Terms are built from variables and constants by repeated function application. Crucially, **a term is not a formula**. A term has no truth value — you cannot ask whether `x + 3` is true or false; you can only ask what object it denotes.

## Examples to make the distinction sharp

In the language of arithmetic with `0, 1, +, ×, <, =`:

- `0` — term (a constant)
- `x` — term (a variable)
- `x + 1` — term (function `+` applied to terms `x` and `1`)
- `(x + 1) × y` — term (nested function application)
- `x + 1 < y` — **not a term**. This is an *atomic formula*: the predicate `<` applied to the terms `x + 1` and `y`. It has a truth value.
- `∀x. x + 1 < y` — a formula, not a term.

In Ivy's setting, with a node type and an `id` function:

- `x` — term (variable of type node)
- `id(x)` — term (function `id` applied to term `x`; denotes some id value)
- `id(x) < id(y)` — atomic formula (the `<` predicate applied to two terms)
- `leader(x)` — atomic formula (the `leader` predicate applied to a term)

Notice that `id(x)` and `leader(x)` look syntactically identical but live in different worlds. `id` is a **function symbol** — it takes a node and returns an id, so `id(x)` is a term denoting an object. `leader` is a **predicate symbol** — it takes a node and returns a truth value, so `leader(x)` is a formula. The distinction is in the symbol's declaration, not its surface syntax.

## Why this matters in Ivy

Ivy keeps function symbols and predicate symbols (relation symbols) syntactically distinct. When you declare `relation leader(N: node)` versus `function id(N: node): id_t`, you're telling Ivy which symbols produce terms and which produce formulas. This shows up in the decidable-fragment rules: the "function cycle" check Ivy enforces on EPR fragments is literally a check on the graph of function symbols (the things that build terms), because function symbols are what generate infinite Herbrand universes when quantifiers chase them.

The slogan: **terms denote things, formulas say things**. Predicate logic is the layer that lets you build complex things (nested terms) and then say complex stuff about them (with predicates and quantifiers). Propositional logic, by contrast, has no terms at all — its atoms are just opaque Booleans with no internal structure to refer to.

# follow-up: What are Sorts?

Excellent question — this gets at something that's often glossed over in introductory treatments.

## Two flavors of predicate logic

There's a fork in the road here:

**Unsorted (single-sorted) first-order logic** is the version in classic textbooks. There's one domain of discourse — call it the universe — and every variable, constant, and function output lives in that one universe. No sorts, no types. When you write `∀x. P(x)`, the `x` ranges over *everything*.

**Many-sorted first-order logic** partitions the universe into disjoint **sorts**. Every term has a sort. Function and predicate symbols have sort signatures specifying what sorts their arguments must have and (for functions) what sort their result is. Variables are declared with a sort, and quantifiers range only over the sort.

These are formally equivalent in expressive power — you can always encode sorts in unsorted FOL by introducing unary predicates `IsNode(x)`, `IsMessage(x)`, etc., and relativizing every quantifier. But this encoding is awful in practice. It bloats formulas, hides structure, and lets you write nonsense like `succ(alice) < bob.color` that should be a type error caught immediately. So real verification tools — Ivy, SMT-LIB, Z3, Dafny, F\*, Coq — all use many-sorted logic.

## What is a sort?

Q: Okay, so what is a "Sort"? In Predicate logic, do terms have Sorts (types)? Do formulas have Sorts?

A (Claude):

A sort is just a *name for a domain*. Syntactically, it's a symbol like `node`, `message`, `int`, `bool`. Semantically, when you give a model, each sort gets interpreted as a specific non-empty set, and these sets are disjoint.

In Ivy you write:

```
type node
type id_t
relation leader(N: node)
function id(N: node): id_t
```

This declares two sorts (`node` and `id_t`), a predicate `leader` taking a `node`, and a function `id` mapping `node` to `id_t`. The signature `node → id_t` is the sort signature of `id`.

## Do terms have sorts? Yes.

In many-sorted logic, every term has exactly one sort, determined inductively:

1. A variable's sort is fixed by its declaration.
2. A constant's sort is fixed by its declaration.
3. If `f` has signature `s₁ × … × sₙ → s` and each `tᵢ` has sort `sᵢ`, then `f(t₁, …, tₙ)` has sort `s`. If the argument sorts don't match, `f(t₁, …, tₙ)` is **not a well-formed term at all** — it's a syntax error, just like a type error in a programming language.

So in the Ivy example:
- `x` (declared as `node`) has sort `node`.
- `id(x)` has sort `id_t` (because `id : node → id_t`).
- `id(id(x))` is **not well-formed** — `id` expects a `node`, but `id(x)` has sort `id_t`. You'd catch this at parse time.

This is exactly analogous to type-checking in a programming language. Sorts in many-sorted FOL are types, and the well-formedness rule for terms is type-checking. Your background with Go's type system maps directly: function signatures, no implicit conversions, well-typedness is a syntactic check.

## Do formulas have sorts?

Here the answer splits depending on how puristic you want to be.

**Strict traditional answer: no.** Formulas don't have sorts — they have a *truth value*. There's a categorical distinction between the term level (sorted objects in the domain) and the formula level (truth values, which live in a separate meta-level Boolean space, not in any sort of the logic). Atomic formulas are built by applying a predicate symbol to terms of matching sorts, but the resulting formula is not itself a term and doesn't have a sort. This is the standard view in mathematical logic textbooks.

**Pragmatic SMT/programming-language answer: yes, formulas have sort `Bool`.** SMT-LIB, Z3, and most modern verification tools treat `Bool` as just another sort. Formulas are terms of sort `Bool`. This collapses the term/formula distinction and is enormously convenient: you can have functions that return `Bool`, you can substitute Boolean expressions like ordinary terms, you can write `ite(P, x, y)` (if-then-else) where `P` is a `Bool` term and `x, y` are terms of some other sort. The whole logic becomes uniformly term-based.

Ivy, sitting on top of Z3, effectively uses this second view internally — relations are functions returning `Bool` under the hood. But Ivy's surface language preserves the predicate/function distinction syntactically (you write `relation` vs. `function`) because it matters for the decidable-fragment analysis. When Ivy's fragment checker hunts for function cycles, it's looking at non-`Bool` function symbols specifically, because those are the ones whose Herbrand expansion can blow up.

## Why the sort discipline pays off in Ivy

Two concrete reasons sorts matter for your verification work:

**Sort-stratified Herbrand universes.** When Z3 instantiates quantifiers, it generates ground terms. In unsorted logic, every term is a candidate substitution for every variable, so the search space explodes. With sorts, only terms of the matching sort are candidates. This is one of the things that makes EPR practically tractable: in EPR with sorts, each sort's Herbrand universe is finite and small.

**Sort-based decidability conditions.** Ivy's "function cycle" check operates on the graph whose nodes are sorts and whose edges are function symbols. A cycle in this graph (e.g., a function from `node` to `node`, or a chain `node → id_t → node`) means quantifier instantiation can generate arbitrarily deep nested terms, breaking decidability. Without sorts, you couldn't even formulate this check — every function would trivially "cycle" through the one universal domain.

So the slogan from the previous message gets refined: **terms denote things of a specific sort, and formulas say things relating terms of matching sorts**. The sort discipline is exactly the type discipline you'd expect from a programming-language background, applied to logic.

# follow-up: Nested terms

Q: I am still not clear on the definition of a "nested term". What makes a term "nested"? You say that this is a danger zone, why?

A (Claude):

Good — let's pin this down precisely, because the danger is the whole reason Ivy's fragment checker exists.

## What "nested" means

A term is **nested** when a function application appears *inside* another function application. The depth of nesting is just how many layers of function symbols you have to peel off before reaching variables and constants.

Take a function `f : node → node`. Starting from a variable `x : node`:

- `x` — depth 0, not nested. Just a variable.
- `f(x)` — depth 1. A function applied to a variable. Some people call this "nested," some don't; the standard usage is that nesting begins at depth 2.
- `f(f(x))` — depth 2, **nested**. `f` applied to a term that is itself a function application.
- `f(f(f(x)))` — depth 3, nested deeper.
- `f(f(f(f(x))))` — depth 4. And so on.

With multiple function symbols, nesting can mix them. Given `g : node → id_t` and `h : id_t → node`:

- `g(x)` — depth 1.
- `h(g(x))` — depth 2, nested. `h` applied to the result of `g`.
- `g(h(g(x)))` — depth 3, nested.

The structure is a tree. The depth of nesting is the height of that tree.

A non-nested term, by contrast, has function symbols applied only to variables or constants — never to other function applications. So `f(x)`, `g(x, y)`, `h(c, x)` are all unnested (or "flat"). The moment a function's argument is itself a function call, you have nesting.

## Why this is the danger zone

The danger is **unbounded Herbrand expansion**, which destroys decidability. Here's the mechanism, step by step.

When Z3 (or any SMT solver) tries to decide a formula with universal quantifiers, the standard technique is *quantifier instantiation*: replace `∀x. φ(x)` with a conjunction of instances `φ(t₁) ∧ φ(t₂) ∧ …` where the `tᵢ` are ground terms (terms with no variables) drawn from the **Herbrand universe** — the set of all ground terms you can build from the constants and function symbols in the formula.

Now ask: how big is this Herbrand universe?

**Case 1: no function symbols of the relevant sort** (this is EPR). Suppose your only `node`-producing symbols are the constants `alice` and `bob`. The Herbrand universe for sort `node` is just `{alice, bob}` — *finite*. Quantifier instantiation terminates after considering both. Decidable.

**Case 2: function symbols whose result sort is not their input sort**. Suppose `g : node → id_t` and you have constants `alice, bob : node` and no functions returning `node`. Then `node`-typed ground terms are still just `{alice, bob}`, and `id_t`-typed ground terms are `{g(alice), g(bob)}`. Still finite. Still decidable.

**Case 3: a function from a sort back to itself, e.g. `f : node → node`**. Now the `node`-typed Herbrand universe is:

```
alice, bob,
f(alice), f(bob),
f(f(alice)), f(f(bob)),
f(f(f(alice))), f(f(f(bob))),
…
```

**Infinite.** Every level of nesting produces fresh ground terms. Quantifier instantiation never runs out of candidates. The solver can loop forever generating ever-deeper nested instances without ever finding a model or a contradiction.

**Case 4: a cycle through multiple sorts**, e.g. `g : node → id_t` and `h : id_t → node`. Then:

```
alice : node
g(alice) : id_t
h(g(alice)) : node
g(h(g(alice))) : id_t
h(g(h(g(alice)))) : node
…
```

Also infinite. The cycle in the function-symbol-graph (`node → id_t → node → …`) means you can compose `h ∘ g` arbitrarily many times to keep producing fresh `node` terms.

## The connection to Ivy's "function cycle" rule

Now the fragment checker's behavior should look transparent rather than mysterious. Ivy builds a directed graph:

- **Nodes** of the graph: sorts (`node`, `id_t`, `message`, …).
- **Edges** of the graph: for each function symbol `f : s₁ × … × sₙ → s`, draw an edge from each `sᵢ` to `s`.

If this graph is **acyclic**, the Herbrand universe of every sort is finite (you can only build terms of bounded depth before running out of edges to follow). Quantifier instantiation terminates. The formula is in a decidable fragment.

If the graph has a **cycle**, you can build nested terms of unbounded depth around the cycle. Herbrand universe infinite. Decidability gone.

This is exactly why Ivy complains about function cycles and not, say, predicate cycles. Predicates produce `Bool`, not new terms of the sorts being quantified over — they don't grow the Herbrand universe of `node` or `id_t`. Only functions do. Skolemization counts here too: an `∃y` inside a `∀x` becomes a Skolem function `sk(x)`, and if that function's signature creates a sort cycle, you've quietly entered the danger zone even though you didn't write any function symbols explicitly. That's why Ivy also yells at you about quantifier alternation — `∀∃` patterns introduce Skolem functions during normalization, and those Skolem functions can close cycles.

## The intuition in one line

**Nested terms are dangerous when their sorts form a cycle, because then nesting has no ceiling — and an unbounded ceiling on term depth is an unbounded ceiling on instantiation work, which is decidability's grave.**

EPR forbids function symbols entirely (only constants and predicates), which trivially makes the function-sort graph have no edges, hence no cycles, hence bounded Herbrand universes. FAU is more permissive: it allows functions, but enforces the acyclicity condition on the sort graph. Both fragments are decidable for exactly this reason.

This is also why your earlier instinct about EPR being "effectively propositional" lands: with no nesting possible, the Herbrand universe is finite, and the whole problem reduces to a finite (large but finite) propositional SAT problem on the ground instances. The decidability of EPR is, in a real sense, the decidability of SAT scaled up to a known-finite term space.
