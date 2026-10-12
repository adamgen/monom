import { useCallback, useEffect, useRef, useState } from "react";
import { commands, paths } from "~/lib/project";
import { FileTree, segClass } from "./FileTree";

// Stage 3 of the adoption story: the tree on the left, a terminal on the
// right. It types each real command in turn and lights the matching path in
// the tree, one segment at a time, in the same color on both sides. Hovering,
// focusing or tapping a command file in the tree shows its command instead.

const PROMPT = "~/acme $";
const TYPE_MS = 55;
const HOLD_MS = 1900;
const RESUME_MS = 5000;

type Shown = { path: string; typed: number; done: boolean };

/** "db/migrate.py" → [" db", " migrate.py"], the words after "monom" with their leading space. */
const wordsOf = (path: string) => path.split("/").map((w) => ` ${w}`);
const lengthOf = (path: string) => "monom".length + wordsOf(path).join("").length;

export function PathMap() {
  // The server render and a reduced-motion visitor both see a finished first
  // command; the cycle starts only in the browser.
  const [shown, setShown] = useState<Shown>({ path: paths[0], typed: lengthOf(paths[0]), done: true });
  const [still, setStill] = useState(false);
  // Earlier commands stay on screen, dimmed, like a real scrollback.
  const [history, setHistory] = useState<string[]>([]);
  const paused = useRef(false);
  const resumeTimer = useRef<number>();
  const index = useRef(0);

  useEffect(() => {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      setStill(true);
      return;
    }
    let cancelled = false;
    let timer = 0;
    const step = (path: string, typed: number) => {
      if (cancelled) return;
      if (paused.current) {
        timer = window.setTimeout(() => step(path, typed), 250);
        return;
      }
      const total = lengthOf(path);
      if (typed < total) {
        setShown({ path, typed, done: false });
        timer = window.setTimeout(() => step(path, typed + 1), TYPE_MS);
      } else {
        setShown({ path, typed: total, done: true });
        timer = window.setTimeout(() => {
          setHistory((h) => [...h, path].slice(-2));
          index.current = (index.current + 1) % paths.length;
          step(paths[index.current], "monom".length);
        }, HOLD_MS);
      }
    };
    timer = window.setTimeout(() => step(paths[index.current], "monom".length), 600);
    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, []);

  const pick = useCallback((path: string) => {
    paused.current = true;
    index.current = paths.indexOf(path);
    setHistory([]);
    setShown({ path, typed: lengthOf(path), done: true });
    window.clearTimeout(resumeTimer.current);
    resumeTimer.current = window.setTimeout(() => {
      paused.current = false;
    }, RESUME_MS);
  }, []);

  useEffect(() => () => window.clearTimeout(resumeTimer.current), []);

  const words = wordsOf(shown.path);
  // How many path segments are fully typed; the one being typed lights too.
  let seen = "monom".length;
  let lit = 0;
  for (const w of words) {
    if (shown.typed > seen) lit++;
    seen += w.length;
  }

  return (
    <div className={`pathmap ${still ? "is-still" : ""}`}>
      <div className="pathmap-tree">
        <div className="panel-label">file tree</div>
        <FileTree variant="linked" mapped={{ path: shown.path, lit }} onPick={pick} />
        <p className="pathmap-hint">Hover or tap a file to see its command.</p>
      </div>
      <div className="terminal pathmap-term" aria-live="polite">
        <div className="terminal-bar">
          <span className="dot" />
          <span className="dot" />
          <span className="dot" />
          <span className="terminal-title">command line</span>
        </div>
        <div className="terminal-body">
          {still ? (
            paths.map((p) => <CommandLine key={p} path={p} typed={lengthOf(p)} done onPick={pick} />)
          ) : (
            <>
              {history.map((p, i) => (
                <div key={`${i}-${p}`} className="pathmap-past">
                  <CommandLine path={p} typed={lengthOf(p)} done />
                </div>
              ))}
              <CommandLine path={shown.path} typed={shown.typed} done={shown.done} />
            </>
          )}
        </div>
        {!still && (
          <div className="terminal-foot pathmap-foot" aria-hidden="true">
            <span className="pathmap-foot-label">runs</span>
            <span className="mono">
              ./
              {shown.path.split("/").map((seg, i, all) => (
                <span key={i}>
                  <span className={i < lit ? `is-lit ${segClass(i, i === all.length - 1)}` : "is-dim"}>{seg}</span>
                  {i < all.length - 1 ? "/" : ""}
                </span>
              ))}
            </span>
          </div>
        )}
      </div>
    </div>
  );
}

function CommandLine({
  path,
  typed,
  done,
  onPick,
}: {
  path: string;
  typed: number;
  done: boolean;
  onPick?: (path: string) => void;
}) {
  const words = wordsOf(path);
  let left = typed - "monom".length;
  return (
    <div className="pathmap-cmd" onMouseEnter={onPick ? () => onPick(path) : undefined}>
      <div className="line">
        <span className="ps1">{PROMPT}</span> monom
        {words.map((w, i) => {
          const part = w.slice(0, Math.max(0, left));
          left -= w.length;
          return part ? (
            <span key={i} className={`is-lit ${segClass(i, i === words.length - 1)}`}>
              {part}
            </span>
          ) : null;
        })}
        {!done && <span className="caret" />}
      </div>
      {done && <div className="line is-out">{commands[path]}</div>}
    </div>
  );
}
