const fs = require("node:fs");
const path = require("node:path");

const extension = path.join(process.cwd(), "extension");
const oldFile = path.join(extension, "lib", "rrweb.min.js");
const newFile = path.join(extension, "lib", "rrweb.js");

let source = fs.readFileSync(oldFile, "utf8");

const encoded = source.match(/const encodedJs = "([A-Za-z0-9+/=]+)";/);
if (!encoded) {
  throw new Error("Could not find rrweb's encoded worker payload.");
}

const workerSource = Buffer.from(encoded[1], "base64").toString("utf8");
const templateLiteral = workerSource
  .replace(/\\/g, "\\\\")
  .replace(/`/g, "\\`")
  .replace(/\$\{/g, "\\${");

source = source.replace(
  encoded[0],
  `const imageBitmapDataUrlWorkerSource = \`${templateLiteral}\`;`,
);

source = source.replace(/const decodeBase64 = .*?;\n/, "");
source = source.replace(
  "new Blob([decodeBase64(encodedJs)]",
  "new Blob([imageBitmapDataUrlWorkerSource]",
);
source = source.replace(
  /return new Worker\(\s*"data:text\/javascript;base64," \+ encodedJs,\s*\{\s*name: options == null \? void 0 : options.name\s*\}\s*\);/s,
  'throw new Error("Unable to create the rrweb image bitmap worker");',
);

if (/const encodedJs\s*=|data:text\/javascript;base64|[A-Za-z0-9+/]{200,}={0,2}/.test(source)) {
  throw new Error("Encoded code still exists; aborting.");
}

fs.writeFileSync(newFile, `${source.trimEnd()}\n`);
fs.unlinkSync(oldFile);

function replaceInFile(relativePath) {
  const file = path.join(extension, relativePath);
  const content = fs.readFileSync(file, "utf8")
    .replaceAll("lib/rrweb.min.js", "lib/rrweb.js");
  fs.writeFileSync(file, content);
}

replaceInFile("manifest.json");
replaceInFile("background/service_worker.js");
replaceInFile("tests/service_worker.test.cjs");

const readme = path.join(extension, "README.md");
fs.writeFileSync(
  readme,
  fs.readFileSync(readme, "utf8").replace(
    "working on the extension.\n",
    `working on the extension.

## Third-party source

The extension ships rrweb \`2.0.0-alpha.18\` as its official unminified UMD
distribution in [\`lib/rrweb.js\`](./lib/rrweb.js). Its image-bitmap worker
source is also stored as readable JavaScript rather than an encoded payload,
for Chrome Web Store review. Do not add a minified or encoded rrweb build to
the extension package.
`,
  ),
);

console.log("Created extension/lib/rrweb.js and removed the encoded payload.");