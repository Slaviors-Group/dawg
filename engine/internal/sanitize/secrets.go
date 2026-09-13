package sanitize

import (
	"regexp"
)

var secretPatterns = []*regexp.Regexp{
	// AWS Access Key ID
	regexp.MustCompile(`\b(A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}\b`),
	// GCP API Key
	regexp.MustCompile(`\bAIza[0-9A-Za-z\-_]{35}\b`),
	// GitHub Token
	regexp.MustCompile(`\b(ghp|gho|ghu|ghs|ghr)_[0-9a-zA-Z]{36}\b`),
	// Generic Private Key
	regexp.MustCompile(`-----BEGIN (RSA|OPENSSH|DSA|EC|PGP) PRIVATE KEY-----`),
	// Slack Token
	regexp.MustCompile(`\bxox[baprs]-[0-9]+-[0-9a-zA-Z]+\b`),
}

// matchesSecretPattern returns true if the value matches any known secret format.
func matchesSecretPattern(value string) bool {
	for _, pattern := range secretPatterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}
