// Persistent state keys written to chrome.storage.session on every change
// so the service worker can recover its recording state after Chrome kills
// and revives the MV3 worker (which can happen after ~30s of inactivity).
const STATE_KEY = "dawg_recording_state";

let isRecording = false;
let daemonUrl = "ws://127.0.0.1:8082/ws";
let httpDaemonUrl = "http://127.0.0.1:8082";
let ws = null;
let connectRetryTimer = null;
const pendingRequests = new Map();

function debugLog(msg) {
  console.log(`[debug] ${msg}`);
}

// persistState writes the current recording state to chrome.storage.session
// so it survives service worker termination and can be read on revival.
function persistState() {
  chrome.storage.session.set({
    [STATE_KEY]: { isRecording, daemonUrl }
  }).catch((e) => debugLog(`persistState: error: ${e.message}`));
}

// restoreState reads persisted state from chrome.storage.session and, if a
// recording was in progress when the worker was killed, schedules a reconnect
// to resume the session transparently.
async function restoreState() {
  try {
    const result = await chrome.storage.session.get(STATE_KEY);
    const saved = result[STATE_KEY];
    if (!saved) return;

    debugLog(`restoreState: found saved state isRecording=${saved.isRecording} daemonUrl=${saved.daemonUrl}`);

    if (saved.isRecording) {
      // Restore in-memory state without re-triggering a full startCapture()
      // (which would notify tabs to start a new rrweb session).
      // We only need to reconnect the WebSocket to resume event streaming.
      isRecording = true;
      if (saved.daemonUrl) {
        daemonUrl = saved.daemonUrl;
        if (daemonUrl.startsWith("ws://")) {
          httpDaemonUrl = daemonUrl.replace("ws://", "http://").replace(/\/ws$/, "");
        } else if (daemonUrl.startsWith("wss://")) {
          httpDaemonUrl = daemonUrl.replace("wss://", "https://").replace(/\/ws$/, "");
        }
      }
      debugLog("restoreState: recording was active — scheduling WebSocket reconnect");
      if (!connectRetryTimer) {
        connectRetryTimer = setTimeout(() => {
          debugLog("restoreState: attempting reconnect to resume session");
          connectWebSocket(daemonUrl, true);
        }, 500);
      }
    }
  } catch (e) {
    debugLog(`restoreState: error: ${e.message}`);
  }
}

// Restore state immediately when the service worker starts (after a kill/revival).
restoreState();

function connectWebSocket(url, isReconnect = false) {
  return new Promise((resolve) => {
    if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
      debugLog(`connectWebSocket: already open/connecting to ${url}, skipping`);
      resolve(true);
      return;
    }

    debugLog(`connectWebSocket: attempting to connect to ${url} (isReconnect=${isReconnect})`);

    try {
      ws = new WebSocket(url);

      ws.onopen = () => {
        debugLog(`connectWebSocket: connected successfully to ${url}`);
        if (connectRetryTimer) {
          clearTimeout(connectRetryTimer);
          connectRetryTimer = null;
        }
        // Only send the SESSION_START handshake if this is a reconnect attempt
        // (i.e., extension started before the desktop daemon was ready, or the
        // worker was revived after being killed by Chrome).
        // For the initial connection, startCapture() sends it after this resolves.
        if (isReconnect && isRecording) {
          debugLog("connectWebSocket: sending late DAWG_SESSION_START (reconnect while recording)");
          sendToDaemon("DAWG_SESSION_START", { startedAt: new Date().toISOString() });
        }
        resolve(true);
      };

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          debugLog(`ws.onmessage: received type=${msg.type}`);
          if (msg.type === "DAWG_COMMAND_STOP") {
            debugLog("ws.onmessage: desktop requested stop, stopping capture");
            stopCapture();
          }
        } catch (e) {
          debugLog(`ws.onmessage: parse error: ${e.message}`);
        }
      };

      ws.onerror = (err) => {
        debugLog(`connectWebSocket: WebSocket error (daemon may not be running yet): ${err.type}`);
        resolve(false);
      };

      ws.onclose = () => {
        debugLog("ws.onclose: WebSocket connection closed");
        ws = null;
        // If still recording when WS closes unexpectedly, schedule a reconnect.
        if (isRecording) {
          debugLog("ws.onclose: still recording — scheduling reconnect in 2s");
          connectRetryTimer = setTimeout(() => {
            debugLog("reconnect: retrying WebSocket connection");
            connectWebSocket(daemonUrl, true); // isReconnect = true
          }, 2000);
        }
      };
    } catch (e) {
      debugLog(`connectWebSocket: exception opening WebSocket: ${e.message}`);
      resolve(false);
    }
  });
}

function sendToDaemon(type, data) {
  const payload = JSON.stringify({ type, data, timestamp: Date.now() });
  debugLog(`sendToDaemon: type=${type} wsState=${ws ? ws.readyState : "null"}`);

  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(payload);
  } else {
    // HTTP Fallback
    const endpoint = type === "DAWG_RRWEB_EVENT" ? "/api/v1/stream/rrweb" :
                     type === "DAWG_ACTION_EVENT" ? "/api/v1/stream/actions" :
                     type === "DAWG_HTTP_EVENT" ? "/api/v1/stream/http" : "/api/v1/stream/event";
    debugLog(`sendToDaemon: WS not ready, falling back to HTTP POST ${endpoint}`);
    fetch(httpDaemonUrl + endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: payload
    }).catch((e) => debugLog(`sendToDaemon: HTTP fallback error: ${e.message}`));
  }
}

async function startCapture(targetDaemonUrl) {
  debugLog(`startCapture: called with daemonUrl=${targetDaemonUrl || daemonUrl}`);

  // Guard against double-start: if state was recovered from storage after a
  // worker kill, isRecording is already true. Return early to avoid sending a
  // duplicate session-start that would confuse the engine's session guard.
  if (isRecording) {
    debugLog("startCapture: already recording (recovered from storage or duplicate call), skipping");
    return { isRecording: true, daemonUrl };
  }

  if (targetDaemonUrl) {
    daemonUrl = targetDaemonUrl;
    if (daemonUrl.startsWith("ws://")) {
      httpDaemonUrl = daemonUrl.replace("ws://", "http://").replace(/\/ws$/, "");
    } else if (daemonUrl.startsWith("wss://")) {
      httpDaemonUrl = daemonUrl.replace("wss://", "https://").replace(/\/ws$/, "");
    }
    debugLog(`startCapture: resolved httpDaemonUrl=${httpDaemonUrl}`);
  }

  // Mark as recording first so reconnect logic and content scripts
  // know to start sending data even if the WS is not yet up.
  isRecording = true;
  persistState();

  const connected = await connectWebSocket(daemonUrl);
  if (connected) {
    // Send START handshake only if connected immediately.
    // If connection fails now, the onopen handler will send it when
    // the daemon eventually becomes available.
    debugLog("startCapture: WS connected, sending DAWG_SESSION_START");
    sendToDaemon("DAWG_SESSION_START", { startedAt: new Date().toISOString() });
  } else {
    debugLog("startCapture: WS not connected yet, scheduling retry — recording is marked active");
    // Schedule a retry so we connect as soon as the desktop daemon is ready.
    if (!connectRetryTimer) {
      connectRetryTimer = setTimeout(() => {
        debugLog("reconnect: retrying after initial connect failure");
        connectWebSocket(daemonUrl, true); // isReconnect=true so onopen sends SESSION_START
      }, 2000);
    }
  }

  // Notify all tabs to start recording rrweb
  const tabs = await chrome.tabs.query({});
  debugLog(`startCapture: notifying ${tabs.length} tabs to start recording`);
  for (const tab of tabs) {
    if (tab.id && tab.url && !tab.url.startsWith("chrome://")) {
      chrome.tabs.sendMessage(tab.id, { type: "START_RECORDING" }).catch(() => {});
    }
  }

  return { isRecording: true, daemonUrl };
}

async function stopCapture() {
  debugLog(`stopCapture: called, isRecording=${isRecording}`);
  if (!isRecording) return { isRecording: false };

  isRecording = false;
  persistState();

  // Cancel any pending reconnect timers — no point reconnecting if we're stopping.
  if (connectRetryTimer) {
    clearTimeout(connectRetryTimer);
    connectRetryTimer = null;
    debugLog("stopCapture: cancelled pending reconnect timer");
  }

  // Send STOP handshake
  debugLog("stopCapture: sending DAWG_SESSION_STOP");
  sendToDaemon("DAWG_SESSION_STOP", { stoppedAt: new Date().toISOString() });

  // Notify all tabs to stop recording rrweb
  const tabs = await chrome.tabs.query({});
  debugLog(`stopCapture: notifying ${tabs.length} tabs to stop recording`);
  for (const tab of tabs) {
    if (tab.id) {
      chrome.tabs.sendMessage(tab.id, { type: "STOP_RECORDING" }).catch(() => {});
    }
  }

  // Give the STOP message a moment to be flushed before closing WS.
  if (ws) {
    setTimeout(() => {
      if (ws) {
        debugLog("stopCapture: closing WebSocket after 500ms flush delay");
        ws.close();
        ws = null;
      }
    }, 500);
  }

  return { isRecording: false };
}

// Intercept Network HTTP traffic via chrome.webRequest
chrome.webRequest.onBeforeRequest.addListener(
  (details) => {
    if (!isRecording) return;
    if (details.url.startsWith("http://127.0.0.1:8082") || details.url.startsWith("ws://127.0.0.1:8082")) return;
    if (details.url.includes("chrome-extension://")) return;

    let body = "";
    if (details.requestBody) {
      if (details.requestBody.raw && details.requestBody.raw[0] && details.requestBody.raw[0].bytes) {
        body = new TextDecoder().decode(details.requestBody.raw[0].bytes);
      } else if (details.requestBody.formData) {
        body = JSON.stringify(details.requestBody.formData);
      }
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
    if (!isRecording) return;
    const req = pendingRequests.get(details.requestId);
    if (req) {
      const headers = {};
      if (details.requestHeaders) {
        for (const h of details.requestHeaders) {
          headers[h.name] = h.value;
        }
      }
      req.requestHeaders = headers;
    }
  },
  { urls: ["<all_urls>"] },
  ["requestHeaders"]
);

chrome.webRequest.onCompleted.addListener(
  (details) => {
    if (!isRecording) return;
    const req = pendingRequests.get(details.requestId);
    if (!req) return;
    pendingRequests.delete(details.requestId);

    const responseHeaders = {};
    if (details.responseHeaders) {
      for (const h of details.responseHeaders) {
        responseHeaders[h.name] = h.value;
      }
    }

    const durationMs = Date.now() - req.startedAt;
    const httpRecord = {
      id: req.id,
      timestamp: new Date().toISOString(),
      request: {
        method: req.method,
        url: req.url,
        headers: req.requestHeaders,
        body: req.requestBody
      },
      response: {
        status: details.statusCode,
        headers: responseHeaders,
        body: ""
      },
      direction: "frontend-to-backend",
      durationMs: durationMs
    };

    sendToDaemon("DAWG_HTTP_EVENT", httpRecord);
  },
  { urls: ["<all_urls>"] },
  ["responseHeaders"]
);

chrome.webRequest.onErrorOccurred.addListener(
  (details) => {
    if (!isRecording) return;
    pendingRequests.delete(details.requestId);
  },
  { urls: ["<all_urls>"] }
);

// Listen to messages from content scripts and popup UI
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === "CHECK_RECORDING_STATE") {
    sendResponse({ isRecording, daemonUrl });
    return true;
  } else if (message.type === "DAWG_RRWEB_EVENT") {
    if (isRecording) {
      sendToDaemon("DAWG_RRWEB_EVENT", message.payload);
    }
  } else if (message.type === "DAWG_ACTION_EVENT") {
    if (isRecording) {
      sendToDaemon("DAWG_ACTION_EVENT", message.payload);
    }
  } else if (message.type === "CMD_START_CAPTURE") {
    startCapture(message.daemonUrl).then((res) => sendResponse(res));
    return true;
  } else if (message.type === "CMD_STOP_CAPTURE") {
    stopCapture().then((res) => sendResponse(res));
    return true;
  }
});
