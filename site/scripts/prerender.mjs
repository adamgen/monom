// Prerender the Remix SSR build into static HTML so build/client can be served as pure static assets.
import { createRequestHandler } from "@remix-run/node";
import { copyFile, mkdir, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";

const build = await import(new URL("../build/server/index.js", import.meta.url));
const handler = createRequestHandler(build, "production");
const out = new URL("../build/client/", import.meta.url).pathname;
const pages = { "/": "index.html", "/docs": "docs/index.html", "/__404": "404.html" };

for (const [path, file] of Object.entries(pages)) {
  const res = await handler(new Request(`http://localhost${path}`));
  const html = await res.text();
  if (path !== "/__404" && res.status !== 200) throw new Error(`${path} -> ${res.status}`);
  const dest = join(out, file);
  await mkdir(dirname(dest), { recursive: true });
  await writeFile(dest, html);
  console.log(`${path} -> ${file} (${res.status}, ${html.length} bytes)`);
}

// monom.dev/install.sh is the repo-root install.sh, copied here so there is one
// source of truth (never commit a copy under site/). public/_headers serves it
// as text/plain; CI checks the built file is byte-identical to the original.
const installScript = new URL("../../install.sh", import.meta.url).pathname;
await copyFile(installScript, join(out, "install.sh"));
console.log(`../install.sh -> install.sh`);
