package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jhowk14/commit-ai/v2/internal/i18n"
)

const (
	PushAlways = "always"
	PushAsk    = "ask"
	PushNever  = "never"
)

const (
	FileName         = ".commit-ai.conf"
	CustomPromptName = ".commit-ai-prompt.txt"
)

type Config struct {
	Format          string
	AutoConfirm     bool
	PushMode        string
	UseCustomPrompt bool
	Language        string
	Provider        string
	Model           string
	OpenAIBaseURL   string
	GeminiAPIKey    string
	OpenAIAPIKey    string
}

func Default() Config {
	return Config{
		Format:        "conventional",
		PushMode:      PushNever,
		Language:      string(i18n.Portuguese),
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
	case "push_mode":
		cfg.PushMode = NormalizePushMode(value)
	case "ask_push": // Compatibility with commit-ai 1.x and 2.0.0–2.0.2.
		if value == "true" {
			cfg.PushMode = PushAsk
		} else {
			cfg.PushMode = PushNever
		}
	case "language":
		if i18n.IsValid(value) {
			cfg.Language = string(i18n.Normalize(value))
		}
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
	cfg.PushMode = NormalizePushMode(cfg.PushMode)
	cfg.Language = string(i18n.Normalize(cfg.Language))
	content := fmt.Sprintf(`# commit-ai configuration
# This file is compatible with commit-ai 1.x.
language=%s
format=%s
auto_confirm=%t
# Legacy clients understand ask_push. Automatic push is unavailable in 1.x.
ask_push=%t
push_mode=%s
use_custom_prompt=%t
provider=%s
model=%s
openai_base_url=%s
gemini_api_key=%s
openai_api_key=%s
`, cfg.Language, cfg.Format, cfg.AutoConfirm, cfg.PushMode == PushAsk, cfg.PushMode, cfg.UseCustomPrompt, cfg.Provider, cfg.Model, cfg.OpenAIBaseURL, cfg.GeminiAPIKey, cfg.OpenAIAPIKey)
	return os.WriteFile(path, []byte(content), 0o600)
}

func NormalizePushMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case PushAlways, "auto", "automatic", "automatico", "automático":
		return PushAlways
	case PushAsk, "prompt", "perguntar":
		return PushAsk
	default:
		return PushNever
	}
}

func (cfg Config) UILanguage() i18n.Language { return i18n.Normalize(cfg.Language) }

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
	language := cfg.UILanguage()
	mask := func(value string) string {
		if value == "" {
			return i18n.T(language, "not_configured")
		}
		if len(value) <= 4 {
			return "****"
		}
		return "****" + value[len(value)-4:]
	}
	fmt.Fprintln(out, i18n.T(language, "config_title"))
	fmt.Fprintf(out, i18n.T(language, "config"), cfg.Language, cfg.Format, cfg.AutoConfirm, pushModeLabel(language, cfg.PushMode), cfg.UseCustomPrompt, cfg.Provider, cfg.Model, cfg.OpenAIBaseURL, mask(cfg.GeminiAPIKey), mask(cfg.OpenAIAPIKey))
}

func pushModeLabel(language i18n.Language, mode string) string {
	return i18n.T(language, "push_mode_"+NormalizePushMode(mode))
}
