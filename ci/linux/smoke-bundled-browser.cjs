"use strict";

const assert = require("node:assert/strict");
const path = require("node:path");

async function main() {
    const resources = process.argv[2];
    const executablePath = process.env.DAWG_CHROMIUM_EXECUTABLE_PATH;
    if (!resources || !executablePath || !process.env.APPDIR) {
        throw new Error("usage: APPDIR=... DAWG_CHROMIUM_EXECUTABLE_PATH=... smoke-bundled-browser.cjs RESOURCES");
    }

    const { chromium } = require(path.join(resources, "node_modules", "playwright"));
    const { hostBrowserEnvironment } = require(path.join(resources, "scripts", "review-common.cjs"));
    const browserEnvironment = hostBrowserEnvironment();
    const appDir = path.resolve(process.env.APPDIR);

    assert.equal(browserEnvironment.APPDIR, undefined);
    for (const entry of String(browserEnvironment.LD_LIBRARY_PATH || "").split(path.delimiter).filter(Boolean)) {
        const relative = path.relative(appDir, path.resolve(entry));
        assert.ok(relative === ".." || relative.startsWith(`..${path.sep}`), `Chromium inherited AppDir library path: ${entry}`);
    }

    const browser = await chromium.launch({
        env: browserEnvironment,
        executablePath,
        headless: true,
    });
    try {
        const page = await browser.newPage();
        await page.setContent("<title>DAWG bundled browser smoke test</title><main>ok</main>");
        assert.equal(await page.title(), "DAWG bundled browser smoke test");
        assert.equal(await page.textContent("main"), "ok");
    } finally {
        await browser.close().catch(() => {});
    }

    process.stdout.write("Bundled Chromium smoke test passed.\n");
}

main().catch(error => {
    process.stderr.write(`${error.stack || error.message}\n`);
    process.exitCode = 1;
});