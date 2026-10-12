// A browser port of internal/filter/filter.go, so the homepage demo completes
// exactly the way `mnmd filter` does. Keep the two in step.
export function filter(commands: string[], words: string[]): string[] {
  let completed: string[] = [];
  let partial = "";
  if (words.length > 0) {
    const last = words[words.length - 1];
    completed = words.slice(0, -1);
    partial = last;
  }

  const seen = new Set<string>();
  const result: string[] = [];
  for (const cmd of commands) {
    const parts = cmd.split("/");
    if (parts.some((p) => p.includes(" "))) continue;
    if (parts.length <= completed.length) continue;
    if (!completed.every((w, i) => parts[i] === w)) continue;
    const token = parts[completed.length];
    if (!token.startsWith(partial)) continue;
    if (!seen.has(token)) {
      seen.add(token);
      result.push(token);
    }
  }
  return result;
}

export function commonPrefix(items: string[]): string {
  if (items.length === 0) return "";
  let prefix = items[0];
  for (const item of items.slice(1)) {
    while (!item.startsWith(prefix)) prefix = prefix.slice(0, -1);
  }
  return prefix;
}
