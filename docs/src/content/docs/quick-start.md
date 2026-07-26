---
title: "Quick Start Guide"
description: "Get TDES up and running in minutes — build, package, distribute, solve, and grade"
section: "get-started"
order: 0
---

# Quick Start Guide

Get TDES running on your machine and walk through a complete **offline drive workflow** end-to-end: **build → package an exercise → prepare a drive → fetch as a student → solve → submit → batch grade** — all without a network.

> TDES supports multiple transport layers. This guide uses the **Drive Transport** (flash drive / offline) as the primary pathway. If you have a networked environment, see the [Remote Server Alternative](#remote-server-alternative) section at the end.

---

## Prerequisites

Before you begin, make sure the following tools are installed:

| Tool | Version | Required For |
|---|---|---|
| **Go** | 1.25+ | Building the `euc2` CLI *(skip if downloading a pre-built binary)* |
| **Docker** | Any recent version | Sandboxed exercise execution and grading |
| **Node.js & npm** | 18+ | *(Optional)* Only needed if building or modifying the Web UI |

> **Tip:** Verify Docker is running with `docker info`. The evaluator and `euc2 run` both require a reachable Docker daemon.

---

## 1. Install `euc2`

### Option A: Download a Pre-built Binary

Grab the latest binary for your platform from the GitHub Releases page:

> **https://github.com/piyush-bit/Lab-Test-Evaluation-and-Delivery-System/releases**

Extract it, place it in your `PATH`, and skip ahead to [Step 2](#2-build-a-runner-image).

### Option B: Build from Source

```bash
git clone https://github.com/piyush-bit/Lab-Test-Evaluation-and-Delivery-System.git
cd Lab-Test-Evaluation-and-Delivery-System/TDES
go build -o euc2 ./cmd/euc2
```

*(Optional)* Build the Web UI into the binary:

```bash
cd ../frontend
npm install
npm run build     # outputs to TDES/cmd/euc2/ui_dist
cd ../TDES
go build -o euc2 ./cmd/euc2
```

---

## 2. Build a Runner Image

Exercises execute inside Docker containers using a **runner image**. Build the included Go runner:

```bash
cd runner-images/lab-go-runner
docker build -t lab-go-runner:v1.0 .
```

This image is based on `golang:1.22-alpine` with `make` installed — everything needed to compile and test Go exercises in isolation.

---

## 3. Package an Exercise (Instructor)

TDES ships with demo exercises. Let's package one for distribution:

```bash
./euc2 package ../demo_exercises/go101-lab01-stack
```

**What happens under the hood:**
1. The CLI runs the reference solution against private tests inside Docker to verify correctness.
2. It splits the exercise into a **public package** (student boilerplate) and a **private package** (hidden tests + reference solution).
3. Both packages are cached locally at `~/.euc2/cache`.

---

## 4. Prepare the Drive (Instructor)

This is where the offline workflow begins. You'll set up a flash drive (or any directory) with the exercise and a cryptographic keypair for secure submissions.

### Initialize the drive

```bash
./euc2 drive prepare /Volumes/ExamUSB
```

### Add the exercise to the drive

```bash
./euc2 drive add-exercise /Volumes/ExamUSB go101-lab01-stack 1.0.0
```

### Set up encrypted submissions

Generate an X25519 keypair for the exam session (use any standard tool), then prepare the drive with the public key:

```bash
./euc2 drive prepare-submission /Volumes/ExamUSB \
  --recipient-public-key <base64-x25519-public-key>
```

> **Keep the private key safe.** You'll need it later to decrypt and grade student submissions.

The drive is now ready to hand out to students. It contains the exercise packages and the recipient public key — no network needed.

---

## 5. Fetch & Solve (Student)

Students plug in the drive and work entirely offline.

### Fetch the exercise from the drive

```bash
./euc2 fetch go101-lab01-stack --drive /Volumes/ExamUSB
```

This copies the public package from the drive into the local cache at `~/.euc2/cache`.

### Initialize a workspace

```bash
./euc2 init go101-lab01-stack ./my-workspace
cd ./my-workspace
```

Your workspace now contains the boilerplate source files, public tests, a Makefile, and a README with the problem statement.

### Solve and test locally

Edit the source files (look for `// TODO:` markers), then run the public tests:

```bash
# Run inside a Docker sandbox (recommended)
../euc2 run

# Or run directly on your host
../euc2 run --local
```

Iterate until the public tests pass.

---

## 6. Submit to the Drive (Student)

When ready, submit the solution back to the drive:

```bash
../euc2 submit --drive /Volumes/ExamUSB --student-id STU001
```

The CLI reads the recipient public key from the drive, generates an ephemeral X25519 keypair, derives a shared AES-256-GCM key, and writes the **encrypted submission envelope** to the drive. No network, no server — the submission is cryptographically sealed and tamper-proof.

---

## 7. Batch Evaluate & Grade (Instructor)

After collecting all drives, the instructor decrypts and grades every submission in one pass:

```bash
./euc2 drive evaluate-batch /Volumes/ExamUSB \
  --recipient-private-key <base64-x25519-private-key> \
  --csv grades.csv \
  --output-dir results/
```

This:
1. Decrypts every student envelope using the private key.
2. Runs each submission through Docker sandbox evaluation against the private tests.
3. Exports a **summary CSV** (`grades.csv`) and individual **detailed JSON reports** in `results/`.

---

## 8. Launch the Web UI

TDES includes an embedded React-based workbench with a Monaco code editor, file browser, submit wizard, and admin dashboard:

```bash
./euc2 ui --port 8082
```

Your browser opens automatically at `http://127.0.0.1:8082`. From here you can browse exercises, edit code, run tests, and submit — all from a graphical interface.

---

## Remote Server Alternative

If you have a networked environment, TDES also supports an **online transport** using a central registry server. Here's the equivalent workflow:

### Start the Registry Server

```bash
# Build the server
go build -o registry-server ./cmd/server

# Set environment variables and start
export PORT=8080
export EUC2_INSTRUCTOR_TOKENS=instructor-secret-token
export EUC2_PIN_SALT=my-secure-salt
./registry-server
```

> **Health check:** `curl http://localhost:8080/healthz`

### Publish an exercise

```bash
./euc2 publish ../demo_exercises/go101-lab01-stack \
  --remote http://localhost:8080 \
  --org-id default \
  --bearer-token instructor-secret-token
```

### Fetch as a student

```bash
./euc2 fetch go101-lab01-stack \
  --remote http://localhost:8080 \
  --org-id default
```

### Submit online

```bash
./euc2 submit \
  --remote http://localhost:8080 \
  --org-id default \
  --student-id STU001 \
  --pin 1234
```

On your **first submission**, the PIN is registered via TOFU (Trust On First Use). The server evaluates your code in a Docker sandbox and returns your grade immediately.

> **Note:** The instructor must first onboard student IDs via `euc2 admin onboard-students`. See [Student Onboarding](/docs/userflow#2-student-onboarding-roster-ingestion) for details.

### Retrieve grades

```bash
./euc2 admin get-grades \
  --remote http://localhost:8080 \
  --org-id default \
  --lab-id go101-lab01-stack \
  --bearer-token instructor-secret-token \
  --csv grades.csv
```

### Environment Variables (Server)

| Variable | Description |
|---|---|
| `EUC2_INSTRUCTOR_TOKENS` | Comma-separated bearer tokens authorized for admin APIs |
| `EUC2_PIN_SALT` | Salt for hashing student TOFU PINs |
| `EUC2_REMOTE_BEARER_TOKEN` | Default bearer token for `--remote` commands |
| `PORT` | HTTP listen port (default: `8080`) |

---

## What's Next?

- **[User Flow Overview](/docs/userflow)** — Detailed sequence diagrams for every actor and workflow
- **[Exercise Authoring Specification](/docs/exercise-authoring-spec)** — How to create your own exercises
- **[Student CLI Reference](/docs/student-cli-reference)** — Full command reference for students
- **[Admin CLI Reference](/docs/admin-cli-reference)** — Instructor and admin command reference
- **[euc2 Web UI & Workflows](/docs/euc2-ui-workflow)** — Web UI features and workflows
