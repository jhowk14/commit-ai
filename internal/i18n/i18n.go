// Package i18n contains the small, dependency-free set of messages shown by
// commit-ai. Commit-message prompts deliberately remain in English so that
// generated Git history stays consistent regardless of the UI language.
package i18n

import (
	"fmt"
	"strings"
)

type Language string

const (
	Portuguese Language = "pt-BR"
	English    Language = "en"
)

func Normalize(value string) Language {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "en", "en-us", "english":
		return English
	case "pt", "pt-br", "portuguese", "português":
		return Portuguese
	default:
		return Portuguese
	}
}

func IsValid(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "en" || value == "en-us" || value == "english" || value == "pt" || value == "pt-br" || value == "portuguese" || value == "português"
}

func T(language Language, key string, values ...any) string {
	if language != English {
		language = Portuguese
	}
	message := messages[language][key]
	if message == "" {
		message = key
	}
	if len(values) == 0 {
		return message
	}
	return fmt.Sprintf(message, values...)
}

var messages = map[Language]map[string]string{
	Portuguese: {
		"language":         "Idioma (pt-BR/en)",
		"setup_title":      "\n🤖 commit-ai 2.0 — configuração",
		"format":           "Formato (conventional/gitmoji)",
		"auto_confirm":     "Confirmar commits automaticamente",
		"custom_prompt":    "Usar prompt customizado",
		"provider":         "Provedor (gemini/openai)",
		"gemini_model":     "Modelo Gemini",
		"openai_base_url":  "Base URL compatível com OpenAI",
		"model":            "Modelo",
		"gemini_key":       "Chave Gemini (vazio mantém)",
		"api_key":          "Chave da API (vazio mantém)",
		"saved":            "✅ Configuração salva.",
		"current":          "atual: %s",
		"push_title":       "Envio ao remoto após o commit:",
		"push_always":      "1. Enviar automaticamente",
		"push_ask":         "2. Perguntar antes de enviar",
		"push_never":       "3. Não enviar automaticamente",
		"push_choice":      "Escolha (1/2/3)",
		"push_mode_always": "automático",
		"push_mode_ask":    "perguntar",
		"push_mode_never":  "não enviar",
		"config_title":     "Configuração do commit-ai",
		"not_configured":   "(não configurada)",
		"config":           "  idioma: %s\n  formato: %s\n  confirmação automática: %t\n  envio ao remoto: %s\n  prompt customizado: %t\n  provedor: %s\n  modelo: %s\n  base URL: %s\n  chave Gemini: %s\n  chave OpenAI/compatível: %s\n",
		"auto_sync":        "🔄 Auto-sync ativado. Sincronizando com a branch remota...",
		"undo":             "✅ Último commit desfeito. Alterações permanecem preparadas.\nMensagem anterior: %s\n",
		"branch":           "🌱 Abrindo branch %q...",
		"no_staged":        "nenhuma alteração preparada; use git add <arquivo> ou commit-ai --sync",
		"empty_message":    "a IA retornou uma mensagem de commit vazia",
		"preview":          "\nPrévia da mensagem:\n\n  %s\n",
		"committed":        "✅ Commit criado: %s\n",
		"pushed_branch":    "🚀 Enviado para origin/%s\n",
		"pushed":           "🚀 Alterações enviadas.",
		"push_prompt":      "Enviar para o repositório remoto? (s/N)",
		"editor_fallback":  "\n📝 Mensagem gerada (Enter mantém a sugestão):\n  %s\n\n> ",
		"editor_prompt":    "📝 Mensagem de commit: ",
		"commit_canceled":  "commit cancelado",
		"sync_stage":       "➕ Adicionando alterações locais ao staging...",
		"sync_stash":       "📦 Guardando alterações locais temporariamente...",
		"sync_pull":        "⬇️ Baixando atualizações de origin/%s...",
		"sync_restore":     "📂 Restaurando alterações salvas...",
		"sync_prepare":     "➕ Preparando os arquivos para o commit...",
	},
	English: {
		"language":         "Language (pt-BR/en)",
		"setup_title":      "\n🤖 commit-ai 2.0 — setup",
		"format":           "Format (conventional/gitmoji)",
		"auto_confirm":     "Confirm commits automatically",
		"custom_prompt":    "Use a custom prompt",
		"provider":         "Provider (gemini/openai)",
		"gemini_model":     "Gemini model",
		"openai_base_url":  "OpenAI-compatible base URL",
		"model":            "Model",
		"gemini_key":       "Gemini API key (leave blank to keep)",
		"api_key":          "API key (leave blank to keep)",
		"saved":            "✅ Configuration saved.",
		"current":          "current: %s",
		"push_title":       "Push after committing:",
		"push_always":      "1. Push automatically",
		"push_ask":         "2. Ask before pushing",
		"push_never":       "3. Do not push automatically",
		"push_choice":      "Choose (1/2/3)",
		"push_mode_always": "automatic",
		"push_mode_ask":    "ask",
		"push_mode_never":  "do not push",
		"config_title":     "commit-ai configuration",
		"not_configured":   "(not configured)",
		"config":           "  language: %s\n  format: %s\n  auto-confirm: %t\n  remote push: %s\n  custom prompt: %t\n  provider: %s\n  model: %s\n  base URL: %s\n  Gemini key: %s\n  OpenAI-compatible key: %s\n",
		"auto_sync":        "🔄 Auto-sync enabled. Synchronizing with the remote branch...",
		"undo":             "✅ Last commit undone. Changes remain staged.\nPrevious message: %s\n",
		"branch":           "🌱 Opening branch %q...",
		"no_staged":        "no changes are staged; use git add <file> or commit-ai --sync",
		"empty_message":    "the AI returned an empty commit message",
		"preview":          "\nCommit message preview:\n\n  %s\n",
		"committed":        "✅ Commit created: %s\n",
		"pushed_branch":    "🚀 Pushed to origin/%s\n",
		"pushed":           "🚀 Changes pushed.",
		"push_prompt":      "Push to the remote repository? (y/N)",
		"editor_fallback":  "\n📝 Generated message (Enter keeps the suggestion):\n  %s\n\n> ",
		"editor_prompt":    "📝 Commit message: ",
		"commit_canceled":  "commit canceled",
		"sync_stage":       "➕ Staging local changes...",
		"sync_stash":       "📦 Temporarily stashing local changes...",
		"sync_pull":        "⬇️ Pulling updates from origin/%s...",
		"sync_restore":     "📂 Restoring stashed changes...",
		"sync_prepare":     "➕ Preparing files for the commit...",
	},
}
