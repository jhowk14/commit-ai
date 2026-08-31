package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/jhowk14/commit-ai/v2/internal/i18n"
)

const setupDivider = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

type setupOption struct {
	Value string
	Label string
}

type setupPrompter struct {
	reader   *bufio.Reader
	out      io.Writer
	language i18n.Language
}

// Setup keeps the guided, menu-driven flow introduced in commit-ai 1.x while
// persisting the richer v2 configuration. Secrets are never echoed as defaults.
func Setup(current Config, in io.Reader, out io.Writer) (Config, error) {
	prompt := setupPrompter{reader: bufio.NewReader(in), out: out, language: current.UILanguage()}
	prompt.banner()

	language, _, err := prompt.choose("setup_language_title", []setupOption{
		{Value: string(i18n.Portuguese), Label: prompt.t("setup_language_pt")},
		{Value: string(i18n.English), Label: prompt.t("setup_language_en")},
	}, current.Language)
	if err != nil {
		return current, err
	}
	current.Language = string(i18n.Normalize(language))
	prompt.language = current.UILanguage()

	format, _, err := prompt.choose("setup_format_title", []setupOption{
		{Value: "conventional", Label: prompt.t("setup_format_conventional")},
		{Value: "gitmoji", Label: prompt.t("setup_format_gitmoji")},
	}, current.Format)
	if err != nil {
		return current, err
	}
	current.Format = format

	autoConfirm, _, err := prompt.choose("setup_auto_confirm_title", boolOptions(&prompt), strconv.FormatBool(current.AutoConfirm))
	if err != nil {
		return current, err
	}
	current.AutoConfirm = autoConfirm == "true"

	pushMode, _, err := prompt.choose("push_title", []setupOption{
		{Value: PushAlways, Label: prompt.t("setup_push_always")},
		{Value: PushAsk, Label: prompt.t("setup_push_ask")},
		{Value: PushNever, Label: prompt.t("setup_push_never")},
	}, NormalizePushMode(current.PushMode))
	if err != nil {
		return current, err
	}
	current.PushMode = pushMode

	customPrompt, _, err := prompt.choose("setup_custom_prompt_title", customPromptOptions(&prompt), strconv.FormatBool(current.UseCustomPrompt))
	if err != nil {
		return current, err
	}
	current.UseCustomPrompt = customPrompt == "true"

	previousProvider, previousEndpoint := current.Provider, current.OpenAIBaseURL
	provider, _, err := prompt.choose("setup_provider_title", []setupOption{
		{Value: "gemini", Label: prompt.t("setup_provider_gemini")},
		{Value: "openai", Label: prompt.t("setup_provider_openai")},
	}, current.Provider)
	if err != nil {
		return current, err
	}
	current.Provider = provider

	preset := endpointPresetForURL(current.OpenAIBaseURL)
	endpointChanged := false
	if current.Provider == "openai" {
		preset, endpointChanged, err = prompt.choose("setup_endpoint_title", endpointOptions(&prompt), preset)
		if err != nil {
			return current, err
		}
		if endpointChanged {
			if preset == "custom" {
				url, entered, readErr := prompt.input(prompt.t("setup_custom_url"))
				if readErr != nil {
					return current, readErr
				}
				if entered {
					current.OpenAIBaseURL = url
				}
			} else {
				current.OpenAIBaseURL = endpointURL(preset)
			}
		}
		preset = endpointPresetForURL(current.OpenAIBaseURL)
	}

	if previousProvider != current.Provider || (current.Provider == "openai" && previousEndpoint != current.OpenAIBaseURL) {
		current.Model = defaultModel(current.Provider, preset)
	}
	model, _, err := prompt.choose("setup_model_title", modelOptions(&prompt, current.Provider, preset), current.Model)
	if err != nil {
		return current, err
	}
	if model == "custom" {
		model, entered, readErr := prompt.input(prompt.t("setup_custom_model"))
		if readErr != nil {
			return current, readErr
		}
		if entered {
			current.Model = model
		}
	} else {
		current.Model = model
	}

	if err := prompt.configureKey(&current); err != nil {
		return current, err
	}
	if err := Save(current); err != nil {
		return current, err
	}
	path, err := Path()
	if err != nil {
		return current, err
	}
	prompt.complete(path)
	return current, nil
}

func boolOptions(prompt *setupPrompter) []setupOption {
	return []setupOption{
		{Value: "true", Label: prompt.t("setup_yes")},
		{Value: "false", Label: prompt.t("setup_no")},
	}
}

func customPromptOptions(prompt *setupPrompter) []setupOption {
	return []setupOption{
		{Value: "true", Label: prompt.t("setup_custom_yes")},
		{Value: "false", Label: prompt.t("setup_custom_no")},
	}
}

func (p *setupPrompter) banner() {
	fmt.Fprintln(p.out, setupDivider)
	fmt.Fprintln(p.out, p.t("setup_title"))
	fmt.Fprintln(p.out, setupDivider)
}

func (p *setupPrompter) complete(path string) {
	fmt.Fprintln(p.out)
	fmt.Fprintln(p.out, p.t("saved"))
	fmt.Fprintf(p.out, p.t("setup_config_path"), path)
	fmt.Fprintln(p.out, setupDivider)
	fmt.Fprintln(p.out, p.t("setup_complete"))
	fmt.Fprintln(p.out, setupDivider)
}

func (p *setupPrompter) choose(title string, options []setupOption, current string) (string, bool, error) {
	fmt.Fprintln(p.out)
	fmt.Fprintln(p.out, p.t(title))
	currentLabel := current
	for index, option := range options {
		fmt.Fprintf(p.out, "  %d. %s\n", index+1, option.Label)
		if option.Value == current {
			currentLabel = option.Label
		}
	}
	for {
		fmt.Fprintf(p.out, "%s [%s] (1-%d): ", p.t("setup_choose"), p.t("current", currentLabel), len(options))
		value, entered, err := p.input("")
		if err != nil {
			return current, false, err
		}
		if !entered {
			return current, false, nil
		}
		choice, err := strconv.Atoi(value)
		if err == nil && choice >= 1 && choice <= len(options) {
			return options[choice-1].Value, true, nil
		}
		fmt.Fprintf(p.out, p.t("setup_invalid_choice"), len(options))
	}
}

func (p *setupPrompter) input(label string) (string, bool, error) {
	if label != "" {
		fmt.Fprint(p.out, label+" ")
	}
	value, err := p.reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", false, err
	}
	value = strings.TrimSpace(value)
	return value, value != "", nil
}

func (p *setupPrompter) configureKey(cfg *Config) error {
	fmt.Fprintln(p.out)
	fmt.Fprintln(p.out, p.t("setup_api_key_title"))
	if cfg.Provider == "gemini" {
		p.keyStatus(cfg.GeminiAPIKey, "GEMINI_API_KEY", false)
		value, entered, err := p.input(p.t("setup_gemini_key"))
		if err != nil {
			return err
		}
		if entered {
			cfg.GeminiAPIKey = value
		}
		return nil
	}

	local := isLocalURL(cfg.OpenAIBaseURL)
	fmt.Fprintf(p.out, "%s\n", p.t("setup_compatible_key", cfg.OpenAIBaseURL))
	p.keyStatus(cfg.OpenAIAPIKey, configuredKeyVariable(cfg.OpenAIBaseURL), local)
	label := p.t("setup_api_key")
	if local {
		label = p.t("setup_local_api_key")
	}
	value, entered, err := p.input(label)
	if err != nil {
		return err
	}
	if entered {
		cfg.OpenAIAPIKey = value
	}
	return nil
}

func (p *setupPrompter) keyStatus(stored, environmentVariable string, optional bool) {
	switch {
	case stored != "":
		fmt.Fprintf(p.out, "%s ****%s\n", p.t("setup_key_current"), lastFour(stored))
	case environmentVariable != "" && os.Getenv(environmentVariable) != "":
		fmt.Fprintf(p.out, p.t("setup_key_environment"), environmentVariable)
	case optional:
		fmt.Fprintln(p.out, p.t("setup_key_optional"))
	default:
		fmt.Fprintln(p.out, p.t("setup_key_missing"))
	}
}

func (p *setupPrompter) t(key string, values ...any) string {
	return i18n.T(p.language, key, values...)
}

func lastFour(value string) string {
	if len(value) <= 4 {
		return ""
	}
	return value[len(value)-4:]
}

func configuredKeyVariable(baseURL string) string {
	for _, name := range compatibleKeyVariables(baseURL) {
		if os.Getenv(name) != "" {
			return name
		}
	}
	return ""
}

func endpointOptions(prompt *setupPrompter) []setupOption {
	return []setupOption{
		{Value: "openai", Label: prompt.t("endpoint_openai")},
		{Value: "openrouter", Label: prompt.t("endpoint_openrouter")},
		{Value: "groq", Label: prompt.t("endpoint_groq")},
		{Value: "deepseek", Label: prompt.t("endpoint_deepseek")},
		{Value: "ollama", Label: prompt.t("endpoint_ollama")},
		{Value: "lmstudio", Label: prompt.t("endpoint_lmstudio")},
		{Value: "cerebras", Label: prompt.t("endpoint_cerebras")},
		{Value: "custom", Label: prompt.t("endpoint_custom")},
	}
}

func endpointURL(preset string) string {
	return map[string]string{
		"openai":     "https://api.openai.com/v1",
		"openrouter": "https://openrouter.ai/api/v1",
		"groq":       "https://api.groq.com/openai/v1",
		"deepseek":   "https://api.deepseek.com/v1",
		"ollama":     "http://localhost:11434/v1",
		"lmstudio":   "http://localhost:1234/v1",
		"cerebras":   "https://api.cerebras.ai/v1",
	}[preset]
}

func endpointPresetForURL(value string) string {
	value = strings.TrimRight(strings.ToLower(strings.TrimSpace(value)), "/")
	for _, preset := range []string{"openai", "openrouter", "groq", "deepseek", "ollama", "lmstudio", "cerebras"} {
		if value == endpointURL(preset) {
			return preset
		}
	}
	return "custom"
}

func defaultModel(provider, preset string) string {
	if provider == "gemini" {
		return "gemini-3-flash-preview"
	}
	if model := map[string]string{
		"openrouter": "meta-llama/llama-3.3-70b-instruct",
		"groq":       "llama-3.3-70b-versatile",
		"deepseek":   "deepseek-chat",
		"ollama":     "llama3.2",
		"lmstudio":   "llama3.2",
		"cerebras":   "gpt-oss-120b",
	}[preset]; model != "" {
		return model
	}
	return "gpt-4o-mini"
}

func modelOptions(prompt *setupPrompter, provider, preset string) []setupOption {
	if provider == "gemini" {
		return []setupOption{
			{Value: "gemini-3-flash-preview", Label: "gemini-3-flash-preview — " + prompt.t("model_recommended")},
			{Value: "gemini-2.5-flash", Label: "gemini-2.5-flash — " + prompt.t("model_fast")},
			{Value: "gemini-2.5-pro", Label: "gemini-2.5-pro — " + prompt.t("model_advanced")},
			{Value: "custom", Label: prompt.t("model_custom")},
		}
	}
	options := map[string][]setupOption{
		"openai": {
			{Value: "gpt-4o-mini", Label: "gpt-4o-mini — " + prompt.t("model_fast_recommended")},
			{Value: "gpt-4o", Label: "gpt-4o — " + prompt.t("model_advanced")},
			{Value: "gpt-5-mini", Label: "gpt-5-mini — " + prompt.t("model_reasoning")},
		},
		"openrouter": {
			{Value: "meta-llama/llama-3.3-70b-instruct", Label: "Llama 3.3 70B — " + prompt.t("model_recommended")},
			{Value: "openai/gpt-4o-mini", Label: "OpenAI GPT-4o mini"},
			{Value: "deepseek/deepseek-chat", Label: "DeepSeek Chat"},
		},
		"groq": {
			{Value: "llama-3.3-70b-versatile", Label: "Llama 3.3 70B — " + prompt.t("model_recommended")},
			{Value: "llama-3.1-8b-instant", Label: "Llama 3.1 8B — " + prompt.t("model_fast")},
		},
		"deepseek": {
			{Value: "deepseek-chat", Label: "deepseek-chat — " + prompt.t("model_recommended")},
			{Value: "deepseek-reasoner", Label: "deepseek-reasoner — " + prompt.t("model_reasoning")},
		},
		"ollama": {
			{Value: "llama3.2", Label: "llama3.2 — " + prompt.t("model_recommended")},
			{Value: "qwen2.5-coder:7b", Label: "qwen2.5-coder:7b"},
			{Value: "mistral", Label: "mistral"},
		},
		"lmstudio": {
			{Value: "llama3.2", Label: "llama3.2 — " + prompt.t("model_recommended")},
			{Value: "qwen2.5-coder:7b", Label: "qwen2.5-coder:7b"},
			{Value: "mistral", Label: "mistral"},
		},
		"cerebras": {
			{Value: "gpt-oss-120b", Label: "gpt-oss-120b — " + prompt.t("model_recommended")},
			{Value: "llama-3.3-70b", Label: "llama-3.3-70b"},
			{Value: "llama3.1-8b", Label: "llama3.1-8b — " + prompt.t("model_fast")},
		},
	}[preset]
	return append(options, setupOption{Value: "custom", Label: prompt.t("model_custom")})
}
