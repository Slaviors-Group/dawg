# Sanitizer Policy

DAWG's sanitizer removes sensitive data from captured artifacts before export. It uses a deny-by-default approach powered by Open Policy Agent (OPA) Rego policies.

## How It Works

```
Capture Output ──▶ Rule Engine ──▶ OPA Policy Gate ──▶ Sanitized Artifact
                    (field-name     (deny-by-default)    (safe to export)
                     matching +
                     regex)
```

**Core principle:** If the sanitizer can't identify a field as public, it redacts it. Export is blocked (not warned) if the OPA policy fails.

---

## Detection Layers

### 1. Field-Name Matching

Matches field names against known patterns:

```yaml
# Patterns that trigger redaction
password
*.password
secret
*.secret
token
*.token
authorization
*.authorization
cookie
*.cookie
session_id
*.session_id
api_key
*.api_key
credit_card
*.credit_card
```

### 2. Regex Patterns

Detects sensitive data by content pattern:

| Pattern | Type | Example |
|---|---|---|
| `\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b` | Credit card | `4111 1111 1111 1111` |
| `eyJ[A-Za-z0-9-_]+\.eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+` | JWT | `eyJhbGci...` |
| `[A-Za-z0-9]{32,}` | Generic secret | Long alphanumeric strings |
| `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z\|a-z]{2,}\b` | Email | `user@example.com` |

### 3. Gitleaks Patterns

Reuses Gitleaks' battle-tested secret detection rules:

- AWS access keys
- GitHub/GitLab tokens
- Slack webhooks
- Private keys (PEM format)
- API keys for common services

### 4. Format-Preserving Faker

When a field is redacted, DAWG replaces it with synthetic data that maintains the same format:

| Original | Replacement | Format preserved |
|---|---|---|
| `john@example.com` | `uq4k7m@x8v2j.io` | Yes (valid email) |
| `555-123-4567` | `555-987-6543` | Yes (valid phone) |
| `John Smith` | `Alex Johnson` | Yes (valid name) |
| `4111-1111-1111-1111` | `5555-4444-3333-2222` | Yes (valid card format) |

This ensures replay validation still passes with synthetic data.

---

## OPA Policy Gate

The OPA policy is the **hard gate** for export. It evaluates the entire sanitized output and determines whether the artifact is safe to export.

### Default Policy (`default.rego`)

```text
package dawg.sanitize

default allow = false

# Allow export if no unredacted secrets remain
allow {
    count(secrets) == 0
}

# Allow export if all secrets were redacted
allow {
    count(unredacted_secrets) == 0
}

secrets[field] {
    field := input.captured_data[_]
    is_secret(field)
}

unredacted_secrets[field] {
    field := secrets[field]
    not field.redacted
}
```

### Policy Behavior

| Condition | Result |
|---|---|
| No secrets detected | ✅ Export allowed |
| All secrets redacted | ✅ Export allowed |
| Any secret unredacted | ❌ **Export blocked** |
| Policy file missing | ❌ **Export blocked** |

### Policy Presets

The Desktop UI provides preset policies:

| Preset | Strictness | Use case |
|---|---|---|
| **Default** | Standard | General web-app debugging |
| **Strict** | Maximum | Security-sensitive environments |
| **Minimal** | Low | Internal/trusted networks only |

---

## Sanitization Report

Every sanitized artifact includes a `sanitize-report.json`:

```json
{
  "policyVersion": "1.2.0",
  "fieldsRedacted": 14,
  "fieldsSynthetic": 8,
  "secretsDetected": 3,
  "secretsRedacted": 3,
  "reviewedBy": "auto",
  "timestamp": "2026-09-06T12:00:00Z",
  "redactions": [
    {
      "path": "http_pairs[3].request.headers.authorization",
      "type": "secret",
      "action": "redacted",
      "replacement": "[REDACTED]"
    },
    {
      "path": "db_diff[0].rows[2].email",
      "type": "pii",
      "action": "synthetic",
      "replacement": "uq4k7m@x8v2j.io"
    }
  ]
}
```

---

## Custom Policies

### Writing Custom Rego Policies

Create a `.rego` file in your project:

```text
package dawg.sanitize

default allow = false

# Your custom rules here
allow {
    # Allow if all PII fields are redacted
    count(input.unredacted_pii) == 0
}

# Block export of specific fields
deny {
    field := input.captured_data[_]
    field.path == "payment.card_number"
}
```

### Applying Custom Policies

```yaml
# dawg.config.yaml
sanitize:
  policy: ./my-custom-policy.rego
  deny_unmatched: true
```

---

## Security Guarantees

| Guarantee | Implementation |
|---|---|
| **No PII leaks** | Hard OPA gate blocks export if any field is unredacted |
| **No secret leaks** | Gitleaks patterns + field-name matching + OPA policy |
| **Audit trail** | `sanitize-report.json` logs every redaction |
| **Deterministic** | Same input produces same redaction (reproducible) |
| **Versioned policies** | Policy version tracked in manifest for reproducibility |

---

## Known Limitations

- **Rule-based detection is not 100%** — novel PII patterns may be missed. Mitigation: `deny_unmatched: true` redacts unknown fields.
- **ORM-tap blind spot** — raw SQL queries outside the ORM are not captured. Planned upgrade: Debezium CDC (P1).
- **Format preservation** — some synthetic replacements may break strict validation. Manual review recommended for critical fields.
