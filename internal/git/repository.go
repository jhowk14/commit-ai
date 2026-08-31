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
	out, err := r.run(ctx, "diff", "--cached", "--quiet")
	if err == nil {
		return false, nil
	}
	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
		return true, nil
	}
	return false, r.commandError("verificar alterações preparadas", out, err)
}

func (r Repository) Context(ctx context.Context) (Context, error) {
	diff, err := r.run(ctx, "diff", "--cached", "--unified=3")
	if err != nil {
		return Context{}, r.commandError("ler diff", diff, err)
	}
	files, err := r.run(ctx, "diff", "--cached", "--name-only")
	if err != nil {
		return Context{}, r.commandError("listar arquivos", files, err)
	}
	history := ""
	hasCommit, err := r.hasCommit(ctx)
	if err != nil {
		return Context{}, err
	}
	if hasCommit {
		history, err = r.run(ctx, "log", "--oneline", "-n", "20")
		if err != nil {
			return Context{}, r.commandError("ler histórico", history, err)
		}
	}
	return Context{Files: strings.TrimSpace(files), Diff: relevantDiff(diff), History: strings.TrimSpace(history)}, nil
}

func (r Repository) hasCommit(ctx context.Context) (bool, error) {
	out, err := r.run(ctx, "rev-parse", "--verify", "--quiet", "HEAD")
	if err == nil {
		return true, nil
	}
	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
		return false, nil
	}
	return false, r.commandError("verificar histórico", out, err)
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
	var out string
	var err error
	if _, lookupErr := r.run(ctx, "show-ref", "--verify", "--quiet", "refs/heads/"+branch); lookupErr == nil {
		out, err = r.run(ctx, "checkout", branch)
	} else {
		out, err = r.run(ctx, "checkout", "-b", branch)
	}
	return r.commandError("criar ou trocar a branch", out, err)
}

func (r Repository) Sync(ctx context.Context, language i18n.Language, progress Progress) (resultErr error) {
	report := func(message string) {
		if progress != nil {
			progress(message)
		}
	}
	report(i18n.T(language, "sync_stage"))
	if out, err := r.run(ctx, "add", "-A"); err != nil {
		return r.commandError("adicionar alterações", out, err)
	}
	status, err := r.run(ctx, "status", "--porcelain")
	if err != nil {
		return r.commandError("verificar alterações", status, err)
	}
	stashed := strings.TrimSpace(status) != ""
	needsRestore := false
	restoreStash := func() error {
		report(i18n.T(language, "sync_restore"))
		if out, err := r.run(context.WithoutCancel(ctx), "stash", "pop", "--index"); err != nil {
			return r.commandError("restaurar alterações", out, err)
		}
		needsRestore = false
		return nil
	}
	if stashed {
		report(i18n.T(language, "sync_stash"))
		if out, err := r.run(ctx, "stash", "push", "-u", "-m", "commit-ai-auto-stash"); err != nil {
			return r.commandError("guardar alterações temporariamente", out, err)
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
		return r.commandError("identificar branch atual", branch, err)
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
	out, err := r.run(ctx, "add", "-A")
	return r.commandError("adicionar alterações após sincronização", out, err)
}

func (r Repository) Commit(ctx context.Context, message string) error {
	out, err := r.run(ctx, "commit", "-m", message)
	return r.commandError("criar commit", out, err)
}

func (r Repository) Undo(ctx context.Context) (string, error) {
	message, err := r.run(ctx, "log", "-1", "--pretty=%B")
	if err != nil {
		return "", r.commandError("ler o último commit", message, err)
	}
	out, err := r.run(ctx, "reset", "--soft", "HEAD~1")
	if err != nil {
		return "", r.commandError("desfazer o último commit", out, err)
	}
	return strings.TrimSpace(message), nil
}

func (r Repository) Push(ctx context.Context, branch string) error {
	if branch != "" {
		if out, err := r.run(ctx, "push", "-u", "origin", branch); err == nil {
			return nil
		} else if _, err2 := r.run(ctx, "push", "origin", branch); err2 == nil {
			return nil
		} else {
			return r.commandError("enviar alterações", out, err)
		}
	}

	out, err := r.run(ctx, "push")
	if err == nil {
		return nil
	}

	currentBranch, branchErr := r.run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if branchErr == nil {
		cb := strings.TrimSpace(currentBranch)
		if cb != "" && cb != "HEAD" {
			if _, err2 := r.run(ctx, "push", "-u", "origin", cb); err2 == nil {
				return nil
			} else if _, err3 := r.run(ctx, "push", "origin", cb); err3 == nil {
				return nil
			}
		}
	}

	return r.commandError("enviar alterações", out, err)
}

func (r Repository) run(ctx context.Context, args ...string) (string, error) {
	command := exec.CommandContext(ctx, "git", args...)
	command.Dir = r.Dir
	output, err := command.CombinedOutput()
	return string(output), err
}

func (r Repository) commandError(action string, output string, err error) error {
	if err == nil {
		return nil
	}
	out := strings.TrimSpace(output)
	if out != "" {
		lines := strings.Split(out, "\n")
		for _, line := range lines {
			l := strings.TrimSpace(line)
			if l != "" && !strings.HasPrefix(l, "On branch") && !strings.HasPrefix(l, "Your branch") {
				return fmt.Errorf("não foi possível %s: %s", action, l)
			}
		}
		return fmt.Errorf("não foi possível %s: %s", action, lines[0])
	}
	return fmt.Errorf("não foi possível %s: %w", action, err)
}
