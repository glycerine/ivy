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

~~~
Package Dependency Layers (reference)

Layer 0 (foundation):  ast, lexer, ivyutils
Layer 1:               logic, parser, ivylogic
Layer 2:               module (+clauseops)
Layer 3:               z3bridge (+solver), logicutil, logicparser
Layer 4:               actions (+transrel), isolate
Layer 5:               interp, compiler, art
Layer 6:               proof(imports module, compiler), trace, bmc, temporal
Layer 7:               tactics, webui
Layer 8:               iupdr, check (top-level consumer)

Key constraint: imports only pull from lower layer number.
A package can import anything in a lower layer, 
but not vice-versa. e.g.

* `module` (L3) cannot import `actions` (L4). 

* `compiler` (L5) cannot import `proof` (L6).
~~~

## Context

Ivy is a formal analysis language (~45K lines of Python) using Z3 for model checking. We're porting it to Go incrementally, bottom-up, with comprehensive unit tests and fuzz tests per chunk. The Go code will live at `goivy/` in the repo root. Parser is hand-written recursive descent, and also a goyacc version to cross-check the precedence parsing because the original uses PLY (Python Lex Yacc).

## Go Package Structure

```
goivy/
  go.mod                    # module github.com/kenmcmil/ivy/goivy
  logic/                    # Chunks 1-3: sorts, terms, formulas, errors
    sort.go                 # Sort interface + concrete sort types
    sort_test.go
    term.go                 # Var, Const, Apply
    term_test.go
    formula.go              # Eq, Not, And, Or, Implies, Iff, ForAll, Exists, etc.
    formula_test.go
    node.go                 # Node interface
    error.go                # IvyError, SortError
  typeinfer/                # Chunk 4: union-find type inference
    sortvar.go
    unify.go
    infer.go
    typeinfer_test.go
  logicutil/                # Chunk 5: free_variables, substitution, etc.
  ivyutils/                 # Chunk 6: parameters, naming, utilities
  parser/                   # Chunk 7+: hand-written recursive descent
    lexer.go
    parser.go
```

## Chunk 1: Sort Types + Error Types

**Source**: `logic.py` lines 14-101, `general.py`

**Files to create**: `goivy/go.mod`, `goivy/logic/error.go`, `goivy/logic/sort.go`, `goivy/logic/sort_test.go`

### error.go
- `IvyError` struct implementing `error` — wraps a message
- `SortError` struct implementing `error` — distinct type for `errors.As`

### sort.go
- `Sort` interface: `String() string`, `Equal(Sort) bool`, unexported `sortSeal()` method
- Concrete types:
  - `UninterpretedSort{Name string}`
  - `BooleanSort struct{}` — singleton `var Boolean = BooleanSort{}`
  - `FunctionSort{Sorts []Sort}` — last element is range, rest is domain
    - Methods: `Domain() []Sort`, `Range() Sort`, `Arity() int`
    - Constructor `NewFunctionSort(sorts ...Sort) (*FunctionSort, error)` — validates ≥1 sort, all first-order
  - `EnumeratedSort{Name string, Extension []string}`
    - Methods: `Card() int`
  - `RangeSort{Name string, Lb, Ub string}`
  - `TopSort{Name string}` — default `var TopS = TopSort{Name: "TopSort"}`
    - Method: `IsSortVariable() bool` (true when Name != "TopSort")
- Functions: `FirstOrderSort(s Sort) bool`

### Tests & Fuzz
- Table-driven tests for each sort constructor, `String()`, `Equal()`
- `FunctionSort` validation: zero args → error, higher-order args → error
- `FirstOrderSort` returns false only for `FunctionSort`
- Fuzz: `FuzzNewFunctionSort` with random sort slices — no panics, valid error or valid object
- Fuzz: `FuzzUninterpretedSort` with random name strings — `Equal` reflexivity

---

## Chunk 2: Terms (Var, Const, Apply)

**Source**: `logic.py` lines 115-184

**Files to create**: `goivy/logic/node.go`, `goivy/logic/term.go`, `goivy/logic/term_test.go`

### node.go
- `Node` interface: `NodeSort() Sort`, `Children() []Node`, `String() string`, `Equal(Node) bool`
- All Sort types get `NodeSort()` returning themselves, `Children()` returning nil

### term.go
- `Var{Name string, VSort Sort}` — constructor validates name starts with uppercase
- `Const{Name string, CSort Sort}` — no name restriction (commented out in Python)
- `Apply{Func Node, Terms []Node}` — constructor validates:
  - Func sort is FunctionSort or TopSort
  - Arity matches
  - Argument sorts match domain sorts (unless TopSort involved)
  - `NodeSort()` returns range of func sort (or TopS if func is TopSort)
- `Var` and `Const` implement `func (v *Var) Call(terms ...Node) (Node, error)` returning Apply or self

### Tests & Fuzz
- Var: valid names, invalid names (lowercase → error), `String()`, `Equal()`
- Const: construction, `String()`, `Equal()`
- Apply: valid application, arity mismatch → error, sort mismatch → error, TopSort passthrough
- Reproduce Python's `__main__` examples: `leq(X,Y)` etc.
- Fuzz: random Var names, random Apply constructions

---

## Chunk 3: Formulas (Eq through NamedBinder)

**Source**: `logic.py` lines 186-448

**Files to create**: `goivy/logic/formula.go`, `goivy/logic/formula_test.go`

### formula.go

All formula types — constructors validate sort constraints:

| Type | Fields | Validation |
|------|--------|------------|
| `Eq{T1, T2 Node}` | two terms | same sort, first-order |
| `Ite{ISort Sort, Cond, Then, Else Node}` | sort from Then | cond is Boolean |
| `Not{Body Node}` | one sub | body is Boolean |
| `Globally{Environ *string, Body Node}` | | body is Boolean |
| `Eventually{Environ *string, Body Node}` | | body is Boolean |
| `WhenOperator{WSort Sort, Name string, T1, T2 Node}` | sort from T1 | T2 is Boolean |
| `Cond{CSort Sort, T1, T2 Node}` | sort from T2 | |
| `And{Terms []Node}` | variadic | all Boolean |
| `Or{Terms []Node}` | variadic | all Boolean |
| `Implies{T1, T2 Node}` | | both Boolean |
| `Iff{T1, T2 Node}` | | both Boolean |
| `ForAll{Variables []*Var, Body Node}` | | ≥1 var, body Boolean |
| `Exists{Variables []*Var, Body Node}` | | ≥1 var, body Boolean |
| `Lambda{Variables []*Var, Body Node}` | | all must be Var |
| `NamedBinder{Name string, Variables []*Var, Environ *string, Body Node}` | | sort is FunctionSort of var sorts → body sort |

Package-level:
```go
var True Node = &And{}    // And with no terms
var False Node = &Or{}    // Or with no terms
```

### Tests & Fuzz
- Each formula type: valid construction, sort validation errors
- `Children()` returns only sub-structures (not meta fields like Variables, Environ)
- `String()` matches Python output format
- Reproduce `__main__` examples: transitive, antisymmetric formulas, NamedBinder
- `ContainsTopSort(n Node) bool` and `IsPolymorphic(n Node) bool` — test with Python's assertions
- Fuzz: build random well-typed formula trees, verify `NodeSort()` invariants

---

## Chunk 4: Type Inference (Union-Find + Sort Unification)

**Source**: `type_inference.py` (~305 lines)

**Files to create**: `goivy/typeinfer/sortvar.go`, `goivy/typeinfer/unify.go`, `goivy/typeinfer/infer.go`, `goivy/typeinfer/typeinfer_test.go`

### sortvar.go
- `SortVar struct{ Instance SortOrVar }` — mutable, used in union-find
- `SortOrVar` interface implemented by both `logic.Sort` and `*SortVar`

### unify.go
- `Find(x SortOrVar) SortOrVar` — path-compressing find
- `Unify(s1, s2 SortOrVar) error` — unification with occurs check
- `OccursIn(s1, s2 SortOrVar) bool`

### infer.go
- `InferSorts(node logic.Node, env map[string]SortOrVar) error`
- `ConcretizeSorts(node logic.Node, sort logic.Sort) (logic.Node, error)`

### Tests & Fuzz
- Union-find: chains, path compression, cycle detection
- Unify: same sorts, TopSort wildcards, SortVar assignment, FunctionSort element-wise, incompatible → error
- Reproduce `type_inference.py` `__main__` block examples
- Fuzz: random sort pairs for `Unify` — no panics

---

## Verification Plan

After each chunk:
1. `go build ./...` — compiles
2. `go test ./... -v` — all unit tests pass
3. `go test ./... -fuzz=. -fuzztime=30s` — fuzz tests find no crashes
4. `go vet ./...` — no warnings
5. Cross-reference with Python: run Python `logic.py __main__` block, verify Go produces identical `String()` output

## Future Chunks (brief)

- **Chunk 5**: `logicutil` — `FreeVariables`, `UsedVariables`, `Substitute`, etc. from `logic_util.py`
- **Chunk 6**: `ivyutils` — `Parameter`, `UniqueRenamer`, `TopologicalSort`, name utilities from `ivy_utils.py`
- **Chunk 7+**: `parser` — hand-written recursive descent lexer+parser (replaces PLY lex/yacc)

# PROGRESS UPDATE 1

# Ivy Python → Go Port: Progress & Next Steps

## Context

Ivy is a formal analysis language (~45K lines of Python) using Z3 for model checking. We're porting it to Go incrementally, bottom-up, with comprehensive unit tests and fuzz tests per chunk. The Go code lives at `goivy/` in the repo root. Module path: `github.com/glycerine/goivy`. Branch: `goport`.

---

## PART 1: COMPLETED WORK

### Package Structure (current)

```
goivy/                          # 6,495 lines across 24 files
├── go.mod                      # module github.com/glycerine/goivy, go 1.21
├── go.sum
├── logic/                      # Chunks 1-3: sorts, terms, formulas (1,361 lines)
│   ├── error.go                # IvyError, SortError
│   ├── node.go                 # Node interface + Sort→Node implementations
│   ├── sort.go                 # Sort interface + 6 concrete types
│   ├── term.go                 # Var, Const, Apply
│   ├── formula.go              # 15 formula types + True/False
│   ├── sort_test.go            # 12 tests + 2 fuzz
│   ├── term_test.go            # 15 tests + 1 fuzz
│   └── formula_test.go         # 25 tests + 1 fuzz
├── typeinfer/                  # Chunk 4: union-find type inference (912 lines)
│   ├── sortvar.go              # SortOrVar interface, SortWrapper, SortVar
│   ├── unify.go                # Find, Unify, OccursIn, ConvertFrom/ToSortVars
│   ├── infer.go                # InferSorts, ConcretizeSorts (all 15+ node types)
│   └── typeinfer_test.go       # 14 tests + 1 fuzz
├── logicutil/                  # Chunk 5: logic utilities (1,124 lines)
│   ├── logicutil.go            # FreeVariables, Substitute, EqualModAlpha, etc.
│   └── logicutil_test.go       # 22 tests + 1 fuzz
├── ivyutils/                   # Chunk 6: general utilities (647 lines)
│   ├── renamer.go              # UniqueRenamer, VariableGenerator, ConstantNameGenerator
│   ├── parameter.go            # Parameter, BooleanParameter, EnumeratedParameter, Registry
│   ├── graph.go                # TopologicalSort[T,K], Reachable[T,K], FindCycle[K]
│   ├── names.go                # ComposeNames, SplitName, BaseName, PolymorphicSymbols
│   └── ivyutils_test.go        # 27 tests + 2 fuzz
└── z3bridge/                   # Chunk 7: Z3 integration (1,233 lines)
    ├── quantifier.go           # Self-contained Z3 CGo wrapper (Context, Sort, Expr, FuncDecl, Solver, Model, ForAll, Exists)
    ├── translate.go            # Translator: Ivy logic.Node → Z3 Expr, Implies(), IsSat()
    └── z3bridge_test.go        # 22 tests (incl. Python z3_utils.py __main__ reproduction)
```

### Chunk Details

**Chunk 1-3: `logic/`** — Ported from `logic.py` (488 lines)
- `Sort` interface (embeds `Node`) with seal method; 6 concrete types: `UninterpretedSort`, `BooleanSort`, `FunctionSort`, `EnumeratedSort`, `RangeSort`, `TopSort`
- `Node` interface: `NodeSort()`, `Children()`, `String()`, `Equal(Node)`
- Terms: `Var` (uppercase name validation), `Const`, `Apply` (arity/sort validation)
- 15 formula types: `Eq`, `Ite`, `Not`, `Globally`, `Eventually`, `WhenOperator`, `Cond`, `And`, `Or`, `Implies`, `Iff`, `ForAll`, `Exists`, `Lambda`, `NamedBinder`
- `True = &And{}`, `False = &Or{}`
- Helpers: `SortEqual`, `FirstOrderSort`, `ContainsTopSort`, `IsPolymorphic`, `IsBooleanOrTop`

**Chunk 4: `typeinfer/`** — Ported from `type_inference.py` (419 lines)
- `SortOrVar` interface with `SortWrapper` (wraps `logic.Sort`) and `SortVar` (mutable union-find node)
- `Find()` with path compression, `Unify()` with occurs check, `OccursIn()`
- `ConvertFromSortVars()`, `ConvertToSortVars()`, `InsertSortVars()`
- `InferSorts()` handles all 15+ node types via type switch, returns `InferResult{Sort, Concretize}`
- `ConcretizeSorts()` — replaces TopSorts with concrete sorts
- Note: `collectNames()` is a simplified stand-in; should eventually use `logicutil.FreeVariables`

**Chunk 5: `logicutil/`** — Ported from `logic_util.py` (335 lines)
- `FreeVariables(t)` — by pointer identity; `FreeVariablesByName(t)` — by name
- `UsedVariables(t)`, `BoundVariables(t)`, `UsedConstants(t)`
- `Substitute(t, subs)` — simultaneous substitution with `CaptureError` detection
- `IsTautologyEquality(t)` — detects `Eq(x,x)`
- `EqualModAlpha(t, u)` — alpha-equivalence using `pushableMap`
- Convenience: `FreeVariablesList()`, `UsedConstantsList()`

**Chunk 6: `ivyutils/`** — Ported from `ivy_utils.py` (773 lines)
- `UniqueRenamer` — generates unique names with suffix collision avoidance
- `VariableGenerator` — generates A,B,...,Z,AA,BB,... (skipping O)
- `ConstantNameGenerator()` — yields a-z, a0-z0, a1-z1,...
- `Parameter`, `BooleanParameter`, `EnumeratedParameter` with `Registry` and `Parameterize`
- `TopologicalSort[T,K]`, `Reachable[T,K]`, `FindCycle[K]` (generics)
- `ComposeNames`, `SplitName`, `BaseName`, `ParentChildName`, `ExtractParametersName`
- `Distinct[T]`, `PolymorphicSymbols`

**Chunk 7: `z3bridge/`** — Ported from `z3_utils.py` (197 lines) + `ivy_solver.py` (1,716 lines)
- Self-contained Z3 CGo wrapper (NOT using go-z3, which lacks quantifier support)
- `Context` with thread-safe mutex, `Sort`, `Expr`, `FuncDecl` with proper ref counting via `runtime.SetFinalizer`
- Boolean ops: `Not`, `And`, `Or`, `Implies`, `Iff`, `Eq`, `Ite`
- **Quantifiers**: `ForAll`, `Exists` via `Z3_mk_forall_const`/`Z3_mk_exists_const`
- `Solver` with `Assert`, `Check`, `Push`, `Pop`, `Model`
- `Translator` — converts all Ivy `logic.Node` types to Z3 `Expr`
- `Implies(f1, f2)`, `IsSat(f)` convenience functions
- CGo flags: `-I/usr/local/opt/z3/include -L/usr/local/opt/z3/lib -lz3`

### Test Summary

**138 tests total, all passing, `go vet` clean:**
| Package | Tests | Fuzz | Lines |
|---------|-------|------|-------|
| logic | 52 | 4 | 934 |
| typeinfer | 14 | 1 | 250 |
| logicutil | 22 | 1 | 453 |
| ivyutils | 27 | 2 | 368 |
| z3bridge | 22 | 0 | 419 |
| **Total** | **137** | **8** | **2,424** |

### Key Design Decisions Made
1. `Sort` embeds `Node` (solves `Equal(Sort)` vs `Equal(Node)` conflict); `SortEqual()` helper
2. Python `recstruct.__iter__` → Go `Children()` returns only sub-fields, not meta fields
3. Python `frozenset(variables)` → Go sorted display in `varSortList()` for deterministic `String()`
4. Python `__call__` on Var/Const/NamedBinder → Go `Call()` methods
5. Z3 integration uses direct CGo (not go-z3) because go-z3 lacks ForAll/Exists
6. `newExpr`/`newSort`/`newFuncDecl` must be called with lock held (avoids deadlock from nested `ctx.do()`)

---

## PART 2: REMAINING PYTHON CODEBASE

### Dependency Graph (what depends on what)

```
ALREADY PORTED:
  logic.py → logic/
  type_inference.py → typeinfer/
  logic_util.py → logicutil/
  ivy_utils.py → ivyutils/
  z3_utils.py + ivy_solver.py → z3bridge/

NEXT TO PORT:
  ivy_lexer.py (304 lines) — PLY lexer, 47 tokens + 173 reserved keywords
       ↓
  ivy_ast.py (1,965 lines) — 157 AST node classes
       ↓
  ivy_logic_parser.py (655 lines) — formula sub-parser
       ↓
  ivy_parser.py (3,161 lines) — 279 grammar rules (PLY yacc)
       ↓
  ivy_logic.py (1,774 lines) — higher-level IR, Symbol/Sig management
       ↓
  ivy_actions.py (1,687 lines) — 49 action classes (imperative semantics)
       ↓
  ivy_transrel.py (669 lines) — transition relation composition
       ↓
  ivy_module.py (412 lines) — module context/definitions
       ↓
  ivy_compiler.py (2,320 lines) — AST → logic IR compilation
       ↓
  ivy_isolate.py (2,022 lines) — modular verification
       ↓
  ivy_check.py (1,041 lines) — top-level verification
       ↓
  ivy_proof.py (1,653 lines) — proof state/tactics
       ↓
  ivy_mc.py (1,772 lines) — model checking
       ↓
  Code generators (ivy_to_cpp.py 6,715 lines, etc.)
       ↓
  UI/Visualization (ivy_ui*.py, tk_*.py, cy_*.py — ~7,000 lines)
```

### File Inventory (unported, by priority)

**Tier 1 — Language Frontend (~6,085 lines):**
| File | Lines | Role |
|------|-------|------|
| `ivy_lexer.py` | 304 | Tokenizer (47 tokens, 173 keywords, version-gated) |
| `ivy_ast.py` | 1,965 | 157 AST node classes (Formula, Decl, Action, Sort hierarchies) |
| `ivy_logic_parser.py` | 655 | Formula sub-parser (terms, atoms, quantifiers) |
| `ivy_parser.py` | 3,161 | Main parser (279 grammar rules, modules/actions/declarations) |

**Tier 2 — Core Semantics (~4,542 lines):**
| File | Lines | Role |
|------|-------|------|
| `ivy_logic.py` | 1,774 | Higher-level logic IR, Symbol/Sig, sort inference |
| `ivy_actions.py` | 1,687 | 49 action classes (Assume, Assert, Assign, If, While, Call, etc.) |
| `ivy_transrel.py` | 669 | Transition relations, state versioning, interpolation |
| `ivy_module.py` | 412 | Module context storage |

**Tier 3 — Compiler & Verification (~7,036 lines):**
| File | Lines | Role |
|------|-------|------|
| `ivy_compiler.py` | 2,320 | AST → logic IR transformation |
| `ivy_isolate.py` | 2,022 | Modular verification strategies |
| `ivy_check.py` | 1,041 | Top-level verification checker |
| `ivy_proof.py` | 1,653 | Proof state management and tactics |

**Tier 4 — Advanced Analysis (~3,219 lines):**
| File | Lines | Role |
|------|-------|------|
| `ivy_mc.py` | 1,772 | Model checking engine |
| `ivy_ranking.py` | 1,195 | Termination/ranking analysis |
| `ivy_theory.py` | 181 | Built-in theories (int arithmetic) |
| `ivy_interp.py` | 642 | Symbolic interpreter |
| `ivy_art.py` | 523 | Abstraction Refinement Tree (CEGAR) |

**Tier 5 — Code Generation (~8,375 lines):**
| File | Lines | Role |
|------|-------|------|
| `ivy_to_cpp.py` | 6,715 | C++ code generation |
| `ivy_cpp_types.py` | 525 | C++ type system |
| `ivy_cpp.py` | 414 | C++ compilation utilities |
| `ivy_dafny_compiler.py` | 478 | Dafny output |
| Others | ~243 | Lean, Markdown outputs |

**Tier 6 — UI/Visualization (~7,000+ lines):** Not a porting priority.

---

## PART 3: NEXT CHUNKS — DETAILED PLANS

### Chunk 8: `ast/` — AST Node Types

**Source**: `ivy_ast.py` (1,965 lines, 157 classes)

**Files to create**: `goivy/ast/ast.go`, `goivy/ast/formula.go`, `goivy/ast/decl.go`, `goivy/ast/action.go`, `goivy/ast/sort.go`, `goivy/ast/ast_test.go`

This is a separate AST from `logic/` — it represents the *parsed syntax tree* before compilation to the logic IR. The `logic/` package represents the *semantic* IR.

**Key base types:**
```go
// AST is the base for all syntax tree nodes (has line number info)
type AST struct {
    Lineno  Location    // source location
    Args    []ASTNode   // sub-expressions (recstruct pattern)
}

type ASTNode interface {
    GetLineno() Location
    GetArgs() []ASTNode
    String() string
}
```

**Formula hierarchy** (AST-level, NOT logic-level):
- `Symbol{Name string}` — identifier
- `Atom{Relname string, Terms []ASTNode}` — atomic formula/predicate
- `App{Func ASTNode, Args []ASTNode}` — function application
- `Variable{Name string, Sort ASTNode}` — variable with sort annotation
- `And`, `Or`, `Not`, `Implies`, `Iff`, `Ite` — logical connectives
- `Forall`, `Exists` — quantifiers with bounds
- `Globally`, `Eventually` — temporal operators
- `NamedBinder{Name string, Params []ASTNode, Body ASTNode}`
- `Old{Term ASTNode}` — temporal "old" operator
- `This{}` — self-reference

**Declaration hierarchy:**
- `ModuleDecl{Name, Args, Body}` — module definition
- `ObjectDecl{Name, Body}` — object definition
- `ActionDecl{Name, Params, Returns, Body}`
- `TypeDecl{Name, Sort}`, `VariantDecl`
- `RelationDecl{Name, Params}`, `ConstantDecl{Name, Sort}`
- `AxiomDecl`, `PropertyDecl`, `ConjectureDecl` — labeled formulas
- `InitDecl`, `ExportDecl`, `ImportDecl`, `IsolateDecl`
- `NativeDecl{Code string}`, `AttributeDecl`
- `MixinDecl{Kind string}` — before/after/implement mixins
- `InterpretDecl`, `DefinitionDecl`

**Action AST nodes** (syntax level, distinct from `ivy_actions.py` semantic actions):
- `AssignAction{Lhs, Rhs ASTNode}`
- `CallAction{Callee ASTNode, Args []ASTNode}`
- `IfAction{Cond, Then, Else ASTNode}`
- `WhileAction{Cond, Body ASTNode, Invariant ASTNode}`
- `Sequence{Actions []ASTNode}`
- `AssumeAction{Formula ASTNode}`, `AssertAction{Formula ASTNode}`
- `LocalAction{Vars []ASTNode, Body ASTNode}`

**AST Sort nodes:**
- `ConstantSort{Name string}` — uninterpreted
- `EnumeratedSort{Name string, Values []string}`
- `StructSort{Name string, Fields []ASTNode}`
- `FunctionSort{Domain []ASTNode, Range ASTNode}`
- `RelationSort{Domain []ASTNode}`

**Root node:**
- `Ivy{Body []ASTNode}` — top-level program, contains all declarations

**Tests**: Construction of each node type, String() output, Children/Args traversal.

### Chunk 9: `lexer/` — Hand-Written Lexer

**Source**: `ivy_lexer.py` (304 lines)

**Files to create**: `goivy/lexer/token.go`, `goivy/lexer/lexer.go`, `goivy/lexer/lexer_test.go`

**Design**: Hand-written lexer (no PLY equivalent in Go). Simple `Lexer` struct with `NextToken()`.

**token.go:**
```go
type TokenType int
const (
    // Punctuation
    LPAREN TokenType = iota; RPAREN; LCB; RCB; LB; RB
    COMMA; SEMI; COLON; DOT; DOTS; DOTDOTDOT; ARROW; ASSIGN
    // Operators
    PLUS; MINUS; TIMES; DIV; EQ; TILDAEQ; TILDA; LE; LT; GE; GT
    AND; OR; IFF; DOLLAR; CARET; PTO
    // Literals/identifiers
    SYMBOL      // lowercase identifier or quoted string
    VARIABLE    // uppercase identifier
    NATIVEQUOTE // <<<...>>>
    // Special
    EOF; NEWLINE; ERROR
)

type Token struct {
    Type   TokenType
    Value  string
    Line   int
    Column int
}
```

**Reserved keywords** (173 total, version-gated):
- Core: `relation`, `individual`, `function`, `axiom`, `conjecture`, `action`, `init`, `module`, `object`, ...
- Version 1.4+: `struct`, `ghost`, `alias`, `variant`, `of`, `constructor`
- Version 1.6+: `specification`, `implementation`, `parameter`, `tactic`
- Version 1.7+: `debug`, `for`, `process`, `subclass`, `template`

**lexer.go:**
```go
type Lexer struct {
    input   string
    pos     int
    line    int
    col     int
    version [2]int  // e.g., [1, 7]
}

func NewLexer(input string, version string) *Lexer
func (l *Lexer) NextToken() Token
func (l *Lexer) Peek() Token
```

Key behaviors:
- Identifiers starting with uppercase → `VARIABLE`
- Identifiers starting with lowercase → check reserved words → `SYMBOL` or keyword token
- Quoted strings `"..."` → `SYMBOL`
- Native code `<<<...>>>` → `NATIVEQUOTE`
- Comments `#...` → skip to end of line
- Multi-character operators: `->`, `<->`, `~=`, `*>`, `..`, `...`
- Version-gated keywords (e.g., `struct` is only a keyword in version ≥ 1.4)

**Tests**: Token-by-token verification for sample inputs, version-gated keyword tests, edge cases (empty input, unterminated strings, native quotes).

### Chunk 10: `parser/` — Hand-Written Recursive Descent Parser

**Source**: `ivy_parser.py` (3,161 lines, 279 grammar rules) + `ivy_logic_parser.py` (655 lines)

**Files to create**: `goivy/parser/parser.go`, `goivy/parser/expr.go`, `goivy/parser/decl.go`, `goivy/parser/action.go`, `goivy/parser/parser_test.go`

**Design**: Hand-written recursive descent with Pratt parsing for expressions. No PLY/yacc.

**parser.go:**
```go
type Parser struct {
    lexer   *lexer.Lexer
    current lexer.Token
    peek    lexer.Token
    errors  []ParseError
    version [2]int
}

func NewParser(input string, version string) *Parser
func (p *Parser) Parse() (*ast.Ivy, error)
```

**expr.go** — Expression/formula parsing (Pratt/precedence climbing):
```
Precedence (low to high):
1. SEMI (sequence in actions)
2. GLOBALLY, EVENTUALLY, temporal operators
3. IF/ELSE
4. OR (|)
5. AND (&)
6. TILDA (~, NOT)
7. Comparisons (=, <, <=, >, >=, ~=, *>)
8. COLON (sort annotation)
9. PLUS, MINUS
10. TIMES, DIV
11. DOLLAR ($, named binders)
12. DOT (field access)
13. Primary (atoms, parenthesized, quantifiers)
```

Key expression rules:
- `parseExpr(minPrec)` — Pratt parser core
- `parsePrimary()` — atoms, variables, parenthesized, quantifiers
- `parseQuantifier()` — `forall`/`exists` with variable lists
- `parseApp()` — function application `f(x, y)`
- `parseSortAnnotation()` — `X:Sort`

**decl.go** — Top-level declarations:
- `parseTopLevel()` — dispatch on keyword
- `parseModule()`, `parseObject()`, `parseClass()`
- `parseAction()`, `parseInit()`
- `parseRelation()`, `parseFunction()`, `parseIndividual()`
- `parseAxiom()`, `parseProperty()`, `parseConjecture()`
- `parseType()`, `parseVariant()`
- `parseIsolate()`, `parseExport()`, `parseImport()`
- `parseInterpret()`, `parseDefinition()`
- `parseNative()`, `parseAttribute()`

**action.go** — Action body parsing:
- `parseActionBody()` — sequence of statements
- `parseStatement()` — dispatch on keyword/pattern
- `parseIf()`, `parseWhile()`
- `parseAssign()` — `lhs := rhs`
- `parseCall()` — `call action(args)`
- `parseLocal()`, `parseLet()`
- `parseAssume()`, `parseAssert()`

**Tests**: Parse sample Ivy programs, verify AST structure. Key test cases:
- Simple declarations: `type t`, `relation r(X:t, Y:t)`
- Axioms: `axiom forall X:t. r(X, X)`
- Actions: `action a(x:t) = { ... }`
- Modules: `module m(t) = { ... }`
- Full programs from Ivy test suite

### Chunk 11: `ivylogic/` — Higher-Level Logic IR

**Source**: `ivy_logic.py` (1,774 lines)

**Files to create**: `goivy/ivylogic/sig.go`, `goivy/ivylogic/symbol.go`, `goivy/ivylogic/ivylogic.go`, `goivy/ivylogic/ivylogic_test.go`

This wraps `logic/` with:
- `Sig` — signature maintaining all symbols, sorts, and their relationships
- `Symbol` — higher-level constant/relation with name mangling
- Sort inference and resolution
- Formula classification: `IsQF()`, `IsPrenexUniversal()`, etc.
- `WithSymbols`, `WithSorts` — context-like sort/symbol scoping
- `VariableUniqifier` — unique variable renaming
- `PolySymsDict` — polymorphic symbol resolution

### Chunk 12: `actions/` — Action Semantics

**Source**: `ivy_actions.py` (1,687 lines, 49 classes)

Action class hierarchy (semantic level, post-compilation):
- `Action` base with `UpdateAction`, `AssumeAction`, `AssertAction`
- `AssignAction`, `HavocAction`, `SetAction`
- `IfAction`, `WhileAction`, `ChoiceAction`, `Sequence`
- `CallAction`, `LocalAction`, `LetAction`
- `NativeAction`, `CrashAction`, `ThunkAction`
- Type checking: `TypeCheckAction()`, `HasCode()`

### Chunk 13: `transrel/` — Transition Relations

**Source**: `ivy_transrel.py` (669 lines)

- State variable versioning (`new()`, `old()`, `rename()`)
- `ComposeUpdates()`, `JoinAction()`, `IteAction()`
- Forward interpolation and interpolant computation
- State-action composition

### Chunk 14: `compiler/` — AST → Logic IR

**Source**: `ivy_compiler.py` (2,320 lines)

- AST node → logic node compilation
- Type inference during compilation
- Field reference resolution
- Module instantiation with parameter substitution
- Context managers for scoping

---

## PART 4: IMPLEMENTATION STRATEGY

### Recommended Order

1. **Chunk 8: `ast/`** — AST node types (no parsing yet, just data structures). ~800 lines Go.
2. **Chunk 9: `lexer/`** — Hand-written tokenizer. ~400 lines Go.
3. **Chunk 10: `parser/`** — Recursive descent parser producing AST. ~1,500 lines Go.
4. **Chunk 11: `ivylogic/`** — Higher-level IR wrapping `logic/`. ~800 lines Go.
5. **Chunk 12: `actions/`** — Action semantics. ~700 lines Go.
6. **Chunk 13: `transrel/`** — Transition relations. ~400 lines Go.
7. **Chunk 14: `compiler/`** — AST → logic compilation. ~1,200 lines Go.

### Estimated Total: ~5,800 lines of Go (excluding tests)

### Verification After Each Chunk
1. `go build ./...` — compiles
2. `go test ./... -v` — all unit tests pass
3. `go test ./... -fuzz=. -fuzztime=30s` — fuzz tests find no crashes
4. `go vet ./...` — no warnings
5. Cross-reference with Python where applicable

### Key Python Source Files (for reference during implementation)
- `/Users/jaten/pyivy/ivy/ivy/ivy_ast.py` — AST node definitions
- `/Users/jaten/pyivy/ivy/ivy/ivy_lexer.py` — token definitions and lexer
- `/Users/jaten/pyivy/ivy/ivy/ivy_parser.py` — 279 grammar rules
- `/Users/jaten/pyivy/ivy/ivy/ivy_logic_parser.py` — formula sub-parser
- `/Users/jaten/pyivy/ivy/ivy/ivy_logic.py` — higher-level logic IR
- `/Users/jaten/pyivy/ivy/ivy/ivy_actions.py` — action semantics
- `/Users/jaten/pyivy/ivy/ivy/ivy_transrel.py` — transition relations
- `/Users/jaten/pyivy/ivy/ivy/ivy_compiler.py` — AST → logic compilation
- `/Users/jaten/pyivy/ivy/ivy/ivy_module.py` — module context

# PROGRESS UPDATE 2

# Ivy Python → Go Port: Detailed Status & Progress Report

**Date**: 2026-03-14
**Module path**: `github.com/glycerine/goivy`
**Branch**: `goport` (base: `master`)
**Python source**: `~/pyivy/ivy/ivy/` — 54,099 lines across 68 `.py` files

---

## PART 1: COMPLETED WORK — 12 Chunks

### Summary

| # | Package | Source | Tests | Fuzz | Python Source | Description |
|---|---------|--------|-------|------|---------------|-------------|
| 1-3 | `logic/` | 1,020 | 934 | 4 | `logic.py` (488 lines) | Sort, Node, Term, Formula types |
| 4 | `typeinfer/` | 857 | 250 | 1 | `type_inference.py` (419 lines) | Union-find type inference |
| 5 | `logicutil/` | 671 | 453 | 1 | `logic_util.py` (335 lines) | FreeVars, Substitute, EqualModAlpha |
| 6 | `ivyutils/` | 609 | 368 | 2 | `ivy_utils.py` (773 lines) | Renamer, Parameters, Graph algos |
| 7 | `z3bridge/` | 914 | 419 | **0** | `z3_utils.py` (197) + `ivy_solver.py` (1,716) | Z3 CGo wrapper + Translator |
| 8 | `ast/` | 2,466 | 642 | **0** | `ivy_ast.py` (1,965 lines) | 100+ AST node types |
| 9 | `lexer/` | 793 | 424 | **2** | `ivy_lexer.py` (304 lines) | Hand-written tokenizer |
| 10 | `parser/` | 1,803 | 787 | **2** | `ivy_parser.py` (3,161) + `ivy_logic_parser.py` (655) | Recursive descent parser |
| 11 | `ivylogic/` | 1,977 | 808 | 1 | `ivy_logic.py` (1,774 lines) | Higher-level logic IR, Sig, polymorphic symbols |
| 12 | `module/` | 491 | 268 | 1 | `ivy_module.py` (412 lines) | Module context/state container |
| | **TOTAL** | **11,601** | **5,353** | **14** | | |

**Grand total**: 16,954 Go lines across 41 files, 369 tests + 14 fuzz tests, 10 packages.
All `go vet` clean. All tests passing.

---

### Chunk 1-3: `logic/` — Core Logic Types

**Source**: `logic.py` (488 lines)
**Go files**: `sort.go` (209), `node.go` (82), `term.go` (161), `formula.go` (549), `error.go` (19)
**Tests**: `sort_test.go` (253), `term_test.go` (248), `formula_test.go` (433) — **53 tests, 4 fuzz**

**What's in it**:
- `Sort` interface (embeds `Node`) with seal method; 6 concrete types: `UninterpretedSort`, `BooleanSort`, `FunctionSort`, `EnumeratedSort`, `RangeSort`, `TopSort`
- `Node` interface: `NodeSort()`, `Children()`, `String()`, `Equal(Node)`
- Terms: `Var` (uppercase name validation), `Const`, `Apply` (arity/sort validation)
- 15 formula types: `Eq`, `Ite`, `Not`, `Globally`, `Eventually`, `WhenOperator`, `Cond`, `And`, `Or`, `Implies`, `Iff`, `ForAll`, `Exists`, `Lambda`, `NamedBinder`
- `True = &And{}`, `False = &Or{}`
- Helpers: `SortEqual`, `FirstOrderSort`, `ContainsTopSort`, `IsPolymorphic`, `IsBooleanOrTop`

**Fuzz tests**:
- `FuzzUninterpretedSort` — arbitrary name strings for sort construction
- `FuzzNewFunctionSort` — variable arity (0-20 args)
- 2 more in sort_test.go and term_test.go

**Design decisions**: `Sort` embeds `Node` (solves `Equal(Sort)` vs `Equal(Node)` conflict). Python `recstruct.__iter__` → Go `Children()`. Python `frozenset(variables)` → sorted display for deterministic `String()`.

---

### Chunk 4: `typeinfer/` — Type Inference

**Source**: `type_inference.py` (419 lines)
**Go files**: `sortvar.go` (52), `unify.go` (195), `infer.go` (610)
**Tests**: `typeinfer_test.go` (250) — **14 tests, 1 fuzz**

**What's in it**:
- `SortOrVar` interface with `SortWrapper` (wraps `logic.Sort`) and `SortVar` (mutable union-find node)
- `Find()` with path compression, `Unify()` with occurs check, `OccursIn()`
- `ConvertFromSortVars()`, `ConvertToSortVars()`, `InsertSortVars()`
- `InferSorts()` handles all 15+ node types, returns `InferResult{Sort, Concretize}`
- `ConcretizeSorts()` — replaces TopSorts with concrete sorts

**Fuzz tests**:
- `FuzzUnify` — unification of sort variables against various SortOrVar combinations

**Note**: `collectNames()` is a simplified stand-in; should eventually use `logicutil.FreeVariables`.

---

### Chunk 5: `logicutil/` — Logic Utilities

**Source**: `logic_util.py` (335 lines)
**Go files**: `logicutil.go` (671)
**Tests**: `logicutil_test.go` (453) — **22 tests, 1 fuzz**

**What's in it**:
- `FreeVariables(t)` — by pointer identity; `FreeVariablesByName(t)` — by name
- `UsedVariables(t)`, `BoundVariables(t)`, `UsedConstants(t)`
- `Substitute(t, subs)` — simultaneous substitution with `CaptureError` detection
- `IsTautologyEquality(t)` — detects `Eq(x,x)`
- `EqualModAlpha(t, u)` — alpha-equivalence using `pushableMap`

**Fuzz tests**:
- `FuzzSubstitute` — variable substitution with boolean formula combinations

---

### Chunk 6: `ivyutils/` — General Utilities

**Source**: `ivy_utils.py` (773 lines)
**Go files**: `renamer.go` (135), `parameter.go` (189), `graph.go` (144), `names.go` (141)
**Tests**: `ivyutils_test.go` (368) — **27 tests, 2 fuzz**

**What's in it**:
- `UniqueRenamer` — unique names with suffix collision avoidance
- `VariableGenerator` — A,B,...,Z,AA,BB,... (skipping O)
- `ConstantNameGenerator()` — yields a-z, a0-z0, a1-z1,...
- `Parameter`, `BooleanParameter`, `EnumeratedParameter` with `Registry`
- `TopologicalSort[T,K]`, `Reachable[T,K]`, `FindCycle[K]` (Go generics)
- `ComposeNames`, `SplitName`, `BaseName`, `ParentChildName`, `ExtractParametersName`

**Fuzz tests**:
- `FuzzUniqueRenamer` — arbitrary prefixes and names
- `FuzzSplitName` — robustness of name parsing

---

### Chunk 7: `z3bridge/` — Z3 Integration

**Source**: `z3_utils.py` (197 lines) + partial `ivy_solver.py` (1,716 lines)
**Go files**: `quantifier.go` (549), `translate.go` (365)
**Tests**: `z3bridge_test.go` (419) — **22 tests, 0 fuzz** ⚠️

**What's in it**:
- Self-contained Z3 CGo wrapper (NOT using go-z3, which lacks quantifier support)
- `Context` with thread-safe mutex, `Sort`, `Expr`, `FuncDecl` with `runtime.SetFinalizer` ref counting
- Boolean ops: `Not`, `And`, `Or`, `Implies`, `Iff`, `Eq`, `Ite`
- **Quantifiers**: `ForAll`, `Exists` via `Z3_mk_forall_const`/`Z3_mk_exists_const`
- `Solver` with `Assert`, `Check`, `Push`, `Pop`, `Model`
- `Translator` — converts all Ivy `logic.Node` types to Z3 `Expr`
- `Implies(f1, f2)`, `IsSat(f)` convenience functions
- CGo flags: `-I/usr/local/opt/z3/include -L/usr/local/opt/z3/lib -lz3`

**⚠️ Missing fuzz tests**: Should fuzz `Translator.Translate()` with random logic trees.

**Design decisions**: Direct CGo (not go-z3) because go-z3 lacks ForAll/Exists. `newExpr`/`newSort`/`newFuncDecl` must be called with lock already held (avoids nested `ctx.do()` deadlock).

---

### Chunk 8: `ast/` — AST Node Types

**Source**: `ivy_ast.py` (1,965 lines, 157 Python classes)
**Go files**: `ast.go` (564), `formula.go` (318), `decl.go` (1,145), `sort.go` (147), `tactic.go` (292)
**Tests**: `ast_test.go` (642) — **51 tests, 0 fuzz** ⚠️

**What's in it**:
- Base types: `Node` interface, `Base` struct, `Location`, `NoneAST`
- Core: `Symbol`, `Atom`, `App`, `Variable`, `Old`, `This`, `Literal`, `Dot`, `Bracket`, `Tuple`
- Formulas: `And`, `Or`, `Not`, `Implies`, `Iff`, `Ite`, `Forall`, `Exists`, `Globally`, `Eventually`, `WhenOperator`, `Let`, `Definition`, `NamedBinder`, `Isa`, `Trigger`
- 40+ declaration types: `ModuleDecl`, `ObjectDecl`, `ActionDecl`, `TypeDecl`, `RelationDecl`, `ConstantDecl`, `AxiomDecl`, `PropertyDecl`, `IsolateDecl`, `MixinDecl`, all variants
- 15+ tactic types for proof scripts
- AST-level sort types: `ConstantSort`, `EnumeratedSort`, `StructSort`, `FunctionSort`, `RelationSort`, `Range`
- Helpers: `IsTrue()`, `IsFalse()`, `HasTemporal()`, `IsEquals()`
- `LabeledFormula` with atomic counter for unique IDs

**⚠️ Missing fuzz tests**: Should fuzz `Clone()` round-trips and `String()` output.

---

### Chunk 9: `lexer/` — Hand-Written Tokenizer

**Source**: `ivy_lexer.py` (304 lines)
**Go files**: `token.go` (241), `lexer.go` (552)
**Tests**: `lexer_test.go` (382) — **35 tests, 0 fuzz** ⚠️

**What's in it**:
- 47 token types + 120+ reserved keywords with version-gating (v1.0 through v2.0)
- Multi-character operator disambiguation (`->` vs `-`, `<->` vs `<=` vs `<`, `...` vs `..` vs `.`)
- Native quote (`<<<...>>>`) support, quoted strings, comment handling
- Unicode temporal operators (□, ◇)
- `Peek()` / `NextToken()` interface, `Tokenize()` convenience function

**Fuzz tests**:
- `FuzzLexer` — Arbitrary byte input with Version{1,7}, loops NextToken() up to 10K times
- `FuzzLexerAllVersions` — Fuzzes both input and version numbers

---

### Chunk 10: `parser/` — Recursive Descent Parser

**Source**: `ivy_parser.py` (3,161 lines) + `ivy_logic_parser.py` (655 lines)
**Go files**: `parser.go` (328), `expr.go` (470), `decl.go` (738), `action.go` (267)
**Tests**: `parser_test.go` (693) — **59 tests, 0 fuzz** ⚠️

**What's in it**:
- Pratt/precedence-climbing expression parser with 15 precedence levels
- All top-level declarations: type, relation, individual, function, axiom, property, conjecture, action, init, module, object, isolate, export, import, mixin, variant, definition, schema, theorem, attribute, etc.
- Action body parsing: if/else, while, for, local, let, var, assignment, call, assume/assert/require/ensure
- Proof body parsing (simplified)

**Fuzz tests**:
- `FuzzParser` — Arbitrary strings to Parse(), 40 seed inputs covering all declaration types
- `FuzzParserExpr` — Expression parsing with 51 seed inputs covering all expression forms

---

### Chunk 11: `ivylogic/` — Higher-Level Logic IR

**Source**: `ivy_logic.py` (1,774 lines)
**Go files**: `sig.go` (381), `ivylogic.go` (382), `classify.go` (240), `formula.go` (247), `poly.go` (114), `util.go` (613)
**Tests**: `ivylogic_test.go` (808) — **54 tests, 1 fuzz**

**What's in it**:
- `Sig` — first-order signature (sorts, symbols, constructors, interpretations)
- `SymbolEntry` with `UnionSort` for polymorphic symbol overloading
- `WithSymbols`, `WithSorts` — scoped symbol/sort additions with Enter/Exit
- Formula classification: `IsQF`, `IsPrenexUniversal`, `IsPrenexExistential`, `IsAE`, `IsEA`
- Formula types: `Some`, `Definition`, `DefinitionSchema`, `Let`, `Literal`, `Predicate`
- Simplification: `SimpAnd`, `SimpOr`, `SimpNot`, `SimpIte`
- Polymorphic symbols table: 21 built-in entries + dynamic `bfe[...]`
- `VariableUniqifier` — alpha-converts formulas for unique variable names
- Utilities: `CloneNode`, `CloneBinder`, `NodeArgs`, `CloseFormula`, `Extensionality`, `PartialFunction`, `ASTMatch`, `LabelTemporal`, `NormalizeOps`, `AlphaAvoid`
- Type predicates: `IsVariable`, `IsConstant`, `IsApp`, `IsAtom`, `IsQuantifier`, `IsBinder`, `IsTemporal`, `HasTemporal`, `IsNumeral`, `IsGprop`, etc.

**Fuzz tests**:
- `FuzzSigAddSymbol` — arbitrary symbol names and sort names

---

### Chunk 12: `module/` — Module Context

**Source**: `ivy_module.py` (412 lines)
**Go files**: `module.go` (491)
**Tests**: `module_test.go` (268) — **16 tests, 1 fuzz**

**What's in it**:
- `Module` — central container for all module-level state (40+ fields)
- `LabeledFormula` — formula with label, line number, temporal flag
- `IsolateInfo`, `MixinTriple` — isolate metadata
- `Clear()`, `Copy()` — reset and shallow-copy
- `AddToHierarchy()`, `AddObject()` — hierarchy management
- `FindAction()`, `IsVariant()`, `VariantIndex()`, `SortCard()`, `SortDependencies()`
- Helper types: `NamedAction`, `ProofEntry`, `NamedEntry`, `SubgoalEntry`

**Fuzz tests**:
- `FuzzModuleCopy` — copy module with random sorts/symbols, verify isolation

---

## PART 2: FUZZ TEST GAP ANALYSIS

### Current State

| Package | Tests | Fuzz | Status |
|---------|-------|------|--------|
| `logic/` | 53 | 4 | ✅ Good coverage |
| `typeinfer/` | 14 | 1 | ✅ Adequate |
| `logicutil/` | 22 | 1 | ✅ Adequate |
| `ivyutils/` | 27 | 2 | ✅ Good coverage |
| `z3bridge/` | 22 | **0** | ⚠️ Missing |
| `ast/` | 51 | **0** | ⚠️ Missing |
| `lexer/` | 37 | **2** | ✅ Covered |
| `parser/` | 61 | **2** | ✅ Covered |
| `ivylogic/` | 54 | 1 | ✅ Covered |
| `module/` | 16 | 1 | ✅ Covered |

### Recommended Fuzz Tests to Add

**Priority 1 — Parser & Lexer** (crash-prone, arbitrary input):
1. **`FuzzLexer`** — Feed arbitrary `[]byte` to `Tokenize()`. Verify no panics, no infinite loops, always terminates with EOF or ERROR.
2. **`FuzzParser`** — Feed arbitrary strings to `Parse()`. Verify no panics. Optionally verify that if `Parse` succeeds, every returned node satisfies the `Node` interface.
3. **`FuzzParserRoundTrip`** — Parse valid Ivy snippets, `String()` the AST, re-parse. Check structural equivalence.

**Priority 2 — AST** (data structure integrity):
4. **`FuzzASTClone`** — Build random AST nodes, clone them, verify the clone is independent (modifying the clone doesn't affect the original).
5. **`FuzzASTString`** — Build random AST nodes, call `String()`, verify no panics.

**Priority 3 — Z3 Bridge** (CGo boundary):
6. **`FuzzTranslator`** — Build random `logic.Node` trees (using the existing logic types), translate to Z3 via `Translator.Translate()`. Verify no panics or Z3 crashes.

---

## PART 3: REMAINING PYTHON CODEBASE

### Python File Inventory (54,099 total lines)

**Already ported** (9,962 lines of Python → 9,133 lines of Go source):

| Python File | Lines | Go Package |
|-------------|-------|------------|
| `logic.py` | 488 | `logic/` |
| `type_inference.py` | 419 | `typeinfer/` |
| `logic_util.py` | 335 | `logicutil/` |
| `ivy_utils.py` | 773 | `ivyutils/` |
| `z3_utils.py` | 197 | `z3bridge/` |
| `ivy_solver.py` | 1,716 | `z3bridge/` (partial) |
| `ivy_ast.py` | 1,965 | `ast/` |
| `ivy_lexer.py` | 304 | `lexer/` |
| `ivy_parser.py` | 3,161 | `parser/` |
| `ivy_logic_parser.py` | 655 | `parser/` |

**Not yet ported** — organized by tier:

#### Tier 1: Core Semantics (4,542 lines) — Next to port

| File | Lines | Role | Key Dependencies |
|------|-------|------|------------------|
| `ivy_module.py` | 412 | Module context — holds all declarations, axioms, actions | logic, utils, solver, ast |
| `ivy_logic.py` | 1,774 | Higher-level logic IR — Symbol/Sig management, sort inference | (already partially in logic/) |
| `ivy_actions.py` | 1,687 | 49 action classes — imperative action semantics | logic, utils, transrel, module, ast |
| `ivy_transrel.py` | 669 | Transition relations — state versioning, interpolation | logic, utils, solver |

**Note**: `ivy_logic.py` (1,774 lines) and `ivy_transrel.py` (669 lines) are *partially* ported — the core data types are in `logic/` and `logicutil/`, but the higher-level wrappers (Sig management, state versioning, interpolation) still need porting.

#### Tier 2: Compiler & Verification (7,035 lines)

| File | Lines | Role | Key Dependencies |
|------|-------|------|------------------|
| `ivy_compiler.py` | 2,320 | AST → logic IR compilation | parser, actions, logic, module, theory, isolate |
| `ivy_isolate.py` | 2,022 | Modular verification — isolate extraction | logic, actions, module, solver |
| `ivy_check.py` | 1,040 | Top-level verification checker | compiler, isolate, proof, actions, module |
| `ivy_proof.py` | 1,653 | Proof state management and tactics | logic, utils, ast |

#### Tier 3: Symbolic Execution & Model Checking (3,118 lines)

| File | Lines | Role | Key Dependencies |
|------|-------|------|------------------|
| `ivy_interp.py` | 642 | Symbolic interpreter | logic, solver, transrel, actions |
| `ivy_art.py` | 523 | Abstract Reachability Tree (CEGAR) | actions, interp, module |
| `ivy_mc.py` | 1,772 | Model checking engine + AIGER encoding | module, actions, logic, transrel, art |
| `ivy_theory.py` | 181 | Built-in theories (int, nat, bv[n]) | actions, module, logic |

#### Tier 4: Support Libraries (5,577 lines)

| File | Lines | Role |
|------|-------|------|
| `ivy_logic_utils.py` | 1,635 | Clause operations, formula manipulation (partially in logicutil/) |
| `ivy_fragment.py` | 591 | EPR/decidable fragment detection |
| `ivy_unitres.py` | 546 | Unit resolution |
| `ivy_ranking.py` | 1,195 | Termination/ranking analysis |
| `ivy_l2s.py` | 1,522 | Liveness-to-safety reduction |
| `ivy_temporal.py` | 436 | Temporal property handling |
| `ivy_trace.py` | 427 | Counterexample trace generation |
| `ivy_alpha.py` | 225 | Alpha-renaming utilities |

#### Tier 5: Code Generation (8,375 lines) — Lower priority

| File | Lines | Role |
|------|-------|------|
| `ivy_to_cpp.py` | 6,715 | C++ code generation |
| `ivy_cpp_types.py` | 525 | C++ type system |
| `ivy_cpp.py` | 414 | C++ compilation utilities |
| `ivy_dafny_compiler.py` | 478 | Dafny output |
| `ivy_to_md.py` | 47 | Markdown output |
| Others | ~196 | Lean, SMT-LIB, VMT outputs |

#### Tier 6: UI/Visualization (~7,000 lines) — Not a porting priority

| File | Lines | Role |
|------|-------|------|
| `widget_analysis_session.py` | 1,648 | Analysis UI widget |
| `concept.py` | 865 | Concept visualization |
| `ivy_graph.py` | 783 | Graph display |
| `ivy_ui_cti.py` | 722 | CTI UI |
| `ivy_ui.py` | 622 | Main UI framework |
| `ivy_graph_ui.py` | 559 | Graph UI |
| `tk_graph_ui.py` | 430 | Tk graph widgets |
| Others | ~1,400 | Misc UI components |

### Totals by Tier

| Tier | Lines | Status |
|------|-------|--------|
| Already ported | 9,962 | ✅ Done (9,133 Go + 4,141 test) |
| Tier 1: Core Semantics | 4,542 | Next up |
| Tier 2: Compiler & Verification | 7,035 | Pending |
| Tier 3: Symbolic Exec & MC | 3,118 | Pending |
| Tier 4: Support Libraries | 5,577 | Pending |
| Tier 5: Code Generation | 8,375 | Lower priority |
| Tier 6: UI/Visualization | ~7,000 | Not planned |
| Small utilities | ~1,490 | As needed |

---

## PART 4: DETAILED PLANS FOR NEXT CHUNKS

### Chunk 11: `ivylogic/` — Higher-Level Logic IR

**Source**: `ivy_logic.py` (1,774 lines) — the parts NOT already in `logic/`

The `logic/` package has the core data types. `ivy_logic.py` adds:
- `Sig` — signature maintaining all symbols, sorts, and their relationships
- `Symbol` — higher-level constant/relation with name mangling
- Sort inference and resolution: `find_sort()`, `add_sort()`, `find_symbol()`
- Formula classification: `IsQF()`, `IsPrenexUniversal()`, etc.
- `WithSymbols`, `WithSorts` — context-like sort/symbol scoping
- `VariableUniqifier` — unique variable renaming
- `PolySymsDict` — polymorphic symbol resolution
- `extensionality()`, `exclusivity()` — axiom generators for enumerated sorts

**Estimated**: ~800 lines Go + 300 test

**Fuzz tests needed**:
- `FuzzSigAddSymbol` — add random symbols to a Sig, verify consistency

---

### Chunk 12: `module/` — Module Context

**Source**: `ivy_module.py` (412 lines)

Central container for all module-level state:
- `Module` class — holds declarations, relations, definitions, actions, axioms, schemata
- `clear()`, `copy()`, `get_axioms()`, `background_theory()`
- `call_graph()` — action call graph construction
- `sort_card()` — sort cardinality
- `IsolateInfo` — metadata about isolates

**Estimated**: ~400 lines Go + 200 test

**Fuzz tests needed**:
- `FuzzModuleCopy` — copy module, modify copy, verify original unchanged

---

### Chunk 13: `actions/` — Action Semantics

**Source**: `ivy_actions.py` (1,687 lines, 49 classes)

Action class hierarchy (semantic level, post-compilation):
- `Action` base with `UpdateAction`, `AssumeAction`, `AssertAction`
- `AssignAction`, `HavocAction`, `SetAction`
- `IfAction`, `WhileAction`, `ChoiceAction`, `Sequence`
- `CallAction`, `LocalAction`, `LetAction`
- `NativeAction`, `CrashAction`, `ThunkAction`
- `Schema` — parametrized action schemas
- `ActionContext`, `UnrollContext` — evaluation contexts
- Type checking: `type_check()`, `type_ast()`

**Dependencies**: logic, logicutil, ivyutils, transrel (partial), module, ast

**Estimated**: ~1,200 lines Go + 400 test

**Fuzz tests needed**:
- `FuzzActionTypeCheck` — random action trees, verify type checking doesn't panic

---

### Chunk 14: `transrel/` — Transition Relations (remaining)

**Source**: `ivy_transrel.py` (669 lines) — the parts NOT already in `logicutil/`

The remaining functions:
- Symbol renaming: `new()`, `is_new()`, `old()`, `is_old()`, `rename()`
- State operations: `null_update()`, `pure_state()`, `state_to_action()`, `action_to_state()`
- Update composition: `compose_updates()`, `join_action()`, `ite_action()`
- Interpolation: `History` class, `forward_image()`, `interpolant()`
- Constraint handling: `frame()`, `hide()`, `constrain_state()`

**Estimated**: ~500 lines Go + 200 test

---

### Chunk 15: `proof/` — Proof Checking

**Source**: `ivy_proof.py` (1,653 lines)

Self-contained proof engine:
- `ProofChecker` — main proof verification engine
- Schema instantiation: `inst_schema()`, `unify_with_goal()`, `match_goal()`
- Goal management: `make_goal()`, `traverse_goals()`, `list_goals()`
- `MatchProblem` — matching premises to goals
- Error types: `Redefinition`, `Circular`, `NoMatch`, `ProofError`, `CaptureError`

**Dependencies**: logic, logicutil, ast, ivyutils (relatively self-contained)

**Estimated**: ~900 lines Go + 400 test

**Fuzz tests needed**:
- `FuzzProofCheckerMatch` — random formula matching

---

### Chunk 16: `compiler/` — AST → Logic IR

**Source**: `ivy_compiler.py` (2,320 lines)

The big one:
- `IvyDeclInterp` — declaration interpreter
- Context classes: `Context`, `ReturnContext`, `ExprContext`
- Expression compilation: `compile_app()`, `compile_inline_call()`, `compile_method_call()`
- Declaration compilation: `ConstantDecl_cmpl()`, `compile_local()`
- Sort inference during compilation
- Module instantiation with parameter substitution

**Dependencies**: parser, actions, logic, module, theory, isolate — nearly everything

**Estimated**: ~1,500 lines Go + 600 test

---

### Chunk 17: `isolate/` — Modular Verification

**Source**: `ivy_isolate.py` (2,022 lines)

Isolate extraction and analysis:
- `lookup_action()`, `add_mixins()`, `summarize_action()`
- Strip/abstraction: `strip_sort()`, `strip_action()`, `strip_isolate()`
- Side effect analysis: `has_side_effect()`
- Dependency analysis: `get_calls_mods()`, `check_interference()`
- Path analysis utilities

**Dependencies**: logic, actions, module, solver, ast

**Estimated**: ~1,200 lines Go + 400 test

---

### Chunk 18: `theory/` — Built-in Theories

**Source**: `ivy_theory.py` (181 lines) — smallest file

- `Theory` base, `IntegerTheory`, `NaturalTheory`, `BitVectorTheory`
- `theories()`, `parse_theory()`, `get_sort_theory()`, `has_integer_interp()`

**Estimated**: ~150 lines Go + 100 test

---

### Chunk 19: `interp/` — Symbolic Interpreter

**Source**: `ivy_interp.py` (642 lines)

- `State` — abstract state with clauses and preconditions
- `EvalContext` — evaluation context
- State operations: `new_state()`, `concrete_post()`, `concrete_join()`
- `apply_action()` — symbolic action application

**Estimated**: ~500 lines Go + 200 test

---

### Chunk 20: `art/` — Abstract Reachability Tree

**Source**: `ivy_art.py` (523 lines)

- `AnalysisGraph` — CEGAR-style reachability analysis
- State tracking: `add()`, `push_state()`, `pop_state()`
- `apply_action()`, `join_states()`, `refine()`
- Witness management

**Estimated**: ~400 lines Go + 200 test

---

### Chunk 21: `check/` — Verification Driver

**Source**: `ivy_check.py` (1,040 lines)

Top-level verification:
- `Checker`, `ConjChecker` — proof checking drivers
- `check_conjectures()`, `check_temporals()`, `check_properties()`
- Counterexample display

**Dependencies**: Nearly everything

**Estimated**: ~700 lines Go + 300 test

---

### Chunk 22: `mc/` — Model Checking

**Source**: `ivy_mc.py` (1,772 lines)

- `Aiger` class — AIGER circuit format
- `Encoder` — formula-to-circuit encoding
- Schema/axiom instantiation for bounded checking
- Quantifier elimination: `Qelim` class

**Estimated**: ~1,000 lines Go + 400 test

---

## PART 5: IMPLEMENTATION STRATEGY

### Recommended Order

```
DONE: logic → typeinfer → logicutil → ivyutils → z3bridge → ast → lexer → parser
                                                                              |
DONE: ─── ivylogic ─── module                                                |
                          |                                                   |
NEXT: ─── theory ─── proof ───────────────────────────┐                       |
                          |                           |                       |
                     actions ── transrel              |                       |
                          |                           |                       |
                     interp ── art                    |                       |
                          |        \                  |                       |
                    compiler ──── isolate             |                       |
                          |                           |                       |
                     check ───────────────────────────┘                       |
                          |                                                   |
                       mc ─────────────────────────────────── (complete core) |
```

### Immediate Next Steps (Priority Order)

1. ~~Add fuzz tests to lexer and parser~~ ✅ Done
2. **Add fuzz test to z3bridge** — CGo boundary is crash-prone.
3. ~~Port Chunk 11: `ivylogic/`~~ ✅ Done
4. ~~Port Chunk 12: `module/`~~ ✅ Done
5. **Port Chunk 13: `actions/`** — action semantics (49 classes).
6. **Port Chunk 14: `transrel/`** — transition relations.
7. **Port Chunk 15: `proof/`** — proof checking.

### Verification After Each Chunk

1. `go build ./...` — compiles
2. `go test ./... -v` — all unit tests pass
3. `go test ./... -fuzz=. -fuzztime=30s` — fuzz tests find no crashes
4. `go vet ./...` — no warnings
5. Cross-reference with Python where applicable

### Estimated Remaining Work

| Phase | Chunks | Estimated Go Lines | Estimated Tests |
|-------|--------|-------------------|-----------------|
| ~~Fuzz gap fill~~ | ✅ | ~~+300~~ done | +4 fuzz added |
| ~~Core Semantics pt1~~ | ✅ 11-12 | ~~~1,200~~ done | ~1,076 done |
| Core Semantics pt2 | 13-14 | ~1,700 | ~600 |
| Compiler & Verification (Tier 2) | 15-17 | ~3,600 | ~1,400 |
| Symbolic Exec & MC (Tier 3) | 18-22 | ~2,750 | ~1,200 |
| **Subtotal (remaining core)** | | **~8,050** | **~3,200** |
| Support Libraries (Tier 4) | — | ~3,000 | ~1,200 |
| Code Generation (Tier 5) | — | ~5,000 | ~1,500 |
| **Full remaining total** | | **~16,050** | **~5,900** |

Combined with existing 16,954 lines, the full core port (through Tier 3) would be ~25,004 lines of Go.

---

## PART 6: KEY PYTHON SOURCE FILE PATHS

For reference during implementation:

```
/Users/jaten/pyivy/ivy/ivy/ivy_logic.py         — higher-level logic IR (1,774 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_module.py         — module context (412 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_actions.py        — action semantics (1,687 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_transrel.py       — transition relations (669 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_proof.py          — proof checking (1,653 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_compiler.py       — AST → logic compilation (2,320 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_isolate.py        — modular verification (2,022 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_theory.py         — built-in theories (181 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_interp.py         — symbolic interpreter (642 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_art.py            — abstract reachability (523 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_check.py          — verification driver (1,040 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_mc.py             — model checking (1,772 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_logic_utils.py    — clause operations (1,635 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_solver.py         — Z3 solver bridge (1,716 lines)
```

# PROGRESS UPDATE 3

# Ivy Python → Go Port: Detailed Status & Progress Report

**Date**: 2026-03-14
**Module path**: `github.com/glycerine/goivy`
**Branch**: `goport` (base: `master`)
**Python source**: `~/pyivy/ivy/ivy/` — 54,099 lines across 68 `.py` files

---

## PART 1: COMPLETED WORK — 23 Chunks

### Summary

| # | Package | Source | Tests | Fuzz | Python Source | Description |
|---|---------|--------|-------|------|---------------|-------------|
| 1-3 | `logic/` | 1,020 | 934 | 4 | `logic.py` (488 lines) | Sort, Node, Term, Formula types |
| 4 | `typeinfer/` | 857 | 250 | 1 | `type_inference.py` (419 lines) | Union-find type inference |
| 5 | `logicutil/` | 671 | 453 | 1 | `logic_util.py` (335 lines) | FreeVars, Substitute, EqualModAlpha |
| 6 | `ivyutils/` | 609 | 368 | 2 | `ivy_utils.py` (773 lines) | Renamer, Parameters, Graph algos |
| 7 | `z3bridge/` | 914 | 419 | **0** | `z3_utils.py` (197) + `ivy_solver.py` (1,716) | Z3 CGo wrapper + Translator |
| 8 | `ast/` | 2,466 | 642 | **0** | `ivy_ast.py` (1,965 lines) | 100+ AST node types |
| 9 | `lexer/` | 793 | 424 | **2** | `ivy_lexer.py` (304 lines) | Hand-written tokenizer |
| 10 | `parser/` | 1,803 | 787 | **2** | `ivy_parser.py` (3,161) + `ivy_logic_parser.py` (655) | Recursive descent parser |
| 11 | `ivylogic/` | 2,590 | 1,150 | 1 | `ivy_logic.py` (1,774 lines) | Higher-level logic IR, Sig, sort inference, EPR |
| 12 | `module/` | 1,050 | 580 | 1 | `ivy_module.py` (412 lines) | Module context, theory gen, canonize types |
| 13 | `actions/` | 1,216 | 647 | 2 | `ivy_actions.py` (1,687 lines) | Action semantics, 20+ action types, annotations |
| 14 | `transrel/` | 1,200 | 1,560 | 5 | `ivy_transrel.py` (669 lines) | Transition relations, composition, images |
| 15 | `proof/` | 1,261 | 846 | 2 | `ivy_proof.py` (1,653 lines) | Proof checker, matching, skolemization, goals |
| 16 | `theory/` | 279 | 337 | 1 | `ivy_theory.py` (181 lines) | Built-in theories (int/nat/bv) |
| 17 | `clauseops/` | 1,314 | 801 | 2 | `ivy_logic_utils.py` (1,635 lines) | Clauses type, clause ops, AST utilities |
| 18 | `compiler/` | 2,012 | 634 | 1 | `ivy_compiler.py` (2,320 lines) | AST→logic IR compilation, DeclInterp |
| 19 | `isolate/` | 939 | 855 | 1 | `ivy_isolate.py` (2,022 lines) | Modular verification, stripping, deps |
| 20 | `interp/` | 1,200 | 1,100 | 2 | `ivy_interp.py` (643 lines) | Symbolic interpreter, State, EvalContext |
| 21 | `art/` | 750 | 900 | 2 | `ivy_art.py` (523 lines) | Analysis graph, Z3-backed cover/BMC |
| 22 | `check/` | 1,050 | 700 | 1 | `ivy_check.py` (1,041 lines) | Top-level verification driver |
| 23 | `mc/` | 2,200 | 1,800 | 3 | `ivy_mc.py` (1,772 lines) | AIGER encoding, model checking engine |
| | **TOTAL** | **25,975** | **16,669** | **39** | | |

**Grand total**: 42,644 Go lines across 103 files, ~1,360 tests + 39 fuzz tests, 21 packages.
All `go vet` clean. All tests passing.

---

### Chunk 1-3: `logic/` — Core Logic Types

**Source**: `logic.py` (488 lines)
**Go files**: `sort.go` (209), `node.go` (82), `term.go` (161), `formula.go` (549), `error.go` (19)
**Tests**: `sort_test.go` (253), `term_test.go` (248), `formula_test.go` (433) — **53 tests, 4 fuzz**

**What's in it**:
- `Sort` interface (embeds `Node`) with seal method; 6 concrete types: `UninterpretedSort`, `BooleanSort`, `FunctionSort`, `EnumeratedSort`, `RangeSort`, `TopSort`
- `Node` interface: `NodeSort()`, `Children()`, `String()`, `Equal(Node)`
- Terms: `Var` (uppercase name validation), `Const`, `Apply` (arity/sort validation)
- 15 formula types: `Eq`, `Ite`, `Not`, `Globally`, `Eventually`, `WhenOperator`, `Cond`, `And`, `Or`, `Implies`, `Iff`, `ForAll`, `Exists`, `Lambda`, `NamedBinder`
- `True = &And{}`, `False = &Or{}`
- Helpers: `SortEqual`, `FirstOrderSort`, `ContainsTopSort`, `IsPolymorphic`, `IsBooleanOrTop`

**Fuzz tests**:
- `FuzzUninterpretedSort` — arbitrary name strings for sort construction
- `FuzzNewFunctionSort` — variable arity (0-20 args)
- 2 more in sort_test.go and term_test.go

**Design decisions**: `Sort` embeds `Node` (solves `Equal(Sort)` vs `Equal(Node)` conflict). Python `recstruct.__iter__` → Go `Children()`. Python `frozenset(variables)` → sorted display for deterministic `String()`.

---

### Chunk 4: `typeinfer/` — Type Inference

**Source**: `type_inference.py` (419 lines)
**Go files**: `sortvar.go` (52), `unify.go` (195), `infer.go` (610)
**Tests**: `typeinfer_test.go` (250) — **14 tests, 1 fuzz**

**What's in it**:
- `SortOrVar` interface with `SortWrapper` (wraps `logic.Sort`) and `SortVar` (mutable union-find node)
- `Find()` with path compression, `Unify()` with occurs check, `OccursIn()`
- `ConvertFromSortVars()`, `ConvertToSortVars()`, `InsertSortVars()`
- `InferSorts()` handles all 15+ node types, returns `InferResult{Sort, Concretize}`
- `ConcretizeSorts()` — replaces TopSorts with concrete sorts

**Fuzz tests**:
- `FuzzUnify` — unification of sort variables against various SortOrVar combinations

**Note**: `collectNames()` is a simplified stand-in; should eventually use `logicutil.FreeVariables`.

---

### Chunk 5: `logicutil/` — Logic Utilities

**Source**: `logic_util.py` (335 lines)
**Go files**: `logicutil.go` (671)
**Tests**: `logicutil_test.go` (453) — **22 tests, 1 fuzz**

**What's in it**:
- `FreeVariables(t)` — by pointer identity; `FreeVariablesByName(t)` — by name
- `UsedVariables(t)`, `BoundVariables(t)`, `UsedConstants(t)`
- `Substitute(t, subs)` — simultaneous substitution with `CaptureError` detection
- `IsTautologyEquality(t)` — detects `Eq(x,x)`
- `EqualModAlpha(t, u)` — alpha-equivalence using `pushableMap`

**Fuzz tests**:
- `FuzzSubstitute` — variable substitution with boolean formula combinations

---

### Chunk 6: `ivyutils/` — General Utilities

**Source**: `ivy_utils.py` (773 lines)
**Go files**: `renamer.go` (135), `parameter.go` (189), `graph.go` (144), `names.go` (141)
**Tests**: `ivyutils_test.go` (368) — **27 tests, 2 fuzz**

**What's in it**:
- `UniqueRenamer` — unique names with suffix collision avoidance
- `VariableGenerator` — A,B,...,Z,AA,BB,... (skipping O)
- `ConstantNameGenerator()` — yields a-z, a0-z0, a1-z1,...
- `Parameter`, `BooleanParameter`, `EnumeratedParameter` with `Registry`
- `TopologicalSort[T,K]`, `Reachable[T,K]`, `FindCycle[K]` (Go generics)
- `ComposeNames`, `SplitName`, `BaseName`, `ParentChildName`, `ExtractParametersName`

**Fuzz tests**:
- `FuzzUniqueRenamer` — arbitrary prefixes and names
- `FuzzSplitName` — robustness of name parsing

---

### Chunk 7: `z3bridge/` — Z3 Integration

**Source**: `z3_utils.py` (197 lines) + partial `ivy_solver.py` (1,716 lines)
**Go files**: `quantifier.go` (549), `translate.go` (365)
**Tests**: `z3bridge_test.go` (419) — **22 tests, 0 fuzz** ⚠️

**What's in it**:
- Self-contained Z3 CGo wrapper (NOT using go-z3, which lacks quantifier support)
- `Context` with thread-safe mutex, `Sort`, `Expr`, `FuncDecl` with `runtime.SetFinalizer` ref counting
- Boolean ops: `Not`, `And`, `Or`, `Implies`, `Iff`, `Eq`, `Ite`
- **Quantifiers**: `ForAll`, `Exists` via `Z3_mk_forall_const`/`Z3_mk_exists_const`
- `Solver` with `Assert`, `Check`, `Push`, `Pop`, `Model`
- `Translator` — converts all Ivy `logic.Node` types to Z3 `Expr`
- `Implies(f1, f2)`, `IsSat(f)` convenience functions
- CGo flags: `-I/usr/local/opt/z3/include -L/usr/local/opt/z3/lib -lz3`

**⚠️ Missing fuzz tests**: Should fuzz `Translator.Translate()` with random logic trees.

**Design decisions**: Direct CGo (not go-z3) because go-z3 lacks ForAll/Exists. `newExpr`/`newSort`/`newFuncDecl` must be called with lock already held (avoids nested `ctx.do()` deadlock).

---

### Chunk 8: `ast/` — AST Node Types

**Source**: `ivy_ast.py` (1,965 lines, 157 Python classes)
**Go files**: `ast.go` (564), `formula.go` (318), `decl.go` (1,145), `sort.go` (147), `tactic.go` (292)
**Tests**: `ast_test.go` (642) — **51 tests, 0 fuzz** ⚠️

**What's in it**:
- Base types: `Node` interface, `Base` struct, `Location`, `NoneAST`
- Core: `Symbol`, `Atom`, `App`, `Variable`, `Old`, `This`, `Literal`, `Dot`, `Bracket`, `Tuple`
- Formulas: `And`, `Or`, `Not`, `Implies`, `Iff`, `Ite`, `Forall`, `Exists`, `Globally`, `Eventually`, `WhenOperator`, `Let`, `Definition`, `NamedBinder`, `Isa`, `Trigger`
- 40+ declaration types: `ModuleDecl`, `ObjectDecl`, `ActionDecl`, `TypeDecl`, `RelationDecl`, `ConstantDecl`, `AxiomDecl`, `PropertyDecl`, `IsolateDecl`, `MixinDecl`, all variants
- 15+ tactic types for proof scripts
- AST-level sort types: `ConstantSort`, `EnumeratedSort`, `StructSort`, `FunctionSort`, `RelationSort`, `Range`
- Helpers: `IsTrue()`, `IsFalse()`, `HasTemporal()`, `IsEquals()`
- `LabeledFormula` with atomic counter for unique IDs

**⚠️ Missing fuzz tests**: Should fuzz `Clone()` round-trips and `String()` output.

---

### Chunk 9: `lexer/` — Hand-Written Tokenizer

**Source**: `ivy_lexer.py` (304 lines)
**Go files**: `token.go` (241), `lexer.go` (552)
**Tests**: `lexer_test.go` (382) — **35 tests, 0 fuzz** ⚠️

**What's in it**:
- 47 token types + 120+ reserved keywords with version-gating (v1.0 through v2.0)
- Multi-character operator disambiguation (`->` vs `-`, `<->` vs `<=` vs `<`, `...` vs `..` vs `.`)
- Native quote (`<<<...>>>`) support, quoted strings, comment handling
- Unicode temporal operators (□, ◇)
- `Peek()` / `NextToken()` interface, `Tokenize()` convenience function

**Fuzz tests**:
- `FuzzLexer` — Arbitrary byte input with Version{1,7}, loops NextToken() up to 10K times
- `FuzzLexerAllVersions` — Fuzzes both input and version numbers

---

### Chunk 10: `parser/` — Recursive Descent Parser

**Source**: `ivy_parser.py` (3,161 lines) + `ivy_logic_parser.py` (655 lines)
**Go files**: `parser.go` (328), `expr.go` (470), `decl.go` (738), `action.go` (267)
**Tests**: `parser_test.go` (693) — **59 tests, 0 fuzz** ⚠️

**What's in it**:
- Pratt/precedence-climbing expression parser with 15 precedence levels
- All top-level declarations: type, relation, individual, function, axiom, property, conjecture, action, init, module, object, isolate, export, import, mixin, variant, definition, schema, theorem, attribute, etc.
- Action body parsing: if/else, while, for, local, let, var, assignment, call, assume/assert/require/ensure
- Proof body parsing (simplified)

**Fuzz tests**:
- `FuzzParser` — Arbitrary strings to Parse(), 40 seed inputs covering all declaration types
- `FuzzParserExpr` — Expression parsing with 51 seed inputs covering all expression forms

---

### Chunk 11: `ivylogic/` — Higher-Level Logic IR

**Source**: `ivy_logic.py` (1,774 lines)
**Go files**: `sig.go` (381), `ivylogic.go` (382), `classify.go` (240), `formula.go` (247), `poly.go` (114), `util.go` (613)
**Tests**: `ivylogic_test.go` (808) — **54 tests, 1 fuzz**

**What's in it**:
- `Sig` — first-order signature (sorts, symbols, constructors, interpretations)
- `SymbolEntry` with `UnionSort` for polymorphic symbol overloading
- `WithSymbols`, `WithSorts` — scoped symbol/sort additions with Enter/Exit
- Formula classification: `IsQF`, `IsPrenexUniversal`, `IsPrenexExistential`, `IsAE`, `IsEA`
- Formula types: `Some`, `Definition`, `DefinitionSchema`, `Let`, `Literal`, `Predicate`
- Simplification: `SimpAnd`, `SimpOr`, `SimpNot`, `SimpIte`
- Polymorphic symbols table: 21 built-in entries + dynamic `bfe[...]`
- `VariableUniqifier` — alpha-converts formulas for unique variable names
- Utilities: `CloneNode`, `CloneBinder`, `NodeArgs`, `CloseFormula`, `Extensionality`, `PartialFunction`, `ASTMatch`, `LabelTemporal`, `NormalizeOps`, `AlphaAvoid`
- Type predicates: `IsVariable`, `IsConstant`, `IsApp`, `IsAtom`, `IsQuantifier`, `IsBinder`, `IsTemporal`, `HasTemporal`, `IsNumeral`, `IsGprop`, etc.

**Fuzz tests**:
- `FuzzSigAddSymbol` — arbitrary symbol names and sort names

---

### Chunk 12: `module/` — Module Context

**Source**: `ivy_module.py` (412 lines)
**Go files**: `module.go` (491)
**Tests**: `module_test.go` (268) — **16 tests, 1 fuzz**

**What's in it**:
- `Module` — central container for all module-level state (40+ fields)
- `LabeledFormula` — formula with label, line number, temporal flag
- `IsolateInfo`, `MixinTriple` — isolate metadata
- `Clear()`, `Copy()` — reset and shallow-copy
- `AddToHierarchy()`, `AddObject()` — hierarchy management
- `FindAction()`, `IsVariant()`, `VariantIndex()`, `SortCard()`, `SortDependencies()`
- Helper types: `NamedAction`, `ProofEntry`, `NamedEntry`, `SubgoalEntry`

**Fuzz tests**:
- `FuzzModuleCopy` — copy module with random sorts/symbols, verify isolation

---

### Chunk 13: `actions/` — Action Semantics

**Source**: `ivy_actions.py` (1,687 lines, 49 Python classes)
**Go files**: `action.go` (849), `annotation.go` (191), `helpers.go` (176)
**Tests**: `actions_test.go` (647) — **41 tests, 2 fuzz**

**What's in it**:
- `Action` interface with `Clone`, `String`, `IterCalls`, `IterSubactions`, `GetFormalParams`/`GetFormalReturns`
- 20+ concrete action types: `Sequence`, `AssumeAction`, `AssertAction`, `RequireAction`, `EnsureAction`, `AssignAction`, `HavocAction`, `SetAction`, `IfAction`, `WhileAction`, `ChoiceAction`, `CallAction`, `LocalAction`, `LetAction`, `BindOldsAction`, `NativeAction`, `CrashAction`, `ThunkAction`, `EnvAction`, `ReturnAction`, `IgnoreAction`
- `Schema` — parametrized action schema wrapping a definition
- `RME` — requires-modifies-ensures clause
- Annotation types: `EmptyAnnotation`, `ConjAnnotation`, `ComposeAnnotation`, `RenameAnnotation`, `IteAnnotation`
- Helper functions: `ConcatActions`, `HasCode`, `CallSet`, `PrefixAction`, `PostfixAction`, `ApplyMixin`, `ParamsToStr`, `ActionDefToStr`

**Note**: `IntUpdate()` methods (computing transition relations) are stubbed — they depend on `Clauses` from unported `ivy_logic_utils.py`.

**Fuzz tests**:
- `FuzzActionClone` — clone random action types, verify independence
- `FuzzAnnotation` — construct random annotation trees, verify String()

---

### Chunk 14: `transrel/` — Transition Relations

**Source**: `ivy_transrel.py` (669 lines)
**Go files**: `transrel.go` (459)
**Tests**: `transrel_test.go` (782) — **38 tests, 5 fuzz**

**What's in it**:
- Symbol renaming: `New()`, `IsNew()`, `NewOf()`, `Old()`, `IsOld()`, `OldOf()` — prefix-based symbol versioning
- Skolem detection: `IsSkolem()`, `IsGlobalSkolem()`
- `Update` type: `(Modified []string, TR lg.Node, Pre lg.Node)` — the core transition relation triple
- State operations: `NullUpdate()`, `PureState()`, `IsPureState()`, `TopState()`, `BottomState()`
- Frame operations: `FrameDef()`, `Frame()`, `DiffFrame()`, `UpdatedJoin()`
- Composition stubs: `ComposeUpdates()`, `JoinAction()`, `IteAction()`, `Hide()`, `StateToAction()`, `ActionToState()`, `ForwardImage()`
- Types: `CounterExample`, `ActionFailed`, `History`
- Utility: `ListDiff()`, `ListUnion()`

**Note**: Composition and image operations are stubbed — they require `Clauses` infrastructure from `ivy_logic_utils.py`.

**Fuzz tests**:
- `FuzzNewOld` — round-trip symbol renaming
- `FuzzIsSkolem` — arbitrary strings for skolem detection
- `FuzzIsGlobalSkolem` — arbitrary strings for global skolem detection
- `FuzzUpdatedJoin` — union of modified symbol lists
- `FuzzFrameDef` — frame definition construction

---

### Chunk 15: `proof/` — Proof Checker

**Source**: `ivy_proof.py` (1,653 lines)
**Go files**: `errors.go` (72), `match.go` (578), `goal.go` (338), `checker.go` (138), `skolem.go` (289)
**Tests**: `proof_test.go` (846) — **48 tests, 2 fuzz**

**What's in it**:
- Error types: `Redefinition`, `Circular`, `NoMatch`, `ProofError`, `CaptureError`
- `MatchProblem` — schema-instantiation matching problem
- `ProofChecker` — main proof verification engine with axioms, definitions, schemata
- Match/unification: `Match()`, `MatchQuants()`, `FOMatch()`, `HeadsMatch()`, `FuncsMatch()`, `MergeMatches()`, `EquivAlpha()`
- Match operations: `ApplyMatch()`, `ApplyMatchSym()`, `ApplyMatchFunc()`, `ApplyMatchFreesyms()`, `ComposeMatches()`, `ExtractTerms()`
- Goal operations: `GoalPrems()`, `GoalConc()`, `CloneGoal()`, `NormalizeGoal()`, `GoalVocab()`, `GoalFree()`, `TrivialGoal()`, `CheckConcsMatch()`
- Skolemization: `SkolemizeGoal()`, `SkolemizeFmla()` — converts goals to skolem normal form
- Symbol scope: `AddSymbols`, `RemoveSymbols` — temporary set modification with `Restore()`
- Tactic registry: `RegisteredTactics`, `RegisterTactic()`
- Helper types: `Vocab` (sorts, symbols, variables)

**Fuzz tests**:
- `FuzzMergeMatches` — random match merging with conflict detection
- `FuzzMatch` — pattern matching with free/bound variable combinations

---

## PART 2: FUZZ TEST GAP ANALYSIS

### Current State

| Package | Tests | Fuzz | Status |
|---------|-------|------|--------|
| `logic/` | 53 | 4 | ✅ Good coverage |
| `typeinfer/` | 14 | 1 | ✅ Adequate |
| `logicutil/` | 22 | 1 | ✅ Adequate |
| `ivyutils/` | 27 | 2 | ✅ Good coverage |
| `z3bridge/` | 22 | **0** | ⚠️ Missing |
| `ast/` | 51 | **0** | ⚠️ Missing |
| `lexer/` | 37 | **2** | ✅ Covered |
| `parser/` | 61 | **2** | ✅ Covered |
| `ivylogic/` | 54 | 1 | ✅ Covered |
| `module/` | 16 | 1 | ✅ Covered |
| `actions/` | 41 | 2 | ✅ Covered |
| `transrel/` | 38 | 5 | ✅ Good coverage |
| `proof/` | 48 | 2 | ✅ Covered |

### Recommended Fuzz Tests to Add

**Priority 1 — Parser & Lexer** (crash-prone, arbitrary input):
1. **`FuzzLexer`** — Feed arbitrary `[]byte` to `Tokenize()`. Verify no panics, no infinite loops, always terminates with EOF or ERROR.
2. **`FuzzParser`** — Feed arbitrary strings to `Parse()`. Verify no panics. Optionally verify that if `Parse` succeeds, every returned node satisfies the `Node` interface.
3. **`FuzzParserRoundTrip`** — Parse valid Ivy snippets, `String()` the AST, re-parse. Check structural equivalence.

**Priority 2 — AST** (data structure integrity):
4. **`FuzzASTClone`** — Build random AST nodes, clone them, verify the clone is independent (modifying the clone doesn't affect the original).
5. **`FuzzASTString`** — Build random AST nodes, call `String()`, verify no panics.

**Priority 3 — Z3 Bridge** (CGo boundary):
6. **`FuzzTranslator`** — Build random `logic.Node` trees (using the existing logic types), translate to Z3 via `Translator.Translate()`. Verify no panics or Z3 crashes.

---

## PART 3: REMAINING PYTHON CODEBASE

### Python File Inventory (54,099 total lines)

**Already ported** (13,971 lines of Python → 14,537 lines of Go source):

| Python File | Lines | Go Package |
|-------------|-------|------------|
| `logic.py` | 488 | `logic/` |
| `type_inference.py` | 419 | `typeinfer/` |
| `logic_util.py` | 335 | `logicutil/` |
| `ivy_utils.py` | 773 | `ivyutils/` |
| `z3_utils.py` | 197 | `z3bridge/` |
| `ivy_solver.py` | 1,716 | `z3bridge/` (partial) |
| `ivy_ast.py` | 1,965 | `ast/` |
| `ivy_lexer.py` | 304 | `lexer/` |
| `ivy_parser.py` | 3,161 | `parser/` |
| `ivy_logic_parser.py` | 655 | `parser/` |
| `ivy_logic.py` | 1,774 | `ivylogic/` |
| `ivy_module.py` | 412 | `module/` |
| `ivy_actions.py` | 1,687 | `actions/` |
| `ivy_transrel.py` | 669 | `transrel/` |
| `ivy_proof.py` | 1,653 | `proof/` |

**Not yet ported** — organized by tier:

**Note**: `ivy_logic.py` (1,774 lines) and `ivy_transrel.py` (669 lines) are *partially* ported — the core data types are in `logic/` and `logicutil/`, but the higher-level wrappers (Sig management, state versioning, interpolation) still need porting.

#### Tier 2: Compiler & Verification (7,035 lines)

| File | Lines | Role | Key Dependencies |
|------|-------|------|------------------|
| `ivy_compiler.py` | 2,320 | AST → logic IR compilation | parser, actions, logic, module, theory, isolate |
| `ivy_isolate.py` | 2,022 | Modular verification — isolate extraction | logic, actions, module, solver |
| `ivy_check.py` | 1,040 | Top-level verification checker | compiler, isolate, proof, actions, module |
| `ivy_proof.py` | 1,653 | Proof state management and tactics | logic, utils, ast |

#### Tier 3: Symbolic Execution & Model Checking (3,118 lines)

| File | Lines | Role | Key Dependencies |
|------|-------|------|------------------|
| `ivy_interp.py` | 642 | Symbolic interpreter | logic, solver, transrel, actions |
| `ivy_art.py` | 523 | Abstract Reachability Tree (CEGAR) | actions, interp, module |
| `ivy_mc.py` | 1,772 | Model checking engine + AIGER encoding | module, actions, logic, transrel, art |
| `ivy_theory.py` | 181 | Built-in theories (int, nat, bv[n]) | actions, module, logic |

#### Tier 4: Support Libraries (5,577 lines)

| File | Lines | Role |
|------|-------|------|
| `ivy_logic_utils.py` | 1,635 | Clause operations, formula manipulation (partially in logicutil/) |
| `ivy_fragment.py` | 591 | EPR/decidable fragment detection |
| `ivy_unitres.py` | 546 | Unit resolution |
| `ivy_ranking.py` | 1,195 | Termination/ranking analysis |
| `ivy_l2s.py` | 1,522 | Liveness-to-safety reduction |
| `ivy_temporal.py` | 436 | Temporal property handling |
| `ivy_trace.py` | 427 | Counterexample trace generation |
| `ivy_alpha.py` | 225 | Alpha-renaming utilities |

#### Tier 5: Code Generation (8,375 lines) — Lower priority

| File | Lines | Role |
|------|-------|------|
| `ivy_to_cpp.py` | 6,715 | C++ code generation |
| `ivy_cpp_types.py` | 525 | C++ type system |
| `ivy_cpp.py` | 414 | C++ compilation utilities |
| `ivy_dafny_compiler.py` | 478 | Dafny output |
| `ivy_to_md.py` | 47 | Markdown output |
| Others | ~196 | Lean, SMT-LIB, VMT outputs |

#### Tier 6: UI/Visualization (~7,000 lines) — Not a porting priority

| File | Lines | Role |
|------|-------|------|
| `widget_analysis_session.py` | 1,648 | Analysis UI widget |
| `concept.py` | 865 | Concept visualization |
| `ivy_graph.py` | 783 | Graph display |
| `ivy_ui_cti.py` | 722 | CTI UI |
| `ivy_ui.py` | 622 | Main UI framework |
| `ivy_graph_ui.py` | 559 | Graph UI |
| `tk_graph_ui.py` | 430 | Tk graph widgets |
| Others | ~1,400 | Misc UI components |

### Totals by Tier

| Tier | Lines | Status |
|------|-------|--------|
| Already ported | 9,962 | ✅ Done (9,133 Go + 4,141 test) |
| Tier 1: Core Semantics | 4,542 | Next up |
| Tier 2: Compiler & Verification | 7,035 | Pending |
| Tier 3: Symbolic Exec & MC | 3,118 | Pending |
| Tier 4: Support Libraries | 5,577 | Pending |
| Tier 5: Code Generation | 8,375 | Lower priority |
| Tier 6: UI/Visualization | ~7,000 | Not planned |
| Small utilities | ~1,490 | As needed |

---

## PART 4: DETAILED PLANS FOR NEXT CHUNKS

### Chunk 11: `ivylogic/` — Higher-Level Logic IR

**Source**: `ivy_logic.py` (1,774 lines) — the parts NOT already in `logic/`

The `logic/` package has the core data types. `ivy_logic.py` adds:
- `Sig` — signature maintaining all symbols, sorts, and their relationships
- `Symbol` — higher-level constant/relation with name mangling
- Sort inference and resolution: `find_sort()`, `add_sort()`, `find_symbol()`
- Formula classification: `IsQF()`, `IsPrenexUniversal()`, etc.
- `WithSymbols`, `WithSorts` — context-like sort/symbol scoping
- `VariableUniqifier` — unique variable renaming
- `PolySymsDict` — polymorphic symbol resolution
- `extensionality()`, `exclusivity()` — axiom generators for enumerated sorts

**Estimated**: ~800 lines Go + 300 test

**Fuzz tests needed**:
- `FuzzSigAddSymbol` — add random symbols to a Sig, verify consistency

---

### Chunk 12: `module/` — Module Context

**Source**: `ivy_module.py` (412 lines)

Central container for all module-level state:
- `Module` class — holds declarations, relations, definitions, actions, axioms, schemata
- `clear()`, `copy()`, `get_axioms()`, `background_theory()`
- `call_graph()` — action call graph construction
- `sort_card()` — sort cardinality
- `IsolateInfo` — metadata about isolates

**Estimated**: ~400 lines Go + 200 test

**Fuzz tests needed**:
- `FuzzModuleCopy` — copy module, modify copy, verify original unchanged

---

### Chunk 13: `actions/` — Action Semantics

**Source**: `ivy_actions.py` (1,687 lines, 49 classes)

Action class hierarchy (semantic level, post-compilation):
- `Action` base with `UpdateAction`, `AssumeAction`, `AssertAction`
- `AssignAction`, `HavocAction`, `SetAction`
- `IfAction`, `WhileAction`, `ChoiceAction`, `Sequence`
- `CallAction`, `LocalAction`, `LetAction`
- `NativeAction`, `CrashAction`, `ThunkAction`
- `Schema` — parametrized action schemas
- `ActionContext`, `UnrollContext` — evaluation contexts
- Type checking: `type_check()`, `type_ast()`

**Dependencies**: logic, logicutil, ivyutils, transrel (partial), module, ast

**Estimated**: ~1,200 lines Go + 400 test

**Fuzz tests needed**:
- `FuzzActionTypeCheck` — random action trees, verify type checking doesn't panic

---

### Chunk 14: `transrel/` — Transition Relations (remaining)

**Source**: `ivy_transrel.py` (669 lines) — the parts NOT already in `logicutil/`

The remaining functions:
- Symbol renaming: `new()`, `is_new()`, `old()`, `is_old()`, `rename()`
- State operations: `null_update()`, `pure_state()`, `state_to_action()`, `action_to_state()`
- Update composition: `compose_updates()`, `join_action()`, `ite_action()`
- Interpolation: `History` class, `forward_image()`, `interpolant()`
- Constraint handling: `frame()`, `hide()`, `constrain_state()`

**Estimated**: ~500 lines Go + 200 test

---

### Chunk 15: `proof/` — Proof Checking

**Source**: `ivy_proof.py` (1,653 lines)

Self-contained proof engine:
- `ProofChecker` — main proof verification engine
- Schema instantiation: `inst_schema()`, `unify_with_goal()`, `match_goal()`
- Goal management: `make_goal()`, `traverse_goals()`, `list_goals()`
- `MatchProblem` — matching premises to goals
- Error types: `Redefinition`, `Circular`, `NoMatch`, `ProofError`, `CaptureError`

**Dependencies**: logic, logicutil, ast, ivyutils (relatively self-contained)

**Estimated**: ~900 lines Go + 400 test

**Fuzz tests needed**:
- `FuzzProofCheckerMatch` — random formula matching

---

### Chunk 16: `compiler/` — AST → Logic IR

**Source**: `ivy_compiler.py` (2,320 lines)

The big one:
- `IvyDeclInterp` — declaration interpreter
- Context classes: `Context`, `ReturnContext`, `ExprContext`
- Expression compilation: `compile_app()`, `compile_inline_call()`, `compile_method_call()`
- Declaration compilation: `ConstantDecl_cmpl()`, `compile_local()`
- Sort inference during compilation
- Module instantiation with parameter substitution

**Dependencies**: parser, actions, logic, module, theory, isolate — nearly everything

**Estimated**: ~1,500 lines Go + 600 test

---

### Chunk 17: `isolate/` — Modular Verification

**Source**: `ivy_isolate.py` (2,022 lines)

Isolate extraction and analysis:
- `lookup_action()`, `add_mixins()`, `summarize_action()`
- Strip/abstraction: `strip_sort()`, `strip_action()`, `strip_isolate()`
- Side effect analysis: `has_side_effect()`
- Dependency analysis: `get_calls_mods()`, `check_interference()`
- Path analysis utilities

**Dependencies**: logic, actions, module, solver, ast

**Estimated**: ~1,200 lines Go + 400 test

---

### Chunk 18: `theory/` — Built-in Theories

**Source**: `ivy_theory.py` (181 lines) — smallest file

- `Theory` base, `IntegerTheory`, `NaturalTheory`, `BitVectorTheory`
- `theories()`, `parse_theory()`, `get_sort_theory()`, `has_integer_interp()`

**Estimated**: ~150 lines Go + 100 test

---

### Chunk 19: `interp/` — Symbolic Interpreter

**Source**: `ivy_interp.py` (642 lines)

- `State` — abstract state with clauses and preconditions
- `EvalContext` — evaluation context
- State operations: `new_state()`, `concrete_post()`, `concrete_join()`
- `apply_action()` — symbolic action application

**Estimated**: ~500 lines Go + 200 test

---

### Chunk 20: `art/` — Abstract Reachability Tree

**Source**: `ivy_art.py` (523 lines)

- `AnalysisGraph` — CEGAR-style reachability analysis
- State tracking: `add()`, `push_state()`, `pop_state()`
- `apply_action()`, `join_states()`, `refine()`
- Witness management

**Estimated**: ~400 lines Go + 200 test

---

### Chunk 21: `check/` — Verification Driver

**Source**: `ivy_check.py` (1,040 lines)

Top-level verification:
- `Checker`, `ConjChecker` — proof checking drivers
- `check_conjectures()`, `check_temporals()`, `check_properties()`
- Counterexample display

**Dependencies**: Nearly everything

**Estimated**: ~700 lines Go + 300 test

---

### Chunk 22: `mc/` — Model Checking

**Source**: `ivy_mc.py` (1,772 lines)

- `Aiger` class — AIGER circuit format
- `Encoder` — formula-to-circuit encoding
- Schema/axiom instantiation for bounded checking
- Quantifier elimination: `Qelim` class

**Estimated**: ~1,000 lines Go + 400 test

---

## PART 5: IMPLEMENTATION STRATEGY

### Recommended Order

```
DONE: logic → typeinfer → logicutil → ivyutils → z3bridge → ast → lexer → parser
                                                                              |
DONE: ─── ivylogic ─── module ─── actions ── transrel ─── proof              |
                                                              |               |
NEXT: ─── theory ─── compiler ───────────────────────┐        |               |
                          |                          |        |               |
                     interp ── art                   |        |               |
                          |        \                 |        |               |
                     isolate ──────────              |        |               |
                          |                          |        |               |
                     check ──────────────────────────┘        |               |
                          |                                   |               |
                       mc ────────────────────────────────────┘               |
```

### Immediate Next Steps (Priority Order)

1. ~~Add fuzz tests to lexer and parser~~ ✅ Done
2. **Add fuzz test to z3bridge** — CGo boundary is crash-prone.
3. ~~Port Chunk 11: `ivylogic/`~~ ✅ Done
4. ~~Port Chunk 12: `module/`~~ ✅ Done
5. ~~Port Chunk 13: `actions/`~~ ✅ Done
6. ~~Port Chunk 14: `transrel/`~~ ✅ Done
7. ~~Port Chunk 15: `proof/`~~ ✅ Done
8. **Port `ivy_logic_utils.py`** — Clauses type and clause operations (needed to fill in transrel/actions stubs).
9. **Port Chunk 16: `compiler/`** — AST → logic IR compilation.
10. **Port Chunk 17: `isolate/`** — modular verification.
11. **Port Chunk 18: `theory/`** — built-in theories.

### Verification After Each Chunk

1. `go build ./...` — compiles
2. `go test ./... -v` — all unit tests pass
3. `go test ./... -fuzz=. -fuzztime=30s` — fuzz tests find no crashes
4. `go vet ./...` — no warnings
5. Cross-reference with Python where applicable

### Estimated Remaining Work

| Phase | Chunks | Estimated Go Lines | Estimated Tests |
|-------|--------|-------------------|-----------------|
| ~~Fuzz gap fill~~ | ✅ | ~~+300~~ done | +4 fuzz added |
| ~~Core Semantics pt1~~ | ✅ 11-12 | ~~~1,200~~ done | ~1,076 done |
| ~~Core Semantics pt2~~ | ✅ 13-15 | ~~2,936~~ done | ~2,275 done |
| Compiler & Verification (Tier 2) | 16-17 | ~2,700 | ~1,000 |
| Symbolic Exec & MC (Tier 3) | 18-22 | ~2,750 | ~1,200 |
| **Subtotal (remaining core)** | | **~5,450** | **~2,200** |
| Support Libraries (Tier 4) | — | ~3,000 | ~1,200 |
| Code Generation (Tier 5) | — | ~5,000 | ~1,500 |
| **Full remaining total** | | **~13,450** | **~4,900** |

Combined with existing 22,165 lines, the full core port (through Tier 3) would be ~27,615 lines of Go.

---

## PART 6: KEY PYTHON SOURCE FILE PATHS

For reference during implementation:

```
/Users/jaten/pyivy/ivy/ivy/ivy_logic.py         — higher-level logic IR (1,774 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_module.py         — module context (412 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_actions.py        — action semantics (1,687 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_transrel.py       — transition relations (669 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_proof.py          — proof checking (1,653 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_compiler.py       — AST → logic compilation (2,320 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_isolate.py        — modular verification (2,022 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_theory.py         — built-in theories (181 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_interp.py         — symbolic interpreter (642 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_art.py            — abstract reachability (523 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_check.py          — verification driver (1,040 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_mc.py             — model checking (1,772 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_logic_utils.py    — clause operations (1,635 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_solver.py         — Z3 solver bridge (1,716 lines)
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

# "make golden" progress

~~~
255401  go : XTRACE: compiler.Thing ENTER type=LabeledFormula
        py : XTRACE: compiler.Thing ENTER type=LabeledFormula

255402  go : XTRACE: compiler.CompileNode ENTER type=LabeledFormula
        py : XTRACE: compiler.CompileNode ENTER type=LabeledFormula

255403  go : XTRACE: compiler.CompileNode return case=LabeledFormula
        py : XTRACE: compiler.CompileNode return case=LabeledFormula

255404  go : XTRACE: compiler.CompileLabeledFormula ENTER
        py : XTRACE: compiler.CompileLabeledFormula ENTER

255405  go : XTRACE: compiler.sortify_with_inference ENTER
        py : XTRACE: compiler.sortify_with_inference ENTER

255406  go : XTRACE: compiler.SigCheck@SortifyWithInference HASH leaf=blake3.33B-FsUCcfFu54vUdRZKUwmctGiFT3mysb1P0ETr2IrpLU4i root=blake3.33B-_p3bJ0ti2_n9Ic9XHTFGnEYHDMMe28WlRXPro1Zq4GFe canon=(sig sorts:[addr_type bool index lclock loc_type mem_loc_type mem_type op_ack_type op_type ph_type proc tar_cf_clock tar ...(truncated long line to 300 bytes)

        py : XTRACE: compiler.SigCheck@SortifyWithInference HASH leaf=blake3.33B-2Lceoz5TS6sIt0S-1T90V9FVWr7Anr2JXJhGfu6RQQ0f root=blake3.33B-UqT6FGtnaSmAbHPoFjC_VKf21Kwm6QNHxX2E_ie68q2Z canon=(sig sorts:[S addr_type bool index lclock loc_type mem_loc_type mem_type op_ack_type op_type ph_type proc tar_cf_clock t
 ...(truncated long line to 300 bytes)


=== S-expression diff (go '-' vs py '+') ===
   (sig
-   sorts:[addr_type
+   sorts:[S
-     bool
+     addr_type
-     index
+     bool
-     lclock
+     index
-     loc_type
+     lclock
-     mem_loc_type
+     loc_type
-     mem_type
+     mem_loc_type
-     op_ack_type
+     mem_type
-     op_type
+     op_ack_type
-     ph_type
+     op_type

    golden_test.go:452: ivy_check and goivy_check differ at line 255406, counting from 0.
~~~
