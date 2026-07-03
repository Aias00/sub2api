package wechat

import "testing"

func TestNormalizeExportFormats(t *testing.T) {
	tests := []struct {
		name string
		raw  []string
		want []ExportFormat
		ok   bool
	}{
		{name: "defaults", want: []ExportFormat{ExportFormatHTML, ExportFormatMarkdown}, ok: true},
		{name: "trims lowercases and dedupes", raw: []string{" Markdown ", "HTML", "markdown"}, want: []ExportFormat{ExportFormatMarkdown, ExportFormatHTML}, ok: true},
		{name: "invalid", raw: []string{"pdf"}, ok: false},
		{name: "blank invalid", raw: []string{""}, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := NormalizeExportFormats(tt.raw)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("formats = %#v, want %#v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("formats = %#v, want %#v", got, tt.want)
				}
			}
		})
	}
}

func TestNormalizeExportRetentionDays(t *testing.T) {
	if got := NormalizeExportRetentionDays(0); got != DefaultExportRetentionDays {
		t.Fatalf("default retention = %d, want %d", got, DefaultExportRetentionDays)
	}
	if got := NormalizeExportRetentionDays(30); got != 30 {
		t.Fatalf("retention = %d, want 30", got)
	}
}

func TestEstimateExportCost(t *testing.T) {
	if got := EstimateExportCost(2, 2, false, 0.05); got != 0.2 {
		t.Fatalf("cost = %v, want 0.2", got)
	}
	if got := EstimateExportCost(2, 0, false, 0.05); got != 0.1 {
		t.Fatalf("zero format cost = %v, want 0.1", got)
	}
	if got := EstimateExportCost(2, 1, true, 0.05); got != 0.2 {
		t.Fatalf("engagement cost = %v, want 0.2", got)
	}
	if got := EstimateExportCost(1, 1, false, -1); got != DefaultExportCostPerArticle {
		t.Fatalf("default cost = %v, want %v", got, DefaultExportCostPerArticle)
	}
}
