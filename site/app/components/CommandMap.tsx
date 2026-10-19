import { useState } from "react";
import { ENTRIES, PROJECT, type Entry } from "~/lib/project";

// The homepage's mapping. Same example project as every other tree
// (lib/project.ts). Folders are nouns, the file you pick is the verb, and
// the line above the tree is the command that file becomes.

type Kind = "root" | "group" | "cmd";

type Node = {
  id: string;
  label: string;
  depth: number;
  kind: Kind;
  words: string[];
  file?: string;
  lang?: string;
  /** Command words a group would list, the way `monom <group>` does. */
  avail?: string;
  ancestors: string[];
};

type Branch = {
  id: string;
  name: string;
  depth: number;
  kind: Kind;
  words: string[];
  file?: string;
  lang?: string;
  ancestors: string[];
  children: Branch[];
};

function fileName(entry: Entry) {
  const path = entry.path.replace(/\/$/, "");
  return path.slice(path.lastIndexOf("/") + 1);
}

function build(): Node[] {
  const root: Branch = {
    id: PROJECT,
    name: PROJECT,
    depth: 0,
    kind: "root",
    words: [],
    ancestors: [],
    children: [],
  };

  for (const entry of ENTRIES) {
    if (entry.outcome !== "command" || !entry.command) continue;
    const words = entry.command.split("/");
    let parent = root;
    const ancestors = [PROJECT];
    words.forEach((part, i) => {
      const last = i === words.length - 1;
      const id = words.slice(0, i + 1).join("/");
      let child = parent.children.find((c) => c.id === id);
      if (!child) {
        child = {
          id,
          name: last ? fileName(entry) : part,
          depth: i + 1,
          kind: last ? "cmd" : "group",
          words: words.slice(0, i + 1),
          ancestors: [...ancestors],
          children: [],
          ...(last ? { file: entry.path, lang: entry.lang } : {}),
        };
        parent.children.push(child);
      }
      ancestors.push(id);
      parent = child;
    });
  }

  const flat: Node[] = [];
  const walk = (node: Branch) => {
    const avail =
      node.kind === "group"
        ? node.children.map((c) => (c.kind === "cmd" ? c.words[c.words.length - 1] : c.name)).join(", ")
        : undefined;
    flat.push({
      id: node.id,
      label: node.kind === "cmd" ? node.name : `${node.name.replace(/\/$/, "")}/`,
      depth: node.depth,
      kind: node.kind,
      words: node.words,
      file: node.file,
      lang: node.lang,
      avail,
      ancestors: node.ancestors,
    });
    node.children.forEach(walk);
  };
  walk(root);
  return flat;
}

const NODES = build();
const COMMANDS = NODES.filter((n) => n.kind === "cmd");
const INITIAL = "infra/cloud/deploy";

export function CommandMap() {
  const [current, setCurrent] = useState(INITIAL);
  const [tabDown, setTabDown] = useState(false);
  const selected = NODES.find((n) => n.id === current) ?? COMMANDS[0];
  const ancestors = new Set(selected.ancestors);

  const cycle = () => {
    setTabDown(true);
    window.setTimeout(() => setTabDown(false), 120);
    const at = COMMANDS.findIndex((n) => n.id === current);
    if (at === -1) {
      const from = NODES.findIndex((n) => n.id === current);
      const next = NODES.slice(from + 1).find((n) => n.kind === "cmd") ?? COMMANDS[0];
      setCurrent(next.id);
      return;
    }
    setCurrent(COMMANDS[(at + 1) % COMMANDS.length].id);
  };

  return (
    <div className="command-map">
      <div className="command-result" aria-live="polite">
        <p className="command-you">you type</p>
        <p className="command-line">
          <span className="command-ps">monom</span>
          {selected.words.map((word, i) => {
            const last = i === selected.words.length - 1;
            const verb = selected.kind === "cmd" && last;
            return (
              <span key={word} className={verb ? "is-command-text" : "is-group-text"}>
                {" "}
                {word}
              </span>
            );
          })}
        </p>
        <p className="command-meta">
          {selected.kind === "group" ? (
            <>
              {selected.words[selected.words.length - 1]} is a command group · available: {selected.avail}
            </>
          ) : (
            <>
              runs <code>{selected.file}</code>
              {selected.lang ? ` · ${selected.lang}` : ""}
            </>
          )}
        </p>
      </div>

      <div className="picker">
        <div className="picker-head">
          <span>{PROJECT}</span>
          <span className="picker-legend">
            <span>
              <i className="is-folder" /> folder
            </span>
            <span>
              <i className="is-script" /> script
            </span>
          </span>
        </div>
        <div className="picker-tree" role="listbox" aria-label={`${PROJECT} project`}>
          {NODES.map((node) => {
            const on = node.id === current;
            if (node.kind === "root") {
              return (
                <div key={node.id} className="picker-row is-root" style={{ paddingLeft: 14 }}>
                  {node.label}
                </div>
              );
            }
            return (
              <button
                key={node.id}
                type="button"
                role="option"
                aria-selected={on}
                className={[
                  "picker-row",
                  node.kind === "group" ? "is-group" : "is-cmd",
                  on && node.kind === "cmd" ? "is-on" : "",
                  on && node.kind === "group" ? "is-on-group" : "",
                  ancestors.has(node.id) ? "is-ancestor" : "",
                ].join(" ")}
                style={{ paddingLeft: 14 + node.depth * 18 }}
                onClick={() => setCurrent(node.id)}
              >
                {node.depth > 0 && (
                  <span className="picker-guides" style={{ width: node.depth * 18 - 8 }} aria-hidden="true" />
                )}
                <span>{node.label}</span>
                {node.lang && <span className="picker-lang">{node.lang}</span>}
              </button>
            );
          })}
        </div>
        <div className="picker-foot">
          <span>Folders list. Scripts run.</span>
          <button type="button" className={`tab-key ${tabDown ? "is-down" : ""}`} onClick={cycle}>
            Tab ⇥
          </button>
        </div>
      </div>
    </div>
  );
}
