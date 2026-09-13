const fs = require("node:fs");
const path = require("node:path");
const readline = require("node:readline");
const { chromium } = require("playwright");

let activeBrowser;
let activeControl;
let activeContext;
let activePage;
let activeBrowserOwned = false;

function parseArguments(argumentsList) {
    const values = {};
    for (let index = 0; index < argumentsList.length; index += 2) {
        values[argumentsList[index].replace(/^--/, "")] = argumentsList[index + 1];
    }
    return values;
}

function appendJSONL(filePath, value) {
    fs.appendFileSync(filePath, `${JSON.stringify(value)}\n`);
}

function wait(milliseconds) {
    return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

async function settlePending(pending, timeoutMilliseconds) {
    if (pending.size === 0) return;
    await Promise.race([
        Promise.allSettled([...pending]),
        wait(timeoutMilliseconds),
    ]);
}

async function closeCaptureBrowser() {
    if (activePage && !activePage.isClosed()) {
        await activePage.close().catch(() => {});
    }
    if (activeBrowserOwned && activeContext) {
        await activeContext.close().catch(() => {});
    } else if (activeBrowser?.isConnected()) {
        // connectOverCDP Browser.close() disconnects this Playwright client; it
        // does not terminate the independently-owned Chrome/Edge process.
        await activeBrowser.close().catch(() => {});
    }
    activePage = undefined;
    activeContext = undefined;
    activeBrowser = undefined;
    activeBrowserOwned = false;
}

async function main() {
    const options = parseArguments(process.argv.slice(2));
    for (const required of ["url", "rrweb-output", "http-output", "actions-output"]) {
        if (!options[required]) {
            throw new Error(`missing --${required}`);
        }
    }
    for (const output of [options["rrweb-output"], options["http-output"], options["actions-output"]]) {
        fs.mkdirSync(path.dirname(output), { recursive: true });
    }

    const endpoint = options["cdp-endpoint"] || "http://127.0.0.1:9222";
    const rrwebRoot = path.dirname(require.resolve("rrweb"));
    const rrwebBundle = fs.readFileSync(path.join(rrwebRoot, "rrweb.umd.cjs"), "utf8");
    const stopSignal = {};
    stopSignal.promise = new Promise((resolve) => {
        stopSignal.resolve = resolve;
    });
    let stopRequested = false;
    let terminalError;

    const requestStop = (error) => {
        if (error && !terminalError) terminalError = error;
        if (stopRequested) return;
        stopRequested = true;
        stopSignal.resolve();
    };

    const control = readline.createInterface({ input: process.stdin });
    activeControl = control;
    control.on("line", (line) => {
        try {
            const message = JSON.parse(line);
            if (message.command === "stop") requestStop();
        } catch (error) {
            process.stderr.write(`Invalid recorder command: ${error.message}\n`);
        }
    });
    process.once("SIGTERM", () => requestStop());
    process.once("SIGINT", () => requestStop());

    let browser;
    let browserMode = "cdp";
    let context;
    try {
        browser = await chromium.connectOverCDP(endpoint, { timeout: 10_000 });
        context = browser.contexts()[0];
        if (!context) {
            await browser.close();
            throw new Error(`The CDP browser at ${endpoint} has no default browser context`);
        }
    } catch (cdpError) {
        if (!options["chromium-path"] || !options["browser-profile"]) {
            throw new Error(
                `Cannot connect to Chrome/Edge at ${endpoint}, and bundled Chromium fallback is unavailable. ` +
                `Start Chrome/Edge with --remote-debugging-port=9222 and a non-default --user-data-dir. ${cdpError.message}`,
            );
        }

        browserMode = "bundled";
        fs.mkdirSync(options["browser-profile"], { recursive: true, mode: 0o700 });
        const launchOptions = {
            executablePath: options["chromium-path"],
            headless: false,
            ignoreHTTPSErrors: true,
        };
        if (options["proxy-server"]) {
            launchOptions.proxy = { server: options["proxy-server"] };
        }
        try {
            context = await chromium.launchPersistentContext(options["browser-profile"], launchOptions);
            browser = context.browser();
            activeBrowserOwned = true;
        } catch (launchError) {
            throw new Error(
                `Cannot connect to Chrome/Edge at ${endpoint}, and bundled Chromium failed to start. ` +
                `CDP error: ${cdpError.message}. Bundled browser error: ${launchError.message}`,
            );
        }
    }
    activeBrowser = browser;
    activeContext = context;

    browser.once("disconnected", () => {
        if (!stopRequested) requestStop(new Error(`${browserMode} browser disconnected during capture`));
    });

    const existingPages = browserMode === "bundled" ? context.pages() : [];
    const page = existingPages[0] || await context.newPage();
    activePage = page;
    const pending = new Set();
    const track = (promise) => {
        pending.add(promise);
        promise.finally(() => pending.delete(promise));
    };

    await page.exposeBinding("__dawgRecordEvent", (_, event) => appendJSONL(options["rrweb-output"], event));
    await page.exposeBinding("__dawgRecordAction", (_, action) => appendJSONL(options["actions-output"], action));
    await page.addInitScript({ content: rrwebBundle });
    await page.addInitScript(() => {
        // Playwright re-runs addInitScript callbacks for every frame it
        // attaches to this page, including same-origin iframes the target
        // page creates itself (ads, analytics beacons, widgets, etc. often
        // use hidden `about:blank` iframes for this). Without this guard,
        // rrweb.record() starts once per iframe, and every one of those
        // recorders emits its own Meta/FullSnapshot into the same flat
        // event stream via __dawgRecordEvent. Replaying that merged stream
        // is undefined: a single rrweb.Replayer can only rebuild one
        // document, so it ends up applying whichever frame's snapshot it
        // saw last - typically one of the near-empty iframe snapshots -
        // instead of the real top-level page, producing a blank replay.
        if (window.top !== window.self) return;

        rrweb.record({ emit: (event) => window.__dawgRecordEvent(event) });
        document.addEventListener("click", (event) => {
            const element = event.target;
            if (!(element instanceof Element)) return;
            window.__dawgRecordAction({
                type: "click",
                timestamp: Date.now(),
                selector: element.id ? `#${CSS.escape(element.id)}` : element.tagName.toLowerCase(),
            });
        }, true);
        document.addEventListener("input", (event) => {
            const element = event.target;
            if (!(element instanceof HTMLInputElement || element instanceof HTMLTextAreaElement || element instanceof HTMLSelectElement)) return;
            window.__dawgRecordAction({
                type: "fill",
                timestamp: Date.now(),
                selector: element.id ? `#${CSS.escape(element.id)}` : element.tagName.toLowerCase(),
                value: element.value,
            });
        }, true);
    });

    const captureRequest = async (request) => {
        try {
            if (request.url().startsWith("data:")) return;
            const response = await request.response();
            if (!response) return;

            let bodyText = "";
            try {
                bodyText = await response.text();
            } catch {
                // Binary, streaming, and cached responses may not expose a text body.
            }

            const timing = request.timing();
            const durationMs = timing.responseEnd > 0
                ? Math.max(0, Math.round(timing.responseEnd - timing.requestStart))
                : 0;
            appendJSONL(options["http-output"], {
                id: crypto.randomUUID(),
                timestamp: new Date().toISOString(),
                request: {
                    method: request.method(),
                    url: request.url(),
                    headers: await request.allHeaders(),
                    body: request.postData() || "",
                },
                response: {
                    status: response.status(),
                    headers: response.headers(),
                    body: bodyText,
                },
                direction: "frontend-to-backend",
                durationMs,
            });
        } catch (error) {
            process.stderr.write(`Failed to capture request ${request.url()}: ${error.message}\n`);
        }
    };
    const onRequestFinished = (request) => track(captureRequest(request));
    page.on("requestfinished", onRequestFinished);

    try {
        await page.goto(options.url, { waitUntil: "domcontentloaded" });
    } catch (error) {
        page.off("requestfinished", onRequestFinished);
        await closeCaptureBrowser();
        throw new Error(`Cannot navigate the ${browserMode} capture tab to ${options.url}: ${error.message}`);
    }

    process.stdout.write(`${JSON.stringify({ status: "capturing", browserMode, cdpEndpoint: endpoint })}\n`);
    await stopSignal.promise;

    page.off("requestfinished", onRequestFinished);
    await settlePending(pending, 2_000);
    await closeCaptureBrowser();
    control.close();
    activeControl = undefined;
    process.stdout.write(`${JSON.stringify({ status: "stopped" })}\n`);

    if (terminalError) throw terminalError;
}

main().catch(async (error) => {
    await closeCaptureBrowser();
    activeControl?.close();
    process.stderr.write(`${error.stack || error.message}\n`);
    process.exitCode = 1;
});
