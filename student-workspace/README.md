# go101-lab01: Slice-Backed Stack in Go

Implement a LIFO stack in Go using a slice as the underlying storage.

## Learning Objectives

- Use struct types and pointer receivers in Go.
- Manage a slice as a dynamic, bounded sequence.
- Return multiple values to signal success or failure without panicking.
- Write code that passes both visible and hidden tests.

## Problem Contract

Implement the five TODO methods in `stack.go`.  
Do **not** change the struct definition, field names, method signatures, or package name.

| Method | Signature | Behaviour |
|---|---|---|
| `Push` | `(val int)` | Add `val` to the top of the stack |
| `Pop` | `() (int, bool)` | Remove and return the top element; `(0, false)` if empty |
| `Peek` | `() (int, bool)` | Return the top element without removing it; `(0, false)` if empty |
| `IsEmpty` | `() bool` | Return `true` when the stack contains no elements |
| `Size` | `() int` | Return the number of elements currently in the stack |

## Files in this Lab Unit

- `manifest.json` — metadata and execution limits for the evaluator
- `go.mod` — Go module definition (do not edit)
- `stack.go` — your workspace; implement the TODOs here
- `stack_public_test.go` — public tests you can run locally
- `tests_private/` — private grading tests (not shipped to students)
- `reference/stack.go` — reference implementation
- `Makefile` — standardized entrypoints
- `run` — convenience wrapper

## Student Workflow

1. Implement the TODOs in `stack.go`.
2. Run local tests:

```bash
./run public
```

3. Iterate until all public tests pass.
4. Submit through the platform CLI.

## Instructor / Platform Workflow

```bash
make test-public      # validate with public tests
make test-submission-lifo
make test-submission-interleaved
make test-submission-size
make test-submission-pop-empty
make test-submission-peek
```

Each `test-submission-*` target copies `reference/stack.go` into the live slot,
injects `tests_private/stack_private_test.go`, runs exactly one private test,
then restores the directory whether the test passed or failed.
