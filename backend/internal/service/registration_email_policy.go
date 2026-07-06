package service

import (
	"github.com/Aias00/cloudbase/internal/identity"
)

// RegistrationEmailSuffix extracts normalized suffix in "@domain" form.
func RegistrationEmailSuffix(email string) string {
	return identity.RegistrationEmailSuffix(email)
}

// IsRegistrationEmailSuffixAllowed checks whether an email is allowed by suffix whitelist.
// Empty whitelist means allow all.
func IsRegistrationEmailSuffixAllowed(email string, whitelist []string) bool {
	return identity.IsRegistrationEmailSuffixAllowed(email, whitelist)
}

// NormalizeRegistrationEmailSuffixWhitelist normalizes and validates suffix whitelist items.
func NormalizeRegistrationEmailSuffixWhitelist(raw []string) ([]string, error) {
	return identity.NormalizeRegistrationEmailSuffixWhitelist(raw)
}

// ParseRegistrationEmailSuffixWhitelist parses persisted JSON into normalized suffixes.
// Invalid entries are ignored to keep old misconfigurations from breaking runtime reads.
func ParseRegistrationEmailSuffixWhitelist(raw string) []string {
	return identity.ParseRegistrationEmailSuffixWhitelist(raw)
}

func normalizeRegistrationEmailSuffix(raw string) (string, error) {
	return identity.NormalizeRegistrationEmailSuffix(raw)
}

func isValidRegistrationEmailDomain(domain string) bool {
	return identity.IsValidRegistrationEmailDomain(domain)
}

func registrationEmailDomainMatchesWildcard(domain string, allowed string) bool {
	return identity.RegistrationEmailDomainMatchesWildcard(domain, allowed)
}

func splitEmailForPolicy(raw string) (local string, domain string, ok bool) {
	return identity.SplitEmailForPolicy(raw)
}
