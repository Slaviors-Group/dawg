package sanitize

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

const redactedValue = "[REDACTED]"

func replacement(_ string, value string, category classification) string {
	if !category.synthetic {
		return redactedValue
	}

	digest := sha256.Sum256([]byte(category.reason + "\x00" + value))
	suffix := fmt.Sprintf("%x", digest[:4])
	switch category.reason {
	case "pii:email":
		return "user_" + suffix + "@example.com"
	case "pii:phone":
		return "+1555" + numericSuffix(digest[:6], 7)
	case "pii:name":
		return "User " + strings.ToUpper(suffix)
	default:
		return redactedValue
	}
}

func numericSuffix(bytes []byte, length int) string {
	var builder strings.Builder
	for _, value := range bytes {
		builder.WriteByte('0' + value%10)
		if builder.Len() == length {
			return builder.String()
		}
	}
	for builder.Len() < length {
		builder.WriteByte('0')
	}
	return builder.String()
}
