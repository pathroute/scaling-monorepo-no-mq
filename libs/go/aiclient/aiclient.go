// Package aiclient is the client services use to call LLMs through ai-gateway.
package aiclient

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Request is one completion request.
type Request struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

// Response is a completion and its token usage.
type Response struct {
	Text         string `json:"text"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

// ErrRetryable marks provider errors worth retrying (rate limits, overload).
var ErrRetryable = errors.New("aiclient: retryable")

// Provider completes prompts.
type Provider interface {
	Complete(ctx context.Context, req Request) (Response, error)
}

// Retry wraps a provider with exponential backoff on ErrRetryable.
type Retry struct {
	Next     Provider
	Attempts int
	Base     time.Duration
	Sleep    func(time.Duration) // time.Sleep outside tests
}

// Complete calls Next up to Attempts times.
func (r Retry) Complete(ctx context.Context, req Request) (Response, error) {
	sleep := r.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	var err error
	delay := r.Base
	for i := range max(r.Attempts, 1) {
		var resp Response
		resp, err = r.Next.Complete(ctx, req)
		if err == nil || !errors.Is(err, ErrRetryable) {
			return resp, err
		}
		if i < r.Attempts-1 {
			sleep(delay)
			delay *= 2
		}
	}
	return Response{}, err
}

// Fake is a deterministic provider for tests and local runs: it echoes the prompt.
type Fake struct{}

// Complete returns "echo: <prompt>", truncated to MaxTokens words.
func (Fake) Complete(_ context.Context, req Request) (Response, error) {
	words := strings.Fields(req.Prompt)
	in := len(words)
	if req.MaxTokens > 0 && len(words) > req.MaxTokens {
		words = words[:req.MaxTokens]
	}
	return Response{Text: "echo: " + strings.Join(words, " "), InputTokens: in, OutputTokens: len(words)}, nil
}
