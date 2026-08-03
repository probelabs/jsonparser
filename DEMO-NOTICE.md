# DEMO NOTICE — this branch intentionally contains defects

**Branch:** `proof-demo` · **Status:** synthetic showcase content · **Never merge, tag, or release.**

This `proof-demo` branch **deliberately re-introduces 12 defects** into the jsonparser library so
that the [proof-portal](https://github.com/) public showcase dashboard has a realistic, well-annotated
catalogue of findings to render. It is **not** a bug report against the real library and it is **not**
a release candidate.

- The defects are catalogued as known-issues in `proof/known-issues/KI-5.yaml` .. `KI-16.yaml`.
- Ten of them (KI-5 .. KI-14) are **open** — real, surgical regressions spanning several bug classes
  and severities (memory-safety / OOB, integer overflow, input-validation, contract/robustness).
- Two of them (KI-15, KI-16) are **fixed** — the current code is already correct; they document
  resolved behavior with GREEN reproducers.
- Each buggy site carries an in-code `// KI-<N> (proof-portal demo — INTENTIONAL DEFECT, do not ship)`
  comment, and each known-issue has a matching `// Reproduces: KI-<N>` reproducer test in
  `proof_demo_reproducers_test.go`.

## Do NOT

- **Do not merge this branch into `master`.**
- **Do not tag or cut a release from this branch.**
- **Do not "fix" the failing open-bug reproducer tests** — they are supposed to fail while the demo
  defects are present; that is the entire point of the showcase.

## The upstream library is unaffected

`master` (and every published release / tag) does **not** contain any of these defects. They exist
**only** on this throwaway showcase branch and were never shipped.
