---
title: "PIN Authentication Design"
description: "Authentication and integrity verification for offline and online delivery"
section: "get-started"
order: 2
---

# Design & Implementation Plan: PIN-Based TOFU Authentication

This document outlines the design and future implementation plan for a **PIN-Based Trust-On-First-Use (TOFU)** authentication system in TDES. It ensures student identity verification while remaining completely friction-free for legitimate candidates.

---

## 1. Core Architecture

The authentication model relies on a **Trust-On-First-Use (TOFU)** scheme. 
* On their **first submission** for a given `(org_id, student_id)`, the candidate registers a private PIN.
* On **subsequent submissions**, the candidate must supply this PIN.
* The system securely hashes and stores the PIN on the server.
* The CLI caches the PIN locally on the student’s workstation so they do not need to re-enter it for every submission.

---

## 2. Server-Side Changes

### Database Schema
A new SQLite table `student_credentials` will be added to the registry database:

```sql
CREATE TABLE IF NOT EXISTS student_credentials (
    org_id TEXT NOT NULL,
    student_id TEXT NOT NULL,
    pin_hash TEXT NOT NULL,      -- Salted bcrypt or PBKDF2 hash of the PIN
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (org_id, student_id)
);
```

### API Endpoint Modifications (`POST /v1/submissions`)
The submission handler will accept two additional fields in the multipart form:
1. `pin` (string): The current PIN of the student.
2. `new_pin` (string, optional): A new PIN if the student wishes to update their credential.

#### Verification Logic:
1. **Retrieve Credentials**: Look up the `(org_id, student_id)` in the `student_credentials` table.
2. **If Record Does Not Exist (First Time / Onboarding)**:
   * Validate that the provided `pin` meets basic criteria (e.g., minimum 4 digits).
   * Hash the `pin` using a secure hashing function.
   * Write the new record containing `(org_id, student_id, pin_hash)` to the database.
   * Proceed with evaluation.
3. **If Record Exists**:
   * Compare the hash of the provided `pin` with `pin_hash` stored in the database.
   * **If mismatch**: Abort evaluation and return `401 Unauthorized` (Invalid PIN).
   * **If match**:
     * If `new_pin` is supplied: Hash the `new_pin`, update the record in the database, and save.
     * Proceed with evaluation.

### Security Enhancements
* **PIN Hashing**: Use bcrypt (with a cost factor of at least 10) or Argon2id to prevent offline brute-force attacks if the database is leaked.
* **Rate Limiting**: To prevent online brute-force guessing of short PINs, implement a rate limiter on the submission route (e.g., max 5 failed attempts per `student_id` within a 15-minute window).

---

## 3. CLI Changes (`euc2`)

### Workstation Configuration (`~/.euc2/config.json`)
The CLI will cache the student's identity details locally.
```json
{
  "student_id": "student-1234",
  "org_id": "org-acme",
  "pin": "123456"
}
```

### CLI Submit Flow (`euc2 submit`)
1. **Check Local Config**:
   * Look for a stored PIN in `~/.euc2/config.json`.
   * **If missing**: Prompt the student:
     ```text
     No PIN found on this workstation.
     Please enter a new PIN (minimum 4 digits) to secure your student ID: ______
     ```
     Save the entered PIN to the local config.
2. **Embed in Payload**:
   * Attach the `pin` from the config file to the multipart form upload when sending the remote request.
3. **Handle Response**:
   * If the server returns `401 Unauthorized` due to a PIN mismatch, alert the student that their PIN is incorrect (indicating they may have typed it incorrectly or another student has already claimed that ID).

### CLI Pin Update Command
Provide a command to update the cached and server-side PIN:
```bash
euc2 set-pin --old <old-pin> --new <new-pin>
```
This will send an update request to the server containing both values, and update the local configuration on success.
