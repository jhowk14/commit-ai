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

func TestRunPushesAutomaticallyWhenConfigured(t *testing.T) {
	home, repoDir, remote := t.TempDir(), t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.name", "Commit AI Test")
	runGit(t, repoDir, "config", "user.email", "commit-ai@example.test")
	if err := os.WriteFile(filepath.Join(repoDir, "file.txt"), []byte("before\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", "file.txt")
	runGit(t, repoDir, "commit", "-m", "chore: initial")
	runGit(t, remote, "init", "--bare")
	runGit(t, repoDir, "remote", "add", "origin", remote)
	runGit(t, repoDir, "push", "-u", "origin", "HEAD")
	if err := os.WriteFile(filepath.Join(repoDir, "file.txt"), []byte("before\nafter\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", "file.txt")

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"choices":[{"finish_reason":"stop","message":{"content":"fix: push automatically"}}]}`))
	}))
	defer server.Close()
	cfg := config.Default()
	cfg.Provider, cfg.Model, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey, cfg.PushMode = "openai", "model", server.URL, "test", config.PushAlways
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	application := New("test", strings.NewReader(""), &output, &output)
	application.workDir, application.client = repoDir, ai.Client{HTTPClient: server.Client()}
	if err := application.Run(context.Background(), []string{"--yes"}); err != nil {
		t.Fatal(err)
	}
	branch := strings.TrimSpace(runGit(t, repoDir, "branch", "--show-current"))
	if got := strings.TrimSpace(runGit(t, repoDir, "ls-remote", "--heads", "origin", branch)); got == "" {
		t.Fatal("commit não foi enviado automaticamente")
	}
	if !strings.Contains(output.String(), "Alterações enviadas") {
		t.Fatalf("saída de push: %s", output.String())
	}
}

func TestRunSyncReportsEveryStep(t *testing.T) {
	home, repoDir, remote := t.TempDir(), t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.name", "Commit AI Test")
	runGit(t, repoDir, "config", "user.email", "commit-ai@example.test")
	if err := os.WriteFile(filepath.Join(repoDir, "file.txt"), []byte("before\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", "file.txt")
	runGit(t, repoDir, "commit", "-m", "chore: initial")
	runGit(t, remote, "init", "--bare")
	runGit(t, repoDir, "remote", "add", "origin", remote)
	runGit(t, repoDir, "push", "-u", "origin", "HEAD")
	if err := os.WriteFile(filepath.Join(repoDir, "file.txt"), []byte("before\nafter\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"choices":[{"finish_reason":"stop","message":{"content":"fix: report sync progress"}}]}`))
	}))
	defer server.Close()
	cfg := config.Default()
	cfg.Provider, cfg.Model, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey = "openai", "test-model", server.URL, "test"
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	application := New("2.0.0-test", strings.NewReader(""), &output, &output)
	application.workDir, application.client = repoDir, ai.Client{HTTPClient: server.Client()}
	if err := application.Run(context.Background(), []string{"--sync", "--yes"}); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"Auto-sync ativado",
		"Verificando atualizações",
		"Preparando os arquivos",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("saída não contém %q:\n%s", expected, output.String())
		}
	}
}

func TestParseSupportsEveryPublicOption(t *testing.T) {
	opts, err := parse([]string{"-e", "-c", "-C", "-p", "-y", "-s", "-u", "-m", "hint", "-b", "feature/test", "-B", "http://localhost:1234/v1", "--setup", "--config", "--edit-prompt", "--help", "--version"}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.emoji || !opts.conventional || !opts.custom || !opts.preview || !opts.yes || !opts.sync || !opts.undo || !opts.setup || !opts.showConfig || !opts.editPrompt || !opts.help || !opts.version {
		t.Fatalf("opções: %#v", opts)
	}
	if opts.message != "hint" || opts.branch != "feature/test" || opts.baseURL != "http://localhost:1234/v1" {
		t.Fatalf("valores: %#v", opts)
	}
	if _, err := parse([]string{"--unknown"}, &bytes.Buffer{}); err == nil {
		t.Fatal("flag desconhecida deveria falhar")
	}
}

func TestHelpVersionConfigAndCustomPromptErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	var output bytes.Buffer
	application := New("2.0.0-test", strings.NewReader(""), &output, &output)
	if err := application.Run(context.Background(), []string{"--help"}); err != nil {
		t.Fatal(err)
	}
	if err := application.Run(context.Background(), []string{"--version"}); err != nil {
		t.Fatal(err)
	}
	if err := application.Run(context.Background(), []string{"--config"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "commit-ai v2.0.0-test") || !strings.Contains(output.String(), "Uso:") {
		t.Fatalf("saída: %s", output.String())
	}
	if _, err := application.customPrompt(true); err == nil {
		t.Fatal("prompt ausente deveria falhar")
	}
	path, err := config.CustomPromptPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("# comentário\n{FILES}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	prompt, err := application.customPrompt(true)
	if err != nil {
		t.Fatal(err)
	}
	if prompt != "{FILES}" {
		t.Fatalf("prompt: %q", prompt)
	}
}

func TestMessageEditorAcceptsReplacementAndDefault(t *testing.T) {
	var output bytes.Buffer
	application := New("test", strings.NewReader("fix: custom\n"), &output, &output)
	message, err := application.editCommitMessage(config.Default().UILanguage(), "fix: generated")
	if err != nil || message != "fix: custom" {
		t.Fatalf("substituição: %q %v", message, err)
	}
	application.in = strings.NewReader("\n")
	message, err = application.editCommitMessage(config.Default().UILanguage(), "fix: generated")
	if err != nil || message != "fix: generated" {
		t.Fatalf("sugestão padrão: %q %v", message, err)
	}
	if !strings.Contains(output.String(), "Enter mantém a sugestão") {
		t.Fatalf("saída do editor: %s", output.String())
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
