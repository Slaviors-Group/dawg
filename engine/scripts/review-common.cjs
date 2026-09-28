"use strict";

const MAX_FLAGS = 500;
const MAX_TITLE_LENGTH = 120;
const MAX_NOTE_LENGTH = 2000;
const FLAG_ID_PATTERN = /^[a-f0-9]{32}$/;
const ARTIFACT_ID_PATTERN = /^sha256:[a-f0-9]{64}$/;
const CATEGORIES = new Set(["bug", "error", "network", "console", "action", "note"]);
const SEVERITIES = new Set(["info", "warning", "error"]);

function fail(message) {
    throw new Error(`invalid review: ${message}`);
}

function isPlainObject(value) {
    return value !== null && typeof value === "object" && !Array.isArray(value);
}

function integer(value, field, minimum, maximum = Number.MAX_SAFE_INTEGER) {
    if (!Number.isSafeInteger(value) || value < minimum || value > maximum) {
        fail(`${field} must be an integer between ${minimum} and ${maximum}`);
    }
    return value;
}

function string(value, field, maximum, { required = true, trim = false } = {}) {
    if (value === undefined && !required) return undefined;
    if (typeof value !== "string") fail(`${field} must be a string`);
    if (value.length > maximum) fail(`${field} exceeds ${maximum} characters`);
    if (trim && value !== value.trim()) fail(`${field} must be trimmed`);
    if (required && value.length === 0) fail(`${field} is required`);
    return value;
}

function utcRfc3339(value) {
    string(value, "reviewedAt", 64);
    if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?Z$/.test(value)
        || Number.isNaN(Date.parse(value))) {
        fail("reviewedAt must be a UTC RFC3339 timestamp");
    }
    return value;
}

function validateFlags(flags, durationMs = Number.MAX_SAFE_INTEGER) {
    if (!Array.isArray(flags) || flags.length === 0) fail("flags must contain at least one item");
    if (flags.length > MAX_FLAGS) fail(`flags may contain at most ${MAX_FLAGS} items`);
    integer(durationMs, "durationMs", 0);

    const ids = new Set();
    const normalized = flags.map((flag, index) => {
        if (!isPlainObject(flag)) fail(`flags[${index}] must be an object`);
        const id = string(flag.id, `flags[${index}].id`, 32);
        if (!FLAG_ID_PATTERN.test(id)) fail(`flags[${index}].id must be 32 lowercase hexadecimal characters`);
        if (ids.has(id)) fail(`flags[${index}].id is duplicated`);
        ids.add(id);

        const startOffsetMs = integer(flag.startOffsetMs, `flags[${index}].startOffsetMs`, 0, durationMs);
        const endOffsetMs = flag.endOffsetMs === undefined || flag.endOffsetMs === null
            ? undefined
            : integer(flag.endOffsetMs, `flags[${index}].endOffsetMs`, 0, durationMs);
        if (endOffsetMs !== undefined && endOffsetMs <= startOffsetMs) {
            fail(`flags[${index}].endOffsetMs must be greater than startOffsetMs`);
        }

        const title = string(flag.title, `flags[${index}].title`, MAX_TITLE_LENGTH, { trim: true });
        const note = string(flag.note, `flags[${index}].note`, MAX_NOTE_LENGTH, { required: false });
        const category = string(flag.category, `flags[${index}].category`, 16);
        const severity = string(flag.severity, `flags[${index}].severity`, 16);
        if (!CATEGORIES.has(category)) fail(`flags[${index}].category is not supported`);
        if (!SEVERITIES.has(severity)) fail(`flags[${index}].severity is not supported`);

        const normalizedFlag = { id, startOffsetMs, title, category, severity };
        if (endOffsetMs !== undefined) normalizedFlag.endOffsetMs = endOffsetMs;
        if (note !== undefined) normalizedFlag.note = note;
        return normalizedFlag;
    });

    return normalized.sort((left, right) => left.startOffsetMs - right.startOffsetMs || left.id.localeCompare(right.id));
}

function unwrapReview(value) {
    if (isPlainObject(value) && isPlainObject(value.review)) return value.review;
    return value;
}

function validateReview(value, durationMs = Number.MAX_SAFE_INTEGER, { allowEmpty = false } = {}) {
    if (value === undefined || value === null) {
        if (allowEmpty) return null;
        fail("review is required");
    }
    const review = unwrapReview(value);
    if (!isPlainObject(review)) fail("review must be an object");
    if (review.formatVersion !== "1") fail("formatVersion must be exactly 1");
    if (review.kind !== "flagged") fail("kind must be exactly flagged");
    for (const field of ["rootArtifactId", "parentArtifactId"]) {
        const artifactId = string(review[field], field, 71);
        if (!ARTIFACT_ID_PATTERN.test(artifactId)) fail(`${field} must be a DAWG SHA-256 artifact ID`);
    }
    const revision = integer(review.revision, "revision", 1);
    const reviewedAt = utcRfc3339(review.reviewedAt);
    const flags = validateFlags(review.flags, durationMs);
    return {
        formatVersion: "1",
        kind: "flagged",
        rootArtifactId: review.rootArtifactId,
        parentArtifactId: review.parentArtifactId,
        revision,
        reviewedAt,
        flags,
    };
}

function reviewFlags(value, durationMs) {
    if (value === undefined || value === null) return [];
    return validateReview(value, durationMs).flags;
}

function parseReviewFile(filePath, durationMs, { allowEmpty = true } = {}) {
    if (filePath === undefined || filePath === null) return null;
    if (typeof filePath !== "string" || filePath.length === 0) throw new Error("review file path is required");
    const fs = require("node:fs");
    let value;
    try {
        value = JSON.parse(fs.readFileSync(filePath, "utf8"));
    } catch (error) {
        throw new Error(`cannot read review file: ${error.message}`);
    }
    return validateReview(value, durationMs, { allowEmpty });
}

function makeFlagId(randomBytes) {
    if (!randomBytes || randomBytes.length !== 16) throw new Error("flag ID requires 16 random bytes");
    return Buffer.from(randomBytes).toString("hex");
}

module.exports = {
    ARTIFACT_ID_PATTERN,
    CATEGORIES: [...CATEGORIES],
    MAX_FLAGS,
    MAX_NOTE_LENGTH,
    MAX_TITLE_LENGTH,
    SEVERITIES: [...SEVERITIES],
    makeFlagId,
    parseReviewFile,
    reviewFlags,
    validateFlags,
    validateReview,
};
