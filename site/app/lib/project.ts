// The one example project every file tree, command, and output on the site
// uses. Its real counterpart is fixtures/site-project (tools/lint and
// tools/fmt are compiled from tests/harness/testdata/site-binaries);
// tests/cases/site_homepage.yaml runs it in bash and zsh, and
// TestSiteExamples (tests/harness/site_test.go) checks project.json against
// that fixture and those cases. Change all three together.
import entries from "./project.json";

export type Lang = "python" | "node" | "bash" | "sh" | "ruby" | "go" | "rust" | "text";

/** What default discovery does with an entry. */
export type Outcome =
  | "command" // registered: shows up in Tab and runs
  | "warn" // executable but rejected by the gate: `mnmd check` warns
  | "ignored" // not executable
  | "hidden"; // never scanned (_-prefixed, node_modules, ...)

export type Entry = {
  /** Path from the project root. A trailing slash marks a directory entry. */
  path: string;
  /** The command path discovery registers it as, suffix stripped. Commands only. */
  command?: string;
  outcome: Outcome;
  lang?: Lang;
  /** Why the gate decided what it did. Starts with the shebang when there is one. */
  why: string;
  /** What running it prints. Commands only. */
  output?: string;
};

export const PROJECT = "acme";

export const ENTRIES = entries as Entry[];

const COMMANDS = ENTRIES.filter((e) => e.outcome === "command");

/** Registered command paths, in `mnmd discover` order (sorted). */
export const paths = COMMANDS.map((e) => e.command!).sort();

/** Command path → what it prints when run. */
export const commands: Record<string, string> = Object.fromEntries(COMMANDS.map((e) => [e.command!, e.output ?? ""]));

/** Command path → the file it runs. */
export const fileOf: Record<string, string> = Object.fromEntries(COMMANDS.map((e) => [e.command!, e.path]));

/** File path → the command path it registers as. */
export const commandOf: Record<string, string> = Object.fromEntries(COMMANDS.map((e) => [e.path, e.command!]));

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
  rust: "rs",
  text: "md",
};
