const fs = require("node:fs");
const path = require("node:path");
const { chromium } = require("playwright");

const REPLAY_PROTOCOL = "dawg.replay.v1";
let replaySequence = 0;

function emitReplayEvent(event) {
    replaySequence += 1;
    process.stdout.write(`${JSON.stringify({ protocol: REPLAY_PROTOCOL, sequence: replaySequence, ...event })}\n`);
}

function parseArguments(argumentsList) {
    const values = {};
    for (let index = 0; index < argumentsList.length; index += 1) {
        const argument = argumentsList[index];
        if (!argument.startsWith("--")) {
            continue;
        }

        const name = argument.slice(2);
        const next = argumentsList[index + 1];
        if (next && !next.startsWith("--")) {
            values[name] = next;
            index += 1;
        } else {
            values[name] = true;
        }
    }
    return values;
}

function getRecordedViewport(events) {
    const fallback = { width: 1280, height: 720 };
    const metaEvent = events.find(event => {
        if (event.type !== 4 || !event.data) {
            return false;
        }
        const width = Math.round(Number(event.data.width));
        const height = Math.round(Number(event.data.height));
        return Number.isFinite(width) && width >= 240 && width <= 3840
            && Number.isFinite(height) && height >= 160 && height <= 2160;
    });

    if (!metaEvent) {
        process.stderr.write(
            `WARNING: no usable rrweb Meta viewport found; using ${fallback.width}x${fallback.height}.\n`,
        );
        return fallback;
    }

    return {
        width: Math.round(Number(metaEvent.data.width)),
        height: Math.round(Number(metaEvent.data.height)),
    };
}

function interactiveDocument() {
    return `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>DAWG Replay</title>
    <link rel="stylesheet" href="/rrweb.css">
    <link rel="stylesheet" href="/replay.css">
    <script src="/rrweb.js"></script>
    <script src="/replay.js" defer></script>
</head>
<body>
    <main id="replay-app">
        <section id="viewport-host" aria-label="Recorded browser viewport">
            <div id="recorded-stage"><div id="replay-root"></div></div>
        </section>
        <section id="replay-controls" aria-label="Replay controls">
            <button id="play-pause" type="button" aria-label="Play replay">Play</button>
            <button id="rewind" type="button" aria-label="Rewind 10 seconds">-10s</button>
            <button id="forward" type="button" aria-label="Forward 10 seconds">+10s</button>
            <button id="speed" type="button" aria-label="Playback speed">1x</button>
            <input id="timeline" type="range" min="0" value="0" step="100" aria-label="Replay timeline">
            <output id="time-label" for="timeline">0:00 / 0:00</output>
        </section>
    </main>
</body>
</html>`;
}

function interactiveStyles(viewport) {
    return `:root { color-scheme: dark; font-family: system-ui, sans-serif; }
* { box-sizing: border-box; }
body { margin: 0; min-width: 320px; overflow: hidden; background: #141414; }
#replay-app { display: grid; grid-template-rows: minmax(0, 1fr) auto; height: 100vh; }
#viewport-host { display: flex; min-height: 0; align-items: center; justify-content: center; overflow: hidden; background: #050505; }
#recorded-stage { width: ${viewport.width}px; height: ${viewport.height}px; flex: 0 0 auto; overflow: hidden; transform-origin: center center; background: white; box-shadow: 0 0 28px rgba(0, 0, 0, .7); }
#replay-root, #replay-root > .replayer-wrapper { width: 100%; height: 100%; }
#replay-controls { display: grid; grid-template-columns: auto auto auto auto minmax(120px, 1fr) auto; gap: 8px; align-items: center; padding: 10px 14px; border-top: 1px solid #343434; background: #1e1e1e; color: #f3f3f3; }
#replay-controls button { min-width: 44px; min-height: 32px; border: 1px solid #5a5a5a; border-radius: 4px; background: #2c2c2c; color: inherit; cursor: pointer; font: inherit; }
#replay-controls button:hover { background: #3a3a3a; }
#replay-controls button:focus-visible, #timeline:focus-visible { outline: 2px solid #77a7ff; outline-offset: 2px; }
#timeline { width: 100%; accent-color: #77a7ff; cursor: pointer; }
#time-label { min-width: 118px; color: #d0d0d0; font-variant-numeric: tabular-nums; text-align: right; white-space: nowrap; }
@media (max-width: 560px) { #replay-controls { grid-template-columns: auto auto auto auto minmax(80px, 1fr); } #time-label { grid-column: 1 / -1; text-align: left; } }`;
}

function interactiveScript(recordedViewport) {
    return `(async () => {
    try {
        const response = await fetch("/events.json");
        if (!response.ok) throw new Error(\`events request failed: \${response.status}\`);
        const events = await response.json();
        const root = document.getElementById("replay-root");
        const controls = document.getElementById("replay-controls");
        const viewportHost = document.getElementById("viewport-host");
        const stage = document.getElementById("recorded-stage");
        const playPause = document.getElementById("play-pause");
        const rewind = document.getElementById("rewind");
        const forward = document.getElementById("forward");
        const speed = document.getElementById("speed");
        const timeline = document.getElementById("timeline");
        const timeLabel = document.getElementById("time-label");
        if (!root || !controls || !viewportHost || !stage || !playPause || !rewind || !forward || !speed || !timeline || !timeLabel) {
            throw new Error("interactive replay controls could not be initialized");
        }

        const replayer = new rrweb.Replayer(events, { root, unpackFn: rrweb.unpack });
        const replayerMetadata = replayer.getMetaData();
        const replayDuration = Math.max(0, replayerMetadata.totalTime);
        const metadata = {
            ...replayerMetadata,
            eventCount: events.length,
            firstTimestamp: events[0].timestamp,
            lastTimestamp: events[events.length - 1].timestamp,
            recordedViewport: ${JSON.stringify(recordedViewport)},
        };
        const speeds = [0.5, 1, 1.5, 2, 4];
        let speedIndex = speeds.indexOf(1);
        let playing = false;
        let scrubbing = false;

        const replayState = (type = "state", reason = "update") => ({
            type,
            reason,
            currentTimeMs: Math.max(0, Math.min(replayDuration, replayer.getCurrentTime())),
            durationMs: replayDuration,
            firstTimestamp: metadata.firstTimestamp,
            lastTimestamp: metadata.lastTimestamp,
            playing,
            speed: speeds[speedIndex],
        });
        const reportState = (type = "state", reason = "update") => {
            const state = replayState(type, reason);
            if (typeof window.__dawgReportReplayState === "function") {
                void window.__dawgReportReplayState(state);
            }
            return state;
        };

        const formatTime = milliseconds => {
            const totalSeconds = Math.floor(Math.max(0, milliseconds) / 1000);
            const seconds = totalSeconds % 60;
            const totalMinutes = Math.floor(totalSeconds / 60);
            const minutes = totalMinutes % 60;
            const hours = Math.floor(totalMinutes / 60);
            const base = \`\${minutes}:\${String(seconds).padStart(2, "0")}\`;
            return hours > 0 ? \`\${hours}:\${base.padStart(5, "0")}\` : base;
        };
        const clamp = value => Math.min(replayDuration, Math.max(0, Number(value) || 0));
        const updateControls = timeOffset => {
            const currentTime = clamp(timeOffset);
            if (!scrubbing) timeline.value = String(currentTime);
            timeLabel.textContent = \`\${formatTime(scrubbing ? Number(timeline.value) : currentTime)} / \${formatTime(replayDuration)}\`;
            playPause.textContent = playing ? "Pause" : "Play";
            playPause.setAttribute("aria-label", playing ? "Pause replay" : "Play replay");
            speed.textContent = \`\${speeds[speedIndex]}x\`;
        };
        const play = () => {
            if (replayer.getCurrentTime() >= replayDuration) replayer.pause(0);
            replayer.play(clamp(replayer.getCurrentTime()));
        };
        const pause = () => replayer.pause();
        const seek = timeOffset => {
            const target = clamp(timeOffset);
            const shouldPlay = playing;
            replayer.pause(target);
            if (shouldPlay) replayer.play(target);
            updateControls(target);
            reportState("state", "seek");
            return target;
        };
        const setSpeed = value => {
            const index = speeds.indexOf(Number(value));
            if (index < 0) throw new Error("unsupported replay speed");
            speedIndex = index;
            replayer.setConfig({ speed: speeds[speedIndex] });
            updateControls(replayer.getCurrentTime());
            reportState("state", "speed");
            return speeds[speedIndex];
        };
        const getState = () => replayState("state", "requested");
        const restart = () => seek(0);
        const resizeStage = () => {
            const scale = Math.max(0.01, Math.min(
                viewportHost.clientWidth / metadata.recordedViewport.width,
                viewportHost.clientHeight / metadata.recordedViewport.height,
            ));
            stage.style.transform = \`scale(\${scale})\`;
        };

        timeline.max = String(replayDuration);
        replayer.on(rrweb.ReplayerEvents.Start, () => {
            playing = true;
            updateControls(replayer.getCurrentTime());
            reportState("state", "play");
        });
        replayer.on(rrweb.ReplayerEvents.Pause, () => {
            playing = false;
            updateControls(replayer.getCurrentTime());
            reportState("state", "pause");
        });
        replayer.on(rrweb.ReplayerEvents.Finish, () => {
            playing = false;
            updateControls(replayDuration);
            reportState("finished", "finish");
        });
        playPause.addEventListener("click", () => playing ? pause() : play());
        rewind.addEventListener("click", () => seek(replayer.getCurrentTime() - 10000));
        forward.addEventListener("click", () => seek(replayer.getCurrentTime() + 10000));
        speed.addEventListener("click", () => {
            setSpeed(speeds[(speedIndex + 1) % speeds.length]);
        });
        timeline.addEventListener("pointerdown", () => { scrubbing = true; });
        timeline.addEventListener("input", () => updateControls(Number(timeline.value)));
        timeline.addEventListener("change", () => {
            scrubbing = false;
            seek(Number(timeline.value));
        });
        timeline.addEventListener("pointerup", () => { scrubbing = false; });
        new ResizeObserver(resizeStage).observe(viewportHost);
        resizeStage();
        window.setInterval(() => {
            updateControls(replayer.getCurrentTime());
            if (playing) reportState("state", "tick");
        }, 100);

        replayer.pause(0);
        updateControls(0);
        const iframe = root.querySelector("iframe");
        if (!iframe) throw new Error("rrweb replay iframe was not created");
        iframe.id = "dawg-replay-frame";
        iframe.name = "dawg-replay-frame";
        iframe.title = "DAWG replayed page";
        iframe.dataset.dawgReplay = "true";
        window.__DAWG_REPLAY__ = { events, replayer, iframe, play, pause, seek, setSpeed, getState, restart, metadata };
    } catch (error) {
        window.__DAWG_REPLAY_ERROR__ = error && (error.stack || error.message) || String(error);
        console.error(error);
    }
})();`;
}

function installInteractiveControlChannel(page) {
    let buffered = "";
    let closed = false;
    const processLine = async line => {
        if (!line.trim()) return;
        let command;
        try {
            command = JSON.parse(line);
            if (!command || typeof command.type !== "string") throw new Error("missing command type");
            if (command.protocol && command.protocol !== REPLAY_PROTOCOL) throw new Error("unsupported replay protocol");
            if (command.type === "seek" && !Number.isFinite(command.offsetMs)) throw new Error("seek requires a finite offsetMs");
            if (command.type === "setSpeed" && ![0.5, 1, 1.5, 2, 4].includes(command.speed)) throw new Error("unsupported replay speed");
            if (!["play", "pause", "seek", "setSpeed", "getState"].includes(command.type)) throw new Error("unsupported replay command");
        } catch (error) {
            emitReplayEvent({ type: "error", commandId: command?.id || null, message: error.message || String(error) });
            return;
        }
        try {
            const state = await page.evaluate(replayCommand => {
                const controller = window.__DAWG_REPLAY__;
                if (!controller) throw new Error("replay controls are not ready");
                switch (replayCommand.type) {
                    case "play": controller.play(); break;
                    case "pause": controller.pause(); break;
                    case "seek": controller.seek(replayCommand.offsetMs); break;
                    case "setSpeed": controller.setSpeed(replayCommand.speed); break;
                    case "getState": break;
                    default: throw new Error("unsupported replay command");
                }
                return controller.getState();
            }, command);
            emitReplayEvent({ type: "ack", commandId: command.id || null, command: command.type });
            emitReplayEvent({ ...state, type: "state", reason: "command" });
        } catch (error) {
            if (!closed) emitReplayEvent({ type: "error", commandId: command.id || null, message: error.message || String(error) });
        }
    };
    process.stdin.setEncoding("utf8");
    process.stdin.on("data", chunk => {
        buffered += chunk;
        const lines = buffered.split("\n");
        buffered = lines.pop();
        for (const line of lines) void processLine(line);
    });
    process.stdin.on("end", () => { closed = true; });
    process.stdin.resume();
}

async function main() {
    const options = parseArguments(process.argv.slice(2));
    for (const required of ["rrweb-input", "screenshot-output"]) {
        if (!options[required]) {
            throw new Error(`missing --${required}`);
        }
    }
    const interactive = options.interactive === true;

    const rrwebRoot = path.dirname(require.resolve("rrweb"));
    const rrwebBundle = fs.readFileSync(path.join(rrwebRoot, "rrweb.umd.cjs"), "utf8");
    // Without rrweb's official replay stylesheet, `.replayer-wrapper` isn't
    // `position: relative` and `.replayer-mouse`/`.replayer-mouse-tail`
    // aren't `position: absolute`. The mouse-tail canvas then renders as a
    // normal block box (sized to the recorded viewport) ABOVE the replayed
    // iframe instead of overlaying it, pushing all real page content down
    // by a full viewport height and making the replay look blank/broken.
    const rrwebCss = fs.readFileSync(path.join(rrwebRoot, "style.css"), "utf8");

    // Read and parse JSONL events.
    const rawEvents = fs.readFileSync(options["rrweb-input"], "utf8");
    const events = rawEvents
        .split("\n")
        .map(line => line.trim())
        .filter(line => line.length > 0)
        .map(line => JSON.parse(line));

    if (events.length === 0) {
        throw new Error("no rrweb events found in input");
    }

    const recordedViewport = interactive ? getRecordedViewport(events) : null;

    // Diagnostics: a replayed page that renders blank despite the process
    // exiting successfully is otherwise invisible to the user (no error is
    // ever thrown). Log the shape of the trace up front so a degenerate
    // capture (e.g. no FullSnapshot, or a snapshot of an unrelated blank
    // document) shows up immediately in DAWG's Execution Logs.
    const typeCounts = {};
    for (const event of events) {
        typeCounts[event.type] = (typeCounts[event.type] || 0) + 1;
    }
    const metaEvents = events.filter(event => event.type === 4);
    process.stderr.write(`Loaded ${events.length} rrweb events, type counts: ${JSON.stringify(typeCounts)}\n`);
    process.stderr.write(`Meta events (href/viewport): ${JSON.stringify(metaEvents.map(event => event.data))}\n`);
    if (!typeCounts[2]) {
        process.stderr.write("WARNING: no FullSnapshot (type 2) event found in trace — replay will render nothing.\n");
    }

    // Use the Playwright-managed Chromium bundled with DAWG. An explicit
    // executable remains available for CI and advanced deployments, but replay
    // must not silently depend on a separately installed Google Chrome.
    const launchOptions = { headless: false };
    if (recordedViewport) {
        // The outer Chromium window needs room for browser chrome and the
        // player controls below the fixed-size recorded stage.
        const windowHeight = Math.min(2400, recordedViewport.height + 160);
        launchOptions.args = [`--window-size=${recordedViewport.width},${windowHeight}`];
    }
    if (process.env.DAWG_CHROMIUM_EXECUTABLE_PATH) {
        launchOptions.executablePath = process.env.DAWG_CHROMIUM_EXECUTABLE_PATH;
    }
    if (options["proxy-server"]) {
        launchOptions.proxy = { server: options["proxy-server"] };
    }

    const browser = await chromium.launch(launchOptions);
    try {
        const contextOptions = { ignoreHTTPSErrors: true };
        if (interactive) {
            // A null viewport follows the Chromium window size. The fixed replay
            // stage below then scales into that available area on resize/maximize.
            contextOptions.viewport = null;
        }
        const context = await browser.newContext(contextOptions);
        const page = await context.newPage();

        // Surface in-page console/errors in the engine output. A replayed
        // page can render blank because of an in-page JS error (e.g. an
        // incompatible/legacy rrweb payload) that never bubbles up through
        // page.evaluate() if it happens asynchronously (rAF callbacks,
        // mutation processing, etc.).
        page.on("console", msg => {
            process.stderr.write(`[replay page console:${msg.type()}] ${msg.text()}\n`);
        });
        page.on("pageerror", err => {
            process.stderr.write(`[replay page error] ${err.message}\n`);
        });

        const eventsJson = JSON.stringify(events);
        let evaluation;

        if (interactive) {
            await page.exposeFunction("__dawgReportReplayState", event => emitReplayEvent(event));
            const resources = new Map([
                ["/", { contentType: "text/html", body: interactiveDocument() }],
                ["/index.html", { contentType: "text/html", body: interactiveDocument() }],
                ["/replay.js", { contentType: "text/javascript", body: interactiveScript(recordedViewport) }],
                ["/replay.css", { contentType: "text/css", body: interactiveStyles(recordedViewport) }],
                ["/rrweb.js", { contentType: "text/javascript", body: rrwebBundle }],
                ["/rrweb.css", { contentType: "text/css", body: rrwebCss }],
                ["/events.json", { contentType: "application/json", body: eventsJson }],
            ]);
            await page.route("http://dawg-replay.local/**", route => {
                const resource = resources.get(new URL(route.request().url()).pathname);
                if (resource) {
                    return route.fulfill(resource);
                }
                return route.fulfill({ status: 404, contentType: "text/plain", body: "Not found" });
            });
            await page.goto("http://dawg-replay.local/", { waitUntil: "load" });
            await page.waitForFunction(() => window.__DAWG_REPLAY__ || window.__DAWG_REPLAY_ERROR__);
            evaluation = await page.evaluate(() => ({
                duration: window.__DAWG_REPLAY__
                    ? window.__DAWG_REPLAY__.metadata.totalTime
                    : 0,
                replayerError: window.__DAWG_REPLAY_ERROR__ || null,
                iframeCount: document.querySelectorAll("iframe").length,
                replayState: window.__DAWG_REPLAY__ ? window.__DAWG_REPLAY__.getState() : null,
            }));
        } else {
            await page.route("http://dawg-replay.local/events.json", route => {
                route.fulfill({ contentType: "application/json", body: eventsJson });
            });
            await page.setContent("<!DOCTYPE html><html><head><style>body { margin: 0; padding: 0; }</style></head><body></body></html>");
            await page.addScriptTag({ content: rrwebBundle });
            await page.addStyleTag({ content: rrwebCss });
            evaluation = await page.evaluate(async () => {
                const response = await fetch("http://dawg-replay.local/events.json");
                const events = await response.json();
                const duration = events[events.length - 1].timestamp - events[0].timestamp;
                let replayerError = null;
                try {
                    const replayer = new rrweb.Replayer(events, {
                        root: document.body,
                        unpackFn: rrweb.unpack,
                    });
                    // Preserve automated replay behavior: autoplay through the
                    // trace, wait, capture the screenshot, and close Chromium.
                    replayer.play();
                } catch (error) {
                    replayerError = error && (error.stack || error.message);
                }
                return {
                    duration,
                    replayerError,
                    iframeCount: document.querySelectorAll("iframe").length,
                };
            });
        }

        if (evaluation.replayerError) {
            throw new Error(`rrweb.Replayer failed: ${evaluation.replayerError}`);
        }
        process.stderr.write(`Replayer attached ${evaluation.iframeCount} iframe(s) to the page.\n`);

        if (interactive) {
            installInteractiveControlChannel(page);
            emitReplayEvent({ ...evaluation.replayState, type: "ready", reason: "ready" });
            process.stderr.write(`Interactive replay ready at http://dawg-replay.local/ with recorded viewport ${recordedViewport.width}x${recordedViewport.height}. Close the replay window when finished.\n`);
            await Promise.race([
                new Promise(resolve => browser.once("disconnected", resolve)),
                new Promise(resolve => page.once("close", resolve)),
            ]);
            emitReplayEvent({ type: "closed" });
            return;
        }

        // Wait for the full replay duration plus a buffer for final rendering.
        const waitMs = evaluation.duration + 2000;
        process.stderr.write(`Replay duration: ${Math.round(evaluation.duration / 1000)}s — waiting ${Math.round(waitMs / 1000)}s for playback...\n`);
        await new Promise(resolve => setTimeout(resolve, waitMs));

        const replayedTextLength = await page.evaluate(() => {
            const iframe = document.querySelector("iframe");
            try {
                return iframe && iframe.contentDocument ? iframe.contentDocument.body.innerText.length : -1;
            } catch (_error) {
                return -1;
            }
        });
        process.stderr.write(`Replayed iframe visible text length: ${replayedTextLength}\n`);

        fs.mkdirSync(path.dirname(options["screenshot-output"]), { recursive: true });
        await page.screenshot({ path: options["screenshot-output"], fullPage: true });
    } finally {
        // Always close the browser, even if replay failed partway through, so
        // a malformed trace or a runtime error never leaves an orphaned
        // Chromium window stuck open on the user's desktop.
        await browser.close().catch(() => {});
    }
}

main().catch((error) => {
    process.stderr.write(`${error.stack || error.message}\n`);
    process.exitCode = 1;
});
