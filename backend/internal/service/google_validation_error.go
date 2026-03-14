package service

import (
	"context"
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	googleValidationReason        = "VALIDATION_REQUIRED"
	googleValidationVerifyDefault = "Verify account"
	googleValidationLearnDefault  = "Learn more"
)

var (
	googleValidationURLRegex        = regexp.MustCompile(`"validation_url"\s*:\s*"([^"]+)"`)
	googleValidationLabelRegex      = regexp.MustCompile(`"validation_url_link_text"\s*:\s*"([^"]+)"`)
	googleValidationLearnURLRegex   = regexp.MustCompile(`"validation_learn_more_url"\s*:\s*"([^"]+)"`)
	googleValidationLearnLabelRegex = regexp.MustCompile(`"validation_learn_more_link_text"\s*:\s*"([^"]+)"`)
	googleValidationMessageRegex    = regexp.MustCompile(`"validation_error_message"\s*:\s*"([^"]+)"`)
)

type GoogleValidationRequiredInfo struct {
	Message         string
	Reason          string
	Domain          string
	ValidationURL   string
	ValidationLabel string
	LearnMoreURL    string
	LearnMoreLabel  string
}

type googleValidationErrorPayload struct {
	Error struct {
		Message string            `json:"message"`
		Details []json.RawMessage `json:"details"`
	} `json:"error"`
}

type googleValidationErrorInfoDetail struct {
	Type     string            `json:"@type"`
	Reason   string            `json:"reason"`
	Domain   string            `json:"domain"`
	Metadata map[string]string `json:"metadata"`
}

type googleValidationHelpDetail struct {
	Type  string `json:"@type"`
	Links []struct {
		Description string `json:"description"`
		URL         string `json:"url"`
	} `json:"links"`
}

func ExtractGoogleValidationRequired(body []byte) *GoogleValidationRequiredInfo {
	if len(body) == 0 {
		return nil
	}

	var payload googleValidationErrorPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}

	info := &GoogleValidationRequiredInfo{
		Message: strings.TrimSpace(payload.Error.Message),
	}
	var matched bool

	for _, raw := range payload.Error.Details {
		var marker struct {
			Type string `json:"@type"`
		}
		if err := json.Unmarshal(raw, &marker); err != nil {
			continue
		}

		switch marker.Type {
		case "type.googleapis.com/google.rpc.ErrorInfo":
			var detail googleValidationErrorInfoDetail
			if err := json.Unmarshal(raw, &detail); err != nil {
				continue
			}
			if !strings.EqualFold(strings.TrimSpace(detail.Reason), googleValidationReason) {
				continue
			}
			matched = true
			info.Reason = strings.TrimSpace(detail.Reason)
			info.Domain = strings.TrimSpace(detail.Domain)
			info.ValidationURL = firstHTTPSURL(
				detail.Metadata["validation_url"],
				info.ValidationURL,
			)
			info.ValidationLabel = firstNonEmpty(
				detail.Metadata["validation_url_link_text"],
				info.ValidationLabel,
				googleValidationVerifyDefault,
			)
			info.LearnMoreURL = firstHTTPSURL(
				detail.Metadata["validation_learn_more_url"],
				info.LearnMoreURL,
			)
			info.LearnMoreLabel = firstNonEmpty(
				detail.Metadata["validation_learn_more_link_text"],
				info.LearnMoreLabel,
				googleValidationLearnDefault,
			)
			if info.Message == "" {
				info.Message = strings.TrimSpace(detail.Metadata["validation_error_message"])
			}
		case "type.googleapis.com/google.rpc.Help":
			var detail googleValidationHelpDetail
			if err := json.Unmarshal(raw, &detail); err != nil {
				continue
			}
			for _, link := range detail.Links {
				description := strings.TrimSpace(link.Description)
				linkURL := normalizeHTTPSURL(link.URL)
				if linkURL == "" {
					continue
				}
				lowerDesc := strings.ToLower(description)
				switch {
				case info.ValidationURL == "" && (strings.Contains(lowerDesc, "verify") || strings.Contains(lowerDesc, "validation")):
					info.ValidationURL = linkURL
					info.ValidationLabel = firstNonEmpty(description, info.ValidationLabel, googleValidationVerifyDefault)
				case info.LearnMoreURL == "" && strings.Contains(lowerDesc, "learn"):
					info.LearnMoreURL = linkURL
					info.LearnMoreLabel = firstNonEmpty(description, info.LearnMoreLabel, googleValidationLearnDefault)
				case info.ValidationURL == "":
					info.ValidationURL = linkURL
					info.ValidationLabel = firstNonEmpty(description, info.ValidationLabel, googleValidationVerifyDefault)
				case info.LearnMoreURL == "":
					info.LearnMoreURL = linkURL
					info.LearnMoreLabel = firstNonEmpty(description, info.LearnMoreLabel, googleValidationLearnDefault)
				}
			}
		}
	}

	if !matched || info.ValidationURL == "" {
		return nil
	}

	info.ValidationLabel = firstNonEmpty(info.ValidationLabel, googleValidationVerifyDefault)
	info.LearnMoreLabel = firstNonEmpty(info.LearnMoreLabel, googleValidationLearnDefault)
	return info
}

func ExtractGoogleValidationRequiredFromText(text string) *GoogleValidationRequiredInfo {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	if info := ExtractGoogleValidationRequired([]byte(text)); info != nil {
		return info
	}
	if idx := strings.IndexByte(text, '{'); idx >= 0 {
		if info := ExtractGoogleValidationRequired([]byte(text[idx:])); info != nil {
			return info
		}
	}

	validationURL := firstHTTPSURL(extractRegexGroup(googleValidationURLRegex, text))
	if validationURL == "" {
		return nil
	}

	info := &GoogleValidationRequiredInfo{
		Message:         firstNonEmpty(extractRegexGroup(googleValidationMessageRegex, text)),
		Reason:          googleValidationReason,
		ValidationURL:   validationURL,
		ValidationLabel: firstNonEmpty(extractRegexGroup(googleValidationLabelRegex, text), googleValidationVerifyDefault),
		LearnMoreURL:    firstHTTPSURL(extractRegexGroup(googleValidationLearnURLRegex, text)),
		LearnMoreLabel:  firstNonEmpty(extractRegexGroup(googleValidationLearnLabelRegex, text), googleValidationLearnDefault),
	}
	return info
}

func persistGoogleValidationRequired(ctx context.Context, repo AccountRepository, account *Account, body []byte) *GoogleValidationRequiredInfo {
	if repo == nil || account == nil || account.Platform != PlatformGemini {
		return nil
	}

	info := ExtractGoogleValidationRequired(body)
	if info == nil {
		return nil
	}

	updates := googleValidationExtraUpdates(info)
	if len(updates) == 0 {
		return info
	}
	_ = repo.UpdateExtra(ctx, account.ID, updates)
	mergeAccountExtra(account, updates)
	return info
}

func persistGoogleValidationRequiredFromText(ctx context.Context, repo AccountRepository, account *Account, text string) *GoogleValidationRequiredInfo {
	if repo == nil || account == nil || account.Platform == "" {
		return nil
	}
	info := ExtractGoogleValidationRequiredFromText(text)
	if info == nil {
		return nil
	}
	updates := googleValidationExtraUpdates(info)
	if len(updates) == 0 {
		return info
	}
	_ = repo.UpdateExtra(ctx, account.ID, updates)
	mergeAccountExtra(account, updates)
	return info
}

func googleValidationExtraUpdates(info *GoogleValidationRequiredInfo) map[string]any {
	if info == nil || strings.TrimSpace(info.ValidationURL) == "" {
		return nil
	}

	updates := map[string]any{
		"google_validation_required":   true,
		"google_validation_url":        info.ValidationURL,
		"google_validation_label":      firstNonEmpty(info.ValidationLabel, googleValidationVerifyDefault),
		"google_validation_updated_at": time.Now().UTC().Format(time.RFC3339),
	}
	if v := strings.TrimSpace(info.Message); v != "" {
		updates["google_validation_message"] = v
	}
	if v := strings.TrimSpace(info.Reason); v != "" {
		updates["google_validation_reason"] = v
	}
	if v := strings.TrimSpace(info.Domain); v != "" {
		updates["google_validation_domain"] = v
	}
	if v := strings.TrimSpace(info.LearnMoreURL); v != "" {
		updates["google_validation_learn_more_url"] = v
	}
	if v := strings.TrimSpace(info.LearnMoreLabel); v != "" {
		updates["google_validation_learn_more_label"] = v
	}
	return updates
}

func firstHTTPSURL(values ...string) string {
	for _, value := range values {
		if normalized := normalizeHTTPSURL(value); normalized != "" {
			return normalized
		}
	}
	return ""
}

func normalizeHTTPSURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if !strings.EqualFold(parsed.Scheme, "https") || parsed.Host == "" {
		return ""
	}
	return parsed.String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func extractRegexGroup(re *regexp.Regexp, text string) string {
	if re == nil || text == "" {
		return ""
	}
	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}
