package sanitize

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

const sanitizeReportFileName = "sanitize-report.json"

var captureFiles = []string{
	"http/frontend.jsonl",
	"http/backend.jsonl",
	"db/diff.jsonl",
	"logs/structured.jsonl",
	"traces/rrweb.jsonl",
	"actions/browser.jsonl",
	"cassettes/thirdparty.jsonl",
	"diagnostics/console.jsonl",
	"diagnostics/network.jsonl",
	"diagnostics/errors.jsonl",
}

// SanitizeFiles redacts supported JSONL capture files in place and returns a report before policy evaluation.
func SanitizeFiles(sessionDirectory, policyFile, policyVersion string) (dawgtypes.SanitizeReport, error) {
	report := dawgtypes.SanitizeReport{
		PolicyVersion: policyVersion,
		PolicyFile:    policyFile,
		OPAResult:     "not-evaluated",
		Redactions:    []dawgtypes.Redaction{},
		BlockedFields: []string{},
	}
	for _, relativePath := range captureFiles {
		path := filepath.Join(sessionDirectory, filepath.FromSlash(relativePath))
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return dawgtypes.SanitizeReport{}, fmt.Errorf("sanitizer: inspect %s: %w", path, err)
		}
		fileReport, err := sanitizeJSONLFile(path, relativePath)
		if err != nil {
			return dawgtypes.SanitizeReport{}, err
		}
		report.FieldsScanned += fileReport.fieldsScanned
		report.FieldsRedacted += len(fileReport.redactions)
		report.Redactions = append(report.Redactions, fileReport.redactions...)
	}
	bodyReport, err := sanitizeDiagnosticBodies(sessionDirectory)
	if err != nil {
		return dawgtypes.SanitizeReport{}, err
	}
	report.FieldsScanned += bodyReport.fieldsScanned
	report.FieldsRedacted += len(bodyReport.redactions)
	report.Redactions = append(report.Redactions, bodyReport.redactions...)
	return report, nil
}

// SanitizeDirectory redacts supported capture files, evaluates the export policy, and writes its audit report.
func SanitizeDirectory(ctx context.Context, sessionDirectory, policyFile, policyVersion string) (dawgtypes.SanitizeReport, error) {
	report, err := SanitizeFiles(sessionDirectory, policyFile, policyVersion)
	if err != nil {
		return dawgtypes.SanitizeReport{}, err
	}
	policyResult, err := EvaluatePolicy(ctx, policyFile, map[string]any{
		"fieldsRedacted": report.FieldsRedacted,
		"blockedFields":  report.BlockedFields,
	})
	if err != nil {
		return dawgtypes.SanitizeReport{}, err
	}
	if policyResult.Allowed {
		report.OPAResult = "allow"
		report.ExportAllowed = true
	} else {
		report.OPAResult = "deny"
		report.BlockedFields = append(report.BlockedFields, policyResult.Reasons...)
	}
	if err := writeReport(sessionDirectory, report); err != nil {
		return dawgtypes.SanitizeReport{}, err
	}
	if !report.ExportAllowed {
		return report, fmt.Errorf("sanitizer: %w by policy %s", dawgtypes.ErrExportBlocked, policyFile)
	}
	return report, nil
}

func writeReport(sessionDirectory string, report dawgtypes.SanitizeReport) error {
	contents, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("sanitizer: serialize report: %w", err)
	}
	contents = append(contents, '\n')
	path := filepath.Join(sessionDirectory, sanitizeReportFileName)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		return fmt.Errorf("sanitizer: write report %s: %w", path, err)
	}
	return nil
}

type fileSanitization struct {
	fieldsScanned int
	redactions    []dawgtypes.Redaction
}

func sanitizeDiagnosticBodies(sessionDirectory string) (fileSanitization, error) {
	root := filepath.Join(sessionDirectory, "diagnostics", "bodies")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return fileSanitization{}, nil
	} else if err != nil {
		return fileSanitization{}, fmt.Errorf("sanitizer: inspect diagnostic bodies: %w", err)
	}
	result := fileSanitization{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return fmt.Errorf("sanitizer: refuse non-regular diagnostic body %s", path)
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if len(contents) > 256<<10 {
			return fmt.Errorf("sanitizer: diagnostic body %s exceeds %d bytes", path, 256<<10)
		}
		relative, _ := filepath.Rel(sessionDirectory, path)
		var value any
		if err := json.Unmarshal(contents, &value); err == nil {
			sanitized, scanned, redactions := sanitizeValue(value, "", filepath.ToSlash(relative), 1)
			updated, err := json.Marshal(sanitized)
			if err != nil {
				return err
			}
			if err := os.WriteFile(path, append(updated, '\n'), 0o600); err != nil {
				return err
			}
			result.fieldsScanned += scanned
			result.redactions = append(result.redactions, redactions...)
			return nil
		}
		if category, matched := classify("diagnostic.body", string(contents)); matched {
			replacementValue := replacement("body", string(contents), category)
			if err := os.WriteFile(path, []byte(replacementValue), 0o600); err != nil {
				return err
			}
			result.fieldsScanned++
			result.redactions = append(result.redactions, dawgtypes.Redaction{File: filepath.ToSlash(relative), Line: 1, Field: "body", Reason: category.reason, Action: redactionAction(category), SyntheticValue: syntheticValue(category, replacementValue)})
		}
		return nil
	})
	if err != nil {
		return fileSanitization{}, err
	}
	return result, nil
}

func sanitizeJSONLFile(path, relativePath string) (fileSanitization, error) {
	input, err := os.Open(path)
	if err != nil {
		return fileSanitization{}, fmt.Errorf("sanitizer: open %s: %w", path, err)
	}
	defer input.Close()

	temporaryPath := path + ".sanitizing"
	output, err := os.OpenFile(temporaryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fileSanitization{}, fmt.Errorf("sanitizer: create temporary file %s: %w", temporaryPath, err)
	}

	result := fileSanitization{}
	scanner := bufio.NewScanner(input)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, dawgtypes.MaxJSONLRecordBytes)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var document any
		if err := json.Unmarshal(line, &document); err != nil {
			output.Close()
			os.Remove(temporaryPath)
			return fileSanitization{}, fmt.Errorf("sanitizer: decode %s line %d: %w", relativePath, lineNumber, err)
		}
		sanitized, scanCount, redactions := sanitizeValue(document, "", relativePath, lineNumber)
		result.fieldsScanned += scanCount
		result.redactions = append(result.redactions, redactions...)
		contents, err := json.Marshal(sanitized)
		if err != nil {
			output.Close()
			os.Remove(temporaryPath)
			return fileSanitization{}, fmt.Errorf("sanitizer: encode %s line %d: %w", relativePath, lineNumber, err)
		}
		if _, err := output.Write(append(contents, '\n')); err != nil {
			output.Close()
			os.Remove(temporaryPath)
			return fileSanitization{}, fmt.Errorf("sanitizer: write %s line %d: %w", relativePath, lineNumber, err)
		}
	}
	if err := scanner.Err(); err != nil {
		output.Close()
		os.Remove(temporaryPath)
		return fileSanitization{}, fmt.Errorf("sanitizer: scan %s: %w", relativePath, err)
	}
	if err := input.Close(); err != nil {
		output.Close()
		os.Remove(temporaryPath)
		return fileSanitization{}, fmt.Errorf("sanitizer: close %s: %w", relativePath, err)
	}
	if err := output.Close(); err != nil {
		os.Remove(temporaryPath)
		return fileSanitization{}, fmt.Errorf("sanitizer: close %s: %w", temporaryPath, err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fileSanitization{}, fmt.Errorf("sanitizer: replace %s: %w", path, err)
	}
	return result, nil
}

func sanitizeValue(value any, path, file string, line int) (any, int, []dawgtypes.Redaction) {
	switch typedValue := value.(type) {
	case map[string]any:
		fieldsScanned := 0
		redactions := []dawgtypes.Redaction{}
		// Preserve rrweb DocumentType names; replacing "html" with a
		// synthetic value makes createDocumentType fail during replay.
		_, hasPublicID := typedValue["publicId"]
		_, hasSystemID := typedValue["systemId"]
		isDocumentTypeNode := hasPublicID && hasSystemID
		for key, child := range typedValue {
			childPath := key
			if path != "" {
				childPath = path + "." + key
			}
			fieldsScanned++
			if key == "name" && isDocumentTypeNode {
				continue
			}
			// Preserve SVG geometry that numeric PII rules could otherwise
			// replace with values Chromium rejects during replay.
			if isSVGGeometryAttribute(path, key) {
				continue
			}
			if stringValue, ok := child.(string); ok {
				if key == "body" {
					if decoded, ok := decodeJSONBody(stringValue); ok {
						sanitized, nestedScanned, nestedRedactions := sanitizeValue(decoded, childPath, file, line)
						fieldsScanned += nestedScanned
						redactions = append(redactions, nestedRedactions...)
						encoded, err := json.Marshal(sanitized)
						if err == nil {
							typedValue[key] = string(encoded)
						}
						continue
					}
				}
				classificationPath := childPath
				if key == "value" {
					for _, contextKey := range []string{"selector", "fieldName", "inputType"} {
						if contextValue, ok := typedValue[contextKey].(string); ok {
							classificationPath += "." + contextValue
						}
					}
				}
				if category, matched := classify(classificationPath, stringValue); matched {
					replacementValue := replacement(childPath, stringValue, category)
					typedValue[key] = replacementValue
					redactions = append(redactions, dawgtypes.Redaction{
						File:           file,
						Line:           line,
						Field:          childPath,
						Reason:         category.reason,
						Action:         redactionAction(category),
						SyntheticValue: syntheticValue(category, replacementValue),
					})
				}
				continue
			}
			sanitized, nestedScanned, nestedRedactions := sanitizeValue(child, childPath, file, line)
			typedValue[key] = sanitized
			fieldsScanned += nestedScanned
			redactions = append(redactions, nestedRedactions...)
		}
		return typedValue, fieldsScanned, redactions
	case []any:
		fieldsScanned := 0
		redactions := []dawgtypes.Redaction{}
		for index, child := range typedValue {
			sanitized, nestedScanned, nestedRedactions := sanitizeValue(child, fmt.Sprintf("%s[%d]", path, index), file, line)
			typedValue[index] = sanitized
			fieldsScanned += nestedScanned
			redactions = append(redactions, nestedRedactions...)
		}
		return typedValue, fieldsScanned, redactions
	default:
		return value, 0, nil
	}
}

func isSVGGeometryAttribute(path, key string) bool {
	return strings.HasSuffix(path, ".attributes") && (key == "viewBox" || key == "points")
}

func decodeJSONBody(value string) (any, bool) {
	var document any
	if err := json.Unmarshal([]byte(value), &document); err != nil {
		return nil, false
	}
	return document, true
}

func redactionAction(category classification) string {
	if category.synthetic {
		return "replaced-synthetic"
	}
	return "redacted"
}

func syntheticValue(category classification, value string) string {
	if category.synthetic {
		return value
	}
	return ""
}
