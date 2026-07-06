package identity

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var userAgentVersionRegex = regexp.MustCompile(`/(\d+)\.(\d+)\.(\d+)`)

type Fingerprint struct {
	ClientID                string
	UserAgent               string
	StainlessLang           string
	StainlessPackageVersion string
	StainlessOS             string
	StainlessArch           string
	StainlessRuntime        string
	StainlessRuntimeVersion string
	UpdatedAt               int64
}

func FingerprintFromHeaders(headers http.Header, defaults Fingerprint) Fingerprint {
	return Fingerprint{
		UserAgent:               headerOrDefault(headers, "User-Agent", defaults.UserAgent),
		StainlessLang:           headerOrDefault(headers, "X-Stainless-Lang", defaults.StainlessLang),
		StainlessPackageVersion: headerOrDefault(headers, "X-Stainless-Package-Version", defaults.StainlessPackageVersion),
		StainlessOS:             headerOrDefault(headers, "X-Stainless-OS", defaults.StainlessOS),
		StainlessArch:           headerOrDefault(headers, "X-Stainless-Arch", defaults.StainlessArch),
		StainlessRuntime:        headerOrDefault(headers, "X-Stainless-Runtime", defaults.StainlessRuntime),
		StainlessRuntimeVersion: headerOrDefault(headers, "X-Stainless-Runtime-Version", defaults.StainlessRuntimeVersion),
	}
}

func MergeHeadersIntoFingerprint(fp *Fingerprint, headers http.Header) {
	if fp == nil {
		return
	}
	if ua := headers.Get("User-Agent"); ua != "" {
		fp.UserAgent = ua
	}
	mergeHeader(headers, "X-Stainless-Lang", &fp.StainlessLang)
	mergeHeader(headers, "X-Stainless-Package-Version", &fp.StainlessPackageVersion)
	mergeHeader(headers, "X-Stainless-OS", &fp.StainlessOS)
	mergeHeader(headers, "X-Stainless-Arch", &fp.StainlessArch)
	mergeHeader(headers, "X-Stainless-Runtime", &fp.StainlessRuntime)
	mergeHeader(headers, "X-Stainless-Runtime-Version", &fp.StainlessRuntimeVersion)
}

func ApplyFingerprintHeaders(headers http.Header, fp *Fingerprint) {
	if headers == nil || fp == nil {
		return
	}
	setHeaderRaw(headers, "User-Agent", fp.UserAgent)
	setHeaderRaw(headers, "X-Stainless-Lang", fp.StainlessLang)
	setHeaderRaw(headers, "X-Stainless-Package-Version", fp.StainlessPackageVersion)
	setHeaderRaw(headers, "X-Stainless-OS", fp.StainlessOS)
	setHeaderRaw(headers, "X-Stainless-Arch", fp.StainlessArch)
	setHeaderRaw(headers, "X-Stainless-Runtime", fp.StainlessRuntime)
	setHeaderRaw(headers, "X-Stainless-Runtime-Version", fp.StainlessRuntimeVersion)
}

func IsNewerUserAgentVersion(newUA, cachedUA string) bool {
	newProduct := extractProduct(newUA)
	cachedProduct := extractProduct(cachedUA)
	if newProduct == "" || cachedProduct == "" || newProduct != cachedProduct {
		return false
	}

	newMajor, newMinor, newPatch, newOk := parseUserAgentVersion(newUA)
	cachedMajor, cachedMinor, cachedPatch, cachedOk := parseUserAgentVersion(cachedUA)
	if !newOk || !cachedOk {
		return false
	}
	if newMajor > cachedMajor {
		return true
	}
	if newMajor < cachedMajor {
		return false
	}
	if newMinor > cachedMinor {
		return true
	}
	if newMinor < cachedMinor {
		return false
	}
	return newPatch > cachedPatch
}

func headerOrDefault(headers http.Header, key, defaultValue string) string {
	if v := getHeader(headers, key); v != "" {
		return v
	}
	return defaultValue
}

func mergeHeader(headers http.Header, key string, target *string) {
	if v := getHeader(headers, key); v != "" {
		*target = v
	}
}

func getHeader(headers http.Header, key string) string {
	if v := headers.Get(key); v != "" {
		return v
	}
	for existing, values := range headers {
		if strings.EqualFold(existing, key) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func parseUserAgentVersion(ua string) (major, minor, patch int, ok bool) {
	matches := userAgentVersionRegex.FindStringSubmatch(ua)
	if len(matches) != 4 {
		return 0, 0, 0, false
	}
	major, _ = strconv.Atoi(matches[1])
	minor, _ = strconv.Atoi(matches[2])
	patch, _ = strconv.Atoi(matches[3])
	return major, minor, patch, true
}

func extractProduct(ua string) string {
	if idx := strings.Index(ua, "/"); idx > 0 {
		return strings.ToLower(ua[:idx])
	}
	return ""
}

func setHeaderRaw(h http.Header, key, value string) {
	if h == nil || strings.TrimSpace(key) == "" || value == "" {
		return
	}
	for existing := range h {
		if strings.EqualFold(existing, key) {
			delete(h, existing)
		}
	}
	h[key] = []string{value}
}
