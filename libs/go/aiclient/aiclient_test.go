package aiclient

import (
	"context"
	"errors"
	"testing"
	"time"
)

type flaky struct {
	fails int
	err   error
	calls int
}

func (f *flaky) Complete(ctx context.Context, req Request) (Response, error) {
	f.calls++
	if f.calls <= f.fails {
		return Response{}, f.err
	}
	return Fake{}.Complete(ctx, req)
}

func TestRetryBacksOff(t *testing.T) {
	f := &flaky{fails: 2, err: ErrRetryable}
	var slept []time.Duration
	r := Retry{Next: f, Attempts: 3, Base: time.Second, Sleep: func(d time.Duration) { slept = append(slept, d) }}
	resp, err := r.Complete(context.Background(), Request{Prompt: "hi"})
	if err != nil || resp.Text != "echo: hi" || f.calls != 3 {
		t.Fatalf("got %+v %v after %d calls", resp, err, f.calls)
	}
	if len(slept) != 2 || slept[0] != time.Second || slept[1] != 2*time.Second {
		t.Fatalf("backoff: %v", slept)
	}
}

func TestRetryGivesUp(t *testing.T) {
	f := &flaky{fails: 5, err: ErrRetryable}
	r := Retry{Next: f, Attempts: 2, Sleep: func(time.Duration) {}}
	if _, err := r.Complete(context.Background(), Request{}); !errors.Is(err, ErrRetryable) || f.calls != 2 {
		t.Fatalf("err %v after %d calls", err, f.calls)
	}
}

func TestRetryStopsOnPermanentError(t *testing.T) {
	f := &flaky{fails: 5, err: errors.New("bad request")}
	r := Retry{Next: f, Attempts: 3, Sleep: func(time.Duration) {}}
	if _, err := r.Complete(context.Background(), Request{}); err == nil || f.calls != 1 {
		t.Fatalf("err %v after %d calls", err, f.calls)
	}
}

func TestFakeTruncates(t *testing.T) {
	resp, _ := Fake{}.Complete(context.Background(), Request{Prompt: "a b c d", MaxTokens: 2})
	if resp.Text != "echo: a b" || resp.InputTokens != 4 || resp.OutputTokens != 2 {
		t.Fatalf("got %+v", resp)
	}
}
