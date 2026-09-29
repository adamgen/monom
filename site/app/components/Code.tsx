import { Fragment, type ReactNode } from "react";

// A deliberately small shell highlighter: comments, strings, variables,
// keywords, and the first word of a `$ ` prompt line. Enough for the snippets
// on this site; not a general-purpose lexer.
const KEYWORDS = new Set([
  "case", "esac", "in", "if", "then", "else", "fi", "for", "do", "done",
  "echo", "exit", "source", "export", "printf", "find", "sed", "sort", "cd",
]);

const TOKEN = /(#.*$)|("(?:[^"\\]|\\.)*"|'[^']*')|(\$\{?[A-Za-z_@0-9#*]+\}?|\$\()|([A-Za-z_][\w-]*)|(\s+)|(.)/gm;

function highlightLine(line: string): ReactNode[] {
  if (line.startsWith("$ ")) {
    return [
      <span key="p" className="tok-ps1">$ </span>,
      <span key="c" className="tok-cmd">{line.slice(2)}</span>,
    ];
  }
  if (line.startsWith("#!")) return [<span key="s" className="tok-comment">{line}</span>];

  const out: ReactNode[] = [];
  let m: RegExpExecArray | null;
  let i = 0;
  TOKEN.lastIndex = 0;
  while ((m = TOKEN.exec(line)) !== null) {
    const [text, comment, str, variable, word] = m;
    const k = i++;
    if (comment) out.push(<span key={k} className="tok-comment">{text}</span>);
    else if (str) out.push(<span key={k} className="tok-string">{text}</span>);
    else if (variable) out.push(<span key={k} className="tok-var">{text}</span>);
    else if (word && KEYWORDS.has(word)) out.push(<span key={k} className="tok-kw">{text}</span>);
    else out.push(<Fragment key={k}>{text}</Fragment>);
    if (text === "") TOKEN.lastIndex++;
  }
  return out;
}

export function Code({ children, title, output }: { children: string; title?: string; output?: boolean }) {
  const lines = children.replace(/^\n/, "").replace(/\n\s*$/, "").split("\n");
  return (
    <figure className="code">
      {title && <figcaption>{title}</figcaption>}
      <pre>
        <code>
          {lines.map((line, i) => (
            <div key={i} className={output && !line.startsWith("$ ") ? "tok-out" : undefined}>
              {output && !line.startsWith("$ ") ? line : highlightLine(line)}
              {line === "" ? " " : null}
            </div>
          ))}
        </code>
      </pre>
    </figure>
  );
}
