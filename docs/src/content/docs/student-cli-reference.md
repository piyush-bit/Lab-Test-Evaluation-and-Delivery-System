---
title: "Student CLI Reference"
description: "Command reference guide for student workflows (init, fetch, run, submit)"
section: "reference-guides"
order: 1
---

# `euc2` Student CLI Reference

This document provides a complete reference guide for the student-facing subcommands of the `euc2` command-line utility. Students use these commands to initialize workspaces, fetch lab templates, run verification tests, and submit their solutions.

> [!TIP]
> You can download the latest pre-compiled `euc2` executable for macOS, Linux, or Windows from the [GitHub Releases](https://github.com/piyush-bit/Lab-Test-Evaluation-and-Delivery-System/releases/latest) page.

---

## Command Hierarchy

Below is the structure of the student-facing subcommands in the `euc2` command tree:

* [`euc2`](#euc2-root) — Base CLI utility
  * [`init`](#euc2-init) — Initialize exercise workspace files
  * [`fetch`](#euc2-fetch) — Retrieve public exercise packages
  * [`run`](#euc2-run) — Run public tests in Docker sandbox or host
  * [`submit`](#euc2-submit) — Submit student solutions (Online/Drive)
  * [`ui`](#euc2-ui) — Start the local TDES Web UI dashboard

---

## `euc2` Root

### Description
Base command for packaging, caching, initializing, and running coding exercises.

### Usage
```bash
euc2 [command]
```

---

## `euc2 init`

### Description
Extracts the public template files of a cached exercise package into a target working directory to scaffold the student's workspace.

### Usage
```bash
euc2 init <lab-id>[@version] [working-directory]
```

### Arguments
* `<lab-id>[@version]`: The ID of the exercise. If version is omitted, uses the latest cached version.
* `[working-directory]`: Target directory (defaults to current directory `./`).

### Example
```bash
$ euc2 init go101-lab01@1.1.0 ./stack-lab
Exercise initialized in ./stack-lab
$ ls -A ./stack-lab
Makefile          README.md         manifest.json     run               stack.go
```

---

## `euc2 fetch`

### Description
Fetches the public package of an exercise from a physical drive path or remote registry server, saving the `.tar.gz` archive to the local workstation cache (`~/.euc2/cache/`).

### Usage
```bash
euc2 fetch <lab-id> [flags]
```

### Flags
* `-d, --drive <path>`: Local drive directory (for offline exams).
* `-r, --remote <url>`: Registry server URL (falls back to `EUC2_REGISTRY_URL` env var).
* `--org-id <id>`: Organization ID (required if using `--remote`).

### Example
```bash
# Offline/Drive Mode:
$ euc2 fetch go101-lab01 --drive /Volumes/ExamUSB
Fetching go101-lab01 from drive source...
Saved public archive to cache.

# Online Registry Mode:
$ euc2 fetch go101-lab01 --remote http://localhost:8080 --org-id default
Querying server registry for go101-lab01...
Fetched package successfully.
```

---

## `euc2 run`

### Description
Executes the public test suite defined in `manifest.json → local_entrypoint` (usually `make test-public`) inside an isolated Docker sandbox. Mounts the student's workspace files as read-only, except for designated submission paths.

### Usage
```bash
euc2 run [exercise-directory] [flags]
```

### Flags
* `--local`: Bypasses Docker entirely and runs tests directly on the host machine's environment.

### Example
```bash
# Docker sandboxed mode:
$ euc2 run
ℹ️ Running public tests in container...
✓ TestStackPush (0.01s)
✓ TestStackPop (0.01s)
PASS: 5/5 public tests completed successfully.

# Local host mode:
$ euc2 run --local
go test -v ./...
=== RUN   TestStackPush
--- PASS: TestStackPush (0.00s)
PASS
ok  	TDES/stack	0.005s
```

---

## `euc2 submit`

### Description
Packages the student's submission files (as defined in `manifest.json → submission.include_paths`) and delivers them either over HTTP to the registry server, or encrypts them as a secure JSON envelope on a local drive.

#### Student Authentication (Online Mode)
On the **first submission** (TOFU), the CLI prompts the student to create a private PIN (4+ characters). This PIN is cached locally in `~/.euc2/config.json` and attached to all future submissions for authentication.

### Usage
```bash
euc2 submit [flags]
```

### Flags
* `-d, --drive <path>`: Local drive directory to write the encrypted envelope.
* `-r, --remote <url>`: Remote server URL (falls back to `EUC2_REGISTRY_URL`).
* `--org-id <id>`: Organization ID.
* `--student-id <id>`: Student/Candidate ID.
* `--pin <pin>`: Override the cached PIN.
* `--update-pin <new-pin>`: Change the student's PIN on the registry database.

### Example
```bash
# Online Submission (TOFU Prompt):
$ euc2 submit --remote http://localhost:8080 --org-id org1 --student-id student1
No PIN found on this workstation.
Please enter a new PIN (minimum 4 characters) to secure your student ID: 1234
Submission accepted. Evaluation Score: 10/10

# Offline Drive Submission (Saves encrypted package to USB):
$ euc2 submit --drive /Volumes/ExamUSB --org-id org1 --student-id student1
Submission envelope created successfully.
Encrypted file written to: /Volumes/ExamUSB/submissions/go101-lab01/envelope_student1.json
```

---

## `euc2 ui`

### Description
Starts a local HTTP server serving the embedded TDES Web Console. The Web Console provides an interactive visual dashboard for students to view exercise instructions, edit files, execute tests, and submit grading packages without using the terminal directly.

### Usage
```bash
euc2 ui [flags]
```

### Flags
* `-p, --port <port>`: Port to bind local web console to (defaults to `8082`).
* `-o, --host <host>`: Host bind address (defaults to `127.0.0.1`).
* `--no-open`: Disable auto-opening the web browser.

### Example
```bash
$ euc2 ui
Starting TDES Web Console server on http://127.0.0.1:8082
Opening browser window...
```
