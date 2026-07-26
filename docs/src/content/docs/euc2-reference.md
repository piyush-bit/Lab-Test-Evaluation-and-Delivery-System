---
title: "euc2 Workstation Reference"
description: "Workstation profile config, environment variables, subcommands and grading strategies"
section: "reference-guides"
order: 2
---

# Command line tool (`euc2`)

The `euc2` command-line tool packages, caches, initializes, and runs coding exercises for the Test Delivery and Evaluation System (TDES). It operates in both online (registry-backed) and offline (physical drive-backed) modes to deliver lab tests and evaluate submissions securely.

---

## Syntax

Use the following syntax to run `euc2` commands from your terminal window:

```bash
euc2 [command] [ARGS] [flags]
```

where `command`, `ARGS`, and `flags` are:

* **`command`**: Specifies the operation that you want to perform, for example `init`, `run`, `submit`, `fetch`, or a command group like `drive` or `admin`.
* **`ARGS`**: Specifies arguments unique to the subcommand, such as `<lab-id>[@version]`, `<working-directory>`, `<exercise-directory>`, `<submission-tar>`, or `<envelope-json-path>`.
* **`flags`**: Specifies optional flags. For example, you can use the `-r` or `--remote` flags to specify the registry server URL, or `-d` or `--drive` to specify a physical drive path.

> [!CAUTION]
> Flags that you specify from the command line override default values and any corresponding environment variables or configuration file values.

If you need help, run `euc2 --help` or `euc2 [command] --help` from the terminal window.

---

## In-Workstation Configuration and Environment Overrides

The `euc2` tool manages a local student profile and respects several environment variables for connection and security overrides.

### Student Profile Configuration (`~/.euc2/config.json`)

For student operations (such as `submit`), `euc2` looks for a student profile configuration file at `~/.euc2/config.json`. This file caches student identifiers to avoid repeated prompt entries:

```json
{
  "student_id": "student-12345",
  "org_id": "default",
  "pin": "9876"
}
```

* **`student_id`**: The candidate's unique identifier.
* **`org_id`**: The organization or course identifier (defaults to `default`).
* **`pin`**: The cached Trust-On-First-Use (TOFU) PIN. Used to authenticate subsequent submissions.

### Environment Variables

| Variable Name | Description | Subcommands Affected |
| :--- | :--- | :--- |
| `EUC2_REGISTRY_URL` | Specifies the default registry server URL. | `fetch`, `publish`, `submit`, `evaluate`, `drive evaluate-batch`, `admin onboard-students`, `admin get-grades` |
| `EUC2_REMOTE_BEARER_TOKEN` | Specifies the Bearer auth token for authenticating registry uploads or admin tasks. | `publish`, `evaluate`, `drive evaluate-batch`, `admin onboard-students`, `admin get-grades` |
| `EUC2_PRIVATE_STORE_DIR` | Overrides the path to the private exercise package store. | `evaluate`, `drive evaluate-batch` |

---

## Operations

Operations are grouped logically by workflow categories.

### Student core workflow
Commands used by candidates to initialize workspaces, run test suites locally, and submit solutions.

| Command | Usage | Description |
| :--- | :--- | :--- |
| [`init`](#euc2-init) | `euc2 init <lab-id>[@version] <working-dir>` | Initializes an exercise from the local cache into a working workspace. |
| [`run`](#euc2-run) | `euc2 run [exercise-dir]` | Runs the public test suite for the exercise in Docker or locally. |
| [`submit`](#euc2-submit) | `euc2 submit` | Packages and submits the candidate's files. |
| [`fetch`](#euc2-fetch) | `euc2 fetch <lab-id>` | Pulls public exercise templates from a drive or registry. |

### Authoring workflow (Exercise Setters)
Commands used by exercise authors to build and distribute labs.

| Command | Usage | Description |
| :--- | :--- | :--- |
| [`package`](#euc2-package) | `euc2 package <exercise-dir>` | Runs verification tests and splits private/public packages locally. |
| [`publish`](#euc2-publish) | `euc2 publish <exercise-dir>` | Packages and uploads the exercise to the remote registry. |

### Offline/Drive workflow (Examiners)
Utility commands to configure physical media and grade offline submissions.

| Command | Usage | Description |
| :--- | :--- | :--- |
| [`drive prepare`](#euc2-drive-prepare) | `euc2 drive prepare <drive-path>` | Scaffolds directories on physical media for offline exams. |
| [`drive prepare-submission`](#euc2-drive-prepare-submission) | `euc2 drive prepare-submission <drive-path>` | Registers the instructor's public key on a drive to enable encrypted uploads. |
| [`drive decrypt-submission`](#euc2-drive-decrypt-submission) | `euc2 drive decrypt-submission <envelope-json-path>` | Decrypts an encrypted JSON submission envelope to plaintext tar. |
| [`drive evaluate-batch`](#euc2-drive-evaluate-batch) | `euc2 drive evaluate-batch <drive-path>` | Decrypts and grades all student drive submissions in batch. |

### Registry administration workflow
Commands used by administrators to onboard classes and retrieve results.

| Command | Usage | Description |
| :--- | :--- | :--- |
| [`admin onboard-students`](#euc2-admin-onboard-students) | `euc2 admin onboard-students <roster-csv-path>` | Uploads the candidate roster to pre-register student IDs. |
| [`admin get-grades`](#euc2-admin-get-grades) | `euc2 admin get-grades` | Pulls graded submissions from the registry server. |

---

### `euc2 init`

Initialize an exercise into a working directory.

#### Usage
```bash
euc2 init <lab-id>[@version] <working-directory>
```

#### Arguments
* `<lab-id>[@version]`: The identifier of the lab to extract. An optional version can be specified using `@version` (e.g., `go101-lab01@1.1.0`). If the version is omitted, the command defaults to the latest cached version.
* `<working-directory>`: The target path where the student workspace files will be unpacked.

#### Examples
```bash
# Initialize a Go stack exercise into the stack-workspace folder
euc2 init go101-lab01@1.1.0 ./stack-workspace
```

---

### `euc2 run`

Run public tests for the current exercise.

#### Usage
```bash
euc2 run [exercise-directory] [flags]
```

#### Flags
* `--local`: Run the public test suite entrypoint directly on the host machine instead of inside an isolated Docker sandbox container.

#### Examples
```bash
# Run public tests inside the default Docker sandbox container
euc2 run ./stack-workspace

# Run public tests directly on the local host machine
euc2 run ./stack-workspace --local
```

---

### `euc2 submit`

Submit the candidate's solution files.

#### Usage
```bash
euc2 submit [flags]
```

#### Flags
* `-d, --drive <path>`: Local physical drive path (used to save the encrypted submission envelope offline).
* `-r, --remote <url>`: Remote registry server URL (used to post the submission online).
* `--org-id <id>`: Identifies the organization/course. Overrides cached configuration.
* `--student-id <id>`: Identifies the candidate. Overrides cached configuration.
* `--pin <pin>`: Candidate authentication PIN. Overrides cached configuration.
* `--update-pin <new-pin>`: Updates the current student PIN to a new value on successful submission.

> [!NOTE]
> During online remote submission, if no PIN is found on the workstation, the tool prompts the student to create a new secure PIN (TOFU) to lock their student ID.

#### Examples
```bash
# Submit online to a registry server (prompts for Student ID/PIN if not cached)
euc2 submit --remote http://localhost:8080 --org-id default

# Submit offline to a prepared USB drive (creates an encrypted envelope)
euc2 submit --drive /Volumes/ExamUSB --student-id student123 --org-id default
```

---

### `euc2 fetch`

Fetch a public exercise package.

#### Usage
```bash
euc2 fetch <lab-id>[@version] [flags]
```

#### Flags
* `-d, --drive <path>`: Local drive path containing the exercise repository.
* `-r, --remote <url>`: Registry server URL to retrieve the public package.
* `--org-id <id>`: Organization identifier (required if fetching from a remote server).

#### Examples
```bash
# Fetch from a physical drive path
euc2 fetch go101-lab01 --drive /Volumes/ExamUSB

# Fetch from a remote registry server
euc2 fetch go101-lab01 --remote http://localhost:8080 --org-id default
```

---

### `euc2 package`

Package an exercise locally for validation.

#### Usage
```bash
euc2 package <exercise-directory>
```

> [!IMPORTANT]
> The `package` command executes the public and private test suites against the author's reference solution. It fails if any tests fail, ensuring only passing exercise packages are built.

#### Examples
```bash
# Validate and package the stack exercise
euc2 package ./go101-lab01-stack
```

---

### `euc2 publish`

Package and publish an exercise to the remote registry.

#### Usage
```bash
euc2 publish <exercise-directory> [flags]
```

#### Flags
* `-r, --remote <url>`: Remote registry base URL.
* `--org-id <id>`: Organization ID.
* `--exercise-id <id>`: Optional override for the exercise identifier.
* `--version <version>`: Optional override for the exercise version.
* `--status <status>`: Optional override for the release status (defaults to `published`).

#### Examples
```bash
# Publish an exercise to the server registry
euc2 publish ./go101-lab01-stack --remote http://localhost:8080 --org-id default
```

---

### `euc2 evaluate`

Evaluate a student submission package locally.

#### Usage
```bash
euc2 evaluate <submission-tar> [flags]
```

#### Flags
* `-o, --output <path>`: Writes the evaluation JSON result to the specified file.
* `--private-store <path>`: Path to the private exercise package store (defaults to the local cache).
* `--docker-binary <path/URI>`: Docker command binary or socket URI used for the grading sandbox.
* `--registry-url <url>`: Registry base URL to pull private exercises if missing from local cache.
* `--bearer-token <token>`: Bearer token to authorize private exercise pulls.

#### Examples
```bash
# Grade a student tar package and save grading details to result.json
euc2 evaluate ./submission-student123.tar -o result.json
```

---

### `euc2 drive prepare`

Prepare a drive directory for offline exercise distribution.

#### Usage
```bash
euc2 drive prepare <drive-path>
```

#### Examples
```bash
# Prepare a USB drive folder structure
euc2 drive prepare /Volumes/ExamUSB
```

---

### `euc2 drive prepare-submission`

Prepare a drive directory for encrypted submissions.

#### Usage
```bash
euc2 drive prepare-submission <drive-path> [flags]
```

#### Flags
* `--recipient-public-key <base64-key>`: **(Required)** Base64 encoded X25519 public key.

#### Examples
```bash
# Configure a drive for encrypted uploads using an instructor's public key
euc2 drive prepare-submission /Volumes/ExamUSB --recipient-public-key MCowBQYDK2VuAyEA...
```

---

### `euc2 drive decrypt-submission`

Decrypt an encrypted JSON submission envelope.

#### Usage
```bash
euc2 drive decrypt-submission <envelope-json-path> [flags]
```

#### Flags
* `-o, --output <path>`: **(Required)** Path where the decrypted `.tar` file should be written.
* `--recipient-private-key <base64-key>`: **(Required)** Base64 encoded X25519 private key.

#### Examples
```bash
# Decrypt a student's envelope to a grading archive
euc2 drive decrypt-submission /Volumes/ExamUSB/submissions/go101-lab01/envelope_student123.json -o ./student123.tar --recipient-private-key MC4CAQAwBQYDK2VuBCIE...
```

---

### `euc2 drive evaluate-batch`

Decrypt and evaluate all student submissions on a drive.

#### Usage
```bash
euc2 drive evaluate-batch <drive-path> [flags]
```

#### Flags
* `--recipient-private-key <base64-key>`: **(Required)** Base64 encoded X25519 private key.
* `--csv <path>`: Path to write a summary CSV grading report.
* `--json <path>`: Path to write a summary JSON grading report.
* `--output-dir <dir>`: Directory where individual student detailed evaluation results (JSON) will be saved.
* `--private-store <path>`: Private exercise store directory.
* `--docker-binary <path/URI>`: Docker socket or binary path.
* `--registry-url <url>`: Registry URL to fetch missing private test archives.
* `--bearer-token <token>`: Registry authorization token.

#### Examples
```bash
# Decrypt and grade all submissions on a USB drive in batch
euc2 drive evaluate-batch /Volumes/ExamUSB --recipient-private-key MC4CAQAwBQYDK2VuBCIE... --csv report.csv --output-dir results/
```

---

### `euc2 admin onboard-students`

Onboard candidate student IDs to the registry server course roster.

#### Usage
```bash
euc2 admin onboard-students <roster-csv-path> [flags]
```

#### Flags
* `-r, --remote <url>`: Registry base URL.
* `--bearer-token <token>`: Instructor authorization token.

#### Examples
```bash
# Pre-register a class roster of student IDs on the registry server
euc2 admin onboard-students ./cs101-roster.csv --remote http://localhost:8080
```

---

### `euc2 admin get-grades`

Retrieve student grades from the registry server.

#### Usage
```bash
euc2 admin get-grades [flags]
```

#### Flags
* `-r, --remote <url>`: Registry base URL.
* `--bearer-token <token>`: Instructor authorization token.
* `--org-id <id>`: Filter results by organization.
* `--lab-id <id>`: Filter results by lab.
* `--csv <path>`: Save the grades directly to a CSV file.
* `--json <path>`: Save the grades directly to a JSON file.

#### Examples
```bash
# Fetch and print all grades for lab01 to the console
euc2 admin get-grades --remote http://localhost:8080 --lab-id go101-lab01

# Fetch and save organization grades to a CSV file
euc2 admin get-grades --remote http://localhost:8080 --org-id default --csv course-grades.csv
```

---

## Resource Types

In TDES, the primary managed resource is the **Exercise (Lab Unit)**.

### The Exercise Philosophy
An Exercise is self-contained and self-verifying. It contains public test suites for candidate validation and private test targets for final grading.

### Directory Structure of an Exercise

```
<exercise-root>/
├── manifest.json              ← Machine-readable metadata (schema, limits, image runner)
├── README.md                  ← Problem description and specifications
├── Makefile                   ← Entry points for tests (test-public, test-submission-*)
├── run                        ← Executable wrapper script
├── reference/                 ← Solution slot (contains author's passing solution files)
└── <source & test files>      ← Source code files and test files
```

### `manifest.json` Schema

| Property | Type | Description |
| :--- | :--- | :--- |
| `lab_id` | `string` | Unique exercise identifier matching folder prefix. |
| `title` | `string` | Human-readable title of the exercise. |
| `version` | `string` | Semantic versioning (`semver`). |
| `language` | `string` | Programming language (e.g., `go`, `python`, `c`). |
| `runner_image` | `string` | Docker image name in registry used for sandboxed execution. |
| `local_entrypoint`| `string` | Command run locally by students (always `make test-public`). |
| `grading` | `array` | List of objects containing `command` and `points` pairs. |
| `submission` | `object` | Specifies `include_paths` (files to package) and globs. |
| `limits` | `object` | Specifies `memory_mb` and `timeout_seconds` per test command. |

---

## Output Options

The `euc2` CLI outputs structured information for machines and clean console interfaces for users.

### Format Options

* **JSON Grading Result (`euc2 evaluate`)**:
  Grading outputs an evaluation summary schema:
  ```json
  {
    "student_id": "student-12345",
    "lab_id": "go101-lab01",
    "version": "1.1.0",
    "status": "success",
    "earned_points": 8,
    "max_points": 10
  }
  ```

* **CSV Grading Report (`euc2 drive evaluate-batch` or `euc2 admin get-grades`)**:
  Tabulates student scores into standard comma-separated fields:
  `student_id,lab_id,version,earned_points,max_points,status,envelope_file,error`

---

## Examples: Common Operations

Familiarize yourself with typical workflow cycles using the following examples:

### Student workflow cycle

```bash
# 1. Fetch the public lab package from USB drive
euc2 fetch go101-lab01 --drive /Volumes/ExamUSB

# 2. Extract and initialize workspace stack-lab
euc2 init go101-lab01@1.1.0 ./stack-lab

# 3. Solve the exercise and run local public tests
cd ./stack-lab
euc2 run --local

# 4. Submit the completed solution offline to the USB drive
euc2 submit --drive /Volumes/ExamUSB --student-id student-12345
```

### Exercise setter workflow cycle

```bash
# 1. Package and validate reference solution locally
euc2 package ./go101-lab01-stack

# 2. Publish exercise to the remote registry
euc2 publish ./go101-lab01-stack --remote http://localhost:8080 --org-id default
```

### Administrator offline grading cycle

```bash
# 1. Decrypt student envelope to plaintext tar
euc2 drive decrypt-submission /Volumes/ExamUSB/submissions/go101-lab01/envelope_student123.json -o ./student123.tar --recipient-private-key MC4CAQAwBQYDK2VuBCIE...

# 2. Evaluate decrypted package in sandbox container
euc2 evaluate ./student123.tar -o result.json
```

---

## Examples: Implementing Grading Strategies

Exercise designers use different strategies in the lab's `Makefile` to evaluate student code placed in the `reference/` folder.

### Strategy A: Compile directly from `reference/`
Pass the files in `reference/` directly to the compiler command:

```makefile
test-submission-group1:
	$(CC) $(CFLAGS) reference/src/solution.c tests/private_group1.c -o bin/test_group1
	@./bin/test_group1
```

### Strategy B: Inject and restore
Copy the file from `reference/` into the live code path, run tests, and restore the original workspace state:

```makefile
define _grade
	@cp src/solution.go src/solution.go.bak
	@cp reference/src/solution.go src/solution.go
	@go test -v -run $(1); EXIT=$$?; \
	cp src/solution.go.bak src/solution.go; \
	rm -f src/solution.go.bak; exit $$EXIT
endef

test-submission-lifo:
	$(call _grade,TestStackLifo)
```

### Strategy C: Environment variable injection
Provide the solution file path dynamically via an environment variable:

```makefile
test-submission-group1:
	SUBMISSION=reference/src/solution.py python3 tests/private_group1.py
```
