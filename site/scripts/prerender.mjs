// Prerender the Remix SSR build into static HTML so build/client can be served as pure static assets.
import { createRequestHandler } from "@remix-run/node";
import { mkdir, writeFile } from "node:fs/promises";
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
