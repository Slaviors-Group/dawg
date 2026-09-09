let isRecording = false;
let daemonUrl = "ws://127.0.0.1:8082/ws";
let httpDaemonUrl = "http://127.0.0.1:8082";
let ws = null;
const pendingRequests = new Map();

function connectWebSocket(url) {
  return new Promise((resolve) => {
    if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
      resolve(true);
      return;
    }

    try {
      ws = new WebSocket(url);

      ws.onopen = () => {
        console.log("[DAWG ServiceWorker] Connected to DAWG Daemon at", url);
        resolve(true);
      };

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          if (msg.type === "DAWG_COMMAND_STOP") {
            stopCapture();
          }
        } catch (e) {}
      };

      ws.onerror = (err) => {
        console.warn("[DAWG ServiceWorker] WebSocket error:", err);
        resolve(false);
      };

      ws.onclose = () => {
        console.log("[DAWG ServiceWorker] WebSocket connection closed.");
        ws = null;
      };
    } catch (e) {
      console.warn("[DAWG ServiceWorker] Exception opening WebSocket:", e);
      resolve(false);
    }
  });
}

function sendToDaemon(type, data) {
  const payload = JSON.stringify({ type, data, timestamp: Date.now() });

  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(payload);
  } else {
    // HTTP Fallback
    const endpoint = type === "DAWG_RRWEB_EVENT" ? "/api/v1/stream/rrweb" :
                     type === "DAWG_ACTION_EVENT" ? "/api/v1/stream/actions" :
                     type === "DAWG_HTTP_EVENT" ? "/api/v1/stream/http" : "/api/v1/stream/event";
    fetch(httpDaemonUrl + endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: payload
    }).catch(() => {});
  }
}

async function startCapture(targetDaemonUrl) {
  if (targetDaemonUrl) {
    daemonUrl = targetDaemonUrl;
    if (daemonUrl.startsWith("ws://")) {
      httpDaemonUrl = daemonUrl.replace("ws://", "http://").replace(/\/ws$/, "");
    } else if (daemonUrl.startsWith("wss://")) {
      httpDaemonUrl = daemonUrl.replace("wss://", "https://").replace(/\/ws$/, "");
    }
  }

  await connectWebSocket(daemonUrl);
  isRecording = true;

  // Send START handshake
  sendToDaemon("DAWG_SESSION_START", { startedAt: new Date().toISOString() });

  // Notify all tabs to start recording rrweb
  const tabs = await chrome.tabs.query({});
  for (const tab of tabs) {
    if (tab.id && tab.url && !tab.url.startsWith("chrome://")) {
      chrome.tabs.sendMessage(tab.id, { type: "START_RECORDING" }).catch(() => {});
    }
  }

  return { isRecording: true, daemonUrl };
}

async function stopCapture() {
  if (!isRecording) return { isRecording: false };

  isRecording = false;

  // Send STOP handshake
  sendToDaemon("DAWG_SESSION_STOP", { stoppedAt: new Date().toISOString() });

  // Notify all tabs to stop recording rrweb
  const tabs = await chrome.tabs.query({});
  for (const tab of tabs) {
    if (tab.id) {
      chrome.tabs.sendMessage(tab.id, { type: "STOP_RECORDING" }).catch(() => {});
    }
  }

  if (ws) {
    setTimeout(() => {
      if (ws) {
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

