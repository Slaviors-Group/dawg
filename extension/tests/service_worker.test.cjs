const assert = require("node:assert/strict");
const { after, test } = require("node:test");

class ChromeEvent {
  constructor() {
    this.listeners = [];
  }

  addListener(listener) {
    this.listeners.push(listener);
  }
}

class FakeWebSocket {
  static CONNECTING = 0;
  static OPEN = 1;
  static CLOSED = 3;
  static instances = [];

  constructor(url) {
    this.url = url;
    this.readyState = FakeWebSocket.CONNECTING;
    this.sent = [];
    FakeWebSocket.instances.push(this);
    queueMicrotask(() => {
      this.readyState = FakeWebSocket.OPEN;
      this.onopen?.();
    });
  }

  send(payload) {
    const message = JSON.parse(payload);
    this.sent.push(message);
    if (message.type === "DAWG_SESSION_STOP") {
      queueMicrotask(() => {
        this.onmessage?.({ data: JSON.stringify({ type: "DAWG_SESSION_STOP_ACK" }) });
      });
    }
  }

  receive(message) {
    this.onmessage?.({ data: JSON.stringify(message) });
  }

  close() {
    this.readyState = FakeWebSocket.CLOSED;
    this.onclose?.();
  }
}

const runtimeMessages = new ChromeEvent();
const actionClicked = new ChromeEvent();
const tabMessages = [];
const scriptExecutions = [];
const operationLog = [];
const stored = {};

globalThis.__DAWG_TEST_MODE__ = true;
globalThis.WebSocket = FakeWebSocket;
globalThis.chrome = {
  storage: {
    session: {
      async get(key) {
        return key in stored ? { [key]: stored[key] } : {};
      },
      async set(values) {
        Object.assign(stored, values);
      }
    }
  },
  scripting: {
    async executeScript({ target, files }) {
      operationLog.push(`script:${files.join(",")}`);
      scriptExecutions.push({ target, files });
    }
  },
  tabs: {
    async query() {
      return [{ id: 7, windowId: 2, url: "http://localhost:3000" }];
    },
    async update() {},
    async create() {
      throw new Error("the existing matching tab should be reused");
    },
    async sendMessage(tabId, message) {
      tabMessages.push({ tabId, message });
      if (message.type === "START_RECORDING") return { status: "started" };
      if (message.type === "STOP_RECORDING") return { status: "stopped" };
      return {};
    },
    onRemoved: new ChromeEvent()
  },
  action: { onClicked: actionClicked },
  windows: { async update() {} },
  debugger: {
    async attach() { operationLog.push("debugger:attach"); },
    async detach() { operationLog.push("debugger:detach"); },
    async sendCommand(_target, command) { operationLog.push(`debugger:${command}`); return {}; },
    onEvent: new ChromeEvent(),
    onDetach: new ChromeEvent()
  },
  webRequest: {
    onBeforeRequest: new ChromeEvent(),
    onBeforeSendHeaders: new ChromeEvent(),
    onCompleted: new ChromeEvent(),
    onErrorOccurred: new ChromeEvent()
  },
  runtime: {
    onMessage: runtimeMessages,
    onStartup: new ChromeEvent(),
    onInstalled: new ChromeEvent()
  }
};

require("../background/service_worker.js");

after(() => globalThis.__DAWG_EXTENSION_TEST__.shutdown());

function waitFor(predicate, message) {
  return new Promise((resolve, reject) => {
    const deadline = Date.now() + 1000;
    const poll = () => {
      if (predicate()) {
        resolve();
      } else if (Date.now() >= deadline) {
        reject(new Error(message));
      } else {
        setTimeout(poll, 5);
      }
    };
    poll();
  });
}

test("daemon starts and stops one target tab with an acknowledged drain", async () => {
  await waitFor(() => FakeWebSocket.instances[0]?.sent.some((item) => item.type === "DAWG_EXTENSION_READY"), "extension did not announce readiness");
  const socket = FakeWebSocket.instances[0];

  const startAt = operationLog.length;
  socket.receive({
    type: "DAWG_COMMAND_START",
    data: { targetUrl: "http://localhost:3000", sessionToken: "session-test", diagnosticsProfile: "enhanced" }
  });
  await waitFor(() => socket.sent.some((item) => item.type === "DAWG_SESSION_START"), "extension did not acknowledge capture start");

  assert.equal(globalThis.__DAWG_EXTENSION_TEST__.publicState().isRecording, true);
  assert.equal(globalThis.__DAWG_EXTENSION_TEST__.publicState().recordingTabId, 7);
  assert.equal(socket.sent.find((item) => item.type === "DAWG_SESSION_START").sessionToken, "session-test");
  assert.ok(tabMessages.some(({ tabId, message }) => tabId === 7 && message.type === "START_RECORDING"));
  assert.ok(
    scriptExecutions.some(
      ({ target, files }) =>
        target.tabId === 7 &&
        files.join(",") === "lib/rrweb.js,content/recorder.js"
    )
  );
  const operations = operationLog.slice(startAt);
  const attachAt = operations.indexOf("debugger:attach");
  const recorderAt = operations.indexOf("script:lib/rrweb.js,content/recorder.js");
  assert.ok(attachAt >= 0, "enhanced capture did not attach the debugger");
  assert.ok(recorderAt > attachAt, `rrweb started before debugger attachment: ${operations.join(", ")}`);

  const runtimeListener = runtimeMessages.listeners[0];
  runtimeListener(
    { type: "DAWG_RRWEB_EVENT", payload: { type: 2, data: { node: "snapshot" } } },
    { tab: { id: 99 } },
    () => {}
  );
  runtimeListener(
    { type: "DAWG_RRWEB_EVENT", payload: { type: 3, data: { node: "mutation" } } },
    { tab: { id: 7 } },
    () => {}
  );
  assert.equal(socket.sent.filter((item) => item.type === "DAWG_RRWEB_EVENT").length, 1);

  socket.receive({ type: "DAWG_COMMAND_STOP", data: null });
  await waitFor(() => socket.sent.some((item) => item.type === "DAWG_SESSION_STOP"), "extension did not send the ordered stop marker");
  await waitFor(() => !globalThis.__DAWG_EXTENSION_TEST__.publicState().isRecording, "extension did not leave recording state");
  assert.ok(tabMessages.some(({ tabId, message }) => tabId === 7 && message.type === "STOP_RECORDING"));
});


test("toolbar action injects and toggles the in-page capture panel", async () => {
  await actionClicked.listeners[0]({ id: 7 });

  assert.ok(
    scriptExecutions.some(
      ({ target, files }) => target.tabId === 7 && files.join(",") === "content/popup-panel.js"
    )
  );
  assert.ok(
    tabMessages.some(({ tabId, message }) => tabId === 7 && message.type === "DAWG_TOGGLE_PANEL")
  );
});
