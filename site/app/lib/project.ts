// The one example project every file tree, command, and output on the site
// uses. It was built as a real project (a git repo, no monom file) and checked
// against the real binaries: `mnmd discover`, `mnmd check`, running each
// command, and Tab completion in bash and zsh. Change it only by rebuilding
// that project and re-checking what the site quotes.

export type Lang = "python" | "node" | "bash" | "sh" | "ruby" | "go" | "text";

/** What default discovery does with an entry. */
export type Outcome =
  | "command" // registered: shows up in Tab and runs
  | "warn" // executable but rejected by the gate: `mnmd check` warns
  | "ignored" // not executable
  | "hidden"; // never scanned (_-prefixed, node_modules, ...)

export type Entry = {
  /** Path from the project root. A trailing slash marks a directory entry. */
  path: string;
  outcome: Outcome;
  lang?: Lang;
  /** Why the gate decided what it did. */
  why: string;
  /** What running it prints. Commands only. */
  output?: string;
};

export const PROJECT = "acme";

export const ENTRIES: Entry[] = [
  { path: "db/migrate.py", outcome: "command", lang: "python", why: "#!/usr/bin/env python3", output: "running db migrations..." },
  { path: "db/seed.js", outcome: "command", lang: "node", why: "#!/usr/bin/env node", output: "seeding database..." },
  { path: "infra/cloud/deploy.sh", outcome: "command", lang: "bash", why: "#!/usr/bin/env bash", output: "deploying to cloud..." },
  { path: "infra/local/start", outcome: "command", lang: "bash", why: "#!/usr/bin/env bash, no extension", output: "starting local environment..." },
  { path: "tools/lint", outcome: "command", lang: "go", why: "compiled Go binary: no shebang, command-shaped name", output: "lint: 0 issues" },
  { path: "tools/Format.sh", outcome: "warn", lang: "sh", why: "no shebang, and the name has a capital and a dot" },
  { path: "_lib/log.sh", outcome: "hidden", lang: "bash", why: "_-prefixed: private, never scanned" },
  { path: "node_modules/", outcome: "hidden", why: "built-in skip list, like vendor, venv, dist" },
  { path: "README.md", outcome: "ignored", lang: "text", why: "not executable" },
  { path: "release.rb", outcome: "command", lang: "ruby", why: "#!/usr/bin/env ruby", output: "releasing..." },
];

/** Registered command paths, in `mnmd discover` order (sorted). */
export const paths = ENTRIES.filter((e) => e.outcome === "command")
  .map((e) => e.path)
  .sort();

/** Command path → what it prints when run. */
export const commands: Record<string, string> = Object.fromEntries(
  ENTRIES.filter((e) => e.outcome === "command").map((e) => [e.path, e.output ?? ""]),
);

export const cmd = (path: string) => `monom ${path.split("/").join(" ")}`;

export const warnings = ENTRIES.filter((e) => e.outcome === "warn").map(
  (e) =>
    `warning: [root-contents] executable not registered: ${e.path} (no shebang, and the name does not match ^[a-z0-9][a-z0-9_-]*$); add a shebang, rename it, or declare it in the monom config file`,
);

export const DISCOVER = ["$ mnmd discover", ...paths].join("\n");

export const CHECK = [
  "$ mnmd check",
  ...warnings,
  `✔ ${paths.length} commands OK (default discovery), ${warnings.length} warning(s)`,
].join("\n");

export const LANG_LABEL: Record<Lang, string> = {
  python: "py",
  node: "js",
  bash: "bash",
  sh: "sh",
  ruby: "rb",
  go: "go",
  text: "md",
};
