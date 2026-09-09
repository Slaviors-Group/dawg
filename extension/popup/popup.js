document.addEventListener("DOMContentLoaded", () => {
  const daemonInput = document.getElementById("daemonAddress");
  const btnStart = document.getElementById("btnStart");
  const btnStop = document.getElementById("btnStop");
  const statusBadge = document.getElementById("statusBadge");
  const infoBox = document.getElementById("infoBox");

  function updateUI(isRecording, daemonUrl) {
    if (daemonUrl) daemonInput.value = daemonUrl;

    if (isRecording) {
      statusBadge.textContent = "Recording";
      statusBadge.className = "badge status-recording";
      btnStart.classList.add("hidden");
      btnStop.classList.remove("hidden");
      infoBox.innerHTML = "<p style='color:#f87171;'>🔴 Capture session active. Intercepting DOM events & HTTP network calls.</p>";
      daemonInput.disabled = true;
    } else {
      statusBadge.textContent = "Idle";
      statusBadge.className = "badge status-idle";
      btnStart.classList.remove("hidden");
      btnStop.classList.add("hidden");
      infoBox.innerHTML = "<p>Ready to record DOM events & network traffic directly from your active browser session.</p>";
      daemonInput.disabled = false;
    }
  }

  chrome.runtime.sendMessage({ type: "CHECK_RECORDING_STATE" }, (response) => {
    if (chrome.runtime.lastError) return;
    if (response) {
      updateUI(response.isRecording, response.daemonUrl);
    }
  });

  btnStart.addEventListener("click", () => {
    const url = daemonInput.value.trim();
    chrome.runtime.sendMessage({ type: "CMD_START_CAPTURE", daemonUrl: url }, (response) => {
      if (response && response.isRecording) {
        updateUI(true, response.daemonUrl);
      }
    });
  });

  btnStop.addEventListener("click", () => {
    chrome.runtime.sendMessage({ type: "CMD_STOP_CAPTURE" }, (response) => {
      if (response) {
        updateUI(false);
      }
    });
  });
});

