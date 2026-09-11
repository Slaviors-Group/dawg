document.addEventListener("DOMContentLoaded", () => {
  const btnStop = document.getElementById("btnStop");
  const statusBadge = document.getElementById("statusBadge");
  const infoBox = document.getElementById("infoBox");

  function updateUI(state) {
    if (state?.isRecording) {
      statusBadge.textContent = "Recording";
      statusBadge.className = "badge status-recording";
      btnStop.classList.remove("hidden");
      infoBox.innerHTML = "<p style=\"color:#f87171;\">Recording the desktop-selected browser tab.</p>";
    } else {
      statusBadge.textContent = "Idle";
      statusBadge.className = "badge status-idle";
      btnStop.classList.add("hidden");
      infoBox.innerHTML = "<p>Start a capture from the DAWG desktop app.</p>";
    }
  }

  chrome.runtime.sendMessage({ type: "CHECK_RECORDING_STATE" }, (response) => {
    if (!chrome.runtime.lastError) updateUI(response);
  });

  btnStop.addEventListener("click", () => {
    chrome.runtime.sendMessage({ type: "CMD_STOP_CAPTURE" }, (response) => {
      if (!chrome.runtime.lastError) updateUI(response);
    });
  });
});
