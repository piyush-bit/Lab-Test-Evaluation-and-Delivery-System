# `euc2` CLI Command Suite Reference

This document provides a complete reference guide for the `euc2` command-line utility. The utility is used by **Exercise Authors**, **Students (Candidates)**, and **Instructors/Admins** to manage, package, solve, submit, and evaluate lab exercises.

---

## Command Hierarchy

Below is the complete structure of the `euc2` command tree:

* [`euc2`](#euc2-root) — Base CLI utility
  * [`init`](#euc2-init) — Initialize exercise workspace files
  * [`fetch`](#euc2-fetch) — Retrieve public exercise packages
  * [`package`](#euc2-package) — Build public & private packages (Exercise Setter)
  * [`publish`](#euc2-publish) — Deploy exercise to registry server (Exercise Setter)
  * [`run`](#euc2-run) — Run public tests in Docker sandbox or host
  * [`submit`](#euc2-submit) — Submit student solutions (Online/Drive)
  * [`evaluate`](#euc2-evaluate) — Grade a student submission archive locally
  * [`drive`](#euc2-drive-group) — Manage drive-backed (offline) storage
    * [`prepare`](#euc2-drive-prepare) — Setup a drive directory for lab distribution
    * [`prepare-submission`](#euc2-drive-prepare-submission) — Setup a drive for receiving encrypted student envelopes
    * [`decrypt-submission`](#euc2-drive-decrypt-submission) — Decrypt a single student's envelope
    * [`evaluate-batch`](#euc2-drive-evaluate-batch) — Decrypt and grade all submissions in batch
  * [`admin`](#euc2-admin-group) — Server-based administrative commands
    * [`onboard-students`](#euc2-admin-onboard-students) — Pre-register student IDs (roster upload)
    * [`get-grades`](#euc2-admin-get-grades) — Download student scores (CSV/JSON/Table)

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

## `euc2 package`

### Description
(For Exercise Setters) Runs tests against the reference solution, strips sensitive assets according to `manifest.json` rules, and generates separate public (`_public.tar.gz`) and private (`_private.tar.gz`) archives in the local cache.

### Usage
```bash
euc2 package [exercise-directory]
```

### Example
```bash
$ euc2 package ./go101-lab01-stack
Packaging exercise: /Users/piyush/Developer/Test-delivery-and-evalution-system/demo_exercises/go101-lab01-stack
Public packages: [/var/tmp/euc2/cache/go101-lab01_1.1.0_public.tar.gz]
Private packages: [/var/tmp/euc2/private_packages/go101-lab01_1.1.0_private.tar.gz]
Exercise packaged and cached successfully.
```

---

## `euc2 publish`

### Description
(For Exercise Setters) Packages the exercise (executing validation tests locally) and publishes both public and private tarballs to the remote registry server.

### Usage
```bash
euc2 publish [exercise-directory] [flags]
```

### Flags
* `-r, --remote <url>`: Registry server URL (falls back to `EUC2_REGISTRY_URL`).
* `-t, --bearer-token <token>`: Instructor authentication token (falls back to `EUC2_REMOTE_BEARER_TOKEN`).

### Example
```bash
$ euc2 publish ./go101-lab01-stack --remote http://localhost:8080 --bearer-token admin-secret
Exercise published successfully. Version: 1.1.0
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

## `euc2 evaluate`

### Description
Grades a student's plaintext submission `.tar` archive locally inside a Docker container using private tests. If the private tests are missing in the local cache, the CLI pulls them from the registry server.

### Usage
```bash
euc2 evaluate <submission-tar-path> [flags]
```

### Flags
* `-o, --output <path>`: Path to write grading JSON output.
* `--docker-binary <path>`: Docker binary path or Docker socket URI.
* `--private-store <dir>`: Local private exercises package directory.
* `--registry-url <url>`: Registry server URL to retrieve private test packages.
* `--bearer-token <token>`: Registry authorization token.

### Example
```bash
$ euc2 evaluate ./student1.tar --output result.json
{
  "lab_id": "go101-lab01",
  "score": 8,
  "max_score": 10,
  "passed": false,
  "evaluated_at": "2026-06-21T12:05:00Z"
}
```

---

## `euc2 drive` Group

Local offline drive setup and grading utility commands.

---

### `euc2 drive prepare`

#### Description
Prepares a target directory structure (e.g. on a USB drive) for offline lab distribution, creating directories for public exercises.

#### Usage
```bash
euc2 drive prepare [drive-path]
```

#### Example
```bash
$ euc2 drive prepare /Volumes/ExamUSB
Drive prepared at: /Volumes/ExamUSB
```

---

### `euc2 drive prepare-submission`

#### Description
Sets up a submissions directory on a drive and registers the instructor's X25519 public key. This allows students to encrypt submissions directly on the drive.

#### Usage
```bash
euc2 drive prepare-submission [drive-path] [flags]
```

#### Flags
* `--recipient-public-key <base64-key>`: **(Required)** Base64 encoded X25519 public key.

#### Example
```bash
$ euc2 drive prepare-submission /Volumes/ExamUSB --recipient-public-key MCowBQYDK2VuAyEA...
Drive submission module prepared at: /Volumes/ExamUSB
```

---

### `euc2 drive decrypt-submission`

#### Description
Decrypts a student's encrypted JSON submission envelope using the instructor's X25519 private key, producing a plaintext `.tar` archive.

#### Usage
```bash
euc2 drive decrypt-submission [envelope-json-path] [flags]
```

#### Flags
* `-o, --output <path>`: **(Required)** Path to write the decrypted `.tar` file.
* `--recipient-private-key <base64-key>`: **(Required)** Base64 encoded X25519 private key.

#### Example
```bash
$ euc2 drive decrypt-submission /Volumes/ExamUSB/submissions/go101-lab01/envelope_student1.json -o ./student1.tar --recipient-private-key MC4CAQAwBQYDK2VuBCIE...
Decrypted submission package written to: ./student1.tar
```

---

### `euc2 drive evaluate-batch`

#### Description
Scans all submission envelopes in a drive directory, decrypts each using the private key, runs Docker evaluation sandboxes on the fly, and outputs consolidated CSV/JSON grade reports.

#### Usage
```bash
euc2 drive evaluate-batch [drive-path] [flags]
```

#### Flags
* `--recipient-private-key <base64-key>`: **(Required)** Base64 X25519 private key.
* `--csv <path>`: File path to save grading report in CSV format.
* `--json <path>`: File path to save grading report in JSON format.
* `--output-dir <dir>`: Directory to save individual student detailed result JSONs.
* `--docker-binary <path>`: Custom Docker binary path.
* `--registry-url <url>`: Registry server URL (if pulling private tests).
* `--bearer-token <token>`: Registry server bearer token.

#### Example
```bash
$ euc2 drive evaluate-batch /Volumes/ExamUSB --recipient-private-key MC4CAQAwBQYDK2VuBCIE... --csv summary.csv --output-dir results/
Scanning and decrypting submissions in: /Volumes/ExamUSB
Found 2 submission files. Starting batch evaluation...
[1/2] Processing envelope_student1.json...
  -> Success: Student=student1 Score=10/10
[2/2] Processing envelope_student2.json...
  -> Success: Student=student2 Score=8/10

=================================== BATCH SUMMARY ===================================
STUDENT ID      LAB ID          VERSION  STATUS       SCORE 
-------------------------------------------------------------------------------------
student1        go101-lab01     v1.0.0   success      10/10 
student2        go101-lab01     v1.0.0   success      8/10  
=====================================================================================
CSV report written to: summary.csv
```

---

## `euc2 admin` Group

Server-backed administration commands. Require valid instructor token credentials.

---

### `euc2 admin onboard-students`

#### Description
Pre-registers candidate student IDs on the remote registry server course roster via a CSV file to prevent hijackings.

#### Usage
```bash
euc2 admin onboard-students [roster-csv-path] [flags]
```

#### Flags
* `-r, --remote <url>`: Registry server base URL.
* `--bearer-token <token>`: Instructor authorization token.

#### Example
```bash
$ euc2 admin onboard-students ./roster.csv --remote http://localhost:8080 --bearer-token admin-token
Successfully onboarded roster. Server response: {"onboarded":12,"status":"ok"}
```

---

### `euc2 admin get-grades`

#### Description
Downloads student grades and scores from the registry server. Supports organization and lab filtering, and prints results to stdout or writes them to CSV/JSON files.

#### Usage
```bash
euc2 admin get-grades [flags]
```

#### Flags
* `-r, --remote <url>`: Registry server base URL.
* `--bearer-token <token>`: Instructor authorization token.
* `--org-id <id>`: Filter results by organization.
* `--lab-id <id>`: Filter results by lab.
* `--csv <path>`: Write grades report directly as a CSV file.
* `--json <path>`: Write grades report directly as a JSON file.

#### Example
```bash
$ euc2 admin get-grades --remote http://localhost:8080 --bearer-token admin-token --org-id org1
================================= REGISTRY GRADES ==================================
STUDENT ID      LAB ID          VERSION  STATUS       SCORE  SUBMITTED AT        
------------------------------------------------------------------------------------
student2        go101-lab01     v1.0.0   success      8/10   2026-06-21 12:05:00 
student1        go101-lab01     v1.0.0   success      10/10  2026-06-21 12:00:00 
====================================================================================
```
