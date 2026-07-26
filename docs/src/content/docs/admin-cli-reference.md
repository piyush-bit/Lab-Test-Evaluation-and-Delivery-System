---
title: "Admin CLI Reference"
description: "Command reference guide for exercise authors and administrators (package, publish, evaluate, drive, admin)"
section: "reference-guides"
order: 2
---

# `euc2` Admin CLI Reference

This document provides a complete reference guide for the administrator and author-facing subcommands of the `euc2` command-line utility. Exercise authors and administrators use these commands to build, publish, grade, and manage exams.

> [!TIP]
> You can download the latest pre-compiled `euc2` executable for macOS, Linux, or Windows from the [GitHub Releases](https://github.com/piyush-bit/Lab-Test-Evaluation-and-Delivery-System/releases/latest) page.

---

## Command Hierarchy

Below is the structure of the admin-facing subcommands in the `euc2` command tree:

* [`euc2`](#euc2-root) — Base CLI utility
  * [`package`](#euc2-package) — Build public & private packages (Exercise Setter)
  * [`publish`](#euc2-publish) — Deploy exercise to registry server (Exercise Setter)
  * [`evaluate`](#euc2-evaluate) — Grade a student submission archive locally
  * [`drive`](#euc2-drive-group) — Manage drive-backed (offline) storage
    * [`prepare`](#euc2-drive-prepare) — Setup a drive directory for lab distribution
    * [`prepare-submission`](#euc2-drive-prepare-submission) — Setup a drive for receiving encrypted student envelopes
    * [`decrypt-submission`](#euc2-drive-decrypt-submission) — Decrypt a single student's envelope
    * [`evaluate-batch`](#euc2-drive-evaluate-batch) — Decrypt and grade all submissions in batch
  * [`admin`](#euc2-admin-group) — Server-based administrative commands
    * [`onboard-students`](#euc2-admin-onboard-students) — Pre-register student IDs (roster upload)
    * [`get-grades`](#euc2-admin-get-grades) — Download student scores (CSV/JSON/Table)
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

---

## `euc2 ui`

### Description
Starts a local HTTP server serving the embedded TDES Web Console. The Web Console provides instructors and administrators with an interactive visual dashboard to configure physical drives (preparing directories, generating X25519 keys), view collected student submission envelopes, and run batch evaluations with custom docker sandbox constraints and report exports.

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
