# CON-017 analyze-input note: single writer conformance class satisfied by absence

**Requirement:** CON-017 — "The Protodoc format SHALL define exactly one conformance class for writers."

**Status:** Satisfied.

**How it is satisfied:** By the *absence* of a second writer conformance class, not by
new logic. There is exactly one writer conformance class: any document produced by any
conforming writer must satisfy the identical validation pipeline (`pkg/validate`), with no
separate lenient / strict / legacy / transitional / migration writer-class flag or branch
anywhere in the writer or validation path.

**Evidence (CI-checkable):** Task **T-0126** ships the A-CLASS lint guard
`TestCON_017_SingleWriterConformanceClassAsserted` in
`pkg/validate/writerclass_test.go`. It scans every non-test `.go` file in the `validate`
package for any identifier naming a writer-class concept, or pairing a class-tier word
(`lenient`, `strict`, `transitional`, `legacy`) with a writer / class / mode context, and
fails CI if one is introduced. The guard currently passes with zero findings, which is the
positive evidence that no second writer conformance class exists. Introducing one would fail
that test and this note would no longer hold.

**For the analyze phase:** This note records the CON-017 disposition so the phase-5 analyze
step can consume it without further clarification: CON-017 is satisfied by absence, evidenced
by T-0126's passing lint guard. No document, code, or spec change is pending for CON-017.

**Split rationale:** This documentation deliverable was split out of T-0126's definition of
done as task **T-0360**, because the CI lint guard (T-0126) is independently verifiable by
test, whereas the "note added to phase-5 analyze inputs" clause is a separate documentation
artifact that the same test cannot verify. The unit test `TestCON_017_AnalyzeInputNotePresent`
verifies this note is present and states the required disposition.
