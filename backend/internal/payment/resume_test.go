//go:build unit

package payment

import (
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestNormalizeVisibleMethods(t *testing.T) {
	t.Parallel()

	got := NormalizeVisibleMethods([]string{
		"alipay_direct",
		"alipay",
		" wxpay_direct ",
		"wxpay",
		"stripe",
	})

	want := []string{"alipay", "wxpay", "stripe"}
	if len(got) != len(want) {
		t.Fatalf("NormalizeVisibleMethods len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("NormalizeVisibleMethods[%d] = %q, want %q (full=%v)", i, got[i], want[i], got)
		}
	}
}

func TestNormalizePaymentSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{name: "empty uses default", input: "", expect: PaymentSourceHostedRedirect},
		{name: "wechat alias normalized", input: "wechat_in_app", expect: PaymentSourceWechatInAppResume},
		{name: "canonical value preserved", input: PaymentSourceWechatInAppResume, expect: PaymentSourceWechatInAppResume},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizePaymentSource(tt.input); got != tt.expect {
				t.Fatalf("NormalizePaymentSource(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestCanonicalizeReturnURL(t *testing.T) {
	t.Parallel()

	got, err := CanonicalizeReturnURL("https://example.com/payment/result?b=2#a", "example.com", "")
	if err != nil {
		t.Fatalf("CanonicalizeReturnURL returned error: %v", err)
	}
	if got != "https://example.com/payment/result?b=2" {
		t.Fatalf("CanonicalizeReturnURL = %q, want %q", got, "https://example.com/payment/result?b=2")
	}
}

func TestCanonicalizeReturnURLRejectsRelativeURL(t *testing.T) {
	t.Parallel()

	if _, err := CanonicalizeReturnURL("/payment/result", "example.com", ""); err == nil {
		t.Fatal("CanonicalizeReturnURL should reject relative URLs")
	}
}

func TestCanonicalizeReturnURLAllowsConfiguredFrontendHost(t *testing.T) {
	t.Parallel()

	got, err := CanonicalizeReturnURL(
		"https://app.example.com/payment/result?from=checkout",
		"api.example.com",
		"https://app.example.com/purchase",
	)
	if err != nil {
		t.Fatalf("CanonicalizeReturnURL returned error: %v", err)
	}
	if got != "https://app.example.com/payment/result?from=checkout" {
		t.Fatalf("CanonicalizeReturnURL = %q, want %q", got, "https://app.example.com/payment/result?from=checkout")
	}
}

func TestBuildPaymentReturnURL(t *testing.T) {
	t.Parallel()

	got, err := BuildPaymentReturnURL("https://example.com/payment/result?from=checkout#fragment", 42, "sub2_42", "resume-token")
	if err != nil {
		t.Fatalf("BuildPaymentReturnURL returned error: %v", err)
	}

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("url.Parse returned error: %v", err)
	}
	if parsed.Fragment != "" {
		t.Fatalf("BuildPaymentReturnURL should strip fragments, got %q", parsed.Fragment)
	}
	query := parsed.Query()
	if query.Get("from") != "checkout" {
		t.Fatalf("expected original query to be preserved, got %q", query.Get("from"))
	}
	if query.Get("order_id") != strconv.FormatInt(42, 10) {
		t.Fatalf("order_id = %q", query.Get("order_id"))
	}
	if query.Get("out_trade_no") != "sub2_42" {
		t.Fatalf("out_trade_no = %q", query.Get("out_trade_no"))
	}
	if query.Get("resume_token") != "resume-token" {
		t.Fatalf("resume_token = %q", query.Get("resume_token"))
	}
	if query.Get("status") != "success" {
		t.Fatalf("status = %q", query.Get("status"))
	}
}

func TestResumeTokenRoundTrip(t *testing.T) {
	t.Parallel()

	svc := NewResumeService([]byte("0123456789abcdef0123456789abcdef"))
	token, err := svc.CreateToken(ResumeTokenClaims{
		OrderID:            42,
		UserID:             7,
		ProviderInstanceID: "19",
		ProviderKey:        "easypay",
		PaymentType:        "wxpay",
		CanonicalReturnURL: "https://example.com/payment/result",
		IssuedAt:           1234567890,
	})
	if err != nil {
		t.Fatalf("CreateToken returned error: %v", err)
	}

	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}
	if claims.OrderID != 42 || claims.UserID != 7 {
		t.Fatalf("claims mismatch: %+v", claims)
	}
	if claims.ProviderInstanceID != "19" || claims.ProviderKey != "easypay" || claims.PaymentType != "wxpay" {
		t.Fatalf("claims provider snapshot mismatch: %+v", claims)
	}
}

func TestParseTokenRejectsExpiredToken(t *testing.T) {
	t.Parallel()

	svc := NewResumeService([]byte("0123456789abcdef0123456789abcdef"))
	token, err := svc.CreateToken(ResumeTokenClaims{
		OrderID:   42,
		UserID:    7,
		IssuedAt:  time.Now().Add(-25 * time.Hour).Unix(),
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("CreateToken returned error: %v", err)
	}

	_, err = svc.ParseToken(token)
	if err == nil {
		t.Fatal("ParseToken should reject expired tokens")
	}
}

func TestWeChatPaymentResumeTokenRoundTrip(t *testing.T) {
	t.Parallel()

	svc := NewResumeService([]byte("0123456789abcdef0123456789abcdef"))
	token, err := svc.CreateWeChatPaymentResumeToken(WeChatPaymentResumeClaims{
		OpenID:      "openid-123",
		PaymentType: TypeWxpay,
		Amount:      "12.50",
		OrderType:   OrderTypeSubscription,
		PlanID:      7,
		RedirectTo:  "/purchase?from=wechat",
		Scope:       "snsapi_base",
		IssuedAt:    1234567890,
	})
	if err != nil {
		t.Fatalf("CreateWeChatPaymentResumeToken returned error: %v", err)
	}

	claims, err := svc.ParseWeChatPaymentResumeToken(token)
	if err != nil {
		t.Fatalf("ParseWeChatPaymentResumeToken returned error: %v", err)
	}
	if claims.OpenID != "openid-123" || claims.PaymentType != TypeWxpay {
		t.Fatalf("claims mismatch: %+v", claims)
	}
	if claims.Amount != "12.50" || claims.OrderType != OrderTypeSubscription || claims.PlanID != 7 {
		t.Fatalf("claims payment context mismatch: %+v", claims)
	}
	if claims.RedirectTo != "/purchase?from=wechat" || claims.Scope != "snsapi_base" {
		t.Fatalf("claims redirect/scope mismatch: %+v", claims)
	}
}

func TestNormalizeVisibleMethodSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		input  string
		want   string
	}{
		{name: "alipay official alias", method: TypeAlipay, input: "alipay", want: VisibleMethodSourceOfficialAlipay},
		{name: "alipay easypay alias", method: TypeAlipay, input: "easypay", want: VisibleMethodSourceEasyPayAlipay},
		{name: "wxpay official alias", method: TypeWxpay, input: "wxpay", want: VisibleMethodSourceOfficialWechat},
		{name: "wxpay easypay alias", method: TypeWxpay, input: "easypay", want: VisibleMethodSourceEasyPayWechat},
		{name: "unsupported source", method: TypeWxpay, input: "stripe", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeVisibleMethodSource(tt.method, tt.input); got != tt.want {
				t.Fatalf("NormalizeVisibleMethodSource(%q, %q) = %q, want %q", tt.method, tt.input, got, tt.want)
			}
		})
	}
}

func TestVisibleMethodProviderKeyForSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		source string
		want   string
		ok     bool
	}{
		{name: "official alipay", method: TypeAlipay, source: VisibleMethodSourceOfficialAlipay, want: TypeAlipay, ok: true},
		{name: "easypay alipay", method: TypeAlipay, source: VisibleMethodSourceEasyPayAlipay, want: TypeEasyPay, ok: true},
		{name: "official wechat", method: TypeWxpay, source: VisibleMethodSourceOfficialWechat, want: TypeWxpay, ok: true},
		{name: "easypay wechat", method: TypeWxpay, source: VisibleMethodSourceEasyPayWechat, want: TypeEasyPay, ok: true},
		{name: "mismatched method and source", method: TypeAlipay, source: VisibleMethodSourceOfficialWechat, want: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := VisibleMethodProviderKeyForSource(tt.method, tt.source)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("VisibleMethodProviderKeyForSource(%q, %q) = (%q, %v), want (%q, %v)", tt.method, tt.source, got, ok, tt.want, tt.ok)
			}
		})
	}
}
