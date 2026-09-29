import { useCallback, useEffect, useRef, useState } from "react";
import { commonPrefix, filter } from "~/lib/filter";
import { CHECK, commands, paths } from "~/lib/project";
import { FileTree } from "./FileTree";

type Line =
  | { kind: "prompt"; text: string; tab?: boolean }
  | { kind: "out"; text: string }
  | { kind: "err"; text: string }
  | { kind: "list"; items: { name: string; group: boolean }[] };

const PROMPT = "~/acme $";
const MAX_LINES = 14;

const isGroup = (prefix: string) => paths.some((p) => p.startsWith(prefix + "/"));

function wordsOf(input: string) {
  const tokens = input.split(" ").filter(Boolean);
  const words = tokens.slice(1);
  if (input.endsWith(" ")) words.push("");
  return { head: tokens[0] ?? "", words };
}

// What `monom()` prints for a command group, matching src/monom.
function groupMessage(label: string, tokens: string[]): Line[] {
  const children = filter(paths, [...tokens, ""]);
  const lines: Line[] = [{ kind: "err", text: `monom: ${label}` }];
  if (children.length) lines.push({ kind: "err", text: `available: ${children.join(", ")}` });
  return lines;
}

function run(input: string): Line[] {
  const trimmed = input.trim();
  if (trimmed === "") return [];
  if (trimmed === "mnmd check") {
    return CHECK.split("\n").slice(1).map((text) => ({ kind: text.startsWith("warning") ? ("err" as const) : ("out" as const), text }));
  }
  if (trimmed === "mnmd root") return [{ kind: "out", text: "/home/you/acme" }];
  const tokens = trimmed.split(/\s+/);
  if (tokens[0] !== "monom") {
    return [{ kind: "err", text: `bash: ${tokens[0]}: try 'monom' and press Tab` }];
  }
  const args = tokens.slice(1);
  if (args.length === 0) return groupMessage("monom", []);
  const path = args.join("/");
  if (commands[path]) return [{ kind: "out", text: commands[path] }];
  if (isGroup(path)) return groupMessage(`'${args[args.length - 1]}' is a command group`, args);
  return [{ kind: "err", text: `mnmd pack: pack: command not found: /home/you/acme/${path}` }];
}

type TabResult = { input: string; list?: Line };

function complete(input: string): TabResult {
  if ("monom".startsWith(input.trim()) && !input.includes(" ")) return { input: "monom " };
  const { head, words } = wordsOf(input);
  if (head !== "monom") return { input };
  if (words.length === 0) return { input: input + " " };

  const partial = words[words.length - 1];
  const prefix = words.slice(0, -1);
  const matches = filter(paths, words);
  if (matches.length === 0) return { input };

  const base = input.slice(0, input.length - partial.length);
  if (matches.length === 1) return { input: base + matches[0] + " " };

  const shared = commonPrefix(matches);
  if (shared.length > partial.length) return { input: base + shared };
  return {
    input,
    list: {
      kind: "list",
      items: matches.map((name) => ({ name, group: isGroup([...prefix, name].join("/")) })),
    },
  };
}

type Step = { type: string } | { tab: true } | { enter: true } | { pause: number };

// Real bash behaviour on lib/project.ts: an ambiguous word lists on Tab, a
// unique one completes and adds a space.
const SCRIPT: Step[] = [
  { type: "monom " },
  { tab: true },
  { pause: 700 },
  { type: "in" },
  { tab: true },
  { pause: 350 },
  { tab: true },
  { pause: 700 },
  { type: "c" },
  { tab: true },
  { pause: 350 },
  { tab: true },
  { pause: 600 },
  { enter: true },
  { pause: 900 },
  { type: "monom d" },
  { tab: true },
  { pause: 350 },
  { tab: true },
  { pause: 700 },
  { type: "m" },
  { tab: true },
  { pause: 500 },
  { enter: true },
];

export function TabDemo() {
  const [lines, setLines] = useState<Line[]>([]);
  const [input, setInput] = useState("");
  const [flash, setFlash] = useState<string | null>(null);
  const [live, setLive] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const autoplay = useRef(true);

  const push = useCallback((more: Line[]) => {
    setLines((prev) => [...prev, ...more].slice(-MAX_LINES));
  }, []);

  const pressTab = useCallback(
    (current: string) => {
      const result = complete(current);
      if (result.list) push([{ kind: "prompt", text: current, tab: true }, result.list]);
      if (result.input === current && !result.list) {
        setFlash("bell");
        window.setTimeout(() => setFlash(null), 180);
      }
      setInput(result.input);
      return result.input;
    },
    [push],
  );

  const pressEnter = useCallback(
    (current: string) => {
      if (current.trim() === "clear") {
        setLines([]);
      } else {
        push([{ kind: "prompt", text: current }, ...run(current)]);
      }
      setInput("");
    },
    [push],
  );

  useEffect(() => {
    const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (reduced) {
      autoplay.current = false;
      setLive(true);
      return;
    }
    let cancelled = false;
    let current = "";
    const timers: number[] = [];
    const wait = (ms: number) =>
      new Promise<void>((resolve) => timers.push(window.setTimeout(resolve, ms)));

    (async () => {
      await wait(900);
      for (const step of SCRIPT) {
        if (cancelled || !autoplay.current) return;
        if ("type" in step) {
          for (const ch of step.type) {
            if (cancelled || !autoplay.current) return;
            current += ch;
            setInput(current);
            await wait(70 + Math.random() * 60);
          }
        } else if ("tab" in step) {
          setFlash("tab");
          current = pressTab(current);
          await wait(160);
          setFlash(null);
        } else if ("enter" in step) {
          pressEnter(current);
          current = "";
        } else {
          await wait(step.pause);
        }
      }
      if (!cancelled) setLive(true);
    })();

    return () => {
      cancelled = true;
      timers.forEach(clearTimeout);
    };
  }, [pressTab, pressEnter]);

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight });
  }, [lines, input]);

  const takeOver = () => {
    if (autoplay.current) {
      autoplay.current = false;
      setInput("");
      setLive(true);
    }
  };

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    takeOver();
    if (e.key === "Tab") {
      e.preventDefault();
      pressTab(input);
      return;
    }
    if (e.key === "Enter") {
      e.preventDefault();
      pressEnter(input);
    } else if (e.key === "l" && e.ctrlKey) {
      e.preventDefault();
      setLines([]);
    }
  };

  const { head, words } = wordsOf(input);
  const typed = words.filter(Boolean);
  const highlighting = head === "monom" && (words.length > 0 || input.endsWith(" "));
  const activePath = input.endsWith(" ") ? typed : typed.slice(0, -1);
  const partial = input.endsWith(" ") ? "" : (typed[typed.length - 1] ?? "");

  return (
    <div className="demo">
      <div className="demo-tree" aria-hidden="true">
        <div className="panel-label">file tree</div>
        <FileTree active={highlighting ? activePath : undefined} partial={partial} />
      </div>
      <div
        className={`terminal ${flash === "bell" ? "is-bell" : ""}`}
        onClick={() => inputRef.current?.focus({ preventScroll: true })}
      >
        <div className="terminal-bar">
          <span className="dot" />
          <span className="dot" />
          <span className="dot" />
          <span className="terminal-title">command tree</span>
        </div>
        <div className="terminal-body" ref={scrollRef} aria-live="polite">
          {lines.map((line, i) => (
            <TerminalLine key={i} line={line} />
          ))}
          <label className="line prompt-line">
            <span className="ps1">{PROMPT}</span>
            <span className="input-wrap">
              <span className="input-mirror">
                {input}
                <span className="caret" />
              </span>
              <input
                ref={inputRef}
                value={input}
                onChange={(e) => {
                  takeOver();
                  setInput(e.target.value);
                }}
                onKeyDown={onKeyDown}
                onFocus={takeOver}
                spellCheck={false}
                autoCapitalize="off"
                autoComplete="off"
                autoCorrect="off"
                aria-label="Try monom: type a command and press Tab"
              />
            </span>
          </label>
        </div>
        <div className="terminal-foot">
          <span className={live ? "hint is-live" : "hint"}>
            {live ? (
              <>
                your turn — type <kbd>monom</kbd>, press <kbd>Tab</kbd>
              </>
            ) : (
              "autoplaying — click to take over"
            )}
          </span>
          <button
            type="button"
            className={`tab-key ${flash === "tab" ? "is-down" : ""}`}
            onClick={(e) => {
              e.stopPropagation();
              takeOver();
              pressTab(autoplay.current ? "" : input);
              inputRef.current?.focus({ preventScroll: true });
            }}
          >
            Tab ⇥
          </button>
        </div>
      </div>
    </div>
  );
}

function TerminalLine({ line }: { line: Line }) {
  if (line.kind === "prompt") {
    return (
      <div className="line">
        <span className="ps1">{PROMPT}</span> {line.text}
        {line.tab && <span className="tab-mark">⇥ tab</span>}
      </div>
    );
  }
  if (line.kind === "list") {
    return (
      <div className="line completions">
        {line.items.map((item) => (
          <span key={item.name} className={item.group ? "is-group" : "is-command"}>
            {item.name}
          </span>
        ))}
      </div>
    );
  }
  return <div className={`line ${line.kind === "err" ? "is-err" : "is-out"}`}>{line.text}</div>;
}
