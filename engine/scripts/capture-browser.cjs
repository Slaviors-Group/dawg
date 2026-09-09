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

function appendJSONL(filePath, value) {
    fs.appendFileSync(filePath, `${JSON.stringify(value)}\n`);
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

    const rrwebRoot = path.dirname(require.resolve("rrweb"));
    const rrwebBundle = fs.readFileSync(path.join(rrwebRoot, "rrweb.umd.cjs"), "utf8");
    const browser = await chromium.launch(process.env.DAWG_CHROMIUM_EXECUTABLE_PATH
        ? { executablePath: process.env.DAWG_CHROMIUM_EXECUTABLE_PATH }
        : {});
    const context = await browser.newContext();
    const page = await context.newPage();
    let stopped = false;

    await page.exposeBinding("__dawgRecordEvent", (_, event) => appendJSONL(options["rrweb-output"], event));
    await page.exposeBinding("__dawgRecordAction", (_, action) => appendJSONL(options["actions-output"], action));
    await page.addInitScript({ content: rrwebBundle });
    await page.addInitScript(() => {
        rrweb.record({ emit: (event) => window.__dawgRecordEvent(event) });
        document.addEventListener("click", (event) => {
            const element = event.target;
            window.__dawgRecordAction({ type: "click", timestamp: Date.now(), selector: element.id ? `#${element.id}` : element.tagName.toLowerCase() });
        }, true);
        document.addEventListener("input", (event) => {
            const element = event.target;
            window.__dawgRecordAction({ type: "fill", timestamp: Date.now(), selector: element.id ? `#${element.id}` : element.tagName.toLowerCase(), value: element.value });
        }, true);
    });

    await page.route("**/*", async (route) => {
        const request = route.request();
        const startedAt = Date.now();
        const response = await route.fetch();
        appendJSONL(options["http-output"], {
            id: crypto.randomUUID(),
            timestamp: new Date().toISOString(),
            request: { method: request.method(), url: request.url(), headers: await request.allHeaders(), body: request.postData() || "" },
            response: { status: response.status(), headers: response.headers(), body: await response.text() },
            direction: "frontend-to-backend",
            durationMs: Date.now() - startedAt
        });
        await route.fulfill({ response });
    });

    const stop = async () => {
        if (stopped) return;
        stopped = true;
        await browser.close();
    };
    process.once("SIGTERM", () => stop().then(() => process.exit(0)));
    process.once("SIGINT", () => stop().then(() => process.exit(0)));

    process.stdout.write(`${JSON.stringify({ status: "capturing" })}\n`);
    await page.goto(options.url).catch((error) => process.stderr.write(`${error.message}\n`));
}

main().catch((error) => {
    process.stderr.write(`${error.stack || error.message}\n`);
    process.exitCode = 1;
});