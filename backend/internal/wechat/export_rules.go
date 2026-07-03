package wechat

import "strings"

const (
	DefaultExportCostPerArticle    = 0.05
	ExportEngagementCostMultiplier = 2.0
	DefaultExportRetentionDays     = 7
)

type ExportFormat string

const (
	ExportFormatHTML     ExportFormat = "html"
	ExportFormatMarkdown ExportFormat = "markdown"
)

func NormalizeExportFormats(raw []string) ([]ExportFormat, bool) {
	if len(raw) == 0 {
		raw = []string{string(ExportFormatHTML), string(ExportFormatMarkdown)}
	}
	formats := make([]ExportFormat, 0, len(raw))
	seen := make(map[ExportFormat]struct{}, len(raw))
	for _, item := range raw {
		format := ExportFormat(strings.ToLower(strings.TrimSpace(item)))
		switch format {
		case ExportFormatHTML, ExportFormatMarkdown:
			if _, ok := seen[format]; !ok {
				seen[format] = struct{}{}
				formats = append(formats, format)
			}
		default:
			return nil, false
		}
	}
	if len(formats) == 0 {
		return nil, false
	}
	return formats, true
}

func NormalizeExportRetentionDays(days int) int {
	if days <= 0 {
		return DefaultExportRetentionDays
	}
	return days
}

func EstimateExportCost(articleCount, formatCount int, includeEngagement bool, costPerArticle float64) float64 {
	if costPerArticle <= 0 {
		costPerArticle = DefaultExportCostPerArticle
	}
	engagementMultiplier := 1.0
	if includeEngagement {
		engagementMultiplier = ExportEngagementCostMultiplier
	}
	return costPerArticle * float64(articleCount) * float64(maxIntValue(1, formatCount)) * engagementMultiplier
}

func maxIntValue(a, b int) int {
	if a > b {
		return a
	}
	return b
}
