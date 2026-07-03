package ops

import (
	"errors"
	"strings"
)

type QueryMode string

const (
	QueryModeAuto   QueryMode = "auto"
	QueryModeRaw    QueryMode = "raw"
	QueryModePreagg QueryMode = "preagg"
)

const (
	ListViewErrors   = "errors"
	ListViewExcluded = "excluded"
	ListViewAll      = "all"
)

// ErrPreaggregatedNotPopulated indicates that raw logs exist for a window, but
// the pre-aggregation tables are not populated yet.
var ErrPreaggregatedNotPopulated = errors.New("ops pre-aggregated tables not populated")

func ParseQueryMode(raw string) QueryMode {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case string(QueryModeRaw):
		return QueryModeRaw
	case string(QueryModePreagg):
		return QueryModePreagg
	default:
		return QueryModeAuto
	}
}

func ParseOptionalQueryMode(raw string) QueryMode {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	return ParseQueryMode(raw)
}

func (m QueryMode) IsValid() bool {
	switch m {
	case QueryModeAuto, QueryModeRaw, QueryModePreagg:
		return true
	default:
		return false
	}
}

func ShouldFallbackPreagg(mode QueryMode, err error) bool {
	return mode == QueryModeAuto && errors.Is(err, ErrPreaggregatedNotPopulated)
}

func ParseListView(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "", ListViewErrors:
		return ListViewErrors
	case ListViewExcluded:
		return ListViewExcluded
	case ListViewAll:
		return ListViewAll
	default:
		return ListViewErrors
	}
}
