import type React from "react";
import { useState } from "react";

const POLICIES = {
  "default.rego": {
    name: "Default Gate Policy (default.rego)",
    description: "Standard OPA export gate. Refuses packaging if any unresolved violation exists.",
    code: `package dawg.sanitize

default allow = false

# Hard export gate rules
allow {
    count(input.violations) == 0
}`,
  },
  "strict-pci.rego": {
    name: "Strict PCI Policy (strict-pci.rego)",
    description: "Enforces strict PCI-DSS secret redaction (credit cards, CVVs, Bearer tokens).",
    code: `package dawg.sanitize

default allow = false

# Strict PCI redaction rule
allow {
    count(input.violations) == 0
    count(input.unredactedPciTokens) == 0
}`,
  },
  "relaxed-dev.rego": {
    name: "Dev Relaxed Policy (relaxed-dev.rego)",
    description: "Dev-only policy allowing localhost HTTP traffic without strict export blocking.",
    code: `package dawg.sanitize

default allow = true

# Warn on unredacted tokens but allow local dev exports
warn_only {
    input.isLocalhost == true
}`,
  },
};

export const PolicyConfig: React.FC = () => {
  const [selectedPolicyKey, setSelectedPolicyKey] = useState<keyof typeof POLICIES>("default.rego");
  const activePolicy = POLICIES[selectedPolicyKey];

  const copyPolicyCode = () => {
    navigator.clipboard.writeText(activePolicy.code);
  };

  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 max-w-4xl mx-auto space-y-6">
      <div>
        <h2 className="text-2xl font-semibold mb-1">Sanitizer Policy Configuration</h2>
        <p className="text-slate-400 text-sm">
          OPA Rego Policy Gating &amp; Secret Redaction Rules (PRD §6.2 / architecture.md §Schema)
        </p>
      </div>

      <div className="flex flex-col gap-2">
        <label htmlFor="policy-preset" className="text-xs font-medium text-slate-300">
          Select Policy Preset
        </label>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {(Object.keys(POLICIES) as Array<keyof typeof POLICIES>).map((key) => {
            const isSelected = key === selectedPolicyKey;
            return (
              <button
                key={key}
                type="button"
                onClick={() => setSelectedPolicyKey(key)}
                className={`p-3 rounded-lg border text-left transition-all ${
                  isSelected
                    ? "bg-slate-700 border-sky-400 text-slate-100 shadow-md"
                    : "bg-slate-900/50 border-slate-700 text-slate-400 hover:text-slate-200 hover:border-slate-600"
                }`}
              >
                <div className="font-mono text-xs font-semibold">{key}</div>
                <div className="text-[11px] text-slate-400 mt-1 line-clamp-2">
                  {POLICIES[key].name}
                </div>
              </button>
            );
          })}
        </div>
      </div>

      <div className="pt-4 border-t border-slate-700 space-y-3">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-medium text-slate-100">{activePolicy.name}</h3>
            <p className="text-xs text-slate-400">{activePolicy.description}</p>
          </div>
          <button
            type="button"
            onClick={copyPolicyCode}
            className="px-3 py-1.5 bg-slate-900 hover:bg-slate-700 border border-slate-700 rounded text-xs text-slate-200 transition-colors"
          >
            Copy Rego Code
          </button>
        </div>

        <pre className="bg-slate-950 border border-slate-800 p-4 rounded-lg font-mono text-xs text-slate-200 overflow-x-auto">
          {activePolicy.code}
        </pre>
      </div>
    </div>
  );
};
