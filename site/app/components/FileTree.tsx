type Node = { name: string; path: string[]; children: Node[] };

function build(paths: string[]): Node[] {
  const root: Node = { name: "", path: [], children: [] };
  for (const p of paths) {
    let node = root;
    for (const part of p.split("/")) {
      let child = node.children.find((c) => c.name === part);
      if (!child) {
        child = { name: part, path: [...node.path, part], children: [] };
        node.children.push(child);
      }
      node = child;
    }
  }
  return root.children;
}

type Props = {
  paths: string[];
  /** Completed tokens the user has typed — highlights the branch. Omit for no highlighting. */
  active?: string[];
  /** The token being typed — highlights matching siblings. */
  partial?: string;
  rootName?: string;
};

export function FileTree({ paths, active, partial = "", rootName = "acme/" }: Props) {
  const nodes = build(paths);
  return (
    <ul className="tree">
      <li>
        <span className="tree-row is-root">{rootName}</span>
        <ul>
          <li>
            <span className="tree-row is-config">
              monom <em>config</em>
            </span>
          </li>
          {nodes.map((n) => (
            <TreeNode key={n.name} node={n} active={active} partial={partial} />
          ))}
        </ul>
      </li>
    </ul>
  );
}

function TreeNode({ node, active, partial }: { node: Node; active?: string[]; partial: string }) {
  const depth = node.path.length - 1;
  const onBranch = !!active && node.path.every((seg, i) => active[i] === seg);
  const isCandidate =
    !!active &&
    depth === active.length &&
    node.path.slice(0, -1).every((seg, i) => active[i] === seg) &&
    node.name.startsWith(partial);
  const isGroup = node.children.length > 0;
  const cls = [
    "tree-row",
    isGroup ? "is-group" : "is-command",
    onBranch && depth < active!.length ? "is-active" : "",
    isCandidate ? "is-candidate" : "",
  ].join(" ");

  return (
    <li>
      <span className={cls}>
        {node.name}
        {isGroup ? "/" : ""}
      </span>
      {isGroup && (
        <ul>
          {node.children.map((c) => (
            <TreeNode key={c.name} node={c} active={active} partial={partial} />
          ))}
        </ul>
      )}
    </li>
  );
}
