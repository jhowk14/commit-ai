package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAndLoadLegacyCompatibleConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg := Default()
	cfg.Format = "gitmoji"
	cfg.AutoConfirm = true
	cfg.PushMode = PushAlways
	cfg.Language = "en"
	cfg.UseCustomPrompt = true
	cfg.Provider = "openai"
	cfg.Model = "gpt-oss-120b"
	cfg.OpenAIBaseURL = "https://api.cerebras.ai/v1"
	cfg.OpenAIAPIKey = "test-key"
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded != cfg {
		t.Fatalf("configuração diferente:\nquer: %#v\nobteve: %#v", cfg, loaded)
	}
	info, err := os.Stat(filepath.Join(home, FileName))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("permissão insegura: %o", info.Mode().Perm())
	}
}

func TestLoadAcceptsLegacyAliasesAndIgnoresInvalidValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	content := "format=gitmoji\nauto_confirm=true\nask_push=true\npush_mode=always\nlanguage=en\nbase_url=http://localhost:11434/v1\nprovider=openai\nmodel=test\nunknown=value\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Format != "gitmoji" || !cfg.AutoConfirm || cfg.PushMode != PushAlways || cfg.Language != "en" || cfg.OpenAIBaseURL != "http://localhost:11434/v1" || cfg.Model != "test" {
		t.Fatalf("configuração: %#v", cfg)
	}
}

func TestValidateProviderAndShowDoNotLeakFullKeys(t *testing.T) {
	cfg := Default()
	if err := cfg.ValidateProvider(); err == nil {
		t.Fatal("Gemini sem chave deveria falhar")
	}
	cfg.Provider, cfg.OpenAIBaseURL = "openai", "http://localhost:11434/v1"
	if err := cfg.ValidateProvider(); err != nil {
		t.Fatalf("servidor local não exige chave: %v", err)
	}
	cfg.OpenAIBaseURL, cfg.OpenAIAPIKey = "https://api.example.test/v1", "secret-value"
	if err := cfg.ValidateProvider(); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	Show(cfg, &out)
	if bytes.Contains(out.Bytes(), []byte("secret-value")) {
		t.Fatal("Show expôs a chave inteira")
	}
	if bytes.Contains(out.Bytes(), []byte("MISSING")) || !bytes.Contains(out.Bytes(), []byte("envio ao remoto:")) {
		t.Fatalf("Show formatou a configuração incorretamente: %s", out.String())
	}
}

func TestSetupPersistsInteractiveValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	input := "2\n2\n1\n1\n1\n2\n7\n1\ntest-key\n"
	var out bytes.Buffer
	cfg, err := Setup(Default(), bytes.NewBufferString(input), &out)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Language != "en" || cfg.Format != "gitmoji" || !cfg.AutoConfirm || cfg.PushMode != PushAlways || !cfg.UseCustomPrompt || cfg.Provider != "openai" || cfg.Model != "gpt-oss-120b" {
		t.Fatalf("setup: %#v", cfg)
	}
	if !bytes.Contains(out.Bytes(), []byte("Cerebras")) || !bytes.Contains(out.Bytes(), []byte("gpt-oss-120b")) || bytes.Contains(out.Bytes(), []byte("test-key")) {
		t.Fatalf("setup não exibiu os presets ou expôs a chave:\n%s", out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("Choose")) || !bytes.Contains(out.Bytes(), []byte("Configuration complete")) {
		t.Fatalf("setup não aplicou o idioma escolhido:\n%s", out.String())
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded != cfg {
		t.Fatalf("configuração persistida: %#v != %#v", loaded, cfg)
	}
}

func TestProviderPresetsChooseCompatibleEndpointsAndModels(t *testing.T) {
	for _, testCase := range []struct {
		url, preset, model string
	}{
		{"https://api.cerebras.ai/v1", "cerebras", "gpt-oss-120b"},
		{"https://api.groq.com/openai/v1", "groq", "llama-3.3-70b-versatile"},
		{"http://localhost:11434/v1", "ollama", "llama3.2"},
		{"https://example.test/v1", "custom", "gpt-4o-mini"},
	} {
		if got := endpointPresetForURL(testCase.url); got != testCase.preset {
			t.Fatalf("preset para %s: %q", testCase.url, got)
		}
		if got := defaultModel("openai", testCase.preset); got != testCase.model {
			t.Fatalf("modelo padrão para %s: %q", testCase.preset, got)
		}
	}
}

func TestSetupMasksStoredKeyAndRejectsInvalidMenuChoice(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	current := Default()
	current.GeminiAPIKey = "secret-key"
	// The first invalid choice must be rejected before the normal menu flow.
	input := "bad\n1\n\n\n\n\n\n\n\n"
	var out bytes.Buffer
	if _, err := Setup(current, bytes.NewBufferString(input), &out); err != nil {
		t.Fatal(err)
	}
	contents := out.String()
	if !strings.Contains(contents, "Opção inválida") || strings.Contains(contents, "secret-key") || !strings.Contains(contents, "****-key") {
		t.Fatalf("saída insegura ou sem validação:\n%s", contents)
	}
}

func TestLegacyAskPushMigratesToAskMode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("ask_push=true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PushMode != PushAsk {
		t.Fatalf("modo migrado: %q", cfg.PushMode)
	}
}

func TestEnvironmentKeyWinsOverConfig(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "from-environment")
	cfg := Default()
	cfg.OpenAIAPIKey = "from-file"
	if got := cfg.ResolvedOpenAIKey(); got != "from-environment" {
		t.Fatalf("chave: %q", got)
	}
}

func TestResolvedGeminiAndCustomPromptPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg := Default()
	cfg.GeminiAPIKey = "from-file"
	if got := cfg.ResolvedGeminiKey(); got != "from-file" {
		t.Fatalf("chave Gemini: %q", got)
	}
	t.Setenv("GEMINI_API_KEY", "from-environment")
	if got := cfg.ResolvedGeminiKey(); got != "from-environment" {
		t.Fatalf("chave Gemini: %q", got)
	}
	path, err := CustomPromptPath()
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(home, CustomPromptName) {
		t.Fatalf("caminho: %q", path)
	}
}

func TestResolvedCompatibleProviderKeysPreferNativeEnvironmentVariable(t *testing.T) {
	testCases := []struct {
		baseURL, variable string
	}{
		{"https://api.cerebras.ai/v1", "CEREBRAS_API_KEY"},
		{"https://openrouter.ai/api/v1", "OPENROUTER_API_KEY"},
		{"https://api.groq.com/openai/v1", "GROQ_API_KEY"},
		{"https://api.deepseek.com/v1", "DEEPSEEK_API_KEY"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.variable, func(t *testing.T) {
			t.Setenv("OPENAI_API_KEY", "fallback-key")
			t.Setenv(testCase.variable, "native-key")
			cfg := Default()
			cfg.OpenAIBaseURL, cfg.OpenAIAPIKey = testCase.baseURL, "saved-key"
			if got := cfg.ResolvedOpenAIKey(); got != "native-key" {
				t.Fatalf("chave resolvida: %q", got)
			}
		})
	}
}
