(function () {
  if (window.__dawgPanelInjected) return;
  window.__dawgPanelInjected = true;

  const FONT_WEIGHTS = [400, 500, 600, 700];

  function loadOutfitFonts() {
    if (typeof FontFace !== "function" || !document.fonts)
      return Promise.resolve();
    return Promise.all(
      FONT_WEIGHTS.map(async (weight) => {
        try {
          const url = chrome.runtime.getURL(`popup/outfit-${weight}.ttf`);
          const face = new FontFace("Outfit", `url(${url})`, {
            weight: String(weight),
            style: "normal",
          });
          const loaded = await face.load();
          document.fonts.add(loaded);
        } catch (_error) {}
      }),
    );
  }

  const host = document.createElement("div");
  host.id = "dawg-panel-host";
  host.style.cssText =
    "all:initial;position:fixed;top:12px;right:12px;z-index:2147483647;display:none;";
  const root = host.attachShadow({ mode: "open" });

  root.innerHTML = `
    <style>
      :host { all: initial; }
      * { box-sizing: border-box; margin: 0; padding: 0; }
      .card {
        width: 340px;
        display: flex;
        flex-direction: column;
        gap: 18px;
        padding: 16px;
        background: hsl(0 0% 100%);
        border: 1px solid hsl(248 20% 90%);
        border-radius: 18px;
        box-shadow: 0 12px 32px -6px hsl(240 15% 12% / 0.28),
          0 0 3px hsl(240 15% 12% / 0.08);
        font-family: "Outfit", ui-sans-serif, system-ui, -apple-system,
          BlinkMacSystemFont, "Segoe UI", sans-serif;
        color: hsl(240 15% 12%);
      }
      .header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
      }
      .logo { display: flex; align-items: center; min-width: 0; gap: 10px; }
      .logo .icon { width: 28px; height: 28px; flex: none; }
      .logo h1 {
        color: hsl(240 15% 12%);
        font-size: 16px;
        font-weight: 700;
        letter-spacing: -0.02em;
      }
      .badge {
        padding: 5px 9px;
        border: 1px solid transparent;
        border-radius: 9999px;
        font-size: 10px;
        font-weight: 700;
        letter-spacing: 0.06em;
        line-height: 1;
        text-transform: uppercase;
      }
      .status-idle {
        background: hsl(258 70% 95%);
        border-color: hsl(258 70% 90%);
        color: hsl(258 65% 45%);
      }
      .status-recording {
        background: hsl(348 80% 96%);
        border-color: hsl(348 70% 88%);
        color: hsl(348 80% 40%);
        animation: pulse 1.5s infinite;
      }
      @keyframes pulse {
        0%, 100% { opacity: 1; }
        50% { opacity: 0.65; }
      }
      .content { display: flex; flex-direction: column; gap: 12px; }
      .info-box {
        padding: 12px;
        background: hsl(248 40% 95%);
        border: 1px solid hsl(248 20% 90%);
        border-radius: 12px;
        color: hsl(240 15% 35%);
        font-size: 12px;
        line-height: 1.5;
      }
      .info-box p[style] { color: hsl(348 80% 40%) !important; }
      .actions { display: flex; gap: 8px; margin-top: 2px; }
      .btn {
        width: 100%;
        padding: 10px 12px;
        border: 1px solid transparent;
        border-radius: 8px;
        color: #fff;
        cursor: pointer;
        font: inherit;
        font-size: 13px;
        font-weight: 700;
        transition: background-color 120ms ease, transform 120ms ease;
      }
      .btn:active { transform: translateY(1px); }
      .btn-danger { background: hsl(348 80% 40%); }
      .btn-danger:hover { background: hsl(348 80% 35%); }
      .hidden { display: none !important; }
    </style>
    <div class="card">
      <div class="header">
        <div class="logo">
          <img class="icon" src="${chrome.runtime.getURL("icons/64x64.png")}" alt="">
          <h1>DAWG Capture</h1>
        </div>
        <div id="statusBadge" class="badge status-idle">Idle</div>
      </div>
      <div class="content">
        <div class="info-box" id="infoBox">
          <p>Start a capture from the DAWG desktop app.</p>
        </div>
        <div class="actions">
          <button id="btnStop" class="btn btn-danger hidden">Stop Recording</button>
        </div>
      </div>
    </div>
  `;

  // (document.body || document.documentElement).appendChild(host);

  const statusBadge = root.getElementById("statusBadge");
  const infoBox = root.getElementById("infoBox");
  const btnStop = root.getElementById("btnStop");

  function updateUI(state) {
    if (state && state.isRecording) {
      statusBadge.textContent = "Recording";
      statusBadge.className = "badge status-recording";
      btnStop.classList.remove("hidden");
      infoBox.innerHTML =
        '<p style="color:#f87171;">Recording the desktop-selected browser tab.</p>';
    } else {
      statusBadge.textContent = "Idle";
      statusBadge.className = "badge status-idle";
      btnStop.classList.add("hidden");
      infoBox.innerHTML = "<p>Start a capture from the DAWG desktop app.</p>";
    }
  }

  function refreshState() {
    chrome.runtime.sendMessage(
      { type: "CHECK_RECORDING_STATE" },
      (response) => {
        if (!chrome.runtime.lastError) updateUI(response);
      },
    );
  }

  const BLUR_TOGGLE_GUARD_MS = 250;
  let lastBlurHideAt = 0;

  function isOpen() {
    return host.style.display !== "none";
  }

  function isInsidePanel(event) {
    const path =
      typeof event.composedPath === "function" ? event.composedPath() : null;
    if (path) return path.includes(host);
    return event.target === host || host.contains(event.target);
  }

  function handleOutsidePointer(event) {
    if (isInsidePanel(event)) return;
    hide();
  }

  function handleKeydown(event) {
    if (event.key === "Escape") hide();
  }

  function handleWindowBlur() {
    if (!isOpen()) return;
    lastBlurHideAt = Date.now();
    hide();
  }

  function attachDismissListeners() {
    document.addEventListener("pointerdown", handleOutsidePointer, true);
    document.addEventListener("keydown", handleKeydown, true);
    window.addEventListener("blur", handleWindowBlur);
  }

  function detachDismissListeners() {
    document.removeEventListener("pointerdown", handleOutsidePointer, true);
    document.removeEventListener("keydown", handleKeydown, true);
    window.removeEventListener("blur", handleWindowBlur);
  }

  function show() {
    if (!host.isConnected) {
      (document.body || document.documentElement).appendChild(host);
    }
    host.style.display = "";
    attachDismissListeners();
    refreshState();
  }

  function hide() {
    detachDismissListeners();
    host.style.display = "none";
    if (host.isConnected) host.remove();
  }

  btnStop.addEventListener("click", () => {
    chrome.runtime.sendMessage({ type: "CMD_STOP_CAPTURE" }, (response) => {
      if (!chrome.runtime.lastError) updateUI(response);
    });
  });

  chrome.runtime.onMessage.addListener((message) => {
    if (!message || message.type !== "DAWG_TOGGLE_PANEL") return;
    if (Date.now() - lastBlurHideAt < BLUR_TOGGLE_GUARD_MS) return;
    if (isOpen()) hide();
    else show();
  });

  void loadOutfitFonts();
})();
