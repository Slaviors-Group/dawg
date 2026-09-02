import type React from "react";

export const CaptureControls: React.FC = () => {
  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 max-w-4xl mx-auto">
      <h2 className="text-2xl font-semibold mb-1">Capture Session Controls</h2>
      <p className="text-slate-400 text-sm mb-6">
        Record browser, HTTP traffic, database diffs, and structured logs
      </p>

      <form onSubmit={(e) => e.preventDefault()} className="flex flex-col gap-2 mb-6">
        <label htmlFor="target-url" className="text-sm font-medium text-slate-300">
          Target Application URL
        </label>
        <div className="flex gap-2">
          <input
            id="target-url"
            type="url"
            placeholder="https://staging.example.com"
            defaultValue="http://localhost:3000"
            disabled
            className="flex-1 bg-slate-900 border border-slate-700 text-slate-100 px-3 py-2 rounded-md disabled:opacity-60"
          />
          <button
            type="button"
            disabled
            className="px-4 py-2 bg-sky-500 text-slate-950 font-semibold rounded-md disabled:opacity-50 cursor-not-allowed"
          >
            Start Capture
          </button>
          <button
            type="button"
            disabled
            className="px-4 py-2 bg-slate-700 text-slate-100 font-semibold rounded-md disabled:opacity-50 cursor-not-allowed"
          >
            Stop Capture
          </button>
        </div>
      </form>

      <div className="mt-6 pt-4 border-t border-slate-700">
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
    </div>
  );
};
