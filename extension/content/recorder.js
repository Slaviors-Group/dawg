(function () {
  if (window.__dawgRecorderInjected) return;
  window.__dawgRecorderInjected = true;

  let stopRecord = null;
  let isRecording = false;

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
        value: event.target.value || ""
      }
    }).catch(() => {});
  }

  function startRecording() {
    if (isRecording) return;
    isRecording = true;

    document.addEventListener("click", handleInteractionClick, true);
    document.addEventListener("input", handleInteractionInput, true);

    if (typeof rrweb !== "undefined" && typeof rrweb.record === "function") {
      stopRecord = rrweb.record({
        emit(event) {
          if (!isRecording) return;
          chrome.runtime.sendMessage({
            type: "DAWG_RRWEB_EVENT",
            payload: event
          }).catch(() => {});
        }
      });
    }
  }

  function stopRecording() {
    if (!isRecording) return;
    isRecording = false;

    document.removeEventListener("click", handleInteractionClick, true);
    document.removeEventListener("input", handleInteractionInput, true);

    if (typeof stopRecord === "function") {
      stopRecord();
      stopRecord = null;
    }
  }

  chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type === "START_RECORDING") {
      startRecording();
      sendResponse({ status: "started" });
    } else if (message.type === "STOP_RECORDING") {
      stopRecording();
      sendResponse({ status: "stopped" });
    } else if (message.type === "GET_RECORDING_STATUS") {
      sendResponse({ isRecording });
    }
  });

  // Query status on initialization
  chrome.runtime.sendMessage({ type: "CHECK_RECORDING_STATE" }, (response) => {
    if (chrome.runtime.lastError) return;
    if (response && response.isRecording) {
      startRecording();
    }
  });
})();

