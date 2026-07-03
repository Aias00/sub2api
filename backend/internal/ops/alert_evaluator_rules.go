package ops

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type AlertRuleScope struct {
	Platform string
	GroupID  *int64
	Region   *string
}

type AlertDescriptionRule struct {
	MetricType string
	Operator   string
	Threshold  float64
}

func ParseAlertRuleScope(filters map[string]any) AlertRuleScope {
	scope := AlertRuleScope{}
	if filters == nil {
		return scope
	}
	if v, ok := filters["platform"]; ok {
		if s, ok := v.(string); ok {
			scope.Platform = strings.TrimSpace(s)
		}
	}
	if v, ok := filters["group_id"]; ok {
		scope.GroupID = parseAlertScopeGroupID(v)
	}
	if v, ok := filters["region"]; ok {
		if s, ok := v.(string); ok {
			vv := strings.TrimSpace(s)
			if vv != "" {
				scope.Region = &vv
			}
		}
	}
	return scope
}

func BuildAlertDimensions(platform string, groupID *int64) map[string]any {
	dims := map[string]any{}
	if strings.TrimSpace(platform) != "" {
		dims["platform"] = strings.TrimSpace(platform)
	}
	if groupID != nil && *groupID > 0 {
		dims["group_id"] = *groupID
	}
	if len(dims) == 0 {
		return nil
	}
	return dims
}

func BuildAlertDescription(rule AlertDescriptionRule, value float64, windowMinutes int, platform string, groupID *int64) string {
	scope := "overall"
	if strings.TrimSpace(platform) != "" {
		scope = fmt.Sprintf("platform=%s", strings.TrimSpace(platform))
	}
	if groupID != nil && *groupID > 0 {
		scope = fmt.Sprintf("%s group_id=%d", scope, *groupID)
	}
	if windowMinutes <= 0 {
		windowMinutes = 1
	}
	return fmt.Sprintf("%s %s %.2f (current %.2f) over last %dm (%s)",
		strings.TrimSpace(rule.MetricType),
		strings.TrimSpace(rule.Operator),
		rule.Threshold,
		value,
		windowMinutes,
		strings.TrimSpace(scope),
	)
}

func RequiredSustainedBreaches(sustainedMinutes int, interval time.Duration) int {
	if sustainedMinutes <= 0 {
		return 1
	}
	if interval <= 0 {
		return sustainedMinutes
	}
	required := int(math.Ceil(float64(sustainedMinutes*60) / interval.Seconds()))
	if required < 1 {
		return 1
	}
	return required
}

func CompareAlertMetric(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}

func parseAlertScopeGroupID(value any) *int64 {
	switch t := value.(type) {
	case float64:
		if t > 0 {
			id := int64(t)
			return &id
		}
	case int64:
		if t > 0 {
			id := t
			return &id
		}
	case int:
		if t > 0 {
			id := int64(t)
			return &id
		}
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		if err == nil && n > 0 {
			return &n
		}
	}
	return nil
}
