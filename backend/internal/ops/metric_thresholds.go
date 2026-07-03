package ops

import "errors"

type MetricThresholds struct {
	SLAPercentMin               *float64 `json:"sla_percent_min,omitempty"`
	TTFTp99MsMax                *float64 `json:"ttft_p99_ms_max,omitempty"`
	RequestErrorRatePercentMax  *float64 `json:"request_error_rate_percent_max,omitempty"`
	UpstreamErrorRatePercentMax *float64 `json:"upstream_error_rate_percent_max,omitempty"`
}

func DefaultMetricThresholds() *MetricThresholds {
	slaMin := 99.5
	ttftMax := 500.0
	reqErrMax := 5.0
	upstreamErrMax := 5.0
	return &MetricThresholds{
		SLAPercentMin:               &slaMin,
		TTFTp99MsMax:                &ttftMax,
		RequestErrorRatePercentMax:  &reqErrMax,
		UpstreamErrorRatePercentMax: &upstreamErrMax,
	}
}

func ValidateMetricThresholds(cfg *MetricThresholds) error {
	if cfg == nil {
		return errors.New("invalid config")
	}
	if cfg.SLAPercentMin != nil && (*cfg.SLAPercentMin < 0 || *cfg.SLAPercentMin > 100) {
		return errors.New("sla_percent_min must be between 0 and 100")
	}
	if cfg.TTFTp99MsMax != nil && *cfg.TTFTp99MsMax < 0 {
		return errors.New("ttft_p99_ms_max must be >= 0")
	}
	if cfg.RequestErrorRatePercentMax != nil && (*cfg.RequestErrorRatePercentMax < 0 || *cfg.RequestErrorRatePercentMax > 100) {
		return errors.New("request_error_rate_percent_max must be between 0 and 100")
	}
	if cfg.UpstreamErrorRatePercentMax != nil && (*cfg.UpstreamErrorRatePercentMax < 0 || *cfg.UpstreamErrorRatePercentMax > 100) {
		return errors.New("upstream_error_rate_percent_max must be between 0 and 100")
	}
	return nil
}
