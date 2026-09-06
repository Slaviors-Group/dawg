import { useEffect, useRef, useState } from "react";
import { useEngine } from "../context/EngineContext";
import { Input } from "./ui/Input";
import { Button } from "./ui/Button";
import { Toggle } from "./ui/Toggle";
import {
  MagnifyingGlass,
  TerminalWindow,
  CopySimple,
  Trash,
  Check,
} from "@phosphor-icons/react";

export function LogStreamer() {
  const { logs, clearLogs } = useEngine();
  const [filter, setFilter] = useState("");
  const [autoScroll, setAutoScroll] = useState(true);
  const [copied, setCopied] = useState(false);
  const logContainerRef = useRef<HTMLDivElement>(null);

  const filteredLogs = logs.filter((log) => log.toLowerCase().includes(filter.toLowerCase()));

  // biome-ignore lint/correctness/useExhaustiveDependencies: auto scroll on log updates
  useEffect(() => {
    if (autoScroll && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [logs.length, autoScroll]);

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(logs.join("\n"));
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="flex flex-col h-72 bg-surface rounded-lg border border-border overflow-hidden shadow-sm">
      {/* Header bar */}
      <div className="flex justify-between items-center px-4 py-2 bg-canvas-subtle border-b border-border">
        <div className="flex items-center gap-2 shrink-0 min-w-0 pr-2">
          <TerminalWindow size={16} className="text-text-secondary shrink-0" />
          <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider truncate">
            Log Console
          </span>
        </div>
        
        <div className="flex items-center gap-2 shrink-0">
          <div className="w-24 sm:w-48">
            <Input
              type="text"
              placeholder="Filter logs..."
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              iconLeft={<MagnifyingGlass size={14} />}
              className="h-7 text-xs bg-canvas text-text-primary placeholder:text-text-disabled"
            />
          </div>
          
          <div className="flex items-center gap-1 sm:gap-2 border-l border-border pl-2 sm:pl-4">
            <div className="flex items-center gap-1.5 mr-1 sm:mr-2" title="Auto-scroll">
              <span className="text-xs text-text-secondary hidden xl:inline">Auto-scroll</span>
              <Toggle checked={autoScroll} onChange={setAutoScroll} />
            </div>
            
            <Button
              variant="ghost"
              size="sm"
              onClick={copyToClipboard}
              iconLeft={copied ? <Check size={14} className="text-success-text" /> : <CopySimple size={14} />}
              className="text-text-secondary hover:text-text-primary h-7 w-7 p-0 flex items-center justify-center"
              title="Copy Logs"
            />
            
            <Button
              variant="ghost"
              size="sm"
              onClick={clearLogs}
              iconLeft={<Trash size={14} />}
              className="text-text-secondary hover:text-error-text hover:bg-error-bg/10 h-7 w-7 p-0 flex items-center justify-center"
              title="Clear Logs"
            />
          </div>
        </div>
      </div>

      {/* Log view area */}
      <div className="flex-1 p-4 overflow-y-auto font-mono text-[11px] leading-relaxed bg-surface text-text-secondary flex flex-col gap-1" ref={logContainerRef}>
        {filteredLogs.length === 0 ? (
          <div className="flex items-center justify-center h-full text-text-disabled italic">
            No log entries matching filter.
          </div>
        ) : (
          filteredLogs.map((log, index) => {
            const isError = log.includes("[ERROR]");
            const isSystem = log.includes("[SYSTEM]");
            return (
              <div
                key={`${index}-${log.slice(0, 10)}`}
                className={[
                  "leading-relaxed",
                  isError
                    ? "text-error-text font-semibold"
                    : isSystem
                      ? "text-brand-400 font-medium"
                      : "text-text-primary",
                ].join(" ")}
              >
                {log}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
