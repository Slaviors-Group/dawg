import {
  ArrowsLeftRight,
  Browser,
  Globe,
  Record,
  StopCircle,
} from "@phosphor-icons/react";
import type React from "react";
import { useEffect, useState } from "react";
import { useEngine } from "../context/EngineContext";
import { LogStreamer } from "./LogStreamer";
import { Badge } from "./ui/Badge";
import { Button } from "./ui/Button";
import { Card, CardHeader, CardTitle } from "./ui/Card";
import { Input } from "./ui/Input";
import { PageShell } from "./ui/PageShell";
import { Separator } from "./ui/Separator";

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
    label: "Browser Extension",
    description: "rrweb DOM and user-action trace",
    variant: "success" as const,
  },
  {
    icon: ArrowsLeftRight,
    label: "Network Metadata",
    description: "Requests from the recorded tab",
    variant: "success" as const,
  },
];

export const CaptureControls: React.FC = () => {
  const {
    isCapturing,
    isStopping,
    targetUrl,
    setTargetUrl,
    startCaptureSession,
    stopCaptureSession,
    activeSessionId,
  } = useEngine();

  const [urlError, setUrlError] = useState("");
  const [packagingProgress, setPackagingProgress] = useState(0);

  useEffect(() => {
    let interval: ReturnType<typeof setInterval>;
    if (isStopping) {
      setPackagingProgress(0);
      interval = setInterval(() => {
        setPackagingProgress((prev) => {
          if (prev < 95) {
            // Slow down as it gets closer to 95%
            const increment = prev > 80 ? 0.5 : 1.2;
            return Math.min(95, prev + increment);
          }
          return prev;
        });
      }, 1000);
    } else {
      setPackagingProgress(0);
    }
    return () => clearInterval(interval);
  }, [isStopping]);

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
      subtitle="Record one browser tab through the installed DAWG extension"
      actions={
        isCapturing ? (
          <Badge variant="error" dot>
            Live session
          </Badge>
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
            hint="DAWG will focus an existing matching tab or open this URL in your browser."
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
              <Button type="submit" variant="primary" size="md" iconLeft={<Record size={14} />}>
                Start Capture
              </Button>
            ) : (
              <Button
                type="button"
                variant="danger"
                size="md"
                onClick={stopCaptureSession}
                disabled={isStopping}
                iconLeft={<StopCircle size={14} />}
                className="shrink-0"
              >
                {isStopping ? "Packaging..." : "Stop Capture"}
              </Button>
            )}
            <div className="flex flex-col flex-1 max-w-md">
              <p className="text-xs text-text-tertiary">
                {isStopping
                  ? "Sanitizing and packaging the artifact — this may take 1–2 minutes for a large session."
                  : isCapturing
                    ? "Capture is running. Stop to finalize and package the artifact."
                    : "Requires the DAWG browser extension to be installed, enabled, and reloaded after updates."}
              </p>
              {isStopping && (
                <div className="mt-2 w-full">
                  <div className="flex justify-between text-[10px] font-medium text-text-tertiary mb-1">
                    <span>Processing telemetry</span>
                    <span>{Math.floor(packagingProgress)}%</span>
                  </div>
                  <div className="h-1.5 w-full bg-canvas-subtle overflow-hidden rounded-full border border-border">
                    <div
                      className="h-full bg-brand-500 transition-all duration-1000 ease-linear"
                      style={{ width: `${packagingProgress}%` }}
                    />
                  </div>
                </div>
              )}
            </div>
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
                <p className="text-sm font-semibold text-text-primary">Session Active</p>
                <p className="text-xs text-text-tertiary font-mono mt-0.5">
                  {activeSessionId ?? "Initializing…"}
                </p>
              </div>
            </div>
            <Badge variant="error" dot>
              Recording telemetry
            </Badge>
          </div>
        </Card>
      )}

      {/* Configuration summary */}
      <div className="flex flex-col gap-4 mt-2">
        <div>
          <h3 className="text-base font-bold text-text-primary">Session Configuration</h3>
          <p className="text-sm text-text-tertiary">Active capture components for this session</p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {CAPTURE_COMPONENTS.map(({ icon: Icon, label, description, variant }) => (
            <div
              key={label}
              className="flex items-start gap-3 p-3 rounded-md bg-canvas-subtle border border-border"
            >
              <div className="w-7 h-7 rounded-sm bg-success-bg border border-success-border flex items-center justify-center shrink-0">
                <Icon size={14} className="text-success-text" />
              </div>
              <div>
                <p className="text-xs font-semibold text-text-primary">{label}</p>
                <p className="text-[11px] text-text-tertiary mt-0.5">{description}</p>
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
