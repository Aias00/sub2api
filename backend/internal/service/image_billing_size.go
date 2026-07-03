package service

import (
	"strings"

	billingctx "github.com/Aias00/cloudbase/internal/billing"
)

const (
	ImageBillingSize1K = billingctx.ImageBillingSize1K
	ImageBillingSize2K = billingctx.ImageBillingSize2K
	ImageBillingSize4K = billingctx.ImageBillingSize4K

	ImageSizeSourceOutput  = billingctx.ImageSizeSourceOutput
	ImageSizeSourceInput   = billingctx.ImageSizeSourceInput
	ImageSizeSourceDefault = billingctx.ImageSizeSourceDefault
	ImageSizeSourceLegacy  = billingctx.ImageSizeSourceLegacy
)

type ImageBillingSizeResolution = billingctx.ImageBillingSizeResolution

func ClassifyImageBillingTier(size string) (string, bool) {
	return billingctx.ClassifyImageBillingTier(size)
}

func NormalizeImageBillingTierOrDefault(size string) string {
	return billingctx.NormalizeImageBillingTierOrDefault(size)
}

func ResolveImageBillingSize(inputSize string, outputSizes []string) ImageBillingSizeResolution {
	return billingctx.ResolveImageBillingSize(inputSize, outputSizes)
}

func ApplyOpenAIImageBillingResolution(result *OpenAIForwardResult) {
	if result == nil || result.ImageCount <= 0 {
		return
	}
	inputSize := strings.TrimSpace(result.ImageInputSize)
	if inputSize == "" && strings.TrimSpace(result.ImageSize) != ImageBillingSize2K {
		inputSize = strings.TrimSpace(result.ImageSize)
	}
	outputSizes := result.ImageOutputSizes
	if len(outputSizes) == 0 && strings.TrimSpace(result.ImageOutputSize) != "" {
		outputSizes = []string{result.ImageOutputSize}
	}
	resolved := ResolveImageBillingSize(inputSize, outputSizes)
	applyImageBillingResolution(
		&result.ImageSize,
		&result.ImageInputSize,
		&result.ImageOutputSize,
		&result.ImageSizeSource,
		&result.ImageSizeBreakdown,
		resolved,
	)
}

func ApplyForwardImageBillingResolution(result *ForwardResult) {
	if result == nil || result.ImageCount <= 0 {
		return
	}
	inputSize := strings.TrimSpace(result.ImageInputSize)
	if inputSize == "" && strings.TrimSpace(result.ImageSize) != ImageBillingSize2K {
		inputSize = strings.TrimSpace(result.ImageSize)
	}
	outputSizes := result.ImageOutputSizes
	if len(outputSizes) == 0 && strings.TrimSpace(result.ImageOutputSize) != "" {
		outputSizes = []string{result.ImageOutputSize}
	}
	resolved := ResolveImageBillingSize(inputSize, outputSizes)
	applyImageBillingResolution(
		&result.ImageSize,
		&result.ImageInputSize,
		&result.ImageOutputSize,
		&result.ImageSizeSource,
		&result.ImageSizeBreakdown,
		resolved,
	)
}

func applyImageBillingResolution(
	billingSize *string,
	inputSize *string,
	outputSize *string,
	source *string,
	breakdown *map[string]int,
	resolved ImageBillingSizeResolution,
) {
	*billingSize = resolved.BillingSize
	*inputSize = resolved.InputSize
	*outputSize = resolved.OutputSize
	*source = resolved.Source
	*breakdown = resolved.Breakdown
}

func SortedImageBillingBreakdownKeys(breakdown map[string]int) []string {
	return billingctx.SortedImageBillingBreakdownKeys(breakdown)
}
