import { ArrowsLeftRight, Browser, Globe, Record, StopCircle } from "@phosphor-icons/react";
import type React from "react";
import { useEffect, useState } from "react";
import { useEngine } from "../context/EngineContext";
import type { CaptureDiagnosticsProfile } from "../lib/engine";
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

const DIAGNOSTICS_PROFILES: Array<{
  id: CaptureDiagnosticsProfile;
  title: string;
  description: string;
}> = [
  {
    id: "safe",
    title: "Safe",
    description: "Sanitized data for replay.",
  },
  {
    id: "enhanced",
    title: "Enhanced Diagnostics",
    description: "More evidence for troubleshooting.",
  },
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
  const [artifactTitle, setArtifactTitle] = useState("");
  const [diagnosticsProfile, setDiagnosticsProfile] = useState<CaptureDiagnosticsProfile>("safe");
  const [enhancedConsent, setEnhancedConsent] = useState(false);
  const [profileError, setProfileError] = useState("");
  const [packagingProgress, setPackagingProgress] = useState(0);

  useEffect(() => {
    let interval: ReturnType<typeof setInterval>;
    if (isStopping) {
      setPackagingProgress(0);
      interval = setInterval(() => {
        setPackagingProgress((prev) => {
          if (prev < 95) {
            // Reserve completion for the engine's stop response.
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

  const handleDiagnosticsProfileChange = (profile: CaptureDiagnosticsProfile) => {
    setDiagnosticsProfile(profile);
    setEnhancedConsent(false);
    setProfileError("");
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
    if (diagnosticsProfile === "enhanced" && !enhancedConsent) {
      setProfileError("Confirm the enhanced diagnostics warning before starting.");
      return;
    }
    void startCaptureSession({
      url: targetUrl,
      title: artifactTitle.trim() || undefined,
      diagnosticsProfile,
    });
  };

  return (
    <PageShell
      title="Capture"
      subtitle="Record a browser session with the DAWG extension."
      actions={
        isCapturing ? (
          <Badge variant="error" dot>
            Live session
          </Badge>
        ) : undefined
      }
    >
      <output className="rounded-2xl border border-warning-border bg-warning-bg px-4 py-3 text-sm text-warning-text">
        <p className="font-semibold">Use Chrome, Edge, or another Chromium browser to record.</p>
      </output>

      {/* Session controls */}
      <Card>
        <CardHeader>
          <CardTitle title="Target application" />
        </CardHeader>

        <form onSubmit={handleStart} className="flex flex-col gap-4">
          <div data-tour="capture-target" className="grid grid-cols-1 gap-4 lg:grid-cols-2">
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
              hint="DAWG opens or focuses this URL."
              wrapperClassName="min-w-0"
            />

            <Input
              id="artifact-title"
              label="Artifact name (optional)"
              placeholder="Checkout validation regression"
              value={artifactTitle}
              onChange={(event) => setArtifactTitle(event.target.value)}
              disabled={isCapturing}
              hint="Leave blank to name it automatically."
              wrapperClassName="min-w-0"
            />
          </div>

          <fieldset disabled={isCapturing} className="flex flex-col gap-2">
            <legend className="text-xs font-medium text-text-secondary">Capture profile</legend>
            <p className="text-xs text-text-tertiary">Safe is the default.</p>
            <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
              {DIAGNOSTICS_PROFILES.map((profile) => {
                const isSelected = diagnosticsProfile === profile.id;
                return (
                  <label
                    key={profile.id}
                    data-tour={`capture-profile-${profile.id}`}
                    className={[
                      "flex cursor-pointer gap-3 rounded-xl border p-4 transition-colors",
                      isSelected
                        ? "border-brand-400 bg-brand-500/10"
                        : "border-border bg-canvas-subtle hover:border-text-tertiary",
                      "has-disabled:cursor-not-allowed has-disabled:opacity-50",
                    ].join(" ")}
                  >
                    <input
                      type="radio"
                      name="diagnostics-profile"
                      value={profile.id}
                      checked={isSelected}
                      onChange={() => handleDiagnosticsProfileChange(profile.id)}
                      className="mt-0.5 accent-[--color-brand-500]"
                    />
                    <span>
                      <span className="block text-sm font-medium text-text-primary">
                        {profile.title}
                      </span>
                      <span className="mt-0.5 block text-xs text-text-tertiary">
                        {profile.description}
                      </span>
                    </span>
                  </label>
                );
              })}
            </div>

            {diagnosticsProfile === "enhanced" && (
              <div className="rounded-xl border border-warning-border bg-warning-bg p-4">
                <p className="text-xs font-semibold text-warning-text">
                  Enhanced Diagnostics warning
                </p>
                <p className="mt-1 text-xs text-warning-text">
                  This may retain extra evidence. Check your data policy before recording.
                </p>
                <label className="mt-3 flex cursor-pointer items-start gap-2 text-xs text-warning-text">
                  <input
                    type="checkbox"
                    checked={enhancedConsent}
                    onChange={(event) => {
                      setEnhancedConsent(event.target.checked);
                      setProfileError("");
                    }}
                    className="mt-0.5 accent-[--color-brand-500]"
                  />
                  <span>I understand and agree to enhanced diagnostics.</span>
                </label>
                {profileError && (
                  <p className="mt-2 text-xs text-error-text" role="alert">
                    {profileError}
                  </p>
                )}
              </div>
            )}
          </fieldset>

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
                  "rounded-full border px-3 py-1 text-[11px] font-mono transition-all duration-[--duration-fast]",
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
                data-tour="capture-action"
                type="submit"
                variant="primary"
                size="md"
                disabled={diagnosticsProfile === "enhanced" && !enhancedConsent}
                iconLeft={<Record size={14} />}
              >
                Start Capture
              </Button>
            ) : (
              <Button
                data-tour="capture-action"
                type="button"
                variant="danger"
                size="md"
                onClick={stopCaptureSession}
                disabled={isStopping}
                iconLeft={<StopCircle size={14} />}
                className="shrink-0"
              >
                {isStopping ? "Packaging..." : "Stop capture"}
              </Button>
            )}
            <div className="flex flex-col flex-1 max-w-md">
              <p className="text-xs text-text-tertiary">
                {isStopping
                  ? "Packaging the artifact. Large sessions may take 1 to 2 minutes."
                  : isCapturing
                    ? "Stop capture to save the artifact."
                    : "Requires the DAWG browser extension."}
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
        <Card>
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
          <h3 className="text-base font-semibold text-text-primary">Capture components</h3>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {CAPTURE_COMPONENTS.map(({ icon: Icon, label, description, variant }) => (
            <div
              key={label}
              className="flex items-start gap-3 rounded-2xl border border-border bg-surface p-4 shadow-card"
            >
              <div className="w-8 h-8 rounded-xl bg-brand-100 flex items-center justify-center shrink-0">
                <Icon size={15} className="text-brand-600" />
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
