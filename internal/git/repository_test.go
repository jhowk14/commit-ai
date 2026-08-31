package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func testRepository(t *testing.T) Repository {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Commit AI Test")
	runGit(t, dir, "config", "user.email", "commit-ai@example.test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "chore: initial")
	return Repository{Dir: dir}
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

func TestContextCommitAndUndoUseRealGitRepository(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()
	if err := os.WriteFile(filepath.Join(repo.Dir, "README.md"), []byte("base\nchanged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo.Dir, "add", "README.md")
	hasStaged, err := repo.HasStaged(ctx)
	if err != nil || !hasStaged {
		t.Fatalf("has staged=%t err=%v", hasStaged, err)
	}
	commitContext, err := repo.Context(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(commitContext.Files, "README.md") || !strings.Contains(commitContext.Diff, "+changed") {
		t.Fatalf("contexto inesperado: %#v", commitContext)
	}
	if err := repo.Commit(ctx, "feat: add change"); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runGit(t, repo.Dir, "log", "-1", "--pretty=%s")); got != "feat: add change" {
		t.Fatalf("commit: %q", got)
	}
	message, err := repo.Undo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if message != "feat: add change" {
		t.Fatalf("mensagem desfeita: %q", message)
	}
	hasStaged, err = repo.HasStaged(ctx)
	if err != nil || !hasStaged {
		t.Fatalf("alterações não foram preservadas: %t %v", hasStaged, err)
	}
}

func TestCreateOrSwitchBranch(t *testing.T) {
	repo := testRepository(t)
	if err := repo.CreateOrSwitchBranch(context.Background(), "feature/v2"); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runGit(t, repo.Dir, "branch", "--show-current")); got != "feature/v2" {
		t.Fatalf("branch: %q", got)
	}
	if err := repo.CreateOrSwitchBranch(context.Background(), "feature/v2"); err != nil {
		t.Fatal(err)
	}
}

func TestSyncPreservesAndStagesLocalChanges(t *testing.T) {
	repo := testRepository(t)
	remote := t.TempDir()
	runGit(t, remote, "init", "--bare")
	runGit(t, repo.Dir, "remote", "add", "origin", remote)
	runGit(t, repo.Dir, "push", "-u", "origin", "HEAD")
	if err := os.WriteFile(filepath.Join(repo.Dir, "README.md"), []byte("base\nlocal change\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := repo.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	hasStaged, err := repo.HasStaged(context.Background())
	if err != nil || !hasStaged {
		t.Fatalf("mudanças não foram preparadas após sync: %t %v", hasStaged, err)
	}
	if got := strings.TrimSpace(runGit(t, repo.Dir, "diff", "--cached", "--", "README.md")); !strings.Contains(got, "+local change") {
		t.Fatalf("alteração perdida: %s", got)
	}
}
