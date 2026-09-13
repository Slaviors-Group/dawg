package sanitize

import (
	"context"
	"fmt"
	"os"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/open-policy-agent/opa/rego"
)

// EvaluatePolicy evaluates the DAWG sanitizer allow rule against one report input.
func EvaluatePolicy(ctx context.Context, policyPath string, input map[string]any) (dawgtypes.OPAResult, error) {
	policy, err := os.ReadFile(policyPath)
	if err != nil {
		return dawgtypes.OPAResult{}, fmt.Errorf("sanitizer: read policy %s: %w", policyPath, err)
	}
	query, err := rego.New(
		rego.Query("data.dawg.sanitizer.allow"),
		rego.Module(policyPath, string(policy)),
	).PrepareForEval(ctx)
	if err != nil {
		return dawgtypes.OPAResult{}, fmt.Errorf("sanitizer: prepare policy %s: %w", policyPath, err)
	}

	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return dawgtypes.OPAResult{}, fmt.Errorf("sanitizer: evaluate policy %s: %w", policyPath, err)
	}
	allowed := false
	if len(results) > 0 && len(results[0].Expressions) > 0 {
		allowed, _ = results[0].Expressions[0].Value.(bool)
	}
	result := dawgtypes.OPAResult{Allowed: allowed, PolicyFile: policyPath}
	if !allowed {
		result.Reasons = []string{"policy denied export"}
	}
	return result, nil
}
