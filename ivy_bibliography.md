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
[https://www.wisdom.weizmann.ac.il/~padon/ivy.pdf](https://www.wisdom.weizmann.ac.il/~padon/ivy.pdf)
> Comprehensive treatment of the theory and implementation behind Ivy's interactive generalization approach.

---

## Project Resources

| Resource | URL |
|---|---|
| Microsoft Research project page | [https://www.microsoft.com/en-us/research/project/ivy/](https://www.microsoft.com/en-us/research/project/ivy/) |
| IVy documentation & tutorials | [https://microsoft.github.io/ivy/](https://microsoft.github.io/ivy/) |
| GitHub repository (microsoft/ivy) | [https://github.com/microsoft/ivy](https://github.com/microsoft/ivy) |
| QUIC formal specification (in Ivy) | [https://github.com/microsoft/ivy/blob/master/doc/examples/quic/README.md](https://github.com/microsoft/ivy/blob/master/doc/examples/quic/README.md) |
| Apple generic memory model (contributed to Ivy repo) | [https://github.com/microsoft/ivy](https://github.com/microsoft/ivy) |
| Ken McMillan's home page | [http://mcmil.net/wordpress/](http://mcmil.net/wordpress/) |
| McMillan publication list (dblp) | [https://dblp.org/pid/m/KennethLMcMillan.html](https://dblp.org/pid/m/KennethLMcMillan.html) |

---

## Notes

- The **Apple memory subsystem work** is described in the CAV 2024 paper "Toward Liveness Proofs at Scale." Apple engineers used Ivy to prove *safety* properties of a ~1,200-line memory subsystem model (78 invariants, ~500 lines of proof). McMillan extended Ivy's liveness support to handle their liveness conjecture (that every core operation is eventually retired). The generic model and proofs were contributed to the open-source Ivy repository.

- The **QUIC specification** work (SIGCOMM 2019) is among the highest-impact applications of Ivy's testing methodology outside hardware, and directly influenced the IETF standardization process for HTTP/3.

- The **RISC-V cache coherence** work (FMCAD 2016) predates the Apple engagement and established the pattern of using Ivy specifications to test hardware implementations compositionally.

- McMillan left Microsoft Research for **UT Austin** (Department of Computer Science), where Ivy research continues.

---

*Bibliography compiled March 2026. For the most current publication list, see McMillan's dblp page linked above.*
