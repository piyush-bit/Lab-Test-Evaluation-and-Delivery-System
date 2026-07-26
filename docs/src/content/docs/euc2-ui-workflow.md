---
title: "euc2 Web UI & Workflows"
description: "Walkthrough of the student workspace editor and admin drive console inside the euc2 local Web UI"
section: "get-started"
order: 3
---

# euc2 Web UI & Workflows

TDES provides a local graphical Web Console interface alongside its CLI commands. This interface can be launched via:

```bash
euc2 ui
```

This starts an HTTP server (defaulting to `http://127.0.0.1:8082`) and opens the Web Console in your default browser. The Web Console provides custom dashboards tailored for both **students** (for writing code, running tests, and submitting solutions) and **instructors/administrators** (for key pair setup, drive preparation, and offline batch evaluation).

---

## Student Workflow

Students use the TDES Web Console to manage their workstation environment, solve tasks in the browser, and submit solutions without relying on command-line flags.

### 1. Student Dashboard
Upon opening `euc2 ui`, students can browse remote registry libraries, scan for active exercise caches, and initialize their workspace directory with a single click.

![Student Dashboard](/images/ui/student-dashboard.png)

### 2. Monaco Workspace Editor
The Web Console integrates a full-featured Monaco text editor, allowing students to browse code skeletons, edit source files, and run public test suites inside a Docker container with real-time logging output.

![Monaco Code Workspace & Test Panel](/images/ui/student-monaco.png)

---

## Instructor / Admin Workflow

Administrators and exercise authors use the Web Console to onboard students, prepare offline drive structures, and grade candidate submissions in bulk.

### 1. Admin Dashboard
Instructors can generate X25519 public/private key pairs directly from the web console to encrypt offline student packages and onboard class roster CSV files.

![Admin Dashboard & Keys Manager](/images/ui/admin-dashboard.png)

### 2. Drive Management & Batch Evaluation
Instructors can scan external storage devices, click "Prepare Drive" to configure key folders, view collected student envelopes, and run batch evaluation sequences showing grades and output logs on the fly.

![Admin Drive Management & Batch Evaluation](/images/ui/admin-drive.png)
