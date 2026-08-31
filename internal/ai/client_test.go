package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/jhowk14/commit-ai/v2/internal/config"
	gitctx "github.com/jhowk14/commit-ai/v2/internal/git"
)

func TestPromptBuildersRespectFormatHintAndCustomTemplate(t *testing.T) {
	input := Input{Git: gitctx.Context{History: "abc", Files: "main.go", Diff: "+line"}, Format: "gitmoji", Hint: "improve setup"}
	prompt := buildPrompt(input)
	if !strings.Contains(prompt, "Gitmoji") || !strings.Contains(prompt, "improve setup") || !strings.Contains(prompt, "main.go") {
		t.Fatalf("prompt: %s", prompt)
	}
	input.UseCustomPrompt, input.CustomPrompt = true, "{FILES}|{DIFF}|{HISTORY}"
	if got := buildPrompt(input); got != "main.go|+line|abc" {
		t.Fatalf("custom prompt: %q", got)
	}
	compact := compactCerebrasPrompt(input)
	if strings.Contains(compact, "abc") || !strings.Contains(compact, "Gitmoji") {
		t.Fatalf("prompt compacto: %s", compact)
	}
}

func TestModelTokenLimitProfiles(t *testing.T) {
	for _, testCase := range []struct {
		model string
		want  bool
	}{
		{"gpt-4o-mini", false},
		{"gpt-5-mini", true},
		{"o1", true},
		{"o3-mini", true},
		{"o4-mini", true},
		{"deepseek-chat", false},
	} {
		if got := usesCompletionTokenLimit(testCase.model); got != testCase.want {
			t.Fatalf("modelo %s: got %t, want %t", testCase.model, got, testCase.want)
		}
	}
	if !isCerebrasGPTOSS("https://api.cerebras.ai/v1", "GPT-OSS-120B") {
		t.Fatal("perfil GPT-OSS do Cerebras não identificado")
	}
	if isCerebrasGPTOSS("https://api.openai.com/v1", "gpt-oss-120b") {
		t.Fatal("perfil GPT-OSS não deve ser aplicado fora do Cerebras")
	}
}

func TestCerebrasRetriesOnlyAfterTruncatedResponse(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var payload openAIRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		call := calls.Add(1)
		if payload.ReasoningEffort != "low" || payload.ReasoningFormat != "hidden" || payload.MaxCompletionTokens == nil {
			t.Fatalf("payload Cerebras incorreto: %#v", payload)
		}
		if call == 1 {
			if *payload.MaxCompletionTokens != 512 {
				t.Fatalf("limite inicial: %d", *payload.MaxCompletionTokens)
			}
			_, _ = writer.Write([]byte(`{"choices":[{"finish_reason":"length","message":{"content":""}}]}`))
			return
		}
		if *payload.MaxCompletionTokens != 1024 {
			t.Fatalf("limite de retry: %d", *payload.MaxCompletionTokens)
		}
		_, _ = writer.Write([]byte(`{"choices":[{"finish_reason":"stop","message":{"content":"fix: recover response"}}]}`))
	}))
	defer server.Close()
	cfg := config.Default()
	cfg.Provider, cfg.Model, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey = "openai", "gpt-oss-120b", server.URL+"/cerebras.ai", "test"
	message, err := NewClient().Generate(context.Background(), cfg, Input{Git: gitctx.Context{Files: "a.go", Diff: "+change"}, Format: "conventional"})
	if err != nil {
		t.Fatal(err)
	}
	if message != "fix: recover response" || calls.Load() != 2 {
		t.Fatalf("resultado=%q chamadas=%d", message, calls.Load())
	}
}

func TestOpenAICompatibleUsesOneRequest(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		_, _ = writer.Write([]byte(`{"choices":[{"finish_reason":"stop","message":{"content":"feat: add support"}}]}`))
	}))
	defer server.Close()
	cfg := config.Default()
	cfg.Provider, cfg.Model, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey = "openai", "model", server.URL, "test"
	message, err := NewClient().Generate(context.Background(), cfg, Input{Git: gitctx.Context{Files: "a.go", Diff: "+change"}, Format: "conventional"})
	if err != nil {
		t.Fatal(err)
	}
	if message != "feat: add support" || calls.Load() != 1 {
		t.Fatalf("resultado=%q chamadas=%d", message, calls.Load())
	}
}

func TestMainOpenAICompatibleProviderContracts(t *testing.T) {
	testCases := []struct {
		name, baseURL, model, key, authorization string
		openRouter, cerebras                     bool
	}{
		{"OpenAI", "https://api.openai.com/v1", "gpt-4o-mini", "openai-key", "Bearer openai-key", false, false},
		{"OpenAI GPT-5", "https://api.openai.com/v1", "gpt-5-mini", "openai-key", "Bearer openai-key", false, false},
		{"OpenRouter", "https://openrouter.ai/api/v1", "openai/gpt-4o-mini", "router-key", "Bearer router-key", true, false},
		{"Groq", "https://api.groq.com/openai/v1", "llama-3.3-70b-versatile", "groq-key", "Bearer groq-key", false, false},
		{"DeepSeek", "https://api.deepseek.com/v1", "deepseek-chat", "deepseek-key", "Bearer deepseek-key", false, false},
		{"Ollama", "http://localhost:11434/v1", "llama3.2", "", "Bearer none", false, false},
		{"LM Studio", "http://127.0.0.1:1234/v1", "local-model", "", "Bearer none", false, false},
		{"Cerebras", "https://api.cerebras.ai/v1", "gpt-oss-120b", "cerebras-key", "Bearer cerebras-key", false, true},
		{"Cerebras Gemma", "https://api.cerebras.ai/v1", "gemma-4-31b", "cerebras-key", "Bearer cerebras-key", false, false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/v1/chat/completions" && request.URL.Path != "/api/v1/chat/completions" && request.URL.Path != "/openai/v1/chat/completions" && request.URL.Path != "/chat/completions" {
					t.Fatalf("rota: %s", request.URL.Path)
				}
				if got := request.Header.Get("Authorization"); got != testCase.authorization {
					t.Fatalf("autorização: %q", got)
				}
				if testCase.openRouter && (request.Header.Get("HTTP-Referer") == "" || request.Header.Get("X-Title") != "commit-ai") {
					t.Fatal("cabeçalhos OpenRouter ausentes")
				}
				var payload openAIRequest
				if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
					t.Fatal(err)
				}
				if testCase.cerebras {
					if payload.MaxCompletionTokens == nil || *payload.MaxCompletionTokens != 512 || payload.ReasoningEffort != "low" {
						t.Fatalf("payload Cerebras: %#v", payload)
					}
				} else if testCase.model == "gpt-5-mini" {
					if payload.MaxCompletionTokens == nil || *payload.MaxCompletionTokens != 100 || payload.MaxTokens != nil {
						t.Fatalf("payload GPT-5: %#v", payload)
					}
				} else if payload.MaxTokens == nil || *payload.MaxTokens != 100 {
					t.Fatalf("payload compatível: %#v", payload)
				}
				_, _ = writer.Write([]byte(`{"choices":[{"finish_reason":"stop","message":{"content":"fix: validate provider"}}]}`))
			}))
			defer server.Close()
			client := NewClient()
			transport := server.Client().Transport
			client.HTTPClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
				clone := request.Clone(request.Context())
				clone.URL.Scheme, clone.URL.Host = "http", strings.TrimPrefix(server.URL, "http://")
				return transport.RoundTrip(clone)
			})
			cfg := config.Default()
			cfg.Provider, cfg.Model, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey = "openai", testCase.model, testCase.baseURL, testCase.key
			message, err := client.Generate(context.Background(), cfg, Input{Git: gitctx.Context{Files: "a.go", Diff: "+change"}, Format: "conventional"})
			if err != nil {
				t.Fatal(err)
			}
			if message != "fix: validate provider" {
				t.Fatalf("mensagem: %q", message)
			}
		})
	}
}

func TestGeminiAndResponseFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("x-goog-api-key") != "gemini-key" {
			t.Fatal("chave Gemini ausente")
		}
		_, _ = writer.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"fix: support Gemini\nextra"}]}}]}`))
	}))
	defer server.Close()
	serverTransport := server.Client().Transport
	client := NewClient()
	// Intercepta a URL oficial no transporte para exercitar o caminho Gemini.
	client.HTTPClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		clone := request.Clone(request.Context())
		clone.URL.Scheme, clone.URL.Host = "http", strings.TrimPrefix(server.URL, "http://")
		return serverTransport.RoundTrip(clone)
	})
	cfg := config.Default()
	cfg.Provider, cfg.GeminiAPIKey = "gemini", "gemini-key"
	message, err := client.Generate(context.Background(), cfg, Input{Git: gitctx.Context{Files: "a", Diff: "+b"}, Format: "conventional"})
	if err != nil {
		t.Fatal(err)
	}
	if message != "fix: support Gemini" {
		t.Fatalf("mensagem Gemini: %q", message)
	}
	if _, err := client.Generate(context.Background(), cfg, Input{}); err != nil {
		t.Fatalf("entrada vazia ainda é válida: %v", err)
	}
	if err := responseError(openAIResponse{}); err == nil {
		t.Fatal("resposta vazia deveria falhar")
	}
	if got := normalize("```fix: test\nextra"); got != "fix: test" {
		t.Fatalf("normalização: %q", got)
	}
}

func TestOpenAIAndGeminiReturnUsefulErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(`{"error":{"message":"modelo inválido"}}`))
	}))
	defer server.Close()
	cfg := config.Default()
	cfg.Provider, cfg.Model, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey = "openai", "model", server.URL, "test"
	if _, err := NewClient().Generate(context.Background(), cfg, Input{}); err == nil || !strings.Contains(err.Error(), "modelo inválido") {
		t.Fatalf("erro OpenAI: %v", err)
	}
	server.Config.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(`{"error":{"message":"falha Gemini"}}`))
	})
	client := NewClient()
	transport := server.Client().Transport
	client.HTTPClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		clone := request.Clone(request.Context())
		clone.URL.Scheme, clone.URL.Host = "http", strings.TrimPrefix(server.URL, "http://")
		return transport.RoundTrip(clone)
	})
	cfg = config.Default()
	cfg.Provider, cfg.GeminiAPIKey = "gemini", "key"
	if _, err := client.Generate(context.Background(), cfg, Input{}); err == nil || !strings.Contains(err.Error(), "falha Gemini") {
		t.Fatalf("erro Gemini: %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return fn(request) }
