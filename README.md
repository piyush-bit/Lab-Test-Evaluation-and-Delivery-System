# TDES — Transport, Delivery, and Evaluation System

A pluggable system for packaging, distributing, submitting, and automatically grading coding lab exercises — online or fully offline.

🌐 **[tdes.bypiyush.com](https://tdes.bypiyush.com)** · 📦 **[Releases](https://github.com/piyush-bit/Lab-Test-Evaluation-and-Delivery-System/releases)** · 📖 **[Documentation](https://tdes.bypiyush.com/docs/quick-start)**

---

## Overview

TDES separates the exercise lifecycle (authoring → distribution → submission → evaluation) from the **transport mechanism**. The CLI (`euc2`) dispatches to different transport implementations at runtime, and the evaluation engine grades submissions identically regardless of how they arrived.

| Transport | Flag | Description |
|---|---|---|
| **Server Transport** | `--remote <url>` | Online mode via a central HTTP registry server |
| **Drive Transport** | `--drive <path>` | Fully end-to-end offline using USB flash drives with X25519/AES-GCM encryption |

### Key Features

- **Pluggable transports** — online server or offline flash drive, same CLI
- **Docker-sandboxed evaluation** — language-agnostic grading inside isolated containers
- **Encrypted offline submissions** — X25519 ECDH key exchange + AES-256-GCM
- **TOFU PIN authentication** — Trust On First Use identity binding for students
- **Embedded Web UI** — React + Monaco Editor workbench baked into the CLI binary
- **Batch grading** — decrypt and evaluate an entire drive of submissions in one command

---

## Quick Start

### Download a Pre-built Binary

Grab the latest release for your platform:

> **https://github.com/piyush-bit/Lab-Test-Evaluation-and-Delivery-System/releases**

Extract it, place it in your `PATH`, and you're ready to go.

### Build from Source

**Prerequisites:** Go 1.25+, Docker, Node.js 18+


```bash
# Clone repository
git clone https://github.com/piyush-bit/Lab-Test-Evaluation-and-Delivery-System.git
cd Lab-Test-Evaluation-and-Delivery-System

# 1. Build frontend assets (REQUIRED)
cd frontend
npm install
npm run build     # outputs to TDES/cmd/euc2/ui_dist

# 2. Build the euc2 CLI and registry server
cd ../TDES
go build -o euc2 ./cmd/euc2
go build -o registry-server ./cmd/server
```

### Build a Runner Image

```bash
cd runner-images/lab-go-runner
docker build -t lab-go-runner:v1.0 .
```

---

## Offline Drive Workflow (End-to-End)

The primary workflow — no server, no network:

```bash
# 1. Package an exercise
./euc2 package ./demo_exercises/go101-lab01-stack

# 2. Prepare a USB drive
./euc2 drive prepare /Volumes/ExamUSB
./euc2 drive add-exercise /Volumes/ExamUSB go101-lab01-stack 1.0.0
./euc2 drive prepare-submission /Volumes/ExamUSB \
  --recipient-public-key <base64-public-key>

# 3. Student fetches from the drive
./euc2 fetch go101-lab01-stack --drive /Volumes/ExamUSB
./euc2 init go101-lab01-stack ./workspace
cd ./workspace

# 4. Student solves and submits
./euc2 run                    # run public tests in Docker
./euc2 submit --drive /Volumes/ExamUSB --student-id STU001

# 5. Instructor batch evaluates
./euc2 drive evaluate-batch /Volumes/ExamUSB \
  --recipient-private-key <base64-private-key> \
  --csv grades.csv
```

---

## Online Server Workflow

```bash
# Start the registry server
export EUC2_INSTRUCTOR_TOKENS=instructor-secret-token
export EUC2_PIN_SALT=my-secure-salt
./registry-server

# Publish an exercise
./euc2 publish ./demo_exercises/go101-lab01-stack \
  --remote http://localhost:8080 --org-id default \
  --bearer-token instructor-secret-token

# Fetch and solve as a student
./euc2 fetch go101-lab01-stack --remote http://localhost:8080 --org-id default
./euc2 init go101-lab01-stack ./workspace && cd ./workspace
./euc2 run

# Submit online
./euc2 submit --remote http://localhost:8080 \
  --org-id default --student-id STU001 --pin 1234

# Retrieve grades
./euc2 admin get-grades --remote http://localhost:8080 \
  --org-id default --lab-id go101-lab01-stack \
  --bearer-token instructor-secret-token --csv grades.csv
```

---

## Web UI

Launch the embedded graphical workbench:

```bash
./euc2 ui --port 8082
```

Opens a browser at `http://127.0.0.1:8082` with a Monaco code editor, file browser, submit wizard, and admin dashboard.

---

## Project Structure

```
├── TDES/                    # Go monorepo
│   ├── cmd/euc2/            # euc2 CLI (Cobra)
│   ├── cmd/server/          # Registry HTTP server
│   └── internals/           # Core packages
│       ├── drive/           # Offline transport (X25519 + AES-GCM)
│       ├── remote/          # Online transport (HTTP REST client)
│       ├── evaluator-core/  # Docker-sandboxed grading engine
│       ├── exercise/        # Packaging & manifest parsing
│       ├── exercise_store/  # Local cache (~/.euc2)
│       ├── registry/        # SQLite + artifact store
│       └── init/            # Workspace scaffolding
├── frontend/                # React 19 + Vite + Monaco Editor
├── runner-images/           # Docker images for sandboxed execution
├── demo_exercises/          # Example exercises
├── docs/                    # Astro documentation site
└── mock-drive/              # Simulated USB drive for development
```

---

## Documentation

Full documentation is available at **[tdes.bypiyush.com](https://tdes.bypiyush.com)**:

- [Quick Start Guide](https://tdes.bypiyush.com/docs/quick-start)
- [User Flow Overview](https://tdes.bypiyush.com/docs/userflow)
- [Exercise Authoring Spec](https://tdes.bypiyush.com/docs/exercise-authoring-spec)
- [Student CLI Reference](https://tdes.bypiyush.com/docs/student-cli-reference)
- [Admin CLI Reference](https://tdes.bypiyush.com/docs/admin-cli-reference)
- [API Reference](https://tdes.bypiyush.com/docs/api-reference)

