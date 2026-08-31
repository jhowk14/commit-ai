package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	FileName         = ".commit-ai.conf"
	CustomPromptName = ".commit-ai-prompt.txt"
)

type Config struct {
	Format          string
	AutoConfirm     bool
	AskPush         bool
	UseCustomPrompt bool
	Provider        string
	Model           string
	OpenAIBaseURL   string
	GeminiAPIKey    string
	OpenAIAPIKey    string
}

func Default() Config {
	return Config{
		Format:        "conventional",
		Provider:      "gemini",
		Model:         "gemini-3-flash-preview",
		OpenAIBaseURL: "https://api.openai.com/v1",
	}
}

func HomePath(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("não foi possível localizar a pasta do usuário: %w", err)
	}
	return filepath.Join(home, name), nil
}

func Path() (string, error)             { return HomePath(FileName) }
func CustomPromptPath() (string, error) { return HomePath(CustomPromptName) }

func Load() (Config, error) {
	cfg := Default()
	path, err := Path()
	if err != nil {
		return cfg, err
	}
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("não foi possível ler %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		apply(&cfg, strings.TrimSpace(key), strings.TrimSpace(value))
	}
	if err := scanner.Err(); err != nil {
		return cfg, fmt.Errorf("não foi possível ler a configuração: %w", err)
	}
	return cfg, nil
}

func apply(cfg *Config, key, value string) {
	switch key {
	case "format":
		if value == "gitmoji" || value == "conventional" {
			cfg.Format = value
		}
	case "auto_confirm":
		cfg.AutoConfirm = value == "true"
	case "ask_push":
		cfg.AskPush = value == "true"
	case "use_custom_prompt":
		cfg.UseCustomPrompt = value == "true"
	case "provider":
		if value == "gemini" || value == "openai" {
			cfg.Provider = value
		}
	case "model":
		if value != "" {
			cfg.Model = value
		}
	case "openai_base_url", "base_url":
		if value != "" {
			cfg.OpenAIBaseURL = value
		}
	case "gemini_api_key":
		cfg.GeminiAPIKey = value
	case "openai_api_key":
		cfg.OpenAIAPIKey = value
	}
}

func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	content := fmt.Sprintf(`# commit-ai configuration
# This file is compatible with commit-ai 1.x.
format=%s
auto_confirm=%t
ask_push=%t
use_custom_prompt=%t
provider=%s
model=%s
openai_base_url=%s
gemini_api_key=%s
openai_api_key=%s
`, cfg.Format, cfg.AutoConfirm, cfg.AskPush, cfg.UseCustomPrompt, cfg.Provider, cfg.Model, cfg.OpenAIBaseURL, cfg.GeminiAPIKey, cfg.OpenAIAPIKey)
	return os.WriteFile(path, []byte(content), 0o600)
}

func (cfg Config) ResolvedGeminiKey() string {
	if value := os.Getenv("GEMINI_API_KEY"); value != "" {
		return value
	}
	return cfg.GeminiAPIKey
}

func (cfg Config) ResolvedOpenAIKey() string {
	for _, name := range compatibleKeyVariables(cfg.OpenAIBaseURL) {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return cfg.OpenAIAPIKey
}

func compatibleKeyVariables(baseURL string) []string {
	baseURL = strings.ToLower(baseURL)
	switch {
	case strings.Contains(baseURL, "cerebras.ai"):
		return []string{"CEREBRAS_API_KEY", "OPENAI_API_KEY"}
	case strings.Contains(baseURL, "openrouter.ai"):
		return []string{"OPENROUTER_API_KEY", "OPENAI_API_KEY"}
	case strings.Contains(baseURL, "groq.com"):
		return []string{"GROQ_API_KEY", "OPENAI_API_KEY"}
	case strings.Contains(baseURL, "deepseek.com"):
		return []string{"DEEPSEEK_API_KEY", "OPENAI_API_KEY"}
	default:
		return []string{"OPENAI_API_KEY"}
	}
}

func (cfg Config) ValidateProvider() error {
	if cfg.Provider == "gemini" && cfg.ResolvedGeminiKey() == "" {
		return errors.New("a chave Gemini não está configurada; execute commit-ai --setup")
	}
	if cfg.Provider == "openai" && cfg.ResolvedOpenAIKey() == "" && !isLocalURL(cfg.OpenAIBaseURL) {
		return errors.New("a chave OpenAI/compatível não está configurada; execute commit-ai --setup")
	}
	return nil
}

func isLocalURL(value string) bool {
	value = strings.ToLower(value)
	return strings.Contains(value, "localhost") || strings.Contains(value, "127.0.0.1")
}

func Show(cfg Config, out io.Writer) {
	mask := func(value string) string {
		if value == "" {
			return "(não configurada)"
		}
		if len(value) <= 4 {
			return "****"
		}
		return "****" + value[len(value)-4:]
	}
	fmt.Fprintln(out, "Configuração do commit-ai")
	fmt.Fprintf(out, "  formato: %s\n  confirmação automática: %t\n  perguntar push: %t\n  prompt customizado: %t\n  provedor: %s\n  modelo: %s\n  base URL: %s\n  chave Gemini: %s\n  chave OpenAI/compatível: %s\n", cfg.Format, cfg.AutoConfirm, cfg.AskPush, cfg.UseCustomPrompt, cfg.Provider, cfg.Model, cfg.OpenAIBaseURL, mask(cfg.GeminiAPIKey), mask(cfg.OpenAIAPIKey))
}

func Setup(current Config, in io.Reader, out io.Writer) (Config, error) {
	reader := bufio.NewReader(in)
	ask := func(label, current string) (string, error) {
		fmt.Fprintf(out, "%s [atual: %s]: ", label, current)
		value, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return current, nil
		}
		return value, nil
	}
	askBool := func(label string, current bool) (bool, error) {
		value, err := ask(label+" (s/n)", map[bool]string{true: "s", false: "n"}[current])
		if err != nil {
			return false, err
		}
		switch strings.ToLower(value) {
		case "s", "sim", "y", "yes", "true":
			return true, nil
		case "n", "não", "nao", "no", "false":
			return false, nil
		}
		return current, nil
	}

	fmt.Fprintln(out, "\n🤖 commit-ai 2.0 — configuração")
	format, err := ask("Formato (conventional/gitmoji)", current.Format)
	if err != nil {
		return current, err
	}
	if format == "conventional" || format == "gitmoji" {
		current.Format = format
	}
	if current.AutoConfirm, err = askBool("Confirmar commits automaticamente", current.AutoConfirm); err != nil {
		return current, err
	}
	if current.AskPush, err = askBool("Perguntar antes de enviar ao remoto", current.AskPush); err != nil {
		return current, err
	}
	if current.UseCustomPrompt, err = askBool("Usar prompt customizado", current.UseCustomPrompt); err != nil {
		return current, err
	}
	provider, err := ask("Provedor (gemini/openai)", current.Provider)
	if err != nil {
		return current, err
	}
	if provider == "gemini" || provider == "openai" {
		current.Provider = provider
	}
	if current.Provider == "gemini" {
		current.Model, err = ask("Modelo Gemini", current.Model)
		if err != nil {
			return current, err
		}
		current.GeminiAPIKey, err = ask("Chave Gemini (vazio mantém)", current.GeminiAPIKey)
		if err != nil {
			return current, err
		}
	} else {
		current.OpenAIBaseURL, err = ask("Base URL compatível com OpenAI", current.OpenAIBaseURL)
		if err != nil {
			return current, err
		}
		current.Model, err = ask("Modelo", current.Model)
		if err != nil {
			return current, err
		}
		current.OpenAIAPIKey, err = ask("Chave da API (vazio mantém)", current.OpenAIAPIKey)
		if err != nil {
			return current, err
		}
	}
	if err := Save(current); err != nil {
		return current, err
	}
	fmt.Fprintln(out, "✅ Configuração salva.")
	return current, nil
}
