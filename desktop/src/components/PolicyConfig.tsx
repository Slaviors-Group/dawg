import type React from "react";

export const PolicyConfig: React.FC = () => {
  return (
    <div className="panel">
      <h2>Sanitizer Policy Configuration</h2>
      <p className="subtitle">OPA Rego Policy Gating &amp; Secret Redaction Rules</p>

      <div className="section">
        <h3>Active Policy: default.rego</h3>
        <pre className="code-block">
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
