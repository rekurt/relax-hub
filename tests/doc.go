// Package tests contains integration tests against a real Postgres + Redis stack.
// They are excluded from `go test ./...` runs that lack a database (no _test.go
// build constraints required — the suite is opt-in via TEST_DB env var).
//
// This file gives the package a buildable unit so `go build ./...` does not
// error with "no non-test Go files in tests".
package tests
