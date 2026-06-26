package service

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

const legacySMTPChannelID = "primary"

// SMTPChannelConfig is persisted in settings as JSON for additional SMTP senders.
type SMTPChannelConfig struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Enabled            bool   `json:"enabled"`
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Username           string `json:"username"`
	Password           string `json:"password,omitempty"`
	PasswordConfigured bool   `json:"password_configured,omitempty"`
	From               string `json:"from_email"`
	FromName           string `json:"from_name"`
	UseTLS             bool   `json:"use_tls"`
	DailyLimit         int    `json:"daily_limit"`
	SortOrder          int    `json:"sort_order"`
}

func parseSMTPDailyLimit(raw string) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func normalizeSMTPDailyLimit(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func parseSMTPChannels(raw string) []SMTPChannelConfig {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var channels []SMTPChannelConfig
	if err := json.Unmarshal([]byte(raw), &channels); err != nil {
		return nil
	}
	return normalizeSMTPChannels(channels)
}

func marshalSMTPChannels(channels []SMTPChannelConfig) (string, error) {
	normalized := normalizeSMTPChannels(channels)
	if len(normalized) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func normalizeSMTPChannels(channels []SMTPChannelConfig) []SMTPChannelConfig {
	if len(channels) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(channels))
	normalized := make([]SMTPChannelConfig, 0, len(channels))
	for i, channel := range channels {
		channel.ID = normalizeSMTPChannelID(channel.ID, i)
		if _, exists := seen[channel.ID]; exists {
			channel.ID = normalizeSMTPChannelID("", i)
		}
		seen[channel.ID] = struct{}{}
		channel.Name = strings.TrimSpace(channel.Name)
		channel.Host = strings.TrimSpace(channel.Host)
		channel.Username = strings.TrimSpace(channel.Username)
		channel.Password = strings.TrimSpace(channel.Password)
		channel.From = strings.TrimSpace(channel.From)
		channel.FromName = strings.TrimSpace(channel.FromName)
		if channel.Port <= 0 {
			channel.Port = 587
		}
		channel.DailyLimit = normalizeSMTPDailyLimit(channel.DailyLimit)
		if channel.SortOrder <= 0 {
			channel.SortOrder = i + 1
		}
		channel.PasswordConfigured = channel.Password != ""
		normalized = append(normalized, channel)
	}

	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].SortOrder == normalized[j].SortOrder {
			return normalized[i].ID < normalized[j].ID
		}
		return normalized[i].SortOrder < normalized[j].SortOrder
	})
	return normalized
}

func normalizeSMTPChannelID(raw string, index int) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	var b strings.Builder
	lastDash := false
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			_, _ = b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			_ = b.WriteByte('-')
			lastDash = true
		}
	}
	id := strings.Trim(b.String(), "-")
	if id == "" || id == legacySMTPChannelID {
		return "channel-" + strconv.Itoa(index+1)
	}
	return id
}

func CopySMTPChannelsForAdmin(channels []SMTPChannelConfig) []SMTPChannelConfig {
	if len(channels) == 0 {
		return nil
	}
	out := make([]SMTPChannelConfig, 0, len(channels))
	for _, channel := range normalizeSMTPChannels(channels) {
		channel.Password = ""
		channel.PasswordConfigured = channel.PasswordConfigured || strings.TrimSpace(channel.Password) != ""
		out = append(out, channel)
	}
	return out
}

func MergeSMTPChannelPasswords(channels, previous []SMTPChannelConfig) []SMTPChannelConfig {
	if len(channels) == 0 {
		return nil
	}

	previousByID := make(map[string]SMTPChannelConfig, len(previous))
	for _, channel := range normalizeSMTPChannels(previous) {
		previousByID[channel.ID] = channel
	}

	merged := make([]SMTPChannelConfig, 0, len(channels))
	for _, channel := range channels {
		channel.ID = normalizeSMTPChannelID(channel.ID, len(merged))
		if strings.TrimSpace(channel.Password) == "" {
			if previousChannel, ok := previousByID[channel.ID]; ok {
				channel.Password = previousChannel.Password
			}
		}
		merged = append(merged, channel)
	}
	return normalizeSMTPChannels(merged)
}
