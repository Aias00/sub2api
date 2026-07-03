//go:build unit

package payment

import "testing"

func TestEnabledVisibleMethodsForProvider(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		providerKey    string
		supportedTypes string
		want           []string
	}{
		{name: "official alipay empty support", providerKey: TypeAlipay, supportedTypes: "", want: []string{TypeAlipay}},
		{name: "official wxpay direct support", providerKey: TypeWxpay, supportedTypes: TypeWxpayDirect, want: []string{TypeWxpay}},
		{name: "easypay both", providerKey: TypeEasyPay, supportedTypes: "alipay_direct,wxpay", want: []string{TypeAlipay, TypeWxpay}},
		{name: "stripe none", providerKey: TypeStripe, supportedTypes: "card,link", want: nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := EnabledVisibleMethodsForProvider(tc.providerKey, tc.supportedTypes)
			if len(got) != len(tc.want) {
				t.Fatalf("len = %d, want %d (%v)", len(got), len(tc.want), got)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("got[%d] = %q, want %q (full=%v)", i, got[i], tc.want[i], got)
				}
			}
		})
	}
}

func TestFilterEnabledVisibleMethodSources(t *testing.T) {
	t.Parallel()

	sources := []VisibleMethodProviderSource{
		{ID: 1, ProviderKey: TypeAlipay, SupportedTypes: TypeAlipay, Enabled: true},
		{ID: 2, ProviderKey: TypeEasyPay, SupportedTypes: "alipay,wxpay", Enabled: true},
		{ID: 3, ProviderKey: TypeWxpay, SupportedTypes: TypeWxpay, Enabled: false},
		{ID: 4, ProviderKey: TypeStripe, SupportedTypes: "card,link", Enabled: true},
	}

	alipay := FilterEnabledVisibleMethodSources(sources, TypeAlipay)
	if len(alipay) != 2 || alipay[0].ID != 1 || alipay[1].ID != 2 {
		t.Fatalf("alipay sources = %+v, want ids 1,2", alipay)
	}

	wxpay := FilterEnabledVisibleMethodSources(sources, TypeWxpay)
	if len(wxpay) != 1 || wxpay[0].ID != 2 {
		t.Fatalf("wxpay sources = %+v, want id 2", wxpay)
	}
}

func TestVisibleMethodProviderKeySelection(t *testing.T) {
	t.Parallel()

	sources := []VisibleMethodProviderSource{
		{ID: 1, ProviderKey: TypeAlipay, SupportedTypes: TypeAlipay, Enabled: true},
		{ID: 2, ProviderKey: TypeEasyPay, SupportedTypes: TypeAlipay, Enabled: true},
		{ID: 3, ProviderKey: TypeEasyPay, SupportedTypes: TypeWxpay, Enabled: true},
	}

	keys := DistinctVisibleMethodProviderKeys(sources)
	if len(keys) != 2 || keys[0] != TypeAlipay || keys[1] != TypeEasyPay {
		t.Fatalf("keys = %+v, want [alipay easypay]", keys)
	}

	selected, ok := SelectVisibleMethodSourceByProviderKey(sources, "EasyPay")
	if !ok || selected.ID != 2 {
		t.Fatalf("selected = %+v, ok=%v; want id 2", selected, ok)
	}

	filtered := FilterVisibleMethodSourcesByProviderKey(sources, TypeAlipay, TypeEasyPay)
	if len(filtered) != 1 || filtered[0].ID != 2 {
		t.Fatalf("filtered = %+v, want id 2", filtered)
	}
}

func TestBuildVisibleMethodSourceAvailability(t *testing.T) {
	t.Parallel()

	got := BuildVisibleMethodSourceAvailability([]VisibleMethodProviderSource{
		{ProviderKey: TypeAlipay, SupportedTypes: "alipay"},
		{ProviderKey: TypeEasyPay, SupportedTypes: "wxpay_direct, alipay"},
		{ProviderKey: TypeWxpay, SupportedTypes: "wxpay_direct"},
	})
	if !got[VisibleMethodSourceOfficialAlipay] {
		t.Fatalf("expected %q to be available", VisibleMethodSourceOfficialAlipay)
	}
	if !got[VisibleMethodSourceEasyPayAlipay] {
		t.Fatalf("expected %q to be available", VisibleMethodSourceEasyPayAlipay)
	}
	if !got[VisibleMethodSourceOfficialWechat] {
		t.Fatalf("expected %q to be available", VisibleMethodSourceOfficialWechat)
	}
	if !got[VisibleMethodSourceEasyPayWechat] {
		t.Fatalf("expected %q to be available", VisibleMethodSourceEasyPayWechat)
	}
}

func TestApplyVisibleMethodRoutingToEnabledTypes(t *testing.T) {
	t.Parallel()

	base := []string{TypeAlipay, TypeWxpay, TypeStripe}
	vals := map[string]string{
		SettingPaymentVisibleMethodAlipayEnabled: "true",
		SettingPaymentVisibleMethodAlipaySource:  VisibleMethodSourceOfficialAlipay,
		SettingPaymentVisibleMethodWxpayEnabled:  "true",
		SettingPaymentVisibleMethodWxpaySource:   VisibleMethodSourceOfficialWechat,
	}
	available := map[string]bool{
		VisibleMethodSourceOfficialAlipay: true,
		VisibleMethodSourceOfficialWechat: false,
	}

	got := ApplyVisibleMethodRoutingToEnabledTypes(base, vals, available)
	want := []string{TypeAlipay, TypeStripe}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q (full=%v)", i, got[i], want[i], got)
		}
	}
}

func TestApplyVisibleMethodRoutingAddsConfiguredVisibleMethod(t *testing.T) {
	t.Parallel()

	got := ApplyVisibleMethodRoutingToEnabledTypes(
		[]string{TypeStripe},
		map[string]string{
			SettingPaymentVisibleMethodAlipayEnabled: "true",
			SettingPaymentVisibleMethodAlipaySource:  VisibleMethodSourceEasyPayAlipay,
		},
		map[string]bool{VisibleMethodSourceEasyPayAlipay: true},
	)
	want := []string{TypeStripe, TypeAlipay}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q (full=%v)", i, got[i], want[i], got)
		}
	}
}
