package service

import "github.com/Aias00/cloudbase/internal/identity"

// NewMetadataFormatMinVersion is the minimum Claude Code version that uses
// JSON-formatted metadata.user_id instead of the legacy concatenated string.
const NewMetadataFormatMinVersion = "2.1.78"

// ParsedUserID represents the components extracted from a metadata.user_id value.
type ParsedUserID struct {
	DeviceID    string // 64-char hex (or arbitrary client id)
	AccountUUID string // may be empty
	SessionID   string // UUID
	IsNewFormat bool   // true if the original was JSON format
}

// ParseMetadataUserID parses a metadata.user_id string in either format.
// Returns nil if the input cannot be parsed.
func ParseMetadataUserID(raw string) *ParsedUserID {
	parsed := identity.ParseMetadataUserID(raw)
	if parsed == nil {
		return nil
	}
	return &ParsedUserID{
		DeviceID:    parsed.DeviceID,
		AccountUUID: parsed.AccountUUID,
		SessionID:   parsed.SessionID,
		IsNewFormat: parsed.IsNewFormat,
	}
}

// FormatMetadataUserID builds a metadata.user_id string in the format
// appropriate for the given CLI version. Components are the rewritten values
// (not necessarily the originals).
func FormatMetadataUserID(deviceID, accountUUID, sessionID, uaVersion string) string {
	return identity.FormatMetadataUserID(deviceID, accountUUID, sessionID, uaVersion)
}

// IsNewMetadataFormatVersion returns true if the given CLI version uses the
// new JSON metadata.user_id format (>= 2.1.78).
func IsNewMetadataFormatVersion(version string) bool {
	return identity.IsNewMetadataFormatVersion(version)
}

// ExtractCLIVersion extracts the Claude Code version from a User-Agent string.
// Returns "" if the UA doesn't match the expected pattern.
func ExtractCLIVersion(ua string) string {
	return identity.ExtractCLIVersion(ua)
}
