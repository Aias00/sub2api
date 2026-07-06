package identity

import "testing"

func TestParseMetadataUserID(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantDevice string
		wantAcct   string
		wantSess   string
		wantNew    bool
	}{
		{
			name:       "legacy_without_account_uuid",
			raw:        "user_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2_account__session_123e4567-e89b-12d3-a456-426614174000",
			wantDevice: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			wantSess:   "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:       "json_with_account_uuid",
			raw:        `{"device_id":"d61f76d0aabbccdd00112233445566778899aabbccddeeff0011223344556677","account_uuid":"550e8400-e29b-41d4-a716-446655440000","session_id":"c72554f2-1234-5678-abcd-123456789abc"}`,
			wantDevice: "d61f76d0aabbccdd00112233445566778899aabbccddeeff0011223344556677",
			wantAcct:   "550e8400-e29b-41d4-a716-446655440000",
			wantSess:   "c72554f2-1234-5678-abcd-123456789abc",
			wantNew:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMetadataUserID(tt.raw)
			if got == nil {
				t.Fatal("ParseMetadataUserID() = nil")
			}
			if got.DeviceID != tt.wantDevice || got.AccountUUID != tt.wantAcct || got.SessionID != tt.wantSess || got.IsNewFormat != tt.wantNew {
				t.Fatalf("ParseMetadataUserID() = %#v", got)
			}
		})
	}
}

func TestParseMetadataUserIDRejectsInvalidInputs(t *testing.T) {
	tests := []string{
		"",
		"   ",
		"not-a-valid-user-id",
		`{"device_id":}`,
		`{"account_uuid":"","session_id":"c72554f2-1234-5678-abcd-123456789abc"}`,
		`{"device_id":"d61f76d0aabbccdd00112233445566778899aabbccddeeff0011223344556677","account_uuid":""}`,
		"user_a1b2c3d4_account__session_123e4567-e89b-12d3-a456-426614174000",
	}
	for _, raw := range tests {
		if got := ParseMetadataUserID(raw); got != nil {
			t.Fatalf("ParseMetadataUserID(%q) = %#v, want nil", raw, got)
		}
	}
}

func TestFormatMetadataUserID(t *testing.T) {
	deviceID := "deadbeef00112233445566778899aabbccddeeff0011223344556677"
	if got := FormatMetadataUserID(deviceID, "acc-uuid", "sess-uuid", "2.1.77"); got != "user_"+deviceID+"_account_acc-uuid_session_sess-uuid" {
		t.Fatalf("legacy FormatMetadataUserID() = %q", got)
	}
	if got := FormatMetadataUserID(deviceID, "acc-uuid", "sess-uuid", "2.1.78"); got != `{"device_id":"deadbeef00112233445566778899aabbccddeeff0011223344556677","account_uuid":"acc-uuid","session_id":"sess-uuid"}` {
		t.Fatalf("json FormatMetadataUserID() = %q", got)
	}
}

func TestMetadataVersionRules(t *testing.T) {
	if IsNewMetadataFormatVersion("") {
		t.Fatal("empty version should use legacy format")
	}
	if IsNewMetadataFormatVersion("2.1.77") {
		t.Fatal("2.1.77 should use legacy format")
	}
	if !IsNewMetadataFormatVersion("2.1.78") {
		t.Fatal("2.1.78 should use JSON format")
	}
	if got := ExtractCLIVersion("claude-cli/2.1.22-beta"); got != "2.1.22" {
		t.Fatalf("ExtractCLIVersion() = %q", got)
	}
	if got := CompareVersions("v2.1.0", "2.1.0"); got != 0 {
		t.Fatalf("CompareVersions() = %d", got)
	}
}
