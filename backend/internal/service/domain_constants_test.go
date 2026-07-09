//go:build unit

package service

import "testing"

// TestSettingKeyDefaultPlatformQuotas 验证新的系统层 JSON key 常量值正确。
func TestSettingKeyDefaultPlatformQuotas(t *testing.T) {
	if SettingKeyDefaultPlatformQuotas != "default_platform_quotas" {
		t.Errorf("SettingKeyDefaultPlatformQuotas = %q, want %q",
			SettingKeyDefaultPlatformQuotas, "default_platform_quotas")
	}
}

// TestSettingKeyAuthSourcePlatformQuotas 验证新的 auth-source JSON key 函数返回值正确。
func TestSettingKeyAuthSourcePlatformQuotas(t *testing.T) {
	if got := SettingKeyAuthSourcePlatformQuotas("email"); got != "auth_source_default_email_platform_quotas" {
		t.Fatalf("got %q, want %q", got, "auth_source_default_email_platform_quotas")
	}
	if got := SettingKeyAuthSourcePlatformQuotas("github"); got != "auth_source_default_github_platform_quotas" {
		t.Fatalf("got %q, want %q", got, "auth_source_default_github_platform_quotas")
	}
}

// TestAffiliateRebateFreezeHoursDefault 验证返利冻结默认值为 72h。
// 默认部署下返利默认冻结 72h，给退款回扣留出冻结窗口，避免"B 充值→A 秒领返利→B 退款"净套现。
func TestAffiliateRebateFreezeHoursDefault(t *testing.T) {
	if AffiliateRebateFreezeHoursDefault != 72 {
		t.Fatalf("AffiliateRebateFreezeHoursDefault = %d, want 72", AffiliateRebateFreezeHoursDefault)
	}
	if AffiliateRebateFreezeHoursDefault > AffiliateRebateFreezeHoursMax {
		t.Fatalf("default %d exceeds max %d", AffiliateRebateFreezeHoursDefault, AffiliateRebateFreezeHoursMax)
	}
}
