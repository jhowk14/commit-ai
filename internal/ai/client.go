package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/jhowk14/commit-ai/v2/internal/config"
	gitctx "github.com/jhowk14/commit-ai/v2/internal/git"
)

type Input struct {
	Git             gitctx.Context
	Format          string
	Hint            string
	CustomPrompt    string
	UseCustomPrompt bool
}

type Client struct{ HTTPClient *http.Client }

func NewClient() Client { return Client{HTTPClient: &http.Client{Timeout: 45 * time.Second}} }

type gitmojiAlias struct {
	shortcode string
	emoji     string
}

// Keep the aliases most often emitted by OpenAI-compatible models here. The
// model is instructed to return Unicode, but accepting these aliases makes the
// command robust when a provider ignores that output-format instruction.
var gitmojiAliases = []gitmojiAlias{
	{":adhesive_bandage:", "🩹"}, {":airplane:", "✈️"}, {":alembic:", "⚗️"},
	{":alien:", "👽️"},
	{":ambulance:", "🚑"}, {":art:", "🎨"}, {":arrow_down:", "⬇️"},
	{":arrow_up:", "⬆️"}, {":beers:", "🍻"}, {":bookmark:", "🔖"},
	{":bento:", "🍱"}, {":boom:", "💥"}, {":bricks:", "🧱"}, {":bug:", "🐛"},
	{":building_construction:", "🏗️"}, {":bulb:", "💡"}, {":busts_in_silhouette:", "👥"},
	{":camera_flash:", "📸"}, {":card_file_box:", "🗃️"},
	{":chart_with_upwards_trend:", "📈"}, {":children_crossing:", "🚸"},
	{":closed_lock_with_key:", "🔐"}, {":clown_face:", "🤡"}, {":coffin:", "⚰️"},
	{":construction:", "🚧"}, {":construction_worker:", "👷"}, {":dizzy:", "💫"},
	{":egg:", "🥚"}, {":fire:", "🔥"}, {":globe_with_meridians:", "🌐"},
	{":green_heart:", "💚"}, {":hammer:", "🔨"}, {":heavy_minus_sign:", "➖"},
	{":heavy_plus_sign:", "➕"}, {":iphone:", "📱"}, {":label:", "🏷️"},
	{":lipstick:", "💄"}, {":lock:", "🔒"}, {":loud_sound:", "🔊"}, {":mag:", "🔍️"},
	{":memo:", "📝"}, {":monocle_face:", "🧐"}, {":money_with_wings:", "💸"},
	{":mute:", "🔇"}, {":necktie:", "👔"}, {":package:", "📦"},
	{":page_facing_up:", "📄"}, {":passport_control:", "🛂"}, {":pencil2:", "✏️"},
	{":poop:", "💩"}, {":pushpin:", "📌"}, {":recycle:", "♻️"}, {":rewind:", "⏪"},
	{":rocket:", "🚀"}, {":rotating_light:", "🚨"}, {":safety_vest:", "🦺"},
	{":seedling:", "🌱"}, {":see_no_evil:", "🙈"}, {":sparkles:", "✨"},
	{":speech_balloon:", "💬"}, {":stethoscope:", "🩺"}, {":t-rex:", "🦖"},
	{":tada:", "🎉"}, {":technologist:", "🧑‍💻"}, {":test_tube:", "🧪"},
	{":thread:", "🧵"}, {":triangular_flag_on_post:", "🚩"}, {":truck:", "🚚"},
	{":twisted_rightwards_arrows:", "🔀"}, {":warning:", "⚠️"}, {":wastebasket:", "🗑️"},
	{":wheelchair:", "♿️"}, {":white_check_mark:", "✅"}, {":wrench:", "🔧"},
	{":zap:", "⚡"}, {":goal_net:", "🥅"},
}

func (c Client) Generate(ctx context.Context, cfg config.Config, input Input) (string, error) {
	if err := cfg.ValidateProvider(); err != nil {
		return "", err
	}
	prompt := buildPrompt(input)
	if cfg.Provider == "gemini" {
		return c.generateGemini(ctx, cfg, prompt, input.Format)
	}
	return c.generateOpenAI(ctx, cfg, input, prompt)
}

func buildPrompt(input Input) string {
	if input.UseCustomPrompt && input.CustomPrompt != "" {
		return strings.NewReplacer("{HISTORY}", input.Git.History, "{FILES}", input.Git.Files, "{DIFF}", input.Git.Diff).Replace(input.CustomPrompt)
	}
	formatRules := "EXACT format: <type>: <message>. Choose one of feat, fix, docs, style, refactor, perf, test, build, ci, chore or revert. Message starts lowercase."
	if input.Format == "gitmoji" {
		formatRules = "You are a senior Git and Gitmoji expert. Choose ONLY ONE Gitmoji. EXACT format: <Unicode Gitmoji><space><Message>. The first character MUST be the actual Unicode emoji, never an alias or text such as :sparkles:, sparkles, feat:, or emoji:. Start the imperative English message with a capital letter."
	}
	hint := ""
	if input.Hint != "" {
		hint = "\nUser context: " + input.Hint + "\n"
	}
	return fmt.Sprintf("You are a Git commit-message generator.%s\nRecent history:\n%s\n\nStaged files:\n%s\n\nRelevant diff:\n%s\n\nRules: %s Use imperative English, at most 72 characters, no final period. Return only the final commit message.", hint, input.Git.History, input.Git.Files, input.Git.Diff, formatRules)
}

func compactCerebrasPrompt(input Input) string {
	rules := "Use exactly <type>: <message> with type feat, fix, docs, style, refactor, perf, test, build, ci, chore, or revert. Use imperative English and a lowercase message."
	if input.Format == "gitmoji" {
		rules = "Choose ONLY ONE Gitmoji. Use exactly <Unicode Gitmoji><space><imperative English message>. Start the message with a capital letter. Never use aliases or text such as :sparkles:, sparkles, feat:, or emoji:."
	}
	hint := ""
	if input.Hint != "" {
		hint = "\nUser context: " + input.Hint + "\n"
	}
	return fmt.Sprintf("Generate one commit message for the staged change.%s\nStaged files:\n%s\n\nRelevant diff:\n%s\n\nRules: %s Maximum 72 characters. No period. Return only the commit message.", hint, input.Git.Files, input.Git.Diff, rules)
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type openAIRequest struct {
	Model               string    `json:"model"`
	Messages            []message `json:"messages"`
	MaxTokens           *int      `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int      `json:"max_completion_tokens,omitempty"`
	ReasoningEffort     string    `json:"reasoning_effort,omitempty"`
	ReasoningFormat     string    `json:"reasoning_format,omitempty"`
	Temperature         *float64  `json:"temperature,omitempty"`
}
type openAIResponse struct {
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c Client) generateOpenAI(ctx context.Context, cfg config.Config, input Input, prompt string) (string, error) {
	baseURL := strings.TrimRight(cfg.OpenAIBaseURL, "/")
	endpoint := baseURL
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint += "/chat/completions"
	}
	isCerebrasReasoning := isCerebrasGPTOSS(baseURL, cfg.Model)
	request := openAIRequest{Model: cfg.Model, Messages: []message{{Role: "user", Content: prompt}}}
	if isCerebrasReasoning {
		limit, temperature := 512, 0.0
		request.Messages = []message{{Role: "system", Content: "Return only the final commit message. Do not include analysis or explanation."}, {Role: "user", Content: compactCerebrasPrompt(input)}}
		request.MaxCompletionTokens, request.ReasoningEffort, request.ReasoningFormat, request.Temperature = &limit, "low", "hidden", &temperature
	} else if usesCompletionTokenLimit(cfg.Model) {
		limit := 100
		request.MaxCompletionTokens = &limit
	} else {
		limit := 100
		request.MaxTokens = &limit
	}
	response, err := c.postOpenAI(ctx, endpoint, cfg.ResolvedOpenAIKey(), request)
	if err != nil {
		return "", err
	}
	if message, truncated := responseMessage(response); message != "" {
		return normalizeCommitMessage(message, input.Format), nil
	} else if !isCerebrasReasoning || !truncated {
		return "", responseError(response)
	}
	limit := 1024
	request.MaxCompletionTokens = &limit
	response, err = c.postOpenAI(ctx, endpoint, cfg.ResolvedOpenAIKey(), request)
	if err != nil {
		return "", err
	}
	if message, _ := responseMessage(response); message != "" {
		return normalizeCommitMessage(message, input.Format), nil
	}
	return "", responseError(response)
}

func isCerebrasGPTOSS(baseURL, model string) bool {
	return strings.Contains(strings.ToLower(baseURL), "cerebras.ai") && strings.HasPrefix(strings.ToLower(model), "gpt-oss-")
}

// usesCompletionTokenLimit covers current OpenAI reasoning and GPT-5 models.
// They reject max_tokens, while the broadly compatible Chat Completions APIs
// (including OpenAI's GPT-4 family and local providers) continue to accept it.
func usesCompletionTokenLimit(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(model, "gpt-5") ||
		strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3") ||
		strings.HasPrefix(model, "o4")
}

func (c Client) postOpenAI(ctx context.Context, endpoint, key string, payload openAIRequest) (openAIResponse, error) {
	var response openAIResponse
	body, err := json.Marshal(payload)
	if err != nil {
		return response, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return response, err
	}
	request.Header.Set("Content-Type", "application/json")
	if key != "" {
		request.Header.Set("Authorization", "Bearer "+key)
	} else {
		request.Header.Set("Authorization", "Bearer none")
	}
	if strings.Contains(endpoint, "openrouter.ai") {
		request.Header.Set("HTTP-Referer", "https://github.com/jhowk14/commit-ai")
		request.Header.Set("X-Title", "commit-ai")
	}
	httpResponse, err := c.HTTPClient.Do(request)
	if err != nil {
		return response, fmt.Errorf("erro ao chamar a API: %w", err)
	}
	defer httpResponse.Body.Close()
	if err := json.NewDecoder(io.LimitReader(httpResponse.Body, 2<<20)).Decode(&response); err != nil {
		return response, fmt.Errorf("resposta inválida da API: %w", err)
	}
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode > 299 {
		return response, responseError(response)
	}
	return response, nil
}

func responseMessage(response openAIResponse) (string, bool) {
	if len(response.Choices) == 0 {
		return "", false
	}
	choice := response.Choices[0]
	return strings.TrimSpace(choice.Message.Content), choice.FinishReason == "length"
}

func responseError(response openAIResponse) error {
	if response.Error.Message != "" {
		return errors.New(response.Error.Message)
	}
	if len(response.Choices) > 0 && response.Choices[0].FinishReason == "length" {
		return errors.New("o modelo atingiu o limite de saída antes de gerar a mensagem de commit")
	}
	return errors.New("a API não retornou uma mensagem de commit")
}

func (c Client) generateGemini(ctx context.Context, cfg config.Config, prompt, format string) (string, error) {
	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", url.PathEscape(cfg.Model))
	payload := map[string]any{"contents": []any{map[string]any{"parts": []any{map[string]string{"text": prompt}}}}}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-goog-api-key", cfg.ResolvedGeminiKey())
	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("erro ao chamar Gemini: %w", err)
	}
	defer response.Body.Close()
	var decoded struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&decoded); err != nil {
		return "", fmt.Errorf("resposta inválida do Gemini: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		if decoded.Error.Message != "" {
			return "", errors.New(decoded.Error.Message)
		}
		return "", fmt.Errorf("Gemini retornou HTTP %d", response.StatusCode)
	}
	if len(decoded.Candidates) == 0 || len(decoded.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("Gemini não retornou uma mensagem de commit")
	}
	return normalizeCommitMessage(decoded.Candidates[0].Content.Parts[0].Text, format), nil
}

func normalize(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "`\"")
	if line, _, found := strings.Cut(value, "\n"); found {
		value = line
	}
	return strings.TrimSpace(value)
}

func normalizeCommitMessage(value, format string) string {
	message := normalize(value)
	if format != "gitmoji" {
		return message
	}
	return normalizeGitmoji(message)
}

func normalizeGitmoji(value string) string {
	lower := strings.ToLower(value)
	for _, alias := range gitmojiAliases {
		if !strings.HasPrefix(lower, alias.shortcode) {
			continue
		}
		return formatGitmoji(alias.emoji, value[len(alias.shortcode):])
	}
	for _, alias := range gitmojiAliases {
		if strings.HasPrefix(value, alias.emoji) {
			return formatGitmoji(alias.emoji, strings.TrimPrefix(value, alias.emoji))
		}
	}
	return value
}

func formatGitmoji(emoji, message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return emoji
	}
	runes := []rune(message)
	runes[0] = unicode.ToUpper(runes[0])
	return emoji + " " + string(runes)
}
