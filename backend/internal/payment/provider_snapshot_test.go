package payment

import (
	"strings"
	"testing"
)

func TestParseProviderSnapshot(t *testing.T) {
	t.Parallel()

	snapshot := ParseProviderSnapshot(map[string]any{
		"schema_version":       float64(2),
		"provider_instance_id": " 12 ",
		"provider_key":         TypeWxpay,
		"payment_mode":         " jsapi ",
		"merchant_app_id":      " wx-app ",
		"merchant_id":          " mch-1 ",
		"currency":             " CNY ",
	})
	if snapshot == nil {
		t.Fatal("expected snapshot")
	}
	if snapshot.SchemaVersion != 2 ||
		snapshot.ProviderInstanceID != "12" ||
		snapshot.ProviderKey != TypeWxpay ||
		snapshot.PaymentMode != "jsapi" ||
		snapshot.MerchantAppID != "wx-app" ||
		snapshot.MerchantID != "mch-1" ||
		snapshot.Currency != "CNY" {
		t.Fatalf("snapshot = %#v", snapshot)
	}
}

func TestParseProviderSnapshotIgnoresEmptySnapshot(t *testing.T) {
	t.Parallel()

	if snapshot := ParseProviderSnapshot(map[string]any{"config": "ignored"}); snapshot != nil {
		t.Fatalf("snapshot = %#v, want nil", snapshot)
	}
}

func TestValidateProviderSnapshotMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		snapshot    *ProviderSnapshot
		providerKey string
		metadata    map[string]string
		wantErr     string
	}{
		{
			name:        "wxpay accepts matching merchant metadata",
			providerKey: TypeWxpay,
			snapshot: &ProviderSnapshot{
				MerchantAppID: "wx-app-expected",
				MerchantID:    "mch-expected",
				Currency:      "CNY",
			},
			metadata: map[string]string{
				"appid":       "wx-app-expected",
				"mchid":       "mch-expected",
				"currency":    "cny",
				"trade_state": "SUCCESS",
			},
		},
		{
			name:        "wxpay rejects appid mismatch",
			providerKey: TypeWxpay,
			snapshot:    &ProviderSnapshot{MerchantAppID: "wx-app-expected"},
			metadata:    map[string]string{"appid": "wx-app-other"},
			wantErr:     "wxpay appid mismatch",
		},
		{
			name:        "stripe rejects currency mismatch",
			providerKey: TypeStripe,
			snapshot:    &ProviderSnapshot{Currency: "HKD"},
			metadata:    map[string]string{"currency": "USD"},
			wantErr:     "stripe currency mismatch",
		},
		{
			name:        "airwallex rejects failed status",
			providerKey: TypeAirwallex,
			snapshot:    &ProviderSnapshot{MerchantID: "acct-1", Currency: "USD"},
			metadata:    map[string]string{"account_id": "acct-1", "currency": "USD", "status": "FAILED"},
			wantErr:     "airwallex status mismatch",
		},
		{
			name:        "legacy snapshot without relevant fields is allowed",
			providerKey: TypeWxpay,
			snapshot:    &ProviderSnapshot{ProviderInstanceID: "9", ProviderKey: TypeWxpay},
			metadata:    map[string]string{"appid": "runtime-app", "mchid": "runtime-mch", "currency": "CNY"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateProviderSnapshotMetadata(tt.snapshot, tt.providerKey, tt.metadata)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
