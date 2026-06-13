# TDES Feature Development & Verification Tracker

This document serves as a tracking sheet for the Test Delivery and Evaluation System (TDES). It registers the implementation status and verification status of all components.

---

## 1. CLI Commands (`euc2`)

### Initialize Workspace (`euc2 init`)
* **Description**: Unpacks public boilerplate templates from the local cache into a target directory.
* **Status**:
  - [x] Implemented
  - [x] Manually Verified

### Run Public Tests (`euc2 run`)
* **Description**: Mounts the workspace in Docker and executes public Makefile targets inside the sandbox.
* **Status**:
  - [x] Implemented
  - [x] Manually Verified

### Run Local Public Tests (`euc2 run --local`)
* **Description**: Runs public tests directly on the host system without spawning Docker.
* **Status**:
  - [x] Implemented
  - [x] Manually Verified

### Package Exercise (`euc2 package`)
* **Description**: Tests the author's reference solution, splits private/public directories using globs, and saves them locally.
* **Status**:
  - [x] Implemented
  - [x] Manually Verified

### Publish Exercise (`euc2 publish`)
* **Description**: Packages the exercise (running verification tests) and uploads it to the remote registry server.
* **Status**:
  - [x] Implemented
  - [x] Manually Verified

### Fetch Exercise (`euc2 fetch`)
* **Description**: Pulls public archives from local drive volumes or remote REST server caches.
* **Status**:
  - [x] Implemented
  - [x] Manually Verified

### Submit Solution (`euc2 submit`)
* **Description**: Creates a submission archive, attaches student metadata, and pushes it to a local drive folder (encrypted) or remote server.
* **Status**:
  - [x] Implemented
  - [x] Manually Verified

### Evaluate Submission (`euc2 evaluate`)
* **Description**: Grades a student's submission tar archive in a local Docker sandbox using private tests. If the private tests package is not found locally, it automatically pulls it from the configured registry server.
* **Status**:
  - [x] Implemented
  - [x] Manually Verified
  - [x] Registry Pull Integration
    * **Note**: Currently uses a simplified/dummy Bearer Token authorization header in transit, which must be replaced with proper role-based authentication later.

### Drive Setup Utilities (`euc2 drive prepare/prepare-submission`)
* **Description**: Prepares folders with submission envelopes and configures X25519 recipient keys.
* **Status**:
  - [x] Implemented
  - [x] Manually Verified

---

## 2. Server REST Endpoints (`cmd/server`)

### Health Check (`GET /healthz`)
* **Description**: Simple ping to check if the server is healthy.
* **Status**:
  - [x] Implemented
  - [ ] Manually Verified

### Publish Exercise Version (`POST /v1/exercises/publish`)
* **Description**: Accepts multipart form uploads for public/private exercise tarballs and persists them.
* **Status**:
  - [x] Implemented
  - [x] Manually Verified

### Get Version Metadata (`GET /v1/exercises/{orgID}/{exerciseID}/versions/{version}`)
* **Description**: Retrieves exercise details, title, languages, and SHA-256 hashes of artifacts.
* **Status**:
  - [x] Implemented
  - [ ] Manually Verified

### Download Raw Artifact (`GET /v1/artifacts/{sha256}`)
* **Description**: Downloads raw binary packages directly by their SHA-256 content hash.
* **Status**:
  - [x] Implemented
  - [ ] Manually Verified

### Download Versioned Artifact (`GET /v1/exercises/{orgID}/{exerciseID}/versions/{version}/download`)
* **Description**: Resolves and streams either the `public` or `private` package for an exercise.
* **Status**:
  - [x] Implemented
  - [ ] Manually Verified

### List Exercises (`GET /v1/exercises`)
* **Description**: Lists all registered exercises, optionally filtered by organization ID or status.
* **Status**:
  - [x] Implemented
  - [ ] Manually Verified

### List Exercise Versions (`GET /v1/exercises/{orgID}/{exerciseID}/versions`)
* **Description**: Lists all version releases for a single exercise.
* **Status**:
  - [x] Implemented
  - [ ] Manually Verified

### Update Version Status (`POST /v1/exercises/{orgID}/{exerciseID}/versions/{version}/status`)
* **Description**: Updates release statuses (e.g. `draft`, `published`, `retired`).
* **Status**:
  - [x] Implemented
  - [ ] Manually Verified

### Delete Version (`DELETE /v1/exercises/{orgID}/{exerciseID}/versions/{version}`)
* **Description**: Deletes metadata records for a specific exercise version.
* **Status**:
  - [x] Implemented
  - [ ] Manually Verified

---

## 3. Cryptography & Security

### Key Agreement (ECDH)
* **Description**: Uses X25519 key agreements to establish secrets dynamically between students and recipients.
* **Status**:
  - [x] Implemented
  - [ ] Manually Verified

### Symmetric Encryption (AES-GCM)
* **Description**: Packages are encrypted/decrypted securely using AES-256-GCM.
* **Status**:
  - [x] Implemented
  - [ ] Manually Verified

---

## 4. Frontend Web Documentation (`docs/`)

### Theme Switcher
* **Description**: Supports Light and Dark mode UI matching modern aesthetics.
* **Status**:
  - [x] Implemented
  - [ ] Manually Verified

### Terminal Simulator
* **Description**: Simulates typing animations and visual responses for running and evaluating exercises.
* **Status**:
  - [x] Implemented
  - [ ] Manually Verified

### CLI Command tab references
* **Description**: Interactive UI component displaying syntax, details, and example runs for all `euc2` commands.
* **Status**:
  - [x] Implemented
  - [ ] Manually Verified

---

## 5. System Backlog (To Be Implemented)

### Scaffold Exercise templates (`euc2 new-exercise`)
* **Description**: Command to generate template boilerplate structures for exercise setters.
* **Status**:
  - [ ] Implemented
  - [ ] Manually Verified
  - [ ] **Priority**: Low (Deprioritized)

### Remote API Submission Evaluator
* **Description**: REST API endpoint on the server to submit and grade student solutions directly over HTTP.
* **Status**:
  - [x] Implemented
  - [x] Manually Verified
