import { ENTRIES, LANG_LABEL, PROJECT, cmd, type Entry, type Outcome } from "~/lib/project";

// The site's only file tree. Every tree on every page renders the same
// project (lib/project.ts); sections differ only in what they annotate.

type Node = { name: string; path: string[]; entry?: Entry; children: Node[]; dir: boolean };

function build(): Node[] {
  const root: Node = { name: "", path: [], children: [], dir: true };
  for (const entry of ENTRIES) {
    const isDir = entry.path.endsWith("/");
    const parts = entry.path.replace(/\/$/, "").split("/");
    let node = root;
    parts.forEach((part, i) => {
      let child = node.children.find((c) => c.name === part);
      if (!child) {
        child = { name: part, path: [...node.path, part], children: [], dir: i < parts.length - 1 || isDir };
        node.children.push(child);
      }
      node = child;
    });
    node.entry = entry;
  }
  return root.children;
}

const NODES = build();

/** A directory takes the outcome of its best child: a group if it holds any command. */
function outcomeOf(node: Node): Outcome {
  if (node.entry) return node.entry.outcome;
  const kids = node.children.map(outcomeOf);
  for (const o of ["command", "warn", "ignored", "hidden"] as const) if (kids.includes(o)) return o;
  return "hidden";
}

const NOTE: Record<Outcome, string> = {
  command: "",
  warn: "mnmd check warns",
  ignored: "ignored",
  hidden: "never scanned",
};

type Props = {
  /**
   * What to annotate beside each file:
   * - "none": names and language badges only
   * - "gate": the discovery gate's decision and its reason
   */
  annotate?: "none" | "gate";
  /** Completed tokens the user has typed — highlights the branch (the Tab demo). */
  active?: string[];
  /** The token being typed — highlights matching siblings. */
  partial?: string;
};

export function FileTree({ annotate = "none", active, partial = "" }: Props) {
  return (
    <ul className={`tree ${annotate === "gate" ? "is-annotated" : ""}`}>
      <li>
        <span className="tree-row is-root">{PROJECT}/</span>
        <ul>
          {NODES.map((n) => (
            <TreeNode key={n.name} node={n} annotate={annotate} active={active} partial={partial} />
          ))}
        </ul>
      </li>
    </ul>
  );
}

function TreeNode({
  node,
  annotate,
  active,
  partial,
}: {
  node: Node;
  annotate: "none" | "gate";
  active?: string[];
  partial: string;
}) {
  const outcome = outcomeOf(node);
  const usable = outcome === "command";
  const depth = node.path.length - 1;
  const onBranch = !!active && usable && node.path.every((seg, i) => active[i] === seg);
  const isCandidate =
    !!active &&
    usable &&
    depth === active.length &&
    node.path.slice(0, -1).every((seg, i) => active[i] === seg) &&
    node.name.startsWith(partial);
  const cls = [
    "tree-row",
    !usable ? `is-${outcome}` : node.dir ? "is-group" : "is-command",
    onBranch && depth < active!.length ? "is-active" : "",
    isCandidate ? "is-candidate" : "",
  ].join(" ");
  const lang = node.entry?.lang;
  const showNote = annotate === "gate" && !!node.entry;

  return (
    <li>
      <span className="tree-line">
        <span className={cls}>
          {node.name}
          {node.dir ? "/" : ""}
        </span>
        {lang && (
          <span className={`lang lang-${lang}`} title={node.entry?.why}>
            {LANG_LABEL[lang]}
          </span>
        )}
        {showNote && (
          <span className={`tree-note is-${outcome}`}>
            {outcome === "command" ? cmd(node.entry!.path) : NOTE[outcome]}
            <span className="tree-why">{node.entry!.why}</span>
          </span>
        )}
      </span>
      {node.dir && node.children.length > 0 && (
        <ul>
          {node.children.map((c) => (
            <TreeNode key={c.name} node={c} annotate={annotate} active={active} partial={partial} />
          ))}
        </ul>
      )}
    </li>
  );
}
