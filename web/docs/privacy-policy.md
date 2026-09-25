---
description: "How DAWG handles local captures, sanitization, artifact storage, browser diagnostics, device metadata, and third-party resources."
---

# Privacy Policy

**Effective Date:** September 13, 2026

This Privacy Policy explains how DAWG ("we," "us," or "our") handles data when you use our browser extension, desktop application, and CLI tools. Because DAWG is a local-first platform designed for capturing and reproducing web application glitches, our data collection and processing methods are fundamentally designed to keep you in control of your data.

## 1. Local-First Application Architecture

The DAWG desktop application, browser extension, and CLI operate as **local-first** software. Captures and artifacts are stored on your machine unless you explicitly export or push them.

DAWG does not automatically upload or synchronize captures, bug reports, or artifacts to a DAWG-operated service. Sharing a `.dawg` archive, pushing an artifact to an OCI registry, or sending an artifact through another service is a manual action under your control.

## 2. What Data DAWG Captures (The "Artifact")

When you initiate a capture sequence, the DAWG browser extension and local engine intercept and record high-fidelity telemetry to create a reproducible artifact. This artifact (`application/vnd.dawg.*`) may contain the following data:

*   **Browser Session Data:**
    *   DOM snapshots, mutations, mouse activity, and other interaction events recorded through `rrweb`.
    *   Separate click and fill action records. Although rrweb uses `maskAllInputs: true`, entered values may still appear in this separate action stream before sanitization.
*   **Frontend Request Metadata:**
    *   Requests associated with the selected tab, including URLs, methods, headers, and request bodies, plus response status and headers. Frontend response bodies are not currently captured by this stream.
*   **Optional Environment Data:**
    *   Compose and environment lock information, database-diff or fixture data, structured logs, and third-party HTTP cassettes when those capture inputs are configured and available.
*   **Artifact Metadata:**
    *   Source commit, timestamps, target URL, tool versions, policy information, and determinism metadata.

## 3. How Data is Processed & Sanitized

To reduce accidental exposure of Personally Identifiable Information (PII) and credentials, the DAWG engine applies heuristic field-name and value rules to supported JSONL capture streams, then evaluates an Open Policy Agent (OPA) allow decision before packaging.

*   **Heuristic Sanitization:** Recognized emails, phone numbers, names, credentials, tokens, and other supported patterns are replaced or redacted locally. Unsupported files, unfamiliar field names, and unrecognized formats may remain unchanged.
*   **Policy Gate:** Packaging stops when sanitization fails, the selected Rego policy cannot be evaluated, or the policy denies the report. The bundled default policy allows every successfully completed sanitization pass; it does not independently inspect the sanitized content for remaining secrets.
*   **Audit Report:** `sanitize-report.json` records the selected policy, redaction counts, individual changes, the OPA result, and whether export is allowed.

Sanitization reduces risk but does not guarantee that an artifact contains no sensitive data. Review artifacts before sharing them.

## 4. Data Storage, Security, and Retention

Because DAWG is a local-first tool:

*   **Storage:** Local artifacts are stored as OCI Image Layout directories. Portable `.dawg` exports are ZIP archives containing a validated OCI layout; individual OCI layers may use tar or zstd-compressed tar formats.
*   **Security:** We strongly advise that you treat these artifacts as sensitive data, especially if captured from staging or production environments. Always review the artifact's contents locally before sharing it with third parties.
*   **Retention:** You are entirely responsible for the retention and deletion of your artifacts. We do not impose any automated deletion schedules on your local machine.

## 5. User Rights (Data Access and Deletion)

Under data protection laws (such as GDPR, CCPA, or the Indonesian PDP Law), you have rights regarding your personal data. 

Because we do not collect or store your data on our servers, you exercise these rights directly on your own machine. You can view, modify, or permanently delete any captured data simply by deleting the DAWG artifact files from your local storage.

## 6. Application Telemetry and Project Website

The DAWG desktop application, browser extension, and CLI do not include usage analytics or automated crash reporting in the current release.

The documentation website does not include an analytics service, but it currently requests fonts, icon resources, and selected technology logos from third-party hosts such as Google Fonts, unpkg, GitHub, and Simple Icons. Those providers can receive ordinary network metadata, including your IP address and browser request headers, when your browser loads those resources.

## 7. Your Responsibility as a Data Controller

If you use DAWG to capture data from your users or customers, you act as the **Data Controller** under applicable privacy laws. It is your responsibility to ensure that any data you capture, sanitize, and distribute using DAWG complies with your organization's legal obligations and privacy notices.

## 8. Contact & Support

If you have questions about this Privacy Policy or encounter potential vulnerabilities in the sanitization engine, please open an issue in the [project repository](https://github.com/Slaviors-Group/dawg/issues).
