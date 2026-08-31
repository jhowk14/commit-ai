package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhowk14/commit-ai/v2/internal/ai"
	"github.com/jhowk14/commit-ai/v2/internal/config"
)

func TestRunCreatesAndUndoesRealCommit(t *testing.T) {
	home, repoDir := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.name", "Commit AI Test")
	runGit(t, repoDir, "config", "user.email", "commit-ai@example.test")
	if err := os.WriteFile(filepath.Join(repoDir, "file.txt"), []byte("before\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", "file.txt")
	runGit(t, repoDir, "commit", "-m", "chore: initial")
	if err := os.WriteFile(filepath.Join(repoDir, "file.txt"), []byte("before\nafter\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", "file.txt")

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/chat/completions" {
			t.Fatalf("rota: %s", request.URL.Path)
		}
		_, _ = writer.Write([]byte(`{"choices":[{"finish_reason":"stop","message":{"content":"feat: create v2 commit"}}]}`))
	}))
	defer server.Close()
	cfg := config.Default()
	cfg.Provider, cfg.Model, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey = "openai", "test-model", server.URL, "test"
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	application := New("2.0.0-test", strings.NewReader(""), &output, &output)
	application.workDir = repoDir
	application.client = ai.Client{HTTPClient: server.Client()}
	if err := application.Run(context.Background(), []string{"-y"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runGit(t, repoDir, "log", "-1", "--pretty=%s")); got != "feat: create v2 commit" {
		t.Fatalf("commit: %q", got)
	}
	if err := application.Run(context.Background(), []string{"--undo"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runGit(t, repoDir, "log", "-1", "--pretty=%s")); got != "chore: initial" {
		t.Fatalf("undo: %q", got)
	}
	if !strings.Contains(output.String(), "Commit criado") || !strings.Contains(output.String(), "desfeito") {
		t.Fatalf("saída: %s", output.String())
	}
}

func TestPreviewDoesNotCommit(t *testing.T) {
	home, repoDir := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.name", "Commit AI Test")
	runGit(t, repoDir, "config", "user.email", "commit-ai@example.test")
	if err := os.WriteFile(filepath.Join(repoDir, "file.txt"), []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", "file.txt")
	runGit(t, repoDir, "commit", "-m", "chore: initial")
	if err := os.WriteFile(filepath.Join(repoDir, "file.txt"), []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", "file.txt")
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"choices":[{"finish_reason":"stop","message":{"content":"fix: preview change"}}]}`))
	}))
	defer server.Close()
	cfg := config.Default()
	cfg.Provider, cfg.Model, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey = "openai", "model", server.URL, "test"
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	application := New("2.0.0-test", strings.NewReader(""), &output, &output)
	application.workDir, application.client = repoDir, ai.Client{HTTPClient: server.Client()}
	if err := application.Run(context.Background(), []string{"--preview"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runGit(t, repoDir, "log", "-1", "--pretty=%s")); got != "chore: initial" {
		t.Fatalf("preview criou commit: %q", got)
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return string(output)
}
