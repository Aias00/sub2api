package service

import opsctx "github.com/Aias00/cloudbase/internal/ops"

type OpsQueryMode = opsctx.QueryMode

const (
	OpsQueryModeAuto   OpsQueryMode = opsctx.QueryModeAuto
	OpsQueryModeRaw    OpsQueryMode = opsctx.QueryModeRaw
	OpsQueryModePreagg OpsQueryMode = opsctx.QueryModePreagg
)

// ErrOpsPreaggregatedNotPopulated indicates that raw logs exist for a window, but the
// pre-aggregation tables are not populated yet. This is primarily used to implement
// the forced `preagg` mode UX.
var ErrOpsPreaggregatedNotPopulated = opsctx.ErrPreaggregatedNotPopulated

func ParseOpsQueryMode(raw string) OpsQueryMode {
	return opsctx.ParseQueryMode(raw)
}

func shouldFallbackOpsPreagg(filter *OpsDashboardFilter, err error) bool {
	return filter != nil &&
		opsctx.ShouldFallbackPreagg(filter.QueryMode, err)
}

func cloneOpsFilterWithMode(filter *OpsDashboardFilter, mode OpsQueryMode) *OpsDashboardFilter {
	if filter == nil {
		return nil
	}
	cloned := *filter
	cloned.QueryMode = mode
	return &cloned
}
