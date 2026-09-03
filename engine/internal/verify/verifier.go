package verify

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

// Verifier executes checks against a replay outcome to determine reproduction status.
type Verifier struct{}

// Verify compares the ReplayOutput against expected baselines and assertions.
func (v *Verifier) Verify(ctx context.Context, replay dawgtypes.ReplayOutput, originalSessionDir string, expected dawgtypes.ExpectedOutcome) (dawgtypes.VerifyResult, error) {
	result := dawgtypes.VerifyResult{
		ArtifactID:      replay.ArtifactID,
		VerifiedAt:      time.Now(),
		VerifiedAgainst: "local",
		Result:          "pass", // default to pass until a check fails
		Checks:          []dawgtypes.Check{},
	}

	// 1. Exit Code Check
	exitCheck := dawgtypes.Check{
		Type:     dawgtypes.CheckTypeExitCode,
		Expected: 0,
		Actual:   replay.Outcomes.ExitCode,
		Passed:   replay.Outcomes.ExitCode == 0,
	}
	result.Checks = append(result.Checks, exitCheck)

	// 2. HTTP Response Diff Check
	// Compare replayed frontend.jsonl with original frontend.jsonl
	expectedHTTPPath := filepath.Join(originalSessionDir, "http", "frontend.jsonl")
	actualHTTPPath := replay.Outcomes.HTTPResponses
	if actualHTTPPath != "" {
		mismatches, err := CompareHTTPResponses(expectedHTTPPath, actualHTTPPath)
		if err == nil {
			httpCheck := dawgtypes.Check{
				Type:     dawgtypes.CheckTypeHTTPResponse,
				Expected: "0 mismatches",
				Actual:   fmt.Sprintf("%d mismatches", mismatches),
				Passed:   mismatches == 0,
			}
			result.Checks = append(result.Checks, httpCheck)
		}
	}

	// 3. Screenshot Diff Check (if original screenshot existed)
	// For MVP, if no expected assertion screenshot, we skip image diff.
	if len(replay.Outcomes.Screenshots) > 0 && expected.AssertionFile != "" {
		baselinePath := filepath.Join(originalSessionDir, expected.AssertionFile)
		actualScreenshot := replay.Outcomes.Screenshots[0]
		diffPixels, err := CompareImages(baselinePath, actualScreenshot)
		if err == nil {
			threshold := 100 // Arbitrary MVP threshold
			screenshotCheck := dawgtypes.Check{
				Type:       dawgtypes.CheckTypeScreenshot,
				File:       actualScreenshot,
				DiffPixels: diffPixels,
				Threshold:  threshold,
				Passed:     diffPixels <= threshold,
			}
			result.Checks = append(result.Checks, screenshotCheck)
		}
	}

	// Calculate overall result
	passedCount := 0
	for _, check := range result.Checks {
		if check.Passed {
			passedCount++
		} else {
			result.Result = "fail"
		}
	}

	result.Summary = fmt.Sprintf("%d/%d checks passed", passedCount, len(result.Checks))
	if result.Result == "fail" {
		result.Summary += " — bug still reproduces"
	} else {
		result.Summary += " — bug no longer reproduces (fix verified)"
	}

	return result, nil
}
