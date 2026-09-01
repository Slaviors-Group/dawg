package sanitize

import "testing"

func TestClassifyDetectsRequiredP0Patterns(t *testing.T) {
	testCases := []struct {
		name      string
		field     string
		value     string
		reason    string
		synthetic bool
	}{
		{name: "email", field: "email", value: "person@example.test", reason: "pii:email", synthetic: true},
		{name: "phone", field: "phone", value: "+62 812-3456-7890", reason: "pii:phone", synthetic: true},
		{name: "card", field: "payment.card", value: "4111 1111 1111 1111", reason: "secret:payment-card"},
		{name: "JWT", field: "payload", value: "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.signature", reason: "secret:token"},
		{name: "bearer token", field: "authorization", value: "Bearer top-secret", reason: "secret:token"},
		{name: "password field", field: "password", value: "not-a-pattern", reason: "secret:token"},
		{name: "unrecognized sensitive position", field: "profile.ssn", value: "123-45-6789", reason: "secret:token"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			category, matched := classify(testCase.field, testCase.value)
			if !matched {
				t.Fatal("expected a sensitive value match")
			}
			if category.reason != testCase.reason || category.synthetic != testCase.synthetic {
				t.Fatalf("unexpected classification: %#v", category)
			}
		})
	}
}

func TestClassifyRedactsHyphenatedAPIKeyHeader(t *testing.T) {
	category, matched := classify("request.headers.x-api-key", "private-key")
	if !matched || category.reason != "secret:token" {
		t.Fatalf("expected API key header to be redacted, got %#v, matched=%t", category, matched)
	}
}

func TestReplacementIsStableAndFormatPreserving(t *testing.T) {
	category, matched := classify("email", "person@example.test")
	if !matched {
		t.Fatal("expected email to match")
	}
	first := replacement("email", "person@example.test", category)
	second := replacement("email", "person@example.test", category)
	if first != second {
		t.Fatalf("replacement must be stable: %q != %q", first, second)
	}
	if !emailPattern.MatchString(first) || first == "person@example.test" {
		t.Fatalf("replacement must be a different valid email: %q", first)
	}
}

func TestReplacementRedactsSecrets(t *testing.T) {
	category, matched := classify("authorization", "Bearer top-secret")
	if !matched {
		t.Fatal("expected authorization to match")
	}
	if actual := replacement("authorization", "Bearer top-secret", category); actual != redactedValue {
		t.Fatalf("expected redacted value, got %q", actual)
	}
}
