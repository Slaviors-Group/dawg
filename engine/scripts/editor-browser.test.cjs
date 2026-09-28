"use strict";

const assert = require("node:assert/strict");
const test = require("node:test");
const { chromium } = require("playwright");
const { clientScript, pageHtml, styles } = require("./editor-browser.cjs");

const point = {
    id: "11111111111111111111111111111111",
    startOffsetMs: 2_000,
    title: "Point finding",
    note: "",
    category: "note",
    severity: "info",
};
const range = {
    id: "22222222222222222222222222222222",
    startOffsetMs: 5_000,
    endOffsetMs: 5_001,
    title: "Short range",
    note: "",
    category: "bug",
    severity: "warning",
};

const rrwebMock = `window.__dawgEditorIntent=async()=>({});window.rrweb={unpack:value=>value,ReplayerEvents:{Start:"start",Pause:"pause",Resize:"resize"},Replayer:class{constructor(events,options){this.time=0;this.handlers={};options.root.append(document.createElement("iframe"))}getMetaData(){return{totalTime:10000}}getCurrentTime(){return this.time}on(name,handler){this.handlers[name]=handler}pause(value){if(value!==undefined)this.time=Number(value);this.handlers.pause?.()}play(value){if(value!==undefined)this.time=Number(value);this.handlers.start?.()}setConfig(){}}};`;

async function editorPage(browser) {
    const page = await browser.newPage();
    const resources = new Map([
        ["/", ["text/html", pageHtml()]],
        ["/rrweb.js", ["text/javascript", rrwebMock]],
        ["/rrweb.css", ["text/css", ""]],
        ["/editor.css", ["text/css", styles({ width: 1280, height: 720 })]],
        ["/editor.js", ["text/javascript", clientScript(
            { width: 1280, height: 720 },
            ["bug", "note"],
            ["info", "warning"],
            { defaultTitle: "Review 1", durationMs: 10_000 },
        )]],
        ["/events.json", ["application/json", JSON.stringify([{ type: 4, timestamp: 1_000, data: { width: 1280, height: 720 } }])]],
        ["/review.json", ["application/json", JSON.stringify({ flags: [point, range] })]],
    ]);
    await page.route("**/*", route => {
        const url = new URL(route.request().url());
        const resource = resources.get(url.pathname);
        return resource
            ? route.fulfill({ contentType: resource[0], body: resource[1] })
            : route.fulfill({ status: 404, body: "Not found" });
    });
    await page.goto("http://dawg-editor.local/");
    await page.waitForFunction(() => window.__DAWG_EDITOR__ || window.__DAWG_EDITOR_ERROR__);
    const error = await page.evaluate(() => window.__DAWG_EDITOR_ERROR__ || null);
    assert.equal(error, null);
    return page;
}

test("editor timeline selection and delete undo remain usable", { timeout: 30_000 }, async t => {
    const browser = await chromium.launch({ headless: true });
    t.after(() => browser.close());
    const page = await editorPage(browser);

    assert.equal(await page.locator(".marker.point").count(), 1);
    assert.equal(await page.locator(".marker.range").count(), 1);
    assert.equal(await page.locator(".flag-handle").count(), 1);
    assert.match(await page.locator("#form-heading").textContent(), /Edit point · 1 of 2/);

    await page.locator(".marker.range .marker-select").click();
    assert.equal(await page.locator(".flag-handle").count(), 2);
    assert.match(await page.locator("#form-heading").textContent(), /Edit range · 2 of 2/);
    const rangeHitWidth = await page.locator(".marker.range .marker-select").evaluate(element => element.getBoundingClientRect().width);
    assert.ok(rangeHitWidth >= 24, `expected a usable range hit target, got ${rangeHitWidth}`);

    await page.locator("#delete").click();
    assert.equal(await page.locator(".marker.range").count(), 0);
    assert.equal(await page.locator("#undo-row").isVisible(), true);

    await page.locator("#undo-delete").click();
    assert.equal(await page.locator(".marker.range").count(), 1);
    assert.equal(await page.locator(".flag-handle").count(), 2);

    await page.evaluate(() => window.__DAWG_EDITOR__.publicationConfirmed({ artifactTitle: "Review 1" }));
    assert.equal(await page.locator("#publication-message").isVisible(), true);
    assert.match(await page.locator("#publication-message").textContent(), /Artifact saved as a new artifact “Review 1”\. It is safe to close/);

    await page.evaluate(() => window.__DAWG_EDITOR__.replaceFlags([]));
    assert.equal(await page.locator("#form").isVisible(), false);
    assert.equal(await page.locator("#undo-row").isVisible(), false);
    assert.equal(await page.locator("#save").isDisabled(), true);
    assert.equal(await page.locator("#draft").isDisabled(), true);
});
