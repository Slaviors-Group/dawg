import type React from "react";
import { useEffect, useState } from "react";
import { useEngine } from "../context/EngineContext";
import { LogStreamer } from "./LogStreamer";
import { PageShell } from "./ui/PageShell";
import { Card, CardHeader, CardTitle } from "./ui/Card";
import { Input } from "./ui/Input";
import { Button } from "./ui/Button";
import { Badge } from "./ui/Badge";
import { Separator } from "./ui/Separator";
import {
  Globe,
  Record,
  StopCircle,
  Browser,
  ArrowsLeftRight,
  Database,
  FileText,
} from "@phosphor-icons/react";

const URL_PRESETS = [
  "http://localhost:5173",
  "http://localhost:3000",
  "http://localhost:4200",
  "http://localhost:4321",
  "http://localhost:8000",
  "http://localhost:8080",
];

const CAPTURE_COMPONENTS = [
  {
    icon: Browser,
    label: "Browser Capture",
    description: "Playwright + rrweb DOM trace",
    variant: "success" as const,
  },
  {
    icon: ArrowsLeftRight,
    label: "HTTP Proxy",
    description: "mitmproxy cassette capture",
    variant: "success" as const,
  },
  {
    icon: Database,
    label: "Database Tap",
    description: "ORM diff collector",
    variant: "success" as const,
  },
  {
    icon: FileText,
    label: "Log Capture",
    description: "Application stdout / JSONL stream",
    variant: "success" as const,
  },
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

  const [urlError, setUrlError] = useState("");

  useEffect(() => {
    const savedUrl = localStorage.getItem("dawg_last_target_url");
    if (savedUrl) setTargetUrl(savedUrl);
  }, [setTargetUrl]);

  const handleUrlChange = (url: string) => {
    setTargetUrl(url);
    setUrlError("");
    localStorage.setItem("dawg_last_target_url", url);
  };

  const handleStart = (e: React.FormEvent) => {
    e.preventDefault();
    if (!targetUrl) {
      setUrlError("Target URL is required before starting a session.");
      return;
    }
    if (!targetUrl.startsWith("http://") && !targetUrl.startsWith("https://")) {
      setUrlError("URL must start with http:// or https://");
      return;
    }
    startCaptureSession(targetUrl);
  };

  return (
    <PageShell
      title="Capture"
      subtitle="Record browser activity, HTTP traffic, database diffs, and structured logs"
      actions={
        isCapturing ? (
          <Badge variant="error" dot>Live session</Badge>
        ) : undefined
      }
    >
      {/* Session controls */}
      <Card>
        <CardHeader>
          <CardTitle
            title="Target Application"
            subtitle="Enter the URL of the app you want to capture"
          />
        </CardHeader>

        <form onSubmit={handleStart} className="flex flex-col gap-4">
          <Input
            id="target-url"
            type="url"
            label="Target URL"
            placeholder="https://staging.example.com"
            value={targetUrl}
            onChange={(e) => handleUrlChange(e.target.value)}
            disabled={isCapturing}
            error={urlError}
            iconLeft={<Globe size={14} />}
            hint="The DAWG proxy will intercept all requests through this origin."
          />

          {/* URL presets */}
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-xs text-text-tertiary">Presets:</span>
            {URL_PRESETS.map((preset) => (
              <button
                key={preset}
                type="button"
                disabled={isCapturing}
                onClick={() => handleUrlChange(preset)}
                className={[
                  "px-2 py-0.5 rounded-sm border text-[11px] font-mono transition-all duration-[--duration-fast]",
                  "disabled:opacity-50 disabled:cursor-not-allowed",
                  targetUrl === preset
                    ? "bg-brand-100 border-brand-300 text-brand-700"
                    : "bg-canvas-subtle border-border text-text-tertiary hover:text-text-primary hover:border-border-strong",
                ].join(" ")}
              >
                {preset.replace("http://", "").replace("https://", "")}
              </button>
            ))}
          </div>

          <Separator />

          {/* Action */}
          <div className="flex items-center gap-3">
            {!isCapturing ? (
              <Button
                type="submit"
                variant="primary"
                size="md"
                iconLeft={<Record size={14} />}
              >
                Start Capture
              </Button>
            ) : (
              <Button
                type="button"
                variant="danger"
                size="md"
                onClick={stopCaptureSession}
                iconLeft={<StopCircle size={14} />}
              >
                Stop Capture
              </Button>
            )}
            <p className="text-xs text-text-tertiary">
              {isCapturing
                ? "Capture is running. Stop to finalize and package the artifact."
                : "Starts the proxy, browser agent, and log tap simultaneously."}
            </p>
          </div>
        </form>
      </Card>

      {/* Live session banner */}
      {isCapturing && (
        <Card accent>
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <span className="w-2.5 h-2.5 rounded-full bg-error-dot animate-ping shrink-0" />
              <div>
                <p className="text-sm font-semibold text-text-primary">
                  Session Active
                </p>
                <p className="text-xs text-text-tertiary font-mono mt-0.5">
                  {activeSessionId ?? "Initializing…"}
                </p>
              </div>
            </div>
            <Badge variant="error" dot>Recording telemetry</Badge>
          </div>
        </Card>
      )}

      {/* Configuration summary */}
      <div className="flex flex-col gap-4 mt-2">
        <div>
          <h3 className="text-base font-bold text-text-primary">Session Configuration</h3>
          <p className="text-sm text-text-tertiary">Active capture components for this session</p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-3">
          {CAPTURE_COMPONENTS.map(({ icon: Icon, label, description, variant }) => (
            <div
              key={label}
              className="flex items-start gap-3 p-3 rounded-md bg-canvas-subtle border border-border"
            >
              <div className="w-7 h-7 rounded-sm bg-success-bg border border-success-border flex items-center justify-center shrink-0">
                <Icon size={14} className="text-success-text" />
              </div>
              <div>
                <p className="text-xs font-semibold text-text-primary">
                  {label}
                </p>
                <p className="text-[11px] text-text-tertiary mt-0.5">
                  {description}
                </p>
              </div>
              <Badge variant={variant} size="sm" className="ml-auto shrink-0">
                on
              </Badge>
            </div>
          ))}
        </div>
      </div>

      {/* Log stream */}
      <LogStreamer />
    </PageShell>
  );
};
