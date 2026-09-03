// Package sanitize implements PRD §6.2 redaction and export gating for captured data.
package sanitize

import (
	"regexp"
	"strings"
)

var (
	emailPattern      = regexp.MustCompile(`(?i)^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	phonePattern      = regexp.MustCompile(`^\+?[0-9][0-9 .()\-]{6,}[0-9]$`)
	creditCardPattern = regexp.MustCompile(`^[0-9][0-9 -]{11,}[0-9]$`)
	jwtPattern        = regexp.MustCompile(`^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`)
)

type classification struct {
	reason    string
	synthetic bool
}

func classify(field, value string) (classification, bool) {
	field = strings.ToLower(field)
	value = strings.TrimSpace(value)
	if strings.HasSuffix(field, "tagname") || strings.HasSuffix(field, "nodename") || strings.HasSuffix(field, "localname") {
		return classification{}, false
	}

	if isSecretField(field) || strings.HasPrefix(strings.ToLower(value), "bearer ") || jwtPattern.MatchString(value) || matchesSecretPattern(value) {
		return classification{reason: "secret:token"}, true
	}
	if strings.Contains(field, "email") || emailPattern.MatchString(value) {
		return classification{reason: "pii:email", synthetic: true}, true
	}
	if strings.Contains(field, "card") || strings.Contains(field, "credit") || creditCardPattern.MatchString(value) {
		return classification{reason: "secret:payment-card"}, true
	}
	if strings.Contains(field, "phone") || strings.Contains(field, "mobile") || phonePattern.MatchString(value) {
		return classification{reason: "pii:phone", synthetic: true}, true
	}
	if strings.Contains(field, "name") {
		return classification{reason: "pii:name", synthetic: true}, true
	}
	return classification{}, false
}

func isSecretField(field string) bool {
	for _, marker := range []string{
		"password", "authorization", "cookie", "token", "secret", "api_key", "api-key", "apikey", "session",
		"ssn", "social_security", "passport", "national_id", "tax_id", "date_of_birth", "birthdate", "address",
	} {
		if strings.Contains(field, marker) {
			return true
		}
	}
	return false
}
