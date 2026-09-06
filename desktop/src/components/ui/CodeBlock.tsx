/**
 * CodeBlock — monospaced read-only code/log surface.
 * Dark background regardless of theme (like a terminal).
 * Supports copy-to-clipboard action.
 */
import { useState } from "react";
import { CopySimple, Check } from "@phosphor-icons/react";

interface CodeBlockProps {
  code: string;
  language?: string;
  maxHeight?: string;
  showCopy?: boolean;
  className?: string;
}

export function CodeBlock({
  code,
  language,
  maxHeight = "20rem",
  showCopy = true,
  className = "",
}: CodeBlockProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 1800);
  };

  return (
    <div className={["relative group rounded-md overflow-hidden", className].join(" ")}>
      {/* Header bar */}
      <div className="flex items-center justify-between bg-[hsl(240,15%,10%)] px-3 py-1.5 border-b border-[hsl(240,10%,18%)]">
        <span className="text-[10px] font-mono text-[hsl(240,6%,48%)] uppercase tracking-wider">
          {language ?? "code"}
        </span>
        {showCopy && (
          <button
            type="button"
            onClick={handleCopy}
            aria-label={copied ? "Copied!" : "Copy to clipboard"}
            className={[
              "flex items-center gap-1 text-[10px] font-medium px-2 py-0.5 rounded",
              "transition-all duration-[--duration-fast]",
              copied
                ? "text-success-dot bg-success-bg/20"
                : "text-[hsl(240,6%,55%)] hover:text-[hsl(240,6%,80%)] hover:bg-[hsl(240,10%,20%)]",
            ].join(" ")}
          >
            {copied ? (
              <Check size={10} weight="bold" />
            ) : (
              <CopySimple size={10} />
            )}
            {copied ? "Copied" : "Copy"}
          </button>
        )}
      </div>
      {/* Code surface */}
      <pre
        style={{ maxHeight }}
        className={[
          "overflow-auto p-4 m-0 text-xs leading-relaxed font-[--font-mono]",
          "bg-[hsl(240,18%,8%)] text-[hsl(142,60%,68%)]",
        ].join(" ")}
      >
        <code>{code}</code>
      </pre>
    </div>
  );
}
