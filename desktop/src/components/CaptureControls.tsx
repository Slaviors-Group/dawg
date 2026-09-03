import type React from "react";
import { useEffect } from "react";
import { useEngine } from "../context/EngineContext";
import { LogStreamer } from "./LogStreamer";

const URL_PRESETS = [
  "http://localhost:3000",
  "http://localhost:8080",
  "http://localhost:5173",
  "https://staging.example.com",
];

export const CaptureControls: React.FC = () => {
  const {
    isCapturing,
    targetUrl,
    setTargetUrl,
    startCaptureSession,
    stopCaptureSession,
    activeSessionId,
  } = useEngine();

  useEffect(() => {
    const savedUrl = localStorage.getItem("dawg_last_target_url");
    if (savedUrl) {
      setTargetUrl(savedUrl);
    }
  }, [setTargetUrl]);

  const handleUrlChange = (url: string) => {
    setTargetUrl(url);
    localStorage.setItem("dawg_last_target_url", url);
  };

  const handleStart = (e: React.FormEvent) => {
    e.preventDefault();
    if (!isCapturing && targetUrl) {
      startCaptureSession(targetUrl);
    }
  };

  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 max-w-4xl mx-auto space-y-6">
      <div>
        <h2 className="text-2xl font-semibold mb-1">Capture Session Controls</h2>
        <p className="text-slate-400 text-sm">
          Record browser, HTTP traffic, database diffs, and structured logs
        </p>
      </div>

      <form onSubmit={handleStart} className="flex flex-col gap-3">
        <label htmlFor="target-url" className="text-sm font-medium text-slate-300">
          Target Application URL
        </label>
        <div className="flex gap-2">
          <input
            id="target-url"
            type="url"
            placeholder="https://staging.example.com"
            value={targetUrl}
            onChange={(e) => handleUrlChange(e.target.value)}
            disabled={isCapturing}
            className="flex-1 bg-slate-900 border border-slate-700 text-slate-100 px-3 py-2 rounded-md disabled:opacity-60 focus:outline-none focus:border-sky-500 text-sm font-mono"
          />
          {!isCapturing ? (
            <button
              type="submit"
              className="px-4 py-2 bg-sky-500 hover:bg-sky-400 text-slate-950 font-semibold rounded-md transition-colors text-sm"
            >
              Start Capture
            </button>
          ) : (
            <button
              type="button"
              onClick={stopCaptureSession}
              className="px-4 py-2 bg-rose-500 hover:bg-rose-400 text-white font-semibold rounded-md transition-colors animate-pulse text-sm"
            >
              Stop Capture
            </button>
          )}
        </div>

        <div className="flex items-center gap-2 text-xs">
          <span className="text-slate-400">Presets:</span>
          {URL_PRESETS.map((preset) => (
            <button
              key={preset}
              type="button"
              disabled={isCapturing}
              onClick={() => handleUrlChange(preset)}
              className={`px-2 py-0.5 rounded border text-[11px] font-mono transition-colors ${
                targetUrl === preset
                  ? "bg-sky-500/20 border-sky-500 text-sky-300"
                  : "bg-slate-900 border-slate-700 text-slate-400 hover:text-slate-200"
              }`}
            >
              {preset.replace("http://", "").replace("https://", "")}
            </button>
          ))}
        </div>
      </form>

      {isCapturing && (
        <div className="bg-emerald-500/10 border border-emerald-500/40 rounded-md p-3 flex items-center justify-between text-xs text-emerald-300">
          <div className="flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-ping" />
            <span>
              Active Session ID:{" "}
              <strong className="font-mono">{activeSessionId || "Initializing..."}</strong>
            </span>
          </div>
          <span>Recording live telemetry stream...</span>
        </div>
      )}

      <div className="pt-4 border-t border-slate-700">
        <h3 className="text-lg font-medium mb-3">Session Configuration</h3>
        <ul className="list-disc pl-5 text-slate-400 space-y-1 text-sm">
          <li>
            <strong className="text-slate-200">Browser Capture:</strong> Playwright + rrweb trace
            enabled
          </li>
          <li>
            <strong className="text-slate-200">HTTP Proxy:</strong> mitmproxy cassette capture
            enabled
          </li>
          <li>
            <strong className="text-slate-200">Database Tap:</strong> ORM diff collector enabled
          </li>
          <li>
            <strong className="text-slate-200">Log Capture:</strong> Application stdout / JSONL
            stream enabled
          </li>
        </ul>
      </div>

      <div className="pt-4 border-t border-slate-700">
        <LogStreamer />
      </div>
    </div>
  );
};
