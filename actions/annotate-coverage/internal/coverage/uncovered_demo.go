package coverage

// UncoveredDemo exists only as a fixture for the test-annotate-coverage
// workflow. That workflow runs this action against this module's own coverage
// profile and asserts that the body below is reported as uncovered, which
// exercises the whole pipeline end to end: real `go test` coverage -> diff
// intersection -> `::notice` output. It is intentionally never called from a
// test. Please do not add a test that covers it (and do not delete it) without
// updating .github/workflows/test-annotate-coverage.yaml, or the end-to-end
// assertion loses its signal.
func UncoveredDemo(n int) int {
	if n < 0 {
		return -n
	}
	return n * 2
}
