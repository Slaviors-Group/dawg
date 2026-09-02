import { useEffect, useRef, useState } from "react";
import { useEngine } from "../context/EngineContext";

export function LogStreamer() {
  const { logs, clearLogs } = useEngine();
  const [filter, setFilter] = useState("");
  const [autoScroll, setAutoScroll] = useState(true);
  const logContainerRef = useRef<HTMLDivElement>(null);

  const filteredLogs = logs.filter((log) =>
    log.toLowerCase().includes(filter.toLowerCase())
  );

  useEffect(() => {
    if (autoScroll && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  const copyToClipboard = () => {
    navigator.clipboard.writeText(logs.join("\n"));
  };

  return (
    <div className="bg-slate-900 border border-slate-700 rounded-lg flex flex-col h-72">
      <div className="flex justify-between items-center px-4 py-2 bg-slate-800/80 border-b border-slate-700">
        <div className="flex items-center gap-2">
          <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse" />
          <span className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
            Engine Log Console
          </span>
        </div>
        <div className="flex items-center gap-2">
          <input
            type="text"
            placeholder="Filter logs..."
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="bg-slate-950 border border-slate-700 text-xs px-2 py-1 rounded text-slate-200 focus:outline-none focus:border-sky-500"
          />
          <button
            type="button"
            onClick={() => setAutoScroll(!autoScroll)}
            className={`text-xs px-2 py-1 rounded border ${
              autoScroll
                ? "bg-sky-500/20 border-sky-500 text-sky-300"
                : "bg-slate-800 border-slate-700 text-slate-400"
            }`}
          >
            Auto-Scroll
          </button>
          <button
            type="button"
            onClick={copyToClipboard}
            className="text-xs px-2 py-1 bg-slate-800 hover:bg-slate-700 border border-slate-700 rounded text-slate-300 transition-colors"
          >
            Copy
          </button>
          <button
            type="button"
            onClick={clearLogs}
            className="text-xs px-2 py-1 bg-slate-800 hover:bg-slate-700 border border-slate-700 rounded text-slate-400 hover:text-slate-200 transition-colors"
          >
            Clear
          </button>
        </div>
      </div>

      <div
        ref={logContainerRef}
        className="flex-1 p-3 overflow-y-auto font-mono text-xs text-slate-300 space-y-1 select-text"
      >
        {filteredLogs.length === 0 ? (
          <div className="text-slate-500 italic">No log entries matching filter.</div>
        ) : (
          filteredLogs.map((log, index) => {
            const isError = log.includes("[ERROR]");
            const isSystem = log.includes("[SYSTEM]");
            return (
              <div
                key={`${index}-${log.slice(0, 10)}`}
                className={`leading-relaxed ${
                  isError ? "text-rose-400 font-semibold" : isSystem ? "text-sky-300" : "text-slate-300"
                }`}
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
