.PHONY: test-public \
        test-submission-lifo \
        test-submission-interleaved \
        test-submission-size \
        test-submission-pop-empty \
        test-submission-peek \
        clean

# Run the public test suite against the student's stack.go.
test-public:
	@go test -v -count=1 .

# ---------------------------------------------------------------------------
# Grading targets — one per private test, each worth 2 points (total: 10).
#
# Each target is self-contained:
#   1. Back up the live stack.go
#   2. Inject reference/stack.go into the live slot
#   3. Inject the private test file
#   4. Run ONLY the named test
#   5. Restore originals (whether the test passed or failed)
# ---------------------------------------------------------------------------

# Helper macro: inject, run one test by name, restore.
# Usage:  $(call _grade,TestFuncName)
define _grade
	@cp stack.go stack.go.bak
	@cp reference/stack.go stack.go
	@cp tests_private/stack_private_test.go stack_private_test.go
	@go test -v -count=1 -run ^$(1)$$ .; EXIT=$$?; \
	  cp stack.go.bak stack.go; \
	  rm -f stack.go.bak stack_private_test.go; \
	  exit $$EXIT
endef

# Private 1 — LIFO ordering across 5 elements  (2 pts)
test-submission-lifo:
	$(call _grade,TestPrivate_LIFOOrder)

# Private 2 — Interleaved Push / Pop / Peek    (2 pts)
test-submission-interleaved:
	$(call _grade,TestPrivate_InterleavedOps)

# Private 3 — Size tracking under load         (2 pts)
test-submission-size:
	$(call _grade,TestPrivate_SizeUnderLoad)

# Private 4 — Pop past empty does not corrupt  (2 pts)
test-submission-pop-empty:
	$(call _grade,TestPrivate_PopUntilEmpty)

# Private 5 — Repeated Peek does not mutate    (2 pts)
test-submission-peek:
	$(call _grade,TestPrivate_PeekPreservesOrder)

clean:
	@go clean ./...
	@rm -f stack_private_test.go stack.go.bak
