// A browser port of internal/pack's resolution, so the homepage terminal runs
// what `monom` would for the example project's commands. Keep the two in step.
// Real pack ignores the gate and runs any executable; the site knows output
// only for commands, so it resolves only their files.
import { ENTRIES } from "./project";

export type Packed =
  | { file: string } // runs this file
  | { group: true } // a command group: exit 3
  | { error: string }; // what mnmd pack prints

const COMMANDS = ENTRIES.filter((e) => e.outcome === "command");
const FILES = COMMANDS.map((e) => e.path);
// Only a script, a file with a shebang, resolves by its stem.
const SCRIPTS = COMMANDS.filter((e) => e.why.startsWith("#!")).map((e) => e.path);

/** internal/discover's Stem: name without its last "." suffix, or "" when it has none. */
function stem(name: string): string {
  const i = name.lastIndexOf(".");
  return i <= 0 || i === name.length - 1 ? "" : name.slice(0, i);
}

export function pack(root: string, words: string[]): Packed {
  const path = words.join("/");
  if (FILES.includes(path)) return { file: path };
  if (FILES.some((f) => f.startsWith(path + "/"))) return { group: true };
  const under = FILES.find((f) => path.startsWith(f + "/"));
  if (under) {
    return { error: `pack: cannot stat ${root}/${path}: stat ${root}/${path}: not a directory` };
  }

  // Nothing at the path: the last word may be a stripped command name.
  const dir = path.slice(0, path.lastIndexOf("/") + 1);
  const last = path.slice(dir.length);
  const matches = SCRIPTS.filter((f) => {
    const name = f.slice(dir.length);
    return f.startsWith(dir) && !name.includes("/") && stem(name) === last;
  });
  if (matches.length === 1) return { file: matches[0] };
  if (matches.length > 1) {
    const names = matches.map((f) => f.slice(dir.length)).sort();
    return { error: `pack: ambiguous command ${path}: matches ${names.join(", ")}` };
  }
  return { error: `pack: command not found: ${root}/${path}` };
}
