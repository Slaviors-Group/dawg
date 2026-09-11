const STATE_KEY = "dawg_recording_state";
const DEFAULT_DAEMON_URL = "ws://127.0.0.1:8082/ws";
const HTTP_DELIVERY_TIMEOUT_MS = 2000;
const CONTENT_READY_TIMEOUT_MS = 8000;
const RECONNECT_DELAY_MS = 2000;
const KEEPALIVE_INTERVAL_MS = 20000;

let isRecording = false;
let recordingTabId = null;
let targetUrl = null;
let sessionToken = null;
let daemonUrl = DEFAULT_DAEMON_URL;
let httpDaemonUrl = "http://127.0.0.1:8082";
let ws = null;
let connectPromise = null;
let connectRetryTimer = null;
let keepaliveTimer = null;
let startCapturePromise = null;
let stopCapturePromise = null;
let stopAckWaiter = null;
let stateReadyPromise = null;

const pendingRequests = new Map();
const pendingDeliveries = new Set();

function debugLog(message) {
  console.log(`[DAWG] ${message}`);
}

function setDaemonUrl(value) {
  if (typeof value !== "string" || !value.trim()) return;
  daemonUrl = value.trim();
  if (daemonUrl.startsWith("ws://")) {
    httpDaemonUrl = daemonUrl.replace("ws://", "http://").replace(/\/ws$/, "");
  } else if (daemonUrl.startsWith("wss://")) {
    httpDaemonUrl = daemonUrl.replace("wss://", "https://").replace(/\/ws$/, "");
  }
}

function publicState() {
  return { isRecording, recordingTabId, targetUrl, sessionToken, daemonUrl };
}

async function persistState() {
  try {
    await chrome.storage.session.set({ [STATE_KEY]: publicState() });
  } catch (error) {
    debugLog(`Could not persist state: ${error.message}`);
  }
}

async function restoreState() {
  try {
    const result = await chrome.storage.session.get(STATE_KEY);
    const saved = result[STATE_KEY];
    if (saved) {
      isRecording = saved.isRecording === true;
      recordingTabId = Number.isInteger(saved.recordingTabId) ? saved.recordingTabId : null;
      targetUrl = typeof saved.targetUrl === "string" ? saved.targetUrl : null;
      sessionToken = typeof saved.sessionToken === "string" ? saved.sessionToken : null;
      setDaemonUrl(saved.daemonUrl || DEFAULT_DAEMON_URL);
    }
  } catch (error) {
    debugLog(`Could not restore state: ${error.message}`);
  }
  void ensureDaemonConnection();
}

function initializeState() {
  if (!stateReadyPromise) stateReadyPromise = restoreState();
  return stateReadyPromise;
}

function clearReconnectTimer() {
  if (connectRetryTimer !== null) {
    clearTimeout(connectRetryTimer);
    connectRetryTimer = null;
  }
}

function scheduleReconnect() {
  if (connectRetryTimer !== null) return;
  connectRetryTimer = setTimeout(() => {
    connectRetryTimer = null;
    void ensureDaemonConnection();
  }, RECONNECT_DELAY_MS);
}

function clearKeepalive() {
  if (keepaliveTimer !== null) {
    clearInterval(keepaliveTimer);
    keepaliveTimer = null;
  }
}

function sendSocketEnvelope(type, data = null, token = sessionToken) {
  if (!ws || ws.readyState !== WebSocket.OPEN) return false;
  try {
    ws.send(JSON.stringify({ type, data, sessionToken: token, timestamp: Date.now() }));
    return true;
  } catch (error) {
    debugLog(`WebSocket send failed for ${type}: ${error.message}`);
    return false;
  }
}

function startKeepalive() {
  clearKeepalive();
  keepaliveTimer = setInterval(() => {
    sendSocketEnvelope("DAWG_KEEPALIVE", { recording: isRecording });
  }, KEEPALIVE_INTERVAL_MS);
}

function settleStopAck(acknowledged) {
  if (!stopAckWaiter) return;
  const waiter = stopAckWaiter;
  stopAckWaiter = null;
  clearTimeout(waiter.timer);
  waiter.resolve(acknowledged);
}

function waitForStopAck(timeoutMs) {
  settleStopAck(false);
  return new Promise((resolve) => {
    const timer = setTimeout(() => {
      if (stopAckWaiter && stopAckWaiter.timer === timer) {
        stopAckWaiter = null;
        resolve(false);
      }
    }, timeoutMs);
    stopAckWaiter = { resolve, timer };
  });
}

function handleDaemonMessage(event) {
  let message;
  try {
    message = JSON.parse(event.data);
  } catch (error) {
    debugLog(`Ignored invalid daemon message: ${error.message}`);
    return;
  }

  if (message.type === "DAWG_COMMAND_START") {
    const requestedUrl = message.data && message.data.targetUrl;
    const requestedToken = message.data && message.data.sessionToken;
    void startCaptureForTarget(requestedUrl, requestedToken).catch((error) => {
      debugLog(`Daemon start command failed: ${error.message}`);
      void sendToDaemon("DAWG_SESSION_ERROR", { message: error.message }, requestedToken);
    });
  } else if (message.type === "DAWG_COMMAND_STOP") {
    void stopCapture();
  } else if (message.type === "DAWG_SESSION_STOP_ACK") {
    settleStopAck(true);
  }
}

function connectWebSocket(url) {
  if (ws && ws.readyState === WebSocket.OPEN) return Promise.resolve(true);
  if (connectPromise) return connectPromise;

  connectPromise = new Promise((resolve) => {
    let settled = false;
    const finish = (connected) => {
      if (settled) return;
      settled = true;
      resolve(connected);
    };

    try {
      const socket = new WebSocket(url);
      ws = socket;

      socket.onopen = () => {
        if (ws !== socket) {
          socket.close();
          finish(false);
          return;
        }
        clearReconnectTimer();
        startKeepalive();
        sendSocketEnvelope("DAWG_EXTENSION_READY", publicState());
        if (isRecording) {
          sendSocketEnvelope("DAWG_SESSION_START", {
            targetUrl,
            tabId: recordingTabId,
            resumed: true
          });
        }
        finish(true);
      };

      socket.onmessage = handleDaemonMessage;
      socket.onerror = () => finish(false);
      socket.onclose = () => {
        if (ws === socket) ws = null;
        clearKeepalive();
        settleStopAck(false);
        finish(false);
        scheduleReconnect();
      };
    } catch (error) {
      debugLog(`Could not open daemon WebSocket: ${error.message}`);
      ws = null;
      finish(false);
      scheduleReconnect();
    }
  }).finally(() => {
    connectPromise = null;
  });

  return connectPromise;
}

async function ensureDaemonConnection() {
  const connected = await connectWebSocket(daemonUrl);
  if (!connected) scheduleReconnect();
  return connected;
}

function sendToDaemon(type, data, token = sessionToken) {
  const delivery = deliverToDaemon(type, data, token);
  pendingDeliveries.add(delivery);
  void delivery.finally(() => pendingDeliveries.delete(delivery));
  return delivery;
}

async function deliverToDaemon(type, data, token) {
  const payload = JSON.stringify({ type, data, sessionToken: token, timestamp: Date.now() });
  if (sendSocketEnvelope(type, data, token)) return true;

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), HTTP_DELIVERY_TIMEOUT_MS);
  try {
    const response = await fetch(`${httpDaemonUrl}/api/v1/stream/event`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: payload,
      signal: controller.signal
    });
    return response.ok;
  } catch (_error) {
    return false;
  } finally {
    clearTimeout(timeout);
  }
}

function normalizedUrl(value) {
  const parsed = new URL(value);
  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    throw new Error("DAWG can only record http:// or https:// pages");
  }
  parsed.hash = "";
  if (parsed.pathname.length > 1) parsed.pathname = parsed.pathname.replace(/\/+$/, "");
  return parsed.toString();
}

async function findOrCreateTargetTab(requestedUrl) {
  const wanted = normalizedUrl(requestedUrl);
  const tabs = await chrome.tabs.query({});
  const exact = tabs.find((tab) => {
    if (!tab.id || !tab.url) return false;
    try {
      return normalizedUrl(tab.url) === wanted;
    } catch (_error) {
      return false;
    }
  });
  if (exact) {
    await chrome.tabs.update(exact.id, { active: true }).catch(() => {});
    if (exact.windowId !== undefined) {
      await chrome.windows.update(exact.windowId, { focused: true }).catch(() => {});
    }
    return exact;
  }
  return chrome.tabs.create({ url: requestedUrl, active: true });
}

async function startRecorderInTab(tabId) {
  const deadline = Date.now() + CONTENT_READY_TIMEOUT_MS;
  let lastError = null;
  while (Date.now() < deadline) {
    try {
      await chrome.scripting.executeScript({
        target: { tabId },
        files: ["lib/rrweb.min.js", "content/recorder.js"]
      });
      const response = await chrome.tabs.sendMessage(tabId, { type: "START_RECORDING" });
      if (response && response.status === "started") return;
      lastError = new Error("the page recorder did not acknowledge start");
    } catch (error) {
      lastError = error;
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`could not inject the recorder into the target tab: ${lastError ? lastError.message : "page did not become ready"}`);
}

async function performStartCapture(requestedUrl, requestedToken) {
  if (!requestedUrl) throw new Error("the desktop did not provide a target URL");
  if (!requestedToken) throw new Error("the desktop did not provide a capture session token");
  const wanted = normalizedUrl(requestedUrl);
  if (isRecording) {
    if (sessionToken === requestedToken && targetUrl && normalizedUrl(targetUrl) === wanted) {
      await sendToDaemon("DAWG_SESSION_START", {
        startedAt: new Date().toISOString(), targetUrl, tabId: recordingTabId, resumed: true
      });
      return publicState();
    }
    // A daemon can disappear without completing its stop handshake (host
    // shutdown, crash, or update). A new authenticated command supersedes that
    // stale local state and starts a fresh rrweb snapshot for the new session.
    if (recordingTabId !== null) {
      await chrome.tabs.sendMessage(recordingTabId, { type: "STOP_RECORDING" }).catch(() => {});
    }
    isRecording = false;
    recordingTabId = null;
    targetUrl = null;
    sessionToken = null;
    pendingRequests.clear();
  }

  const tab = await findOrCreateTargetTab(requestedUrl);
  if (!tab.id) throw new Error("Chrome did not provide an ID for the target tab");
  recordingTabId = tab.id;
  targetUrl = requestedUrl;
  sessionToken = requestedToken;
  isRecording = true;
  await persistState();

  try {
    // State is active before the content recorder starts so its initial rrweb
    // full-snapshot event is accepted instead of being dropped as "too early".
    await startRecorderInTab(tab.id);
    if (!isRecording || recordingTabId !== tab.id) {
      throw new Error("capture start was cancelled before the target tab became ready");
    }
    const delivered = await sendToDaemon("DAWG_SESSION_START", {
      startedAt: new Date().toISOString(),
      targetUrl,
      tabId: recordingTabId
    });
    if (!delivered) {
      throw new Error("the browser extension lost its connection to the DAWG desktop engine");
    }
  } catch (error) {
    await chrome.tabs.sendMessage(tab.id, { type: "STOP_RECORDING" }).catch(() => {});
    if (recordingTabId === tab.id) {
      isRecording = false;
      recordingTabId = null;
      targetUrl = null;
      sessionToken = null;
      await persistState();
    }
    throw error;
  }
  return publicState();
}

function startCaptureForTarget(requestedUrl, requestedToken) {
  if (startCapturePromise) return startCapturePromise;
  startCapturePromise = performStartCapture(requestedUrl, requestedToken).finally(() => {
    startCapturePromise = null;
  });
  return startCapturePromise;
}

async function performStopCapture() {
  if (!isRecording) return { ...publicState(), acknowledged: true };

  const tabId = recordingTabId;
  const stoppingToken = sessionToken;
  if (tabId !== null) {
    await chrome.tabs.sendMessage(tabId, { type: "STOP_RECORDING" }).catch(() => {});
  }
  await new Promise((resolve) => setTimeout(resolve, 50));

  isRecording = false;
  recordingTabId = null;
  targetUrl = null;
  await persistState();
  pendingRequests.clear();
  await Promise.allSettled([...pendingDeliveries]);

  let acknowledged = false;
  if (ws && ws.readyState === WebSocket.OPEN) {
    const ack = waitForStopAck(3500);
    const sent = await sendToDaemon("DAWG_SESSION_STOP", {
      stoppedAt: new Date().toISOString(),
      tabId
    }, stoppingToken);
    acknowledged = sent ? await ack : false;
  } else {
    acknowledged = await sendToDaemon("DAWG_SESSION_STOP", {
      stoppedAt: new Date().toISOString(),
      tabId
    }, stoppingToken);
  }

  sessionToken = null;
  await persistState();

  if (ws) {
    const socket = ws;
    ws = null;
    socket.close(1000, "capture stopped");
  }
  return { ...publicState(), acknowledged };
}

function stopCapture() {
  if (stopCapturePromise) return stopCapturePromise;
  stopCapturePromise = performStopCapture().finally(() => {
    stopCapturePromise = null;
  });
  return stopCapturePromise;
}

function belongsToRecordingTab(details) {
  return isRecording && recordingTabId !== null && details.tabId === recordingTabId;
}

chrome.webRequest.onBeforeRequest.addListener(
  (details) => {
    if (!belongsToRecordingTab(details)) return;
    if (details.url.startsWith(httpDaemonUrl) || details.url.startsWith(daemonUrl)) return;

    let body = "";
    if (details.requestBody?.raw?.[0]?.bytes) {
      body = new TextDecoder().decode(details.requestBody.raw[0].bytes);
    } else if (details.requestBody?.formData) {
      body = JSON.stringify(details.requestBody.formData);
    }
    pendingRequests.set(details.requestId, {
      id: crypto.randomUUID(),
      url: details.url,
      method: details.method,
      startedAt: Date.now(),
      requestHeaders: {},
      requestBody: body
    });
  },
  { urls: ["<all_urls>"] },
  ["requestBody"]
);

chrome.webRequest.onBeforeSendHeaders.addListener(
  (details) => {
    if (!belongsToRecordingTab(details)) return;
    const request = pendingRequests.get(details.requestId);
    if (!request) return;
    request.requestHeaders = Object.fromEntries(
      (details.requestHeaders || []).map((header) => [header.name, header.value || ""])
    );
  },
  { urls: ["<all_urls>"] },
  ["requestHeaders"]
);

chrome.webRequest.onCompleted.addListener(
  (details) => {
    if (!belongsToRecordingTab(details)) return;
    const request = pendingRequests.get(details.requestId);
    if (!request) return;
    pendingRequests.delete(details.requestId);
    const responseHeaders = Object.fromEntries(
      (details.responseHeaders || []).map((header) => [header.name, header.value || ""])
    );
    void sendToDaemon("DAWG_HTTP_EVENT", {
      id: request.id,
      timestamp: new Date().toISOString(),
      request: {
        method: request.method,
        url: request.url,
        headers: request.requestHeaders,
        body: request.requestBody
      },
      response: { status: details.statusCode, headers: responseHeaders, body: "" },
      direction: "frontend-to-backend",
      durationMs: Date.now() - request.startedAt
    });
  },
  { urls: ["<all_urls>"] },
  ["responseHeaders"]
);

chrome.webRequest.onErrorOccurred.addListener(
  (details) => {
    if (belongsToRecordingTab(details)) pendingRequests.delete(details.requestId);
  },
  { urls: ["<all_urls>"] }
);

chrome.tabs.onRemoved.addListener((tabId) => {
  if (isRecording && tabId === recordingTabId) {
    void sendToDaemon("DAWG_SESSION_ERROR", { message: "the recorded browser tab was closed" });
    void stopCapture();
  }
});

chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === "CHECK_RECORDING_STATE" || message.type === "DAWG_EXTENSION_WAKE") {
    void initializeState().then(() => {
      void ensureDaemonConnection();
      sendResponse(publicState());
    });
    return true;
  }
  if (message.type === "DAWG_RRWEB_EVENT") {
    if (isRecording && sender.tab?.id === recordingTabId) {
      void sendToDaemon("DAWG_RRWEB_EVENT", message.payload);
    }
    return false;
  }
  if (message.type === "DAWG_ACTION_EVENT") {
    if (isRecording && sender.tab?.id === recordingTabId) {
      void sendToDaemon("DAWG_ACTION_EVENT", message.payload);
    }
    return false;
  }
  if (message.type === "CMD_START_CAPTURE") {
    void (async () => {
      setDaemonUrl(message.daemonUrl || DEFAULT_DAEMON_URL);
      await persistState();
      await ensureDaemonConnection();
      const tabs = await chrome.tabs.query({ active: true, currentWindow: true });
      const activeTab = tabs[0];
      if (!activeTab?.url) throw new Error("no recordable active tab was found");
      // Popup starts are retained for diagnostics, but a recorder session still
      // needs a daemon-issued token. The normal product path starts in desktop.
      if (!message.sessionToken) throw new Error("start capture from the DAWG desktop app");
      return startCaptureForTarget(message.targetUrl || activeTab.url, message.sessionToken);
    })().then(sendResponse).catch((error) => sendResponse({ ...publicState(), error: error.message }));
    return true;
  }
  if (message.type === "CMD_STOP_CAPTURE") {
    void stopCapture().then(sendResponse).catch((error) => sendResponse({ ...publicState(), error: error.message }));
    return true;
  }
  return false;
});

chrome.runtime.onStartup.addListener(() => void initializeState());
chrome.runtime.onInstalled.addListener(() => void initializeState());
void initializeState();

if (globalThis.__DAWG_TEST_MODE__) {
  globalThis.__DAWG_EXTENSION_TEST__ = {
    publicState,
    shutdown() {
      clearReconnectTimer();
      clearKeepalive();
      if (ws) {
        ws.onclose = null;
        ws.close();
        ws = null;
      }
    }
  };
}
