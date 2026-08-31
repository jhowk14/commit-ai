package config

import (
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

func TestEnvironmentKeyWinsOverConfig(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "from-environment")
	cfg := Default()
	cfg.OpenAIAPIKey = "from-file"
	if got := cfg.ResolvedOpenAIKey(); got != "from-environment" {
		t.Fatalf("chave: %q", got)
	}
}
