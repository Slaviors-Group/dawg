import type React from "react";

export const PolicyConfig: React.FC = () => {
  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 max-w-4xl mx-auto">
      <h2 className="text-2xl font-semibold mb-1">Sanitizer Policy Configuration</h2>
      <p className="text-slate-400 text-sm mb-6">
        OPA Rego Policy Gating &amp; Secret Redaction Rules
      </p>

      <div className="mt-6 pt-4 border-t border-slate-700">
        <h3 className="text-lg font-medium mb-3">Active Policy: default.rego</h3>
        <pre className="bg-slate-900 border border-slate-700 p-4 rounded-md font-mono text-sm text-slate-200 overflow-x-auto">
          {`package dawg.sanitize

default allow = false

# Hard export gate rules
allow {
    count(input.violations) == 0
}`}
        </pre>
      </div>
    </div>
  );
};
