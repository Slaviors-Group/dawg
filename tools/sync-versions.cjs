#!/usr/bin/env node
/**
 * Synchronize DAWG version metadata from ../version.json.
 *
 * Usage:
 *   node tools/sync-versions.cjs          # update tracked metadata
 *   node tools/sync-versions.cjs --check  # report drift without modifying files
 *
 * Third-party dependency versions (Playwright, rrweb, Rust crates) deliberately
 * remain in their own lockfiles/package manifests. Updating those requires a
 * dependency upgrade workflow, not a release-version bump.
 */
const fs = require("node:fs");
const path = require("node:path");

const root = path.resolve(__dirname, "..");
const checkOnly = process.argv.slice(2).includes("--check");
const unexpectedArgs = process.argv.slice(2).filter((arg) => arg !== "--check");
if (unexpectedArgs.length > 0) {
  throw new Error(`Unknown argument(s): ${unexpectedArgs.join(", ")}`);
}

const configPath = path.join(root, "version.json");
const versions = JSON.parse(fs.readFileSync(configPath, "utf8"));
const semver = /^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/;
const majorMinorRelease = /^(\d+)\.(\d+)$/;
const namedRelease = /^(\d+)\.(\d+)[-_]([0-9A-Za-z._-]+)$/;
const extensionSemver = /^\d+\.\d+\.\d+(?:\.\d+)?$/;
const namedExtensionRelease = /^(\d+)\.(\d+)(?:\.(\d+))?[-_]([0-9A-Za-z._-]+)$/;

// Allow convenient MAJOR.MINOR and MAJOR.MINOR-label shorthand in
// version.json, e.g. "0.2" or "0.2-naughty". Package managers and Tauri
// require strict three-component SemVer, so generated metadata receives
// "0.2.0" or "0.2.0-naughty" respectively. The label is preserved.
function normalizeReleaseVersion(value, field) {
  if (semver.test(value)) return value;
  const plainMatch = majorMinorRelease.exec(value);
  if (plainMatch) return `${plainMatch[1]}.${plainMatch[2]}.0`;
  const namedMatch = namedRelease.exec(value);
  if (namedMatch) {
    // `_` is convenient in a user-entered label but invalid in SemVer
    // prerelease identifiers. Convert it to `-` in generated metadata.
    return `${namedMatch[1]}.${namedMatch[2]}.0-${namedMatch[3].replaceAll("_", "-")}`;
  }
  throw new Error(
    `version.json ${field} must be MAJOR.MINOR[.PATCH][-_label] ` +
    `(for example 0.2, 0.2-naughty, or 0.2.0-naughty): ${value}`,
  );
}

function normalizeExtensionVersion(value) {
  if (extensionSemver.test(value)) return { version: value, versionName: null };
  const plainMatch = majorMinorRelease.exec(value);
  if (plainMatch) return { version: `${plainMatch[1]}.${plainMatch[2]}.0`, versionName: null };
  const namedMatch = namedExtensionRelease.exec(value);
  if (namedMatch) {
    // Chrome requires a numeric `version`, but supports an arbitrary display
    // label in `version_name`. Keep the label there while producing a valid
    // installable extension manifest.
    return {
      version: `${namedMatch[1]}.${namedMatch[2]}.${namedMatch[3] ?? "0"}`,
      versionName: value,
    };
  }
  throw new Error(
    `version.json extensionVersion must begin with MAJOR.MINOR[.PATCH] ` +
    `(for example 0.2, 0.2_naughty, 0.2.3, or 0.2.3_naughty): ${value}`,
  );
}

const appVersion = normalizeReleaseVersion(versions.appVersion, "appVersion");
const desktopVersion = normalizeReleaseVersion(versions.desktopVersion, "desktopVersion");
const extensionVersion = normalizeExtensionVersion(versions.extensionVersion);
for (const key of ["mitmproxy", "node"]) {
  if (!semver.test(versions.runtime?.[key])) {
    throw new Error(`version.json runtime.${key} is not valid semver: ${versions.runtime?.[key]}`);
  }
}

let changed = false;
const drift = [];

function relative(file) {
  return path.relative(root, file).replaceAll(path.sep, "/");
}

function writeText(file, contents) {
  const previous = fs.readFileSync(file, "utf8");
  if (previous === contents) return;
  drift.push(relative(file));
  if (!checkOnly) fs.writeFileSync(file, contents);
  changed = true;
}

function updateJSON(relativePath, mutate) {
  const file = path.join(root, relativePath);
  const original = fs.readFileSync(file, "utf8");
  const document = JSON.parse(original);
  const before = JSON.stringify(document);
  mutate(document);
  // Avoid unrelated formatting churn when the managed value is already in
  // sync. When a release value changes, JSON.stringify provides predictable
  // formatting for the affected manifest/package file.
  if (JSON.stringify(document) === before) return;
  writeText(file, `${JSON.stringify(document, null, 2)}\n`);
}

function replaceRequired(file, pattern, replacement) {
  const original = fs.readFileSync(file, "utf8");
  if (!pattern.test(original)) {
    throw new Error(`Expected version field was not found in ${relative(file)}`);
  }
  writeText(file, original.replace(pattern, replacement));
}

function currentSchemaVersion() {
  const manifestDir = path.join(root, "schema", "manifest");
  const files = fs.readdirSync(manifestDir).filter((file) => /^v.+\.json$/.test(file));
  if (files.length !== 1) {
    throw new Error(`Expected one schema manifest in ${relative(manifestDir)}, found: ${files.join(", ")}`);
  }
  return files[0].slice(1, -".json".length);
}

const schemaVersion = appVersion;
const oldSchemaVersion = currentSchemaVersion();

updateJSON("engine/package.json", (pkg) => { pkg.version = appVersion; });
updateJSON("desktop/package.json", (pkg) => { pkg.version = appVersion; });
updateJSON("extension/manifest.json", (manifest) => {
  manifest.version = extensionVersion.version;
  if (extensionVersion.versionName) {
    manifest.version_name = extensionVersion.versionName;
  } else {
    delete manifest.version_name;
  }
});
updateJSON("desktop/src-tauri/tauri.conf.json", (config) => { config.version = desktopVersion; });

for (const lockFile of ["engine/package-lock.json", "desktop/package-lock.json"]) {
  updateJSON(lockFile, (lock) => {
    if (lock.packages?.[""]) lock.packages[""].version = appVersion;
  });
}

replaceRequired(
  path.join(root, "desktop", "src-tauri", "Cargo.toml"),
  /^version = ".+"$/m,
  `version = "${desktopVersion}"`,
);
replaceRequired(
  path.join(root, "engine", "cmd", "dawg", "main.go"),
  /^const version = ".+"$/m,
  `const version = "${appVersion}"`,
);
replaceRequired(
  path.join(root, "desktop", "build-bundle.ps1"),
  /\[string\]\$MitmVersion = "[^"]+"/,
  `[string]$MitmVersion = "${versions.runtime.mitmproxy}"`,
);
replaceRequired(
  path.join(root, "desktop", "build-bundle.ps1"),
  /\[string\]\$NodeVersion = "[^"]+"/,
  `[string]$NodeVersion = "${versions.runtime.node}"`,
);
replaceRequired(
  path.join(root, "desktop", "build-bundle.sh"),
  /^MITM_VERSION="[^"]+"$/m,
  `MITM_VERSION="${versions.runtime.mitmproxy}"`,
);
replaceRequired(
  path.join(root, "desktop", "build-bundle.sh"),
  /^NODE_VERSION="[^"]+"$/m,
  `NODE_VERSION="${versions.runtime.node}"`,
);
replaceRequired(
  path.join(root, "desktop", "src", "components", "ui", "SettingsPanel.tsx"),
  /v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?/,
  `v${appVersion}`,
);
replaceRequired(
  path.join(root, "desktop", "src", "components", "ArtifactInspectorModal.tsx"),
  /nodeVersion: "[^"]+"/,
  `nodeVersion: "${versions.runtime.node}"`,
);

// Schema-version references must change as a group with its filename. This
// includes Go tests/test fixtures, keeping a release bump testable immediately.
const schemaReferenceFiles = [
  "engine/cmd/dawg/init.go",
  "engine/cmd/dawg/main_test.go",
  "engine/internal/capture/envsnap.go",
  "engine/internal/capture/envsnap_test.go",
  "engine/internal/dawgenv/env.go",
  "engine/internal/manifest/manifest.go",
  "engine/internal/manifest/manifest_test.go",
  "engine/internal/manifest/testdata/invalid-missing-title.json",
  "engine/internal/manifest/testdata/valid.json",
  "engine/internal/packager/layout_test.go",
  "README.md",
];
for (const relativePath of schemaReferenceFiles) {
  const file = path.join(root, relativePath);
  const original = fs.readFileSync(file, "utf8");
  writeText(file, original.split(oldSchemaVersion).join(schemaVersion));
}

const oldSchemaFile = path.join(root, "schema", "manifest", `v${oldSchemaVersion}.json`);
const newSchemaFile = path.join(root, "schema", "manifest", `v${schemaVersion}.json`);
const schemaContents = fs.readFileSync(oldSchemaFile, "utf8").split(oldSchemaVersion).join(schemaVersion);
if (oldSchemaFile !== newSchemaFile) {
  drift.push(`${relative(oldSchemaFile)} -> ${relative(newSchemaFile)}`);
  if (!checkOnly) {
    fs.writeFileSync(newSchemaFile, schemaContents);
    fs.unlinkSync(oldSchemaFile);
  }
  changed = true;
} else {
  writeText(oldSchemaFile, schemaContents);
}

if (checkOnly) {
  if (changed) {
    console.error(`Version metadata is out of sync with version.json:\n${drift.map((file) => `  - ${file}`).join("\n")}`);
    process.exitCode = 1;
  } else {
    console.log(`Version metadata is in sync (${appVersion}).`);
  }
} else if (changed) {
  console.log(`Synchronized version metadata from version.json (${appVersion}).`);
  for (const file of drift) console.log(`  updated ${file}`);
} else {
  console.log(`Version metadata already in sync (${appVersion}).`);
}
