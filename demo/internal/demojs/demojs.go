package demojs

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func RootDir() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("runtime.Caller failed")
	}
	// This file lives in <repo>/demo/internal/demojs/demojs.go
	return filepath.Dir(filepath.Dir(filepath.Dir(thisFile))), nil
}

// Run executes a JavaScript snippet and returns its stdout (trimmed).
// The snippet is prefixed with: `const qs = require('qs');`
func Run(t *testing.T, code string) string {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skipf("node is required for demo tests: %v", err)
	}

	demoDir, err := RootDir()
	if err != nil {
		t.Fatalf("failed to locate demo directory: %v", err)
	}

	fullCode := `const qs = require('qs');` + code

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

		// #nosec G204 -- executing a fixed binary ("node") with fixed args ("-e");
		// the code is test-controlled and run in a sandboxed CI context.
		cmd := exec.CommandContext(ctx, "node", "-e", fullCode)
	cmd.Dir = demoDir
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("JS execution timed out: %v\nOutput: %s", ctx.Err(), string(out))
	}
	if err != nil {
		t.Fatalf("JS execution failed: %v\nOutput: %s", err, string(out))
	}
	return strings.TrimSpace(string(out))
}
