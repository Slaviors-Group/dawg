import { open, save } from "@tauri-apps/plugin-dialog";

const dawgFilter = { name: "DAWG artifact", extensions: ["dawg"] };

export async function chooseArtifactArchive(): Promise<string | null> {
  const selected = await open({
    multiple: false,
    directory: false,
    filters: [dawgFilter],
  });
  return typeof selected === "string" ? selected : null;
}

export async function chooseArtifactExportPath(defaultName: string): Promise<string | null> {
  const selected = await save({
    defaultPath: ensureDawgExtension(defaultName),
    filters: [dawgFilter],
  });
  return selected ? ensureDawgExtension(selected) : null;
}

export function defaultArtifactExportName(title: string, createdAt: string): string {
  const timestamp = new Date(createdAt)
    .toISOString()
    .replace(/[-:]/g, "")
    .replace(/\.\d{3}Z$/, "");
  const readableTitle = slugify(title) || "dawg-artifact";
  return `${timestamp}-${readableTitle}.dawg`;
}

function ensureDawgExtension(path: string): string {
  return path.toLowerCase().endsWith(".dawg") ? path : `${path}.dawg`;
}

function slugify(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}
