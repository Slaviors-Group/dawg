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

    // Launch the user's installed Chrome in isolated mode (incognito)
    const launchOptions = { headless: false, channel: "chrome" };

    if (options["proxy-server"]) {
        launchOptions.proxy = { server: options["proxy-server"] };
    }

    const browser = await chromium.launch(launchOptions);
    const context = await browser.newContext({ ignoreHTTPSErrors: true });
    const page = await context.newPage();

    // Make the events available to the browser context via routing
    await page.route("http://dawg-replay.local/events.json", route => {
        route.fulfill({
            contentType: "application/json",
            body: JSON.stringify(events)
        });
    });

    await page.setContent('<!DOCTYPE html><html><head><style>body { margin: 0; padding: 0; }</style></head><body></body></html>');
    await page.addScriptTag({ content: rrwebBundle });

    const replayDurationMs = await page.evaluate(async () => {
        const response = await fetch("http://dawg-replay.local/events.json");
        const events = await response.json();

        const firstTimestamp = events[0].timestamp;
        const lastTimestamp = events[events.length - 1].timestamp;
        const duration = lastTimestamp - firstTimestamp;

        const replayer = new rrweb.Replayer(events, {
            root: document.body,
            unpackFn: rrweb.unpack,
        });

        // Play events in real-time instead of seeking to the end
        replayer.play();
        return duration;
    });

    // Wait for the full replay duration plus a buffer for final rendering
    const waitMs = replayDurationMs + 2000;
    process.stderr.write(`Replay duration: ${Math.round(replayDurationMs / 1000)}s — waiting ${Math.round(waitMs / 1000)}s for playback...\n`);
    await new Promise(resolve => setTimeout(resolve, waitMs));

    fs.mkdirSync(path.dirname(options["screenshot-output"]), { recursive: true });
    await page.screenshot({ path: options["screenshot-output"], fullPage: true });

    await browser.close();
}

main().catch((error) => {
    process.stderr.write(`${error.stack || error.message}\n`);
    process.exitCode = 1;
});
