package billing

func ResolveImageRateMultiplier(imageRateIndependent bool, imageRateMultiplier float64, effectiveGroupMultiplier float64) float64 {
	if imageRateIndependent {
		if imageRateMultiplier < 0 {
			return 0
		}
		return imageRateMultiplier
	}
	return effectiveGroupMultiplier
}
