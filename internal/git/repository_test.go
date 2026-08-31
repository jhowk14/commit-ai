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

func TestEnsureAndEmptyRepositoryStates(t *testing.T) {
	ctx := context.Background()
	if err := (Repository{Dir: t.TempDir()}).Ensure(ctx); err == nil {
		t.Fatal("diretório sem Git deveria falhar")
	}
	emptyDir := t.TempDir()
	runGit(t, emptyDir, "init")
	if err := os.WriteFile(filepath.Join(emptyDir, "first.txt"), []byte("first change\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, emptyDir, "add", "first.txt")
	firstContext, err := (Repository{Dir: emptyDir}).Context(ctx)
	if err != nil || firstContext.History != "" || !strings.Contains(firstContext.Diff, "+first change") {
		t.Fatalf("contexto do primeiro commit=%#v err=%v", firstContext, err)
	}
	repo := testRepository(t)
	if err := repo.Ensure(ctx); err != nil {
		t.Fatal(err)
	}
	hasStaged, err := repo.HasStaged(ctx)
	if err != nil || hasStaged {
		t.Fatalf("staged=%t err=%v", hasStaged, err)
	}
	if _, err := repo.Undo(ctx); err == nil {
		t.Fatal("undo do commit inicial deveria falhar")
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

func TestPushSendsNewBranchToBareRemote(t *testing.T) {
	repo := testRepository(t)
	remote := t.TempDir()
	runGit(t, remote, "init", "--bare")
	runGit(t, repo.Dir, "remote", "add", "origin", remote)
	if err := repo.CreateOrSwitchBranch(context.Background(), "feature/push"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo.Dir, "README.md"), []byte("base\npushed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo.Dir, "add", "README.md")
	if err := repo.Commit(context.Background(), "feat: push branch"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Push(context.Background(), "feature/push"); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runGit(t, repo.Dir, "ls-remote", "--heads", "origin", "feature/push")); got == "" {
		t.Fatal("branch não foi enviada")
	}
	if err := repo.Push(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
}

func TestRelevantDiffFiltersAndLimitsOutput(t *testing.T) {
	input := "context\ndiff --git a/a b/a\n@@ -1 +1 @@\n-old\n+new\nignored"
	got := relevantDiff(input)
	if strings.Contains(got, "context") || !strings.Contains(got, "+new") {
		t.Fatalf("filtro: %q", got)
	}
	long := "+" + strings.Repeat("x", maxDiffBytes+100)
	if got := relevantDiff(long); len(got) > maxDiffBytes {
		t.Fatalf("limite: %d", len(got))
	}
}
