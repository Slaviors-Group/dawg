import type React from "react";
import { useState } from "react";
import { PageShell } from "./ui/PageShell";
import { Card, CardHeader, CardTitle } from "./ui/Card";
import { CodeBlock } from "./ui/CodeBlock";
import { CheckCircle } from "@phosphor-icons/react";

const POLICIES = {
  "default.rego": {
    name: "Default Gate Policy",
    description: "Standard OPA export gate. Refuses packaging if any unresolved violation exists.",
    code: `package dawg.sanitize

default allow = false

# Hard export gate rules
allow {
    count(input.violations) == 0
}`,
  },
  "strict-pci.rego": {
    name: "Strict PCI Policy",
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
    name: "Dev Relaxed Policy",
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

  return (
    <PageShell
      title="Sanitizer Policy Configuration"
      subtitle="OPA Rego Policy Gating & Secret Redaction Rules (PRD §6.2 / architecture.md §Schema)"
    >
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {(Object.keys(POLICIES) as Array<keyof typeof POLICIES>).map((key) => {
          const isSelected = key === selectedPolicyKey;
          const policy = POLICIES[key];

          return (
            <button
              key={key}
              type="button"
              onClick={() => setSelectedPolicyKey(key)}
              className={[
                "text-left p-4 rounded-lg border transition-all duration-[--duration-fast]",
                "flex flex-col gap-1 relative",
                isSelected
                  ? "bg-brand-50 border-brand-400 shadow-sm"
                  : "bg-surface border-border hover:border-brand-300 hover:bg-surface-hover",
              ].join(" ")}
            >
              <span
                className={[
                  "text-xs font-mono font-semibold",
                  isSelected ? "text-brand-700" : "text-text-secondary",
                ].join(" ")}
              >
                {key}
              </span>
              <span
                className={[
                  "text-[11px] leading-snug line-clamp-2",
                  isSelected ? "text-brand-800" : "text-text-tertiary",
                ].join(" ")}
              >
                {policy.name}
              </span>
              {isSelected && (
                <CheckCircle
                  size={18}
                  weight="fill"
                  className="absolute top-4 right-4 text-brand-500"
                />
              )}
            </button>
          );
        })}
      </div>

      <Card>
        <CardHeader>
          <CardTitle
            title={activePolicy.name}
            subtitle={activePolicy.description}
          />
        </CardHeader>
        
        <CodeBlock code={activePolicy.code} language="rego" />
      </Card>
    </PageShell>
  );
};
