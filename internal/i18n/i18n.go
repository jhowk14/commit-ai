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
		"language":                  "Idioma (pt-BR/en)",
		"setup_title":               "\n🤖 commit-ai 2.0 — configuração",
		"setup_language_title":      "🌐 Idioma da interface",
		"setup_language_pt":         "Português (Brasil)",
		"setup_language_en":         "English",
		"setup_format_title":        "📝 Formato da mensagem de commit",
		"setup_format_conventional": "Conventional Commits — feat:, fix:, docs:",
		"setup_format_gitmoji":      "Gitmoji — ✨, 🐛, 📝",
		"setup_auto_confirm_title":  "⚡ Confirmar commits automaticamente?",
		"setup_yes":                 "Sim — não pedir confirmação",
		"setup_no":                  "Não — revisar a mensagem antes do commit",
		"setup_push_always":         "Enviar automaticamente após o commit",
		"setup_push_ask":            "Perguntar antes de enviar",
		"setup_push_never":          "Não enviar automaticamente",
		"setup_custom_prompt_title": "📝 Usar arquivo de prompt personalizado?",
		"setup_custom_yes":          "Sim — usar ~/.commit-ai-prompt.txt",
		"setup_custom_no":           "Não — usar o prompt padrão",
		"setup_provider_title":      "🔌 Provedor de IA",
		"setup_provider_gemini":     "Gemini — Google",
		"setup_provider_openai":     "OpenAI ou API compatível — OpenAI, Cerebras, Groq, Ollama e mais",
		"setup_endpoint_title":      "🌐 Endpoint compatível com OpenAI",
		"endpoint_openai":           "OpenAI oficial — https://api.openai.com/v1",
		"endpoint_openrouter":       "OpenRouter — vários provedores",
		"endpoint_groq":             "Groq — inferência rápida",
		"endpoint_deepseek":         "DeepSeek",
		"endpoint_ollama":           "Ollama local — http://localhost:11434/v1",
		"endpoint_lmstudio":         "LM Studio / LocalAI — http://localhost:1234/v1",
		"endpoint_cerebras":         "Cerebras — modelos rápidos e gpt-oss",
		"endpoint_custom":           "URL personalizada",
		"setup_custom_url":          "Digite a Base URL personalizada (Enter mantém a atual):",
		"setup_model_title":         "🧠 Modelo",
		"model_recommended":         "recomendado",
		"model_fast":                "rápido",
		"model_advanced":            "avançado",
		"model_reasoning":           "raciocínio",
		"model_fast_recommended":    "rápido, recomendado",
		"model_custom":              "Nome de modelo personalizado",
		"setup_custom_model":        "Digite o nome do modelo personalizado:",
		"setup_api_key_title":       "🔐 Chave da API",
		"setup_gemini_key":          "Digite a chave da API Gemini (Enter mantém a atual):",
		"setup_compatible_key":      "Chave para o provedor compatível com OpenAI (%s):",
		"setup_api_key":             "Digite a chave da API (Enter mantém a atual):",
		"setup_local_api_key":       "Digite a chave da API (opcional para servidor local):",
		"setup_key_current":         "  Chave configurada:",
		"setup_key_environment":     "  Usando a variável de ambiente %s (tem prioridade).\n",
		"setup_key_optional":        "  Não configurada — opcional para servidor local.",
		"setup_key_missing":         "  Não configurada.",
		"setup_choose":              "Escolha",
		"setup_invalid_choice":      "⚠️  Opção inválida. Digite um número entre 1 e %d.\n",
		"setup_config_path":         "📂 Configuração salva em %s\n\n",
		"setup_complete":            "Configuração concluída. Execute 'commit-ai' para começar.",
		"format":                    "Formato (conventional/gitmoji)",
		"auto_confirm":              "Confirmar commits automaticamente",
		"custom_prompt":             "Usar prompt customizado",
		"provider":                  "Provedor (gemini/openai)",
		"gemini_model":              "Modelo Gemini",
		"openai_base_url":           "Base URL compatível com OpenAI",
		"model":                     "Modelo",
		"gemini_key":                "Chave Gemini (vazio mantém)",
		"api_key":                   "Chave da API (vazio mantém)",
		"saved":                     "✅ Configuração salva.",
		"current":                   "atual: %s",
		"push_title":                "Envio ao remoto após o commit:",
		"push_always":               "1. Enviar automaticamente",
		"push_ask":                  "2. Perguntar antes de enviar",
		"push_never":                "3. Não enviar automaticamente",
		"push_choice":               "Escolha (1/2/3)",
		"push_mode_always":          "automático",
		"push_mode_ask":             "perguntar",
		"push_mode_never":           "não enviar",
		"config_title":              "Configuração do commit-ai",
		"not_configured":            "(não configurada)",
		"config":                    "  idioma: %s\n  formato: %s\n  confirmação automática: %t\n  envio ao remoto: %s\n  prompt customizado: %t\n  provedor: %s\n  modelo: %s\n  base URL: %s\n  chave Gemini: %s\n  chave OpenAI/compatível: %s\n",
		"auto_sync":                 "🔄 Auto-sync ativado. Sincronizando com a branch remota...",
		"undo":                      "✅ Último commit desfeito. Alterações permanecem preparadas.\nMensagem anterior: %s\n",
		"branch":                    "🌱 Abrindo branch %q...",
		"no_staged":                 "nenhuma alteração preparada; use git add <arquivo> ou commit-ai --sync",
		"empty_message":             "a IA retornou uma mensagem de commit vazia",
		"preview":                   "\nPrévia da mensagem:\n\n  %s\n",
		"committed":                 "✅ Commit criado: %s\n",
		"pushed_branch":             "🚀 Enviado para origin/%s\n",
		"pushed":                    "🚀 Alterações enviadas.",
		"push_prompt":               "Enviar para o repositório remoto? (s/N)",
		"editor_fallback":           "\n📝 Mensagem gerada (Enter mantém a sugestão):\n  %s\n\n> ",
		"editor_prompt":             "📝 Mensagem de commit: ",
		"commit_canceled":           "commit cancelado",
		"sync_fetch":                "⬇️ Verificando atualizações de origin/%s...",
		"sync_stash":                "📦 Guardando alterações locais temporariamente...",
		"sync_update":               "🔄 Aplicando atualizações de origin/%s...",
		"sync_restore":              "📂 Restaurando alterações salvas...",
		"sync_prepare":              "➕ Preparando os arquivos para o commit...",
	},
	English: {
		"language":                  "Language (pt-BR/en)",
		"setup_title":               "\n🤖 commit-ai 2.0 — setup",
		"setup_language_title":      "🌐 Interface language",
		"setup_language_pt":         "Português (Brazil)",
		"setup_language_en":         "English",
		"setup_format_title":        "📝 Commit message format",
		"setup_format_conventional": "Conventional Commits — feat:, fix:, docs:",
		"setup_format_gitmoji":      "Gitmoji — ✨, 🐛, 📝",
		"setup_auto_confirm_title":  "⚡ Confirm commits automatically?",
		"setup_yes":                 "Yes — skip the confirmation prompt",
		"setup_no":                  "No — review the message before committing",
		"setup_push_always":         "Push automatically after committing",
		"setup_push_ask":            "Ask before pushing",
		"setup_push_never":          "Do not push automatically",
		"setup_custom_prompt_title": "📝 Use a custom prompt file?",
		"setup_custom_yes":          "Yes — use ~/.commit-ai-prompt.txt",
		"setup_custom_no":           "No — use the default prompt",
		"setup_provider_title":      "🔌 AI provider",
		"setup_provider_gemini":     "Gemini — Google",
		"setup_provider_openai":     "OpenAI or compatible API — OpenAI, Cerebras, Groq, Ollama, and more",
		"setup_endpoint_title":      "🌐 OpenAI-compatible endpoint",
		"endpoint_openai":           "Official OpenAI — https://api.openai.com/v1",
		"endpoint_openrouter":       "OpenRouter — multiple providers",
		"endpoint_groq":             "Groq — fast inference",
		"endpoint_deepseek":         "DeepSeek",
		"endpoint_ollama":           "Local Ollama — http://localhost:11434/v1",
		"endpoint_lmstudio":         "LM Studio / LocalAI — http://localhost:1234/v1",
		"endpoint_cerebras":         "Cerebras — fast models and gpt-oss",
		"endpoint_custom":           "Custom URL",
		"setup_custom_url":          "Enter the custom Base URL (Enter keeps the current one):",
		"setup_model_title":         "🧠 Model",
		"model_recommended":         "recommended",
		"model_fast":                "fast",
		"model_advanced":            "advanced",
		"model_reasoning":           "reasoning",
		"model_fast_recommended":    "fast, recommended",
		"model_custom":              "Custom model name",
		"setup_custom_model":        "Enter the custom model name:",
		"setup_api_key_title":       "🔐 API key",
		"setup_gemini_key":          "Enter the Gemini API key (Enter keeps the current one):",
		"setup_compatible_key":      "API key for the OpenAI-compatible provider (%s):",
		"setup_api_key":             "Enter the API key (Enter keeps the current one):",
		"setup_local_api_key":       "Enter the API key (optional for a local server):",
		"setup_key_current":         "  Configured key:",
		"setup_key_environment":     "  Using the %s environment variable (it takes precedence).\n",
		"setup_key_optional":        "  Not configured — optional for a local server.",
		"setup_key_missing":         "  Not configured.",
		"setup_choose":              "Choose",
		"setup_invalid_choice":      "⚠️  Invalid option. Enter a number from 1 to %d.\n",
		"setup_config_path":         "📂 Configuration saved to %s\n\n",
		"setup_complete":            "Configuration complete. Run 'commit-ai' to start.",
		"format":                    "Format (conventional/gitmoji)",
		"auto_confirm":              "Confirm commits automatically",
		"custom_prompt":             "Use a custom prompt",
		"provider":                  "Provider (gemini/openai)",
		"gemini_model":              "Gemini model",
		"openai_base_url":           "OpenAI-compatible base URL",
		"model":                     "Model",
		"gemini_key":                "Gemini API key (leave blank to keep)",
		"api_key":                   "API key (leave blank to keep)",
		"saved":                     "✅ Configuration saved.",
		"current":                   "current: %s",
		"push_title":                "Push after committing:",
		"push_always":               "1. Push automatically",
		"push_ask":                  "2. Ask before pushing",
		"push_never":                "3. Do not push automatically",
		"push_choice":               "Choose (1/2/3)",
		"push_mode_always":          "automatic",
		"push_mode_ask":             "ask",
		"push_mode_never":           "do not push",
		"config_title":              "commit-ai configuration",
		"not_configured":            "(not configured)",
		"config":                    "  language: %s\n  format: %s\n  auto-confirm: %t\n  remote push: %s\n  custom prompt: %t\n  provider: %s\n  model: %s\n  base URL: %s\n  Gemini key: %s\n  OpenAI-compatible key: %s\n",
		"auto_sync":                 "🔄 Auto-sync enabled. Synchronizing with the remote branch...",
		"undo":                      "✅ Last commit undone. Changes remain staged.\nPrevious message: %s\n",
		"branch":                    "🌱 Opening branch %q...",
		"no_staged":                 "no changes are staged; use git add <file> or commit-ai --sync",
		"empty_message":             "the AI returned an empty commit message",
		"preview":                   "\nCommit message preview:\n\n  %s\n",
		"committed":                 "✅ Commit created: %s\n",
		"pushed_branch":             "🚀 Pushed to origin/%s\n",
		"pushed":                    "🚀 Changes pushed.",
		"push_prompt":               "Push to the remote repository? (y/N)",
		"editor_fallback":           "\n📝 Generated message (Enter keeps the suggestion):\n  %s\n\n> ",
		"editor_prompt":             "📝 Commit message: ",
		"commit_canceled":           "commit canceled",
		"sync_fetch":                "⬇️ Checking updates from origin/%s...",
		"sync_stash":                "📦 Temporarily stashing local changes...",
		"sync_update":               "🔄 Applying updates from origin/%s...",
		"sync_restore":              "📂 Restoring stashed changes...",
		"sync_prepare":              "➕ Preparing files for the commit...",
	},
}
