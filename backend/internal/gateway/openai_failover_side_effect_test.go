package gateway

import "testing"

func TestBuildOpenAIUpstreamFailoverSideEffect(t *testing.T) {
	effect := BuildOpenAIUpstreamFailoverSideEffect(OpenAIUpstreamFailoverSideEffectInput{
		Platform:     "openai",
		AccountID:    42,
		AccountName:  "acct",
		StatusCode:   429,
		RequestID:    "req-1",
		Message:      "rate limited",
		ResponseBody: []byte("body"),
		Decision: OpenAIUpstreamFailureDecision{
			RetryableOnSameAccount: true,
			Detail:                 "decision detail",
		},
		DefaultDetail: "fallback",
	})
	if effect.Platform != "openai" || effect.AccountID != 42 || effect.StatusCode != 429 {
		t.Fatalf("unexpected effect: %+v", effect)
	}
	if effect.Detail != "decision detail" {
		t.Fatalf("detail = %q, want decision detail", effect.Detail)
	}
	if !effect.RetryableOnSameAccount {
		t.Fatal("retryable flag should be preserved")
	}
}

func TestBuildOpenAIUpstreamFailoverSideEffectUsesDefaultDetail(t *testing.T) {
	effect := BuildOpenAIUpstreamFailoverSideEffect(OpenAIUpstreamFailoverSideEffectInput{
		DefaultDetail: "fallback",
	})
	if effect.Detail != "fallback" {
		t.Fatalf("detail = %q, want fallback", effect.Detail)
	}
}
