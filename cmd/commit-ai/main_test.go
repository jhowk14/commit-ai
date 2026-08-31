package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuiltBinaryReportsHelpAndVersion(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "commit-ai")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build do binário: %v\n%s", err, output)
	}

	versionOutput, err := exec.Command(binary, "--version").CombinedOutput()
	if err != nil || strings.TrimSpace(string(versionOutput)) != "commit-ai v"+version {
		t.Fatalf("versão: %v / %q", err, versionOutput)
	}
	helpOutput, err := exec.Command(binary, "--help").CombinedOutput()
	if err != nil || !strings.Contains(string(helpOutput), "--base-url") || !strings.Contains(string(helpOutput), "--setup") {
		t.Fatalf("ajuda: %v / %q", err, helpOutput)
	}
}
