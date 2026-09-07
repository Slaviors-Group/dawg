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

    const browser = await chromium.launch(process.env.DAWG_CHROMIUM_EXECUTABLE_PATH
        ? { executablePath: process.env.DAWG_CHROMIUM_EXECUTABLE_PATH }
        : {});
    const context = await browser.newContext();
    const page = await context.newPage();

    // Make the events available to the browser context via routing to avoid large evaluate payloads
    await page.route("http://dawg-replay.local/events.json", route => {
        route.fulfill({
            contentType: "application/json",
            body: JSON.stringify(events)
        });
    });

    await page.setContent('<!DOCTYPE html><html><head><style>body { margin: 0; padding: 0; }</style></head><body></body></html>');
    await page.addScriptTag({ content: rrwebBundle });

    await page.evaluate(async () => {
        const response = await fetch("http://dawg-replay.local/events.json");
        const events = await response.json();

        const replayer = new rrweb.Replayer(events, {
            root: document.body,
            unpackFn: rrweb.unpack,
        });

        // Seek to the end of the recording to render the final state
        const firstTimestamp = events[0].timestamp;
        const lastTimestamp = events[events.length - 1].timestamp;
        replayer.pause(lastTimestamp - firstTimestamp);

        // Give the DOM a tiny bit of time to settle just in case there are images loading
        await new Promise(resolve => setTimeout(resolve, 500));
    });

    fs.mkdirSync(path.dirname(options["screenshot-output"]), { recursive: true });
    await page.screenshot({ path: options["screenshot-output"], fullPage: true });

    await browser.close();
}

main().catch((error) => {
    process.stderr.write(`${error.stack || error.message}\n`);
    process.exitCode = 1;
});
