package app

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/jhowk14/commit-ai/v2/internal/ai"
	"github.com/jhowk14/commit-ai/v2/internal/config"
	gitrepo "github.com/jhowk14/commit-ai/v2/internal/git"
)

type App struct {
	version string
	in      io.Reader
	out     io.Writer
	err     io.Writer
	workDir string
	client  ai.Client
}

func New(version string, in io.Reader, out, err io.Writer) *App {
	workDir, _ := os.Getwd()
	return &App{version: version, in: in, out: out, err: err, workDir: workDir, client: ai.NewClient()}
}

type options struct {
	emoji, conventional, custom, preview, yes, sync, undo bool
	setup, showConfig, editPrompt, help, version          bool
	message, branch, baseURL                              string
}

func (a *App) Run(ctx context.Context, args []string) error {
	opts, err := parse(args, a.err)
	if err != nil {
		return err
	}
	if opts.help {
		a.printHelp()
		return nil
	}
	if opts.version {
		fmt.Fprintf(a.out, "commit-ai v%s\n", a.version)
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if opts.setup {
		_, err := config.Setup(cfg, a.in, a.out)
		return err
	}
	if opts.showConfig {
		config.Show(cfg, a.out)
		return nil
	}
	if opts.editPrompt {
		return a.editPrompt()
	}

	if opts.baseURL != "" {
		cfg.OpenAIBaseURL = opts.baseURL
		cfg.Provider = "openai"
	}
	if opts.emoji {
		cfg.Format = "gitmoji"
	}
	if opts.conventional {
		cfg.Format = "conventional"
	}
	if opts.custom {
		cfg.UseCustomPrompt = true
	}

	repo := gitrepo.Repository{Dir: a.workDir}
	if err := repo.Ensure(ctx); err != nil {
		return err
	}
	if opts.undo {
		message, err := repo.Undo(ctx)
		if err != nil {
			return err
		}
		fmt.Fprintf(a.out, "✅ Último commit desfeito. Alterações permanecem preparadas.\nMensagem anterior: %s\n", message)
		return nil
	}
	if opts.branch != "" {
		fmt.Fprintf(a.out, "🌱 Abrindo branch %q...\n", opts.branch)
		if err := repo.CreateOrSwitchBranch(ctx, opts.branch); err != nil {
			return err
		}
	}
	if opts.sync {
		fmt.Fprintln(a.out, "🔄 Sincronizando alterações com a branch remota...")
		if err := repo.Sync(ctx); err != nil {
			return err
		}
	}
	hasStaged, err := repo.HasStaged(ctx)
	if err != nil {
		return err
	}
	if !hasStaged {
		return errors.New("nenhuma alteração preparada; use git add <arquivo> ou commit-ai --sync")
	}
	gitContext, err := repo.Context(ctx)
	if err != nil {
		return err
	}
	customPrompt, err := a.customPrompt(cfg.UseCustomPrompt)
	if err != nil {
		return err
	}
	message, err := a.client.Generate(ctx, cfg, ai.Input{Git: gitContext, Format: cfg.Format, Hint: opts.message, UseCustomPrompt: cfg.UseCustomPrompt, CustomPrompt: customPrompt})
	if err != nil {
		return err
	}
	if message == "" {
		return errors.New("a IA retornou uma mensagem de commit vazia")
	}

	if opts.preview {
		fmt.Fprintf(a.out, "\nPrévia da mensagem:\n\n  %s\n", message)
		return nil
	}
	if !opts.yes && !cfg.AutoConfirm {
		message, err = a.confirmMessage(message)
		if err != nil {
			return err
		}
	}
	if err := repo.Commit(ctx, message); err != nil {
		return err
	}
	fmt.Fprintf(a.out, "✅ Commit criado: %s\n", message)
	if opts.branch != "" {
		if err := repo.Push(ctx, opts.branch); err != nil {
			return err
		}
		fmt.Fprintf(a.out, "🚀 Enviado para origin/%s\n", opts.branch)
	} else if cfg.AskPush && a.confirm("Enviar para o repositório remoto? (s/N)") {
		if err := repo.Push(ctx, ""); err != nil {
			return err
		}
		fmt.Fprintln(a.out, "🚀 Alterações enviadas.")
	}
	return nil
}

func parse(args []string, stderr io.Writer) (options, error) {
	var opts options
	flags := flag.NewFlagSet("commit-ai", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.BoolVar(&opts.emoji, "e", false, "usar Gitmoji")
	flags.BoolVar(&opts.emoji, "emoji", false, "usar Gitmoji")
	flags.BoolVar(&opts.conventional, "c", false, "usar Conventional Commits")
	flags.BoolVar(&opts.conventional, "conv", false, "usar Conventional Commits")
	flags.BoolVar(&opts.custom, "C", false, "usar prompt customizado")
	flags.BoolVar(&opts.custom, "custom", false, "usar prompt customizado")
	flags.BoolVar(&opts.preview, "p", false, "prévia")
	flags.BoolVar(&opts.preview, "preview", false, "prévia")
	flags.BoolVar(&opts.yes, "y", false, "confirmar automaticamente")
	flags.BoolVar(&opts.yes, "yes", false, "confirmar automaticamente")
	flags.BoolVar(&opts.sync, "s", false, "sincronizar antes")
	flags.BoolVar(&opts.sync, "S", false, "sincronizar antes")
	flags.BoolVar(&opts.sync, "sync", false, "sincronizar antes")
	flags.BoolVar(&opts.undo, "u", false, "desfazer último commit")
	flags.BoolVar(&opts.undo, "undo", false, "desfazer último commit")
	flags.StringVar(&opts.message, "m", "", "contexto")
	flags.StringVar(&opts.message, "message", "", "contexto")
	flags.StringVar(&opts.branch, "b", "", "branch")
	flags.StringVar(&opts.branch, "branch", "", "branch")
	flags.StringVar(&opts.baseURL, "B", "", "base URL OpenAI compatível")
	flags.StringVar(&opts.baseURL, "base-url", "", "base URL OpenAI compatível")
	flags.BoolVar(&opts.setup, "setup", false, "configurar")
	flags.BoolVar(&opts.showConfig, "config", false, "mostrar configuração")
	flags.BoolVar(&opts.editPrompt, "edit-prompt", false, "editar prompt")
	flags.BoolVar(&opts.help, "h", false, "ajuda")
	flags.BoolVar(&opts.help, "help", false, "ajuda")
	flags.BoolVar(&opts.version, "v", false, "versão")
	flags.BoolVar(&opts.version, "version", false, "versão")
	if err := flags.Parse(args); err != nil {
		return opts, err
	}
	if len(flags.Args()) > 0 {
		return opts, fmt.Errorf("argumentos desconhecidos: %s", strings.Join(flags.Args(), " "))
	}
	return opts, nil
}

func (a *App) customPrompt(enabled bool) (string, error) {
	if !enabled {
		return "", nil
	}
	path, err := config.CustomPromptPath()
	if err != nil {
		return "", err
	}
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("o prompt customizado não existe; execute commit-ai --edit-prompt")
	}
	if err != nil {
		return "", err
	}
	lines := make([]string, 0)
	for _, line := range strings.Split(string(contents), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			lines = append(lines, line)
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n")), nil
}

func (a *App) editPrompt() error {
	path, err := config.CustomPromptPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		content := "# Use {HISTORY}, {FILES} e {DIFF} como variáveis.\nReturn only a valid commit message.\n"
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			return err
		}
	}
	var command *exec.Cmd
	if runtime.GOOS == "windows" {
		command = exec.Command("notepad", path)
	} else if editor := os.Getenv("EDITOR"); editor != "" {
		command = exec.Command(editor, path)
	} else {
		command = exec.Command("vi", path)
	}
	command.Stdin, command.Stdout, command.Stderr = a.in, a.out, a.err
	return command.Run()
}

func (a *App) confirmMessage(message string) (string, error) {
	fmt.Fprintf(a.out, "\nMensagem gerada:\n  %s\n\nPressione Enter para confirmar, digite outra mensagem ou 'n' para cancelar: ", message)
	value, err := bufio.NewReader(a.in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "n") || strings.EqualFold(value, "não") || strings.EqualFold(value, "nao") {
		return "", errors.New("commit cancelado")
	}
	if value != "" {
		return value, nil
	}
	return message, nil
}

func (a *App) confirm(label string) bool {
	fmt.Fprint(a.out, label+" ")
	value, _ := bufio.NewReader(a.in).ReadString('\n')
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "s" || value == "sim" || value == "y" || value == "yes"
}

func (a *App) printHelp() {
	fmt.Fprint(a.out, `commit-ai — mensagens de commit geradas por IA

Uso: commit-ai [opções]
  -e, --emoji              usa formato Gitmoji
  -c, --conv               usa Conventional Commits
  -m, --message <texto>    contexto adicional
  -b, --branch <nome>      cria/troca branch e envia após o commit
  -s, -S, --sync           sincroniza e prepara alterações antes do commit
  -C, --custom             usa ~/.commit-ai-prompt.txt
  -p, --preview            mostra a mensagem sem criar commit
  -y, --yes                confirma sem interação
  -u, --undo               desfaz o último commit, mantendo alterações preparadas
  -B, --base-url <url>     base URL de API OpenAI compatível
      --setup              configuração interativa
      --config             mostra a configuração atual
      --edit-prompt        cria/edita o prompt customizado
  -h, --help               mostra esta ajuda
  -v, --version            mostra a versão
`)
}
