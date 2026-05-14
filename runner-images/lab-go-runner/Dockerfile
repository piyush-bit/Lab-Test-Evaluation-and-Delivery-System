# lab-go-runner
# Lightweight Go runner image for lab exercise evaluation.
# Based on Alpine to minimize size (~235MB vs ~800MB for Debian-based golang image).
# Adds make so Makefile entrypoints (test-public, test-submission) work correctly.

FROM golang:1.22-alpine
RUN apk add --no-cache make
