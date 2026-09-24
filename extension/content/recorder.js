(function () {
  if (window.__dawgRecorderInjected) return;
  window.__dawgRecorderInjected = true;

  let stopRecord = null;
  let isRecording = false;
  const MAX_DIAGNOSTIC_BYTES = 32 * 1024;

  function forwardDiagnostic(event) {
    if (!isRecording || !event.detail || typeof event.detail !== "object") return;
    const { kind, payload } = event.detail;
    if ((kind !== "console" && kind !== "error") || !payload || typeof payload !== "object") return;
    let serialized;
    try {
      serialized = JSON.stringify(payload);
    } catch (_error) {
      return;
    }
    if (serialized.length > MAX_DIAGNOSTIC_BYTES) return;
    chrome.runtime.sendMessage({
      type: kind === "console" ? "DAWG_DIAGNOSTIC_CONSOLE" : "DAWG_DIAGNOSTIC_ERROR",
      payload
    }).catch(() => {});
  }

  window.addEventListener("dawg-diagnostic", forwardDiagnostic);

  function buildSelector(element) {
    if (!element || element === document) return "";
    if (element.id) return `#${element.id}`;
    if (element.className && typeof element.className === "string") {
      const classes = element.className.trim().split(/\s+/).filter(Boolean).join(".");
      if (classes) return `${element.tagName.toLowerCase()}.${classes}`;
    }
    return element.tagName ? element.tagName.toLowerCase() : "";
  }

  function handleInteractionClick(event) {
    if (!isRecording) return;
    const selector = buildSelector(event.target);
    chrome.runtime.sendMessage({
      type: "DAWG_ACTION_EVENT",
      payload: {
        type: "click",
        timestamp: Date.now(),
        selector: selector
      }
    }).catch(() => {});
  }

  function handleInteractionInput(event) {
    if (!isRecording) return;
    const selector = buildSelector(event.target);
    chrome.runtime.sendMessage({
      type: "DAWG_ACTION_EVENT",
      payload: {
        type: "fill",
        timestamp: Date.now(),
        selector: selector,
        fieldName: event.target.name || "",
        inputType: event.target.type || "",
        value: event.target.value || ""
      }
    }).catch(() => {});
  }

  function setDiagnosticsActive(active) {
    window.dispatchEvent(new CustomEvent("dawg-diagnostics-control", { detail: { active } }));
  }

  function browserIdentity(userAgent, brands) {
    const brand = (brands || []).find((item) => /Chrome|Chromium|Edge|Opera/i.test(item.brand));
    if (brand) return { name: brand.brand, version: brand.version };
    const match = userAgent.match(/(Edg|OPR|Chrome|Chromium|Firefox|Version)\/([\d.]+)/);
    const names = { Edg: "Microsoft Edge", OPR: "Opera", Version: /Safari/.test(userAgent) ? "Safari" : "Unknown" };
    return match ? { name: names[match[1]] || match[1], version: match[2] } : { name: "Unknown", version: "" };
  }

  function operatingSystem(userAgent, platform, platformVersion) {
    let name = platform || "Unknown";
    if (/Windows/i.test(userAgent)) name = "Windows";
    else if (/Android/i.test(userAgent)) name = "Android";
    else if (/iPhone|iPad|iPod/i.test(userAgent)) name = "iOS";
    else if (/Mac OS X/i.test(userAgent)) name = "macOS";
    else if (/Linux/i.test(userAgent)) name = "Linux";
    return { name, version: platformVersion || "" };
  }

  async function collectDeviceProfile() {
    const userAgent = navigator.userAgent || "";
    const userAgentData = navigator.userAgentData;
    let highEntropy = {};
    if (userAgentData?.getHighEntropyValues) {
      highEntropy = await userAgentData
        .getHighEntropyValues(["architecture", "bitness", "model", "platformVersion", "uaFullVersion", "fullVersionList"])
        .catch(() => ({}));
    }
    const brands = userAgentData?.brands || [];
    const fullVersionBrand = (highEntropy.fullVersionList || []).find((item) =>
      /Chrome|Chromium|Edge|Opera/i.test(item.brand)
    );
    const connection = navigator.connection || navigator.mozConnection || navigator.webkitConnection;
    return {
      formatVersion: "1",
      capturedAt: new Date().toISOString(),
      browser: {
        ...browserIdentity(userAgent, brands),
        fullVersion: fullVersionBrand?.version || highEntropy.uaFullVersion || "",
        brands
      },
      operatingSystem: {
        ...operatingSystem(userAgent, userAgentData?.platform || navigator.platform, highEntropy.platformVersion),
        architecture: highEntropy.architecture || "",
        bitness: highEntropy.bitness || ""
      },
      device: {
        model: highEntropy.model || "",
        type: userAgentData?.mobile ? "mobile" : navigator.maxTouchPoints > 0 ? "touch-capable" : "desktop",
        mobile: Boolean(userAgentData?.mobile),
        hostname: { state: "unavailable", reason: "Browser security prevents access to the device hostname." },
        manufacturer: { state: "unavailable", reason: "Browser APIs do not expose a reliable device manufacturer." }
      },
      viewport: { width: window.innerWidth, height: window.innerHeight },
      screen: {
        width: window.screen.width,
        height: window.screen.height,
        availableWidth: window.screen.availWidth,
        availableHeight: window.screen.availHeight,
        colorDepth: window.screen.colorDepth,
        pixelRatio: window.devicePixelRatio
      },
      locale: {
        language: navigator.language || "",
        languages: Array.from(navigator.languages || []),
        timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || ""
      },
      hardware: {
        logicalProcessors: navigator.hardwareConcurrency || null,
        deviceMemoryGiB: navigator.deviceMemory || null,
        processorModel: { state: "unavailable", reason: "Browser APIs expose logical capacity but not the CPU model." },
        storage: { state: "unavailable", reason: "Browser APIs do not expose physical storage devices." }
      },
      network: {
        connectionType: connection?.type || "",
        effectiveType: connection?.effectiveType || "",
        downlinkMbps: connection?.downlink ?? null,
        roundTripTimeMs: connection?.rtt ?? null,
        saveData: connection?.saveData ?? null,
        ipAddresses: { state: "unavailable", reason: "Browser APIs do not expose reliable IP addresses without an external service." },
        macAddress: { state: "unavailable", reason: "Browser security prevents access to MAC addresses." }
      },
      page: { url: location.href },
      userAgent
    };
  }

  function sendDeviceProfile() {
    void collectDeviceProfile()
      .then((payload) => chrome.runtime.sendMessage({ type: "DAWG_DEVICE_INFO", payload }))
      .catch(() => {});
  }

  function startRecording() {
    if (isRecording) return;
    if (typeof rrweb === "undefined" || typeof rrweb.record !== "function") {
      throw new Error("rrweb recorder is unavailable");
    }

    stopRecord = rrweb.record({
        // Mask values in rrweb snapshots. The separate action stream is
        // sanitized by the engine before packaging.
        maskAllInputs: true,
        emit(event) {
          if (!isRecording) return;
          chrome.runtime.sendMessage({
            type: "DAWG_RRWEB_EVENT",
            payload: event
          }).catch(() => {});
        }
      });
    if (typeof stopRecord !== "function") {
      stopRecord = null;
      throw new Error("rrweb recorder did not start");
    }

    isRecording = true;
    setDiagnosticsActive(true);
    document.addEventListener("click", handleInteractionClick, true);
    document.addEventListener("input", handleInteractionInput, true);
    sendDeviceProfile();
  }

  function stopRecording() {
    if (!isRecording) return;
    isRecording = false;
    setDiagnosticsActive(false);

    document.removeEventListener("click", handleInteractionClick, true);
    document.removeEventListener("input", handleInteractionInput, true);

    if (typeof stopRecord === "function") {
      stopRecord();
      stopRecord = null;
    }
  }

  chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type === "START_RECORDING") {
      try {
        startRecording();
        sendResponse({ status: "started" });
      } catch (error) {
        sendResponse({ status: "error", error: error.message });
      }
    } else if (message.type === "STOP_RECORDING") {
      stopRecording();
      sendResponse({ status: "stopped" });
    } else if (message.type === "GET_RECORDING_STATUS") {
      sendResponse({ isRecording });
    }
  });


  chrome.runtime.sendMessage({ type: "CHECK_RECORDING_STATE" }, (response) => {
    if (chrome.runtime.lastError) return;
    if (response && response.isRecording) {
      startRecording();
    }
  });

  // Wake the MV3 worker while a normal page is open so it can discover a
  // newly-started desktop daemon. Once connected, its WebSocket keeps it alive.
  setInterval(() => {
    chrome.runtime.sendMessage({ type: "DAWG_EXTENSION_WAKE" }).catch(() => {});
  }, 2000);
})();
