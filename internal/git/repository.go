package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jhowk14/commit-ai/v2/internal/i18n"
)

const maxDiffBytes = 14000

type Repository struct{ Dir string }

type Progress func(string)

func (r Repository) Ensure(ctx context.Context) error {
	output, err := r.run(ctx, "rev-parse", "--is-inside-work-tree")
	if err != nil || strings.TrimSpace(output) != "true" {
		return fmt.Errorf("execute commit-ai dentro de um repositório Git")
	}
	return nil
}

func (r Repository) HasStaged(ctx context.Context) (bool, error) {
	_, err := r.run(ctx, "diff", "--cached", "--quiet")
	if err == nil {
		return false, nil
	}
	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
		return true, nil
	}
	return false, r.commandError("verificar alterações preparadas", err)
}

func (r Repository) Context(ctx context.Context) (Context, error) {
	diff, err := r.run(ctx, "diff", "--cached", "--unified=3")
	if err != nil {
		return Context{}, r.commandError("ler diff", err)
	}
	files, err := r.run(ctx, "diff", "--cached", "--name-only")
	if err != nil {
		return Context{}, r.commandError("listar arquivos", err)
	}
	history := ""
	hasCommit, err := r.hasCommit(ctx)
	if err != nil {
		return Context{}, err
	}
	if hasCommit {
		history, err = r.run(ctx, "log", "--oneline", "-n", "20")
		if err != nil {
			return Context{}, r.commandError("ler histórico", err)
		}
	}
	return Context{Files: strings.TrimSpace(files), Diff: relevantDiff(diff), History: strings.TrimSpace(history)}, nil
}

func (r Repository) hasCommit(ctx context.Context) (bool, error) {
	_, err := r.run(ctx, "rev-parse", "--verify", "--quiet", "HEAD")
	if err == nil {
		return true, nil
	}
	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
		return false, nil
	}
	return false, r.commandError("verificar histórico", err)
}

type Context struct{ Files, Diff, History string }

func relevantDiff(diff string) string {
	var builder strings.Builder
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "@@") || strings.HasPrefix(line, "diff --git") {
			if builder.Len()+len(line)+1 > maxDiffBytes {
				break
			}
			builder.WriteString(line)
			builder.WriteByte('\n')
		}
	}
	return strings.TrimSpace(builder.String())
}

func (r Repository) CreateOrSwitchBranch(ctx context.Context, branch string) error {
	if branch == "" {
		return nil
	}
	var err error
	if _, lookupErr := r.run(ctx, "show-ref", "--verify", "--quiet", "refs/heads/"+branch); lookupErr == nil {
		_, err = r.run(ctx, "checkout", branch)
	} else {
		_, err = r.run(ctx, "checkout", "-b", branch)
	}
	return r.commandError("criar ou trocar a branch", err)
}

func (r Repository) Sync(ctx context.Context, language i18n.Language, progress Progress) (resultErr error) {
	report := func(message string) {
		if progress != nil {
			progress(message)
		}
	}
	report(i18n.T(language, "sync_stage"))
	if _, err := r.run(ctx, "add", "-A"); err != nil {
		return r.commandError("adicionar alterações", err)
	}
	status, err := r.run(ctx, "status", "--porcelain")
	if err != nil {
		return r.commandError("verificar alterações", err)
	}
	stashed := strings.TrimSpace(status) != ""
	needsRestore := false
	restoreStash := func() error {
		report(i18n.T(language, "sync_restore"))
		if _, err := r.run(context.WithoutCancel(ctx), "stash", "pop", "--index"); err != nil {
			return r.commandError("restaurar alterações", err)
		}
		needsRestore = false
		return nil
	}
	if stashed {
		report(i18n.T(language, "sync_stash"))
		if _, err := r.run(ctx, "stash", "push", "-u", "-m", "commit-ai-auto-stash"); err != nil {
			return r.commandError("guardar alterações temporariamente", err)
		}
		needsRestore = true
		// A synchronization failure or Ctrl+C must never leave a hidden stash
		// behind. WithoutCancel lets this cleanup finish after the user cancels.
		defer func() {
			if !needsRestore {
				return
			}
			if restoreErr := restoreStash(); restoreErr != nil {
				if resultErr == nil {
					resultErr = restoreErr
				} else {
					resultErr = fmt.Errorf("%w; %v", resultErr, restoreErr)
				}
			}
		}()
	}
	branch, err := r.run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return r.commandError("identificar branch atual", err)
	}
	branch = strings.TrimSpace(branch)
	report(i18n.T(language, "sync_pull", branch))
	if _, err := r.run(ctx, "pull", "origin", branch, "--rebase"); err != nil {
		if _, fallbackErr := r.run(ctx, "pull", "origin", branch); fallbackErr != nil {
			return fmt.Errorf("não foi possível sincronizar com origin/%s: %w", branch, fallbackErr)
		}
	}
	if stashed {
		// The explicit restore preserves the familiar progress order on success.
		// The deferred path above is reserved for errors and cancellations.
		if err := restoreStash(); err != nil {
			return err
		}
	}
	report(i18n.T(language, "sync_prepare"))
	_, err = r.run(ctx, "add", "-A")
	return r.commandError("adicionar alterações após sincronização", err)
}

func (r Repository) Commit(ctx context.Context, message string) error {
	_, err := r.run(ctx, "commit", "-m", message)
	return r.commandError("criar commit", err)
}

func (r Repository) Undo(ctx context.Context) (string, error) {
	message, err := r.run(ctx, "log", "-1", "--pretty=%B")
	if err != nil {
		return "", r.commandError("ler o último commit", err)
	}
	if _, err := r.run(ctx, "reset", "--soft", "HEAD~1"); err != nil {
		return "", r.commandError("desfazer o último commit", err)
	}
	return strings.TrimSpace(message), nil
}

func (r Repository) Push(ctx context.Context, branch string) error {
	if branch == "" {
		_, err := r.run(ctx, "push")
		return r.commandError("enviar alterações", err)
	}
	if _, err := r.run(ctx, "push", "-u", "origin", branch); err == nil {
		return nil
	}
	_, err := r.run(ctx, "push", "origin", branch)
	return r.commandError("enviar alterações", err)
}

func (r Repository) run(ctx context.Context, args ...string) (string, error) {
	command := exec.CommandContext(ctx, "git", args...)
	command.Dir = r.Dir
	output, err := command.CombinedOutput()
	return string(output), err
}

func (r Repository) commandError(action string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("não foi possível %s: %w", action, err)
}
