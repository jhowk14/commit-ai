package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadLegacyCompatibleConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg := Default()
	cfg.Format = "gitmoji"
	cfg.AutoConfirm = true
	cfg.AskPush = true
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
	content := "format=gitmoji\nauto_confirm=true\nbase_url=http://localhost:11434/v1\nprovider=openai\nmodel=test\nunknown=value\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Format != "gitmoji" || !cfg.AutoConfirm || cfg.OpenAIBaseURL != "http://localhost:11434/v1" || cfg.Model != "test" {
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
}

func TestSetupPersistsInteractiveValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	input := "gitmoji\ns\nn\ns\nopenai\nhttps://api.cerebras.ai/v1\ngpt-oss-120b\ntest-key\n"
	var out bytes.Buffer
	cfg, err := Setup(Default(), bytes.NewBufferString(input), &out)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Format != "gitmoji" || !cfg.AutoConfirm || cfg.AskPush || !cfg.UseCustomPrompt || cfg.Provider != "openai" || cfg.Model != "gpt-oss-120b" {
		t.Fatalf("setup: %#v", cfg)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded != cfg {
		t.Fatalf("configuração persistida: %#v != %#v", loaded, cfg)
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
