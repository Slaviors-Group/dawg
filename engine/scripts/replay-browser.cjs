const fs = require("node:fs");
const path = require("node:path");
const { chromium } = require("playwright");

function parseArguments(argumentsList) {
    const values = {};
    for (let index = 0; index < argumentsList.length; index += 2) {
        values[argumentsList[index].replace(/^--/, "")] = argumentsList[index + 1];
    }
    return values;
}

async function main() {
    const options = parseArguments(process.argv.slice(2));
    for (const required of ["rrweb-input", "screenshot-output"]) {
        if (!options[required]) {
            throw new Error(`missing --${required}`);
        }
    }

    const rrwebRoot = path.dirname(require.resolve("rrweb"));
    const rrwebBundle = fs.readFileSync(path.join(rrwebRoot, "rrweb.umd.cjs"), "utf8");
    // Without rrweb's official replay stylesheet, `.replayer-wrapper` isn't
    // `position: relative` and `.replayer-mouse`/`.replayer-mouse-tail`
    // aren't `position: absolute`. The mouse-tail canvas then renders as a
    // normal block box (sized to the recorded viewport) ABOVE the replayed
    // iframe instead of overlaying it, pushing all real page content down
    // by a full viewport height and making the replay look blank/broken.
    const rrwebCss = fs.readFileSync(path.join(rrwebRoot, "style.css"), "utf8");

    // Read and parse JSONL events
    const rawEvents = fs.readFileSync(options["rrweb-input"], "utf8");
    const events = rawEvents
        .split("\n")
        .map(line => line.trim())
        .filter(line => line.length > 0)
        .map(line => JSON.parse(line));

    if (events.length === 0) {
        throw new Error("no rrweb events found in input");
    }

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
    if (process.env.DAWG_CHROMIUM_EXECUTABLE_PATH) {
        launchOptions.executablePath = process.env.DAWG_CHROMIUM_EXECUTABLE_PATH;
    }

    if (options["proxy-server"]) {
        launchOptions.proxy = { server: options["proxy-server"] };
    }

    const browser = await chromium.launch(launchOptions);
    try {
        const context = await browser.newContext({ ignoreHTTPSErrors: true });
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

        // Make the events available to the browser context via routing
        await page.route("http://dawg-replay.local/events.json", route => {
            route.fulfill({
                contentType: "application/json",
                body: JSON.stringify(events)
            });
        });

        await page.setContent('<!DOCTYPE html><html><head><style>body { margin: 0; padding: 0; }</style></head><body></body></html>');
        await page.addScriptTag({ content: rrwebBundle });
        await page.addStyleTag({ content: rrwebCss });

        const evaluation = await page.evaluate(async () => {
            const response = await fetch("http://dawg-replay.local/events.json");
            const events = await response.json();

            const firstTimestamp = events[0].timestamp;
            const lastTimestamp = events[events.length - 1].timestamp;
            const duration = lastTimestamp - firstTimestamp;

            let replayerError = null;
            try {
                const replayer = new rrweb.Replayer(events, {
                    root: document.body,
                    unpackFn: rrweb.unpack,
                });

                // Play events in real-time instead of seeking to the end
                replayer.play();
            } catch (error) {
                replayerError = error && (error.stack || error.message);
            }

            return { duration, replayerError, iframeCount: document.querySelectorAll("iframe").length };
        });

        if (evaluation.replayerError) {
            throw new Error(`rrweb.Replayer failed: ${evaluation.replayerError}`);
        }
        process.stderr.write(`Replayer attached ${evaluation.iframeCount} iframe(s) to the page.\n`);

        // Wait for the full replay duration plus a buffer for final rendering
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
