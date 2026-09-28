"use strict";

const assert = require("node:assert/strict");
const test = require("node:test");
const { validateFlags, validateReview } = require("./review-common.cjs");
const { interactiveScript, interactiveStyles } = require("./replay-browser.cjs");
const { clientScript, styles: editorStyles } = require("./editor-browser.cjs");

const artifactId = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
const flag = {
    id: "0123456789abcdef0123456789abcdef",
    startOffsetMs: 200,
    title: "Checkout failed",
    note: "The server returned an error.",
    category: "error",
    severity: "error",
};

function review(flags) {
    return {
        formatVersion: "1",
        kind: "flagged",
        rootArtifactId: artifactId,
        parentArtifactId: artifactId,
        revision: 1,
        reviewedAt: "2026-09-28T12:00:00Z",
        flags,
    };
}

test("validates and deterministically orders review flags", () => {
    const later = { ...flag, id: "fedcba9876543210fedcba9876543210", startOffsetMs: 900, title: "Later" };
    const result = validateReview(review([later, flag]), 1_000);
    assert.deepEqual(result.flags.map(value => value.id), [flag.id, later.id]);
});

test("rejects unsafe text, duplicate IDs, and invalid ranges", () => {
    assert.throws(() => validateFlags([{ ...flag, endOffsetMs: 200 }], 1_000), /greater/);
    assert.throws(() => validateFlags([flag, { ...flag, title: "Duplicate" }], 1_000), /duplicated/);
    assert.throws(() => validateFlags([{ ...flag, title: " untrimmed" }], 1_000), /trimmed/);
    assert.throws(() => validateReview(review([{ ...flag, startOffsetMs: 1_001 }]), 1_000), /between/);
});

test("generated Chromium clients are valid JavaScript", () => {
    assert.doesNotThrow(() => new Function(interactiveScript({ width: 1280, height: 720 }, { preRollMs: 0, postRollMs: 0 })));
    const editor = clientScript(
        { width: 1280, height: 720 },
        ["note"],
        ["info"],
        { defaultTitle: "Checkout review 1", durationMs: 1_000 },
    );
    assert.doesNotThrow(() => new Function(editor));
    assert.match(editor, /flag-handle/);
    assert.match(editor, /artifactTitle/);
    assert.match(editor, /publicationConfirmed/);
    assert.match(editor, /It is safe to close the Editor browser/);
    assert.match(editor, /if\(isSelected\).*flag-handle/);
    assert.match(editor, /marker-select/);
    assert.match(editor, /flags\.some\(flag=>flag\.id===previousId\)/);
    assert.match(editor, /lastDeleted/);
    assert.doesNotMatch(editor, /confirm\(/);
    assert.ok(editor.includes("const integer=value=>/^\\d+$/.test"));
});

test("replay reserves review panel width only when flags are shown", () => {
    const css = interactiveStyles({ width: 1280, height: 720 });
    assert.match(css, /#replay-app \{[^}]*grid-template-columns: minmax\(0, 1fr\);/);
    assert.match(css, /#replay-app\.has-review \{[^}]*grid-template-columns: minmax\(0, 1fr\) minmax\(230px, 300px\);/);
});

test("editor scales the fixed recorded viewport only once", () => {
    const css = editorStyles({ width: 1280, height: 720 });
    assert.match(css, /#stage\{[^}]*flex:0 0 auto;/);
    assert.match(css, /#view\{[^}]*min-width:0;[^}]*min-height:0;/);
    assert.match(css, /\*\[hidden\]\{display:none!important\}/);
    assert.match(css, /\.marker\.point\{top:25%;/);
    assert.match(css, /\.marker\.point::after\{[^}]*#58a6ff/);
    assert.match(css, /\.marker\.range\{top:75%;[^}]*rgba\(192,132,252/);
    assert.match(css, /\.marker-select\{[^}]*min-width:24px;/);
});
