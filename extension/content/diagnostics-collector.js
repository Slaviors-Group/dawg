(() => {
  if (window.__dawgDiagnosticsCollectorInjected) return;
  window.__dawgDiagnosticsCollectorInjected = true;

  const MAX_TEXT = 16 * 1024;
  let active = false;
  let forwarding = false;
  const originals = {};

  function bounded(value) {
    let text;
    try {
      text = typeof value === "string" ? value : JSON.stringify(value);
    } catch (_error) {
      text = String(value);
    }
    return text.length > MAX_TEXT ? `${text.slice(0, MAX_TEXT)}…` : text;
  }

  function argument(value) {
    if (value === null || ["string", "number", "boolean", "undefined"].includes(typeof value)) {
      return { type: typeof value, value: bounded(value), state: "captured" };
    }
    return { type: typeof value, preview: bounded(value), state: "preview-only" };
  }

  function emit(kind, payload) {
    if (!active || forwarding) return;
    forwarding = true;
    try {
      window.dispatchEvent(new CustomEvent("dawg-diagnostic", { detail: { kind, payload } }));
    } finally {
      forwarding = false;
    }
  }

  function installConsole(level) {
    const original = console[level];
    if (typeof original !== "function") return;
    originals[level] = original;
    console[level] = function (...args) {
      // Preserve the page-visible call exactly before observing it.
      const result = Reflect.apply(original, this, args);
      emit("console", {
        level,
        kind: "console-api",
        text: args.map(bounded).join(" "),
        arguments: args.slice(0, 20).map(argument),
        pageUrl: location.href
      });
      return result;
    };
  }

  function start() {
    if (active) return;
    active = true;
    for (const level of ["log", "info", "warn", "error", "debug"]) installConsole(level);
  }

  function stop() {
    active = false;
    for (const [level, original] of Object.entries(originals)) console[level] = original;
    for (const key of Object.keys(originals)) delete originals[key];
  }

  window.addEventListener("error", (event) => emit("error", {
    kind: "exception",
    level: "error",
    text: bounded(event.message || "Uncaught error"),
    pageUrl: location.href,
    location: { url: event.filename || location.href, line: event.lineno || 0, column: event.colno || 0 },
    stack: bounded(event.error?.stack || "")
  }), true);
  window.addEventListener("unhandledrejection", (event) => emit("error", {
    kind: "unhandled-rejection",
    level: "error",
    text: bounded(event.reason),
    pageUrl: location.href,
    stack: bounded(event.reason?.stack || "")
  }));
  window.addEventListener("dawg-diagnostics-control", (event) => {
    if (event.detail?.active) start(); else stop();
  });
})();
