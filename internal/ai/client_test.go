package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/jhowk14/commit-ai/v2/internal/config"
	gitctx "github.com/jhowk14/commit-ai/v2/internal/git"
)

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
