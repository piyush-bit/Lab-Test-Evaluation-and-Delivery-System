# Exercise Authoring Specification

This document defines the rules and contracts every exercise designer must follow
when authoring a lab unit for this system.

---

## 1. Core Philosophy

> **The exercise is self-contained and self-verifying.**

The exercise package knows how to:
- Give the student a feedback loop while solving (`make test-public`)
- Grade a real student submission — by running each grading command declared in `manifest.json → grading`

The evaluator is intentionally dumb. It extracts the submission archive into
`reference/`, runs each grading command in sequence, and checks only exit codes.
The exercise does the rest.

---

## 2. Required Directory Structure

Every lab unit must follow this layout:

```
<course-id>-<lab-id>-<slug>/
├── manifest.json              ← machine-readable metadata (required)
├── README.md                  ← problem statement, objectives, contract (required)
├── Makefile                   ← must expose test-public and all grading targets (required)
├── run                        ← convenience wrapper script (required)
│
├── <...editable files...>     ← student's working files, wherever the project demands
│
├── <...test files...>         ← public and private tests, wherever the project demands
│
└── reference/                 ← the problem setter's own submission (see §4)
    └── <mirrors editable file paths exactly>
```

**The location of test files, source files, and build artifacts is up to the designer.**
The only structural rules are `manifest.json`, `Makefile`, `run`, and `reference/`.

---

## 3. Make Targets

Every exercise exposes one public target and one or more grading targets.

### `make test-public`

- **Used by:** The student, locally, while solving the exercise.
- **Tests against:** The student's editable file(s) at their live paths in the exercise.
- **Must:** Run the public test suite and exit `0` on full pass, non-zero on any failure.
- **Can live anywhere:** The public tests can be at any path the language or project demands.

### `make test-submission-<group>` (one per grading entry)

- **Used by:** The exercise designer (to validate the exercise) and the evaluator (to grade submissions).
- **These are the same commands, run the same way — that is the whole point.**
- **Precondition:** The submitted file(s) have been placed into `reference/` at their corresponding relative paths (see §4).
- **Must:** Read from `reference/`, run the designated subset of private tests, and exit `0` on pass, non-zero on failure.
- **Each target is self-contained** — it does its own inject-and-restore independently.
- **Strategy is up to the designer** — see §6 for allowed patterns.
- **Named in `manifest.json → grading`** alongside the points awarded for a passing exit code.

> There is no single `test-submission` target. Grading is composed of N independent commands, each worth a declared number of points.

---

## 4. The `reference/` Folder

`reference/` is the **submission slot** — the active interface that every grading target reads from.

- **Locally (designer):** contains the problem setter's correct solution. Running all grading targets against it validates that the exercise works.
- **During grading:** the evaluator drops the student's submitted files here before running each grading command. The same commands now grade the student's code.

Files inside `reference/` sit at the **same relative path** as they do in the live exercise:

```
exercise-root/
├── src/
│   └── module/
│       └── solution.go       ← student's working file
└── reference/
    └── src/
        └── module/
            └── solution.go   ← submission slot (designer's solution locally;
                                 student's submission during grading)
```

**Rules:**
- Must contain only the files listed in `manifest.json → submission.include_paths`.
- The designer's solution placed here must pass **all** grading targets.
- Every grading target must always read from `reference/` — never from the live paths directly.
- It is the **canonical example of a passing submission** and the grading interface in one.

---

## 5. Submission & Grading Flow

Packaging and grading follow the same principle: **preserve relative paths, exclude everything else.**

### Packaging (on submit)

```
Student triggers submission
        ↓
CLI reads manifest.json → submission.private_globs + submission.exclude_globs
        ↓
CLI packages the exercise directory, stripping all matched paths.
After packaging, CLI asserts that no private_globs file is present
in the archive — packaging fails loudly if any leaked in.
        ↓
Result: a .tar.gz archive where every remaining file sits
at the same relative path as in the live exercise
```

### Grading (by the evaluator)

```
Evaluator receives the submission archive
        ↓
Evaluator extracts submitted files into reference/
(each submitted file lands at its reference/ path,
e.g. stack.go → reference/stack.go)
        ↓
Evaluator reads manifest.json → grading[]
        ↓
For each { "command": "make test-submission-<group>", "points": N }:
    Evaluator runs the command
    Exit code 0 → award N points
    Non-zero   → award 0 points
        ↓
Final score = sum of awarded points
Recorded alongside max possible score (sum of all points in grading[])
```

The evaluator neither knows nor cares about the programming language, build system,
or test framework. The Makefile handles all of that.

---

## 6. Allowed Grading Target Implementation Strategies

The designer chooses the strategy that fits their language and build system.
The contract is always: **read from `reference/`, run the designated private tests, exit with the result.**
All three strategies share this invariant — they differ only in how the test runner
consumes the file from `reference/`.

Each grading target is self-contained: it handles its own setup, runs exactly its
assigned test group, and cleans up, regardless of whether it passes or fails.

### Strategy A — Compile directly from `reference/`

Pass `reference/<file>` directly to the compiler or interpreter.
No file movement needed.

```makefile
test-submission-group1:
    $(CC) $(CFLAGS) reference/src/solution.c tests/private_group1.c -o bin/test_group1
    @./bin/test_group1

test-submission-group2:
    $(CC) $(CFLAGS) reference/src/solution.c tests/private_group2.c -o bin/test_group2
    @./bin/test_group2
```

**When to use:** When the build tool accepts an explicit file path argument.

### Strategy B — Inject and restore

Copy `reference/<file>` into the live path, run the designated private tests
(which read from the live path), then restore the original file regardless of outcome.
Useful when the test runner or compiler requires the file to be at a fixed location.

A shared `define` macro keeps the inject-and-restore logic DRY across all grading targets:

```makefile
# Usage: $(call _grade,TestFuncName)
define _grade
    @cp src/solution.c src/solution.c.bak
    @cp reference/src/solution.c src/solution.c
    @$(CC) $(CFLAGS) src/solution.c tests/private_tests.c -o bin/test_sub && ./bin/test_sub -run $(1); EXIT=$$?; \
      cp src/solution.c.bak src/solution.c; \
      rm -f src/solution.c.bak; exit $$EXIT
endef

test-submission-group1:
    $(call _grade,group1_suite)

test-submission-group2:
    $(call _grade,group2_suite)
```

**When to use:** When the language runtime, import system, or build tool requires
the file to be at a specific path (e.g. Go test packages, Python imports).

### Strategy C — Environment variable or argument

Pass the `reference/` path to the test runner via an env var or flag.

```makefile
test-submission-group1:
    SUBMISSION=reference/src/solution.py python3 tests/private_group1.py

test-submission-group2:
    SUBMISSION=reference/src/solution.py python3 tests/private_group2.py
```

**When to use:** When the test runner already supports a configurable source path.

> **Rule for all strategies:** `make test-public` must never touch `reference/`.
> It always runs against the file at its live path, exactly as the student left it.

---

## 7. `manifest.json` Schema

```jsonc
{
  "lab_id": "go101-lab01",              // unique, matches directory name prefix
  "title": "Slice-Backed Stack in Go",
  "version": "1.1.0",                   // semver
  "language": "go",                     // c | python | go | lex | ...
  "runner_image": "lab-go-runner:v1.0", // Docker image used for grading.
                                       // Must provide /bin/sh and make.
                                       // Use a custom lab-runner image from the
                                       // private registry for languages whose
                                       // stock images lack these tools (Go, Python, Lex).
  "local_entrypoint": "make test-public",
  "grading": [                          // Ordered list of grading commands.
    {                                  // Evaluator runs each in sequence.
      "command": "make test-public",                  "points": 0, "public": true
    },
    {
      "command": "make test-submission-lifo", // Must match a Makefile target exactly.
      "points": 2,                     // Points awarded if exit code is 0.
      "public": false
    },
    {
      "command": "make test-submission-interleaved",
      "points": 2,
      "public": false
    },
    {
      "command": "make test-submission-size",
      "points": 2,
      "public": false
    },
    {
      "command": "make test-submission-pop-empty",
      "points": 2,
      "public": false
    },
    {
      "command": "make test-submission-peek",
      "points": 2,
      "public": false
    }
  ],                                   // Max score = sum of all points fields.
  "submission": {
    "include_paths": [                  // files the student edits and submits
      "stack.go"                        // paths relative to exercise root
    ],
    "private_globs": [                  // SECURITY: files that must never reach students.
      "reference/*",                   // CLI strips these AND asserts none leaked into
      "tests_private/*"                // the archive. Evaluator validates on arrival too.
    ],
    "exclude_globs": [                  // CLEANLINESS: build artifacts and temp files.
      "*.tar.gz"
    ]
  },
  "limits": {
    "memory_mb": 256,
    "timeout_seconds": 5               // Applied per grading command, not total.
    // "pids_limit": 32                // Optional process limit, when supported.
  }
}
```

**Rules:**
- `submission.include_paths` must match **exactly** the files inside `reference/`
  (at their corresponding relative paths under `reference/`).
- `grading[].command` values must exactly match Makefile target names. Mismatches are
  caught during exercise validation (the reference solution must pass every grading command).
- `grading[].public` is a boolean flag indicating whether the test command is public (executable locally by the student in the IDE) or private (only executed during grading).
- `grading[].points` defines points earned for a passing run. Diagnostic or public loop test commands (like `make test-public`) should be marked with `"public": true` and `"points": 0` as they do not affect the student's submission grade.
- `local_entrypoint` must always be `"make test-public"`.
- `grading` replaces `server_entrypoint`. There is no single `server_entrypoint` field.
- `reference/` and all private test directories must appear in `private_globs` — never in `exclude_globs`.
- `exclude_globs` is for build artifacts only (e.g. `bin/*`, `*.o`). Do not mix security-sensitive paths here.
- The CLI treats `private_globs` as a hard constraint: packaging fails if any match is found in the resulting archive.
- `runner_image` must provide `/bin/sh` and `make`. Stock language images (e.g. `golang:alpine`, distroless) often lack these — use a custom image from the private registry (`192.168.1.100:5000/lab-<lang>-runner:<tag>`) that layers the language runtime on top of a base that includes the required tools.
- `limits.timeout_seconds` is applied **per grading command**, not as a total budget.

---

## 8. Student Boilerplate Rules

The editable file(s) listed in `include_paths` must:

- **Compile and run as-is** (even with TODOs) — the student should be able to run
  `make test-public` immediately on a fresh checkout without fighting the build system.
- **Use `// TODO:` comments** (or language-appropriate equivalent) to mark every
  task the student must complete.
- **Stub unimplemented functions** so they compile cleanly (returning a safe default).
- **Contain no test logic** and no references to private tests or `reference/`.

---

## 9. The `run` Script

Every exercise must include a `run` script (executable, no extension) as a
convenience wrapper for students unfamiliar with Make:

```bash
#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-public}"

case "$MODE" in
  public) make test-public ;;
  clean)  make clean ;;
  *)
    echo "Usage: ./run [public|clean]"
    echo "  Grading is performed by the evaluator via manifest.json → grading commands."
    exit 1
    ;;
esac
```

---

## 10. Exercise Validation Checklist

Before publishing an exercise, the designer must verify all of the following:

- [ ] `make test-public` passes when run against the **reference** files
- [ ] `make test-public` produces failures when run against the **unmodified boilerplate**
- [ ] Every command in `manifest.json → grading` passes with the reference files — exit code `0`
- [ ] Every command in `manifest.json → grading` fails when the unmodified boilerplate is used — exit code non-zero
- [ ] `manifest.json → grading[].command` values match Makefile target names exactly
- [ ] `manifest.json → grading[].points` values are non-zero and sum to the intended max score
- [ ] `manifest.json → include_paths` matches exactly the files in `reference/` (by relative path)
- [ ] `manifest.json → private_globs` covers `reference/*` and all private test directories
- [ ] `manifest.json → exclude_globs` contains only build artifacts — no sensitive paths
- [ ] `reference/` contains no extra files beyond the submission files
- [ ] Student boilerplate compiles without errors out of the box
- [ ] `make clean` removes all build artifacts cleanly
- [ ] The `run` script is executable (`chmod +x run`)
