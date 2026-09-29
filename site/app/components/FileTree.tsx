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

/** Which part of a mapped command a tree node is: a folder at some depth, or the file. */
export const segClass = (i: number, last: boolean) => (last ? "seg-file" : `seg-${Math.min(i, 2)}`);

type Props = {
  /**
   * How the same tree is drawn:
   * - "plain": monochrome names only (no badges, colors or annotations)
   * - "colored": groups, commands and skipped entries colored, with language badges
   * - "linked": colored, and the path of `mapped` lit segment by segment in the
   *   colors the command line uses; command files report hover/tap via `onPick`
   */
  variant?: "plain" | "colored" | "linked";
  /** Colored only: the discovery gate's decision and reason beside each file. */
  annotate?: boolean;
  /** Completed tokens the user has typed — highlights the branch (the Tab demo). */
  active?: string[];
  /** The token being typed — highlights matching siblings. */
  partial?: string;
  /** Linked only: the command path being shown, and how many of its segments are lit. */
  mapped?: { path: string; lit: number };
  /** Linked only: a command file was hovered, focused or tapped. */
  onPick?: (path: string) => void;
};

type Ctx = Required<Pick<Props, "variant" | "annotate" | "partial">> & Omit<Props, "variant" | "annotate" | "partial">;

export function FileTree({ variant = "colored", annotate = false, active, partial = "", mapped, onPick }: Props) {
  const ctx: Ctx = { variant, annotate: annotate && variant !== "plain", active, partial, mapped, onPick };
  return (
    <ul className={`tree is-${variant} ${ctx.annotate ? "is-annotated" : ""}`}>
      <li>
        <span className="tree-row is-root">{PROJECT}/</span>
        <ul>
          {NODES.map((n) => (
            <TreeNode key={n.name} node={n} ctx={ctx} />
          ))}
        </ul>
      </li>
    </ul>
  );
}

function TreeNode({ node, ctx }: { node: Node; ctx: Ctx }) {
  const { variant, annotate, active, partial, mapped, onPick } = ctx;
  const plain = variant === "plain";
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
  const segs = mapped?.path.split("/") ?? [];
  const lit =
    variant === "linked" && !!mapped && depth < mapped.lit && node.path.every((seg, i) => segs[i] === seg);
  const pickable = variant === "linked" && !!onPick && !node.dir && usable && !!node.entry;
  const cls = [
    "tree-row",
    plain ? "" : !usable ? `is-${outcome}` : node.dir ? "is-group" : "is-command",
    onBranch && depth < active!.length ? "is-active" : "",
    isCandidate ? "is-candidate" : "",
    lit ? `is-lit ${segClass(depth, depth === segs.length - 1)}` : "",
    pickable ? "is-pickable" : "",
  ].join(" ");
  const lang = plain ? undefined : node.entry?.lang;
  const showNote = annotate && !!node.entry;
  const pick = pickable ? () => onPick!(node.entry!.path) : undefined;

  return (
    <li>
      <span className="tree-line">
        <span
          className={cls}
          {...(pickable
            ? {
                role: "button",
                tabIndex: 0,
                "aria-label": `Show ${cmd(node.entry!.path)}`,
                onMouseEnter: pick,
                onFocus: pick,
                onClick: pick,
              }
            : {})}
        >
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
            <TreeNode key={c.name} node={c} ctx={ctx} />
          ))}
        </ul>
      )}
    </li>
  );
}
