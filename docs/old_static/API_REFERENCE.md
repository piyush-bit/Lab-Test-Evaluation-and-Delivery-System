# TDES Registry Server REST API Reference

The TDES Registry Server is implemented using Go's standard library `net/http` router (`http.NewServeMux()`). It does not natively generate Swagger UI or OpenAPI specifications. This document serves as the manual API reference.

---

## Global API Design & Security

### Base URL
```
http://<host>:<port>
```
*Default port: `8080` (or configured via the `PORT` environment variable).*

### Authentication Schemes
The server enforces authentication on specific endpoints using an HTTP `Authorization` header containing a static Bearer Token:

| Role | Auth Header | Verification Source |
| :--- | :--- | :--- |
| **Instructor / Admin** | `Authorization: Bearer <instructor_token>` | Checked against the server's `EUC2_INSTRUCTOR_TOKENS` environment variable (comma-separated list). |
| **Student / Candidate** | Attached in request payload / headers | Verified via a PIN-Based Trust-On-First-Use (TOFU) scheme checked against `student_credentials` database table. |
| **Public / Anonymous** | None | Public read-only endpoints (e.g., health check, downloading public exercise template packages). |

---

## Endpoint Summary Table

| Method | Path | Authentication | Description |
| :--- | :--- | :--- | :--- |
| **GET** | `/healthz` | None (Public) | Service health check status. |
| **POST** | `/v1/admin/onboard` | Instructor | Upload a CSV file containing authorized student IDs. |
| **POST** | `/v1/exercises/publish` | Instructor | Publish a new exercise version with public & private archives. |
| **GET** | `/v1/exercises` | None (Public) | List all exercises, optionally filtered. |
| **GET** | `/v1/exercises/{orgID}/{exerciseID}/versions` | None (Public) | List all version releases of a specific exercise. |
| **GET** | `/v1/exercises/{orgID}/{exerciseID}/versions/{version}` | None (Public) | Get metadata for a specific exercise version. |
| **GET** | `/v1/exercises/{orgID}/{exerciseID}/versions/{version}/download` | None (Public)* | Download either the `public` or `private` package archive. |
| **POST** | `/v1/exercises/{orgID}/{exerciseID}/versions/{version}/status` | None (Public) | Update an exercise version's status (e.g. draft, retired). |
| **DELETE**| `/v1/exercises/{orgID}/{exerciseID}/versions/{version}` | None (Public) | Delete a specific exercise version. |
| **GET** | `/v1/artifacts/{sha256}` | None (Public) | Download a raw package archive directly by its SHA-256 hash. |
| **POST** | `/v1/submissions` | Student (TOFU/PIN) | Submit student code for containerized sandbox evaluation. |
| **GET** | `/v1/submissions` | Instructor | Retrieve/export student grades (supports table/CSV/JSON). |

---

## Endpoint Details

### 1. Health Check
Checks if the registry server is running.

* **Route**: `GET /healthz`
* **Authentication**: None
* **Success Response**: `200 OK`
  ```json
  {
    "status": "ok"
  }
  ```

---

### 2. Onboard Students (Roster Upload)
Uploads student IDs to the course roster. Roster entries start with an empty PIN. Students activate their IDs on first submission.

* **Route**: `POST /v1/admin/onboard`
* **Authentication**: Instructor Bearer Token required.
* **Content-Type**: `multipart/form-data`
* **Form Fields**:
  * `roster_csv` (file): **(Required)** CSV file with headers containing `student_id` (or `studentid`) and optional `org_id` (or `org`).
* **Responses**:
  * `200 OK` (Onboarding successful)
    ```json
    {
      "status": "ok",
      "onboarded": 3
    }
    ```
  * `401 Unauthorized` (Invalid or missing Bearer token)
  * `400 Bad Request` (Missing headers, parse error, or empty file)

---

### 3. Publish Exercise Version
Registers metadata and uploads the public (boilerplate) and private (tests) package tarballs for a lab.

* **Route**: `POST /v1/exercises/publish`
* **Authentication**: Instructor Bearer Token required.
* **Content-Type**: `multipart/form-data`
* **Form Fields**:
  * `org_id` (string): **(Required)** Organization identifier.
  * `exercise_id` (string): **(Required)** Exercise identifier.
  * `version` (string): **(Required)** SemVer version string (e.g. `1.1.0`).
  * `status` (string): Initial status (`draft`, `published`, etc.).
  * `public_artifact` (file): **(Required)** The student's template `.tar` archive.
  * `private_artifact` (file): **(Required)** The evaluation/private test `.tar` archive.
* **Responses**:
  * `201 Created` (Exercise version successfully registered)
    ```json
    {
      "org_id": "org1",
      "exercise_id": "lab1",
      "version": "1.1.0",
      "status": "published",
      "public_artifact_sha256": "8a32bfa8f...",
      "private_artifact_sha256": "7a35eab2e...",
      "created_at": "2026-06-21T12:00:00Z"
    }
    ```
  * `409 Conflict` (Version already exists)
  * `401 Unauthorized` (Invalid token)
  * `400 Bad Request` (Missing fields or file errors)

---

### 4. List Exercises
Retrieves metadata of registered exercises.

* **Route**: `GET /v1/exercises`
* **Query Parameters**:
  * `org_id` (string, optional): Filter by organization.
  * `status` (string, optional): Filter by status (e.g. `published`).
* **Success Response**: `200 OK`
  ```json
  [
    {
      "org_id": "org1",
      "exercise_id": "lab1",
      "latest_version": "1.1.0",
      "title": "Stack Exercise"
    }
  ]
  ```

---

### 5. List Exercise Versions
Lists all version releases registered for a single exercise.

* **Route**: `GET /v1/exercises/{orgID}/{exerciseID}/versions`
* **Success Response**: `200 OK`
  ```json
  [
    {
      "org_id": "org1",
      "exercise_id": "lab1",
      "version": "1.0.0",
      "status": "published",
      "public_artifact_sha256": "4b32a...",
      "private_artifact_sha256": "5c92d..."
    },
    {
      "org_id": "org1",
      "exercise_id": "lab1",
      "version": "1.1.0",
      "status": "published",
      "public_artifact_sha256": "8a32b...",
      "private_artifact_sha256": "7a35e..."
    }
  ]
  ```

---

### 6. Get Version Metadata
Gets complete details for a specific exercise version.

* **Route**: `GET /v1/exercises/{orgID}/{exerciseID}/versions/{version}`
* **Responses**:
  * `200 OK`
    ```json
    {
      "org_id": "org1",
      "exercise_id": "lab1",
      "version": "1.1.0",
      "status": "published",
      "public_artifact_sha256": "8a32bfa...",
      "private_artifact_sha256": "7a35eab...",
      "created_at": "2026-06-21T12:00:00Z"
    }
    ```
  * `404 Not Found` (Exercise or version not found)

---

### 7. Download Versioned Artifact
Streams the public or private archive associated with an exercise version.

* **Route**: `GET /v1/exercises/{orgID}/{exerciseID}/versions/{version}/download`
* **Query Parameters**:
  * `type` (string, optional): Download type. Can be `public` (default) or `private`.
* **Headers Returned**:
  * `Content-Type: application/octet-stream`
  * `X-Artifact-SHA256`: SHA-256 checksum of the file
* **Responses**:
  * `200 OK` (Streams binary archive)
  * `404 Not Found` (Version or file not found)

---

### 8. Update Version Status
Updates the lifecycle status of an exercise version.

* **Route**: `POST /v1/exercises/{orgID}/{exerciseID}/versions/{version}/status`
* **Request Body** (or Form Value or Query Parameter):
  ```json
  {
    "status": "retired"
  }
  ```
* **Success Response**: `200 OK`
  ```json
  {
    "status": "ok"
  }
  ```

---

### 9. Delete Version
Deletes metadata records for a specific exercise version.

* **Route**: `DELETE /v1/exercises/{orgID}/{exerciseID}/versions/{version}`
* **Success Response**: `200 OK`
  ```json
  {
    "status": "ok"
  }
  ```

---

### 10. Download Raw Artifact by SHA-256
Streams an archive file directly using its unique SHA-256 hash.

* **Route**: `GET /v1/artifacts/{sha256}`
* **Success Response**: `200 OK` (Streams binary archive)
* **Headers Returned**:
  * `Content-Type: application/octet-stream`
  * `Content-Length`: Size in bytes
* **Fail Response**: `404 Not Found`

---

### 11. Submit Student Solution (Remote Evaluator)
Uploads student solution code for grading. The server spins up a Docker container on-the-fly, grades it, persists metadata in the database, and returns the result receipt.

* **Route**: `POST /v1/submissions`
* **Authentication**: Student TOFU/PIN authentication (verified against course roster).
* **Content-Type**: `multipart/form-data`
* **Form Fields**:
  * `submission_package` (file): **(Required)** A `.tar` package containing student workspace files including `submission-manifest.json` at its root.
  * `pin` (string): **(Required)** The student's private authentication PIN.
  * `new_pin` (string, optional): A new PIN value if rotating credentials.
* **Responses**:
  * `200 OK` (Grading successfully executed and saved)
    ```json
    {
      "org_id": "org1",
      "student_id": "student1",
      "lab_id": "lab1",
      "version": "v1.0.0",
      "status": "success",
      "earned_points": 10,
      "max_points": 10,
      "results": [
        {
          "command": "make test-submission-lifo",
          "status": "passed",
          "points_earned": 5,
          "points_possible": 5
        },
        {
          "command": "make test-submission-fifo",
          "status": "passed",
          "points_earned": 5,
          "points_possible": 5
        }
      ]
    }
    ```
  * `403 Forbidden` (Student ID is not pre-registered in the roster)
  * `401 Unauthorized` (PIN verification failed)
  * `400 Bad Request` (PIN length < 4, missing fields, or manifest package parsing error)
  * `500 Internal Server Error` (Docker sandbox sandbox failure or database write failure)

---

### 12. Retrieve Submissions (Grades Export)
Retrieves or exports all student grades from the registry database.

* **Route**: `GET /v1/submissions`
* **Authentication**: Instructor Bearer Token required.
* **Query Parameters**:
  * `org_id` (string, optional): Filter records by organization.
  * `lab_id` (string, optional): Filter records by lab ID.
  * `format` (string, optional): Export format. If set to `csv`, returns a CSV download. Otherwise, returns a JSON array.
* **Responses**:
  * `200 OK` (JSON response format)
    ```json
    [
      {
        "id": "sub1",
        "org_id": "org1",
        "student_id": "student1",
        "lab_id": "lab1",
        "version": "v1.0.0",
        "status": "success",
        "earned_points": 10,
        "max_points": 10,
        "created_at": "2026-06-21T12:00:00Z"
      }
    ]
    ```
  * `200 OK` (CSV response format if `format=csv`)
    * **Headers**: 
      * `Content-Type: text/csv`
      * `Content-Disposition: attachment; filename="submissions.csv"`
    * **Payload**:
      ```csv
      id,org_id,student_id,lab_id,version,status,earned_points,max_points,created_at
      sub1,org1,student1,lab1,v1.0.0,success,10,10,2026-06-21T12:00:00Z
      ```
  * `401 Unauthorized` (Invalid instructor token)
  * `500 Internal Server Error` (Database read error)
